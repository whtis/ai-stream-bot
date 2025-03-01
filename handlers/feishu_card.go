package handlers

import (
	"ai-stream-bot/dal/cache"
	"ai-stream-bot/model"
	"ai-stream-bot/service"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
)

type FeishuCardHandler struct {
}

func NewFeishuCardHandler() *FeishuCardHandler {
	return &FeishuCardHandler{}
}

func (h *FeishuCardHandler) Handle(ctx context.Context, event *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
	actionValue, _ := json.Marshal(event.Event.Action.Value)
	actionInfo := model.CardActionInfo{}
	err := json.Unmarshal(actionValue, &actionInfo)
	if err != nil {
		hlog.Errorf("unmarshal card action failed: %v", err)
		return nil, err
	}
	actionInfo.Ctx = ctx
	actionInfo.SessionCache = cache.GetSessionCache()
	actions := []model.CardAction{
		&service.ClearCardService{},
	}
	card, ok := cardChain(&actionInfo, actions...)
	if !ok {
		return nil, fmt.Errorf("card chain failed")
	}
	callbackCard := &callback.Card{
		Type: "raw",
		Data: card,
	}
	return &callback.CardActionTriggerResponse{Card: callbackCard}, nil
}

// 责任链
func cardChain(data *model.CardActionInfo, actions ...model.CardAction) (*larkcard.MessageCard, bool) {
	for _, v := range actions {
		card, ok := v.Execute(data)
		if !ok {
			return nil, false
		}
		if card != nil {
			return card, true
		}
	}
	return nil, true
}
