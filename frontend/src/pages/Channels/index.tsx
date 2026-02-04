import { useState } from 'react';
import {
  Button,
  Table,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Tag,
  message,
  Popconfirm,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PoweroffOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { channelsApi } from '@/api/client';
import type { Channel, ChannelCreate } from '@/types';

const providers = [
  { label: 'Anthropic', value: 'anthropic' },
  { label: 'OpenAI', value: 'openai' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: '自定义', value: 'custom' },
];

const columns = (onEdit: (channel: Channel) => void, onDelete: (id: number) => void) => [
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: '提供商', dataIndex: 'provider', key: 'provider' },
  { title: '基础地址', dataIndex: 'base_url', key: 'base_url', ellipsis: true },
  {
    title: '状态',
    dataIndex: 'is_active',
    key: 'is_active',
    render: (active: boolean) =>
      active ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>,
  },
  { title: '优先级', dataIndex: 'priority', key: 'priority' },
  { title: '超时(秒)', dataIndex: 'timeout', key: 'timeout' },
  {
    title: '操作',
    key: 'actions',
    render: (_: unknown, record: Channel) => (
      <Space>
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          编辑
        </Button>
        <Button
          type="link"
          icon={<PoweroffOutlined />}
          onClick={() => {
            if (record.is_active) {
              channelsApi.deactivate(record.id);
            } else {
              channelsApi.activate(record.id);
            }
            window.location.reload();
          }}
        >
          {record.is_active ? '禁用' : '启用'}
        </Button>
        <Popconfirm
          title="确定删除此渠道吗？"
          onConfirm={() => onDelete(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      </Space>
    ),
  },
];

export default function Channels() {
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingChannel, setEditingChannel] = useState<Channel | null>(null);
  const [form] = Form.useForm();

  const { data: channelsData, isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: async () => {
      const { data } = await channelsApi.list();
      return data;
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: ChannelCreate) => channelsApi.create(data),
    onSuccess: () => {
      message.success('渠道创建成功');
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setIsModalOpen(false);
      form.resetFields();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Channel> }) =>
      channelsApi.update(id, data),
    onSuccess: () => {
      message.success('渠道更新成功');
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setIsModalOpen(false);
      setEditingChannel(null);
      form.resetFields();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => channelsApi.delete(id),
    onSuccess: () => {
      message.success('渠道删除成功');
      queryClient.invalidateQueries({ queryKey: ['channels'] });
    },
  });

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingChannel) {
        updateMutation.mutate({ id: editingChannel.id, data: values });
      } else {
        createMutation.mutate(values);
      }
    } catch (err) {
      // Validation failed
    }
  };

  const handleEdit = (channel: Channel) => {
    setEditingChannel(channel);
    form.setFieldsValue(channel);
    setIsModalOpen(true);
  };

  const handleDelete = (id: number) => {
    deleteMutation.mutate(id);
  };

  const handleCancel = () => {
    setIsModalOpen(false);
    setEditingChannel(null);
    form.resetFields();
  };

  const handleTest = async () => {
    const baseUrl = form.getFieldValue('base_url');
    const apiKey = form.getFieldValue('api_key');
    if (!baseUrl || !apiKey) {
      message.warning('请先输入基础地址和API密钥');
      return;
    }
    try {
      await channelsApi.test(baseUrl, apiKey);
      message.success('连接测试成功');
    } catch (err) {
      message.error('连接测试失败');
    }
  };

  const cardStyle = {
    borderRadius: '12px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
    border: 'none',
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsModalOpen(true)}
          style={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
            borderRadius: '8px',
            height: '40px',
            fontSize: '14px',
          }}
        >
          添加渠道
        </Button>
      </div>

      <Table
        loading={isLoading}
        dataSource={channelsData?.data || []}
        columns={columns(handleEdit, handleDelete)}
        rowKey="id"
        style={cardStyle}
      />

      <Modal
        title={editingChannel ? '编辑渠道' : '添加渠道'}
        open={isModalOpen}
        onOk={handleModalOk}
        onCancel={handleCancel}
        width={600}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        okButtonProps={{
          style: {
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
          },
        }}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="渠道名称"
            name="name"
            rules={[{ required: true, message: '请输入渠道名称' }]}
          >
            <Input placeholder="我的 Anthropic 渠道" />
          </Form.Item>

          <Form.Item
            label="提供商"
            name="provider"
            rules={[{ required: true, message: '请选择提供商' }]}
          >
            <Select options={providers} placeholder="选择提供商" />
          </Form.Item>

          <Form.Item
            label="基础地址"
            name="base_url"
            rules={[{ required: true, message: '请输入基础地址' }]}
          >
            <Input placeholder="https://api.anthropic.com" />
          </Form.Item>

          <Form.Item
            label="API 密钥"
            name="api_key"
            rules={[{ required: true, message: '请输入API密钥' }]}
          >
            <Input.Password placeholder="sk-ant-..." />
          </Form.Item>

          <Form.Item label="优先级" name="priority" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} placeholder="数字越小优先级越高" />
          </Form.Item>

          <Form.Item label="超时时间（秒）" name="timeout" initialValue={60}>
            <InputNumber min={1} max={300} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="最大重试次数" name="max_retries" initialValue={3}>
            <InputNumber min={0} max={10} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="速率限制（0 = 无限制）" name="rate_limit" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item>
            <Button
              onClick={handleTest}
              block
              style={{ borderRadius: '8px' }}
            >
              测试连接
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
