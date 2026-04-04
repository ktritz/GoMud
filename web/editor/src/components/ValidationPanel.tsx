import type { ValidationWarning } from '../hooks/useValidation';

interface ValidationPanelProps {
  warnings: ValidationWarning[];
}

export function ValidationPanel({ warnings }: ValidationPanelProps) {
  if (warnings.length === 0) return null;

  const errors = warnings.filter((w) => w.severity === 'error');
  const warns = warnings.filter((w) => w.severity === 'warning');

  return (
    <div className="mb-4 space-y-1">
      {errors.map((w, i) => (
        <div key={`e${i}`} className="bg-red-900/30 border border-red-700 text-red-300 px-3 py-1.5 rounded text-sm flex items-start gap-2">
          <span className="font-bold shrink-0">!!</span>
          <span><span className="font-mono text-red-400">{w.field}</span>: {w.message}</span>
        </div>
      ))}
      {warns.map((w, i) => (
        <div key={`w${i}`} className="bg-yellow-900/30 border border-yellow-700 text-yellow-300 px-3 py-1.5 rounded text-sm flex items-start gap-2">
          <span className="font-bold shrink-0">!</span>
          <span><span className="font-mono text-yellow-400">{w.field}</span>: {w.message}</span>
        </div>
      ))}
    </div>
  );
}
