import { useState } from 'react';
import {
  Button,
  Table,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  Space,
  Tag,
  message,
  Popconfirm,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { channelsApi, mappingsApi } from '@/api/client';
import type { ModelMapping, MappingCreate, Channel } from '@/types';

const columns = (onEdit: (mapping: ModelMapping) => void, onDelete: (id: number) => void) => [
  { title: 'Display Model', dataIndex: 'display_model', key: 'display_model' },
  { title: 'Upstream Model', dataIndex: 'upstream_model', key: 'upstream_model' },
  { title: 'Channel', dataIndex: 'channel_name', key: 'channel_name' },
  {
    title: 'Status',
    dataIndex: 'is_enabled',
    key: 'is_enabled',
    render: (enabled: boolean) =>
      enabled ? <Tag color="green">Enabled</Tag> : <Tag color="red">Disabled</Tag>,
  },
  {
    title: 'Actions',
    key: 'actions',
    render: (_: unknown, record: ModelMapping) => (
      <Space>
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          Edit
        </Button>
        <Popconfirm
          title="Delete this mapping?"
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

export default function Mappings() {
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMapping, setEditingMapping] = useState<ModelMapping | null>(null);
  const [form] = Form.useForm();

  const { data: mappingsData, isLoading } = useQuery({
    queryKey: ['mappings'],
    queryFn: async () => {
      const { data } = await mappingsApi.list();
      return data;
    },
  });

  const { data: channelsData } = useQuery({
    queryKey: ['channels'],
    queryFn: async () => {
      const { data } = await channelsApi.list();
      return data;
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: MappingCreate) => mappingsApi.create(data),
    onSuccess: () => {
      message.success('Mapping created');
      queryClient.invalidateQueries({ queryKey: ['mappings'] });
      setIsModalOpen(false);
      form.resetFields();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ModelMapping> }) =>
      mappingsApi.update(id, data),
    onSuccess: () => {
      message.success('Mapping updated');
      queryClient.invalidateQueries({ queryKey: ['mappings'] });
      setIsModalOpen(false);
      setEditingMapping(null);
      form.resetFields();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => mappingsApi.delete(id),
    onSuccess: () => {
      message.success('Mapping deleted');
      queryClient.invalidateQueries({ queryKey: ['mappings'] });
    },
  });

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingMapping) {
        updateMutation.mutate({ id: editingMapping.id, data: values });
      } else {
        createMutation.mutate(values);
      }
    } catch (err) {
      // Validation failed
    }
  };

  const handleEdit = (mapping: ModelMapping) => {
    setEditingMapping(mapping);
    form.setFieldsValue(mapping);
    setIsModalOpen(true);
  };

  const handleDelete = (id: number) => {
    deleteMutation.mutate(id);
  };

  const handleCancel = () => {
    setIsModalOpen(false);
    setEditingMapping(null);
    form.resetFields();
  };

  const channelOptions = channelsData?.data
    ?.filter((c: Channel) => c.is_active)
    .map((c: Channel) => ({ label: c.name, value: c.id })) || [];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setIsModalOpen(true)}
        >
          Add Model Mapping
        </Button>
      </div>

      <Table
        loading={isLoading}
        dataSource={mappingsData?.data || []}
        columns={columns(handleEdit, handleDelete)}
        rowKey="id"
      />

      <Modal
        title={editingMapping ? 'Edit Model Mapping' : 'Add Model Mapping'}
        open={isModalOpen}
        onOk={handleModalOk}
        onCancel={handleCancel}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="Channel"
            name="channel_id"
            rules={[{ required: true, message: 'Please select a channel' }]}
          >
            <Select options={channelOptions} placeholder="Select channel" />
          </Form.Item>

          <Form.Item
            label="Display Model Name"
            name="display_model"
            rules={[{ required: true, message: 'Please enter display model name' }]}
            extra="The model name that clients will use (e.g., claude-sonnet-4-5)"
          >
            <Input placeholder="claude-sonnet-4-5" />
          </Form.Item>

          <Form.Item
            label="Upstream Model Name"
            name="upstream_model"
            rules={[{ required: true, message: 'Please enter upstream model name' }]}
            extra="The actual model name to send to the API (e.g., claude-3-5-sonnet-20241022)"
          >
            <Input placeholder="claude-3-5-sonnet-20241022" />
          </Form.Item>

          <Form.Item
            label="Enabled"
            name="is_enabled"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
