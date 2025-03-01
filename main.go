package main

import (
	"ai-stream-bot/dal/cache"
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	hertzlogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/sirupsen/logrus"

	"ai-stream-bot/client/ai"
	"ai-stream-bot/client/im"
	"ai-stream-bot/config"
	"ai-stream-bot/consts"
	"ai-stream-bot/handlers"
)

// setupRouter 设置路由
func setupRouter(h *server.Hertz) {
	// 健康检查接口
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, utils.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	// 根据配置文件启动对应的机器人
	// 启动飞书机器人
	if config.IsFeishuEnabled() {
		hlog.Info("启动飞书机器人")
		feishuCfg := config.GetFeishuConfig()
		eventHandler := dispatcher.NewEventDispatcher(feishuCfg.AppVerificationToken, feishuCfg.AppEncryptKey)
		msgHandler := handlers.GetMsgReceiveHandler(consts.BotFeishu).(*handlers.FeishuMsgHandler)
		eventHandler.OnP2MessageReceiveV1(msgHandler.Handle)

		cardHandler := handlers.GetCardActionHandler(consts.BotFeishu).(*handlers.FeishuCardHandler)
		eventHandler.OnP2CardActionTrigger(cardHandler.Handle)

		im.NewFeishuClient(feishuCfg, eventHandler)
	}

	// 启动火山引擎 AI 客户端
	aiManager := ai.GetManager()
	if config.IsVolcEnabled() {
		volcCfg := config.GetVolcConfig()
		aiManager.RegisterClient(ai.NewVolcClient(volcCfg))
		aiManager.SetDefaultClient(ai.ProviderVolc)
	}
	if config.IsOpenAIEnabled() {
		openaiCfg := config.GetOpenAIConfig()
		aiManager.RegisterClient(ai.NewOpenAIClient(openaiCfg))
		aiManager.SetDefaultClient(ai.ProviderOpenAI)
	}
}

func main() {
	err := config.LoadConfig()
	if err != nil {
		hlog.Errorf("加载配置文件失败: %v", err)
		os.Exit(1)
	}
	// 初始化日志
	logger := hertzlogrus.NewLogger()
	logger.Logger().SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			for i := 0; i < 15; i++ { // 遍历调用栈找到第一个非 logrus 和非 hertz 的调用
				pc, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				if !strings.Contains(file, "logrus") && !strings.Contains(file, "hertz") {
					fn := runtime.FuncForPC(pc)
					return fmt.Sprintf("%s:%d", file, line), fn.Name()
				}
			}
			return frame.File, frame.Function
		},
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyFile: "source",
			logrus.FieldKeyFunc: "caller",
		},
	})
	logger.Logger().SetReportCaller(true)
	hlog.SetLogger(logger)
	hlog.SetLevel(hlog.LevelDebug)
	// 初始化 cache
	cache.NewMsgCache()
	cache.NewSessionCache()

	// 创建 Hertz 实例
	h := server.Default(
		server.WithHostPorts(":8888"),
	)
	// 设置路由
	setupRouter(h)

	// 启动服务器
	hlog.Info("服务器启动在 :8888")
	h.Spin()
}
