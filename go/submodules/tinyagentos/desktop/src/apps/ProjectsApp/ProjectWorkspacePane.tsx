import { useRef, useState } from "react";
import type { Project } from "@/lib/projects";
import { MessagesApp } from "@/apps/MessagesApp";
import { CanvasView } from "./canvas/CanvasView";
import { useIsMobile } from "../../hooks/use-is-mobile";
import styles from "./ProjectsApp.module.css";

type PreviewMode = "preview" | "code" | "canvas";

const MIN_RIGHT = 320;
const MIN_LEFT = 360;

/**
 * The Workspace tab (task #59 hero): a split pane.
 *
 *   LEFT:  the project channel thread + composer. Reuses the existing
 *          project-scoped MessagesApp (humans + agents, send logic, A2A bus),
 *          so this is real data, not a mock.
 *   RIGHT: a live-preview pane with a Preview | Code | Canvas segmented
 *          toggle and a small toolbar.
 *
 * Phase 1 scope: the Canvas toggle embeds the real project canvas. Preview and
 * Code render honest placeholders rather than faking a running app build. A
 * true streamed live build preview is #59 phase 2/3 (see TODO below).
 */
export function ProjectWorkspacePane({ project }: { project: Project }) {
  const isMobile = useIsMobile();
  const [mode, setMode] = useState<PreviewMode>("preview");
  const [rightWidth, setRightWidth] = useState(472);
  const containerRef = useRef<HTMLDivElement>(null);
  const draggingRef = useRef(false);

  const onDividerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    e.preventDefault();
    draggingRef.current = true;
    e.currentTarget.setPointerCapture(e.pointerId);
  };
  const onDividerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!draggingRef.current || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const next = rect.right - e.clientX;
    setRightWidth(Math.max(MIN_RIGHT, Math.min(rect.width - MIN_LEFT, next)));
  };
  const onDividerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    draggingRef.current = false;
    try {
      e.currentTarget.releasePointerCapture(e.pointerId);
    } catch {
      // pointer may already be released
    }
  };

  return (
    <div className={styles.ws} ref={containerRef}>
      {/* LEFT: real project channel thread + composer. Desktop only -- on mobile
          the squeezed split made messages unreadable, so Messages is its own
          full-page tab and the Workspace pane shows just the preview. */}
      {!isMobile && (
        <>
          <div className={styles.wsLeft}>
            <div className={styles.wsThread}>
              <MessagesApp
                key={project.id}
                windowId={`project-workspace-messages-${project.id}`}
                scope={{ projectId: project.id }}
              />
            </div>
          </div>

          {/* draggable resize divider */}
          <div
            className={styles.wsDivider}
            role="separator"
            aria-orientation="vertical"
            aria-label="Resize panes"
            onPointerDown={onDividerDown}
            onPointerMove={onDividerMove}
            onPointerUp={onDividerUp}
          />
        </>
      )}

      {/* RIGHT: preview / code / canvas (full width on mobile) */}
      <div className={styles.wsRight} style={isMobile ? { width: "100%", flex: 1 } : { width: rightWidth }}>
        <div className={styles.pvBar}>
          <div className={styles.seg} role="tablist" aria-label="Preview mode">
            {(["preview", "code", "canvas"] as PreviewMode[]).map((m) => (
              <button
                key={m}
                type="button"
                role="tab"
                aria-selected={mode === m}
                className={mode === m ? styles.segOn : ""}
                onClick={() => setMode(m)}
              >
                {m.charAt(0).toUpperCase() + m.slice(1)}
              </button>
            ))}
          </div>
          <div className={styles.pvLive} aria-hidden>
            <span className={styles.pvPulse}><i /></span>
            Live
          </div>
          <div className={styles.pvTools}>
            {/* These toolbar actions are intentionally inert in phase 1: refresh,
                device-size and open-in-window all hang off a real streamed build
                preview, which is deferred. They are shown so the chrome matches
                the approved mock without faking behavior. */}
            <button type="button" className={styles.pvTool} title="Refresh preview" disabled aria-label="Refresh preview">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M3 12a9 9 0 0 1 15-6.7L21 8" /><path d="M21 3v5h-5" />
                <path d="M21 12a9 9 0 0 1-15 6.7L3 16" /><path d="M3 21v-5h5" />
              </svg>
            </button>
            <button type="button" className={styles.pvTool} title="Device size" disabled aria-label="Device size">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                <rect x="7" y="3" width="10" height="18" rx="2" /><path d="M11 18h2" />
              </svg>
            </button>
            <button type="button" className={styles.pvTool} title="Open in window" disabled aria-label="Open in window">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8">
                <path d="M14 4h6v6" /><path d="M20 4 10 14" />
                <path d="M20 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1h5" />
              </svg>
            </button>
          </div>
        </div>

        {mode === "canvas" ? (
          <div className={styles.pvCanvas}>
            <CanvasView projectId={project.id} projectSlug={project.slug} />
          </div>
        ) : mode === "code" ? (
          <div className={styles.pvStage}>
            <PreviewPlaceholder
              title="Code view"
              body="When this project builds an app, its source will stream here alongside the live preview. Wiring the build pipeline is the next phase of the Workspace."
            />
          </div>
        ) : (
          <div className={styles.pvStage}>
            {/* TODO(#59 phase 2/3): replace with a real streamed live build
                preview (iframe / StreamedBrowser to the running app). Until the
                project build pipeline is wired, show an honest placeholder
                rather than a hardcoded fake running app. */}
            <PreviewPlaceholder
              title="Live preview"
              body="A live preview of the app this project builds will render here. Until a build is running, there is nothing to show. Use the Canvas tab for the shared diagram."
            />
          </div>
        )}
      </div>
    </div>
  );
}

function PreviewPlaceholder({ title, body }: { title: string; body: string }) {
  return (
    <div className={styles.placeholder}>
      <div className={styles.phIc}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6">
          <rect x="3" y="3" width="18" height="18" rx="3" />
          <path d="M3 9h18M9 3v18" opacity=".5" />
        </svg>
      </div>
      <h3>{title}</h3>
      <p>{body}</p>
      <span className={styles.tagb}>Phase 2</span>
    </div>
  );
}
