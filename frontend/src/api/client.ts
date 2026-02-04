import axios from 'axios';
import type {
  Channel,
  ChannelCreate,
  ModelMapping,
  MappingCreate,
  RequestLog,
  OverallStats,
  PaginatedResponse,
  StatsFilter,
  ChannelStats,
  DailyStats,
  ModelStats,
} from '@/types';

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Response interceptor - handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Only redirect to login if:
    // 1. It's a 401 error
    // 2. Not already on login page
    // 3. Not an auth endpoint request (to avoid infinite loops)
    if (
      error.response?.status === 401 &&
      !window.location.pathname.startsWith('/login') &&
      !error.config?.url?.startsWith('/auth/')
    ) {
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Channels API
export const channelsApi = {
  list: () => api.get<{ data: Channel[]; total: number }>('/channels'),
  get: (id: number) => api.get<Channel>(`/channels/${id}`),
  create: (data: ChannelCreate) => api.post<Channel>('/channels', data),
  update: (id: number, data: Partial<Channel>) => api.put<Channel>(`/channels/${id}`, data),
  delete: (id: number) => api.delete(`/channels/${id}`),
  activate: (id: number) => api.put(`/channels/${id}/activate`),
  deactivate: (id: number) => api.put(`/channels/${id}/deactivate`),
  test: (baseUrl: string, apiKey: string) => api.post('/channels/test', { base_url: baseUrl, api_key: apiKey }),
  getMappings: (id: number) => api.get<{ data: ModelMapping[]; total: number }>(`/channels/${id}/mappings`),
};

// Mappings API
export const mappingsApi = {
  list: () => api.get<{ data: ModelMapping[]; total: number }>('/mappings'),
  get: (id: number) => api.get<ModelMapping>(`/mappings/${id}`),
  create: (data: MappingCreate) => api.post<ModelMapping>('/mappings', data),
  update: (id: number, data: Partial<ModelMapping>) => api.put<ModelMapping>(`/mappings/${id}`, data),
  delete: (id: number) => api.delete(`/mappings/${id}`),
};

// Stats API
export const statsApi = {
  getOverall: (filter?: StatsFilter) => api.get<OverallStats>('/stats', { params: filter }),
  getChannelStats: (filter?: StatsFilter) => api.get<ChannelStats[]>('/stats/channels', { params: filter }),
  getDailyStats: (filter?: StatsFilter) => api.get<DailyStats[]>('/stats/daily', { params: filter }),
  getModelStats: (filter?: StatsFilter) => api.get<ModelStats[]>('/stats/models', { params: filter }),
  getLogs: (filter?: StatsFilter, page = 1, pageSize = 20) =>
    api.get<PaginatedResponse<RequestLog>>('/stats/logs', { params: { ...filter, page, page_size: pageSize } }),
  export: (filter?: StatsFilter) => api.get('/stats/export', { params: filter, responseType: 'blob' }),
};

// Auth API
export const authApi = {
  login: (apiKey: string) => api.post<{ success: boolean; message: string; token: string }>('/auth/login', { api_key: apiKey }),
  logout: () => api.post('/auth/logout'),
  verify: () => api.get<{ authenticated: boolean }>('/auth/verify'),
};

export default api;
