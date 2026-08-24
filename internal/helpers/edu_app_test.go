// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");

package helpers

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func withEduAppCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{dryRun: true}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func runEduApp(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newEduAppCommand()
	cmd.SetArgs(args)
	return cmd.Execute()
}

// TestEduAppHappyPathsFullFlags exercises each leaf command with every flag
// populated, so all optional-field branches and the dispatch line are covered.
func TestCrossPlatformCoverageEduAppHappyPathsFullFlags(t *testing.T) {
	cases := [][]string{
		{"message", "summary-list", "--class-id", "1", "--cid", "c", "--target-role", "guardian", "--status", "0"},

		{"task", "publish-list", "--cursor", "5", "--limit", "10", "--need-statistic", "--task-sources", "EDU_HOMEWORK,EDU_NOTICE"},
		{"task", "all-list", "--biz-id", "1", "--cursor", "5", "--limit", "10", "--need-statistic", "--task-sources", "EDU_CARD"},
		{"task", "student-list", "--students", `[{"userId":"u1","bizId":"1"}]`, "--query-all", "--cursor", "c", "--limit", "10", "--task-sources", "EDU_SR"},

		{"report", "get", "--ids", "1001,1002"},
		{"report", "by-teacher", "--page", "1", "--limit", "20", "--status", "1"},
		{"report", "by-class", "--report-id", "1001", "--class-id", "12345", "--student-ids", "u1,u2"},
		{"report", "by-student-list", "--class-id", "12345", "--student-id", "u1", "--page", "1", "--limit", "20"},
		{"report", "by-student-detail", "--report-id", "1001", "--student-id", "u1", "--class-id", "12345"},

		{"notice", "confirm", "--notice-id", "n1", "--student-id", "u1", "--device-id", "d1", "--parent-name", "张三", "--update-sign"},
		{"notice", "create", "--identifer", "org1-staff1-uuid", "--content", "明天放假", "--title", "放假",
			"--class-ids", "1,2", "--class-names", "一班,二班", "--class-selected-students", `{"1":["u1"]}`,
			"--type", "SCHOOL", "--scope", "ALL", "--target-role", "guardian", "--is-signed", "true",
			"--photo", "p", "--media", "m", "--audio", "a", "--send-ding", "--scheduled-release", "2026-07-29",
			"--notice-deadline", "100", "--notice-deadline-open", "true", "--notice-deadline-setting", "s",
			"--attributes", `{"k":"v"}`, "--user-name", "张三"},
		{"notice", "delete", "--notice-id", "12345", "--user-name", "张三"},
		{"notice", "list-by-teacher", "--class-id", "c", "--type", "SCHOOL", "--status", "FINISHED", "--user-name", "u", "--page", "1", "--page-size", "20"},
		{"notice", "get", "--notice-id", "12345", "--user-name", "张三"},
		{"notice", "confirm-status", "--notice-id", "12345", "--class-id", "c", "--status", "CONFIRMED", "--user-name", "u", "--page", "1", "--page-size", "20"},
		{"notice", "list-by-student", "--student-id", "u1", "--class-id", "c", "--status", "FINISHED", "--user-name", "u", "--page", "1", "--page-size", "20"},

		{"circle", "posts", "--class-id", "12345", "--student-id", "u1", "--target-role", "guardian"},

		{"card", "update", "--card-id", "1", "--identifier", "org1-staff1-uuid", "--title", "新标题", "--content", "新内容", "--should-send-update-msg"},
		{"card", "end", "--card-id", "1"},
		{"card", "list", "--status", "UNFINISH", "--class-id", "5", "--page", "1", "--limit", "10"},
		{"card", "user-statistic", "--card-id", "1", "--task-code", "code1", "--class-id", "cid1", "--finish", "--page", "1", "--limit", "20"},
		{"card", "finish-info", "--card-id", "1", "--card-biz-id", "bid1", "--target-role", "guardian", "--student-id", "stu1"},

		{"diploma", "create", "--identifier", "org1-staff1-uuid", "--content", "期末三好学生", "--user-name", "张三",
			"--title", "三好学生", "--unit-name", "实验小学", "--tag", "三好", "--photo", "p", "--publish-time", "2026-07-29",
			"--biz-code", "bc", "--biz-category", "cat", "--msg-type", "mt", "--template-url", "tpl",
			"--class-ids", "1,2", "--select-class", `[{"classId":"1"}]`, "--attributes", `{"k":"v"}`},
		{"diploma", "read", "--diploma-id", "1", "--class-id", "c", "--student-id", "u1", "--user-name", "张三"},
		{"diploma", "list-by-teacher", "--page", "1", "--limit", "20", "--status", "PUBLISHED", "--tag", "三好", "--user-name", "u"},
		{"diploma", "get", "--diploma-id", "1", "--user-name", "张三"},
		{"diploma", "statistics", "--diploma-id", "1", "--user-name", "张三"},
		{"diploma", "detail", "--diploma-id", "1", "--class-id", "c", "--user-name", "张三"},
		{"diploma", "list-by-student", "--student-id", "u1", "--class-id", "c", "--page", "1", "--limit", "20", "--user-name", "张三"},
		{"diploma", "student-detail", "--diploma-id", "1", "--student-id", "u1", "--class-id", "c", "--user-name", "张三"},
		{"diploma", "delete", "--diploma-id", "1", "--user-name", "张三"},

		{"homework", "create", "--identifier", "org1-staff1-uuid", "--hw-content", "完成练习",
			"--hw-title", "数学作业", "--hw-photo", "p", "--hw-media", "m", "--hw-video", "v",
			"--class-ids", "1,2", "--class-names", "一班,二班", "--class-selected-students", `{"1":["u1"]}`,
			"--feedback", "fb", "--hw-deadline", "100", "--hw-deadline-open", "true", "--hw-deadline-setting", "s",
			"--submit-types", "TEXT,PHOTO", "--hw-type", "HOMEWORK", "--target-role", "guardian", "--publish-type", "NOW",
			"--biz-code", "bc", "--scheduled-release", "2026-07-29", "--task-plan-duration", "5", "--attributes", `{"k":"v"}`, "--user-name", "张三"},
		{"homework", "delete", "--homework-id", "1", "--user-name", "张三"},
		{"homework", "submit", "--hw-content-detail-id", "1", "--homework-id", "2", "--student-id", "u1", "--class-id", "c",
			"--content", "已完成", "--photo", "p", "--media", "m", "--video", "v", "--user-name", "张三"},
		{"homework", "get", "--homework-id", "1", "--user-name", "张三"},
		{"homework", "class-by-homework", "--homework-id", "1", "--class-id", "c", "--user-name", "张三"},
		{"homework", "class-detail", "--homework-id", "1", "--class-id", "c", "--user-name", "张三"},
		{"homework", "submit-statistics", "--homework-id", "1", "--class-id", "c", "--user-name", "张三"},
		{"homework", "list-by-student", "--student-id", "u1", "--class-id", "c", "--user-name", "张三", "--status", "FINISHED", "--page", "1", "--page-size", "20"},
		{"homework", "student-detail", "--homework-id", "1", "--student-id", "u1", "--class-id", "c", "--user-name", "张三"},
		{"homework", "list-by-teacher", "--class-id", "c", "--type", "HOMEWORK", "--status", "FINISHED", "--user-name", "u", "--page", "1", "--page-size", "20"},
		{"homework", "create-comment", "--comment", "做得很好", "--hw-content-detail-id", "1", "--homework-id", "2",
			"--student-id", "u1", "--photo", "p", "--video", "v", "--media", "m", "--user-name", "张三"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			withEduAppCaller(t)
			if err := runEduApp(t, args...); err != nil {
				t.Fatalf("Execute(%v) = %v, want nil", args, err)
			}
		})
	}
}

// TestEduAppZeroPagination drives the pagination defaulting branches (page<=0 /
// page-size<=0 / the >0 else arms) that the positive-value happy paths skip.
func TestCrossPlatformCoverageEduAppZeroPagination(t *testing.T) {
	cases := [][]string{
		{"report", "by-teacher", "--page", "0", "--limit", "0"},
		{"report", "by-student-list", "--class-id", "c", "--student-id", "u1", "--page", "0", "--limit", "0"},
		{"notice", "list-by-teacher", "--page", "0", "--page-size", "0"},
		{"notice", "confirm-status", "--notice-id", "1", "--class-id", "c", "--page", "0", "--page-size", "0"},
		{"notice", "list-by-student", "--student-id", "u1", "--class-id", "c", "--page", "0", "--page-size", "0"},
		{"card", "list", "--status", "FINISH", "--page", "0", "--limit", "0"},
		{"card", "user-statistic", "--card-id", "1", "--task-code", "code1", "--class-id", "cid1", "--page", "0", "--limit", "0"},
		{"diploma", "list-by-teacher", "--page", "0", "--limit", "0"},
		{"diploma", "list-by-student", "--student-id", "u1", "--class-id", "c", "--page", "0", "--limit", "0"},
		{"homework", "list-by-student", "--student-id", "u1", "--class-id", "c", "--user-name", "张三", "--page", "0", "--page-size", "0"},
		{"homework", "list-by-teacher", "--page", "0", "--page-size", "0"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			withEduAppCaller(t)
			if err := runEduApp(t, args...); err != nil {
				t.Fatalf("Execute(%v) = %v, want nil", args, err)
			}
		})
	}
}

func TestCrossPlatformCoverageEduAppErrorPaths(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"summary-list missing class-id", []string{"message", "summary-list", "--cid", "c", "--target-role", "guardian", "--status", "0"}, "class-id"},
		{"summary-list non-int class-id", []string{"message", "summary-list", "--class-id", "abc", "--cid", "c", "--target-role", "guardian", "--status", "0"}, "整数"},
		{"summary-list missing cid", []string{"message", "summary-list", "--class-id", "1", "--target-role", "guardian", "--status", "0"}, "cid"},
		{"summary-list missing target-role", []string{"message", "summary-list", "--class-id", "1", "--cid", "c", "--status", "0"}, "target-role"},
		{"summary-list missing status", []string{"message", "summary-list", "--class-id", "1", "--cid", "c", "--target-role", "guardian"}, "status"},
		{"summary-list non-int status", []string{"message", "summary-list", "--class-id", "1", "--cid", "c", "--target-role", "guardian", "--status", "x"}, "status"},

		{"all-list missing biz-id", []string{"task", "all-list"}, "biz-id"},
		{"student-list missing students", []string{"task", "student-list"}, "students"},
		{"student-list bad json", []string{"task", "student-list", "--students", "{"}, "JSON"},

		{"report get missing ids", []string{"report", "get"}, "ids"},
		{"report get non-int ids", []string{"report", "get", "--ids", "abc"}, "整数"},
		{"report by-class missing report-id", []string{"report", "by-class", "--class-id", "c"}, "report-id"},
		{"report by-class non-int report-id", []string{"report", "by-class", "--report-id", "x", "--class-id", "c"}, "整数"},
		{"report by-class missing class-id", []string{"report", "by-class", "--report-id", "1"}, "class-id"},
		{"report by-student-list missing class-id", []string{"report", "by-student-list", "--student-id", "u1"}, "class-id"},
		{"report by-student-list missing student-id", []string{"report", "by-student-list", "--class-id", "c"}, "student-id"},
		{"report by-student-detail missing report-id", []string{"report", "by-student-detail", "--student-id", "u1", "--class-id", "c"}, "report-id"},
		{"report by-student-detail non-int report-id", []string{"report", "by-student-detail", "--report-id", "x", "--student-id", "u1", "--class-id", "c"}, "整数"},
		{"report by-student-detail missing student-id", []string{"report", "by-student-detail", "--report-id", "1", "--class-id", "c"}, "student-id"},
		{"report by-student-detail missing class-id", []string{"report", "by-student-detail", "--report-id", "1", "--student-id", "u1"}, "class-id"},

		{"notice confirm missing notice-id", []string{"notice", "confirm", "--student-id", "u1"}, "notice-id"},
		{"notice confirm missing student-id", []string{"notice", "confirm", "--notice-id", "n1"}, "student-id"},
		{"notice create missing identifer", []string{"notice", "create", "--content", "c"}, "identifer"},
		{"notice create missing content", []string{"notice", "create", "--identifer", "x"}, "content"},
		{"notice create bad selected-students", []string{"notice", "create", "--identifer", "x", "--content", "c", "--class-selected-students", "{"}, "class-selected-students"},
		{"notice create bad attributes", []string{"notice", "create", "--identifer", "x", "--content", "c", "--attributes", "{"}, "attributes"},
		{"notice delete missing notice-id", []string{"notice", "delete"}, "notice-id"},
		{"notice delete non-int notice-id", []string{"notice", "delete", "--notice-id", "x"}, "整数"},
		{"notice get missing notice-id", []string{"notice", "get"}, "notice-id"},
		{"notice confirm-status missing notice-id", []string{"notice", "confirm-status", "--class-id", "c"}, "notice-id"},
		{"notice confirm-status missing class-id", []string{"notice", "confirm-status", "--notice-id", "1"}, "class-id"},
		{"notice list-by-student missing student-id", []string{"notice", "list-by-student", "--class-id", "c"}, "student-id"},
		{"notice list-by-student missing class-id", []string{"notice", "list-by-student", "--student-id", "u1"}, "class-id"},

		{"circle posts missing class-id", []string{"circle", "posts", "--student-id", "u1", "--target-role", "guardian"}, "class-id"},
		{"circle posts missing student-id", []string{"circle", "posts", "--class-id", "c", "--target-role", "guardian"}, "student-id"},
		{"circle posts missing target-role", []string{"circle", "posts", "--class-id", "c", "--student-id", "u1"}, "target-role"},

		{"card update missing card-id", []string{"card", "update", "--identifier", "i", "--title", "t"}, "card-id"},
		{"card update missing identifier", []string{"card", "update", "--card-id", "1", "--title", "t"}, "identifier"},
		{"card update no title no content", []string{"card", "update", "--card-id", "1", "--identifier", "i"}, "至少传一个"},
		{"card end missing card-id", []string{"card", "end"}, "card-id"},
		{"card list missing status", []string{"card", "list"}, "status"},
		{"card list invalid status", []string{"card", "list", "--status", "OTHER"}, "FINISH"},
		{"card user-statistic missing card-id", []string{"card", "user-statistic", "--task-code", "c", "--class-id", "c"}, "card-id"},
		{"card user-statistic missing task-code", []string{"card", "user-statistic", "--card-id", "1", "--class-id", "c"}, "task-code"},
		{"card user-statistic missing class-id", []string{"card", "user-statistic", "--card-id", "1", "--task-code", "c"}, "class-id"},
		{"card finish-info missing card-id", []string{"card", "finish-info", "--card-biz-id", "b"}, "card-id"},
		{"card finish-info missing card-biz-id", []string{"card", "finish-info", "--card-id", "1"}, "card-biz-id"},
		{"card finish-info invalid target-role", []string{"card", "finish-info", "--card-id", "1", "--card-biz-id", "b", "--target-role", "boss"}, "target-role"},

		{"diploma create missing identifier", []string{"diploma", "create", "--content", "c", "--user-name", "u"}, "identifier"},
		{"diploma create missing content", []string{"diploma", "create", "--identifier", "i", "--user-name", "u"}, "content"},
		{"diploma create missing user-name", []string{"diploma", "create", "--identifier", "i", "--content", "c"}, "user-name"},
		{"diploma create bad select-class", []string{"diploma", "create", "--identifier", "i", "--content", "c", "--user-name", "u", "--select-class", "{"}, "select-class"},
		{"diploma create bad attributes", []string{"diploma", "create", "--identifier", "i", "--content", "c", "--user-name", "u", "--attributes", "{"}, "attributes"},
		{"diploma read missing diploma-id", []string{"diploma", "read"}, "diploma-id"},
		{"diploma get missing diploma-id", []string{"diploma", "get"}, "diploma-id"},
		{"diploma statistics missing diploma-id", []string{"diploma", "statistics"}, "diploma-id"},
		{"diploma detail missing diploma-id", []string{"diploma", "detail"}, "diploma-id"},
		{"diploma list-by-student missing student-id", []string{"diploma", "list-by-student", "--class-id", "c"}, "student-id"},
		{"diploma list-by-student missing class-id", []string{"diploma", "list-by-student", "--student-id", "u1"}, "class-id"},
		{"diploma student-detail missing diploma-id", []string{"diploma", "student-detail", "--student-id", "u1", "--class-id", "c"}, "diploma-id"},
		{"diploma student-detail missing student-id", []string{"diploma", "student-detail", "--diploma-id", "1", "--class-id", "c"}, "student-id"},
		{"diploma student-detail missing class-id", []string{"diploma", "student-detail", "--diploma-id", "1", "--student-id", "u1"}, "class-id"},
		{"diploma delete missing diploma-id", []string{"diploma", "delete"}, "diploma-id"},

		{"homework create missing identifier", []string{"homework", "create", "--hw-content", "c"}, "identifier"},
		{"homework create missing hw-content", []string{"homework", "create", "--identifier", "i"}, "hw-content"},
		{"homework create bad hw-deadline", []string{"homework", "create", "--identifier", "i", "--hw-content", "c", "--hw-deadline", "x"}, "hw-deadline"},
		{"homework create bad task-plan-duration", []string{"homework", "create", "--identifier", "i", "--hw-content", "c", "--task-plan-duration", "x"}, "task-plan-duration"},
		{"homework create bad selected-students", []string{"homework", "create", "--identifier", "i", "--hw-content", "c", "--class-selected-students", "{"}, "class-selected-students"},
		{"homework create bad attributes", []string{"homework", "create", "--identifier", "i", "--hw-content", "c", "--attributes", "{"}, "attributes"},
		{"homework delete missing homework-id", []string{"homework", "delete"}, "homework-id"},
		{"homework submit missing detail-id", []string{"homework", "submit"}, "hw-content-detail-id"},
		{"homework submit bad homework-id", []string{"homework", "submit", "--hw-content-detail-id", "1", "--homework-id", "x"}, "homework-id"},
		{"homework get missing homework-id", []string{"homework", "get"}, "homework-id"},
		{"homework class-by-homework missing homework-id", []string{"homework", "class-by-homework"}, "homework-id"},
		{"homework class-detail missing homework-id", []string{"homework", "class-detail", "--class-id", "c", "--user-name", "u"}, "homework-id"},
		{"homework class-detail missing class-id", []string{"homework", "class-detail", "--homework-id", "1", "--user-name", "u"}, "class-id"},
		{"homework class-detail missing user-name", []string{"homework", "class-detail", "--homework-id", "1", "--class-id", "c"}, "user-name"},
		{"homework submit-statistics missing homework-id", []string{"homework", "submit-statistics", "--class-id", "c"}, "homework-id"},
		{"homework submit-statistics missing class-id", []string{"homework", "submit-statistics", "--homework-id", "1"}, "class-id"},
		{"homework list-by-student missing student-id", []string{"homework", "list-by-student", "--class-id", "c", "--user-name", "u"}, "student-id"},
		{"homework list-by-student missing class-id", []string{"homework", "list-by-student", "--student-id", "u1", "--user-name", "u"}, "class-id"},
		{"homework list-by-student missing user-name", []string{"homework", "list-by-student", "--student-id", "u1", "--class-id", "c"}, "user-name"},
		{"homework student-detail missing homework-id", []string{"homework", "student-detail", "--student-id", "u1", "--class-id", "c"}, "homework-id"},
		{"homework student-detail missing student-id", []string{"homework", "student-detail", "--homework-id", "1", "--class-id", "c"}, "student-id"},
		{"homework student-detail missing class-id", []string{"homework", "student-detail", "--homework-id", "1", "--student-id", "u1"}, "class-id"},
		{"homework create-comment missing comment", []string{"homework", "create-comment", "--hw-content-detail-id", "1"}, "comment"},
		{"homework create-comment missing detail-id", []string{"homework", "create-comment", "--comment", "cm"}, "hw-content-detail-id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withEduAppCaller(t)
			err := runEduApp(t, tc.args...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Execute(%v) error = %v, want contains %q", tc.args, err, tc.want)
			}
		})
	}
}

func TestCrossPlatformCoverageEduAppParseHelpers(t *testing.T) {
	if got := eduAppParseCSV(" a , , b "); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("eduAppParseCSV = %#v", got)
	}
	ids, err := eduAppParseIntCSV(" 1 , , 2 ")
	if err != nil || len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("eduAppParseIntCSV = %#v, err = %v", ids, err)
	}
	if _, err := eduAppParseIntCSV("1,bad"); err == nil {
		t.Fatalf("eduAppParseIntCSV invalid = nil error")
	}
}

func withEduAppDispatchCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func TestCrossPlatformCoverageEduAppDispatch(t *testing.T) {
	t.Run("message summary-list dispatches get_ai_message_summary_list", func(t *testing.T) {
		caller := withEduAppDispatchCaller(t)
		cmd := newEduAppCommand()
		cmd.SetArgs([]string{"message", "summary-list", "--class-id", "100", "--cid", "cidxxx", "--target-role", "guardian", "--status", "1"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-app" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-app")
		}
		if caller.tool != "get_ai_message_summary_list" {
			t.Fatalf("tool = %q, want %q", caller.tool, "get_ai_message_summary_list")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] type = %T, want map[string]any", caller.args["input"])
		}
		if input["classId"] != int64(100) {
			t.Fatalf("classId = %v, want 100", input["classId"])
		}
		if input["cid"] != "cidxxx" {
			t.Fatalf("cid = %v, want %q", input["cid"], "cidxxx")
		}
		if input["targetRole"] != "guardian" {
			t.Fatalf("targetRole = %v, want %q", input["targetRole"], "guardian")
		}
		if input["status"] != int64(1) {
			t.Fatalf("status = %v, want 1", input["status"])
		}
	})

	t.Run("report get dispatches get_report", func(t *testing.T) {
		caller := withEduAppDispatchCaller(t)
		cmd := newEduAppCommand()
		cmd.SetArgs([]string{"report", "get", "--ids", "1001,1002"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-app" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-app")
		}
		if caller.tool != "get_report" {
			t.Fatalf("tool = %q, want %q", caller.tool, "get_report")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] type = %T, want map[string]any", caller.args["input"])
		}
		ids, ok := input["schoolReportIdList"].([]int64)
		if !ok || len(ids) != 2 || ids[0] != 1001 || ids[1] != 1002 {
			t.Fatalf("schoolReportIdList = %v, want [1001 1002]", input["schoolReportIdList"])
		}
	})

	t.Run("circle posts dispatches query_student_circle_posts", func(t *testing.T) {
		caller := withEduAppDispatchCaller(t)
		cmd := newEduAppCommand()
		cmd.SetArgs([]string{"circle", "posts", "--class-id", "12345", "--student-id", "stu1", "--target-role", "guardian"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-app" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-app")
		}
		if caller.tool != "query_student_circle_posts" {
			t.Fatalf("tool = %q, want %q", caller.tool, "query_student_circle_posts")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] type = %T, want map[string]any", caller.args["input"])
		}
		if input["classId"] != "12345" {
			t.Fatalf("classId = %v, want %q", input["classId"], "12345")
		}
		if input["studentId"] != "stu1" {
			t.Fatalf("studentId = %v, want %q", input["studentId"], "stu1")
		}
		if input["targetRole"] != "guardian" {
			t.Fatalf("targetRole = %v, want %q", input["targetRole"], "guardian")
		}
	})

	t.Run("card end dispatches end_card", func(t *testing.T) {
		caller := withEduAppDispatchCaller(t)
		cmd := newEduAppCommand()
		cmd.SetArgs([]string{"card", "end", "--card-id", "999"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-app" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-app")
		}
		if caller.tool != "end_card" {
			t.Fatalf("tool = %q, want %q", caller.tool, "end_card")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] type = %T, want map[string]any", caller.args["input"])
		}
		if input["cardId"] != int64(999) {
			t.Fatalf("cardId = %v, want 999", input["cardId"])
		}
	})

	t.Run("card update dispatches update_card", func(t *testing.T) {
		caller := withEduAppDispatchCaller(t)
		cmd := newEduAppCommand()
		cmd.SetArgs([]string{"card", "update", "--card-id", "77", "--identifier", "org1-staff1-uuid", "--title", "新标题", "--should-send-update-msg"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-app" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-app")
		}
		if caller.tool != "update_card" {
			t.Fatalf("tool = %q, want %q", caller.tool, "update_card")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] type = %T, want map[string]any", caller.args["input"])
		}
		if input["cardId"] != int64(77) {
			t.Fatalf("cardId = %v, want 77", input["cardId"])
		}
		if input["identifier"] != "org1-staff1-uuid" {
			t.Fatalf("identifier = %v, want %q", input["identifier"], "org1-staff1-uuid")
		}
		if input["title"] != "新标题" {
			t.Fatalf("title = %v, want %q", input["title"], "新标题")
		}
		if input["shouldSendUpdateMsg"] != true {
			t.Fatalf("shouldSendUpdateMsg = %v, want true", input["shouldSendUpdateMsg"])
		}
	})
}

// newEduAppConfirmRoot 模拟真实运行时的根命令：核心框架在 rootCmd 上注册
// 全局 persistent --yes flag，叶子命令通过合并后的 Flags() 读取。
func newEduAppConfirmRoot() *cobra.Command {
	root := &cobra.Command{Use: "dws"}
	root.PersistentFlags().BoolP("yes", "y", false, "跳过确认提示")
	root.AddCommand(newEduAppCommand())
	return root
}

// TestCrossPlatformCoverageEduAppDestructiveConfirmGate 对 edu-app 每个
// user_required 破坏性叶子做成对验证：
//   - 未显式确认：返回 confirmation_required 错误，且 caller 调用次数为零。
//   - 显式确认后：恰好一次 MCP 调用，且 productID、tool、完整参数均准确。
func TestCrossPlatformCoverageEduAppDestructiveConfirmGate(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantTool  string
		wantInput map[string]any
	}{
		{
			"notice delete",
			[]string{"edu-app", "notice", "delete", "--notice-id", "12345"},
			"delete_notice",
			map[string]any{"noticeId": int64(12345)},
		},
		{
			"notice delete with user-name",
			[]string{"edu-app", "notice", "delete", "--notice-id", "12345", "--user-name", "张三"},
			"delete_notice",
			map[string]any{"noticeId": int64(12345), "userName": "张三"},
		},
		{
			"homework delete",
			[]string{"edu-app", "homework", "delete", "--homework-id", "12345"},
			"delete_homework",
			map[string]any{"homeworkId": int64(12345)},
		},
		{
			"diploma delete",
			[]string{"edu-app", "diploma", "delete", "--diploma-id", "12345"},
			"delete_diploma",
			map[string]any{"diplomaId": int64(12345)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/rejected_without_yes", func(t *testing.T) {
			caller := &recruitCaptureCaller{dryRun: false}
			InitDepsForTest(t, caller)
			deps.Out.w = io.Discard

			root := newEduAppConfirmRoot()
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

			root := newEduAppConfirmRoot()
			root.SetArgs(append(append([]string{}, tc.args...), "--yes"))
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute() with --yes error = %v", err)
			}
			if len(caller.calls) != 1 {
				t.Fatalf("expected exactly 1 MCP call with --yes, got %d", len(caller.calls))
			}
			if caller.calls[0].productID != "edu-app" {
				t.Errorf("productID = %q, want %q", caller.calls[0].productID, "edu-app")
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
