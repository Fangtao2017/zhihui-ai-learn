import React from 'react';
import { Card, Typography, Space, Button } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

const RechargePage = () => {
  const navigate = useNavigate();
  
  const rechargeOptions = [
    { amount: 10, sgd: 13.50 },
    { amount: 20, sgd: 27.00 },
    { amount: 50, sgd: 67.50 },
    { amount: 100, sgd: 135.00 }
  ];

  const handleRecharge = (amount) => {
    // 处理充值逻辑
    console.log(`Recharging $${amount}`);
  };

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px' }}>
      <Button 
        icon={<ArrowLeftOutlined />} 
        style={{ marginBottom: 24 }}
        onClick={() => navigate('/pay')}
      >
        Back
      </Button>
      
      <Title level={2}>Select Recharge Amount</Title>
      
      <Space size="large" wrap style={{ marginTop: 24 }}>
        {rechargeOptions.map(({ amount, sgd }) => (
          <Card
            key={amount}
            hoverable
            style={{ 
              width: 240,
              textAlign: 'center',
              cursor: 'pointer'
            }}
            onClick={() => handleRecharge(amount)}
          >
            <Space direction="vertical" size="large">
              <Title level={2} style={{ margin: 0 }}>${amount}</Title>
              <Text type="secondary">≈ S${sgd.toFixed(2)}</Text>
              <Button type="primary" size="large" block>
                Pay with PayNow
              </Button>
            </Space>
          </Card>
        ))}
      </Space>
    </div>
  );
};

export default RechargePage;