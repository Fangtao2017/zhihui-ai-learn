# ZHIHUI AI LEARN Platform

An integrated learning platform with advanced AI technologies, providing intelligent conversation, knowledge base management, and interactive learning capabilities.

## Project Overview

ZHIHUI AI LEARN is an intelligent learning tool based on advanced large language model technologies, designed to help users improve learning efficiency through AI conversations and knowledge base management. The system consists of two core components: **AI Conversation Tool** and **RAG Knowledge Base System**, both adopting a front-end and back-end separation design to provide a smooth user experience and powerful functional support.

## System Architecture

### Project Structure

```
ZHIHUI-AI-LEARN/
├── AI Tools Web/           # AI Conversation Tool
│   ├── frontend/          # React-based front-end
│   └── backend/           # Go-implemented back-end
├── RAGSystem/             # RAG Knowledge Base System
│   ├── rag-frontend/      # React-based front-end
│   └── rag-backend/       # Go-implemented back-end
├── docker-compose.yml     # Docker container orchestration
├── start-services.sh      # Linux/Mac startup script
└── start-services.bat     # Windows startup script
```

### Technology Stack

#### Frontend Technologies

- **Framework**: React 18
- **UI Component Library**: Ant Design 5.x
- **State Management**: React Context API
- **HTTP Client**: Axios
- **WebSocket**: For streaming responses

#### Backend Technologies

- **Language**: Go 1.20+
- **Web Frameworks**: 
  - AI Tools Web: Gorilla Mux
  - RAGSystem: Gin
- **Databases**: 
  - MongoDB: User and chat data storage
  - Qdrant/Milvus: Vector database for RAG system
- **Authentication**: JWT (JSON Web Tokens)
- **AI API Integration**: 
  - OpenAI API (GPT-4o, GPT-4, GPT-3.5 series)
  - Anthropic API (Claude 3.5 Sonnet, Claude 3 Opus)

## Core Features

### AI Conversation Tool

- **Multi-model Support**: Integration of various latest models from OpenAI and Anthropic
- **Session Management**: Create, save, restore, and delete conversations
- **Streaming Response**: Real-time display of AI replies for a more natural conversation experience
- **Language Adaptation**: Automatic detection and adaptation to Chinese and English environments
- **User Authentication**: Complete registration, login, and password management functions
- **Personalized Settings**: Allow users to adjust model parameters and interface preferences

### RAG Knowledge Base System

- **Document Management**: Upload, process, query, and delete documents
- **Multi-format Support**: Support for PDF, Word, Text, and other formats
- **Vector Storage**: Using advanced vector databases for semantic retrieval
- **Intelligent Query**: Extract relevant document fragments based on user questions and generate answers
- **Multi-Agent Processing**: Use multiple agents to collaboratively extract and process information for complex documents
- **Status Tracking**: Real-time monitoring of document processing progress

## Installation and Configuration

### Prerequisites

- Go 1.20+
- Node.js 16+
- MongoDB 5.0+
- Qdrant/Milvus (Docker deployment)
- OpenAI API key and/or Anthropic API key

### Manual Installation

#### 1. AI Conversation Tool Frontend

```bash
cd AI\ Tools\ Web/frontend
npm install
npm start
```

#### 2. AI Conversation Tool Backend

```bash
cd AI\ Tools\ Web/backend
go mod tidy
go run cmd/api/main.go
```

#### 3. RAG System Frontend

```bash
cd RAGSystem/rag-frontend
npm install
npm start
```

#### 4. RAG System Backend

```bash
cd RAGSystem/rag-backend
go mod tidy
go run cmd/api/main.go
```

### Docker Deployment

The project provides complete Docker configuration for one-click startup of all services:

```bash
docker-compose up -d
```

### Environment Variable Configuration

#### AI Conversation Tool Backend (.env)

```
# JWT Settings
JWT_SECRET_KEY=your_jwt_secret_key_here

# MongoDB Settings
MONGODB_URI=mongodb://localhost:27017
DB_NAME=admin

# Server Settings
PORT=8080

# OpenAI API Configuration
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-4o
OPENAI_BASE_URL=https://api.openai.com

# Anthropic API Configuration
ANTHROPIC_API_KEY=your_anthropic_api_key_here
ANTHROPIC_BASE_URL=https://api.anthropic.com

# RAG Service Configuration
RAG_SERVICE_URL=http://localhost:8081
```

#### RAG System Backend (.env)

```
# Server Settings
SERVER_PORT=8081

# MongoDB Settings
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=rag_system

# Qdrant Settings
QDRANT_URL=http://localhost:6333
COLLECTION_NAME=documents

# OpenAI API Configuration
OPENAI_API_KEY=your_openai_api_key_here
EMBEDDING_MODEL=text-embedding-3-large
```

## User Guide

### System Startup

The project provides convenient startup scripts to launch all services with one click:

- **Linux/Mac**: `./start-services.sh`
- **Windows**: `start-services.bat`

After startup, you can access the various services at the following addresses:

- AI Tools Web Frontend: http://localhost:3000
- AI Tools Web Backend: http://localhost:8080
- RAG System Backend: http://localhost:8081

### Basic Usage Flow

1. Register/login on the homepage
2. Visit the chat page, select an AI model to start a conversation
3. Upload documents to the knowledge base
4. Ask questions and learn based on your personal knowledge base

## Developer Guide

### API Interfaces

#### AI Conversation Tool API

- User Authentication:
  - `POST /api/register`: Register a new user
  - `POST /api/login`: User login
  - `POST /api/user/change-password`: Change password

- Chat Functionality:
  - `GET /api/chat/history`: Get chat history
  - `POST /api/chat/new`: Create a new chat
  - `GET /api/chat/{id}/messages`: Get chat messages
  - `POST /api/chat/{id}/messages`: Send a message
  - `GET|POST /api/chat/{id}/messages/stream`: Stream send/receive messages
  - `GET /api/chat/models`: Get list of available models

#### RAG System API

- Document Management:
  - `POST /upload`: Upload document
  - `GET /documents`: Get document list
  - `DELETE /document/:doc_id`: Delete document
  - `POST /document/:doc_id/reprocess`: Reprocess document
  - `POST /document/:doc_id/multi-agent`: Process document with multiple agents

- Query Functionality:
  - `POST /query`: Query knowledge base
  - `GET /status/:task_id`: Get processing status
  - `POST /clear-vectors`: Clear vector database

### Project Extension

- **Adding New Models**: Implement new service interfaces in the `internal/services` directory
- **Extending RAG Functionality**: Add new processing methods in `rag-backend/internal/app`
- **Optimizing User Interface**: Modify frontend components to enhance user experience

## License

© 2024 ZHIHUI AI LEARN Platform. All rights reserved.

---

*Note: Using this system requires valid OpenAI API keys and/or Anthropic API keys.*
