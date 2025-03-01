package ai

import (
	"context"
	"fmt"
	"sync"
)

// Provider 定义 AI 服务提供商类型
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderVolc   Provider = "volc"
)

const (
	MaxTokens = 8092
)

// AiMessage 定义消息结构
type AiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiChatStreamRequest struct {
	Ctx          context.Context
	Msgs         []AiMessage `json:"msgs"`
	ThinkStream  chan string `json:"think_stream"`
	AnswerStream chan string `json:"answer_stream"`
	RefStream    chan string `json:"ref_stream"`
}

// Client 定义 AI 客户端接口
type Client interface {
	// GetProvider 获取服务提供商类型
	GetProvider() Provider

	StreamChat(ctx context.Context, req *AiChatStreamRequest) error
}

// Manager AI 客户端管理器
type Manager struct {
	clients       map[Provider]Client
	defaultClient Client
	mu            sync.RWMutex
}

var (
	manager *Manager
	once    sync.Once
)

// GetManager 获取 AI 客户端管理器单例
func GetManager() *Manager {
	once.Do(func() {
		manager = &Manager{
			clients: make(map[Provider]Client),
		}
	})
	return manager
}

// RegisterClient 注册 AI 客户端
func (m *Manager) RegisterClient(client Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.GetProvider()] = client
}

// SetDefaultClient 设置默认 AI 客户端
func (m *Manager) SetDefaultClient(provider Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[provider]
	if !ok {
		return fmt.Errorf("provider %s not registered", provider)
	}

	m.defaultClient = client
	return nil
}

// GetClient 获取指定提供商的客户端
func (m *Manager) GetClient(provider Provider) (Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", provider)
	}
	return client, nil
}

// Chat 使用默认客户端发送聊天请求
func (m *Manager) StreamChat(ctx context.Context, req *AiChatStreamRequest) error {
	m.mu.RLock()
	client := m.defaultClient
	m.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("no default client set")
	}

	return client.StreamChat(ctx, req)
}

// Factory 定义 AI 客户端工厂接口
type Factory interface {
	// CreateClient 创建 AI 客户端
	CreateClient(provider Provider) (Client, error)
}
