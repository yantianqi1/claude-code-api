# Claude API Gateway

一个功能完善的 Claude Code API 聚合中转服务，支持多上游渠道管理、模型名称映射、快速渠道切换以及详细的 Token 消耗统计。

## 功能特性

- **多渠道管理** - 支持配置多个上游 API 渠道，按优先级自动切换
- **模型名称映射** - 将自定义模型名映射到上游实际模型名
- **Token 统计** - 详细的 Token 消耗统计和成本分析
- **请求日志** - 完整的请求/响应日志记录
- **Web 管理界面** - 基于 React + Ant Design 的管理控制台
- **Docker 部署** - 一键部署，开箱即用

## 技术栈

- **后端**: Go (Gin) + SQLite
- **前端**: React + TypeScript + Vite + Ant Design
- **部署**: Docker + Docker Compose

## 项目结构

```
claude-api-gateway/
├── backend/                    # Go 后端服务
│   ├── cmd/server/main.go      # 应用入口
│   ├── internal/
│   │   ├── api/handler/        # HTTP 处理器
│   │   ├── api/middleware/     # 中间件
│   │   ├── api/router/         # 路由配置
│   │   ├── service/            # 业务逻辑层
│   │   ├── repository/         # 数据访问层
│   │   ├── model/              # 数据模型
│   │   ├── config/             # 配置管理
│   │   └── proxy/              # 代理转发核心
│   ├── pkg/
│   │   ├── database/           # 数据库初始化
│   │   └── logger/             # 日志组件
│   ├── migrations/             # 数据库迁移
│   └── go.mod
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── api/                # API 客户端
│   │   ├── components/         # 通用组件
│   │   ├── pages/              # 页面组件
│   │   ├── types/              # TypeScript 类型
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
├── docker/                     # Docker 配置
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   ├── nginx.conf
│   └── docker-compose.yml
└── README.md
```

## 快速开始

### 使用 Docker Compose (推荐)

```bash
# 进入 docker 目录
cd docker

# 启动服务
docker-compose up -d

# 访问管理界面
# 前端: http://localhost
# 后端 API: http://localhost:8080
```

### 手动部署

#### 后端

```bash
cd backend

# 安装依赖
go mod download

# 运行服务
go run cmd/server/main.go

# 或编译后运行
go build -o gateway cmd/server/main.go
./gateway
```

#### 前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
```

## API 端点

### 代理端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/messages` | POST | Anthropic API 兼容的消息端点 |

### 管理端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/channels` | GET | 获取渠道列表 |
| `/api/channels` | POST | 创建渠道 |
| `/api/channels/:id` | GET | 获取单个渠道 |
| `/api/channels/:id` | PUT | 更新渠道 |
| `/api/channels/:id` | DELETE | 删除渠道 |
| `/api/channels/:id/activate` | PUT | 激活渠道 |
| `/api/channels/:id/deactivate` | PUT | 停用渠道 |
| `/api/channels/test` | POST | 测试连接 |
| `/api/mappings` | GET | 获取映射列表 |
| `/api/mappings` | POST | 创建映射 |
| `/api/mappings/:id` | PUT | 更新映射 |
| `/api/mappings/:id` | DELETE | 删除映射 |
| `/api/stats` | GET | 获取统计数据 |
| `/api/stats/logs` | GET | 获取请求日志 |
| `/api/stats/export` | GET | 导出 CSV |

## 使用示例

### 代理请求

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -d '{
    "model": "claude-sonnet-4-5",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 创建渠道

```bash
curl -X POST http://localhost:8080/api/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Anthropic Primary",
    "base_url": "https://api.anthropic.com",
    "api_key": "sk-ant-xxx",
    "provider": "anthropic",
    "priority": 10
  }'
```

### 创建模型映射

```bash
curl -X POST http://localhost:8080/api/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": 1,
    "display_model": "claude-sonnet-4-5",
    "upstream_model": "claude-3-5-sonnet-20241022"
  }'
```

## 配置

### 环境变量

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `SERVER_PORT` | 8080 | 服务端口 |
| `DATA_DIR` | ./data | 数据存储目录 |
| `DEBUG` | false | 调试模式 |
| `ENABLE_CORS` | true | 启用 CORS |
| `ALLOWED_ORIGINS` | * | 允许的跨域来源 |

## 数据库

项目使用 SQLite 作为数据库，数据文件默认存储在 `./data/gateway.db`。

数据库表结构：
- `channels` - 上游渠道配置
- `model_mappings` - 模型名称映射
- `request_logs` - 请求日志
- `system_configs` - 系统配置

## 许可证

MIT License
