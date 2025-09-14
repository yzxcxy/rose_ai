package custom_tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strconv"
	"time"
)

type DeleteTodosTool struct {
	BaseURL string // API base URL，例如 "http://localhost:8080"
}

type DeleteTodoReq struct {
	TodoId int64 `json:"todoId"`
}

type DeleteTodoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewDeleteTodoTool(baseURL string) *DeleteTodosTool {
	return &DeleteTodosTool{
		BaseURL: baseURL,
	}
}

func (a *DeleteTodosTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "delete_todo",
		Desc: "删除一个todo项，输入是待删除代办项的todoId",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"todoId": {
				Type:     schema.Integer,
				Required: true,
				Desc:     "待删除代办项的todoId",
			},
		}),
	}, nil
}

func (a *DeleteTodosTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析输入 JSON
	var req DeleteTodoReq
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", fmt.Errorf("无效的输入 JSON: %v", err)
	}

	// 验证字段
	v := validator.New()
	if err := v.Struct(req); err != nil {
		return "", fmt.Errorf("验证失败: %v", err)
	}

	// 准备 HTTP 请求
	body, _ := json.MarshalIndent(req, "", " ")

	httpReq, _ := http.NewRequest(http.MethodDelete, a.BaseURL+"/todoList/"+strconv.FormatInt(req.TodoId, 10), bytes.NewReader(body))
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

	var res DeleteTodoResp
	if err := json.Unmarshal(respBody, &res); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if res.Code != 0 {
		return "", fmt.Errorf("API 返回错误: %s", res.Msg)
	}
	return "删除成功", nil
}
