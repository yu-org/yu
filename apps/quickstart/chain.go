package main

import (
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/startup"
	"github.com/yu-org/yu/core/tripod"
	"net/http"
)

type QuickStart struct {
	*tripod.Tripod
}

func NewQuickStart() *QuickStart {
	tri := &QuickStart{
		tripod.NewTripod(),
	}
	// 此处需要手动将自定义的 Writing 注册到 tripod 中，
	tri.SetWritings(tri.WriteA)
	// 此处需要手动将自定义的 Reading 注册到 tripod 中
	tri.SetReadings(tri.ReadA)
	return tri
}

type WriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 此处定制开发一个 Writing
// Writing会被全网节点共识并执行
func (q *QuickStart) WriteA(ctx *context.WriteContext) error {
	// 设置该 writing 所需消耗的lei (lei和gas同义)
	ctx.SetLei(100)
	// 解析请求体
	req := new(WriteRequest)
	err := ctx.BindJson(req)
	if err != nil {
		return err
	}
	// 将数据存入链上状态中。
	q.Set([]byte(req.Key), []byte(req.Value))
	// 向链外发射一个event
	ctx.EmitStringEvent("execute success")
	return nil
}

type ReadRequest struct {
	Key string `json:"key"`
}

type ReadResponse struct {
	Value string `json:"value"`
}

// 此处定制开发一个 Reading
func (q *QuickStart) ReadA(ctx *context.ReadContext) {
	req := new(ReadRequest)
	err := ctx.BindJson(req)
	if err != nil {
		ctx.Err(http.StatusBadRequest, err)
		return
	}
	value, err := q.Get([]byte(req.Key))
	if err != nil {
		ctx.ErrOk(err)
		return
	}
	ctx.JsonOk(ReadResponse{Value: string(value)})
}

func main() {
	// 启用poa tripod的默认配置
	poaCfg := poa.DefaultCfg(0)
	// 启用yu的默认配置
	yuCfg := startup.InitDefaultKernelConfig()

	poaTri := poa.NewPoa(poaCfg)
	qsTri := NewQuickStart()
	startup.DefaultStartup(yuCfg, poaTri, qsTri)
}
