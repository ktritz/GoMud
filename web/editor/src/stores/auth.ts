import { create } from 'zustand';
import { login as apiLogin, setToken } from '../api/client';

interface AuthState {
  isAuthenticated: boolean;
  user: { userId: number; username: string; role: string } | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuth = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,

  login: async (username: string, password: string) => {
    const data = await apiLogin(username, password);
    set({ isAuthenticated: true, user: data.user });
  },

  logout: () => {
    setToken(null);
    set({ isAuthenticated: false, user: null });
  },
}));
