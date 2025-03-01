package service

import (
	"ai-stream-bot/client/im"
	"ai-stream-bot/config"
	"ai-stream-bot/consts"
	"ai-stream-bot/model"
	"ai-stream-bot/pkg/feishu"
	"strings"
	"time"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
)

type Service interface {
	Execute(a *model.MsgActionInfo) bool
}

type ProcessedUniqueService struct {
}

func (s *ProcessedUniqueService) Execute(action *model.MsgActionInfo) bool {
	msgId := action.ActionMsgInfo.MsgId
	if msgId == nil {
		return false
	}
	_, found := action.MsgCache.IfProcessed(*msgId)
	if found {
		return false
	}

	action.MsgCache.Process(action.Ctx, *msgId, true, time.Hour*10)
	return true
}

type ProcessMentionService struct {
}

func (s *ProcessMentionService) Execute(action *model.MsgActionInfo) bool {
	// 私聊消息，直接返回
	if action.ActionMsgInfo.ChatType == consts.UserChatType {
		return true
	}
	// 群聊消息，判断是否@了机器人，且仅仅艾特了机器人
	if action.ActionMsgInfo.ChatType == consts.GroupChatType {
		mention := action.ActionMsgInfo.Mention
		if mention == nil {
			return false
		}
		if len(mention) != 1 {
			return false
		}
		if *mention[0].Name == config.GetFeishuConfig().BotName {
			return true
		}
		return false
	}
	return true
}

type EmptyService struct {
}

func (s *EmptyService) Execute(action *model.MsgActionInfo) bool {
	if action.ActionMsgInfo.Content == "" {
		// 空消息，直接返回
		card := feishu.BuildMessageCard(
			feishu.BuildCardHeader("️🆑 DeepSeek友情提示", larkcard.TemplateGrey),
			feishu.BuildCardNote("🤖️：你想知道什么呢~"),
		)
		cardStr, _ := card.String()
		im.GetFeishuClient().FeishuReplyMsg(action.Ctx, *action.ActionMsgInfo.MsgId, cardStr)
		return false
	}
	return true
}

type CommandService struct {
}

func (s *CommandService) Execute(action *model.MsgActionInfo) bool {
	content := action.ActionMsgInfo.Content
	commandGroups := map[string][]string{
		"clearCommands": {"/clear", "开始新会话"},
		"helpCommands":  {"/help", "帮助"},
	}

	commandActions := map[string]func(){
		"clearCommands": func() {
			action.SessionCache.Clear(*action.ActionMsgInfo.SessionId)
			card := feishu.BuildMessageCard(
				feishu.BuildCardHeader("️🆑 DeepSeek友情提示", larkcard.TemplateGrey),
				feishu.BuildCardNote("已清除此话题的上下文信息"),
				feishu.BuildCardNote("我们可以开始一个全新的话题，继续找我聊天吧"),
			)
			cardStr, _ := card.String()
			im.GetFeishuClient().FeishuReplyMsg(action.Ctx, *action.ActionMsgInfo.MsgId, cardStr)
		},
		"helpCommands": func() {
			card := feishu.BuildMessageCard(
				feishu.BuildCardHeader("🎒需要帮助吗？", larkcard.TemplateBlue),
				feishu.BuildCardMainMd("**我是您的贴心助手**"),
				feishu.BuildCardSplitLine(),
				feishu.BuildCardMdAndButton("** 🆑 清除话题上下文**\n文本回复*/clear*",
					feishu.BuildEmbedButton("开始新会话", map[string]interface{}{
						"kind":      consts.ClearCard,
						"chatType":  action.ActionMsgInfo.ChatType,
						"sessionId": *action.ActionMsgInfo.SessionId,
						"msgId":     *action.ActionMsgInfo.MsgId,
					}, larkcard.MessageCardButtonTypeDanger),
				),
				feishu.BuildCardSplitLine(),
				feishu.BuildCardMainMd("🎒 **需要更多帮助**\n文本回复 *帮助* 或 */help*"),
				feishu.BuildCardSplitLine(),
				feishu.BuildCardMainMd("🎒 **有啥想法反馈，请随时告诉我！**"),
			)
			cardStr, _ := card.String()
			im.GetFeishuClient().FeishuReplyMsg(action.Ctx, *action.ActionMsgInfo.MsgId, cardStr)
		},
	}

	for group, cmds := range commandGroups {
		for _, cmd := range cmds {
			if strings.HasPrefix(content, cmd) {
				commandActions[group]()
				return false
			}
		}
	}
	return true
}

type ClearCardService struct {
}

func (s *ClearCardService) Execute(action *model.CardActionInfo) (*larkcard.MessageCard, bool) {
	if action.Kind != consts.ClearCard {
		return nil, false
	}
	// 清除上下文
	action.SessionCache.Clear(action.SessionId)
	// 更新卡片内容
	card := feishu.BuildMessageCard(
		feishu.BuildCardHeader("️🆑 DeepSeek友情提示", larkcard.TemplateGrey),
		feishu.BuildCardNote("已清除此话题的上下文信息"),
		feishu.BuildCardNote("我们可以开始一个全新的话题，继续找我聊天吧"),
	)
	return card, true
}
