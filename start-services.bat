@echo off
echo 启动AI Tools Web和RAGSystem服务...

:: 启动RAGSystem后端服务
start cmd /k "cd RAGSystem\rag-backend && go run cmd/api/main.go"

:: 等待2秒，确保RAGSystem后端服务已启动
timeout /t 2 /nobreak > nul

:: 启动AI Tools Web后端服务
start cmd /k "cd AI Tools Web\backend && go run cmd/api/main.go"

:: 等待2秒，确保AI Tools Web后端服务已启动
timeout /t 2 /nobreak > nul

:: 启动AI Tools Web前端服务
start cmd /k "cd AI Tools Web\frontend && npm start"

echo 所有服务已启动！
echo RAGSystem后端: http://localhost:8081
echo AI Tools Web后端: http://localhost:8080
echo AI Tools Web前端: http://localhost:3000 