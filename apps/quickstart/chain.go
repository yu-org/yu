package main

import (
	"net/http"

	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/startup"
	"github.com/yu-org/yu/core/tripod"
)

type QuickStart struct {
	*tripod.Tripod
}

func NewQuickStart() *QuickStart {
	tri := &QuickStart{
		tripod.NewTripod(),
	}
	// Register custom Writing to the tripod manually
	tri.SetWritings(tri.WriteA)
	// Register custom Reading to the tripod manually
	tri.SetReadings(tri.ReadA)
	return tri
}

type WriteRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// WriteA is a custom Writing handler.
// Writings are executed by consensus across all network nodes.
func (q *QuickStart) WriteA(ctx *context.WriteContext) error {
	// Set the lei (equivalent to gas) cost for this writing
	ctx.SetLei(100)
	// Parse the request body
	req := new(WriteRequest)
	err := ctx.BindJson(req)
	if err != nil {
		return err
	}
	// Store data into on-chain state
	q.Set([]byte(req.Key), []byte(req.Value))
	// Emit an event to off-chain listeners
	ctx.EmitStringEvent("execute success")
	return nil
}

type ReadRequest struct {
	Key string `json:"key"`
}

type ReadResponse struct {
	Value string `json:"value"`
}

// ReadA is a custom Reading handler.
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
	// Use the default PoA tripod configuration
	poaCfg := poa.DefaultCfg(0)
	// Use the default yu kernel configuration
	yuCfg := startup.InitDefaultKernelConfig()

	poaTri := poa.NewPoa(poaCfg)
	qsTri := NewQuickStart()
	startup.InitDefaultKernel(yuCfg).WithTripods(poaTri, qsTri).Startup()
}
