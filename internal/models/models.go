package models

type SseMessage struct {
	EventId int64
	Message []byte
}

type BridgeMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}
