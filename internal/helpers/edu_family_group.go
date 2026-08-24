// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");

package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd/contract"
)

// ──────────────────────────────────────────────────────────
// dws edu-familygroup — 家庭群管理
// 共 6 个工具，按 group（读操作）/ manage（写操作）分组
// 参考 wukong/extensions/vendors/dingtalk/eduFamilyGroup.go 迁移
// ──────────────────────────────────────────────────────────

func newEduFamilyGroupCommand() *cobra.Command {
	contract.RegisterProductDecl(contract.ProductDecl{
		ID: "edu-familygroup",
		Selection: contract.ProductSelectionDecl{
			AgentSummary: "家庭群查询/创建、孩子管理、家长邀请、学生应用权限控制",
			UseWhen: []string{
				"用户要查询或管理钉钉家庭群、添加孩子、邀请家长或控制学生应用权限。",
			},
			AvoidWhen: []string{
				"家校通讯录用 edu-contact；班级师生群用 edu-group；家校应用/作业/打卡用 edu-app。",
			},
		},
	})
	root := &cobra.Command{
		Use:    "edu-familygroup",
		Short:  "家庭群",
		Long:   `钉钉家庭群管理：家庭群查询/创建、孩子管理、家长邀请、学生应用权限控制等。`,
		Hidden: true,
		RunE:   groupRunE,
	}

	// ════════════════════════════════════════════════════════════
	// group 子命令组 — 家庭群读操作
	// ════════════════════════════════════════════════════════════

	groupCmd := &cobra.Command{Use: "group", Short: "家庭群查询", RunE: groupRunE}

	groupCheckExistsCmd := &cobra.Command{
		Use:   "check-exists",
		Short: "检查家庭群是否存在",
		Long: `根据传入的 uid 拉取该用户所有家庭组织，按家庭群名称匹配判断家庭群是否存在。
仅当存在同名家庭且其群会话 cid 非空时，才认为家庭群存在，返回 true，否则返回 false。
面向家长（GUARDIAN）角色。`,
		Example: `  dws edu-familygroup group check-exists --uid 12345 --group-name "小明一家"
  dws edu-familygroup group check-exists --uid 12345 --group-name "小明一家" -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			groupName, err := eduFamilyGroupRequiredStringFlag(cmd, "group-name")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-familygroup", "check_family_group_exists", map[string]any{
				"input": map[string]any{"uid": uid, "groupName": groupName},
			})
		},
	}
	DeclareLeafMetadata(groupCheckExistsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "check_family_group_exists",
				CanonicalPath:  "edu-familygroup.check_family_group_exists",
				CLIPath:        "edu-familygroup group check-exists",
				PrimaryCLIPath: "edu-familygroup group check-exists",
			},
			Description: "检查家庭群是否存在",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "check_family_group_exists"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "检查家庭群是否存在",
				UseWhen:      []string{"需要判断指定用户名下是否存在同名家庭群时"},
				AvoidWhen:    []string{"查询家庭成员信息用 list-children"},
				Examples: []string{
					"dws edu-familygroup group check-exists --uid 12345 --group-name \"小明一家\" --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
				{Name: "group-name", Property: "input.groupName", Required: boolPtr(true)},
			},
		},
	})

	groupListChildrenCmd := &cobra.Command{
		Use:   "list-children",
		Short: "查询家长绑定的孩子列表",
		Long: `查询当前用户（uid）作为家长身份所在家庭中的所有孩子信息（不限家庭组织），
包含孩子基本信息及关联的学生号列表。底层按 uid 读扩散并完成家长身份校验，无孩子时返回空列表。
面向家长（GUARDIAN）角色。`,
		Example: `  dws edu-familygroup group list-children --uid 12345
  dws edu-familygroup group list-children --uid 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-familygroup", "listBoundChildren", map[string]any{
				"input": map[string]any{"uid": uid},
			})
		},
	}
	DeclareLeafMetadata(groupListChildrenCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "listBoundChildren",
				CanonicalPath:  "edu-familygroup.listBoundChildren",
				CLIPath:        "edu-familygroup group list-children",
				PrimaryCLIPath: "edu-familygroup group list-children",
			},
			Description: "查询家长绑定的孩子列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "listBoundChildren"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询家长绑定的孩子列表",
				UseWhen:      []string{"需要查询指定家长 uid 绑定的所有孩子及关联学生号信息时"},
				AvoidWhen:    []string{"查看家庭群是否存在用 check-exists"},
				Examples: []string{
					"dws edu-familygroup group list-children --uid 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// manage 子命令组 — 家庭群写操作
	// ════════════════════════════════════════════════════════════

	manageCmd := &cobra.Command{Use: "manage", Short: "家庭群管理", RunE: groupRunE}

	manageCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建家庭群",
		Long: `以 uid 作为创建人创建一个新的家庭，创建家庭时会同时创建家庭群，返回结果中携带群会话 cid。
children 为 JSON 数组，每个元素包含 name（必填）、students（必填，含 corpId + staffId）、
birthday / gender / nick / avatar / period / grade / mobile（均可选）。
addGroup 为 JSON 对象，含 schoolCorpId / schoolStaffId / inviteDingtalkId / inviteId（均可选）。
新建家庭场景下无需前置家长身份校验，创建人合法性由底层校验单元完成。`,
		Example: `  dws edu-familygroup manage create --uid 12345 --children '[{"name":"小明","students":[{"corpId":"dingxxx","staffId":"stu001"}]}]'
  dws edu-familygroup manage create --uid 12345 --children '[{"name":"小明","students":[{"corpId":"dingxxx","staffId":"stu001"}]}]' --source 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			childrenRaw, _ := cmd.Flags().GetString("children")
			if strings.TrimSpace(childrenRaw) == "" {
				return fmt.Errorf("--children 为必填参数")
			}
			var children []any
			if err := json.Unmarshal([]byte(childrenRaw), &children); err != nil {
				return fmt.Errorf("--children 须为合法 JSON 数组: %w", err)
			}
			if err := eduFamilyGroupValidateChildren(children); err != nil {
				return err
			}
			input := map[string]any{"uid": uid, "children": children}
			if v, _ := cmd.Flags().GetString("add-group"); strings.TrimSpace(v) != "" {
				var addGroup map[string]any
				if err := json.Unmarshal([]byte(v), &addGroup); err != nil {
					return fmt.Errorf("--add-group 须为合法 JSON 对象: %w", err)
				}
				if addGroup == nil {
					return fmt.Errorf("--add-group 须为 JSON 对象，不能为 null")
				}
				input["addGroup"] = addGroup
			}
			if v, _ := cmd.Flags().GetInt("source"); cmd.Flags().Changed("source") {
				input["source"] = v
			}
			return callMCPToolOnServer("edu-familygroup", "create_family_group", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(manageCreateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "create_family_group",
				CanonicalPath:  "edu-familygroup.create_family_group",
				CLIPath:        "edu-familygroup manage create",
				PrimaryCLIPath: "edu-familygroup manage create",
			},
			Description: "创建家庭群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "create_family_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建家庭群并同步创建家庭组织",
				UseWhen:      []string{"需要以指定 uid 创建新的家庭群，同时注册孩子信息并生成群会话时"},
				AvoidWhen:    []string{"已有家庭群要加孩子用 add-child"},
				Examples: []string{
					"dws edu-familygroup manage create --uid 12345 --children '[{\"name\":\"小明\",\"students\":[{\"corpId\":\"dingxxx\",\"staffId\":\"stu001\"}]}]' --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
				{Name: "children", Property: "input.children", Required: boolPtr(true)},
				{Name: "add-group", Property: "input.addGroup"},
				{Name: "source", Property: "input.source"},
			},
		},
	})

	manageInviteParentCmd := &cobra.Command{
		Use:   "invite-parent",
		Short: "短信邀请家长加入家庭群",
		Long: `通过短信向指定手机号的家长发送家庭群邀请链接，家长点击链接后加入当前家庭组织及家庭群。
返回 true 表示邀请短信已成功发送。仅家长（GUARDIAN）角色可调用。`,
		Example: `  dws edu-familygroup manage invite-parent --org-id 12345 --uid 67890 --mobile 13800138000
  dws edu-familygroup manage invite-parent --org-id 12345 --uid 67890 --mobile 13800138000 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID, err := eduFamilyGroupRequiredIntFlag(cmd, "org-id")
			if err != nil {
				return err
			}
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			mobile, err := eduFamilyGroupRequiredStringFlag(cmd, "mobile")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-familygroup", "invite_parent_to_familygroup", map[string]any{
				"input": map[string]any{"orgId": orgID, "uid": uid, "mobile": mobile},
			})
		},
	}
	DeclareLeafMetadata(manageInviteParentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "invite_parent_to_familygroup",
				CanonicalPath:  "edu-familygroup.invite_parent_to_familygroup",
				CLIPath:        "edu-familygroup manage invite-parent",
				PrimaryCLIPath: "edu-familygroup manage invite-parent",
			},
			Description: "短信邀请家长加入家庭群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "invite_parent_to_familygroup"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "通过短信邀请家长加入家庭群",
				UseWhen:      []string{"需要向指定手机号发送家庭群邀请短信时"},
				AvoidWhen:    []string{"添加孩子到家庭群用 add-child"},
				Examples: []string{
					"dws edu-familygroup manage invite-parent --org-id 12345 --uid 67890 --mobile 13800138000 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "org-id", Property: "input.orgId", Required: boolPtr(true)},
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(true)},
			},
		},
	})

	manageAddChildCmd := &cobra.Command{
		Use:   "add-child",
		Short: "为家庭群添加孩子",
		Long: `为指定家庭群添加孩子，支持三种方式（由底层自动路由）：
  - 仅传 --mobile：手机号邀请链路，该手机号对应的钉钉账号被邀请加入家庭群
  - 仅传 --students：直接生成学生号链路，选中学生后创建孩子并绑定关系
  - 同时传 --mobile + --students：mobile 优先，走手机号邀请链路
mobile 与 students 至少传一个。
students 为 JSON 数组，每个元素含 schoolOrgId（整数）和 studentStaffId（字符串），均必填。
仅家长（GUARDIAN）角色可调用。`,
		Example: `  dws edu-familygroup manage add-child --org-id 12345 --uid 67890 --name 小明 --mobile 13900139000
  dws edu-familygroup manage add-child --org-id 12345 --uid 67890 --name 小明 --students '[{"schoolOrgId":111,"studentStaffId":"stu001"}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID, err := eduFamilyGroupRequiredIntFlag(cmd, "org-id")
			if err != nil {
				return err
			}
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			name, err := eduFamilyGroupRequiredStringFlag(cmd, "name")
			if err != nil {
				return err
			}
			mobile, _ := cmd.Flags().GetString("mobile")
			mobile = strings.TrimSpace(mobile)
			studentsRaw, _ := cmd.Flags().GetString("students")
			studentsRaw = strings.TrimSpace(studentsRaw)
			if mobile == "" && studentsRaw == "" {
				return fmt.Errorf("--mobile 与 --students 至少传一个")
			}
			input := map[string]any{"orgId": orgID, "uid": uid, "name": name}
			if mobile != "" {
				input["mobile"] = mobile
			}
			if studentsRaw != "" {
				var students []any
				if err := json.Unmarshal([]byte(studentsRaw), &students); err != nil {
					return fmt.Errorf("--students 须为合法 JSON 数组: %w", err)
				}
				if err := eduFamilyGroupValidateStudents(students); err != nil {
					return err
				}
				input["students"] = students
			}
			return callMCPToolOnServer("edu-familygroup", "add_child_to_family_group", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(manageAddChildCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "add_child_to_family_group",
				CanonicalPath:  "edu-familygroup.add_child_to_family_group",
				CLIPath:        "edu-familygroup manage add-child",
				PrimaryCLIPath: "edu-familygroup manage add-child",
			},
			Description: "为家庭群添加孩子",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "add_child_to_family_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "为已有家庭群添加孩子",
				UseWhen:      []string{"需要向已有家庭群添加新孩子（通过手机号邀请或直接绑定学生号）时"},
				AvoidWhen:    []string{"创建全新家庭群用 create"},
				Examples: []string{
					"dws edu-familygroup manage add-child --org-id 12345 --uid 67890 --name 小明 --mobile 13900139000 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "org-id", Property: "input.orgId", Required: boolPtr(true)},
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
				{Name: "name", Property: "input.name", Required: boolPtr(true)},
				{Name: "mobile", Property: "input.mobile"},
				{Name: "students", Property: "input.students"},
			},
		},
	})

	manageToggleAppCmd := &cobra.Command{
		Use:   "toggle-app",
		Short: "开启或关闭学生应用权限",
		Long: `为指定学生号开启或关闭应用权限。
支持的应用类型（--app-type）：
  - XIAOTIANDI：小天地（学生圈）
  - LEARNING_VIDEO：学习视频
直接覆写权限状态，天然幂等。仅家长（GUARDIAN）角色可调用。`,
		Example: `  dws edu-familygroup manage toggle-app --org-id 12345 --uid 67890 --child-staff-id staff001 --app-type XIAOTIANDI --open true
  dws edu-familygroup manage toggle-app --org-id 12345 --uid 67890 --child-staff-id staff001 --app-type LEARNING_VIDEO --open false`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgID, err := eduFamilyGroupRequiredIntFlag(cmd, "org-id")
			if err != nil {
				return err
			}
			uid, err := eduFamilyGroupRequiredIntFlag(cmd, "uid")
			if err != nil {
				return err
			}
			childStaffID, err := eduFamilyGroupRequiredStringFlag(cmd, "child-staff-id")
			if err != nil {
				return err
			}
			appType, err := eduFamilyGroupRequiredStringFlag(cmd, "app-type")
			if err != nil {
				return err
			}
			if appType != "XIAOTIANDI" && appType != "LEARNING_VIDEO" {
				return fmt.Errorf("--app-type 须为 XIAOTIANDI 或 LEARNING_VIDEO")
			}
			openStr, err := eduFamilyGroupRequiredStringFlag(cmd, "open")
			if err != nil {
				return err
			}
			var open bool
			switch strings.ToLower(openStr) {
			case "true":
				open = true
			case "false":
				open = false
			default:
				return fmt.Errorf("--open 须为 true 或 false")
			}
			return callMCPToolOnServer("edu-familygroup", "toggle_student_app", map[string]any{
				"input": map[string]any{
					"orgId": orgID, "uid": uid, "childStaffId": childStaffID,
					"appType": appType, "open": open,
				},
			})
		},
	}
	DeclareLeafMetadata(manageToggleAppCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-familygroup",
				Name:           "toggle_student_app",
				CanonicalPath:  "edu-familygroup.toggle_student_app",
				CLIPath:        "edu-familygroup manage toggle-app",
				PrimaryCLIPath: "edu-familygroup manage toggle-app",
			},
			Description: "开启或关闭学生应用权限",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-familygroup", RPCName: "toggle_student_app"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "开启或关闭学生应用权限",
				UseWhen:      []string{"需要为指定学生号开启或关闭小天地/学习视频应用权限时"},
				AvoidWhen:    []string{"管理家庭群成员用 add-child / invite-parent"},
				Examples: []string{
					"dws edu-familygroup manage toggle-app --org-id 12345 --uid 67890 --child-staff-id staff001 --app-type XIAOTIANDI --open true --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "org-id", Property: "input.orgId", Required: boolPtr(true)},
				{Name: "uid", Property: "input.uid", Required: boolPtr(true)},
				{Name: "child-staff-id", Property: "input.childStaffId", Required: boolPtr(true)},
				{Name: "app-type", Property: "input.appType", Required: boolPtr(true)},
				{Name: "open", Property: "input.open", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// flags + 构建命令树
	// ════════════════════════════════════════════════════════════

	// group（读操作）flags
	groupCheckExistsCmd.Flags().String("uid", "", "用户 uid（必填）")
	groupCheckExistsCmd.Flags().String("group-name", "", "家庭群名称（必填）")
	groupListChildrenCmd.Flags().String("uid", "", "家长 uid（必填）")

	// manage（写操作）flags
	manageCreateCmd.Flags().String("uid", "", "创建人 uid（必填）")
	manageCreateCmd.Flags().String("children", "", "孩子信息 JSON 数组（必填）")
	manageCreateCmd.Flags().String("add-group", "", "同学群信息 JSON 对象（可选）")
	manageCreateCmd.Flags().Int("source", 0, "渠道来源（可选）")

	manageInviteParentCmd.Flags().String("org-id", "", "家庭组织 ID（必填）")
	manageInviteParentCmd.Flags().String("uid", "", "操作人 uid（必填）")
	manageInviteParentCmd.Flags().String("mobile", "", "被邀请家长手机号（必填）")

	manageAddChildCmd.Flags().String("org-id", "", "家庭组织 ID（必填）")
	manageAddChildCmd.Flags().String("uid", "", "操作人 uid（必填）")
	manageAddChildCmd.Flags().String("name", "", "孩子姓名（必填）")
	manageAddChildCmd.Flags().String("mobile", "", "孩子手机号（可选，与 --students 至少传一个）")
	manageAddChildCmd.Flags().String("students", "", "待关联学生号 JSON 数组（可选，每项含 schoolOrgId + studentStaffId）")

	manageToggleAppCmd.Flags().String("org-id", "", "家庭组织 ID（必填）")
	manageToggleAppCmd.Flags().String("uid", "", "家长 uid（必填）")
	manageToggleAppCmd.Flags().String("child-staff-id", "", "孩子在家庭组织中的 staffId（必填）")
	manageToggleAppCmd.Flags().String("app-type", "", "应用类型：XIAOTIANDI / LEARNING_VIDEO（必填）")
	manageToggleAppCmd.Flags().String("open", "", "true=开启 / false=关闭（必填）")

	groupCmd.AddCommand(groupCheckExistsCmd, groupListChildrenCmd)
	manageCmd.AddCommand(manageCreateCmd, manageInviteParentCmd, manageAddChildCmd, manageToggleAppCmd)

	root.AddCommand(groupCmd, manageCmd)

	return root
}

// eduFamilyGroupValidateChildren validates the --children payload: the array
// must be non-empty, each child must carry a non-empty name and a non-empty
// students array, and each student must carry corpId + staffId.
func eduFamilyGroupValidateChildren(children []any) error {
	if len(children) == 0 {
		return fmt.Errorf("--children 不能为空数组，至少需包含一个孩子")
	}
	for i, c := range children {
		child, ok := c.(map[string]any)
		if !ok {
			return fmt.Errorf("--children[%d] 须为 JSON 对象", i)
		}
		name, _ := child["name"].(string)
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("--children[%d].name 为必填字段", i)
		}
		students, ok := child["students"].([]any)
		if !ok || len(students) == 0 {
			return fmt.Errorf("--children[%d].students 为必填字段且不能为空数组", i)
		}
		for j, s := range students {
			student, ok := s.(map[string]any)
			if !ok {
				return fmt.Errorf("--children[%d].students[%d] 须为 JSON 对象", i, j)
			}
			corpID, _ := student["corpId"].(string)
			if strings.TrimSpace(corpID) == "" {
				return fmt.Errorf("--children[%d].students[%d].corpId 为必填字段", i, j)
			}
			staffID, _ := student["staffId"].(string)
			if strings.TrimSpace(staffID) == "" {
				return fmt.Errorf("--children[%d].students[%d].staffId 为必填字段", i, j)
			}
		}
	}
	return nil
}

// eduFamilyGroupValidateStudents validates the --students payload: the array
// must be non-empty and each element must carry a numeric schoolOrgId and a
// non-empty studentStaffId.
func eduFamilyGroupValidateStudents(students []any) error {
	if len(students) == 0 {
		return fmt.Errorf("--students 不能为空数组，至少需包含一个学生号")
	}
	for i, s := range students {
		student, ok := s.(map[string]any)
		if !ok {
			return fmt.Errorf("--students[%d] 须为 JSON 对象", i)
		}
		switch student["schoolOrgId"].(type) {
		case float64, int, int64, json.Number:
		default:
			return fmt.Errorf("--students[%d].schoolOrgId 为必填字段且须为整数", i)
		}
		staffID, _ := student["studentStaffId"].(string)
		if strings.TrimSpace(staffID) == "" {
			return fmt.Errorf("--students[%d].studentStaffId 为必填字段", i)
		}
	}
	return nil
}

// eduFamilyGroupRequiredIntFlag extracts a required integer flag, returning an
// error if the flag is empty or not a valid integer.
func eduFamilyGroupRequiredIntFlag(cmd *cobra.Command, name string) (int64, error) {
	v, _ := cmd.Flags().GetString(name)
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, fmt.Errorf("--%s 为必填参数", name)
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("--%s 须为整数: %w", name, err)
	}
	return n, nil
}

// eduFamilyGroupRequiredStringFlag extracts a required string flag.
func eduFamilyGroupRequiredStringFlag(cmd *cobra.Command, name string) (string, error) {
	v, _ := cmd.Flags().GetString(name)
	v = strings.TrimSpace(v)
	if v == "" {
		return "", fmt.Errorf("--%s 为必填参数", name)
	}
	return v, nil
}
