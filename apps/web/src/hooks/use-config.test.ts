import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";

let mockKey: string | undefined;
let mockData: unknown;
vi.mock("swr", () => ({
  default: (key: string, _fetcher: unknown, options?: { fallbackData?: unknown }) => {
    mockKey = key;
    return {
      data: mockData ?? options?.fallbackData ?? undefined,
    };
  },
}));

vi.mock("../api/client.ts", () => ({
  fetcher: vi.fn(),
}));

import { useConfig } from "./use-config.ts";
import { DEFAULT_EFFORTS } from "../components/tasks/TaskTable/constants.ts";

describe("useConfig", () => {
  beforeEach(() => {
    mockData = undefined;
  });

  it("calls SWR with /api/config when no project", () => {
    renderHook(() => useConfig());
    expect(mockKey).toBe("/api/config");
  });

  it("includes project in query string", () => {
    renderHook(() => useConfig("myproject"));
    expect(mockKey).toBe("/api/config?project=myproject");
  });

  it("returns defaults when data is undefined", () => {
    const { result } = renderHook(() => useConfig());
    expect(result.current.readonly).toBe(false);
    expect(result.current.version).toBe("");
    expect(result.current.phases).toEqual([]);
  });

  it("falls back to the default effort vocabulary when the server omits it", () => {
    mockData = { readonly: false, version: "1.0", phases: [] };
    const { result } = renderHook(() => useConfig());
    expect(result.current.efforts).toEqual(DEFAULT_EFFORTS);
  });

  it("returns the server's effort vocabulary when present", () => {
    mockData = {
      readonly: false,
      version: "1.0",
      phases: [],
      efforts: ["xs", "s", "m", "l", "xl"],
    };
    const { result } = renderHook(() => useConfig());
    expect(result.current.efforts).toEqual(["xs", "s", "m", "l", "xl"]);
  });

  it("falls back when the server sends an empty vocabulary", () => {
    mockData = { readonly: false, version: "1.0", phases: [], efforts: [] };
    const { result } = renderHook(() => useConfig());
    expect(result.current.efforts).toEqual(DEFAULT_EFFORTS);
  });
});
