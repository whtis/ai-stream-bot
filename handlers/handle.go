package handlers

import "ai-stream-bot/consts"

func GetMsgReceiveHandler(bot string) interface{} {
	switch bot {
	case consts.BotFeishu:
		return NewFeishuMsgHandler()
	default:
		return nil
	}
}

func GetCardActionHandler(bot string) interface{} {
	switch bot {
	case consts.BotFeishu:
		return NewFeishuCardHandler()
	default:
		return nil
	}
}
