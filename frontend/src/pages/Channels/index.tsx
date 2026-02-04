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
  { label: 'Custom', value: 'custom' },
];

const columns = (onEdit: (channel: Channel) => void, onDelete: (id: number) => void) => [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Provider', dataIndex: 'provider', key: 'provider' },
  { title: 'Base URL', dataIndex: 'base_url', key: 'base_url', ellipsis: true },
  {
    title: 'Status',
    dataIndex: 'is_active',
    key: 'is_active',
    render: (active: boolean) =>
      active ? <Tag color="green">Active</Tag> : <Tag color="red">Inactive</Tag>,
  },
  { title: 'Priority', dataIndex: 'priority', key: 'priority' },
  { title: 'Timeout (s)', dataIndex: 'timeout', key: 'timeout' },
  {
    title: 'Actions',
    key: 'actions',
    render: (_: unknown, record: Channel) => (
      <Space>
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          Edit
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
          {record.is_active ? 'Disable' : 'Enable'}
        </Button>
        <Popconfirm
          title="Delete this channel?"
          onConfirm={() => onDelete(record.id)}
          okText="Yes"
          cancelText="No"
        >
          <Button type="link" danger icon={<DeleteOutlined />}>
            Delete
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
      message.success('Channel created');
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setIsModalOpen(false);
      form.resetFields();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Channel> }) =>
      channelsApi.update(id, data),
    onSuccess: () => {
      message.success('Channel updated');
      queryClient.invalidateQueries({ queryKey: ['channels'] });
      setIsModalOpen(false);
      setEditingChannel(null);
      form.resetFields();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => channelsApi.delete(id),
    onSuccess: () => {
      message.success('Channel deleted');
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
      message.warning('Please enter base URL and API key first');
      return;
    }
    try {
      await channelsApi.test(baseUrl, apiKey);
      message.success('Connection test successful');
    } catch (err) {
      message.error('Connection test failed');
    }
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsModalOpen(true)}
        >
          Add Channel
        </Button>
      </div>

      <Table
        loading={isLoading}
        dataSource={channelsData?.data || []}
        columns={columns(handleEdit, handleDelete)}
        rowKey="id"
      />

      <Modal
        title={editingChannel ? 'Edit Channel' : 'Add Channel'}
        open={isModalOpen}
        onOk={handleModalOk}
        onCancel={handleCancel}
        width={600}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="Channel Name"
            name="name"
            rules={[{ required: true, message: 'Please enter channel name' }]}
          >
            <Input placeholder="My Anthropic Channel" />
          </Form.Item>

          <Form.Item
            label="Provider"
            name="provider"
            rules={[{ required: true, message: 'Please select provider' }]}
          >
            <Select options={providers} />
          </Form.Item>

          <Form.Item
            label="Base URL"
            name="base_url"
            rules={[{ required: true, message: 'Please enter base URL' }]}
          >
            <Input placeholder="https://api.anthropic.com" />
          </Form.Item>

          <Form.Item
            label="API Key"
            name="api_key"
            rules={[{ required: true, message: 'Please enter API key' }]}
          >
            <Input.Password placeholder="sk-ant-..." />
          </Form.Item>

          <Form.Item label="Priority" name="priority" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="Timeout (seconds)" name="timeout" initialValue={60}>
            <InputNumber min={1} max={300} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="Max Retries" name="max_retries" initialValue={3}>
            <InputNumber min={0} max={10} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item label="Rate Limit (0 = unlimited)" name="rate_limit" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item>
            <Button onClick={handleTest} block>
              Test Connection
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
