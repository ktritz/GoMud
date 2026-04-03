import { useState } from 'react';
import { Link } from 'react-router-dom';

interface Column<T> {
  key: string;
  label: string;
  render?: (item: T) => React.ReactNode;
}

interface EntityListProps<T> {
  title: string;
  data: T[] | undefined;
  isLoading: boolean;
  columns: Column<T>[];
  linkPrefix: string;
  idField: string;
  searchPlaceholder?: string;
  onSearch?: (query: string) => void;
}

export function EntityList<T extends Record<string, any>>({
  title,
  data,
  isLoading,
  columns,
  linkPrefix,
  idField,
  searchPlaceholder = 'Search...',
  onSearch,
}: EntityListProps<T>) {
  const [search, setSearch] = useState('');

  const filtered = data?.filter((item) => {
    if (!search) return true;
    const lower = search.toLowerCase();
    return columns.some((col) => {
      const val = item[col.key];
      return val && String(val).toLowerCase().includes(lower);
    });
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">{title}</h1>
        <div className="flex gap-2">
          <input
            type="text"
            placeholder={searchPlaceholder}
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              onSearch?.(e.target.value);
            }}
            className="px-3 py-1.5 bg-gray-800 border border-gray-600 rounded text-sm text-white focus:outline-none focus:border-blue-500 w-64"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="text-gray-400">Loading...</div>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-700">
              {columns.map((col) => (
                <th key={col.key} className="text-left py-2 px-3 text-gray-400 font-medium">
                  {col.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered?.map((item) => (
              <tr
                key={item[idField]}
                className="border-b border-gray-800 hover:bg-gray-800/50 transition-colors"
              >
                {columns.map((col) => (
                  <td key={col.key} className="py-2 px-3">
                    {col.key === columns[0].key ? (
                      <Link
                        to={`${linkPrefix}/${item[idField]}`}
                        className="text-blue-400 hover:text-blue-300"
                      >
                        {col.render ? col.render(item) : item[col.key]}
                      </Link>
                    ) : col.render ? (
                      col.render(item)
                    ) : (
                      item[col.key]
                    )}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <div className="mt-2 text-sm text-gray-500">
        {filtered?.length ?? 0} of {data?.length ?? 0} entries
      </div>
    </div>
  );
}
