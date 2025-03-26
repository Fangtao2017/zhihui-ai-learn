import React, { useState } from 'react';
import { 
  Box, Button, Typography, LinearProgress, Alert, 
  Paper, IconButton, Tooltip, Stepper, Step, StepLabel, StepContent
} from '@mui/material';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import CloseIcon from '@mui/icons-material/Close';
import axios from 'axios';
import { styled } from '@mui/material/styles';

const Input = styled('input')({
  display: 'none',
});

// 修改主题色为绿色系
const themeColor = '#2B7A0B';

const UploadComponent = ({ onUploadSuccess }) => {
  const [selectedFile, setSelectedFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [activeStep, setActiveStep] = useState(0);

  // Upload steps
  const steps = [
    { 
      label: 'Select File', 
      description: 'Choose a document file to upload (PDF, TXT, DOCX, MD)' 
    },
    { 
      label: 'Upload File', 
      description: 'Upload the file to the server' 
    },
    { 
      label: 'Process Document', 
      description: 'The system is processing the document, extracting text and creating vector indices' 
    }
  ];

  // Handle file selection
  const handleFileChange = (event) => {
    const file = event.target.files[0];
    if (file) {
      setSelectedFile(file);
      setError(null);
      setSuccess(false);
      setActiveStep(1); // Move to step 2
    }
  };

  // Clear selected file
  const handleClearFile = () => {
    setSelectedFile(null);
    setError(null);
    setSuccess(false);
    setActiveStep(0); // Reset steps
  };

  // Handle file upload
  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a file first');
      return;
    }

    // 检查文件类型
    const validTypes = ['application/pdf', 'text/plain', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'text/markdown'];
    const fileType = selectedFile.type;
    
    if (!validTypes.includes(fileType) && 
        !(selectedFile.name.endsWith('.pdf') || 
          selectedFile.name.endsWith('.txt') || 
          selectedFile.name.endsWith('.docx') ||
          selectedFile.name.endsWith('.md'))) {
      setError('Invalid file type. Please upload PDF, TXT, DOCX, or MD files.');
      return;
    }

    // 检查文件大小
    const maxSize = 50 * 1024 * 1024; // 50MB
    if (selectedFile.size > maxSize) {
      setError(`File size exceeds the limit (50MB). Current size: ${(selectedFile.size / (1024 * 1024)).toFixed(2)}MB`);
      return;
    }

    setUploading(true);
    setUploadProgress(0);
    setError(null);
    setSuccess(false);
    setActiveStep(2); // Move to step 3

    const formData = new FormData();
    formData.append('file', selectedFile);

    try {
      const token = localStorage.getItem('token');
      await axios.post('http://localhost:8080/api/rag/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
          'Authorization': `Bearer ${token}`
        },
        onUploadProgress: (progressEvent) => {
          const progress = Math.round((progressEvent.loaded / progressEvent.total) * 100);
          setUploadProgress(progress);
        },
      });

      setSuccess(true);
      setSelectedFile(null);
      setActiveStep(0); // Reset steps
      if (onUploadSuccess) onUploadSuccess();
    } catch (err) {
      console.error('Upload failed:', err);
      setError(err.response?.data?.message || 'Failed to upload file');
      setActiveStep(1); // Back to upload step
    } finally {
      setUploading(false);
    }
  };

  return (
    <Box>
      <Typography variant="subtitle1" gutterBottom sx={{ mb: 1, fontWeight: 'medium' }}>
        Upload Document
      </Typography>
      
      <Box sx={{ mb: 1 }}>
        <Typography variant="caption" color="text.secondary" gutterBottom>
          Supported file formats: PDF, TXT, DOCX, MD
        </Typography>
        
        <Box sx={{ 
          display: 'flex', 
          flexDirection: { xs: 'column', sm: 'row' }, // Vertical on small screens, horizontal on large screens
          gap: 1, // Add spacing
          mt: 0.5
        }}>
          <input
            accept=".pdf,.txt,.docx,.md"
            style={{ display: 'none' }}
            id="file-upload"
            type="file"
            onChange={handleFileChange}
            disabled={uploading}
          />
          
          <label htmlFor="file-upload" style={{ width: '100%' }}>
            <Button
              variant="contained"
              component="span"
              startIcon={<CloudUploadIcon />}
              disabled={uploading}
              fullWidth // Make button fill available space
              size="small" // Smaller button size
              sx={{ bgcolor: themeColor, '&:hover': { bgcolor: '#1d5407' } }}
            >
              Select File
            </Button>
          </label>
          
          <Button
            variant="contained"
            color="primary"
            onClick={handleUpload}
            disabled={!selectedFile || uploading}
            fullWidth // Make button fill available space
            size="small" // Smaller button size
            sx={{ bgcolor: themeColor, '&:hover': { bgcolor: '#1d5407' } }}
          >
            Upload
          </Button>
        </Box>
      </Box>
      
      {selectedFile && (
        <Paper variant="outlined" sx={{ 
          p: 0.75, 
          mb: 1, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          overflow: 'hidden' // Prevent content overflow
        }}>
          <Typography variant="caption" noWrap sx={{ 
            maxWidth: '80%',
            overflow: 'hidden',
            textOverflow: 'ellipsis'
          }}>
            {selectedFile.name} ({(selectedFile.size / 1024).toFixed(2)} KB)
          </Typography>
          <Tooltip title="Clear">
            <IconButton size="small" onClick={handleClearFile} disabled={uploading} sx={{ p: 0.5 }}>
              <CloseIcon fontSize="small" />
            </IconButton>
          </Tooltip>
        </Paper>
      )}
      
      {/* Upload step indicator */}
      {selectedFile && (
        <Stepper activeStep={activeStep} orientation="vertical" sx={{ mb: 1 }}>
          {steps.map((step, index) => (
            <Step key={index}>
              <StepLabel sx={{ '& .MuiStepLabel-label': { fontSize: '0.75rem' } }}>{step.label}</StepLabel>
              <StepContent>
                <Typography variant="caption" color="text.secondary">
                  {step.description}
                </Typography>
                {index === 1 && uploading && (
                  <Box sx={{ mt: 1, mb: 0.5 }}>
                    <LinearProgress variant="determinate" value={uploadProgress} />
                    <Typography variant="caption" color="text.secondary" align="center" sx={{ mt: 0.5 }}>
                      {uploadProgress}%
                    </Typography>
                  </Box>
                )}
              </StepContent>
            </Step>
          ))}
        </Stepper>
      )}
      
      {error && <Alert severity="error" sx={{ mb: 1, py: 0, fontSize: '0.75rem' }}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 1, py: 0, fontSize: '0.75rem' }}>Document uploaded successfully! The system is processing the document, please wait...</Alert>}
    </Box>
  );
};

export default UploadComponent; 