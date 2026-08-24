package helpers

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestEduContactRoot() *cobra.Command {
	return newEduContactCommand()
}

// ──────────────────────────────────────────────────────────
// 命令注册测试 — 验证所有子命令路径是否正确注册
// ──────────────────────────────────────────────────────────

func TestCrossPlatformCoverageEduContactCommandTree(t *testing.T) {
	root := newTestEduContactRoot()

	paths := [][]string{
		// school
		{"school", "roles"},
		{"school", "structure"},
		{"school", "periods"},
		{"school", "type"},
		{"school", "stats"},
		{"school", "class-list"},
		// class — 原有
		{"class", "detail"},
		{"class", "students"},
		{"class", "teachers"},
		{"class", "same-name"},
		{"class", "user-role"},
		{"class", "search-by-name"},
		{"class", "headmaster"},
		// class — 新增
		{"class", "search-by-teacher"},
		{"class", "add-student"},
		{"class", "add-teachers"},
		{"class", "add-unofficial-student"},
		{"class", "delete-students"},
		{"class", "delete-teacher"},
		{"class", "modify-student-info"},
		{"class", "move-student"},
		{"class", "update-info"},
		{"class", "update-student"},
		{"class", "update-student-mobile"},
		{"class", "update-student-number"},
		// family
		{"family", "children"},
		{"family", "parents"},
		// teacher
		{"teacher", "classes"},
		{"teacher", "update-course"},
	}

	for _, path := range paths {
		if _, _, err := root.Find(path); err != nil {
			t.Errorf("command path %v not found: %v", path, err)
		}
	}
}

// ──────────────────────────────────────────────────────────
// 参数校验测试 — 验证必填参数缺失时返回错误
// ──────────────────────────────────────────────────────────

func executeCommand(root *cobra.Command, args ...string) error {
	root.SetArgs(args)
	return root.Execute()
}

func TestCrossPlatformCoverageClassSearchByTeacher_MissingName(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "search-by-teacher")
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
}

func TestCrossPlatformCoverageClassAddTeachers_MissingDeptId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-teachers", "--teacher-user-ids", "uid1")
	if err == nil {
		t.Fatal("expected error for missing --dept-id, got nil")
	}
}

func TestCrossPlatformCoverageClassAddTeachers_MissingTeacherUserIds(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-teachers", "--dept-id", "12345")
	if err == nil {
		t.Fatal("expected error for missing --teacher-user-ids, got nil")
	}
}

func TestCrossPlatformCoverageClassAddTeachers_InvalidIsAdviser(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-teachers", "--dept-id", "12345", "--teacher-user-ids", "uid1", "--is-adviser", "3")
	if err == nil {
		t.Fatal("expected error for invalid --is-adviser, got nil")
	}
}

func TestCrossPlatformCoverageClassMoveStudent_MissingOriginClassId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "move-student", "--student-user-ids", "uid1", "--target-class-id", "67890")
	if err == nil {
		t.Fatal("expected error for missing --origin-class-id, got nil")
	}
}

func TestCrossPlatformCoverageClassMoveStudent_MissingStudentUserIds(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "move-student", "--origin-class-id", "12345", "--target-class-id", "67890")
	if err == nil {
		t.Fatal("expected error for missing --student-user-ids, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateStudentMobile_MissingMobile(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-student-mobile", "--dept-id", "12345", "--student-user-id", "uid1")
	if err == nil {
		t.Fatal("expected error for missing --mobile, got nil")
	}
}

func TestCrossPlatformCoverageClassDeleteStudents_MissingDeptId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "delete-students", "--student-user-ids", "uid1")
	if err == nil {
		t.Fatal("expected error for missing --dept-id, got nil")
	}
}

func TestCrossPlatformCoverageClassDeleteStudents_MissingStudentUserIds(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "delete-students", "--dept-id", "12345")
	if err == nil {
		t.Fatal("expected error for missing --student-user-ids, got nil")
	}
}

func TestCrossPlatformCoverageClassAddUnofficialStudent_MissingDeptId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-unofficial-student", "--student-staff-ids", "sid1")
	if err == nil {
		t.Fatal("expected error for missing --dept-id, got nil")
	}
}

func TestCrossPlatformCoverageClassAddUnofficialStudent_MissingStaffIds(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-unofficial-student", "--dept-id", "12345")
	if err == nil {
		t.Fatal("expected error for missing --student-staff-ids, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateStudent_MissingClassId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-student", "--student-user-id", "uid1", "--student-name", "张三", "--append-patriarch")
	if err == nil {
		t.Fatal("expected error for missing --class-id, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateStudent_MissingStudentUserId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-student", "--class-id", "12345", "--student-name", "张三", "--append-patriarch")
	if err == nil {
		t.Fatal("expected error for missing --student-user-id, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateStudent_InvalidPatriarchsJSON(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-student", "--class-id", "12345", "--student-user-id", "uid1", "--patriarchs", "invalid-json", "--append-patriarch")
	if err == nil {
		t.Fatal("expected error for invalid --patriarchs JSON, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateStudentNumber_MissingStudentNumber(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-student-number", "--class-id", "12345", "--student-user-id", "uid1")
	if err == nil {
		t.Fatal("expected error for missing --student-number, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateInfo_MissingClassId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-info", "--nick", "火箭班")
	if err == nil {
		t.Fatal("expected error for missing --class-id, got nil")
	}
}

func TestCrossPlatformCoverageClassUpdateInfo_GroupNameWithoutConversationId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "update-info", "--class-id", "12345", "--group-name", "测试群")
	if err == nil {
		t.Fatal("expected error for --group-name without --conversation-id, got nil")
	}
}

func TestCrossPlatformCoverageClassDeleteTeacher_MissingTeacherUserId(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "delete-teacher", "--class-id", "12345")
	if err == nil {
		t.Fatal("expected error for missing --teacher-user-id, got nil")
	}
}

func TestCrossPlatformCoverageClassModifyStudentInfo_MissingBothNickAndPatriarch(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "modify-student-info", "--dept-id", "12345", "--target-user-id", "uid1")
	if err == nil {
		t.Fatal("expected error for missing both --nick and --patriarch-user-id, got nil")
	}
}

func TestCrossPlatformCoverageClassModifyStudentInfo_PatriarchWithoutRelation(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "modify-student-info", "--dept-id", "12345", "--target-user-id", "uid1", "--patriarch-user-id", "pid1")
	if err == nil {
		t.Fatal("expected error for --patriarch-user-id without --relation, got nil")
	}
}

func TestCrossPlatformCoverageClassAddStudent_MissingStudentName(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-student", "--dept-id", "12345", "--student-mobile", "13800138000")
	if err == nil {
		t.Fatal("expected error for missing --student-name, got nil")
	}
}

func TestCrossPlatformCoverageClassAddStudent_MissingMobile(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-student", "--dept-id", "12345", "--student-name", "张三")
	if err == nil {
		t.Fatal("expected error for missing mobile (student or parent), got nil")
	}
}

func TestCrossPlatformCoverageClassAddStudent_InvalidMotherJSON(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "class", "add-student", "--dept-id", "12345", "--student-name", "张三", "--mother", "bad-json")
	if err == nil {
		t.Fatal("expected error for invalid --mother JSON, got nil")
	}
}

func TestCrossPlatformCoverageTeacherUpdateCourse_MissingTeacherClassInfos(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "teacher", "update-course")
	if err == nil {
		t.Fatal("expected error for missing --teacher-class-infos, got nil")
	}
}

func TestCrossPlatformCoverageTeacherUpdateCourse_InvalidJSON(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "teacher", "update-course", "--teacher-class-infos", "not-json")
	if err == nil {
		t.Fatal("expected error for invalid --teacher-class-infos JSON, got nil")
	}
}

func TestCrossPlatformCoverageTeacherUpdateCourse_EmptyArray(t *testing.T) {
	root := newTestEduContactRoot()
	err := executeCommand(root, "teacher", "update-course", "--teacher-class-infos", "[]")
	if err == nil {
		t.Fatal("expected error for empty --teacher-class-infos array, got nil")
	}
}

// ──────────────────────────────────────────────────────────
// Happy-path 测试 — 每个 leaf 命令的成功分支（经由 dry-run caller）
// 以及所有可选字段分支，用于把 changed-code 覆盖率补到 100%。
// ──────────────────────────────────────────────────────────

// withEduContactCaller installs a dry-run capture caller so happy-path command
// execution exercises each RunE up to the callMCPToolOnServer dispatch without
// requiring a live MCP transport.
func withEduContactCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{dryRun: true}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func TestCrossPlatformCoverageEduContactHappyPaths(t *testing.T) {
	withEduContactCaller(t)
	cases := [][]string{
		// school
		{"school", "roles"},
		{"school", "structure"},
		{"school", "periods"},
		{"school", "type"},
		{"school", "stats"},
		{"school", "stats", "--statistics-type", "1"},
		{"school", "class-list"},
		// class — 读操作
		{"class", "detail", "--dept-id", "123"},
		{"class", "students", "--dept-id", "123"},
		{"class", "teachers", "--dept-id", "123"},
		{"class", "same-name", "--dept-id", "123"},
		{"class", "user-role", "--dept-id", "123"},
		{"class", "search-by-name", "--query-type", "student", "--name", "张三"},
		{"class", "headmaster", "--class-name", "一年级1班"},
		{"class", "search-by-teacher", "--name", "张老师"},
		// class — update-student 各分支
		{"class", "update-student", "--class-id", "123", "--student-user-id", "u1",
			"--student-name", "张三", "--student-number", "S1", "--append-patriarch",
			"--patriarchs", `[{"userId":"uid1","relation":"F"}]`},
		{"class", "update-student", "--class-id", "123", "--student-user-id", "u1"},
		// class — add-student 各手机号来源分支
		{"class", "add-student", "--dept-id", "123", "--student-name", "张三",
			"--student-mobile", "13800138000", "--student-user-id", "u1",
			"--student-number", "S1", "--virtual-account-id", "v1"},
		{"class", "add-student", "--dept-id", "123", "--student-name", "张三",
			"--mother", `{"mobile":"13800138000","relation":"M"}`},
		{"class", "add-student", "--dept-id", "123", "--student-name", "张三",
			"--father", `{"mobile":"13900139000","relation":"F"}`},
		{"class", "add-student", "--dept-id", "123", "--student-name", "张三",
			"--other-patriarchs", `[{"mobile":"13700137000","relation":"O"}]`},
		// class — modify-student-info 两个分支
		{"class", "modify-student-info", "--dept-id", "123", "--target-user-id", "u1", "--nick", "张三"},
		{"class", "modify-student-info", "--dept-id", "123", "--target-user-id", "u1",
			"--patriarch-user-id", "p1", "--relation", "父亲"},
		// class — delete-teacher
		{"class", "delete-teacher", "--class-id", "123", "--teacher-user-id", "u1"},
		// class — update-info 三个可选分支
		{"class", "update-info", "--class-id", "123", "--nick", "火箭班"},
		{"class", "update-info", "--class-id", "123", "--expected-student-num", "45"},
		{"class", "update-info", "--class-id", "123", "--group-name", "家长群", "--conversation-id", "cid1"},
		// class — update-student-number
		{"class", "update-student-number", "--class-id", "123", "--student-user-id", "u1", "--student-number", "S1"},
		// class — add-unofficial-student
		{"class", "add-unofficial-student", "--dept-id", "123", "--student-staff-ids", "s1,s2"},
		// class — delete-students
		{"class", "delete-students", "--dept-id", "123", "--student-user-ids", "u1,u2"},
		// class — update-student-mobile
		{"class", "update-student-mobile", "--dept-id", "123", "--student-user-id", "u1", "--mobile", "13800138000"},
		// class — move-student
		{"class", "move-student", "--student-user-ids", "u1,u2", "--origin-class-id", "123", "--target-class-id", "456"},
		// class — add-teachers 默认/班主任
		{"class", "add-teachers", "--dept-id", "123", "--teacher-user-ids", "u1,u2"},
		{"class", "add-teachers", "--dept-id", "123", "--teacher-user-ids", "u1", "--is-adviser", "1"},
		// family
		{"family", "children"},
		{"family", "parents"},
		// teacher
		{"teacher", "classes"},
		{"teacher", "update-course", "--teacher-class-infos", `[{"classId":123,"courseCode":"c1","courseName":"语文"}]`},
	}
	for _, args := range cases {
		root := newTestEduContactRoot()
		if err := executeCommand(root, args...); err != nil {
			t.Errorf("happy path %v returned error: %v", args, err)
		}
	}
}

// TestEduContactErrorPathsRemaining covers validation branches not exercised by
// the existing error tests, ensuring 100% changed-code coverage.
func TestCrossPlatformCoverageEduContactErrorPathsRemaining(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		// eduRequiredIntFlag：空值 + 非整数
		{"stats-invalid-type", []string{"school", "stats", "--statistics-type", "abc"}},
		{"detail-missing-dept", []string{"class", "detail"}},
		{"detail-invalid-dept", []string{"class", "detail", "--dept-id", "abc"}},
		{"students-missing-dept", []string{"class", "students"}},
		{"teachers-missing-dept", []string{"class", "teachers"}},
		{"samename-missing-dept", []string{"class", "same-name"}},
		{"userrole-missing-dept", []string{"class", "user-role"}},
		{"searchbyname-missing-querytype", []string{"class", "search-by-name", "--name", "张三"}},
		{"searchbyname-missing-name", []string{"class", "search-by-name", "--query-type", "student"}},
		{"headmaster-missing-classname", []string{"class", "headmaster"}},
		// update-student
		{"updatestudent-missing-classid", []string{"class", "update-student", "--student-user-id", "u1"}},
		{"updatestudent-missing-userid", []string{"class", "update-student", "--class-id", "123"}},
		{"updatestudent-invalid-patriarchs", []string{"class", "update-student", "--class-id", "123", "--student-user-id", "u1", "--patriarchs", "not-json"}},
		// add-student
		{"addstudent-missing-dept", []string{"class", "add-student", "--student-name", "张三"}},
		{"addstudent-missing-name", []string{"class", "add-student", "--dept-id", "123"}},
		{"addstudent-invalid-father", []string{"class", "add-student", "--dept-id", "123", "--student-name", "张三", "--father", "not-json"}},
		{"addstudent-invalid-other", []string{"class", "add-student", "--dept-id", "123", "--student-name", "张三", "--other-patriarchs", "not-json"}},
		{"addstudent-no-mobile", []string{"class", "add-student", "--dept-id", "123", "--student-name", "张三"}},
		// modify-student-info
		{"modify-missing-dept", []string{"class", "modify-student-info", "--target-user-id", "u1", "--nick", "张三"}},
		{"modify-missing-target", []string{"class", "modify-student-info", "--dept-id", "123", "--nick", "张三"}},
		// delete-teacher
		{"deleteteacher-missing-classid", []string{"class", "delete-teacher", "--teacher-user-id", "u1"}},
		// update-info
		{"updateinfo-missing-classid", []string{"class", "update-info", "--nick", "火箭班"}},
		{"updateinfo-invalid-expected", []string{"class", "update-info", "--class-id", "123", "--expected-student-num", "abc"}},
		// update-student-number
		{"usn-missing-classid", []string{"class", "update-student-number", "--student-user-id", "u1", "--student-number", "S1"}},
		{"usn-missing-userid", []string{"class", "update-student-number", "--class-id", "123", "--student-number", "S1"}},
		// add-unofficial-student
		{"unofficial-missing-dept", []string{"class", "add-unofficial-student", "--student-staff-ids", "s1"}},
		{"unofficial-empty-staffids", []string{"class", "add-unofficial-student", "--dept-id", "123", "--student-staff-ids", ",,"}},
		// delete-students
		{"deletestudents-missing-userids", []string{"class", "delete-students", "--dept-id", "123"}},
		{"deletestudents-empty-userids", []string{"class", "delete-students", "--dept-id", "123", "--student-user-ids", ",,"}},
		// update-student-mobile
		{"usm-missing-dept", []string{"class", "update-student-mobile", "--student-user-id", "u1", "--mobile", "13800138000"}},
		{"usm-missing-userid", []string{"class", "update-student-mobile", "--dept-id", "123", "--mobile", "13800138000"}},
		// move-student
		{"move-missing-target", []string{"class", "move-student", "--student-user-ids", "u1", "--origin-class-id", "123"}},
		{"move-empty-userids", []string{"class", "move-student", "--origin-class-id", "123", "--target-class-id", "456", "--student-user-ids", ",,"}},
		// add-teachers
		{"addteachers-empty-userids", []string{"class", "add-teachers", "--dept-id", "123", "--teacher-user-ids", ",,"}},
		{"addteachers-too-many", []string{"class", "add-teachers", "--dept-id", "123", "--teacher-user-ids", strings.Repeat("u,", 51) + "u"}},
	}
	for _, tc := range cases {
		root := newTestEduContactRoot()
		if err := executeCommand(root, tc.args...); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

// ──────────────────────────────────────────────────────────
// Dispatch 验证测试 — 验证命令正确派发到 MCP Server
// ──────────────────────────────────────────────────────────

// withEduContactDispatchCaller installs a non-dry-run capture caller so that
// callMCPToolOnServer goes through deps.Caller.CallTool and we can verify
// the dispatched productID, tool name, and args.
func withEduContactDispatchCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func TestCrossPlatformCoverageEduContactDispatch(t *testing.T) {
	t.Run("class detail dispatches get_class_detail with deptId", func(t *testing.T) {
		caller := withEduContactDispatchCaller(t)
		root := newEduContactCommand()
		root.SetArgs([]string{"class", "detail", "--dept-id", "123"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if caller.productID != "edu-contact" {
			t.Errorf("productID = %q, want %q", caller.productID, "edu-contact")
		}
		if caller.tool != "get_class_detail" {
			t.Errorf("tool = %q, want %q", caller.tool, "get_class_detail")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] is not map[string]any: %#v", caller.args)
		}
		if input["deptId"] != int64(123) {
			t.Errorf("input[deptId] = %v (%T), want int64(123)", input["deptId"], input["deptId"])
		}
	})

	t.Run("class search-by-name dispatches query_class_by_guardian_name", func(t *testing.T) {
		caller := withEduContactDispatchCaller(t)
		root := newEduContactCommand()
		root.SetArgs([]string{"class", "search-by-name", "--query-type", "student", "--name", "张三"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if caller.productID != "edu-contact" {
			t.Errorf("productID = %q, want %q", caller.productID, "edu-contact")
		}
		if caller.tool != "query_class_by_guardian_name" {
			t.Errorf("tool = %q, want %q", caller.tool, "query_class_by_guardian_name")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] is not map[string]any: %#v", caller.args)
		}
		if input["queryType"] != "student" {
			t.Errorf("input[queryType] = %v, want %q", input["queryType"], "student")
		}
		if input["name"] != "张三" {
			t.Errorf("input[name] = %v, want %q", input["name"], "张三")
		}
	})

	t.Run("school stats dispatches statistics_school with statisticsType", func(t *testing.T) {
		caller := withEduContactDispatchCaller(t)
		root := newEduContactCommand()
		root.SetArgs([]string{"school", "stats", "--statistics-type", "2"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if caller.productID != "edu-contact" {
			t.Errorf("productID = %q, want %q", caller.productID, "edu-contact")
		}
		if caller.tool != "statistics_school" {
			t.Errorf("tool = %q, want %q", caller.tool, "statistics_school")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] is not map[string]any: %#v", caller.args)
		}
		if input["statisticsType"] != int64(2) {
			t.Errorf("input[statisticsType] = %v (%T), want int64(2)", input["statisticsType"], input["statisticsType"])
		}
	})

	t.Run("class add-teachers dispatches batch_add_class_teacher with list and isAdviser", func(t *testing.T) {
		caller := withEduContactDispatchCaller(t)
		root := newEduContactCommand()
		root.SetArgs([]string{"class", "add-teachers", "--dept-id", "789", "--teacher-user-ids", "t1,t2", "--is-adviser", "1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if caller.productID != "edu-contact" {
			t.Errorf("productID = %q, want %q", caller.productID, "edu-contact")
		}
		if caller.tool != "batch_add_class_teacher" {
			t.Errorf("tool = %q, want %q", caller.tool, "batch_add_class_teacher")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] is not map[string]any: %#v", caller.args)
		}
		if input["deptId"] != int64(789) {
			t.Errorf("input[deptId] = %v (%T), want int64(789)", input["deptId"], input["deptId"])
		}
		teacherUserIds, ok := input["teacherUserIds"].([]string)
		if !ok {
			t.Fatalf("input[teacherUserIds] is not []string: %#v", input["teacherUserIds"])
		}
		if len(teacherUserIds) != 2 || teacherUserIds[0] != "t1" || teacherUserIds[1] != "t2" {
			t.Errorf("input[teacherUserIds] = %v, want [t1 t2]", teacherUserIds)
		}
		if input["isAdviser"] != int64(1) {
			t.Errorf("input[isAdviser] = %v (%T), want int64(1)", input["isAdviser"], input["isAdviser"])
		}
	})

	t.Run("class move-student dispatches move_student with list and two int flags", func(t *testing.T) {
		caller := withEduContactDispatchCaller(t)
		root := newEduContactCommand()
		root.SetArgs([]string{"class", "move-student", "--student-user-ids", "u1,u2", "--origin-class-id", "100", "--target-class-id", "200"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if caller.productID != "edu-contact" {
			t.Errorf("productID = %q, want %q", caller.productID, "edu-contact")
		}
		if caller.tool != "move_student" {
			t.Errorf("tool = %q, want %q", caller.tool, "move_student")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] is not map[string]any: %#v", caller.args)
		}
		studentUserIds, ok := input["studentUserIds"].([]string)
		if !ok {
			t.Fatalf("input[studentUserIds] is not []string: %#v", input["studentUserIds"])
		}
		if len(studentUserIds) != 2 || studentUserIds[0] != "u1" || studentUserIds[1] != "u2" {
			t.Errorf("input[studentUserIds] = %v, want [u1 u2]", studentUserIds)
		}
		if input["originClassId"] != int64(100) {
			t.Errorf("input[originClassId] = %v (%T), want int64(100)", input["originClassId"], input["originClassId"])
		}
		if input["targetClassId"] != int64(200) {
			t.Errorf("input[targetClassId] = %v (%T), want int64(200)", input["targetClassId"], input["targetClassId"])
		}
	})
}

// newEduContactConfirmRoot 模拟真实运行时的根命令：核心框架在 rootCmd 上注册
// 全局 persistent --yes flag，叶子命令通过合并后的 Flags() 读取。
func newEduContactConfirmRoot() *cobra.Command {
	root := &cobra.Command{Use: "dws"}
	root.PersistentFlags().BoolP("yes", "y", false, "跳过确认提示")
	root.AddCommand(newEduContactCommand())
	return root
}

// TestCrossPlatformCoverageEduContactDestructiveConfirmGate 对 edu-contact 每个
// user_required 破坏性叶子做成对验证：
//   - 未显式确认：返回 confirmation_required 错误，且 caller 调用次数为零。
//   - 显式确认后：恰好一次 MCP 调用，且 productID、tool、完整参数均准确。
func TestCrossPlatformCoverageEduContactDestructiveConfirmGate(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantTool  string
		wantInput map[string]any
	}{
		{
			"class delete-teacher",
			[]string{"edu-contact", "class", "delete-teacher", "--class-id", "12345", "--teacher-user-id", "userId1"},
			"delete_teacher",
			map[string]any{"classId": int64(12345), "teacherUserId": "userId1"},
		},
		{
			"class delete-students",
			[]string{"edu-contact", "class", "delete-students", "--dept-id", "12345", "--student-user-ids", "userId1,userId2"},
			"delete_students",
			map[string]any{"deptId": int64(12345), "studentUserIds": []string{"userId1", "userId2"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/rejected_without_yes", func(t *testing.T) {
			caller := &recruitCaptureCaller{dryRun: false}
			InitDepsForTest(t, caller)
			deps.Out.w = io.Discard

			root := newEduContactConfirmRoot()
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

			root := newEduContactConfirmRoot()
			root.SetArgs(append(append([]string{}, tc.args...), "--yes"))
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() with --yes error = %v", err)
			}
			if len(caller.calls) != 1 {
				t.Fatalf("expected exactly 1 MCP call with --yes, got %d", len(caller.calls))
			}
			if caller.calls[0].productID != "edu-contact" {
				t.Errorf("productID = %q, want %q", caller.calls[0].productID, "edu-contact")
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
