import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { LoginGate } from "./LoginGate";

describe("LoginGate host reachability", () => {
  afterEach(() => { vi.restoreAllMocks(); });

  it("shows the off-network screen when /auth/status is unreachable", async () => {
    // A thrown fetch is a network failure (host unreachable), not an HTTP error.
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("network down"));
    render(
      <LoginGate>
        <div>the desktop shell</div>
      </LoginGate>,
    );
    expect(await screen.findByText("Can't reach your taOS")).toBeInTheDocument();
    // The broken shell must NOT render behind it.
    expect(screen.queryByText("the desktop shell")).not.toBeInTheDocument();
  });

  it("renders the app shell when authenticated and reachable", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ configured: true, authenticated: true }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    render(
      <LoginGate>
        <div>the desktop shell</div>
      </LoginGate>,
    );
    await waitFor(() =>
      expect(screen.getByText("the desktop shell")).toBeInTheDocument(),
    );
  });
});
