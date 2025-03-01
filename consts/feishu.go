package consts

type MsgType string

const (
	MsgTypeText     MsgType = "text"
	MsgTypePost     MsgType = "post"
	MsgTypeImage    MsgType = "image"
	MsgTypeFile     MsgType = "file"
	MsgTypeShare    MsgType = "share"
	MsgTypeLocation MsgType = "location"
)

type ChatType string

const (
	GroupChatType = "group"
	UserChatType  = "personal"
	OtherChatType = "other"
)

type CardKind string

const (
	ClearCard CardKind = "clear"
	HelpCard  CardKind = "help"
)
