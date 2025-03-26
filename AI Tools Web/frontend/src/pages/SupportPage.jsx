import React, { useState } from 'react';
import { Tabs, Collapse, Form, Input, Select, Button, Upload, Typography, Card, Space, Divider, message } from 'antd';
import { QuestionCircleOutlined, FileTextOutlined, MailOutlined, BugOutlined, CloudUploadOutlined } from '@ant-design/icons';

const { TabPane } = Tabs;
const { Panel } = Collapse;
const { Title, Paragraph, Text, Link } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const SupportPage = () => {
  const [feedbackForm] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  
  const handleSubmit = (values) => {
    setSubmitting(true);
    // 这里添加提交反馈的逻辑
    console.log('Feedback submitted:', values);
    
    // 模拟API调用
    setTimeout(() => {
      setSubmitting(false);
      feedbackForm.resetFields();
      // 显示成功消息
      message.success('Thank you for your feedback! We will review it shortly.');
    }, 1000);
  };

  // 每个标签页内容的样式，确保可滚动
  const tabContentStyle = {
    maxHeight: 'calc(100vh - 220px)',
    overflowY: 'auto',
    paddingRight: '10px'
  };

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto', padding: '0 16px' }}>
      <Title level={2}>Support Center</Title>
      
      <Tabs defaultActiveKey="faq" style={{ overflow: 'visible' }}>
        <TabPane 
          tab={<span><QuestionCircleOutlined /> Frequently Asked Questions</span>} 
          key="faq"
        >
          <div style={tabContentStyle}>
            <Collapse defaultActiveKey={['1']}>
              <Panel header="Using AI Chat" key="1">
                <Paragraph>
                  <Text strong>How do I get better responses from the AI?</Text>
                  <br />
                  Be specific in your questions and provide context. The more detailed your query, the more accurate the response.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>What languages are supported?</Text>
                  <br />
                  Our AI supports English and Chinese for optimal results.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>Is there a limit to how many questions I can ask?</Text>
                  <br />
                  Usage limits depend on your subscription plan. Free users have a daily limit, while paid subscribers have higher or unlimited usage.
                </Paragraph>
              </Panel>
              
              <Panel header="Knowledge Base Management" key="2">
                <Paragraph>
                  <Text strong>What file formats can I upload?</Text>
                  <br />
                  We support PDF, DOCX, TXT, and CSV files for knowledge base processing.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>How do I clear my vector database?</Text>
                  <br />
                  Go to Settings and click "Clear Vector Database". This will remove all stored vector data.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>How large can my uploaded files be?</Text>
                  <br />
                  Individual files should be under 20MB. For larger documents, consider splitting them into smaller files.
                </Paragraph>
              </Panel>
              
              <Panel header="Account Management" key="4">
                <Paragraph>
                  <Text strong>How do I change my password?</Text>
                  <br />
                  Go to Settings page and click "Change Password". You'll need to enter your current password for verification.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>Is my data secure?</Text>
                  <br />
                  Yes, we use industry-standard encryption and security practices to protect your data and conversations.
                </Paragraph>
                
                <Paragraph>
                  <Text strong>Can I export my chat history?</Text>
                  <br />
                  Currently, we don't offer a direct export function, but this feature is on our roadmap.
                </Paragraph>
              </Panel>
            </Collapse>
          </div>
        </TabPane>
        
        <TabPane 
          tab={<span><FileTextOutlined /> User Guides</span>} 
          key="guides"
        >
          <div style={tabContentStyle}>
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <Card title="Getting Started with FangTao AI">
                <Paragraph>
                  A comprehensive guide to help you make the most of our AI platform.
                </Paragraph>
                <Button type="primary">View Guide</Button>
              </Card>
              
              <Card title="Knowledge Base Tutorial">
                <Paragraph>
                  Learn how to effectively upload and query your documents.
                </Paragraph>
                <Button type="primary">View Tutorial</Button>
              </Card>
              
              <Card title="Advanced AI Prompting Techniques">
                <Paragraph>
                  Master the art of crafting prompts to get the best AI responses.
                </Paragraph>
                <Button type="primary">View Guide</Button>
              </Card>
            </Space>
          </div>
        </TabPane>
        
        <TabPane 
          tab={<span><BugOutlined /> Submit Feedback</span>} 
          key="feedback"
        >
          <div style={tabContentStyle}>
            <Form 
              form={feedbackForm}
              layout="vertical" 
              onFinish={handleSubmit}
            >
              <Form.Item name="type" label="Feedback Type" rules={[{ required: true, message: 'Please select a feedback type' }]}>
                <Select placeholder="Select feedback type">
                  <Option value="bug">Bug Report</Option>
                  <Option value="feature">Feature Request</Option>
                  <Option value="account">Account Issue</Option>
                  <Option value="other">Other</Option>
                </Select>
              </Form.Item>
              
              <Form.Item name="subject" label="Subject" rules={[{ required: true, message: 'Please enter a subject' }]}>
                <Input placeholder="Brief description of your issue" />
              </Form.Item>
              
              <Form.Item name="details" label="Details" rules={[{ required: true, message: 'Please provide details' }]}>
                <TextArea rows={6} placeholder="Please provide as much detail as possible" />
              </Form.Item>
              
              <Form.Item name="attachment" label="Attachments (Optional)">
                <Upload maxCount={3} beforeUpload={() => false}>
                  <Button icon={<CloudUploadOutlined />}>Attach Screenshots</Button>
                </Upload>
                <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
                  You can attach up to 3 files (images preferred, max 5MB each)
                </Text>
              </Form.Item>
              
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={submitting}>
                  Submit Feedback
                </Button>
              </Form.Item>
            </Form>
          </div>
        </TabPane>
        
        <TabPane 
          tab={<span><MailOutlined /> Contact Us</span>} 
          key="contact"
        >
          <div style={tabContentStyle}>
            <Card>
              <Title level={4}>Technical Support</Title>
              <Paragraph>
                Email: support@fangtao-ai.com<br />
                Hours: Monday-Friday, 9am-6pm SGT<br />
                Response Time: Within 24 hours
              </Paragraph>
              
              <Divider />
              
              <Title level={4}>System Status</Title>
              <Paragraph>
                <Text type="success">All systems operational</Text>
              </Paragraph>
              <Link href="#status">View detailed system status</Link>
            </Card>
          </div>
        </TabPane>
      </Tabs>
      
      <Divider />
      
      <div style={{ textAlign: 'center', margin: '24px 0' }}>
        <Space>
          <Link href="/terms">Terms of Service</Link>
          <Link href="/privacy">Privacy Policy</Link>
          <Link href="/data">Data Processing</Link>
        </Space>
      </div>
    </div>
  );
};

export default SupportPage;