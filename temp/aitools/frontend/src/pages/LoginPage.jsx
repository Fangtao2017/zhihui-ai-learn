import React, { useState, useEffect } from 'react';
import { Form, Input, Button, Checkbox, message } from 'antd';
import { GoogleOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import request from '../utils/request';

const API_BASE_URL = 'http://localhost:8080';

const LoginPage = () => {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    // Check if token is expired
    const checkTokenExpiration = () => {
        const token = localStorage.getItem('token');
        if (token) {
            try {
                const base64Url = token.split('.')[1];
                const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
                const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
                    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
                }).join(''));

                const { exp } = JSON.parse(jsonPayload);
                if (exp * 1000 > Date.now()) {
                    // Token is not expired, redirect to chat page
                    navigate('/chat');
                } else {
                    // Token has expired, clear local storage
                    localStorage.removeItem('token');
                    localStorage.removeItem('username');
                }
            } catch (error) {
                // Token parsing failed, clear local storage
                localStorage.removeItem('token');
                localStorage.removeItem('username');
            }
        }
    };

    useEffect(() => {
        checkTokenExpiration();
    }, [navigate]);

    const onFinish = async (values) => {
        try {
            setLoading(true);
            const response = await request(`${API_BASE_URL}/api/login`, {
                method: 'POST',
                body: JSON.stringify({
                    email: values.email,
                    password: values.password
                })
            });

            if (response.token) {
                await Promise.all([
                    localStorage.setItem('token', response.token),
                    localStorage.setItem('username', response.username || values.email)
                ]);
                
                // Create a new conversation and get its ID after successful login
                try {
                    const chatResponse = await request(`${API_BASE_URL}/api/chat/new`, {
                        method: 'POST',
                    });
                    if (chatResponse && chatResponse.id) {
                        localStorage.setItem('currentChatId', chatResponse.id);
                    }
                } catch (error) {
                    console.error('Failed to create initial chat:', error);
                }
                
                navigate('/chat', { replace: true });
            }
        } catch (error) {
            message.error('Login failed: ' + error.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ 
            display: 'flex',
            minHeight: '100vh',
            width: '100vw',
            overflow: 'hidden'
        }}>
            <div style={{
                width: '50%',
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                backgroundColor: '#ffffff',
                padding: '0 100px'
            }}>
                <div style={{ width: '100%', maxWidth: '400px' }}>
                    <h1 style={{ 
                        fontSize: "28px", 
                        marginBottom: "8px",
                        fontWeight: "600"
                    }}>Welcome back!</h1>
                    <p style={{ 
                        marginBottom: "32px", 
                        color: "#666",
                        fontSize: "14px"
                    }}>
                        Enter your Credentials to access your account
                    </p>

                    <Form layout="vertical" onFinish={onFinish}>
                        <Form.Item
                            label={<>Email address <span style={{ color: '#ff4d4f' }}>*</span></>}
                            name="email"
                            validateTrigger="onBlur"
                            rules={[
                                { required: true, message: "Please input your email!" },
                                { type: "email", message: "Please enter a valid email!" }
                            ]}
                        >
                            <Input 
                                placeholder="Enter your email" 
                                size="large"
                            />
                        </Form.Item>

                        <Form.Item
                            label={<>Password <span style={{ color: '#ff4d4f' }}>*</span></>}
                            name="password"
                            validateTrigger="onBlur"
                            rules={[{ required: true, message: "Please input your password!" }]}
                        >
                            <Input.Password 
                                placeholder="Enter your password" 
                                size="large"
                            />
                        </Form.Item>

                        <div style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            marginBottom: '24px'
                        }}>
                            <Form.Item name="remember" valuePropName="checked" noStyle>
                                <Checkbox>Remember for 30 days</Checkbox>
                            </Form.Item>
                            <Link to="/forgot-password" style={{ color: '#1677ff' }}>
                                Forgot password
                            </Link>
                        </div>

                        <Form.Item>
                            <Button 
                                type="primary" 
                                htmlType="submit" 
                                loading={loading}
                                block 
                                size="large"
                                style={{ 
                                    backgroundColor: "#2B7A0B",
                                    height: "44px",
                                    borderRadius: "8px",
                                    fontWeight: "500"
                                }}
                            >
                                Sign in
                            </Button>
                        </Form.Item>

                        <div style={{ 
                            textAlign: "center", 
                            margin: "24px 0",
                            color: "#666",
                            position: "relative"
                        }}>
                            <span style={{
                                backgroundColor: "#fff",
                                padding: "0 10px",
                                zIndex: 1,
                                position: "relative"
                            }}>Or</span>
                            <div style={{
                                position: "absolute",
                                top: "50%",
                                left: 0,
                                right: 0,
                                height: "1px",
                                backgroundColor: "#e8e8e8",
                                zIndex: 0
                            }}/>
                        </div>

                        <Button 
                            icon={<GoogleOutlined />}
                            size="large"
                            block
                            style={{ 
                                height: "44px",
                                borderRadius: "8px",
                                border: "1px solid #e2e8f0",
                            }}
                        >
                            Sign in with Google
                        </Button>

                        <div style={{ textAlign: "center", marginTop: "24px" }}>
                            Don't have an account? <Link to="/signup" style={{ color: "#1677ff" }}>Sign up</Link>
                        </div>
                    </Form>
                </div>
            </div>
            
            <div style={{
                width: '50%',
                backgroundImage: `url('https://www.singaporetech.edu.sg/sites/default/files/2024-08/SIT%20Punggol%20Campus_Campus%20Court_University%20Tower.jpg')`,
                backgroundSize: 'cover',
                backgroundPosition: 'center'
            }} />
        </div>
    );
};

export default LoginPage;