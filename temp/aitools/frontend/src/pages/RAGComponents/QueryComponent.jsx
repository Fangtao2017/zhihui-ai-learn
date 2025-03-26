import React, { useState } from 'react';
import { 
  Box, TextField, Button, Typography, Paper, 
  CircularProgress, Alert, Divider, List, ListItem, ListItemText,
  Card, CardContent, Chip, Grid, IconButton, Tooltip
} from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import LightbulbOutlinedIcon from '@mui/icons-material/LightbulbOutlined';
import SchoolOutlinedIcon from '@mui/icons-material/SchoolOutlined';
import axios from 'axios';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

const QueryComponent = () => {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);

  // Example questions list
  const exampleQuestions = [
    "What is the main content of this document?",
    "Please summarize the key points in the document",
    "What are the important data or statistics in the document?",
    "What is the conclusion of this report?"
  ];

  // Fill example question into query box
  const handleExampleClick = (question) => {
    setQuery(question);
    setError(null);
  };

  // Handle query input change
  const handleQueryChange = (event) => {
    setQuery(event.target.value);
    setError(null);
  };

  // Handle query submission
  const handleSubmit = async (event) => {
    event.preventDefault();
    
    if (!query.trim()) {
      setError('Please enter a query');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      const token = localStorage.getItem('token');
      // Build request data
      const requestData = { 
        query: query.trim() 
      };
      
      console.log('Complete request data:', JSON.stringify(requestData));
      
      const response = await axios.post('http://localhost:8080/api/rag/query', 
        requestData,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          }
        }
      );
      setResult(response.data);
      console.log('Received query response:', response.data);
      
      // Print full answer content for debugging
      if (response.data && response.data.answer) {
        console.log('Answer content:', response.data.answer);
      }
    } catch (err) {
      console.error('Query failed:', err);
      setError(err.response?.data?.error || 'Query failed, please try again later');
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  // Format answer as Markdown
  const formatAnswer = (answer) => {
    if (!answer) return '';
    
    // Ensure titles use Markdown format
    let formattedAnswer = answer;
    
    // If answer doesn't have title format, add one
    if (!formattedAnswer.includes('# ')) {
      formattedAnswer = `## Answer\n\n${formattedAnswer}`;
    }
    
    return formattedAnswer;
  };

  // Filter valid sources
  const getValidSources = (sources) => {
    if (!sources || !Array.isArray(sources)) return [];
    
    // Filter out sources without document name or content
    return sources.filter(source => 
      source && 
      (source.document_name || source.documentName) && 
      (source.content || source.text) && 
      (source.content || source.text).trim() !== ''
    );
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <Typography variant="subtitle1" gutterBottom sx={{ mb: 1, fontWeight: 'medium' }}>
        Knowledge Base Query
      </Typography>
      
      <form onSubmit={handleSubmit}>
        <Box sx={{ 
          display: 'flex', 
          flexDirection: { xs: 'column', sm: 'row' },
          alignItems: { sm: 'center' },
          gap: 1
        }}>
          <TextField
            fullWidth
            label="Enter your question"
            variant="outlined"
            value={query}
            onChange={handleQueryChange}
            disabled={loading}
            size="small"
            sx={{ 
              '& .MuiOutlinedInput-root': {
                borderRadius: '4px 0 0 4px'
              }
            }}
          />
          
          <Button
            type="submit"
            variant="contained"
            color="primary"
            endIcon={<SendIcon />}
            disabled={loading || !query.trim()}
            sx={{ 
              minWidth: { xs: '100%', sm: '120px' },
              height: '40px',
              borderRadius: { xs: '4px', sm: '0 4px 4px 0' }
            }}
          >
            {loading ? 'Querying...' : 'Query'}
          </Button>
        </Box>
      </form>
      
      {error && <Alert severity="error" sx={{ mt: 1 }}>{error}</Alert>}
      
      {loading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
          <CircularProgress />
        </Box>
      )}
      
      {!result && !loading && (
        <Box sx={{ mt: 2, mb: 1 }}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 1.5, 
              borderRadius: 2, 
              bgcolor: 'rgba(0, 0, 0, 0.02)',
              border: '1px dashed rgba(0, 0, 0, 0.1)'
            }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
              <SchoolOutlinedIcon sx={{ mr: 1, fontSize: 18, color: '#2B7A0B' }} />
              <Typography variant="subtitle2" sx={{ fontWeight: "medium", color: '#2B7A0B' }}>
                Welcome to the Knowledge Base Query System
              </Typography>
            </Box>
            
            <Typography variant="body2" paragraph sx={{ mb: 1 }}>
              You can ask questions about the uploaded documents, and the system will provide accurate answers based on the document content, along with reference sources.
            </Typography>
            
            <Divider sx={{ my: 1 }} />
            
            <Box sx={{ mb: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
                <HelpOutlineIcon sx={{ mr: 1, fontSize: 16, color: '#2B7A0B' }} />
                <Typography variant="subtitle2" fontWeight="medium">
                  Usage Guide
                </Typography>
              </Box>
              
              <Box component="ul" sx={{ pl: 3, mt: 0.5, mb: 0 }}>
                <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
                  Ensure you have uploaded and processed the relevant documents
                </Typography>
                <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
                  Ask questions as specifically as possible to get more accurate answers
                </Typography>
                <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
                  The system will automatically find relevant content from the documents and generate answers
                </Typography>
                <Typography component="li" variant="body2">
                  Answer below will display quoted document fragments to help you verify information accuracy
                </Typography>
              </Box>
            </Box>
            
            <Divider sx={{ my: 1 }} />
            
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
                <LightbulbOutlinedIcon sx={{ mr: 1, fontSize: 16, color: '#2B7A0B' }} />
                <Typography variant="subtitle2" fontWeight="medium">
                  Example Questions (Click to Use)
                </Typography>
              </Box>
              
              <Box sx={{ 
                display: 'flex', 
                flexWrap: 'wrap', 
                gap: 0.5
              }}>
                {exampleQuestions.map((question, index) => (
                  <Chip
                    key={index}
                    label={question}
                    onClick={() => handleExampleClick(question)}
                    size="small"
                    sx={{ 
                      height: 'auto',
                      py: 0.5,
                      '& .MuiChip-label': { 
                        whiteSpace: 'normal'
                      }
                    }}
                  />
                ))}
              </Box>
            </Box>
          </Paper>
        </Box>
      )}
      
      {result && (
        <Box sx={{ mt: 1.5, overflow: 'auto', flexGrow: 1 }}>
          <Paper elevation={2} sx={{ 
            p: { xs: 2, sm: 3 },
            borderRadius: 2,
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.05)'
          }}>
            <Box sx={{ 
              '& a': { color: '#2B7A0B' },
              '& h1, & h2, & h3, & h4, & h5, & h6': { 
                mt: 2, 
                mb: 1,
                fontWeight: 'bold',
                color: '#2B7A0B'
              },
              '& code': { 
                backgroundColor: 'rgba(0, 0, 0, 0.05)', 
                p: 0.5, 
                borderRadius: 1,
                fontFamily: 'Consolas, Monaco, "Andale Mono", monospace'
              },
              '& pre': { 
                backgroundColor: 'rgba(0, 0, 0, 0.05)', 
                p: 2,
                borderRadius: 1,
                overflowX: 'auto',
                border: '1px solid rgba(0, 0, 0, 0.1)'
              },
              '& blockquote': { 
                borderLeft: '4px solid rgba(0, 0, 0, 0.1)', 
                pl: 2,
                ml: 0,
                my: 1.5,
                color: 'text.secondary'
              },
              '& ul, & ol': { pl: 3, mb: 2 },
              '& img': { maxWidth: '100%' },
              '& p': { mb: 1.5, lineHeight: 1.6 }
            }}>
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {formatAnswer(result.answer)}
              </ReactMarkdown>
            </Box>
            
            {/* Quoted document fragments */}
            {getValidSources(result.sources).length > 0 && (
              <>
                <Divider sx={{ my: 2 }} />
                <Typography variant="subtitle1" gutterBottom sx={{ 
                  color: '#2B7A0B',
                  fontWeight: 'bold',
                  mb: 1.5
                }}>
                  Reference Sources
                </Typography>
                <List sx={{ 
                  bgcolor: 'background.paper', 
                  borderRadius: 2,
                  p: 0
                }}>
                  {getValidSources(result.sources).map((source, index) => (
                    <ListItem 
                      key={index} 
                      sx={{ 
                        display: 'block',
                        mb: 1.5,
                        p: 1.5,
                        bgcolor: 'rgba(0, 0, 0, 0.02)',
                        borderRadius: 2,
                        borderLeft: '4px solid',
                        borderColor: '#2B7A0B',
                        transition: 'all 0.2s',
                        '&:hover': {
                          bgcolor: 'rgba(0, 0, 0, 0.04)',
                          transform: 'translateY(-2px)'
                        }
                      }}
                    >
                      <Typography variant="subtitle2" sx={{ 
                        fontWeight: 'bold', 
                        mb: 0.5,
                        color: '#2B7A0B'
                      }}>
                        {index + 1}. Document {(source.document_name || source.documentName).replace(/^文档-/i, '')}
                      </Typography>
                      <Paper 
                        variant="outlined" 
                        sx={{ 
                          p: 1.5,
                          bgcolor: 'rgba(255, 255, 255, 0.7)',
                          borderRadius: 1,
                          maxHeight: '150px',
                          overflow: 'auto',
                          border: '1px solid rgba(0, 0, 0, 0.1)'
                        }}
                      >
                        <Typography 
                          variant="body2" 
                          sx={{ 
                            whiteSpace: 'pre-wrap',
                            fontFamily: 'Consolas, Monaco, monospace',
                            fontSize: '0.8rem',
                            lineHeight: 1.4
                          }}
                        >
                          {source.content || source.text}
                        </Typography>
                      </Paper>
                    </ListItem>
                  ))}
                </List>
              </>
            )}
          </Paper>
        </Box>
      )}
    </Box>
  );
};

export default QueryComponent; 