// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");

package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd/contract"
)

// ──────────────────────────────────────────────────────────
// dws edu-group — 家校群管理
// 共 14 个工具，按 student-group / class-group / batch 分组
// ──────────────────────────────────────────────────────────

func newEduGroupCommand() *cobra.Command {
	contract.RegisterProductDecl(contract.ProductDecl{
		ID: "edu-group",
		Selection: contract.ProductSelectionDecl{
			AgentSummary: "师生群查询/创建/解散、班级群会话信息查询、批量操作",
			UseWhen: []string{
				"用户要查询或管理钉钉师生群、班级群会话信息，或批量检查/创建师生群。",
			},
			AvoidWhen: []string{
				"家庭群用 edu-familygroup；家校通讯录用 edu-contact；家校应用/作业/打卡用 edu-app。",
			},
		},
	})
	root := &cobra.Command{
		Use:    "edu-group",
		Short:  "家校群",
		Long:   `钉钉家校群管理：师生群查询/创建/解散、班级群会话信息查询、批量操作等。`,
		Hidden: true,
		RunE:   groupRunE,
	}

	// ════════════════════════════════════════════════════════════
	// student-group 子命令组 — 师生群管理
	// ════════════════════════════════════════════════════════════

	studentGroupCmd := &cobra.Command{Use: "student-group", Short: "师生群管理", RunE: groupRunE}

	studentGroupInfoCmd := &cobra.Command{
		Use:   "info",
		Short: "查询班级师生群信息",
		Long:  `查询指定班级的师生群信息。返回师生群的群会话ID（cid）。管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group student-group info --dept-id 12345
  dws edu-group student-group info --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "get_class_group_info", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupInfoCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "get_class_group_info",
				CanonicalPath:  "edu-group.get_class_group_info",
				CLIPath:        "edu-group student-group info",
				PrimaryCLIPath: "edu-group student-group info",
			},
			Description: "查询班级师生群信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "get_class_group_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询班级师生群信息",
				UseWhen:      []string{"需要查询指定班级的师生群会话ID时"},
				AvoidWhen:    []string{"查询班级群会话详情用 student-group conversation"},
				Examples: []string{
					"dws edu-group student-group info --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupExistsCmd := &cobra.Command{
		Use:   "exists",
		Short: "检查组织是否已创建师生群",
		Long:  `检查指定组织下是否已创建师生群。返回是否存在师生群（true/false）。仅限管理员角色调用。`,
		Example: `  dws edu-group student-group exists --dept-id 12345
  dws edu-group student-group exists --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "check_class_group_exists", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupExistsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "check_class_group_exists",
				CanonicalPath:  "edu-group.check_class_group_exists",
				CLIPath:        "edu-group student-group exists",
				PrimaryCLIPath: "edu-group student-group exists",
			},
			Description: "检查组织是否已创建师生群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "check_class_group_exists"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "检查组织是否已创建师生群",
				UseWhen:      []string{"需要判断指定班级是否已创建师生群时"},
				AvoidWhen:    []string{"查询师生群成员用 student-group members"},
				Examples: []string{
					"dws edu-group student-group exists --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupMembersCmd := &cobra.Command{
		Use:   "members",
		Short: "查询师生群成员列表",
		Long:  `查询指定班级师生群的所有成员userId列表。管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group student-group members --dept-id 12345
  dws edu-group student-group members --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "get_group_members", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupMembersCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "get_group_members",
				CanonicalPath:  "edu-group.get_group_members",
				CLIPath:        "edu-group student-group members",
				PrimaryCLIPath: "edu-group student-group members",
			},
			Description: "查询师生群成员列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "get_group_members"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询师生群成员列表",
				UseWhen:      []string{"需要查询指定班级师生群的所有成员userId列表时"},
				AvoidWhen:    []string{"判断用户是否在群中用 student-group is-in"},
				Examples: []string{
					"dws edu-group student-group members --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupIsInCmd := &cobra.Command{
		Use:   "is-in",
		Short: "判断用户是否在师生群中",
		Long:  `判断当前用户是否在指定班级的师生群中。返回是否在群中（true/false）。管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group student-group is-in --dept-id 12345
  dws edu-group student-group is-in --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "is_in_class_group", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupIsInCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "is_in_class_group",
				CanonicalPath:  "edu-group.is_in_class_group",
				CLIPath:        "edu-group student-group is-in",
				PrimaryCLIPath: "edu-group student-group is-in",
			},
			Description: "判断用户是否在师生群中",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "is_in_class_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "判断用户是否在师生群中",
				UseWhen:      []string{"需要判断当前用户是否在指定班级师生群中时"},
				AvoidWhen:    []string{"查询群成员列表用 student-group members"},
				Examples: []string{
					"dws edu-group student-group is-in --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupConversationCmd := &cobra.Command{
		Use:   "conversation",
		Short: "查询班级群会话详情",
		Long: `查询指定班级师生群的会话详情。
返回群会话ID（cid）、群标题（title）、群成员数量（memberCount）和群图标URL（icon）。
管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group student-group conversation --dept-id 12345
  dws edu-group student-group conversation --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "get_group_conversation_info", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupConversationCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "get_group_conversation_info",
				CanonicalPath:  "edu-group.get_group_conversation_info",
				CLIPath:        "edu-group student-group conversation",
				PrimaryCLIPath: "edu-group student-group conversation",
			},
			Description: "查询班级群会话详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "get_group_conversation_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询班级师生群会话详情",
				UseWhen:      []string{"需要查询指定班级师生群的会话ID、标题、成员数、图标时"},
				AvoidWhen:    []string{"仅需群会话ID用 student-group info"},
				Examples: []string{
					"dws edu-group student-group conversation --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建班级师生群",
		Long: `为指定班级创建师生群。自动将班级的班主任设为群主，并拉入所有老师和学生。
前提是班级必须已设置班主任。返回创建成功的群会话ID（cid）。仅限管理员或班主任角色调用。`,
		Example: `  dws edu-group student-group create --dept-id 12345
  dws edu-group student-group create --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "create_class_group", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupCreateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "create_class_group",
				CanonicalPath:  "edu-group.create_class_group",
				CLIPath:        "edu-group student-group create",
				PrimaryCLIPath: "edu-group student-group create",
			},
			Description: "创建班级师生群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "create_class_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "为指定班级创建师生群",
				UseWhen:      []string{"需要为指定班级创建师生群，自动拉入班主任、老师和学生时"},
				AvoidWhen:    []string{"批量创建师生群用 batch create-student-groups"},
				Examples: []string{
					"dws edu-group student-group create --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	studentGroupDisbandCmd := &cobra.Command{
		Use:   "disband",
		Short: "解散班级师生群",
		Long:  `解散指定班级的师生群，同时删除班级与群的关联关系。仅限管理员或班主任角色调用。`,
		Example: `  dws edu-group student-group disband --dept-id 12345
  dws edu-group student-group disband --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "disband_class_group", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(studentGroupDisbandCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "disband_class_group",
				CanonicalPath:  "edu-group.disband_class_group",
				CLIPath:        "edu-group student-group disband",
				PrimaryCLIPath: "edu-group student-group disband",
			},
			Description: "解散班级师生群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "disband_class_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "解散班级师生群并删除关联关系",
				UseWhen:      []string{"需要解散指定班级的师生群并删除班级与群的关联关系时"},
				AvoidWhen:    []string{"查询师生群信息用 student-group info"},
				Examples: []string{
					"dws edu-group student-group disband --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// class-group 子命令组 — 班级群（家校群）会话管理
	// ════════════════════════════════════════════════════════════

	classGroupCmd := &cobra.Command{Use: "class-group", Short: "班级群会话管理", RunE: groupRunE}

	classGroupConversationIDCmd := &cobra.Command{
		Use:   "conversation-id",
		Short: "获取班级群会话ID",
		Long:  `获取指定班级的班级群会话ID，可用于后续发送群消息等操作。管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group class-group conversation-id --dept-id 12345
  dws edu-group class-group conversation-id --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "get_class_conversation_id", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classGroupConversationIDCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "get_class_conversation_id",
				CanonicalPath:  "edu-group.get_class_conversation_id",
				CLIPath:        "edu-group class-group conversation-id",
				PrimaryCLIPath: "edu-group class-group conversation-id",
			},
			Description: "获取班级群会话ID",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "get_class_conversation_id"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "获取班级群会话ID",
				UseWhen:      []string{"需要获取指定班级的班级群会话ID以便后续发送群消息时"},
				AvoidWhen:    []string{"需要完整群信息用 class-group conversation"},
				Examples: []string{
					"dws edu-group class-group conversation-id --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classGroupConversationCmd := &cobra.Command{
		Use:   "conversation",
		Short: "获取班级群完整会话信息",
		Long: `获取指定班级的班级群完整会话信息。
返回班级群的会话ID（cid）、群标题（title）、群成员数量（memberCount）和群图标URL（icon）。
管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group class-group conversation --dept-id 12345
  dws edu-group class-group conversation --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "get_class_conversation", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classGroupConversationCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "get_class_conversation",
				CanonicalPath:  "edu-group.get_class_conversation",
				CLIPath:        "edu-group class-group conversation",
				PrimaryCLIPath: "edu-group class-group conversation",
			},
			Description: "获取班级群完整会话信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "get_class_conversation"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "获取班级群完整会话信息",
				UseWhen:      []string{"需要查询指定班级的班级群会话ID、标题、成员数、图标时"},
				AvoidWhen:    []string{"仅需会话ID用 class-group conversation-id"},
				Examples: []string{
					"dws edu-group class-group conversation --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classGroupExistsCmd := &cobra.Command{
		Use:   "exists",
		Short: "检查班级群是否存在",
		Long:  `检查指定班级是否已创建班级群。返回班级群是否存在（true/false）。管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group class-group exists --dept-id 12345
  dws edu-group class-group exists --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduGroupRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-group", "check_class_conversation_exists", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classGroupExistsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "check_class_conversation_exists",
				CanonicalPath:  "edu-group.check_class_conversation_exists",
				CLIPath:        "edu-group class-group exists",
				PrimaryCLIPath: "edu-group class-group exists",
			},
			Description: "检查班级群是否存在",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "check_class_conversation_exists"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "检查班级群是否存在",
				UseWhen:      []string{"需要判断指定班级是否已创建班级群时"},
				AvoidWhen:    []string{"查询班级群会话信息用 class-group conversation"},
				Examples: []string{
					"dws edu-group class-group exists --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classGroupListByConversationIDsCmd := &cobra.Command{
		Use:   "list-by-cids",
		Short: "根据会话ID列表批量查询群信息",
		Long: `根据群会话ID列表批量查询群的详细信息。
返回群会话详情列表，每项包含会话ID（cid）、群标题（title）、群成员数量（memberCount）和群图标URL（icon）。
管理员、班主任、老师角色可调用。`,
		Example: `  dws edu-group class-group list-by-cids --conversation-ids cid1,cid2,cid3
  dws edu-group class-group list-by-cids --conversation-ids cid1,cid2 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("conversation-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--conversation-ids 为必填参数")
			}
			conversationIDs := eduGroupParseCSV(raw)
			if len(conversationIDs) == 0 {
				return fmt.Errorf("--conversation-ids 不能为空")
			}
			return callMCPToolOnServer("edu-group", "list_groups_by_conversation_ids", map[string]any{
				"input": map[string]any{"conversationIds": conversationIDs},
			})
		},
	}
	DeclareLeafMetadata(classGroupListByConversationIDsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "list_groups_by_conversation_ids",
				CanonicalPath:  "edu-group.list_groups_by_conversation_ids",
				CLIPath:        "edu-group class-group list-by-cids",
				PrimaryCLIPath: "edu-group class-group list-by-cids",
			},
			Description: "根据会话ID列表批量查询群信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "list_groups_by_conversation_ids"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "根据会话ID列表批量查询群信息",
				UseWhen:      []string{"需要根据一组群会话ID批量查询群的详细信息时"},
				AvoidWhen:    []string{"按班级ID批量查询用 batch get-class-groups"},
				Examples: []string{
					"dws edu-group class-group list-by-cids --conversation-ids cid1,cid2,cid3 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "conversation-ids", Property: "input.conversationIds", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// batch 子命令组 — 批量操作
	// ════════════════════════════════════════════════════════════

	batchCmd := &cobra.Command{Use: "batch", Short: "批量操作", RunE: groupRunE}

	batchCheckClassGroupCmd := &cobra.Command{
		Use:   "check-student-group",
		Short: "批量检查班级是否已创建师生群",
		Long:  `批量检查多个班级是否已创建师生群。返回班级ID与群会话ID（cid）的映射关系。仅限管理员角色调用。`,
		Example: `  dws edu-group batch check-student-group --class-ids 12345,67890
  dws edu-group batch check-student-group --class-ids 12345,67890 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("class-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--class-ids 为必填参数")
			}
			classIDs, err := eduGroupParseIntCSV(raw)
			if err != nil {
				return fmt.Errorf("--class-ids 须为逗号分隔的整数列表: %w", err)
			}
			return callMCPToolOnServer("edu-group", "batch_check_class_group", map[string]any{
				"input": map[string]any{"classIds": classIDs},
			})
		},
	}
	DeclareLeafMetadata(batchCheckClassGroupCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "batch_check_class_group",
				CanonicalPath:  "edu-group.batch_check_class_group",
				CLIPath:        "edu-group batch check-student-group",
				PrimaryCLIPath: "edu-group batch check-student-group",
			},
			Description: "批量检查班级是否已创建师生群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "batch_check_class_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量检查班级是否已创建师生群",
				UseWhen:      []string{"需要批量检查多个班级是否已创建师生群时"},
				AvoidWhen:    []string{"单个班级检查用 student-group exists"},
				Examples: []string{
					"dws edu-group batch check-student-group --class-ids 12345,67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-ids", Property: "input.classIds", Required: boolPtr(true)},
			},
		},
	})

	batchGetClassConversationsCmd := &cobra.Command{
		Use:   "get-class-groups",
		Short: "批量获取班级群信息",
		Long:  `批量获取多个班级的班级群会话信息。返回班级ID与群会话信息的映射关系。仅限管理员角色调用。`,
		Example: `  dws edu-group batch get-class-groups --class-ids 12345,67890
  dws edu-group batch get-class-groups --class-ids 12345,67890 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("class-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--class-ids 为必填参数")
			}
			classIDs := eduGroupParseCSV(raw)
			if len(classIDs) == 0 {
				return fmt.Errorf("--class-ids 不能为空")
			}
			return callMCPToolOnServer("edu-group", "batch_get_class_conversations", map[string]any{
				"input": map[string]any{"classIds": classIDs},
			})
		},
	}
	DeclareLeafMetadata(batchGetClassConversationsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "batch_get_class_conversations",
				CanonicalPath:  "edu-group.batch_get_class_conversations",
				CLIPath:        "edu-group batch get-class-groups",
				PrimaryCLIPath: "edu-group batch get-class-groups",
			},
			Description: "批量获取班级群信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "batch_get_class_conversations"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量获取班级群会话信息",
				UseWhen:      []string{"需要批量获取多个班级的班级群会话信息时"},
				AvoidWhen:    []string{"按会话ID批量查询用 class-group list-by-cids"},
				Examples: []string{
					"dws edu-group batch get-class-groups --class-ids 12345,67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-ids", Property: "input.classIds", Required: boolPtr(true)},
			},
		},
	})

	batchCreateClassGroupCmd := &cobra.Command{
		Use:   "create-student-groups",
		Short: "批量创建师生群",
		Long: `为组织下所有已设置班主任但尚未创建师生群的班级批量创建师生群。
小学和幼儿园学段的班级不会创建师生群。仅限管理员角色调用。`,
		Example: `  dws edu-group batch create-student-groups
  dws edu-group batch create-student-groups -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-group", "batch_create_class_group", map[string]any{
				"input": map[string]any{},
			})
		},
	}
	DeclareLeafMetadata(batchCreateClassGroupCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-group",
				Name:           "batch_create_class_group",
				CanonicalPath:  "edu-group.batch_create_class_group",
				CLIPath:        "edu-group batch create-student-groups",
				PrimaryCLIPath: "edu-group batch create-student-groups",
			},
			Description: "批量创建师生群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-group", RPCName: "batch_create_class_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "为组织下符合条件的班级批量创建师生群",
				UseWhen:      []string{"需要为组织下所有已设置班主任但尚未创建师生群的班级批量创建师生群时"},
				AvoidWhen:    []string{"单个班级创建用 student-group create"},
				Examples: []string{
					"dws edu-group batch create-student-groups --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ════════════════════════════════════════════════════════════
	// flags + 构建命令树
	// ════════════════════════════════════════════════════════════

	// student-group flags
	studentGroupInfoCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupExistsCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupMembersCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupIsInCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupConversationCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupCreateCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	studentGroupDisbandCmd.Flags().String("dept-id", "", "班级 ID（必填）")

	// class-group flags
	classGroupConversationIDCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classGroupConversationCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classGroupExistsCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classGroupListByConversationIDsCmd.Flags().String("conversation-ids", "", "群会话 ID 列表，逗号分隔（必填）")

	// batch flags
	batchCheckClassGroupCmd.Flags().String("class-ids", "", "班级 ID 列表，逗号分隔（必填）")
	batchGetClassConversationsCmd.Flags().String("class-ids", "", "班级 ID 列表，逗号分隔（必填）")

	studentGroupCmd.AddCommand(
		studentGroupInfoCmd, studentGroupExistsCmd, studentGroupMembersCmd,
		studentGroupIsInCmd, studentGroupConversationCmd,
		studentGroupCreateCmd, studentGroupDisbandCmd,
	)
	classGroupCmd.AddCommand(
		classGroupConversationIDCmd, classGroupConversationCmd,
		classGroupExistsCmd, classGroupListByConversationIDsCmd,
	)
	batchCmd.AddCommand(batchCheckClassGroupCmd, batchGetClassConversationsCmd, batchCreateClassGroupCmd)

	root.AddCommand(studentGroupCmd, classGroupCmd, batchCmd)

	return root
}

// eduGroupRequiredIntFlag extracts a required integer flag, returning an error
// if the flag is empty or not a valid integer.
func eduGroupRequiredIntFlag(cmd *cobra.Command, name string) (int64, error) {
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

// eduGroupParseCSV splits a comma-separated string into trimmed non-empty values.
func eduGroupParseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

// eduGroupParseIntCSV splits a comma-separated string into int64 values.
func eduGroupParseIntCSV(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q", v)
		}
		result = append(result, n)
	}
	return result, nil
}
