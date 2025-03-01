package im

import (
	"ai-stream-bot/config"
	"ai-stream-bot/model"
	"ai-stream-bot/pkg"
	"context"
	"errors"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

var (
	feishuClient *FeishuClient
)

type FeishuClient struct {
	*lark.Client
}

func NewFeishuClient(cfg *config.FeishuConfig, eventHandler *dispatcher.EventDispatcher) {
	larkClient := lark.NewClient(cfg.AppID, cfg.AppSecret)
	larkWsClient := larkws.NewClient(cfg.AppID, cfg.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelDebug))

	feishuClient = &FeishuClient{
		larkClient,
	}
	// 启动飞书 WebSocket 连接
	go func() {
		err := larkWsClient.Start(context.Background())
		if err != nil {
			hlog.Errorf("启动飞书 WebSocket 连接失败: %v", err)
			os.Exit(1)
		}
	}()

}

func GetFeishuClient() *FeishuClient {
	return feishuClient
}

func (f *FeishuClient) FeishuReplyMsg(ctx context.Context, msgId string, content string) (*string, error) {
	resp, err := f.Client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeInteractive).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		hlog.Errorf("replyMsg returned error: %v", err)
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		hlog.Errorf("replyMsg returned error: %v", resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.MessageId, nil
}

func (f *FeishuClient) FeishuCreateCard(ctx context.Context) (*string, error) {
	req := larkcardkit.NewCreateCardReqBuilder().
		Body(larkcardkit.NewCreateCardReqBodyBuilder().
			Type(`card_json`).
			Data(`{
    "schema": "2.0",
    "config": {
        "update_multi": true,
        "streaming_mode": true,
        "streaming_config": {
            "print_step": {
                "default": 1
            },
            "print_frequency_ms": {
                "default": 70
            },
            "print_strategy": "fast"
        },
        "style": {
            "text_size": {
                "normal_v2": {
                    "default": "normal",
                    "pc": "normal",
                    "mobile": "heading"
                }
            }
        }
    },
    "body": {
        "direction": "vertical",
        "padding": "12px 12px 12px 12px",
        "elements": [
            {
                "tag": "markdown",
                "content": "",
                "text_align": "left",
                "text_size": "notation",
                "margin": "0px 0px 0px 0px",
                "element_id": "think"
            },
            {
                "tag": "markdown",
                "content": "",
                "text_align": "left",
                "text_size": "normal_v2",
                "margin": "0px 0px 0px 0px",
                "element_id": "answer"
            },
            {
                "tag": "markdown",
                "content": "",
                "text_align": "left",
                "text_size": "normal_v2",
                "margin": "0px 0px 0px 0px",
                "element_id": "reference"
            }
        ]
    }
}`).
			Build()).
		Build()
	// 发起请求
	resp, err := f.Client.Cardkit.V1.Card.Create(ctx, req)

	// 处理错误
	if err != nil {
		hlog.Errorf("FeishuCreateCard returned error: %v", err)
		return nil, err
	}
	// 服务端错误处理
	if !resp.Success() {
		hlog.Errorf("FeishuCreateCard returned error: %v", resp.Code, resp.Msg, resp.RequestId())
		return nil, errors.New(resp.Msg)
	}
	return resp.Data.CardId, nil
}

func (f *FeishuClient) FeishuUpdateCard(ctx context.Context, update model.StreamUpdateMessage, cardId string) error {
	var resp *larkcardkit.ContentCardElementResp
	var err error
	needRequest := false

	// 创建请求对象
	if update.Thinking != "" {
		req := larkcardkit.NewContentCardElementReqBuilder().
			CardId(cardId).
			ElementId(`think`).
			Body(larkcardkit.NewContentCardElementReqBodyBuilder().
				Uuid(uuid.New().String()).
				Content(update.Thinking).
				Sequence(pkg.NextSequence()).
				Build()).
			Build()
		// 发起请求
		resp, err = f.Client.Cardkit.V1.CardElement.Content(ctx, req)
		needRequest = true
	}
	if update.Answer != "" {
		// 创建请求对象
		req := larkcardkit.NewContentCardElementReqBuilder().
			CardId(cardId).
			ElementId(`answer`).
			Body(larkcardkit.NewContentCardElementReqBodyBuilder().
				Uuid(uuid.New().String()).
				Content(update.Answer).
				Sequence(pkg.NextSequence()).
				Build()).
			Build()
		// 发起请求
		resp, err = f.Client.Cardkit.V1.CardElement.Content(ctx, req)
		needRequest = true
	}
	if update.Reference != "" {
		// 创建请求对象
		req := larkcardkit.NewContentCardElementReqBuilder().
			CardId(cardId).
			ElementId(`reference`).
			Body(larkcardkit.NewContentCardElementReqBodyBuilder().
				Uuid(uuid.New().String()).
				Content(update.Reference).
				Sequence(pkg.NextSequence()).
				Build()).
			Build()
		// 发起请求
		resp, err = f.Client.Cardkit.V1.CardElement.Content(ctx, req)
		needRequest = true
	}
	if !needRequest {
		return nil
	}
	// 处理错误
	if err != nil {
		hlog.Errorf("FeishuUpdateCard returned error: %v", err)
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		hlog.Errorf("FeishuUpdateCard returned error: %v", resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func (f *FeishuClient) FeishuUpdateCardSetting(ctx context.Context, cardId string) error {
	time.Sleep(500 * time.Millisecond)
	// 创建请求对象
	req := larkcardkit.NewSettingsCardReqBuilder().
		CardId(cardId).
		Body(larkcardkit.NewSettingsCardReqBodyBuilder().
			Settings(`{"config":{"streaming_mode":false,"enable_forward":true,"update_multi":true,"width_mode":"fill","enable_forward_interaction":false},"card_link":{"url":"https://applink.feishu.cn/T8UcoLiLPJyV","android_url":"https://applink.feishu.cn/T8UcoLiLPJyV","ios_url":"https://applink.feishu.cn/T8UcoLiLPJyV","pc_url":"https://applink.feishu.cn/T8UcoLiLPJyV"}}`).
			Uuid(uuid.New().String()).
			Sequence(pkg.NextSequence()).
			Build()).
		Build()

	// 发起请求
	resp, err := f.Client.Cardkit.V1.Card.Settings(ctx, req)

	// 处理错误
	if err != nil {
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		return errors.New(resp.Msg)
	}
	return nil
}
