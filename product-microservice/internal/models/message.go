package models

import "time"

// ErrorMessage
type ErrorMessage struct {
	MessageID string    `json:"message_id"`
	Offset    int64     `json:"offset"`
	Topic     string    `json:"topic"`
	Partition int       `json:"partition"`
	Error     string    `json:"error"`
	Time      time.Time `json:"time"`
}
