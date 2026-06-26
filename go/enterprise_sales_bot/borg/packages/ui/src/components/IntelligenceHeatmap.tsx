'use client';

import React, { useRef, useEffect, useState, useCallback, useMemo } from 'react';
import * as THREE from 'three';

// ─── Types ───────────────────────────────────────────────────────────────────

export interface VaultRecord3D {
  id: string;
  type: string;
  content: string;
  importance: number;
  heatScore: number;
  embedding?: number[];
  createdAt: string;
  sessionId?: string;
}

export interface IntelligenceHeatmapProps {
  records: VaultRecord3D[];
  width?: number;
  height?: number;
  onRecordClick?: (record: VaultRecord3D) => void;
  onRecordHover?: (record: VaultRecord3D | null) => void;
  className?: string;
}

// ─── Constants ───────────────────────────────────────────────────────────────

const TYPE_COLORS: Record<string, number> = {
  working: 0x38bdf8,    // sky-400
  long_term: 0x818cf8,  // indigo-400
  archive: 0x64748b,    // slate-500
  heal: 0x34d399,       // emerald-400
  consensus: 0xfbbf24,  // amber-400
  session: 0xf472b6,    // pink-400
  skill: 0xa78bfa,      // violet-400
};

const DEFAULT_COLOR = 0x7dd3fc; // sky-300

// ─── Component ───────────────────────────────────────────────────────────────

/**
 * IntelligenceHeatmap renders a 3D scatter plot of L2 Vault records
 * positioned by their embedding vectors. Point size reflects heat score,
 * color reflects memory type, and opacity reflects importance.
 *
 * Records without embeddings are positioned using a deterministic hash
 * of their ID, clustered by type.
 */
export function IntelligenceHeatmap({
  records,
  width = 800,
  height = 600,
  onRecordClick,
  onRecordHover,
  className,
}: IntelligenceHeatmapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const sceneRef = useRef<{
    scene: THREE.Scene;
    camera: THREE.PerspectiveCamera;
    renderer: THREE.WebGLRenderer;
    points: THREE.Points | null;
    raycaster: THREE.Raycaster;
    mouse: THREE.Vector2;
    animationId: number;
    recordMap: Map<number, VaultRecord3D>;
  } | null>(null);

  const [hoveredRecord, setHoveredRecord] = useState<VaultRecord3D | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<VaultRecord3D | null>(null);

  // ─── Scene Setup ──────────────────────────────────────────────────────────

  useEffect(() => {
    if (!containerRef.current) return;

    const container = containerRef.current;

    // Create canvas if not present
    if (!canvasRef.current) {
      canvasRef.current = document.createElement('canvas');
      container.appendChild(canvasRef.current);
    }
    const canvas = canvasRef.current;

    // Scene
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x0a0f1a);
    scene.fog = new THREE.Fog(0x0a0f1a, 30, 80);

    // Camera
    const camera = new THREE.PerspectiveCamera(60, width / height, 0.1, 200);
    camera.position.set(0, 8, 25);
    camera.lookAt(0, 0, 0);

    // Renderer
    const renderer = new THREE.WebGLRenderer({
      canvas,
      antialias: true,
      alpha: true,
    });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

    // Lighting
    const ambient = new THREE.AmbientLight(0x334155, 0.5);
    scene.add(ambient);

    const point = new THREE.PointLight(0x38bdf8, 1, 50);
    point.position.set(10, 15, 10);
    scene.add(point);

    // Grid helper
    const grid = new THREE.GridHelper(40, 20, 0x1e293b, 0x0f172a);
    grid.position.y = -5;
    scene.add(grid);

    // Raycaster for interaction
    const raycaster = new THREE.Raycaster();
    raycaster.params.Points.threshold = 0.5;
    const mouse = new THREE.Vector2();

    const state = {
      scene,
      camera,
      renderer,
      points: null as THREE.Points | null,
      raycaster,
      mouse,
      animationId: 0,
      recordMap: new Map<number, VaultRecord3D>(),
    };
    sceneRef.current = state;

    // ─── Animation Loop ──────────────────────────────────────────────────

    const animate = () => {
      state.animationId = requestAnimationFrame(animate);

      // Slow auto-rotation
      if (state.points) {
        state.points.rotation.y += 0.001;
      }

      // Hover detection
      raycaster.setFromCamera(mouse, camera);

      renderer.render(scene, camera);
    };
    animate();

    // ─── Mouse Handlers ──────────────────────────────────────────────────

    const onMouseMove = (event: MouseEvent) => {
      const rect = canvas.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      // Raycast for hover
      if (state.points) {
        const intersects = raycaster.intersectObject(state.points);
        if (intersects.length > 0) {
          const idx = intersects[0].index ?? -1;
          const record = state.recordMap.get(idx);
          if (record) {
            setHoveredRecord(record);
            onRecordHover?.(record);
          }
        } else {
          if (hoveredRecord) {
            setHoveredRecord(null);
            onRecordHover?.(null);
          }
        }
      }
    };

    const onClick = () => {
      if (hoveredRecord) {
        setSelectedRecord(hoveredRecord);
        onRecordClick?.(hoveredRecord);
      }
    };

    canvas.addEventListener('mousemove', onMouseMove);
    canvas.addEventListener('click', onClick);

    // ─── Cleanup ─────────────────────────────────────────────────────────

    return () => {
      cancelAnimationFrame(state.animationId);
      canvas.removeEventListener('mousemove', onMouseMove);
      canvas.removeEventListener('click', onClick);
      renderer.dispose();
      if (canvas.parentNode === container) {
        container.removeChild(canvas);
      }
      canvasRef.current = null;
      sceneRef.current = null;
    };
  }, [width, height]); // eslint-disable-line react-hooks/exhaustive-deps

  // ─── Update Points ────────────────────────────────────────────────────────

  useEffect(() => {
    if (!sceneRef.current || records.length === 0) return;

    const { scene, recordMap } = sceneRef.current;

    // Remove existing points
    if (sceneRef.current.points) {
      scene.remove(sceneRef.current.points);
      sceneRef.current.points.geometry.dispose();
      (sceneRef.current.points.material as THREE.Material).dispose();
    }

    recordMap.clear();

    const positions = new Float32Array(records.length * 3);
    const colors = new Float32Array(records.length * 3);
    const sizes = new Float32Array(records.length);

    records.forEach((record, i) => {
      // Position: use embedding if available, otherwise deterministic position
      if (record.embedding && record.embedding.length >= 3) {
        // Normalize embedding to 3D space [-10, 10]
        const dim = Math.min(record.embedding.length, 3);
        for (let d = 0; d < dim; d++) {
          positions[i * 3 + d] = record.embedding[d] * 10;
        }
        // Fill remaining dimensions
        for (let d = dim; d < 3; d++) {
          positions[i * 3 + d] = 0;
        }
      } else {
        // Deterministic position based on record ID hash + type
        const hash = simpleHash(record.id);
        const typeOffset = TYPE_COLORS[record.type] ? Object.keys(TYPE_COLORS).indexOf(record.type) : 0;

        positions[i * 3 + 0] = ((hash % 200) - 100) / 10 + typeOffset * 3;
        positions[i * 3 + 1] = ((Math.floor(hash / 200) % 200) - 100) / 10;
        positions[i * 3 + 2] = ((Math.floor(hash / 40000) % 200) - 100) / 10;
      }

      // Color from type
      const colorHex = TYPE_COLORS[record.type] ?? DEFAULT_COLOR;
      const color = new THREE.Color(colorHex);
      colors[i * 3 + 0] = color.r;
      colors[i * 3 + 1] = color.g;
      colors[i * 3 + 2] = color.b;

      // Size from heat score
      sizes[i] = Math.max(2, Math.min(12, (record.heatScore || 50) / 10));

      // Map index to record
      recordMap.set(i, record);
    });

    // Create geometry
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));
    geometry.setAttribute('size', new THREE.BufferAttribute(sizes, 1));

    // Custom shader material for varying point sizes
    const material = new THREE.ShaderMaterial({
      uniforms: {
        uTime: { value: 0 },
      },
      vertexShader: `
        attribute float size;
        varying vec3 vColor;
        uniform float uTime;
        void main() {
          vColor = color;
          vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
          gl_PointSize = size * (200.0 / -mvPosition.z);
          gl_Position = projectionMatrix * mvPosition;
        }
      `,
      fragmentShader: `
        varying vec3 vColor;
        void main() {
          float dist = length(gl_PointCoord - vec2(0.5));
          if (dist > 0.5) discard;
          float alpha = 1.0 - smoothstep(0.3, 0.5, dist);
          gl_FragColor = vec4(vColor, alpha * 0.85);
        }
      `,
      transparent: true,
      vertexColors: true,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
    });

    const points = new THREE.Points(geometry, material);
    scene.add(points);
    sceneRef.current.points = points;
  }, [records]);

  return (
    <div className={`relative ${className ?? ''}`}>
      <div ref={containerRef} style={{ width, height }} />

      {/* Hover tooltip */}
      {hoveredRecord && (
        <div className="absolute bottom-4 left-4 max-w-xs rounded-lg border border-slate-700 bg-slate-900/95 p-3 text-xs backdrop-blur">
          <div className="flex items-center gap-2 mb-1">
            <span
              className="h-2.5 w-2.5 rounded-full"
              style={{ backgroundColor: typeColorCSS(hoveredRecord.type) }}
            />
            <span className="font-semibold text-slate-200">{hoveredRecord.type}</span>
            <span className="text-slate-500">· Heat {Math.round(hoveredRecord.heatScore)}</span>
          </div>
          <div className="text-slate-300 line-clamp-3">{hoveredRecord.content}</div>
          <div className="mt-1 text-slate-500">
            Importance: {Math.round(hoveredRecord.importance * 100)}%
          </div>
        </div>
      )}

      {/* Selected detail */}
      {selectedRecord && (
        <div className="absolute top-4 right-4 max-w-sm rounded-lg border border-cyan-500/30 bg-slate-950/95 p-4 text-xs backdrop-blur">
          <div className="flex items-center justify-between gap-2 mb-2">
            <div className="flex items-center gap-2">
              <span
                className="h-3 w-3 rounded-full"
                style={{ backgroundColor: typeColorCSS(selectedRecord.type) }}
              />
              <span className="font-bold text-cyan-200">{selectedRecord.type.toUpperCase()}</span>
            </div>
            <button
              onClick={() => setSelectedRecord(null)}
              className="text-slate-500 hover:text-slate-300"
            >
              ✕
            </button>
          </div>
          <div className="text-slate-200 leading-relaxed mb-3">{selectedRecord.content}</div>
          <div className="grid grid-cols-2 gap-2 text-slate-400">
            <div>Heat: <span className="text-amber-300">{Math.round(selectedRecord.heatScore)}</span></div>
            <div>Importance: <span className="text-blue-300">{Math.round(selectedRecord.importance * 100)}%</span></div>
            <div>Created: {new Date(selectedRecord.createdAt).toLocaleDateString()}</div>
            <div>ID: {selectedRecord.id.slice(0, 8)}...</div>
          </div>
        </div>
      )}

      {/* Legend */}
      <div className="absolute bottom-4 right-4 flex flex-wrap gap-2 text-[10px]">
        {Object.entries(TYPE_COLORS).map(([type, _]) => (
          <div key={type} className="flex items-center gap-1">
            <span
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: typeColorCSS(type) }}
            />
            <span className="text-slate-500">{type}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function simpleHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit int
  }
  return Math.abs(hash);
}

function typeColorCSS(type: string): string {
  const hex = TYPE_COLORS[type] ?? DEFAULT_COLOR;
  return '#' + hex.toString(16).padStart(6, '0');
}

export default IntelligenceHeatmap;
