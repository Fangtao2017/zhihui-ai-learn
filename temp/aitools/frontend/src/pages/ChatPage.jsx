import React, { useState, useEffect, useRef } from 'react';
import { Layout, Input, Button, Select, List, Avatar, message, Spin, Typography, Modal } from 'antd';
import { SendOutlined, PlusOutlined, EditOutlined, DeleteOutlined, UserOutlined, RobotOutlined } from '@ant-design/icons';
import request from '../utils/request';
import dayjs from 'dayjs';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';

const { Sider, Content } = Layout;
const { TextArea } = Input;
const { Title, Text } = Typography;

const API_BASE_URL = 'http://localhost:8080';

// Add inline styles
const typingIndicatorStyle = `
.ai-message-loading {
    padding: 8px 0;
}

.typing-indicator {
    display: inline-flex;
    align-items: center;
}

.typing-indicator span {
    height: 8px;
    width: 8px;
    margin: 0 2px;
    background-color: #1890ff;
    border-radius: 50%;
    display: inline-block;
    opacity: 0.4;
}

.typing-indicator span:nth-child(1) {
    animation: pulse 1s infinite ease-in-out;
}

.typing-indicator span:nth-child(2) {
    animation: pulse 1s infinite ease-in-out 0.2s;
}

.typing-indicator span:nth-child(3) {
    animation: pulse 1s infinite ease-in-out 0.4s;
}

@keyframes pulse {
    0% {
        opacity: 0.4;
        transform: scale(1);
    }
    50% {
        opacity: 1;
        transform: scale(1.2);
    }
    100% {
        opacity: 0.4;
        transform: scale(1);
    }
}

/* 自定义滚动条样式 */
.chat-history-list::-webkit-scrollbar {
    width: 6px;
    height: 6px;
}

.chat-history-list::-webkit-scrollbar-thumb {
    background-color: rgba(0, 0, 0, 0.2);
    border-radius: 3px;
}

.chat-history-list::-webkit-scrollbar-track {
    background-color: rgba(0, 0, 0, 0.05);
}

/* 聊天历史项悬停效果 */
.chat-history-item {
    position: relative;
    transition: all 0.3s ease;
}

.chat-history-item:hover {
    background-color: #f5f5f5 !important;
}

.chat-history-item.active {
    background-color: #e6f7ff !important;
    border-left: 3px solid #1890ff !important;
}

/* 聊天记录列表容器样式 */
.chat-history-list {
    overflow-y: auto;
    scrollbar-width: thin;
    height: calc(100vh - 204px) !important; /* 视口高度减去头部64px和选择器区域约140px */
}

/* Add custom styles to optimize Markdown content display */
.markdown-content h1,
.markdown-content h2,
.markdown-content h3,
.markdown-content h4,
.markdown-content h5,
.markdown-content h6 {
    margin-top: 8px;
    margin-bottom: 4px;
}

.markdown-content p {
    margin-top: 2px;
    margin-bottom: 6px;
}

.markdown-content ul,
.markdown-content ol {
    margin-top: 2px;
    margin-bottom: 6px;
    padding-left: 20px;
}

.markdown-content li {
    margin-bottom: 1px;
}

.markdown-content pre {
    margin-bottom: 6px;
}

.markdown-content blockquote {
    margin-top: 4px;
    margin-bottom: 6px;
    padding-left: 12px;
    border-left: 3px solid #ddd;
}

.markdown-content img {
    max-width: 100%;
    margin: 4px 0;
}

.markdown-content table {
    margin: 4px 0;
}

.markdown-content hr {
    margin: 6px 0;
}
`;

const ChatPage = () => {
    const navigate = useNavigate();
    const [selectedModel, setSelectedModel] = useState('gpt-4o');
    const [inputMessage, setInputMessage] = useState('');
    const [chatHistory, setChatHistory] = useState([]);
    const [currentChat, setCurrentChat] = useState([]);
    const [loading, setLoading] = useState(false);
    const [currentChatId, setCurrentChatId] = useState(null);
    const [modelLocked, setModelLocked] = useState(false);
    const [currentChatModel, setCurrentChatModel] = useState('');
    const messagesEndRef = useRef(null);
    const [editingTitle, setEditingTitle] = useState(null);
    const [newTitle, setNewTitle] = useState('');
    const messagesContainerRef = useRef(null);
    // Add delete confirmation dialog state
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [chatToDelete, setChatToDelete] = useState(null);
    // Add language preference state, get initial value from localStorage
    const [languagePreference, setLanguagePreference] = useState(() => {
        const savedPreference = localStorage.getItem('languagePreference');
        return savedPreference || 'auto'; // 默认使用自动检测，而不是固定的english或chinese
    });
    // 添加新的状态
    const [availableModels, setAvailableModels] = useState([]);
    // 只保留聊天历史列表的ref
    const chatHistoryListRef = useRef(null);

    // Listen for language preference changes, save to localStorage
    useEffect(() => {
        localStorage.setItem('languagePreference', languagePreference);
    }, [languagePreference]);

    const scrollToBottom = () => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        } else if (messagesContainerRef.current) {
            messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
        }
    };

    useEffect(() => {
        scrollToBottom();
    }, [currentChat]);

    useEffect(() => {
        const token = localStorage.getItem('token');
        if (!token) {
            navigate('/login', { replace: true });
            return;
        }

        const init = async () => {
            try {
                await fetchChatHistory();
                
                // Try to get current chat ID from localStorage
                const savedChatId = localStorage.getItem('currentChatId');
                if (savedChatId) {
                    setCurrentChatId(savedChatId);
                    await fetchChatMessages(savedChatId);
                }
            } catch (err) {
                console.error('Initialization error:', err);
                message.error('Initialization failed');
            }
        };

        init();
    }, []);

    const fetchChatHistory = async () => {
        try {
            const history = await request(`${API_BASE_URL}/api/chat/history`);
            if (Array.isArray(history)) {
                setChatHistory(history);
                // 在聊天历史更新后滚动到当前选中的聊天
                setTimeout(scrollToCurrentChat, 100);
            }
        } catch (err) {
            console.error('Error fetching chat history:', err);
            message.error('Failed to fetch chat history');
        }
    };

    // 添加一个函数来检查聊天是否为空
    const isChatEmpty = (chatId) => {
        return currentChat.length === 0;
    };

    // 添加函数来删除空聊天
    const deleteEmptyChat = async (chatId) => {
        if (!chatId) return;
        
        try {
            console.log(`Deleting empty chat: ${chatId}`);
            await request.delete(`${API_BASE_URL}/api/chat/${chatId}`);
            await fetchChatHistory(); // 刷新聊天列表
        } catch (err) {
            console.error("Error deleting empty chat:", err);
            // 静默失败，不显示错误消息
        }
    };
    
    // 修改fetchChatMessages函数，增加错误处理和日志记录
    const fetchChatMessages = async (chatId) => {
        if (!chatId) {
            console.error('Invalid chat ID');
            return;
        }
        
        try {
            setLoading(true);
            console.log(`Fetching messages for chat ID: ${chatId}`);
            
            const response = await request(`${API_BASE_URL}/api/chat/${chatId}/messages`);
            
            // 检查响应是否为null，如果是则视为空数组
            if (response === null) {
                console.log(`Response is null for chat ID: ${chatId}, treating as empty array`);
                setCurrentChat([]);
                setCurrentChatId(chatId);
                localStorage.setItem('currentChatId', chatId);
                setModelLocked(false);
                setCurrentChatModel('');
                setLoading(false);
                return;
            }
            
            // 确保messages是数组，即使是空数组
            if (Array.isArray(response)) {
                console.log(`Received ${response.length} messages for chat ID: ${chatId}`);
                setCurrentChat(response);
                setCurrentChatId(chatId);
                localStorage.setItem('currentChatId', chatId);
                
                // 检查是否有消息，如果有，则锁定模型
                if (response.length > 0) {
                    setModelLocked(true);
                    
                    // 获取聊天信息以检索使用的模型
                    try {
                        const chatInfo = await request(`${API_BASE_URL}/api/chat/${chatId}/info`);
                        if (chatInfo && chatInfo.model) {
                            setSelectedModel(chatInfo.model);
                            setCurrentChatModel(chatInfo.model);
                            console.log(`Using model from chat: ${chatInfo.model}`);
                        }
                    } catch (err) {
                        console.error('Error fetching chat info:', err);
                        // 如果无法获取模型信息，仍然锁定但保持当前选择的模型
                        setCurrentChatModel(selectedModel);
                    }
                } else {
                    console.log(`Chat ${chatId} is empty, unlocking model selection`);
                    setModelLocked(false);
                    // 对于空聊天，允许用户选择模型
                    setCurrentChatModel('');
                }
            } else {
                console.error('Received non-array response for messages:', response);
                // 处理非数组响应，仍然打开这个聊天但使用空数组
                setCurrentChat([]);
                setCurrentChatId(chatId);
                localStorage.setItem('currentChatId', chatId);
                setModelLocked(false);
                setCurrentChatModel('');
                // 通知用户但不阻止操作
                message.warning('获取聊天记录格式异常，但仍然打开该聊天');
            }
        } catch (err) {
            console.error('Error fetching chat messages:', err);
            // 如果是404错误，可能聊天记录已被删除
            if (err.status === 404) {
                console.log(`Chat ${chatId} may have been deleted, creating a new empty chat view`);
                setCurrentChat([]);
                setCurrentChatId(chatId);
                localStorage.setItem('currentChatId', chatId);
                setModelLocked(false);
                message.info('没有找到聊天记录，已为您创建新的聊天');
            } else {
                message.error('获取聊天记录失败');
                
                // 尝试恢复到之前的聊天
                if (currentChatId && currentChatId !== chatId) {
                    console.log(`Attempting to recover by staying on current chat: ${currentChatId}`);
                } else {
                    // 如果没有当前聊天可以恢复，仍然打开这个聊天但使用空数组
                    setCurrentChat([]);
                    setCurrentChatId(chatId);
                    localStorage.setItem('currentChatId', chatId);
                    setModelLocked(false);
                    setCurrentChatModel('');
                }
            }
        } finally {
            setLoading(false);
        }
    };
    
    // 修改handleCreateChat函数，保留检查并删除当前空聊天的逻辑，并添加检查其他空聊天的逻辑
    const handleCreateChat = async () => {
        try {
            setLoading(true);
            
            // 检查当前聊天是否为空，如果为空则删除
            if (currentChatId && currentChat.length === 0) {
                console.log(`Current chat ${currentChatId} is empty, deleting it before creating new chat`);
                await deleteEmptyChat(currentChatId);
            }
            
            // 查找并删除所有其他空聊天
            await findAndDeleteEmptyChats();
            
            // 设置请求超时
            const timeoutId = setTimeout(() => {
                message.error('创建聊天超时，请重试');
                setLoading(false);
                throw new Error('创建聊天超时');
            }, 10000); // 10秒超时
            
            const response = await request.post(`${API_BASE_URL}/api/chat/new`);
            
            // 清除超时计时器
            clearTimeout(timeoutId);
            
            if (response && response.id) {
                const newChatId = response.id;
                console.log('New chat created with ID:', newChatId);
                
                // 更新聊天历史
                await fetchChatHistory();
                
                // 设置当前聊天ID
                setCurrentChatId(newChatId);
                localStorage.setItem('currentChatId', newChatId);
                setCurrentChat([]);
                setInputMessage('');
                
                // 创建新聊天时解锁模型选择
                setModelLocked(false);
                setCurrentChatModel('');
                
                await fetchChatHistory();
                return newChatId; // 返回新创建的聊天ID
            } else {
                console.error('Failed to create chat: No ID returned');
                throw new Error('Failed to create chat');
            }
        } catch (error) {
            console.error('Create chat error:', error);
            message.error(error.message || '创建新聊天失败');
        } finally {
            // 确保在任何情况下都重置加载状态
            setLoading(false);
        }
    };
    
    // 添加一个新函数，用于查找和删除所有空聊天（除了当前聊天）
    const findAndDeleteEmptyChats = async () => {
        try {
            // 对每个聊天记录进行检查
            for (const chat of chatHistory) {
                // 跳过当前打开的聊天
                if (chat.id === currentChatId) {
                    continue;
                }
                
                // 获取该聊天的消息
                const messages = await request(`${API_BASE_URL}/api/chat/${chat.id}/messages`);
                
                // 如果没有消息，删除这个聊天
                if (!messages || messages.length === 0) {
                    console.log(`Found empty chat: ${chat.id}, deleting it`);
                    await deleteEmptyChat(chat.id);
                }
            }
        } catch (err) {
            console.error('Error finding and deleting empty chats:', err);
            // 静默失败，不影响正常流程
        }
    };

    const handleSend = async () => {
        if (!inputMessage.trim()) return;
        
        // 存储原始消息，以便在出错时恢复
        const originalMessage = inputMessage;
        
        // 检查是否已经在加载中，防止重复发送
        if (loading) {
            message.info('消息正在发送中，请稍候');
            return;
        }
        
        // Check if it's a language preference request
        const englishRequestPattern = /(?:please|can you|would you)?\s*(?:use|speak|answer in|respond in|reply in|talk in)\s*(?:english)/i;
        const chineseRequestPattern = /(?:please|can you|would you)?\s*(?:use|speak|answer in|respond in|reply in|talk in)\s*(?:chinese|mandarin)/i;
        
        if (englishRequestPattern.test(inputMessage)) {
            // Set language preference to English
            setLanguagePreference('english');
            
            // Add user message to UI
            const userMessage = {
                role: 'user',
                content: inputMessage,
                timestamp: new Date().toISOString()
            };
            
            setCurrentChat(prev => [...prev, userMessage]);
            setInputMessage('');
            
            // Add system response
            const systemResponse = {
                role: 'assistant',
                content: 'I will now respond in English. How can I help you today?',
                timestamp: new Date().toISOString()
            };
            
            setCurrentChat(prev => [...prev, systemResponse]);
            
            // Save messages to database
            if (currentChatId) {
                try {
                    await request.post(`${API_BASE_URL}/api/chat/${currentChatId}/messages`, { content: inputMessage });
                    await saveAIMessage(currentChatId, systemResponse.content);
                } catch (err) {
                    console.error('Error saving language preference messages:', err);
                }
            }
            
            // Scroll to bottom
            setTimeout(scrollToBottom, 100);
            return;
        }
        
        if (chineseRequestPattern.test(inputMessage)) {
            // Set language preference to Chinese
            setLanguagePreference('chinese');
            
            // Add user message to UI
            const userMessage = {
                role: 'user',
                content: inputMessage,
                timestamp: new Date().toISOString()
            };
            
            setCurrentChat(prev => [...prev, userMessage]);
            setInputMessage('');
            
            // Add system response
            const systemResponse = {
                role: 'assistant',
                content: '我现在将用中文回答。今天有什么可以帮到您的吗？',
                timestamp: new Date().toISOString()
            };
            
            setCurrentChat(prev => [...prev, systemResponse]);
            
            // Save messages to database
            if (currentChatId) {
                try {
                    await request.post(`${API_BASE_URL}/api/chat/${currentChatId}/messages`, { content: inputMessage });
                    await saveAIMessage(currentChatId, systemResponse.content);
                } catch (err) {
                    console.error('Error saving language preference messages:', err);
                }
            }
            
            // Scroll to bottom
            setTimeout(scrollToBottom, 100);
            return;
        }
        
        // Prioritize using currentChatId from component state, if not available try to get from localStorage
        let chatId = currentChatId || localStorage.getItem('currentChatId');
        console.log('Current chat ID:', chatId);
        
        if (!chatId) {
            console.log('No active chat, creating new chat...');
            // If no current chat ID, try to create a new chat
            try {
                chatId = await handleCreateChat();
                if (!chatId) {
                    message.error('Failed to create new chat');
                    return;
                }
                console.log('New chat created with ID:', chatId);
            } catch (err) {
                console.error('Error creating new chat:', err.message);
                message.error('Failed to create new conversation');
                return; // 创建失败时直接返回，不继续尝试发送消息
            }
        }

        // 如果这是当前聊天的第一条消息，锁定模型并记录所使用的模型
        if (currentChat.length === 0) {
            setModelLocked(true);
            setCurrentChatModel(selectedModel);
            
            // 向后端发送当前使用的模型信息
            try {
                await request.put(`${API_BASE_URL}/api/chat/${chatId}/model`, { model: selectedModel });
            } catch (err) {
                console.error('Error saving chat model:', err);
                // 即使保存失败，我们仍然在前端锁定模型
            }
        }

        // Add user message to UI
        const userMessage = {
            role: 'user',
            content: inputMessage,
            timestamp: new Date().toISOString()
        };
        
        setCurrentChat(prev => [...prev, userMessage]);
        setInputMessage('');
        
        // Add an empty AI message for streaming updates
        const aiMessageId = Date.now().toString();
        const aiMessage = {
            id: aiMessageId,
            role: 'assistant',
            content: '',
            timestamp: new Date().toISOString(),
            isStreaming: true // Mark as streaming
        };
        
        setCurrentChat(prev => [...prev, aiMessage]);
        setLoading(true); // Start loading
        
        // 设置超时处理
        const loadingTimeoutId = setTimeout(() => {
            setLoading(false); // 强制重置加载状态
            message.error('响应超时，请重试');
            setCurrentChat(prev => prev.map(msg => 
                msg.id === aiMessageId 
                    ? { ...msg, content: '**错误: 响应超时，请重试**', isStreaming: false }
                    : msg
            ));
        }, 60000); // 60秒超时
        
        // Scroll to bottom to ensure user sees the latest messages
        setTimeout(scrollToBottom, 100);
        
        try {
            console.log('Sending message to chat ID:', chatId);
            
            // Build URL parameters, ensure proper encoding
            const params = new URLSearchParams();
            params.append('message', inputMessage);
            params.append('model', selectedModel);
            // Add language preference parameter
            params.append('language', languagePreference);
            
            // Create EventSource connection
            const eventSourceUrl = `${API_BASE_URL}/api/chat/${chatId}/messages/stream?${params.toString()}`;
            console.log('EventSource URL:', eventSourceUrl);
            
            // Use fetch API instead of EventSource for better request control
            try {
                const controller = new AbortController();
                const signal = controller.signal;
                
                const response = await fetch(eventSourceUrl, {
                    method: 'GET',
                    headers: {
                        'Accept': 'text/event-stream',
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    signal
                });
                
                // 清除超时计时器
                clearTimeout(loadingTimeoutId);
                
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                let buffer = '';
                let fullResponse = '';
                
                // Process streaming response
                const processStream = async () => {
                    while (true) {
                        const { value, done } = await reader.read();
                        
                        if (done) {
                            console.log('Stream complete');
                            setLoading(false); // 确保在流完成时重置加载状态
                            break;
                        }
                        
                        // Decode received data
                        const chunk = decoder.decode(value, { stream: true });
                        buffer += chunk;
                        
                        // Process received data lines
                        const lines = buffer.split('\n\n');
                        buffer = lines.pop() || '';
                        
                        for (const line of lines) {
                            if (line.trim() === '') continue;
                            
                            // Extract data portion
                            const dataLine = line.replace(/^data: /, '').trim();
                            
                            if (dataLine === '[DONE]') {
                                // Stream ended
                                console.log('Stream complete, saving response');
                                
                                // Update message state, remove isStreaming flag
                                setCurrentChat(prev => 
                                    prev.map(msg => 
                                        msg.id === aiMessageId 
                                            ? { ...msg, content: fullResponse, isStreaming: false } 
                                            : msg
                                    )
                                );
                                
                                // Save complete AI response to database
                                if (fullResponse.trim()) {
                                    console.log('Saving AI response:', fullResponse.substring(0, 100) + '...');
                                    // 延迟50ms保存消息，确保Claude的完整响应能被处理
                                    setTimeout(() => {
                                        saveAIMessage(chatId, fullResponse);
                                        console.log('AI response saved after delay for Claude model');
                                    }, 50);
                                } else {
                                    console.warn('Empty response, not saving');
                                }
                                
                                setLoading(false);
                                return;
                            }
                            
                            if (dataLine.startsWith('ERROR:')) {
                                // Handle error
                                const errorMessage = dataLine.substring(6).trim();
                                console.error('Stream error:', errorMessage);
                                
                                // 检查是否是模型回退通知
                                if (errorMessage.includes('Claude模型不可用') || errorMessage.includes('using gpt-')) {
                                    // 更新聊天模型信息
                                    const fallbackModelMatch = errorMessage.match(/使用(.*?)代替/);
                                    if (fallbackModelMatch && fallbackModelMatch[1]) {
                                        const fallbackModel = fallbackModelMatch[1];
                                        setCurrentChatModel(fallbackModel);
                                        setSelectedModel(fallbackModel);
                                        
                                        console.log(`Auto-switched to fallback model: ${fallbackModel}`);
                                        message.warning(`Claude模型不可用，已自动切换到${fallbackModel}`);
                                        
                                        // 不显示错误消息，继续流式处理
                                        continue;
                                    }
                                }
                                
                                setCurrentChat(prev => 
                                    prev.map(msg => 
                                        msg.id === aiMessageId 
                                            ? { ...msg, content: `**错误: ${errorMessage || '获取AI响应失败'}**`, isStreaming: false } 
                                            : msg
                                    )
                                );
                                
                                message.error(`AI响应错误: ${errorMessage || '未知错误'}`);
                                setLoading(false);
                                return;
                            }
                            
                            // Try to parse JSON-encoded content
                            let contentChunk = dataLine;
                            try {
                                // 检查是否明显是JSON格式
                                if ((dataLine.startsWith('{') && dataLine.endsWith('}')) || 
                                    (dataLine.startsWith('"') && dataLine.endsWith('"'))) {
                                    // 尝试解析JSON
                                    contentChunk = JSON.parse(dataLine);
                                    
                                    // 检查是否包含Claude API的元数据格式
                                    if (typeof contentChunk === 'object' && 
                                        (contentChunk.type === 'stream_start' || 
                                         contentChunk.type === 'message_start' ||
                                         contentChunk.type === 'content_block_start' ||
                                         contentChunk.type === 'message_delta' ||
                                         contentChunk.type === 'content_block_delta' ||
                                         contentChunk.type === 'ping')) {
                                        
                                        console.log('Detected Claude API metadata:', contentChunk.type);
                                        
                                        // 从Claude API元数据中提取实际内容
                                        if (contentChunk.content_block && contentChunk.content_block.text) {
                                            contentChunk = contentChunk.content_block.text;
                                        } else if (contentChunk.delta && contentChunk.delta.text) {
                                            contentChunk = contentChunk.delta.text;
                                        } else {
                                            // 如果无法从元数据中提取内容，跳过此块
                                            console.log('Skipping metadata without content');
                                            continue;
                                        }
                                    }
                                }
                            } catch (e) {
                                console.warn('非JSON格式内容或解析失败，使用原始内容:', e);
                                // 解析失败时使用原始内容
                            }
                            
                            // 确保contentChunk是字符串
                            if (typeof contentChunk !== 'string') {
                                contentChunk = String(contentChunk);
                            }
                            
                            // 更新AI响应内容
                            fullResponse += contentChunk;
                            
                            // Real-time update AI message in UI
                            setCurrentChat(prev => 
                                prev.map(msg => 
                                    msg.id === aiMessageId 
                                        ? { ...msg, content: fullResponse } 
                                        : msg
                                )
                            );
                            
                            // Scroll to bottom to ensure user sees newest content
                            scrollToBottom();
                        }
                    }
                };
                
                // Start processing stream
                processStream().catch(error => {
                    console.error('Error processing stream:', error);
                    message.error('处理响应流时出错');
                    
                    // 在出错时更新消息显示
                    setCurrentChat(prev => 
                        prev.map(msg => 
                            msg.id === aiMessageId 
                                ? { ...msg, content: '**错误: 处理响应流时出错**', isStreaming: false } 
                                : msg
                        )
                    );
                    
                    setLoading(false);
                });
                
            } catch (error) {
                console.error('Error processing response stream:', error);
                message.error('处理响应流时出错');
                
                // 在错误时清除超时计时器
                clearTimeout(loadingTimeoutId);
                
                // 在出错时更新消息显示
                setCurrentChat(prev => 
                    prev.map(msg => 
                        msg.id === aiMessageId 
                            ? { ...msg, content: '**错误: 处理响应流时出错**', isStreaming: false } 
                            : msg
                    )
                );
                
                setLoading(false); // 确保在出错时重置加载状态
            }
        } catch (err) {
            console.error('Error sending message:', err);
            message.error('发送消息失败');
            
            // 在错误时清除超时计时器
            clearTimeout(loadingTimeoutId);
            
            // 恢复输入框的消息，以便用户可以重试
            setInputMessage(originalMessage);
            
            // 移除失败的消息
            setCurrentChat(prev => prev.filter(msg => msg.id !== aiMessageId));
            
            setLoading(false); // 确保在出错时重置加载状态
        }
    };
    
    // Save AI response to database
    const saveAIMessage = async (chatId, content) => {
        if (!chatId || !content) {
            console.error('Invalid chat ID or content');
            return;
        }
        
        try {
            console.log('Saving AI message to chat ID:', chatId);
            // Ensure using full URL
            const url = `${API_BASE_URL}/api/chat/${chatId}/messages/ai`;
            console.log('Request URL:', url);
            
            // 添加重试机制
            let retryCount = 0;
            const maxRetries = 3;
            
            const saveWithRetry = async () => {
                try {
                    // Use native fetch instead of request tool function
                    const token = localStorage.getItem('token');
                    const response = await fetch(url, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': token ? `Bearer ${token}` : '',
                        },
                        body: JSON.stringify({ content })
                    });
                    
                    if (!response.ok) {
                        const errorText = await response.text();
                        throw new Error(errorText || response.statusText);
                    }
                    
                    console.log('AI message saved successfully');
                } catch (err) {
                    console.error(`Error saving AI response (attempt ${retryCount + 1}/${maxRetries}):`, err);
                    
                    if (retryCount < maxRetries - 1) {
                        retryCount++;
                        console.log(`Retrying save... (${retryCount}/${maxRetries})`);
                        
                        // 增加延迟时间，避免立即重试
                        const delay = 500 * Math.pow(2, retryCount);
                        await new Promise(resolve => setTimeout(resolve, delay));
                        return await saveWithRetry();
                    } else {
                        message.error('保存AI回复失败，但不影响当前对话');
                        throw err;
                    }
                }
            };
            
            await saveWithRetry();
        } catch (err) {
            console.error('最终错误: 保存AI响应失败:', err);
            message.error('无法保存AI回复，但不影响当前对话');
        }
    };

    const handleTitleEdit = async (chatId, newTitle) => {
        try {
            await request.put(`${API_BASE_URL}/api/chat/${chatId}/title`, { title: newTitle });
            await fetchChatHistory();
            setEditingTitle(null);
        } catch (err) {
            console.error('Error updating title:', err);
            message.error('Failed to update title');
        }
    };

    // Add update chat title function
    const handleUpdateChatTitle = async (chatId, title) => {
        try {
            await request.put(`${API_BASE_URL}/api/chat/${chatId}/title`, { title });
            await fetchChatHistory();
            setEditingTitle(null);
        } catch (err) {
            console.error('Error updating chat title:', err);
            message.error('Failed to update chat title');
        }
    };

    // Add delete chat function
    const handleDeleteChat = async (chatId) => {
        try {
            await request.delete(`${API_BASE_URL}/api/chat/${chatId}`);
            
            // If deleting is current chat, clear current chat
            if (chatId === currentChatId) {
                setCurrentChatId(null);
                localStorage.removeItem('currentChatId');
                setCurrentChat([]);
            }
            
            await fetchChatHistory();
            message.success('Chat deleted successfully');
        } catch (err) {
            console.error('Error deleting chat:', err);
            message.error('Failed to delete chat');
        } finally {
            setDeleteDialogOpen(false);
            setChatToDelete(null);
        }
    };

    // Open delete confirmation dialog
    const openDeleteDialog = (chatId, e) => {
        e.stopPropagation(); // Stop event bubbling
        setChatToDelete(chatId);
        setDeleteDialogOpen(true);
    };

    // Render message content
    const renderMessageContent = (message) => {
        if (message.role === 'assistant') {
            // If it's an AI message with empty content, show loading indicator
            if (message.content === '' && message.isStreaming) {
                return (
                    <div className="ai-message-loading">
                        <div className="typing-indicator">
                            <span></span>
                            <span></span>
                            <span></span>
                        </div>
                    </div>
                );
            }
            
            // Use ReactMarkdown to render Markdown content
            return (
                <div className="markdown-content" style={{ maxWidth: '100%', wordWrap: 'break-word' }}>
                    <ReactMarkdown 
                        remarkPlugins={[remarkGfm]}
                        components={{
                            a: ({node, ...props}) => <a style={{color: '#1890ff'}} target="_blank" rel="noopener noreferrer" {...props} />,
                            h1: ({node, ...props}) => <h1 style={{marginTop: '12px', marginBottom: '6px', fontWeight: 'bold'}} {...props} />,
                            h2: ({node, ...props}) => <h2 style={{marginTop: '10px', marginBottom: '6px', fontWeight: 'bold'}} {...props} />,
                            h3: ({node, ...props}) => <h3 style={{marginTop: '8px', marginBottom: '4px', fontWeight: 'bold'}} {...props} />,
                            h4: ({node, ...props}) => <h4 style={{marginTop: '8px', marginBottom: '4px', fontWeight: 'bold'}} {...props} />,
                            h5: ({node, ...props}) => <h5 style={{marginTop: '6px', marginBottom: '4px', fontWeight: 'bold'}} {...props} />,
                            h6: ({node, ...props}) => <h6 style={{marginTop: '6px', marginBottom: '4px', fontWeight: 'bold'}} {...props} />,
                            code: ({node, inline, className, children, ...props}) => {
                                const match = /language-(\w+)/.exec(className || '');
                                return !inline && match ? (
                                    <div style={{position: 'relative', marginBottom: '8px'}}>
                                        <pre style={{
                                            backgroundColor: 'rgba(0, 0, 0, 0.05)', 
                                            padding: '12px', 
                                            borderRadius: '4px', 
                                            overflowX: 'auto',
                                            marginBottom: 0
                                        }}>
                                            <code className={className} {...props}>
                                                {children}
                                            </code>
                                        </pre>
                                    </div>
                                ) : (
                                    <code style={{
                                        backgroundColor: 'rgba(0, 0, 0, 0.05)', 
                                        padding: '2px 4px', 
                                        borderRadius: '3px'
                                    }} {...props}>
                                        {children}
                                    </code>
                                );
                            },
                            pre: ({node, ...props}) => <pre style={{margin: 0}} {...props} />,
                            blockquote: ({node, ...props}) => <blockquote style={{borderLeft: '4px solid #ddd', paddingLeft: '16px', margin: '0', marginBottom: '8px'}} {...props} />,
                            ul: ({node, ...props}) => <ul style={{paddingLeft: '24px', marginTop: '4px', marginBottom: '8px'}} {...props} />,
                            ol: ({node, ...props}) => <ol style={{paddingLeft: '24px', marginTop: '4px', marginBottom: '8px'}} {...props} />,
                            li: ({node, ...props}) => <li style={{marginBottom: '2px'}} {...props} />,
                            img: ({node, ...props}) => <img style={{maxWidth: '100%', marginBottom: '8px'}} {...props} />,
                            p: ({node, ...props}) => <p style={{marginTop: '4px', marginBottom: '8px', lineHeight: '1.5'}} {...props} />,
                            table: ({node, ...props}) => <div style={{overflowX: 'auto', marginBottom: '8px'}}><table style={{borderCollapse: 'collapse', width: '100%'}} {...props} /></div>,
                            th: ({node, ...props}) => <th style={{border: '1px solid #ddd', padding: '8px', backgroundColor: 'rgba(0, 0, 0, 0.05)'}} {...props} />,
                            td: ({node, ...props}) => <td style={{border: '1px solid #ddd', padding: '8px'}} {...props} />
                        }}
                    >
                        {message.content}
                    </ReactMarkdown>
                </div>
            );
        }
        
        // User messages are displayed as plain text, preserving spaces and line breaks
        return <div style={{whiteSpace: 'pre-wrap'}}>{message.content}</div>;
    };

    // 添加函数获取可用模型
    const fetchAvailableModels = async () => {
        try {
            const response = await request.get(`${API_BASE_URL}/api/chat/models`);
            if (response && Array.isArray(response)) {
                // 按照提供商分组模型
                const openAIModels = response.filter(model => model.provider === 'openai')
                    .map(model => ({ value: model.id, label: model.name }));
                
                const anthropicModels = response.filter(model => model.provider === 'anthropic')
                    .map(model => ({ value: model.id, label: model.name }));
                
                const modelOptions = [];
                
                if (openAIModels.length > 0) {
                    modelOptions.push({
                        label: 'OpenAI',
                        options: openAIModels
                    });
                }
                
                if (anthropicModels.length > 0) {
                    modelOptions.push({
                        label: 'Anthropic',
                        options: anthropicModels
                    });
                }
                
                setAvailableModels(modelOptions);
                console.log('Available models loaded:', modelOptions);
            }
        } catch (error) {
            console.error('Error fetching available models:', error);
            // 如果API获取失败，使用硬编码的备用选项
            setAvailableModels([
                {
                    label: 'OpenAI',
                    options: [
                        { value: 'gpt-4o', label: 'GPT-4o' },
                        { value: 'gpt-4', label: 'GPT-4' },
                        { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
                        { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
                    ]
                },
                {
                    label: 'Anthropic',
                    options: [
                        { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet' },
                        { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus' },
                    ]
                }
            ]);
        }
    };

    // 在组件加载时获取可用模型
    useEffect(() => {
        fetchAvailableModels();
    }, []);

    // 添加函数让当前选中的聊天记录保持可见
    const scrollToCurrentChat = () => {
        if (currentChatId && chatHistoryListRef.current) {
            const activeElement = chatHistoryListRef.current.querySelector('.chat-history-item.active');
            if (activeElement) {
                activeElement.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        }
    };

    // 在聊天记录加载或当前聊天ID变化时调用滚动函数
    useEffect(() => {
        scrollToCurrentChat();
    }, [currentChatId, chatHistory]);

    return (
        <Layout style={{ 
            height: 'calc(100vh - 64px)',
            backgroundColor: '#fff'
        }}>
            {/* Add inline styles */}
            <style>{typingIndicatorStyle}</style>
            
            <Sider width={300} theme="light" style={{ 
                borderRight: '1px solid #f0f0f0',
                height: '100%',
                position: 'fixed',
                left: 0,
                top: 64,
                backgroundColor: '#fff',
                zIndex: 1,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden'
            }}>
                {/* 固定在顶部的选择器和按钮 */}
                <div style={{ 
                    padding: '20px', 
                    borderBottom: '1px solid #f0f0f0',
                    flexShrink: 0 // 防止内容压缩
                }}>
                    <Select
                        value={modelLocked ? currentChatModel : selectedModel}
                        onChange={setSelectedModel}
                        style={{ width: '100%', marginBottom: '10px' }}
                        disabled={modelLocked}
                        options={availableModels}
                    />
                    
                    {/* 语言偏好选择器 */}
                    <Select
                        value={languagePreference}
                        onChange={setLanguagePreference}
                        style={{ width: '100%', marginBottom: '10px' }}
                        options={[
                            { value: 'english', label: 'English' },
                            { value: 'chinese', label: '中文' },
                            { value: 'auto', label: 'Auto Detect' }
                        ]}
                    />
                    
                    <Button 
                        type="primary" 
                        icon={<PlusOutlined />} 
                        onClick={handleCreateChat}
                        loading={loading}
                        block
                    >
                        New Chat
                    </Button>
                </div>
                
                {/* 可滚动的聊天列表区域 */}
                <div 
                    className="chat-history-list" 
                    ref={chatHistoryListRef}
                    style={{ 
                        flex: 1,
                        overflow: 'auto',
                        paddingTop: '8px',
                        paddingBottom: '8px',
                        position: 'relative'
                    }}
                >
                    {chatHistory.length > 0 ? (
                        <List
                            dataSource={chatHistory}
                            renderItem={chat => (
                                <List.Item 
                                    key={chat.id}
                                    onClick={() => fetchChatMessages(chat.id)}
                                    style={{ 
                                        cursor: 'pointer', 
                                        padding: '10px 20px',
                                        backgroundColor: chat.id === currentChatId ? '#e6f7ff' : 'transparent',
                                        borderLeft: chat.id === currentChatId ? '3px solid #1890ff' : '3px solid transparent',
                                        transition: 'all 0.3s',
                                    }}
                                    className={`chat-history-item ${chat.id === currentChatId ? 'active' : ''}`}
                                >
                                    <div style={{ width: '100%', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                        <div style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                            {editingTitle === chat.id ? (
                                                <Input 
                                                    value={newTitle} 
                                                    onChange={e => setNewTitle(e.target.value)}
                                                    onPressEnter={() => handleUpdateChatTitle(chat.id, newTitle)}
                                                    onBlur={() => setEditingTitle(null)}
                                                    autoFocus
                                                    size="small"
                                                    style={{ width: '80%' }}
                                                />
                                            ) : (
                                                <span 
                                                    onDoubleClick={(e) => {
                                                        e.stopPropagation();
                                                        setEditingTitle(chat.id);
                                                        setNewTitle(chat.title || 'New Chat');
                                                    }}
                                                    style={{
                                                        color: chat.id === currentChatId ? '#1890ff' : 'rgba(0, 0, 0, 0.85)',
                                                        fontWeight: chat.id === currentChatId ? '500' : 'normal',
                                                    }}
                                                >
                                                    {chat.title || 'New Chat'}
                                                </span>
                                            )}
                                        </div>
                                        <div style={{ display: 'flex', alignItems: 'center' }}>
                                            <div style={{ color: '#999', fontSize: '12px', marginRight: '10px' }}>
                                                {dayjs(chat.created_at).format('MM-DD HH:mm')}
                                            </div>
                                            <Button 
                                                type="text" 
                                                icon={<EditOutlined />} 
                                                size="small"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    setEditingTitle(chat.id);
                                                    setNewTitle(chat.title || 'New Chat');
                                                }}
                                            />
                                            <Button 
                                                type="text" 
                                                icon={<DeleteOutlined />} 
                                                size="small"
                                                onClick={(e) => openDeleteDialog(chat.id, e)}
                                            />
                                        </div>
                                    </div>
                                </List.Item>
                            )}
                        />
                    ) : (
                        <div style={{ textAlign: 'center', padding: '30px 0', color: '#999' }}>
                            没有聊天记录
                        </div>
                    )}
                </div>
            </Sider>
            
            {/* Delete confirmation dialog */}
            <Modal
                title="Confirm Delete"
                open={deleteDialogOpen}
                onOk={() => handleDeleteChat(chatToDelete)}
                onCancel={() => {
                    setDeleteDialogOpen(false);
                    setChatToDelete(null);
                }}
                okText="Delete"
                cancelText="Cancel"
            >
                <p>Are you sure you want to delete this chat? This action cannot be undone.</p>
            </Modal>
            
            <Layout style={{ marginLeft: 300, height: '100%' }}>
                <Content style={{ 
                    padding: '20px',
                    display: 'flex',
                    flexDirection: 'column',
                    height: '100%',
                    maxWidth: 'calc(100% - 40px)', // 减去左右padding
                    overflow: 'hidden'
                }}>
                    <div 
                        ref={messagesContainerRef}
                        style={{ 
                            flex: 1, 
                            overflowY: 'auto',
                            padding: '10px 0',
                            marginBottom: '10px',
                            maxWidth: '100%'
                        }}
                    >
                        {currentChat && currentChat.length > 0 ? (
                            <List
                                itemLayout="horizontal"
                                dataSource={currentChat}
                                style={{
                                    width: '100%',
                                    overflow: 'hidden',
                                    wordWrap: 'break-word'
                                }}
                                renderItem={msg => (
                                    <List.Item style={{ 
                                        padding: '6px 0',
                                        borderBottom: 'none',
                                        width: '100%',
                                        overflow: 'hidden',
                                        wordWrap: 'break-word'
                                    }}>
                                        <List.Item.Meta
                                            avatar={
                                                <Avatar 
                                                    style={{ 
                                                        backgroundColor: msg.role === 'user' ? '#1890ff' : '#52c41a',
                                                        verticalAlign: 'middle',
                                                        display: 'flex',
                                                        justifyContent: 'center',
                                                        alignItems: 'center',
                                                        fontSize: '18px'
                                                    }}
                                                >
                                                    {msg.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
                                                </Avatar>
                                            }
                                            description={
                                                <div style={{ 
                                                    whiteSpace: 'pre-wrap', 
                                                    color: 'rgba(0, 0, 0, 0.85)',
                                                    fontSize: '14px',
                                                    wordBreak: 'break-word',
                                                    maxWidth: '100%',
                                                    overflowWrap: 'break-word',
                                                    wordWrap: 'break-word',
                                                    marginTop: '0'
                                                }}>
                                                    {renderMessageContent(msg)}
                                                </div>
                                            }
                                        />
                                    </List.Item>
                                )}
                            />
                        ) : (
                            <div style={{ 
                                display: 'flex', 
                                justifyContent: 'center', 
                                alignItems: 'center', 
                                height: '100%',
                                color: '#999'
                            }}>
                                Select or create a chat to start
                            </div>
                        )}
                        {/* Add a reference point for scrolling to the bottom */}
                        <div ref={messagesEndRef} />
                    </div>
                    
                    <div style={{ 
                        display: 'flex',
                        padding: '10px 0',
                        borderTop: '1px solid #f0f0f0',
                        backgroundColor: '#fff'
                    }}>
                        <TextArea
                            value={inputMessage}
                            onChange={e => setInputMessage(e.target.value)}
                            placeholder="Enter message, press Enter to send"
                            autoSize={{ minRows: 2, maxRows: 6 }}
                            onPressEnter={e => {
                                if (!e.shiftKey) {
                                    e.preventDefault();
                                    handleSend();
                                }
                            }}
                            style={{ 
                                flex: 1, 
                                marginRight: '10px',
                                borderRadius: '4px'
                            }}
                        />
                        <Button 
                            type="primary" 
                            icon={<SendOutlined />} 
                            onClick={handleSend}
                            loading={loading}
                            style={{
                                height: '40px'
                            }}
                        />
                    </div>
                </Content>
            </Layout>
        </Layout>
    );
};

export default ChatPage;