// PayPage.jsx
import React, { useState } from 'react';
import { Card, Button, Typography, Space, Radio, Tag, Row, Col, message, Modal } from 'antd';
import { ArrowLeftOutlined, CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text, Paragraph } = Typography;
const { confirm } = Modal;

const PayPage = () => {
  const navigate = useNavigate();
  const [selectedPlan, setSelectedPlan] = useState('monthly');
  
  const subscriptionPlans = [
    {
      key: 'monthly',
      title: 'Monthly',
      price: 20,
      details: 'SGD 20 per month',
      popular: false,
      features: ['Access to all AI tools', 'Monthly new features', 'Priority support']
    },
    {
      key: 'quarterly',
      title: 'Quarterly',
      price: 50,
      originalPrice: 60,
      details: 'SGD 50 per 3 months (Save SGD 10)',
      popular: true,
      features: ['Access to all AI tools', 'Monthly new features', 'Priority support', 'Custom AI model config']
    },
    {
      key: 'halfyear',
      title: '6 Months',
      price: 100,
      originalPrice: 120,
      details: 'SGD 100 per 6 months (Save SGD 20)',
      popular: false,
      features: ['Access to all AI tools', 'Monthly new features', 'Priority support', 'Custom AI model config', 'Dedicated advisor']
    },
    {
      key: 'annual',
      title: 'Annual',
      price: 210,
      originalPrice: 240,
      details: 'SGD 210 per year (Save SGD 30)',
      popular: false,
      features: ['Access to all AI tools', 'Monthly new features', 'Priority support', 'Custom AI model config', 'Dedicated advisor', 'Unlimited usage']
    }
  ];

  const handleSubmit = () => {
    confirm({
      title: 'Payment Unavailable',
      icon: <ExclamationCircleOutlined />,
      content: 'The payment feature is currently unavailable. I apologize for the inconvenience and are working to restore this functionality soon.',
      okText: 'Back to Home',
      cancelText: 'Stay on Page',
      onOk() {
        message.info('Redirecting to home page');
        navigate('/');
      },
    });
  };

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: '16px' }}>
      <Button 
        icon={<ArrowLeftOutlined />} 
        style={{ marginBottom: 16 }}
        onClick={() => navigate('/')}
      >
        Back
      </Button>
      
      <Card bordered={false} style={{ paddingBottom: 0 }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: 24 }}>Choose Subscription Plan</Title>
        
        <Radio.Group 
          value={selectedPlan} 
          onChange={(e) => setSelectedPlan(e.target.value)}
          style={{ width: '100%', marginBottom: 16 }}
        >
          <Row gutter={[16, 16]}>
            {subscriptionPlans.map(plan => (
              <Col xs={24} sm={24} md={12} lg={6} key={plan.key}>
                <Card
                  style={{ 
                    borderRadius: '8px',
                    borderColor: selectedPlan === plan.key ? '#1890ff' : '#f0f0f0',
                    backgroundColor: selectedPlan === plan.key ? '#f0f8ff' : '#f9f9f9',
                    position: 'relative',
                    paddingTop: '16px'
                  }}
                  bodyStyle={{ padding: '16px' }}
                  hoverable
                  onClick={() => setSelectedPlan(plan.key)}
                >
                  <div style={{ position: 'absolute', top: '12px', right: '12px' }}>
                    <Radio value={plan.key} />
                  </div>
                  
                  {plan.popular && (
                    <Tag color="#faad14" style={{ 
                      position: 'absolute', 
                      top: '12px', 
                      right: '45px',
                      borderRadius: '12px',
                      padding: '0 8px'
                    }}>
                      Most Popular
                    </Tag>
                  )}

                  <Title level={4} style={{ margin: '0 0 4px 0' }}>{plan.title}</Title>
                  <Text type="secondary" style={{ display: 'block', marginBottom: '8px' }}>{plan.details}</Text>

                  <div style={{ margin: '8px 0' }}>
                    <Title level={2} style={{ margin: 0, display: 'inline-block' }}>S${plan.price}</Title>
                    {plan.originalPrice && (
                      <Text delete type="secondary" style={{ marginLeft: '8px' }}>S${plan.originalPrice}</Text>
                    )}
                  </div>
                  
                  <Space direction="vertical" size="small" style={{ width: '100%', marginTop: '12px' }}>
                    {plan.features.map((feature, index) => (
                      <div key={index} style={{ display: 'flex', alignItems: 'center' }}>
                        <CheckCircleOutlined style={{ color: '#52c41a', marginRight: '8px', fontSize: '14px' }} />
                        <Text style={{ fontSize: '13px' }}>{feature}</Text>
                      </div>
                    ))}
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>
        </Radio.Group>
        
        <div style={{ textAlign: 'center', marginTop: '16px', paddingBottom: '16px' }}>
          <Space size="middle">
            <Button 
              type="primary" 
              size="large" 
              onClick={handleSubmit}
              style={{ minWidth: 160 }}
            >
              Subscribe Now
            </Button>
            
            <Button
              size="large"
              onClick={handleSubmit}
              style={{ minWidth: 160 }}
            >
              Confirm Selection
            </Button>
          </Space>
        </div>
      </Card>
    </div>
  );
};

export default PayPage;