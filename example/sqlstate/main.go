// Command sqlstate demonstrates table storage inside a business tripod.
package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/apps/synchronizer"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/startup"
	"github.com/yu-org/yu/core/state/sqlstate"
	"github.com/yu-org/yu/core/tripod"
)

// Order is a complete row; amounts use integers rather than floating point.
type Order struct {
	ID     string `gorm:"primaryKey" json:"id"`
	Owner  string `gorm:"index" json:"owner"`
	Amount int64  `json:"amount"`
	Status string `gorm:"index" json:"status"`
}

// OrderBook owns its table handle, not a separate database connection.
type OrderBook struct {
	*tripod.Tripod
	orders *sqlstate.Table
}

type orderRequest struct {
	ID string `json:"id"`
}

type findRequest struct {
	Owner string `json:"owner"`
}

// NewOrderBook registers transaction and query handlers before kernel startup.
func NewOrderBook() *OrderBook {
	o := &OrderBook{Tripod: tripod.NewTripodWithName("orderbook")}
	o.SetWritings(o.CreateOrder, o.MarkMatched, o.DeleteOrder)
	o.SetReadings(o.GetOrder, o.FindOrders)
	return o
}

// InitChain registers the orderbook_orders schema before synchronization runs.
func (o *OrderBook) InitChain() {
	table, err := o.RegisterStateTable("orders", &Order{})
	if err != nil {
		panic(err)
	}
	o.orders = table
}

// CreateOrder inserts a new pending order. ID and owner must be nonempty,
// and amount must be positive. Authorization is application-specific.
func (o *OrderBook) CreateOrder(ctx *context.WriteContext) error {
	ctx.SetLei(100)
	var row Order
	if err := ctx.BindJson(&row); err != nil {
		return err
	}
	if row.ID == "" || row.Owner == "" || row.Amount <= 0 {
		return fmt.Errorf("id, owner and a positive amount are required")
	}
	var existing Order
	err := o.orders.Get(row.ID, &existing)
	if err == nil {
		return fmt.Errorf("order %s already exists", row.ID)
	}
	if !errors.Is(err, sqlstate.ErrNotFound) {
		return err
	}
	row.Status = "pending"
	return o.orders.Put(&row)
}

// MarkMatched demonstrates read-modify-write in the current transaction view.
// The order must exist and be pending; a real exchange supplies matching rules.
func (o *OrderBook) MarkMatched(ctx *context.WriteContext) error {
	ctx.SetLei(100)
	var req orderRequest
	if err := ctx.BindJson(&req); err != nil {
		return err
	}
	var row Order
	if err := o.orders.Get(req.ID, &row); err != nil {
		return err
	}
	if row.Status != "pending" {
		return fmt.Errorf("order %s is not pending", req.ID)
	}
	row.Status = "matched"
	return o.orders.Put(&row)
}

// DeleteOrder stages a tombstone for an existing order in the current transaction.
func (o *OrderBook) DeleteOrder(ctx *context.WriteContext) error {
	ctx.SetLei(100)
	var req orderRequest
	if err := ctx.BindJson(&req); err != nil {
		return err
	}
	return o.orders.Delete(req.ID)
}

// GetOrder reads latest finalized state by default, or a retained block when
// the reading request specifies block_hash. Old finalized versions are unavailable.
func (o *OrderBook) GetOrder(ctx *context.ReadContext) {
	var req orderRequest
	if err := ctx.BindJson(&req); err != nil {
		ctx.Err(http.StatusBadRequest, err)
		return
	}
	var row Order
	var err error
	if block := ctx.GetBlockHash(); block != nil {
		err = o.orders.GetPending(block.String(), req.ID, &row)
	} else {
		err = o.orders.GetFinalized(req.ID, &row)
	}
	if err != nil {
		ctx.ErrOk(err)
		return
	}
	ctx.JsonOk(row)
}

// FindOrders queries indexed SQL columns of the latest finalized table.
// It does not include pending blocks or promise historical query support.
func (o *OrderBook) FindOrders(ctx *context.ReadContext) {
	var req findRequest
	if err := ctx.BindJson(&req); err != nil {
		ctx.Err(http.StatusBadRequest, err)
		return
	}
	if ctx.GetBlockHash() != nil {
		ctx.Err(http.StatusBadRequest, fmt.Errorf("indexed queries support latest finalized state only"))
		return
	}
	var rows []Order
	if err := o.orders.FindFinalized(map[string]any{"owner": req.Owner}, &rows); err != nil {
		ctx.ErrOk(err)
		return
	}
	ctx.JsonOk(rows)
}

// main starts a single-node example with SQLite and initializes business tables
// before the synchronizer. It is a storage example, not a complete exchange.
func main() {
	cfg := config.InitDefaultCfg()
	cfg.DataDir = "yu-sqlstate-example"
	startup.InitKernel(cfg).WithTripods(
		poa.NewPoa(poa.SingleNodeCfg()),
		NewOrderBook(),
		synchronizer.NewSynchronizer(cfg.SyncMode),
	).Startup()
}
