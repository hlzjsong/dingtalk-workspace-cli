// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edition

import (
	"testing"

	"github.com/DingTalk-Real-AI/dingtalk-workspace-cli/pkg/agentproduct"
)

func TestClawTypeDefaultsToOSSValue(t *testing.T) {
	t.Setenv(agentproduct.EnvName, "")
	prev := Get()
	defer Override(prev)

	Override(defaultHooks())
	if got := ClawType(); got != DefaultOSSClawType {
		t.Fatalf("ClawType() = %q, want %q", got, DefaultOSSClawType)
	}
}

func TestClawTypeUsesOverlayValue(t *testing.T) {
	t.Setenv(agentproduct.EnvName, "")
	prev := Get()
	defer Override(prev)

	Override(&Hooks{Name: "overlay", ClawTypeValue: "wukong"})
	if got := ClawType(); got != "wukong" {
		t.Fatalf("ClawType() = %q, want overlay value %q", got, "wukong")
	}
}

func TestClawTypeUsesValidAgentProduct(t *testing.T) {
	t.Setenv(agentproduct.EnvName, " qwenwork ")
	prev := Get()
	defer Override(prev)

	Override(&Hooks{Name: "overlay", ClawTypeValue: "wukong"})
	if got := ClawType(); got != "qwenwork" {
		t.Fatalf("ClawType() = %q, want qwenwork", got)
	}
}

func TestClawTypeInvalidAgentProductFallsBackToOverlay(t *testing.T) {
	t.Setenv(agentproduct.EnvName, "qwen work")
	prev := Get()
	defer Override(prev)

	Override(&Hooks{Name: "overlay", ClawTypeValue: "wukong"})
	if got := ClawType(); got != "wukong" {
		t.Fatalf("ClawType() = %q, want overlay fallback wukong", got)
	}
}

func TestOpenStaticServersIncludesCoreProducts(t *testing.T) {
	servers := openStaticServers()
	byID := make(map[string]ServerInfo, len(servers))
	for _, server := range servers {
		byID[server.ID] = server
	}

	required := []string{"aitable", "aitable-helper", "calendar", "todo", "doc", "chat", "mail", "oa"}
	for _, id := range required {
		if _, ok := byID[id]; !ok {
			t.Errorf("openStaticServers() missing required product %q", id)
		}
	}

	helper := byID["aitable-helper"]
	if helper.Endpoint == "" {
		t.Fatal("aitable-helper has empty endpoint")
	}
}

func TestOpenVisibleProductsExcludesCompatibilityOnlyCommands(t *testing.T) {
	visible := openVisibleProducts()
	byID := make(map[string]bool, len(visible))
	for _, id := range visible {
		byID[id] = true
	}
	if byID["conference"] {
		t.Fatal("conference must remain hidden compatibility-only and not appear in VisibleProducts")
	}

	for _, server := range openStaticServers() {
		if server.ID == "conference" {
			t.Fatal("conference must remain compatibility-only and not be added to StaticServers")
		}
	}
	if byID["mcp-meta"] {
		t.Fatal("mcp-meta is helper-only and must not appear in VisibleProducts")
	}
}

func TestOpenSupplementServersIncludesMCPMeta(t *testing.T) {
	servers := openSupplementServers()
	foundMCPMeta := false
	foundWhiteboard := false
	foundRecruit := false
	for _, server := range servers {
		if server.ID == "recruit" {
			foundRecruit = server.Endpoint == "https://mcp-gw.dingtalk.com/server/f69b54ada16c57b603c0e5e1c36f464ba73dcee28d64bb701ff2682c259c0cff" &&
				len(server.Prefixes) == 2 && server.Prefixes[0] == "recruit" && server.Prefixes[1] == "job"
		}
		if server.ID == "whiteboard" {
			foundWhiteboard = server.Endpoint == "https://mcp-gw.dingtalk.com/server/whiteboard"
		}
		if server.ID != "mcp-meta" {
			continue
		}
		foundMCPMeta = true
		if server.Endpoint == "" {
			t.Fatal("mcp-meta has empty endpoint")
		}
		if len(server.Prefixes) != 0 {
			t.Fatal("mcp-meta must remain helper-only without command prefixes")
		}
	}
	if !foundMCPMeta {
		t.Fatal("openSupplementServers() missing mcp-meta")
	}
	if !foundWhiteboard {
		t.Fatal("openSupplementServers() missing helper-only whiteboard endpoint")
	}
	if !foundRecruit {
		t.Fatal("openSupplementServers() missing explicitly wired recruit endpoint")
	}
}

func TestCrossPlatformCoverageOpenSupplementServersIncludesEduEndpoints(t *testing.T) {
	servers := openSupplementServers()
	byID := make(map[string]ServerInfo, len(servers))
	for _, s := range servers {
		byID[s.ID] = s
	}

	edu := []struct {
		id       string
		endpoint string
		prefixes []string
	}{
		{"edu-contact", "https://mcp-gw.dingtalk.com/server/d24759cc1c6e424e2de4e9901ea0202136e6707991ffc33b473878ec1cd688a2", []string{"edu-contact", "edu"}},
		{"edu-group", "https://mcp-gw.dingtalk.com/server/14624b71ac9bc1a03b1b60e5b0403a48b346361f86cc9f555f98f89eb383875a", []string{"edu-group"}},
		{"edu-app", "https://mcp-gw.dingtalk.com/server/905eef591d16e2a1d95b235bcc780ce2fadb6ebe1f25648a279f8a2d97907a1e", []string{"edu-app"}},
		{"edu-familygroup", "https://mcp-gw.dingtalk.com/server/1cd153fb5296df340507c3e9ee20c938f9feeefec3147e9cc32317032f1a2944", []string{"edu-familygroup"}},
		{"college-contact", "https://mcp-gw.dingtalk.com/server/45bb310b388b9c39e0b80e08236782880cb51ad536e1292f9a40933c428a7474", []string{"college-contact"}},
	}

	for _, want := range edu {
		got, ok := byID[want.id]
		if !ok {
			t.Errorf("openSupplementServers() missing edu endpoint %q", want.id)
			continue
		}
		if got.Endpoint != want.endpoint {
			t.Errorf("%s endpoint = %q, want %q", want.id, got.Endpoint, want.endpoint)
		}
		if len(got.Prefixes) != len(want.prefixes) {
			t.Errorf("%s prefixes length = %d, want %d", want.id, len(got.Prefixes), len(want.prefixes))
			continue
		}
		for i, p := range want.prefixes {
			if got.Prefixes[i] != p {
				t.Errorf("%s prefixes[%d] = %q, want %q", want.id, i, got.Prefixes[i], p)
			}
		}
	}
}
