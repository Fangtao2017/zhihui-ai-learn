package models

import "time"

type Chat struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Title     string    `json:"title" bson:"title"`
	Model     string    `json:"model" bson:"model"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Message struct {
	ID        string    `json:"id" bson:"_id"`
	ChatID    string    `json:"chat_id" bson:"chat_id"`
	Role      string    `json:"role" bson:"role"`
	Content   string    `json:"content" bson:"content"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type ChatResponse struct {
	Chat     Chat      `json:"chat"`
	Messages []Message `json:"messages"`
}
