// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0

package app

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/internal/output"
)

func TestCrossPlatformCoveragePATGrantDeliveredContract(t *testing.T) {
	root := NewRootCommand()
	cmd := exactCommandForTest(root, "pat grant")
	if cmd == nil || !output.UsesUnifiedResult(cmd) {
		t.Fatal("missing unified pat grant command")
	}
	full := contactOrgListSchemaLeaf(t, "pat.grant", false)
	compact := contactOrgListSchemaLeaf(t, "pat.grant", true)
	if full["primary_cli_path"] != "pat grant" || full["confirmation"] != "user_required" || full["effect"] != "write" {
		t.Fatalf("identity/safety drift: %#v", full)
	}
	if schemaInterfaceObject(full["interface_ref"])["rpc_name"] != "pat.grant" {
		t.Fatal("grant must bind to pat.grant, not batch_grant")
	}
	if problem := schemaHelpFlagCompletenessProblem("pat.grant", "pat grant", cmd, full); problem != "" {
		t.Fatal(problem)
	}
	params := schemaContractMap(full["parameters"])
	if params["grant-type"]["default"] != "permanent" || params["grant-type"]["property"] != "grantType" || params["session-id"]["property"] != "sessionId" {
		t.Fatalf("parameter mapping/default drift: %#v", params)
	}
	positionals, _ := full["positionals"].([]any)
	if len(positionals) != 1 {
		t.Fatalf("positionals = %#v", positionals)
	}
	scope := positionals[0].(map[string]any)
	if scope["name"] != "scope" || scope["required"] != true || scope["variadic"] == true {
		t.Fatalf("grant must require exactly one scope: %#v", scope)
	}
	if full["result"] == nil || !reflect.DeepEqual(full["result"], compact["result"]) {
		t.Fatal("full and compact must publish the same result contract")
	}
}

func TestCrossPlatformCoveragePATGrantRootDryRun(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"pat", "grant", "chat.message:list", "--grant-type", "once", "--dry-run", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		OK   bool `json:"ok"`
		Data struct {
			Tool      string         `json:"tool"`
			Executed  bool           `json:"executed"`
			Arguments map[string]any `json:"arguments"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid unified JSON: %v; output=%s", err, out.String())
	}
	if !envelope.OK || envelope.Data.Executed || envelope.Data.Tool != "pat.grant" || envelope.Data.Arguments["grantType"] != "once" {
		t.Fatalf("unexpected preview: %s", out.String())
	}
}
