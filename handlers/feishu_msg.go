package handlers

import (
	"ai-stream-bot/client/ai"
	"ai-stream-bot/consts"
	"ai-stream-bot/dal/cache"
	"ai-stream-bot/model"
	"ai-stream-bot/pkg"
	"ai-stream-bot/service"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type FeishuMsgHandler struct {
}

func NewFeishuMsgHandler() *FeishuMsgHandler {
	return &FeishuMsgHandler{}
}

func (h *FeishuMsgHandler) Handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	handlerType := judgeChatType(event)
	if handlerType == consts.OtherChatType {
		hlog.Infof("unknown chat type")
		return nil
	}

	msgType, err := judgeMsgType(event)
	if err != nil {
		hlog.Errorf("error getting message type: %v", err)
		return nil
	}

	msgContent := event.Event.Message.Content
	msgId := event.Event.Message.MessageId
	rootId := event.Event.Message.RootId
	chatId := event.Event.Message.ChatId
	mention := event.Event.Message.Mentions

	sessionId := rootId
	if sessionId == nil || *sessionId == "" {
		sessionId = msgId
	}
	actionMsgInfo := model.ActionMsgInfo{
		ChatType:  handlerType,
		MsgType:   msgType,
		MsgId:     msgId,
		UserId:    *event.Event.Sender.SenderId.UserId,
		ChatId:    chatId,
		Content:   strings.Trim(parseContent(*msgContent, msgType), " "),
		SessionId: sessionId,
		Mention:   mention,
	}
	data := &model.MsgActionInfo{
		Ctx:           ctx,
		ActionMsgInfo: &actionMsgInfo,
		MsgCache:      cache.GetMsgCache(),
		SessionCache:  cache.GetSessionCache(),
		MessageEvent:  event,
	}
	actions := []model.MsgAction{
		&service.ProcessedUniqueService{},            // 避免重复处理
		&service.ProcessMentionService{},             // 判断机器人是否应该被调用
		&service.EmptyService{},                      // 空消息处理
		&service.CommandService{},                    // 清除消息处理
		service.NewFeishuMsgService(ai.GetManager()), // 消息处理
	}

	msgChain(data, actions...)
	return nil
}

func parseContent(content string, msgType consts.MsgType) string {
	if msgType == consts.MsgTypeText {
		//"{\"text\":\"@_user_1  hahaha\"}",
		//only get text content hahaha
		var contentMap map[string]interface{}
		err := json.Unmarshal([]byte(content), &contentMap)
		if err != nil {
			hlog.Errorf("error unmarshalling content: %v", err)
			return ""
		}
		if contentMap["text"] == nil {
			return ""
		}
		text := contentMap["text"].(string)
		return msgFilter(text)
	} else if msgType == consts.MsgTypePost {
		result, err := pkg.ExtractTextFromFeishuMessage(content)
		if err != nil {
			hlog.Errorf("error extracting text from feishu message: %v", err)
			return ""
		}
		return msgFilter(result)
	}
	return ""
}

func msgFilter(msg string) string {
	//replace @到下一个非空的字段 为 ''
	regex := regexp.MustCompile(`@[^ ]*`)
	return regex.ReplaceAllString(msg, "")

}

func judgeChatType(event *larkim.P2MessageReceiveV1) consts.ChatType {
	chatType := event.Event.Message.ChatType
	if *chatType == "group" {
		return consts.GroupChatType
	}
	if *chatType == "p2p" {
		return consts.UserChatType
	}
	return consts.OtherChatType
}

func judgeMsgType(event *larkim.P2MessageReceiveV1) (consts.MsgType, error) {
	msgType := event.Event.Message.MessageType

	switch *msgType {
	case string(consts.MsgTypeText):
		return consts.MsgTypeText, nil
	case string(consts.MsgTypePost):
		return consts.MsgTypePost, nil
	default:
		return "", fmt.Errorf("unknown message type: %v", *msgType)
	}
}

// 责任链
func msgChain(data *model.MsgActionInfo, actions ...model.MsgAction) bool {
	for _, v := range actions {
		if !v.Execute(data) {
			return false
		}
	}
	return true
}
