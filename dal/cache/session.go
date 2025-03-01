package cache

import (
	"ai-stream-bot/client/ai"
	"ai-stream-bot/consts"
	"ai-stream-bot/pkg"
	"time"

	"github.com/patrickmn/go-cache"
)

type SessionCache struct {
	cache *cache.Cache
}

var sessionCache *SessionCache

func GetSessionCache() *SessionCache {
	return sessionCache
}

func NewSessionCache() {
	sessionCache = &SessionCache{cache: cache.New(12*time.Hour, 12*time.Hour)}
}

func (s *SessionCache) GetMsg(sessionId string) []ai.AiMessage {
	msgs, ok := s.cache.Get(sessionId)
	if !ok {
		return nil
	}
	return msgs.([]ai.AiMessage)
}

func (s *SessionCache) SetMsg(sessionId string, msgs []ai.AiMessage) {
	// 限制上下文长度
	if pkg.GetStrPoolTotalLength(msgs) > consts.MaxContextLength {
		// 创建新的切片来存储单数位置的消息
		newMsgs := make([]ai.AiMessage, 0)
		for i := 0; i < len(msgs); i++ {
			if i%2 == 0 { // 只保留索引为偶数的消息（对应1,3,5...位置）
				newMsgs = append(newMsgs, msgs[i])
			}
		}
		msgs = newMsgs
	}

	s.cache.Set(sessionId, msgs, 12*time.Hour)
}

func (s *SessionCache) Clear(sessionId string) {
	s.cache.Delete(sessionId)
}
