# RAG System

This is a Retrieval-Augmented Generation (RAG) question-answering system, developed with Go language for the backend and React for the frontend.

## Features

- Document upload and processing
- Vector storage and retrieval
- OpenAI-based intelligent Q&A
- Document management and reprocessing
- Multi-agent analysis system
- Knowledge extraction and note generation
- Automatic Anki card generation
- Structured document content processing

## Technology Stack

### Backend

- Go language
- Gin Web framework
- MongoDB (document metadata storage)
- Qdrant (vector database)
- OpenAI API (vector embedding and answer generation)
- LangChain multi-agent system

### Frontend

- React
- Axios (HTTP requests)
- React Markdown (formatted display)

## Installation and Usage

### Requirements

- Go 1.16+
- Node.js 14+
- MongoDB
- Qdrant

### Backend Setup

1. Navigate to the backend directory:
   ```
   cd rag-backend
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Create a `.env` file and configure:
   ```
   # Database configuration
   MONGODB_URI=mongodb://localhost:27017
   MONGODB_DATABASE=admin
   MONGODB_COLLECTION=RAGDATA

   # Qdrant configuration
   QDRANT_URL=http://localhost:6333
   QDRANT_HOST=localhost:6334
   QDRANT_COLLECTION=documents

   # OpenAI configuration
   OPENAI_API_KEY=your_openai_api_key_here
   OPENAI_MODEL=gpt-4o
   OPENAI_EMBEDDING_MODEL=text-embedding-ada-002

   # Server configuration
   SERVER_PORT=8080
   ```

4. Start the backend service:
   ```
   go run cmd/api/main.go cmd/api/route.go
   ```

### Frontend Setup

1. Navigate to the frontend directory:
   ```
   cd rag-frontend
   ```

2. Install dependencies:
   ```
   npm install
   ```

3. Start the frontend development server:
   ```
   npm start
   ```

## How to Use

1. Upload documents: Supports PDF and TXT formats
2. Wait for document processing to complete (status becomes "ready")
3. Enter your question in the query box
4. The system will generate answers based on the uploaded document content

## Advanced Features

### Basic Document Processing

- **Reprocess Document**: If you need to update a document's vector representation, use the "Reprocess" function
- **Clear Vector Database**: Remove all vector data while retaining document records
- **Clear All Documents**: Completely remove all documents and related data
- **Clean Invalid Records**: Fix invalid records in the database

### Multi-Agent System

The newly added multi-agent system provides more powerful document analysis capabilities:

- **Content Analysis Agent**: Analyzes document structure and chapters, extracting the document's hierarchical structure and organization
- **Knowledge Extraction Agent**: Extracts key concepts and technical definitions from documents, forming a knowledge base
- **Summary Agent**: Generates summaries for each chapter and section of the document, helping to quickly understand the content
- **Formatting Agent**: Integrates analysis results into structured Markdown notes and Anki cards

### Anki Card Generation

The system can automatically generate Anki learning cards based on document content:

- Generate question and answer cards based on key concepts
- Create specialized learning cards for technical architectures and diagrams
- Create dedicated cards for protocol and standard-related content
- Support export in Anki-importable format

### Document Content Enhancement

- Identify and mark technical diagrams and flowcharts in documents
- Add better structure and navigation to document content
- Optimize document presentation format to improve readability
- Add emojis and styling to enhance visual experience

## Usage Instructions

1. **RAG Q&A Mode**: Enter questions in the query box, and the system will retrieve relevant document fragments and generate answers
2. **Multi-Agent Processing**: Click the "Agent Processing" button in the document list to start multi-agent analysis
3. **Browse Notes**: After processing is complete, you can view the generated structured notes
4. **Export Cards**: View and export the generated Anki learning cards

## License

MIT 