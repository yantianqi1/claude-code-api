import { Row, Col, Card, Statistic, Table } from 'antd';
import { useQuery } from '@tanstack/react-query';
import {
  CloudServerOutlined,
  CheckCircleOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { statsApi } from '@/api/client';
import type { OverallStats, ChannelStats, DailyStats } from '@/types';

const getChartOption = (data: DailyStats[], type: 'requests' | 'tokens') => ({
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: data.map((d) => d.date).reverse(),
  },
  yAxis: { type: 'value' },
  series: [
    {
      name: type === 'requests' ? 'Requests' : 'Tokens (K)',
      type: 'line',
      smooth: true,
      data: data
        .map((d) => (type === 'requests' ? d.total_requests : Math.round(d.total_tokens / 1000)))
        .reverse(),
      itemStyle: { color: type === 'requests' ? '#1890ff' : '#52c41a' },
      areaStyle: { opacity: 0.3 },
    },
  ],
});

const channelColumns = [
  { title: 'Channel', dataIndex: 'channel_name', key: 'channel_name' },
  {
    title: 'Total Requests',
    dataIndex: 'total_requests',
    key: 'total_requests',
    render: (val: number) => val.toLocaleString(),
  },
  {
    title: 'Success Rate',
    key: 'success_rate',
    render: (_: unknown, r: ChannelStats) =>
      r.total_requests > 0
        ? `${Math.round((r.success_requests / r.total_requests) * 100)}%`
        : '-',
  },
  {
    title: 'Total Tokens',
    dataIndex: 'total_tokens',
    key: 'total_tokens',
    render: (val: number) => val.toLocaleString(),
  },
  {
    title: 'Avg Latency',
    dataIndex: 'avg_latency_ms',
    key: 'avg_latency_ms',
    render: (val: number) => `${Math.round(val)}ms`,
  },
];

export default function Dashboard() {
  const { data: stats, isLoading } = useQuery<OverallStats>({
    queryKey: ['stats', 'overall'],
    queryFn: async () => {
      const { data } = await statsApi.getOverall();
      return data;
    },
  });

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Channels"
              value={stats?.total_channels || 0}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Active Channels"
              value={stats?.active_channels || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Requests"
              value={stats?.total_requests || 0}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Tokens"
              value={stats?.total_tokens || 0}
              prefix={<DollarOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="Request Trends" loading={isLoading}>
            <ReactECharts
              option={getChartOption(stats?.daily_stats || [], 'requests')}
              style={{ height: 300 }}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Token Usage Trends" loading={isLoading}>
            <ReactECharts
              option={getChartOption(stats?.daily_stats || [], 'tokens')}
              style={{ height: 300 }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Channel Statistics" style={{ marginTop: 16 }} loading={isLoading}>
        <Table
          dataSource={stats?.channel_stats || []}
          columns={channelColumns}
          rowKey="channel_id"
          pagination={false}
          size="small"
        />
      </Card>
    </div>
  );
}
