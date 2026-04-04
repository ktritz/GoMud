const API_BASE = '/api';
const TOKEN_KEY = 'gomud_editor_token';

let authToken: string | null = localStorage.getItem(TOKEN_KEY);

export function setToken(token: string | null) {
  authToken = token;
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
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
  setToken(data.token);
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
  create: (data: { Zone: string; Title?: string; Description?: string }) =>
    api.post<any>('/admin/rooms', data),
};

export const mobs = {
  list: (params?: { zone?: string; search?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return api.get<{ mobs: MobSummary[]; total: number }>(`/admin/mobs${query ? '?' + query : ''}`);
  },
  get: (id: number) => api.get<any>(`/admin/mobs/${id}`),
  create: (data: { Name: string; Zone: string; Level?: number }) =>
    api.post<any>('/admin/mobs', data),
};

export const items = {
  list: (params?: { type?: string; search?: string }) => {
    const query = new URLSearchParams(params as Record<string, string>).toString();
    return api.get<{ items: ItemSummary[]; total: number }>(`/admin/items${query ? '?' + query : ''}`);
  },
  get: (id: number) => api.get<any>(`/admin/items/${id}`),
  create: (data: { Name: string; Type: string; Subtype?: string }) =>
    api.post<any>('/admin/items', data),
};

export const quests = {
  list: () => api.get<{ quests: any[]; total: number }>('/admin/quests'),
  get: (id: string) => api.get<any>(`/admin/quests/${id}`),
  create: (data: { Name: string; Description?: string }) =>
    api.post<any>('/admin/quests', data),
};

export const buffs = {
  list: () => api.get<{ buffs: any[]; total: number }>('/admin/buffs'),
  get: (id: number) => api.get<any>(`/admin/buffs/${id}`),
};

export const spells = {
  list: () => api.get<{ spells: any[]; total: number }>('/admin/spells'),
  get: (id: string) => api.get<any>(`/admin/spells/${id}`),
};

export const races = {
  list: () => api.get<{ races: any[]; total: number }>('/admin/races'),
  get: (id: number) => api.get<any>(`/admin/races/${id}`),
};

export const zones = {
  list: () => api.get<{ zones: any[]; total: number }>('/admin/zones'),
};

export const server = {
  stats: () => api.get<{ connections: number; disconnections: number; onlineUsers: number }>('/admin/stats'),
};

export interface VCSCommit {
  hash: string;
  shortHash: string;
  author: string;
  date: string;
  message: string;
}

export const vcs = {
  status: () => api.get<{ active: boolean; dirty: boolean; changed: string[] }>('/admin/vcs/status'),
  history: (limit = 50) => api.get<{ commits: VCSCommit[]; total: number }>(`/admin/vcs/history?limit=${limit}`),
  diff: (hash: string) => api.get<{ hash: string; diff: string }>(`/admin/vcs/diff/${hash}`),
  checkpoint: (message: string) => api.post<{ committed: boolean; message: string }>('/admin/vcs/checkpoint', { message }),
  revert: (hash: string) => api.post<{ reverted: boolean; hash: string }>(`/admin/vcs/revert/${hash}`),
};
