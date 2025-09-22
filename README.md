# Rose 项目说明

Rose 是一个基于 Go 语言开发的智能代理服务，支持文档检索、嵌入、分割、索引、问答等多种功能，适用于知识管理和智能问答场景。

## 目录结构
- `internal/agent/`：核心代理逻辑，包括检索器、分割器、文档加载器、嵌入器、索引器等。
- `internal/config/`：配置文件解析。
- `internal/handler/`：API 路由及处理逻辑。
- `internal/logic/`：业务逻辑实现。
- `internal/types/`：通用类型定义。
- `internal/utils/`：工具函数。
- `model/`：数据库模型。
- `pkg/`：通用包，如加密、JWT、Milvus 等。
- `etc/rose-api.yaml`：主配置文件。
- `uploads/`：上传的文档。

## 快速开始
1. 安装依赖：
   ```bash
   go mod tidy
   ```
2. 配置 `etc/rose-api.yaml`。
3. 启动服务：
   ```bash
   go run rose.go
   ```

## 测试
- 单元测试位于各模块的 `*_test.go` 文件。
- 运行所有测试：
   ```bash
   go test ./...
   ```

## 主要功能
- 文档检索（Milvus/VikingDB）
- 文档分割（HTML/Markdown/语义分割）
- 文档嵌入与索引
- 智能问答与聊天
- 待办事项管理

## 依赖
- Go 1.20+
- [go-zero](https://github.com/zeromicro/go-zero)
- Milvus/VikingDB

## 贡献
欢迎提交 issue 和 PR。

## License
MIT

