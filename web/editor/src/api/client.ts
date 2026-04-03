const API_BASE = '/api';

let authToken: string | null = null;

export function setToken(token: string | null) {
  authToken = token;
}

export function getToken(): string | null {
  return authToken;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    authToken = null;
    window.location.href = '/';
    throw new Error('Unauthorized');
  }

  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.message || data.error || 'Request failed');
  }

  return data as T;
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T>(path: string) => request<T>('DELETE', path),
};

// Auth
export interface LoginResponse {
  token: string;
  user: { userId: number; username: string; role: string };
  expiresAt: string;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const data = await request<LoginResponse>('POST', '/auth/login', { username, password });
  authToken = data.token;
  return data;
}

// Entity types
export interface RoomSummary {
  roomId: number;
  zone: string;
  title: string;
  biome: string;
  exitCount: number;
}

export interface MobSummary {
  mobId: number;
  zone: string;
  name: string;
  level: number;
  hostile: boolean;
}

export interface ItemSummary {
  itemId: number;
  name: string;
  type: string;
  subtype: string;
  value: number;
}

export interface ListResponse<T> {
  total: number;
  [key: string]: T[] | number;
}

// API functions
export const rooms = {
  list: (params?: { zone?: string; search?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return api.get<{ rooms: RoomSummary[]; total: number }>(`/admin/rooms${query ? '?' + query : ''}`);
  },
  get: (id: number) => api.get<any>(`/admin/rooms/${id}`),
};

export const mobs = {
  list: (params?: { zone?: string; search?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return api.get<{ mobs: MobSummary[]; total: number }>(`/admin/mobs${query ? '?' + query : ''}`);
  },
  get: (id: number) => api.get<any>(`/admin/mobs/${id}`),
};

export const items = {
  list: (params?: { type?: string; search?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return api.get<{ items: ItemSummary[]; total: number }>(`/admin/items${query ? '?' + query : ''}`);
  },
  get: (id: number) => api.get<any>(`/admin/items/${id}`),
};

export const quests = {
  list: () => api.get<{ quests: any[]; total: number }>('/admin/quests'),
  get: (id: string) => api.get<any>(`/admin/quests/${id}`),
};

export const buffs = {
  list: () => api.get<{ buffs: any[]; total: number }>('/admin/buffs'),
};

export const spells = {
  list: () => api.get<{ spells: any[]; total: number }>('/admin/spells'),
};

export const races = {
  list: () => api.get<{ races: any[]; total: number }>('/admin/races'),
};

export const zones = {
  list: () => api.get<{ zones: any[]; total: number }>('/admin/zones'),
};

export const server = {
  stats: () => api.get<{ connections: number; disconnections: number; onlineUsers: number }>('/admin/stats'),
};
