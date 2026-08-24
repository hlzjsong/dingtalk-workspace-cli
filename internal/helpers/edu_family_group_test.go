// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");

package helpers

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func withEduFamilyGroupCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{dryRun: true}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func TestCrossPlatformCoverageEduFamilyGroupHappyPaths(t *testing.T) {
	cases := [][]string{
		{"group", "check-exists", "--uid", "1", "--group-name", "小明一家"},
		{"group", "list-children", "--uid", "1"},
		{"manage", "create", "--uid", "1",
			"--children", `[{"name":"小明","students":[{"corpId":"c","staffId":"s"}]}]`,
			"--add-group", `{"schoolCorpId":"x"}`, "--source", "1"},
		{"manage", "invite-parent", "--org-id", "1", "--uid", "2", "--mobile", "13800138000"},
		{"manage", "add-child", "--org-id", "1", "--uid", "2", "--name", "小明", "--mobile", "13900139000"},
		{"manage", "add-child", "--org-id", "1", "--uid", "2", "--name", "小明",
			"--students", `[{"schoolOrgId":111,"studentStaffId":"stu001"}]`},
		{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s",
			"--app-type", "XIAOTIANDI", "--open", "true"},
		{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s",
			"--app-type", "LEARNING_VIDEO", "--open", "false"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			withEduFamilyGroupCaller(t)
			cmd := newEduFamilyGroupCommand()
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute(%v) = %v, want nil", args, err)
			}
		})
	}
}

func TestCrossPlatformCoverageEduFamilyGroupErrorPaths(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"check-exists missing uid", []string{"group", "check-exists", "--group-name", "x"}, "uid"},
		{"check-exists non-int uid", []string{"group", "check-exists", "--uid", "abc", "--group-name", "x"}, "整数"},
		{"check-exists missing group-name", []string{"group", "check-exists", "--uid", "1"}, "group-name"},
		{"list-children missing uid", []string{"group", "list-children"}, "uid"},
		{"create missing uid", []string{"manage", "create", "--children", "[]"}, "uid"},
		{"create missing children", []string{"manage", "create", "--uid", "1"}, "children"},
		{"create children bad json", []string{"manage", "create", "--uid", "1", "--children", "{"}, "JSON"},
		{"create children empty array", []string{"manage", "create", "--uid", "1", "--children", "[]"}, "不能为空数组"},
		{"create add-group bad json", []string{"manage", "create", "--uid", "1",
			"--children", `[{"name":"小明","students":[{"corpId":"c","staffId":"s"}]}]`,
			"--add-group", "{"}, "add-group"},
		{"create add-group null", []string{"manage", "create", "--uid", "1",
			"--children", `[{"name":"小明","students":[{"corpId":"c","staffId":"s"}]}]`,
			"--add-group", "null"}, "null"},
		{"invite-parent missing org-id", []string{"manage", "invite-parent", "--uid", "2", "--mobile", "138"}, "org-id"},
		{"invite-parent missing uid", []string{"manage", "invite-parent", "--org-id", "1", "--mobile", "138"}, "uid"},
		{"invite-parent missing mobile", []string{"manage", "invite-parent", "--org-id", "1", "--uid", "2"}, "mobile"},
		{"add-child missing org-id", []string{"manage", "add-child", "--uid", "2", "--name", "x", "--mobile", "138"}, "org-id"},
		{"add-child missing uid", []string{"manage", "add-child", "--org-id", "1", "--name", "x", "--mobile", "138"}, "uid"},
		{"add-child missing name", []string{"manage", "add-child", "--org-id", "1", "--uid", "2", "--mobile", "138"}, "name"},
		{"add-child no mobile no students", []string{"manage", "add-child", "--org-id", "1", "--uid", "2", "--name", "x"}, "至少传一个"},
		{"add-child students bad json", []string{"manage", "add-child", "--org-id", "1", "--uid", "2", "--name", "x", "--students", "{"}, "students"},
		{"add-child students empty", []string{"manage", "add-child", "--org-id", "1", "--uid", "2", "--name", "x", "--students", "[]"}, "不能为空数组"},
		{"toggle-app missing org-id", []string{"manage", "toggle-app", "--uid", "2", "--child-staff-id", "s", "--app-type", "XIAOTIANDI", "--open", "true"}, "org-id"},
		{"toggle-app missing uid", []string{"manage", "toggle-app", "--org-id", "1", "--child-staff-id", "s", "--app-type", "XIAOTIANDI", "--open", "true"}, "uid"},
		{"toggle-app missing child-staff-id", []string{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--app-type", "XIAOTIANDI", "--open", "true"}, "child-staff-id"},
		{"toggle-app missing app-type", []string{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s", "--open", "true"}, "app-type"},
		{"toggle-app invalid app-type", []string{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s", "--app-type", "OTHER", "--open", "true"}, "XIAOTIANDI"},
		{"toggle-app missing open", []string{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s", "--app-type", "XIAOTIANDI"}, "open"},
		{"toggle-app invalid open", []string{"manage", "toggle-app", "--org-id", "1", "--uid", "2", "--child-staff-id", "s", "--app-type", "XIAOTIANDI", "--open", "maybe"}, "true 或 false"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withEduFamilyGroupCaller(t)
			cmd := newEduFamilyGroupCommand()
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Execute(%v) error = %v, want contains %q", tc.args, err, tc.want)
			}
		})
	}
}

func TestCrossPlatformCoverageEduFamilyGroupValidateChildren(t *testing.T) {
	valid := []any{map[string]any{
		"name":     "小明",
		"students": []any{map[string]any{"corpId": "c", "staffId": "s"}},
	}}
	if err := eduFamilyGroupValidateChildren(valid); err != nil {
		t.Fatalf("valid children error = %v", err)
	}
	cases := []struct {
		name     string
		children []any
		want     string
	}{
		{"empty", []any{}, "不能为空数组"},
		{"non-object", []any{1}, "须为 JSON 对象"},
		{"missing name", []any{map[string]any{"students": []any{map[string]any{"corpId": "c", "staffId": "s"}}}}, "name 为必填"},
		{"missing students", []any{map[string]any{"name": "x"}}, "students 为必填"},
		{"student non-object", []any{map[string]any{"name": "x", "students": []any{1}}}, "须为 JSON 对象"},
		{"missing corpId", []any{map[string]any{"name": "x", "students": []any{map[string]any{"staffId": "s"}}}}, "corpId 为必填"},
		{"missing staffId", []any{map[string]any{"name": "x", "students": []any{map[string]any{"corpId": "c"}}}}, "staffId 为必填"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := eduFamilyGroupValidateChildren(tc.children); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestCrossPlatformCoverageEduFamilyGroupValidateStudents(t *testing.T) {
	valid := []any{map[string]any{"schoolOrgId": float64(111), "studentStaffId": "stu001"}}
	if err := eduFamilyGroupValidateStudents(valid); err != nil {
		t.Fatalf("valid students error = %v", err)
	}
	cases := []struct {
		name     string
		students []any
		want     string
	}{
		{"empty", []any{}, "不能为空数组"},
		{"non-object", []any{1}, "须为 JSON 对象"},
		{"missing schoolOrgId", []any{map[string]any{"studentStaffId": "s"}}, "schoolOrgId 为必填"},
		{"missing studentStaffId", []any{map[string]any{"schoolOrgId": float64(1)}}, "studentStaffId 为必填"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := eduFamilyGroupValidateStudents(tc.students); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func withEduFamilyGroupDispatchCaller(t *testing.T) *recruitCaptureCaller {
	t.Helper()
	caller := &recruitCaptureCaller{dryRun: false}
	InitDepsForTest(t, caller)
	deps.Out.w = io.Discard
	return caller
}

func TestCrossPlatformCoverageEduFamilyGroupDispatch(t *testing.T) {
	t.Run("check-exists", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"group", "check-exists", "--uid", "123", "--group-name", "测试家庭"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "check_family_group_exists" {
			t.Fatalf("tool = %q, want %q", caller.tool, "check_family_group_exists")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] = %#v, want map[string]any", caller.args["input"])
		}
		if input["uid"] != int64(123) {
			t.Fatalf("input[\"uid\"] = %#v, want int64(123)", input["uid"])
		}
		if input["groupName"] != "测试家庭" {
			t.Fatalf("input[\"groupName\"] = %#v, want %q", input["groupName"], "测试家庭")
		}
	})

	t.Run("list-children", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"group", "list-children", "--uid", "456"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "listBoundChildren" {
			t.Fatalf("tool = %q, want %q", caller.tool, "listBoundChildren")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] = %#v, want map[string]any", caller.args["input"])
		}
		if input["uid"] != int64(456) {
			t.Fatalf("input[\"uid\"] = %#v, want int64(456)", input["uid"])
		}
	})

	t.Run("create", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"manage", "create", "--uid", "789",
			"--children", `[{"name":"小明","students":[{"corpId":"c","staffId":"s"}]}]`})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "create_family_group" {
			t.Fatalf("tool = %q, want %q", caller.tool, "create_family_group")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] = %#v, want map[string]any", caller.args["input"])
		}
		if input["uid"] != int64(789) {
			t.Fatalf("input[\"uid\"] = %#v, want int64(789)", input["uid"])
		}
		if input["children"] == nil {
			t.Fatalf("input[\"children\"] is nil, want non-nil")
		}
	})

	t.Run("invite-parent", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"manage", "invite-parent", "--org-id", "1", "--uid", "2", "--mobile", "13800138000"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "invite_parent_to_familygroup" {
			t.Fatalf("tool = %q, want %q", caller.tool, "invite_parent_to_familygroup")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] = %#v, want map[string]any", caller.args["input"])
		}
		if input["orgId"] != int64(1) {
			t.Fatalf("input[\"orgId\"] = %#v, want int64(1)", input["orgId"])
		}
		if input["uid"] != int64(2) {
			t.Fatalf("input[\"uid\"] = %#v, want int64(2)", input["uid"])
		}
		if input["mobile"] != "13800138000" {
			t.Fatalf("input[\"mobile\"] = %#v, want %q", input["mobile"], "13800138000")
		}
	})

	t.Run("toggle-app", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"manage", "toggle-app", "--org-id", "1", "--uid", "2",
			"--child-staff-id", "s", "--app-type", "XIAOTIANDI", "--open", "true"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "toggle_student_app" {
			t.Fatalf("tool = %q, want %q", caller.tool, "toggle_student_app")
		}
		input, ok := caller.args["input"].(map[string]any)
		if !ok {
			t.Fatalf("args[\"input\"] = %#v, want map[string]any", caller.args["input"])
		}
		if input["appType"] != "XIAOTIANDI" {
			t.Fatalf("input[\"appType\"] = %#v, want %q", input["appType"], "XIAOTIANDI")
		}
		if input["open"] != true {
			t.Fatalf("input[\"open\"] = %#v, want true", input["open"])
		}
	})

	t.Run("add-child", func(t *testing.T) {
		caller := withEduFamilyGroupDispatchCaller(t)
		cmd := newEduFamilyGroupCommand()
		cmd.SetArgs([]string{"manage", "add-child", "--org-id", "12345", "--uid", "67890",
			"--name", "小明", "--mobile", "13900139000",
			"--students", `[{"schoolOrgId":111,"studentStaffId":"stu001"}]`})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() = %v", err)
		}
		if caller.productID != "edu-familygroup" {
			t.Fatalf("productID = %q, want %q", caller.productID, "edu-familygroup")
		}
		if caller.tool != "add_child_to_family_group" {
			t.Fatalf("tool = %q, want %q", caller.tool, "add_child_to_family_group")
		}
		want := map[string]any{
			"input": map[string]any{
				"orgId":  int64(12345),
				"uid":    int64(67890),
				"name":   "小明",
				"mobile": "13900139000",
				"students": []any{
					map[string]any{"schoolOrgId": float64(111), "studentStaffId": "stu001"},
				},
			},
		}
		if !reflect.DeepEqual(caller.args, want) {
			t.Fatalf("args = %#v, want %#v", caller.args, want)
		}
	})
}
