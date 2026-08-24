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
// dws edu-contact — 家校通讯录扩展
// 共 16 个工具，按 school / class / family / teacher 分组
// ──────────────────────────────────────────────────────────

func newEduContactCommand() *cobra.Command {
	contract.RegisterProductDecl(contract.ProductDecl{
		ID: "edu-contact",
		Selection: contract.ProductSelectionDecl{
			AgentSummary: "家校通讯录：学校组织架构/班级/家庭关系/教师信息查询与管理",
			UseWhen: []string{
				"用户要查询或管理钉钉家校通讯录，包括学校组织、班级、学生、教师、家长等信息。",
			},
			AvoidWhen: []string{
				"家庭群用 edu-familygroup；班级师生群用 edu-group；家校应用/作业/打卡用 edu-app。",
			},
		},
	})
	root := &cobra.Command{
		Use:    "edu-contact",
		Short:  "家校通讯录",
		Long:   `钉钉家校通讯录扩展：查询学校组织/班级/家庭关系/教师信息等。`,
		Hidden: true,
		RunE:   groupRunE,
	}

	// ════════════════════════════════════════════════════════════
	// school 子命令组 — 学校/组织级查询
	// ════════════════════════════════════════════════════════════

	schoolCmd := &cobra.Command{Use: "school", Short: "学校/组织管理", RunE: groupRunE}

	schoolRolesCmd := &cobra.Command{
		Use:   "roles",
		Short: "查询用户在组织内的身份",
		Long: `查询当前用户在指定组织（学校）内拥有的角色身份。
返回用户拥有的角色列表：teacher、headmaster、student、guardian、superAdmin、admin 等。`,
		Example: `  dws edu-contact school roles
  dws edu-contact school roles -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "query_user_roles_in_org", nil)
		},
	}
	DeclareLeafMetadata(schoolRolesCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_user_roles_in_org",
				CanonicalPath:  "edu-contact.query_user_roles_in_org",
				CLIPath:        "edu-contact school roles",
				PrimaryCLIPath: "edu-contact school roles",
			},
			Description: "查询用户在组织内的身份",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_user_roles_in_org"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询用户在组织内的身份角色",
				UseWhen:      []string{"需要查询当前用户在学校组织内拥有哪些角色身份时"},
				AvoidWhen:    []string{"查询班级内角色用 class user-role"},
				Examples: []string{
					"dws edu-contact school roles --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	schoolStructureCmd := &cobra.Command{
		Use:   "structure",
		Short: "查询学校组织架构",
		Long:  `查询指定学校的完整组织架构树。仅限管理员角色调用。`,
		Example: `  dws edu-contact school structure
  dws edu-contact school structure -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "get_school_structure", nil)
		},
	}
	DeclareLeafMetadata(schoolStructureCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_school_structure",
				CanonicalPath:  "edu-contact.get_school_structure",
				CLIPath:        "edu-contact school structure",
				PrimaryCLIPath: "edu-contact school structure",
			},
			Description: "查询学校组织架构",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_school_structure"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学校完整组织架构树",
				UseWhen:      []string{"需要查看学校的完整组织架构时"},
				AvoidWhen:    []string{"查询学校班级列表用 school class-list"},
				Examples: []string{
					"dws edu-contact school structure --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	schoolPeriodsCmd := &cobra.Command{
		Use:   "periods",
		Short: "查询学校学段信息",
		Long:  `查询指定学校配置的所有学段信息，如小学、初中、高中等。所有已认证用户均可调用。`,
		Example: `  dws edu-contact school periods
  dws edu-contact school periods -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "get_school_periods", nil)
		},
	}
	DeclareLeafMetadata(schoolPeriodsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_school_periods",
				CanonicalPath:  "edu-contact.get_school_periods",
				CLIPath:        "edu-contact school periods",
				PrimaryCLIPath: "edu-contact school periods",
			},
			Description: "查询学校学段信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_school_periods"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学校配置的所有学段信息",
				UseWhen:      []string{"需要查询学校有哪些学段（小学/初中/高中等）时"},
				AvoidWhen:    []string{"查询学校类型用 school type"},
				Examples: []string{
					"dws edu-contact school periods --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	schoolTypeCmd := &cobra.Command{
		Use:   "type",
		Short: "查询学校组织类型",
		Long: `查询指定学校的教育组织类型。
返回值：EDU_BASIC（基础教育/K12）、SELF_CLASS（自建班级）、UNIVERSITY（高校）、EDU_TRAINING（教培机构）。`,
		Example: `  dws edu-contact school type
  dws edu-contact school type -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "get_edu_org_type", map[string]any{
				"input": "",
			})
		},
	}
	DeclareLeafMetadata(schoolTypeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_edu_org_type",
				CanonicalPath:  "edu-contact.get_edu_org_type",
				CLIPath:        "edu-contact school type",
				PrimaryCLIPath: "edu-contact school type",
			},
			Description: "查询学校组织类型",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_edu_org_type"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学校的教育组织类型",
				UseWhen:      []string{"需要查询学校属于哪种教育组织类型时"},
				AvoidWhen:    []string{"查询学段信息用 school periods"},
				Examples: []string{
					"dws edu-contact school type --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	schoolStatsCmd := &cobra.Command{
		Use:   "stats",
		Short: "查询学校统计数据",
		Long: `查询指定学校的统计数据。
统计类型 --statistics-type 取值：1-学生、2-家长、4-班级；不指定则返回全部统计。
仅限老师、班主任、管理员角色调用。`,
		Example: `  dws edu-contact school stats
  dws edu-contact school stats --statistics-type 1
  dws edu-contact school stats --statistics-type 4 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if stStr, _ := cmd.Flags().GetString("statistics-type"); strings.TrimSpace(stStr) != "" {
				st, err := strconv.ParseInt(strings.TrimSpace(stStr), 10, 64)
				if err != nil {
					return fmt.Errorf("--statistics-type 须为整数: %w", err)
				}
				input["statisticsType"] = st
			}
			return callMCPToolOnServer("edu-contact", "statistics_school", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(schoolStatsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "statistics_school",
				CanonicalPath:  "edu-contact.statistics_school",
				CLIPath:        "edu-contact school stats",
				PrimaryCLIPath: "edu-contact school stats",
			},
			Description: "查询学校统计数据",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "statistics_school"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学校统计数据（学生/家长/班级数量）",
				UseWhen:      []string{"需要查询学校的学生数、家长数或班级数等统计信息时"},
				AvoidWhen:    []string{"查询班级列表用 school class-list"},
				Examples: []string{
					"dws edu-contact school stats --statistics-type 1 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "statistics-type", Property: "input.statisticsType"},
			},
		},
	})

	schoolClassListCmd := &cobra.Command{
		Use:   "class-list",
		Short: "查询学校所有班级列表",
		Long:  `查询指定学校下的所有班级列表。仅限老师、班主任、管理员角色调用。`,
		Example: `  dws edu-contact school class-list
  dws edu-contact school class-list -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "get_org_class_list", nil)
		},
	}
	DeclareLeafMetadata(schoolClassListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_org_class_list",
				CanonicalPath:  "edu-contact.get_org_class_list",
				CLIPath:        "edu-contact school class-list",
				PrimaryCLIPath: "edu-contact school class-list",
			},
			Description: "查询学校所有班级列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_org_class_list"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学校下的所有班级列表",
				UseWhen:      []string{"需要获取学校下所有班级的列表时"},
				AvoidWhen:    []string{"查询班级详情用 class detail"},
				Examples: []string{
					"dws edu-contact school class-list --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ════════════════════════════════════════════════════════════
	// class 子命令组 — 班级级查询
	// ════════════════════════════════════════════════════════════

	classCmd := &cobra.Command{Use: "class", Short: "班级管理", RunE: groupRunE}

	classDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "查询班级详情",
		Long:  `查询指定班级的详细信息，包括班级名称、所属年级、学段、学生人数、班主任等。`,
		Example: `  dws edu-contact class detail --dept-id 12345
  dws edu-contact class detail --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-contact", "get_class_detail", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_class_detail",
				CanonicalPath:  "edu-contact.get_class_detail",
				CLIPath:        "edu-contact class detail",
				PrimaryCLIPath: "edu-contact class detail",
			},
			Description: "查询班级详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_class_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询指定班级的详细信息",
				UseWhen:      []string{"需要查询某个班级的名称、年级、学段、学生人数、班主任等详情时"},
				AvoidWhen:    []string{"查询班级学生列表用 class students"},
				Examples: []string{
					"dws edu-contact class detail --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classStudentsCmd := &cobra.Command{
		Use:   "students",
		Short: "查询班级学生信息",
		Long:  `查询指定班级内的学生列表及班级详情。`,
		Example: `  dws edu-contact class students --dept-id 12345
  dws edu-contact class students --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-contact", "query_classes_student_info", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classStudentsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_classes_student_info",
				CanonicalPath:  "edu-contact.query_classes_student_info",
				CLIPath:        "edu-contact class students",
				PrimaryCLIPath: "edu-contact class students",
			},
			Description: "查询班级学生信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_classes_student_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询指定班级内的学生列表",
				UseWhen:      []string{"需要查询某个班级的学生列表及班级详情时"},
				AvoidWhen:    []string{"查询班级老师用 class teachers"},
				Examples: []string{
					"dws edu-contact class students --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classTeachersCmd := &cobra.Command{
		Use:   "teachers",
		Short: "查询班级的老师列表",
		Long:  `查询指定班级的所有老师信息，包括教授课程、是否为班主任等。`,
		Example: `  dws edu-contact class teachers --dept-id 12345
  dws edu-contact class teachers --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-contact", "query_class_teachers", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classTeachersCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_class_teachers",
				CanonicalPath:  "edu-contact.query_class_teachers",
				CLIPath:        "edu-contact class teachers",
				PrimaryCLIPath: "edu-contact class teachers",
			},
			Description: "查询班级的老师列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_class_teachers"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询指定班级的所有老师信息",
				UseWhen:      []string{"需要查询某个班级的老师列表（含课程、是否班主任）时"},
				AvoidWhen:    []string{"查询班级学生用 class students"},
				Examples: []string{
					"dws edu-contact class teachers --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classSameNameCmd := &cobra.Command{
		Use:   "same-name",
		Short: "查询班级内同名学生",
		Long:  `查询指定班级内是否存在同名学生。适用于点名、成绩录入等场景。`,
		Example: `  dws edu-contact class same-name --dept-id 12345
  dws edu-contact class same-name --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-contact", "get_class_same_name_students", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classSameNameCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_class_same_name_students",
				CanonicalPath:  "edu-contact.get_class_same_name_students",
				CLIPath:        "edu-contact class same-name",
				PrimaryCLIPath: "edu-contact class same-name",
			},
			Description: "查询班级内同名学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_class_same_name_students"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询班级内是否存在同名学生",
				UseWhen:      []string{"需要查询班级内是否有重名学生时"},
				AvoidWhen:    []string{"查询班级学生列表用 class students"},
				Examples: []string{
					"dws edu-contact class same-name --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classUserRoleCmd := &cobra.Command{
		Use:   "user-role",
		Short: "查询用户在班级内的角色",
		Long:  `查询当前用户在指定班级内的角色身份。`,
		Example: `  dws edu-contact class user-role --dept-id 12345
  dws edu-contact class user-role --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-contact", "get_user_role_in_class", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}
	DeclareLeafMetadata(classUserRoleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "get_user_role_in_class",
				CanonicalPath:  "edu-contact.get_user_role_in_class",
				CLIPath:        "edu-contact class user-role",
				PrimaryCLIPath: "edu-contact class user-role",
			},
			Description: "查询用户在班级内的角色",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "get_user_role_in_class"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询当前用户在指定班级内的角色身份",
				UseWhen:      []string{"需要查询用户在某个班级内的角色时"},
				AvoidWhen:    []string{"查询用户在组织内角色用 school roles"},
				Examples: []string{
					"dws edu-contact class user-role --dept-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	classSearchByNameCmd := &cobra.Command{
		Use:   "search-by-name",
		Short: "根据学生或家长姓名查询班级",
		Long: `根据学生或家长姓名模糊搜索其所在的班级信息。
--query-type 取值：student（学生）、guardian（家长）。仅限管理员角色调用。`,
		Example: `  dws edu-contact class search-by-name --query-type student --name 张三
  dws edu-contact class search-by-name --query-type guardian --name 李四 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			queryType, _ := cmd.Flags().GetString("query-type")
			queryType = strings.TrimSpace(queryType)
			if queryType == "" {
				return fmt.Errorf("--query-type 为必填参数")
			}
			name, _ := cmd.Flags().GetString("name")
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("--name 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "query_class_by_guardian_name", map[string]any{
				"input": map[string]any{
					"queryType": queryType,
					"name":      name,
				},
			})
		},
	}
	DeclareLeafMetadata(classSearchByNameCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_class_by_guardian_name",
				CanonicalPath:  "edu-contact.query_class_by_guardian_name",
				CLIPath:        "edu-contact class search-by-name",
				PrimaryCLIPath: "edu-contact class search-by-name",
			},
			Description: "根据学生或家长姓名查询班级",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_class_by_guardian_name"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "根据学生或家长姓名模糊搜索所在班级",
				UseWhen:      []string{"需要根据学生或家长的姓名查找其所在班级时"},
				AvoidWhen:    []string{"根据老师姓名查班级用 class search-by-teacher"},
				Examples: []string{
					"dws edu-contact class search-by-name --query-type student --name 张三 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "query-type", Property: "input.queryType", Required: boolPtr(true)},
				{Name: "name", Property: "input.name", Required: boolPtr(true)},
			},
		},
	})

	classHeadmasterCmd := &cobra.Command{
		Use:   "headmaster",
		Short: "根据班级名称查询班主任",
		Long:  `根据班级名称查询班主任信息。仅限管理员角色调用。`,
		Example: `  dws edu-contact class headmaster --class-name 一年级1班
  dws edu-contact class headmaster --class-name 三年级2班 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			className, _ := cmd.Flags().GetString("class-name")
			className = strings.TrimSpace(className)
			if className == "" {
				return fmt.Errorf("--class-name 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "query_headmaster_by_className", map[string]any{
				"input": map[string]any{"className": className},
			})
		},
	}
	DeclareLeafMetadata(classHeadmasterCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_headmaster_by_className",
				CanonicalPath:  "edu-contact.query_headmaster_by_className",
				CLIPath:        "edu-contact class headmaster",
				PrimaryCLIPath: "edu-contact class headmaster",
			},
			Description: "根据班级名称查询班主任",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_headmaster_by_className"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "根据班级名称查询班主任信息",
				UseWhen:      []string{"需要通过班级名称查找班主任时"},
				AvoidWhen:    []string{"查询班级所有老师用 class teachers"},
				Examples: []string{
					"dws edu-contact class headmaster --class-name 一年级1班 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-name", Property: "input.className", Required: boolPtr(true)},
			},
		},
	})

	classSearchByTeacherCmd := &cobra.Command{
		Use:   "search-by-teacher",
		Short: "根据老师姓名查询班级",
		Long: `根据教师姓名模糊搜索其任教的班级信息。
通过通讯录模糊匹配教师，返回该教师所任教的班级列表及班级详情（班级ID、格式化班级名称等），
结果已按班级去重。仅限管理员角色调用。`,
		Example: `  dws edu-contact class search-by-teacher --name 张老师
  dws edu-contact class search-by-teacher --name 李 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			name = strings.TrimSpace(name)
			if name == "" {
				return fmt.Errorf("--name 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "query_class_by_teacher_name", map[string]any{
				"input": map[string]any{
					"name": name,
				},
			})
		},
	}
	DeclareLeafMetadata(classSearchByTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_class_by_teacher_name",
				CanonicalPath:  "edu-contact.query_class_by_teacher_name",
				CLIPath:        "edu-contact class search-by-teacher",
				PrimaryCLIPath: "edu-contact class search-by-teacher",
			},
			Description: "根据老师姓名查询班级",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_class_by_teacher_name"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "根据教师姓名模糊搜索其任教的班级",
				UseWhen:      []string{"需要通过教师姓名查找其任教的班级时"},
				AvoidWhen:    []string{"根据学生/家长姓名查班级用 class search-by-name"},
				Examples: []string{
					"dws edu-contact class search-by-teacher --name 张老师 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "name", Property: "input.name", Required: boolPtr(true)},
			},
		},
	})

	classUpdateStudentCmd := &cobra.Command{
		Use:   "update-student",
		Short: "更新学生信息",
		Long: `更新指定班级中某个学生的信息。
可选更新学生姓名、学号及家长信息列表。

通过 --append-patriarch 控制家长更新模式：
  - 完全更新家长信息（默认）：不传 --append-patriarch 或 --append-patriarch=false，并传入完整的 --patriarchs JSON
  - 增量添加新家长：--append-patriarch=true，仅传入需要新增的 --patriarchs JSON
  - 不修改家长信息：--append-patriarch=true 且不传 --patriarchs

--patriarchs JSON 格式示例：
  [{"userId":"uid1","relation":"F","relationName":"父亲","phone":"13800138000"}]
  relation 取值：F-父亲 M-母亲 GF-爷爷 GM-奶奶 O-其他

仅限管理员或班主任角色调用。`,
		Example: `  # 仅更新学生姓名（不修改家长，需传 --append-patriarch 避免清空家长）
  dws edu-contact class update-student --class-id 12345 --student-user-id userId1 --student-name 张三 --append-patriarch
  # 仅更新学号
  dws edu-contact class update-student --class-id 12345 --student-user-id userId1 --student-number S001 --append-patriarch
  # 全量更新家长信息（覆盖原有家长列表，不传 --append-patriarch）
  dws edu-contact class update-student --class-id 12345 --student-user-id userId1 --patriarchs '[{"userId":"uid1","relation":"F","relationName":"父亲","phone":"13800138000"}]'
  # 增量添加新家长（不影响已有家长）
  dws edu-contact class update-student --class-id 12345 --student-user-id userId1 --patriarchs '[{"userId":"uid2","relation":"M","relationName":"母亲","phone":"13900139000"}]' --append-patriarch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, err := eduRequiredIntFlag(cmd, "class-id")
			if err != nil {
				return err
			}
			studentUserID, _ := cmd.Flags().GetString("student-user-id")
			studentUserID = strings.TrimSpace(studentUserID)
			if studentUserID == "" {
				return fmt.Errorf("--student-user-id 为必填参数")
			}

			input := map[string]any{
				"classId":       classID,
				"studentUserId": studentUserID,
			}

			if studentName, _ := cmd.Flags().GetString("student-name"); strings.TrimSpace(studentName) != "" {
				input["studentName"] = strings.TrimSpace(studentName)
			}
			if studentNumber, _ := cmd.Flags().GetString("student-number"); strings.TrimSpace(studentNumber) != "" {
				input["studentNumber"] = strings.TrimSpace(studentNumber)
			}

			appendPatriarch, _ := cmd.Flags().GetBool("append-patriarch")
			input["appendPatriarch"] = appendPatriarch

			patriarchsJSON, _ := cmd.Flags().GetString("patriarchs")
			if strings.TrimSpace(patriarchsJSON) != "" {
				var patriarchs []map[string]any
				if err := json.Unmarshal([]byte(patriarchsJSON), &patriarchs); err != nil {
					return fmt.Errorf("--patriarchs JSON 格式错误: %w", err)
				}
				input["patriarchs"] = patriarchs
			}

			return callMCPToolOnServer("edu-contact", "update_student", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(classUpdateStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "update_student",
				CanonicalPath:  "edu-contact.update_student",
				CLIPath:        "edu-contact class update-student",
				PrimaryCLIPath: "edu-contact class update-student",
			},
			Description: "更新学生信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "update_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "更新指定班级中某个学生的姓名、学号或家长信息",
				UseWhen:      []string{"需要更新学生的姓名、学号或家长信息时"},
				AvoidWhen:    []string{"修改学号用 class update-student-number；修改学生名称/家长关系用 class modify-student-info"},
				Examples: []string{
					"dws edu-contact class update-student --class-id 12345 --student-user-id userId1 --student-name 张三 --append-patriarch --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "student-user-id", Property: "input.studentUserId", Required: boolPtr(true)},
				{Name: "student-name", Property: "input.studentName"},
				{Name: "student-number", Property: "input.studentNumber"},
				{Name: "patriarchs", Property: "input.patriarchs"},
				{Name: "append-patriarch", Property: "input.appendPatriarch"},
			},
		},
	})

	classAddStudentCmd := &cobra.Command{
		Use:   "add-student",
		Short: "添加班级学生",
		Long: `向指定班级添加一名学生。
必须提供学生手机号（--student-mobile）或至少一个家长手机号（通过 --mother/--father/--other-patriarchs 传入）。
可选传入学生userId（已有学生时传入）、学号、虚拟账号ID。
仅限管理员角色调用。

家长信息通过 --mother、--father、--other-patriarchs 传入 JSON 格式：
  {"mobile":"13800138000","relation":"M","relationName":"母亲"}
  relation 取值：F-父亲 M-母亲 GF-爷爷 GM-奶奶 O-其他

--other-patriarchs 为 JSON 数组，可传入多个其他家长。`,
		Example: `  # 添加学生（学生手机号）
  dws edu-contact class add-student --dept-id 12345 --student-name 张三 --student-mobile 13800138000
  # 添加学生并指定母亲信息（家长手机号）
  dws edu-contact class add-student --dept-id 12345 --student-name 张三 --mother '{"mobile":"13800138000","relation":"M","relationName":"母亲"}'
  # 添加学生并指定学号和父亲信息
  dws edu-contact class add-student --dept-id 12345 --student-name 张三 --student-number S001 --father '{"mobile":"13900139000","relation":"F","relationName":"父亲"}'
  # 添加已有学生到班级
  dws edu-contact class add-student --dept-id 12345 --student-name 张三 --student-user-id existingUserId --student-mobile 13800138000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			studentName, _ := cmd.Flags().GetString("student-name")
			studentName = strings.TrimSpace(studentName)
			if studentName == "" {
				return fmt.Errorf("--student-name 为必填参数")
			}

			input := map[string]any{
				"deptId":      deptID,
				"studentName": studentName,
			}

			if studentUserID, _ := cmd.Flags().GetString("student-user-id"); strings.TrimSpace(studentUserID) != "" {
				input["studentUserId"] = strings.TrimSpace(studentUserID)
			}
			if studentNumber, _ := cmd.Flags().GetString("student-number"); strings.TrimSpace(studentNumber) != "" {
				input["studentNumber"] = strings.TrimSpace(studentNumber)
			}
			if virtualAccountID, _ := cmd.Flags().GetString("virtual-account-id"); strings.TrimSpace(virtualAccountID) != "" {
				input["virtualAccountId"] = strings.TrimSpace(virtualAccountID)
			}

			studentMobile, _ := cmd.Flags().GetString("student-mobile")
			studentMobile = strings.TrimSpace(studentMobile)
			if studentMobile != "" {
				input["studentMobile"] = studentMobile
			}

			hasMobile := studentMobile != ""

			if motherJSON, _ := cmd.Flags().GetString("mother"); strings.TrimSpace(motherJSON) != "" {
				var mother map[string]any
				if err := json.Unmarshal([]byte(motherJSON), &mother); err != nil {
					return fmt.Errorf("--mother JSON 格式错误: %w", err)
				}
				input["mother"] = mother
				hasMobile = true
			}
			if fatherJSON, _ := cmd.Flags().GetString("father"); strings.TrimSpace(fatherJSON) != "" {
				var father map[string]any
				if err := json.Unmarshal([]byte(fatherJSON), &father); err != nil {
					return fmt.Errorf("--father JSON 格式错误: %w", err)
				}
				input["father"] = father
				hasMobile = true
			}
			if otherJSON, _ := cmd.Flags().GetString("other-patriarchs"); strings.TrimSpace(otherJSON) != "" {
				var others []map[string]any
				if err := json.Unmarshal([]byte(otherJSON), &others); err != nil {
					return fmt.Errorf("--other-patriarchs JSON 格式错误: %w", err)
				}
				input["otherPatriarchs"] = others
				hasMobile = true
			}

			if !hasMobile {
				return fmt.Errorf("必须提供学生手机号（--student-mobile）或至少一个家长手机号（通过 --mother/--father/--other-patriarchs 传入）")
			}

			return callMCPToolOnServer("edu-contact", "add_student", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(classAddStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "add_student",
				CanonicalPath:  "edu-contact.add_student",
				CLIPath:        "edu-contact class add-student",
				PrimaryCLIPath: "edu-contact class add-student",
			},
			Description: "添加班级学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "add_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "向指定班级添加一名学生",
				UseWhen:      []string{"需要向班级添加新学生时"},
				AvoidWhen:    []string{"添加非行政班学生用 class add-unofficial-student"},
				Examples: []string{
					"dws edu-contact class add-student --dept-id 12345 --student-name 张三 --student-mobile 13800138000 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "student-name", Property: "input.studentName", Required: boolPtr(true)},
				{Name: "student-mobile", Property: "input.studentMobile"},
				{Name: "student-user-id", Property: "input.studentUserId"},
				{Name: "student-number", Property: "input.studentNumber"},
				{Name: "virtual-account-id", Property: "input.virtualAccountId"},
				{Name: "mother", Property: "input.mother"},
				{Name: "father", Property: "input.father"},
				{Name: "other-patriarchs", Property: "input.otherPatriarchs"},
			},
		},
	})

	classModifyStudentInfoCmd := &cobra.Command{
		Use:   "modify-student-info",
		Short: "修改学生信息",
		Long: `修改指定班级中某个学生的名称或家长关系信息。
可选修改学生名称和家长关系，两者至少填一项。
修改家长关系时需同时传入家长 userId 和关系类型。
操作有并发控制，同一用户短时间内不可重复操作。仅限班主任角色调用。

--relation 取值：父亲、母亲、爷爷、奶奶、哥哥、姐姐、其他`,
		Example: `  # 修改学生名称
  dws edu-contact class modify-student-info --dept-id 12345 --target-user-id userId1 --nick 张三
  # 修改家长关系
  dws edu-contact class modify-student-info --dept-id 12345 --target-user-id userId1 --patriarch-user-id parentId1 --relation 父亲
  # 同时修改名称和家长关系
  dws edu-contact class modify-student-info --dept-id 12345 --target-user-id userId1 --nick 张三 --patriarch-user-id parentId1 --relation 母亲`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			targetUserID, _ := cmd.Flags().GetString("target-user-id")
			targetUserID = strings.TrimSpace(targetUserID)
			if targetUserID == "" {
				return fmt.Errorf("--target-user-id 为必填参数")
			}

			input := map[string]any{
				"deptId":       deptID,
				"targetUserId": targetUserID,
			}

			nick, _ := cmd.Flags().GetString("nick")
			nick = strings.TrimSpace(nick)
			patriarchUserID, _ := cmd.Flags().GetString("patriarch-user-id")
			patriarchUserID = strings.TrimSpace(patriarchUserID)
			relation, _ := cmd.Flags().GetString("relation")
			relation = strings.TrimSpace(relation)

			if nick == "" && patriarchUserID == "" {
				return fmt.Errorf("--nick 和 --patriarch-user-id/--relation 至少填一项")
			}
			if nick != "" {
				input["nick"] = nick
			}
			if patriarchUserID != "" {
				if relation == "" {
					return fmt.Errorf("修改家长关系时需同时传入 --relation")
				}
				input["patriarchUserId"] = patriarchUserID
				input["relation"] = relation
			}

			return callMCPToolOnServer("edu-contact", "modify_student_info", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(classModifyStudentInfoCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "modify_student_info",
				CanonicalPath:  "edu-contact.modify_student_info",
				CLIPath:        "edu-contact class modify-student-info",
				PrimaryCLIPath: "edu-contact class modify-student-info",
			},
			Description: "修改学生信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "modify_student_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "修改学生名称或家长关系信息",
				UseWhen:      []string{"需要修改学生名称或调整家长关系时"},
				AvoidWhen:    []string{"更新学生姓名/学号/家长列表用 class update-student"},
				Examples: []string{
					"dws edu-contact class modify-student-info --dept-id 12345 --target-user-id userId1 --nick 张三 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "target-user-id", Property: "input.targetUserId", Required: boolPtr(true)},
				{Name: "nick", Property: "input.nick"},
				{Name: "patriarch-user-id", Property: "input.patriarchUserId"},
				{Name: "relation", Property: "input.relation"},
			},
		},
	})

	classDeleteTeacherCmd := &cobra.Command{
		Use:   "delete-teacher",
		Short: "删除班级教师",
		Long: `从指定班级中删除一名教师。
返回操作是否成功。仅限管理员或班主任角色调用。`,
		Example: `  dws edu-contact class delete-teacher --class-id 12345 --teacher-user-id userId1
  dws edu-contact class delete-teacher --class-id 12345 --teacher-user-id userId1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, err := eduRequiredIntFlag(cmd, "class-id")
			if err != nil {
				return err
			}
			teacherUserID, _ := cmd.Flags().GetString("teacher-user-id")
			teacherUserID = strings.TrimSpace(teacherUserID)
			if teacherUserID == "" {
				return fmt.Errorf("--teacher-user-id 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "delete_teacher", map[string]any{
				"input": map[string]any{
					"classId":       classID,
					"teacherUserId": teacherUserID,
				},
			})
		},
	}
	DeclareLeafMetadata(classDeleteTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "delete_teacher",
				CanonicalPath:  "edu-contact.delete_teacher",
				CLIPath:        "edu-contact class delete-teacher",
				PrimaryCLIPath: "edu-contact class delete-teacher",
			},
			Description: "删除班级教师",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "delete_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "从指定班级中删除一名教师",
				UseWhen:      []string{"需要将某位教师从班级中移除时"},
				AvoidWhen:    []string{"添加教师用 class add-teachers"},
				Examples: []string{
					"dws edu-contact class delete-teacher --class-id 12345 --teacher-user-id userId1 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "teacher-user-id", Property: "input.teacherUserId", Required: boolPtr(true)},
			},
		},
	})

	classUpdateClassInfoCmd := &cobra.Command{
		Use:   "update-info",
		Short: "更新班级信息",
		Long: `更新指定班级的基本信息。
可选更新班级别名、预期学生人数及班级群名称。
班级别名和群名称有长度限制且需通过内容安全校验。
修改群名称时需同时传入群会话ID（--conversation-id）。
仅限管理员或班主任角色调用。`,
		Example: `  # 更新班级别名
  dws edu-contact class update-info --class-id 12345 --nick 火箭班
  # 更新预期学生人数
  dws edu-contact class update-info --class-id 12345 --expected-student-num 45
  # 更新班级群名称（需同时传 conversation-id）
  dws edu-contact class update-info --class-id 12345 --group-name 一年级1班家长群 --conversation-id cid123456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, err := eduRequiredIntFlag(cmd, "class-id")
			if err != nil {
				return err
			}

			input := map[string]any{
				"classId": classID,
			}

			if nick, _ := cmd.Flags().GetString("nick"); strings.TrimSpace(nick) != "" {
				input["nick"] = strings.TrimSpace(nick)
			}
			if expectedStr, _ := cmd.Flags().GetString("expected-student-num"); strings.TrimSpace(expectedStr) != "" {
				expectedNum, err := strconv.ParseInt(strings.TrimSpace(expectedStr), 10, 64)
				if err != nil {
					return fmt.Errorf("--expected-student-num 须为整数: %w", err)
				}
				input["expectedStudentNum"] = expectedNum
			}

			groupName, _ := cmd.Flags().GetString("group-name")
			conversationID, _ := cmd.Flags().GetString("conversation-id")
			groupName = strings.TrimSpace(groupName)
			conversationID = strings.TrimSpace(conversationID)
			if groupName != "" {
				if conversationID == "" {
					return fmt.Errorf("修改群名称时需同时传入 --conversation-id")
				}
				input["groupName"] = groupName
				input["conversationId"] = conversationID
			}

			return callMCPToolOnServer("edu-contact", "update_class_info", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(classUpdateClassInfoCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "update_class_info",
				CanonicalPath:  "edu-contact.update_class_info",
				CLIPath:        "edu-contact class update-info",
				PrimaryCLIPath: "edu-contact class update-info",
			},
			Description: "更新班级信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "update_class_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "更新班级别名、预期学生人数或班级群名称",
				UseWhen:      []string{"需要修改班级别名、预期人数或群名称时"},
				AvoidWhen:    []string{"更新学生信息用 class update-student"},
				Examples: []string{
					"dws edu-contact class update-info --class-id 12345 --nick 火箭班 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "nick", Property: "input.nick"},
				{Name: "expected-student-num", Property: "input.expectedStudentNum"},
				{Name: "group-name", Property: "input.groupName"},
				{Name: "conversation-id", Property: "input.conversationId"},
			},
		},
	})

	classUpdateStudentNumberCmd := &cobra.Command{
		Use:   "update-student-number",
		Short: "修改学生学号",
		Long: `修改指定班级中某个学生的学号。
学号有长度限制且需通过内容安全校验。仅限班主任或教师角色调用。`,
		Example: `  dws edu-contact class update-student-number --class-id 12345 --student-user-id userId1 --student-number S001
  dws edu-contact class update-student-number --class-id 12345 --student-user-id userId1 --student-number S002 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, err := eduRequiredIntFlag(cmd, "class-id")
			if err != nil {
				return err
			}
			studentUserID, _ := cmd.Flags().GetString("student-user-id")
			studentUserID = strings.TrimSpace(studentUserID)
			if studentUserID == "" {
				return fmt.Errorf("--student-user-id 为必填参数")
			}
			studentNumber, _ := cmd.Flags().GetString("student-number")
			studentNumber = strings.TrimSpace(studentNumber)
			if studentNumber == "" {
				return fmt.Errorf("--student-number 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "update_student_number", map[string]any{
				"input": map[string]any{
					"classId":       classID,
					"studentUserId": studentUserID,
					"studentNumber": studentNumber,
				},
			})
		},
	}
	DeclareLeafMetadata(classUpdateStudentNumberCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "update_student_number",
				CanonicalPath:  "edu-contact.update_student_number",
				CLIPath:        "edu-contact class update-student-number",
				PrimaryCLIPath: "edu-contact class update-student-number",
			},
			Description: "修改学生学号",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "update_student_number"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "修改指定班级中某个学生的学号",
				UseWhen:      []string{"需要修改学生学号时"},
				AvoidWhen:    []string{"更新学生姓名/家长用 class update-student"},
				Examples: []string{
					"dws edu-contact class update-student-number --class-id 12345 --student-user-id userId1 --student-number S001 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "student-user-id", Property: "input.studentUserId", Required: boolPtr(true)},
				{Name: "student-number", Property: "input.studentNumber", Required: boolPtr(true)},
			},
		},
	})

	classAddUnofficialStudentCmd := &cobra.Command{
		Use:   "add-unofficial-student",
		Short: "添加非行政班学生",
		Long: `向指定非行政班级批量添加学生。
支持批量操作，学生数量不可超过上限。仅限管理员角色调用。`,
		Example: `  dws edu-contact class add-unofficial-student --dept-id 12345 --student-staff-ids staffId1,staffId2
  dws edu-contact class add-unofficial-student --dept-id 12345 --student-staff-ids staffId1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			staffIdsStr, _ := cmd.Flags().GetString("student-staff-ids")
			staffIdsStr = strings.TrimSpace(staffIdsStr)
			if staffIdsStr == "" {
				return fmt.Errorf("--student-staff-ids 为必填参数")
			}
			var studentStaffIds []string
			for _, sid := range strings.Split(staffIdsStr, ",") {
				if s := strings.TrimSpace(sid); s != "" {
					studentStaffIds = append(studentStaffIds, s)
				}
			}
			if len(studentStaffIds) == 0 {
				return fmt.Errorf("--student-staff-ids 不能为空")
			}
			return callMCPToolOnServer("edu-contact", "add_unofficial_class_student", map[string]any{
				"input": map[string]any{
					"deptId":          deptID,
					"studentStaffIds": studentStaffIds,
				},
			})
		},
	}
	DeclareLeafMetadata(classAddUnofficialStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "add_unofficial_class_student",
				CanonicalPath:  "edu-contact.add_unofficial_class_student",
				CLIPath:        "edu-contact class add-unofficial-student",
				PrimaryCLIPath: "edu-contact class add-unofficial-student",
			},
			Description: "添加非行政班学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "add_unofficial_class_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "向非行政班级批量添加学生",
				UseWhen:      []string{"需要向非行政班级批量添加学生时"},
				AvoidWhen:    []string{"添加行政班学生用 class add-student"},
				Examples: []string{
					"dws edu-contact class add-unofficial-student --dept-id 12345 --student-staff-ids staffId1,staffId2 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "student-staff-ids", Property: "input.studentStaffIds", Required: boolPtr(true)},
			},
		},
	})

	classDeleteStudentsCmd := &cobra.Command{
		Use:   "delete-students",
		Short: "批量删除学生",
		Long: `从指定班级中批量删除学生。
返回操作是否成功。仅限管理员角色调用。`,
		Example: `  dws edu-contact class delete-students --dept-id 12345 --student-user-ids userId1,userId2
  dws edu-contact class delete-students --dept-id 12345 --student-user-ids userId1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			studentIdsStr, _ := cmd.Flags().GetString("student-user-ids")
			studentIdsStr = strings.TrimSpace(studentIdsStr)
			if studentIdsStr == "" {
				return fmt.Errorf("--student-user-ids 为必填参数")
			}
			var studentUserIds []string
			for _, uid := range strings.Split(studentIdsStr, ",") {
				if u := strings.TrimSpace(uid); u != "" {
					studentUserIds = append(studentUserIds, u)
				}
			}
			if len(studentUserIds) == 0 {
				return fmt.Errorf("--student-user-ids 不能为空")
			}
			return callMCPToolOnServer("edu-contact", "delete_students", map[string]any{
				"input": map[string]any{
					"deptId":         deptID,
					"studentUserIds": studentUserIds,
				},
			})
		},
	}
	DeclareLeafMetadata(classDeleteStudentsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "delete_students",
				CanonicalPath:  "edu-contact.delete_students",
				CLIPath:        "edu-contact class delete-students",
				PrimaryCLIPath: "edu-contact class delete-students",
			},
			Description: "批量删除学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "delete_students"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "从指定班级中批量删除学生",
				UseWhen:      []string{"需要从班级中批量移除学生时"},
				AvoidWhen:    []string{"学生移班用 class move-student"},
				Examples: []string{
					"dws edu-contact class delete-students --dept-id 12345 --student-user-ids userId1,userId2 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "student-user-ids", Property: "input.studentUserIds", Required: boolPtr(true)},
			},
		},
	})

	classUpdateStudentMobileCmd := &cobra.Command{
		Use:   "update-student-mobile",
		Short: "修改学生手机号",
		Long: `修改指定班级中某个学生的手机号。
手机号会进行格式校验。仅限管理员角色调用。`,
		Example: `  dws edu-contact class update-student-mobile --dept-id 12345 --student-user-id userId1 --mobile 13800138000
  dws edu-contact class update-student-mobile --dept-id 12345 --student-user-id userId1 --mobile 13800138000 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			studentUserID, _ := cmd.Flags().GetString("student-user-id")
			studentUserID = strings.TrimSpace(studentUserID)
			if studentUserID == "" {
				return fmt.Errorf("--student-user-id 为必填参数")
			}
			mobile, _ := cmd.Flags().GetString("mobile")
			mobile = strings.TrimSpace(mobile)
			if mobile == "" {
				return fmt.Errorf("--mobile 为必填参数")
			}
			return callMCPToolOnServer("edu-contact", "update_student_mobile", map[string]any{
				"input": map[string]any{
					"deptId":        deptID,
					"studentUserId": studentUserID,
					"mobile":        mobile,
				},
			})
		},
	}
	DeclareLeafMetadata(classUpdateStudentMobileCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "update_student_mobile",
				CanonicalPath:  "edu-contact.update_student_mobile",
				CLIPath:        "edu-contact class update-student-mobile",
				PrimaryCLIPath: "edu-contact class update-student-mobile",
			},
			Description: "修改学生手机号",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "update_student_mobile"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "修改指定班级中某个学生的手机号",
				UseWhen:      []string{"需要修改学生手机号时"},
				AvoidWhen:    []string{"更新学生姓名/学号用 class update-student"},
				Examples: []string{
					"dws edu-contact class update-student-mobile --dept-id 12345 --student-user-id userId1 --mobile 13800138000 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "student-user-id", Property: "input.studentUserId", Required: boolPtr(true)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(true)},
			},
		},
	})

	classMoveStudentCmd := &cobra.Command{
		Use:   "move-student",
		Short: "学生移班",
		Long: `将学生从一个班级转移到另一个班级。
支持批量操作，学生数量不可超过上限。仅限管理员或班主任角色调用。`,
		Example: `  dws edu-contact class move-student --student-user-ids userId1,userId2 --origin-class-id 12345 --target-class-id 67890
  dws edu-contact class move-student --student-user-ids userId1 --origin-class-id 12345 --target-class-id 67890 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			originClassID, err := eduRequiredIntFlag(cmd, "origin-class-id")
			if err != nil {
				return err
			}
			targetClassID, err := eduRequiredIntFlag(cmd, "target-class-id")
			if err != nil {
				return err
			}
			studentIdsStr, _ := cmd.Flags().GetString("student-user-ids")
			studentIdsStr = strings.TrimSpace(studentIdsStr)
			if studentIdsStr == "" {
				return fmt.Errorf("--student-user-ids 为必填参数")
			}
			var studentUserIds []string
			for _, uid := range strings.Split(studentIdsStr, ",") {
				if u := strings.TrimSpace(uid); u != "" {
					studentUserIds = append(studentUserIds, u)
				}
			}
			if len(studentUserIds) == 0 {
				return fmt.Errorf("--student-user-ids 不能为空")
			}
			return callMCPToolOnServer("edu-contact", "move_student", map[string]any{
				"input": map[string]any{
					"studentUserIds": studentUserIds,
					"originClassId":  originClassID,
					"targetClassId":  targetClassID,
				},
			})
		},
	}
	DeclareLeafMetadata(classMoveStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "move_student",
				CanonicalPath:  "edu-contact.move_student",
				CLIPath:        "edu-contact class move-student",
				PrimaryCLIPath: "edu-contact class move-student",
			},
			Description: "学生移班",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "move_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "将学生从一个班级转移到另一个班级",
				UseWhen:      []string{"需要将学生从原班级转移到目标班级时"},
				AvoidWhen:    []string{"删除学生用 class delete-students"},
				Examples: []string{
					"dws edu-contact class move-student --student-user-ids userId1 --origin-class-id 12345 --target-class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "student-user-ids", Property: "input.studentUserIds", Required: boolPtr(true)},
				{Name: "origin-class-id", Property: "input.originClassId", Required: boolPtr(true)},
				{Name: "target-class-id", Property: "input.targetClassId", Required: boolPtr(true)},
			},
		},
	})

	classAddTeachersCmd := &cobra.Command{
		Use:   "add-teachers",
		Short: "批量添加班级教师",
		Long: `批量将教师添加到指定班级。
教师列表会自动去重，单次最多添加50名教师。仅限管理员角色调用。

--is-adviser 取值：
  0 = 普通老师（默认）
  1 = 班主任`,
		Example: `  dws edu-contact class add-teachers --dept-id 12345 --teacher-user-ids userId1,userId2
  dws edu-contact class add-teachers --dept-id 12345 --teacher-user-ids userId1,userId2 --is-adviser 1
  dws edu-contact class add-teachers --dept-id 12345 --teacher-user-ids userId1 --is-adviser 0 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptID, err := eduRequiredIntFlag(cmd, "dept-id")
			if err != nil {
				return err
			}
			teacherIdsStr, _ := cmd.Flags().GetString("teacher-user-ids")
			teacherIdsStr = strings.TrimSpace(teacherIdsStr)
			if teacherIdsStr == "" {
				return fmt.Errorf("--teacher-user-ids 为必填参数")
			}
			var teacherUserIds []string
			for _, uid := range strings.Split(teacherIdsStr, ",") {
				if u := strings.TrimSpace(uid); u != "" {
					teacherUserIds = append(teacherUserIds, u)
				}
			}
			if len(teacherUserIds) == 0 {
				return fmt.Errorf("--teacher-user-ids 不能为空")
			}
			if len(teacherUserIds) > 50 {
				return fmt.Errorf("单次最多添加50名教师，当前 %d 名", len(teacherUserIds))
			}
			isAdviserStr, _ := cmd.Flags().GetString("is-adviser")
			isAdviser := int64(0)
			if strings.TrimSpace(isAdviserStr) != "" {
				isAdviser, err = strconv.ParseInt(strings.TrimSpace(isAdviserStr), 10, 64)
				if err != nil || (isAdviser != 0 && isAdviser != 1) {
					return fmt.Errorf("--is-adviser 须为 0（普通老师）或 1（班主任）")
				}
			}
			return callMCPToolOnServer("edu-contact", "batch_add_class_teacher", map[string]any{
				"input": map[string]any{
					"deptId":         deptID,
					"teacherUserIds": teacherUserIds,
					"isAdviser":      isAdviser,
				},
			})
		},
	}
	DeclareLeafMetadata(classAddTeachersCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "batch_add_class_teacher",
				CanonicalPath:  "edu-contact.batch_add_class_teacher",
				CLIPath:        "edu-contact class add-teachers",
				PrimaryCLIPath: "edu-contact class add-teachers",
			},
			Description: "批量添加班级教师",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "batch_add_class_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量将教师添加到指定班级",
				UseWhen:      []string{"需要向班级批量添加教师时"},
				AvoidWhen:    []string{"删除教师用 class delete-teacher"},
				Examples: []string{
					"dws edu-contact class add-teachers --dept-id 12345 --teacher-user-ids userId1,userId2 --is-adviser 1 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "teacher-user-ids", Property: "input.teacherUserIds", Required: boolPtr(true)},
				{Name: "is-adviser", Property: "input.isAdviser"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// family 子命令组 — 家庭关系查询
	// ════════════════════════════════════════════════════════════

	familyCmd := &cobra.Command{Use: "family", Short: "家庭关系查询", RunE: groupRunE}

	familyChildrenCmd := &cobra.Command{
		Use:   "children",
		Short: "查询家长的孩子信息",
		Long:  `查询当前家长在指定组织（学校）内关联的所有孩子信息。仅限家长角色调用。`,
		Example: `  dws edu-contact family children
  dws edu-contact family children -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "query_children_in_org", nil)
		},
	}
	DeclareLeafMetadata(familyChildrenCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_children_in_org",
				CanonicalPath:  "edu-contact.query_children_in_org",
				CLIPath:        "edu-contact family children",
				PrimaryCLIPath: "edu-contact family children",
			},
			Description: "查询家长的孩子信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_children_in_org"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询当前家长关联的所有孩子信息",
				UseWhen:      []string{"需要查询家长在学校内关联的孩子信息时"},
				AvoidWhen:    []string{"查询学生的家长信息用 family parents"},
				Examples: []string{
					"dws edu-contact family children --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	familyParentsCmd := &cobra.Command{
		Use:   "parents",
		Short: "查询学生的家长信息",
		Long:  `查询当前学生在指定组织（学校）内关联的所有家长信息。`,
		Example: `  dws edu-contact family parents
  dws edu-contact family parents -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "query_patriarch_in_org", nil)
		},
	}
	DeclareLeafMetadata(familyParentsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_patriarch_in_org",
				CanonicalPath:  "edu-contact.query_patriarch_in_org",
				CLIPath:        "edu-contact family parents",
				PrimaryCLIPath: "edu-contact family parents",
			},
			Description: "查询学生的家长信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_patriarch_in_org"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询当前学生关联的所有家长信息",
				UseWhen:      []string{"需要查询学生在学校内关联的家长信息时"},
				AvoidWhen:    []string{"查询家长的孩子信息用 family children"},
				Examples: []string{
					"dws edu-contact family parents --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ════════════════════════════════════════════════════════════
	// teacher 子命令组 — 教师查询
	// ════════════════════════════════════════════════════════════

	teacherCmd := &cobra.Command{Use: "teacher", Short: "教师管理", RunE: groupRunE}

	teacherClassesCmd := &cobra.Command{
		Use:   "classes",
		Short: "查询老师管理的班级列表",
		Long:  `查询当前老师或班主任在指定组织（学校）内管理的班级列表。`,
		Example: `  dws edu-contact teacher classes
  dws edu-contact teacher classes -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("edu-contact", "query_teacher_classes", nil)
		},
	}
	DeclareLeafMetadata(teacherClassesCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "query_teacher_classes",
				CanonicalPath:  "edu-contact.query_teacher_classes",
				CLIPath:        "edu-contact teacher classes",
				PrimaryCLIPath: "edu-contact teacher classes",
			},
			Description: "查询老师管理的班级列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "query_teacher_classes"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询当前老师管理的班级列表",
				UseWhen:      []string{"需要查询老师或班主任管理的班级列表时"},
				AvoidWhen:    []string{"查询班级老师列表用 class teachers"},
				Examples: []string{
					"dws edu-contact teacher classes --format json",
				},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	teacherUpdateCourseCmd := &cobra.Command{
		Use:   "update-course",
		Short: "更新教师任教科目",
		Long: `批量更新教师在各班级的任教科目信息。
返回操作是否成功。仅限班主任或教师角色调用。

--teacher-class-infos JSON 格式示例：
  [{"classId":12345,"courseCode":"c1","courseName":"语文"}]

每项包含：
  - classId: 班级ID
  - courseCode: 科目编码
  - courseName: 科目名称`,
		Example: `  dws edu-contact teacher update-course --teacher-class-infos '[{"classId":12345,"courseCode":"c1","courseName":"语文"}]'
  dws edu-contact teacher update-course --teacher-class-infos '[{"classId":12345,"courseCode":"c1","courseName":"语文"},{"classId":67890,"courseCode":"c2","courseName":"数学"}]' -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			infosJSON, _ := cmd.Flags().GetString("teacher-class-infos")
			infosJSON = strings.TrimSpace(infosJSON)
			if infosJSON == "" {
				return fmt.Errorf("--teacher-class-infos 为必填参数")
			}
			var teacherClassInfos []map[string]any
			if err := json.Unmarshal([]byte(infosJSON), &teacherClassInfos); err != nil {
				return fmt.Errorf("--teacher-class-infos JSON 格式错误: %w", err)
			}
			if len(teacherClassInfos) == 0 {
				return fmt.Errorf("--teacher-class-infos 不能为空数组")
			}
			return callMCPToolOnServer("edu-contact", "update_teacher_course", map[string]any{
				"input": map[string]any{
					"teacherClassInfos": teacherClassInfos,
				},
			})
		},
	}
	DeclareLeafMetadata(teacherUpdateCourseCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-contact",
				Name:           "update_teacher_course",
				CanonicalPath:  "edu-contact.update_teacher_course",
				CLIPath:        "edu-contact teacher update-course",
				PrimaryCLIPath: "edu-contact teacher update-course",
			},
			Description: "更新教师任教科目",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-contact", RPCName: "update_teacher_course"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量更新教师在各班级的任教科目信息",
				UseWhen:      []string{"需要更新教师任教的科目信息时"},
				AvoidWhen:    []string{"添加教师到班级用 class add-teachers"},
				Examples: []string{
					"dws edu-contact teacher update-course --teacher-class-infos '[{\"classId\":12345,\"courseCode\":\"c1\",\"courseName\":\"语文\"}]' --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "teacher-class-infos", Property: "input.teacherClassInfos", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// flags + 构建命令树
	// ════════════════════════════════════════════════════════════

	// school flags
	schoolStatsCmd.Flags().String("statistics-type", "", "统计类型：1-学生 2-家长 4-班级")

	// class flags
	classDetailCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classStudentsCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classTeachersCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classSameNameCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classUserRoleCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classSearchByNameCmd.Flags().String("query-type", "", "查询类型（必填）：student / guardian")
	classSearchByNameCmd.Flags().String("name", "", "姓名（必填）")
	classHeadmasterCmd.Flags().String("class-name", "", "班级名称（必填）")
	classSearchByTeacherCmd.Flags().String("name", "", "教师姓名（必填）")
	classUpdateStudentCmd.Flags().String("class-id", "", "班级 ID（必填）")
	classUpdateStudentCmd.Flags().String("student-user-id", "", "学生 userId（必填）")
	classUpdateStudentCmd.Flags().String("student-name", "", "学生姓名（可选）")
	classUpdateStudentCmd.Flags().String("student-number", "", "学号（可选）")
	classUpdateStudentCmd.Flags().String("patriarchs", "", "家长信息 JSON 数组（可选）")
	classUpdateStudentCmd.Flags().Bool("append-patriarch", false, "家长更新模式：传入则增量添加，不传则全量更新（默认全量更新）")
	classAddStudentCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classAddStudentCmd.Flags().String("student-name", "", "学生姓名（必填）")
	classAddStudentCmd.Flags().String("student-mobile", "", "学生手机号（与家长手机号至少填一个）")
	classAddStudentCmd.Flags().String("student-user-id", "", "学生 userId（可选，已有学生时传入）")
	classAddStudentCmd.Flags().String("student-number", "", "学号（可选）")
	classAddStudentCmd.Flags().String("virtual-account-id", "", "虚拟账号 ID（可选）")
	classAddStudentCmd.Flags().String("mother", "", "母亲信息 JSON（可选）")
	classAddStudentCmd.Flags().String("father", "", "父亲信息 JSON（可选）")
	classAddStudentCmd.Flags().String("other-patriarchs", "", "其他家长信息 JSON 数组（可选）")
	classModifyStudentInfoCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classModifyStudentInfoCmd.Flags().String("target-user-id", "", "学生 userId（必填）")
	classModifyStudentInfoCmd.Flags().String("nick", "", "学生名称（可选）")
	classModifyStudentInfoCmd.Flags().String("patriarch-user-id", "", "家长 userId（修改家长关系时必填）")
	classModifyStudentInfoCmd.Flags().String("relation", "", "家长关系：父亲/母亲/爷爷/奶奶/哥哥/姐姐/其他（修改家长关系时必填）")
	classDeleteTeacherCmd.Flags().String("class-id", "", "班级 ID（必填）")
	classDeleteTeacherCmd.Flags().String("teacher-user-id", "", "教师 userId（必填）")
	classUpdateClassInfoCmd.Flags().String("class-id", "", "班级 ID（必填）")
	classUpdateClassInfoCmd.Flags().String("nick", "", "班级别名（可选）")
	classUpdateClassInfoCmd.Flags().String("expected-student-num", "", "预期学生人数（可选）")
	classUpdateClassInfoCmd.Flags().String("group-name", "", "班级群名称（可选，需同时传 --conversation-id）")
	classUpdateClassInfoCmd.Flags().String("conversation-id", "", "群会话 ID（修改群名称时必填）")
	classUpdateStudentNumberCmd.Flags().String("class-id", "", "班级 ID（必填）")
	classUpdateStudentNumberCmd.Flags().String("student-user-id", "", "学生 userId（必填）")
	classUpdateStudentNumberCmd.Flags().String("student-number", "", "新学号（必填）")
	classAddUnofficialStudentCmd.Flags().String("dept-id", "", "非行政班级 ID（必填）")
	classAddUnofficialStudentCmd.Flags().String("student-staff-ids", "", "学生 staffId 列表，逗号分隔（必填）")
	classDeleteStudentsCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classDeleteStudentsCmd.Flags().String("student-user-ids", "", "学生 userId 列表，逗号分隔（必填）")
	classUpdateStudentMobileCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classUpdateStudentMobileCmd.Flags().String("student-user-id", "", "学生 userId（必填）")
	classUpdateStudentMobileCmd.Flags().String("mobile", "", "修改后的手机号（必填）")
	classMoveStudentCmd.Flags().String("student-user-ids", "", "学生 userId 列表，逗号分隔（必填）")
	classMoveStudentCmd.Flags().String("origin-class-id", "", "原班级 ID（必填）")
	classMoveStudentCmd.Flags().String("target-class-id", "", "目标班级 ID（必填）")
	classAddTeachersCmd.Flags().String("dept-id", "", "班级 ID（必填）")
	classAddTeachersCmd.Flags().String("teacher-user-ids", "", "教师 userId 列表，逗号分隔（必填，最多50个）")
	classAddTeachersCmd.Flags().String("is-adviser", "0", "教师角色：0-普通老师 1-班主任（默认 0）")

	schoolCmd.AddCommand(schoolRolesCmd, schoolStructureCmd, schoolPeriodsCmd, schoolTypeCmd, schoolStatsCmd, schoolClassListCmd)
	classCmd.AddCommand(classDetailCmd, classStudentsCmd, classTeachersCmd, classSameNameCmd, classUserRoleCmd, classSearchByNameCmd, classHeadmasterCmd, classSearchByTeacherCmd, classAddStudentCmd, classModifyStudentInfoCmd, classDeleteTeacherCmd, classUpdateClassInfoCmd, classUpdateStudentCmd, classUpdateStudentNumberCmd, classAddUnofficialStudentCmd, classDeleteStudentsCmd, classUpdateStudentMobileCmd, classMoveStudentCmd, classAddTeachersCmd)
	familyCmd.AddCommand(familyChildrenCmd, familyParentsCmd)
	teacherUpdateCourseCmd.Flags().String("teacher-class-infos", "", "教师班级课程信息 JSON 数组（必填）")
	teacherCmd.AddCommand(teacherClassesCmd, teacherUpdateCourseCmd)

	root.AddCommand(schoolCmd, classCmd, familyCmd, teacherCmd)

	return root
}

func eduRequiredIntFlag(cmd *cobra.Command, name string) (int64, error) {
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
