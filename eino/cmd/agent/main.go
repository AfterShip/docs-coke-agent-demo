package main

import (
	"context"
	"eino/pkg/langfuse/api/resources/prompts/types"
	"encoding/json"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/cloudwego/eino-ext/devops"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"log"
	"strings"
)

func main() {
	ctx := context.Background()

	log.Printf("===setup agent tracing===\n")
	setupTracing()

	log.Printf("===setup debug console===\n")
	_ = devops.Init(ctx)

	log.Printf("===load system prompts ===\n")
	langfuseAPI, _ := initLangfuseAPIClient()
	mainSystemPrompt, _ := getPromptByName(ctx, langfuseAPI, "eino-demo-main-system")
	mainSystemMessage := getSystemMessage(mainSystemPrompt)

	config := &react.AgentConfig{MaxStep: 25}

	// 创建 LLM
	log.Printf("===create llm===\n")
	cm := createClaudeChatModel(ctx)
	config.ToolCallingModel = cm
	// 对于 Claude 模型，Eino 需要我们自己实现一个 tool checker，否则无法触发 MCP 调用
	config.StreamToolCallChecker = claudeStreamToolChecker
	log.Printf("create llm success\n\n")

	// 绑定 Tools
	log.Printf("===bind tools===\n")
	tools := make([]tool.BaseTool, 0)
	tools = append(tools, newAfterShipConnectorSDKTool()...)
	tools = append(tools, newFileSystemTool()...)
	config.ToolsConfig.Tools = tools

	// 创建 Agent
	log.Printf("===create agent===\n")
	agent, newAgentErr := react.NewAgent(ctx, config)
	if newAgentErr != nil {
		panic(newAgentErr)
	}

	createGinServer(func(ctx context.Context, request *Request) (*schema.StreamReader[*schema.Message], error) {
		log.Printf("===llm stream generate===\n")
		log.Printf("request messages: %+v\n", request.Messages)
		return startAgentFlow(agent, ctx, request, mainSystemMessage)
	})
}

func startAgentFlow(cm *react.Agent, ctx context.Context, request *Request, systems []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	// 这里可以添加更多的业务逻辑
	log.Printf("Starting agent flow with chat model: %T\n", cm)

	// 处理系统消息：移除现有系统消息并插入新的系统提示
	filteredMessages := convertMessages(request)

	// 插入新的系统提示到消息开头
	allMessages := append(systems, filteredMessages...)

	// 初始化 responseStream
	responseStream, streamWriter := schema.Pipe[*schema.Message](10)

	// 将 responseStream 存储到 context 中
	ctx = context.WithValue(ctx, "responseStreamWriter", streamWriter)

	// cm.Stream 的 streamReader 会自动读取并写入到 responseStream
	go forwardStream(ctx, cm, allMessages, streamWriter)

	return responseStream, nil
}

func convertMessages(request *Request) []*schema.Message {
	filteredMessages := make([]*schema.Message, 0)
	for _, msg := range request.Messages {
		if msg.Role == anthropic.MessageParamRoleUser || msg.Role == anthropic.MessageParamRoleAssistant {
			schemaMessage := schema.Message{
				Role: schema.RoleType(msg.Role),
			}

			// Convert content blocks to string content
			// Extract text content from the message
			if len(msg.Content) > 0 {
				// Extract only text content for processing
				var textContent strings.Builder
				for _, content := range msg.Content {
					// Check if this is a text block
					if content.OfText != nil {
						textContent.WriteString(content.OfText.Text)
					}
				}

				if textContent.Len() > 0 {
					schemaMessage.Content = textContent.String()
				} else {
					// If no text content, convert to JSON as fallback
					contentBytes, err := json.Marshal(msg.Content)
					if err != nil {
						log.Printf("Error marshaling content: %v", err)
						schemaMessage.Content = "Content conversion error"
					} else {
						schemaMessage.Content = string(contentBytes)
					}
				}
			}

			filteredMessages = append(filteredMessages, &schemaMessage)
		}
	}
	return filteredMessages
}

func getSystemMessage(prompt *types.Prompt) []*schema.Message {
	messages := make([]*schema.Message, len(prompt.Prompt))
	for i, content := range prompt.Prompt {
		messages[i] = &schema.Message{
			Role:    schema.System,
			Content: content.Content,
		}
	}
	return messages
}

// forwardStream 优雅地将源流转发到目标流写入器
func forwardStream(ctx context.Context, agent *react.Agent, messages []*schema.Message, writer *schema.StreamWriter[*schema.Message]) {
	defer writer.Close()

	//正常来说，我们应该读取这里的 streamReader, 给到用户；
	//之所以我没这样做，是因为 Claude 模型在触发 tool 调用的时候，tool request 在最后一个 chunk 才返回；
	// react.go 这里没有很好地适配 claude 模型的这种特性，从而导致 claude 模型下 stream 打字机效果被破坏。
	// 正式解法，可能需要重写 react.go，这里 demo 我简单在 #claudeStreamToolChecker 那边处理消息返回,从而解决这个问题。
	_, err := agent.Stream(ctx, messages)
	if err != nil {
		log.Printf("Error creating stream: %v", err)
		writer.Send(nil, err) // 将错误传播到 responseStream
		return
	}
}
