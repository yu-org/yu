# SQL 表状态存储

[English](README.md) | 简体中文

`sqlstate` 让业务 tripod 通过表接口读写共识状态。业务只需要定义行结构和主键，
框架负责交易回滚、pending 分叉、Merkle 提交以及 finalized 数据入表。

已确认业务表、pending 改动、不可变行内容、Merkle 节点和区块版本都保存在同一个
SQL 数据库中，不需要为状态树额外部署 KV 数据库。接口接受已有的 GORM 连接；
默认使用 SQLite，也可以通过现有驱动连接 MySQL/InnoDB、PostgreSQL。
一个状态数据库只能由一个节点执行实例管理，不能让多个独立节点共同写入。

## 在 tripod 中存储订单

完整、可编译的示例见 [example/sqlstate/main.go](../../../example/sqlstate/main.go)。

首先定义一行数据的结构，并在 tripod 中保存表句柄：

```go
type Order struct {
    ID     string `gorm:"primaryKey" json:"id"`
    Owner  string `gorm:"index" json:"owner"`
    Amount int64  `json:"amount"`
    Status string `gorm:"index" json:"status"`
}

type OrderBook struct {
    *tripod.Tripod
    orders *sqlstate.Table
}

// InitChain registers the schema before block execution or synchronization.
func (o *OrderBook) InitChain() {
    table, err := o.RegisterStateTable("orders", &Order{})
    if err != nil {
        panic(err)
    }
    o.orders = table
}
```

当 tripod name 是 `orderbook` 时，SQL 表名就是 `orderbook_orders`。模型自己的
`TableName()` 方法不会覆盖这个名字。同一 tripod 可以继续注册其他业务表。

在 Writing 中新增或更新数据：

```go
// SaveOrder stages a complete order during an active Writing execution.
func (o *OrderBook) SaveOrder(row *Order) error {
    return o.orders.Put(row)
}

// MarkOrderMatched updates an existing order in the current transaction view.
func (o *OrderBook) MarkOrderMatched(id string) error {
    var row Order
    if err := o.orders.Get(id, &row); err != nil {
        return err
    }
    row.Status = "matched"
    return o.orders.Put(&row)
}
```

`Put` 按主键保存**完整行**：主键不存在时新增，存在时替换。因此业务若要求“只新增，
不覆盖”，需要像完整示例中的 `CreateOrder` 那样，先通过 `Get` 检查是否已经存在。
部分字段更新要先读取完整行，修改后再 `Put`，不能只传主键和一个待更新字段。

删除使用 `o.orders.Delete(id)`。它暂存删除标记，不会立即删除已确认业务表中的行。
查不到行时返回 `sqlstate.ErrNotFound`，可以用 `errors.Is` 判断。

普通 Writing 不需要手动调用 `Commit`、`NextTxn` 或 `Finalize`。Writing 返回错误时，
kernel 撤销本交易全部表的修改；成功交易的改动参与区块 Commit 和 StateRoot 计算。
区块最终确认后，相关行才更新到已确认业务表。

## 查询哪个状态

| 接口 | 读取范围 |
| --- | --- |
| `Get(key, &row)` | 当前执行视图；包含当前交易及该分支之前的有效修改 |
| `Scan(&rows)` | 当前完整表视图，按编码后的主键顺序返回 |
| `GetFinalized(key, &row)` | 最新 finalized 状态，按状态树读取并校验行内容 |
| `FindFinalized(equal, &rows)` | 最新 finalized SQL 业务表，可使用字段索引 |
| `GetPending(version, key, &row)` | 指定保留版本：pending 区块或最新 finalized 区块 |
| `ScanPending(version, &rows)` | 指定保留版本的完整表视图 |

`Get` 适合交易执行时读取自己的修改；对外查询可明确选择最新 finalized 或指定的
已提交 pending 区块。完整示例中的 `GetOrder` 默认读取 finalized，并支持通过
Reading 请求的 `block_hash` 选择仍保留的版本。

例如，查询某个所有者的已确认订单：

```go
var rows []Order
err := o.orders.FindFinalized(map[string]any{"owner": owner}, &rows)
```

这个查询不包含 pending 改动，也不支持查询旧 finalized 区块。需要在当前执行视图
撮合订单时，可以先 `Scan`，再按业务条件筛选；当前实现尚未为 pending 查询建立
可使用 SQL 二级索引的视图。

## 配置和启动顺序

```toml
statedb_type = "sql"

[state_sql]
sql_db_type = "sqlite"
dsn = "state.db"
```

SQLite 的相对路径按 `data_dir` 解析。MySQL 使用 `mysql`；PostgreSQL 可以使用
`postgre`、`postgres` 或 `postgresql`，并提供相应驱动 DSN。

`config.InitDefaultCfg()` 为新链选择 SQL 状态后端。显式选择旧后端的配置仍可使用
原有实现。新后端使用新的状态和封块承诺格式，**不会自动迁移已有 KV 状态数据库**。

注册模型必须早于区块执行和同步。示例显式将 synchronizer 放在业务 tripod 后面：

```go
startup.InitKernel(cfg).WithTripods(
    poa.NewPoa(poa.SingleNodeCfg()),
    NewOrderBook(),
    synchronizer.NewSynchronizer(cfg.SyncMode),
).Startup()
```

从仓库根目录运行 `go run ./example/sqlstate` 可启动单节点存储示例，数据目录为
`yu-sqlstate-example`。示例用于演示接口；业务权限和实际撮合规则由应用补充。

## 模型和表名约束

- 表名采用 `tripodName_businessTableName`。两部分只能使用小写字母、数字和下划线，
  且以字母开头；拼接后的名字最长 63 字节。冲突或有歧义的注册会被拒绝。
- 主键支持字符串、整数和复合主键。复合主键查询时，传入填好主键字段的模型。
  数值主键需要显式设置 `autoIncrement:false`。
- 字段支持字符串、布尔值、整数、字节切片及其可空指针。复杂数据可以显式编码到
  字符串或字节列。浮点数和原生时间字段不在首版支持范围内；时间可以使用整数时间戳。
- 所有整数值必须处于可移植的有符号 SQL 整数范围内。更大的整数可以用十进制字符串。
  索引字符串最长 191 字节，字符串主键不能以空格结尾。
- 不支持自动生成 ID、时间戳、默认值、关联模型、自定义 SQL 类型、GORM serializer、
  check 约束和二级唯一约束。相关业务校验需要针对当前 pending 视图完成。
- 编码只包含实际持久化字段，并按 SQL 列名排序。模型类型、标签和编码规则属于协议
  配置；不兼容的模型注册会被拒绝，结构升级需要显式迁移。

所有共识写入都必须经过表接口。直接通过 GORM 或 SQL 修改业务表，会绕过回滚和
Merkle 承诺，属于不受支持的用法。

## Merkle 提交与保留范围

所有业务表共用一棵具有 256 位路径、压缩末端路径的 Sparse Merkle Tree：

- 用无歧义编码的 `(物理表名, 编码后的主键)` 计算路径。
- 完整行记录包含表名、主键和确定性编码的字段值，并按内容 hash 存储。
- 叶子承诺路径与行内容 hash；分支承诺左右子节点的 hash。
- 行、叶子、分支、路径和空子树使用不同的域标记。
- 修改时只创建受影响路径上的新节点，复用未变化子树；保留版本引用的节点和行
  不会被覆盖。

只保留**最新 finalized 状态及其 pending 后继分支**。推进 finality 时，在同一个 SQL
事务中更新业务表和 finalized 指针，淘汰旧 finalized 版本及竞争分支，并清理不可达
节点和行。pending 后继仍需引用的共享内容继续保留。

不提供旧 finalized 状态查询。当前垃圾回收采用同步 mark-and-sweep。

## Kernel 与共识接入

`sqlstate.State` 实现 `state.IState`。旧 `Tripod.Set/Get` 调用通过 `tripod_kv` 表兼容；
新增具体业务表使用 `RegisterStateTable`。

kernel 从真实父区块的状态开始执行，管理交易帧，并在 SQL Commit 事务内部比较收到的
StateRoot。校验不通过时，节点、行、pending 日志和版本记录一起回滚。

Commit 先使用本地执行 ID，计算 StateRoot 后才能确定最终区块 hash。共识在封块后、
追加区块前必须调用 `ChainEnv.BindState(block)`，绑定状态版本，并保存带最终区块 hash
的回执。会进入 StateRoot 的业务数据不能依赖当前区块的最终 hash，否则会形成循环依赖。

PoA 已接入“先执行、再签名”的流程，接收节点会重新执行并比较 root。SQL 状态路径的
receipt root 排除本地保存时间和最终区块 hash，使用确定的顺序和带域标记的哈希。
其他共识需要显式接入相同生命周期；旧 KV 和无状态后端保留原有路径。

同步时先执行、校验和绑定，再追加区块。收到区块本身不代表最终确认，共识判断可以
finalize 后，应调用可返回错误的 `FinalizeState`。`DropPending` 丢弃落败分支及其后继；
`AbortExecution` 撤销当前失败执行。

创世状态在 tripod 初始化前锚定。自定义创世分配需要提前通过 `Store` 注册模型并
初始化数据，绑定真实创世身份及匹配的 root。创世身份单独保存，不会随旧版本被清理。

重启时，若链的 finality 标记落后于已持久化的 SQL 状态，kernel 按区块 hash、祖先
关系和 StateRoot 修复。SQL 状态内部提交原子化；区块存储和回执存储仍属于独立子系统。

## 验证范围

测试覆盖增量 root 与独立全量重建的比较、交易回滚、分叉隔离、删除、pending 扫描、
重启恢复、旧 finalized 版本淘汰、后继版本保留、垃圾回收、跨表提交失败回滚、内容损坏、
root 不匹配回滚、执行重放、回执元数据、finality 标记恢复和三节点 PoA 重放。

真实数据库测试目前覆盖 SQLite。MySQL、PostgreSQL 接入了现有驱动，但仍需要在部署
环境中进行真实服务的一致性验证。
