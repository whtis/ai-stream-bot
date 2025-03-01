package ai

import (
	"ai-stream-bot/config"
	"context"
)

type OpenAIClient struct {
	cfg *config.OpenAIConfig
}

func NewOpenAIClient(cfg *config.OpenAIConfig) *OpenAIClient {
	return &OpenAIClient{cfg: cfg}
}

func (c *OpenAIClient) StreamChat(ctx context.Context, req *AiChatStreamRequest) error {
	// to be implemented
	return nil
}

func (c *OpenAIClient) GetProvider() Provider {
	return ProviderOpenAI
}
