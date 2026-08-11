import { describe, it, expect } from "vitest";
import { hasActiveFilters, defaultFilterState } from "./filters.ts";
import { STATUSES, PRIORITIES, DEFAULT_EFFORTS, TYPES } from "./constants.ts";

describe("hasActiveFilters", () => {
  it("returns false for default filter state", () => {
    expect(hasActiveFilters(defaultFilterState())).toBe(false);
  });

  it("returns true when a status is deselected", () => {
    const filters = { ...defaultFilterState(), selectedStatuses: new Set(STATUSES.slice(0, 3)) };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("returns true when a priority is deselected", () => {
    const filters = { ...defaultFilterState(), selectedPriorities: new Set(PRIORITIES.slice(0, 2)) };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("returns true when a type is deselected", () => {
    const filters = { ...defaultFilterState(), selectedTypes: new Set(TYPES.slice(0, 3)) };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("returns true when tags are selected", () => {
    const filters = { ...defaultFilterState(), selectedTags: new Set(["api"]) };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("returns true when effort is selected", () => {
    const filters = { ...defaultFilterState(), selectedEffort: new Set(["small"]) };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("is inactive when a full custom vocabulary is selected", () => {
    const efforts = ["xs", "s", "m", "l", "xl"];
    expect(hasActiveFilters(defaultFilterState(efforts), efforts)).toBe(false);
  });

  it("is active when a custom vocabulary is partly selected", () => {
    const efforts = ["xs", "s", "m", "l", "xl"];
    const filters = { ...defaultFilterState(efforts), selectedEffort: new Set(["xs"]) };
    expect(hasActiveFilters(filters, efforts)).toBe(true);
  });

  it("returns true when global filter has text", () => {
    const filters = { ...defaultFilterState(), globalFilter: "search" };
    expect(hasActiveFilters(filters)).toBe(true);
  });

  it("returns false when all statuses/priorities/types selected and nothing else active", () => {
    const filters = {
      selectedStatuses: new Set(STATUSES),
      selectedPriorities: new Set(PRIORITIES),
      selectedTypes: new Set(TYPES),
      selectedTags: new Set<string>(),
      selectedEffort: new Set(DEFAULT_EFFORTS),
      selectedPhases: new Set<string>(),
      globalFilter: "",
    };
    expect(hasActiveFilters(filters)).toBe(false);
  });
});

describe("defaultFilterState", () => {
  it("has all statuses selected", () => {
    expect(defaultFilterState().selectedStatuses).toEqual(new Set(STATUSES));
  });

  it("includes in-review in default statuses", () => {
    expect(defaultFilterState().selectedStatuses.has("in-review")).toBe(true);
  });

  it("has all priorities selected", () => {
    expect(defaultFilterState().selectedPriorities).toEqual(new Set(PRIORITIES));
  });

  it("has all types selected", () => {
    expect(defaultFilterState().selectedTypes).toEqual(new Set(TYPES));
  });

  it("has no tags selected", () => {
    expect(defaultFilterState().selectedTags.size).toBe(0);
  });

  it("has all efforts selected", () => {
    expect(defaultFilterState().selectedEffort).toEqual(new Set(DEFAULT_EFFORTS));
  });

  it("has empty global filter", () => {
    expect(defaultFilterState().globalFilter).toBe("");
  });
});
