import React from 'react';
import { Layout, Menu, Image } from 'antd';
import {
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  FileSearchOutlined
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
const { Header, Content } = Layout;

const HomePage = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    // Clear all authentication information
    localStorage.removeItem('username');
    localStorage.removeItem('token');
    navigate('/login');
  };

  // Check authentication status
  React.useEffect(() => {
    const checkAuth = () => {
      const token = localStorage.getItem('token');
      if (!token) {
        navigate('/login', { replace: true });
        return;
      }

      try {
        // Check if token is expired
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join(''));

        const { exp } = JSON.parse(jsonPayload);
        if (exp * 1000 <= Date.now()) {
          // Token expired, clear local storage and redirect to login page
          localStorage.removeItem('token');
          localStorage.removeItem('username');
          navigate('/login', { replace: true });
          return;
        }

        // Token is valid and not expired, if at root path, redirect to chat page
        if (location.pathname === '/') {
          navigate('/chat', { replace: true });
        }
      } catch (error) {
        // Token parsing failed, clear local storage and redirect to login page
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        navigate('/login', { replace: true });
      }
    };

    checkAuth();
  }, [navigate, location]);

  const username = localStorage.getItem('username');

  const menuItems = [
    { key: '', label: 'Chat', onClick: () => navigate('/') },
    { key: 'rag', label: 'Knowledge Base', icon: <FileSearchOutlined />, onClick: () => navigate('/rag') },
    { key: 'support', label: 'Support', onClick: () => navigate('/support') },
    { key: 'settings', label: 'Settings', onClick: () => navigate('/settings') },
  ];

  // Get current path
  const currentPath = location.pathname.slice(1) || '';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ 
        padding: '0 24px', 
        background: '#fff', 
        display: 'flex', 
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
        position: 'sticky',
        top: 0,
        zIndex: 1,
        width: '100%'
      }}>
        {/* Header content remains unchanged */}
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <div style={{ 
            display: 'flex',
            alignItems: 'center',
            marginRight: '48px'
          }}>
            <span style={{ 
              fontSize: '20px', 
              fontWeight: 'bold',
              color: '#2B7A0B',
            }}>
              ZHIHUI AI LEARN
            </span>
          </div>
          <Menu mode="horizontal" selectedKeys={[currentPath]}>
            {menuItems.map(item => (
              <Menu.Item key={item.key} onClick={item.onClick}>
                {item.label}
              </Menu.Item>
            ))}
          </Menu>
        </div>
        <Menu mode="horizontal" selectedKeys={[]}>
          <Menu.SubMenu 
            key="user" 
            icon={<UserOutlined />}
            title={username}
          >
            <Menu.Item 
              key="settings" 
              icon={<SettingOutlined />}
              onClick={() => navigate('/settings')}
            >
              Settings
            </Menu.Item>
            <Menu.Item 
              key="logout" 
              icon={<LogoutOutlined />}
              onClick={handleLogout}
            >
              Logout
            </Menu.Item>
          </Menu.SubMenu>
        </Menu>
      </Header>

      <Layout>
        <Content style={{ 
          padding: '24px',
          minHeight: 'calc(100vh - 64px)',
          background: '#fff'
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default HomePage;