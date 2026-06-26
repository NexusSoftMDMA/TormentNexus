import { Suspense, lazy } from "react";
import { InstallPromptBanner } from "./shell/InstallPromptBanner";

const MessagesApp = lazy(() => import("./apps/MessagesApp").then((m) => ({ default: m.MessagesApp })));

export function ChatStandalone() {
  return (
    <div
      className="w-screen flex flex-col overflow-hidden"
      style={{
        // 100dvh exactly fills the visible standalone area; 100vh (h-screen)
        // resolves to the larger viewport in an installed iOS PWA and leaves
        // dead space at the bottom.
        height: "100dvh",
        backgroundColor: "var(--color-shell-bg)",
        paddingTop: "env(safe-area-inset-top, 0px)",
      }}
    >
      <InstallPromptBanner />
      <Suspense fallback={
        <div className="flex items-center justify-center h-full" style={{ color: "rgba(255,255,255,0.4)" }}>
          Loading…
        </div>
      }>
        <MessagesApp windowId="standalone-chat" title="taOS talk" />
      </Suspense>
    </div>
  );
}
