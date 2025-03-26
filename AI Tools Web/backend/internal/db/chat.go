package db

import (
	"backend/internal/models"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatRepository 定义
type ChatRepository struct {
	collection *mongo.Collection
}

// NewChatRepository 创建新的 ChatRepository 实例
func NewChatRepository() *ChatRepository {
	return &ChatRepository{
		collection: GetCollection(ChatCollection),
	}
}

// CreateChat 创建新的聊天
func (r *ChatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {
	if chat.ID == "" {
		chat.ID = primitive.NewObjectID().Hex()
	}
	chat.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, chat)
	return err
}

// SaveMessage 保存消息
func (r *ChatRepository) SaveMessage(ctx context.Context, message *models.Message) error {
	// 确保消息有唯一ID
	if message.ID == "" {
		message.ID = primitive.NewObjectID().Hex()
	}

	// 确保消息有有效的时间戳
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	// 记录函数调用信息
	log.Printf("Saving message: ChatID=%s, Role=%s, ContentLength=%d",
		message.ChatID, message.Role, len(message.Content))

	collection := GetCollection(MessageCollection)
	_, err := collection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("Error inserting message: %v", err)
		return err
	}

	log.Printf("Message saved successfully: ChatID=%s, Role=%s", message.ChatID, message.Role)
	return nil
}

// GetChatHistory 获取聊天历史
func (r *ChatRepository) GetChatHistory(ctx context.Context) ([]models.Chat, error) {
	cursor, err := r.collection.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}

	var chats []models.Chat
	if err = cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

// GetMessages 获取消息
func (r *ChatRepository) GetMessages(ctx context.Context, chatID string) ([]models.Message, error) {
	messageCollection := GetCollection(MessageCollection)

	cursor, err := messageCollection.Find(ctx,
		bson.M{"chat_id": chatID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}

	var messages []models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// UpdateChatTitle 更新聊天标题
func (r *ChatRepository) UpdateChatTitle(ctx context.Context, chatID string, title string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": chatID},
		bson.M{"$set": bson.M{
			"title":      title,
			"updated_at": time.Now(),
		}},
	)

	if err != nil {
		log.Printf("Error updating chat title: %v", err)
		return err
	}

	if result.MatchedCount == 0 {
		log.Printf("No chat found with ID: %s", chatID)
		return fmt.Errorf("chat not found: %s", chatID)
	}

	return nil
}

// DeleteChat 删除聊天及其所有消息
func (r *ChatRepository) DeleteChat(ctx context.Context, chatID string) error {
	// 删除聊天记录
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": chatID})
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	// 删除该聊天的所有消息
	_, err = GetCollection(MessageCollection).DeleteMany(ctx, bson.M{"chat_id": chatID})
	if err != nil {
		return fmt.Errorf("failed to delete chat messages: %w", err)
	}

	return nil
}

// GetChat 获取单个聊天信息
func (r *ChatRepository) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	var chat models.Chat
	err := r.collection.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("chat not found: %s", chatID)
		}
		return nil, fmt.Errorf("error finding chat: %w", err)
	}
	return &chat, nil
}

// UpdateChatModel 更新聊天使用的模型
func (r *ChatRepository) UpdateChatModel(ctx context.Context, chatID string, model string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": chatID},
		bson.M{"$set": bson.M{
			"model":      model,
			"updated_at": time.Now(),
		}},
	)

	if err != nil {
		log.Printf("Error updating chat model: %v", err)
		return err
	}

	if result.MatchedCount == 0 {
		log.Printf("No chat found with ID: %s", chatID)
		return fmt.Errorf("chat not found: %s", chatID)
	}

	return nil
}
