import { useRef, useEffect, useState, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

interface MapRoom {
  roomId: number;
  x: number;
  y: number;
  z: number;
  symbol: string;
  title: string;
  exits: Record<string, number>;
}

interface MapData {
  zone: string;
  rootRoom: number;
  rooms: MapRoom[];
  total: number;
}

const CELL = 60;
const NODE_R = 20;
const COLORS = {
  bg: '#111827',
  grid: '#1f2937',
  node: '#374151',
  nodeHover: '#4b5563',
  nodeSelected: '#2563eb',
  text: '#d1d5db',
  textDim: '#6b7280',
  exit: '#4b5563',
  exitArrow: '#6b7280',
};

export function ZoneMap() {
  const { zoneName } = useParams();
  const navigate = useNavigate();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [dragging, setDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [hoveredRoom, setHoveredRoom] = useState<number | null>(null);
  const [selectedZ, setSelectedZ] = useState(0);

  const { data: mapData, isLoading } = useQuery({
    queryKey: ['zone-map', zoneName],
    queryFn: () => api.get<MapData>(`/admin/zones/${encodeURIComponent(zoneName!)}/map`),
    enabled: !!zoneName,
  });

  // Compute z-layers
  const zLayers = [...new Set((mapData?.rooms || []).map((r) => r.z))].sort((a, b) => a - b);

  // Filter rooms by current z-layer
  const visibleRooms = (mapData?.rooms || []).filter((r) => r.z === selectedZ);

  // Compute center offset on first load
  useEffect(() => {
    if (!mapData?.rooms?.length || !canvasRef.current) return;
    const canvas = canvasRef.current;
    const filtered = mapData.rooms.filter((r) => r.z === selectedZ);
    if (filtered.length === 0) return;

    const minX = Math.min(...filtered.map((r) => r.x));
    const maxX = Math.max(...filtered.map((r) => r.x));
    const minY = Math.min(...filtered.map((r) => r.y));
    const maxY = Math.max(...filtered.map((r) => r.y));

    const centerX = ((minX + maxX) / 2) * CELL;
    const centerY = ((minY + maxY) / 2) * CELL;

    setPan({
      x: canvas.width / 2 - centerX,
      y: canvas.height / 2 - centerY,
    });
  }, [mapData, selectedZ]);

  const roomAt = useCallback(
    (screenX: number, screenY: number): MapRoom | null => {
      for (const room of visibleRooms) {
        const rx = room.x * CELL + pan.x;
        const ry = room.y * CELL + pan.y;
        const dx = screenX - rx;
        const dy = screenY - ry;
        if (dx * dx + dy * dy <= NODE_R * NODE_R) return room;
      }
      return null;
    },
    [visibleRooms, pan]
  );

  // Draw
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Size canvas to container
    const rect = canvas.parentElement!.getBoundingClientRect();
    canvas.width = rect.width;
    canvas.height = rect.height;

    ctx.fillStyle = COLORS.bg;
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Draw exits as lines
    ctx.strokeStyle = COLORS.exit;
    ctx.lineWidth = 2;
    for (const room of visibleRooms) {
      const rx = room.x * CELL + pan.x;
      const ry = room.y * CELL + pan.y;
      for (const [, targetId] of Object.entries(room.exits)) {
        const target = visibleRooms.find((r) => r.roomId === targetId);
        if (!target) continue;
        const tx = target.x * CELL + pan.x;
        const ty = target.y * CELL + pan.y;
        ctx.beginPath();
        ctx.moveTo(rx, ry);
        ctx.lineTo(tx, ty);
        ctx.stroke();
      }
    }

    // Draw nodes
    for (const room of visibleRooms) {
      const rx = room.x * CELL + pan.x;
      const ry = room.y * CELL + pan.y;
      const isHovered = room.roomId === hoveredRoom;

      ctx.beginPath();
      ctx.arc(rx, ry, NODE_R, 0, Math.PI * 2);
      ctx.fillStyle = isHovered ? COLORS.nodeHover : COLORS.node;
      ctx.fill();
      ctx.strokeStyle = isHovered ? '#60a5fa' : '#4b5563';
      ctx.lineWidth = isHovered ? 2 : 1;
      ctx.stroke();

      // Room ID
      ctx.fillStyle = COLORS.text;
      ctx.font = '11px monospace';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(`${room.roomId}`, rx, ry);

      // Title below
      ctx.fillStyle = COLORS.textDim;
      ctx.font = '9px sans-serif';
      ctx.fillText(
        room.title.length > 14 ? room.title.slice(0, 12) + '..' : room.title,
        rx,
        ry + NODE_R + 10
      );
    }
  }, [visibleRooms, pan, hoveredRoom]);

  const handleMouseDown = (e: React.MouseEvent) => {
    const room = roomAt(e.nativeEvent.offsetX, e.nativeEvent.offsetY);
    if (room) return; // Don't start drag on a room
    setDragging(true);
    setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (dragging) {
      setPan({ x: e.clientX - dragStart.x, y: e.clientY - dragStart.y });
    } else {
      const room = roomAt(e.nativeEvent.offsetX, e.nativeEvent.offsetY);
      setHoveredRoom(room?.roomId || null);
    }
  };

  const handleMouseUp = () => {
    setDragging(false);
  };

  const handleClick = (e: React.MouseEvent) => {
    const room = roomAt(e.nativeEvent.offsetX, e.nativeEvent.offsetY);
    if (room) {
      navigate(`/rooms/${room.roomId}`);
    }
  };

  if (isLoading) return <div className="text-gray-400">Loading map...</div>;
  if (!mapData) return <div className="text-gray-400">No map data</div>;

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Link to="/zones" className="text-gray-400 hover:text-white">&larr; Zones</Link>
          <span className="text-gray-600">/</span>
          <span className="text-gray-300">{mapData.zone}</span>
          <span className="text-gray-500 text-sm">({mapData.total} rooms)</span>
        </div>
        <div className="flex items-center gap-2">
          {zLayers.length > 1 && (
            <>
              <span className="text-gray-500 text-sm">Layer:</span>
              {zLayers.map((z) => (
                <button
                  key={z}
                  onClick={() => setSelectedZ(z)}
                  className={`px-2 py-0.5 rounded text-sm ${
                    z === selectedZ
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
                  }`}
                >
                  Z={z}
                </button>
              ))}
            </>
          )}
        </div>
      </div>

      <div className="flex-1 relative border border-gray-700 rounded overflow-hidden" style={{ cursor: dragging ? 'grabbing' : 'grab' }}>
        <canvas
          ref={canvasRef}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onClick={handleClick}
          className="w-full h-full"
        />
        {hoveredRoom && (
          <div className="absolute bottom-3 left-3 bg-gray-800/90 border border-gray-700 rounded px-3 py-1.5 text-sm">
            <span className="text-gray-400">Room </span>
            <span className="text-blue-400 font-mono">#{hoveredRoom}</span>
            <span className="text-gray-400"> — </span>
            <span className="text-gray-300">
              {visibleRooms.find((r) => r.roomId === hoveredRoom)?.title}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
