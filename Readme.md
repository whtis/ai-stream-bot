# AI Stream Bot

一个基于 Go 语言开发的 AI 对话机器人，支持流式输出思考过程及参考文献和上下文管理。适配 DeepSeek-R1 等深度思考模型，通过简单的配置即可使用及二次开发。
- 本项目模型代码以火山为例，因为 openAI 官方 API 暂未实现深度思考和参考文献输出；代码已预留扩展接口，新的 SDK 接入只需实现对应 SDK 逻辑，修改少量接口即可实现。
- 本项目机器人实现以飞书为例，代码已预留其它类型 BOT 接口，接入只需实现对应 BOT 的方法接口。

> ⚠️ 免责声明：本项目仅供学习和研究使用，请遵守相关法律法规和服务条款。

## ✨ 功能特点

- 🚀 流式输出：支持 AI 回复的流式实时输出，展示思考过程
- 💭 深度思考：支持 DeepSeek-R1 等深度思考模型，输出参考文献
- 💬 智能对话：支持群聊和单聊场景，自动区分会话上下文
- 📦 扩展性强：预留模型和机器人接口，易于扩展其他平台
- ⚡️ 高性能：基于 Go 语言开发，并发性能优异
- 🔄 会话管理：支持 12 小时会话缓存，可自定义配置
- 🎯 智能优化：超出长度限制时自动保留关键对话内容
- 🛡️ 内存优化：智能管理内存占用，自动清理过期会话

## 🛠️ 技术栈

- 🚀 [Hertz](https://github.com/cloudwego/hertz) - 高性能 HTTP 框架
- 💾 [go-cache](https://github.com/patrickmn/go-cache) - 高性能内存缓存
- 🤖 AI 模型集成
  - DeepSeek-R1
  - 火山引擎
  - 预留其他模型接口
- 🔧 其他依赖
  - Go >= 1.20
  - 支持 Linux/MacOS/Windows

## 📦 安装与配置

```bash
# 克隆项目
git clone https://github.com/whtis/ai-stream-bot.git

# 进入项目目录
cd ai-stream-bot

# 安装依赖
go mod download

# 复制配置文件模板
cp config_example.yaml config.yaml

# 修改配置文件
vim config.yaml  # 或使用其他编辑器

# 运行项目
go run main.go
```

### ⚙️ 配置说明

项目通过 `config.yaml` 进行配置，主要配置项包括：

```yaml
bot:
  # 飞书机器人配置
  app_id: "your_app_id"           # 飞书应用 ID
  app_secret: "your_app_secret"   # 飞书应用密钥
  encrypt_key: "your_encrypt_key" # 飞书消息加密密钥（可选）
  verification_token: "your_token" # 飞书验证令牌（可选）

ai:
  # AI 模型配置
  model: "deepseek-r1"           # 使用的模型，支持 deepseek-r1
  api_key: "your_api_key"        # API 密钥
  base_url: "your_base_url"      # API 基础URL（可选）

```

> 注意：首次运行前必须配置 bot 和 ai 模型相关参数，否则机器人将无法正常工作。

## 📝 特性说明

### 上下文管理
- 自动管理对话上下文长度
- 超过限制时智能保留奇数位置的消息
- 默认支持最大 8192 长度的上下文

### 会话管理
- 支持多会话并发
- 12 小时自动过期
- 支持手动清理会话

## 🙏 致谢

本项目在开发过程中参考和借鉴了以下优秀的开源项目：

- [Feishu-OpenAI-Stream-Chatbot](https://github.com/ConnectAI-E/Feishu-OpenAI-Stream-Chatbot) - 可以流式输出文本的飞书openai机器人 Feishu-OpenAI robot that can stream chat

感谢这些项目的作者和贡献者们！

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[Apache License 2.0](LICENSE)
