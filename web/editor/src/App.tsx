import { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuth } from './stores/auth';
import { Login } from './pages/Login';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { Rooms } from './pages/Rooms';
import { RoomDetail } from './pages/RoomDetail';
import { Mobs } from './pages/Mobs';
import { MobDetail } from './pages/MobDetail';
import { Items } from './pages/Items';
import { ItemDetail } from './pages/ItemDetail';
import { Quests } from './pages/Quests';
import { Zones } from './pages/Zones';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

export default function App() {
  const { isAuthenticated, loading, checkAuth } = useAuth();

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 text-gray-400">
        Loading...
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/zones" element={<Zones />} />
            <Route path="/rooms" element={<Rooms />} />
            <Route path="/rooms/:roomId" element={<RoomDetail />} />
            <Route path="/mobs" element={<Mobs />} />
            <Route path="/mobs/:mobId" element={<MobDetail />} />
            <Route path="/items" element={<Items />} />
            <Route path="/items/:itemId" element={<ItemDetail />} />
            <Route path="/quests" element={<Quests />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
