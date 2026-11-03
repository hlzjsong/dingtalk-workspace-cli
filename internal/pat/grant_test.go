// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0

package pat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/corecmd"
	apperrors "github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/errors"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/output"
	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/pkg/edition"
)

type directGrantCaller struct {
	fakeToolCaller
	errors []error
	result *edition.ToolResult
	tools  []string
}

func (f *directGrantCaller) CallTool(ctx context.Context, product, tool string, args map[string]any) (*edition.ToolResult, error) {
	f.tools = append(f.tools, tool)
	_, _ = f.fakeToolCaller.CallTool(ctx, product, tool, args)
	if len(f.errors) >= f.callN && f.errors[f.callN-1] != nil {
		return nil, f.errors[f.callN-1]
	}
	return f.result, nil
}

func newDirectGrantCaller() *directGrantCaller {
	return &directGrantCaller{result: &edition.ToolResult{Content: []edition.ContentBlock{{Type: "text", Text: `{"success":true,"data":{"grantType":"once"}}`}}}}
}

func executeDirectGrant(t *testing.T, caller edition.ToolCaller, argv ...string) (string, error) {
	t.Helper()
	cmd := newGrantCommand(caller)
	cmd.Flags().Bool("yes", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("format", "json", "")
	cmd.SetIn(bytes.NewReader(nil))
	cmd.SetErr(io.Discard)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(argv)
	if err := corecmd.ExecuteForTest(cmd); err != nil {
		return out.String(), err
	}
	_, _, err := output.EmitStoredResult(cmd)
	return out.String(), err
}

func clearGrantEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{agentCodeEnv, agentCodeEnvCompat, sessionIDEnvDingtalk, sessionIDEnvDWS, sessionIDEnvRewind} {
		t.Setenv(key, "")
	}
}

func TestCrossPlatformCoverageGrantDirectSingleScope(t *testing.T) {
	clearGrantEnvironment(t)
	caller := newDirectGrantCaller()
	out, err := executeDirectGrant(t, caller, "chat.message:list", "--grant-type", "once", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if caller.callN != 1 || caller.gotTool != "pat.grant" {
		t.Fatalf("calls = %v; must call only pat.grant", caller.tools)
	}
	scopes, ok := caller.gotArgs["scopes"].([]string)
	if !ok || len(scopes) != 1 || scopes[0] != "chat.message:list" || caller.gotArgs["grantType"] != "once" {
		t.Fatalf("arguments = %#v", caller.gotArgs)
	}
	if _, exists := caller.gotArgs["agentCode"]; exists {
		t.Fatal("unspecified agentCode must be omitted")
	}
	var envelope struct {
		OK      bool           `json:"ok"`
		Outcome string         `json:"outcome"`
		Data    map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil || !envelope.OK || envelope.Outcome != "success" || envelope.Data["executed"] != true {
		t.Fatalf("unified output = %s, error = %v", out, err)
	}
}

func TestCrossPlatformCoverageGrantValidationAndConfirmation(t *testing.T) {
	clearGrantEnvironment(t)
	for _, argv := range [][]string{
		{"--yes"}, {"scope:a", "scope:b", "--yes"}, {" ", "--yes"},
		{"scope:a", "--grant-type", "forever", "--yes"},
		{"scope:a", "--grant-type", "session", "--yes"},
		{"scope:a", "--agentCode", "invalid agent", "--yes"},
	} {
		t.Run(jsonStringForGrantTest(argv), func(t *testing.T) {
			caller := newDirectGrantCaller()
			_, err := executeDirectGrant(t, caller, argv...)
			var typed *apperrors.Error
			if !errors.As(err, &typed) || typed.Category != apperrors.CategoryValidation || caller.callN != 0 {
				t.Fatalf("err = %v, calls = %d", err, caller.callN)
			}
		})
	}
	caller := newDirectGrantCaller()
	if _, err := executeDirectGrant(t, caller, "scope:a"); err == nil || caller.callN != 0 {
		t.Fatalf("unconfirmed invocation: err = %v, calls = %d", err, caller.callN)
	}
}

func jsonStringForGrantTest(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestCrossPlatformCoverageGrantSessionIdentityAndFallback(t *testing.T) {
	clearGrantEnvironment(t)
	t.Setenv(agentCodeEnv, "env-agent")
	t.Setenv(sessionIDEnvDWS, "env-session")
	caller := newDirectGrantCaller()
	caller.errors = []error{errors.New("tool not found")}
	_, err := executeDirectGrant(t, caller, "chat.message:list", "--grant-type", "session", "--agentCode", "flag-agent", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if caller.callN != 2 || caller.tools[0] != "pat.grant" || caller.tools[1] != "个人授权" {
		t.Fatalf("fallback tools = %v", caller.tools)
	}
	if caller.gotArgs["sessionId"] != "env-session" || caller.gotAgentEnv != "flag-agent" || caller.gotDingSessionEnv != "env-session" {
		t.Fatalf("identity not forwarded: args=%v agent=%q session=%q", caller.gotArgs, caller.gotAgentEnv, caller.gotDingSessionEnv)
	}
	if _, ok := caller.gotArgs["scope"]; !ok {
		t.Fatalf("legacy scope argument missing: %v", caller.gotArgs)
	}
	if os.Getenv(agentCodeEnv) != "env-agent" || os.Getenv(sessionIDEnvDingtalk) != "" {
		t.Fatal("temporary authorization context was not restored")
	}
}

func TestCrossPlatformCoverageGrantDryRunAndFailures(t *testing.T) {
	clearGrantEnvironment(t)
	caller := newDirectGrantCaller()
	out, err := executeDirectGrant(t, caller, "scope:a", "--dry-run")
	if err != nil || caller.callN != 0 || !bytes.Contains([]byte(out), []byte(`"executed": false`)) {
		t.Fatalf("dry-run = %s, err=%v, calls=%d", out, err, caller.callN)
	}
	for _, result := range []*edition.ToolResult{
		nil, {},
		{Content: []edition.ContentBlock{{Type: "text", Text: `not json`}}},
		{Content: []edition.ContentBlock{{Type: "text", Text: `null`}}},
		{Content: []edition.ContentBlock{{Type: "text", Text: `{}`}}},
		{Content: []edition.ContentBlock{{Type: "text", Text: `{"success":false,"success":true}`}}},
		{Content: []edition.ContentBlock{{Type: "text", Text: `{"success":false,"code":"PAT_NO_PERMISSION","message":"denied"}`}}},
		{Content: []edition.ContentBlock{{Type: "text", Text: `{"success":true,"data":{"success":false,"code":"PAT_NO_PERMISSION","message":"denied"}}`}}},
	} {
		caller := newDirectGrantCaller()
		caller.result = result
		if _, err := executeDirectGrant(t, caller, "scope:a", "--yes"); err == nil || caller.callN != 1 {
			t.Fatalf("result=%#v err=%v calls=%d", result, err, caller.callN)
		}
	}
	caller = newDirectGrantCaller()
	caller.errors = []error{context.DeadlineExceeded}
	if _, err := executeDirectGrant(t, caller, "scope:a", "--yes"); !errors.Is(err, context.DeadlineExceeded) || caller.callN != 1 {
		t.Fatalf("transport error=%v calls=%d", err, caller.callN)
	}
}

func TestCrossPlatformCoverageGrantResponseSemantics(t *testing.T) {
	clearGrantEnvironment(t)
	caller := newDirectGrantCaller()
	caller.result.Content[0].Text = `{"success":true,"data":{"success":true,"status":"PENDING_CONFIRMATION","authRequestId":"request-1","uid":9007199254740993}}`
	out, err := executeDirectGrant(t, caller, "scope:a", "--yes")
	var typed *apperrors.Error
	if !errors.As(err, &typed) || typed.Category != apperrors.CategoryAuth || typed.Reason != "pat_grant_pending_confirmation" || caller.callN != 1 {
		t.Fatalf("pending authorization was not surfaced: output=%s err=%v", out, err)
	}
	caller = newDirectGrantCaller()
	caller.result.Content[0].Text = `{"success":true,"data":{"success":true,"status":"OK","uid":9007199254740993}}`
	out, err = executeDirectGrant(t, caller, "scope:a", "--yes")
	if err != nil || !bytes.Contains([]byte(out), []byte(`9007199254740993`)) {
		t.Fatalf("number preservation: output=%s err=%v", out, err)
	}
	caller = newDirectGrantCaller()
	caller.result.Content[0].Text = `{"success":true,"data":{"agentCode":"wrong-agent"}}`
	if _, err := executeDirectGrant(t, caller, "scope:a", "--agentCode", "requested-agent", "--yes"); err == nil {
		t.Fatal("mismatched authorization target accepted")
	}
	if _, err := executeDirectGrant(t, nil, "scope:a", "--yes"); err == nil {
		t.Fatal("missing runtime accepted")
	}
}

func TestCrossPlatformCoverageGrantRegistered(t *testing.T) {
	root := &cobra.Command{Use: "dws"}
	RegisterCommands(root, nil)
	cmd, _, err := root.Find([]string{"pat", "grant"})
	if err != nil || cmd.Name() != "grant" {
		t.Fatalf("grant command missing: cmd=%v err=%v", cmd, err)
	}
}
