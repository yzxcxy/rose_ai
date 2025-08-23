package custom_tools

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"time"
)

// TimeTool 一个获得当前时间的工具
type TimeTool struct{}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

type TimeRequest struct {
	TimeInterval int `json:"timeInterval" validate:"required"`
}

func (t *TimeTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",
		Desc: "获取距离当前时间多少天的时间,0表示当前时间，负数表示多少天前，正数表示多少天后，返回格式为：2006-01-02 15:04:05",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"timeInterval": {
				Type:     schema.Integer,
				Desc:     "获取距离当前时间多少天前或多少天后的时间，正数表示多少天后，负数表示多少天前，默认为0表示当前时间",
				Required: true,
			},
		}),
	}, nil
}

func (t *TimeTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	timeRequest := &TimeRequest{}
	err := json.Unmarshal([]byte(argumentsInJSON), timeRequest)
	if err != nil {
		return "获取时间失败，输入了错误的参数，不能转为JSON type", err
	}

	currentTime := time.Now().AddDate(0, 0, timeRequest.TimeInterval).Format("2006-01-02 15:04:05")
	return currentTime, nil
}
