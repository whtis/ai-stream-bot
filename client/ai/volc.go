package ai

import (
	"ai-stream-bot/config"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type VolcClient struct {
	cfg    *config.VolcConfig
	client *arkruntime.Client
}

var volcClient *VolcClient

func NewVolcClient(cfg *config.VolcConfig) *VolcClient {
	if volcClient == nil {
		client := arkruntime.NewClientWithApiKey(
			cfg.APIKey,
			arkruntime.WithBaseUrl(cfg.APIURL),
		)
		return &VolcClient{
			cfg:    cfg,
			client: client,
		}
	}
	return volcClient
}

func (c *VolcClient) GetProvider() Provider {
	return ProviderVolc
}

func (c *VolcClient) StreamChat(ctx context.Context, req *AiChatStreamRequest) error {
	chatMsgs := make([]*model.ChatCompletionMessage, len(req.Msgs))
	for i, m := range req.Msgs {
		chatMsgs[i] = &model.ChatCompletionMessage{
			Role:    m.Role,
			Content: &model.ChatCompletionMessageContent{StringValue: volcengine.String(m.Content)},
		}
	}
	return c.StreamChatWithHistory(ctx, chatMsgs, MaxTokens, req.ThinkStream, req.AnswerStream, req.RefStream)
}

func (c *VolcClient) StreamChatWithHistory(ctx context.Context, msg []*model.ChatCompletionMessage, maxTokens int, thinkStream, answerStream, refStream chan string) error {
	req := model.BotChatCompletionRequest{
		BotId:       c.cfg.Model,
		Messages:    msg,
		N:           1,
		Temperature: 0.7,
		MaxTokens:   maxTokens,
		TopP:        1,
	}
	stream, err := c.client.CreateBotChatCompletionStream(ctx, req)
	if err != nil {
		hlog.Errorf("CreateBotChatCompletionStream returned error: %v", err)
		return err
	}
	defer stream.Close()
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			hlog.Errorf("Stream error: %v\n", err)
			return err
		}
		if len(response.Choices) > 0 {
			if response.References != nil {
				for i, ref := range response.References {
					num := i + 1
					refStream <- fmt.Sprintf("[%d] [%s](%s)\n", num, ref.Title, ref.Url)
				}
			}
			if response.Choices[0].Delta.ReasoningContent != nil {
				thinkStream <- *response.Choices[0].Delta.ReasoningContent
			} else {
				answerStream <- response.Choices[0].Delta.Content
			}
		}
	}
}
