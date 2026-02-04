import { useState, useEffect } from 'react';
import { Table, Tag, DatePicker, Button, Space, Select } from 'antd';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import dayjs, { Dayjs } from 'dayjs';
import { statsApi } from '@/api/client';
import type { RequestLog, StatsFilter } from '@/types';
import locale from 'antd/es/date-picker/locale/zh_CN';

const { RangePicker } = DatePicker;

const columns = [
  { title: '时间', dataIndex: 'request_time', key: 'request_time', render: (t: string) => dayjs(t).format('MM-DD HH:mm:ss') },
  { title: '渠道', dataIndex: 'channel_name', key: 'channel_name' },
  { title: '模型', dataIndex: 'model_name', key: 'model_name' },
  {
    title: 'Token流向',
    key: 'tokens',
    render: (_: unknown, r: RequestLog) => `${r.input_tokens}→${r.output_tokens}`,
  },
  { title: '总Token数', dataIndex: 'total_tokens', key: 'total_tokens' },
  { title: '延迟', dataIndex: 'latency_ms', key: 'latency_ms', render: (v: number) => `${v}ms` },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) =>
      status === 'success' ? (
        <Tag color="green">成功</Tag>
      ) : (
        <Tag color="red">失败</Tag>
      ),
  },
  {
    title: '错误信息',
    dataIndex: 'error_message',
    key: 'error_message',
    ellipsis: true,
    render: (msg: string) => msg || '-',
  },
];

export default function Logs() {
  const [filter, setFilter] = useState<StatsFilter>({});
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20 });
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([
    dayjs().subtract(7, 'day'),
    dayjs(),
  ]);

  // Initialize filter with date range on mount
  useEffect(() => {
    setFilter({
      start_date: dateRange[0].format('YYYY-MM-DD'),
      end_date: dateRange[1].format('YYYY-MM-DD'),
    });
  }, []);

  const { data: logsData, isLoading, refetch } = useQuery({
    queryKey: ['logs', filter, pagination.current, pagination.pageSize],
    queryFn: async () => {
      const { data } = await statsApi.getLogs(
        filter,
        pagination.current,
        pagination.pageSize
      );
      return data;
    },
  });

  const handleDateChange = (dates: any) => {
    if (dates && dates[0] && dates[1]) {
      setDateRange([dates[0], dates[1]]);
      setFilter({
        ...filter,
        start_date: dates[0].format('YYYY-MM-DD'),
        end_date: dates[1].format('YYYY-MM-DD'),
      });
    }
  };

  const handleExport = async () => {
    try {
      const { data } = await statsApi.export(filter);
      const url = URL.createObjectURL(new Blob([data], { type: 'text/csv' }));
      const link = document.createElement('a');
      link.href = url;
      link.download = `logs_${Date.now()}.csv`;
      link.click();
    } catch (err) {
      console.error('Export failed', err);
    }
  };

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <RangePicker
          value={dateRange}
          onChange={handleDateChange}
          allowClear={false}
          locale={locale}
        />
        <Select
          placeholder="状态筛选"
          allowClear
          style={{ width: 120 }}
          onChange={(value) => setFilter({ ...filter, status: value })}
        >
          <Select.Option value="success">成功</Select.Option>
          <Select.Option value="error">失败</Select.Option>
        </Select>
        <Button
          icon={<ReloadOutlined />}
          onClick={() => refetch()}
          style={{ borderRadius: '8px' }}
        >
          刷新
        </Button>
        <Button
          icon={<DownloadOutlined />}
          onClick={handleExport}
          style={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
            color: '#fff',
            borderRadius: '8px',
          }}
        >
          导出CSV
        </Button>
      </Space>

      <Table
        loading={isLoading}
        dataSource={logsData?.data || []}
        columns={columns}
        rowKey="id"
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: logsData?.total || 0,
          showSizeChanger: true,
          showTotal: (t) => `共 ${t} 条记录`,
          onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
        }}
        scroll={{ x: 1000 }}
      />
    </div>
  );
}
