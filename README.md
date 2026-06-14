# Go Examples

Go 语言示例代码合集，按主题分类组织，作为 [GoLang 知识库 / 博客系列](/Users/admin/blogs/GoLang) 的配套代码仓库。

每个示例对应博客中讲解的核心概念，提供可编译、可运行的最小代码片段，帮助你从"理解原理"到"动手实践"。

---

## 目录结构

```
go-examples/
├── 01-basics/          # Go 基础语法
│   ├── variables/          — 变量声明、常量、零值
│   ├── control-flow/       — if/for/switch/defer
│   ├── functions/          — 函数、多返回值、闭包、defer
│   ├── types-interfaces/   — 类型系统、接口、类型断言、类型嵌入
│   ├── errors/             — 错误处理、error 接口、panic/recover
│   └── generics/           — 泛型函数与类型约束
│
├── 02-datastructures/  # 数据结构
│   ├── slice-string/       — 切片底层结构、扩容、字符串与 []byte 互转
│   ├── map/                — 哈希表、桶、扩容策略、并发安全
│   └── struct/             — 结构体内存对齐、tag、嵌入
│
├── 03-concurrency/     # 并发编程
│   ├── goroutine-channel/  — goroutine 创建、channel 通信、缓冲 vs 非缓冲
│   ├── sync-package/       — WaitGroup、Mutex、RWMutex、Once、Pool
│   ├── atomic/             — 原子操作、atomic.Value
│   ├── context/            — 请求上下文、取消传播、超时控制
│   ├── select-timer/       — select 多路复用、超时、Ticker/Timer
│   └── patterns/           — 并发模式：Fan-in/Fan-out、Pipeline、Worker Pool
│
├── 04-runtime/         # 运行时原理
│   ├── scheduler/          — GMP 模型、goroutine 调度、抢占
│   ├── gc/                 — GC 触发条件、三色标记、STW 分析
│   ├── memory/             — 内存分配、span、mcache/mcentral/mheap
│   └── panic-recover/      — panic 传播链、recover 恢复、栈保护
│
├── 05-performance/     # 性能调优
│   ├── pprof/              — CPU/Heap/Goroutine/Mutex profiling
│   ├── benchmark/          — 基准测试、内存分配统计、编译优化
│   ├── trace/              — 执行追踪、Goroutine 分析、调度延迟
│   └── escape-analysis/    — 逃逸分析、堆栈分配决策
│
├── 06-engineering/     # 工程实践
│   ├── testing/            — 单元测试、table-driven test、mock、fuzz
│   ├── logging/            — 结构化日志、zap/logrus/slog
│   ├── config/             — 配置管理、viper、环境变量
│   ├── di/                 — 依赖注入、wire、fx
│   ├── middleware/         — 中间件模式、链式处理
│   └── ci-cd/              — CI/CD 编排、Docker 多阶段构建
│
├── 07-web/             # Web 框架与中间件
│   ├── gin/                — Gin 路由、参数绑定、中间件、验证
│   ├── gorm/               — GORM CRUD、关联、事务、钩子
│   ├── grpc/               — Protocol Buffers、服务端/客户端、拦截器
│   ├── jwt/                — JWT 签发与验证、RBAC 权限控制
│   ├── redis/              — Redis 连接池、Pub/Sub、分布式缓存
│   └── kafka/              — sarama 生产者/消费者、重试、Exactly-Once
│
├── 08-advanced/        # 高级特性
│   ├── reflection/         — reflect.Type/Value、反射调用、性能代价
│   ├── unsafe/             — unsafe.Pointer、uintptr、内存布局
│   ├── pgo/                — Profile Guided Optimization 生产优化
│   └── plugin/             — plugin 包、动态加载、热替换
│
└── 09-projects/        # 实战项目迷你版
    ├── cli-tool/           — CLI 工具：cobra 命令行解析与交互
    ├── http-server/        — HTTP 服务器：路由、中间件、REST API
    ├── cache/              — 并发安全缓存：LRU + TTL + 持久化
    ├── ratelimit/          — 限流器：令牌桶、滑动窗口、自适应限流
    └── distlock/           — 分布式锁：Redis/etcd 实现、租约与续期
```

## 如何使用

### 前置条件

- Go 1.22+（部分示例依赖泛型、PGO 等特性）
- 部分项目需要 Docker（Redis、Kafka 等中间件）

### 运行单个示例

每个子目录都是独立的 `main` 包，直接运行：

```bash
# 运行基础语法示例
go run 01-basics/variables/main.go

# 运行并发编程示例
go run 03-concurrency/goroutine-channel/main.go
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./05-performance/benchmark/... -bench=.

# 带详细输出的测试
go test -v ./06-engineering/testing/...
```

### 浏览器分析

pprof 和 trace 示例需要浏览器查看：

```bash
# pprof 示例
cd 05-performance/pprof && go run main.go

# trace 示例
cd 05-performance/trace && go run main.go
```

### 需要中间件的示例

```bash
# 启动依赖服务
docker compose -f 07-web/redis/docker-compose.yml up -d
docker compose -f 07-web/kafka/docker-compose.yml up -d

# 运行示例
go run 07-web/redis/main.go
```

## 关联知识库

本仓库是 GoLang 知识库 / 博客系列的配套代码，两者对应关系如下：

| 代码目录 | 博客章节 |
|---|---|
| 01-basics/ | 02-核心语法与数据结构 |
| 02-datastructures/ | 02-核心语法与数据结构 |
| 03-concurrency/ | 03-并发编程 |
| 04-runtime/ | 04-底层原理与运行时 |
| 05-performance/ | 05-性能调优与故障排查 |
| 06-engineering/ | 06-工程实践 |
| 07-web/ | 07-Web 框架与中间件 |
| 08-advanced/ | 09-高级特性与 Go 未来 |
| 09-projects/ | 10-实战项目与案例 |

建议先阅读对应博客文章理解原理，再动手运行示例代码加深印象。

## License

MIT License