'use client';

import React, { useEffect, useRef } from 'react';

export default function Confetti() {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;

        interface Particle {
            x: number;
            y: number;
            w: number;
            h: number;
            vx: number;
            vy: number;
            color: string;
            gravity: number;
            drag: number;
        }

        const particles: Particle[] = [];
        const colors = ['#f59e0b', '#ef4444', '#3b82f6', '#10b981', '#ffffff'];

        for (let i = 0; i < 150; i++) {
            particles.push({
                x: window.innerWidth / 2,
                y: window.innerHeight / 2,
                w: Math.random() * 8 + 2,
                h: Math.random() * 8 + 2,
                vx: (Math.random() - 0.5) * 20,
                vy: (Math.random() - 0.5) * 20 - 5,
                color: colors[Math.floor(Math.random() * colors.length)],
                gravity: 0.5,
                drag: 0.95
            });
        }

        let animationId: number;

        const render = () => {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            let active = false;

            particles.forEach(p => {
                p.x += p.vx;
                p.y += p.vy;
                p.vy += p.gravity;
                p.vx *= p.drag;
                p.vy *= p.drag;

                if (p.y < canvas.height + 20) {
                    active = true;
                    ctx.fillStyle = p.color;
                    ctx.fillRect(p.x, p.y, p.w, p.h);
                }
            });

            if (active) {
                animationId = requestAnimationFrame(render);
            }
        };

        render();

        return () => {
            cancelAnimationFrame(animationId);
        };
    }, []);

    return (
        <canvas
            ref={canvasRef}
            className="fixed inset-0 pointer-events-none z-[100]"
        />
    );
}
