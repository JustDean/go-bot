package poller

type TgUpdateResponseBase struct {
	Ok      bool `json:"ok"`
	Updates []struct {
		UpdateId uint `json:"update_id"`
	} `json:"result"`
}

type TgUpdateResponse struct {
	Ok      bool       `json:"ok"`
	Updates []TgUpdate `json:"result"`
}

type TgUpdate struct {
	UpdateId uint      `json:"update_id"`
	Message  TgMessage `json:"message,omitempty"`
}

type TgMessage struct {
	Id   uint32 `json:"message_id"`
	User TgUser `json:"from,omitempty"`
	Chat TgChat `json:"sender_chat,omitempty"`
	Date uint   `json:"date"`
	Text string `json:"text,omitempty"`
}

type TgUser struct {
	Id        uint32 `json:"id"`
	IsBot     bool   `json:"is_bot"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
}

type TgChat struct {
	Id        uint32 `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
