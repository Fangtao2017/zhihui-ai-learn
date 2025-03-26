import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import LoginPage from './pages/LoginPage';
import SignUpPage from './pages/SignUpPage';
import ChatPage from './pages/ChatPage';
import HomePage from './pages/HomePage';
import SupportPage from './pages/SupportPage';
import SettingsPage from './pages/SettingsPage.jsx';
import RAGPage from './pages/RAGPage';
import { App as AntdApp } from 'antd';

// Protected Route Component
const ProtectedRoute = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return children;
};

// Public Route Component
const PublicRoute = ({ children }) => {
  const token = localStorage.getItem('token');
  if (token) {
    return <Navigate to="/chat" replace />;
  }
  return children;
};

function App() {
  return (
    <AntdApp>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#2B7A0B',
            borderRadius: 8,
            fontSize: 14,
          },
          components: {
            Button: {
              borderRadius: 8,
              fontSize: 14,
            },
            Input: {
              borderRadius: 8,
            },
          },
        }}
      >
        <Router>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<PublicRoute><LoginPage /></PublicRoute>} />
            <Route path="/signup" element={<PublicRoute><SignUpPage /></PublicRoute>} />

            {/* Protected Routes */}
            <Route path="/" element={<ProtectedRoute><HomePage /></ProtectedRoute>}>
              <Route index element={<Navigate to="/chat" replace />} />
              <Route path="chat" element={<ChatPage />} />
              <Route path="support" element={<SupportPage />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="rag" element={<RAGPage />} />
            </Route>

            {/* Catch all route */}
            <Route path="*" element={
              localStorage.getItem('token') 
                ? <Navigate to="/chat" replace /> 
                : <Navigate to="/login" replace />
            } />
          </Routes>
        </Router>
      </ConfigProvider>
    </AntdApp>
  );
}

export default App;
