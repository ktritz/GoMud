import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vcs, type VCSCommit } from '../api/client';

export function History() {
  const queryClient = useQueryClient();
  const [selectedCommit, setSelectedCommit] = useState<string | null>(null);
  const [checkpointMsg, setCheckpointMsg] = useState('');

  const { data: status } = useQuery({
    queryKey: ['vcs-status'],
    queryFn: () => vcs.status(),
    refetchInterval: 30_000,
  });

  const { data: history, isLoading } = useQuery({
    queryKey: ['vcs-history'],
    queryFn: () => vcs.history(100),
  });

  const { data: diffData, isLoading: diffLoading } = useQuery({
    queryKey: ['vcs-diff', selectedCommit],
    queryFn: () => vcs.diff(selectedCommit!),
    enabled: !!selectedCommit,
  });

  const checkpointMut = useMutation({
    mutationFn: (msg: string) => vcs.checkpoint(msg),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vcs-history'] });
      queryClient.invalidateQueries({ queryKey: ['vcs-status'] });
      setCheckpointMsg('');
    },
  });

  const revertMut = useMutation({
    mutationFn: (hash: string) => vcs.revert(hash),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vcs-history'] });
      queryClient.invalidateQueries({ queryKey: ['vcs-status'] });
      setSelectedCommit(null);
    },
  });

  if (status && !status.active) {
    return (
      <div>
        <h1 className="text-2xl font-bold mb-4">Version History</h1>
        <div className="bg-yellow-900/30 border border-yellow-700 text-yellow-300 px-4 py-3 rounded">
          Version control is not active. Initialize a git repository in the data directory.
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Version History</h1>
        {status?.dirty && (
          <span className="px-2 py-0.5 bg-yellow-900/50 text-yellow-300 text-sm rounded">
            {status.changed?.length || 0} unsaved changes
          </span>
        )}
      </div>

      {/* Checkpoint form */}
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-6">
        <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-2">Create Checkpoint</h2>
        <div className="flex gap-2">
          <input
            type="text"
            value={checkpointMsg}
            onChange={(e) => setCheckpointMsg(e.target.value)}
            placeholder="Describe your changes..."
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && checkpointMsg.trim()) {
                checkpointMut.mutate(checkpointMsg.trim());
              }
            }}
          />
          <button
            onClick={() => checkpointMsg.trim() && checkpointMut.mutate(checkpointMsg.trim())}
            disabled={checkpointMut.isPending || !checkpointMsg.trim()}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white text-sm rounded font-medium"
          >
            {checkpointMut.isPending ? 'Saving...' : 'Save Checkpoint'}
          </button>
        </div>
        {checkpointMut.isSuccess && !(checkpointMut.data as any)?.committed && (
          <p className="text-gray-500 text-sm mt-2">No changes to commit.</p>
        )}
        {checkpointMut.isError && (
          <p className="text-red-400 text-sm mt-2">{(checkpointMut.error as Error).message}</p>
        )}

        {status?.dirty && status.changed && status.changed.length > 0 && (
          <details className="mt-3">
            <summary className="text-gray-500 text-sm cursor-pointer hover:text-gray-300">
              {status.changed.length} pending changes
            </summary>
            <ul className="mt-1 text-xs text-gray-400 font-mono max-h-32 overflow-y-auto">
              {status.changed.map((f, i) => (
                <li key={i}>{f}</li>
              ))}
            </ul>
          </details>
        )}
      </div>

      {/* Commit list + diff viewer */}
      <div className="flex gap-4">
        {/* Commit list */}
        <div className="w-1/2">
          {isLoading ? (
            <div className="text-gray-400">Loading history...</div>
          ) : (
            <div className="space-y-1">
              {history?.commits?.map((commit: VCSCommit) => (
                <button
                  key={commit.hash}
                  onClick={() => setSelectedCommit(commit.hash === selectedCommit ? null : commit.hash)}
                  className={`w-full text-left px-3 py-2 rounded text-sm transition-colors ${
                    commit.hash === selectedCommit
                      ? 'bg-blue-900/40 border border-blue-700'
                      : 'bg-gray-800 border border-gray-800 hover:border-gray-700'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-gray-500 text-xs">{commit.shortHash}</span>
                    <span className="text-gray-500 text-xs">{formatDate(commit.date)}</span>
                  </div>
                  <div className="text-gray-300 mt-0.5">{commit.message}</div>
                  <div className="text-gray-500 text-xs mt-0.5">by {commit.author}</div>
                </button>
              ))}
              {history?.commits?.length === 0 && (
                <p className="text-gray-500">No history yet.</p>
              )}
            </div>
          )}
        </div>

        {/* Diff viewer */}
        <div className="w-1/2">
          {selectedCommit && (
            <div className="bg-gray-800 border border-gray-700 rounded-lg overflow-hidden">
              <div className="flex items-center justify-between px-3 py-2 border-b border-gray-700">
                <span className="text-sm font-mono text-gray-400">{selectedCommit.slice(0, 8)}</span>
                <button
                  onClick={() => {
                    if (confirm('Are you sure you want to revert this commit? This creates a new commit undoing these changes.')) {
                      revertMut.mutate(selectedCommit);
                    }
                  }}
                  disabled={revertMut.isPending}
                  className="px-3 py-1 bg-red-700 hover:bg-red-600 disabled:bg-gray-600 text-white text-xs rounded"
                >
                  {revertMut.isPending ? 'Reverting...' : 'Revert'}
                </button>
              </div>
              {revertMut.isError && (
                <div className="px-3 py-1 bg-red-900/30 text-red-300 text-sm">
                  {(revertMut.error as Error).message}
                </div>
              )}
              {diffLoading ? (
                <div className="p-4 text-gray-400 text-sm">Loading diff...</div>
              ) : (
                <pre className="p-3 text-xs font-mono overflow-auto max-h-[600px] whitespace-pre">
                  {diffData?.diff?.split('\n').map((line: string, i: number) => (
                    <div
                      key={i}
                      className={
                        line.startsWith('+') && !line.startsWith('+++')
                          ? 'text-green-400 bg-green-900/20'
                          : line.startsWith('-') && !line.startsWith('---')
                          ? 'text-red-400 bg-red-900/20'
                          : line.startsWith('@@')
                          ? 'text-blue-400'
                          : 'text-gray-400'
                      }
                    >
                      {line}
                    </div>
                  ))}
                </pre>
              )}
            </div>
          )}
          {!selectedCommit && (
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-8 text-center text-gray-500">
              Select a commit to view its diff
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  if (diff < 60_000) return 'just now';
  if (diff < 3600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86400_000) return `${Math.floor(diff / 3600_000)}h ago`;
  return d.toLocaleDateString();
}
