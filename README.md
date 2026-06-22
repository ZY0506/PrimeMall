# PrimeMall

基于 **go-zero** 微服务框架的电商系统，支持高并发、高可用。

## 技术栈

| 层级    | 技术                                                      | 说明                 |
| ----- | ------------------------------------------------------- | ------------------ |
| 开发语言  | Go 1.25+                                                | —                  |
| 微服务框架 | [go-zero](https://github.com/zeromicro/go-zero) v1.10.1 | 限流、熔断、自适应负载        |
| 服务发现  | etcd v3.5                                               | 服务注册与发现            |
| 网关    | go-zero rest + Nginx                                    | 3 实例网关集群           |
| 数据库   | MySQL 8.0                                               | 7 个业务数据库           |
| 缓存    | Redis 7.2                                               | 会话、库存、商品缓存         |
| 消息队列  | RabbitMQ 3.13                                           | 异步任务、事件驱动          |
| 搜索引擎  | Elasticsearch 7.17                                      | 商品搜索               |
| 鉴权    | JWT + Casbin                                            | Token 认证 + RBAC 权限 |
| 容器化   | Docker Compose                                          | 一键部署基础设施           |

## 项目结构

```
PrimeMall-Server/
├── apps/                                   # 应用层
│   ├── gateway/                            # 网关层（HTTP 统一入口）
│   │   ├── shop/                           # 前台商城网关
│   │   │   ├── etc/                        #   配置（含 3 实例 + Casbin）
│   │   │   └── internal/
│   │   │       ├── handler/                #   路由处理（auth/cart/order/payment/...）
│   │   │       ├── logic/                  #   业务逻辑
│   │   │       ├── svc/                    #   服务上下文
│   │   │       └── types/                  #   请求响应类型
│   │   └── admin/                          # 管理后台网关
│   │       └── internal/
│   │           ├── handler/ (admin,auth)
│   │           ├── logic/
│   │           ├── svc/
│   │           └── types/
│   └── service/                            # RPC 微服务层（gRPC）
│       ├── user/                           # 用户服务（8081）
│       │   ├── rpc/ (client/ etc/ internal/ types/)
│       │   └── sql/
│       ├── product/                        # 商品服务（8082）
│       │   ├── rpc/
│       │   └── sql/
│       ├── order/                          # 订单服务（8083）
│       │   ├── rpc/
│       │   └── sql/
│       ├── payment/                        # 支付服务（8084）
│       │   ├── rpc/
│       │   └── sql/
│       ├── marketing/                      # 营销服务（8085）
│       │   ├── rpc/
│       │   └── sql/
│       ├── search/                         # 搜索服务（8087）
│       │   ├── rpc/
│       │   └── sql/
│       └── admin/                          # 后台管理 RPC（8086）
│           ├── rpc/ (client/ etc/ internal/)
│           └── sql/
├── common/                                 # 公共模块
│   ├── middleware/                         #   网关中间件（CORS/限流/JWT黑名单）
│   ├── jwt/                                #   JWT 工具
│   ├── snowflakes/                         #   雪花算法 ID
│   ├── errorx/                             #   错误定义
│   ├── interceptor/                        #   gRPC 拦截器
│   ├── response/                           #   统一响应
│   ├── constants/                          #   全局常量
│   ├── captcha/                            #   验证码
│   ├── ctxdata/                            #   上下文数据
│   ├── base62/                             #   唯一编码
│   └── utils/                              #   通用工具
├── pkg/                                    # 工具包
│   ├── database/                           #   MySQL 初始化
│   ├── rdb/                                #   Redis 初始化
│   ├── mq/rabbitmq/                        #   消息队列客户端
│   ├── oss/                                #   阿里云 OSS
│   └── pwd/                                #   密码加密（bcrypt）
├── deployments/                            # 部署配置
│   ├── docker/mysql/init/                  #   7 个业务库 DDL + 种子数据
│   └── nginx/                              #   Nginx 负载均衡配置
├── doc/                                    # 文档
│   ├── api/                                #   API 接口文档
│   └── database/                           #   数据库设计
├── Dockerfile                              # 多阶段构建（编译 9 个服务）
├── docker-compose.yml                      # 全服务编排（基础设施 + 应用）
├── .env.example                            # 环境变量模板
├── .dockerignore                           # Docker 构建上下文排除
├── go.mod / go.sum                         # Go 依赖管理
├── start-all.bat / stop-all.bat            # 本地开发脚本
└── struct.md                               # 项目结构说明
```

## 启动项目

### 🐳 方式一：纯 Docker 运行（推荐，无需 Go 环境）

只要目标机器有 **Docker Desktop**，无需安装 Go / MySQL / Redis 等任何依赖。

#### 前置条件

| 依赖             | 说明             |
| -------------- | -------------- |
| Docker Desktop | 最新稳定版          |
| 4 GB+ 可用内存     | 全部服务约占用 3.5 GB |

#### 快速启动

```bash
# 1. 创建环境配置（首次只需一次）
cp .env.example .env

# 2. 构建 + 启动所有服务
docker compose up -d --build
```

> ⏱ 首次构建需要编译 9 个 Go 服务，约 3-8 分钟；后续启动只需几秒。

#### 访问

| 地址                    | 说明                    |
| --------------------- | --------------------- |
| http://localhost:8080 | 🏪 商城前端 API（通过 Nginx） |
| http://localhost:8088 | 🔧 管理后台 API           |

#### 查看日志

```bash
docker compose logs -f          # 所有服务实时日志
docker compose logs -f user-rpc # 只看某个服务
docker compose logs --tail=50   # 最近 50 行
```

> `Ctrl+C` 退出日志追踪。

#### 常用命令

```bash
docker compose ps                # 查看所有容器状态
docker compose up -d             # 启动全部
docker compose down              # 停止全部（数据保留）
docker compose down -v           # 停止全部 + 清除数据
docker compose build             # 重新构建镜像
```

#### 端口总览

| 服务              | 容器内端口 | 宿主机映射 | 协议   |
| --------------- | ----- | ----- | ---- |
| Nginx（统一入口）     | 8080  | 8080  | HTTP |
| admin-api       | 8088  | 8088  | HTTP |
| shop-api × 3 实例 | 9091  | —     | HTTP |
| user.rpc        | 8081  | —     | gRPC |
| product.rpc     | 8082  | —     | gRPC |
| order.rpc       | 8083  | —     | gRPC |
| payment.rpc     | 8084  | —     | gRPC |
| marketing.rpc   | 8085  | —     | gRPC |
| admin.rpc       | 8086  | —     | gRPC |
| search.rpc      | 8087  | —     | gRPC |
| MySQL           | 3306  | 13306 | TCP  |
| Redis           | 6379  | 16379 | TCP  |
| Etcd            | 2379  | 12379 | TCP  |
| RabbitMQ        | 5672  | 5672  | TCP  |
| RabbitMQ UI     | 15672 | 15672 | HTTP |
| Elasticsearch   | 9200  | 19200 | TCP  |

> 容器内 RPC 服务通过 Docker 内部网络相互通信，无需映射到宿主机。

---

### 本地开发

如果你有 Go 开发环境，也可用 `start-all.bat` 在宿主机上直接运行服务。

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
