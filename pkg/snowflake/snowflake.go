package snowflake

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"time"
)

var node *snowflake.Node

func Init(startTime string, machineID int64) (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 转化为毫秒
	// 设置起始的时间点
	snowflake.Epoch = st.UnixNano() / 1000000
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		fmt.Printf("NewNode error: %v\n", err)
		return
	}
	return
}

func GenID() int64 {
	return node.Generate().Int64()
}
