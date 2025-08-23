package agent

import (
	"context"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"rose/internal/config"
	"rose/internal/custom_tools"
)

type Agent struct {
	ChatModel        model.ToolCallingChatModel
	ReAct            *react.Agent
	CallBackForReact []agent.AgentOption
}

// NewAgent creates a new Agent instance with the provided configuration.
// 基本结构: https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/react_agent_manual/
func NewAgent(conf *config.Config) (agent *Agent, err error) {
	agent = new(Agent)
	agent.ChatModel, err = NewDeepSeekModel(conf)
	if err != nil {
		return nil, err
	}

	toolsConfig := compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			custom_tools.NewAddTodoTool("http://localhost:8888"),
			custom_tools.NewTimeTool(),
		},
	}

	messageModifier := func(ctx context.Context, input []*schema.Message) []*schema.Message {
		res := make([]*schema.Message, 0, len(input)+1)
		res = append(res, schema.SystemMessage(prompt))
		res = append(res, input...)
		return res
	}

	ctx := context.Background()
	agent.ReAct, _ = react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: agent.ChatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  messageModifier,
	})

	agent.CallBackForReact = getOpts()

	return agent, nil
}

// QA method uses the ReAct agent to generate a response based on the provided messages.
func (agent *Agent) QA(ctx context.Context, message []*schema.Message) (*schema.Message, error) {
	generate, err := agent.ReAct.Generate(ctx, message, agent.CallBackForReact...)
	if err != nil {
		return nil, err
	}

	return generate, nil
}

// getOpts returns a slice of agent options
func getOpts() []agent.AgentOption {
	opt := []agent.AgentOption{
		agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})),
	}

	return opt
}

var prompt string = `
	你是一个高效的TODOs系统智能助手，帮助用户管理TODOs,但是不的功能还不仅仅如此。你的目标是提供清晰、简洁且实用的任务管理支持。以下是你的行为准则和提示词：
	1.任务管理：
	1.1 帮助用户创建、查看、编辑、删除和标记任务完成。
	2. 交互方式：
	2.1 使用简洁、友好的中文与用户沟通。
	2.2 主动询问用户是否需要添加新任务、查看任务列表或更新任务状态。
	2.3 如果用户输入不完整或者不能直接满足调用工具的格式的时候，可以礼貌地请求补充必要信息或者自己补充或者请求其他工具补充完输入信息。
	2.4 如果询问的不是任务相关内容，按照正常的逻辑回答用户。
`
