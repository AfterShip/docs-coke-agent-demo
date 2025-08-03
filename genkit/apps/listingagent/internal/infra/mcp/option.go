package mcp

import (
	"github.com/firebase/genkit/go/plugins/mcp"
)

// Option MCP配置选项
type Option struct {
	Name       string             `json:"name" mapstructure:"name" validate:"required"`
	MCPServers []MCPServerOptions `json:"mcp_servers" mapstructure:"mcp_servers" validate:"required"`
}

// MCPServerOptions MCP服务器配置选项
type MCPServerOptions struct {
	Name   string           `json:"name" mapstructure:"name" validate:"required"`
	Config MCPClientOptions `json:"config" mapstructure:"config" validate:"required"`
}

// MCPClientOptions MCP客户端配置选项
type MCPClientOptions struct {
	Name    string       `json:"name" mapstructure:"name" validate:"required"`
	Version string       `json:"version" mapstructure:"version" validate:"required"`
	Stdio   *StdioConfig `json:"stdio" mapstructure:"stdio" validate:"required"`
}

// StdioConfig 标准输入输出配置
type StdioConfig struct {
	Command string   `json:"command" mapstructure:"command" validate:"required"`
	Args    []string `json:"args" mapstructure:"args" validate:"omitempty"`
}

// NewOption 创建新的MCP配置选项
func NewOption() *Option {
	return &Option{
		Name:       "mcp-manager",
		MCPServers: []MCPServerOptions{},
	}
}

// ToMCPManagerOptions 将配置选项转换为MCP管理器选项
func (o *Option) ToMCPManagerOptions() mcp.MCPManagerOptions {
	var servers []mcp.MCPServerConfig
	for _, server := range o.MCPServers {
		servers = append(servers, mcp.MCPServerConfig{
			Name: server.Name,
			Config: mcp.MCPClientOptions{
				Name:    server.Config.Name,
				Version: server.Config.Version,
				Stdio: &mcp.StdioConfig{
					Command: server.Config.Stdio.Command,
					Args:    server.Config.Stdio.Args,
				},
			},
		})
	}

	return mcp.MCPManagerOptions{
		Name:       o.Name,
		MCPServers: servers,
	}
}
