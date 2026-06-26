import { AppSidebar } from "./app-sidebar";
import { MobileTopBar } from "./mobile-top-bar";

interface AppShellProps {
  children: React.ReactNode;
}

/**
 * App shell. The desktop top bar was retired — Search + New Chat now live
 * inside the sidebar so the run page (and every other text-heavy page)
 * gets the full vertical space. Mobile still needs a thin top bar with
 * just the hamburger that opens the mobile nav.
 */
export function AppShell({ children }: AppShellProps) {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-background text-foreground">
      <AppSidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <MobileTopBar />
        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}
