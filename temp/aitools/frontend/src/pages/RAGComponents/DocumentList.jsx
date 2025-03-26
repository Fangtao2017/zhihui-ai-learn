import React, { useState, useRef, useEffect } from 'react';
import { 
  Box, Typography, List, ListItem, ListItemText, ListItemSecondaryAction, ListItemAvatar,
  IconButton, Divider, Button, Dialog, DialogActions, DialogContent, 
  DialogContentText, DialogTitle, CircularProgress, Alert, Tooltip, Avatar, Chip,
  Tab, Tabs, Card, CardContent, Paper
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import DescriptionOutlinedIcon from '@mui/icons-material/DescriptionOutlined';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import ArticleIcon from '@mui/icons-material/Article';
import CodeIcon from '@mui/icons-material/Code';
import PsychologyIcon from '@mui/icons-material/Psychology';
import DownloadIcon from '@mui/icons-material/Download';
import SaveAltIcon from '@mui/icons-material/SaveAlt';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';
import axios from 'axios';
import html2pdf from 'html2pdf.js';
import { styled, keyframes } from '@mui/material/styles';

import './markdown-styles.css';

// 添加关键帧动画
const pulse = keyframes`
  0% {
    box-shadow: 0 0 0 0 rgba(43, 122, 11, 0.4);
  }
  70% {
    box-shadow: 0 0 0 20px rgba(43, 122, 11, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(43, 122, 11, 0);
  }
`;

// 样式化组件
const PulseBox = styled(Box)`
  animation: ${pulse} 2s infinite;
`;

const DocumentList = ({ documents, loading, error, onDocumentDeleted, onDocumentReprocessed }) => {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedDocId, setSelectedDocId] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState(null);
  
  const [multiAgentDialogOpen, setMultiAgentDialogOpen] = useState(false);
  const [selectedDocForMultiAgent, setSelectedDocForMultiAgent] = useState(null);
  const [selectedDocName, setSelectedDocName] = useState("");
  const [multiAgentLoading, setMultiAgentLoading] = useState(false);
  const [multiAgentResult, setMultiAgentResult] = useState(null);
  const [multiAgentTabValue, setMultiAgentTabValue] = useState(0);
  const [multiAgentProcessingSteps] = useState([
    { id: 'content', label: "Content Analysis", description: "Identify document structure and sections", icon: "📑", color: "#2B7A0B" },
    { id: 'knowledge', label: "Knowledge Extraction", description: "Extract key concepts and definitions", icon: "🔍", color: "#72A764" },
    { id: 'summary', label: "Content Summary", description: "Generate section summaries", icon: "🔑", color: "#507D45" },
    { id: 'format', label: "Formatting", description: "Format as Markdown and Anki cards", icon: "✨", color: "#345F29" }
  ]);
  const [activeStep, setActiveStep] = useState(0);
  const [stepsCompleted, setStepsCompleted] = useState({});

  // 增加处理状态跟踪
  const [processingState, setProcessingState] = useState({
    started: false,
    contentAnalysis: false,
    knowledgeExtraction: false,
    summarization: false,
    formatting: false,
    completed: false,
    startTime: null
  });

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return 'Unknown';
    
    try {
      const date = new Date(dateString);
      // Check if date is valid
      if (isNaN(date.getTime())) {
        return 'Invalid date';
      }
      return date.toLocaleString('en-US');
    } catch (error) {
      console.error('Date formatting error:', error);
      return 'Invalid date';
    }
  };

  // Get file icon
  const getFileIcon = (filename) => {
    if (!filename) return <DescriptionOutlinedIcon />;
    
    const extension = filename.split('.').pop().toLowerCase();
    
    switch (extension) {
      case 'pdf':
        return <PictureAsPdfIcon style={{ color: '#f44336' }} />;
      case 'txt':
        return <TextSnippetIcon style={{ color: '#2196f3' }} />;
      case 'docx':
      case 'doc':
        return <ArticleIcon style={{ color: '#1976d2' }} />;
      case 'md':
        return <TextSnippetIcon style={{ color: '#9c27b0' }} />;
      default:
        return <DescriptionOutlinedIcon style={{ color: '#757575' }} />;
    }
  };

  // Open delete confirmation dialog
  const handleOpenDeleteDialog = (docId) => {
    console.log(`Opening delete dialog, document ID: ${docId}`);
    setSelectedDocId(docId);
    setDeleteDialogOpen(true);
  };

  // Close delete confirmation dialog
  const handleCloseDeleteDialog = () => {
    setDeleteDialogOpen(false);
    setSelectedDocId(null);
  };

  // Delete document
  const handleDeleteDocument = async () => {
    if (!selectedDocId) {
      console.error('No document ID selected for deletion');
      return;
    }
    
    console.log(`Starting document deletion: ${selectedDocId}`);
    setActionLoading(true);
    setActionError(null);
    
    try {
      const token = localStorage.getItem('token');
      console.log(`Using token: ${token ? token.substring(0, 10) + '...' : 'null'}`);
      console.log(`Sending delete request to: http://localhost:8080/api/rag/document/${selectedDocId}`);
      
      const response = await axios.delete(`http://localhost:8080/api/rag/document/${selectedDocId}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      console.log('Delete request successful:', response.data);
      handleCloseDeleteDialog();
      if (onDocumentDeleted) {
        console.log('Calling onDocumentDeleted callback');
        onDocumentDeleted();
      }
    } catch (err) {
      console.error('Delete document failed:', err);
      console.error('Error details:', err.response ? err.response.data : 'No response data');
      setActionError('Failed to delete document, please try again later');
    } finally {
      setActionLoading(false);
    }
  };

  // Reprocess document
  const handleReprocessDocument = async (docId) => {
    setActionLoading(true);
    setActionError(null);
    
    try {
      const token = localStorage.getItem('token');
      await axios.post(`http://localhost:8080/api/rag/document/${docId}/reprocess`, {}, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      if (onDocumentReprocessed) onDocumentReprocessed();
    } catch (err) {
      console.error('Failed to reprocess document:', err);
      setActionError('Failed to reprocess document, please try again later');
    } finally {
      setActionLoading(false);
    }
  };

  // 处理多Agent分析
  const handleProcessWithMultiAgent = (docId, docName) => {
    console.log(`Processing document ${docId} using Multi-Agent system`);
    setSelectedDocForMultiAgent(docId);
    setSelectedDocName(docName);
    setMultiAgentDialogOpen(true);
  };

  // 检查当前进度状态，更新activeStep
  useEffect(() => {
    if (multiAgentLoading) {
      const checkProgress = async () => {
        try {
          // 根据后端输出的关键词判断当前处理阶段
          // 目前没有API可用，因此根据日志查看处理状态

          // 假设后台处理需要一定时间，我们创建一个更逼真的模拟
          let newActiveStep = 0;
          
          if (processingState.started) {
            newActiveStep = 0; // 内容分析
            
            // 定义步骤持续时间（单位：秒）
            const stepDurations = [5, 6, 4, 3];
            const totalDuration = stepDurations.reduce((a, b) => a + b, 0);
            const stepStartTimes = [0];
            
            // 计算每个步骤的开始时间
            for (let i = 1; i < stepDurations.length; i++) {
              stepStartTimes[i] = stepStartTimes[i-1] + stepDurations[i-1];
            }
            
            // 计算经过的时间（模拟）
            const startAt = processingState.startTime || Date.now();
            const elapsed = (Date.now() - startAt) / 1000; // 转换为秒
            
            // 根据经过的时间决定当前阶段
            if (elapsed < stepDurations[0]) {
              // 第一阶段：内容分析
              newActiveStep = 0;
              if (!processingState.contentAnalysis && elapsed > stepDurations[0] * 0.5) {
                // 进行中的步骤设置为已完成
                setStepsCompleted(prev => ({...prev, 0: true}));
                setProcessingState(prev => ({...prev, contentAnalysis: true}));
              }
            } else if (elapsed < stepStartTimes[1] + stepDurations[1]) {
              // 第二阶段：知识提取
              newActiveStep = 1;
              if (!processingState.knowledgeExtraction) {
                setProcessingState(prev => ({...prev, knowledgeExtraction: true}));
                setStepsCompleted(prev => ({...prev, 0: true}));
              }
              if (elapsed > stepStartTimes[1] + stepDurations[1] * 0.7) {
                setStepsCompleted(prev => ({...prev, 1: true}));
              }
            } else if (elapsed < stepStartTimes[2] + stepDurations[2]) {
              // 第三阶段：总结生成
              newActiveStep = 2;
              if (!processingState.summarization) {
                setProcessingState(prev => ({...prev, summarization: true}));
                setStepsCompleted(prev => ({...prev, 0: true, 1: true}));
              }
              if (elapsed > stepStartTimes[2] + stepDurations[2] * 0.6) {
                setStepsCompleted(prev => ({...prev, 2: true}));
              }
            } else if (elapsed < totalDuration) {
              // 第四阶段：格式排版
              newActiveStep = 3;
              if (!processingState.formatting) {
                setProcessingState(prev => ({...prev, formatting: true}));
                setStepsCompleted(prev => ({...prev, 0: true, 1: true, 2: true}));
              }
              if (elapsed > totalDuration * 0.95) {
                setStepsCompleted(prev => ({...prev, 3: true}));
              }
            }
          }
          
          setActiveStep(newActiveStep);
        } catch (err) {
          console.error('检查处理进度失败:', err);
        }
      };
      
      // 设置开始时间（如果尚未设置）
      if (processingState.started && !processingState.startTime) {
        setProcessingState(prev => ({...prev, startTime: Date.now()}));
      }
      
      // 创建进度轮询
      const intervalId = setInterval(checkProgress, 500);
      
      // 清理函数
      return () => {
        clearInterval(intervalId);
      };
    }
  }, [multiAgentLoading, processingState, multiAgentProcessingSteps]);

  // 确认使用多Agent处理 - 修改使用轮询方式
  const handleConfirmMultiAgentProcess = async () => {
    if (!selectedDocForMultiAgent) {
      console.error('No document selected for Multi-Agent processing');
      return;
    }
    
    // 重置进度状态
    setMultiAgentLoading(true);
    setActionError(null);
    setActiveStep(0);
    setStepsCompleted({});
    
    // 重置处理状态
    setProcessingState({
      started: true,
      contentAnalysis: false,
      knowledgeExtraction: false,
      summarization: false,
      formatting: false,
      completed: false,
      startTime: Date.now() // 直接设置开始时间
    });
    
    console.log('Processing document ID:', selectedDocForMultiAgent);
    
    try {
      // 开始处理请求
      const response = await axios.post(`http://localhost:8081/document/${selectedDocForMultiAgent}/multi-agent`, {});
      
      console.log('Multi-Agent processing result:', response.data);
      
      // 确保所有步骤显示为已完成
      const allCompleted = {};
      multiAgentProcessingSteps.forEach((_, index) => {
        allCompleted[index] = true;
      });
      setStepsCompleted(allCompleted);
      
      // 设置处理状态为完成
      setProcessingState(prev => ({...prev, completed: true}));
      
      // 显示结果
      setMultiAgentResult(response.data);
      setMultiAgentLoading(false);
      
    } catch (err) {
      console.error('Multi-Agent processing failed:', err);
      setActionError('Multi-Agent processing failed, please try again later');
      setStepsCompleted({});
      setActiveStep(0);
      setMultiAgentLoading(false);
      setProcessingState({
        started: false,
        contentAnalysis: false,
        knowledgeExtraction: false,
        summarization: false,
        formatting: false,
        completed: false,
        startTime: null
      });
    }
  };

  // 关闭多Agent对话框
  const handleCloseMultiAgentDialog = () => {
    setMultiAgentDialogOpen(false);
    setSelectedDocForMultiAgent(null);
    setSelectedDocName("");
    setMultiAgentResult(null);
    setMultiAgentTabValue(0);
  };

  // 处理多Agent结果选项卡切换
  const handleMultiAgentTabChange = (event, newValue) => {
    setMultiAgentTabValue(newValue);
  };

  // 渲染多Agent结果
  const renderMultiAgentResultContent = () => {
    if (!multiAgentResult) {
      return null;
    }

    switch(multiAgentTabValue) {
      case 0: // Markdown内容
        return (
          <div>
            {/* Export buttons row */}
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', gap: 1, mb: 2 }}>
              <Button
                variant="outlined"
                startIcon={<TextSnippetIcon />}
                onClick={handleDownloadMarkdown}
                disabled={!multiAgentResult.markdown_content}
                size="small"
                sx={{ 
                  color: '#2B7A0B', 
                  borderColor: '#2B7A0B',
                  '&:hover': { 
                    borderColor: '#1d5407',
                    backgroundColor: 'rgba(43, 122, 11, 0.04)'
                  }
                }}
              >
                Download Markdown
              </Button>
              <Button
                variant="contained"
                startIcon={<PictureAsPdfIcon />}
                onClick={handleExportToPDF}
                disabled={!multiAgentResult.markdown_content || multiAgentLoading}
                size="small"
                sx={{ 
                  bgcolor: '#2B7A0B', 
                  '&:hover': { bgcolor: '#1d5407' } 
                }}
              >
                Export to PDF
              </Button>
            </Box>
            
            {/* Markdown content display area */}
            <Paper 
              elevation={1}
              sx={{ 
                p: 3, 
                bgcolor: '#ffffff',
                borderRadius: 1, 
                height: 'calc(70vh - 120px)',
                overflow: 'auto',
                overflowX: 'hidden',
                '& h1': {
                  marginTop: '0.5em',
                  paddingTop: 0,
                  scrollMarginTop: '2em' // Top margin when scrolling to heading
                },
                '& h1, & h2, & h3, & h4, & h5, & h6': {
                  marginTop: '1em',
                  marginBottom: '0.5em',
                  fontWeight: 600,
                  color: '#333',
                  scrollMarginTop: '2em', // Top margin when scrolling to heading
                  '& a': {
                    color: 'inherit',
                    textDecoration: 'none',
                    '&:hover': {
                      textDecoration: 'none',
                      color: 'primary.main'
                    },
                    '&::before': {
                      content: '"#"',
                      position: 'absolute',
                      left: '-1em',
                      color: 'transparent',
                      fontWeight: 'bold'
                    },
                    '&:hover::before': {
                      color: 'primary.main'
                    }
                  }
                },
                '& p': {
                  marginBottom: '1em'
                },
                '& ul, & ol': {
                  paddingLeft: '1.5em',
                  marginBottom: '1em'
                },
                '& blockquote': {
                  borderLeft: '4px solid #e0e0e0',
                  paddingLeft: '1em',
                  fontStyle: 'italic',
                  margin: '1em 0',
                  color: '#555'
                },
                '& hr': {
                  margin: '1em 0'
                },
                '& code': {
                  backgroundColor: '#f5f5f5',
                  padding: '0.2em 0.4em',
                  borderRadius: '3px',
                  fontFamily: 'monospace'
                },
                '& img': {
                  maxWidth: '100%',
                  height: 'auto'
                },
                '& table': {
                  borderCollapse: 'collapse',
                  width: '100%',
                  overflowX: 'auto',
                  display: 'block'
                },
                '& th, & td': {
                  border: '1px solid #ddd',
                  padding: '8px'
                }
              }}
              className="markdown-content"
              id="markdown-content"
            >
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                rehypePlugins={[rehypeSlug, [rehypeAutolinkHeadings, { behavior: 'wrap' }]]}
              >
                {multiAgentResult.markdown_content || "No Markdown content generated"}
              </ReactMarkdown>
            </Paper>

            {/* Hidden print area - for PDF export */}
            <div id="print-container" style={{ display: 'none' }}>
              <div id="printable-content" style={{ 
                padding: '20px', 
                fontFamily: '"Noto Sans SC", "Roboto", sans-serif',
                lineHeight: '1.6',
                color: '#333'
              }}>
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  rehypePlugins={[rehypeSlug]}
                >
                  {multiAgentResult.markdown_content || "No Markdown content generated"}
                </ReactMarkdown>
              </div>
            </div>
          </div>
        );
      case 1: // Anki卡片
        return (
          <Box sx={{ height: 'calc(70vh - 120px)', overflow: 'auto' }}>
            {multiAgentResult.anki_cards && multiAgentResult.anki_cards.length > 0 ? (
              multiAgentResult.anki_cards.map((card, index) => (
                <Card key={index} sx={{ mb: 2, border: '1px solid rgba(0,0,0,0.1)' }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      {card.front}
                    </Typography>
                    <Divider sx={{ my: 1 }} />
                    <Typography variant="body2">
                      {card.back}
                    </Typography>
                  </CardContent>
                </Card>
              ))
            ) : (
              <Typography variant="body2" color="text.secondary">
                No Anki cards generated
              </Typography>
            )}
          </Box>
        );
      case 2: // 处理详情
        return (
          <Box sx={{ height: 'calc(70vh - 120px)', overflow: 'auto' }}>
            <Typography variant="subtitle2" gutterBottom>
              Document Information:
            </Typography>
            <Typography variant="body2" paragraph>
              Document ID: {multiAgentResult.doc_id}<br />
              File Name: {multiAgentResult.file_name}<br />
              Processing Time: {new Date(multiAgentResult.processed_at).toLocaleString('en-US')}<br />
              Status: {multiAgentResult.status}
            </Typography>
            
            <Typography variant="subtitle2" gutterBottom>
              Agent Processing Results:
            </Typography>
            <Typography variant="body2" component="div">
              <ul>
                <li>Content Analysis Agent: {multiAgentResult.content_agent.status}</li>
                <li>Knowledge Extraction Agent: {multiAgentResult.knowledge_agent.status}</li>
                <li>Summary Agent: {multiAgentResult.summary_agent.status}</li>
                <li>Formatting Agent: {multiAgentResult.format_agent.status}</li>
              </ul>
            </Typography>
            
            {multiAgentResult.error && (
              <Alert severity="error" sx={{ mt: 2 }}>
                Error: {multiAgentResult.error}
              </Alert>
            )}
          </Box>
        );
      default:
        return null;
    }
  };

  // 下载Markdown内容
  const handleDownloadMarkdown = () => {
    if (!multiAgentResult || !multiAgentResult.markdown_content) return;
    
    const blob = new Blob([multiAgentResult.markdown_content || ""], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${selectedDocName.replace(/\.[^/.]+$/, '')}_notes.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  // 导出PDF功能
  const handleExportToPDF = () => {
    if (!multiAgentResult || !multiAgentResult.markdown_content) return;
    
    setMultiAgentLoading(true);
    
    setTimeout(() => {
      try {
        // 创建一个新窗口
        const printWindow = window.open('', '_blank');
        
        if (!printWindow) {
          alert('Please allow pop-ups to export PDF');
          setMultiAgentLoading(false);
          return;
        }
        
        // 写入HTML内容
        printWindow.document.write(`
          <!DOCTYPE html>
          <html>
          <head>
            <title>${selectedDocName} - Learning Notes</title>
            <style>
              body {
                font-family: "Noto Sans SC", "Roboto", sans-serif;
                line-height: 1.6;
                color: #333;
                padding: 20px;
                margin: 0;
              }
              h1, h2, h3, h4, h5, h6 {
                margin-top: 1.5em;
                margin-bottom: 0.5em;
                font-weight: 600;
                color: #333;
              }
              h1 { font-size: 2em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
              h2 { font-size: 1.5em; border-bottom: 1px solid #eee; padding-bottom: 0.3em; }
              h3 { font-size: 1.25em; }
              p { margin-bottom: 1em; }
              ul, ol { padding-left: 2em; margin-bottom: 1em; }
              blockquote {
                border-left: 4px solid #e0e0e0;
                padding-left: 1em;
                font-style: italic;
                margin: 1em 0;
                color: #555;
              }
              hr { margin: 1.5em 0; border: none; border-top: 1px solid #eee; }
              code {
                background-color: #f5f5f5;
                padding: 0.2em 0.4em;
                border-radius: 3px;
                font-family: monospace;
              }
              pre {
                background-color: #f6f8fa;
                padding: 16px;
                overflow: auto;
                border-radius: 3px;
                margin-bottom: 16px;
              }
              table {
                width: 100%;
                border-collapse: collapse;
                margin-bottom: 16px;
              }
              table th, table td {
                padding: 6px 13px;
                border: 1px solid #dfe2e5;
              }
              table th {
                background-color: #f6f8fa;
                font-weight: 600;
              }
              @media print {
                body { padding: 0; }
                @page { margin: 1.5cm; }
              }
            </style>
          </head>
          <body>
            ${document.getElementById('markdown-content').innerHTML}
          </body>
          </html>
        `);
        
        // 关闭document写入
        printWindow.document.close();
        
        // 等待内容加载完成后打印
        printWindow.onload = function() {
          printWindow.print();
          setMultiAgentLoading(false);
        };
        
      } catch (err) {
        console.error('PDF导出失败:', err);
        setActionError('PDF导出失败，请稍后重试');
        setMultiAgentLoading(false);
      }
    }, 100);
  };

  // 渲染简化的进度指示器组件 - 单一状态
  const renderProgressStepper = () => {
    return (
      <Box sx={{ 
        width: '100%', 
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 2
      }}>
        {/* 中心加载动画 */}
        <Box sx={{ position: 'relative' }}>
          {/* 旋转外圈 */}
          <Box sx={{
            width: 90,
            height: 90,
            borderRadius: '50%',
            border: '3px solid #f3f3f3',
            borderTop: '3px solid #2B7A0B',
            borderRight: '3px solid #72A764',
            borderBottom: '3px solid #507D45',
            borderLeft: '3px solid #345F29',
            animation: 'spin 1.5s linear infinite',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
          }}>
            {/* 内圈脉动图标 */}
            <Box sx={{
              width: 60,
              height: 60,
              borderRadius: '50%',
              backgroundColor: 'rgba(43, 122, 11, 0.1)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              animation: 'pulse 1.5s ease-in-out infinite',
              fontSize: '32px'
            }}>
              <PsychologyIcon sx={{ fontSize: 32, color: '#2B7A0B' }} />
            </Box>
          </Box>
        </Box>
        
        {/* 当前状态文本 */}
        <Typography variant="subtitle1" sx={{ mt: 2, mb: 0.5, textAlign: 'center', fontWeight: 'medium' }}>
          Analyzing
        </Typography>
        
        <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', maxWidth: '80%' }}>
          The system is performing intelligent analysis on the document, please be patient
        </Typography>
        
        {/* 添加动画关键帧 */}
        <style jsx global>{`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
          
          @keyframes pulse {
            0% { transform: scale(0.95); opacity: 0.8; }
            50% { transform: scale(1.05); opacity: 1; }
            100% { transform: scale(0.95); opacity: 0.8; }
          }
        `}</style>
      </Box>
    );
  };

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <Typography variant="subtitle1" gutterBottom sx={{ mb: 1, fontWeight: 'medium' }}>
        Document List
      </Typography>
      
      {/* Add debug information */}
      <Box sx={{ mb: 2, p: 2, bgcolor: 'background.paper', borderRadius: 1, display: 'none' }}>
        <Typography variant="subtitle2">Debug Info:</Typography>
        <pre style={{ overflow: 'auto', maxHeight: '200px' }}>
          {documents && documents.length > 0 ? JSON.stringify(documents[0], null, 2) : 'No documents'}
        </pre>
      </Box>
      
      {error && <Alert severity="error" sx={{ mb: 1 }}>{error}</Alert>}
      {actionError && <Alert severity="error" sx={{ mb: 1 }}>{actionError}</Alert>}
      
      <Box sx={{ flexGrow: 1, overflow: 'auto' }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
            <CircularProgress size={24} sx={{ color: '#2B7A0B' }} />
          </Box>
        ) : !documents || documents.length === 0 ? (
          <Box sx={{ 
            p: 2, 
            textAlign: 'center',
            border: '2px dashed rgba(0, 0, 0, 0.1)',
            borderRadius: 2,
            bgcolor: 'rgba(0, 0, 0, 0.01)'
          }}>
            <Box sx={{ mb: 1 }}>
              <DescriptionOutlinedIcon sx={{ fontSize: 40, color: 'text.secondary', opacity: 0.5 }} />
            </Box>
            <Typography variant="subtitle2" color="text.secondary" gutterBottom>
              No documents
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph sx={{ mb: 1 }}>
              Please use the upload feature above to add documents to the knowledge base
            </Typography>
            <Divider sx={{ my: 1 }} />
            <Typography variant="caption" color="text.secondary" gutterBottom>
              Supported file formats:
            </Typography>
            <Box sx={{ display: 'flex', justifyContent: 'center', flexWrap: 'wrap', gap: 0.5, mt: 0.5 }}>
              <Chip size="small" label="PDF" icon={<PictureAsPdfIcon fontSize="small" />} />
              <Chip size="small" label="TXT" icon={<TextSnippetIcon fontSize="small" />} />
              <Chip size="small" label="DOCX" icon={<ArticleIcon fontSize="small" />} />
              <Chip size="small" label="MD" icon={<CodeIcon fontSize="small" />} />
            </Box>
            <Box sx={{ mt: 1 }}>
              <Typography variant="caption" color="text.secondary">
                Documents uploaded will be automatically processed and added to the knowledge base
              </Typography>
            </Box>
          </Box>
        ) : (
          <List dense>
            {documents.map((doc, index) => (
              <React.Fragment key={doc.id || doc._id || index}>
                <ListItem>
                  <ListItemAvatar>
                    <Avatar sx={{ bgcolor: 'background.paper', width: 32, height: 32 }}>
                      {getFileIcon(doc.filename || doc.name)}
                    </Avatar>
                  </ListItemAvatar>
                  <ListItemText
                    primary={
                      <Typography 
                        variant="body2" 
                        sx={{ 
                          fontWeight: 'medium', 
                          color: 'primary.main',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap'
                        }}
                      >
                        {doc.filename || doc.name || doc.document_name || doc.file_name || "Unnamed Document"}
                      </Typography>
                    }
                    secondary={
                      <>
                        <Typography 
                          component="span" 
                          variant="caption" 
                          sx={{
                            display: 'inline-block',
                            fontWeight: 'bold',
                            color: doc.status === 'processing' ? 'info.main' : 
                                  doc.status === 'failed' ? 'error.main' : 'success.main',
                            bgcolor: doc.status === 'processing' ? 'rgba(2, 136, 209, 0.1)' : 
                                    doc.status === 'failed' ? 'rgba(211, 47, 47, 0.1)' : 'rgba(46, 125, 50, 0.1)',
                            px: 0.75,
                            py: 0.25,
                            borderRadius: 1,
                            mb: 0.25
                          }}
                        >
                          Status: {
                            doc.status === 'processed' ? 'Processed' : 
                            doc.status === 'processing' ? 'Processing' : 
                            doc.status === 'ready' ? 'Processed' :
                            doc.status === 'failed' ? 'Failed' : 'Processed'
                          }
                        </Typography>
                        <br />
                        <Typography 
                          component="span" 
                          variant="caption"
                          sx={{
                            display: 'block',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap'
                          }}
                        >
                          Upload time: {formatDate(doc.upload_time || doc.uploadedAt || doc.processedAt || doc.createdAt)}
                          {doc.size > 0 && ` | Size: ${formatFileSize(doc.size)}`}
                          {doc.chunks > 0 && ` | Chunks: ${doc.chunks}`}
                        </Typography>
                      </>
                    }
                  />
                  <ListItemSecondaryAction>
                    {(doc.status === 'processed' || doc.status === 'ready') && (
                      <Tooltip title="Multi-agent generated study notes">
                        <IconButton 
                          edge="end" 
                          aria-label="multi-agent" 
                          onClick={() => handleProcessWithMultiAgent(doc.id || doc._id, doc.filename || doc.name || "Unnamed document")}
                          disabled={actionLoading || multiAgentLoading}
                          size="small"
                          sx={{ mr: 0.5, p: 0.5 }}
                        >
                          <PsychologyIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    )}
                    <Tooltip title="Reprocess">
                      <IconButton 
                        edge="end" 
                        aria-label="reprocess" 
                        onClick={() => handleReprocessDocument(doc.id || doc._id)}
                        disabled={actionLoading}
                        size="small"
                        sx={{ mr: 0.5, p: 0.5 }}
                      >
                        <RefreshIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete">
                      <IconButton 
                        edge="end" 
                        aria-label="delete" 
                        onClick={() => {
                          const docId = doc.id || doc._id;
                          console.log(`Clicked delete button, document object:`, doc);
                          console.log(`Clicked delete button, document ID: ${docId}`);
                          handleOpenDeleteDialog(docId);
                        }}
                        disabled={actionLoading}
                        size="small"
                        sx={{ p: 0.5 }}
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </ListItemSecondaryAction>
                </ListItem>
                {index < documents.length - 1 && <Divider />}
              </React.Fragment>
            ))}
          </List>
        )}
      </Box>

      {/* Delete confirmation dialog */}
      <Dialog
        open={deleteDialogOpen}
        onClose={handleCloseDeleteDialog}
      >
        <DialogTitle>Confirm Delete</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete this document? This operation cannot be undone.
            {selectedDocId && <div style={{ marginTop: '10px', fontWeight: 'bold' }}>Document ID: {selectedDocId}</div>}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDeleteDialog} disabled={actionLoading}>
            Cancel
          </Button>
          <Button 
            onClick={handleDeleteDocument} 
            color="error" 
            disabled={actionLoading}
            startIcon={actionLoading ? <CircularProgress size={20} sx={{ color: 'inherit' }} /> : null}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      {/* 多Agent处理对话框 */}
      <Dialog
        open={multiAgentDialogOpen}
        onClose={handleCloseMultiAgentDialog}
        maxWidth="md"
        fullWidth
        PaperProps={{ 
          sx: { 
            height: 'auto',
            maxHeight: '90vh',
            display: 'flex',
            flexDirection: 'column',
            overflowX: 'hidden'
          } 
        }}
      >
        <DialogTitle sx={{ px: 3, py: 1.5 }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <PsychologyIcon sx={{ mr: 1, color: '#2B7A0B' }} />
            Multi-Agent System Analysis - {selectedDocName}
          </Box>
        </DialogTitle>
        <DialogContent sx={{ overflow: 'auto', flex: '1 1 auto', px: 3, py: 1 }}>
          {actionError && (
            <Alert severity="error" sx={{ mb: 2 }}>{actionError}</Alert>
          )}
          
          {!multiAgentResult ? (
            <Box>
              <DialogContentText sx={{ display: 'none' }}>
                Using the Multi-Agent system to analyze this document will perform the following operations:
                <Box component="ul" sx={{ mt: 1 }}>
                  <li><strong>Content Analysis Agent</strong>: Identify sections, headings, and paragraph structure</li>
                  <li><strong>Knowledge Extraction Agent</strong>: Extract key definitions, concepts, and formulas</li>
                  <li><strong>Summary Agent</strong>: Generate concise summaries for each section</li>
                  <li><strong>Formatting Agent</strong>: Output as Markdown and Anki card format</li>
                </Box>
              </DialogContentText>
              
              {/* 显示处理进度 */}
              {multiAgentLoading ? (
                renderProgressStepper()
              ) : (
                <Box sx={{ mt: 2 }}>
                  <Box sx={{ 
                    display: 'flex', 
                    flexDirection: 'column', 
                    alignItems: 'center',
                    justifyContent: 'center',
                    textAlign: 'center',
                    mb: 2
                  }}>
                    <Box sx={{ 
                      position: 'relative',
                      width: 100,
                      height: 100,
                      mb: 2
                    }}>
                      <PulseBox sx={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        height: '100%',
                        borderRadius: '50%',
                        background: 'linear-gradient(135deg, rgba(43, 122, 11, 0.1), rgba(43, 122, 11, 0.2))'
                      }} />
                      <Box sx={{
                        position: 'absolute',
                        top: '50%',
                        left: '50%',
                        transform: 'translate(-50%, -50%)',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center'
                      }}>
                        <PsychologyIcon sx={{ fontSize: 48, color: '#2B7A0B' }} />
                      </Box>
                    </Box>
                    <Typography variant="h5" sx={{ fontWeight: 'bold', color: '#333', mb: 1 }}>
                      Multi-Agent Learning Note Generator
                    </Typography>
                    <Typography variant="body2" sx={{ color: '#666', mb: 2, maxWidth: '80%' }}>
                      AI-based complex document understanding and knowledge extraction system that generates structured learning notes and review cards
                    </Typography>
                  </Box>
                  
                  <Box sx={{ 
                    display: 'flex', 
                    flexWrap: 'wrap', 
                    gap: 1.5,
                    justifyContent: 'center',
                    mb: 2,
                  }}>
                    {multiAgentProcessingSteps.map((step, index) => (
                      <Box key={step.id} sx={{ 
                        width: {xs: '100%', sm: 'calc(50% - 12px)', md: 'calc(25% - 12px)'},
                        p: 1.5,
                        borderRadius: 2,
                        bgcolor: `${step.color}10`,
                        border: `1px solid ${step.color}30`,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        transition: 'transform 0.2s ease, box-shadow 0.2s ease',
                        '&:hover': {
                          transform: 'translateY(-2px)',
                          boxShadow: '0 4px 8px rgba(0,0,0,0.05)'
                        }
                      }}>
                        <Box sx={{ 
                          fontSize: '24px', 
                          width: 40, 
                          height: 40, 
                          display: 'flex', 
                          alignItems: 'center', 
                          justifyContent: 'center',
                          borderRadius: '50%',
                          bgcolor: `${step.color}20`,
                          mb: 0.5
                        }}>
                          {step.icon}
                        </Box>
                        <Typography variant="subtitle1" sx={{ fontWeight: '600', fontSize: '0.9rem', mb: 0.5, color: step.color }}>
                          {step.label}
                        </Typography>
                        <Typography variant="caption" sx={{ color: '#666', textAlign: 'center' }}>
                          {step.description}
                        </Typography>
                      </Box>
                    ))}
                  </Box>
                  
                  <Box sx={{ 
                    p: 2, 
                    bgcolor: 'rgba(43, 122, 11, 0.08)', 
                    borderRadius: 2, 
                    border: '1px dashed rgba(43, 122, 11, 0.4)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1.5
                  }}>
                    <Box sx={{ color: '#2B7A0B', display: 'flex', alignItems: 'center', fontSize: '20px' }}>
                      ⏱️
                    </Box>
                    <Typography variant="body2" sx={{ color: '#2B7A0B', fontWeight: 'medium' }}>
                      This process may take a few minutes. Please be patient. The system will automatically process the document and generate high-quality learning notes and Anki review cards.
                    </Typography>
                  </Box>
                </Box>
              )}
              
              <Box sx={{ mt: 0.5, fontSize: '0.75rem', color: 'text.secondary' }}>
                Document ID: {selectedDocForMultiAgent}
              </Box>
            </Box>
          ) : (
            <Box sx={{ mt: 0 }}>
              <Tabs 
                value={multiAgentTabValue} 
                onChange={handleMultiAgentTabChange}
                sx={{ 
                  mb: 2, 
                  borderBottom: 1, 
                  borderColor: 'divider',
                  '& .MuiTab-root': { color: '#555' },
                  '& .Mui-selected': { color: '#2B7A0B' },
                  '& .MuiTabs-indicator': { backgroundColor: '#2B7A0B' }
                }}
                variant="fullWidth"
              >
                <Tab label="Markdown Notes" />
                <Tab label="Anki Cards" />
                <Tab label="Processing Details" />
              </Tabs>
              
              {renderMultiAgentResultContent()}
            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ flex: '0 0 auto', px: 3, py: 1 }}>
          <Button onClick={handleCloseMultiAgentDialog} variant="outlined" sx={{ color: '#555', borderColor: '#ccc' }}>
            {multiAgentResult ? 'Close' : 'Cancel'}
          </Button>
          {!multiAgentResult && (
            <Button 
              onClick={handleConfirmMultiAgentProcess} 
              disabled={multiAgentLoading}
              startIcon={multiAgentLoading ? <CircularProgress size={20} sx={{ color: 'inherit' }} /> : null}
              variant="contained"
              size="medium"
              sx={{ 
                bgcolor: '#2B7A0B', 
                '&:hover': { bgcolor: '#1d5407' },
                '&.Mui-disabled': { 
                  bgcolor: '#cfe8d6', 
                  color: '#5c8968' 
                }
              }}
            >
              {multiAgentLoading ? 'Processing...' : 'Start Processing'}
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default DocumentList; 