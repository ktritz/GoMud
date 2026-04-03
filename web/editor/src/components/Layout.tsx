import { NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../stores/auth';

const navSections = [
  {
    label: 'World',
    links: [
      { to: '/zones', label: 'Zones' },
      { to: '/rooms', label: 'Rooms' },
    ],
  },
  {
    label: 'Entities',
    links: [
      { to: '/mobs', label: 'Mobs' },
      { to: '/items', label: 'Items' },
      { to: '/races', label: 'Races' },
    ],
  },
  {
    label: 'Systems',
    links: [
      { to: '/quests', label: 'Quests' },
      { to: '/buffs', label: 'Buffs' },
      { to: '/spells', label: 'Spells' },
    ],
  },
];

export function Layout() {
  const { user, logout } = useAuth();

  return (
    <div className="flex h-screen bg-gray-900 text-gray-100">
      {/* Sidebar */}
      <aside className="w-56 bg-gray-800 border-r border-gray-700 flex flex-col">
        <div className="p-4 border-b border-gray-700">
          <h1 className="text-lg font-bold text-blue-400">GoMud Editor</h1>
        </div>

        <nav className="flex-1 overflow-y-auto py-2">
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `block px-4 py-2 text-sm ${isActive ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-750'}`
            }
          >
            Dashboard
          </NavLink>

          {navSections.map((section) => (
            <div key={section.label} className="mt-4">
              <div className="px-4 py-1 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                {section.label}
              </div>
              {section.links.map((link) => (
                <NavLink
                  key={link.to}
                  to={link.to}
                  className={({ isActive }) =>
                    `block px-4 py-2 text-sm ${isActive ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white hover:bg-gray-750'}`
                  }
                >
                  {link.label}
                </NavLink>
              ))}
            </div>
          ))}
        </nav>

        <div className="p-4 border-t border-gray-700">
          <div className="text-sm text-gray-400">{user?.username}</div>
          <button
            onClick={logout}
            className="text-sm text-red-400 hover:text-red-300 mt-1"
          >
            Logout
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto p-6">
        <Outlet />
      </main>
    </div>
  );
}
