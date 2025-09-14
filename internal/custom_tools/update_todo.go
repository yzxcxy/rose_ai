package custom_tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"rose/internal/types"
	"strconv"
	"time"
)

// 使用直接实现接口的方式来定义工具

type UpdateTodoTool struct {
	BaseURL string // API base URL，例如 "http://localhost:8080"
}

type UpdateTodoResp struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data types.UpdateTodoResp `json:"data"`
}

type UpdateTodoReq struct {
	TodoId      int64  `json:"todoId"`
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description,optional" validate:"omitempty,max=500"`
	Status      string `json:"status,optional" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    string `json:"priority,optional" validate:"omitempty,oneof=low medium high"`
	DueDate     string `json:"dueDate,optional"`
}

type SendUpdateTodoReq struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description,optional" validate:"omitempty,max=500"`
	Status      string `json:"status,optional" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    string `json:"priority,optional" validate:"omitempty,oneof=low medium high"`
	DueDate     string `json:"dueDate,optional"`
}

func NewUpdateTodoTool(baseURL string) *UpdateTodoTool {
	return &UpdateTodoTool{
		BaseURL: baseURL,
	}
}

func (a *UpdateTodoTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_todo",
		Desc: "按照todoId去更新一个待办项",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"todoId": {
				Type:     schema.String,
				Required: true,
				Desc:     "待更新待办项的todoId",
			},
			"name": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的名称，长度在3到100个字符之间",
			},
			"description": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的描述，可以省略,最大长度500个字符",
			},
			"status": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的状态，必须是 pending in_progress completed 之一，默认是 pending",
				Enum:     []string{"pending", "in_progress", "completed"},
			},
			"priority": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的优先级，必须是 low medium high 之一，默认是 low",
				Enum:     []string{"low", "medium", "high"},
			},
			"dueDate": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的截止日期，格式为：2006-01-02 15:04:05，支持使用自然语言描述日期，比如明天，需要将明天的时间转换为标准格式",
			},
		}),
	}, nil
}

func (a *UpdateTodoTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析输入 JSON
	var req UpdateTodoReq
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("无效的输入 JSON: %v", err)
	}

	// 验证字段
	v := validator.New()
	if err := v.Struct(req); err != nil {
		return "", fmt.Errorf("验证失败: %v", err)
	}

	var send SendUpdateTodoReq = SendUpdateTodoReq{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	// 准备 HTTP 请求
	body, err := json.MarshalIndent(send, "", " ")
	body = bytes.ReplaceAll(body, []byte{'\n'}, []byte{'\r', '\n'})
	logx.Infof("请求的json： %s", string(body))
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	httpReq, _ := http.NewRequest(http.MethodPut, a.BaseURL+"/todos/"+strconv.FormatInt(req.TodoId, 10), bytes.NewReader(body))
	token, _ := ctx.Value("Authorization").(string)        // 从上下文中获取 Authorization
	httpReq.Header.Set("Authorization", token)             // 透传过去
	httpReq.Header.Set("Content-Type", "application/json") // 一定要设置 Content-Type，否则请求会失败
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("API 调用失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误状态: %d", resp.StatusCode)
	}

	// 读取并解析响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var res UpdateTodoResp
	if err := json.Unmarshal(respBody, &res); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if res.Code != 0 {
		return "", fmt.Errorf("API 返回错误: %s", res.Msg)
	}

	return "更新成功", nil
}
