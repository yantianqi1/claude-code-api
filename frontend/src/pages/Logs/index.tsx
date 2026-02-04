import { useState, useEffect } from 'react';
import { Table, Tag, DatePicker, Button, Space, Select } from 'antd';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import dayjs, { Dayjs } from 'dayjs';
import { statsApi } from '@/api/client';
import type { RequestLog, StatsFilter } from '@/types';

const { RangePicker } = DatePicker;

const columns = [
  { title: 'Time', dataIndex: 'request_time', key: 'request_time', render: (t: string) => dayjs(t).format('MM-DD HH:mm:ss') },
  { title: 'Channel', dataIndex: 'channel_name', key: 'channel_name' },
  { title: 'Model', dataIndex: 'model_name', key: 'model_name' },
  {
    title: 'Tokens',
    key: 'tokens',
    render: (_: unknown, r: RequestLog) => `${r.input_tokens}→${r.output_tokens}`,
  },
  { title: 'Total Tokens', dataIndex: 'total_tokens', key: 'total_tokens' },
  { title: 'Latency', dataIndex: 'latency_ms', key: 'latency_ms', render: (v: number) => `${v}ms` },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    render: (status: string) =>
      status === 'success' ? (
        <Tag color="green">Success</Tag>
      ) : (
        <Tag color="red">Error</Tag>
      ),
  },
  {
    title: 'Error',
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
        />
        <Select
          placeholder="Status"
          allowClear
          style={{ width: 120 }}
          onChange={(value) => setFilter({ ...filter, status: value })}
        >
          <Select.Option value="success">Success</Select.Option>
          <Select.Option value="error">Error</Select.Option>
        </Select>
        <Button icon={<ReloadOutlined />} onClick={() => refetch()}>
          Refresh
        </Button>
        <Button icon={<DownloadOutlined />} onClick={handleExport}>
          Export CSV
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
          showTotal: (t) => `Total ${t} logs`,
          onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
        }}
        scroll={{ x: 1000 }}
      />
    </div>
  );
}
