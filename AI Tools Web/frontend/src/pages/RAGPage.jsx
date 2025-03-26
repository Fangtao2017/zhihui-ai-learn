import React, { useState, useEffect } from 'react';
import { Container, Typography, Box, Paper, Grid, Divider } from '@mui/material';
import axios from 'axios';
import DocumentList from './RAGComponents/DocumentList';
import UploadComponent from './RAGComponents/UploadComponent';
import QueryComponent from './RAGComponents/QueryComponent';
import StorageIcon from '@mui/icons-material/Storage';

const RAGPage = () => {
  const [documents, setDocuments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [refreshInterval, setRefreshInterval] = useState(null);

  // Fetch document list
  const fetchDocuments = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      console.log('Fetching document list, using token:', token ? token.substring(0, 10) + '...' : 'null');
      
      const response = await axios.get('http://localhost:8080/api/rag/documents', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      console.log('Retrieved document list:', response.data);
      if (response.data && Array.isArray(response.data)) {
        console.log('Document count:', response.data.length);
        if (response.data.length > 0) {
          console.log('First document example:', response.data[0]);
          console.log('Document ID field:', response.data[0].id || response.data[0]._id || 'ID field not found');
        }
        setDocuments(response.data);
      } else {
        console.log('Response data is not an array, setting to empty array');
        setDocuments([]);
      }
      setError(null);
    } catch (err) {
      console.error('Failed to fetch document list:', err);
      setError('Failed to retrieve document list, please try again later');
    } finally {
      setLoading(false);
    }
  };

  // Check if there are documents being processed
  const hasProcessingDocuments = () => {
    return documents && documents.length > 0 && documents.some(doc => doc.status === 'processing');
  };

  useEffect(() => {
    fetchDocuments();

    // Set up periodic refresh
    const interval = setInterval(() => {
      if (hasProcessingDocuments()) {
        fetchDocuments();
      }
    }, 5000); // Refresh every 5 seconds

    setRefreshInterval(interval);

    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  }, []);

  // When document list changes, check if we need to continue periodic refresh
  useEffect(() => {
    if (!hasProcessingDocuments() && refreshInterval) {
      clearInterval(refreshInterval);
      setRefreshInterval(null);
    } else if (hasProcessingDocuments() && !refreshInterval) {
      const interval = setInterval(() => {
        fetchDocuments();
      }, 5000);
      setRefreshInterval(interval);
    }
  }, [documents]);

  // Handle refresh after document upload success
  const handleUploadSuccess = () => {
    fetchDocuments();
    
    // Ensure we start periodic refresh after uploading new document
    if (!refreshInterval) {
      const interval = setInterval(() => {
        fetchDocuments();
      }, 5000);
      setRefreshInterval(interval);
    }
  };

  return (
    <Container maxWidth="lg" sx={{ 
      mt: 2, // Add margin top to prevent overlap with navigation
      pb: 2,
      height: 'calc(100vh - 76px)', // Adjust container height to account for top margin
      overflow: 'hidden', // Prevent overall overflow
      display: 'flex',
      flexDirection: 'column'
    }}>
      {/* Page header with icon and title */}
      <Box sx={{ 
        display: 'flex', 
        alignItems: 'center', 
        mb: 2, 
        pb: 1.5,
        borderBottom: '1px solid rgba(0, 0, 0, 0.1)'
      }}>
        <StorageIcon sx={{ 
          fontSize: 30, 
          color: '#2B7A0B', 
          mr: 1.5,
          verticalAlign: 'middle'
        }} />
        <Box>
          <Typography 
            variant="h5" 
            component="h1" 
            sx={{ 
              fontWeight: 'bold', 
              color: '#2B7A0B', 
              lineHeight: 1.2,
              display: 'flex',
              alignItems: 'center'
            }}
          >
            Knowledge Base QA System
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Upload documents, ask questions, and get precise answers based on document content
          </Typography>
        </Box>
      </Box>
      
      <Grid container spacing={2} sx={{ flexGrow: 1, overflow: 'hidden' }}>
        {/* Left panel: upload and document list */}
        <Grid item xs={12} md={4} sx={{ 
          height: '100%',
          display: 'flex',
          flexDirection: 'column'
        }}>
          <Paper elevation={3} sx={{ p: 1.5, mb: 2 }}>
            <UploadComponent onUploadSuccess={handleUploadSuccess} />
          </Paper>
          
          <Paper elevation={3} sx={{ 
            p: 1.5, 
            flexGrow: 1,
            overflow: 'auto', // Allow scrolling when content overflows
            maxHeight: 'calc(100vh - 250px)' // Increase maximum height
          }}>
            <DocumentList 
              documents={documents} 
              loading={loading} 
              error={error} 
              onDocumentDeleted={fetchDocuments}
              onDocumentReprocessed={fetchDocuments}
            />
          </Paper>
        </Grid>
        
        {/* Right panel: query and results */}
        <Grid item xs={12} md={8} sx={{ 
          height: '100%',
          display: 'flex',
          flexDirection: 'column'
        }}>
          <Paper elevation={3} sx={{ 
            p: { xs: 1.5, sm: 2 },
            flexGrow: 1,
            overflow: 'auto', // Allow scrolling when content overflows
            maxHeight: 'calc(100vh - 80px)', // Increase maximum height
            borderRadius: 2
          }}>
            <QueryComponent />
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
};

export default RAGPage; 