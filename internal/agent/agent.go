package agent

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cloudwego/eino/components/model"
	prompt2 "github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"rose/internal/agent/custom_retriver"
	"rose/internal/config"
	"rose/internal/custom_tools"
	"rose/internal/types"
	"rose/internal/utils"
	"strconv"
)

type Agent struct {
	ChatModel        model.ToolCallingChatModel
	ReAct            *react.Agent
	CallBackForReact []agent.AgentOption
	ChatTemplate     prompt2.ChatTemplate
	Retriever        *custom_retriver.MilvusRetriever
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
	retriever, err := custom_retriver.NewMilvusRetriever(conf)
	if err != nil {
		return nil, err
	}
	agent.Retriever = retriever

	return agent, nil
}

// QA method uses the ReAct agent to generate a response based on the provided messages.
func (agent *Agent) QA(ctx context.Context, req *types.QaRequest) (*schema.Message, error) {
	var history []*schema.Message
	// 获得历史聊天记录（从redis中获取）
	var redisClient = ctx.Value("redis").(*redis.Client)
	uid, _, _ := utils.GetUserIdAndUserNameFromContext(ctx)
	uidStr := strconv.FormatInt(uid, 10)
	// 先判断有没有历史聊天消息
	historyKey := uidStr + "::" + req.SessionID
	// LLen函数中key不存在会返回长度为0
	length, err := redisClient.LLen(ctx, historyKey).Result()
	var summaryIndex int
	if err == nil && length > 0 {
		// 如果有历史聊天消息的时候才去获取聊天记录
		var start = 0
		summaryIndexKey := uidStr + "::" + "summary_index::" + req.SessionID
		res, err := redisClient.Get(ctx, summaryIndexKey).Result()
		if err != nil {
			if !errors.Is(redis.Nil, err) {
				return nil, err
			}
		}
		summaryIndex, _ = strconv.Atoi(res)
		start = summaryIndex + 1
		// 获得summary数据
		if summaryIndex > -1 {
			summaryKey := uidStr + "::summary::" + req.SessionID
			summary, err := redisClient.Get(ctx, summaryKey).Result()
			if err != nil {
				return nil, err
			}
			var summaryMessage *schema.Message
			err = json.Unmarshal([]byte(summary), &summaryMessage)
			if err != nil {
				return nil, err
			}
			history = append(history, summaryMessage)
		}
		// 获得未summary的数据
		vals, err := redisClient.LRange(ctx, historyKey, int64(start), -1).Result()
		if err != nil {
			return nil, err
		}
		for idx, _ := range vals {
			var val *schema.Message
			err = json.Unmarshal([]byte(vals[idx]), &val)
			if err != nil {
				return nil, err
			}
			history = append(history, val)
		}
	}
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

	// 将聊天记录添加到redis中
	// 为了规则redis，只将他们的对话保存，不存储ChatTemplate生成的消息
	userMessage := schema.Message{
		Role:    schema.User,
		Content: req.Input,
	}
	messageJson, err := json.Marshal(userMessage)
	redisClient.RPush(ctx, historyKey, string(messageJson))

	jsonData, err := json.Marshal(generate)
	jsonString := string(jsonData)
	redisClient.RPush(ctx, historyKey, jsonString)

	// 提交总结
	go agent.summaryMessages(context.Background(), redisClient, uidStr, req.SessionID, int(length)+1, summaryIndex)
	return generate, nil
}

// getOpts returns a slice of agent options
func getOpts() []agent.AgentOption {
	opt := []agent.AgentOption{
		agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})),
	}

	return opt
}

// 通过协程进行调用，判断是否对消息进行压缩
func (agent *Agent) summaryMessages(ctx context.Context, rdb *redis.Client, user string, sessionId string, length int, summaryIndex int) {
	// 原则：保留前三条且字符数大于1K才进行总结
	// TODO: 后续可以将总结的缓存消息字数也保存下来，假如策略选择当中
	currIndex := length - 1
	if currIndex < 3 || currIndex-summaryIndex+1 <= 3 {
		return
	}

	summaryEnd := currIndex - 3
	historyKey := user + "::" + sessionId
	vals, err := rdb.LRange(ctx, historyKey, int64(summaryIndex+1), int64(summaryEnd)).Result()
	if err != nil {
		logx.Error(err)
		return
	}

	var count int
	for idx, _ := range vals {
		count += len(vals[idx])
	}

	// 启动总结
	if count > 1024 {
		// 取新的数据
		vals, err = rdb.LRange(ctx, historyKey, 0, int64(summaryEnd)).Result()
		if err != nil {
			logx.Error(err)
			return
		}
		summaryIndexKey := user + "::" + "summary_index::" + sessionId
		summaryKey := user + "::summary::" + sessionId
		var message []*schema.Message
		// 添加需要总结的提示词
		message = append(message, &schema.Message{
			Role:    schema.User,
			Content: "总结下面的消息",
		})
		for idx, _ := range vals {
			var val schema.Message
			err := json.Unmarshal([]byte(vals[idx]), &val)
			if err != nil {
				logx.Error(err)
				return
			}
			message = append(message, &val)
		}

		generate, err := agent.ChatModel.Generate(ctx, message)
		if err != nil {
			logx.Error(err)
			return
		}

		jsonData, err := json.Marshal(generate)
		if err != nil {
			logx.Error(err)
			return
		}

		rdb.Set(ctx, summaryKey, string(jsonData), 0)
		rdb.Set(ctx, summaryIndexKey, summaryEnd, 0)
	}
}
