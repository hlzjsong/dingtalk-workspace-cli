package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd/contract"
)

// ──────────────────────────────────────────────────────────
// dws college-contact — 高校通讯录扩展
// 共 65 个工具，按 dept / employee / alumni / graduate / group 分组
// ──────────────────────────────────────────────────────────

func newCollegeContactCommand() *cobra.Command {
	contract.RegisterProductDecl(contract.ProductDecl{
		ID: "college-contact",
		Selection: contract.ProductSelectionDecl{
			AgentSummary: "高校通讯录：高校组织架构/院系部门管理/师生员工管理/校友/毕业年级/规则管理",
			UseWhen: []string{
				"用户要查询或管理钉钉高校通讯录，包括组织架构、部门、教职工、学生、校友、毕业年级等信息。",
			},
			AvoidWhen: []string{
				"家校通讯录用 edu-contact；家校群用 edu-group；家庭群用 edu-familygroup；家校应用用 edu-app。",
			},
		},
	})
	root := &cobra.Command{
		Use:    "college-contact",
		Short:  "高校通讯录",
		Long:   `钉钉高校通讯录扩展：查询高校组织架构/院系/师生信息等。`,
		Hidden: true,
		RunE:   groupRunE,
	}

	// ════════════════════════════════════════════════════════════
	// dept 子命令组 — 部门管理
	// ════════════════════════════════════════════════════════════

	deptCmd := &cobra.Command{Use: "dept", Short: "部门管理", RunE: groupRunE}

	getStandardStructureCmd := &cobra.Command{
		Use:   "get-standard-structure",
		Short: "查询高校标准架构信息",
		Long: `查询高校标准架构信息，返回组织 ID、行政架构部门 ID 及架构固定部门 ID 映射。
所有参数均可选：
  --dept-id   按部门 ID 过滤
  --staff-id  按员工 staffId 过滤
  --keyword   按关键词搜索
  --offset / --size  分页参数`,
		Example: `  dws college-contact dept get-standard-structure
  dws college-contact dept get-standard-structure --dept-id 12345
  dws college-contact dept get-standard-structure --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("dept-id"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-id 须为整数: %w", err)
				}
				input["deptId"] = n
			}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("keyword"); strings.TrimSpace(v) != "" {
				input["keyword"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "get_college_standard_structure", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(getStandardStructureCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_college_standard_structure",
				CanonicalPath:  "college-contact.get_college_standard_structure",
				CLIPath:        "college-contact dept get-standard-structure",
				PrimaryCLIPath: "college-contact dept get-standard-structure",
			},
			Description: "查询高校标准架构信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_college_standard_structure"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询高校标准架构信息",
				UseWhen:      []string{"需要查询高校标准架构信息时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept get-standard-structure --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	getDeptDetailCmd := &cobra.Command{
		Use:   "get-detail",
		Short: "查询部门详情",
		Long: `查询高校部门详情，返回组织 ID、部门 ID、部门类型、部门编码及行政架构部门 ID。
--dept-id 为必填，其余可选：
  --staff-id  按员工 staffId 过滤
  --keyword   按关键词搜索
  --offset / --size  分页参数（size 默认 20）`,
		Example: `  dws college-contact dept get-detail --dept-id 12345
  dws college-contact dept get-detail --dept-id 12345 --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input := map[string]any{"deptId": deptID}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("keyword"); strings.TrimSpace(v) != "" {
				input["keyword"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "get_college_dept_detail", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(getDeptDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_college_dept_detail",
				CanonicalPath:  "college-contact.get_college_dept_detail",
				CLIPath:        "college-contact dept get-detail",
				PrimaryCLIPath: "college-contact dept get-detail",
			},
			Description: "查询部门详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_college_dept_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询部门详情",
				UseWhen:      []string{"需要查询部门详情时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept get-detail --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	getDeptChainCmd := &cobra.Command{
		Use:   "get-chain",
		Short: "查询部门链",
		Long: `查询部门链信息，返回从根节点到当前部门的部门列表及部门行业化信息。
--dept-id 为必填，其余可选：
  --staff-id  按员工 staffId 过滤
  --keyword   按关键词搜索
  --offset / --size  分页参数（size 默认 20）`,
		Example: `  dws college-contact dept get-chain --dept-id 12345
  dws college-contact dept get-chain --dept-id 12345 --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input := map[string]any{"deptId": deptID}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("keyword"); strings.TrimSpace(v) != "" {
				input["keyword"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "get_college_dept_chain", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(getDeptChainCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_college_dept_chain",
				CanonicalPath:  "college-contact.get_college_dept_chain",
				CLIPath:        "college-contact dept get-chain",
				PrimaryCLIPath: "college-contact dept get-chain",
			},
			Description: "查询部门链",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_college_dept_chain"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询部门链",
				UseWhen:      []string{"需要查询部门链时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept get-chain --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	searchContactCmd := &cobra.Command{
		Use:   "search",
		Short: "搜索通讯录",
		Long: `按关键词搜索高校通讯录，返回匹配的人员、部门及角色集合。
--dept-id 与 --keyword 均为必填，其余可选：
  --staff-id  按员工 staffId 过滤
  --offset / --size  分页参数（size 默认 20）`,
		Example: `  dws college-contact dept search --dept-id 12345 --keyword 张三
  dws college-contact dept search --dept-id 12345 --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			keyword, _ := cmd.Flags().GetString("keyword")
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				return fmt.Errorf("--keyword 为必填参数")
			}
			input := map[string]any{"deptId": deptID, "keyword": keyword}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "search_college_contact", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(searchContactCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "search_college_contact",
				CanonicalPath:  "college-contact.search_college_contact",
				CLIPath:        "college-contact dept search",
				PrimaryCLIPath: "college-contact dept search",
			},
			Description: "搜索通讯录",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "search_college_contact"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "搜索通讯录",
				UseWhen:      []string{"需要搜索通讯录时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept search --dept-id 12345 --keyword 计算机"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	// ════════════════════════════════════════════════════════════
	// flags + 构建命令树
	// ════════════════════════════════════════════════════════════

	getStandardStructureCmd.Flags().String("dept-id", "", "部门 ID")
	getStandardStructureCmd.Flags().String("staff-id", "", "员工 staffId")
	getStandardStructureCmd.Flags().String("keyword", "", "搜索关键词")
	getStandardStructureCmd.Flags().String("offset", "", "分页偏移量")
	getStandardStructureCmd.Flags().String("size", "", "分页大小")

	getDeptDetailCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	getDeptDetailCmd.Flags().String("staff-id", "", "员工 staffId")
	getDeptDetailCmd.Flags().String("keyword", "", "搜索关键词")
	getDeptDetailCmd.Flags().String("offset", "", "分页偏移量")
	getDeptDetailCmd.Flags().String("size", "", "分页大小（默认 20）")

	getDeptChainCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	getDeptChainCmd.Flags().String("staff-id", "", "员工 staffId")
	getDeptChainCmd.Flags().String("keyword", "", "搜索关键词")
	getDeptChainCmd.Flags().String("offset", "", "分页偏移量")
	getDeptChainCmd.Flags().String("size", "", "分页大小（默认 20）")

	searchContactCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	searchContactCmd.Flags().String("keyword", "", "搜索关键词（必填）")
	searchContactCmd.Flags().String("staff-id", "", "员工 staffId")
	searchContactCmd.Flags().String("offset", "", "分页偏移量")
	searchContactCmd.Flags().String("size", "", "分页大小（默认 20）")

	createDeptCmd := &cobra.Command{
		Use:   "create",
		Short: "创建部门",
		Long: `创建高校部门，返回创建成功的部门 ID。
必填：
  --super-id          上级部门 ID
  --stru-dept-id      所属组织架构部门 ID
  --name              部门名称
  --dept-type         部门类型
  --create-dept-group 是否创建部门群（true/false）
可选：
  --dept-id     部门 ID
  --dept-code   部门编码
  --brief       简介
  --phone       电话`,
		Example: `  dws college-contact dept create --super-id 100 --stru-dept-id 200 --name 计算机学院 --dept-type college --create-dept-group true
  dws college-contact dept create --super-id 100 --stru-dept-id 200 --name 计算机学院 --dept-type college --create-dept-group false --brief "学院简介" --phone 010-12345678 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"super-id", "superId"},
				{"stru-dept-id", "struDeptId"},
			} {
				raw, _ := cmd.Flags().GetString(pair.flag)
				raw = strings.TrimSpace(raw)
				if raw == "" {
					return fmt.Errorf("--%s 为必填参数", pair.flag)
				}
				n, err := strconv.ParseInt(raw, 10, 64)
				if err != nil {
					return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
				}
				input[pair.key] = n
			}

			// 必填 string
			for _, flag := range []string{"name", "dept-type"} {
				v, _ := cmd.Flags().GetString(flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", flag)
				}
				key := flag
				if flag == "dept-type" {
					key = "deptType"
				}
				input[key] = v
			}

			// 必填 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			if !cmd.Flags().Changed("create-dept-group") {
				return fmt.Errorf("--create-dept-group 为必填参数（true/false）")
			}
			cdgRaw, _ := cmd.Flags().GetString("create-dept-group")
			createDeptGroup, err := strconv.ParseBool(strings.TrimSpace(cdgRaw))
			if err != nil {
				return fmt.Errorf("--create-dept-group 须为 true 或 false: %w", err)
			}
			input["createDeptGroup"] = createDeptGroup

			// 可选 int64
			if v, _ := cmd.Flags().GetString("dept-id"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-id 须为整数: %w", err)
				}
				input["deptId"] = n
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-code", "deptCode"},
				{"brief", "brief"},
				{"phone", "phone"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			return callMCPToolOnServer("college-contact", "create_college_contact_dept", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(createDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "create_college_contact_dept",
				CanonicalPath:  "college-contact.create_college_contact_dept",
				CLIPath:        "college-contact dept create",
				PrimaryCLIPath: "college-contact dept create",
			},
			Description: "创建部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "create_college_contact_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建部门",
				UseWhen:      []string{"需要创建部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept create --super-id 100 --stru-dept-id 200 --name 计算机学院 --dept-type college --create-dept-group true"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "super-id", Property: "input.superId", Required: boolPtr(true)},
				{Name: "stru-dept-id", Property: "input.struDeptId", Required: boolPtr(true)},
				{Name: "name", Property: "input.name", Required: boolPtr(true)},
				{Name: "dept-type", Property: "input.deptType", Required: boolPtr(true)},
				{Name: "create-dept-group", Property: "input.createDeptGroup", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "dept-code", Property: "input.deptCode", Required: boolPtr(false)},
				{Name: "brief", Property: "input.brief", Required: boolPtr(false)},
				{Name: "phone", Property: "input.phone", Required: boolPtr(false)},
			},
		},
	})

	createDeptCmd.Flags().String("super-id", "", "上级部门 ID（必填）")
	createDeptCmd.Flags().String("stru-dept-id", "", "所属组织架构部门 ID（必填）")
	createDeptCmd.Flags().String("name", "", "部门名称（必填）")
	createDeptCmd.Flags().String("dept-type", "", "部门类型（必填）")
	createDeptCmd.Flags().String("create-dept-group", "", "是否创建部门群（必填，true/false）")
	createDeptCmd.Flags().String("dept-id", "", "部门 ID")
	createDeptCmd.Flags().String("dept-code", "", "部门编码")
	createDeptCmd.Flags().String("brief", "", "简介")
	createDeptCmd.Flags().String("phone", "", "电话")

	updateDeptCmd := &cobra.Command{
		Use:   "update",
		Short: "更新部门",
		Long: `更新高校部门信息，返回操作是否成功。
必填：
  --dept-id     部门 ID
  --dept-type   部门类型
可选：
  --name              部门名称
  --stru-dept-id      所属组织架构部门 ID
  --super-id          上级部门 ID
  --dept-code         部门编码
  --brief             简介
  --phone             电话
  --create-dept-group 是否创建部门群（true/false）`,
		Example: `  dws college-contact dept update --dept-id 12345 --dept-type college
  dws college-contact dept update --dept-id 12345 --dept-type college --name 计算机学院 --brief 学院简介 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input["deptId"] = deptID

			// 必填 string
			v, _ := cmd.Flags().GetString("dept-type")
			v = strings.TrimSpace(v)
			if v == "" {
				return fmt.Errorf("--dept-type 为必填参数")
			}
			input["deptType"] = v

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"stru-dept-id", "struDeptId"},
				{"super-id", "superId"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			if cmd.Flags().Changed("create-dept-group") {
				cdgRaw, _ := cmd.Flags().GetString("create-dept-group")
				createDeptGroup, err := strconv.ParseBool(strings.TrimSpace(cdgRaw))
				if err != nil {
					return fmt.Errorf("--create-dept-group 须为 true 或 false: %w", err)
				}
				input["createDeptGroup"] = createDeptGroup
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"name", "name"},
				{"dept-code", "deptCode"},
				{"brief", "brief"},
				{"phone", "phone"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			return callMCPToolOnServer("college-contact", "update_college_contact_dept", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(updateDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "update_college_contact_dept",
				CanonicalPath:  "college-contact.update_college_contact_dept",
				CLIPath:        "college-contact dept update",
				PrimaryCLIPath: "college-contact dept update",
			},
			Description: "更新部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "update_college_contact_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "更新部门",
				UseWhen:      []string{"需要更新部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept update --dept-id 12345 --dept-type college"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "dept-type", Property: "input.deptType", Required: boolPtr(true)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "stru-dept-id", Property: "input.struDeptId", Required: boolPtr(false)},
				{Name: "super-id", Property: "input.superId", Required: boolPtr(false)},
				{Name: "dept-code", Property: "input.deptCode", Required: boolPtr(false)},
				{Name: "brief", Property: "input.brief", Required: boolPtr(false)},
				{Name: "phone", Property: "input.phone", Required: boolPtr(false)},
				{Name: "create-dept-group", Property: "input.createDeptGroup", Required: boolPtr(false)},
			},
		},
	})

	updateDeptCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	updateDeptCmd.Flags().String("dept-type", "", "部门类型（必填）")
	updateDeptCmd.Flags().String("name", "", "部门名称")
	updateDeptCmd.Flags().String("stru-dept-id", "", "所属组织架构部门 ID")
	updateDeptCmd.Flags().String("super-id", "", "上级部门 ID")
	updateDeptCmd.Flags().String("dept-code", "", "部门编码")
	updateDeptCmd.Flags().String("brief", "", "简介")
	updateDeptCmd.Flags().String("phone", "", "电话")
	updateDeptCmd.Flags().String("create-dept-group", "", "是否创建部门群（true/false）")

	deleteDeptCmd := &cobra.Command{
		Use:   "delete",
		Short: "删除部门",
		Long: `删除高校部门，返回操作是否成功。
--dept-id 为必填。

注意：删除操作不可逆。非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact dept delete --dept-id 12345 --dry-run
  dws college-contact dept delete --dept-id 12345 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			return callMCPToolOnServer("college-contact", "delete_college_contact_dept", map[string]any{
				"input": map[string]any{"deptId": deptID},
			})
		},
	}

	DeclareLeafMetadata(deleteDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_college_contact_dept",
				CanonicalPath:  "college-contact.delete_college_contact_dept",
				CLIPath:        "college-contact dept delete",
				PrimaryCLIPath: "college-contact dept delete",
			},
			Description: "删除部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_college_contact_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除部门",
				UseWhen:      []string{"需要删除部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept delete --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	deleteDeptCmd.Flags().String("dept-id", "", "部门 ID（必填）")

	batchUpdateDeptTypeCmd := &cobra.Command{
		Use:   "batch-update-type",
		Short: "批量修改部门类型",
		Long: `批量修改部门类型，返回操作是否成功。
必填：
  --dept-ids          部门 ID 列表，逗号分隔（如 1,2,3）
  --target-dept-type  目标部门类型`,
		Example: `  dws college-contact dept batch-update-type --dept-ids 100,200,300 --target-dept-type college
  dws college-contact dept batch-update-type --dept-ids 100,200 --target-dept-type department -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIDsRaw, _ := cmd.Flags().GetString("dept-ids")
			deptIDsRaw = strings.TrimSpace(deptIDsRaw)
			if deptIDsRaw == "" {
				return fmt.Errorf("--dept-ids 为必填参数")
			}
			var deptIDs []int64
			for _, part := range strings.Split(deptIDsRaw, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-ids 中 %q 须为整数: %w", part, err)
				}
				deptIDs = append(deptIDs, n)
			}
			if len(deptIDs) == 0 {
				return fmt.Errorf("--dept-ids 不能为空列表")
			}

			targetDeptType, _ := cmd.Flags().GetString("target-dept-type")
			targetDeptType = strings.TrimSpace(targetDeptType)
			if targetDeptType == "" {
				return fmt.Errorf("--target-dept-type 为必填参数")
			}

			input := map[string]any{
				"deptIds":        deptIDs,
				"targetDeptType": targetDeptType,
			}

			return callMCPToolOnServer("college-contact", "batch_update_dept_type", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(batchUpdateDeptTypeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "batch_update_dept_type",
				CanonicalPath:  "college-contact.batch_update_dept_type",
				CLIPath:        "college-contact dept batch-update-type",
				PrimaryCLIPath: "college-contact dept batch-update-type",
			},
			Description: "批量修改部门类型",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "batch_update_dept_type"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量修改部门类型",
				UseWhen:      []string{"需要批量修改部门类型时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept batch-update-type --dept-ids 12345 --target-dept-type college"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-ids", Property: "input.deptIds", Required: boolPtr(true)},
				{Name: "target-dept-type", Property: "input.targetDeptType", Required: boolPtr(true)},
			},
		},
	})

	batchUpdateDeptTypeCmd.Flags().String("dept-ids", "", "部门 ID 列表，逗号分隔（必填）")
	batchUpdateDeptTypeCmd.Flags().String("target-dept-type", "", "目标部门类型（必填）")

	listEmployeesCmd := &cobra.Command{
		Use:   "list-employees",
		Short: "查询部门员工列表",
		Long: `查询指定部门的员工列表，返回员工总数及员工列表。
必填：
  --dept-id  部门 ID
可选：
  --staff-id          员工 staffId
  --name              员工姓名
  --mobile            手机号
  --job-number        工号
  --main-dept-id      主部门 ID
  --login-id-type     登录类型
  --exclusive-account 是否独占账号（true/false）
  --send-active-sms   是否发送激活短信（true/false）
  --staff-ids         批量操作的 staffId 列表，逗号分隔
  --target-dept-id    变更目标部门 ID
  --offset / --size   分页参数
  --order-field       排序字段（如 job_number / order_in_dept）
  --ordering          排序顺序（asc / desc）`,
		Example: `  dws college-contact employee list-employees --dept-id 12345
  dws college-contact employee list-employees --dept-id 12345 --name 张三 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input["deptId"] = deptID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "list_employees", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(listEmployeesCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "list_employees",
				CanonicalPath:  "college-contact.list_employees",
				CLIPath:        "college-contact employee list-employees",
				PrimaryCLIPath: "college-contact employee list-employees",
			},
			Description: "查询部门员工列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "list_employees"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询部门员工列表",
				UseWhen:      []string{"需要查询部门员工列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee list-employees --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	listEmployeesCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	listEmployeesCmd.Flags().String("staff-id", "", "员工 staffId")
	listEmployeesCmd.Flags().String("name", "", "员工姓名")
	listEmployeesCmd.Flags().String("mobile", "", "手机号")
	listEmployeesCmd.Flags().String("job-number", "", "工号")
	listEmployeesCmd.Flags().String("main-dept-id", "", "主部门 ID")
	listEmployeesCmd.Flags().String("login-id-type", "", "登录类型")
	listEmployeesCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	listEmployeesCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	listEmployeesCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	listEmployeesCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	listEmployeesCmd.Flags().String("offset", "", "分页偏移量")
	listEmployeesCmd.Flags().String("size", "", "分页大小")
	listEmployeesCmd.Flags().String("order-field", "", "排序字段（如 job_number / order_in_dept）")
	listEmployeesCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	listUnacceptedEmployeesCmd := &cobra.Command{
		Use:   "list-unaccepted",
		Short: "查询未接受邀请的员工列表",
		Long: `查询指定部门下未接受邀请的员工列表，返回员工总数及员工列表。
必填：
  --dept-id  部门 ID
可选：
  --staff-id          员工 staffId
  --name              员工姓名
  --mobile            手机号
  --job-number        工号
  --emp-type          员工类型（college_student / college_teacher）
  --main-dept-id      主部门 ID
  --login-id-type     登录类型
  --exclusive-account 是否独占账号（true/false）
  --send-active-sms   是否发送激活短信（true/false）
  --staff-ids         批量操作的 staffId 列表，逗号分隔
  --target-dept-id    变更目标部门 ID
  --offset / --size   分页参数
  --order-field       排序字段
  --ordering          排序顺序（asc / desc）`,
		Example: `  dws college-contact employee list-unaccepted --dept-id 12345
  dws college-contact employee list-unaccepted --dept-id 12345 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input["deptId"] = deptID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"emp-type", "empType"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "list_unaccepted_employees", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(listUnacceptedEmployeesCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "list_unaccepted_employees",
				CanonicalPath:  "college-contact.list_unaccepted_employees",
				CLIPath:        "college-contact employee list-unaccepted",
				PrimaryCLIPath: "college-contact employee list-unaccepted",
			},
			Description: "查询未接受邀请的员工列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "list_unaccepted_employees"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询未接受邀请的员工列表",
				UseWhen:      []string{"需要查询未接受邀请的员工列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee list-unaccepted --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	listUnacceptedEmployeesCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	listUnacceptedEmployeesCmd.Flags().String("staff-id", "", "员工 staffId")
	listUnacceptedEmployeesCmd.Flags().String("name", "", "员工姓名")
	listUnacceptedEmployeesCmd.Flags().String("mobile", "", "手机号")
	listUnacceptedEmployeesCmd.Flags().String("job-number", "", "工号")
	listUnacceptedEmployeesCmd.Flags().String("emp-type", "", "员工类型（college_student / college_teacher）")
	listUnacceptedEmployeesCmd.Flags().String("main-dept-id", "", "主部门 ID")
	listUnacceptedEmployeesCmd.Flags().String("login-id-type", "", "登录类型")
	listUnacceptedEmployeesCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	listUnacceptedEmployeesCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	listUnacceptedEmployeesCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	listUnacceptedEmployeesCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	listUnacceptedEmployeesCmd.Flags().String("offset", "", "分页偏移量")
	listUnacceptedEmployeesCmd.Flags().String("size", "", "分页大小")
	listUnacceptedEmployeesCmd.Flags().String("order-field", "", "排序字段")
	listUnacceptedEmployeesCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	listUnactiveEmployeesCmd := &cobra.Command{
		Use:   "list-unactive",
		Short: "查询未激活的员工列表",
		Long: `查询指定部门下未激活的员工列表，返回员工总数及员工列表。
必填：
  --dept-id  部门 ID
可选：
  --staff-id          员工 staffId
  --name              员工姓名
  --mobile            手机号
  --job-number        工号
  --main-dept-id      主部门 ID
  --login-id-type     登录类型
  --exclusive-account 是否独占账号（true/false）
  --send-active-sms   是否发送激活短信（true/false）
  --staff-ids         批量操作的 staffId 列表，逗号分隔
  --target-dept-id    变更目标部门 ID
  --offset / --size   分页参数
  --order-field       排序字段
  --ordering          排序顺序（asc / desc）`,
		Example: `  dws college-contact employee list-unactive --dept-id 12345
  dws college-contact employee list-unactive --dept-id 12345 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input["deptId"] = deptID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "list_unactive_employees", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(listUnactiveEmployeesCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "list_unactive_employees",
				CanonicalPath:  "college-contact.list_unactive_employees",
				CLIPath:        "college-contact employee list-unactive",
				PrimaryCLIPath: "college-contact employee list-unactive",
			},
			Description: "查询未激活的员工列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "list_unactive_employees"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询未激活的员工列表",
				UseWhen:      []string{"需要查询未激活的员工列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee list-unactive --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	listUnactiveEmployeesCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	listUnactiveEmployeesCmd.Flags().String("staff-id", "", "员工 staffId")
	listUnactiveEmployeesCmd.Flags().String("name", "", "员工姓名")
	listUnactiveEmployeesCmd.Flags().String("mobile", "", "手机号")
	listUnactiveEmployeesCmd.Flags().String("job-number", "", "工号")
	listUnactiveEmployeesCmd.Flags().String("main-dept-id", "", "主部门 ID")
	listUnactiveEmployeesCmd.Flags().String("login-id-type", "", "登录类型")
	listUnactiveEmployeesCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	listUnactiveEmployeesCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	listUnactiveEmployeesCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	listUnactiveEmployeesCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	listUnactiveEmployeesCmd.Flags().String("offset", "", "分页偏移量")
	listUnactiveEmployeesCmd.Flags().String("size", "", "分页大小")
	listUnactiveEmployeesCmd.Flags().String("order-field", "", "排序字段")
	listUnactiveEmployeesCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	deptCmd.AddCommand(getStandardStructureCmd, getDeptDetailCmd, getDeptChainCmd, searchContactCmd, createDeptCmd, updateDeptCmd, deleteDeptCmd, batchUpdateDeptTypeCmd)

	// ════════════════════════════════════════════════════════════
	// dept 子命令组 - 部门管理
	// ════════════════════════════════════════════════════════════

	overviewStatisticsCmd := &cobra.Command{
		Use:   "overview",
		Short: "查询高校概览统计",
		Long: `查询高校概览统计，返回学生数、教师数、全校人数、未激活员工数、未接受邀请员工数、院系数、行政班数。
所有参数均可选：
  --dept-id   按部门 ID 过滤
  --staff-id  按员工 staffId 过滤
  --keyword   按关键词搜索
  --offset / --size  分页参数`,
		Example: `  dws college-contact dept overview
  dws college-contact dept overview --dept-id 12345
  dws college-contact dept overview --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("dept-id"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-id 须为整数: %w", err)
				}
				input["deptId"] = n
			}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("keyword"); strings.TrimSpace(v) != "" {
				input["keyword"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "get_college_overview_statistics", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(overviewStatisticsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_college_overview_statistics",
				CanonicalPath:  "college-contact.get_college_overview_statistics",
				CLIPath:        "college-contact dept overview",
				PrimaryCLIPath: "college-contact dept overview",
			},
			Description: "查询高校概览统计",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_college_overview_statistics"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询高校概览统计",
				UseWhen:      []string{"需要查询高校概览统计时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact dept overview --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	overviewStatisticsCmd.Flags().String("dept-id", "", "部门 ID")
	overviewStatisticsCmd.Flags().String("staff-id", "", "员工 staffId")
	overviewStatisticsCmd.Flags().String("keyword", "", "搜索关键词")
	overviewStatisticsCmd.Flags().String("offset", "", "分页偏移量")
	overviewStatisticsCmd.Flags().String("size", "", "分页大小")

	deptCmd.AddCommand(overviewStatisticsCmd)

	upgradeStatusCmd := &cobra.Command{
		Use:   "upgrade-status",
		Short: "查询升级状态",
		Long: `查询高校升级状态，返回升级状态码及描述信息。
所有参数均可选：
  --dept-id   按部门 ID 过滤
  --staff-id  按员工 staffId 过滤
  --keyword   按关键词搜索
  --offset / --size  分页参数`,
		Example: `  dws college-contact employee upgrade-status
  dws college-contact employee upgrade-status --dept-id 12345
  dws college-contact employee upgrade-status --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("dept-id"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-id 须为整数: %w", err)
				}
				input["deptId"] = n
			}
			if v, _ := cmd.Flags().GetString("staff-id"); strings.TrimSpace(v) != "" {
				input["staffId"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("keyword"); strings.TrimSpace(v) != "" {
				input["keyword"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}
			return callMCPToolOnServer("college-contact", "get_college_upgrade_status", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(upgradeStatusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_college_upgrade_status",
				CanonicalPath:  "college-contact.get_college_upgrade_status",
				CLIPath:        "college-contact employee upgrade-status",
				PrimaryCLIPath: "college-contact employee upgrade-status",
			},
			Description: "查询升级状态",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_college_upgrade_status"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询升级状态",
				UseWhen:      []string{"需要查询升级状态时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee upgrade-status --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	upgradeStatusCmd.Flags().String("dept-id", "", "部门 ID")
	upgradeStatusCmd.Flags().String("staff-id", "", "员工 staffId")
	upgradeStatusCmd.Flags().String("keyword", "", "搜索关键词")
	upgradeStatusCmd.Flags().String("offset", "", "分页偏移量")
	upgradeStatusCmd.Flags().String("size", "", "分页大小")

	startUpgradeCmd := &cobra.Command{
		Use:   "start-upgrade",
		Short: "启动升级",
		Long:  `启动高校通讯录升级，返回操作是否成功。`,
		Example: `  dws college-contact employee start-upgrade
  dws college-contact employee start-upgrade -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("college-contact", "start_upgrade", map[string]any{
				"input": "",
			})
		},
	}

	DeclareLeafMetadata(startUpgradeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "start_upgrade",
				CanonicalPath:  "college-contact.start_upgrade",
				CLIPath:        "college-contact employee start-upgrade",
				PrimaryCLIPath: "college-contact employee start-upgrade",
			},
			Description: "启动升级",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "start_upgrade"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "启动升级",
				UseWhen:      []string{"需要启动升级时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee start-upgrade --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ════════════════════════════════════════════════════════════
	// employee 子命令组 — 员工管理
	// ════════════════════════════════════════════════════════════

	employeeCmd := &cobra.Command{Use: "employee", Short: "员工管理", RunE: groupRunE}

	getEmployeeDetailCmd := &cobra.Command{
		Use:   "get-detail",
		Short: "查询员工详情",
		Long: `查询员工详情，返回员工基本信息、主部门、所在部门列表、账号状态等。
必填：
  --staff-id  员工 staffId
可选：
  --dept-id             部门 ID
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --emp-type            员工类型（college_student / college_teacher）
  --main-dept-id        主部门 ID
  --login-id-type       登录类型
  --exclusive-account   是否独占账号（true/false）
  --send-active-sms     是否发送激活短信（true/false）
  --staff-ids           批量操作的 staffId 列表，逗号分隔
  --target-dept-id      变更目标部门 ID
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）`,
		Example: `  dws college-contact employee get-detail --staff-id S12345
  dws college-contact employee get-detail --staff-id S12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			staffID, _ := cmd.Flags().GetString("staff-id")
			staffID = strings.TrimSpace(staffID)
			if staffID == "" {
				return fmt.Errorf("--staff-id 为必填参数")
			}
			input["staffId"] = staffID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-id", "deptId"},
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"emp-type", "empType"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "get_employee_detail", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(getEmployeeDetailCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_employee_detail",
				CanonicalPath:  "college-contact.get_employee_detail",
				CLIPath:        "college-contact employee get-detail",
				PrimaryCLIPath: "college-contact employee get-detail",
			},
			Description: "查询员工详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_employee_detail"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询员工详情",
				UseWhen:      []string{"需要查询员工详情时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee get-detail --staff-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	getEmployeeDetailCmd.Flags().String("staff-id", "", "员工 staffId（必填）")
	getEmployeeDetailCmd.Flags().String("dept-id", "", "部门 ID")
	getEmployeeDetailCmd.Flags().String("name", "", "员工姓名")
	getEmployeeDetailCmd.Flags().String("mobile", "", "手机号")
	getEmployeeDetailCmd.Flags().String("job-number", "", "工号")
	getEmployeeDetailCmd.Flags().String("emp-type", "", "员工类型（college_student / college_teacher）")
	getEmployeeDetailCmd.Flags().String("main-dept-id", "", "主部门 ID")
	getEmployeeDetailCmd.Flags().String("login-id-type", "", "登录类型")
	getEmployeeDetailCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	getEmployeeDetailCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	getEmployeeDetailCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	getEmployeeDetailCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	getEmployeeDetailCmd.Flags().String("offset", "", "分页偏移量")
	getEmployeeDetailCmd.Flags().String("size", "", "分页大小")
	getEmployeeDetailCmd.Flags().String("order-field", "", "排序字段")
	getEmployeeDetailCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	addEmployeeCmd := &cobra.Command{
		Use:   "add",
		Short: "添加员工",
		Long: `添加高校员工，返回添加成功/失败数量及邮箱初始密码映射。
必填：
  --emp-type          员工类型（college_student / college_teacher）
  --main-dept-id      主部门 ID
  --exclusive-account 是否独占账号（true/false）
可选：
  --dept-id             部门 ID
  --staff-id            员工 staffId
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --login-id-type       登录类型
  --send-active-sms     是否发送激活短信（true/false）
  --staff-ids           批量操作的 staffId 列表，逗号分隔
  --target-dept-id      变更目标部门 ID
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）`,
		Example: `  dws college-contact employee add --emp-type college_student --main-dept-id 100 --exclusive-account true --name 张三 --mobile 13800138000
  dws college-contact employee add --emp-type college_teacher --main-dept-id 100 --exclusive-account false --name 李老师 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			empType, _ := cmd.Flags().GetString("emp-type")
			empType = strings.TrimSpace(empType)
			if empType == "" {
				return fmt.Errorf("--emp-type 为必填参数")
			}
			input["empType"] = empType

			// 必填 int64
			mainDeptIDRaw, _ := cmd.Flags().GetString("main-dept-id")
			mainDeptIDRaw = strings.TrimSpace(mainDeptIDRaw)
			if mainDeptIDRaw == "" {
				return fmt.Errorf("--main-dept-id 为必填参数")
			}
			mainDeptID, err := strconv.ParseInt(mainDeptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--main-dept-id 须为整数: %w", err)
			}
			input["mainDeptId"] = mainDeptID

			// 必填 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			if !cmd.Flags().Changed("exclusive-account") {
				return fmt.Errorf("--exclusive-account 为必填参数（true/false）")
			}
			eaRaw, _ := cmd.Flags().GetString("exclusive-account")
			exclusiveAccount, err := strconv.ParseBool(strings.TrimSpace(eaRaw))
			if err != nil {
				return fmt.Errorf("--exclusive-account 须为 true 或 false: %w", err)
			}
			input["exclusiveAccount"] = exclusiveAccount

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-id", "deptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			if cmd.Flags().Changed("send-active-sms") {
				sasRaw, _ := cmd.Flags().GetString("send-active-sms")
				v, err := strconv.ParseBool(strings.TrimSpace(sasRaw))
				if err != nil {
					return fmt.Errorf("--send-active-sms 须为 true 或 false: %w", err)
				}
				input["sendActiveSms"] = v
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "add_employee", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(addEmployeeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "add_employee",
				CanonicalPath:  "college-contact.add_employee",
				CLIPath:        "college-contact employee add",
				PrimaryCLIPath: "college-contact employee add",
			},
			Description: "添加员工",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "add_employee"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "添加员工",
				UseWhen:      []string{"需要添加员工时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee add --emp-type college_student --main-dept-id 100 --exclusive-account true --name 张三 --mobile 13800138000"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(true)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(true)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	addEmployeeCmd.Flags().String("emp-type", "", "员工类型（必填，college_student / college_teacher）")
	addEmployeeCmd.Flags().String("main-dept-id", "", "主部门 ID（必填）")
	addEmployeeCmd.Flags().String("exclusive-account", "", "是否独占账号（必填，true/false）")
	addEmployeeCmd.Flags().String("dept-id", "", "部门 ID")
	addEmployeeCmd.Flags().String("staff-id", "", "员工 staffId")
	addEmployeeCmd.Flags().String("name", "", "员工姓名")
	addEmployeeCmd.Flags().String("mobile", "", "手机号")
	addEmployeeCmd.Flags().String("job-number", "", "工号")
	addEmployeeCmd.Flags().String("login-id-type", "", "登录类型")
	addEmployeeCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	addEmployeeCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	addEmployeeCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	addEmployeeCmd.Flags().String("offset", "", "分页偏移量")
	addEmployeeCmd.Flags().String("size", "", "分页大小")
	addEmployeeCmd.Flags().String("order-field", "", "排序字段")
	addEmployeeCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	removeEmployeeCmd := &cobra.Command{
		Use:   "remove",
		Short: "移除员工",
		Long: `移除高校员工，返回移除成功/失败数量。
必填：
  --staff-ids  员工 staffId 列表，逗号分隔
可选：
  --dept-id             部门 ID
  --staff-id            员工 staffId
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --emp-type            员工类型（college_student / college_teacher）
  --main-dept-id        主部门 ID
  --login-id-type       登录类型
  --exclusive-account   是否独占账号（true/false）
  --send-active-sms     是否发送激活短信（true/false）
  --target-dept-id      变更目标部门 ID
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）

注意：移除操作不可逆。非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact employee remove --staff-ids S12345,S12346 --dry-run
  dws college-contact employee remove --staff-ids S12345,S12346 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 []string: --staff-ids
			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIDs []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIDs = append(staffIDs, part)
				}
			}
			if len(staffIDs) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}
			input["staffIds"] = staffIDs

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-id", "deptId"},
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"emp-type", "empType"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			return callMCPToolOnServer("college-contact", "remove_employee", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(removeEmployeeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "remove_employee",
				CanonicalPath:  "college-contact.remove_employee",
				CLIPath:        "college-contact employee remove",
				PrimaryCLIPath: "college-contact employee remove",
			},
			Description: "移除员工",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "remove_employee"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "移除员工",
				UseWhen:      []string{"需要移除员工时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee remove --staff-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	removeEmployeeCmd.Flags().String("staff-ids", "", "员工 staffId 列表，逗号分隔（必填）")
	removeEmployeeCmd.Flags().String("dept-id", "", "部门 ID")
	removeEmployeeCmd.Flags().String("staff-id", "", "员工 staffId")
	removeEmployeeCmd.Flags().String("name", "", "员工姓名")
	removeEmployeeCmd.Flags().String("mobile", "", "手机号")
	removeEmployeeCmd.Flags().String("job-number", "", "工号")
	removeEmployeeCmd.Flags().String("emp-type", "", "员工类型（college_student / college_teacher）")
	removeEmployeeCmd.Flags().String("main-dept-id", "", "主部门 ID")
	removeEmployeeCmd.Flags().String("login-id-type", "", "登录类型")
	removeEmployeeCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	removeEmployeeCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	removeEmployeeCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	removeEmployeeCmd.Flags().String("offset", "", "分页偏移量")
	removeEmployeeCmd.Flags().String("size", "", "分页大小")
	removeEmployeeCmd.Flags().String("order-field", "", "排序字段")
	removeEmployeeCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	changeEmployeeTypeCmd := &cobra.Command{
		Use:   "change-type",
		Short: "变更员工类型",
		Long: `变更高校员工类型，返回变更成功/失败数量。
必填：
  --emp-type   员工类型（college_student / college_teacher）
  --staff-id   员工 staffId
可选：
  --dept-id             部门 ID
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --main-dept-id        主部门 ID
  --login-id-type       登录类型
  --exclusive-account   是否独占账号（true/false）
  --send-active-sms     是否发送激活短信（true/false）
  --staff-ids           批量操作的 staffId 列表，逗号分隔
  --target-dept-id      变更目标部门 ID
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）`,
		Example: `  dws college-contact employee change-type --staff-id S12345 --emp-type college_teacher
  dws college-contact employee change-type --staff-id S12345 --emp-type college_student -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"emp-type", "empType"},
			} {
				v, _ := cmd.Flags().GetString(pair.flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", pair.flag)
				}
				input[pair.key] = v
			}

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-id", "deptId"},
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "change_employee_type", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(changeEmployeeTypeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "change_employee_type",
				CanonicalPath:  "college-contact.change_employee_type",
				CLIPath:        "college-contact employee change-type",
				PrimaryCLIPath: "college-contact employee change-type",
			},
			Description: "变更员工类型",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "change_employee_type"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "变更员工类型",
				UseWhen:      []string{"需要变更员工类型时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee change-type --staff-id 12345 --emp-type standard"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(true)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	changeEmployeeTypeCmd.Flags().String("staff-id", "", "员工 staffId（必填）")
	changeEmployeeTypeCmd.Flags().String("emp-type", "", "员工类型（必填，college_student / college_teacher）")
	changeEmployeeTypeCmd.Flags().String("dept-id", "", "部门 ID")
	changeEmployeeTypeCmd.Flags().String("name", "", "员工姓名")
	changeEmployeeTypeCmd.Flags().String("mobile", "", "手机号")
	changeEmployeeTypeCmd.Flags().String("job-number", "", "工号")
	changeEmployeeTypeCmd.Flags().String("main-dept-id", "", "主部门 ID")
	changeEmployeeTypeCmd.Flags().String("login-id-type", "", "登录类型")
	changeEmployeeTypeCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	changeEmployeeTypeCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	changeEmployeeTypeCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	changeEmployeeTypeCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	changeEmployeeTypeCmd.Flags().String("offset", "", "分页偏移量")
	changeEmployeeTypeCmd.Flags().String("size", "", "分页大小")
	changeEmployeeTypeCmd.Flags().String("order-field", "", "排序字段")
	changeEmployeeTypeCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	changeEmployeeDeptCmd := &cobra.Command{
		Use:   "change-dept",
		Short: "变更员工部门",
		Long: `变更高校员工所属部门，返回变更成功/失败数量。
必填：
  --staff-id       员工 staffId
  --target-dept-id 变更目标部门 ID
可选：
  --dept-id             部门 ID
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --emp-type            员工类型（college_student / college_teacher）
  --main-dept-id        主部门 ID
  --login-id-type       登录类型
  --exclusive-account   是否独占账号（true/false）
  --send-active-sms     是否发送激活短信（true/false）
  --staff-ids           批量操作的 staffId 列表，逗号分隔
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）`,
		Example: `  dws college-contact employee change-dept --staff-id S12345 --target-dept-id 200
  dws college-contact employee change-dept --staff-id S12345 --target-dept-id 200 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			staffID, _ := cmd.Flags().GetString("staff-id")
			staffID = strings.TrimSpace(staffID)
			if staffID == "" {
				return fmt.Errorf("--staff-id 为必填参数")
			}
			input["staffId"] = staffID

			// 必填 int64
			targetDeptIDRaw, _ := cmd.Flags().GetString("target-dept-id")
			targetDeptIDRaw = strings.TrimSpace(targetDeptIDRaw)
			if targetDeptIDRaw == "" {
				return fmt.Errorf("--target-dept-id 为必填参数")
			}
			targetDeptID, err := strconv.ParseInt(targetDeptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--target-dept-id 须为整数: %w", err)
			}
			input["targetDeptId"] = targetDeptID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"dept-id", "deptId"},
				{"main-dept-id", "mainDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"emp-type", "empType"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "change_employee_dept", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(changeEmployeeDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "change_employee_dept",
				CanonicalPath:  "college-contact.change_employee_dept",
				CLIPath:        "college-contact employee change-dept",
				PrimaryCLIPath: "college-contact employee change-dept",
			},
			Description: "变更员工部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "change_employee_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "变更员工部门",
				UseWhen:      []string{"需要变更员工部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee change-dept --staff-id 12345 --target-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(true)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(true)},
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	changeEmployeeDeptCmd.Flags().String("staff-id", "", "员工 staffId（必填）")
	changeEmployeeDeptCmd.Flags().String("target-dept-id", "", "变更目标部门 ID（必填）")
	changeEmployeeDeptCmd.Flags().String("dept-id", "", "部门 ID")
	changeEmployeeDeptCmd.Flags().String("name", "", "员工姓名")
	changeEmployeeDeptCmd.Flags().String("mobile", "", "手机号")
	changeEmployeeDeptCmd.Flags().String("job-number", "", "工号")
	changeEmployeeDeptCmd.Flags().String("emp-type", "", "员工类型（college_student / college_teacher）")
	changeEmployeeDeptCmd.Flags().String("main-dept-id", "", "主部门 ID")
	changeEmployeeDeptCmd.Flags().String("login-id-type", "", "登录类型")
	changeEmployeeDeptCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	changeEmployeeDeptCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	changeEmployeeDeptCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	changeEmployeeDeptCmd.Flags().String("offset", "", "分页偏移量")
	changeEmployeeDeptCmd.Flags().String("size", "", "分页大小")
	changeEmployeeDeptCmd.Flags().String("order-field", "", "排序字段")
	changeEmployeeDeptCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	sendActiveSmsCmd := &cobra.Command{
		Use:   "send-active-sms",
		Short: "发送激活短信",
		Long: `向指定部门下的员工发送激活短信，返回操作是否成功。
必填：
  --dept-id  部门 ID
可选：
  --staff-id            员工 staffId
  --name                员工姓名
  --mobile              手机号
  --job-number          工号
  --emp-type            员工类型（college_student / college_teacher）
  --main-dept-id        主部门 ID
  --login-id-type       登录类型
  --exclusive-account   是否独占账号（true/false）
  --send-active-sms     是否发送激活短信（true/false）
  --staff-ids           批量操作的 staffId 列表，逗号分隔
  --target-dept-id      变更目标部门 ID
  --offset / --size     分页参数
  --order-field         排序字段
  --ordering            排序顺序（asc / desc）`,
		Example: `  dws college-contact employee send-active-sms --dept-id 100
  dws college-contact employee send-active-sms --dept-id 100 --staff-ids S12345,S12346 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			deptIDRaw, _ := cmd.Flags().GetString("dept-id")
			deptIDRaw = strings.TrimSpace(deptIDRaw)
			if deptIDRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptID, err := strconv.ParseInt(deptIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}
			input["deptId"] = deptID

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"main-dept-id", "mainDeptId"},
				{"target-dept-id", "targetDeptId"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if raw, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(raw) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"exclusive-account", "exclusiveAccount"},
				{"send-active-sms", "sendActiveSms"},
			} {
				if cmd.Flags().Changed(pair.flag) {
					raw, _ := cmd.Flags().GetString(pair.flag)
					v, err := strconv.ParseBool(strings.TrimSpace(raw))
					if err != nil {
						return fmt.Errorf("--%s 须为 true 或 false: %w", pair.flag, err)
					}
					input[pair.key] = v
				}
			}

			// 可选 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"staff-id", "staffId"},
				{"name", "name"},
				{"mobile", "mobile"},
				{"job-number", "jobNumber"},
				{"emp-type", "empType"},
				{"login-id-type", "loginIdType"},
				{"order-field", "orderField"},
				{"ordering", "ordering"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					input[pair.key] = strings.TrimSpace(v)
				}
			}

			// 可选 []string: --staff-ids
			if raw, _ := cmd.Flags().GetString("staff-ids"); strings.TrimSpace(raw) != "" {
				var staffIDs []string
				for _, part := range strings.Split(raw, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						staffIDs = append(staffIDs, part)
					}
				}
				if len(staffIDs) > 0 {
					input["staffIds"] = staffIDs
				}
			}

			return callMCPToolOnServer("college-contact", "send_active_sms", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(sendActiveSmsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "send_active_sms",
				CanonicalPath:  "college-contact.send_active_sms",
				CLIPath:        "college-contact employee send-active-sms",
				PrimaryCLIPath: "college-contact employee send-active-sms",
			},
			Description: "发送激活短信",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "send_active_sms"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "发送激活短信",
				UseWhen:      []string{"需要发送激活短信时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact employee send-active-sms --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "job-number", Property: "input.jobNumber", Required: boolPtr(false)},
				{Name: "emp-type", Property: "input.empType", Required: boolPtr(false)},
				{Name: "main-dept-id", Property: "input.mainDeptId", Required: boolPtr(false)},
				{Name: "login-id-type", Property: "input.loginIdType", Required: boolPtr(false)},
				{Name: "exclusive-account", Property: "input.exclusiveAccount", Required: boolPtr(false)},
				{Name: "send-active-sms", Property: "input.sendActiveSms", Required: boolPtr(false)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(false)},
				{Name: "target-dept-id", Property: "input.targetDeptId", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(false)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(false)},
			},
		},
	})

	sendActiveSmsCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	sendActiveSmsCmd.Flags().String("staff-id", "", "员工 staffId")
	sendActiveSmsCmd.Flags().String("name", "", "员工姓名")
	sendActiveSmsCmd.Flags().String("mobile", "", "手机号")
	sendActiveSmsCmd.Flags().String("job-number", "", "工号")
	sendActiveSmsCmd.Flags().String("emp-type", "", "员工类型（college_student / college_teacher）")
	sendActiveSmsCmd.Flags().String("main-dept-id", "", "主部门 ID")
	sendActiveSmsCmd.Flags().String("login-id-type", "", "登录类型")
	sendActiveSmsCmd.Flags().String("exclusive-account", "", "是否独占账号（true/false）")
	sendActiveSmsCmd.Flags().String("send-active-sms", "", "是否发送激活短信（true/false）")
	sendActiveSmsCmd.Flags().String("staff-ids", "", "批量操作的 staffId 列表，逗号分隔")
	sendActiveSmsCmd.Flags().String("target-dept-id", "", "变更目标部门 ID")
	sendActiveSmsCmd.Flags().String("offset", "", "分页偏移量")
	sendActiveSmsCmd.Flags().String("size", "", "分页大小")
	sendActiveSmsCmd.Flags().String("order-field", "", "排序字段")
	sendActiveSmsCmd.Flags().String("ordering", "", "排序顺序（asc / desc）")

	// ════════════════════════════════════════════════════════════
	// alumni 子命令组 — 校友管理
	// ════════════════════════════════════════════════════════════

	alumniCmd := &cobra.Command{Use: "alumni", Short: "校友管理", RunE: groupRunE}

	getAlumniDeptTreeCmd := &cobra.Command{
		Use:   "get-dept-tree",
		Short: "查询校友部门树",
		Long: `查询校友部门树，返回指定校友部门的下级部门列表。
必填：
  --alumni-dept-id  校友部门 ID`,
		Example: `  dws college-contact alumni get-dept-tree --alumni-dept-id 12345
  dws college-contact alumni get-dept-tree --alumni-dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			return callMCPToolOnServer("college-contact", "get_alumni_dept_tree", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId},
			})
		},
	}

	DeclareLeafMetadata(getAlumniDeptTreeCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_alumni_dept_tree",
				CanonicalPath:  "college-contact.get_alumni_dept_tree",
				CLIPath:        "college-contact alumni get-dept-tree",
				PrimaryCLIPath: "college-contact alumni get-dept-tree",
			},
			Description: "查询校友部门树",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_alumni_dept_tree"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友部门树",
				UseWhen:      []string{"需要查询校友部门树时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni get-dept-tree --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	getAlumniDeptTreeCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	getAlumniDeptInfoCmd := &cobra.Command{
		Use:   "get-info",
		Short: "查询校友部门详情",
		Long: `查询校友部门详情，返回部门名称、部门人数、负责人列表、是否有下级部门等。
必填：
  --alumni-dept-id  校友部门 ID`,
		Example: `  dws college-contact alumni get-info --alumni-dept-id 12345
  dws college-contact alumni get-info --alumni-dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			return callMCPToolOnServer("college-contact", "get_alumni_dept_info", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId},
			})
		},
	}

	DeclareLeafMetadata(getAlumniDeptInfoCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_alumni_dept_info",
				CanonicalPath:  "college-contact.get_alumni_dept_info",
				CLIPath:        "college-contact alumni get-info",
				PrimaryCLIPath: "college-contact alumni get-info",
			},
			Description: "查询校友部门详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_alumni_dept_info"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友部门详情",
				UseWhen:      []string{"需要查询校友部门详情时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni get-info --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	getAlumniDeptInfoCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	listAlumniCmd := &cobra.Command{
		Use:   "list",
		Short: "查询校友列表",
		Long: `查询指定校友部门下的校友列表，返回校友列表。
必填：
  --alumni-dept-id  校友部门 ID
  --order-field     排序字段（如 "dept_entry"、"custom"）
  --ordering        排序顺序（"asc" 或 "desc"）
可选：
  --offset / --size  分页参数`,
		Example: `  dws college-contact alumni list --alumni-dept-id 12345 --order-field dept_entry --ordering asc
  dws college-contact alumni list --alumni-dept-id 12345 --order-field dept_entry --ordering asc --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			input["alumniDeptId"] = alumniDeptId

			// 必填 string
			for _, flag := range []string{"order-field", "ordering"} {
				v, _ := cmd.Flags().GetString(flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", flag)
				}
				key := flag
				if flag == "order-field" {
					key = "orderField"
				}
				input[key] = v
			}

			// 可选 int64
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"offset", "offset"},
				{"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(pair.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", pair.flag, err)
					}
					input[pair.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "list_alumni", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(listAlumniCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "list_alumni",
				CanonicalPath:  "college-contact.list_alumni",
				CLIPath:        "college-contact alumni list",
				PrimaryCLIPath: "college-contact alumni list",
			},
			Description: "查询校友列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "list_alumni"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友列表",
				UseWhen:      []string{"需要查询校友列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni list --alumni-dept-id 12345 --order-field dept_entry --ordering asc"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "order-field", Property: "input.orderField", Required: boolPtr(true)},
				{Name: "ordering", Property: "input.ordering", Required: boolPtr(true)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	listAlumniCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")
	listAlumniCmd.Flags().String("order-field", "", "排序字段（必填，如 dept_entry / custom）")
	listAlumniCmd.Flags().String("ordering", "", "排序顺序（必填，asc / desc）")
	listAlumniCmd.Flags().String("offset", "", "分页偏移量")
	listAlumniCmd.Flags().String("size", "", "分页大小")

	queryAlumnusCmd := &cobra.Command{
		Use:   "query",
		Short: "查询校友详情",
		Long: `查询校友详情，返回校友基本信息、入学/毕业时间、工作单位、部门列表等。
必填：
  --staff-id  校友 staffId`,
		Example: `  dws college-contact alumni query --staff-id S12345
  dws college-contact alumni query --staff-id S12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			staffId, _ := cmd.Flags().GetString("staff-id")
			staffId = strings.TrimSpace(staffId)
			if staffId == "" {
				return fmt.Errorf("--staff-id 为必填参数")
			}
			return callMCPToolOnServer("college-contact", "query_alumnus", map[string]any{
				"input": map[string]any{"staffId": staffId},
			})
		},
	}

	DeclareLeafMetadata(queryAlumnusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_alumnus",
				CanonicalPath:  "college-contact.query_alumnus",
				CLIPath:        "college-contact alumni query",
				PrimaryCLIPath: "college-contact alumni query",
			},
			Description: "查询校友详情",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_alumnus"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友详情",
				UseWhen:      []string{"需要查询校友详情时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni query --staff-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(true)},
			},
		},
	})

	queryAlumnusCmd.Flags().String("staff-id", "", "校友 staffId（必填）")

	// ── search ────────────────────────────────────────────────
	searchAlumniCmd := &cobra.Command{
		Use:   "search",
		Short: "搜索校友",
		Long: `按关键词搜索校友，返回匹配校友列表。

参数说明：
  --keyword  搜索关键词（必填）
  --offset   分页偏移量（可选）
  --size     分页大小（可选）`,
		Example: `  dws college-contact alumni search --keyword 张三
  dws college-contact alumni search --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string - keyword
			keyword, _ := cmd.Flags().GetString("keyword")
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				return fmt.Errorf("--keyword 为必填参数")
			}
			input["keyword"] = keyword

			// 可选 int64 - offset, size
			for _, p := range []struct{ flag, key string }{
				{"offset", "offset"}, {"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(p.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", p.flag, err)
					}
					input[p.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "search_alumni", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(searchAlumniCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "search_alumni",
				CanonicalPath:  "college-contact.search_alumni",
				CLIPath:        "college-contact alumni search",
				PrimaryCLIPath: "college-contact alumni search",
			},
			Description: "搜索校友",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "search_alumni"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "搜索校友",
				UseWhen:      []string{"需要搜索校友时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni search --keyword 计算机"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(true)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	searchAlumniCmd.Flags().String("keyword", "", "搜索关键词（必填）")
	searchAlumniCmd.Flags().String("offset", "", "分页偏移量（可选）")
	searchAlumniCmd.Flags().String("size", "", "分页大小（可选）")

	// ── list-unaccepted ───────────────────────────────────────
	listUnacceptedAlumnusCmd := &cobra.Command{
		Use:   "list-unaccepted",
		Short: "查询未接受邀请的校友列表",
		Long: `查询指定校友部门下未接受邀请的校友列表，返回校友列表及总数。

参数说明：
  --alumni-dept-id  校友部门 ID（必填）
  --offset          分页偏移量（可选）
  --size            分页大小（可选）`,
		Example: `  dws college-contact alumni list-unaccepted --alumni-dept-id 12345
  dws college-contact alumni list-unaccepted --alumni-dept-id 12345 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 int64 - alumniDeptId
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			input["alumniDeptId"] = alumniDeptId

			// 可选 int64 - offset, size
			for _, p := range []struct{ flag, key string }{
				{"offset", "offset"}, {"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(p.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", p.flag, err)
					}
					input[p.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "query_unaccepted_alumnus", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(listUnacceptedAlumnusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_unaccepted_alumnus",
				CanonicalPath:  "college-contact.query_unaccepted_alumnus",
				CLIPath:        "college-contact alumni list-unaccepted",
				PrimaryCLIPath: "college-contact alumni list-unaccepted",
			},
			Description: "查询未接受邀请的校友列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_unaccepted_alumnus"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询未接受邀请的校友列表",
				UseWhen:      []string{"需要查询未接受邀请的校友列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni list-unaccepted --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	listUnacceptedAlumnusCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")
	listUnacceptedAlumnusCmd.Flags().String("offset", "", "分页偏移量（可选）")
	listUnacceptedAlumnusCmd.Flags().String("size", "", "分页大小（可选）")

	// ── get-group ─────────────────────────────────────────────
	getAlumniGroupCmd := &cobra.Command{
		Use:   "get-group",
		Short: "查询校友群",
		Long: `查询指定校友部门的校友群信息，返回群名称、群 CID、群主信息等。

必填：
  --alumni-dept-id  校友部门 ID`,
		Example: `  dws college-contact alumni get-group --alumni-dept-id 12345
  dws college-contact alumni get-group --alumni-dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			return callMCPToolOnServer("college-contact", "get_alumni_group", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId},
			})
		},
	}

	DeclareLeafMetadata(getAlumniGroupCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_alumni_group",
				CanonicalPath:  "college-contact.get_alumni_group",
				CLIPath:        "college-contact alumni get-group",
				PrimaryCLIPath: "college-contact alumni get-group",
			},
			Description: "查询校友群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_alumni_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友群",
				UseWhen:      []string{"需要查询校友群时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni get-group --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	getAlumniGroupCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	// ── create-dept ───────────────────────────────────────────
	createAlumniDeptsCmd := &cobra.Command{
		Use:   "create-dept",
		Short: "创建校友子部门",
		Long: `在指定校友部门下创建子部门，返回创建的部门列表。

必填：
  --alumni-dept-id  父校友部门 ID
  --dept-name       新部门名称`,
		Example: `  dws college-contact alumni create-dept --alumni-dept-id 12345 --dept-name 2020级校友
  dws college-contact alumni create-dept --alumni-dept-id 12345 --dept-name 2020级校友 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			deptName, _ := cmd.Flags().GetString("dept-name")
			deptName = strings.TrimSpace(deptName)
			if deptName == "" {
				return fmt.Errorf("--dept-name 为必填参数")
			}
			return callMCPToolOnServer("college-contact", "create_alumni_depts", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId, "deptName": deptName},
			})
		},
	}

	DeclareLeafMetadata(createAlumniDeptsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "create_alumni_depts",
				CanonicalPath:  "college-contact.create_alumni_depts",
				CLIPath:        "college-contact alumni create-dept",
				PrimaryCLIPath: "college-contact alumni create-dept",
			},
			Description: "创建校友子部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "create_alumni_depts"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建校友子部门",
				UseWhen:      []string{"需要创建校友子部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni create-dept --alumni-dept-id 12345 --dept-name 测试"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "dept-name", Property: "input.deptName", Required: boolPtr(true)},
			},
		},
	})

	createAlumniDeptsCmd.Flags().String("alumni-dept-id", "", "父校友部门 ID（必填）")
	createAlumniDeptsCmd.Flags().String("dept-name", "", "新部门名称（必填）")

	// ── update-dept ───────────────────────────────────────────
	updateAlumniDeptCmd := &cobra.Command{
		Use:   "update-dept",
		Short: "更新校友部门",
		Long: `更新校友部门名称。

必填：
  --alumni-dept-id  校友部门 ID
  --dept-name       新部门名称`,
		Example: `  dws college-contact alumni update-dept --alumni-dept-id 12345 --dept-name 2021级校友
  dws college-contact alumni update-dept --alumni-dept-id 12345 --dept-name 2021级校友 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			deptName, _ := cmd.Flags().GetString("dept-name")
			deptName = strings.TrimSpace(deptName)
			if deptName == "" {
				return fmt.Errorf("--dept-name 为必填参数")
			}
			return callMCPToolOnServer("college-contact", "update_alumni_dept", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId, "deptName": deptName},
			})
		},
	}

	DeclareLeafMetadata(updateAlumniDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "update_alumni_dept",
				CanonicalPath:  "college-contact.update_alumni_dept",
				CLIPath:        "college-contact alumni update-dept",
				PrimaryCLIPath: "college-contact alumni update-dept",
			},
			Description: "更新校友部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "update_alumni_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "更新校友部门",
				UseWhen:      []string{"需要更新校友部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni update-dept --alumni-dept-id 12345 --dept-name 测试"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "dept-name", Property: "input.deptName", Required: boolPtr(true)},
			},
		},
	})

	updateAlumniDeptCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")
	updateAlumniDeptCmd.Flags().String("dept-name", "", "新部门名称（必填）")

	// ── delete-dept ───────────────────────────────────────────
	deleteAlumniDeptCmd := &cobra.Command{
		Use:   "delete-dept",
		Short: "删除校友部门",
		Long: `删除指定校友部门。注意：删除操作不可逆。

必填：
  --alumni-dept-id  校友部门 ID

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact alumni delete-dept --alumni-dept-id 12345 --dry-run
  dws college-contact alumni delete-dept --alumni-dept-id 12345 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}
			return callMCPToolOnServer("college-contact", "delete_alumni_dept", map[string]any{
				"input": map[string]any{"alumniDeptId": alumniDeptId},
			})
		},
	}

	DeclareLeafMetadata(deleteAlumniDeptCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_alumni_dept",
				CanonicalPath:  "college-contact.delete_alumni_dept",
				CLIPath:        "college-contact alumni delete-dept",
				PrimaryCLIPath: "college-contact alumni delete-dept",
			},
			Description: "删除校友部门",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_alumni_dept"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除校友部门",
				UseWhen:      []string{"需要删除校友部门时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni delete-dept --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	deleteAlumniDeptCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	// ── update-managers ───────────────────────────────────────
	updateAlumniDeptManagersCmd := &cobra.Command{
		Use:   "update-managers",
		Short: "设置校友部门负责人",
		Long: `设置指定校友部门的负责人（管理员）列表。
必填：
  --alumni-dept-id  校友部门 ID
  --admin-user-ids  管理员 userId 列表，逗号分隔`,
		Example: `  dws college-contact alumni update-managers --alumni-dept-id 12345 --admin-user-ids user001,user002
  dws college-contact alumni update-managers --alumni-dept-id 12345 --admin-user-ids user001 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("admin-user-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--admin-user-ids 为必填参数")
			}
			var adminUserIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					adminUserIds = append(adminUserIds, part)
				}
			}
			if len(adminUserIds) == 0 {
				return fmt.Errorf("--admin-user-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "update_alumni_dept_managers", map[string]any{
				"input": map[string]any{
					"alumniDeptId": alumniDeptId,
					"adminUserIds": adminUserIds,
				},
			})
		},
	}

	DeclareLeafMetadata(updateAlumniDeptManagersCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "update_alumni_dept_managers",
				CanonicalPath:  "college-contact.update_alumni_dept_managers",
				CLIPath:        "college-contact alumni update-managers",
				PrimaryCLIPath: "college-contact alumni update-managers",
			},
			Description: "设置校友部门负责人",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "update_alumni_dept_managers"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "设置校友部门负责人",
				UseWhen:      []string{"需要设置校友部门负责人时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni update-managers --alumni-dept-id 12345 --admin-user-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "admin-user-ids", Property: "input.adminUserIds", Required: boolPtr(true)},
			},
		},
	})

	updateAlumniDeptManagersCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")
	updateAlumniDeptManagersCmd.Flags().String("admin-user-ids", "", "管理员 userId 列表，逗号分隔（必填）")

	// ── add-alumnus ───────────────────────────────────────────
	addAlumnusCmd := &cobra.Command{
		Use:   "add-alumnus",
		Short: "添加校友",
		Long: `添加校友到指定校友部门，返回添加是否成功。
必填：
  --name      校友姓名
  --mobile    手机号
  --dept-ids  校友部门 ID 列表，逗号分隔
可选：
  --student-number  学号
  --email           邮箱
  --intake          入学年份
  --outtake         毕业年份`,
		Example: `  dws college-contact alumni add-alumnus --name 张三 --mobile 13800138000 --dept-ids 12345,67890
  dws college-contact alumni add-alumnus --name 张三 --mobile 13800138000 --dept-ids 12345 --student-number 2020001 --intake 2020 --outtake 2024 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			for _, p := range []struct{ flag, key string }{
				{"name", "name"}, {"mobile", "mobile"},
			} {
				v, _ := cmd.Flags().GetString(p.flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", p.flag)
				}
				input[p.key] = v
			}

			// 必填 []int64: --dept-ids
			raw, _ := cmd.Flags().GetString("dept-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--dept-ids 为必填参数")
			}
			var deptIds []int64
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-ids 元素须为整数: %w", err)
				}
				deptIds = append(deptIds, n)
			}
			if len(deptIds) == 0 {
				return fmt.Errorf("--dept-ids 不能为空列表")
			}
			input["deptIds"] = deptIds

			// 可选 string
			for _, p := range []struct{ flag, key string }{
				{"student-number", "studentNumber"},
				{"email", "email"},
				{"intake", "intake"},
				{"outtake", "outtake"},
			} {
				if v, _ := cmd.Flags().GetString(p.flag); strings.TrimSpace(v) != "" {
					input[p.key] = strings.TrimSpace(v)
				}
			}

			return callMCPToolOnServer("college-contact", "add_alumnus", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(addAlumnusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "add_alumnus",
				CanonicalPath:  "college-contact.add_alumnus",
				CLIPath:        "college-contact alumni add-alumnus",
				PrimaryCLIPath: "college-contact alumni add-alumnus",
			},
			Description: "添加校友",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "add_alumnus"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "添加校友",
				UseWhen:      []string{"需要添加校友时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni add-alumnus --name 张三 --mobile 13800138000 --dept-ids 12345,67890"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "mobile", Property: "input.mobile", Required: boolPtr(false)},
				{Name: "dept-ids", Property: "input.deptIds", Required: boolPtr(true)},
				{Name: "student-number", Property: "input.studentNumber", Required: boolPtr(false)},
				{Name: "email", Property: "input.email", Required: boolPtr(false)},
				{Name: "intake", Property: "input.intake", Required: boolPtr(false)},
				{Name: "outtake", Property: "input.outtake", Required: boolPtr(false)},
			},
		},
	})

	addAlumnusCmd.Flags().String("name", "", "校友姓名（必填）")
	addAlumnusCmd.Flags().String("mobile", "", "手机号（必填）")
	addAlumnusCmd.Flags().String("dept-ids", "", "校友部门 ID 列表，逗号分隔（必填）")
	addAlumnusCmd.Flags().String("student-number", "", "学号")
	addAlumnusCmd.Flags().String("email", "", "邮箱")
	addAlumnusCmd.Flags().String("intake", "", "入学年份")
	addAlumnusCmd.Flags().String("outtake", "", "毕业年份")

	// ── update-alumnus ────────────────────────────────────────
	updateAlumnusCmd := &cobra.Command{
		Use:   "update-alumnus",
		Short: "更新校友信息",
		Long: `更新校友信息，返回更新是否成功。
必填：
  --staff-id  校友 staffId
  --name      校友姓名
  --dept-ids  校友部门 ID 列表，逗号分隔
可选：
  --student-number  学号
  --email           邮箱
  --intake          入学年份
  --outtake         毕业年份`,
		Example: `  dws college-contact alumni update-alumnus --staff-id staff001 --name 张三 --dept-ids 12345,67890
  dws college-contact alumni update-alumnus --staff-id staff001 --name 张三 --dept-ids 12345 --student-number 2020001 --intake 2020 --outtake 2024 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			for _, p := range []struct{ flag, key string }{
				{"staff-id", "staffId"}, {"name", "name"},
			} {
				v, _ := cmd.Flags().GetString(p.flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", p.flag)
				}
				input[p.key] = v
			}

			// 必填 []int64: --dept-ids
			raw, _ := cmd.Flags().GetString("dept-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--dept-ids 为必填参数")
			}
			var deptIds []int64
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return fmt.Errorf("--dept-ids 元素须为整数: %w", err)
				}
				deptIds = append(deptIds, n)
			}
			if len(deptIds) == 0 {
				return fmt.Errorf("--dept-ids 不能为空列表")
			}
			input["deptIds"] = deptIds

			// 可选 string
			for _, p := range []struct{ flag, key string }{
				{"student-number", "studentNumber"},
				{"email", "email"},
				{"intake", "intake"},
				{"outtake", "outtake"},
			} {
				if v, _ := cmd.Flags().GetString(p.flag); strings.TrimSpace(v) != "" {
					input[p.key] = strings.TrimSpace(v)
				}
			}

			return callMCPToolOnServer("college-contact", "update_alumnus", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(updateAlumnusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "update_alumnus",
				CanonicalPath:  "college-contact.update_alumnus",
				CLIPath:        "college-contact alumni update-alumnus",
				PrimaryCLIPath: "college-contact alumni update-alumnus",
			},
			Description: "更新校友信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "update_alumnus"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "更新校友信息",
				UseWhen:      []string{"需要更新校友信息时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni update-alumnus --staff-id staff001 --name 张三 --dept-ids 12345,67890"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(false)},
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "dept-ids", Property: "input.deptIds", Required: boolPtr(true)},
				{Name: "student-number", Property: "input.studentNumber", Required: boolPtr(false)},
				{Name: "email", Property: "input.email", Required: boolPtr(false)},
				{Name: "intake", Property: "input.intake", Required: boolPtr(false)},
				{Name: "outtake", Property: "input.outtake", Required: boolPtr(false)},
			},
		},
	})

	updateAlumnusCmd.Flags().String("staff-id", "", "校友 staffId（必填）")
	updateAlumnusCmd.Flags().String("name", "", "校友姓名（必填）")
	updateAlumnusCmd.Flags().String("dept-ids", "", "校友部门 ID 列表，逗号分隔（必填）")
	updateAlumnusCmd.Flags().String("student-number", "", "学号")
	updateAlumnusCmd.Flags().String("email", "", "邮箱")
	updateAlumnusCmd.Flags().String("intake", "", "入学年份")
	updateAlumnusCmd.Flags().String("outtake", "", "毕业年份")

	// ── remove-alumnus ────────────────────────────────────────
	removeAlumnusCmd := &cobra.Command{
		Use:   "remove-alumnus",
		Short: "删除校友",
		Long: `从校友部门中删除指定校友。注意：删除操作不可逆。

必填：
  --staff-id       校友 staffId
  --alumni-dept-id 校友部门 ID

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact alumni remove-alumnus --staff-id staff001 --alumni-dept-id 12345 --dry-run
  dws college-contact alumni remove-alumnus --staff-id staff001 --alumni-dept-id 12345 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			staffId, _ := cmd.Flags().GetString("staff-id")
			staffId = strings.TrimSpace(staffId)
			if staffId == "" {
				return fmt.Errorf("--staff-id 为必填参数")
			}

			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "delete_alumnus", map[string]any{
				"input": map[string]any{
					"staffId":      staffId,
					"alumniDeptId": alumniDeptId,
				},
			})
		},
	}

	DeclareLeafMetadata(removeAlumnusCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_alumnus",
				CanonicalPath:  "college-contact.delete_alumnus",
				CLIPath:        "college-contact alumni remove-alumnus",
				PrimaryCLIPath: "college-contact alumni remove-alumnus",
			},
			Description: "删除校友",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_alumnus"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除校友",
				UseWhen:      []string{"需要删除校友时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni remove-alumnus --staff-id 12345 --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "staff-id", Property: "input.staffId", Required: boolPtr(true)},
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	removeAlumnusCmd.Flags().String("staff-id", "", "校友 staffId（必填）")
	removeAlumnusCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	// ── cancel-invite ─────────────────────────────────────────
	cancelAlumniInviteCmd := &cobra.Command{
		Use:   "cancel-invite",
		Short: "取消校友邀请",
		Long: `取消校友邀请记录，返回操作是否成功。注意：取消操作不可逆。

必填：
  --alumni-dept-id  校友部门 ID
  --staff-ids       staffId 列表（作为 inviteId 使用），逗号分隔

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact alumni cancel-invite --alumni-dept-id 12345 --staff-ids staff001,staff002 --dry-run
  dws college-contact alumni cancel-invite --alumni-dept-id 12345 --staff-ids staff001,staff002 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIds = append(staffIds, part)
				}
			}
			if len(staffIds) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "delete_alumni_invite_record", map[string]any{
				"input": map[string]any{
					"alumniDeptId": alumniDeptId,
					"staffIds":     staffIds,
				},
			})
		},
	}

	DeclareLeafMetadata(cancelAlumniInviteCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_alumni_invite_record",
				CanonicalPath:  "college-contact.delete_alumni_invite_record",
				CLIPath:        "college-contact alumni cancel-invite",
				PrimaryCLIPath: "college-contact alumni cancel-invite",
			},
			Description: "取消校友邀请",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_alumni_invite_record"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "取消校友邀请",
				UseWhen:      []string{"需要取消校友邀请时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni cancel-invite --alumni-dept-id 12345 --staff-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
			},
		},
	})

	cancelAlumniInviteCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")
	cancelAlumniInviteCmd.Flags().String("staff-ids", "", "staffId 列表（作为 inviteId 使用），逗号分隔（必填）")

	// ── create-group ──────────────────────────────────────────
	createAlumniGroupCmd := &cobra.Command{
		Use:   "create-group",
		Short: "创建校友群",
		Long: `为指定校友部门创建校友群，返回群信息（corpId/deptId/cid/name/owner）。

必填：
  --alumni-dept-id  校友部门 ID`,
		Example: `  dws college-contact alumni create-group --alumni-dept-id 12345 --dry-run
  dws college-contact alumni create-group --alumni-dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "create_alumni_group", map[string]any{
				"input": map[string]any{
					"alumniDeptId": alumniDeptId,
				},
			})
		},
	}

	DeclareLeafMetadata(createAlumniGroupCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "create_alumni_group",
				CanonicalPath:  "college-contact.create_alumni_group",
				CLIPath:        "college-contact alumni create-group",
				PrimaryCLIPath: "college-contact alumni create-group",
			},
			Description: "创建校友群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "create_alumni_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建校友群",
				UseWhen:      []string{"需要创建校友群时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni create-group --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	createAlumniGroupCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	// ── disband-group ─────────────────────────────────────────
	disbandAlumniGroupCmd := &cobra.Command{
		Use:   "disband-group",
		Short: "解散校友群",
		Long: `解散指定校友部门的校友群。注意：解散操作不可逆。

必填：
  --alumni-dept-id  校友部门 ID

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact alumni disband-group --alumni-dept-id 12345 --dry-run
  dws college-contact alumni disband-group --alumni-dept-id 12345 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			alumniDeptIdRaw, _ := cmd.Flags().GetString("alumni-dept-id")
			alumniDeptIdRaw = strings.TrimSpace(alumniDeptIdRaw)
			if alumniDeptIdRaw == "" {
				return fmt.Errorf("--alumni-dept-id 为必填参数")
			}
			alumniDeptId, err := strconv.ParseInt(alumniDeptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--alumni-dept-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "disband_alumni_group", map[string]any{
				"input": map[string]any{
					"alumniDeptId": alumniDeptId,
				},
			})
		},
	}

	DeclareLeafMetadata(disbandAlumniGroupCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "disband_alumni_group",
				CanonicalPath:  "college-contact.disband_alumni_group",
				CLIPath:        "college-contact alumni disband-group",
				PrimaryCLIPath: "college-contact alumni disband-group",
			},
			Description: "解散校友群",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "disband_alumni_group"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "解散校友群",
				UseWhen:      []string{"需要解散校友群时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni disband-group --alumni-dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "alumni-dept-id", Property: "input.alumniDeptId", Required: boolPtr(true)},
			},
		},
	})

	disbandAlumniGroupCmd.Flags().String("alumni-dept-id", "", "校友部门 ID（必填）")

	// ── get-alumni-org-from-graduate ───────────────────────────
	getAlumniOrgFromGraduateCmd := &cobra.Command{
		Use:   "get-alumni-org-from-graduate",
		Short: "查询毕业生校友组织",
		Long:  `查询毕业生校友组织信息，无入参，直接返回毕业生校友组织信息。`,
		Example: `  dws college-contact alumni get-alumni-org-from-graduate --dry-run
  dws college-contact alumni get-alumni-org-from-graduate -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("college-contact", "get_alumni_org_from_graduate", map[string]any{
				"input": map[string]any{},
			})
		},
	}

	DeclareLeafMetadata(getAlumniOrgFromGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_alumni_org_from_graduate",
				CanonicalPath:  "college-contact.get_alumni_org_from_graduate",
				CLIPath:        "college-contact alumni get-alumni-org-from-graduate",
				PrimaryCLIPath: "college-contact alumni get-alumni-org-from-graduate",
			},
			Description: "查询毕业生校友组织",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_alumni_org_from_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询毕业生校友组织",
				UseWhen:      []string{"需要查询毕业生校友组织时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni get-alumni-org-from-graduate --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ── create-alumni-org ──────────────────────────────────────
	createAlumniOrgCmd := &cobra.Command{
		Use:   "create-alumni-org",
		Short: "创建校友会组织",
		Long: `创建校友会组织，返回新组织信息。

必填：
  --org-name  校友会组织名称`,
		Example: `  dws college-contact alumni create-alumni-org --org-name 某大学校友会 --dry-run
  dws college-contact alumni create-alumni-org --org-name 某大学校友会 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName, _ := cmd.Flags().GetString("org-name")
			orgName = strings.TrimSpace(orgName)
			if orgName == "" {
				return fmt.Errorf("--org-name 为必填参数")
			}

			return callMCPToolOnServer("college-contact", "create_alumni_org", map[string]any{
				"input": map[string]any{
					"orgName": orgName,
				},
			})
		},
	}

	DeclareLeafMetadata(createAlumniOrgCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "create_alumni_org",
				CanonicalPath:  "college-contact.create_alumni_org",
				CLIPath:        "college-contact alumni create-alumni-org",
				PrimaryCLIPath: "college-contact alumni create-alumni-org",
			},
			Description: "创建校友会组织",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "create_alumni_org"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建校友会组织",
				UseWhen:      []string{"需要创建校友会组织时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni create-alumni-org --org-name 测试"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "org-name", Property: "input.orgName", Required: boolPtr(true)},
			},
		},
	})

	createAlumniOrgCmd.Flags().String("org-name", "", "校友会组织名称（必填）")

	// ── add-alumni-org-main-admins ────────────────────────────
	addAlumniOrgMainAdminsCmd := &cobra.Command{
		Use:   "add-alumni-org-main-admins",
		Short: "添加校友会组织管理员",
		Long: `为校友会组织添加管理员（主管理员）。

必填：
  --admin-user-ids  管理员 userId 列表，逗号分隔`,
		Example: `  dws college-contact alumni add-alumni-org-main-admins --admin-user-ids user001,user002 --dry-run
  dws college-contact alumni add-alumni-org-main-admins --admin-user-ids user001 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("admin-user-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--admin-user-ids 为必填参数")
			}
			var adminUserIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					adminUserIds = append(adminUserIds, part)
				}
			}
			if len(adminUserIds) == 0 {
				return fmt.Errorf("--admin-user-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "add_alumni_org_main_admins", map[string]any{
				"input": map[string]any{
					"adminUserIds": adminUserIds,
				},
			})
		},
	}

	DeclareLeafMetadata(addAlumniOrgMainAdminsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "add_alumni_org_main_admins",
				CanonicalPath:  "college-contact.add_alumni_org_main_admins",
				CLIPath:        "college-contact alumni add-alumni-org-main-admins",
				PrimaryCLIPath: "college-contact alumni add-alumni-org-main-admins",
			},
			Description: "添加校友会组织管理员",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "add_alumni_org_main_admins"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "添加校友会组织管理员",
				UseWhen:      []string{"需要添加校友会组织管理员时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact alumni add-alumni-org-main-admins --admin-user-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "admin-user-ids", Property: "input.adminUserIds", Required: boolPtr(true)},
			},
		},
	})

	addAlumniOrgMainAdminsCmd.Flags().String("admin-user-ids", "", "管理员 userId 列表，逗号分隔（必填）")

	// ════════════════════════════════════════════════════════════
	// graduate 子命令组 — 毕业年级管理
	// ════════════════════════════════════════════════════════════

	graduateCmd := &cobra.Command{Use: "graduate", Short: "毕业年级管理", RunE: groupRunE}

	// ── query-graduate-years ──────────────────────────────────
	queryGraduateYearsCmd := &cobra.Command{
		Use:   "query-graduate-years",
		Short: "查询毕业年级列表",
		Long:  `查询毕业年级列表，无入参，直接返回所有毕业年级。`,
		Example: `  dws college-contact graduate query-graduate-years --dry-run
  dws college-contact graduate query-graduate-years -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("college-contact", "query_graduate_years", map[string]any{
				"input": map[string]any{},
			})
		},
	}

	DeclareLeafMetadata(queryGraduateYearsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_graduate_years",
				CanonicalPath:  "college-contact.query_graduate_years",
				CLIPath:        "college-contact graduate query-graduate-years",
				PrimaryCLIPath: "college-contact graduate query-graduate-years",
			},
			Description: "查询毕业年级列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_graduate_years"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询毕业年级列表",
				UseWhen:      []string{"需要查询毕业年级列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-graduate-years --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ── query-graduate-depts ──────────────────────────────────
	queryGraduateDeptsCmd := &cobra.Command{
		Use:   "query-graduate-depts",
		Short: "查询待毕业部门列表",
		Long: `查询指定部门下的待毕业部门列表。

必填：
  --dept-id  部门 ID

可选：
  --graduate-year  毕业年份`,
		Example: `  dws college-contact graduate query-graduate-depts --dept-id 12345 --dry-run
  dws college-contact graduate query-graduate-depts --dept-id 12345 --graduate-year 2026 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			input := map[string]any{"deptId": deptId}
			if v, _ := cmd.Flags().GetString("graduate-year"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
				if err != nil {
					return fmt.Errorf("--graduate-year 须为整数: %w", err)
				}
				input["graduateYear"] = n
			}

			return callMCPToolOnServer("college-contact", "query_graduate_depts", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(queryGraduateDeptsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_graduate_depts",
				CanonicalPath:  "college-contact.query_graduate_depts",
				CLIPath:        "college-contact graduate query-graduate-depts",
				PrimaryCLIPath: "college-contact graduate query-graduate-depts",
			},
			Description: "查询待毕业部门列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_graduate_depts"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询待毕业部门列表",
				UseWhen:      []string{"需要查询待毕业部门列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-graduate-depts --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "graduate-year", Property: "input.graduateYear", Required: boolPtr(false)},
			},
		},
	})

	queryGraduateDeptsCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	queryGraduateDeptsCmd.Flags().String("graduate-year", "", "毕业年份")

	// ── query-graduate-sub-depts ──────────────────────────────
	queryGraduateSubDeptsCmd := &cobra.Command{
		Use:   "query-graduate-sub-depts",
		Short: "查询毕业子部门列表",
		Long: `查询指定部门的毕业子部门列表。

必填：
  --dept-id  部门 ID`,
		Example: `  dws college-contact graduate query-graduate-sub-depts --dept-id 12345 --dry-run
  dws college-contact graduate query-graduate-sub-depts --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "query_graduate_sub_depts", map[string]any{
				"input": map[string]any{
					"deptId": deptId,
				},
			})
		},
	}

	DeclareLeafMetadata(queryGraduateSubDeptsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_graduate_sub_depts",
				CanonicalPath:  "college-contact.query_graduate_sub_depts",
				CLIPath:        "college-contact graduate query-graduate-sub-depts",
				PrimaryCLIPath: "college-contact graduate query-graduate-sub-depts",
			},
			Description: "查询毕业子部门列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_graduate_sub_depts"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询毕业子部门列表",
				UseWhen:      []string{"需要查询毕业子部门列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-graduate-sub-depts --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	queryGraduateSubDeptsCmd.Flags().String("dept-id", "", "部门 ID（必填）")

	// ── query-page-graduate-users ─────────────────────────────
	queryPageGraduateUsersCmd := &cobra.Command{
		Use:   "query-page-graduate-users",
		Short: "分页查询待毕业学生列表",
		Long: `分页查询指定部门下的待毕业学生列表。

必填：
  --dept-id  部门 ID

可选：
  --graduate-year  毕业年份
  --offset         分页偏移量
  --size           分页大小`,
		Example: `  dws college-contact graduate query-page-graduate-users --dept-id 12345 --dry-run
  dws college-contact graduate query-page-graduate-users --dept-id 12345 --graduate-year 2026 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			input := map[string]any{"deptId": deptId}
			for _, opt := range []struct {
				flag string
				key  string
			}{
				{"graduate-year", "graduateYear"},
				{"offset", "offset"},
				{"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(opt.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", opt.flag, err)
					}
					input[opt.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "query_page_graduate_users", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(queryPageGraduateUsersCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_page_graduate_users",
				CanonicalPath:  "college-contact.query_page_graduate_users",
				CLIPath:        "college-contact graduate query-page-graduate-users",
				PrimaryCLIPath: "college-contact graduate query-page-graduate-users",
			},
			Description: "分页查询待毕业学生列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_page_graduate_users"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "分页查询待毕业学生列表",
				UseWhen:      []string{"需要分页查询待毕业学生列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-page-graduate-users --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "graduate-year", Property: "input.graduateYear", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	queryPageGraduateUsersCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	queryPageGraduateUsersCmd.Flags().String("graduate-year", "", "毕业年份")
	queryPageGraduateUsersCmd.Flags().String("offset", "", "分页偏移量")
	queryPageGraduateUsersCmd.Flags().String("size", "", "分页大小")

	// ── get-task-result ───────────────────────────────────────
	getTaskResultCmd := &cobra.Command{
		Use:   "get-task-result",
		Short: "查询异步任务执行结果",
		Long: `查询异步任务的执行结果（如毕业、恢复等异步操作）。

必填：
  --request-no  异步任务请求号

可选：
  --type  异步任务类型（RESTORE / ADD_GRADUATE / GRADUATE）`,
		Example: `  dws college-contact graduate get-task-result --request-no xxx123 --dry-run
  dws college-contact graduate get-task-result --request-no xxx123 --type GRADUATE -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestNo, _ := cmd.Flags().GetString("request-no")
			requestNo = strings.TrimSpace(requestNo)
			if requestNo == "" {
				return fmt.Errorf("--request-no 为必填参数")
			}

			input := map[string]any{"requestNo": requestNo}
			if v, _ := cmd.Flags().GetString("type"); strings.TrimSpace(v) != "" {
				input["type"] = strings.TrimSpace(v)
			}

			return callMCPToolOnServer("college-contact", "get_task_result", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(getTaskResultCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_task_result",
				CanonicalPath:  "college-contact.get_task_result",
				CLIPath:        "college-contact graduate get-task-result",
				PrimaryCLIPath: "college-contact graduate get-task-result",
			},
			Description: "查询异步任务执行结果",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_task_result"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询异步任务执行结果",
				UseWhen:      []string{"需要查询异步任务执行结果时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate get-task-result --request-no test"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "request-no", Property: "input.requestNo", Required: boolPtr(true)},
				{Name: "type", Property: "input.type", Required: boolPtr(false)},
			},
		},
	})

	getTaskResultCmd.Flags().String("request-no", "", "异步任务请求号（必填）")
	getTaskResultCmd.Flags().String("type", "", "异步任务类型：RESTORE / ADD_GRADUATE / GRADUATE")

	// ── get-alumni-org ────────────────────────────────────────
	getAlumniOrgCmd := &cobra.Command{
		Use:   "get-alumni-org",
		Short: "查询校友组织信息",
		Long:  `查询校友组织信息（组织 ID、名称、主管理员等），无入参，直接返回组织信息。`,
		Example: `  dws college-contact graduate get-alumni-org --dry-run
  dws college-contact graduate get-alumni-org -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("college-contact", "get_alumni_org", map[string]any{
				"input": map[string]any{},
			})
		},
	}

	DeclareLeafMetadata(getAlumniOrgCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_alumni_org",
				CanonicalPath:  "college-contact.get_alumni_org",
				CLIPath:        "college-contact graduate get-alumni-org",
				PrimaryCLIPath: "college-contact graduate get-alumni-org",
			},
			Description: "查询校友组织信息",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_alumni_org"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询校友组织信息",
				UseWhen:      []string{"需要查询校友组织信息时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate get-alumni-org --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ── query-restore-sub-depts ──────────────────────────────
	queryRestoreSubDeptsCmd := &cobra.Command{
		Use:   "query-restore-sub-depts",
		Short: "查询可恢复子部门列表",
		Long: `查询指定部门下的可恢复子部门列表。

必填：
  --dept-id  部门 ID`,
		Example: `  dws college-contact graduate query-restore-sub-depts --dept-id 12345 --dry-run
  dws college-contact graduate query-restore-sub-depts --dept-id 12345 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "query_restore_sub_depts", map[string]any{
				"input": map[string]any{
					"deptId": deptId,
				},
			})
		},
	}

	DeclareLeafMetadata(queryRestoreSubDeptsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_restore_sub_depts",
				CanonicalPath:  "college-contact.query_restore_sub_depts",
				CLIPath:        "college-contact graduate query-restore-sub-depts",
				PrimaryCLIPath: "college-contact graduate query-restore-sub-depts",
			},
			Description: "查询可恢复子部门列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_restore_sub_depts"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询可恢复子部门列表",
				UseWhen:      []string{"需要查询可恢复子部门列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-restore-sub-depts --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
			},
		},
	})

	queryRestoreSubDeptsCmd.Flags().String("dept-id", "", "部门 ID（必填）")

	// ── query-dept-deleted-emps ──────────────────────────────
	queryDeptDeletedEmpsCmd := &cobra.Command{
		Use:   "query-dept-deleted-emps",
		Short: "查询部门可恢复员工列表",
		Long: `分页查询指定部门下的可恢复员工列表。

必填：
  --dept-id  部门 ID

可选：
  --offset  分页偏移量
  --size    分页大小`,
		Example: `  dws college-contact graduate query-dept-deleted-emps --dept-id 12345 --dry-run
  dws college-contact graduate query-dept-deleted-emps --dept-id 12345 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			input := map[string]any{"deptId": deptId}
			for _, opt := range []struct {
				flag string
				key  string
			}{
				{"offset", "offset"},
				{"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(opt.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", opt.flag, err)
					}
					input[opt.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "query_dept_deleted_emps", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(queryDeptDeletedEmpsCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_dept_deleted_emps",
				CanonicalPath:  "college-contact.query_dept_deleted_emps",
				CLIPath:        "college-contact graduate query-dept-deleted-emps",
				PrimaryCLIPath: "college-contact graduate query-dept-deleted-emps",
			},
			Description: "查询部门可恢复员工列表",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_dept_deleted_emps"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询部门可恢复员工列表",
				UseWhen:      []string{"需要查询部门可恢复员工列表时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate query-dept-deleted-emps --dept-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	queryDeptDeletedEmpsCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	queryDeptDeletedEmpsCmd.Flags().String("offset", "", "分页偏移量")
	queryDeptDeletedEmpsCmd.Flags().String("size", "", "分页大小")

	// ── search-graduate ───────────────────────────────────────
	searchGraduateCmd := &cobra.Command{
		Use:   "search-graduate",
		Short: "搜索毕业部门与员工",
		Long: `按关键词搜索毕业相关的部门与员工。

必填：
  --keyword  搜索关键词

可选：
  --offset  分页偏移量
  --size    分页大小`,
		Example: `  dws college-contact graduate search-graduate --keyword 计算机 --dry-run
  dws college-contact graduate search-graduate --keyword 计算机 --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyword, _ := cmd.Flags().GetString("keyword")
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				return fmt.Errorf("--keyword 为必填参数")
			}

			input := map[string]any{"keyword": keyword}
			for _, opt := range []struct {
				flag string
				key  string
			}{
				{"offset", "offset"},
				{"size", "size"},
			} {
				if v, _ := cmd.Flags().GetString(opt.flag); strings.TrimSpace(v) != "" {
					n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
					if err != nil {
						return fmt.Errorf("--%s 须为整数: %w", opt.flag, err)
					}
					input[opt.key] = n
				}
			}

			return callMCPToolOnServer("college-contact", "search_graduate", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(searchGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "search_graduate",
				CanonicalPath:  "college-contact.search_graduate",
				CLIPath:        "college-contact graduate search-graduate",
				PrimaryCLIPath: "college-contact graduate search-graduate",
			},
			Description: "搜索毕业部门与员工",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "search_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "搜索毕业部门与员工",
				UseWhen:      []string{"需要搜索毕业部门与员工时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate search-graduate --keyword 计算机"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "keyword", Property: "input.keyword", Required: boolPtr(true)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	searchGraduateCmd.Flags().String("keyword", "", "搜索关键词（必填）")
	searchGraduateCmd.Flags().String("offset", "", "分页偏移量")
	searchGraduateCmd.Flags().String("size", "", "分页大小")

	// ── commit-graduate ───────────────────────────────────────
	commitGraduateCmd := &cobra.Command{
		Use:   "commit-graduate",
		Short: "提交毕业",
		Long: `对指定部门提交毕业操作，发起异步毕业任务，返回异步任务编号。
注意：提交毕业不可逆。

必填：
  --graduate-dept-ids  待毕业部门 ID 列表，逗号分隔
  --graduate-year      毕业年份

可选：
  --request-no  异步任务请求号

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate commit-graduate --graduate-dept-ids 12345,12346 --graduate-year 2026 --dry-run
  dws college-contact graduate commit-graduate --graduate-dept-ids 12345,12346 --graduate-year 2026 --request-no xxx123 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("graduate-dept-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--graduate-dept-ids 为必填参数")
			}
			var graduateDeptIds []int64
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return fmt.Errorf("--graduate-dept-ids 须为逗号分隔的整数列表: %w", err)
				}
				graduateDeptIds = append(graduateDeptIds, n)
			}
			if len(graduateDeptIds) == 0 {
				return fmt.Errorf("--graduate-dept-ids 不能为空列表")
			}

			graduateYearRaw, _ := cmd.Flags().GetString("graduate-year")
			graduateYearRaw = strings.TrimSpace(graduateYearRaw)
			if graduateYearRaw == "" {
				return fmt.Errorf("--graduate-year 为必填参数")
			}
			graduateYear, err := strconv.ParseInt(graduateYearRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--graduate-year 须为整数: %w", err)
			}

			input := map[string]any{
				"graduateDeptIds": graduateDeptIds,
				"graduateYear":    graduateYear,
			}
			if v, _ := cmd.Flags().GetString("request-no"); strings.TrimSpace(v) != "" {
				input["requestNo"] = strings.TrimSpace(v)
			}

			return callMCPToolOnServer("college-contact", "commit_graduate", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(commitGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "commit_graduate",
				CanonicalPath:  "college-contact.commit_graduate",
				CLIPath:        "college-contact graduate commit-graduate",
				PrimaryCLIPath: "college-contact graduate commit-graduate",
			},
			Description: "提交毕业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "commit_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "提交毕业",
				UseWhen:      []string{"需要提交毕业时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate commit-graduate --graduate-dept-ids 12345 --graduate-year 2024"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "graduate-dept-ids", Property: "input.graduateDeptIds", Required: boolPtr(true)},
				{Name: "graduate-year", Property: "input.graduateYear", Required: boolPtr(true)},
				{Name: "request-no", Property: "input.requestNo", Required: boolPtr(false)},
			},
		},
	})

	commitGraduateCmd.Flags().String("graduate-dept-ids", "", "待毕业部门 ID 列表，逗号分隔（必填）")
	commitGraduateCmd.Flags().String("graduate-year", "", "毕业年份（必填）")
	commitGraduateCmd.Flags().String("request-no", "", "异步任务请求号")

	// ── all-graduate ──────────────────────────────────────────
	allGraduateCmd := &cobra.Command{
		Use:   "all-graduate",
		Short: "全部毕业",
		Long: `对指定毕业年份的全部待毕业部门提交毕业操作。
注意：全部毕业不可逆。

必填：
  --graduate-year  毕业年份

可选：
  --request-no  异步任务请求号

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate all-graduate --graduate-year 2026 --dry-run
  dws college-contact graduate all-graduate --graduate-year 2026 --request-no xxx123 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			graduateYearRaw, _ := cmd.Flags().GetString("graduate-year")
			graduateYearRaw = strings.TrimSpace(graduateYearRaw)
			if graduateYearRaw == "" {
				return fmt.Errorf("--graduate-year 为必填参数")
			}
			graduateYear, err := strconv.ParseInt(graduateYearRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--graduate-year 须为整数: %w", err)
			}

			input := map[string]any{"graduateYear": graduateYear}
			if v, _ := cmd.Flags().GetString("request-no"); strings.TrimSpace(v) != "" {
				input["requestNo"] = strings.TrimSpace(v)
			}

			return callMCPToolOnServer("college-contact", "all_graduate", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(allGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "all_graduate",
				CanonicalPath:  "college-contact.all_graduate",
				CLIPath:        "college-contact graduate all-graduate",
				PrimaryCLIPath: "college-contact graduate all-graduate",
			},
			Description: "全部毕业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "all_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "全部毕业",
				UseWhen:      []string{"需要全部毕业时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate all-graduate --graduate-year 2024"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "graduate-year", Property: "input.graduateYear", Required: boolPtr(true)},
				{Name: "request-no", Property: "input.requestNo", Required: boolPtr(false)},
			},
		},
	})

	allGraduateCmd.Flags().String("graduate-year", "", "毕业年份（必填）")
	allGraduateCmd.Flags().String("request-no", "", "异步任务请求号")

	// ── batch-graduate ────────────────────────────────────────
	batchGraduateCmd := &cobra.Command{
		Use:   "batch-graduate",
		Short: "批量毕业",
		Long: `对指定部门下的学生批量提交毕业，返回批量毕业结果（成功/失败数量）。
注意：批量毕业不可逆。

必填：
  --dept-id     部门 ID
  --staff-ids   学生 staffId 列表，逗号分隔

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate batch-graduate --dept-id 12345 --staff-ids staff001,staff002 --dry-run
  dws college-contact graduate batch-graduate --dept-id 12345 --staff-ids staff001,staff002 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIds = append(staffIds, part)
				}
			}
			if len(staffIds) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "batch_graduate", map[string]any{
				"input": map[string]any{
					"deptId":   deptId,
					"staffIds": staffIds,
				},
			})
		},
	}

	DeclareLeafMetadata(batchGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "batch_graduate",
				CanonicalPath:  "college-contact.batch_graduate",
				CLIPath:        "college-contact graduate batch-graduate",
				PrimaryCLIPath: "college-contact graduate batch-graduate",
			},
			Description: "批量毕业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "batch_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量毕业",
				UseWhen:      []string{"需要批量毕业时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate batch-graduate --dept-id 12345 --staff-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
			},
		},
	})

	batchGraduateCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	batchGraduateCmd.Flags().String("staff-ids", "", "学生 staffId 列表，逗号分隔（必填）")

	// ── delete-and-graduate ───────────────────────────────────
	deleteAndGraduateCmd := &cobra.Command{
		Use:   "delete-and-graduate",
		Short: "删除并毕业",
		Long: `对指定部门下的学生执行删除并毕业操作，返回批量毕业结果（成功/失败数量）。
注意：删除并毕业不可逆。

必填：
  --dept-id     部门 ID
  --staff-ids   学生 staffId 列表，逗号分隔

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate delete-and-graduate --dept-id 12345 --staff-ids staff001,staff002 --dry-run
  dws college-contact graduate delete-and-graduate --dept-id 12345 --staff-ids staff001,staff002 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIds = append(staffIds, part)
				}
			}
			if len(staffIds) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "delete_and_graduate", map[string]any{
				"input": map[string]any{
					"deptId":   deptId,
					"staffIds": staffIds,
				},
			})
		},
	}

	DeclareLeafMetadata(deleteAndGraduateCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_and_graduate",
				CanonicalPath:  "college-contact.delete_and_graduate",
				CLIPath:        "college-contact graduate delete-and-graduate",
				PrimaryCLIPath: "college-contact graduate delete-and-graduate",
			},
			Description: "删除并毕业",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_and_graduate"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除并毕业",
				UseWhen:      []string{"需要删除并毕业时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate delete-and-graduate --dept-id 12345 --staff-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
			},
		},
	})

	deleteAndGraduateCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	deleteAndGraduateCmd.Flags().String("staff-ids", "", "学生 staffId 列表，逗号分隔（必填）")

	// ── batch-delete-pending ──────────────────────────────────
	batchDeletePendingCmd := &cobra.Command{
		Use:   "batch-delete-pending",
		Short: "批量删除待毕业学生",
		Long: `对指定部门下的待毕业学生执行批量删除，返回操作是否成功。
注意：批量删除不可逆。

必填：
  --dept-id     部门 ID
  --staff-ids   学生 staffId 列表，逗号分隔

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate batch-delete-pending --dept-id 12345 --staff-ids staff001,staff002 --dry-run
  dws college-contact graduate batch-delete-pending --dept-id 12345 --staff-ids staff001,staff002 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIds = append(staffIds, part)
				}
			}
			if len(staffIds) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}

			return callMCPToolOnServer("college-contact", "batch_delete_pending", map[string]any{
				"input": map[string]any{
					"deptId":   deptId,
					"staffIds": staffIds,
				},
			})
		},
	}

	DeclareLeafMetadata(batchDeletePendingCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "batch_delete_pending",
				CanonicalPath:  "college-contact.batch_delete_pending",
				CLIPath:        "college-contact graduate batch-delete-pending",
				PrimaryCLIPath: "college-contact graduate batch-delete-pending",
			},
			Description: "批量删除待毕业学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "batch_delete_pending"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量删除待毕业学生",
				UseWhen:      []string{"需要批量删除待毕业学生时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate batch-delete-pending --dept-id 12345 --staff-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
			},
		},
	})

	batchDeletePendingCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	batchDeletePendingCmd.Flags().String("staff-ids", "", "学生 staffId 列表，逗号分隔（必填）")

	// ── batch-update-pending ──────────────────────────────────
	batchUpdatePendingCmd := &cobra.Command{
		Use:   "batch-update-pending",
		Short: "批量更新待毕业学生",
		Long: `批量更新指定部门下待毕业学生的毕业年份，返回操作是否成功。
注意：批量更新影响面大，请谨慎执行。

必填：
  --dept-id        部门 ID
  --staff-ids      学生 staffId 列表，逗号分隔
  --graduate-year  毕业年份

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate batch-update-pending --dept-id 12345 --staff-ids staff001,staff002 --graduate-year 2026 --dry-run
  dws college-contact graduate batch-update-pending --dept-id 12345 --staff-ids staff001,staff002 --graduate-year 2026 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			deptIdRaw, _ := cmd.Flags().GetString("dept-id")
			deptIdRaw = strings.TrimSpace(deptIdRaw)
			if deptIdRaw == "" {
				return fmt.Errorf("--dept-id 为必填参数")
			}
			deptId, err := strconv.ParseInt(deptIdRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--dept-id 须为整数: %w", err)
			}

			raw, _ := cmd.Flags().GetString("staff-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--staff-ids 为必填参数")
			}
			var staffIds []string
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					staffIds = append(staffIds, part)
				}
			}
			if len(staffIds) == 0 {
				return fmt.Errorf("--staff-ids 不能为空列表")
			}

			graduateYearRaw, _ := cmd.Flags().GetString("graduate-year")
			graduateYearRaw = strings.TrimSpace(graduateYearRaw)
			if graduateYearRaw == "" {
				return fmt.Errorf("--graduate-year 为必填参数")
			}
			graduateYear, err := strconv.ParseInt(graduateYearRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--graduate-year 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "batch_update_pending", map[string]any{
				"input": map[string]any{
					"deptId":       deptId,
					"staffIds":     staffIds,
					"graduateYear": graduateYear,
				},
			})
		},
	}

	DeclareLeafMetadata(batchUpdatePendingCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "batch_update_pending",
				CanonicalPath:  "college-contact.batch_update_pending",
				CLIPath:        "college-contact graduate batch-update-pending",
				PrimaryCLIPath: "college-contact graduate batch-update-pending",
			},
			Description: "批量更新待毕业学生",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "batch_update_pending"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "批量更新待毕业学生",
				UseWhen:      []string{"需要批量更新待毕业学生时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate batch-update-pending --dept-id 12345 --staff-ids staff001,staff002 --graduate-year 2026"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "dept-id", Property: "input.deptId", Required: boolPtr(true)},
				{Name: "staff-ids", Property: "input.staffIds", Required: boolPtr(true)},
				{Name: "graduate-year", Property: "input.graduateYear", Required: boolPtr(true)},
			},
		},
	})

	batchUpdatePendingCmd.Flags().String("dept-id", "", "部门 ID（必填）")
	batchUpdatePendingCmd.Flags().String("staff-ids", "", "学生 staffId 列表，逗号分隔（必填）")
	batchUpdatePendingCmd.Flags().String("graduate-year", "", "毕业年份（必填）")

	// ── commit-restore ───────────────────────────────────────
	commitRestoreCmd := &cobra.Command{
		Use:   "commit-restore",
		Short: "提交恢复",
		Long: `对指定部门提交恢复操作，发起异步恢复任务，返回异步任务编号。
注意：提交恢复不可逆。

必填：
  --graduate-dept-ids  待毕业部门 ID 列表，逗号分隔

可选：
  --request-no  异步任务请求号

非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact graduate commit-restore --graduate-dept-ids 12345,12346 --dry-run
  dws college-contact graduate commit-restore --graduate-dept-ids 12345,12346 --request-no xxx123 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, _ := cmd.Flags().GetString("graduate-dept-ids")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return fmt.Errorf("--graduate-dept-ids 为必填参数")
			}
			var graduateDeptIds []int64
			for _, part := range strings.Split(raw, ",") {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				n, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					return fmt.Errorf("--graduate-dept-ids 须为逗号分隔的整数列表: %w", err)
				}
				graduateDeptIds = append(graduateDeptIds, n)
			}
			if len(graduateDeptIds) == 0 {
				return fmt.Errorf("--graduate-dept-ids 不能为空列表")
			}

			input := map[string]any{"graduateDeptIds": graduateDeptIds}
			if v, _ := cmd.Flags().GetString("request-no"); strings.TrimSpace(v) != "" {
				input["requestNo"] = strings.TrimSpace(v)
			}

			return callMCPToolOnServer("college-contact", "commit_restore", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(commitRestoreCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "commit_restore",
				CanonicalPath:  "college-contact.commit_restore",
				CLIPath:        "college-contact graduate commit-restore",
				PrimaryCLIPath: "college-contact graduate commit-restore",
			},
			Description: "提交恢复",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "commit_restore"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "提交恢复",
				UseWhen:      []string{"需要提交恢复时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact graduate commit-restore --graduate-dept-ids 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "graduate-dept-ids", Property: "input.graduateDeptIds", Required: boolPtr(true)},
				{Name: "request-no", Property: "input.requestNo", Required: boolPtr(false)},
			},
		},
	})

	commitRestoreCmd.Flags().String("graduate-dept-ids", "", "待毕业部门 ID 列表，逗号分隔（必填）")
	commitRestoreCmd.Flags().String("request-no", "", "异步任务请求号")

	// ── group 子命令组 ────────────────────────────────────────
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "规则管理",
		RunE:  groupRunE,
	}

	// ── query-group-rule ─────────────────────────────────────
	queryGroupRuleCmd := &cobra.Command{
		Use:   "query-group-rule",
		Short: "查询规则",
		Long: `分页查询规则列表，支持按规则名称过滤。

可选：
  --name    规则名称
  --offset  分页偏移量
  --size    分页大小`,
		Example: `  dws college-contact group query-group-rule
  dws college-contact group query-group-rule --name 分组规则
  dws college-contact group query-group-rule --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("name"); strings.TrimSpace(v) != "" {
				input["name"] = strings.TrimSpace(v)
			}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}

			return callMCPToolOnServer("college-contact", "query_group_rule", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(queryGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_group_rule",
				CanonicalPath:  "college-contact.query_group_rule",
				CLIPath:        "college-contact group query-group-rule",
				PrimaryCLIPath: "college-contact group query-group-rule",
			},
			Description: "查询规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询规则",
				UseWhen:      []string{"需要查询规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group query-group-rule --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "name", Property: "input.name", Required: boolPtr(false)},
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	queryGroupRuleCmd.Flags().String("name", "", "规则名称")
	queryGroupRuleCmd.Flags().String("offset", "", "分页偏移量")
	queryGroupRuleCmd.Flags().String("size", "", "分页大小")

	// ── get-group-rule-schedule ──────────────────────────────
	getGroupRuleScheduleCmd := &cobra.Command{
		Use:   "get-group-rule-schedule",
		Short: "查询规则调度",
		Long:  `查询规则调度 cron 表达式，无入参。`,
		Example: `  dws college-contact group get-group-rule-schedule
  dws college-contact group get-group-rule-schedule -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return callMCPToolOnServer("college-contact", "get_group_rule_schedule", map[string]any{
				"input": map[string]any{},
			})
		},
	}

	DeclareLeafMetadata(getGroupRuleScheduleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "get_group_rule_schedule",
				CanonicalPath:  "college-contact.get_group_rule_schedule",
				CLIPath:        "college-contact group get-group-rule-schedule",
				PrimaryCLIPath: "college-contact group get-group-rule-schedule",
			},
			Description: "查询规则调度",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "get_group_rule_schedule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询规则调度",
				UseWhen:      []string{"需要查询规则调度时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group get-group-rule-schedule --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	// ── query-preview-data ───────────────────────────────────
	queryPreviewDataCmd := &cobra.Command{
		Use:   "query-preview-data",
		Short: "查询规则预览数据",
		Long: `分页查询规则预览数据。

可选：
  --offset  分页偏移量
  --size    分页大小`,
		Example: `  dws college-contact group query-preview-data
  dws college-contact group query-preview-data --offset 0 --size 20 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("offset"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("--offset 须为整数: %w", err)
				}
				input["offset"] = n
			}
			if v, _ := cmd.Flags().GetString("size"); strings.TrimSpace(v) != "" {
				n, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("--size 须为整数: %w", err)
				}
				input["size"] = n
			}

			return callMCPToolOnServer("college-contact", "query_preview_data", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(queryPreviewDataCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "read", Risk: "low",
			Confirmation: "not_required", Idempotency: "idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "query_preview_data",
				CanonicalPath:  "college-contact.query_preview_data",
				CLIPath:        "college-contact group query-preview-data",
				PrimaryCLIPath: "college-contact group query-preview-data",
			},
			Description: "查询规则预览数据",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "query_preview_data"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "查询规则预览数据",
				UseWhen:      []string{"需要查询规则预览数据时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group query-preview-data --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "offset", Property: "input.offset", Required: boolPtr(false)},
				{Name: "size", Property: "input.size", Required: boolPtr(false)},
			},
		},
	})

	queryPreviewDataCmd.Flags().String("offset", "", "分页偏移量")
	queryPreviewDataCmd.Flags().String("size", "", "分页大小")

	// ── create-group-rule ────────────────────────────────────
	createGroupRuleCmd := &cobra.Command{
		Use:   "create-group-rule",
		Short: "创建规则",
		Long: `创建分组规则，返回操作是否成功。

必填：
  --name       规则名称
  --tag-code   标签编码
  --dept-type  部门类型

可选：
  --auto-admin  是否自动设置管理员（true/false）`,
		Example: `  dws college-contact group create-group-rule --name 毕业生分组 --tag-code T001 --dept-type college
  dws college-contact group create-group-rule --name 毕业生分组 --tag-code T001 --dept-type college --auto-admin true -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}

			// 必填 string
			for _, pair := range []struct {
				flag string
				key  string
			}{
				{"name", "name"},
				{"tag-code", "tagCode"},
				{"dept-type", "deptType"},
			} {
				v, _ := cmd.Flags().GetString(pair.flag)
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("--%s 为必填参数", pair.flag)
				}
				input[pair.key] = v
			}

			// 可选 bool（使用 String flag 以避免 pflag 对 "--flag false" 的错误解析）
			if cmd.Flags().Changed("auto-admin") {
				aaRaw, _ := cmd.Flags().GetString("auto-admin")
				autoAdmin, err := strconv.ParseBool(strings.TrimSpace(aaRaw))
				if err != nil {
					return fmt.Errorf("--auto-admin 须为 true 或 false: %w", err)
				}
				input["autoAdmin"] = autoAdmin
			}

			return callMCPToolOnServer("college-contact", "create_group_rule", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(createGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "create_group_rule",
				CanonicalPath:  "college-contact.create_group_rule",
				CLIPath:        "college-contact group create-group-rule",
				PrimaryCLIPath: "college-contact group create-group-rule",
			},
			Description: "创建规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "create_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "创建规则",
				UseWhen:      []string{"需要创建规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group create-group-rule --name 毕业生分组 --tag-code T001 --dept-type college"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "name", Property: "input.name", Required: boolPtr(true)},
				{Name: "tag-code", Property: "input.tagCode", Required: boolPtr(true)},
				{Name: "dept-type", Property: "input.deptType", Required: boolPtr(true)},
				{Name: "auto-admin", Property: "input.autoAdmin", Required: boolPtr(false)},
			},
		},
	})

	createGroupRuleCmd.Flags().String("name", "", "规则名称（必填）")
	createGroupRuleCmd.Flags().String("tag-code", "", "标签编码（必填）")
	createGroupRuleCmd.Flags().String("dept-type", "", "部门类型（必填）")
	createGroupRuleCmd.Flags().String("auto-admin", "", "是否自动设置管理员（true/false）")

	// ── delete-group-rule ────────────────────────────────────
	deleteGroupRuleCmd := &cobra.Command{
		Use:   "delete-group-rule",
		Short: "删除规则",
		Long: `删除分组规则，返回操作是否成功。
--rule-id 为必填。

注意：删除操作不可逆。非 --dry-run 预览时必须显式传入 --yes 才会真实执行，
未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact group delete-group-rule --rule-id 1 --dry-run
  dws college-contact group delete-group-rule --rule-id 1 --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleIDRaw, _ := cmd.Flags().GetString("rule-id")
			ruleIDRaw = strings.TrimSpace(ruleIDRaw)
			if ruleIDRaw == "" {
				return fmt.Errorf("--rule-id 为必填参数")
			}
			ruleID, err := strconv.ParseInt(ruleIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--rule-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "delete_group_rule", map[string]any{
				"input": map[string]any{"ruleId": ruleID},
			})
		},
	}

	DeclareLeafMetadata(deleteGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "delete_group_rule",
				CanonicalPath:  "college-contact.delete_group_rule",
				CLIPath:        "college-contact group delete-group-rule",
				PrimaryCLIPath: "college-contact group delete-group-rule",
			},
			Description: "删除规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "delete_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "删除规则",
				UseWhen:      []string{"需要删除规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group delete-group-rule --rule-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "rule-id", Property: "input.ruleId", Required: boolPtr(true)},
			},
		},
	})

	deleteGroupRuleCmd.Flags().String("rule-id", "", "规则 ID（必填）")

	// ── enable-group-rule ────────────────────────────────────
	enableGroupRuleCmd := &cobra.Command{
		Use:   "enable-group-rule",
		Short: "启用规则",
		Long: `启用分组规则，返回操作是否成功。
--rule-id 为必填。`,
		Example: `  dws college-contact group enable-group-rule --rule-id 1
  dws college-contact group enable-group-rule --rule-id 1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleIDRaw, _ := cmd.Flags().GetString("rule-id")
			ruleIDRaw = strings.TrimSpace(ruleIDRaw)
			if ruleIDRaw == "" {
				return fmt.Errorf("--rule-id 为必填参数")
			}
			ruleID, err := strconv.ParseInt(ruleIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--rule-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "enable_group_rule", map[string]any{
				"input": map[string]any{"ruleId": ruleID},
			})
		},
	}

	DeclareLeafMetadata(enableGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "enable_group_rule",
				CanonicalPath:  "college-contact.enable_group_rule",
				CLIPath:        "college-contact group enable-group-rule",
				PrimaryCLIPath: "college-contact group enable-group-rule",
			},
			Description: "启用规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "enable_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "启用规则",
				UseWhen:      []string{"需要启用规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group enable-group-rule --rule-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "rule-id", Property: "input.ruleId", Required: boolPtr(true)},
			},
		},
	})

	enableGroupRuleCmd.Flags().String("rule-id", "", "规则 ID（必填）")

	// ── disable-group-rule ───────────────────────────────────
	disableGroupRuleCmd := &cobra.Command{
		Use:   "disable-group-rule",
		Short: "停用规则",
		Long: `停用分组规则，返回操作是否成功。
--rule-id 为必填。`,
		Example: `  dws college-contact group disable-group-rule --rule-id 1
  dws college-contact group disable-group-rule --rule-id 1 -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleIDRaw, _ := cmd.Flags().GetString("rule-id")
			ruleIDRaw = strings.TrimSpace(ruleIDRaw)
			if ruleIDRaw == "" {
				return fmt.Errorf("--rule-id 为必填参数")
			}
			ruleID, err := strconv.ParseInt(ruleIDRaw, 10, 64)
			if err != nil {
				return fmt.Errorf("--rule-id 须为整数: %w", err)
			}

			return callMCPToolOnServer("college-contact", "disable_group_rule", map[string]any{
				"input": map[string]any{"ruleId": ruleID},
			})
		},
	}

	DeclareLeafMetadata(disableGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "disable_group_rule",
				CanonicalPath:  "college-contact.disable_group_rule",
				CLIPath:        "college-contact group disable-group-rule",
				PrimaryCLIPath: "college-contact group disable-group-rule",
			},
			Description: "停用规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "disable_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "停用规则",
				UseWhen:      []string{"需要停用规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group disable-group-rule --rule-id 12345"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "rule-id", Property: "input.ruleId", Required: boolPtr(true)},
			},
		},
	})

	disableGroupRuleCmd.Flags().String("rule-id", "", "规则 ID（必填）")

	// ── set-group-rule-schedule ──────────────────────────────
	setGroupRuleScheduleCmd := &cobra.Command{
		Use:   "set-group-rule-schedule",
		Short: "设置规则调度",
		Long: `设置规则调度 cron 表达式，返回操作是否成功。

可选：
  --cron  cron 表达式`,
		Example: `  dws college-contact group set-group-rule-schedule --cron "0 0 2 * * ?"
  dws college-contact group set-group-rule-schedule --cron "0 0 2 * * ?" -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{}
			if v, _ := cmd.Flags().GetString("cron"); strings.TrimSpace(v) != "" {
				input["cron"] = strings.TrimSpace(v)
			}

			return callMCPToolOnServer("college-contact", "set_group_rule_schedule", map[string]any{
				"input": input,
			})
		},
	}

	DeclareLeafMetadata(setGroupRuleScheduleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "medium",
			Confirmation: "not_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "set_group_rule_schedule",
				CanonicalPath:  "college-contact.set_group_rule_schedule",
				CLIPath:        "college-contact group set-group-rule-schedule",
				PrimaryCLIPath: "college-contact group set-group-rule-schedule",
			},
			Description: "设置规则调度",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "set_group_rule_schedule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "设置规则调度",
				UseWhen:      []string{"需要设置规则调度时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group set-group-rule-schedule --format json"},
			},
			Parameters: []contract.ParamDecl{
				{Name: "cron", Property: "input.cron", Required: boolPtr(false)},
			},
		},
	})

	setGroupRuleScheduleCmd.Flags().String("cron", "", "cron 表达式")

	// ── execute-group-rule ───────────────────────────────────
	executeGroupRuleCmd := &cobra.Command{
		Use:   "execute-group-rule",
		Short: "立即执行规则",
		Long: `立即执行分组规则（按规则对成员进行分组、建群等批量变更），返回操作是否成功。
无入参。

注意：执行会对成员产生批量变更，影响范围大。非 --dry-run 预览时必须显式传入
--yes 才会真实执行，未传 --yes 直接拒绝；请在向用户展示操作摘要并获得确认后再追加 --yes。`,
		Example: `  dws college-contact group execute-group-rule --dry-run
  dws college-contact group execute-group-rule --dry-run -f json`,
		RunE: func(cmd *cobra.Command, args []string) error {

			return callMCPToolOnServer("college-contact", "execute_group_rule", map[string]any{
				"input": map[string]any{},
			})
		},
	}

	DeclareLeafMetadata(executeGroupRuleCmd, LeafSpec{
		Safety: contract.SafetySpec{
			Effect: "destructive", Risk: "high",
			Confirmation: "user_required", Idempotency: "non_idempotent",
		},
		Contract: LeafContract{
			Identity: contract.ToolIdentitySpec{
				ProductID:      "college-contact",
				Name:           "execute_group_rule",
				CanonicalPath:  "college-contact.execute_group_rule",
				CLIPath:        "college-contact group execute-group-rule",
				PrimaryCLIPath: "college-contact group execute-group-rule",
			},
			Description: "立即执行规则",
			Interface: &contract.InterfaceSpec{
				Mode:         "mcp",
				Availability: "available",
				Ref:          &contract.InterfaceRefSpec{ProductID: "college-contact", RPCName: "execute_group_rule"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "立即执行规则",
				UseWhen:      []string{"需要立即执行规则时"},
				AvoidWhen:    []string{"与高校通讯录无关的操作"},
				Examples:     []string{"dws college-contact group execute-group-rule --format json"},
			},
			Parameters: []contract.ParamDecl{},
		},
	})

	groupCmd.AddCommand(queryGroupRuleCmd, getGroupRuleScheduleCmd, queryPreviewDataCmd, createGroupRuleCmd, deleteGroupRuleCmd, enableGroupRuleCmd, disableGroupRuleCmd, setGroupRuleScheduleCmd, executeGroupRuleCmd)

	alumniCmd.AddCommand(getAlumniDeptTreeCmd, getAlumniDeptInfoCmd, listAlumniCmd, queryAlumnusCmd, searchAlumniCmd, listUnacceptedAlumnusCmd, getAlumniGroupCmd, createAlumniDeptsCmd, updateAlumniDeptCmd, deleteAlumniDeptCmd, updateAlumniDeptManagersCmd, addAlumnusCmd, updateAlumnusCmd, removeAlumnusCmd, cancelAlumniInviteCmd, createAlumniGroupCmd, disbandAlumniGroupCmd, getAlumniOrgFromGraduateCmd, createAlumniOrgCmd, addAlumniOrgMainAdminsCmd)
	graduateCmd.AddCommand(queryGraduateYearsCmd, queryGraduateDeptsCmd, queryGraduateSubDeptsCmd, queryPageGraduateUsersCmd, getTaskResultCmd, getAlumniOrgCmd, queryRestoreSubDeptsCmd, queryDeptDeletedEmpsCmd, searchGraduateCmd, commitGraduateCmd, allGraduateCmd, batchGraduateCmd, deleteAndGraduateCmd, batchDeletePendingCmd, batchUpdatePendingCmd, commitRestoreCmd)

	employeeCmd.AddCommand(getEmployeeDetailCmd, addEmployeeCmd, removeEmployeeCmd, changeEmployeeTypeCmd, changeEmployeeDeptCmd, sendActiveSmsCmd, listEmployeesCmd, listUnacceptedEmployeesCmd, listUnactiveEmployeesCmd, upgradeStatusCmd, startUpgradeCmd)
	root.AddCommand(deptCmd, employeeCmd, alumniCmd, graduateCmd, groupCmd)

	return root
}
