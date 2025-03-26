package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// LangChainMultiAgentSystem manages multiple Agents in the system
type LangChainMultiAgentSystem struct {
	agents   []Agent
	results  map[string]interface{}
	filePath string
}

// NewLangChainMultiAgentSystem creates a new multi-Agent system
func NewLangChainMultiAgentSystem(filePath string) *LangChainMultiAgentSystem {
	// Initialize LangChain multi-Agent system
	system := &LangChainMultiAgentSystem{
		results:  make(map[string]interface{}),
		filePath: filePath,
	}

	// Initialize each Agent
	system.agents = []Agent{
		&ContentAnalysisAgent{},
		&KnowledgeExtractionAgent{},
		&SummaryAgent{},
		&FormatAgent{},
	}

	return system
}

// Agent defines the interface that an Agent needs to implement
type Agent interface {
	Execute(input map[string]interface{}) (map[string]interface{}, error)
}

// ContentAnalysisAgent responsible for analyzing document structure
type ContentAnalysisAgent struct{}

func (a *ContentAnalysisAgent) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("[LangChain] Execute Content Analysis Agent...")

	// Get input content
	content, ok := input["content"].(string)
	if !ok || content == "" {
		return nil, fmt.Errorf("Content is empty")
	}

	// Call the actual content analysis function
	result, err := runLangChainContentAnalysisAgent(content)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// KnowledgeExtractionAgent responsible for extracting key knowledge
type KnowledgeExtractionAgent struct{}

func (a *KnowledgeExtractionAgent) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("[LangChain] Execute Knowledge Extraction Agent...")

	// Get chapter content
	chapters, ok := input["chapters"].([]map[string]interface{})
	if !ok || len(chapters) == 0 {
		return nil, fmt.Errorf("No chapter content available for processing")
	}

	// Actual implementation: Call GPT API to extract key concepts and definitions
	result, err := runLangChainKnowledgeExtractionAgent(chapters)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SummaryAgent responsible for generating summaries
type SummaryAgent struct{}

func (a *SummaryAgent) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("[LangChain] Execute Summary Generation Agent...")

	// Get chapter content and knowledge points
	chapters, ok := input["chapters"].([]map[string]interface{})
	if !ok || len(chapters) == 0 {
		return nil, fmt.Errorf("No chapter content available for summarization")
	}

	// Actual implementation: Call GPT API to generate summaries for each chapter
	result, err := runLangChainSummaryAgent(chapters)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FormatAgent responsible for formatting output
type FormatAgent struct{}

func (a *FormatAgent) Execute(input map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("[LangChain] Execute Formatting Agent...")

	// Get chapter content, knowledge points and summaries
	chapters, ok := input["chapters"].([]map[string]interface{})
	if !ok || len(chapters) == 0 {
		return nil, fmt.Errorf("No chapter content available for formatting")
	}

	concepts, ok := input["key_concepts"].([]map[string]interface{})
	if !ok {
		concepts = []map[string]interface{}{}
	}

	summaries, ok := input["summaries"].([]map[string]interface{})
	if !ok {
		summaries = []map[string]interface{}{}
	}

	// Actually call formatting function
	result, err := runLangChainFormatAgent(chapters, concepts, summaries)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GenerateNotesFromPDF processes a PDF document using the LangChain multi-Agent system
func (s *LangChainMultiAgentSystem) GenerateNotesFromPDF() (map[string]interface{}, error) {
	startTime := time.Now()
	fmt.Printf("[LangChain] Start processing document: %s\n", s.filePath)

	// 1. Extract PDF text
	fmt.Println("==================================")
	fmt.Println("📄 Step 1: Start extracting PDF text...")
	text, err := extractLangChainTextFromPDF(s.filePath)
	if err != nil {
		fmt.Printf("❌ PDF text extraction failed: %v\n", err)
		return nil, fmt.Errorf("Text extraction failed: %v", err)
	}
	textLen := len(text)
	fmt.Printf("✅ PDF text extraction succeeded! Extracted %d characters.\n", textLen)
	fmt.Println("==================================")

	// Set initial context
	fmt.Println("🔄 Preparing to initialize processing context...")
	context := map[string]interface{}{
		"content":   text,
		"file_path": s.filePath,
	}
	fmt.Println("✅ Processing context initialized")

	// 2. Execute each Agent in sequence
	for i, agent := range s.agents {
		agentName := fmt.Sprintf("Agent%d", i+1)
		switch agent.(type) {
		case *ContentAnalysisAgent:
			agentName = "Content Analysis Agent"
		case *KnowledgeExtractionAgent:
			agentName = "Knowledge Extraction Agent"
		case *SummaryAgent:
			agentName = "Summary Generation Agent"
		case *FormatAgent:
			agentName = "Formatting Agent"
		}

		fmt.Println("==================================")
		fmt.Printf("🤖 Step %d: Starting to execute %s...\n", i+2, agentName)

		result, err := agent.Execute(context)
		if err != nil {
			fmt.Printf("❌ %s execution failed: %v\n", agentName, err)
			return nil, fmt.Errorf("Agent %d (%s) execution failed: %v", i+1, agentName, err)
		}

		// Print result summary
		resultSummary := ""
		for k, v := range result {
			switch val := v.(type) {
			case string:
				resultSummary += fmt.Sprintf("\n  - %s: %s...(length:%d)", k, truncateString(val, 50), len(val))
			case []map[string]interface{}:
				resultSummary += fmt.Sprintf("\n  - %s: [%d items]", k, len(val))
			default:
				resultSummary += fmt.Sprintf("\n  - %s: %v", k, v)
			}
		}
		fmt.Printf("✅ %s execution succeeded!%s\n", agentName, resultSummary)

		// Update context, add new result
		for k, v := range result {
			context[k] = v
			s.results[k] = v
		}
		fmt.Println("==================================")
	}

	// 3. Add processing metadata
	processingTime := time.Since(startTime).Seconds()
	s.results["processing_time"] = processingTime
	s.results["timestamp"] = time.Now().Format(time.RFC3339)
	s.results["status"] = "success"

	fmt.Println("==================================")
	fmt.Printf("✅ Document processing completed! Total time: %.2f seconds\n", processingTime)
	fmt.Println("==================================")

	return s.results, nil
}

// Use Python script to call PyMuPDF to extract PDF text
func extractLangChainTextFromPDF(pdfPath string) (string, error) {
	fmt.Printf("🔍 Starting to extract PDF file: %s\n", pdfPath)

	// Check if the file exists
	fileInfo, err := os.Stat(pdfPath)
	if os.IsNotExist(err) {
		// Try to find all files in the uploads directory
		uploadsDir := filepath.Dir(pdfPath)
		files, _ := os.ReadDir(uploadsDir)
		fileList := "Files in the directory:\n"
		for _, file := range files {
			fileList += fmt.Sprintf("- %s\n", file.Name())
		}
		return "", fmt.Errorf("PDF file does not exist: %s\n%s", pdfPath, fileList)
	} else if err != nil {
		return "", fmt.Errorf("Failed to check file status: %v", err)
	}

	fmt.Printf("✅ File exists, size: %d bytes, continue processing\n", fileInfo.Size())

	// Try to use alternative method to extract text - if pdftotext tool exists, use it
	pdfToTextPath, _ := exec.LookPath("pdftotext")
	if pdfToTextPath != "" {
		fmt.Println("🛠️ Detected pdftotext tool, trying to use it to extract text")
		cmd := exec.Command(pdfToTextPath, "-layout", pdfPath, "-")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			result := out.String()
			if len(result) > 0 {
				fmt.Printf("✅ Successfully extracted %d characters using pdftotext\n", len(result))
				return result, nil
			}
			fmt.Println("⚠️ pdftotext did not extract text, will try Python method")
		} else {
			fmt.Printf("⚠️ pdftotext execution failed: %v, error: %s\n", err, stderr.String())
		}
	}

	// If PDFTOTEXT fails or does not exist, try using Python method
	fmt.Println("📝 Preparing to use Python method to extract text")

	// Check if Python is available
	pythonPath, err := exec.LookPath("python")
	if err != nil {
		// Try to find python3
		pythonPath, err = exec.LookPath("python3")
		if err != nil {
			return "", fmt.Errorf("Python interpreter not found, cannot extract PDF text")
		}
	}
	fmt.Printf("✅ Found Python interpreter: %s\n", pythonPath)

	// Create temporary Python script
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, "extract_pdf.py")
	fmt.Printf("📝 Creating temporary Python script: %s\n", scriptPath)

	// Check if PyMuPDF is installed
	checkPyMuPDFCmd := exec.Command(pythonPath, "-c", "import fitz; print('PyMuPDF is installed')")
	var checkOutput bytes.Buffer
	checkPyMuPDFCmd.Stdout = &checkOutput
	checkPyMuPDFCmd.Stderr = &checkOutput

	if err := checkPyMuPDFCmd.Run(); err != nil {
		pymuPDFInstallInfo := "Try installing PyMuPDF: pip install PyMuPDF"
		return "", fmt.Errorf("PyMuPDF is not installed, cannot extract PDF text\n%s\nDetection result: %s",
			pymuPDFInstallInfo, checkOutput.String())
	}
	fmt.Println("✅ PyMuPDF is installed")

	// PyMuPDF script content
	scriptContent := `
import sys
import fitz  # PyMuPDF

def extract_text_from_pdf(pdf_path):
    text = ""
    try:
        print(f"Opening PDF: {pdf_path}", file=sys.stderr)
        doc = fitz.open(pdf_path)
        print(f"PDF opened successfully with {len(doc)} pages", file=sys.stderr)
        for page_num in range(len(doc)):
            page = doc.load_page(page_num)
            text += page.get_text()
            if page_num % 10 == 0 and page_num > 0:
                print(f"Processed {page_num} pages...", file=sys.stderr)
        print(f"Extracted {len(text)} characters", file=sys.stderr)
        return text
    except Exception as e:
        print(f"Error extracting text: {str(e)}", file=sys.stderr)
        return ""

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python extract_pdf.py <pdf_path>", file=sys.stderr)
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    print(f"Starting extraction from: {pdf_path}", file=sys.stderr)
    text = extract_text_from_pdf(pdf_path)
    print(text)
`

	// Write temporary script file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		return "", fmt.Errorf("Failed to create temporary Python script: %v", err)
	}
	fmt.Println("✅ Python script created successfully")

	// Execute Python script
	fmt.Println("🚀 Starting to execute Python script to extract text...")
	cmd := exec.Command(pythonPath, scriptPath, pdfPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	stderrOutput := stderr.String()
	if stderrOutput != "" {
		fmt.Printf("⚠️ PyMuPDF output: %s\n", stderrOutput)
	}

	if err != nil {
		return "", fmt.Errorf("Python script execution failed: %v, error: %s", err, stderrOutput)
	}

	// Clean up temporary files
	os.Remove(scriptPath)
	fmt.Println("🧹 Cleaned up temporary files")

	result := out.String()
	if len(result) == 0 {
		return "", fmt.Errorf("Extracted text content is empty")
	}

	fmt.Printf("✅ Text extraction completed, total %d characters\n", len(result))
	return result, nil
}

// Implement actual functionality of Content Analysis Agent, call OpenAI API
func runLangChainContentAnalysisAgent(content string) (map[string]interface{}, error) {
	fmt.Println("🔍 Starting content analysis...")

	// Get original content length
	originalLen := len(content)
	fmt.Printf("📄 Original content length: %d characters\n", originalLen)

	// Check if key content exists
	keyPhrases := []string{
		"approach",
		"Segment",
		"Protocol",
		"Summary",
	}

	fmt.Println("🔎 Checking for key chapters in the document:")
	for _, phrase := range keyPhrases {
		if strings.Contains(content, phrase) {
			fmt.Printf("✅ Found key chapter: '%s'\n", phrase)
		} else {
			fmt.Printf("❌ Key chapter not found: '%s'\n", phrase)
		}
	}

	// Build API request
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	fmt.Println("✅ API key obtained")

	// Build request content
	fmt.Println("📦 Preparing to send request to OpenAI API...")
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a professional document analysis expert, skilled at extracting detailed structure and content from technical documents. Your task is:

1. Treat each page as an independent chapter, ensuring no content is missed
2. If a page contains obvious titles/subtitles, use them as chapter names; otherwise, create descriptive titles for each page
3. If a page is a continuation of the previous page, still treat it as a new chapter, noting "(continued)" in the title
4. Extract complete and detailed content for each chapter, without omitting any details, especially:
   - Detailed descriptions of charts and diagrams
   - Data and relationships in tables
   - Definitions of technical terms and their meanings
   - Complete details of processes and steps
5. Maintain technical accuracy and completeness, do not simplify or omit technical details

Remember, these chapter contents will be used for three different processing:
- Summary: Each chapter will generate a 2-4 sentence brief summary highlighting the core content
- Key Concepts: Each chapter will extract 3-5 core technical terms and their brief definitions
- Content Display: The complete content of each chapter will be retained for detailed learning

Return JSON format, structure as follows:
{
  "chapters": [
    {
      "title": "Chapter Title",
      "level": 1,
      "content": "Detailed chapter content, including complete descriptions of charts, tables, and technical details"
    },
    ...
  ]
}`,
			},
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.2,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	// Send API request
	fmt.Println("🌐 Sending request to OpenAI API...")
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	// Read API response
	fmt.Println("📥 Receiving OpenAI API response...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	// Check API response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// Parse API response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get message")
	}

	content, ok = message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get content")
	}

	// Extract JSON part
	fmt.Println("🔍 Extracting JSON data from API response...")
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Output chapter statistics
		if chapters, ok := result["chapters"].([]interface{}); ok {
			fmt.Printf("✅ Recognized %d chapters\n", len(chapters))
		}

		return result, nil
	}

	return nil, fmt.Errorf("Cannot extract valid JSON structure from API response")
}

// Implement actual functionality of Knowledge Extraction Agent
func runLangChainKnowledgeExtractionAgent(chapters []map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("🔍 Starting knowledge extraction process...")
	fmt.Printf("📚 Processing content of %d chapters\n", len(chapters))

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	fmt.Println("✅ API key obtained")

	// Build input content
	fmt.Println("🔄 Preparing chapter contents...")
	var inputContent string
	for _, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)

		// No content length restriction
		inputContent += fmt.Sprintf("# %s\n%s\n\n", title, content)
	}

	// Get original content length
	originalLen := len(inputContent)
	fmt.Printf("📄 Preparing to send %d characters for knowledge extraction\n", originalLen)

	// Build API request
	fmt.Println("📦 Building knowledge extraction request...")
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a professional knowledge extraction expert. Your task is to extract key concepts, technical terms, processes, chart information, and important knowledge points from technical documents.

Extraction principles:
1. Extract 3-5 independent technical terms or concepts from each chapter, must be specific technical nouns rather than general descriptions
2. Focus on the following types of technical terms:
   - Technical components in communication protocols (such as TDMA, CSMA/CR, etc.)
   - Error detection and fault tolerance mechanisms (such as CRC verification, FTMP algorithm, etc.)
   - Synchronization mechanisms (such as clock synchronization, rate offset, and phase offset, etc.)
   - Bit level frame transmission details (such as bit coding scheme, 8x redundancy, etc.)
   - Segment differences and scheduling methods
   - Professional terms specific to the technical field
   - Components and their relationships described in technical charts
   - Data and corresponding technical parameters in tables
   - Each step and their sequence in technical processes
   - Professional terms and their precise definitions
   - Key elements in protocols and standards
3. Provide concise but technical definitions (1-2 sentences) for each term, including its function and meaning
4. Avoid extracting chapter themes or summaries as key concepts, should extract deeper technical details
5. Terms should be specific enough, not too broad, for example extracting "TDMA(time division multiple access)" rather than "time division technology"

Ensure capturing the most important technical details in the document, especially those that are key to understanding chapter content but may be overlooked in professional terms.

Return JSON format, structure as follows:
{
  "key_concepts": [
    {
      "title": "Technical Term or Component Name(Specific Technical Concept)",
      "description": "Concise and accurate technical definition(1-2 sentences)",
      "chapter": "Belonging Chapter"
    },
    ...
  ]
}`,
			},
			{
				"role":    "user",
				"content": inputContent,
			},
		},
		"temperature": 0.2,
	}

	// Call API to get result
	fmt.Println("🌐 Sending knowledge extraction request to OpenAI API...")
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse API response
	fmt.Println("📥 Receiving knowledge extraction response...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	fmt.Println("🔍 Parsing knowledge extraction response...")
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get content")
	}

	// Extract JSON part
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Output chapter statistics
		if chapters, ok := result["chapters"].([]interface{}); ok {
			fmt.Printf("✅ Recognized %d chapters\n", len(chapters))
		}

		return result, nil
	}

	return nil, fmt.Errorf("Cannot extract valid JSON structure from API response")
}

// Implement actual functionality of Summary Agent
func runLangChainSummaryAgent(chapters []map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("📝 Starting to generate summaries...")
	fmt.Printf("📚 Processing summaries of %d chapters\n", len(chapters))

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	fmt.Println("✅ API key obtained")

	// Build input content
	fmt.Println("🔄 Preparing summary contents...")
	var inputContent string
	for _, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)

		// No content length restriction
		inputContent += fmt.Sprintf("# %s\n%s\n\n", title, content)
	}

	// Get original content length
	originalLen := len(inputContent)
	fmt.Printf("📄 Preparing to send %d characters for summary generation\n", originalLen)

	// Build API request
	fmt.Println("📦 Building summary request...")
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": `You are a summary expert. Generate concise summaries for each chapter in the document, highlighting key points.
				Return JSON format, structure as follows:
				{
					"summaries": [
						{
							"title": "Chapter Title",
							"summary": "Chapter Summary"
						},
						...
					]
				}`,
			},
			{
				"role":    "user",
				"content": inputContent,
			},
		},
		"temperature": 0.3,
	}

	// Call API to get result
	fmt.Println("🌐 Sending summary request to OpenAI API...")
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build API request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Failed to create API request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse API response
	fmt.Println("📥 Receiving summary response...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read API response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response: %v", err)
	}

	// Extract JSON content
	fmt.Println("🔍 Parsing summary response...")
	choices, ok := apiResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("API response format error: No choices field")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get message")
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("API response format error: Cannot get content")
	}

	// Extract JSON part
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := content[jsonStart : jsonEnd+1]
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
			return nil, fmt.Errorf("Failed to parse API returned JSON: %v", err)
		}

		// Output statistics
		if summaries, ok := result["summaries"].([]interface{}); ok {
			fmt.Printf("✅ Generated %d chapter summaries\n", len(summaries))
		} else {
			fmt.Println("⚠️ Summaries field not found")
		}

		return result, nil
	}

	return nil, fmt.Errorf("Cannot extract valid JSON structure from API response")
}

// Implement actual functionality of Format Agent
func runLangChainFormatAgent(chapters []map[string]interface{}, concepts []map[string]interface{}, summaries []map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("📊 Starting to format content...")
	fmt.Printf("📚 Formatting %d chapters, %d concepts, %d summaries\n",
		len(chapters), len(concepts), len(summaries))

	// Check chapter completeness
	fmt.Println("🔎 Checking chapter completeness:")
	keyChapters := []string{
		"Approach",
		"Method",
		"Segment",
		"Summary",
	}

	for _, key := range keyChapters {
		found := false
		for _, chapter := range chapters {
			title, _ := chapter["title"].(string)
			if strings.Contains(title, key) {
				found = true
				content, _ := chapter["content"].(string)
				contentLength := len(content)
				fmt.Printf("✅ Found chapter '%s', content length: %d characters\n", title, contentLength)
				break
			}
		}
		if !found {
			fmt.Printf("❓ Chapter '%s' not found, this may indicate incomplete content extraction\n", key)
		}
	}

	// Generate Markdown format
	fmt.Println("📝 Generating Markdown content...")
	var markdownContent strings.Builder

	// Try to get the title of the first chapter as the document title, if not use default title
	documentTitle := "Technical Document Learning Notes"
	if len(chapters) > 0 {
		if firstTitle, ok := chapters[0]["title"].(string); ok && firstTitle != "" {
			documentTitle = firstTitle + " - Study Note"
		}
	}

	// Add aesthetically pleasing title and introduction - use uniform format
	markdownContent.WriteString("# " + documentTitle + " 📘\n\n")
	markdownContent.WriteString("> *This is an automatically generated set of study notes, including a summary, key technical concepts, and detailed content explanations.*\n\n")
	markdownContent.WriteString("\n") // Add extra line

	// Add table of contents title, with emoji
	markdownContent.WriteString("## 📋 Table of Contents\n\n")

	// Generate table of contents - use correct numbering alignment
	for i, chapter := range chapters {
		title, _ := chapter["title"].(string)
		level, _ := chapter["level"].(float64)

		// Use appropriate indentation based on chapter level
		indent := ""
		if level > 1 {
			for j := 1; j < int(level); j++ {
				indent += "  "
			}
		}

		// Add table of contents item, use correct numbering format and adjust spacing
		markdownContent.WriteString(fmt.Sprintf("%s%d. [%s](#%s)\n",
			indent,
			i+1, // Ensure continuous numbering starting from 1
			title,
			strings.ReplaceAll(strings.ToLower(title), " ", "-")))
	}

	// Use consistent line style, and add spacing
	markdownContent.WriteString("\n\n") // Add line spacing before separator
	markdownContent.WriteString("---\n")
	markdownContent.WriteString("\n\n") // Add line spacing after separator

	// Generate chapter content
	for i, chapter := range chapters {
		title, _ := chapter["title"].(string)
		content, _ := chapter["content"].(string)
		level, _ := chapter["level"].(float64)

		// Use appropriate heading format based on chapter level
		headingLevel := "##"
		if level > 0 {
			headingLevel = strings.Repeat("#", int(level)+1)
		}

		// Chapter title add emoji and style, select different emoji based on chapter type
		chapterEmoji := "📖" // Default to book emoji
		if strings.Contains(strings.ToLower(title), "approach") {
			chapterEmoji = "🔍"
		} else if strings.Contains(strings.ToLower(title), "protocol") {
			chapterEmoji = "🔄"
		} else if strings.Contains(strings.ToLower(title), "segment") {
			chapterEmoji = "🧩"
		} else if strings.Contains(strings.ToLower(title), "synchronization") {
			chapterEmoji = "⏱️"
		} else if strings.Contains(strings.ToLower(title), "error") {
			chapterEmoji = "⚠️"
		} else if strings.Contains(strings.ToLower(title), "frame") {
			chapterEmoji = "📊"
		}

		// Use uniform style heading, ensure enough spacing
		markdownContent.WriteString(fmt.Sprintf("%s %s %s\n\n", headingLevel, chapterEmoji, title))

		// Find corresponding summary
		var summary string
		for _, s := range summaries {
			if sTitle, _ := s["title"].(string); sTitle == title {
				summary, _ = s["summary"].(string)
				break
			}
		}

		// Summary use concise consistent quote block format, add emoji, and ensure line spacing
		if summary != "" {
			markdownContent.WriteString("### 📌 Summary\n\n")
			markdownContent.WriteString("> " + summary + "\n\n")
			markdownContent.WriteString("\n") // Add extra line
		}

		// Find key concepts for this chapter
		conceptCount := 0
		var chapterConcepts []map[string]string
		for _, concept := range concepts {
			if chap, _ := concept["chapter"].(string); chap == title {
				conceptTitle, _ := concept["title"].(string)
				description, _ := concept["description"].(string)

				chapterConcepts = append(chapterConcepts, map[string]string{
					"title":       conceptTitle,
					"description": description,
				})
				conceptCount++
			}
		}

		// If there are key concepts, use more aesthetically pleasing list format to display, avoid table rendering issues, and adjust spacing
		if conceptCount > 0 {
			markdownContent.WriteString("### 🔑 Key Technical Concepts\n\n")

			// Use concise clear list format to display key concepts, avoid table alignment issues, and optimize space
			for _, concept := range chapterConcepts {
				markdownContent.WriteString(fmt.Sprintf("- **%s**：%s\n",
					concept["title"],
					concept["description"]))
			}
			markdownContent.WriteString("\n\n") // Add extra line

			fmt.Printf("📌 Chapter '%s' added %d key concepts\n", title, conceptCount)
		}

		// Check if content contains chart description
		hasChart := strings.Contains(strings.ToLower(content), "图") ||
			strings.Contains(strings.ToLower(content), "表") ||
			strings.Contains(strings.ToLower(content), "流程") ||
			strings.Contains(strings.ToLower(content), "diagram") ||
			strings.Contains(strings.ToLower(content), "chart") ||
			strings.Contains(strings.ToLower(content), "figure")

		// Add document content, add emoji and style, and adjust spacing
		if hasChart {
			markdownContent.WriteString("### 📝 Main Content (with charts and illustrations)\n\n")
		} else {
			markdownContent.WriteString("### 📝 Main Content\n\n")
		}

		// If content exceeds certain length, try segment processing
		if len(content) > 500 {
			paragraphs := strings.Split(content, "\n\n")

			// If able to segment
			if len(paragraphs) > 1 {
				for _, para := range paragraphs {
					paraTrimed := strings.TrimSpace(para)
					if len(paraTrimed) > 0 {
						// Check if it's a list item
						if strings.HasPrefix(paraTrimed, "- ") || strings.HasPrefix(paraTrimed, "* ") ||
							strings.HasPrefix(paraTrimed, "1. ") || strings.HasPrefix(paraTrimed, "2. ") {
							markdownContent.WriteString(para + "\n\n")
						} else {
							// Enhance visual effect of important paragraphs
							if len(paraTrimed) < 100 && (strings.Contains(paraTrimed, "important") ||
								strings.Contains(paraTrimed, "key")) {
								markdownContent.WriteString("**" + para + "**\n\n")
							} else {
								markdownContent.WriteString(para + "\n\n")
							}
						}
					}
				}
			} else {
				// If unable to segment based on paragraph, check if there's a list item
				if strings.Contains(content, "- ") || strings.Contains(content, "* ") ||
					strings.Contains(content, "1. ") || strings.Contains(content, "2. ") {
					markdownContent.WriteString(content + "\n\n")
				} else {
					// Try segment based on period for readability, and add spacing
					sentences := strings.Split(content, ". ")
					for j, sentence := range sentences {
						if len(strings.TrimSpace(sentence)) > 0 {
							if j < len(sentences)-1 {
								markdownContent.WriteString(sentence + ". ")
							} else {
								markdownContent.WriteString(sentence)
							}

							// Add extra line every 2 sentences for readability (changed from 3 to 2)
							if (j+1)%2 == 0 && j < len(sentences)-1 {
								markdownContent.WriteString("\n\n")
							}
						}
					}
					markdownContent.WriteString("\n\n")
				}
			}
		} else {
			// Short content directly display, ensure enough spacing
			markdownContent.WriteString(content + "\n\n\n")
		}

		// If contains chart, add hint, use uniform style, and add spacing
		if hasChart {
			markdownContent.WriteString("> 💡 **Chart Description**: This chapter contains chart or diagram description\n\n\n")
		}

		// If not the last chapter, add concise aesthetically pleasing separator, avoid using HTML, and add spacing
		if i < len(chapters)-1 {
			markdownContent.WriteString("\n") // Extra line
			markdownContent.WriteString("---\n")
			markdownContent.WriteString("\n\n") // Extra line
		}
	}

	// Add document ending, use uniform style, and add spacing
	markdownContent.WriteString("\n\n") // Extra line
	markdownContent.WriteString("## 📌 Learning Tips\n\n")
	markdownContent.WriteString("- Review chapter content with key concepts\n")
	markdownContent.WriteString("- Use Anki cards for spaced repetition learning\n")
	markdownContent.WriteString("- Try explaining each chapter's content in your own words\n")
	markdownContent.WriteString("- Link concepts to practical application scenarios\n\n")
	markdownContent.WriteString("\n") // Extra line
	markdownContent.WriteString("---\n\n")
	markdownContent.WriteString("*Document generated by AI, for assisting learning and reviewing*\n")

	// Generate Anki flashcard content
	fmt.Println("🎴 Generating Anki cards...")
	var ankiCards []map[string]string

	// Generate Anki cards from key concepts, distinguish between general concepts and chart/technical content
	for _, concept := range concepts {
		title, _ := concept["title"].(string)
		description, _ := concept["description"].(string)
		chapter, _ := concept["chapter"].(string)

		// Check if it's a technical chart related concept
		isChartConcept := strings.Contains(strings.ToLower(title), "图") ||
			strings.Contains(strings.ToLower(title), "表") ||
			strings.Contains(strings.ToLower(title), "流程") ||
			strings.Contains(strings.ToLower(title), "structure") ||
			strings.Contains(strings.ToLower(title), "framework") ||
			strings.Contains(strings.ToLower(description), "图表") ||
			strings.Contains(strings.ToLower(description), "diagram")

		// Check if it's a protocol/standard related concept
		isProtocolConcept := strings.Contains(strings.ToLower(title), "protocol") ||
			strings.Contains(strings.ToLower(title), "standard") ||
			strings.Contains(strings.ToLower(title), "specification") ||
			strings.Contains(strings.ToLower(description), "protocol") ||
			strings.Contains(strings.ToLower(description), "standard")

		tags := "concept,definition"
		if isChartConcept {
			tags = "concept,chart,technical architecture"
			// Create additional card for chart related concepts
			ankiCards = append(ankiCards, map[string]string{
				"front": fmt.Sprintf("What are the main components of %s?", title),
				"back":  fmt.Sprintf("%s\n\n(from chapter: %s)", description, chapter),
				"tags":  "component,chart,technical architecture",
			})
		} else if isProtocolConcept {
			tags = "concept,protocol,standard"
			// Create additional card for protocol related concepts
			ankiCards = append(ankiCards, map[string]string{
				"front": fmt.Sprintf("What is the main purpose of %s?", title),
				"back":  fmt.Sprintf("%s\n\n(from chapter: %s)", description, chapter),
				"tags":  "protocol,purpose,standard",
			})
		}

		// Add basic concept card
		ankiCards = append(ankiCards, map[string]string{
			"front": fmt.Sprintf("What is %s?", title),
			"back":  fmt.Sprintf("%s\n\n(from chapter: %s)", description, chapter),
			"tags":  tags,
		})

		// Add additional application scenario card for longer description
		if len(description) > 200 {
			ankiCards = append(ankiCards, map[string]string{
				"front": fmt.Sprintf("What are the application scenarios of %s?", title),
				"back":  fmt.Sprintf("%s\n\n(from chapter: %s)", description, chapter),
				"tags":  "application,scenario",
			})
		}
	}

	// Generate summary question and answer cards
	for _, summary := range summaries {
		title, _ := summary["title"].(string)
		sum, _ := summary["summary"].(string)

		// Basic chapter summary card
		ankiCards = append(ankiCards, map[string]string{
			"front": fmt.Sprintf("What does the %s chapter mainly talk about?", title),
			"back":  sum,
			"tags":  "summary,chapter",
		})

		// Check if it contains technical chart or process
		containsChart := strings.Contains(strings.ToLower(sum), "图") ||
			strings.Contains(strings.ToLower(sum), "表") ||
			strings.Contains(strings.ToLower(sum), "流程") ||
			strings.Contains(strings.ToLower(sum), "component") ||
			strings.Contains(strings.ToLower(sum), "architecture")

		// Add additional card for summary containing chart
		if containsChart {
			ankiCards = append(ankiCards, map[string]string{
				"front": fmt.Sprintf("What are the main technical components described in the %s chapter?", title),
				"back":  sum,
				"tags":  "component,chart,architecture",
			})
		}

		// Check if it contains steps or process
		containsProcess := strings.Contains(strings.ToLower(sum), "steps") ||
			strings.Contains(strings.ToLower(sum), "process") ||
			strings.Contains(strings.ToLower(sum), "stage") ||
			strings.Contains(strings.ToLower(sum), "sequence")

		// Add additional card for summary containing process
		if containsProcess {
			ankiCards = append(ankiCards, map[string]string{
				"front": fmt.Sprintf("What are the main steps or process described in the %s chapter?", title),
				"back":  sum,
				"tags":  "process,steps,stage",
			})
		}
	}

	fmt.Printf("✅ Generated %d Anki cards\n", len(ankiCards))
	fmt.Printf("📊 Markdown content length: %d characters\n", markdownContent.Len())

	// Build return result
	result := map[string]interface{}{
		"markdown_notes": markdownContent.String(),
		"anki_cards":     ankiCards,
		"format_version": "1.0",
	}

	fmt.Println("✅ Formatting completed!")
	return result, nil
}
