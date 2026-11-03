#!/usr/bin/env python3
"""tool_name → dws 指令 映射测试脚本

背景：market_mcp_tools 表中的 MCP tool_name 需要在 DWS CLI 里找到对应指令并验证。

工作流程：
1. 用 `dws schema --all -f json` 生成实时 Schema，按 interface_ref.rpc_name 与
   name 建立 → CLI 路径索引（即官方声明式映射）。
2. 按内置 tool_name 清单解析：优先取手工覆盖表（schema 未声明 rpc 或历史改名），
   否则取 Schema 索引结果；每个 tool 得到「主指令 + 备选指令」。
3. 对每个唯一指令执行 `dws <cmd> --help`，退出码为 0 表示命令存在于真实 Cobra 树。
4. 输出映射 CSV 与汇总表。校验失败时以非零码退出。

用法:
  python3 scripts/dev/test_tool_name_mapping.py             # 默认用仓库根 ./dws
  python3 scripts/dev/test_tool_name_mapping.py --dws /path/to/dws
  python3 scripts/dev/test_tool_name_mapping.py --csv /tmp/map.csv
  python3 scripts/dev/test_tool_name_mapping.py --strict     # 未封装工具也视为失败
  python3 scripts/dev/test_tool_name_mapping.py --no-help-check  # 只解析映射不跑 --help
"""

import argparse
import csv
import json
import os
import subprocess
import sys

# ---------------------------------------------------------------------------
# 用户提供的 tool_name 清单（按原始顺序，含 doc/drive 双归属的 create_folder、
# download_file 各出现两次；本脚本自动去重并合并指令）。
# ---------------------------------------------------------------------------
TOOL_NAMES = """\
get_calendar
get_calendar_detail
get_calendar_participants
list_bases
search_bases
get_base
create_base
update_base
delete_base
get_tables
create_table
update_table
delete_table
get_fields
create_fields
update_field
delete_field
query_records
create_records
update_records
delete_records
search_templates
prepare_attachment_upload
export_data
prepare_import_upload
import_data
get_views
create_view
update_view
delete_view
list_calendar_events
query_busy_status
query_available_meeting_room
list_meeting_room_groups
create_calendar_event
update_calendar_event
add_calendar_participant
add_meeting_room
delete_meeting_room
remove_calendar_participant
delete_calendar_event
get_current_user_profile
search_user_by_key_word
search_contact_by_key_word
search_user_by_mobile
get_user_info_by_user_ids
search_dept_by_keyword
get_sub_depts_by_dept_id
get_dept_members_by_deptId
get_org_labels
search_label_by_name
get_label_members_by_labelId
get_user_todos_in_current_org
query_todo_detail
create_personal_todo
update_todo_task
update_todo_done_status
delete_todo
search_documents
list_nodes
get_document_info
get_document_content
list_document_blocks
create_document
create_folder
create_file
update_document
insert_document_block
update_document_block
delete_document_block
get_file_upload_info
commit_uploaded_file
download_file
list_user_mailboxes
search_emails
get_email_by_message_id
send_email
list_user_visible_process
list_pending_approvals
get_processInstance_detail
get_processInstance_records
list_initiated_instances
list_pending_tasks
approve_processInstance
reject_processInstance
revoke_processInstance
search_groups_by_keyword
get_group_members
list_conversation_message
list_conversation_message_v2
list_individual_chat_message
list_topic_replies
create_internal_org_group
create_internal_group
update_group_name
add_group_member
send_message_as_user
send_direct_message_as_user
send_group_message_as_user_test
remove_group_member
batch_send_robot_msg_to_users
send_robot_group_message
send_message_by_custom_robot
batch_recall_robot_users_msg
recall_robot_group_message
add_robot_to_group
send_ding_message
recall_ding_message
search_my_robots
list_my_created_minutes
list_by_keyword_range
list_shared_minutes
get_minutes_basic_info
get_minutes_keywords
batch_get_minutes_details
get_minutes_ai_summary
get_minutes_transcription
list_minutes_todos
update_minutes_title
update_minutes_summary
create_mind_graph
query_mind_graph_status
replace_speaker
add_personal_hot_word
replace_minutes_text
create_upload_session
complete_upload_session
cancel_upload_session
get_available_report_templates
get_template_details_by_name
get_received_report_list
get_send_report_list
get_report_entry_details
get_report_statistics_by_id
create_report
query_attendance_group_or_rules
get_user_attendance_record
batch_get_employee_shifts
get_attendance_summary
list_files
list_spaces
get_file_info
get_upload_info
commit_upload
search_open_platform_docs
search_open_platform_error_code
get_my_lives
create_group_conversation
enterprise_person_search
search_enterprise_behavior
search_enterprise_group
"""

# ---------------------------------------------------------------------------
# 手工覆盖表：schema 里 rpc 改名 / 未声明 interface_ref / 复合调用的工具。
# 值: (主指令, 说明)。说明会写进 CSV 的 note 列。
# ---------------------------------------------------------------------------
OVERRIDES = {
    # contact —— 服务端旧名 / schema 未声明 rpc 的原子命令
    "search_user_by_key_word": ("contact user search",
                                "服务端现名 search_contact_by_key_word；同族 shortcut: contact +search-user"),
    "get_org_labels": ("contact label list",
                       "原子命令；shortcut: contact +list-roles"),
    "search_label_by_name": ("contact label get", "原子命令"),
    "get_label_members_by_labelId": ("contact label list-members",
                                     "原子命令；shortcut: contact +list-role-members"),
    # todo —— rpc 改名
    "query_todo_detail": ("todo task get", "rpc 现名 get_todo_detail"),
    # doc —— 复合两步上传
    "get_file_upload_info": ("doc upload", "doc upload 内部第一步调用"),
    "commit_uploaded_file": ("doc upload", "doc upload 内部第二步调用"),
    # chat/bot —— 底层 rpc 与 schema 声明不一致或走 bot 服务
    "search_groups_by_keyword": ("chat search", "内部走 im/search_groups"),
    "get_group_members": ("chat group members", "原子命令；shortcut: chat +group-members"),
    "list_conversation_message": ("chat message list",
                                  "DWS 已切换到 v2（list_conversation_message_v2）"),
    "send_robot_group_message": ("chat message send-by-bot", "bot 服务"),
    "batch_send_robot_msg_to_users": ("chat +messages-batch-send-by-bot", "bot 服务"),
    "recall_robot_group_message": ("chat message recall-by-bot", "bot 服务"),
    "batch_recall_robot_users_msg": ("chat +messages-batch-recall-by-bot", "bot 服务"),
    # minutes —— rpc 统一为 list_by_keyword_and_time_range
    "list_my_created_minutes": ("minutes list mine", "rpc list_by_keyword_and_time_range"),
    "list_by_keyword_range": ("minutes list all", "rpc list_by_keyword_and_time_range（可访问范围）"),
}

# ---------------------------------------------------------------------------
# 未封装工具：market 有该 tool_name，但 DWS 没有对应指令（只记录近似命令，不校验）。
# ---------------------------------------------------------------------------
UNWRAPPED = {
    "create_internal_group": "未封装；近似 chat group create",
    "create_internal_org_group": "未封装；近似 chat group create",
    "send_message_as_user": "未封装；近似 chat message send",
    "send_direct_message_as_user": "未封装；近似 chat message send",
    "send_group_message_as_user_test": "未封装（内部测试工具）；近似 chat message send",
    "search_open_platform_error_code": "未封装；近似 devdoc article search",
    "search_enterprise_group": "未封装（ai-search 服务端共 5 个工具，DWS 只封装 3 个）",
}

# ---------------------------------------------------------------------------
# 一个 rpc 命中多个 leaf 时的主指令选择（未列出的默认取第一个非别名、非 +shortcut）。
# ---------------------------------------------------------------------------
PRIMARY_PREF = {
    "list_bases": "aitable base list",
    "search_bases": "aitable base search",
    "get_base": "aitable base get",
    "create_base": "aitable base create",
    "get_tables": "aitable table list",
    "get_fields": "aitable field list",
    "delete_field": "aitable field delete",
    "query_records": "aitable record query",
    "get_views": "aitable view list",
    "update_view": "aitable view update name",
    "create_folder": "doc folder create",
    "download_file": "doc download",
    "create_upload_session": "minutes upload create",
    "search_documents": "doc search",
    "list_nodes": "doc list",
    "create_file": "doc file create",
    "search_open_platform_docs": "devdoc article search",
    "create_group_conversation": "chat group create",
}


def repo_root() -> str:
    return os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def default_dws_bin() -> str:
    return os.path.join(repo_root(), "dws")


def load_schema(dws_bin: str):
    """运行 dws schema --all -f json 并返回 tools 列表。"""
    proc = subprocess.run(
        [dws_bin, "schema", "--all", "-f", "json"],
        capture_output=True, text=True, timeout=120,
    )
    if proc.returncode != 0:
        sys.stderr.write("dws schema --all 失败(exit=%d):\n%s\n%s\n"
                         % (proc.returncode, proc.stdout[-2000:], proc.stderr[-2000:]))
        sys.exit(2)
    data = json.loads(proc.stdout)
    tools = []
    for p in data.get("products", []):
        pid = p.get("id", "")
        for t in p.get("tools", []):
            ir = t.get("interface_ref") or {}
            rpc = ir.get("rpc_name") or "" if isinstance(ir, dict) else ""
            tools.append({
                "rpc": rpc,
                "svc": ir.get("product_id") or "" if isinstance(ir, dict) else "",
                "name": t.get("name") or "",
                "cli": t.get("cli_path") or "",
                "canonical": t.get("canonical_path") or "",
                "product": pid,
                "alias": bool(t.get("is_alias")),
            })
    return tools


def build_index(tools):
    by_rpc, by_name = {}, {}
    for t in tools:
        if t["rpc"]:
            by_rpc.setdefault(t["rpc"], []).append(t)
        if t["name"]:
            by_name.setdefault(t["name"], []).append(t)
    return by_rpc, by_name


def pick_primary(cands, tool, pref):
    if tool in pref:
        for c in cands:
            if c["cli"] == pref[tool]:
                return c
    for c in cands:  # 优先非别名、非 +shortcut 的命令
        if not c["alias"] and "+" not in c["cli"]:
            return c
    for c in cands:  # 其次非别名
        if not c["alias"]:
            return c
    return cands[0]


def resolve_tool(tool, by_rpc, by_name):
    """返回 (primary, alts, note, status)。status ∈ mapped|unwrapped。"""
    if tool in UNWRAPPED:
        return None, [], UNWRAPPED[tool], "unwrapped"
    if tool in OVERRIDES:
        cmd, note = OVERRIDES[tool]
        return {"cli": cmd, "canonical": "", "product": "", "rpc": "", "svc": ""}, [], note, "mapped"
    cands = by_rpc.get(tool, []) or by_name.get(tool, [])
    if not cands:
        return None, [], "未在 schema 中命中且无手工覆盖", "unmapped"
    primary = pick_primary(cands, tool, PRIMARY_PREF)
    alts = [c for c in cands if c is not primary]
    return primary, alts, "", "mapped"


def help_ok(dws_bin, cmd, cache):
    if cmd in cache:
        return cache[cmd]
    proc = subprocess.run([dws_bin] + cmd.split() + ["--help"],
                          capture_output=True, text=True, timeout=60)
    ok = proc.returncode == 0
    cache[cmd] = ok
    return ok


def dedup_cmds(primary, alts):
    cmds = []
    for c in ([primary] if primary else []) + alts:
        if c["cli"] and c["cli"] not in cmds:
            cmds.append(c["cli"])
    return cmds


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--dws", default=None, help="dws 二进制路径（默认仓库根 ./dws）")
    ap.add_argument("--csv", default=None,
                    help="映射 CSV 输出路径（默认 <脚本目录>/tool_name_to_cmd.csv）")
    ap.add_argument("--strict", action="store_true",
                    help="存在未封装/未命中工具时以失败码退出")
    ap.add_argument("--no-help-check", action="store_true",
                    help="跳过 dws --help 校验，只输出解析出的映射")
    args = ap.parse_args()

    dws_bin = args.dws or default_dws_bin()
    if not os.path.exists(dws_bin):
        sys.stderr.write("找不到 dws 二进制: %s（可用 --dws 指定）\n" % dws_bin)
        sys.exit(2)

    names = [l.strip() for l in TOOL_NAMES.splitlines() if l.strip()]
    seen, ordered = set(), []
    for n in names:  # 去重保序（create_folder / download_file 各出现两次）
        if n not in seen:
            seen.add(n)
            ordered.append(n)

    tools = load_schema(dws_bin)
    by_rpc, by_name = build_index(tools)

    cache = {}
    rows, fails, strict_fails = [], 0, 0
    for tool in ordered:
        primary, alts, note, status = resolve_tool(tool, by_rpc, by_name)
        if status == "unmapped":
            rows.append({"tool": tool, "cmd": "", "alts": "", "product": "",
                         "rpc": "", "note": note, "status": "未命中"})
            strict_fails += 1
            continue
        if status == "unwrapped":
            rows.append({"tool": tool, "cmd": "", "alts": "", "product": "",
                         "rpc": "", "note": note, "status": "未封装"})
            strict_fails += 1
            continue
        cmds = dedup_cmds(primary, alts)
        verified = cmds
        if not args.no_help_check:
            verified = [c for c in cmds if help_ok(dws_bin, c, cache)]
            if not verified:
                fails += 1
        rows.append({
            "tool": tool,
            "cmd": primary["cli"],
            "alts": ";".join(c for c in cmds if c != primary["cli"]),
            "product": primary["product"] or primary.get("svc", ""),
            "rpc": primary["rpc"],
            "note": note,
            "status": "OK" if (args.no_help_check or len(verified) == len(cmds)) else "FAIL",
        })

    csv_path = args.csv or os.path.join(os.path.dirname(os.path.abspath(__file__)),
                                        "tool_name_to_cmd.csv")
    with open(csv_path, "w", newline="") as f:
        w = csv.writer(f)
        w.writerow(["tool_name", "dws_cmd", "alt_cmds", "product", "rpc", "note", "status"])
        for r in rows:
            w.writerow([r["tool"], r["cmd"], r["alts"], r["product"], r["rpc"],
                        r["note"], r["status"]])

    # ---- 汇总输出 ----
    ok = sum(1 for r in rows if r["status"] == "OK")
    unwrapped = sum(1 for r in rows if r["status"] == "未封装")
    unmapped = sum(1 for r in rows if r["status"] == "未命中")
    failed = sum(1 for r in rows if r["status"] == "FAIL")
    print("tool_name 总数: %d（去重后 %d）" % (len(names), len(ordered)))
    print("映射成功: %d | 未封装(无对应指令): %d | 未命中: %d | --help 校验失败: %d"
          % (ok, unwrapped, unmapped, failed))
    print("CSV 已写入: %s\n" % csv_path)

    for r in rows:
        mark = {"OK": "ok ", "FAIL": "FAIL", "未封装": "--- ", "未命中": "???"}[r["status"]]
        line = "%-4s %-34s -> %s" % (mark, r["tool"], r["cmd"] or "-")
        if r["alts"]:
            line += "  (alt: %s)" % r["alts"]
        if r["note"]:
            line += "  [%s]" % r["note"]
        print(line)

    if fails:
        print("\n存在 --help 校验失败的指令，请修复映射。")
        sys.exit(1)
    if args.strict and strict_fails:
        print("\n--strict: 存在未封装/未命中工具。")
        sys.exit(1)
    print("\n全部校验通过。")


if __name__ == "__main__":
    main()
