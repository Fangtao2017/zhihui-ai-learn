import React from 'react';
import SignUp from '../components/SignUp';
import { useNavigate } from 'react-router-dom';
import { message } from 'antd';

const SignUpPage = () => {
  const navigate = useNavigate();

  const handleSignUp = async (values) => {
    try {
      const response = await fetch('http://localhost:8080/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: values.username,
          email: values.email,
          password: values.password
        })
      });

      if (response.ok) {
        message.success('Registration successful!');
        navigate('/login');
      } else {
        const errorData = await response.text();
        message.error(errorData || 'Registration failed');
      }
    } catch (error) {
      console.error('Registration failed:', error);
      message.error('Registration failed, please check your network connection');
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
          <SignUp onSignUp={handleSignUp} />
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

export default SignUpPage;