package main

import (
	"fmt"
	"time"

	"github.com/kaimixu/qps"
)

// 字段名须可导出且添加qps Tag
// 成员类型仅支持:uint64,uint32,uint
type HttpQps struct {
	In  uint64 `qps:"in"`
	Out uint64 `qps:"out"`
}

func main() {
	// 持久化缓存近3天的数据，每秒汇总一次qps
	httpQps := qps.New(time.Second, 3*24*60*60, true, HttpQps{})

	// In qps计数加1, 这里"in"为Tag名称
	httpQps.Inc("in")

	// Out Qps计数加1
	httpQps.Inc("out")

	// In qps计数加2
	httpQps.Add("in", 2)

	time.Sleep(2 * time.Second)
	// 获取第1条qps数据
	nodes := httpQps.History(10)

	//format: 2017-05-10 22:10:03.306372453 +0800 CST 3 1
	for _, n := range nodes {
		fmt.Println(n.GetCt(), n.GetData().(HttpQps).In, n.GetData().(HttpQps).Out)
	}
}
