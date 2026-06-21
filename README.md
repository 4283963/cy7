# 云剪贴板

一个简洁的多端云剪贴板工具，通过 4 位提取码在设备间传递文本内容。

## 功能特性

- 上传文本内容，自动生成 4 位数字提取码
- 通过提取码跨设备获取文本内容
- 内容 24 小时后自动过期
- 简洁的响应式网页界面，支持手机和电脑
- 一键复制提取码和内容

## 技术栈

- **后端**: Go + Gin 框架
- **存储**: Redis
- **前端**: 原生 HTML/JavaScript

## 项目结构

```
cy7/
├── backend/
│   ├── main.go              # 主入口，路由设置
│   ├── go.mod               # Go 模块依赖
│   ├── redis/
│   │   └── client.go        # Redis 客户端封装
│   └── handlers/
│       └── clipboard.go     # API 处理器
├── frontend/
│   └── index.html           # 前端页面
└── README.md
```

## API 接口

### 保存内容
```
POST /api/save
Content-Type: application/json

{
  "content": "要保存的文本内容"
}

响应:
{
  "code": "1234"
}
```

### 获取内容
```
GET /api/get/{code}

响应:
{
  "content": "保存的文本内容"
}
```

### 健康检查
```
GET /api/health
```

## 快速开始

### 1. 安装 Redis

确保本地已安装并启动 Redis：

```bash
# macOS
brew install redis
redis-server
```

### 2. 运行服务

```bash
cd backend
go mod download
go run main.go
```

### 3. 访问应用

打开浏览器访问 `http://localhost:8080`

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `REDIS_PASSWORD` | （空） | Redis 密码 |
| `PORT` | `8080` | 服务端口 |
| `GIN_MODE` | `release` | Gin 运行模式 |

## 使用示例

### 命令行测试

保存内容：
```bash
curl -X POST http://localhost:8080/api/save \
  -H "Content-Type: application/json" \
  -d '{"content": "hello world"}'
```

获取内容：
```bash
curl http://localhost:8080/api/get/1234
```

## 部署

### 编译二进制

```bash
cd backend
go build -o cloudclipboard .
```

### 运行

```bash
./cloudclipboard
```

## 安全说明

- 提取码为 4 位随机数字，最多 10000 种组合
- 内容 24 小时后自动删除
- 不适合存储敏感信息
- 生产环境建议添加访问频率限制

## 许可证

MIT
