#!/bin/bash
echo "启动AI Tools Web和RAGSystem服务..."

# 启动RAGSystem后端服务
cd RAGSystem/rag-backend && go run cmd/api/main.go &
RAG_PID=$!

# 等待2秒，确保RAGSystem后端服务已启动
sleep 2

# 启动AI Tools Web后端服务
cd ../../AI\ Tools\ Web/backend && go run cmd/api/main.go &
BACKEND_PID=$!

# 等待2秒，确保AI Tools Web后端服务已启动
sleep 2

# 启动AI Tools Web前端服务
cd ../frontend && npm start &
FRONTEND_PID=$!

echo "所有服务已启动！"
echo "RAGSystem后端: http://localhost:8081"
echo "AI Tools Web后端: http://localhost:8080"
echo "AI Tools Web前端: http://localhost:3000"

# 等待用户按Ctrl+C
trap "kill $RAG_PID $BACKEND_PID $FRONTEND_PID; exit" INT
wait 