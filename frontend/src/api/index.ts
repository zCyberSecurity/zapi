import axios from 'axios'

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/admin',
  timeout: 10000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Provider {
  id: number
  name: string
  base_url: string
  api_key: string
  api_type: 'openai' | 'anthropic'
  enabled: boolean
  created_at: string
}

export interface ProviderModel {
  id: number
  provider_id: number
  model_id: string
  provider_model_id: string
  display_name: string
  enabled: boolean
}

export interface APIKey {
  id: number
  key: string
  name: string
  allowed_models: string
  enabled: boolean
  created_at: string
}

export interface UsageLog {
  id: number
  api_key_id: number
  api_key_name: string
  model: string
  date: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  request_count: number
}

// ─── Providers ────────────────────────────────────────────────────────────────

export const providerApi = {
  list: () => http.get<Provider[]>('/providers'),
  create: (data: Partial<Provider>) => http.post<Provider>('/providers', data),
  update: (id: number, data: Partial<Provider>) => http.put<Provider>(`/providers/${id}`, data),
  remove: (id: number) => http.delete(`/providers/${id}`),
}

// ─── Models ───────────────────────────────────────────────────────────────────

export const modelApi = {
  list: (providerId: number) => http.get<ProviderModel[]>(`/providers/${providerId}/models`),
  add: (providerId: number, data: Partial<ProviderModel>) =>
    http.post<ProviderModel>(`/providers/${providerId}/models`, data),
  update: (id: number, data: Partial<ProviderModel>) =>
    http.put<ProviderModel>(`/models/${id}`, data),
  remove: (id: number) => http.delete(`/models/${id}`),
}

// ─── API Keys ─────────────────────────────────────────────────────────────────

export const keyApi = {
  list: () => http.get<APIKey[]>('/keys'),
  create: (data: { name: string; allowed_models: string[] }) =>
    http.post<APIKey>('/keys', data),
  update: (id: number, data: { name?: string; allowed_models?: string[]; enabled?: boolean }) =>
    http.put<APIKey>(`/keys/${id}`, data),
  remove: (id: number) => http.delete(`/keys/${id}`),
}

// ─── Usage ────────────────────────────────────────────────────────────────────

export const usageApi = {
  list: (params?: { date?: string; key_id?: number }) =>
    http.get<UsageLog[]>('/usage', { params }),
}
