import type { Metadata, Viewport } from "next";
import { Archivo, JetBrains_Mono } from "next/font/google";
import "./globals.css";

const archivo = Archivo({
  subsets: ["latin"],
  weight: ["400", "500", "700", "900"],
  variable: "--font-archivo",
  display: "swap",
});

const jetbrains = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL("https://agentmemory.dev"),
  title: "AGENTMEMORY — PERSISTENT MEMORY FOR AI CODING AGENTS",
  description:
    "The memory layer your coding agent should have had from day one. 95.2% retrieval R@5. 92% fewer tokens. 0 external databases. Works with every agent.",
  icons: {
    icon: [{ url: "/icon.svg", type: "image/svg+xml" }],
    apple: "/icon.svg",
  },
  openGraph: {
    title: "agentmemory",
    description:
      "Persistent memory for AI coding agents. Runs locally. Zero external databases.",
    type: "website",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    title: "agentmemory",
    description:
      "Persistent memory for AI coding agents. Runs locally. Zero external databases.",
  },
};

export const viewport: Viewport = {
  themeColor: "#000000",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`${archivo.variable} ${jetbrains.variable}`}>
      <body>{children}</body>
    </html>
  );
}
