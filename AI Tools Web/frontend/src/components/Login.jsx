import React from "react";
import { Form, Input, Button, Checkbox } from "antd";
import { GoogleOutlined } from "@ant-design/icons";
import { Link } from "react-router-dom";

const Login = ({ onLogin }) => {
  const onFinish = (values) => {
    console.log("Success:", values);
    onLogin(values);
  };

  return (
    <div style={{ maxWidth: 400, width: "100%" }}>
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
          style={{ marginBottom: "12px" }}
        >
          <Input.Password 
            placeholder="Enter your password" 
            size="large"
          />
        </Form.Item>

        <div style={{ 
          display: "flex", 
          justifyContent: "space-between", 
          marginBottom: "24px",
          alignItems: "center"
        }}>
          <Form.Item name="remember" valuePropName="checked" noStyle>
            <Checkbox>Remember for 30 days</Checkbox>
          </Form.Item>
          <Link to="/forgot-password" style={{ color: "#1677ff" }}>Forgot password</Link>
        </div>

        <Form.Item>
          <Button 
            type="primary" 
            htmlType="submit" 
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
  );  
};

export default Login;