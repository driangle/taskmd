import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";

const mockMutate = vi.fn();

vi.mock("swr", () => ({
  useSWRConfig: () => ({ mutate: mockMutate }),
}));

import { useLiveReload } from "./use-live-reload.ts";

describe("useLiveReload", () => {
  let mockEventSource: {
    addEventListener: ReturnType<typeof vi.fn>;
    close: ReturnType<typeof vi.fn>;
  };

  beforeEach(() => {
    mockMutate.mockClear();
    mockEventSource = {
      addEventListener: vi.fn(),
      close: vi.fn(),
    };
    // Must be a function expression, not an arrow: vitest 4 only lets mocks be
    // called with `new` when the implementation is constructible.
    vi.stubGlobal(
      "EventSource",
      vi.fn(function () {
        return mockEventSource;
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("creates an EventSource on /api/events", () => {
    renderHook(() => useLiveReload());
    expect(EventSource).toHaveBeenCalledWith("/api/events");
  });

  it("registers reload and error event listeners", () => {
    renderHook(() => useLiveReload());
    expect(mockEventSource.addEventListener).toHaveBeenCalledWith(
      "reload",
      expect.any(Function),
    );
    expect(mockEventSource.addEventListener).toHaveBeenCalledWith(
      "error",
      expect.any(Function),
    );
  });

  it("calls mutate on reload event", () => {
    renderHook(() => useLiveReload());
    const reloadHandler = mockEventSource.addEventListener.mock.calls.find(
      (call: string[]) => call[0] === "reload",
    )![1] as () => void;
    reloadHandler();
    expect(mockMutate).toHaveBeenCalledWith(expect.any(Function));
  });

  it("closes EventSource on unmount", () => {
    const { unmount } = renderHook(() => useLiveReload());
    unmount();
    expect(mockEventSource.close).toHaveBeenCalled();
  });
});
