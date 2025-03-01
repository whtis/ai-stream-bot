package feishu

import (
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
)

// BuildMessageCard 构建消息卡片
func BuildMessageCard(header *larkcard.MessageCardHeader, elements ...larkcard.MessageCardElement) *larkcard.MessageCard {
	config := larkcard.NewMessageCardConfig().
		WideScreenMode(false).
		EnableForward(true).
		UpdateMulti(true).
		Build()
	var aElementPool []larkcard.MessageCardElement
	aElementPool = append(aElementPool, elements...)
	// 卡片消息体
	cardContent := larkcard.NewMessageCard().
		Config(config).
		Header(header).
		Elements(
			aElementPool,
		).Build()
	return cardContent
}

// BuildCardHeader 构建卡片标题
func BuildCardHeader(title string, color string) *larkcard.MessageCardHeader {
	return larkcard.NewMessageCardHeader().
		Title(larkcard.NewMessageCardPlainText().
			Content(title).
			Build()).
		Template(color).
		Build()
}

// BuildCardNote 构建卡片注释
func BuildCardNote(text string) larkcard.MessageCardElement {
	return larkcard.NewMessageCardNote().
		Elements([]larkcard.MessageCardNoteElement{larkcard.NewMessageCardPlainText().
			Content(text).
			Build()}).
		Build()
}

// BuildCardMainMd 构建卡片Md主内容
func BuildCardMainMd(text string) larkcard.MessageCardElement {
	return larkcard.NewMessageCardDiv().
		Fields([]*larkcard.MessageCardField{larkcard.NewMessageCardField().
			Text(larkcard.NewMessageCardLarkMd().
				Content(text).
				Build()).
			IsShort(true).
			Build()}).
		Build()
}

// BuildCardSplitLine 构建 md 分割线
func BuildCardSplitLine() larkcard.MessageCardElement {
	return larkcard.NewMessageCardHr().Build()
}

// BuildCardMdAndButton 构建 md 和按钮
func BuildCardMdAndButton(text string, btn *larkcard.MessageCardEmbedButton) larkcard.MessageCardElement {
	return larkcard.NewMessageCardDiv().
		Fields(
			[]*larkcard.MessageCardField{
				larkcard.NewMessageCardField().
					Text(larkcard.NewMessageCardLarkMd().
						Content(text).
						Build()).
					IsShort(true).
					Build()}).
		Extra(btn).
		Build()
}

// BuildEmbedButton 构建嵌套按钮
func BuildEmbedButton(content string, value map[string]interface{},
	typename larkcard.MessageCardButtonType) *larkcard.
	MessageCardEmbedButton {
	return larkcard.NewMessageCardEmbedButton().
		Text(larkcard.NewMessageCardPlainText().Content(content).Build()).
		Value(value).
		Type(typename).
		Build()
}
