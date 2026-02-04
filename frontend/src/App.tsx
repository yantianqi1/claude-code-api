import { ConfigProvider } from 'antd';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import zhCN from 'antd/locale/zh_CN';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import Channels from '@/pages/Channels';
import Mappings from '@/pages/Mappings';
import Logs from '@/pages/Logs';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#667eea',
            colorSuccess: '#52c41a',
            colorWarning: '#faad14',
            colorError: '#ff4d4f',
            colorInfo: '#667eea',
            borderRadius: 8,
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
          },
          components: {
            Button: {
              borderRadius: 8,
              controlHeight: 40,
            },
            Input: {
              borderRadius: 8,
              controlHeight: 40,
            },
            Select: {
              borderRadius: 8,
              controlHeight: 40,
            },
            Table: {
              borderRadiusLG: 12,
              headerBg: '#fafafa',
            },
            Card: {
              borderRadiusLG: 12,
            },
            Modal: {
              borderRadiusLG: 12,
            },
          },
        }}
      >
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<Dashboard />} />
              <Route path="channels" element={<Channels />} />
              <Route path="mappings" element={<Mappings />} />
              <Route path="logs" element={<Logs />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}

export default App;
