package helpers

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestRunAllEduCommands 逐条执行 5 个教育产品的全部 156 条叶子命令（dry-run 模式），
// 并打印每条命令对应的 MCP 工具名和参数。相当于在终端逐一执行 dws <cmd> --dry-run。
func TestRunAllEduCommands(t *testing.T) {
	type cmdCase struct {
		product string
		args    []string
	}

	allCmds := []cmdCase{
		// ═══════════════════════════════════════════════════════════
		// edu-contact: 29 commands
		// ═══════════════════════════════════════════════════════════
		// school (6)
		{"edu-contact", []string{"school", "roles"}},
		{"edu-contact", []string{"school", "structure"}},
		{"edu-contact", []string{"school", "periods"}},
		{"edu-contact", []string{"school", "type"}},
		{"edu-contact", []string{"school", "stats", "--statistics-type", "1"}},
		{"edu-contact", []string{"school", "class-list"}},
		// class (19)
		{"edu-contact", []string{"class", "detail", "--dept-id", "12345"}},
		{"edu-contact", []string{"class", "students", "--dept-id", "12345"}},
		{"edu-contact", []string{"class", "teachers", "--dept-id", "12345"}},
		{"edu-contact", []string{"class", "same-name", "--dept-id", "12345"}},
		{"edu-contact", []string{"class", "user-role", "--dept-id", "12345"}},
		{"edu-contact", []string{"class", "search-by-name", "--query-type", "student", "--name", "张三"}},
		{"edu-contact", []string{"class", "headmaster", "--class-name", "一年级1班"}},
		{"edu-contact", []string{"class", "search-by-teacher", "--name", "张老师"}},
		{"edu-contact", []string{"class", "update-student", "--class-id", "123", "--student-user-id", "u1"}},
		{"edu-contact", []string{"class", "add-student", "--dept-id", "123", "--student-name", "张三", "--student-mobile", "13800138000"}},
		{"edu-contact", []string{"class", "modify-student-info", "--dept-id", "123", "--target-user-id", "u1", "--nick", "小明"}},
		{"edu-contact", []string{"class", "delete-teacher", "--class-id", "123", "--teacher-user-id", "u1"}},
		{"edu-contact", []string{"class", "update-info", "--class-id", "123", "--nick", "火箭班"}},
		{"edu-contact", []string{"class", "update-student-number", "--class-id", "123", "--student-user-id", "u1", "--student-number", "S001"}},
		{"edu-contact", []string{"class", "add-unofficial-student", "--dept-id", "123", "--student-staff-ids", "s1,s2"}},
		{"edu-contact", []string{"class", "delete-students", "--dept-id", "123", "--student-user-ids", "u1,u2"}},
		{"edu-contact", []string{"class", "update-student-mobile", "--dept-id", "123", "--student-user-id", "u1", "--mobile", "13800138000"}},
		{"edu-contact", []string{"class", "move-student", "--student-user-ids", "u1,u2", "--origin-class-id", "123", "--target-class-id", "456"}},
		{"edu-contact", []string{"class", "add-teachers", "--dept-id", "123", "--teacher-user-ids", "u1,u2"}},
		// family (2)
		{"edu-contact", []string{"family", "children"}},
		{"edu-contact", []string{"family", "parents"}},
		// teacher (2)
		{"edu-contact", []string{"teacher", "classes"}},
		{"edu-contact", []string{"teacher", "update-course", "--teacher-class-infos", `[{"classId":123,"courseCode":"MATH","courseName":"数学"}]`}},

		// ═══════════════════════════════════════════════════════════
		// edu-group: 14 commands
		// ═══════════════════════════════════════════════════════════
		// student-group (7)
		{"edu-group", []string{"student-group", "info", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "exists", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "members", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "is-in", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "conversation", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "create", "--dept-id", "12345"}},
		{"edu-group", []string{"student-group", "disband", "--dept-id", "12345"}},
		// class-group (4)
		{"edu-group", []string{"class-group", "conversation-id", "--dept-id", "12345"}},
		{"edu-group", []string{"class-group", "conversation", "--dept-id", "12345"}},
		{"edu-group", []string{"class-group", "exists", "--dept-id", "12345"}},
		{"edu-group", []string{"class-group", "list-by-cids", "--conversation-ids", "cid1,cid2,cid3"}},
		// batch (3)
		{"edu-group", []string{"batch", "check-student-group", "--class-ids", "12345,67890"}},
		{"edu-group", []string{"batch", "get-class-groups", "--class-ids", "12345,67890"}},
		{"edu-group", []string{"batch", "create-student-groups"}},

		// ═══════════════════════════════════════════════════════════
		// edu-app: 42 commands
		// ═══════════════════════════════════════════════════════════
		// message (1)
		{"edu-app", []string{"message", "summary-list", "--class-id", "1", "--cid", "c1", "--target-role", "guardian", "--status", "0"}},
		// task (3)
		{"edu-app", []string{"task", "publish-list"}},
		{"edu-app", []string{"task", "all-list", "--biz-id", "1"}},
		{"edu-app", []string{"task", "student-list", "--students", `[{"userId":"u1","bizId":"1"}]`}},
		// report (5)
		{"edu-app", []string{"report", "get", "--ids", "1001,1002"}},
		{"edu-app", []string{"report", "by-teacher"}},
		{"edu-app", []string{"report", "by-class", "--report-id", "1001", "--class-id", "12345"}},
		{"edu-app", []string{"report", "by-student-list", "--class-id", "12345", "--student-id", "u1"}},
		{"edu-app", []string{"report", "by-student-detail", "--report-id", "1001", "--student-id", "u1", "--class-id", "12345"}},
		// notice (7)
		{"edu-app", []string{"notice", "confirm", "--notice-id", "n1", "--student-id", "u1"}},
		{"edu-app", []string{"notice", "create", "--identifer", "org1-staff1-uuid", "--content", "明天放假"}},
		{"edu-app", []string{"notice", "delete", "--notice-id", "12345"}},
		{"edu-app", []string{"notice", "list-by-teacher"}},
		{"edu-app", []string{"notice", "get", "--notice-id", "12345"}},
		{"edu-app", []string{"notice", "confirm-status", "--notice-id", "12345", "--class-id", "c1"}},
		{"edu-app", []string{"notice", "list-by-student", "--student-id", "u1", "--class-id", "c1"}},
		// circle (1)
		{"edu-app", []string{"circle", "posts", "--class-id", "12345", "--student-id", "u1", "--target-role", "guardian"}},
		// card (5)
		{"edu-app", []string{"card", "update", "--card-id", "1", "--identifier", "org1-staff1-uuid", "--title", "新标题"}},
		{"edu-app", []string{"card", "end", "--card-id", "1"}},
		{"edu-app", []string{"card", "list", "--status", "UNFINISH"}},
		{"edu-app", []string{"card", "user-statistic", "--card-id", "1", "--task-code", "code1", "--class-id", "cid1"}},
		{"edu-app", []string{"card", "finish-info", "--card-id", "1", "--card-biz-id", "bid1"}},
		// diploma (9)
		{"edu-app", []string{"diploma", "create", "--identifier", "org1-staff1-uuid", "--content", "三好学生", "--user-name", "张三"}},
		{"edu-app", []string{"diploma", "read", "--diploma-id", "1"}},
		{"edu-app", []string{"diploma", "list-by-teacher"}},
		{"edu-app", []string{"diploma", "get", "--diploma-id", "1"}},
		{"edu-app", []string{"diploma", "statistics", "--diploma-id", "1"}},
		{"edu-app", []string{"diploma", "detail", "--diploma-id", "1"}},
		{"edu-app", []string{"diploma", "list-by-student", "--student-id", "u1", "--class-id", "c1"}},
		{"edu-app", []string{"diploma", "student-detail", "--diploma-id", "1", "--student-id", "u1", "--class-id", "c1"}},
		{"edu-app", []string{"diploma", "delete", "--diploma-id", "1"}},
		// homework (11)
		{"edu-app", []string{"homework", "create", "--identifier", "org1-staff1-uuid", "--hw-content", "完成练习册第3页"}},
		{"edu-app", []string{"homework", "delete", "--homework-id", "1"}},
		{"edu-app", []string{"homework", "submit", "--hw-content-detail-id", "1"}},
		{"edu-app", []string{"homework", "get", "--homework-id", "1"}},
		{"edu-app", []string{"homework", "class-by-homework", "--homework-id", "1"}},
		{"edu-app", []string{"homework", "class-detail", "--homework-id", "1", "--class-id", "c1", "--user-name", "张老师"}},
		{"edu-app", []string{"homework", "submit-statistics", "--homework-id", "1", "--class-id", "c1"}},
		{"edu-app", []string{"homework", "list-by-student", "--student-id", "u1", "--class-id", "c1", "--user-name", "张三"}},
		{"edu-app", []string{"homework", "student-detail", "--homework-id", "1", "--student-id", "u1", "--class-id", "c1"}},
		{"edu-app", []string{"homework", "list-by-teacher"}},
		{"edu-app", []string{"homework", "create-comment", "--comment", "做得很好", "--hw-content-detail-id", "1"}},

		// ═══════════════════════════════════════════════════════════
		// edu-familygroup: 6 commands
		// ═══════════════════════════════════════════════════════════
		// group (2)
		{"edu-familygroup", []string{"group", "check-exists", "--uid", "12345", "--group-name", "小明一家"}},
		{"edu-familygroup", []string{"group", "list-children", "--uid", "12345"}},
		// manage (4)
		{"edu-familygroup", []string{"manage", "create", "--uid", "12345", "--children", `[{"name":"小明","students":[{"corpId":"corp1","staffId":"staff1"}]}]`}},
		{"edu-familygroup", []string{"manage", "invite-parent", "--org-id", "12345", "--uid", "67890", "--mobile", "13800138000"}},
		{"edu-familygroup", []string{"manage", "add-child", "--org-id", "12345", "--uid", "67890", "--name", "小红", "--mobile", "13900139000"}},
		{"edu-familygroup", []string{"manage", "toggle-app", "--org-id", "12345", "--uid", "67890", "--child-staff-id", "staff1", "--app-type", "XIAOTIANDI", "--open", "true"}},

		// ═══════════════════════════════════════════════════════════
		// college-contact: 65 commands
		// ═══════════════════════════════════════════════════════════
		// dept (9)
		{"college-contact", []string{"dept", "get-standard-structure"}},
		{"college-contact", []string{"dept", "get-detail", "--dept-id", "12345"}},
		{"college-contact", []string{"dept", "get-chain", "--dept-id", "12345"}},
		{"college-contact", []string{"dept", "search", "--dept-id", "12345", "--keyword", "计算机"}},
		{"college-contact", []string{"dept", "create", "--super-id", "1", "--stru-dept-id", "2", "--name", "计算机学院", "--dept-type", "COLLEGE", "--create-dept-group", "true"}},
		{"college-contact", []string{"dept", "update", "--dept-id", "12345", "--dept-type", "COLLEGE"}},
		{"college-contact", []string{"dept", "delete", "--dept-id", "12345"}},
		{"college-contact", []string{"dept", "batch-update-type", "--dept-ids", "1,2,3", "--target-dept-type", "COLLEGE"}},
		{"college-contact", []string{"dept", "overview"}},
		// employee (11)
		{"college-contact", []string{"employee", "get-detail", "--staff-id", "staff001"}},
		{"college-contact", []string{"employee", "add", "--emp-type", "TEACHER", "--main-dept-id", "12345", "--exclusive-account", "false"}},
		{"college-contact", []string{"employee", "remove", "--staff-ids", "s1,s2"}},
		{"college-contact", []string{"employee", "change-type", "--staff-id", "staff001", "--emp-type", "STUDENT"}},
		{"college-contact", []string{"employee", "change-dept", "--staff-id", "staff001", "--target-dept-id", "67890"}},
		{"college-contact", []string{"employee", "send-active-sms", "--dept-id", "12345"}},
		{"college-contact", []string{"employee", "list-employees", "--dept-id", "12345"}},
		{"college-contact", []string{"employee", "list-unaccepted", "--dept-id", "12345"}},
		{"college-contact", []string{"employee", "list-unactive", "--dept-id", "12345"}},
		{"college-contact", []string{"employee", "upgrade-status"}},
		{"college-contact", []string{"employee", "start-upgrade"}},
		// alumni (20)
		{"college-contact", []string{"alumni", "get-dept-tree", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "get-info", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "list", "--alumni-dept-id", "12345", "--order-field", "NAME", "--ordering", "ASC"}},
		{"college-contact", []string{"alumni", "query", "--staff-id", "staff001"}},
		{"college-contact", []string{"alumni", "search", "--keyword", "张三"}},
		{"college-contact", []string{"alumni", "list-unaccepted", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "get-group", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "create-dept", "--alumni-dept-id", "12345", "--dept-name", "2020届"}},
		{"college-contact", []string{"alumni", "update-dept", "--alumni-dept-id", "12345", "--dept-name", "2020届计算机"}},
		{"college-contact", []string{"alumni", "delete-dept", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "update-managers", "--alumni-dept-id", "12345", "--admin-user-ids", "u1,u2"}},
		{"college-contact", []string{"alumni", "add-alumnus", "--dept-ids", "1,2", "--name", "张三", "--mobile", "13800138000"}},
		{"college-contact", []string{"alumni", "update-alumnus", "--dept-ids", "1,2", "--staff-id", "s1", "--name", "张三"}},
		{"college-contact", []string{"alumni", "remove-alumnus", "--staff-id", "s1", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "cancel-invite", "--alumni-dept-id", "12345", "--staff-ids", "s1,s2"}},
		{"college-contact", []string{"alumni", "create-group", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "disband-group", "--alumni-dept-id", "12345"}},
		{"college-contact", []string{"alumni", "get-alumni-org-from-graduate"}},
		{"college-contact", []string{"alumni", "create-alumni-org", "--org-name", "计算机校友会"}},
		{"college-contact", []string{"alumni", "add-alumni-org-main-admins", "--admin-user-ids", "u1,u2"}},
		// graduate (16)
		{"college-contact", []string{"graduate", "query-graduate-years"}},
		{"college-contact", []string{"graduate", "query-graduate-depts", "--dept-id", "12345"}},
		{"college-contact", []string{"graduate", "query-graduate-sub-depts", "--dept-id", "12345"}},
		{"college-contact", []string{"graduate", "query-page-graduate-users", "--dept-id", "12345"}},
		{"college-contact", []string{"graduate", "get-task-result", "--request-no", "req001"}},
		{"college-contact", []string{"graduate", "get-alumni-org"}},
		{"college-contact", []string{"graduate", "query-restore-sub-depts", "--dept-id", "12345"}},
		{"college-contact", []string{"graduate", "query-dept-deleted-emps", "--dept-id", "12345"}},
		{"college-contact", []string{"graduate", "search-graduate", "--keyword", "张三"}},
		{"college-contact", []string{"graduate", "commit-graduate", "--graduate-dept-ids", "1,2", "--graduate-year", "2026"}},
		{"college-contact", []string{"graduate", "all-graduate", "--graduate-year", "2026"}},
		{"college-contact", []string{"graduate", "batch-graduate", "--dept-id", "12345", "--staff-ids", "s1,s2"}},
		{"college-contact", []string{"graduate", "delete-and-graduate", "--dept-id", "12345", "--staff-ids", "s1,s2"}},
		{"college-contact", []string{"graduate", "batch-delete-pending", "--dept-id", "12345", "--staff-ids", "s1,s2"}},
		{"college-contact", []string{"graduate", "batch-update-pending", "--dept-id", "12345", "--staff-ids", "s1,s2", "--graduate-year", "2026"}},
		{"college-contact", []string{"graduate", "commit-restore", "--graduate-dept-ids", "1,2"}},
		// group (9)
		{"college-contact", []string{"group", "query-group-rule"}},
		{"college-contact", []string{"group", "get-group-rule-schedule"}},
		{"college-contact", []string{"group", "query-preview-data"}},
		{"college-contact", []string{"group", "create-group-rule", "--name", "自动分组", "--tag-code", "TAG1", "--dept-type", "COLLEGE"}},
		{"college-contact", []string{"group", "delete-group-rule", "--rule-id", "1"}},
		{"college-contact", []string{"group", "enable-group-rule", "--rule-id", "1"}},
		{"college-contact", []string{"group", "disable-group-rule", "--rule-id", "1"}},
		{"college-contact", []string{"group", "set-group-rule-schedule"}},
		{"college-contact", []string{"group", "execute-group-rule"}},
	}

	// 按产品分组的命令构建器
	builders := map[string]func() *cobra.Command{
		"edu-contact":     newEduContactCommand,
		"edu-group":       newEduGroupCommand,
		"edu-app":         newEduAppCommand,
		"edu-familygroup": newEduFamilyGroupCommand,
		"college-contact": newCollegeContactCommand,
	}

	// 安装 dry-run caller
	caller := &recruitCaptureCaller{dryRun: true}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard

	passed := 0
	failed := 0
	currentProduct := ""

	for i, c := range allCmds {
		if c.product != currentProduct {
			currentProduct = c.product
			fmt.Printf("\n══════════════════════════════════════════════════════\n")
			fmt.Printf("  %s\n", strings.ToUpper(currentProduct))
			fmt.Printf("══════════════════════════════════════════════════════\n")
		}

		buildFn, ok := builders[c.product]
		if !ok {
			t.Fatalf("unknown product: %s", c.product)
		}
		root := buildFn()
		root.SetArgs(c.args)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)

		// Reset caller capture
		caller.productID = ""
		caller.tool = ""
		caller.args = nil

		err := root.Execute()
		cmdPath := fmt.Sprintf("dws %s %s", c.product, strings.Join(c.args, " "))

		if err != nil {
			failed++
			fmt.Printf("  [%3d] ✗ FAIL: %s\n", i+1, cmdPath)
			fmt.Printf("        Error: %v\n", err)
			t.Errorf("command %d failed: %s → %v", i+1, cmdPath, err)
		} else {
			passed++
			fmt.Printf("  [%3d] ✓ %s\n", i+1, cmdPath)
			if caller.tool != "" {
				fmt.Printf("        → MCP: %s.%s\n", caller.productID, caller.tool)
			}
		}
	}

	fmt.Printf("\n══════════════════════════════════════════════════════\n")
	fmt.Printf("  SUMMARY: %d passed, %d failed, %d total\n", passed, failed, passed+failed)
	fmt.Printf("══════════════════════════════════════════════════════\n")
}
