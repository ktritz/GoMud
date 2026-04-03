import { create } from 'zustand';
import { login as apiLogin, setToken, getToken, api } from '../api/client';

interface User {
  userId: number;
  username: string;
  role: string;
}

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
}

export const useAuth = create<AuthState>((set) => ({
  isAuthenticated: !!getToken(),
  user: null,
  loading: !!getToken(), // loading if we have a token to verify

  login: async (username: string, password: string) => {
    const data = await apiLogin(username, password);
    set({ isAuthenticated: true, user: data.user, loading: false });
  },

  logout: () => {
    setToken(null);
    set({ isAuthenticated: false, user: null, loading: false });
  },

  checkAuth: async () => {
    if (!getToken()) {
      set({ isAuthenticated: false, user: null, loading: false });
      return;
    }
    try {
      const user = await api.get<User>('/auth/me');
      set({ isAuthenticated: true, user, loading: false });
    } catch {
      setToken(null);
      set({ isAuthenticated: false, user: null, loading: false });
    }
  },
}));
