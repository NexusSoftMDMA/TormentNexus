import { useState } from "react";
import {
  Lock,
  Monitor,
  Tablet,
  Smartphone,
  ClipboardCheck,
  RotateCcw,
  ExternalLink,
  Share2,
} from "lucide-react";

type DeviceMode = "desktop" | "tablet" | "phone";

export function PreviewView() {
  const [deviceMode, setDeviceMode] = useState<DeviceMode>("desktop");

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* view header */}
      <div
        className="flex flex-none items-center gap-3 border-b border-shell-border px-[22px]"
        style={{ height: "54px" }}
      >
        <h2 className="text-[17px] font-bold tracking-[-0.02em]">Preview</h2>
        <span className="text-[12px] text-shell-text-tertiary">
          todo-app - live on fedora-gpu
        </span>
      </div>

      {/* horizontal split */}
      <div className="flex min-h-0 flex-1">
        {/* preview stage */}
        <div className="flex min-w-0 flex-1 flex-col p-[18px]">
          {/* url bar + device toggle */}
          <div className="mb-3.5 flex flex-none items-center gap-2.5">
            <div className="flex h-[34px] flex-1 items-center gap-2.5 rounded-[10px] border border-shell-border bg-shell-surface px-3 text-[12px] text-shell-text-tertiary">
              <Lock size={13} className="text-green-400" />
              todo-app.taos.local
            </div>
            <div className="flex gap-0 rounded-full border border-shell-border bg-shell-surface p-[3px]">
              {(
                [
                  { id: "desktop" as DeviceMode, Icon: Monitor },
                  { id: "tablet" as DeviceMode, Icon: Tablet },
                  { id: "phone" as DeviceMode, Icon: Smartphone },
                ] as { id: DeviceMode; Icon: typeof Monitor }[]
              ).map(({ id, Icon }) => (
                <button
                  key={id}
                  type="button"
                  aria-label={id}
                  onClick={() => setDeviceMode(id)}
                  className={`flex cursor-pointer items-center rounded-full px-[11px] py-[5px] ${
                    deviceMode === id
                      ? "bg-shell-surface-active text-shell-text"
                      : "text-shell-text-tertiary"
                  }`}
                >
                  <Icon size={15} />
                </button>
              ))}
            </div>
          </div>

          {/* frame - simulated todo app in light theme */}
          <div className="flex flex-1 flex-col overflow-hidden rounded-[16px] border border-shell-border-strong">
            {/* app bar */}
            <div
              className="flex flex-none items-center gap-2.5 px-5 text-[15px] font-bold text-white"
              style={{
                height: "54px",
                background: "linear-gradient(135deg,#6f7687,#565d6e)",
              }}
            >
              <ClipboardCheck size={20} color="white" />
              My Tasks
              <span style={{ marginLeft: "auto", fontSize: "12px", fontWeight: 600, opacity: 0.85 }}>
                2 left
              </span>
            </div>

            {/* app body */}
            <div className="flex-1 overflow-auto bg-white px-[22px] py-5">
              {/* add input */}
              <div
                style={{
                  height: "42px",
                  borderRadius: "11px",
                  border: "1px solid #e2e2e6",
                  display: "flex",
                  alignItems: "center",
                  padding: "0 14px",
                  color: "#9a9aa2",
                  fontSize: "13.5px",
                  marginBottom: "14px",
                }}
              >
                Add a task...
              </div>

              {/* todo items */}
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "11px 4px",
                  borderBottom: "1px solid #efeff2",
                  fontSize: "14px",
                  color: "#a6a6ad",
                }}
              >
                <div
                  style={{
                    width: "19px",
                    height: "19px",
                    borderRadius: "6px",
                    background: "#6f7687",
                    border: "1.6px solid #6f7687",
                    flexShrink: 0,
                  }}
                />
                <span style={{ textDecoration: "line-through" }}>Ship Coding Studio mock</span>
                <span
                  style={{
                    marginLeft: "auto",
                    fontSize: "11px",
                    fontWeight: 700,
                    color: "#6f7687",
                    background: "#eef0f3",
                    padding: "3px 9px",
                    borderRadius: "999px",
                  }}
                >
                  today
                </span>
              </div>

              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "11px 4px",
                  borderBottom: "1px solid #efeff2",
                  fontSize: "14px",
                  color: "#26262b",
                }}
              >
                <div
                  style={{
                    width: "19px",
                    height: "19px",
                    borderRadius: "6px",
                    border: "1.6px solid #c2c2c8",
                    flexShrink: 0,
                  }}
                />
                <span>Wire the cluster build runner</span>
                <span
                  style={{
                    marginLeft: "auto",
                    fontSize: "11px",
                    fontWeight: 700,
                    color: "#6f7687",
                    background: "#eef0f3",
                    padding: "3px 9px",
                    borderRadius: "999px",
                  }}
                >
                  soon
                </span>
              </div>

              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "11px 4px",
                  borderBottom: "1px solid #efeff2",
                  fontSize: "14px",
                  color: "#26262b",
                }}
              >
                <div
                  style={{
                    width: "19px",
                    height: "19px",
                    borderRadius: "6px",
                    border: "1.6px solid #c2c2c8",
                    flexShrink: 0,
                  }}
                />
                <span>Add completed-tasks filter</span>
                <span
                  style={{
                    marginLeft: "auto",
                    fontSize: "11px",
                    fontWeight: 700,
                    color: "#6f7687",
                    background: "#eef0f3",
                    padding: "3px 9px",
                    borderRadius: "999px",
                  }}
                >
                  soon
                </span>
              </div>

              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "11px 4px",
                  fontSize: "14px",
                  color: "#a6a6ad",
                }}
              >
                <div
                  style={{
                    width: "19px",
                    height: "19px",
                    borderRadius: "6px",
                    background: "#6f7687",
                    border: "1.6px solid #6f7687",
                    flexShrink: 0,
                  }}
                />
                <span style={{ textDecoration: "line-through" }}>Pick the syntax theme</span>
              </div>
            </div>
          </div>
        </div>

        {/* console panel */}
        <div className="flex w-[300px] flex-none flex-col border-l border-shell-border bg-shell-bg-deep">
          <div className="flex items-center gap-2 px-4 pb-2.5 pt-3.5 text-[12px] font-bold text-shell-text-secondary">
            <div
              className="h-[7px] w-[7px] rounded-full bg-green-400"
              style={{ boxShadow: "0 0 7px #4ade80" }}
            />
            Console - dev server
          </div>

          <div className="flex-1 overflow-auto px-3.5 py-1 font-mono text-[11px] leading-[1.75]">
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:11</span>{" "}
              <span className="text-green-400">ready</span>
              {" vite v5.4 in 412ms"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:11</span>{" "}
              <span className="text-blue-400">hmr</span>
              {" page reload App.tsx"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:18</span>{" "}
              <span className="text-blue-400">hmr</span>
              {" update TodoList.tsx"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:18</span>
              {" useTodos: restored 4 from localStorage"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:24</span>{" "}
              <span className="text-green-400">200</span>
              {" GET / 14ms"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:24</span>
              {" render - 4 todos - 2 remaining"}
            </div>
            <div className="text-shell-text-secondary">
              <span className="text-shell-text-tertiary">18:02:31</span>{" "}
              <span className="text-blue-400">hmr</span>
              {" update styles.css"}
            </div>
          </div>
        </div>
      </div>

      {/* footer bar */}
      <div className="flex flex-none items-center gap-2.5 border-t border-shell-border bg-shell-bg-deep px-[18px] py-3.5">
        <button
          type="button"
          className="flex h-10 cursor-pointer items-center gap-2 rounded-[12px] border border-shell-border bg-shell-surface px-4 text-[12.5px] font-semibold hover:bg-shell-surface-active"
        >
          <RotateCcw size={15} />
          Restart
        </button>
        <button
          type="button"
          className="flex h-10 cursor-pointer items-center gap-2 rounded-[12px] border border-shell-border bg-shell-surface px-4 text-[12.5px] font-semibold hover:bg-shell-surface-active"
        >
          <ExternalLink size={15} />
          Open in browser
        </button>
        <button
          type="button"
          className="ml-auto flex h-10 cursor-pointer items-center gap-2 rounded-[12px] border-0 px-4 text-[12.5px] font-semibold text-white"
          style={{
            background: "linear-gradient(135deg,#a9b0c2,#8b92a3)",
            boxShadow: "0 8px 22px -8px rgba(139,146,163,0.35)",
          }}
        >
          <Share2 size={15} />
          Share to Store
        </button>
      </div>
    </div>
  );
}
