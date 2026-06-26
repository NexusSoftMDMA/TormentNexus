"use client"

import { useState, useEffect, RefObject } from 'react';

export function useResizeObserver(ref: RefObject<HTMLElement>) {
    const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

    useEffect(() => {
        const element = ref.current;
        if (!element) return;

        // Initial dimensions
        setDimensions({
            width: element.clientWidth,
            height: element.clientHeight,
        });

        const ResizeObserver = (window as any).ResizeObserver;
        if (!ResizeObserver) return;

        let animationFrameId: number;

        const ro = new ResizeObserver((entries: ResizeObserverEntry[]) => {
            if (!Array.isArray(entries) || !entries.length) return;

            const entry = entries[0];

            if (animationFrameId) {
                window.cancelAnimationFrame(animationFrameId);
            }

            animationFrameId = window.requestAnimationFrame(() => {
                setDimensions({
                    width: entry.contentRect.width,
                    height: entry.contentRect.height,
                });
            });
        });

        ro.observe(element);

        return () => {
            ro.disconnect();
            if (animationFrameId) {
                window.cancelAnimationFrame(animationFrameId);
            }
        };
    }, [ref]);

    return dimensions;
}
