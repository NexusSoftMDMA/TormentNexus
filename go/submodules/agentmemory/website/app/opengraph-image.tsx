import { ImageResponse } from "next/og";

export const alt = "agentmemory — persistent memory for AI coding agents";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: "72px 80px",
          background:
            "radial-gradient(ellipse at 20% 0%, #1a1407 0%, #0a0a0a 55%, #000 100%)",
          color: "#fff",
          fontFamily: "system-ui, -apple-system, sans-serif",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            fontSize: 20,
            letterSpacing: 3,
            fontWeight: 700,
            color: "#f3b840",
            textTransform: "uppercase",
          }}
        >
          <div
            style={{
              width: 14,
              height: 14,
              background: "#f3b840",
              boxShadow: "0 0 24px #f3b840",
            }}
          />
          <span>AGENTMEMORY</span>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 28 }}>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              fontSize: 128,
              fontWeight: 900,
              lineHeight: 0.95,
              letterSpacing: -3,
              textTransform: "uppercase",
            }}
          >
            <span>AGENT</span>
            <span style={{ color: "#f3b840" }}>MEMORY</span>
          </div>
          <div
            style={{
              display: "flex",
              fontSize: 32,
              color: "#b8b8b8",
              letterSpacing: 0.5,
              maxWidth: 900,
              lineHeight: 1.35,
            }}
          >
            Persistent memory for AI coding agents. Runs locally. Zero external
            databases.
          </div>
        </div>

        <div
          style={{
            display: "flex",
            gap: 48,
            fontSize: 22,
            color: "#888",
            letterSpacing: 2,
            textTransform: "uppercase",
            fontWeight: 700,
            borderTop: "1px solid #222",
            paddingTop: 28,
          }}
        >
          <div style={{ display: "flex", gap: 10 }}>
            <span style={{ color: "#f3b840" }}>95.2%</span>
            <span>RETRIEVAL R@5</span>
          </div>
          <div style={{ display: "flex", gap: 10 }}>
            <span style={{ color: "#f3b840" }}>92%</span>
            <span>FEWER TOKENS</span>
          </div>
          <div style={{ display: "flex", gap: 10 }}>
            <span style={{ color: "#f3b840" }}>0</span>
            <span>EXTERNAL DBs</span>
          </div>
        </div>
      </div>
    ),
    { ...size },
  );
}
