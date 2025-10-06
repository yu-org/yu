# ETH模块

本模块基于go-ethereum实现了以太坊兼容的区块链功能，包括EVM执行、智能合约支持、RPC接口等。

## 依赖版本

- **go-ethereum**: `v1.16.3`
- **Go版本**: `1.24.0`

## 目录结构

```
apps/eth/
├── config/          # 配置文件
├── ethrpc/          # ETH RPC接口实现
├── evm/             # EVM虚拟机实现
├── metrics/         # 监控指标
├── startup.go       # 启动入口
├── test/            # 测试代码
│   ├── cmd/         # 测试命令行工具
│   ├── contracts/   # 智能合约
│   ├── erc20/       # ERC20测试
│   ├── transfer/    # 转账测试
│   └── uniswap/     # Uniswap测试
└── types/           # 类型定义
```

## 测试指南

### 环境要求

1. **Go环境**: Go 1.24.0 或更高版本
2. **依赖**: 确保已运行 `go mod download` 安装所有依赖

### 启动测试

所有测试都需要从 `apps/eth/test/cmd` 目录运行，因为测试需要访问配置文件。

#### 1. ERC20 测试

测试ERC20代币合约的功能：

```bash
cd apps/eth/test/cmd
go run erc20/main.go
```

**功能**：
- 测试ERC20代币合约部署
- 测试代币转账功能
- 验证代币余额

#### 2. ETH转账测试

测试以太坊原生转账功能：

```bash
cd apps/eth/test/cmd
go run transfer/main.go
```

**功能**：
- 测试ETH转账交易
- 验证账户余额变化
- 测试交易冲突处理

#### 3. Uniswap测试

测试Uniswap V2协议功能：

```bash
cd apps/eth/test/cmd
go run uniswap/main.go
```

**功能**：
- 测试Uniswap V2工厂合约
- 测试代币对创建
- 测试流动性添加和移除
- 测试交易功能

#### 4. 基准测试

运行性能基准测试：

```bash
# 转账基准测试
cd apps/eth/test/cmd
go run benchmark/main.go

# Uniswap基准测试
cd apps/eth/test/cmd
go run uniswap_benchmark/main.go
```

### 配置说明

测试使用以下配置文件：

- `conf/eth.toml` - ETH模块配置
- `conf/yu.toml` - YU链配置
- `conf/poa.toml` - PoA共识配置

### 测试参数

所有测试都支持以下命令行参数：

```bash
# 基本参数
-evmConfigPath string     # ETH配置文件路径 (默认: "./conf/eth.toml")
-yuConfigPath string      # YU配置文件路径 (默认: "./conf/yu.toml")
-poaConfigPath string     # PoA配置文件路径 (默认: "./conf/poa.toml")
-nodeUrl string          # 节点RPC地址 (默认: "http://localhost:9092")
-key string              # 私钥 (默认: "32e3b56c9f2763d2332e6e4188e4755815ac96441e899de121969845e343c2ff")
-parallel bool           # 是否并行执行 (默认: true)

# 仅transfer测试支持
-as-client bool          # 作为客户端模式运行 (默认: false)
-chainId int64           # 链ID (默认: 50341)
```

### 测试示例

```bash
# 使用自定义配置运行ERC20测试
cd apps/eth/test/cmd
go run erc20/main.go -nodeUrl "http://localhost:9092" -parallel true

# 作为客户端运行转账测试
cd apps/eth/test/cmd
go run transfer/main.go -as-client true -chainId 50341

# 使用自定义私钥运行Uniswap测试
cd apps/eth/test/cmd
go run uniswap/main.go -key "your_private_key_here"
```

### 故障排除

#### 1. 数据库错误

如果遇到 `missing trie node` 错误：

```bash
# 清理数据库文件
cd apps/eth/test/cmd
rm -rf yu/yu.db
rm -rf ethstate/
rm -f yu/chain.db
```

#### 2. 端口冲突

如果遇到端口占用问题，检查配置文件中的端口设置：

- ETH RPC端口: `conf/eth.toml` 中的 `eth_port`
- YU HTTP端口: `conf/yu.toml` 中的 `http_port`
- YU WebSocket端口: `conf/yu.toml` 中的 `ws_port`

#### 3. 依赖问题

如果遇到依赖问题：

```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download
```

---

**注意**: 所有测试都会启动完整的ETH链，请确保系统资源充足。测试完成后会自动清理资源。
