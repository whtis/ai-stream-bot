package service

import (
	"ai-stream-bot/client/ai"
	"ai-stream-bot/client/im"
	"ai-stream-bot/model"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type FeishuMsgService struct {
	aiManager *ai.Manager
}

func NewFeishuMsgService(aiManager *ai.Manager) *FeishuMsgService {
	return &FeishuMsgService{
		aiManager: aiManager,
	}
}

func (s *FeishuMsgService) Execute(action *model.MsgActionInfo) bool {
	// 1. 返回一个流式卡片
	cardId, _, err := s.sendEditableCard(action)
	if err != nil {
		hlog.Errorf("sendEditableCard returned error: %v", err)
		return false
	}

	thinkingAnswer := "> "
	referenceAnswer := ""
	streamAnswer := ""

	refStream := make(chan string)
	thinkStream := make(chan string)
	answerResponseStream := make(chan string)
	done := make(chan struct{})

	noContentTimeout := time.AfterFunc(10*time.Second, func() {
		hlog.Info("no content timeout")
		close(done)
		err := im.GetFeishuClient().FeishuUpdateCard(action.Ctx, model.StreamUpdateMessage{Answer: "请求超时"}, *cardId)
		if err != nil {
			return
		}
		im.GetFeishuClient().FeishuUpdateCardSetting(action.Ctx, *cardId)
	})

	defer noContentTimeout.Stop()

	msg := action.SessionCache.GetMsg(*action.ActionMsgInfo.SessionId)
	msg = append(msg, ai.AiMessage{
		Role: "user", Content: *&action.ActionMsgInfo.Content,
	})
	go func() {
		defer func() {
			if err := recover(); err != nil {
				err := im.GetFeishuClient().FeishuUpdateCard(action.Ctx, model.StreamUpdateMessage{Answer: "聊天失败"}, *cardId)
				if err != nil {
					hlog.Errorf("FeishuUpdateCard returned error: %v", err)
					return
				}
				im.GetFeishuClient().FeishuUpdateCardSetting(action.Ctx, *cardId)
			}
		}()

		if err := s.aiManager.StreamChat(action.Ctx, &ai.AiChatStreamRequest{
			Msgs:         msg,
			ThinkStream:  thinkStream,
			AnswerStream: answerResponseStream,
			RefStream:    refStream,
		}); err != nil {
			err := im.GetFeishuClient().FeishuUpdateCard(action.Ctx, model.StreamUpdateMessage{Answer: "聊天失败"}, *cardId)
			if err != nil {
				hlog.Errorf("FeishuUpdateCard returned error: %v", err)
				return
			}
			im.GetFeishuClient().FeishuUpdateCardSetting(action.Ctx, *cardId)
			close(done) // 关闭 done 信号
		}

		close(done) // 关闭 done 信号
	}()
	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				updateMsg := model.StreamUpdateMessage{
					Thinking: thinkingAnswer,
					Answer:   streamAnswer,
				}
				err := im.GetFeishuClient().FeishuUpdateCard(action.Ctx, updateMsg, *cardId)
				if err != nil {
					hlog.Errorf("FeishuUpdateCard returned error: %v", err)
					return
				}
			}
		}
	}()

	for {
		select {
		case think, ok := <-thinkStream:
			if !ok {
				continue
			}
			noContentTimeout.Stop()
			thinkingAnswer += think
			thinkingAnswer = strings.ReplaceAll(thinkingAnswer, "\n\n", "\n>")
			hlog.Errorf("think: %s", thinkingAnswer)
		case ref, ok := <-refStream:
			if !ok {
				continue
			}
			noContentTimeout.Stop()
			referenceAnswer += ref
		case res, ok := <-answerResponseStream:
			if !ok {
				continue
			}
			noContentTimeout.Stop()
			streamAnswer += res
		case <-done:
			updateMsg := model.StreamUpdateMessage{
				Thinking:  thinkingAnswer,
				Reference: referenceAnswer,
				Answer:    streamAnswer,
			}
			err := im.GetFeishuClient().FeishuUpdateCard(action.Ctx, updateMsg, *cardId)
			if err != nil {
				hlog.Errorf("FeishuUpdateCard returned error: %v", err)
				return false
			}
			im.GetFeishuClient().FeishuUpdateCardSetting(action.Ctx, *cardId)
			ticker.Stop()
			combinedAnswer := thinkingAnswer + "\n" + streamAnswer + "\n" + referenceAnswer
			msg = append(msg, ai.AiMessage{
				Role:    "assistant",
				Content: streamAnswer,
			})
			action.SessionCache.SetMsg(*action.ActionMsgInfo.SessionId, msg)

			jsonByteArray, err := json.Marshal(msg)
			if err != nil {
				hlog.Errorf("Error marshaling JSON request: UserId: %s , Request: %s , Response: %s", action.ActionMsgInfo.UserId, jsonByteArray, combinedAnswer)
			}
			jsonStr := strings.ReplaceAll(string(jsonByteArray), "\\n", "")
			jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
			hlog.Infof("UserId: %s , Request: %s , Response: %s", action.ActionMsgInfo.UserId, jsonStr, combinedAnswer)
			return false
		}
	}
}

func (s *FeishuMsgService) sendEditableCard(action *model.MsgActionInfo) (*string, *string, error) {
	cardId, err := im.GetFeishuClient().FeishuCreateCard(action.Ctx)
	if err != nil {
		hlog.Errorf("FeishuCreateCard returned error: %v", err)
		return nil, nil, err
	}
	msgId, err := im.GetFeishuClient().FeishuReplyMsg(action.Ctx, *action.ActionMsgInfo.MsgId, fmt.Sprintf(`{ "type": "card","data": {
		"card_id": "%s"
	  }}`, *cardId))
	if err != nil {
		hlog.Errorf("FeishuReplyMsg returned error: %v", err)
		return nil, nil, err
	}

	return cardId, msgId, nil

}
