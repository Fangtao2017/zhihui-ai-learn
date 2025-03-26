import React, { useState } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Button, 
  Divider, 
  Alert, 
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  IconButton,
  InputAdornment
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import LockIcon from '@mui/icons-material/Lock';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import axios from 'axios';

const SettingsPage = () => {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(null);
  const [error, setError] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  
  // Password change states
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState(null);
  const [passwordSuccess, setPasswordSuccess] = useState(null);
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [changingPassword, setChangingPassword] = useState(false);

  // Open confirmation dialog
  const handleOpenDialog = () => {
    setOpenDialog(true);
  };

  // Close confirmation dialog
  const handleCloseDialog = () => {
    setOpenDialog(false);
  };

  // Open password change dialog
  const handleOpenPasswordDialog = () => {
    // Reset password change states
    setOldPassword('');
    setNewPassword('');
    setConfirmPassword('');
    setPasswordError(null);
    setPasswordSuccess(null);
    setOpenPasswordDialog(true);
  };

  // Close password change dialog
  const handleClosePasswordDialog = () => {
    setOpenPasswordDialog(false);
  };

  // Handle password visibility toggle
  const handleClickShowOldPassword = () => {
    setShowOldPassword(!showOldPassword);
  };

  const handleClickShowNewPassword = () => {
    setShowNewPassword(!showNewPassword);
  };

  const handleClickShowConfirmPassword = () => {
    setShowConfirmPassword(!showConfirmPassword);
  };

  // Handle password change
  const handleChangePassword = async () => {
    // Validate inputs
    if (!oldPassword || !newPassword || !confirmPassword) {
      setPasswordError('All fields are required');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('New passwords do not match');
      return;
    }

    if (newPassword.length < 6) {
      setPasswordError('New password must be at least 6 characters long');
      return;
    }

    setChangingPassword(true);
    setPasswordError(null);
    setPasswordSuccess(null);

    try {
      const token = localStorage.getItem('token');
      
      const response = await axios.post(
        'http://localhost:8080/api/user/change-password',
        {
          oldPassword: oldPassword,
          newPassword: newPassword
        },
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          }
        }
      );

      setPasswordSuccess('Password changed successfully');
      // Reset fields
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
      
      // Close dialog after a delay
      setTimeout(() => {
        handleClosePasswordDialog();
        setPasswordSuccess(null);
      }, 2000);
    } catch (err) {
      console.error('Failed to change password:', err);
      setPasswordError(err.response?.data?.error || 'Failed to change password');
    } finally {
      setChangingPassword(false);
    }
  };

  // Clear vector database
  const handleClearVectors = async () => {
    setLoading(true);
    setSuccess(null);
    setError(null);
    handleCloseDialog();

    try {
      console.log('Starting to clear vector database...');
      const token = localStorage.getItem('token');
      console.log('Using token:', token ? token.substring(0, 10) + '...' : 'null');
      
      const response = await axios.post(
        'http://localhost:8080/api/rag/clear-vectors',
        {},
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          }
        }
      );

      console.log('Vector database cleared successfully:', response.data);
      setSuccess('Vector database has been successfully cleared! You can now re-import your documents.');
    } catch (err) {
      console.error('Failed to clear vector database:', err);
      console.error('Error details:', err.response ? err.response.data : 'No response data');
      setError(err.response?.data?.error || 'Failed to clear vector database, please try again later');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{ p: 3, maxWidth: '800px', mx: 'auto' }}>
      <Typography variant="h4" gutterBottom>
        System Settings
      </Typography>
      
      {/* Password Management Section */}
      <Paper elevation={2} sx={{ p: 3, mb: 4 }}>
        <Typography variant="h6" gutterBottom>
          Account Security
        </Typography>
        <Divider sx={{ mb: 2 }} />
        
        <Box sx={{ mb: 2 }}>
          <Typography variant="body1" paragraph>
            You can change your password here. For security reasons, you will need to enter your current password.
          </Typography>
        </Box>
        
        <Button
          variant="contained"
          color="primary"
          startIcon={<LockIcon />}
          onClick={handleOpenPasswordDialog}
          sx={{ mt: 2 }}
        >
          Change Password
        </Button>
      </Paper>
      
      <Paper elevation={2} sx={{ p: 3, mb: 4 }}>
        <Typography variant="h6" gutterBottom>
          Knowledge Base Management
        </Typography>
        <Divider sx={{ mb: 2 }} />
        
        <Box sx={{ mb: 2 }}>
          <Typography variant="body1" paragraph>
            Clearing the vector database will delete all stored vector data. This is useful in the following cases:
          </Typography>
          <ul>
            <li>
              <Typography variant="body2" sx={{ mb: 1 }}>
                When you have deleted documents from MongoDB, but related vector data still exists in the vector database
              </Typography>
            </li>
            <li>
              <Typography variant="body2" sx={{ mb: 1 }}>
                When "unknown document" appears in query results, indicating vectors without corresponding MongoDB records
              </Typography>
            </li>
            <li>
              <Typography variant="body2" sx={{ mb: 1 }}>
                When you want to completely reset the knowledge base and re-import all documents
              </Typography>
            </li>
          </ul>
        </Box>
        
        <Button
          variant="contained"
          color="error"
          startIcon={<DeleteIcon />}
          onClick={handleOpenDialog}
          disabled={loading}
          sx={{ mt: 2 }}
        >
          {loading ? 'Clearing...' : 'Clear Vector Database'}
        </Button>
        
        {success && (
          <Alert severity="success" sx={{ mt: 2 }}>
            {success}
          </Alert>
        )}
        
        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
        
        {loading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
            <CircularProgress size={24} />
          </Box>
        )}
      </Paper>
      
      {/* Confirmation dialog for clearing vector database */}
      <Dialog
        open={openDialog}
        onClose={handleCloseDialog}
      >
        <DialogTitle>
          Confirm Vector Database Clearing
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to clear the vector database? This action will delete all stored vector data and cannot be undone.
            After clearing, you will need to re-upload and process documents to rebuild the knowledge base.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} color="primary">
            Cancel
          </Button>
          <Button onClick={handleClearVectors} color="error" variant="contained">
            Confirm Clearing
          </Button>
        </DialogActions>
      </Dialog>
      
      {/* Password Change Dialog */}
      <Dialog
        open={openPasswordDialog}
        onClose={handleClosePasswordDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          Change Password
        </DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }}>
            To change your password, please enter your current password and then enter a new password.
          </DialogContentText>
          
          {passwordError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {passwordError}
            </Alert>
          )}
          
          {passwordSuccess && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {passwordSuccess}
            </Alert>
          )}
          
          <TextField
            margin="dense"
            label="Current Password"
            type={showOldPassword ? 'text' : 'password'}
            fullWidth
            variant="outlined"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            disabled={changingPassword}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    aria-label="toggle password visibility"
                    onClick={handleClickShowOldPassword}
                    edge="end"
                  >
                    {showOldPassword ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
            sx={{ mb: 2 }}
          />
          
          <TextField
            margin="dense"
            label="New Password"
            type={showNewPassword ? 'text' : 'password'}
            fullWidth
            variant="outlined"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            disabled={changingPassword}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    aria-label="toggle password visibility"
                    onClick={handleClickShowNewPassword}
                    edge="end"
                  >
                    {showNewPassword ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
            sx={{ mb: 2 }}
          />
          
          <TextField
            margin="dense"
            label="Confirm New Password"
            type={showConfirmPassword ? 'text' : 'password'}
            fullWidth
            variant="outlined"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            disabled={changingPassword}
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    aria-label="toggle password visibility"
                    onClick={handleClickShowConfirmPassword}
                    edge="end"
                  >
                    {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClosePasswordDialog} color="primary" disabled={changingPassword}>
            Cancel
          </Button>
          <Button 
            onClick={handleChangePassword} 
            color="primary" 
            variant="contained"
            disabled={changingPassword}
          >
            {changingPassword ? 'Changing...' : 'Change Password'}
          </Button>
        </DialogActions>
        {changingPassword && (
          <Box sx={{ display: 'flex', justifyContent: 'center', pb: 2 }}>
            <CircularProgress size={24} />
          </Box>
        )}
      </Dialog>
    </Box>
  );
};

export default SettingsPage;