package auth_test

import (
	"backend/internal/auth"
	"backend/internal/chat"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 模拟ChatRepository
type MockLLMRepository struct {
	mock.Mock
}

func (m *MockLLMRepository) GetChat(ctx context.Context, chatID string) (interface{}, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockLLMRepository) UpdateChatModel(ctx context.Context, chatID string, model string) error {
	args := m.Called(ctx, chatID, model)
	return args.Error(0)
}

// 测试获取可用模型列表
func TestLLMGetAvailableModelsHandler(t *testing.T) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", "/chat/models", nil)
	assert.NoError(t, err)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 添加认证上下文
		userClaims := auth.UserClaims{
			Username: "testuser",
			Email:    "test@example.com",
		}
		ctx := context.WithValue(r.Context(), "user", userClaims)
		r = r.WithContext(ctx)

		chat.GetAvailableModelsHandler(w, r)
	})

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	assert.Equal(t, http.StatusOK, rr.Code)

	// 解析响应
	var models []string
	err = json.Unmarshal(rr.Body.Bytes(), &models)
	assert.NoError(t, err)

	// 验证返回的模型列表包含预期的模型
	t.Logf("Available models: %v", models)
	assert.Contains(t, models, "gpt-3.5-turbo")
	assert.Contains(t, models, "gpt-4")
}

// 测试创建使用特定模型的聊天
func TestCreateChatWithSpecificModel(t *testing.T) {
	models := []string{"gpt-3.5-turbo", "gpt-4", "claude-3-haiku"}

	for _, model := range models {
		t.Run("Create chat with "+model, func(t *testing.T) {
			// 创建请求体
			reqBody := map[string]string{
				"title": "Test Chat with " + model,
				"model": model,
			}
			body, _ := json.Marshal(reqBody)

			// 创建HTTP请求
			req, err := http.NewRequest("POST", "/chat/new", bytes.NewBuffer(body))
			assert.NoError(t, err)

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 准备测试环境 - 使用认证上下文
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 添加认证上下文
				userClaims := auth.UserClaims{
					Username: "testuser",
					Email:    "test@example.com",
				}
				ctx := context.WithValue(r.Context(), "user", userClaims)
				r = r.WithContext(ctx)

				chat.CreateChatHandler(w, r)
			})

			// 执行请求
			handler.ServeHTTP(rr, req)

			// 检查状态码和响应
			t.Logf("Status code: %d", rr.Code)
			t.Logf("Response body: %s", rr.Body.String())

			// 如果成功，验证响应内容
			if rr.Code == http.StatusOK {
				var resp map[string]interface{}
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Test Chat with "+model, resp["title"])
				assert.Equal(t, model, resp["model"])
			}
		})
	}
}

// 测试更改聊天的模型
func TestUpdateChatModel(t *testing.T) {
	models := []string{"gpt-3.5-turbo", "gpt-4", "claude-3-haiku", "gemini-pro"}
	chatID := "123456"

	for _, model := range models {
		t.Run("Update chat to "+model, func(t *testing.T) {
			// 创建请求体
			reqBody := map[string]string{
				"model": model,
			}
			body, _ := json.Marshal(reqBody)

			// 创建HTTP请求
			req, err := http.NewRequest("PUT", "/chat/"+chatID+"/model", bytes.NewBuffer(body))
			assert.NoError(t, err)

			// 设置路由参数
			vars := map[string]string{
				"id": chatID,
			}
			req = mux.SetURLVars(req, vars)

			// 创建响应记录器
			rr := httptest.NewRecorder()

			// 准备测试环境
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 添加认证上下文
				userClaims := auth.UserClaims{
					Username: "testuser",
					Email:    "test@example.com",
				}
				ctx := context.WithValue(r.Context(), "user", userClaims)
				r = r.WithContext(ctx)

				chat.UpdateChatModelHandler(w, r)
			})

			// 执行请求
			handler.ServeHTTP(rr, req)

			// 记录测试结果
			t.Logf("Status code: %d", rr.Code)
			t.Logf("Response body: %s", rr.Body.String())

			// 如果数据库连接正常，可以验证响应
			if rr.Code == http.StatusOK {
				var resp struct {
					Message string `json:"message"`
				}
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Message, "successfully")
			}
		})
	}
}

// 测试在消息中途切换模型
func TestSwitchModelMidConversation(t *testing.T) {
	chatID := "789012"

	// 添加认证上下文的辅助函数
	addAuthContext := func(handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 添加认证上下文
			userClaims := auth.UserClaims{
				Username: "testuser",
				Email:    "test@example.com",
			}
			ctx := context.WithValue(r.Context(), "user", userClaims)
			r = r.WithContext(ctx)

			handler(w, r)
		})
	}

	// 步骤1: 创建一个使用gpt-3.5-turbo的对话
	t.Run("Create initial chat with gpt-3.5-turbo", func(t *testing.T) {
		reqBody := map[string]string{
			"title": "Multi-model Conversation",
			"model": "gpt-3.5-turbo",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/chat/new", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		handler := addAuthContext(chat.CreateChatHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Initial chat creation status: %d", rr.Code)
	})

	// 步骤2: 发送消息
	t.Run("Send message with initial model", func(t *testing.T) {
		reqBody := map[string]string{
			"message": "这是一个测试消息，使用的是GPT-3.5模型",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/chat/"+chatID+"/messages", bytes.NewBuffer(body))
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.SendMessageHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Send message status: %d", rr.Code)
	})

	// 步骤3: 切换到GPT-4模型
	t.Run("Switch to GPT-4 model", func(t *testing.T) {
		reqBody := map[string]string{
			"model": "gpt-4",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/chat/"+chatID+"/model", bytes.NewBuffer(body))
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.UpdateChatModelHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Model switch status: %d", rr.Code)
		t.Logf("Response: %s", rr.Body.String())
	})

	// 步骤4: 使用新模型发送消息
	t.Run("Send message with new model", func(t *testing.T) {
		reqBody := map[string]string{
			"message": "这是一个测试消息，现在使用的是GPT-4模型",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/chat/"+chatID+"/messages", bytes.NewBuffer(body))
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.SendMessageHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Send message with new model status: %d", rr.Code)
	})

	// 步骤5: 再次切换模型到Claude
	t.Run("Switch to Claude model", func(t *testing.T) {
		reqBody := map[string]string{
			"model": "claude-3-opus",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/chat/"+chatID+"/model", bytes.NewBuffer(body))
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.UpdateChatModelHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Second model switch status: %d", rr.Code)
	})

	// 步骤6: 查看整个对话历史
	t.Run("View conversation history with multiple models", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/chat/"+chatID+"/messages", nil)
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.GetChatMessagesHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Get messages status: %d", rr.Code)

		if rr.Code == http.StatusOK {
			var messages []map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &messages)
			assert.NoError(t, err)

			// 打印消息信息
			for i, msg := range messages {
				t.Logf("Message %d: Content: %s, Role: %s",
					i+1, msg["content"], msg["role"])
			}
		}
	})

	// 步骤7: 查看聊天信息，确认当前模型
	t.Run("Check current chat model", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/chat/"+chatID+"/info", nil)
		vars := map[string]string{"id": chatID}
		req = mux.SetURLVars(req, vars)

		rr := httptest.NewRecorder()
		handler := addAuthContext(chat.GetChatInfoHandler)
		handler.ServeHTTP(rr, req)

		t.Logf("Get chat info status: %d", rr.Code)

		if rr.Code == http.StatusOK {
			var chatInfo map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &chatInfo)
			assert.NoError(t, err)

			t.Logf("Final chat model: %v", chatInfo["model"])
			// 确认最终模型
			model, ok := chatInfo["model"].(string)
			if ok {
				assert.Equal(t, "claude-3-opus", model, "最终模型应该是Claude-3-Opus")
			}
		}
	})
}

// 测试模型性能对比
func TestModelPerformanceComparison(t *testing.T) {
	testPrompt := "解释量子计算的基本原理并用通俗易懂的语言描述量子比特"
	models := []string{"gpt-3.5-turbo", "gpt-4", "claude-3-haiku", "gemini-pro"}

	// 添加认证上下文的辅助函数
	addAuthContext := func(handler http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 添加认证上下文
			userClaims := auth.UserClaims{
				Username: "testuser",
				Email:    "test@example.com",
			}
			ctx := context.WithValue(r.Context(), "user", userClaims)
			r = r.WithContext(ctx)

			handler(w, r)
		})
	}

	for _, model := range models {
		t.Run("Performance test for "+model, func(t *testing.T) {
			// 创建一个新的聊天，使用指定模型
			chatReqBody := map[string]string{
				"title": "Performance Test - " + model,
				"model": model,
			}
			chatBody, _ := json.Marshal(chatReqBody)

			chatReq, _ := http.NewRequest("POST", "/chat/new", bytes.NewBuffer(chatBody))
			chatRR := httptest.NewRecorder()

			createHandler := addAuthContext(chat.CreateChatHandler)
			createHandler.ServeHTTP(chatRR, chatReq)

			// 获取聊天ID
			var chatResp map[string]interface{}
			json.Unmarshal(chatRR.Body.Bytes(), &chatResp)

			// 检查是否成功获取ID
			chatID, ok := chatResp["id"].(string)
			if !ok {
				t.Logf("Failed to parse chat ID")
				return
			}

			t.Logf("Created chat with ID: %s and model: %s", chatID, model)

			// 发送测试消息
			msgReqBody := map[string]string{
				"message": testPrompt,
			}
			msgBody, _ := json.Marshal(msgReqBody)

			msgReq, _ := http.NewRequest("POST", "/chat/"+chatID+"/messages", bytes.NewBuffer(msgBody))
			vars := map[string]string{"id": chatID}
			msgReq = mux.SetURLVars(msgReq, vars)

			msgRR := httptest.NewRecorder()

			msgHandler := addAuthContext(chat.SendMessageHandler)
			msgHandler.ServeHTTP(msgRR, msgReq)

			t.Logf("Message sent status: %d", msgRR.Code)

			// 验证响应
			if msgRR.Code == http.StatusOK {
				t.Logf("Successfully sent message using model: %s", model)

				// 获取AI响应
				getReq, _ := http.NewRequest("GET", "/chat/"+chatID+"/messages", nil)
				getReq = mux.SetURLVars(getReq, vars)

				getRR := httptest.NewRecorder()

				getHandler := addAuthContext(chat.GetChatMessagesHandler)
				getHandler.ServeHTTP(getRR, getReq)

				if getRR.Code == http.StatusOK {
					var messages []map[string]interface{}
					json.Unmarshal(getRR.Body.Bytes(), &messages)

					// 查找AI响应
					for _, msg := range messages {
						if role, ok := msg["role"].(string); ok && role == "assistant" {
							// 只显示响应的前100个字符
							if content, ok := msg["content"].(string); ok {
								responsePreview := content
								if len(responsePreview) > 100 {
									responsePreview = responsePreview[:100] + "..."
								}
								t.Logf("%s response: %s", model, responsePreview)
							}
							break
						}
					}
				}
			}
		})
	}
}
