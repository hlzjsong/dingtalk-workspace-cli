// Copyright 2026 Alibaba Group
// Licensed under the Apache License, Version 2.0 (the "License");

package helpers

import "github.com/spf13/cobra"

// 家校/高校系列（家校通讯录、家校群、家校应用、家庭群、高校通讯录）是开源库
// 显式维护的公开命令，不依赖生成式产品注册表（register_products.go）。
// 其 MCP server 端点由运行时宿主按 serverID 解析。
func init() {
	eduCommands := []struct {
		name    string
		buildFn func() *cobra.Command
	}{
		{"edu-contact", newEduContactCommand},
		{"edu-group", newEduGroupCommand},
		{"edu-app", newEduAppCommand},
		{"edu-familygroup", newEduFamilyGroupCommand},
		{"college-contact", newCollegeContactCommand},
	}
	for _, c := range eduCommands {
		c := c
		RegisterPublic(func() Handler {
			return wukongHandler{name: c.name, buildFn: c.buildFn}
		})
	}
}
