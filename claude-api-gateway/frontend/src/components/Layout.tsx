import { Layout as AntLayout, Menu, Typography, Avatar } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  CloudServerOutlined,
  SwapOutlined,
  FileTextOutlined,
} from '@ant-design/icons';

const { Header, Sider, Content } = AntLayout;
const { Title } = Typography;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '数据概览' },
  { key: '/channels', icon: <CloudServerOutlined />, label: 'API 渠道' },
  { key: '/mappings', icon: <SwapOutlined />, label: '模型映射' },
  { key: '/logs', icon: <FileTextOutlined />, label: '请求日志' },
];

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <AntLayout style={{ minHeight: '100vh', background: '#f5f7fa' }}>
      <Sider
        theme="light"
        width={260}
        style={{
          background: 'linear-gradient(180deg, #667eea 0%, #764ba2 100%)',
          boxShadow: '2px 0 8px rgba(0,0,0,0.1)',
        }}
      >
        <div
          style={{
            padding: '24px 16px',
            textAlign: 'center',
            borderBottom: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '12px',
              marginBottom: '8px',
            }}
          >
            <Avatar
              size={40}
              style={{
                background: 'rgba(255,255,255,0.2)',
                backdropFilter: 'blur(10px)',
              }}
              icon={<CloudServerOutlined />}
            />
          </div>
          <Title
            level={4}
            style={{
              color: '#fff',
              margin: 0,
              fontWeight: 600,
              letterSpacing: '0.5px',
            }}
          >
            AI 创作网关
          </Title>
          <div
            style={{
              color: 'rgba(255,255,255,0.7)',
              fontSize: '12px',
              marginTop: '4px',
            }}
          >
            智能API管理平台
          </div>
        </div>
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{
            background: 'transparent',
            border: 'none',
            marginTop: '16px',
          }}
          getPopupContainer={(node) => node}
          className="custom-menu"
        />
        <style>{`
          .custom-menu .ant-menu-item {
            color: rgba(255,255,255,0.85) !important;
            margin: 4px 12px;
            border-radius: 12px;
            transition: all 0.3s ease;
          }
          .custom-menu .ant-menu-item:hover {
            background: rgba(255,255,255,0.15) !important;
            color: #fff !important;
          }
          .custom-menu .ant-menu-item-selected {
            background: rgba(255,255,255,0.25) !important;
            color: #fff !important;
            font-weight: 500;
          }
          .custom-menu .ant-menu-item .anticon {
            color: inherit;
            font-size: 16px;
          }
        `}</style>
      </Sider>
      <AntLayout>
        <Header
          style={{
            background: '#fff',
            padding: '0 32px',
            display: 'flex',
            alignItems: 'center',
            boxShadow: '0 1px 4px rgba(0,0,0,0.06)',
          }}
        >
          <Title
            level={4}
            style={{
              margin: 0,
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              fontWeight: 600,
            }}
          >
            管理控制台
          </Title>
        </Header>
        <Content
          style={{
            margin: '24px',
            background: '#fff',
            borderRadius: '16px',
            padding: '32px',
            boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
            minHeight: 'calc(100vh - 112px)',
          }}
        >
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  );
}
