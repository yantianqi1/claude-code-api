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
  tooltip: {
    trigger: 'axis',
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    borderColor: '#667eea',
    textStyle: { color: '#333' },
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },
  xAxis: {
    type: 'category',
    data: data.map((d) => d.date).reverse(),
    axisLine: { lineStyle: { color: '#e8e8e8' } },
    axisLabel: { color: '#666' },
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#e8e8e8' } },
    axisLabel: { color: '#666' },
    splitLine: { lineStyle: { color: '#f0f0f0' } },
  },
  series: [
    {
      name: type === 'requests' ? '请求数' : 'Token数 (K)',
      type: 'line',
      smooth: true,
      data: data
        .map((d) => (type === 'requests' ? d.total_requests : Math.round(d.total_tokens / 1000)))
        .reverse(),
      itemStyle: { color: '#667eea' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(102, 126, 234, 0.4)' },
            { offset: 1, color: 'rgba(102, 126, 234, 0.05)' },
          ],
        },
      },
      lineStyle: { width: 3 },
    },
  ],
});

const channelColumns = [
  { title: '渠道名称', dataIndex: 'channel_name', key: 'channel_name' },
  {
    title: '总请求数',
    dataIndex: 'total_requests',
    key: 'total_requests',
    render: (val: number) => val.toLocaleString(),
  },
  {
    title: '成功率',
    key: 'success_rate',
    render: (_: unknown, r: ChannelStats) =>
      r.total_requests > 0
        ? `${Math.round((r.success_requests / r.total_requests) * 100)}%`
        : '-',
  },
  {
    title: '总Token数',
    dataIndex: 'total_tokens',
    key: 'total_tokens',
    render: (val: number) => val.toLocaleString(),
  },
  {
    title: '平均延迟',
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

  const cardStyle = {
    borderRadius: '12px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
    border: 'none',
  };

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading} style={cardStyle}>
            <Statistic
              title="总渠道数"
              value={stats?.total_channels || 0}
              prefix={<CloudServerOutlined style={{ color: '#667eea' }} />}
              valueStyle={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 600,
                fontSize: '28px',
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading} style={cardStyle}>
            <Statistic
              title="活跃渠道"
              value={stats?.active_channels || 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{
                background: 'linear-gradient(135deg, #52c41a 0%, #389e0d 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 600,
                fontSize: '28px',
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading} style={cardStyle}>
            <Statistic
              title="总请求数"
              value={stats?.total_requests || 0}
              valueStyle={{
                background: 'linear-gradient(135deg, #722ed1 0%, #531dab 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 600,
                fontSize: '28px',
              }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card loading={isLoading} style={cardStyle}>
            <Statistic
              title="总Token数"
              value={stats?.total_tokens || 0}
              prefix={<DollarOutlined style={{ color: '#fa8c16' }} />}
              valueStyle={{
                background: 'linear-gradient(135deg, #fa8c16 0%, #d46b08 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                fontWeight: 600,
                fontSize: '28px',
              }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card
            title={
              <span style={{ fontSize: '16px', fontWeight: 600 }}>请求趋势</span>
            }
            loading={isLoading}
            style={cardStyle}
          >
            <ReactECharts
              option={getChartOption(stats?.daily_stats || [], 'requests')}
              style={{ height: 300 }}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card
            title={
              <span style={{ fontSize: '16px', fontWeight: 600 }}>Token 使用趋势</span>
            }
            loading={isLoading}
            style={cardStyle}
          >
            <ReactECharts
              option={getChartOption(stats?.daily_stats || [], 'tokens')}
              style={{ height: 300 }}
            />
          </Card>
        </Col>
      </Row>

      <Card
        title={<span style={{ fontSize: '16px', fontWeight: 600 }}>渠道统计</span>}
        style={{ marginTop: 24, ...cardStyle }}
        loading={isLoading}
      >
        <Table
          dataSource={stats?.channel_stats || []}
          columns={channelColumns}
          rowKey="channel_id"
          pagination={false}
          size="middle"
        />
      </Card>
    </div>
  );
}
