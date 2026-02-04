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
  { title: '显示模型', dataIndex: 'display_model', key: 'display_model' },
  { title: '上游模型', dataIndex: 'upstream_model', key: 'upstream_model' },
  { title: '渠道', dataIndex: 'channel_name', key: 'channel_name' },
  {
    title: '状态',
    dataIndex: 'is_enabled',
    key: 'is_enabled',
    render: (enabled: boolean) =>
      enabled ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>,
  },
  {
    title: '操作',
    key: 'actions',
    render: (_: unknown, record: ModelMapping) => (
      <Space>
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          编辑
        </Button>
        <Popconfirm
          title="确定删除此映射吗？"
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
      message.success('模型映射创建成功');
      queryClient.invalidateQueries({ queryKey: ['mappings'] });
      setIsModalOpen(false);
      form.resetFields();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ModelMapping> }) =>
      mappingsApi.update(id, data),
    onSuccess: () => {
      message.success('模型映射更新成功');
      queryClient.invalidateQueries({ queryKey: ['mappings'] });
      setIsModalOpen(false);
      setEditingMapping(null);
      form.resetFields();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => mappingsApi.delete(id),
    onSuccess: () => {
      message.success('模型映射删除成功');
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
          style={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
            borderRadius: '8px',
            height: '40px',
            fontSize: '14px',
          }}
        >
          添加模型映射
        </Button>
      </div>

      <Table
        loading={isLoading}
        dataSource={mappingsData?.data || []}
        columns={columns(handleEdit, handleDelete)}
        rowKey="id"
      />

      <Modal
        title={editingMapping ? '编辑模型映射' : '添加模型映射'}
        open={isModalOpen}
        onOk={handleModalOk}
        onCancel={handleCancel}
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
            label="渠道"
            name="channel_id"
            rules={[{ required: true, message: '请选择一个渠道' }]}
          >
            <Select options={channelOptions} placeholder="选择渠道" />
          </Form.Item>

          <Form.Item
            label="显示模型名称"
            name="display_model"
            rules={[{ required: true, message: '请输入显示模型名称' }]}
            extra="客户端将使用的模型名称（例如：claude-sonnet-4-5）"
          >
            <Input placeholder="claude-sonnet-4-5" />
          </Form.Item>

          <Form.Item
            label="上游模型名称"
            name="upstream_model"
            rules={[{ required: true, message: '请输入上游模型名称' }]}
            extra="实际发送给API的模型名称（例如：claude-3-5-sonnet-20241022）"
          >
            <Input placeholder="claude-3-5-sonnet-20241022" />
          </Form.Item>

          <Form.Item
            label="启用状态"
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
