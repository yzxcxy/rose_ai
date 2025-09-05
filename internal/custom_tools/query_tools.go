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
	"time"
)

type QueryTodosTool struct {
	BaseURL string // API base URL，例如 "http://localhost:8080"
}

type QueryTodosResp struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data types.ListTodoResp `json:"data"`
}

func NewQueryTodosTool(baseURL string) *QueryTodosTool {
	return &QueryTodosTool{
		BaseURL: baseURL,
	}
}

func (a *QueryTodosTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "query_todo",
		Desc: "查询数据库内的待办事项，需要根据自然语言来构建合适的参数，比如我想查询关于政治的代办事项，那么其就需要将其填入查询参数的search字段中，另外还需要补齐其他参数",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"page": {
				Type:     schema.Integer,
				Required: true,
				Desc:     "查询todo list的页码，默认可以填入1",
			},
			"pageSize": {
				Type:     schema.Integer,
				Required: true,
				Desc:     "查询页大小，默认可以填入10",
			},
			"status": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的状态，必须是 pending in_progress completed 之一",
				Enum:     []string{"pending", "in_progress", "completed"},
			},
			"priority": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的优先级，必须是 low medium high 之一",
				Enum:     []string{"low", "medium", "high"},
			},
			"minDueDate": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的最小截止日期，格式为：2006-01-02 15:04:05，支持使用自然语言描述日期，比如明天，需要将明天的时间转换为标准格式",
			},
			"maxDueDate": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项的最长截止日期，格式为：2006-01-02 15:04:05，支持使用自然语言描述日期，比如明天，需要将明天的时间转换为标准格式",
			},
			"search": {
				Type:     schema.String,
				Required: false,
				Desc:     "用于模糊搜索名称或者描述，注意该字段必须简短且能高度契合查询要求",
			},
			"startDate": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项创建的最小时间，格式为：2006-01-02 15:04:05，支持使用自然语言描述日期，比如明天，需要将明天的时间转换为标准格式",
			},
			"endDate": {
				Type:     schema.String,
				Required: false,
				Desc:     "待办事项创建的最长时间，格式为：2006-01-02 15:04:05，支持使用自然语言描述日期，比如明天，需要将明天的时间转换为标准格式",
			},
			"sortBy": {
				Type:     schema.Array,
				Required: false,
				Desc:     "用于指定查询的结果根据什么字段进行排序，其数据类型是一个字符串数组，元素值必须是指定的，注意其数组长度和sortOrder参数的长度一致",
				Enum: []string{
					"id",
					"todo_id",
					"user_id",
					"name",
					"description",
					"status",
					"priority",
					"due_date",
					"created_at",
				},
			},
			"sortOrder": {
				Type:     schema.Array,
				Required: false,
				Desc:     "用于指定查询的结果根据字段的升序还是降序排序，其数据类型是一个字符串数组，元素值必须是指定的，注意其数组长度和sortBy参数的长度一致",
				Enum:     []string{"asc", "desc"},
			},
		}),
	}, nil
}

func (a *QueryTodosTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析输入 JSON
	var req types.ListTodoReq
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("无效的输入 JSON: %v", err)
	}

	// 验证字段
	v := validator.New()
	if err := v.Struct(req); err != nil {
		return "", fmt.Errorf("验证失败: %v", err)
	}

	// 准备 HTTP 请求
	body, err := json.MarshalIndent(req, "", " ")
	body = bytes.ReplaceAll(body, []byte{'\n'}, []byte{'\r', '\n'})
	logx.Infof("请求的json： %s", string(body))
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	httpReq, _ := http.NewRequest(http.MethodPost, a.BaseURL+"/todoList", bytes.NewReader(body))
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

	var res QueryTodosResp
	if err := json.Unmarshal(respBody, &res); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if res.Code != 0 {
		return "", fmt.Errorf("API 返回错误: %s", res.Msg)
	}

	jsonData, _ := json.Marshal(res.Data)
	data := string(jsonData)
	return fmt.Sprintf("查询结果为: %s", data), nil
}
