package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// RAG服务的基础URL
var ragServiceURL = "http://localhost:8081"

func init() {
	// 从环境变量获取RAG服务URL
	if url := os.Getenv("RAG_SERVICE_URL"); url != "" {
		ragServiceURL = url
	}
	fmt.Printf("RAG服务URL: %s\n", ragServiceURL)
}

// UploadHandler 处理文档上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析多部分表单
	err := r.ParseMultipartForm(10 << 20) // 限制10MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建一个新的multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", handler.Filename)
	if err != nil {
		http.Error(w, "Failed to create form file", http.StatusInternalServerError)
		return
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		http.Error(w, "Failed to copy file content", http.StatusInternalServerError)
		return
	}

	// 关闭multipart writer
	err = writer.Close()
	if err != nil {
		http.Error(w, "Failed to close writer", http.StatusInternalServerError)
		return
	}

	// 创建请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload", ragServiceURL), body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// 尝试解析响应并添加文件名
	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err == nil {
		// 如果解析成功，添加文件名
		responseData["filename"] = handler.Filename
		// 重新编码响应
		modifiedRespBody, err := json.Marshal(responseData)
		if err == nil {
			// 如果编码成功，使用修改后的响应
			respBody = modifiedRespBody
		}
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// ListDocumentsHandler 获取文档列表
func ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/documents", ragServiceURL), nil)
	if err != nil {
		http.Error(w, "创建请求失败", http.StatusInternalServerError)
		return
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "发送请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "读取响应失败", http.StatusInternalServerError)
		return
	}

	// 解析响应
	var responseData []map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		// 如果解析失败，直接返回原始响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 处理文档状态
	for _, doc := range responseData {
		// 确保状态字段存在
		if status, ok := doc["status"].(string); ok {
			// 如果状态是空字符串，设置为"已处理"
			if status == "" {
				doc["status"] = "processed"
			}
		} else {
			// 如果状态字段不存在，添加状态字段
			doc["status"] = "processed"
		}

		// 确保大小字段存在
		if _, ok := doc["size"].(float64); !ok {
			doc["size"] = 0
		}

		// 确保文件名字段存在
		if filename, ok := doc["document_name"].(string); ok {
			// 如果文档名称存在，确保也有filename字段
			doc["filename"] = filename
		} else if filename, ok := doc["name"].(string); ok {
			// 尝试使用name字段
			doc["filename"] = filename
		} else if filename, ok := doc["file_name"].(string); ok {
			// 尝试使用file_name字段
			doc["filename"] = filename
		} else {
			// 如果没有任何文件名字段，设置为未命名文档
			doc["filename"] = "Unnamed Document"
		}

		// 确保上传时间字段存在
		if _, ok := doc["upload_time"].(string); !ok {
			// 尝试使用其他可能的时间字段
			if uploadedAt, ok := doc["uploadedAt"].(string); ok {
				doc["upload_time"] = uploadedAt
			} else if processedAt, ok := doc["processedAt"].(string); ok {
				doc["upload_time"] = processedAt
			} else if createdAt, ok := doc["createdAt"].(string); ok {
				doc["upload_time"] = createdAt
			} else {
				// 如果没有任何时间字段，使用一个固定的时间，而不是当前时间
				doc["upload_time"] = "2023-01-01T00:00:00Z"
			}
		}
	}

	// 重新编码响应
	formattedRespBody, err := json.Marshal(responseData)
	if err != nil {
		// 如果编码失败，直接返回原始响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(formattedRespBody)
}

// DeleteDocumentHandler 删除文档
func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// 处理CORS预检请求
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	// 获取文档ID
	vars := mux.Vars(r)
	docID := vars["doc_id"]
	fmt.Printf("收到删除文档请求，文档ID: %s\n", docID)

	// 创建请求
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/document/%s", ragServiceURL, docID), nil)
	if err != nil {
		fmt.Printf("创建删除请求失败: %v\n", err)
		http.Error(w, "创建请求失败", http.StatusInternalServerError)
		return
	}

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}
	fmt.Printf("发送删除请求到: %s/document/%s\n", ragServiceURL, docID)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送删除请求失败: %v\n", err)
		http.Error(w, "发送请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取删除响应失败: %v\n", err)
		http.Error(w, "读取响应失败", http.StatusInternalServerError)
		return
	}

	fmt.Printf("删除请求响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("删除请求响应内容: %s\n", string(respBody))

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// ReprocessDocumentHandler 重新处理文档
func ReprocessDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// 获取文档ID
	vars := mux.Vars(r)
	docID := vars["doc_id"]

	// 创建请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/document/%s/reprocess", ragServiceURL, docID), nil)
	if err != nil {
		http.Error(w, "创建请求失败", http.StatusInternalServerError)
		return
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "发送请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "读取响应失败", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// QueryHandler 处理查询
func QueryHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var requestBody struct {
		Query string `json:"query"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	// 创建请求体
	queryBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to create request body", http.StatusInternalServerError)
		return
	}

	// 创建请求
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/query", ragServiceURL), bytes.NewBuffer(queryBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// 解析响应
	var responseData map[string]interface{}
	if err := json.Unmarshal(respBody, &responseData); err != nil {
		// 如果解析失败，直接返回原始响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 格式化回答为Markdown
	if answer, ok := responseData["answer"].(string); ok {
		// 添加Markdown格式
		formattedAnswer := formatToMarkdown(answer)
		responseData["answer"] = formattedAnswer
	}

	// 处理参考来源
	if sources, ok := responseData["sources"].([]interface{}); ok {
		// 过滤有效的参考来源
		var validSources []interface{}
		for _, source := range sources {
			if sourceMap, ok := source.(map[string]interface{}); ok {
				// 检查是否有必要的字段
				if docName, hasDocName := sourceMap["document_name"].(string); hasDocName && docName != "" {
					if content, hasContent := sourceMap["content"].(string); hasContent && content != "" {
						// 保留有效的参考来源
						validSources = append(validSources, sourceMap)
					}
				}
			}
		}
		responseData["sources"] = validSources
	}

	// 重新编码响应
	formattedRespBody, err := json.Marshal(responseData)
	if err != nil {
		// 如果编码失败，直接返回原始响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(formattedRespBody)
}

// formatToMarkdown 将普通文本格式化为Markdown
func formatToMarkdown(text string) string {
	// 如果已经是Markdown格式，直接返回
	if strings.Contains(text, "#") || strings.Contains(text, "```") {
		return text
	}

	// 分割文本为段落
	paragraphs := strings.Split(text, "\n\n")
	var formattedParagraphs []string

	// 处理每个段落
	for i, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 第一段作为标题
		if i == 0 {
			// 检查是否已经是标题格式
			if !strings.HasPrefix(para, "# ") && !strings.HasPrefix(para, "## ") {
				// 如果段落很长，可能不是标题，使用通用标题
				if len(para) > 100 {
					formattedParagraphs = append(formattedParagraphs, "## Answer")
					formattedParagraphs = append(formattedParagraphs, para)
				} else {
					formattedParagraphs = append(formattedParagraphs, "## "+para)
				}
			} else {
				formattedParagraphs = append(formattedParagraphs, para)
			}
		} else {
			// 检查是否是列表项
			if strings.HasPrefix(para, "- ") || strings.HasPrefix(para, "* ") ||
				(len(para) > 2 && para[0] >= '0' && para[0] <= '9' && para[1] == '.') {
				formattedParagraphs = append(formattedParagraphs, para)
			} else if strings.HasPrefix(para, "Features:") || strings.HasPrefix(para, "Advantages:") ||
				strings.HasPrefix(para, "Disadvantages:") || strings.HasPrefix(para, "Summary:") ||
				strings.HasPrefix(para, "特点:") || strings.HasPrefix(para, "优点:") ||
				strings.HasPrefix(para, "缺点:") || strings.HasPrefix(para, "总结:") {
				// 将特定前缀转换为小标题
				parts := strings.SplitN(para, ":", 2)
				if len(parts) == 2 {
					// 转换中文标题为英文
					title := parts[0]
					if title == "特点" {
						title = "Features"
					} else if title == "优点" {
						title = "Advantages"
					} else if title == "缺点" {
						title = "Disadvantages"
					} else if title == "总结" {
						title = "Summary"
					}
					formattedParagraphs = append(formattedParagraphs, "### "+title)
					formattedParagraphs = append(formattedParagraphs, parts[1])
				} else {
					formattedParagraphs = append(formattedParagraphs, para)
				}
			} else {
				formattedParagraphs = append(formattedParagraphs, para)
			}
		}
	}

	return strings.Join(formattedParagraphs, "\n\n")
}

// GetStatusHandler 获取文档处理状态
func GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	// 获取任务ID
	vars := mux.Vars(r)
	taskID := vars["task_id"]

	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/status/%s", ragServiceURL, taskID), nil)
	if err != nil {
		http.Error(w, "创建请求失败", http.StatusInternalServerError)
		return
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "发送请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "读取响应失败", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// ClearVectorDBHandler 清理向量数据库中的所有向量
func ClearVectorDBHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 只允许POST请求
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "只允许POST请求"})
		return
	}

	// 调用RAG服务的清理向量数据库API
	fmt.Printf("正在调用RAG服务清理向量数据库: %s/clear-vectors\n", ragServiceURL)

	// 创建一个带超时的客户端
	client := &http.Client{
		Timeout: 30 * time.Second, // 设置30秒超时
	}

	// 创建请求
	req, err := http.NewRequest("POST", ragServiceURL+"/clear-vectors", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "创建请求失败: " + err.Error()})
		return
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("调用RAG服务失败: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "调用RAG服务失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "读取响应失败: " + err.Error()})
		return
	}

	// 检查响应状态码
	fmt.Printf("RAG服务响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("RAG服务响应内容: %s\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "向量数据库已成功清空"})
}
