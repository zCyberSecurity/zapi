# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

zAPI 是一个 AI API 聚合网关：
- 聚合多个 LLM 提供商（Groq、Google AI Studio、OpenRouter、Together AI、Mistral、硅基流动、阿里云百炼、智谱 AI 等）
- 对外提供兼容 OpenAI Chat API 和 Anthropic Messages API 的统一接口
- 后台管理：添加 Provider、管理 API Key、控制模型访问权限
- 每日 Token 用量统计（按 API Key / 模型 / 日期）

## 技术栈

- **后端**：Go 1.24，Gin，GORM + SQLite
- **前端**：Vue 3 + TypeScript，Element Plus，Vite

## Module

`github.com/zCyberSecurity/zapi`

## 后端命令

```bash
# 启动（默认 :8080，admin token: change-me）
go run ./cmd/server

# 环境变量
ADMIN_TOKEN=secret DB_PATH=zapi.db ADDR=:8080 go run ./cmd/server

# 构建
go build -o zapi ./cmd/server

# 测试
go test ./...
go test ./path/to/package -run TestName
```

## 前端命令

```bash
cd frontend
npm run dev      # 开发服务器（代理 /admin 和 /v1 到 :8080）
npm run build    # 生产构建
```

前端访问 http://localhost:5173，用 ADMIN_TOKEN（默认 change-me）登录。

## 架构

### 请求流程

```
客户端
  → middleware/auth.go   (APIKeyAuth: 验证 Bearer token)
  → handler/openai.go    POST /v1/chat/completions, GET /v1/models
  → handler/anthropic.go POST /v1/messages
  → proxy/proxy.go       (FindProvider → 按 model_id 查找启用的 Provider)
  → 上游 LLM API
```

### Anthropic ↔ OpenAI 转换

`proxy/anthropic.go` 负责格式互转：
- `AnthropicToOpenAI`：Anthropic 请求 → OpenAI 请求
- `OpenAIToAnthropic`：OpenAI 响应 → Anthropic 响应
- 若 Provider `api_type = "anthropic"`，则直接透传（不转换）
- 流式（stream）模式目前直接透传 OpenAI SSE，暂未转换为 Anthropic SSE 格式

### 用量统计

非流式请求完成后，从响应体提取 `usage` 字段写入 `usage_logs` 表（按 api_key_id + model + date 聚合）。
流式请求暂不统计用量。

### Admin API（需 Bearer ADMIN_TOKEN）

| 路由 | 说明 |
|------|------|
| GET/POST /admin/providers | Provider 列表 / 创建 |
| PUT/DELETE /admin/providers/:id | 更新 / 删除 |
| GET/POST /admin/providers/:id/models | 模型列表 / 添加 |
| PUT/DELETE /admin/models/:id | 更新 / 删除模型 |
| GET/POST /admin/keys | API Key 列表 / 创建 |
| PUT/DELETE /admin/keys/:id | 更新 / 删除 Key |
| GET /admin/usage?date=&key_id= | 用量查询 |

### ProviderModel 字段说明

- `model_id`：对外暴露的模型名（客户端传此值）
- `provider_model_id`：实际发给上游的模型名（留空则同 model_id）
- 这样可以把 `claude-3-5-sonnet` 映射到 `claude-3-5-sonnet-20241022`
