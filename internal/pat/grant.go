// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0

package pat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd/contract"
	apperrors "github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/errors"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/jsonutil"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/output"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/pkg/edition"
)

// newGrantCommand deliberately bypasses batch planning and batch grants. The
// legacy grantAuth backend handles only one scope, even when it receives scopes.
func newGrantCommand(caller edition.ToolCaller) *cobra.Command {
	return corecmd.New(corecmd.Spec{
		Use:           "grant <scope>",
		Short:         "通过 grantAuth 授予单个权限",
		Long:          "直接调用 pat.grant（后端 grantAuth）授予一个 scope，不调用 pat.batch_grant，并兼容旧的个人授权工具名。批量或按产品授权请使用 pat chmod。真实授权需要用户确认，可用 --yes 明确确认；--dry-run 仅输出请求预览。",
		OutputRollout: output.RolloutUnifiedActive,
		Flags: []corecmd.FlagSpec{
			{Name: "grant-type", Bind: "grantType", Default: grantTypePermanent, Enum: []string{"once", "session", "permanent"}, Usage: "授权策略: once|session|permanent"},
			{Name: "agentCode", OmitEmpty: true, Trim: true, Usage: "Agent 唯一标识；优先使用此参数，其次 DINGTALK_DWS_AGENTCODE，未指定由服务端兜底"},
			{Name: "session-id", Bind: "sessionId", OmitEmpty: true, Trim: true, RequiredWhen: "grant-type is session", Usage: "会话标识；session 模式必填，也可从 DINGTALK_SESSION_ID / DWS_SESSION_ID / REWIND_SESSION_ID 读取"},
		},
		Safety: contract.SafetySpec{
			Effect: "write", Risk: "high", Confirmation: "user_required", Idempotency: "unknown",
		},
		Contract: corecmd.ContractDecl{
			Identity: contract.ToolIdentitySpec{
				ProductID: "pat", Name: "grant", CanonicalPath: "pat.grant", CLIPath: "pat grant", PrimaryCLIPath: "pat grant",
			},
			Description: "通过旧 grantAuth 接口授予单个 PAT scope",
			Positionals: []contract.RuntimeSchemaPositional{
				{Name: "scope", Type: "string", Description: "单个权限 scope，例如 chat.message:list", Required: true, Index: 0},
			},
			Interface: &contract.InterfaceSpec{
				Mode: "mcp", Availability: "available", Ref: &contract.InterfaceRefSpec{ProductID: "pat", RPCName: "pat.grant"},
			},
			DryRun: &contract.DryRunSpec{PreviewKind: contract.DryRunPreviewInvocation, RemoteReads: false},
			Result: &contract.ResultSpec{
				Outcomes: []contract.ResultOutcome{contract.ResultOutcomeSuccess},
				DataSchema: json.RawMessage(`{
					"type":"object",
					"required":["tool","arguments","executed"],
					"properties":{
						"tool":{"type":"string","description":"授权入口 MCP 工具名","const":"pat.grant"},
						"arguments":{"type":"object","description":"授予单个 scope 的规范请求参数"},
						"executed":{"type":"boolean","description":"是否已完成远端授权调用；dry-run 为 false"},
						"response":{"type":"object","description":"远端授权成功响应；dry-run 时省略"}
					},
					"additionalProperties":false
				}`),
				SensitivePaths: []string{"response.patToken", "response.data.patToken"},
			},
			Selection: contract.SelectionSpec{
				AgentSummary: "用户明确要求走旧 grantAuth 接口时，对单个 scope 授权并返回结果",
				UseWhen:      []string{"需要直接调用旧 grantAuth / pat.grant 授予一个明确 scope，且已明确授权类型和应用"},
				AvoidWhen:    []string{"多个 scope、推荐集合或按产品授权时使用 pat chmod", "普通 OAuth 登录使用 auth login"},
				Examples:     []string{"dws pat grant chat.message:list --grant-type once --dry-run --format json"},
			},
		},
		PostMount: func(cmd *cobra.Command) { cmd.Args = cobra.ExactArgs(1) },
		Validate: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(args[0]) == "" {
				return fmt.Errorf("scope must not be empty")
			}
			_, sessionID, err := directGrantIdentity(cmd)
			if err != nil {
				return err
			}
			if commandStringFlag(cmd, "grant-type") == "session" && sessionID == "" {
				return fmt.Errorf("--session-id is required when --grant-type is session")
			}
			return nil
		},
		ResultInvoke: func(c *corecmd.Ctx, args map[string]any) (output.CommandResult, error) {
			agentCode, sessionID, err := directGrantIdentity(c.Command())
			if err != nil {
				return nil, err
			}
			args["scopes"] = []string{strings.TrimSpace(c.Args()[0])}
			if agentCode != "" {
				args["agentCode"] = agentCode
			}
			if sessionID != "" {
				args["sessionId"] = sessionID
			}
			data := map[string]any{"tool": patGrantToolName, "arguments": args, "executed": false}
			if corecmd.BoolFlag(c.Command(), "dry-run") || (caller != nil && caller.DryRun()) {
				return output.Success(data, output.WithDryRun()), nil
			}
			legacyArgs := make(map[string]any, len(args))
			for key, value := range args {
				if key == "scopes" {
					legacyArgs["scope"] = value
				} else {
					legacyArgs[key] = value
				}
			}
			result, err := withPATContextEnv(agentCode, sessionID, func() (*edition.ToolResult, error) {
				return callPATToolWithLegacyFallback(c.Command().Context(), caller, "pat", patGrantToolName, patGrantToolNameLegacyAlias, args, legacyArgs)
			})
			if err != nil {
				return nil, fmt.Errorf("pat grant failed: %w", err)
			}
			response, err := directGrantResponse(result)
			if err != nil {
				return nil, err
			}
			if actual := patResultAgentCode(result); agentCode != "" && actual != "" && actual != agentCode {
				return nil, apperrors.NewAPI("pat grant returned a different agentCode; authorization target does not match the requested agent")
			}
			data["executed"] = true
			data["response"] = response
			return output.Success(data), nil
		},
	})
}

func directGrantResponse(result *edition.ToolResult) (map[string]any, error) {
	raw := []byte(firstToolResultText(result))
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var response map[string]any
	if jsonutil.RejectDuplicateObjectKeys(raw) != nil || decoder.Decode(&response) != nil || len(response) == 0 {
		return nil, apperrors.NewAPI("invalid PAT grant response: expected a non-empty JSON object")
	}
	if err := apperrors.ClassifyMCPResponseText(string(raw)); err != nil {
		return nil, err
	}
	// HSF may retain its ServiceResult envelope around PatAuthGrantResult.
	// Classify that business result too; an outer success only proves RPC success.
	business := response
	if nested, ok := response["data"].(map[string]any); ok {
		business = nested
		nestedJSON, err := json.Marshal(nested)
		if err != nil {
			return nil, err
		}
		if err := apperrors.ClassifyMCPResponseText(string(nestedJSON)); err != nil {
			return nil, err
		}
	}
	if business["status"] == "PENDING_CONFIRMATION" {
		// This legacy response supplies no executable status-query command.
		// Surface the authorization requirement instead of claiming completion
		// or inventing the continuation required by a unified pending result.
		return nil, apperrors.NewAuth("PAT grant is pending user confirmation",
			apperrors.WithReason("pat_grant_pending_confirmation"),
			apperrors.WithHint("请先完成服务端授权确认；当前结果不代表权限已授予。"))
	}
	return response, nil
}

func directGrantIdentity(cmd *cobra.Command) (string, string, error) {
	agentCode, err := resolveAgentCode(commandStringFlag(cmd, "agentCode"))
	if err != nil {
		return "", "", err
	}
	sessionID := strings.TrimSpace(commandStringFlag(cmd, "session-id"))
	if sessionID == "" {
		sessionID = strings.TrimSpace(resolveSessionIDFromEnv())
	}
	return agentCode, sessionID, nil
}
