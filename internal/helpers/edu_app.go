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
// dws edu-app — 家校应用
// 共 42 个工具，按 message / task / report / notice / circle / card / diploma / homework 分组
// ──────────────────────────────────────────────────────────

func newEduAppCommand() *cobra.Command {
	contract.RegisterProductDecl(contract.ProductDecl{
		ID: "edu-app",
		Selection: contract.ProductSelectionDecl{
			AgentSummary: "家校应用：AI消息总结、家校任务查询、成绩单管理、通知确认、班级圈动态、打卡管理、奖状颁发、作业管理",
			UseWhen: []string{
				"用户要管理钉钉家校应用中的消息总结、任务、成绩单、通知、班级圈、打卡、奖状或作业。",
			},
			AvoidWhen: []string{
				"家校通讯录用 edu-contact；班级群用 edu-group；家庭群用 edu-familygroup。",
			},
		},
	})
	root := &cobra.Command{
		Use:    "edu-app",
		Short:  "家校应用",
		Long:   `钉钉家校应用：AI消息总结、家校任务查询、成绩单管理、通知确认、奖状颁发、作业管理等。`,
		Hidden: true,
		RunE:   groupRunE,
	}

	// ════════════════════════════════════════════════════════════
	// message 子命令组 — AI 消息
	// ════════════════════════════════════════════════════════════

	messageCmd := &cobra.Command{Use: "message", Short: "AI 消息管理", RunE: groupRunE}

	messageSummaryListCmd := &cobra.Command{
		Use:   "summary-list",
		Short: "获取AI消息总结列表",
		Long: `查询指定群下用户的AI消息总结任务列表。
--target-role 取值：guardian（家长）、student（学生）。
--status 取值：0（未处理）、1（已处理）。`,
		Example: `  dws edu-app message summary-list --class-id 12345 --cid cidxxx --target-role guardian --status 0
  dws edu-app message summary-list --class-id 12345 --cid cidxxx --target-role student --status 1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, err := eduAppRequiredIntFlag(cmd, "class-id")
			if err != nil {
				return err
			}
			cid, _ := cmd.Flags().GetString("cid")
			cid = strings.TrimSpace(cid)
			if cid == "" {
				return fmt.Errorf("--cid 为必填参数")
			}
			targetRole, _ := cmd.Flags().GetString("target-role")
			targetRole = strings.TrimSpace(targetRole)
			if targetRole == "" {
				return fmt.Errorf("--target-role 为必填参数")
			}
			statusStr, _ := cmd.Flags().GetString("status")
			statusStr = strings.TrimSpace(statusStr)
			if statusStr == "" {
				return fmt.Errorf("--status 为必填参数")
			}
			status, err := strconv.ParseInt(statusStr, 10, 64)
			if err != nil {
				return fmt.Errorf("--status 须为整数: %w", err)
			}
			return callMCPToolOnServer("edu-app", "get_ai_message_summary_list", map[string]any{
				"input": map[string]any{
					"classId":    classID,
					"cid":        cid,
					"targetRole": targetRole,
					"status":     status,
				},
			})
		},
	}
	DeclareLeafMetadata(messageSummaryListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_ai_message_summary_list",
				CanonicalPath:  "edu-app.get_ai_message_summary_list",
				CLIPath:        "edu-app message summary-list",
				PrimaryCLIPath: "edu-app message summary-list",
			},
			Description: "获取AI消息总结列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_ai_message_summary_list"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "获取AI消息总结列表",
				UseWhen:      []string{"需要查询指定群下用户的AI消息总结任务列表时"},
				AvoidWhen:    []string{"查询家校任务用 task publish-list"},
				Examples: []string{
					"dws edu-app message summary-list --class-id 12345 --cid cidxxx --target-role guardian --status 0 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "cid", Property: "input.cid", Required: boolPtr(true)},
				{Name: "target-role", Property: "input.targetRole", Required: boolPtr(true)},
				{Name: "status", Property: "input.status", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// task 子命令组 — 家校任务
	// ════════════════════════════════════════════════════════════

	taskCmd := &cobra.Command{Use: "task", Short: "家校任务管理", RunE: groupRunE}

	taskPublishListCmd := &cobra.Command{
		Use:   "publish-list",
		Short: "查询老师发布的任务列表",
		Long: `查询当前老师已发布的家校任务列表，支持分页。
--task-sources 可选值（逗号分隔）：EDU_HOMEWORK（家校本）、EDU_CARD（打卡）、EDU_NOTICE（通知）、EDU_SR（成绩单）、EDU_DIPLOMA（奖状），不传时默认查询全部类型。`,
		Example: `  dws edu-app task publish-list
  dws edu-app task publish-list --limit 10 --need-statistic
  dws edu-app task publish-list --task-sources EDU_HOMEWORK,EDU_NOTICE -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetInt64("cursor"); cmd.Flags().Changed("cursor") {
				input["cursor"] = v
			}
			if v, _ := cmd.Flags().GetInt64("limit"); cmd.Flags().Changed("limit") {
				input["pageSize"] = v
			}
			if v, _ := cmd.Flags().GetBool("need-statistic"); v {
				input["needStatistic"] = true
			}
			if raw, _ := cmd.Flags().GetString("task-sources"); strings.TrimSpace(raw) != "" {
				input["taskSourceList"] = eduAppParseCSV(raw)
			}
			return callMCPToolOnServer("edu-app", "query_publish_task", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(taskPublishListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_publish_task",
				CanonicalPath:  "edu-app.query_publish_task",
				CLIPath:        "edu-app task publish-list",
				PrimaryCLIPath: "edu-app task publish-list",
			},
			Description: "查询老师发布的任务列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_publish_task"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询老师发布的家校任务列表",
				UseWhen:      []string{"需要查询当前老师已发布的家校任务列表时"},
				AvoidWhen:    []string{"查询全部任务用 task all-list；查询学生待办用 task student-list"},
				Examples: []string{
					"dws edu-app task publish-list --limit 10 --need-statistic --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "cursor", Property: "input.cursor"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "need-statistic", Property: "input.needStatistic"},
				{Name: "task-sources", Property: "input.taskSourceList"},
			},
		},
	})

	taskAllListCmd := &cobra.Command{
		Use:   "all-list",
		Short: "查询全部家校任务列表",
		Long: `查询组织内全部家校任务列表，支持分页，仅限老师角色调用。
--task-sources 可选值（逗号分隔）：EDU_HOMEWORK、EDU_CARD、EDU_NOTICE、EDU_SR、EDU_DIPLOMA，不传时默认查询全部类型。`,
		Example: `  dws edu-app task all-list --biz-id 12345
  dws edu-app task all-list --biz-id 12345 --limit 10 --need-statistic -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			bizID, err := eduAppRequiredIntFlag(cmd, "biz-id")
			if err != nil {
				return err
			}
			input := map[string]any{"bizId": bizID}
			if v, _ := cmd.Flags().GetInt64("cursor"); cmd.Flags().Changed("cursor") {
				input["cursor"] = v
			}
			if v, _ := cmd.Flags().GetInt64("limit"); cmd.Flags().Changed("limit") {
				input["pageSize"] = v
			}
			if v, _ := cmd.Flags().GetBool("need-statistic"); v {
				input["needStatistic"] = true
			}
			if raw, _ := cmd.Flags().GetString("task-sources"); strings.TrimSpace(raw) != "" {
				input["taskSourceList"] = eduAppParseCSV(raw)
			}
			return callMCPToolOnServer("edu-app", "query_all_task", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(taskAllListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_all_task",
				CanonicalPath:  "edu-app.query_all_task",
				CLIPath:        "edu-app task all-list",
				PrimaryCLIPath: "edu-app task all-list",
			},
			Description: "查询全部家校任务列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_all_task"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询组织内全部家校任务列表",
				UseWhen:      []string{"需要查询组织内全部家校任务列表时"},
				AvoidWhen:    []string{"查询老师发布的任务用 task publish-list；查询学生待办用 task student-list"},
				Examples: []string{
					"dws edu-app task all-list --biz-id 12345 --limit 10 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "biz-id", Property: "input.bizId", Required: boolPtr(true)},
				{Name: "cursor", Property: "input.cursor"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "need-statistic", Property: "input.needStatistic"},
				{Name: "task-sources", Property: "input.taskSourceList"},
			},
		},
	})

	taskStudentListCmd := &cobra.Command{
		Use:   "student-list",
		Short: "查询学生待办任务列表",
		Long: `查询多个学生的待办任务列表，支持分页，适用于家长/学生角色调用。
--students 为 JSON 数组，每个元素需包含 userId 和 bizId（班级ID）。
--query-all 为 false 时只查待办任务，为 true 时查询全部任务（含已完成）。
--task-sources 可选值（逗号分隔）：EDU_HOMEWORK、EDU_CARD、EDU_NOTICE、EDU_SR、EDU_DIPLOMA。`,
		Example: `  dws edu-app task student-list --students '[{"userId":"uid1","bizId":"12345"}]' --query-all=false
  dws edu-app task student-list --students '[{"userId":"uid1","bizId":"12345"},{"userId":"uid2","bizId":"67890"}]' --query-all=true -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			studentsRaw, _ := cmd.Flags().GetString("students")
			studentsRaw = strings.TrimSpace(studentsRaw)
			if studentsRaw == "" {
				return fmt.Errorf("--students 为必填参数")
			}
			var students []map[string]any
			if err := json.Unmarshal([]byte(studentsRaw), &students); err != nil {
				return fmt.Errorf("--students JSON 格式错误: %w", err)
			}
			queryAll, _ := cmd.Flags().GetBool("query-all")
			input := map[string]any{
				"students": students,
				"queryAll": queryAll,
			}
			if v, _ := cmd.Flags().GetString("cursor"); strings.TrimSpace(v) != "" {
				input["cursor"] = v
			}
			if v, _ := cmd.Flags().GetString("limit"); strings.TrimSpace(v) != "" {
				input["pageSize"] = v
			}
			if raw, _ := cmd.Flags().GetString("task-sources"); strings.TrimSpace(raw) != "" {
				input["taskSourceList"] = eduAppParseCSV(raw)
			}
			return callMCPToolOnServer("edu-app", "query_multi_user_sub_tasks", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(taskStudentListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_multi_user_sub_tasks",
				CanonicalPath:  "edu-app.query_multi_user_sub_tasks",
				CLIPath:        "edu-app task student-list",
				PrimaryCLIPath: "edu-app task student-list",
			},
			Description: "查询学生待办任务列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_multi_user_sub_tasks"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询多个学生的待办任务列表",
				UseWhen:      []string{"需要查询学生的待办或全部家校任务时"},
				AvoidWhen:    []string{"查询老师发布的任务用 task publish-list"},
				Examples: []string{
					"dws edu-app task student-list --students '[{\"userId\":\"uid1\",\"bizId\":\"12345\"}]' --query-all=false --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "students", Property: "input.students", Required: boolPtr(true)},
				{Name: "query-all", Property: "input.queryAll"},
				{Name: "cursor", Property: "input.cursor"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "task-sources", Property: "input.taskSourceList"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// report 子命令组 — 成绩单
	// ════════════════════════════════════════════════════════════

	reportCmd := &cobra.Command{Use: "report", Short: "成绩单管理", RunE: groupRunE}

	reportGetCmd := &cobra.Command{
		Use:   "get",
		Short: "获取成绩单列表",
		Long:  `按成绩单ID列表批量查询成绩单信息。`,
		Example: `  dws edu-app report get --ids 1001,1002,1003
  dws edu-app report get --ids 1001 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--ids 为必填参数")
			}
			ids, err := eduAppParseIntCSV(raw)
			if err != nil {
				return fmt.Errorf("--ids 须为逗号分隔的整数列表: %w", err)
			}
			return callMCPToolOnServer("edu-app", "get_report", map[string]any{
				"input": map[string]any{"schoolReportIdList": ids},
			})
		},
	}
	DeclareLeafMetadata(reportGetCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_report",
				CanonicalPath:  "edu-app.get_report",
				CLIPath:        "edu-app report get",
				PrimaryCLIPath: "edu-app report get",
			},
			Description: "获取成绩单列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_report"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按成绩单ID列表批量查询成绩单信息",
				UseWhen:      []string{"需要按成绩单ID批量查询成绩单信息时"},
				AvoidWhen:    []string{"查询老师创建的成绩单用 report by-teacher；查询学生成绩用 report by-student-detail"},
				Examples: []string{
					"dws edu-app report get --ids 1001,1002,1003 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "ids", Property: "input.schoolReportIdList", Required: boolPtr(true)},
			},
		},
	})

	reportByTeacherCmd := &cobra.Command{
		Use:   "by-teacher",
		Short: "查询老师创建的成绩单列表",
		Long: `查询指定老师创建的成绩单列表，支持分页。
--status 取值：0（未发布）、1（已发布），不传则查询所有状态。`,
		Example: `  dws edu-app report by-teacher
  dws edu-app report by-teacher --status 1 --page 1 --limit 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber <= 0 {
				pageNumber = 1
			}
			input["pageNumber"] = pageNumber
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 20
			}
			input["pageSize"] = pageSize
			if cmd.Flags().Changed("status") {
				statusVal, _ := cmd.Flags().GetInt64("status")
				input["status"] = statusVal
			}
			return callMCPToolOnServer("edu-app", "list_report_by_teacher", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(reportByTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_report_by_teacher",
				CanonicalPath:  "edu-app.list_report_by_teacher",
				CLIPath:        "edu-app report by-teacher",
				PrimaryCLIPath: "edu-app report by-teacher",
			},
			Description: "查询老师创建的成绩单列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_report_by_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询老师创建的成绩单列表",
				UseWhen:      []string{"需要查询指定老师创建的成绩单列表时"},
				AvoidWhen:    []string{"查询学生收到的成绩单用 report by-student-list"},
				Examples: []string{
					"dws edu-app report by-teacher --status 1 --page 1 --limit 20 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "page", Property: "input.pageNumber"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "status", Property: "input.status"},
			},
		},
	})

	reportByClassCmd := &cobra.Command{
		Use:   "by-class",
		Short: "查询班级学生成绩明细",
		Long: `按成绩单ID和班级ID查询该班级学生的成绩明细。
返回学生成绩明细列表及阅读状态汇总统计（totalCount/readCount/unreadCount）。`,
		Example: `  dws edu-app report by-class --report-id 1001 --class-id 12345
  dws edu-app report by-class --report-id 1001 --class-id 12345 --student-ids uid1,uid2 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportIDStr, _ := cmd.Flags().GetString("report-id")
			reportIDStr = strings.TrimSpace(reportIDStr)
			if reportIDStr == "" {
				return fmt.Errorf("--report-id 为必填参数")
			}
			reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("--report-id 须为整数: %w", err)
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"schoolReportId": reportID,
				"classId":        classID,
			}
			if raw, _ := cmd.Flags().GetString("student-ids"); strings.TrimSpace(raw) != "" {
				input["studentIdList"] = eduAppParseCSV(raw)
			}
			return callMCPToolOnServer("edu-app", "query_detail_by_class", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(reportByClassCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_detail_by_class",
				CanonicalPath:  "edu-app.query_detail_by_class",
				CLIPath:        "edu-app report by-class",
				PrimaryCLIPath: "edu-app report by-class",
			},
			Description: "查询班级学生成绩明细",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_detail_by_class"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按成绩单ID和班级ID查询学生成绩明细",
				UseWhen:      []string{"需要按成绩单ID和班级ID查询班级学生的成绩明细时"},
				AvoidWhen:    []string{"查询单个学生成绩用 report by-student-detail"},
				Examples: []string{
					"dws edu-app report by-class --report-id 1001 --class-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "report-id", Property: "input.schoolReportId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "student-ids", Property: "input.studentIdList"},
			},
		},
	})

	reportByStudentListCmd := &cobra.Command{
		Use:   "by-student-list",
		Short: "查询学生收到的成绩单列表",
		Long:  `查询指定学生收到的已发布成绩单列表，支持分页。`,
		Example: `  dws edu-app report by-student-list --class-id 12345 --student-id uid1
  dws edu-app report by-student-list --class-id 12345 --student-id uid1 --page 1 --limit 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber <= 0 {
				pageNumber = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 20
			}
			return callMCPToolOnServer("edu-app", "list_report_by_student", map[string]any{
				"input": map[string]any{
					"classId":    classID,
					"studentId":  studentID,
					"pageNumber": pageNumber,
					"pageSize":   pageSize,
				},
			})
		},
	}
	DeclareLeafMetadata(reportByStudentListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_report_by_student",
				CanonicalPath:  "edu-app.list_report_by_student",
				CLIPath:        "edu-app report by-student-list",
				PrimaryCLIPath: "edu-app report by-student-list",
			},
			Description: "查询学生收到的成绩单列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_report_by_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学生收到的已发布成绩单列表",
				UseWhen:      []string{"需要查询指定学生收到的成绩单列表时"},
				AvoidWhen:    []string{"查询单个学生成绩明细用 report by-student-detail"},
				Examples: []string{
					"dws edu-app report by-student-list --class-id 12345 --student-id uid1 --page 1 --limit 20 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "limit", Property: "input.pageSize"},
			},
		},
	})

	reportByStudentDetailCmd := &cobra.Command{
		Use:   "by-student-detail",
		Short: "查询学生成绩明细",
		Long: `按成绩单ID查询指定学生的成绩明细。
返回该学生的成绩明细列表，包含各科成绩及已读状态（readStatus：1-未读、2-已读）。`,
		Example: `  dws edu-app report by-student-detail --report-id 1001 --student-id uid1 --class-id 12345
  dws edu-app report by-student-detail --report-id 1001 --student-id uid1 --class-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reportIDStr, _ := cmd.Flags().GetString("report-id")
			reportIDStr = strings.TrimSpace(reportIDStr)
			if reportIDStr == "" {
				return fmt.Errorf("--report-id 为必填参数")
			}
			reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("--report-id 须为整数: %w", err)
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			return callMCPToolOnServer("edu-app", "query_detail_by_student", map[string]any{
				"input": map[string]any{
					"schoolReportId": reportID,
					"studentId":      studentID,
					"classId":        classID,
				},
			})
		},
	}
	DeclareLeafMetadata(reportByStudentDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_detail_by_student",
				CanonicalPath:  "edu-app.query_detail_by_student",
				CLIPath:        "edu-app report by-student-detail",
				PrimaryCLIPath: "edu-app report by-student-detail",
			},
			Description: "查询学生成绩明细",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_detail_by_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按成绩单ID查询指定学生的成绩明细",
				UseWhen:      []string{"需要查询指定学生的成绩明细时"},
				AvoidWhen:    []string{"查询班级整体成绩用 report by-class"},
				Examples: []string{
					"dws edu-app report by-student-detail --report-id 1001 --student-id uid1 --class-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "report-id", Property: "input.schoolReportId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// notice 子命令组 — 通知
	// ════════════════════════════════════════════════════════════

	noticeCmd := &cobra.Command{Use: "notice", Short: "通知管理", RunE: groupRunE}

	noticeConfirmCmd := &cobra.Command{
		Use:   "confirm",
		Short: "确认收到通知",
		Long:  `家长端确认收到通知。返回确认结果（confirmed: true 表示确认成功）。`,
		Example: `  dws edu-app notice confirm --notice-id nid1 --student-id uid1
  dws edu-app notice confirm --notice-id nid1 --student-id uid1 --parent-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noticeID, _ := cmd.Flags().GetString("notice-id")
			noticeID = strings.TrimSpace(noticeID)
			if noticeID == "" {
				return fmt.Errorf("--notice-id 为必填参数")
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			input := map[string]any{
				"noticeId":  noticeID,
				"studentId": studentID,
			}
			if v, _ := cmd.Flags().GetString("device-id"); strings.TrimSpace(v) != "" {
				input["deviceId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("parent-name"); strings.TrimSpace(v) != "" {
				input["parentName"] = strings.TrimSpace(v)
			}
			if cmd.Flags().Changed("update-sign") {
				updateSign, _ := cmd.Flags().GetBool("update-sign")
				input["updateSign"] = updateSign
			}
			return callMCPToolOnServer("edu-app", "confirm_notice", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeConfirmCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "confirm_notice",
				CanonicalPath:  "edu-app.confirm_notice",
				CLIPath:        "edu-app notice confirm",
				PrimaryCLIPath: "edu-app notice confirm",
			},
			Description: "确认收到通知",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "confirm_notice"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "家长端确认收到通知",
				UseWhen:      []string{"需要确认收到家校通知时"},
				AvoidWhen:    []string{"查询通知详情用 notice get；查询确认状态用 notice confirm-status"},
				Examples: []string{
					"dws edu-app notice confirm --notice-id nid1 --student-id uid1 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "notice-id", Property: "input.noticeId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "device-id", Property: "input.deviceId"},
				{Name: "parent-name", Property: "input.parentName"},
				{Name: "update-sign", Property: "input.updateSign"},
			},
		},
	})

	noticeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建并发布通知",
		Long: `创建并发布一条家校通知。
--identifer 为幂等字段（必填），建议格式：orgId-staffId-UUID。
--content 为通知内容（必填）。
--class-ids、--class-names 为逗号分隔的班级列表，--class-selected-students 为 JSON 对象。`,
		Example: `  dws edu-app notice create --identifer org1-staff1-uuid --content "明天放假"
  dws edu-app notice create --identifer org1-staff1-uuid --title "放假通知" --content "明天放假" --class-ids 12345,67890 --target-role guardian --is-signed true -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			identifer, _ := cmd.Flags().GetString("identifer")
			identifer = strings.TrimSpace(identifer)
			if identifer == "" {
				return fmt.Errorf("--identifer 为必填参数")
			}
			content, _ := cmd.Flags().GetString("content")
			content = strings.TrimSpace(content)
			if content == "" {
				return fmt.Errorf("--content 为必填参数")
			}
			input := map[string]any{
				"identifer": identifer,
				"content":   content,
			}
			if v, _ := cmd.Flags().GetString("title"); strings.TrimSpace(v) != "" {
				input["title"] = strings.TrimSpace(v)
			}
			if raw, _ := cmd.Flags().GetString("class-ids"); strings.TrimSpace(raw) != "" {
				input["classIds"] = eduAppParseCSV(raw)
			}
			if raw, _ := cmd.Flags().GetString("class-names"); strings.TrimSpace(raw) != "" {
				input["classNames"] = eduAppParseCSV(raw)
			}
			if raw, _ := cmd.Flags().GetString("class-selected-students"); strings.TrimSpace(raw) != "" {
				var obj any
				if err := json.Unmarshal([]byte(raw), &obj); err != nil {
					return fmt.Errorf("--class-selected-students JSON 格式错误: %w", err)
				}
				input["classSelectedStudents"] = obj
			}
			if v, _ := cmd.Flags().GetString("type"); strings.TrimSpace(v) != "" {
				input["type"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("scope"); strings.TrimSpace(v) != "" {
				input["scope"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("target-role"); strings.TrimSpace(v) != "" {
				input["targetRole"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("is-signed"); strings.TrimSpace(v) != "" {
				input["isSigned"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("photo"); strings.TrimSpace(v) != "" {
				input["photo"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("media"); strings.TrimSpace(v) != "" {
				input["media"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("audio"); strings.TrimSpace(v) != "" {
				input["audio"] = strings.TrimSpace(v)
			}
			if cmd.Flags().Changed("send-ding") {
				v, _ := cmd.Flags().GetBool("send-ding")
				input["sendDing"] = v
			}
			if v, _ := cmd.Flags().GetString("scheduled-release"); strings.TrimSpace(v) != "" {
				input["scheduledRelease"] = strings.TrimSpace(v)
			}
			if cmd.Flags().Changed("notice-deadline") {
				v, _ := cmd.Flags().GetInt64("notice-deadline")
				input["noticeDeadline"] = v
			}
			if v, _ := cmd.Flags().GetString("notice-deadline-open"); strings.TrimSpace(v) != "" {
				input["noticeDeadlineOpen"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("notice-deadline-setting"); strings.TrimSpace(v) != "" {
				input["noticeDeadlineSetting"] = strings.TrimSpace(v)
			}
			if raw, _ := cmd.Flags().GetString("attributes"); strings.TrimSpace(raw) != "" {
				var obj any
				if err := json.Unmarshal([]byte(raw), &obj); err != nil {
					return fmt.Errorf("--attributes JSON 格式错误: %w", err)
				}
				input["attributes"] = obj
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "create_notice", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeCreateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "create_notice",
				CanonicalPath:  "edu-app.create_notice",
				CLIPath:        "edu-app notice create",
				PrimaryCLIPath: "edu-app notice create",
			},
			Description: "创建并发布通知",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "create_notice"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建并发布一条家校通知",
				UseWhen:      []string{"需要创建并发布家校通知时"},
				AvoidWhen:    []string{"删除通知用 notice delete；查询通知用 notice get"},
				Examples: []string{
					"dws edu-app notice create --identifer org1-staff1-uuid --content \"明天放假\" --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "identifer", Property: "input.identifer", Required: boolPtr(true)},
				{Name: "content", Property: "input.content", Required: boolPtr(true)},
				{Name: "title", Property: "input.title"},
				{Name: "class-ids", Property: "input.classIds"},
				{Name: "class-names", Property: "input.classNames"},
				{Name: "class-selected-students", Property: "input.classSelectedStudents"},
				{Name: "type", Property: "input.type"},
				{Name: "scope", Property: "input.scope"},
				{Name: "target-role", Property: "input.targetRole"},
				{Name: "is-signed", Property: "input.isSigned"},
				{Name: "photo", Property: "input.photo"},
				{Name: "media", Property: "input.media"},
				{Name: "audio", Property: "input.audio"},
				{Name: "send-ding", Property: "input.sendDing"},
				{Name: "scheduled-release", Property: "input.scheduledRelease"},
				{Name: "notice-deadline", Property: "input.noticeDeadline"},
				{Name: "notice-deadline-open", Property: "input.noticeDeadlineOpen"},
				{Name: "notice-deadline-setting", Property: "input.noticeDeadlineSetting"},
				{Name: "attributes", Property: "input.attributes"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	noticeDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "删除通知",
		Long: `删除指定的家校通知。
--notice-id 为通知 ID（必填），--user-name 为可选参数。`,
		Example: `  dws edu-app notice delete --notice-id 12345
  dws edu-app notice delete --notice-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noticeID, err := eduAppRequiredIntFlag(cmd, "notice-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"noticeId": noticeID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "delete_notice", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeDeleteCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "delete_notice",
				CanonicalPath:  "edu-app.delete_notice",
				CLIPath:        "edu-app notice delete",
				PrimaryCLIPath: "edu-app notice delete",
			},
			Description: "删除通知",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "delete_notice"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除指定的家校通知",
				UseWhen:      []string{"需要删除指定的家校通知时"},
				AvoidWhen:    []string{"查询通知详情用 notice get；创建通知用 notice create"},
				Examples: []string{
					"dws edu-app notice delete --notice-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "notice-id", Property: "input.noticeId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	noticeListByTeacherCmd := &cobra.Command{
		Use:   "list-by-teacher",
		Short: "查询老师发布的通知列表",
		Long: `查询老师发布的家校通知列表，支持分页和筛选。
所有参数均为可选，--class-id 按班级筛选，--type 按通知类型筛选，--status 按通知状态筛选。`,
		Example: `  dws edu-app notice list-by-teacher
  dws edu-app notice list-by-teacher --class-id 67890 --type SCHOOL --status FINISHED --page 1 --page-size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("type"); strings.TrimSpace(v) != "" {
				input["type"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber > 0 {
				input["pageNumber"] = pageNumber
			} else {
				input["pageNumber"] = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("page-size")
			if pageSize > 0 {
				input["pageSize"] = pageSize
			} else {
				input["pageSize"] = 20
			}
			return callMCPToolOnServer("edu-app", "list_notice_by_teacher", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeListByTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_notice_by_teacher",
				CanonicalPath:  "edu-app.list_notice_by_teacher",
				CLIPath:        "edu-app notice list-by-teacher",
				PrimaryCLIPath: "edu-app notice list-by-teacher",
			},
			Description: "查询老师发布的通知列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_notice_by_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询老师发布的家校通知列表",
				UseWhen:      []string{"需要查询老师发布的家校通知列表时"},
				AvoidWhen:    []string{"查询学生通知用 notice list-by-student"},
				Examples: []string{
					"dws edu-app notice list-by-teacher --class-id 67890 --page 1 --page-size 20 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId"},
				{Name: "type", Property: "input.type"},
				{Name: "status", Property: "input.status"},
				{Name: "user-name", Property: "input.userName"},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "page-size", Property: "input.pageSize"},
			},
		},
	})

	noticeGetCmd := &cobra.Command{
		Use:   "get",
		Short: "查询通知详情",
		Long: `按通知 ID 查询家校通知详情。
--notice-id 为通知 ID（必填），--user-name 为可选参数。`,
		Example: `  dws edu-app notice get --notice-id 12345
  dws edu-app notice get --notice-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noticeID, err := eduAppRequiredIntFlag(cmd, "notice-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"noticeId": noticeID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "get_notice", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeGetCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_notice",
				CanonicalPath:  "edu-app.get_notice",
				CLIPath:        "edu-app notice get",
				PrimaryCLIPath: "edu-app notice get",
			},
			Description: "查询通知详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_notice"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按通知ID查询家校通知详情",
				UseWhen:      []string{"需要查询指定通知的详情时"},
				AvoidWhen:    []string{"查询通知列表用 notice list-by-teacher 或 notice list-by-student"},
				Examples: []string{
					"dws edu-app notice get --notice-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "notice-id", Property: "input.noticeId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	noticeConfirmStatusCmd := &cobra.Command{
		Use:   "confirm-status",
		Short: "查询通知确认状态",
		Long: `按通知 ID 和班级 ID 查询通知的确认状态（已确认/未确认人数及明细）。
--notice-id、--class-id 为必填参数，--status、--page、--page-size、--user-name 为可选参数。`,
		Example: `  dws edu-app notice confirm-status --notice-id 12345 --class-id 67890
  dws edu-app notice confirm-status --notice-id 12345 --class-id 67890 --status CONFIRMED --page 1 --page-size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noticeID, err := eduAppRequiredIntFlag(cmd, "notice-id")
			if err != nil {
				return err
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"noticeId": noticeID,
				"classId":  classID,
			}
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber > 0 {
				input["pageNumber"] = pageNumber
			} else {
				input["pageNumber"] = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("page-size")
			if pageSize > 0 {
				input["pageSize"] = pageSize
			} else {
				input["pageSize"] = 20
			}
			return callMCPToolOnServer("edu-app", "query_notice_confirm_status", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeConfirmStatusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_notice_confirm_status",
				CanonicalPath:  "edu-app.query_notice_confirm_status",
				CLIPath:        "edu-app notice confirm-status",
				PrimaryCLIPath: "edu-app notice confirm-status",
			},
			Description: "查询通知确认状态",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_notice_confirm_status"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按通知ID和班级ID查询通知确认状态",
				UseWhen:      []string{"需要查询通知的确认状态时"},
				AvoidWhen:    []string{"确认通知用 notice confirm；查询通知详情用 notice get"},
				Examples: []string{
					"dws edu-app notice confirm-status --notice-id 12345 --class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "notice-id", Property: "input.noticeId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "status", Property: "input.status"},
				{Name: "user-name", Property: "input.userName"},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "page-size", Property: "input.pageSize"},
			},
		},
	})

	noticeListByStudentCmd := &cobra.Command{
		Use:   "list-by-student",
		Short: "查询学生通知列表",
		Long: `按学生 ID 和班级 ID 查询学生收到的家校通知列表，支持分页和状态筛选。
--student-id、--class-id 为必填参数，--status、--page、--page-size、--user-name 为可选参数。`,
		Example: `  dws edu-app notice list-by-student --student-id uid1 --class-id 67890
  dws edu-app notice list-by-student --student-id uid1 --class-id 67890 --status FINISHED --page 1 --page-size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"studentId": studentID,
				"classId":   classID,
			}
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber > 0 {
				input["pageNumber"] = pageNumber
			} else {
				input["pageNumber"] = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("page-size")
			if pageSize > 0 {
				input["pageSize"] = pageSize
			} else {
				input["pageSize"] = 20
			}
			return callMCPToolOnServer("edu-app", "list_notice_by_student", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(noticeListByStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_notice_by_student",
				CanonicalPath:  "edu-app.list_notice_by_student",
				CLIPath:        "edu-app notice list-by-student",
				PrimaryCLIPath: "edu-app notice list-by-student",
			},
			Description: "查询学生通知列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_notice_by_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按学生ID和班级ID查询学生收到的通知列表",
				UseWhen:      []string{"需要查询学生收到的家校通知列表时"},
				AvoidWhen:    []string{"查询老师发布的通知用 notice list-by-teacher"},
				Examples: []string{
					"dws edu-app notice list-by-student --student-id uid1 --class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "status", Property: "input.status"},
				{Name: "user-name", Property: "input.userName"},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "page-size", Property: "input.pageSize"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// circle 子命令组 — 班级圈
	// ════════════════════════════════════════════════════════════

	circleCmd := &cobra.Command{Use: "circle", Short: "班级圈管理", RunE: groupRunE}

	circlePostsCmd := &cobra.Command{
		Use:   "posts",
		Short: "查询学生班级圈动态",
		Long: `查询指定学生在班级圈中的动态（文字、图片等成长记录）。
--target-role 取值：guardian（家长视角）、student（学生视角）。
返回每条动态的文字内容、图片URL列表、发布者姓名、发布时间、评论数、点赞数等信息。`,
		Example: `  dws edu-app circle posts --class-id 12345 --student-id uid1 --target-role guardian
  dws edu-app circle posts --class-id 12345 --student-id uid1 --target-role student -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			targetRole, _ := cmd.Flags().GetString("target-role")
			targetRole = strings.TrimSpace(targetRole)
			if targetRole == "" {
				return fmt.Errorf("--target-role 为必填参数")
			}
			return callMCPToolOnServer("edu-app", "query_student_circle_posts", map[string]any{
				"input": map[string]any{
					"classId":    classID,
					"studentId":  studentID,
					"targetRole": targetRole,
				},
			})
		},
	}
	DeclareLeafMetadata(circlePostsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_student_circle_posts",
				CanonicalPath:  "edu-app.query_student_circle_posts",
				CLIPath:        "edu-app circle posts",
				PrimaryCLIPath: "edu-app circle posts",
			},
			Description: "查询学生班级圈动态",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_student_circle_posts"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学生在班级圈中的动态",
				UseWhen:      []string{"需要查询学生在班级圈中的成长记录动态时"},
				AvoidWhen:    []string{"查询家校任务用 task；查询作业用 homework"},
				Examples: []string{
					"dws edu-app circle posts --class-id 12345 --student-id uid1 --target-role guardian --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "target-role", Property: "input.targetRole", Required: boolPtr(true)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// card 子命令组 — 打卡
	// ════════════════════════════════════════════════════════════

	cardCmd := &cobra.Command{Use: "card", Short: "打卡管理", RunE: groupRunE}

	cardUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "修改打卡标题或内容",
		Long: `修改指定打卡任务的标题或内容（仅允许修改这两项，其他字段忽略）。仅老师可调用且操作人必须为打卡创建者。
--title 与 --content 至少传一个。
--identifier 建议格式：orgId-staffId-UUID，用于幂等去重。`,
		Example: `  dws edu-app card update --card-id 12345 --identifier org1-staff1-uuid --title "新标题"
  dws edu-app card update --card-id 12345 --identifier org1-staff1-uuid --content "新内容" --should-send-update-msg true -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID, err := eduAppRequiredIntFlag(cmd, "card-id")
			if err != nil {
				return err
			}
			identifier, _ := cmd.Flags().GetString("identifier")
			identifier = strings.TrimSpace(identifier)
			if identifier == "" {
				return fmt.Errorf("--identifier 为必填参数")
			}
			title, _ := cmd.Flags().GetString("title")
			title = strings.TrimSpace(title)
			content, _ := cmd.Flags().GetString("content")
			content = strings.TrimSpace(content)
			if title == "" && content == "" {
				return fmt.Errorf("--title 与 --content 至少传一个")
			}
			input := map[string]any{
				"cardId":     cardID,
				"identifier": identifier,
			}
			if title != "" {
				input["title"] = title
			}
			if content != "" {
				input["content"] = content
			}
			if cmd.Flags().Changed("should-send-update-msg") {
				v, _ := cmd.Flags().GetBool("should-send-update-msg")
				input["shouldSendUpdateMsg"] = v
			}
			return callMCPToolOnServer("edu-app", "update_card", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(cardUpdateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "update_card",
				CanonicalPath:  "edu-app.update_card",
				CLIPath:        "edu-app card update",
				PrimaryCLIPath: "edu-app card update",
			},
			Description: "修改打卡标题或内容",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "update_card"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "修改打卡任务的标题或内容",
				UseWhen:      []string{"需要修改打卡任务的标题或内容时"},
				AvoidWhen:    []string{"结束打卡用 card end；查询打卡用 card list"},
				Examples: []string{
					"dws edu-app card update --card-id 12345 --identifier org1-staff1-uuid --title \"新标题\" --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "card-id", Property: "input.cardId", Required: boolPtr(true)},
				{Name: "identifier", Property: "input.identifier", Required: boolPtr(true)},
				{Name: "title", Property: "input.title"},
				{Name: "content", Property: "input.content"},
				{Name: "should-send-update-msg", Property: "input.shouldSendUpdateMsg"},
			},
		},
	})

	cardEndCmd := &cobra.Command{
		Use:   "end",
		Short: "提前结束打卡任务",
		Long:  `提前结束指定打卡任务，仅老师可调用且操作人必须为打卡创建者。适用于老师在打卡周期未到期时手动终止打卡任务。`,
		Example: `  dws edu-app card end --card-id 12345
  dws edu-app card end --card-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID, err := eduAppRequiredIntFlag(cmd, "card-id")
			if err != nil {
				return err
			}
			return callMCPToolOnServer("edu-app", "end_card", map[string]any{
				"input": map[string]any{
					"cardId": cardID,
				},
			})
		},
	}
	DeclareLeafMetadata(cardEndCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "end_card",
				CanonicalPath:  "edu-app.end_card",
				CLIPath:        "edu-app card end",
				PrimaryCLIPath: "edu-app card end",
			},
			Description: "提前结束打卡任务",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "end_card"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "提前结束打卡任务",
				UseWhen:      []string{"需要提前结束指定打卡任务时"},
				AvoidWhen:    []string{"修改打卡用 card update；查询打卡用 card list"},
				Examples: []string{
					"dws edu-app card end --card-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "card-id", Property: "input.cardId", Required: boolPtr(true)},
			},
		},
	})

	cardListCmd := &cobra.Command{
		Use:   "list",
		Short: "查询打卡列表",
		Long: `查询孩子或本人的全部打卡列表（含进行中与已完结），支持分页。
角色由 uid 真实身份自动判断：学生查自己、家长查孩子，老师调用返回空列表。
--status 取值：FINISH（已完结）、UNFINISH（进行中）。`,
		Example: `  dws edu-app card list --status UNFINISH
  dws edu-app card list --status FINISH --class-id 12345 --page 1 --limit 10 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			status = strings.TrimSpace(status)
			if status == "" {
				return fmt.Errorf("--status 为必填参数")
			}
			if status != "FINISH" && status != "UNFINISH" {
				return fmt.Errorf("--status 取值必须为 FINISH 或 UNFINISH")
			}
			input := map[string]any{
				"status": status,
			}
			if cmd.Flags().Changed("class-id") {
				v, _ := cmd.Flags().GetInt64("class-id")
				input["classId"] = v
			}
			pageNo, _ := cmd.Flags().GetInt64("page")
			if pageNo <= 0 {
				pageNo = 1
			}
			input["pageNo"] = pageNo
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 10
			}
			input["pageSize"] = pageSize
			return callMCPToolOnServer("edu-app", "get_card_list", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(cardListCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_card_list",
				CanonicalPath:  "edu-app.get_card_list",
				CLIPath:        "edu-app card list",
				PrimaryCLIPath: "edu-app card list",
			},
			Description: "查询打卡列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_card_list"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询打卡列表",
				UseWhen:      []string{"需要查询打卡列表时"},
				AvoidWhen:    []string{"查询打卡完成情况用 card user-statistic；查询打卡详情用 card finish-info"},
				Examples: []string{
					"dws edu-app card list --status UNFINISH --page 1 --limit 10 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "status", Property: "input.status", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId"},
				{Name: "page", Property: "input.pageNo"},
				{Name: "limit", Property: "input.pageSize"},
			},
		},
	})

	cardUserStatisticCmd := &cobra.Command{
		Use:   "user-statistic",
		Short: "查询班级打卡完成/未完成人员",
		Long: `查询指定打卡任务下某班级的已完成或未完成人员列表，支持分页。
仅老师/班主任可调用（角色由 uid 真实身份自动判断）。
--finish 为 true 时查已完成人员，为 false 时查未完成人员。
注意：--class-id 为 string 类型。`,
		Example: `  dws edu-app card user-statistic --card-id 12345 --task-code code1 --class-id cid1
  dws edu-app card user-statistic --card-id 12345 --task-code code1 --class-id cid1 --finish --page 1 --limit 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID, err := eduAppRequiredIntFlag(cmd, "card-id")
			if err != nil {
				return err
			}
			taskCode, _ := cmd.Flags().GetString("task-code")
			taskCode = strings.TrimSpace(taskCode)
			if taskCode == "" {
				return fmt.Errorf("--task-code 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"cardId":   cardID,
				"taskCode": taskCode,
				"classId":  classID,
			}
			if cmd.Flags().Changed("finish") {
				v, _ := cmd.Flags().GetBool("finish")
				input["finish"] = v
			}
			pageNo, _ := cmd.Flags().GetInt64("page")
			if pageNo <= 0 {
				pageNo = 1
			}
			input["pageNo"] = pageNo
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 10
			}
			input["pageSize"] = pageSize
			return callMCPToolOnServer("edu-app", "get_card_user_statistic", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(cardUserStatisticCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_card_user_statistic",
				CanonicalPath:  "edu-app.get_card_user_statistic",
				CLIPath:        "edu-app card user-statistic",
				PrimaryCLIPath: "edu-app card user-statistic",
			},
			Description: "查询班级打卡完成/未完成人员",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_card_user_statistic"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询班级打卡完成/未完成人员列表",
				UseWhen:      []string{"需要查询打卡任务下某班级的完成或未完成人员时"},
				AvoidWhen:    []string{"查询打卡列表用 card list；查询打卡详情用 card finish-info"},
				Examples: []string{
					"dws edu-app card user-statistic --card-id 12345 --task-code code1 --class-id cid1 --finish --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "card-id", Property: "input.cardId", Required: boolPtr(true)},
				{Name: "task-code", Property: "input.taskCode", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "finish", Property: "input.finish"},
				{Name: "page", Property: "input.pageNo"},
				{Name: "limit", Property: "input.pageSize"},
			},
		},
	})

	cardFinishInfoCmd := &cobra.Command{
		Use:   "finish-info",
		Short: "查询打卡详情及完成进度",
		Long: `查询指定打卡任务的详情及完成进度。
--target-role 取值：teacher（老师视角）、guardian（家长/学生视角），未传时按 uid 真实身份自动推断。
当 targetRole 为 guardian 时，可通过 --student-id 指定查看某个孩子的进度。`,
		Example: `  dws edu-app card finish-info --card-id 12345 --card-biz-id bid1
  dws edu-app card finish-info --card-id 12345 --card-biz-id bid1 --target-role guardian --student-id stu1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID, err := eduAppRequiredIntFlag(cmd, "card-id")
			if err != nil {
				return err
			}
			cardBizID, _ := cmd.Flags().GetString("card-biz-id")
			cardBizID = strings.TrimSpace(cardBizID)
			if cardBizID == "" {
				return fmt.Errorf("--card-biz-id 为必填参数")
			}
			input := map[string]any{
				"cardId":    cardID,
				"cardBizId": cardBizID,
			}
			if v, _ := cmd.Flags().GetString("target-role"); strings.TrimSpace(v) != "" {
				role := strings.TrimSpace(v)
				if role != "teacher" && role != "guardian" && role != "student" && role != "headmaster" {
					return fmt.Errorf("--target-role 取值必须为 teacher / guardian / student / headmaster")
				}
				input["targetRole"] = role
			}
			if v, _ := cmd.Flags().GetString("student-id"); strings.TrimSpace(v) != "" {
				input["studentId"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "get_card_finish_info", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(cardFinishInfoCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_card_finish_info",
				CanonicalPath:  "edu-app.get_card_finish_info",
				CLIPath:        "edu-app card finish-info",
				PrimaryCLIPath: "edu-app card finish-info",
			},
			Description: "查询打卡详情及完成进度",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_card_finish_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询打卡任务详情及完成进度",
				UseWhen:      []string{"需要查询打卡任务的详情及完成进度时"},
				AvoidWhen:    []string{"查询打卡列表用 card list；查询打卡人员统计用 card user-statistic"},
				Examples: []string{
					"dws edu-app card finish-info --card-id 12345 --card-biz-id bid1 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "card-id", Property: "input.cardId", Required: boolPtr(true)},
				{Name: "card-biz-id", Property: "input.cardBizId", Required: boolPtr(true)},
				{Name: "target-role", Property: "input.targetRole"},
				{Name: "student-id", Property: "input.studentId"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// diploma 子命令组 — 奖状
	// ════════════════════════════════════════════════════════════

	diplomaCmd := &cobra.Command{Use: "diploma", Short: "奖状管理", RunE: groupRunE}

	diplomaCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建并颁发奖状",
		Long: `创建并颁发一张奖状。
--identifier 为幂等标识（必填），建议格式：orgId-staffId-UUID，重复提交同一 identifier 不会重复创建。
--select-class 为选择的班级及学生 JSON 数组，--attributes 为扩展属性 JSON 对象。
返回创建结果：success 表示是否创建成功，diplomaId 为新创建的奖状 ID。`,
		Example: `  dws edu-app diploma create --identifier org1-staff1-uuid --content "期末三好学生" --user-name 张三
  dws edu-app diploma create --identifier org1-staff1-uuid --title "三好学生" --content "表现优秀" --user-name 张三 --unit-name "实验小学" --publish-time "2026-07-29" -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier, _ := cmd.Flags().GetString("identifier")
			identifier = strings.TrimSpace(identifier)
			if identifier == "" {
				return fmt.Errorf("--identifier 为必填参数")
			}
			content, _ := cmd.Flags().GetString("content")
			content = strings.TrimSpace(content)
			if content == "" {
				return fmt.Errorf("--content 为必填参数")
			}
			userName, _ := cmd.Flags().GetString("user-name")
			userName = strings.TrimSpace(userName)
			if userName == "" {
				return fmt.Errorf("--user-name 为必填参数")
			}
			input := map[string]any{
				"identifier": identifier,
				"content":    content,
				"userName":   userName,
			}
			if v, _ := cmd.Flags().GetString("title"); strings.TrimSpace(v) != "" {
				input["title"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("unit-name"); strings.TrimSpace(v) != "" {
				input["unitName"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("tag"); strings.TrimSpace(v) != "" {
				input["tag"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("photo"); strings.TrimSpace(v) != "" {
				input["photo"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("publish-time"); strings.TrimSpace(v) != "" {
				input["publishTime"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("biz-code"); strings.TrimSpace(v) != "" {
				input["bizCode"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("biz-category"); strings.TrimSpace(v) != "" {
				input["bizCategory"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("msg-type"); strings.TrimSpace(v) != "" {
				input["msgType"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("template-url"); strings.TrimSpace(v) != "" {
				input["templateUrl"] = strings.TrimSpace(v)
			}
			if raw, _ := cmd.Flags().GetString("class-ids"); strings.TrimSpace(raw) != "" {
				input["classIds"] = eduAppParseCSV(raw)
			}
			if raw, _ := cmd.Flags().GetString("select-class"); strings.TrimSpace(raw) != "" {
				var selectClass []map[string]any
				if err := json.Unmarshal([]byte(raw), &selectClass); err != nil {
					return fmt.Errorf("--select-class JSON 格式错误: %w", err)
				}
				input["selectClass"] = selectClass
			}
			if raw, _ := cmd.Flags().GetString("attributes"); strings.TrimSpace(raw) != "" {
				var attributes map[string]any
				if err := json.Unmarshal([]byte(raw), &attributes); err != nil {
					return fmt.Errorf("--attributes JSON 格式错误: %w", err)
				}
				input["attributes"] = attributes
			}
			return callMCPToolOnServer("edu-app", "create_diploma", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaCreateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "create_diploma",
				CanonicalPath:  "edu-app.create_diploma",
				CLIPath:        "edu-app diploma create",
				PrimaryCLIPath: "edu-app diploma create",
			},
			Description: "创建并颁发奖状",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "create_diploma"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建并颁发一张奖状",
				UseWhen:      []string{"需要创建并颁发奖状时"},
				AvoidWhen:    []string{"查询奖状用 diploma get；删除奖状用 diploma delete"},
				Examples: []string{
					"dws edu-app diploma create --identifier org1-staff1-uuid --content \"期末三好学生\" --user-name 张三 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "identifier", Property: "input.identifier", Required: boolPtr(true)},
				{Name: "content", Property: "input.content", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName", Required: boolPtr(true)},
				{Name: "title", Property: "input.title"},
				{Name: "unit-name", Property: "input.unitName"},
				{Name: "tag", Property: "input.tag"},
				{Name: "photo", Property: "input.photo"},
				{Name: "publish-time", Property: "input.publishTime"},
				{Name: "biz-code", Property: "input.bizCode"},
				{Name: "biz-category", Property: "input.bizCategory"},
				{Name: "msg-type", Property: "input.msgType"},
				{Name: "template-url", Property: "input.templateUrl"},
				{Name: "class-ids", Property: "input.classIds"},
				{Name: "select-class", Property: "input.selectClass"},
				{Name: "attributes", Property: "input.attributes"},
			},
		},
	})

	diplomaReadCmd := &cobra.Command{
		Use:   "read",
		Short: "标记奖状为已读",
		Long: `标记指定奖状为已读状态。
--diploma-id 为奖状 ID（必填），--class-id、--student-id、--user-name 为可选参数。`,
		Example: `  dws edu-app diploma read --diploma-id 12345
  dws edu-app diploma read --diploma-id 12345 --class-id 67890 --student-id uid1 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"diplomaId": diplomaID,
			}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("student-id"); strings.TrimSpace(v) != "" {
				input["studentId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "read_diploma", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaReadCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "read_diploma",
				CanonicalPath:  "edu-app.read_diploma",
				CLIPath:        "edu-app diploma read",
				PrimaryCLIPath: "edu-app diploma read",
			},
			Description: "标记奖状为已读",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "read_diploma"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "标记指定奖状为已读状态",
				UseWhen:      []string{"需要标记奖状为已读时"},
				AvoidWhen:    []string{"查询奖状详情用 diploma get；查询奖状统计用 diploma statistics"},
				Examples: []string{
					"dws edu-app diploma read --diploma-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId"},
				{Name: "student-id", Property: "input.studentId"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaListByTeacherCmd := &cobra.Command{
		Use:   "list-by-teacher",
		Short: "查询老师创建的奖状列表",
		Long: `查询当前老师创建的奖状列表，支持分页。
--status 为奖状状态筛选，--tag 为标签筛选，--user-name 为获奖人姓名筛选。`,
		Example: `  dws edu-app diploma list-by-teacher
  dws edu-app diploma list-by-teacher --page 1 --limit 20 --tag "三好学生" -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber <= 0 {
				pageNumber = 1
			}
			input["pageNumber"] = pageNumber
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 20
			}
			input["pageSize"] = pageSize
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("tag"); strings.TrimSpace(v) != "" {
				input["tag"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "list_diploma_by_teacher", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaListByTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_diploma_by_teacher",
				CanonicalPath:  "edu-app.list_diploma_by_teacher",
				CLIPath:        "edu-app diploma list-by-teacher",
				PrimaryCLIPath: "edu-app diploma list-by-teacher",
			},
			Description: "查询老师创建的奖状列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_diploma_by_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询老师创建的奖状列表",
				UseWhen:      []string{"需要查询老师创建的奖状列表时"},
				AvoidWhen:    []string{"查询学生收到的奖状用 diploma list-by-student"},
				Examples: []string{
					"dws edu-app diploma list-by-teacher --page 1 --limit 20 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "page", Property: "input.pageNumber"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "status", Property: "input.status"},
				{Name: "tag", Property: "input.tag"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaGetCmd := &cobra.Command{
		Use:   "get",
		Short: "查询奖状详情",
		Long: `按奖状 ID 查询奖状详情信息。
--diploma-id 为奖状 ID（必填），--user-name 为可选参数。`,
		Example: `  dws edu-app diploma get --diploma-id 12345
  dws edu-app diploma get --diploma-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"diplomaId": diplomaID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "get_diploma", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaGetCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_diploma",
				CanonicalPath:  "edu-app.get_diploma",
				CLIPath:        "edu-app diploma get",
				PrimaryCLIPath: "edu-app diploma get",
			},
			Description: "查询奖状详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_diploma"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按奖状ID查询奖状详情",
				UseWhen:      []string{"需要查询指定奖状的详情时"},
				AvoidWhen:    []string{"查询奖状列表用 diploma list-by-teacher 或 diploma list-by-student"},
				Examples: []string{
					"dws edu-app diploma get --diploma-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaStatisticsCmd := &cobra.Command{
		Use:   "statistics",
		Short: "查询奖状阅读统计",
		Long: `按奖状 ID 查询奖状的阅读统计数据（按班级维度）。
返回每个班级的总人数、已读人数、未读人数。
--diploma-id 为奖状 ID（必填），--user-name 为可选参数。`,
		Example: `  dws edu-app diploma statistics --diploma-id 12345
  dws edu-app diploma statistics --diploma-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"diplomaId": diplomaID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_diploma_statistics", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaStatisticsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_diploma_statistics",
				CanonicalPath:  "edu-app.query_diploma_statistics",
				CLIPath:        "edu-app diploma statistics",
				PrimaryCLIPath: "edu-app diploma statistics",
			},
			Description: "查询奖状阅读统计",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_diploma_statistics"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按奖状ID查询奖状阅读统计数据",
				UseWhen:      []string{"需要查询奖状的阅读统计数据时"},
				AvoidWhen:    []string{"查询奖状接收详情用 diploma detail"},
				Examples: []string{
					"dws edu-app diploma statistics --diploma-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "查询奖状接收详情",
		Long: `按奖状 ID 查询奖状的接收详情列表。
返回每个学生的接收状态、接收时间、查看时间等信息。
--diploma-id 为奖状 ID（必填），--class-id、--user-name 为可选参数。`,
		Example: `  dws edu-app diploma detail --diploma-id 12345
  dws edu-app diploma detail --diploma-id 12345 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"diplomaId": diplomaID,
			}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_diploma_detail", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_diploma_detail",
				CanonicalPath:  "edu-app.query_diploma_detail",
				CLIPath:        "edu-app diploma detail",
				PrimaryCLIPath: "edu-app diploma detail",
			},
			Description: "查询奖状接收详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_diploma_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按奖状ID查询奖状接收详情列表",
				UseWhen:      []string{"需要查询奖状的接收详情时"},
				AvoidWhen:    []string{"查询奖状阅读统计用 diploma statistics；查询学生奖状详情用 diploma student-detail"},
				Examples: []string{
					"dws edu-app diploma detail --diploma-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaListByStudentCmd := &cobra.Command{
		Use:   "list-by-student",
		Short: "查询学生收到的奖状列表",
		Long: `查询指定学生收到的奖状列表，支持分页。
--student-id 和 --class-id 为必填参数。`,
		Example: `  dws edu-app diploma list-by-student --student-id uid1 --class-id 12345
  dws edu-app diploma list-by-student --student-id uid1 --class-id 12345 --page 1 --limit 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"studentId": studentID,
				"classId":   classID,
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber <= 0 {
				pageNumber = 1
			}
			input["pageNumber"] = pageNumber
			pageSize, _ := cmd.Flags().GetInt64("limit")
			if pageSize <= 0 {
				pageSize = 20
			}
			input["pageSize"] = pageSize
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "list_diploma_by_student", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaListByStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_diploma_by_student",
				CanonicalPath:  "edu-app.list_diploma_by_student",
				CLIPath:        "edu-app diploma list-by-student",
				PrimaryCLIPath: "edu-app diploma list-by-student",
			},
			Description: "查询学生收到的奖状列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_diploma_by_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询学生收到的奖状列表",
				UseWhen:      []string{"需要查询指定学生收到的奖状列表时"},
				AvoidWhen:    []string{"查询老师创建的奖状用 diploma list-by-teacher"},
				Examples: []string{
					"dws edu-app diploma list-by-student --student-id uid1 --class-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "limit", Property: "input.pageSize"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaStudentDetailCmd := &cobra.Command{
		Use:   "student-detail",
		Short: "查询学生奖状接收详情",
		Long: `按奖状 ID 和学生 ID 查询该学生的奖状接收详情。
--diploma-id、--student-id、--class-id 为必填参数。`,
		Example: `  dws edu-app diploma student-detail --diploma-id 12345 --student-id uid1 --class-id 67890
  dws edu-app diploma student-detail --diploma-id 12345 --student-id uid1 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"diplomaId": diplomaID,
				"studentId": studentID,
				"classId":   classID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_student_diploma_detail", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaStudentDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_student_diploma_detail",
				CanonicalPath:  "edu-app.query_student_diploma_detail",
				CLIPath:        "edu-app diploma student-detail",
				PrimaryCLIPath: "edu-app diploma student-detail",
			},
			Description: "查询学生奖状接收详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_student_diploma_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按奖状ID和学生ID查询学生奖状接收详情",
				UseWhen:      []string{"需要查询指定学生的奖状接收详情时"},
				AvoidWhen:    []string{"查询奖状接收详情列表用 diploma detail"},
				Examples: []string{
					"dws edu-app diploma student-detail --diploma-id 12345 --student-id uid1 --class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// homework 子命令组 — 作业
	// ════════════════════════════════════════════════════════════

	homeworkCmd := &cobra.Command{Use: "homework", Short: "作业管理", RunE: groupRunE}

	homeworkCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "创建并发布作业",
		Long: `创建并发布一份作业。
--identifier 为幂等标识（必填），建议格式：orgId-staffId-UUID，重复提交同一 identifier 不会重复创建。
--hw-content 为作业内容（必填），其余参数均为可选。`,
		Example: `  dws edu-app homework create --identifier org1-staff1-uuid --hw-content "完成课后练习第三题"
  dws edu-app homework create --identifier org1-staff1-uuid --hw-title "数学作业" --hw-content "完成练习册" --class-ids 12345,67890 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			identifier, _ := cmd.Flags().GetString("identifier")
			identifier = strings.TrimSpace(identifier)
			if identifier == "" {
				return fmt.Errorf("--identifier 为必填参数")
			}
			hwContent, _ := cmd.Flags().GetString("hw-content")
			hwContent = strings.TrimSpace(hwContent)
			if hwContent == "" {
				return fmt.Errorf("--hw-content 为必填参数")
			}
			input := map[string]any{
				"identifier": identifier,
				"hwContent":  hwContent,
			}
			if v, _ := cmd.Flags().GetString("hw-title"); strings.TrimSpace(v) != "" {
				input["hwTitle"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("hw-photo"); strings.TrimSpace(v) != "" {
				input["hwPhoto"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("hw-media"); strings.TrimSpace(v) != "" {
				input["hwMedia"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("hw-video"); strings.TrimSpace(v) != "" {
				input["hwVideo"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("class-ids"); strings.TrimSpace(v) != "" {
				input["classIds"] = eduAppParseCSV(v)
			}
			if v, _ := cmd.Flags().GetString("class-names"); strings.TrimSpace(v) != "" {
				input["classNames"] = eduAppParseCSV(v)
			}
			if v, _ := cmd.Flags().GetString("class-selected-students"); strings.TrimSpace(v) != "" {
				var obj any
				if err := json.Unmarshal([]byte(v), &obj); err != nil {
					return fmt.Errorf("--class-selected-students JSON 格式错误: %w", err)
				}
				input["classSelectedStudents"] = obj
			}
			if v, _ := cmd.Flags().GetString("feedback"); strings.TrimSpace(v) != "" {
				input["feedback"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("hw-deadline"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--hw-deadline 须为整数: %w", err)
				}
				input["hwDeadline"] = n
			}
			if v, _ := cmd.Flags().GetString("hw-deadline-open"); strings.TrimSpace(v) != "" {
				input["hwDeadlineOpen"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("hw-deadline-setting"); strings.TrimSpace(v) != "" {
				input["hwDeadlineSetting"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("submit-types"); strings.TrimSpace(v) != "" {
				input["submitTypes"] = eduAppParseCSV(v)
			}
			if v, _ := cmd.Flags().GetString("hw-type"); strings.TrimSpace(v) != "" {
				input["hwType"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("target-role"); strings.TrimSpace(v) != "" {
				input["targetRole"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("publish-type"); strings.TrimSpace(v) != "" {
				input["publishType"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("biz-code"); strings.TrimSpace(v) != "" {
				input["bizCode"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("scheduled-release"); strings.TrimSpace(v) != "" {
				input["scheduledRelease"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("task-plan-duration"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--task-plan-duration 须为整数: %w", err)
				}
				input["taskPlanDuration"] = n
			}
			if v, _ := cmd.Flags().GetString("attributes"); strings.TrimSpace(v) != "" {
				var obj any
				if err := json.Unmarshal([]byte(v), &obj); err != nil {
					return fmt.Errorf("--attributes JSON 格式错误: %w", err)
				}
				input["attributes"] = obj
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "create_homework", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkCreateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "create_homework",
				CanonicalPath:  "edu-app.create_homework",
				CLIPath:        "edu-app homework create",
				PrimaryCLIPath: "edu-app homework create",
			},
			Description: "创建并发布作业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "create_homework"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建并发布一份作业",
				UseWhen:      []string{"需要创建并发布作业时"},
				AvoidWhen:    []string{"删除作业用 homework delete；查询作业用 homework get"},
				Examples: []string{
					"dws edu-app homework create --identifier org1-staff1-uuid --hw-content \"完成课后练习第三题\" --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "identifier", Property: "input.identifier", Required: boolPtr(true)},
				{Name: "hw-content", Property: "input.hwContent", Required: boolPtr(true)},
				{Name: "hw-title", Property: "input.hwTitle"},
				{Name: "hw-photo", Property: "input.hwPhoto"},
				{Name: "hw-media", Property: "input.hwMedia"},
				{Name: "hw-video", Property: "input.hwVideo"},
				{Name: "class-ids", Property: "input.classIds"},
				{Name: "class-names", Property: "input.classNames"},
				{Name: "class-selected-students", Property: "input.classSelectedStudents"},
				{Name: "feedback", Property: "input.feedback"},
				{Name: "hw-deadline", Property: "input.hwDeadline"},
				{Name: "hw-deadline-open", Property: "input.hwDeadlineOpen"},
				{Name: "hw-deadline-setting", Property: "input.hwDeadlineSetting"},
				{Name: "submit-types", Property: "input.submitTypes"},
				{Name: "hw-type", Property: "input.hwType"},
				{Name: "target-role", Property: "input.targetRole"},
				{Name: "publish-type", Property: "input.publishType"},
				{Name: "biz-code", Property: "input.bizCode"},
				{Name: "scheduled-release", Property: "input.scheduledRelease"},
				{Name: "task-plan-duration", Property: "input.taskPlanDuration"},
				{Name: "attributes", Property: "input.attributes"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "删除作业",
		Long: `删除指定的作业。
返回删除结果：success 表示是否删除成功。`,
		Example: `  dws edu-app homework delete --homework-id 12345
  dws edu-app homework delete --homework-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"homeworkId": homeworkID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "delete_homework", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkDeleteCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "delete_homework",
				CanonicalPath:  "edu-app.delete_homework",
				CLIPath:        "edu-app homework delete",
				PrimaryCLIPath: "edu-app homework delete",
			},
			Description: "删除作业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "delete_homework"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除指定的作业",
				UseWhen:      []string{"需要删除指定作业时"},
				AvoidWhen:    []string{"查询作业详情用 homework get；创建作业用 homework create"},
				Examples: []string{
					"dws edu-app homework delete --homework-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkSubmitCmd := &cobra.Command{
		Use:   "submit",
		Short: "提交作业",
		Long: `提交指定作业内容。
--hw-content-detail-id 为作业内容详情 ID（必填），其余参数均为可选。`,
		Example: `  dws edu-app homework submit --hw-content-detail-id 12345
  dws edu-app homework submit --hw-content-detail-id 12345 --homework-id 67890 --content "已完成" --photo "https://example.com/img.jpg" -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			hwContentDetailID, err := eduAppRequiredIntFlag(cmd, "hw-content-detail-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"hwContentDetailId": hwContentDetailID,
			}
			if v, _ := cmd.Flags().GetString("homework-id"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--homework-id 须为整数: %w", err)
				}
				input["homeworkId"] = n
			}
			if v, _ := cmd.Flags().GetString("student-id"); strings.TrimSpace(v) != "" {
				input["studentId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("content"); strings.TrimSpace(v) != "" {
				input["content"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("photo"); strings.TrimSpace(v) != "" {
				input["photo"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("media"); strings.TrimSpace(v) != "" {
				input["media"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("video"); strings.TrimSpace(v) != "" {
				input["video"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "submit_homework", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkSubmitCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "submit_homework",
				CanonicalPath:  "edu-app.submit_homework",
				CLIPath:        "edu-app homework submit",
				PrimaryCLIPath: "edu-app homework submit",
			},
			Description: "提交作业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "submit_homework"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "提交指定作业内容",
				UseWhen:      []string{"需要提交作业时"},
				AvoidWhen:    []string{"查询作业详情用 homework get；创建作业用 homework create"},
				Examples: []string{
					"dws edu-app homework submit --hw-content-detail-id 12345 --content \"已完成\" --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "hw-content-detail-id", Property: "input.hwContentDetailId", Required: boolPtr(true)},
				{Name: "homework-id", Property: "input.homeworkId"},
				{Name: "student-id", Property: "input.studentId"},
				{Name: "class-id", Property: "input.classId"},
				{Name: "content", Property: "input.content"},
				{Name: "photo", Property: "input.photo"},
				{Name: "media", Property: "input.media"},
				{Name: "video", Property: "input.video"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkGetCmd := &cobra.Command{
		Use:   "get",
		Short: "查询作业详情",
		Long: `按作业 ID 查询作业详情。
--homework-id 为作业 ID（必填），--user-name 为可选参数。`,
		Example: `  dws edu-app homework get --homework-id 12345
  dws edu-app homework get --homework-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"homeworkId": homeworkID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "get_homework", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkGetCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "get_homework",
				CanonicalPath:  "edu-app.get_homework",
				CLIPath:        "edu-app homework get",
				PrimaryCLIPath: "edu-app homework get",
			},
			Description: "查询作业详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "get_homework"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按作业ID查询作业详情",
				UseWhen:      []string{"需要查询指定作业的详情时"},
				AvoidWhen:    []string{"查询作业列表用 homework list-by-teacher 或 homework list-by-student"},
				Examples: []string{
					"dws edu-app homework get --homework-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkClassByHomeworkCmd := &cobra.Command{
		Use:   "class-by-homework",
		Short: "查询作业的班级提交情况",
		Long: `按作业 ID 查询作业的班级列表及学生提交明细。
--homework-id 为作业 ID（必填），--class-id、--user-name 为可选参数。`,
		Example: `  dws edu-app homework class-by-homework --homework-id 12345
  dws edu-app homework class-by-homework --homework-id 12345 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"homeworkId": homeworkID,
			}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_class_by_homework", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkClassByHomeworkCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_class_by_homework",
				CanonicalPath:  "edu-app.query_class_by_homework",
				CLIPath:        "edu-app homework class-by-homework",
				PrimaryCLIPath: "edu-app homework class-by-homework",
			},
			Description: "查询作业的班级提交情况",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_class_by_homework"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按作业ID查询作业的班级提交情况",
				UseWhen:      []string{"需要查询作业的班级列表及学生提交明细时"},
				AvoidWhen:    []string{"查询班级作业详情用 homework class-detail"},
				Examples: []string{
					"dws edu-app homework class-by-homework --homework-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkClassDetailCmd := &cobra.Command{
		Use:   "class-detail",
		Short: "查询班级作业详情",
		Long: `按作业 ID 和班级 ID 查询班级作业的提交明细。
--homework-id、--class-id、--user-name 均为必填参数。`,
		Example: `  dws edu-app homework class-detail --homework-id 12345 --class-id 67890 --user-name 张三
  dws edu-app homework class-detail --homework-id 12345 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			userName, _ := cmd.Flags().GetString("user-name")
			userName = strings.TrimSpace(userName)
			if userName == "" {
				return fmt.Errorf("--user-name 为必填参数")
			}
			input := map[string]any{
				"homeworkId": homeworkID,
				"classId":    classID,
				"userName":   userName,
			}
			return callMCPToolOnServer("edu-app", "query_class_homework_detail", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkClassDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_class_homework_detail",
				CanonicalPath:  "edu-app.query_class_homework_detail",
				CLIPath:        "edu-app homework class-detail",
				PrimaryCLIPath: "edu-app homework class-detail",
			},
			Description: "查询班级作业详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_class_homework_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按作业ID和班级ID查询班级作业提交明细",
				UseWhen:      []string{"需要查询班级作业的提交明细时"},
				AvoidWhen:    []string{"查询作业班级列表用 homework class-by-homework"},
				Examples: []string{
					"dws edu-app homework class-detail --homework-id 12345 --class-id 67890 --user-name 张三 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName", Required: boolPtr(true)},
			},
		},
	})

	homeworkSubmitStatisticsCmd := &cobra.Command{
		Use:   "submit-statistics",
		Short: "查询作业提交统计",
		Long: `按作业 ID 和班级 ID 查询作业提交统计（已提交人数、已批改人数）。
--homework-id、--class-id 为必填参数，--user-name 为可选参数。`,
		Example: `  dws edu-app homework submit-statistics --homework-id 12345 --class-id 67890
  dws edu-app homework submit-statistics --homework-id 12345 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"homeworkId": homeworkID,
				"classId":    classID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_submit_statistics", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkSubmitStatisticsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_submit_statistics",
				CanonicalPath:  "edu-app.query_submit_statistics",
				CLIPath:        "edu-app homework submit-statistics",
				PrimaryCLIPath: "edu-app homework submit-statistics",
			},
			Description: "查询作业提交统计",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_submit_statistics"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按作业ID和班级ID查询作业提交统计",
				UseWhen:      []string{"需要查询作业提交统计时"},
				AvoidWhen:    []string{"查询班级作业详情用 homework class-detail"},
				Examples: []string{
					"dws edu-app homework submit-statistics --homework-id 12345 --class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkListByStudentCmd := &cobra.Command{
		Use:   "list-by-student",
		Short: "查询学生作业列表",
		Long: `按学生 ID 和班级 ID 查询学生收到的作业列表。
--student-id、--class-id、--user-name 为必填参数，--status、--page、--page-size 为可选参数。`,
		Example: `  dws edu-app homework list-by-student --student-id uid1 --class-id 67890 --user-name 张三
  dws edu-app homework list-by-student --student-id uid1 --class-id 67890 --user-name 张三 --status FINISHED --page 1 --page-size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			userName, _ := cmd.Flags().GetString("user-name")
			userName = strings.TrimSpace(userName)
			if userName == "" {
				return fmt.Errorf("--user-name 为必填参数")
			}
			input := map[string]any{
				"studentId": studentID,
				"classId":   classID,
				"userName":  userName,
			}
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber > 0 {
				input["pageNumber"] = pageNumber
			} else {
				input["pageNumber"] = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("page-size")
			if pageSize > 0 {
				input["pageSize"] = pageSize
			} else {
				input["pageSize"] = 20
			}
			return callMCPToolOnServer("edu-app", "list_homework_by_student", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkListByStudentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_homework_by_student",
				CanonicalPath:  "edu-app.list_homework_by_student",
				CLIPath:        "edu-app homework list-by-student",
				PrimaryCLIPath: "edu-app homework list-by-student",
			},
			Description: "查询学生作业列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_homework_by_student"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按学生ID和班级ID查询学生收到的作业列表",
				UseWhen:      []string{"需要查询学生收到的作业列表时"},
				AvoidWhen:    []string{"查询老师作业列表用 homework list-by-teacher"},
				Examples: []string{
					"dws edu-app homework list-by-student --student-id uid1 --class-id 67890 --user-name 张三 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName", Required: boolPtr(true)},
				{Name: "status", Property: "input.status"},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "page-size", Property: "input.pageSize"},
			},
		},
	})

	homeworkStudentDetailCmd := &cobra.Command{
		Use:   "student-detail",
		Short: "查询学生作业详情",
		Long: `按作业 ID、学生 ID 和班级 ID 查询单个学生的作业提交明细。
--homework-id、--student-id、--class-id 为必填参数，--user-name 为可选参数。`,
		Example: `  dws edu-app homework student-detail --homework-id 12345 --student-id uid1 --class-id 67890
  dws edu-app homework student-detail --homework-id 12345 --student-id uid1 --class-id 67890 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeworkID, err := eduAppRequiredIntFlag(cmd, "homework-id")
			if err != nil {
				return err
			}
			studentID, _ := cmd.Flags().GetString("student-id")
			studentID = strings.TrimSpace(studentID)
			if studentID == "" {
				return fmt.Errorf("--student-id 为必填参数")
			}
			classID, _ := cmd.Flags().GetString("class-id")
			classID = strings.TrimSpace(classID)
			if classID == "" {
				return fmt.Errorf("--class-id 为必填参数")
			}
			input := map[string]any{
				"homeworkId": homeworkID,
				"studentId":  studentID,
				"classId":    classID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "query_student_homework_detail", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkStudentDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "query_student_homework_detail",
				CanonicalPath:  "edu-app.query_student_homework_detail",
				CLIPath:        "edu-app homework student-detail",
				PrimaryCLIPath: "edu-app homework student-detail",
			},
			Description: "查询学生作业详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "query_student_homework_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "按作业ID和学生ID查询学生作业提交明细",
				UseWhen:      []string{"需要查询单个学生的作业提交明细时"},
				AvoidWhen:    []string{"查询班级作业详情用 homework class-detail"},
				Examples: []string{
					"dws edu-app homework student-detail --homework-id 12345 --student-id uid1 --class-id 67890 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "homework-id", Property: "input.homeworkId", Required: boolPtr(true)},
				{Name: "student-id", Property: "input.studentId", Required: boolPtr(true)},
				{Name: "class-id", Property: "input.classId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	homeworkListByTeacherCmd := &cobra.Command{
		Use:   "list-by-teacher",
		Short: "查询老师作业列表",
		Long: `查询老师发布的家校作业列表，支持分页和筛选。
所有参数均为可选，--class-id 按班级筛选，--type 按作业类型筛选，--status 按作业状态筛选。`,
		Example: `  dws edu-app homework list-by-teacher
  dws edu-app homework list-by-teacher --class-id 67890 --type HOMEWORK --status FINISHED --page 1 --page-size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("class-id"); strings.TrimSpace(v) != "" {
				input["classId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("type"); strings.TrimSpace(v) != "" {
				input["type"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("status"); strings.TrimSpace(v) != "" {
				input["status"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			pageNumber, _ := cmd.Flags().GetInt64("page")
			if pageNumber > 0 {
				input["pageNumber"] = pageNumber
			} else {
				input["pageNumber"] = 1
			}
			pageSize, _ := cmd.Flags().GetInt64("page-size")
			if pageSize > 0 {
				input["pageSize"] = pageSize
			} else {
				input["pageSize"] = 20
			}
			return callMCPToolOnServer("edu-app", "list_homework_by_teacher", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkListByTeacherCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "list_homework_by_teacher",
				CanonicalPath:  "edu-app.list_homework_by_teacher",
				CLIPath:        "edu-app homework list-by-teacher",
				PrimaryCLIPath: "edu-app homework list-by-teacher",
			},
			Description: "查询老师作业列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "list_homework_by_teacher"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询老师发布的家校作业列表",
				UseWhen:      []string{"需要查询老师发布的作业列表时"},
				AvoidWhen:    []string{"查询学生作业列表用 homework list-by-student"},
				Examples: []string{
					"dws edu-app homework list-by-teacher --class-id 67890 --page 1 --page-size 20 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "class-id", Property: "input.classId"},
				{Name: "type", Property: "input.type"},
				{Name: "status", Property: "input.status"},
				{Name: "user-name", Property: "input.userName"},
				{Name: "page", Property: "input.pageNumber"},
				{Name: "page-size", Property: "input.pageSize"},
			},
		},
	})

	homeworkCreateCommentCmd := &cobra.Command{
		Use:   "create-comment",
		Short: "创建作业评语",
		Long: `为指定作业创建评语。
--comment、--hw-content-detail-id 为必填参数，--homework-id、--student-id、--photo、--video、--media、--user-name 为可选参数。`,
		Example: `  dws edu-app homework create-comment --comment "做得很好" --hw-content-detail-id 12345
  dws edu-app homework create-comment --comment "继续努力" --hw-content-detail-id 12345 --homework-id 67890 --student-id uid1 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			comment, _ := cmd.Flags().GetString("comment")
			comment = strings.TrimSpace(comment)
			if comment == "" {
				return fmt.Errorf("--comment 为必填参数")
			}
			hwContentDetailID, err := eduAppRequiredIntFlag(cmd, "hw-content-detail-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"comment":           comment,
				"hwContentDetailId": hwContentDetailID,
			}
			if cmd.Flags().Changed("homework-id") {
				v, _ := cmd.Flags().GetInt64("homework-id")
				input["homeworkId"] = v
			}
			if v, _ := cmd.Flags().GetString("student-id"); strings.TrimSpace(v) != "" {
				input["studentId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("photo"); strings.TrimSpace(v) != "" {
				input["photo"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("video"); strings.TrimSpace(v) != "" {
				input["video"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("media"); strings.TrimSpace(v) != "" {
				input["media"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "create_comment", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(homeworkCreateCommentCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "create_comment",
				CanonicalPath:  "edu-app.create_comment",
				CLIPath:        "edu-app homework create-comment",
				PrimaryCLIPath: "edu-app homework create-comment",
			},
			Description: "创建作业评语",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "create_comment"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "为指定作业创建评语",
				UseWhen:      []string{"需要为作业创建评语时"},
				AvoidWhen:    []string{"提交作业用 homework submit；查询作业用 homework get"},
				Examples: []string{
					"dws edu-app homework create-comment --comment \"做得很好\" --hw-content-detail-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "comment", Property: "input.comment", Required: boolPtr(true)},
				{Name: "hw-content-detail-id", Property: "input.hwContentDetailId", Required: boolPtr(true)},
				{Name: "homework-id", Property: "input.homeworkId"},
				{Name: "student-id", Property: "input.studentId"},
				{Name: "photo", Property: "input.photo"},
				{Name: "video", Property: "input.video"},
				{Name: "media", Property: "input.media"},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	diplomaDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "删除奖状",
		Long: `删除指定的奖状。
返回删除结果：success 表示是否删除成功。`,
		Example: `  dws edu-app diploma delete --diploma-id 12345
  dws edu-app diploma delete --diploma-id 12345 --user-name 张三 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			diplomaID, err := eduAppRequiredIntFlag(cmd, "diploma-id")
			if err != nil {
				return err
			}
			input := map[string]any{
				"diplomaId": diplomaID,
			}
			if v, _ := cmd.Flags().GetString("user-name"); strings.TrimSpace(v) != "" {
				input["userName"] = strings.TrimSpace(v)
			}
			return callMCPToolOnServer("edu-app", "delete_diploma", map[string]any{
				"input": input,
			})
		},
	}
	DeclareLeafMetadata(diplomaDeleteCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "edu-app",
				Name:           "delete_diploma",
				CanonicalPath:  "edu-app.delete_diploma",
				CLIPath:        "edu-app diploma delete",
				PrimaryCLIPath: "edu-app diploma delete",
			},
			Description: "删除奖状",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "edu-app", RPCName: "delete_diploma"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除指定的奖状",
				UseWhen:      []string{"需要删除指定奖状时"},
				AvoidWhen:    []string{"查询奖状详情用 diploma get；创建奖状用 diploma create"},
				Examples: []string{
					"dws edu-app diploma delete --diploma-id 12345 --format json",
				},
			},
			Parameters: []contract.ParamDecl{
				{Name: "diploma-id", Property: "input.diplomaId", Required: boolPtr(true)},
				{Name: "user-name", Property: "input.userName"},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// flags + 构建命令树
	// ════════════════════════════════════════════════════════════

	// message flags
	messageSummaryListCmd.Flags().String("class-id", "", "班级 ID（必填）")
	messageSummaryListCmd.Flags().String("cid", "", "群会话 ID（必填）")
	messageSummaryListCmd.Flags().String("target-role", "", "用户角色：guardian / student（必填）")
	messageSummaryListCmd.Flags().String("status", "", "任务状态：0-未处理 1-已处理（必填）")

	// task flags
	taskPublishListCmd.Flags().Int64("cursor", 0, "分页游标，首次不传")
	taskPublishListCmd.Flags().Int64("limit", 0, "每页条数")
	taskPublishListCmd.Flags().Bool("need-statistic", false, "是否返回任务完成统计")
	taskPublishListCmd.Flags().String("task-sources", "", "任务来源，逗号分隔（EDU_HOMEWORK/EDU_CARD/EDU_NOTICE/EDU_SR/EDU_DIPLOMA）")

	taskAllListCmd.Flags().String("biz-id", "", "班级 ID（必填）")
	taskAllListCmd.Flags().Int64("cursor", 0, "分页游标，首次不传")
	taskAllListCmd.Flags().Int64("limit", 0, "每页条数")
	taskAllListCmd.Flags().Bool("need-statistic", false, "是否返回任务完成统计")
	taskAllListCmd.Flags().String("task-sources", "", "任务来源，逗号分隔")

	taskStudentListCmd.Flags().String("students", "", "学生列表 JSON 数组（必填）")
	taskStudentListCmd.Flags().Bool("query-all", false, "是否查询全部（含已完成），默认 false")
	taskStudentListCmd.Flags().String("cursor", "", "分页游标")
	taskStudentListCmd.Flags().String("limit", "", "每页条数")
	taskStudentListCmd.Flags().String("task-sources", "", "任务来源，逗号分隔")

	// report flags
	reportGetCmd.Flags().String("ids", "", "成绩单 ID 列表，逗号分隔（必填）")

	reportByTeacherCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	reportByTeacherCmd.Flags().Int64("limit", 20, "每页大小")
	reportByTeacherCmd.Flags().Int64("status", 0, "成绩单状态：0-未发布 1-已发布")

	reportByClassCmd.Flags().String("report-id", "", "成绩单 ID（必填）")
	reportByClassCmd.Flags().String("class-id", "", "班级 ID（必填）")
	reportByClassCmd.Flags().String("student-ids", "", "学生 userId 列表，逗号分隔（可选）")

	reportByStudentListCmd.Flags().String("class-id", "", "班级 ID（必填）")
	reportByStudentListCmd.Flags().String("student-id", "", "学生 userId（必填）")
	reportByStudentListCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	reportByStudentListCmd.Flags().Int64("limit", 20, "每页大小")

	reportByStudentDetailCmd.Flags().String("report-id", "", "成绩单 ID（必填）")
	reportByStudentDetailCmd.Flags().String("student-id", "", "学生 userId（必填）")
	reportByStudentDetailCmd.Flags().String("class-id", "", "班级 ID（必填）")

	// notice flags
	noticeConfirmCmd.Flags().String("notice-id", "", "通知 ID（必填）")
	noticeConfirmCmd.Flags().String("student-id", "", "学生 userId（必填）")
	noticeConfirmCmd.Flags().String("device-id", "", "设备 ID（可选）")
	noticeConfirmCmd.Flags().String("parent-name", "", "确认操作人名字（可选）")
	noticeConfirmCmd.Flags().Bool("update-sign", false, "是否更新签名（可选）")

	noticeCreateCmd.Flags().String("identifer", "", "幂等字段，建议 orgId-staffId-UUID（必填）")
	noticeCreateCmd.Flags().String("content", "", "通知内容（必填）")
	noticeCreateCmd.Flags().String("title", "", "通知标题（可选）")
	noticeCreateCmd.Flags().String("class-ids", "", "班级 ID 列表，逗号分隔（可选）")
	noticeCreateCmd.Flags().String("class-names", "", "班级名称列表，逗号分隔（可选）")
	noticeCreateCmd.Flags().String("class-selected-students", "", "班级被选中学生 JSON 对象（可选）")
	noticeCreateCmd.Flags().String("type", "", "类型（可选）")
	noticeCreateCmd.Flags().String("scope", "", "范围（可选）")
	noticeCreateCmd.Flags().String("target-role", "", "目标角色（可选）")
	noticeCreateCmd.Flags().String("is-signed", "", "是否签收（可选）")
	noticeCreateCmd.Flags().String("photo", "", "图片 URL（可选）")
	noticeCreateCmd.Flags().String("media", "", "媒体 URL（可选）")
	noticeCreateCmd.Flags().String("audio", "", "音频 URL（可选）")
	noticeCreateCmd.Flags().Bool("send-ding", false, "是否发送钉钉（可选）")
	noticeCreateCmd.Flags().String("scheduled-release", "", "定时发送（可选）")
	noticeCreateCmd.Flags().Int64("notice-deadline", 0, "通知截止时间戳（可选）")
	noticeCreateCmd.Flags().String("notice-deadline-open", "", "是否开启通知截止时间（可选）")
	noticeCreateCmd.Flags().String("notice-deadline-setting", "", "通知截止时间设置（可选）")
	noticeCreateCmd.Flags().String("attributes", "", "扩展属性 JSON 对象（可选）")
	noticeCreateCmd.Flags().String("user-name", "", "用户名（可选）")

	noticeDeleteCmd.Flags().String("notice-id", "", "通知 ID（必填）")
	noticeDeleteCmd.Flags().String("user-name", "", "用户名（可选）")

	noticeListByTeacherCmd.Flags().String("class-id", "", "班级 ID（可选）")
	noticeListByTeacherCmd.Flags().String("type", "", "通知类型（可选）")
	noticeListByTeacherCmd.Flags().String("status", "", "通知状态筛选（可选）")
	noticeListByTeacherCmd.Flags().String("user-name", "", "用户名（可选）")
	noticeListByTeacherCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	noticeListByTeacherCmd.Flags().Int64("page-size", 20, "每页大小")

	noticeGetCmd.Flags().String("notice-id", "", "通知 ID（必填）")
	noticeGetCmd.Flags().String("user-name", "", "用户名（可选）")

	noticeConfirmStatusCmd.Flags().String("notice-id", "", "通知 ID（必填）")
	noticeConfirmStatusCmd.Flags().String("class-id", "", "班级 ID（必填）")
	noticeConfirmStatusCmd.Flags().String("status", "", "确认状态筛选（可选）")
	noticeConfirmStatusCmd.Flags().String("user-name", "", "用户名（可选）")
	noticeConfirmStatusCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	noticeConfirmStatusCmd.Flags().Int64("page-size", 20, "每页大小")

	noticeListByStudentCmd.Flags().String("student-id", "", "学生 ID（必填）")
	noticeListByStudentCmd.Flags().String("class-id", "", "班级 ID（必填）")
	noticeListByStudentCmd.Flags().String("status", "", "通知状态筛选（可选）")
	noticeListByStudentCmd.Flags().String("user-name", "", "用户名（可选）")
	noticeListByStudentCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	noticeListByStudentCmd.Flags().Int64("page-size", 20, "每页大小")

	// circle flags
	circlePostsCmd.Flags().String("class-id", "", "班级 ID（必填）")
	circlePostsCmd.Flags().String("student-id", "", "学生 userId（必填）")
	circlePostsCmd.Flags().String("target-role", "", "用户角色：guardian（家长）/ student（学生）（必填）")

	// card flags
	cardUpdateCmd.Flags().String("card-id", "", "打卡卡片 ID（必填）")
	cardUpdateCmd.Flags().String("identifier", "", "幂等标识，建议 orgId-staffId-UUID（必填）")
	cardUpdateCmd.Flags().String("title", "", "修改后的标题（与 --content 至少传一个）")
	cardUpdateCmd.Flags().String("content", "", "修改后的内容（与 --title 至少传一个）")
	cardUpdateCmd.Flags().Bool("should-send-update-msg", false, "是否发送更新通知消息，默认 false")

	cardEndCmd.Flags().String("card-id", "", "打卡卡片 ID（必填）")

	cardListCmd.Flags().String("status", "", "打卡状态：FINISH-已完结 UNFINISH-进行中（必填）")
	cardListCmd.Flags().Int64("class-id", 0, "班级 ID（可选）")
	cardListCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	cardListCmd.Flags().Int64("limit", 10, "每页大小")

	cardUserStatisticCmd.Flags().String("class-id", "", "班级 ID（必填）")
	cardUserStatisticCmd.Flags().String("task-code", "", "任务 code（必填）")
	cardUserStatisticCmd.Flags().String("card-id", "", "打卡卡片 ID（必填）")
	cardUserStatisticCmd.Flags().Bool("finish", false, "是否查询已完成人员，默认 false（查未完成）")
	cardUserStatisticCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	cardUserStatisticCmd.Flags().Int64("limit", 10, "每页大小")

	cardFinishInfoCmd.Flags().String("card-id", "", "打卡卡片 ID（必填）")
	cardFinishInfoCmd.Flags().String("card-biz-id", "", "打卡业务 ID（必填）")
	cardFinishInfoCmd.Flags().String("target-role", "", "目标角色：teacher / guardian / student / headmaster（可选，未传按 uid 真实身份自动推断）")
	cardFinishInfoCmd.Flags().String("student-id", "", "学生 ID（可选，targetRole 为 guardian 时用于指定查看某个孩子的进度）")

	// diploma flags
	diplomaCreateCmd.Flags().String("identifier", "", "幂等标识，建议 orgId-staffId-UUID（必填）")
	diplomaCreateCmd.Flags().String("content", "", "奖状内容（必填）")
	diplomaCreateCmd.Flags().String("user-name", "", "获奖人用户名（必填）")
	diplomaCreateCmd.Flags().String("title", "", "奖状标题（可选）")
	diplomaCreateCmd.Flags().String("unit-name", "", "颁发单位名称（可选）")
	diplomaCreateCmd.Flags().String("tag", "", "奖状标签（可选）")
	diplomaCreateCmd.Flags().String("photo", "", "奖状图片 URL（可选）")
	diplomaCreateCmd.Flags().String("publish-time", "", "颁发时间（可选）")
	diplomaCreateCmd.Flags().String("biz-code", "", "业务编码（可选）")
	diplomaCreateCmd.Flags().String("biz-category", "", "业务种类（可选）")
	diplomaCreateCmd.Flags().String("msg-type", "", "消息类型（可选）")
	diplomaCreateCmd.Flags().String("template-url", "", "模板 URL（可选）")
	diplomaCreateCmd.Flags().String("class-ids", "", "班级 ID 列表，逗号分隔（可选）")
	diplomaCreateCmd.Flags().String("select-class", "", "选择的班级及学生 JSON 数组（可选）")
	diplomaCreateCmd.Flags().String("attributes", "", "扩展属性 JSON 对象（可选）")

	diplomaReadCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaReadCmd.Flags().String("class-id", "", "班级 ID（可选）")
	diplomaReadCmd.Flags().String("student-id", "", "学生 ID（可选）")
	diplomaReadCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaListByTeacherCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	diplomaListByTeacherCmd.Flags().Int64("limit", 20, "每页大小")
	diplomaListByTeacherCmd.Flags().String("status", "", "奖状状态筛选（可选）")
	diplomaListByTeacherCmd.Flags().String("tag", "", "标签筛选（可选）")
	diplomaListByTeacherCmd.Flags().String("user-name", "", "获奖人用户名筛选（可选）")

	diplomaGetCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaGetCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaStatisticsCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaStatisticsCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaDetailCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaDetailCmd.Flags().String("class-id", "", "班级 ID（可选）")
	diplomaDetailCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaListByStudentCmd.Flags().String("student-id", "", "学生 ID（必填）")
	diplomaListByStudentCmd.Flags().String("class-id", "", "班级 ID（必填）")
	diplomaListByStudentCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	diplomaListByStudentCmd.Flags().Int64("limit", 20, "每页大小")
	diplomaListByStudentCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaStudentDetailCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaStudentDetailCmd.Flags().String("student-id", "", "学生 ID（必填）")
	diplomaStudentDetailCmd.Flags().String("class-id", "", "班级 ID（必填）")
	diplomaStudentDetailCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkCreateCmd.Flags().String("identifier", "", "幂等标识，建议 orgId-staffId-UUID（必填）")
	homeworkCreateCmd.Flags().String("hw-content", "", "作业内容（必填）")
	homeworkCreateCmd.Flags().String("hw-title", "", "作业标题（可选）")
	homeworkCreateCmd.Flags().String("hw-photo", "", "作业图片 URL（可选）")
	homeworkCreateCmd.Flags().String("hw-media", "", "作业视频 URL（可选）")
	homeworkCreateCmd.Flags().String("hw-video", "", "作业录音 URL（可选）")
	homeworkCreateCmd.Flags().String("class-ids", "", "班级 ID 列表，逗号分隔（可选）")
	homeworkCreateCmd.Flags().String("class-names", "", "班级名称列表，逗号分隔（可选）")
	homeworkCreateCmd.Flags().String("class-selected-students", "", "班级被选中学生 JSON 对象（可选）")
	homeworkCreateCmd.Flags().String("feedback", "", "反馈（可选）")
	homeworkCreateCmd.Flags().String("hw-deadline", "", "作业截止时间戳（可选）")
	homeworkCreateCmd.Flags().String("hw-deadline-open", "", "是否开启作业截止时间（可选）")
	homeworkCreateCmd.Flags().String("hw-deadline-setting", "", "作业截止时间设置（可选）")
	homeworkCreateCmd.Flags().String("submit-types", "", "提交类型列表，逗号分隔（可选）")
	homeworkCreateCmd.Flags().String("hw-type", "", "作业类型（可选）")
	homeworkCreateCmd.Flags().String("target-role", "", "目标角色（可选）")
	homeworkCreateCmd.Flags().String("publish-type", "", "发布类型（可选）")
	homeworkCreateCmd.Flags().String("biz-code", "", "业务代码（可选）")
	homeworkCreateCmd.Flags().String("scheduled-release", "", "定时发送（可选）")
	homeworkCreateCmd.Flags().String("task-plan-duration", "", "计划任务时长（可选）")
	homeworkCreateCmd.Flags().String("attributes", "", "扩展属性 JSON 对象（可选）")
	homeworkCreateCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkDeleteCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkDeleteCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkSubmitCmd.Flags().String("hw-content-detail-id", "", "作业内容详情 ID（必填）")
	homeworkSubmitCmd.Flags().String("homework-id", "", "作业 ID（可选）")
	homeworkSubmitCmd.Flags().String("student-id", "", "学生 ID（可选）")
	homeworkSubmitCmd.Flags().String("class-id", "", "班级 ID（可选）")
	homeworkSubmitCmd.Flags().String("content", "", "提交内容（可选）")
	homeworkSubmitCmd.Flags().String("photo", "", "提交图片 URL（可选）")
	homeworkSubmitCmd.Flags().String("media", "", "提交视频 URL（可选）")
	homeworkSubmitCmd.Flags().String("video", "", "提交录音 URL（可选）")
	homeworkSubmitCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkGetCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkGetCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkClassByHomeworkCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkClassByHomeworkCmd.Flags().String("class-id", "", "班级 ID（可选）")
	homeworkClassByHomeworkCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkClassDetailCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkClassDetailCmd.Flags().String("class-id", "", "班级 ID（必填）")
	homeworkClassDetailCmd.Flags().String("user-name", "", "用户名（必填）")

	homeworkSubmitStatisticsCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkSubmitStatisticsCmd.Flags().String("class-id", "", "班级 ID（必填）")
	homeworkSubmitStatisticsCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkListByStudentCmd.Flags().String("student-id", "", "学生 ID（必填）")
	homeworkListByStudentCmd.Flags().String("class-id", "", "班级 ID（必填）")
	homeworkListByStudentCmd.Flags().String("user-name", "", "用户名（必填）")
	homeworkListByStudentCmd.Flags().String("status", "", "提交状态筛选（可选）")
	homeworkListByStudentCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	homeworkListByStudentCmd.Flags().Int64("page-size", 20, "每页大小")

	homeworkStudentDetailCmd.Flags().String("homework-id", "", "作业 ID（必填）")
	homeworkStudentDetailCmd.Flags().String("student-id", "", "学生 ID（必填）")
	homeworkStudentDetailCmd.Flags().String("class-id", "", "班级 ID（必填）")
	homeworkStudentDetailCmd.Flags().String("user-name", "", "用户名（可选）")

	homeworkListByTeacherCmd.Flags().String("class-id", "", "班级 ID（可选）")
	homeworkListByTeacherCmd.Flags().String("type", "", "作业类型（可选）")
	homeworkListByTeacherCmd.Flags().String("status", "", "作业状态筛选（可选）")
	homeworkListByTeacherCmd.Flags().String("user-name", "", "用户名（可选）")
	homeworkListByTeacherCmd.Flags().Int64("page", 1, "页码，从 1 开始")
	homeworkListByTeacherCmd.Flags().Int64("page-size", 20, "每页大小")

	homeworkCreateCommentCmd.Flags().String("comment", "", "评语文字内容（必填）")
	homeworkCreateCommentCmd.Flags().String("hw-content-detail-id", "", "作业内容详情 ID（必填）")
	homeworkCreateCommentCmd.Flags().Int64("homework-id", 0, "作业 ID（可选）")
	homeworkCreateCommentCmd.Flags().String("student-id", "", "学生 ID（可选）")
	homeworkCreateCommentCmd.Flags().String("photo", "", "评语图片 URL（可选）")
	homeworkCreateCommentCmd.Flags().String("video", "", "评语录音 URL（可选）")
	homeworkCreateCommentCmd.Flags().String("media", "", "评语视频 URL（可选）")
	homeworkCreateCommentCmd.Flags().String("user-name", "", "用户名（可选）")

	diplomaDeleteCmd.Flags().String("diploma-id", "", "奖状 ID（必填）")
	diplomaDeleteCmd.Flags().String("user-name", "", "用户名（可选）")

	messageCmd.AddCommand(messageSummaryListCmd)
	taskCmd.AddCommand(taskPublishListCmd, taskAllListCmd, taskStudentListCmd)
	reportCmd.AddCommand(reportGetCmd, reportByTeacherCmd, reportByClassCmd, reportByStudentListCmd, reportByStudentDetailCmd)
	noticeCmd.AddCommand(noticeConfirmCmd, noticeCreateCmd, noticeDeleteCmd, noticeListByTeacherCmd, noticeGetCmd, noticeConfirmStatusCmd, noticeListByStudentCmd)
	circleCmd.AddCommand(circlePostsCmd)
	cardCmd.AddCommand(cardUpdateCmd, cardEndCmd, cardListCmd, cardUserStatisticCmd, cardFinishInfoCmd)
	homeworkCmd.AddCommand(homeworkCreateCmd, homeworkDeleteCmd, homeworkSubmitCmd, homeworkGetCmd, homeworkClassByHomeworkCmd, homeworkClassDetailCmd, homeworkSubmitStatisticsCmd, homeworkListByStudentCmd, homeworkStudentDetailCmd, homeworkListByTeacherCmd, homeworkCreateCommentCmd)
	diplomaCmd.AddCommand(diplomaCreateCmd, diplomaReadCmd, diplomaListByTeacherCmd, diplomaGetCmd, diplomaStatisticsCmd, diplomaDetailCmd, diplomaListByStudentCmd, diplomaStudentDetailCmd, diplomaDeleteCmd)

	root.AddCommand(messageCmd, taskCmd, reportCmd, noticeCmd, circleCmd, cardCmd, diplomaCmd, homeworkCmd)

	return root
}

// eduAppRequiredIntFlag extracts a required integer flag, returning an error
// if the flag is empty or not a valid integer.
func eduAppRequiredIntFlag(cmd *cobra.Command, name string) (int64, error) {
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

// eduAppParseCSV splits a comma-separated string into trimmed non-empty values.
func eduAppParseCSV(raw string) []string {
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

// eduAppParseIntCSV splits a comma-separated string into int64 values.
func eduAppParseIntCSV(raw string) ([]int64, error) {
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
