import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { AccountSection } from "./AccountPanel";

describe("AccountSection", () => {
  afterEach(() => { vi.restoreAllMocks(); });
  beforeEach(() => { vi.restoreAllMocks(); });

  it("shows the sign-in / create-account form when signed out (401)", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 401 }),
    );
    render(<AccountSection />);
    expect(await screen.findByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByText("Create account")).toBeInTheDocument();
  });

  it("shows an unavailable state when the account service cannot be reached", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("network"));
    render(<AccountSection />);
    expect(
      await screen.findByText(/account service is not reachable/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Retry")).toBeInTheDocument();
  });

  it("shows the signed-in view with email + taOSgo status", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          user_id: "u1",
          email: "jay@example.com",
          taosgo: { status: "none" },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    render(<AccountSection />);
    expect(await screen.findByText("jay@example.com")).toBeInTheDocument();
    expect(screen.getByText("taOSgo")).toBeInTheDocument();
    expect(screen.getByText("Not subscribed")).toBeInTheDocument();
    expect(screen.getByText("Start 7-day free trial")).toBeInTheDocument();
  });

  it("surfaces a backend error message on failed sign-in", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // initial /me
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ error: "Invalid credentials" }), {
          status: 400,
          headers: { "Content-Type": "application/json" },
        }),
      );
    render(<AccountSection />);
    fireEvent.change(await screen.findByLabelText("Email"), {
      target: { value: "jay@example.com" },
    });
    fireEvent.change(screen.getByLabelText("Password"), {
      target: { value: "password1" },
    });
    // Two controls read "Sign in" (the mode tab and the submit button); the
    // submit is the last one.
    fireEvent.click(screen.getAllByRole("button", { name: "Sign in" }).at(-1)!);
    await waitFor(() =>
      expect(screen.getByText("Invalid credentials")).toBeInTheDocument(),
    );
  });
});
