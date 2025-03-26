# ZHIHUI AI LEARN - Intelligent AI Learning Platform

An integrated AI learning platform providing intelligent conversation, knowledge base management, and other features.

## Project Architecture

### Frontend
- React framework
- Ant Design component library
- Streaming response chat interface
- Support for multiple AI models (GPT-4o, Claude 3.5, etc.)
- Knowledge base management and query support

### Backend
- RESTful API developed in Go
- MongoDB data storage
- JWT authentication system
- Integration with OpenAI and Anthropic LLM APIs
- RAG system for knowledge base queries

## Key Features

- **AI Chat**: Support for multiple large language models, including GPT-4o and Claude series
- **Knowledge Base Management**: Upload and manage documents, enabling AI to answer questions based on your knowledge
- **Language Preference**: Support for automatic detection and switching between Chinese and English
- **User Management**: Complete user authentication and authorization system

## Installation Guide

### Frontend Setup

```bash
cd AI\ Tools\ Web/frontend
npm install
npm start
```

### Backend Setup

```bash
cd AI\ Tools\ Web/backend
go mod tidy
go run cmd/api/main.go
```

### Environment Variables Configuration

Create a `.env` file in the backend directory with the following configuration:

```
# JWT settings
JWT_SECRET_KEY=your_jwt_secret

# MongoDB settings
MONGODB_URI=mongodb://localhost:27017
DB_NAME=admin

# Server settings
PORT=8080

# OpenAI API configuration
OPENAI_API_KEY=your_openai_api_key
OPENAI_MODEL=gpt-4o
OPENAI_BASE_URL=https://api.openai.com

# Anthropic API configuration (optional)
ANTHROPIC_API_KEY=your_anthropic_api_key
ANTHROPIC_BASE_URL=https://api.anthropic.com

# RAG service configuration
RAG_SERVICE_URL=http://localhost:8081
```

## Docker Support

The project includes a docker-compose.yml file for containerized deployment:

```bash
docker-compose up -d
```

## License

© 2024 ZHIHUI AI LEARN. All rights reserved.
