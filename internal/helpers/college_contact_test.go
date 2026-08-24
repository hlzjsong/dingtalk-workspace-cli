package helpers

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCrossPlatformCoverageCollegeContactCommand_Structure(t *testing.T) {
	cmd := newCollegeContactCommand()

	if cmd.Name() != "college-contact" {
		t.Errorf("expected name 'college-contact', got %q", cmd.Name())
	}
	if !cmd.Hidden {
		t.Error("extension root command should be Hidden")
	}

	// 分组 → 叶子命令映射
	groups := map[string][]string{
		"dept": {
			"get-standard-structure", "get-detail", "get-chain", "search",
			"create", "update", "delete", "batch-update-type", "overview",
		},
		"employee": {
			"get-detail", "add", "remove", "change-type", "change-dept",
			"send-active-sms", "list-employees", "list-unaccepted",
			"list-unactive", "upgrade-status", "start-upgrade",
		},
		"alumni": {
			"get-dept-tree", "get-info", "list", "query", "search", "list-unaccepted", "get-group", "create-dept", "update-dept", "delete-dept", "update-managers", "add-alumnus", "update-alumnus", "remove-alumnus", "cancel-invite", "create-group", "disband-group", "get-alumni-org-from-graduate", "create-alumni-org", "add-alumni-org-main-admins",
		},
		"graduate": {
			"query-graduate-years", "query-graduate-depts", "query-graduate-sub-depts", "query-page-graduate-users", "get-task-result", "get-alumni-org", "query-restore-sub-depts", "query-dept-deleted-emps", "search-graduate", "commit-graduate", "all-graduate", "batch-graduate", "delete-and-graduate", "batch-delete-pending", "batch-update-pending", "commit-restore",
		},
		"group": {
			"query-group-rule", "get-group-rule-schedule", "query-preview-data", "create-group-rule", "delete-group-rule", "enable-group-rule", "disable-group-rule", "set-group-rule-schedule", "execute-group-rule",
		},
	}

	for groupName, leaves := range groups {
		var groupCmd *cobra.Command
		for _, c := range cmd.Commands() {
			if c.Name() == groupName {
				groupCmd = c
				break
			}
		}
		if groupCmd == nil {
			t.Fatalf("subcommand group %q not found", groupName)
		}
		for _, leaf := range leaves {
			found := false
			for _, c := range groupCmd.Commands() {
				if c.Name() == leaf {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("leaf command %q not found under %q", leaf, groupName)
			}
		}
	}

	// stats 分组已移除
	for _, c := range cmd.Commands() {
		if c.Name() == "stats" {
			t.Error("subcommand group 'stats' should be removed")
		}
	}
}

func TestCrossPlatformCoverageCollegeContactCommand_FindPath(t *testing.T) {
	root := &cobra.Command{Use: "dws"}
	root.AddCommand(newCollegeContactCommand())

	c, _, err := root.Find([]string{"college-contact", "dept", "get-standard-structure"})
	if err != nil {
		t.Fatalf("command path not found: %v", err)
	}
	if c.Name() != "get-standard-structure" {
		t.Errorf("expected leaf 'get-standard-structure', got %q", c.Name())
	}
}

// newCollegeContactTestRoot 模拟真实运行时的根命令：核心框架在 rootCmd 上
// 注册全局 persistent --yes flag，叶子命令通过合并后的 Flags() 读取。
func newCollegeContactTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "dws"}
	root.PersistentFlags().BoolP("yes", "y", false, "跳过确认提示")
	root.AddCommand(newCollegeContactCommand())
	return root
}

// runDestructiveLeaf 执行不可逆叶子命令并捕获 panic。
// 单测环境未初始化 products 运行时依赖，若门禁放行后进入
// CallMCPToolOnServer 会因 deps 为 nil 而 panic，据此区分
// “被门禁拦截（返回错误）”与“已越过门禁到达 MCP 调用层（panic）”。
func runDestructiveLeaf(t *testing.T, args ...string) (err error, panicked bool) {
	t.Helper()
	root := newCollegeContactTestRoot()
	root.SetArgs(args)
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	err = root.Execute()
	return err, false
}

func TestCrossPlatformCoverageCollegeContactDestructive_RejectedWithoutYes(t *testing.T) {
	cases := [][]string{
		{"college-contact", "dept", "delete", "--dept-id", "12345"},
		{"college-contact", "employee", "remove", "--staff-ids", "S12345,S12346"},
	}
	for _, args := range cases {
		err, panicked := runDestructiveLeaf(t, args...)
		if panicked {
			t.Fatalf("%v: 未传 --yes 不应到达 MCP 调用层", args)
		}
		if err == nil {
			t.Fatalf("%v: 未传 --yes 应拒绝执行", args)
		}
		if !strings.Contains(err.Error(), "--yes") {
			t.Errorf("%v: 错误信息应提示 --yes，got: %v", args, err)
		}
	}
}

func TestCrossPlatformCoverageCollegeContactDestructive_ProceedsWithYes(t *testing.T) {
	cases := [][]string{
		{"college-contact", "dept", "delete", "--dept-id", "12345", "--yes"},
		{"college-contact", "employee", "remove", "--staff-ids", "S12345", "--yes"},
	}
	for _, args := range cases {
		err, panicked := runDestructiveLeaf(t, args...)
		if !panicked {
			// 未 panic 意味着未到达 MCP 调用层；若返回的仍是门禁错误则为拦截失败
			if err != nil && strings.Contains(err.Error(), "需要用户确认") {
				t.Fatalf("%v: 已传 --yes 仍被门禁拦截: %v", args, err)
			}
		}
	}
}

// withCollegeContactCaller installs a dry-run capture caller so happy-path
// command execution exercises each RunE up to the callMCPToolOnServer dispatch
// without requiring a live MCP transport. In dry-run mode destructive
// commands' confirm gate short-circuits to nil, so no --yes flag is needed.
func withCollegeContactCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{dryRun: true}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

// TestCollegeContactHappyPaths runs every leaf command with required flags only
// and with all optional flags populated, expecting a nil error (dry-run preview).
func TestCrossPlatformCoverageCollegeContactHappyPaths(t *testing.T) {
	withCollegeContactCaller(t)

	cases := [][]string{
		// ── dept ─────────────────────────────────────────────
		{"dept", "get-standard-structure"},
		{"dept", "get-standard-structure", "--dept-id", "123", "--staff-id", "S1", "--keyword", "k", "--offset", "0", "--size", "20"},
		{"dept", "get-detail", "--dept-id", "123"},
		{"dept", "get-detail", "--dept-id", "123", "--staff-id", "S1", "--keyword", "k", "--offset", "0", "--size", "20"},
		{"dept", "get-chain", "--dept-id", "123"},
		{"dept", "get-chain", "--dept-id", "123", "--staff-id", "S1", "--keyword", "k", "--offset", "0", "--size", "20"},
		{"dept", "search", "--dept-id", "123", "--keyword", "k"},
		{"dept", "search", "--dept-id", "123", "--keyword", "k", "--staff-id", "S1", "--offset", "0", "--size", "20"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college", "--create-dept-group", "true"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college", "--create-dept-group", "false", "--dept-id", "5", "--dept-code", "C", "--brief", "b", "--phone", "p"},
		{"dept", "update", "--dept-id", "123", "--dept-type", "college"},
		{"dept", "update", "--dept-id", "123", "--dept-type", "college", "--stru-dept-id", "200", "--super-id", "100", "--create-dept-group", "true", "--name", "X", "--dept-code", "C", "--brief", "b", "--phone", "p"},
		{"dept", "delete", "--dept-id", "123"},
		{"dept", "batch-update-type", "--dept-ids", "100,200", "--target-dept-type", "college"},
		{"dept", "overview"},
		{"dept", "overview", "--dept-id", "123", "--staff-id", "S1", "--keyword", "k", "--offset", "0", "--size", "20"},

		// ── employee ─────────────────────────────────────────
		{"employee", "get-detail", "--staff-id", "S1"},
		{"employee", "get-detail", "--staff-id", "S1", "--dept-id", "1", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--name", "n", "--mobile", "m", "--job-number", "j", "--emp-type", "college_student", "--login-id-type", "l", "--order-field", "job_number", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "add", "--emp-type", "college_student", "--main-dept-id", "100", "--exclusive-account", "true"},
		{"employee", "add", "--emp-type", "college_student", "--main-dept-id", "100", "--exclusive-account", "true", "--dept-id", "1", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--send-active-sms", "false", "--staff-id", "S1", "--name", "n", "--mobile", "m", "--job-number", "j", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "remove", "--staff-ids", "S1"},
		{"employee", "remove", "--staff-ids", "S1", "--dept-id", "1", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--staff-id", "x", "--name", "n", "--mobile", "m", "--job-number", "j", "--emp-type", "college_student", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc"},
		{"employee", "change-type", "--staff-id", "S1", "--emp-type", "college_teacher"},
		{"employee", "change-type", "--staff-id", "S1", "--emp-type", "college_teacher", "--dept-id", "1", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--name", "n", "--mobile", "m", "--job-number", "j", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "change-dept", "--staff-id", "S1", "--target-dept-id", "200"},
		{"employee", "change-dept", "--staff-id", "S1", "--target-dept-id", "200", "--dept-id", "1", "--main-dept-id", "2", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--name", "n", "--mobile", "m", "--job-number", "j", "--emp-type", "college_student", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "send-active-sms", "--dept-id", "100"},
		{"employee", "send-active-sms", "--dept-id", "100", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--staff-id", "x", "--name", "n", "--mobile", "m", "--job-number", "j", "--emp-type", "college_student", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "list-employees", "--dept-id", "123"},
		{"employee", "list-employees", "--dept-id", "123", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--staff-id", "x", "--name", "n", "--mobile", "m", "--job-number", "j", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "list-unaccepted", "--dept-id", "123"},
		{"employee", "list-unaccepted", "--dept-id", "123", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--staff-id", "x", "--name", "n", "--mobile", "m", "--job-number", "j", "--emp-type", "college_student", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "list-unactive", "--dept-id", "123"},
		{"employee", "list-unactive", "--dept-id", "123", "--main-dept-id", "2", "--target-dept-id", "3", "--offset", "0", "--size", "20", "--exclusive-account", "true", "--send-active-sms", "true", "--staff-id", "x", "--name", "n", "--mobile", "m", "--job-number", "j", "--login-id-type", "l", "--order-field", "f", "--ordering", "asc", "--staff-ids", "s1,s2"},
		{"employee", "upgrade-status"},
		{"employee", "upgrade-status", "--dept-id", "123", "--staff-id", "S1", "--keyword", "k", "--offset", "0", "--size", "20"},
		{"employee", "start-upgrade"},

		// ── alumni ───────────────────────────────────────────
		{"alumni", "get-dept-tree", "--alumni-dept-id", "123"},
		{"alumni", "get-info", "--alumni-dept-id", "123"},
		{"alumni", "list", "--alumni-dept-id", "1", "--order-field", "dept_entry", "--ordering", "asc"},
		{"alumni", "list", "--alumni-dept-id", "1", "--order-field", "dept_entry", "--ordering", "asc", "--offset", "0", "--size", "20"},
		{"alumni", "query", "--staff-id", "S1"},
		{"alumni", "search", "--keyword", "x"},
		{"alumni", "search", "--keyword", "x", "--offset", "0", "--size", "20"},
		{"alumni", "list-unaccepted", "--alumni-dept-id", "1"},
		{"alumni", "list-unaccepted", "--alumni-dept-id", "1", "--offset", "0", "--size", "20"},
		{"alumni", "get-group", "--alumni-dept-id", "1"},
		{"alumni", "create-dept", "--alumni-dept-id", "1", "--dept-name", "D"},
		{"alumni", "update-dept", "--alumni-dept-id", "1", "--dept-name", "D"},
		{"alumni", "delete-dept", "--alumni-dept-id", "1"},
		{"alumni", "update-managers", "--alumni-dept-id", "1", "--admin-user-ids", "u1,u2"},
		{"alumni", "add-alumnus", "--name", "X", "--mobile", "138", "--dept-ids", "1,2"},
		{"alumni", "add-alumnus", "--name", "X", "--mobile", "138", "--dept-ids", "1,2", "--student-number", "2020", "--email", "e", "--intake", "2020", "--outtake", "2024"},
		{"alumni", "update-alumnus", "--staff-id", "S1", "--name", "X", "--dept-ids", "1,2"},
		{"alumni", "update-alumnus", "--staff-id", "S1", "--name", "X", "--dept-ids", "1,2", "--student-number", "2020", "--email", "e", "--intake", "2020", "--outtake", "2024"},
		{"alumni", "remove-alumnus", "--staff-id", "S1", "--alumni-dept-id", "1"},
		{"alumni", "cancel-invite", "--alumni-dept-id", "1", "--staff-ids", "s1,s2"},
		{"alumni", "create-group", "--alumni-dept-id", "1"},
		{"alumni", "disband-group", "--alumni-dept-id", "1"},
		{"alumni", "get-alumni-org-from-graduate"},
		{"alumni", "create-alumni-org", "--org-name", "O"},
		{"alumni", "add-alumni-org-main-admins", "--admin-user-ids", "u1,u2"},

		// ── graduate ─────────────────────────────────────────
		{"graduate", "query-graduate-years"},
		{"graduate", "query-graduate-depts", "--dept-id", "1"},
		{"graduate", "query-graduate-depts", "--dept-id", "1", "--graduate-year", "2026"},
		{"graduate", "query-graduate-sub-depts", "--dept-id", "1"},
		{"graduate", "query-page-graduate-users", "--dept-id", "1"},
		{"graduate", "query-page-graduate-users", "--dept-id", "1", "--graduate-year", "2026", "--offset", "0", "--size", "20"},
		{"graduate", "get-task-result", "--request-no", "r1"},
		{"graduate", "get-task-result", "--request-no", "r1", "--type", "GRADUATE"},
		{"graduate", "get-alumni-org"},
		{"graduate", "query-restore-sub-depts", "--dept-id", "1"},
		{"graduate", "query-dept-deleted-emps", "--dept-id", "1"},
		{"graduate", "query-dept-deleted-emps", "--dept-id", "1", "--offset", "0", "--size", "20"},
		{"graduate", "search-graduate", "--keyword", "x"},
		{"graduate", "search-graduate", "--keyword", "x", "--offset", "0", "--size", "20"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", "1,2", "--graduate-year", "2026"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", "1,2", "--graduate-year", "2026", "--request-no", "r1"},
		{"graduate", "all-graduate", "--graduate-year", "2026"},
		{"graduate", "all-graduate", "--graduate-year", "2026", "--request-no", "r1"},
		{"graduate", "batch-graduate", "--dept-id", "1", "--staff-ids", "s1,s2"},
		{"graduate", "delete-and-graduate", "--dept-id", "1", "--staff-ids", "s1,s2"},
		{"graduate", "batch-delete-pending", "--dept-id", "1", "--staff-ids", "s1,s2"},
		{"graduate", "batch-update-pending", "--dept-id", "1", "--staff-ids", "s1,s2", "--graduate-year", "2026"},
		{"graduate", "commit-restore", "--graduate-dept-ids", "1,2"},
		{"graduate", "commit-restore", "--graduate-dept-ids", "1,2", "--request-no", "r1"},

		// ── group ────────────────────────────────────────────
		{"group", "query-group-rule"},
		{"group", "query-group-rule", "--name", "N", "--offset", "0", "--size", "20"},
		{"group", "get-group-rule-schedule"},
		{"group", "query-preview-data"},
		{"group", "query-preview-data", "--offset", "0", "--size", "20"},
		{"group", "create-group-rule", "--name", "X", "--tag-code", "T", "--dept-type", "college"},
		{"group", "create-group-rule", "--name", "X", "--tag-code", "T", "--dept-type", "college", "--auto-admin", "true"},
		{"group", "delete-group-rule", "--rule-id", "1"},
		{"group", "enable-group-rule", "--rule-id", "1"},
		{"group", "disable-group-rule", "--rule-id", "1"},
		{"group", "set-group-rule-schedule"},
		{"group", "set-group-rule-schedule", "--cron", "0 0 2 * * ?"},
		{"group", "execute-group-rule"},
	}

	for _, args := range cases {
		root := newCollegeContactCommand()
		if err := executeCommand(root, args...); err != nil {
			t.Errorf("%v: expected nil error, got: %v", args, err)
		}
	}
}

// TestCollegeContactValidationErrors exercises every validation-error branch:
// missing required flags, non-integer int flags, invalid bool flags, and
// empty-after-split CSV lists. Each case must return a non-nil error.
func TestCrossPlatformCoverageCollegeContactValidationErrors(t *testing.T) {
	withCollegeContactCaller(t)

	cases := [][]string{
		// ── dept ─────────────────────────────────────────────
		{"dept", "get-standard-structure", "--dept-id", "abc"},
		{"dept", "get-standard-structure", "--offset", "abc"},
		{"dept", "get-standard-structure", "--size", "abc"},
		{"dept", "get-detail"},
		{"dept", "get-detail", "--dept-id", "abc"},
		{"dept", "get-detail", "--dept-id", "1", "--offset", "abc"},
		{"dept", "get-detail", "--dept-id", "1", "--size", "abc"},
		{"dept", "get-chain"},
		{"dept", "get-chain", "--dept-id", "abc"},
		{"dept", "get-chain", "--dept-id", "1", "--offset", "abc"},
		{"dept", "get-chain", "--dept-id", "1", "--size", "abc"},
		{"dept", "search"},
		{"dept", "search", "--dept-id", "abc", "--keyword", "k"},
		{"dept", "search", "--dept-id", "1"},
		{"dept", "search", "--dept-id", "1", "--keyword", "k", "--offset", "abc"},
		{"dept", "search", "--dept-id", "1", "--keyword", "k", "--size", "abc"},
		{"dept", "create"},
		{"dept", "create", "--super-id", "abc"},
		{"dept", "create", "--super-id", "100"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "abc"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college", "--create-dept-group", "maybe"},
		{"dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college", "--create-dept-group", "true", "--dept-id", "abc"},
		{"dept", "update"},
		{"dept", "update", "--dept-id", "abc"},
		{"dept", "update", "--dept-id", "1"},
		{"dept", "update", "--dept-id", "1", "--dept-type", "college", "--stru-dept-id", "abc"},
		{"dept", "update", "--dept-id", "1", "--dept-type", "college", "--super-id", "abc"},
		{"dept", "update", "--dept-id", "1", "--dept-type", "college", "--create-dept-group", "maybe"},
		{"dept", "delete"},
		{"dept", "delete", "--dept-id", "abc"},
		{"dept", "batch-update-type"},
		{"dept", "batch-update-type", "--dept-ids", "abc", "--target-dept-type", "college"},
		{"dept", "batch-update-type", "--dept-ids", ",,", "--target-dept-type", "college"},
		{"dept", "batch-update-type", "--dept-ids", "1,2"},
		{"dept", "overview", "--dept-id", "abc"},
		{"dept", "overview", "--offset", "abc"},
		{"dept", "overview", "--size", "abc"},

		// ── employee ─────────────────────────────────────────
		{"employee", "get-detail"},
		{"employee", "get-detail", "--staff-id", "S1", "--dept-id", "abc"},
		{"employee", "get-detail", "--staff-id", "S1", "--exclusive-account", "maybe"},
		{"employee", "add"},
		{"employee", "add", "--emp-type", "x"},
		{"employee", "add", "--emp-type", "x", "--main-dept-id", "abc"},
		{"employee", "add", "--emp-type", "x", "--main-dept-id", "100"},
		{"employee", "add", "--emp-type", "x", "--main-dept-id", "100", "--exclusive-account", "maybe"},
		{"employee", "add", "--emp-type", "x", "--main-dept-id", "100", "--exclusive-account", "true", "--dept-id", "abc"},
		{"employee", "add", "--emp-type", "x", "--main-dept-id", "100", "--exclusive-account", "true", "--send-active-sms", "maybe"},
		{"employee", "remove"},
		{"employee", "remove", "--staff-ids", ",,"},
		{"employee", "remove", "--staff-ids", "S1", "--dept-id", "abc"},
		{"employee", "remove", "--staff-ids", "S1", "--exclusive-account", "maybe"},
		{"employee", "change-type"},
		{"employee", "change-type", "--staff-id", "S1"},
		{"employee", "change-type", "--staff-id", "S1", "--emp-type", "t", "--dept-id", "abc"},
		{"employee", "change-type", "--staff-id", "S1", "--emp-type", "t", "--exclusive-account", "maybe"},
		{"employee", "change-dept"},
		{"employee", "change-dept", "--staff-id", "S1"},
		{"employee", "change-dept", "--staff-id", "S1", "--target-dept-id", "abc"},
		{"employee", "change-dept", "--staff-id", "S1", "--target-dept-id", "200", "--dept-id", "abc"},
		{"employee", "change-dept", "--staff-id", "S1", "--target-dept-id", "200", "--exclusive-account", "maybe"},
		{"employee", "send-active-sms"},
		{"employee", "send-active-sms", "--dept-id", "abc"},
		{"employee", "send-active-sms", "--dept-id", "1", "--main-dept-id", "abc"},
		{"employee", "send-active-sms", "--dept-id", "1", "--exclusive-account", "maybe"},
		{"employee", "list-employees"},
		{"employee", "list-employees", "--dept-id", "abc"},
		{"employee", "list-employees", "--dept-id", "1", "--main-dept-id", "abc"},
		{"employee", "list-employees", "--dept-id", "1", "--exclusive-account", "maybe"},
		{"employee", "list-unaccepted"},
		{"employee", "list-unaccepted", "--dept-id", "abc"},
		{"employee", "list-unaccepted", "--dept-id", "1", "--main-dept-id", "abc"},
		{"employee", "list-unaccepted", "--dept-id", "1", "--exclusive-account", "maybe"},
		{"employee", "list-unactive"},
		{"employee", "list-unactive", "--dept-id", "abc"},
		{"employee", "list-unactive", "--dept-id", "1", "--main-dept-id", "abc"},
		{"employee", "list-unactive", "--dept-id", "1", "--exclusive-account", "maybe"},
		{"employee", "upgrade-status", "--dept-id", "abc"},
		{"employee", "upgrade-status", "--offset", "abc"},
		{"employee", "upgrade-status", "--size", "abc"},

		// ── alumni ───────────────────────────────────────────
		{"alumni", "get-dept-tree"},
		{"alumni", "get-dept-tree", "--alumni-dept-id", "abc"},
		{"alumni", "get-info"},
		{"alumni", "get-info", "--alumni-dept-id", "abc"},
		{"alumni", "list"},
		{"alumni", "list", "--alumni-dept-id", "abc", "--order-field", "f", "--ordering", "asc"},
		{"alumni", "list", "--alumni-dept-id", "1"},
		{"alumni", "list", "--alumni-dept-id", "1", "--order-field", "f"},
		{"alumni", "list", "--alumni-dept-id", "1", "--order-field", "f", "--ordering", "asc", "--offset", "abc"},
		{"alumni", "query"},
		{"alumni", "search"},
		{"alumni", "search", "--keyword", "x", "--offset", "abc"},
		{"alumni", "list-unaccepted"},
		{"alumni", "list-unaccepted", "--alumni-dept-id", "abc"},
		{"alumni", "list-unaccepted", "--alumni-dept-id", "1", "--offset", "abc"},
		{"alumni", "get-group"},
		{"alumni", "get-group", "--alumni-dept-id", "abc"},
		{"alumni", "create-dept"},
		{"alumni", "create-dept", "--alumni-dept-id", "abc", "--dept-name", "D"},
		{"alumni", "create-dept", "--alumni-dept-id", "1"},
		{"alumni", "update-dept"},
		{"alumni", "update-dept", "--alumni-dept-id", "abc", "--dept-name", "D"},
		{"alumni", "update-dept", "--alumni-dept-id", "1"},
		{"alumni", "delete-dept"},
		{"alumni", "delete-dept", "--alumni-dept-id", "abc"},
		{"alumni", "update-managers"},
		{"alumni", "update-managers", "--alumni-dept-id", "abc", "--admin-user-ids", "u"},
		{"alumni", "update-managers", "--alumni-dept-id", "1"},
		{"alumni", "update-managers", "--alumni-dept-id", "1", "--admin-user-ids", ",,"},
		{"alumni", "add-alumnus"},
		{"alumni", "add-alumnus", "--name", "X"},
		{"alumni", "add-alumnus", "--name", "X", "--mobile", "m"},
		{"alumni", "add-alumnus", "--name", "X", "--mobile", "m", "--dept-ids", "abc"},
		{"alumni", "add-alumnus", "--name", "X", "--mobile", "m", "--dept-ids", ",,"},
		{"alumni", "update-alumnus"},
		{"alumni", "update-alumnus", "--staff-id", "S1"},
		{"alumni", "update-alumnus", "--staff-id", "S1", "--name", "X"},
		{"alumni", "update-alumnus", "--staff-id", "S1", "--name", "X", "--dept-ids", "abc"},
		{"alumni", "update-alumnus", "--staff-id", "S1", "--name", "X", "--dept-ids", ",,"},
		{"alumni", "remove-alumnus"},
		{"alumni", "remove-alumnus", "--staff-id", "S1"},
		{"alumni", "remove-alumnus", "--staff-id", "S1", "--alumni-dept-id", "abc"},
		{"alumni", "cancel-invite"},
		{"alumni", "cancel-invite", "--alumni-dept-id", "abc", "--staff-ids", "s"},
		{"alumni", "cancel-invite", "--alumni-dept-id", "1"},
		{"alumni", "cancel-invite", "--alumni-dept-id", "1", "--staff-ids", ",,"},
		{"alumni", "create-group"},
		{"alumni", "create-group", "--alumni-dept-id", "abc"},
		{"alumni", "disband-group"},
		{"alumni", "disband-group", "--alumni-dept-id", "abc"},
		{"alumni", "create-alumni-org"},
		{"alumni", "add-alumni-org-main-admins"},
		{"alumni", "add-alumni-org-main-admins", "--admin-user-ids", ",,"},

		// ── graduate ─────────────────────────────────────────
		{"graduate", "query-graduate-depts"},
		{"graduate", "query-graduate-depts", "--dept-id", "abc"},
		{"graduate", "query-graduate-depts", "--dept-id", "1", "--graduate-year", "abc"},
		{"graduate", "query-graduate-sub-depts"},
		{"graduate", "query-graduate-sub-depts", "--dept-id", "abc"},
		{"graduate", "query-page-graduate-users"},
		{"graduate", "query-page-graduate-users", "--dept-id", "abc"},
		{"graduate", "query-page-graduate-users", "--dept-id", "1", "--offset", "abc"},
		{"graduate", "get-task-result"},
		{"graduate", "query-restore-sub-depts"},
		{"graduate", "query-restore-sub-depts", "--dept-id", "abc"},
		{"graduate", "query-dept-deleted-emps"},
		{"graduate", "query-dept-deleted-emps", "--dept-id", "abc"},
		{"graduate", "query-dept-deleted-emps", "--dept-id", "1", "--offset", "abc"},
		{"graduate", "search-graduate"},
		{"graduate", "search-graduate", "--keyword", "x", "--offset", "abc"},
		{"graduate", "commit-graduate"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", "abc"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", ",,"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", "1,2"},
		{"graduate", "commit-graduate", "--graduate-dept-ids", "1,2", "--graduate-year", "abc"},
		{"graduate", "all-graduate"},
		{"graduate", "all-graduate", "--graduate-year", "abc"},
		{"graduate", "batch-graduate"},
		{"graduate", "batch-graduate", "--dept-id", "abc"},
		{"graduate", "batch-graduate", "--dept-id", "1"},
		{"graduate", "batch-graduate", "--dept-id", "1", "--staff-ids", ",,"},
		{"graduate", "delete-and-graduate"},
		{"graduate", "delete-and-graduate", "--dept-id", "abc"},
		{"graduate", "delete-and-graduate", "--dept-id", "1"},
		{"graduate", "delete-and-graduate", "--dept-id", "1", "--staff-ids", ",,"},
		{"graduate", "batch-delete-pending"},
		{"graduate", "batch-delete-pending", "--dept-id", "abc"},
		{"graduate", "batch-delete-pending", "--dept-id", "1"},
		{"graduate", "batch-delete-pending", "--dept-id", "1", "--staff-ids", ",,"},
		{"graduate", "batch-update-pending"},
		{"graduate", "batch-update-pending", "--dept-id", "abc"},
		{"graduate", "batch-update-pending", "--dept-id", "1"},
		{"graduate", "batch-update-pending", "--dept-id", "1", "--staff-ids", ",,"},
		{"graduate", "batch-update-pending", "--dept-id", "1", "--staff-ids", "s1"},
		{"graduate", "batch-update-pending", "--dept-id", "1", "--staff-ids", "s1", "--graduate-year", "abc"},
		{"graduate", "commit-restore"},
		{"graduate", "commit-restore", "--graduate-dept-ids", "abc"},
		{"graduate", "commit-restore", "--graduate-dept-ids", ",,"},

		// ── group ────────────────────────────────────────────
		{"group", "query-group-rule", "--offset", "abc"},
		{"group", "query-group-rule", "--size", "abc"},
		{"group", "query-preview-data", "--offset", "abc"},
		{"group", "query-preview-data", "--size", "abc"},
		{"group", "create-group-rule"},
		{"group", "create-group-rule", "--name", "X"},
		{"group", "create-group-rule", "--name", "X", "--tag-code", "T"},
		{"group", "create-group-rule", "--name", "X", "--tag-code", "T", "--dept-type", "college", "--auto-admin", "maybe"},
		{"group", "delete-group-rule"},
		{"group", "delete-group-rule", "--rule-id", "abc"},
		{"group", "enable-group-rule"},
		{"group", "enable-group-rule", "--rule-id", "abc"},
		{"group", "disable-group-rule"},
		{"group", "disable-group-rule", "--rule-id", "abc"},
	}

	for _, args := range cases {
		root := newCollegeContactCommand()
		if err := executeCommand(root, args...); err == nil {
			t.Errorf("%v: expected non-nil error, got nil", args)
		}
	}
}

// TestCrossPlatformCoverageCollegeContactDestructiveConfirmGate verifies every
// user_required destructive leaf in a paired manner:
//   - Without --yes: returns confirmation_required error AND caller is never invoked (zero calls).
//   - With --yes: proceeds to MCP dispatch with exactly one call AND the correct
//     productID, tool name, and complete argument payload.
func TestCrossPlatformCoverageCollegeContactDestructiveConfirmGate(t *testing.T) {
	type destructiveCase struct {
		name      string
		args      []string
		wantTool  string
		wantInput map[string]any
	}

	cases := []destructiveCase{
		{
			"dept delete",
			[]string{"college-contact", "dept", "delete", "--dept-id", "123"},
			"delete_college_contact_dept",
			map[string]any{"deptId": int64(123)},
		},
		{
			"employee remove",
			[]string{"college-contact", "employee", "remove", "--staff-ids", "S1,S2"},
			"remove_employee",
			map[string]any{"staffIds": []string{"S1", "S2"}},
		},
		{
			"alumni delete-dept",
			[]string{"college-contact", "alumni", "delete-dept", "--alumni-dept-id", "1"},
			"delete_alumni_dept",
			map[string]any{"alumniDeptId": int64(1)},
		},
		{
			"alumni remove-alumnus",
			[]string{"college-contact", "alumni", "remove-alumnus", "--staff-id", "S1", "--alumni-dept-id", "1"},
			"delete_alumnus",
			map[string]any{"staffId": "S1", "alumniDeptId": int64(1)},
		},
		{
			"alumni cancel-invite",
			[]string{"college-contact", "alumni", "cancel-invite", "--alumni-dept-id", "1", "--staff-ids", "s1,s2"},
			"delete_alumni_invite_record",
			map[string]any{"alumniDeptId": int64(1), "staffIds": []string{"s1", "s2"}},
		},
		{
			"alumni disband-group",
			[]string{"college-contact", "alumni", "disband-group", "--alumni-dept-id", "1"},
			"disband_alumni_group",
			map[string]any{"alumniDeptId": int64(1)},
		},
		{
			"graduate commit-graduate",
			[]string{"college-contact", "graduate", "commit-graduate", "--graduate-dept-ids", "1,2", "--graduate-year", "2026"},
			"commit_graduate",
			map[string]any{"graduateDeptIds": []int64{1, 2}, "graduateYear": int64(2026)},
		},
		{
			"graduate all-graduate",
			[]string{"college-contact", "graduate", "all-graduate", "--graduate-year", "2026"},
			"all_graduate",
			map[string]any{"graduateYear": int64(2026)},
		},
		{
			"graduate batch-graduate",
			[]string{"college-contact", "graduate", "batch-graduate", "--dept-id", "1", "--staff-ids", "s1,s2"},
			"batch_graduate",
			map[string]any{"deptId": int64(1), "staffIds": []string{"s1", "s2"}},
		},
		{
			"graduate delete-and-graduate",
			[]string{"college-contact", "graduate", "delete-and-graduate", "--dept-id", "1", "--staff-ids", "s1,s2"},
			"delete_and_graduate",
			map[string]any{"deptId": int64(1), "staffIds": []string{"s1", "s2"}},
		},
		{
			"graduate batch-delete-pending",
			[]string{"college-contact", "graduate", "batch-delete-pending", "--dept-id", "1", "--staff-ids", "s1,s2"},
			"batch_delete_pending",
			map[string]any{"deptId": int64(1), "staffIds": []string{"s1", "s2"}},
		},
		{
			"graduate batch-update-pending",
			[]string{"college-contact", "graduate", "batch-update-pending", "--dept-id", "1", "--staff-ids", "s1,s2", "--graduate-year", "2026"},
			"batch_update_pending",
			map[string]any{"deptId": int64(1), "staffIds": []string{"s1", "s2"}, "graduateYear": int64(2026)},
		},
		{
			"graduate commit-restore",
			[]string{"college-contact", "graduate", "commit-restore", "--graduate-dept-ids", "1,2"},
			"commit_restore",
			map[string]any{"graduateDeptIds": []int64{1, 2}},
		},
		{
			"group delete-group-rule",
			[]string{"college-contact", "group", "delete-group-rule", "--rule-id", "1"},
			"delete_group_rule",
			map[string]any{"ruleId": int64(1)},
		},
		{
			"group execute-group-rule",
			[]string{"college-contact", "group", "execute-group-rule"},
			"execute_group_rule",
			map[string]any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/rejected_without_yes", func(t *testing.T) {
			caller := &recruitCaptureCaller{dryRun: false}
			InitDepsForTest(t, caller)
			deps.Out.w = io.Discard

			root := newCollegeContactTestRoot()
			root.SetArgs(tc.args)
			err := root.Execute()
			if err == nil {
				t.Fatalf("expected confirm-gate error without --yes, got nil")
			}
			if !strings.Contains(err.Error(), "需要用户确认") {
				t.Fatalf("expected confirmation gate error, got: %v", err)
			}
			if len(caller.calls) != 0 {
				t.Fatalf("caller should not be invoked without --yes, got %d calls", len(caller.calls))
			}
		})

		t.Run(tc.name+"/dispatched_with_yes", func(t *testing.T) {
			caller := &recruitCaptureCaller{dryRun: false}
			InitDepsForTest(t, caller)
			deps.Out.w = io.Discard

			root := newCollegeContactTestRoot()
			argsWithYes := append(append([]string{}, tc.args...), "--yes")
			root.SetArgs(argsWithYes)
			err := root.Execute()
			if err != nil {
				t.Fatalf("Execute() with --yes error = %v", err)
			}
			if len(caller.calls) != 1 {
				t.Fatalf("expected exactly 1 MCP call with --yes, got %d", len(caller.calls))
			}
			if caller.calls[0].productID != "college-contact" {
				t.Errorf("productID = %q, want %q", caller.calls[0].productID, "college-contact")
			}
			if caller.calls[0].tool != tc.wantTool {
				t.Errorf("tool = %q, want %q", caller.calls[0].tool, tc.wantTool)
			}
			gotArgs := caller.calls[0].args
			if len(gotArgs) != 1 {
				t.Fatalf("args should carry exactly the \"input\" key, got %v", gotArgs)
			}
			gotInput, ok := gotArgs["input"].(map[string]any)
			if !ok {
				t.Fatalf("args[\"input\"] should be map[string]any, got %T", gotArgs["input"])
			}
			if !reflect.DeepEqual(gotInput, tc.wantInput) {
				t.Errorf("input = %#v, want %#v", gotInput, tc.wantInput)
			}
		})
	}
}

// withCollegeContactDispatchCaller installs a non-dry-run capture caller so
// commands go through the full dispatch path (deps.Caller.CallTool) and we can
// verify the productID, tool name, and args passed to callMCPToolOnServer.
func withCollegeContactDispatchCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

// TestCollegeContactDispatch verifies that representative commands from each
// group dispatch to the correct MCP tool with the expected productID and args.
func TestCrossPlatformCoverageCollegeContactDispatch(t *testing.T) {
	type dispatchCase struct {
		name      string
		args      []string
		wantTool  string
		wantProd  string
		checkArgs func(t *testing.T, args map[string]any)
	}

	cases := []dispatchCase{
		{
			name:     "dept get-standard-structure",
			args:     []string{"college-contact", "dept", "get-standard-structure"},
			wantTool: "get_college_standard_structure",
			wantProd: "college-contact",
		},
		{
			name:     "dept get-detail",
			args:     []string{"college-contact", "dept", "get-detail", "--dept-id", "123"},
			wantTool: "get_college_dept_detail",
			wantProd: "college-contact",
			checkArgs: func(t *testing.T, args map[string]any) {
				input := args["input"].(map[string]any)
				if input["deptId"] != int64(123) {
					t.Errorf("deptId = %v (%T), want int64(123)", input["deptId"], input["deptId"])
				}
			},
		},
		{
			name:     "dept create",
			args:     []string{"college-contact", "dept", "create", "--super-id", "100", "--stru-dept-id", "200", "--name", "X", "--dept-type", "college", "--create-dept-group", "true"},
			wantTool: "create_college_contact_dept",
			wantProd: "college-contact",
		},
		{
			name:     "employee get-detail",
			args:     []string{"college-contact", "employee", "get-detail", "--staff-id", "S1"},
			wantTool: "get_employee_detail",
			wantProd: "college-contact",
			checkArgs: func(t *testing.T, args map[string]any) {
				input := args["input"].(map[string]any)
				if input["staffId"] != "S1" {
					t.Errorf("staffId = %v, want S1", input["staffId"])
				}
			},
		},
		{
			name:     "alumni get-dept-tree",
			args:     []string{"college-contact", "alumni", "get-dept-tree", "--alumni-dept-id", "123"},
			wantTool: "get_alumni_dept_tree",
			wantProd: "college-contact",
			checkArgs: func(t *testing.T, args map[string]any) {
				input := args["input"].(map[string]any)
				if input["alumniDeptId"] != int64(123) {
					t.Errorf("alumniDeptId = %v (%T), want int64(123)", input["alumniDeptId"], input["alumniDeptId"])
				}
			},
		},
		{
			name:     "graduate query-graduate-years",
			args:     []string{"college-contact", "graduate", "query-graduate-years"},
			wantTool: "query_graduate_years",
			wantProd: "college-contact",
		},
		{
			name:     "group query-group-rule",
			args:     []string{"college-contact", "group", "query-group-rule"},
			wantTool: "query_group_rule",
			wantProd: "college-contact",
		},
		{
			name:     "dept delete with --yes",
			args:     []string{"college-contact", "dept", "delete", "--dept-id", "123", "--yes"},
			wantTool: "delete_college_contact_dept",
			wantProd: "college-contact",
			checkArgs: func(t *testing.T, args map[string]any) {
				input := args["input"].(map[string]any)
				if input["deptId"] != int64(123) {
					t.Errorf("deptId = %v (%T), want int64(123)", input["deptId"], input["deptId"])
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caller := withCollegeContactDispatchCaller(t)

			root := &cobra.Command{Use: "dws"}
			root.PersistentFlags().BoolP("yes", "y", false, "跳过确认提示")
			root.AddCommand(newCollegeContactCommand())
			root.SetArgs(tc.args)

			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if caller.productID != tc.wantProd {
				t.Errorf("productID = %q, want %q", caller.productID, tc.wantProd)
			}
			if caller.tool != tc.wantTool {
				t.Errorf("tool = %q, want %q", caller.tool, tc.wantTool)
			}
			if tc.checkArgs != nil {
				tc.checkArgs(t, caller.args)
			}
		})
	}
}
