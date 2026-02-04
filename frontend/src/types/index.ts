export interface Channel {
  id: number;
  name: string;
  base_url: string;
  api_key: string;
  provider: string;
  is_active: boolean;
  priority: number;
  max_retries: number;
  timeout: number;
  rate_limit: number;
  created_at: string;
  updated_at: string;
}

export interface ChannelCreate {
  name: string;
  base_url: string;
  api_key: string;
  provider: string;
  priority?: number;
  max_retries?: number;
  timeout?: number;
  rate_limit?: number;
}

export interface ModelMapping {
  id: number;
  channel_id: number;
  channel_name: string;
  upstream_model: string;
  display_model: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface MappingCreate {
  channel_id: number;
  upstream_model: string;
  display_model: string;
}

export interface RequestLog {
  id: number;
  channel_id: number;
  channel_name: string;
  request_id: string;
  model_name: string;
  upstream_model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  request_time: string;
  response_time: string | null;
  latency_ms: number;
  status: string;
  error_code: string;
  error_message: string;
  ip_address: string;
  created_at: string;
}

export interface ChannelStats {
  channel_id: number;
  channel_name: string;
  total_requests: number;
  success_requests: number;
  failed_requests: number;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  avg_latency_ms: number;
}

export interface DailyStats {
  date: string;
  total_requests: number;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

export interface ModelStats {
  model_name: string;
  total_requests: number;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
}

export interface OverallStats {
  total_channels: number;
  active_channels: number;
  total_requests: number;
  total_tokens: number;
  channel_stats: ChannelStats[];
  daily_stats: DailyStats[];
  model_stats: ModelStats[];
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page?: number;
  pageSize?: number;
}

export interface StatsFilter {
  start_date?: string;
  end_date?: string;
  channel_id?: number;
  model_name?: string;
  status?: string;
}
