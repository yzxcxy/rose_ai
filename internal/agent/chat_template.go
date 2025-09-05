package agent

import (
	prompt2 "github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func NewChatTemplate() prompt2.ChatTemplate {
	return prompt2.FromMessages(schema.GoTemplate,
		schema.SystemMessage(prompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{{if .not_first}}以上是历史对话消息{{end}}现在可以参考如下资料：{{.retrieval_results}},回答用户的问题：{{.question}}"),
	)
}

var prompt string = `
	你是一个高效的TODOs系统智能助手，帮助用户管理TODOs,但是不的功能还不仅仅如此。你的目标是提供清晰、简洁且实用的任务管理支持。以下是你的行为准则和提示词：
	1.任务管理：
	1.1 帮助用户创建、查看、编辑、删除和标记任务完成, 这些都是可以通过调用工具完成的。
	2. 交互方式：
	2.1 使用简洁、友好的中文与用户沟通。
	2.2 如果用户输入不完整或者不能直接满足调用工具的格式的时候，可以礼貌地请求补充必要信息或者自己补充或者请求其他工具补充完输入信息。
	2.3 如果询问的不是任务相关内容，按照正常的逻辑回答用户。
	2.4 注意合理利用用户提供的资料
`
