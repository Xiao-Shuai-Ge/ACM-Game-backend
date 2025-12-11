# 后端技术栈与目录设计规则（ACM-Game）

## 总览

- 目标：必须采用微服务架构；每个服务可多实例水平扩展；保持实现简洁清晰。
- 设计原则：模块边界清晰、依赖通过 Uber FX 管理、API 统一由 Protocol Buffers 生成并复用于 gRPC 与 HTTP（grpc-gateway）。

规则：

- 必须按业务域拆分服务，禁止跨服务直接访问他服务数据库。
- 必须支持多副本部署（无状态、优雅停机、幂等写入、请求可重试）。
- 必须以 `.proto` 为唯一 API 真源，HTTP 仅为网关映射。
- 必须统一使用 `fx` 管理依赖与生命周期；禁止手写全局单例。

## 技术栈

- 依赖注入：Uber `fx`
- HTTP 框架：`gin`
- gRPC 与网关：`grpc-go`、`grpc-gateway`、`protoc`（Protocol Buffers）
- ORM 与数据库：`gorm` + `mysql`
- 缓存与消息：`go-redis`
- 配置：`viper`（支持 YAML 文件与环境变量覆盖）
- 日志：`zap`（结构化日志，与 `fx` 集成）

## 顶层目录结构

```
.
├── cmd/
│   └── server/                 # 主服务入口（单进程：HTTP + gRPC + Gateway）
│       └── main.go
├── api/
│   ├── proto/                  # .proto 源文件（分包按业务域与版本）
│   └── gen/go/                 # 由 protoc 生成的 Go 代码（models、grpc、gateway）
├── internal/
│   ├── app/                    # 应用层装配（fx Module 聚合与生命周期）
│   ├── config/                 # 配置解析（viper），统一配置结构体
│   ├── log/                    # 日志初始化（zap），统一 Logger 接口
│   ├── server/
│   │   ├── http/               # Gin 路由、HTTP 中间件、Gateway 注册
│   │   └── grpc/               # gRPC Server、拦截器
│   ├── db/                     # MySQL（gorm）初始化与迁移
│   ├── cache/                  # Redis 客户端与封装
│   ├── model/                  # 领域模型与 GORM 模型定义
│   ├── repository/             # 仓储层（面向 model 的持久化）
│   ├── service/                # 领域服务（业务逻辑）
│   └── handler/
│       ├── http/               # HTTP 适配层（参数校验、DTO、调用 service）
│       └── grpc/               # gRPC 适配层（从 proto 生成的接口实现）
├── configs/
│   ├── config.yaml             # 默认配置
│   ├── config.local.yaml       # 本地开发覆盖
│   └── config.test.yaml        # 测试环境覆盖
├── scripts/                    # 代码生成、运行与工具脚本（可选）
└── Makefile                    # 常用任务（代码生成、lint、run）（可选）
```

说明与对比：

- 相比参考项目的 `internal` 过度聚合，这里按职责分层（server/db/cache/config/log/model/repository/service/handler），降低耦合，提高可测试性。
- `api/proto` 与 `api/gen/go` 明确分离源与生成物，便于清理与再生成。
- 初期采用单二进制 `cmd/server`；若未来需要拆分，可按业务域新增 `cmd/<service>` 并复用 `internal` 的模块。

### 微服务布局

当业务与团队规模增长时，允许按域拆分为多服务。推荐目录如下（保持与单体设计的分层一致）：

```
.
├── cmd/
│   ├── gateway/                 # 统一 HTTP 入口（聚合各服务的 grpc-gateway，可选）
│   ├── user/                    # 用户服务（二进制）
│   ├── platform/                # 资产/素材/任务等服务
│   └── realtime/                # 实时消息/兑换等服务
├── internal/
│   ├── common/                  # 共享模块（可被本仓库内任意包引用）：config/log/clients/tracing
│   ├── user/                    # user 服务的私有实现（server/db/cache/model/repo/service/handler）
│   ├── platform/                # platform 服务的私有实现
│   └── realtime/                # realtime 服务的私有实现
├── api/
│   ├── proto/common/v1/         # 通用模型与错误结构
│   ├── proto/user/v1/           # user 服务的 proto
│   ├── proto/platform/v1/       # platform 服务的 proto
│   └── proto/realtime/v1/       # realtime 服务的 proto
└── api/gen/go/                  # 生成代码（与 proto 目录结构对齐）
```

规则：

- 每个服务拥有独立的入口与生命周期；共享能力（配置、日志、客户端封装）置于 `internal/common`，避免循环依赖。
- 网关可以是独立二进制 `cmd/gateway`，也可由某个服务承载；统一 HTTP 域名与鉴权策略。
- 跨服务调用使用 gRPC，避免直接访问他服务的数据库；公共消息放入 `api/proto/common`。

命名与端口：

- 二进制命名使用短名小写：`user`, `platform`, `realtime`, `gateway`。
- 端口约定：gRPC 使用 `91xx/92xx/93xx`，HTTP 网关使用 `80xx/81xx`（可在 `configs` 中覆盖）。

## 依赖注入（Uber FX）

- 每个子系统提供 `fx.Option`：例如 `internal/config.Module`、`internal/log.Module`、`internal/db.Module`、`internal/cache.Module`、`internal/server/http.Module`、`internal/server/grpc.Module`。
- 在 `internal/app` 聚合模块：`fx.New(app.Module, config.Module, log.Module, db.Module, cache.Module, http.Module, grpc.Module)`。
- 使用 `fx.Lifecycle` 管理启动与关闭（如 DB/Redis 连接、HTTP/gRPC 监听、网关注册）。

规则：

- 每个模块必须暴露 `fx.Option`（如 `Module`）；入口仅通过组合 `fx.Option` 装配。
- 必须在 `OnStart/OnStop` 中处理资源建立与释放，支持优雅停机（超时、拒绝新请求、等待在途完成）。
- 禁止在包初始化阶段建立网络连接或启动 goroutine。

## 配置（Viper + YAML + 环境变量）

- 约定 `configs/config.yaml` 为默认；根据 `ENV` 加载 `config.<env>.yaml` 追加覆盖。
- 环境变量前缀：`ACMGAME_`，如 `ACMGAME_SERVER_HTTP_ADDR`；优先级：环境变量 > 环境 YAML > 默认 YAML。
- 建议统一配置结构体：

```yaml
server:
  http:
    addr: ":8080"
  grpc:
    addr: ":9090"
services:                # 微服务模式下其他服务的地址（供客户端调用）
  user:
    grpcAddr: "localhost:9100"
  platform:
    grpcAddr: "localhost:9200"
  realtime:
    grpcAddr: "localhost:9300"
mysql:
  dsn: "user:pass@tcp(localhost:3306)/acmgame?parseTime=true&loc=Local&charset=utf8mb4"
redis:
  addr: "localhost:6379"
  db: 0
  password: ""
log:
  level: "info"        # debug | info | warn | error
  format: "json"       # json | console
  sampling: true
```

规则：

- 必须提供默认配置与本地覆盖文件；生产以环境变量覆盖敏感与地址。
- 必须为每个外部依赖设置超时与重试参数（客户端与连接池）。
- 禁止在代码中硬编码地址、凭据与端口。

## 日志（Zap）

- 在 `internal/log` 初始化 `zap.Logger`，通过 `fx` 提供；全局禁止使用标准库 `log`。
- HTTP：Gin 使用自定义中间件记录请求（方法、路径、耗时、状态码、trace-id）。
- gRPC：客户端与服务端拦截器统一记录调用与错误。
- 建议开启采样与字段白名单，避免高并发下日志爆炸。

规则：

- 必须在入站/出站请求打关键事件日志（开始、错误、结束、耗时）。
- 禁止打印敏感数据（token、密码、密钥、个人信息）。

## HTTP 与 gRPC

- gRPC 为真源（定义在 `.proto`）；HTTP 通过 `grpc-gateway` 自动映射，确保两端行为一致。
- 进程内同时启动：
  - `internal/server/grpc`：监听 `server.grpc.addr`，注册服务实现（`internal/handler/grpc`）。
  - `internal/server/http`：启动 Gin；注册 Gateway 的 `ServeMux` 到某个路由前缀（如 `/api/v1`）。
- 鉴权与中间件：优先在 gRPC 层实现（便于 gateway 复用），HTTP 额外处理跨域、限流等。

规则：

- 必须以 gRPC 为服务间通信主通道；HTTP 仅用于外部入口。
- 必须在 `.proto` 定义中使用 `google.api.http` 注解来声明网关路由。
- 必须为 RPC 调用设置明确的 `deadline/timeout` 与重试策略；写操作需幂等。

### 服务间通信与网关

- 同步调用：服务间统一使用 gRPC，明确超时、重试与幂等；客户端由 `internal/common/clients` 统一提供（封装负载均衡与拦截器）。
- HTTP 聚合：`cmd/gateway` 挂载各服务的 grpc-gateway，提供统一入口与鉴权；支持按路由前缀划分域（如 `/user`, `/platform`, `/rt`）。
- API 设计：以 `.proto` 为唯一真源；跨服务共享模型置于 `api/proto/common/v1`，避免重复定义。

规则：

- 必须透传 `trace-id/req-id`；在日志与错误中保持关联。
- 网关必须统一鉴权入口与限流策略；禁止在多个服务重复实现外部鉴权。

## Protocol Buffers 约定

- 目录：`api/proto/<domain>/v1/<service>.proto`；包名示例：`acmgame.platform.v1`。
- 生成：输出至 `api/gen/go`，与源相对路径一致；示例命令：

```

protoc -I api/proto \
  --go_out api/gen/go --go_opt paths=source_relative \
  --go-grpc_out api/gen/go --go-grpc_opt paths=source_relative \
  --grpc-gateway_out api/gen/go --grpc-gateway_opt paths=source_relative \
  api/proto/**/*.proto
```

- HTTP 注解：在 `.proto` 中使用 `google.api.http` 注解生成网关路由。
- 版本演进：新增字段采用可选，避免破坏；跨大版本新增 `v2` 包并并行提供。

规则：

- `.proto` 必须按域与版本组织：`api/proto/<domain>/v1/...`。
- 必须开启 `paths=source_relative`，并输出到 `api/gen/go`，禁止提交到其他目录。
- 必须在 CI 中校验生成代码与源一致（代码生成不应手改）。

## 数据访问（GORM + MySQL）

- 初始化：`internal/db` 根据配置创建 `gorm.DB`，设置连接池与命名策略（表名下划线复数）。
- 模型：放在 `internal/model`；迁移在 `internal/db` 提供统一入口（可选集成 `gorm` 的 `AutoMigrate`）。
- 仓储：`internal/repository` 仅做持久化与查询构造；不承载业务规则。
- 事务：在 `service` 层显式控制（传递 `*gorm.DB` 上下文）。

### 数据边界与跨服务一致性

- 每个服务拥有自己的数据库（逻辑或物理隔离）；禁止跨服务直接访问他服务的表。
- 跨服务一致性通过业务层补偿（Saga/Outbox）或异步事件实现；避免分布式事务作为默认方案。

规则：

- 每服务必须拥有独立数据库（逻辑/物理）；禁止跨库 join。
- 写操作必须设计幂等键（如业务主键/去重 token），以支持重试与多副本。
- 必须实现迁移流程（启动或脚本），禁止手工改表。

## Redis（go-redis）

- 初始化：`internal/cache` 提供 `redis.Client`，统一 `Get/Set`/`TTL` 等封装。
- 使用场景：
  - 会话与临时令牌（TTL）
  - 读多写少的短期缓存（热点数据）
  - 简单队列或轻量发布订阅（后期可替换专业 MQ）

规则：

- 必须为缓存键采用命名空间前缀：`acm:<service>:<purpose>:<id>`。
- TTL 必须明确且可配置；禁止永久键除非有业务理由。
- 禁止将 Redis 作为强一致事务存储；仅用于缓存与轻量协调。

## 适配层（Handler）与服务层（Service）

- `internal/handler/grpc`：实现由 proto 生成的接口，将请求转为领域调用，做最少的参数校验与错误映射。
- `internal/handler/http`：处理非 gRPC 的 HTTP 路由（如健康检查、静态资源、管理端），其余通过 gateway。
- `internal/service`：承载业务规则与跨仓储操作；避免直接依赖框架类型（便于单测）。

规则：

- Handler 仅做协议适配与参数校验；业务逻辑必须在 `service` 层。
- Service 层必须通过接口依赖 `repository`；便于单测与替换。
- 错误必须统一映射为标准错误模型（gRPC status/HTTP JSON）。

### 客户端封装（微服务）

- 在 `internal/common/clients` 提供各服务的 gRPC 客户端构造函数，注入到需要调用的服务中（通过 `fx`）。
- 统一拦截器：超时、重试、日志、trace-id 透传；支持基于地址列表的负载均衡。

## 启动与生命周期

- 单体：`cmd/server/main.go` 创建 `fx.App`，组装模块，`fx.Lifecycle` 启动 HTTP 与 gRPC；捕获信号优雅关闭。
- 健康检查：HTTP `/healthz` 与 gRPC `Health` 服务；依赖 DB/Redis 的就绪探针 `/readyz`。

微服务：每个二进制均创建自己的 `fx.App`，注册其私有模块与共享模块；统一使用健康与就绪探针。

## 同步修改

当项目的架构或规范发生变化时，必须同步更新此文档，并在评审中确认遵循情况。

## 多实例与无状态约束

规则：

- 服务必须无状态，所有状态外置（DB/Redis/对象存储）。
- 必须实现优雅停机（拒绝新连接、等待在途、超时强制退出）。
- 必须保证写操作可重试且幂等；读取支持缓存失效与重建。
- 必须提供就绪探针依赖检查（DB/Redis/下游服务）；未就绪时拒绝对外提供服务。
