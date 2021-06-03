package output

type TextMsg struct {
	Msgtype string `json:"msgtype"`
	At      At     `json:"at"`
	Text    Text   `json:"text"`
}

type At struct {
	AtMobiles []string `json:"atMobiles" config:"at_mobiles"`
	AtUserIds []string `json:"atUserIds" config:"at_user_ids"`
	IsAtAll   bool     `json:"isAtAll" config:"is_at_all"`
}

type Text struct {
	Content interface{} `json:"content"`
}

var (
	defaultTextMsg = TextMsg{
		Msgtype: TextMsgType,
	}
)
