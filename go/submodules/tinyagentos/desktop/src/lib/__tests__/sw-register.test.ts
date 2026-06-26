import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { registerServiceWorker } from "../sw-register";

describe("registerServiceWorker", () => {
  let originalSW: any;
  beforeEach(() => {
    originalSW = (navigator as any).serviceWorker;
  });
  afterEach(() => {
    Object.defineProperty(navigator, "serviceWorker", {
      value: originalSW, writable: true, configurable: true,
    });
    vi.restoreAllMocks();
  });

  it("does nothing if serviceWorker is unavailable", async () => {
    Object.defineProperty(navigator, "serviceWorker", {
      value: undefined, writable: true, configurable: true,
    });
    await expect(registerServiceWorker()).resolves.toBeUndefined();
  });

  it("calls navigator.serviceWorker.register('/sw.js')", async () => {
    const register = vi.fn().mockResolvedValue({ scope: "/" });
    Object.defineProperty(navigator, "serviceWorker", {
      value: { register }, writable: true, configurable: true,
    });
    await registerServiceWorker();
    expect(register).toHaveBeenCalledWith("/sw.js");
  });

  it("swallows registration errors (logs only)", async () => {
    const consoleErr = vi.spyOn(console, "warn").mockImplementation(() => {});
    const register = vi.fn().mockRejectedValue(new Error("nope"));
    Object.defineProperty(navigator, "serviceWorker", {
      value: { register }, writable: true, configurable: true,
    });
    await expect(registerServiceWorker()).resolves.toBeUndefined();
    expect(consoleErr).toHaveBeenCalled();
  });
});
