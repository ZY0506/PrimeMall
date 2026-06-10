# PrimeMall

基于 **go-zero** 微服务框架的电商系统，支持高并发、高可用。

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 开发语言 | Go 1.25+ | — |
| 微服务框架 | [go-zero](https://github.com/zeromicro/go-zero) v1.10.1 | 限流、熔断、自适应负载 |
| 服务发现 | etcd v3.5 | 服务注册与发现 |
| 网关 | go-zero rest + Nginx | 3 实例网关集群 |
| 数据库 | MySQL 8.0 | 7 个业务数据库 |
| 缓存 | Redis 7.2 | 会话、库存、商品缓存 |
| 消息队列 | RabbitMQ 3.13 | 异步任务、事件驱动 |
| 搜索引擎 | Elasticsearch 7.17 | 商品搜索 |
| 鉴权 | JWT + Casbin | Token 认证 + RBAC 权限 |
| 容器化 | Docker Compose | 一键部署基础设施 |

## 项目结构

```
PrimeMall-Server/
├── apps/                                  # 应用层
│   ├── gateway/                           # 网关层（统一入口）
│   │   ├── shop/                          # 前台用户端网关
│   │   │   ├── etc/                       #   网关配置
│   │   │   └── internal/
│   │   │       ├── handler/               #   路由处理器
│   │   │       │   ├── auth/
│   │   │       │   ├── marketing/coupon/
│   │   │       │   ├── order/ (cart,core,aftersale,delivery)
│   │   │       │   ├── payment/ (core,callback)
│   │   │       │   ├── product/ (category,list)
│   │   │       │   ├── search/
│   │   │       │   └── user/ (profile,address)
│   │   │       └── logic/                 #   业务逻辑
│   │   └── admin/                         # 管理后台网关
│   │       └── internal/handler/ (admin,auth)
│   └── service/                           # 微服务层
│       ├── user/rpc/                      # 用户服务
│       │   ├── etc/                       #   RPC 配置
│       │   ├── internal/
│       │   │   ├── logic/ (user,admin,userinternal)
│       │   │   ├── model/                 #   DB 模型
│       │   │   └── server/                #   gRPC 服务端
│       │   ├── client/                    #   客户端 SDK
│       │   ├── types/                     #   protobuf 类型
│       │   └── sql/                       #   建表 SQL
│       ├── product/rpc/                   # 商品服务
│       │   └── internal/logic/ (product,productadmin,productinternal)
│       ├── order/rpc/                     # 订单服务
│       │   └── internal/logic/ (order,cart,orderadmin,orderinternal)
│       ├── payment/rpc/                   # 支付服务
│       │   └── internal/logic/ (payment,paymentadmin,paymentinternal)
│       ├── marketing/rpc/                 # 营销服务
│       │   └── internal/logic/ (marketing)
│       ├── search/rpc/                    # 搜索服务
│       │   └── internal/logic/ (search)
│       └── admin/rpc/                     # 后台管理服务
│           └── internal/logic/ (admin,adminuser,adminproduct,adminorder,adminaftersale)
├── common/                                # 公共模块
│   ├── base62/                            # 唯一编码生成
│   ├── captcha/                           # 验证码工具
│   ├── constants/                         # 全局常量
│   ├── ctxdata/                           # 上下文工具
│   ├── errorx/                            # 业务错误定义
│   ├── interceptor/                       # gRPC 拦截器
│   ├── jwt/                               # JWT 工具
│   ├── middleware/                        # 网关中间件
│   │   ├── casbin.go                      #   RBAC 权限
│   │   ├── client_info.go                 #   客户端信息
│   │   ├── cors.go                        #   跨域
│   │   ├── ratelimit.go                   #   限流
│   │   └── tokenblacklist.go              #   Token 黑名单
│   ├── response/                          # 统一响应格式
│   ├── snowflakes/                        # 雪花算法 ID 生成
│   └── utils/                             # 通用工具函数
├── pkg/                                   # 工具包
│   ├── database/                          # 数据库初始化
│   ├── mq/rabbitmq/                       # 消息队列客户端
│   ├── oss/                               # 阿里云 OSS
│   ├── pwd/                               # 密码加密（bcrypt）
│   └── rdb/                               # Redis 初始化
├── deployments/                           # 部署配置
│   ├── docker/mysql/init/                 # 7 业务库建表 SQL
│   └── nginx/                             # Nginx 负载均衡配置
├── doc/                                   # 文档
│   ├── api/                               # API 接口文档
│   ├── database/                          # 数据库表结构设计
│   └── 系统配置文档.md                     # 性能调优配置参考
├── .env.example                           # 环境变量模板
├── docker-compose.yml                     # 基础设施编排
├── go.mod / go.sum                        # Go 依赖管理
├── struct.md                              # 项目结构说明
├── start-all.bat                          # 一键启动脚本
└── stop-all.bat                           # 一键停止脚本
```

## 启动项目

### 前置条件

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.25+ | 编译运行微服务 |
| Docker Desktop | 最新稳定版 | 运行 MySQL / Redis / Etcd / RabbitMQ / ES |

### ⚡ 一键启动（推荐）

```bash
start-all.bat
```

**自动执行流程：**

| 步骤 | 操作 | 说明 |
|------|------|------|
| 1/5 | 构建服务二进制 | 编译 7 个 RPC 服务 + 2 个网关 |
| 2/5 | 启动 Docker 基础设施 | MySQL / Redis / Etcd / RabbitMQ / ES |
| 3/5 | 启动 RPC 后端服务 | 按依赖顺序启动 6 个 RPC 服务 |
| 4/5 | 启动网关集群 | 3 个网关实例 (9091/9092/9093) |
| 5/5 | 启动 Nginx 负载均衡 | 统一入口 http://127.0.0.1:8080 |

> 启动完成后自动验证各服务健康状态，失败的服务会标记 `FAILED`。

### 🔧 手动启动（分步说明）

如需分步操作或排查问题，可按以下顺序手动启动：

#### 1. 启动基础设施

```bash
docker compose up -d mysql redis etcd rabbitmq elasticsearch
```

> 首次启动会自动拉取镜像，耗时取决于网络。确认容器运行：`docker ps`

#### 2. 构建所有服务

```bash
# 编译 RPC 服务
go build -o bin/user.exe      apps/service/user/rpc/user.go
go build -o bin/product.exe   apps/service/product/rpc/product.go
go build -o bin/order.exe     apps/service/order/rpc/order.go
go build -o bin/payment.exe   apps/service/payment/rpc/payment.go
go build -o bin/marketing.exe apps/service/marketing/rpc/marketing.go
go build -o bin/search.exe    apps/service/search/rpc/search.go
go build -o bin/admin.exe     apps/service/admin/rpc/admin.go

# 编译网关
go build -o bin/shop.exe      apps/gateway/shop/shop.go
go build -o bin/admin-api.exe apps/gateway/admin/admin.go
```

#### 3. 启动 RPC 后端服务（按依赖顺序）

```bash
# 先启动基础服务
bin/user.exe -f apps/service/user/rpc/etc/user.yaml &
bin/product.exe -f apps/service/product/rpc/etc/product.yaml &

# 等待 2 秒后启动依赖它们的服务
bin/order.exe -f apps/service/order/rpc/etc/order.yaml &
bin/payment.exe -f apps/service/payment/rpc/etc/payment.yaml &
bin/marketing.exe -f apps/service/marketing/rpc/etc/marketing.yaml &
bin/search.exe -f apps/service/search/rpc/etc/search.yaml &
bin/admin.exe -f apps/service/admin/rpc/etc/admin.yaml &
```

#### 4. 启动网关集群

```bash
# 启动 3 个网关实例
bin/shop.exe -f apps/gateway/shop/etc/shop-api-9091.yaml &
bin/shop.exe -f apps/gateway/shop/etc/shop-api-9092.yaml &
bin/shop.exe -f apps/gateway/shop/etc/shop-api-9093.yaml &

# 启动管理后台网关
bin/admin-api.exe -f apps/gateway/admin/etc/admin-api.yaml &
```

#### 5. （可选）启动 Nginx 负载均衡

```bash
docker compose up -d nginx-shop
```

### ⏹ 停止服务

```bash
stop-all.bat
```

停止顺序：Nginx → 网关实例 → RPC 后端服务。

> ⚠️ Docker 基础设施（MySQL/Redis 等）会保持运行。如需完全停止：`docker compose down`

### ✅ 验证服务状态

```bash
# 直接验证网关实例
curl http://127.0.0.1:9091/health
curl http://127.0.0.1:9092/health
curl http://127.0.0.1:9093/health

# 验证 Nginx 负载均衡入口
curl http://127.0.0.1:8080/health
```

所有服务正常时返回 `200 OK`。

### 端口总览

| 服务 | 端口 | 协议 |
|------|------|------|
| Nginx（统一入口） | 8080 | HTTP |
| shop-api 实例 1~3 | 9091~9093 | HTTP |
| admin-api | 9090 | HTTP |
| user.rpc | 8081 | gRPC |
| product.rpc | 8082 | gRPC |
| order.rpc | 8083 | gRPC |
| payment.rpc | 8084 | gRPC |
| marketing.rpc | 8085 | gRPC |
| admin.rpc | 8086 | gRPC |
| search.rpc | 8087 | gRPC |

## 架构

```
              客户端请求
                  │
            ┌─────┴──────┐
            │ Nginx:8080 │  ← 负载均衡
            └─────┬──────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
shop-api:9091  shop-api:9092  shop-api:9093  ← 网关集群（3实例）
     │            │            │
     └────────────┼────────────┘
                  │
        ┌─────────┼─────────┐
        │         │         │
   user.rpc  product.rpc  order.rpc  ← RPC 服务
   payment.rpc  search.rpc  marketing.rpc
   admin.rpc
        │         │         │
   ┌────┴────┐ ┌──┴──┐   ┌──┴──┐
   │ MySQL   │ │Redis│   │RMQ  │  ← 基础设施
   │   ES    │ └─────┘   └─────┘
   └─────────┘
```

> 系统调优配置见 [`doc/系统配置文档.md`](doc/系统配置文档.md)
