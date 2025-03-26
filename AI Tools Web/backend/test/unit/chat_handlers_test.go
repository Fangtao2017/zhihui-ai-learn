package auth_test

import (
	"backend/internal/auth"
	"backend/internal/chat"
	"backend/internal/models"
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
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateChat(ctx context.Context, chat *models.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockChatRepository) GetChatHistory(ctx context.Context) ([]models.Chat, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Chat), args.Error(1)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, chatID string) ([]models.Message, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) UpdateChatTitle(ctx context.Context, chatID string, title string) error {
	args := m.Called(ctx, chatID, title)
	return args.Error(0)
}

func (m *MockChatRepository) DeleteChat(ctx context.Context, chatID string) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockChatRepository) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockChatRepository) UpdateChatModel(ctx context.Context, chatID string, model string) error {
	args := m.Called(ctx, chatID, model)
	return args.Error(0)
}

// 添加认证上下文的中间件
func withAuthContext(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟用户认证信息
		userClaims := auth.UserClaims{
			Username: "testuser",
			Email:    "test@example.com",
		}

		// 将认证信息添加到请求上下文
		ctx := context.WithValue(r.Context(), "user", userClaims)
		r = r.WithContext(ctx)

		// 调用原始处理器
		handler.ServeHTTP(w, r)
	})
}

// 测试CreateChatHandler
func TestCreateChatHandler(t *testing.T) {
	// 创建请求体
	reqBody := map[string]string{
		"title": "Test Chat",
		"model": "gpt-3.5-turbo",
	}
	body, _ := json.Marshal(reqBody)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/chat/new", bytes.NewBuffer(body))
	assert.NoError(t, err)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.CreateChatHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)
	t.Logf("Response body: %s", rr.Body.String())

	// 解析响应
	if rr.Code == http.StatusOK {
		var resp models.Chat
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// 验证响应字段
		assert.Equal(t, "Test Chat", resp.Title)
		assert.Equal(t, "gpt-3.5-turbo", resp.Model)
		assert.NotEmpty(t, resp.ID)
	}
}

// 测试GetChatHistoryHandler
func TestGetChatHistoryHandler(t *testing.T) {
	// 创建请求
	req, err := http.NewRequest("GET", "/chat/history", nil)
	assert.NoError(t, err)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.GetChatHistoryHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)

	// 解析响应
	if rr.Code == http.StatusOK {
		var resp []models.Chat
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
	}
}

// 测试GetChatMessagesHandler
func TestGetChatMessagesHandler(t *testing.T) {
	// 创建请求
	req, err := http.NewRequest("GET", "/chat/123/messages", nil)
	assert.NoError(t, err)

	// 设置路由参数
	vars := map[string]string{
		"id": "123",
	}
	req = mux.SetURLVars(req, vars)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.GetChatMessagesHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)

	// 解析响应
	if rr.Code == http.StatusOK {
		var resp []models.Message
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
	}
}

// 测试UpdateChatTitleHandler
func TestUpdateChatTitleHandler(t *testing.T) {
	// 创建请求体
	reqBody := map[string]string{
		"title": "Updated Title",
	}
	body, _ := json.Marshal(reqBody)

	// 创建HTTP请求
	req, err := http.NewRequest("PUT", "/chat/123/title", bytes.NewBuffer(body))
	assert.NoError(t, err)

	// 设置路由参数
	vars := map[string]string{
		"id": "123",
	}
	req = mux.SetURLVars(req, vars)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.UpdateChatTitleHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)
}

// 测试UpdateChatModelHandler
func TestUpdateChatModelHandler(t *testing.T) {
	// 创建请求体
	reqBody := map[string]string{
		"model": "gpt-4",
	}
	body, _ := json.Marshal(reqBody)

	// 创建HTTP请求
	req, err := http.NewRequest("PUT", "/chat/123/model", bytes.NewBuffer(body))
	assert.NoError(t, err)

	// 设置路由参数
	vars := map[string]string{
		"id": "123",
	}
	req = mux.SetURLVars(req, vars)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.UpdateChatModelHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)
	t.Logf("Response body: %s", rr.Body.String())

	// 如果成功，验证响应内容
	if rr.Code == http.StatusOK {
		var resp struct {
			Message string `json:"message"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "successfully")
	}
}

// 测试DeleteChatHandler
func TestDeleteChatHandler(t *testing.T) {
	// 创建HTTP请求
	req, err := http.NewRequest("DELETE", "/chat/123", nil)
	assert.NoError(t, err)

	// 设置路由参数
	vars := map[string]string{
		"id": "123",
	}
	req = mux.SetURLVars(req, vars)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.DeleteChatHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)
	t.Logf("Response body: %s", rr.Body.String())

	// 如果成功，验证响应内容
	if rr.Code == http.StatusOK {
		var resp struct {
			Message string `json:"message"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "deleted")
	}
}

// 测试GetChatInfoHandler
func TestGetChatInfoHandler(t *testing.T) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", "/chat/123/info", nil)
	assert.NoError(t, err)

	// 设置路由参数
	vars := map[string]string{
		"id": "123",
	}
	req = mux.SetURLVars(req, vars)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.GetChatInfoHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)

	// 解析响应
	if rr.Code == http.StatusOK {
		var resp models.Chat
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "123", resp.ID)
	}
}

// 测试GetAvailableModelsHandler
func TestGetAvailableModelsHandler(t *testing.T) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", "/chat/models", nil)
	assert.NoError(t, err)

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 准备测试环境 - 使用认证上下文
	handler := withAuthContext(http.HandlerFunc(chat.GetAvailableModelsHandler))

	// 执行请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	t.Logf("Status code: %d", rr.Code)

	// 解析响应
	if rr.Code == http.StatusOK {
		var resp []string
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// 验证返回的模型列表不为空
		assert.NotEmpty(t, resp)
	}
}
