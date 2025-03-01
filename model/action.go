package model

import (
	"ai-stream-bot/consts"
	"ai-stream-bot/dal/cache"
	"context"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type ActionMsgInfo struct {
	ChatType  consts.ChatType
	MsgType   consts.MsgType
	MsgId     *string
	ChatId    *string
	UserId    string
	Content   string
	SessionId *string
	Mention   []*larkim.MentionEvent
}

type MsgActionInfo struct {
	Ctx           context.Context
	ActionMsgInfo *ActionMsgInfo
	MsgCache      *cache.MsgCache
	SessionCache  *cache.SessionCache
	MessageEvent  *larkim.P2MessageReceiveV1
}

type CardActionInfo struct {
	Ctx          context.Context
	Kind         consts.CardKind `json:"kind"`
	ChatType     consts.ChatType `json:"chatType"`
	Value        interface{}     `json:"value"`
	SessionId    string          `json:"sessionId"`
	MsgId        string          `json:"msgId"`
	SessionCache *cache.SessionCache
}

type MsgAction interface {
	Execute(data *MsgActionInfo) bool
}

type CardAction interface {
	Execute(data *CardActionInfo) (*larkcard.MessageCard, bool)
}

type StreamUpdateMessage struct {
	Thinking  string
	Reference string
	Answer    string
}
