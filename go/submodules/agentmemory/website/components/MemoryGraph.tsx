"use client";

import { useEffect, useRef, useState } from "react";
import styles from "./MemoryGraph.module.css";

interface Node {
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
  hot: boolean;
}

export function MemoryGraph() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [running, setRunning] = useState(true);
  const railRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const reduceMotion = matchMedia("(prefers-reduced-motion: reduce)").matches;
    let localRunning = running && !reduceMotion;
    let nodes: Node[] = [];
    let rafId = 0;
    let pulse = 0;

    const size = () => {
      const dpr = Math.max(1, window.devicePixelRatio || 1);
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    const seed = () => {
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      const count = Math.min(52, Math.floor((w * h) / 22000));
      nodes = new Array(count).fill(0).map(() => ({
        x: Math.random() * w,
        y: Math.random() * h,
        vx: (Math.random() - 0.5) * 0.18,
        vy: (Math.random() - 0.5) * 0.18,
        r: 1.2 + Math.random() * 2.2,
        hot: Math.random() < 0.25,
      }));
    };

    const draw = () => {
      const w = canvas.clientWidth;
      const h = canvas.clientHeight;
      ctx.clearRect(0, 0, w, h);

      for (const n of nodes) {
        n.x += n.vx;
        n.y += n.vy;
        if (n.x < 0 || n.x > w) n.vx *= -1;
        if (n.y < 0 || n.y > h) n.vy *= -1;
      }

      const maxDist = 160;
      ctx.lineWidth = 1;
      for (let i = 0; i < nodes.length; i++) {
        for (let j = i + 1; j < nodes.length; j++) {
          const a = nodes[i];
          const b = nodes[j];
          const dx = a.x - b.x;
          const dy = a.y - b.y;
          const d = Math.hypot(dx, dy);
          if (d > maxDist) continue;
          const alpha = (1 - d / maxDist) * 0.35;
          const hot = a.hot && b.hot;
          ctx.strokeStyle = hot
            ? `rgba(255, 192, 0, ${alpha.toFixed(3)})`
            : `rgba(255, 255, 255, ${(alpha * 0.5).toFixed(3)})`;
          ctx.beginPath();
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.stroke();
        }
      }

      for (const n of nodes) {
        const r = n.r + (n.hot ? Math.sin(pulse + n.x) * 0.8 : 0);
        ctx.fillStyle = n.hot ? "#FFC000" : "rgba(255,255,255,0.85)";
        ctx.beginPath();
        ctx.arc(n.x, n.y, Math.max(0.5, r), 0, Math.PI * 2);
        ctx.fill();
        if (n.hot) {
          ctx.fillStyle = "rgba(255, 192, 0, 0.12)";
          ctx.beginPath();
          ctx.arc(n.x, n.y, r * 3.5, 0, Math.PI * 2);
          ctx.fill();
        }
      }

      pulse += 0.04;
    };

    const tick = () => {
      if (!localRunning) return;
      draw();
      rafId = requestAnimationFrame(tick);
    };

    const onResize = () => {
      size();
      seed();
      draw();
    };

    size();
    seed();
    draw();
    if (localRunning) rafId = requestAnimationFrame(tick);
    window.addEventListener("resize", onResize);

    const updateRail = () => {
      const h = document.documentElement;
      const max = h.scrollHeight - h.clientHeight;
      const pct = max <= 0 ? 0 : Math.min(1, h.scrollTop / max);
      if (railRef.current) railRef.current.style.width = `${pct * 100}%`;
    };
    updateRail();
    window.addEventListener("scroll", updateRail, { passive: true });

    return () => {
      cancelAnimationFrame(rafId);
      window.removeEventListener("resize", onResize);
      window.removeEventListener("scroll", updateRail);
      localRunning = false;
    };
  }, [running]);

  return (
    <>
      <canvas ref={canvasRef} className={styles.canvas} aria-hidden />
      <button
        className={styles.pause}
        aria-label={running ? "Pause animation" : "Resume animation"}
        onClick={() => setRunning((v) => !v)}
      >
        <svg viewBox="0 0 48 48" width="44" height="44" aria-hidden>
          <polygon
            points="24,2 44,13 44,35 24,46 4,35 4,13"
            fill="none"
            stroke="#fff"
            strokeWidth="1.8"
          />
          {running ? (
            <g>
              <rect x="17" y="16" width="4" height="16" fill="#fff" />
              <rect x="27" y="16" width="4" height="16" fill="#fff" />
            </g>
          ) : (
            <polygon points="18,14 34,24 18,34" fill="#fff" />
          )}
        </svg>
      </button>
      <div className={styles.rail} aria-hidden>
        <span ref={railRef} />
      </div>
    </>
  );
}
