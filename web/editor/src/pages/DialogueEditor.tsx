import { useState, useRef, useCallback, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';

// --- Types ---

type NodeType = 'entry' | 'condition' | 'response' | 'action';

interface DialogueNode {
  id: string;
  type: NodeType;
  x: number;
  y: number;
  data: Record<string, any>;
}

interface DialogueEdge {
  id: string;
  from: string;
  to: string;
  label?: string;
}

interface DialogueTree {
  nodes: DialogueNode[];
  edges: DialogueEdge[];
}

// --- Constants ---

const NODE_W = 220;
const NODE_H_BASE = 60;
const COLORS: Record<NodeType, { bg: string; border: string; label: string }> = {
  entry:     { bg: 'bg-blue-900/60',   border: 'border-blue-600', label: 'Entry' },
  condition: { bg: 'bg-yellow-900/60', border: 'border-yellow-600', label: 'Condition' },
  response:  { bg: 'bg-green-900/60',  border: 'border-green-600', label: 'Response' },
  action:    { bg: 'bg-purple-900/60', border: 'border-purple-600', label: 'Action' },
};

const ENTRY_TYPES = ['onAsk', 'onSay', 'onGive', 'onShow'];
const CONDITION_TYPES = ['hasQuest', 'missingQuest', 'keywordMatch', 'hasItem', 'hasGold'];
const ACTION_TYPES = ['say', 'emote', 'giveQuest', 'giveItem', 'giveGold', 'command'];

let nextId = 1;
function genId() { return `n${nextId++}`; }
function genEdgeId() { return `e${nextId++}`; }

function emptyTree(): DialogueTree {
  const entryId = genId();
  return {
    nodes: [{
      id: entryId, type: 'entry', x: 50, y: 50,
      data: { event: 'onAsk' },
    }],
    edges: [],
  };
}

// --- Component ---

export function DialogueEditor() {
  const { entityType, entityId } = useParams();
  const id = Number(entityId);
  const queryClient = useQueryClient();

  const [tree, setTree] = useState<DialogueTree>(emptyTree());
  const [selected, setSelected] = useState<string | null>(null);
  const [connecting, setConnecting] = useState<string | null>(null);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [dragging, setDragging] = useState<{ nodeId: string; offsetX: number; offsetY: number } | null>(null);
  const [panning, setPanning] = useState(false);
  const [panStart, setPanStart] = useState({ x: 0, y: 0 });
  const containerRef = useRef<HTMLDivElement>(null);

  // Load existing dialogue tree
  const { data: dialogueData } = useQuery({
    queryKey: ['dialogue', entityType, entityId],
    queryFn: () => api.get<{ exists: boolean; tree: DialogueTree | null }>(`/admin/dialogue/${entityType}/${entityId}`),
    enabled: !!entityType && id > 0,
  });

  useEffect(() => {
    if (dialogueData?.exists && dialogueData.tree) {
      setTree(dialogueData.tree);
      // Update nextId to avoid collisions
      const maxId = Math.max(...dialogueData.tree.nodes.map(n => parseInt(n.id.replace('n', '')) || 0), ...dialogueData.tree.edges.map(e => parseInt(e.id.replace('e', '')) || 0));
      nextId = maxId + 1;
    }
  }, [dialogueData]);

  const saveMut = useMutation({
    mutationFn: () => api.put(`/admin/dialogue/${entityType}/${entityId}`, tree),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['dialogue', entityType, entityId] }),
  });

  const generateMut = useMutation({
    mutationFn: () => {
      const code = generateScript(tree);
      return api.put(`/admin/scripts/${entityType}/${entityId}`, { content: code });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['script', entityType, entityId] }),
  });

  // Node operations
  const addNode = (type: NodeType) => {
    const node: DialogueNode = {
      id: genId(), type, x: 300 - pan.x, y: 200 - pan.y,
      data: type === 'entry' ? { event: 'onAsk' }
          : type === 'condition' ? { condType: 'keywordMatch', value: '' }
          : type === 'response' ? { text: '' }
          : { actionType: 'say', value: '' },
    };
    setTree(prev => ({ ...prev, nodes: [...prev.nodes, node] }));
    setSelected(node.id);
  };

  const deleteNode = (nodeId: string) => {
    setTree(prev => ({
      nodes: prev.nodes.filter(n => n.id !== nodeId),
      edges: prev.edges.filter(e => e.from !== nodeId && e.to !== nodeId),
    }));
    if (selected === nodeId) setSelected(null);
  };

  const updateNode = (nodeId: string, data: Record<string, any>) => {
    setTree(prev => ({
      ...prev,
      nodes: prev.nodes.map(n => n.id === nodeId ? { ...n, data: { ...n.data, ...data } } : n),
    }));
  };

  const addEdge = (from: string, to: string) => {
    if (from === to) return;
    if (tree.edges.some(e => e.from === from && e.to === to)) return;
    setTree(prev => ({
      ...prev,
      edges: [...prev.edges, { id: genEdgeId(), from, to }],
    }));
  };

  const deleteEdge = (edgeId: string) => {
    setTree(prev => ({ ...prev, edges: prev.edges.filter(e => e.id !== edgeId) }));
  };

  // Mouse handlers
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.target === containerRef.current || (e.target as HTMLElement).tagName === 'svg') {
      setPanning(true);
      setPanStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
      setSelected(null);
    }
  }, [pan]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (panning) {
      setPan({ x: e.clientX - panStart.x, y: e.clientY - panStart.y });
    }
    if (dragging) {
      const rect = containerRef.current!.getBoundingClientRect();
      const x = e.clientX - rect.left - pan.x - dragging.offsetX;
      const y = e.clientY - rect.top - pan.y - dragging.offsetY;
      setTree(prev => ({
        ...prev,
        nodes: prev.nodes.map(n => n.id === dragging.nodeId ? { ...n, x, y } : n),
      }));
    }
  }, [panning, panStart, dragging, pan]);

  const handleMouseUp = useCallback(() => {
    setPanning(false);
    setDragging(null);
  }, []);

  const selectedNode = tree.nodes.find(n => n.id === selected);

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center justify-between mb-2 shrink-0">
        <div className="flex items-center gap-2">
          <Link to={`/${entityType}s/${entityId}`} className="text-gray-400 hover:text-white">&larr; Back</Link>
          <span className="text-gray-600">|</span>
          <span className="text-gray-300 text-sm">Dialogue: {entityType} #{entityId}</span>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => addNode('entry')} className="px-2 py-1 bg-blue-900/60 border border-blue-600 text-blue-300 text-xs rounded">+ Entry</button>
          <button onClick={() => addNode('condition')} className="px-2 py-1 bg-yellow-900/60 border border-yellow-600 text-yellow-300 text-xs rounded">+ Condition</button>
          <button onClick={() => addNode('response')} className="px-2 py-1 bg-green-900/60 border border-green-600 text-green-300 text-xs rounded">+ Response</button>
          <button onClick={() => addNode('action')} className="px-2 py-1 bg-purple-900/60 border border-purple-600 text-purple-300 text-xs rounded">+ Action</button>
          <span className="text-gray-700">|</span>
          <button
            onClick={() => saveMut.mutate()}
            disabled={saveMut.isPending}
            className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs rounded"
          >
            {saveMut.isPending ? 'Saving...' : 'Save Tree'}
          </button>
          <button
            onClick={() => generateMut.mutate()}
            disabled={generateMut.isPending}
            className="px-3 py-1 bg-green-700 hover:bg-green-600 text-white text-xs rounded"
          >
            {generateMut.isPending ? 'Generating...' : 'Generate Script'}
          </button>
        </div>
      </div>

      {(saveMut.isSuccess || generateMut.isSuccess) && (
        <div className="bg-green-900/30 border border-green-700 text-green-300 px-3 py-1 rounded mb-2 text-sm shrink-0">
          {generateMut.isSuccess ? 'Script generated!' : 'Tree saved!'}
        </div>
      )}

      <div className="flex flex-1 gap-2 min-h-0">
        {/* Canvas */}
        <div
          ref={containerRef}
          className="flex-1 relative border border-gray-700 rounded overflow-hidden bg-gray-950"
          style={{ cursor: connecting ? 'crosshair' : panning ? 'grabbing' : 'grab' }}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        >
          {/* SVG layer for edges */}
          <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ zIndex: 0 }}>
            {tree.edges.map(edge => {
              const fromNode = tree.nodes.find(n => n.id === edge.from);
              const toNode = tree.nodes.find(n => n.id === edge.to);
              if (!fromNode || !toNode) return null;
              const x1 = fromNode.x + pan.x + NODE_W / 2;
              const y1 = fromNode.y + pan.y + NODE_H_BASE;
              const x2 = toNode.x + pan.x + NODE_W / 2;
              const y2 = toNode.y + pan.y;
              const midY = (y1 + y2) / 2;
              return (
                <g key={edge.id} style={{ pointerEvents: 'auto', cursor: 'pointer' }} onClick={() => deleteEdge(edge.id)}>
                  <path
                    d={`M${x1},${y1} C${x1},${midY} ${x2},${midY} ${x2},${y2}`}
                    stroke="#4b5563" strokeWidth={2} fill="none"
                  />
                  <circle cx={x2} cy={y2} r={4} fill="#4b5563" />
                </g>
              );
            })}
          </svg>

          {/* Nodes */}
          {tree.nodes.map(node => (
            <div
              key={node.id}
              className={`absolute rounded border ${COLORS[node.type].bg} ${COLORS[node.type].border} ${
                selected === node.id ? 'ring-2 ring-white/50' : ''
              }`}
              style={{
                left: node.x + pan.x,
                top: node.y + pan.y,
                width: NODE_W,
                zIndex: selected === node.id ? 10 : 1,
              }}
              onMouseDown={(e) => {
                e.stopPropagation();
                if (connecting) {
                  addEdge(connecting, node.id);
                  setConnecting(null);
                  return;
                }
                setSelected(node.id);
                const rect = (e.target as HTMLElement).closest('[class*="absolute rounded"]')!.getBoundingClientRect();
                setDragging({
                  nodeId: node.id,
                  offsetX: e.clientX - rect.left,
                  offsetY: e.clientY - rect.top,
                });
              }}
            >
              <div className="px-2 py-1 flex items-center justify-between border-b border-gray-700/50">
                <span className="text-xs font-bold text-gray-400">{COLORS[node.type].label}</span>
                <div className="flex gap-1">
                  <button
                    className="text-gray-500 hover:text-blue-400 text-xs"
                    title="Connect to another node"
                    onClick={(e) => { e.stopPropagation(); setConnecting(connecting === node.id ? null : node.id); }}
                  >
                    {connecting === node.id ? '...' : '\u2192'}
                  </button>
                  <button
                    className="text-gray-500 hover:text-red-400 text-xs"
                    onClick={(e) => { e.stopPropagation(); deleteNode(node.id); }}
                  >x</button>
                </div>
              </div>
              <div className="px-2 py-1.5 text-xs text-gray-300">
                <NodeSummary node={node} />
              </div>
            </div>
          ))}

          {connecting && (
            <div className="absolute top-2 left-2 bg-blue-900/80 border border-blue-600 text-blue-300 px-2 py-1 rounded text-xs z-20">
              Click a target node to connect, or click canvas to cancel
            </div>
          )}
        </div>

        {/* Properties panel */}
        <div className="w-72 shrink-0 bg-gray-800 border border-gray-700 rounded p-3 overflow-y-auto">
          {selectedNode ? (
            <NodeProperties
              node={selectedNode}
              onChange={(data) => updateNode(selectedNode.id, data)}
              onDelete={() => deleteNode(selectedNode.id)}
            />
          ) : (
            <div className="text-gray-500 text-sm">
              <p>Select a node to edit its properties.</p>
              <p className="mt-4 text-xs text-gray-600">
                Use the toolbar to add nodes. Click the arrow on a node to draw a connection, then click the target.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// --- Node Summary (rendered inside node boxes) ---

function NodeSummary({ node }: { node: DialogueNode }) {
  switch (node.type) {
    case 'entry':
      return <span className="text-blue-400">{node.data.event || 'onAsk'}</span>;
    case 'condition':
      return (
        <span>
          <span className="text-yellow-400">{node.data.condType}</span>
          {node.data.value && <span className="text-gray-400 ml-1">"{node.data.value}"</span>}
        </span>
      );
    case 'response':
      return <span className="text-green-300 italic">"{(node.data.text || '').slice(0, 30)}{(node.data.text?.length || 0) > 30 ? '...' : ''}"</span>;
    case 'action':
      return (
        <span>
          <span className="text-purple-400">{node.data.actionType}</span>
          {node.data.value && <span className="text-gray-400 ml-1">"{(node.data.value || '').slice(0, 20)}"</span>}
        </span>
      );
    default:
      return null;
  }
}

// --- Properties Panel ---

function NodeProperties({ node, onChange, onDelete }: { node: DialogueNode; onChange: (data: Record<string, any>) => void; onDelete: () => void }) {
  const inputCls = "w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:border-blue-500";
  const labelCls = "text-gray-400 text-xs block mb-1";
  const selectCls = "w-full px-2 py-1 bg-gray-700 border border-gray-600 rounded text-white text-sm";

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-semibold text-gray-300">{COLORS[node.type].label} Node</span>
        <button onClick={onDelete} className="text-red-400 hover:text-red-300 text-xs">Delete</button>
      </div>

      {node.type === 'entry' && (
        <div>
          <label className={labelCls}>Event Trigger</label>
          <select value={node.data.event} onChange={(e) => onChange({ event: e.target.value })} className={selectCls}>
            {ENTRY_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
        </div>
      )}

      {node.type === 'condition' && (
        <>
          <div>
            <label className={labelCls}>Condition Type</label>
            <select value={node.data.condType} onChange={(e) => onChange({ condType: e.target.value })} className={selectCls}>
              {CONDITION_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
            </select>
          </div>
          <div>
            <label className={labelCls}>
              {node.data.condType === 'keywordMatch' ? 'Keywords (comma-separated)' :
               node.data.condType === 'hasQuest' || node.data.condType === 'missingQuest' ? 'Quest Token' :
               node.data.condType === 'hasItem' ? 'Item ID' :
               'Value'}
            </label>
            <input
              type="text" value={node.data.value || ''} className={inputCls}
              onChange={(e) => onChange({ value: e.target.value })}
            />
          </div>
          <div>
            <label className={labelCls}>Else label (for false branch)</label>
            <input
              type="text" value={node.data.elseLabel || ''} className={inputCls}
              placeholder="else"
              onChange={(e) => onChange({ elseLabel: e.target.value })}
            />
          </div>
        </>
      )}

      {node.type === 'response' && (
        <div>
          <label className={labelCls}>NPC Says</label>
          <textarea
            value={node.data.text || ''} className={inputCls + ' min-h-20'} rows={3}
            onChange={(e) => onChange({ text: e.target.value })}
          />
        </div>
      )}

      {node.type === 'action' && (
        <>
          <div>
            <label className={labelCls}>Action Type</label>
            <select value={node.data.actionType} onChange={(e) => onChange({ actionType: e.target.value })} className={selectCls}>
              {ACTION_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
            </select>
          </div>
          <div>
            <label className={labelCls}>
              {node.data.actionType === 'say' ? 'Message' :
               node.data.actionType === 'emote' ? 'Emote text' :
               node.data.actionType === 'giveQuest' ? 'Quest token (e.g. 5-start)' :
               node.data.actionType === 'giveItem' ? 'Item ID' :
               node.data.actionType === 'giveGold' ? 'Gold amount' :
               'Command string'}
            </label>
            <input
              type="text" value={node.data.value || ''} className={inputCls}
              onChange={(e) => onChange({ value: e.target.value })}
            />
          </div>
        </>
      )}

      <div className="text-xs text-gray-600 mt-2">ID: {node.id}</div>
    </div>
  );
}

// --- Code Generation ---

function generateScript(tree: DialogueTree): string {
  const lines: string[] = [];
  lines.push('// Generated by GoMud Dialogue Editor');
  lines.push('// Do not edit directly — changes will be overwritten');
  lines.push('');

  // Group entry nodes by event type
  const entries = tree.nodes.filter(n => n.type === 'entry');
  const eventGroups = new Map<string, DialogueNode[]>();
  for (const entry of entries) {
    const event = entry.data.event || 'onAsk';
    if (!eventGroups.has(event)) eventGroups.set(event, []);
    eventGroups.get(event)!.push(entry);
  }

  // Collect keywords from all keyword conditions
  const allKeywords = new Set<string>();
  tree.nodes.filter(n => n.type === 'condition' && n.data.condType === 'keywordMatch').forEach(n => {
    (n.data.value || '').split(',').map((k: string) => k.trim()).filter(Boolean).forEach((k: string) => allKeywords.add(k));
  });

  if (allKeywords.size > 0) {
    lines.push(`const keywords = [${[...allKeywords].map(k => `"${k}"`).join(', ')}];`);
    lines.push('');
  }

  // Generate each event handler
  for (const [event, entryNodes] of eventGroups) {
    const params = event === 'onIdle' ? 'mob, room' : 'mob, room, eventDetails';
    lines.push(`function ${event}(${params}) {`);
    lines.push('  var user = GetUser(eventDetails.sourceId);');
    lines.push('  if (!user) return false;');
    lines.push('');

    for (const entry of entryNodes) {
      generateNodeCode(tree, entry.id, lines, '  ', new Set());
    }

    lines.push('  return false;');
    lines.push('}');
    lines.push('');
  }

  return lines.join('\n');
}

function generateNodeCode(tree: DialogueTree, nodeId: string, lines: string[], indent: string, visited: Set<string>) {
  if (visited.has(nodeId)) return;
  visited.add(nodeId);

  const node = tree.nodes.find(n => n.id === nodeId);
  if (!node) return;

  const children = tree.edges.filter(e => e.from === nodeId).map(e => tree.nodes.find(n => n.id === e.to)).filter(Boolean) as DialogueNode[];

  switch (node.type) {
    case 'entry':
      // Just follow children
      for (const child of children) {
        generateNodeCode(tree, child.id, lines, indent, visited);
      }
      break;

    case 'condition': {
      const condCode = generateCondition(node);
      if (condCode) {
        lines.push(`${indent}if (${condCode}) {`);
        // True branch: non-condition children, or first connection
        for (const child of children) {
          generateNodeCode(tree, child.id, lines, indent + '  ', new Set(visited));
        }
        lines.push(`${indent}  return true;`);
        lines.push(`${indent}}`);
      }
      break;
    }

    case 'response':
      if (node.data.text) {
        for (const line of node.data.text.split('\n')) {
          if (line.trim()) {
            lines.push(`${indent}mob.Command("say ${escapeSay(line.trim())}");`);
          }
        }
      }
      for (const child of children) {
        generateNodeCode(tree, child.id, lines, indent, visited);
      }
      break;

    case 'action':
      lines.push(`${indent}${generateAction(node)}`);
      for (const child of children) {
        generateNodeCode(tree, child.id, lines, indent, visited);
      }
      break;
  }
}

function generateCondition(node: DialogueNode): string {
  switch (node.data.condType) {
    case 'hasQuest':
      return `user.HasQuest("${node.data.value}")`;
    case 'missingQuest':
      return `!user.HasQuest("${node.data.value}")`;
    case 'keywordMatch': {
      const kws = (node.data.value || '').split(',').map((k: string) => `"${k.trim()}"`).join(', ');
      return `UtilFindMatchIn(eventDetails.askText || eventDetails.msg || "", [${kws}]).found`;
    }
    case 'hasItem':
      return `user.HasItemId(${node.data.value})`;
    case 'hasGold':
      return `eventDetails.gold >= ${node.data.value || 0}`;
    default:
      return 'true';
  }
}

function generateAction(node: DialogueNode): string {
  switch (node.data.actionType) {
    case 'say':
      return `mob.Command("say ${escapeSay(node.data.value)}");`;
    case 'emote':
      return `mob.Command("emote ${escapeSay(node.data.value)}");`;
    case 'giveQuest':
      return `user.GiveQuest("${node.data.value}");`;
    case 'giveItem':
      return `mob.Command("give !${node.data.value} @" + String(eventDetails.sourceId));`;
    case 'giveGold':
      return `mob.Command("give ${node.data.value} gold @" + String(eventDetails.sourceId));`;
    case 'command':
      return `mob.Command("${escapeSay(node.data.value)}");`;
    default:
      return `// unknown action: ${node.data.actionType}`;
  }
}

function escapeSay(text: string): string {
  return (text || '').replace(/\\/g, '\\\\').replace(/"/g, '\\"');
}
