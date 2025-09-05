package agent

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/model"
	prompt2 "github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/custom_retriver"
	"rose/internal/config"
	"rose/internal/custom_tools"
	"rose/internal/types"
	"strconv"
)

type Agent struct {
	ChatModel        model.ToolCallingChatModel
	ReAct            *react.Agent
	CallBackForReact []agent.AgentOption
	ChatTemplate     prompt2.ChatTemplate
	Retriever        *custom_retriver.VikingDBRetriever
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
			custom_tools.NewQueryTodosTool("http://localhost:8888"),
			custom_tools.NewTimeTool(),
		},
	}

	messageModifier := func(ctx context.Context, input []*schema.Message) []*schema.Message {
		return input
	}

	ctx := context.Background()
	agent.ReAct, _ = react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: agent.ChatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  messageModifier,
	})

	agent.CallBackForReact = getOpts()

	agent.ChatTemplate = NewChatTemplate()
	agent.Retriever = custom_retriver.GetRetriever(conf)

	return agent, nil
}

// QA method uses the ReAct agent to generate a response based on the provided messages.
func (agent *Agent) QA(ctx context.Context, req *types.QaRequest) (*schema.Message, error) {
	// 根据输入查询文档
	docs, err := agent.Retriever.Retrieve(ctx, req.Input)
	if err != nil {
		return nil, err
	}
	logx.Infof("召回文档数目为：%d", len(docs))
	// 将文档转化为一个列表信息
	var resources string
	for idx, _ := range docs {
		jsonData, err := json.Marshal(docs[idx])
		if err != nil {
			return nil, err
		}
		metadata := string(jsonData)
		resources += "(" + strconv.Itoa(idx+1) + "): " + docs[idx].String() + "metadata: " + metadata + "\n"
	}

	// 历史消息的mock
	// TODO： 根据sessionID进行查询具体的历史消息，同时判断first的值
	var history = []*schema.Message{
		{
			Role:    schema.User,
			Content: "你好",
		},
		{
			Role:    schema.Assistant,
			Content: "你好，请问我有什么可以帮助到你的吗",
		},
	}

	variables := map[string]any{
		"question":          req.Input,
		"history":           history,
		"not_first":         true,
		"retrieval_results": resources,
	}

	// 格式化
	message, err := agent.ChatTemplate.Format(ctx, variables)
	if err != nil {
		logx.Error(err)
		return nil, err
	}

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
