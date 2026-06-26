import { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import type { SceneKind } from "./types";

/* ------------------------------------------------------------------ */
/*  useGameScene — the real three.js preview                           */
/*                                                                     */
/*  This is the genuinely-working part of phase 1. It mounts a          */
/*  WebGLRenderer (standard materials, no WebGPU/TSL) into a host div,  */
/*  builds a lit demo scene, drives a requestAnimationFrame loop, and   */
/*  exposes play/pause, a real FPS readout and a scene reset.           */
/*                                                                      */
/*  Controls: WASD / arrow keys move the hero on the runner scene;      */
/*  pointer-drag orbits the camera on both scenes. prefers-reduced-     */
/*  motion freezes idle drift but keeps the loop responsive to input.   */
/* ------------------------------------------------------------------ */

export interface GameSceneHandle {
  /** Attach to the host element the canvas mounts into. */
  hostRef: React.RefObject<HTMLDivElement | null>;
  playing: boolean;
  setPlaying: (v: boolean) => void;
  togglePlay: () => void;
  /** Live frames-per-second from the rAF loop (real, integer). */
  fps: number;
  /** Reset the hero + camera to the scene's start pose. */
  reset: () => void;
  /** WebGL availability — false renders an honest fallback instead of a blank stage. */
  supported: boolean;
}

const prefersReducedMotion = (): boolean =>
  typeof window !== "undefined" &&
  window.matchMedia?.("(prefers-reduced-motion: reduce)").matches === true;

export function useGameScene(scene: SceneKind): GameSceneHandle {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const [playing, setPlayingState] = useState(false);
  const [fps, setFps] = useState(60);
  const [supported, setSupported] = useState(true);

  // Mutable engine refs kept out of React state so the loop never re-renders.
  const playingRef = useRef(false);
  const keysRef = useRef<Set<string>>(new Set());
  const resetRef = useRef<() => void>(() => {});
  const setPlayingRef = useRef<(v: boolean) => void>(() => {});

  const setPlaying = (v: boolean) => {
    playingRef.current = v;
    setPlayingState(v);
  };
  setPlayingRef.current = setPlaying;
  const togglePlay = () => setPlayingRef.current(!playingRef.current);
  const reset = () => resetRef.current();

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;

    let renderer: THREE.WebGLRenderer;
    try {
      renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
    } catch {
      setSupported(false);
      return;
    }
    setSupported(true);

    const reduced = prefersReducedMotion();
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    host.appendChild(renderer.domElement);
    renderer.domElement.style.width = "100%";
    renderer.domElement.style.height = "100%";
    renderer.domElement.style.display = "block";

    const sceneObj = new THREE.Scene();
    sceneObj.background = new THREE.Color("#0a0d14");
    sceneObj.fog = new THREE.Fog("#0a0d14", 14, 42);

    const camera = new THREE.PerspectiveCamera(55, 16 / 10, 0.1, 100);

    // ---- lighting (lit standard materials) ----
    sceneObj.add(new THREE.AmbientLight("#5a6680", 0.65));
    const key = new THREE.DirectionalLight("#cfd6ff", 1.15);
    key.position.set(6, 12, 8);
    key.castShadow = true;
    key.shadow.mapSize.set(1024, 1024);
    key.shadow.camera.near = 1;
    key.shadow.camera.far = 50;
    sceneObj.add(key);
    const rim = new THREE.DirectionalLight("#8b92a3", 0.5);
    rim.position.set(-8, 5, -6);
    sceneObj.add(rim);

    // ---- ground plane ----
    const ground = new THREE.Mesh(
      new THREE.PlaneGeometry(60, 120),
      new THREE.MeshStandardMaterial({
        color: "#141d33",
        roughness: 0.95,
        metalness: 0.05,
      }),
    );
    ground.rotation.x = -Math.PI / 2;
    ground.receiveShadow = true;
    sceneObj.add(ground);

    const grid = new THREE.GridHelper(60, 30, "#3a4a7a", "#1d2740");
    (grid.material as THREE.Material).transparent = true;
    (grid.material as THREE.Material).opacity = 0.4;
    sceneObj.add(grid);

    // ---- hero: a lit low-poly cube the player can drive ----
    const hero = new THREE.Mesh(
      new THREE.BoxGeometry(1.1, 1.1, 1.1),
      new THREE.MeshStandardMaterial({
        color: "#ff8a3d",
        roughness: 0.4,
        metalness: 0.1,
        emissive: "#3a1c08",
      }),
    );
    hero.castShadow = true;
    hero.position.set(0, 0.55, 0);
    sceneObj.add(hero);

    // ---- scene-specific props ----
    const movers: THREE.Mesh[] = [];
    const coins: THREE.Mesh[] = [];

    if (scene === "runner") {
      // Oncoming low-poly obstacles + spinning coins down a runway.
      const obstacleColors = ["#c8503e", "#3d7fb5", "#41d0a3", "#b57ad0"];
      for (let i = 0; i < 7; i++) {
        const m = new THREE.Mesh(
          new THREE.BoxGeometry(1.4, 1.4, 1.4),
          new THREE.MeshStandardMaterial({
            color: obstacleColors[i % obstacleColors.length],
            roughness: 0.5,
            metalness: 0.1,
          }),
        );
        m.castShadow = true;
        m.position.set((i % 3) * 3 - 3, 0.7, -8 - i * 6);
        sceneObj.add(m);
        movers.push(m);
      }
      const coinMat = new THREE.MeshStandardMaterial({
        color: "#ffd35c",
        roughness: 0.3,
        metalness: 0.4,
        emissive: "#5a4a10",
      });
      for (let i = 0; i < 6; i++) {
        const c = new THREE.Mesh(
          new THREE.CylinderGeometry(0.45, 0.45, 0.12, 18),
          coinMat,
        );
        c.castShadow = true;
        c.rotation.x = Math.PI / 2;
        c.position.set((i % 3) * 3 - 3, 0.9, -4 - i * 5);
        sceneObj.add(c);
        coins.push(c);
      }
    } else {
      // Orbit scene: a ring of floating low-poly orbs around the hero.
      const orbColors = ["#41d0a3", "#3d7fb5", "#ffd35c", "#b57ad0", "#ff8a3d"];
      for (let i = 0; i < 8; i++) {
        const a = (i / 8) * Math.PI * 2;
        const orb = new THREE.Mesh(
          new THREE.IcosahedronGeometry(0.7, 0),
          new THREE.MeshStandardMaterial({
            color: orbColors[i % orbColors.length],
            roughness: 0.35,
            metalness: 0.15,
            flatShading: true,
          }),
        );
        orb.castShadow = true;
        orb.position.set(Math.cos(a) * 5, 1.4 + Math.sin(a * 2) * 0.6, Math.sin(a) * 5);
        sceneObj.add(orb);
        movers.push(orb);
      }
    }

    // ---- camera orbit state (pointer-drag) ----
    const cam = { theta: scene === "runner" ? 0 : 0.5, phi: 0.9, radius: 12 };
    const applyCamera = () => {
      const r = cam.radius;
      const tx = scene === "runner" ? hero.position.x * 0.4 : 0;
      camera.position.set(
        tx + r * Math.sin(cam.phi) * Math.sin(cam.theta),
        r * Math.cos(cam.phi),
        scene === "runner"
          ? hero.position.z + r * Math.sin(cam.phi) * Math.cos(cam.theta) + 6
          : r * Math.sin(cam.phi) * Math.cos(cam.theta),
      );
      camera.lookAt(
        tx,
        1,
        scene === "runner" ? hero.position.z - 2 : 0,
      );
    };
    applyCamera();

    let dragging = false;
    let lastX = 0;
    let lastY = 0;
    const onPointerDown = (e: PointerEvent) => {
      dragging = true;
      lastX = e.clientX;
      lastY = e.clientY;
      renderer.domElement.setPointerCapture?.(e.pointerId);
    };
    const onPointerMove = (e: PointerEvent) => {
      if (!dragging) return;
      cam.theta -= (e.clientX - lastX) * 0.006;
      cam.phi = Math.min(1.35, Math.max(0.25, cam.phi - (e.clientY - lastY) * 0.005));
      lastX = e.clientX;
      lastY = e.clientY;
    };
    const onPointerUp = () => {
      dragging = false;
    };
    renderer.domElement.addEventListener("pointerdown", onPointerDown);
    renderer.domElement.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", onPointerUp);

    // ---- keyboard capture for WASD / arrows ----
    // Listeners live on the canvas (made focusable) so movement keys are only
    // intercepted while the user is genuinely focused on the game, never OS-wide.
    const MOVE_KEYS = ["w", "a", "s", "d", "arrowup", "arrowdown", "arrowleft", "arrowright"];
    const canvas = renderer.domElement;
    canvas.tabIndex = 0;
    canvas.style.outline = "none";
    // Focus the canvas on pointer-down so a click hands movement input to the game.
    const focusCanvas = () => canvas.focus();
    canvas.addEventListener("pointerdown", focusCanvas);

    const onKeyDown = (e: KeyboardEvent) => {
      const k = e.key.toLowerCase();
      if (MOVE_KEYS.includes(k)) {
        keysRef.current.add(k);
        // Only swallow the key while actively playing; the canvas already holds
        // focus to receive this event, so page scroll/shortcuts stay free elsewhere.
        if (playingRef.current) e.preventDefault();
      }
    };
    const onKeyUp = (e: KeyboardEvent) => {
      keysRef.current.delete(e.key.toLowerCase());
    };
    // Release every held key when focus is lost, so a key held during a
    // tab/window switch can never stick and drive the character forever.
    const clearKeys = () => keysRef.current.clear();
    const onVisibility = () => {
      if (document.hidden) clearKeys();
    };
    canvas.addEventListener("keydown", onKeyDown);
    canvas.addEventListener("keyup", onKeyUp);
    canvas.addEventListener("blur", clearKeys);
    canvas.addEventListener("pointerleave", clearKeys);
    window.addEventListener("blur", clearKeys);
    document.addEventListener("visibilitychange", onVisibility);

    // ---- reset ----
    const startHero = hero.position.clone();
    const startCam = { ...cam };
    resetRef.current = () => {
      hero.position.copy(startHero);
      cam.theta = startCam.theta;
      cam.phi = startCam.phi;
      cam.radius = startCam.radius;
      applyCamera();
      renderer.render(sceneObj, camera);
    };

    // ---- resize to host ----
    const resize = () => {
      const w = host.clientWidth || 1;
      const h = host.clientHeight || 1;
      renderer.setSize(w, h, false);
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
    };
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(host);

    // ---- main loop ----
    const clock = new THREE.Clock();
    let raf = 0;
    let frames = 0;
    let fpsAccum = 0;

    const tick = () => {
      raf = requestAnimationFrame(tick);
      const dt = Math.min(clock.getDelta(), 0.05);

      // FPS sampling once per ~500ms from real frame timing.
      frames++;
      fpsAccum += dt;
      if (fpsAccum >= 0.5) {
        setFps(Math.round(frames / fpsAccum));
        frames = 0;
        fpsAccum = 0;
      }

      if (playingRef.current) {
        const keys = keysRef.current;
        const speed = 7 * dt;
        if (scene === "runner") {
          // WASD / arrows steer the hero across lanes + along the runway.
          if (keys.has("a") || keys.has("arrowleft")) hero.position.x -= speed;
          if (keys.has("d") || keys.has("arrowright")) hero.position.x += speed;
          if (keys.has("w") || keys.has("arrowup")) hero.position.z -= speed;
          if (keys.has("s") || keys.has("arrowdown")) hero.position.z += speed;
          hero.position.x = Math.max(-6, Math.min(6, hero.position.x));
          hero.rotation.y += dt * 1.2;

          // Obstacles roll toward the hero; recycle past the camera.
          for (const m of movers) {
            m.position.z += 9 * dt;
            m.rotation.x += dt;
            if (m.position.z > hero.position.z + 8) m.position.z -= 48;
          }
          for (const c of coins) {
            c.rotation.z += dt * 3;
            c.position.z += 9 * dt;
            if (c.position.z > hero.position.z + 8) c.position.z -= 40;
          }
        } else {
          // Orbit scene: WASD nudges the hero; orbs bob + circle.
          if (keys.has("a") || keys.has("arrowleft")) hero.position.x -= speed;
          if (keys.has("d") || keys.has("arrowright")) hero.position.x += speed;
          if (keys.has("w") || keys.has("arrowup")) hero.position.z -= speed;
          if (keys.has("s") || keys.has("arrowdown")) hero.position.z += speed;
          hero.rotation.y += dt * 0.8;
          const t = clock.elapsedTime;
          movers.forEach((orb, i) => {
            orb.rotation.x += dt * 0.6;
            orb.rotation.y += dt * 0.4;
            orb.position.y = 1.4 + Math.sin(t * 1.4 + i) * 0.5;
          });
        }

        // Idle camera drift, suppressed under reduced-motion.
        if (!reduced && !dragging && scene === "orbit") cam.theta += dt * 0.12;
      }

      applyCamera();
      renderer.render(sceneObj, camera);
    };
    // Render one frame immediately so the stage is never blank before Play.
    tick();

    return () => {
      cancelAnimationFrame(raf);
      ro.disconnect();
      renderer.domElement.removeEventListener("pointerdown", onPointerDown);
      renderer.domElement.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", onPointerUp);
      canvas.removeEventListener("pointerdown", focusCanvas);
      canvas.removeEventListener("keydown", onKeyDown);
      canvas.removeEventListener("keyup", onKeyUp);
      canvas.removeEventListener("blur", clearKeys);
      canvas.removeEventListener("pointerleave", clearKeys);
      window.removeEventListener("blur", clearKeys);
      document.removeEventListener("visibilitychange", onVisibility);
      keysRef.current.clear();
      sceneObj.traverse((obj) => {
        const mesh = obj as THREE.Mesh;
        if (mesh.geometry) mesh.geometry.dispose();
        const mat = mesh.material as THREE.Material | THREE.Material[] | undefined;
        if (Array.isArray(mat)) mat.forEach((m) => m.dispose());
        else mat?.dispose();
      });
      renderer.dispose();
      if (renderer.domElement.parentNode === host) {
        host.removeChild(renderer.domElement);
      }
    };
  }, [scene]);

  return { hostRef, playing, setPlaying, togglePlay, fps, reset, supported };
}
