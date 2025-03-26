package auth

import (
	"github.com/gin-gonic/gin"
)

func ChatAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现聊天相关的认证逻辑
	}
}
