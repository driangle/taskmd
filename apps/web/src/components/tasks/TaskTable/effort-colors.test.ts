import { describe, it, expect } from "vitest";
import { buildEffortBadges } from "./effort-colors.ts";
import { DEFAULT_EFFORTS, EFFORT_COLORS } from "./constants.ts";

const CUSTOM = ["xs", "s", "m", "l", "xl"];

describe("buildEffortBadges", () => {
  it("leaves the default vocabulary on its hand-picked Tailwind classes", () => {
    const badges = buildEffortBadges(DEFAULT_EFFORTS, "light");
    expect(badges.colors).toBe(EFFORT_COLORS);
    expect(badges.styles).toEqual({});
  });

  it("uses the default classes in dark mode too", () => {
    expect(buildEffortBadges(DEFAULT_EFFORTS, "dark").colors).toBe(EFFORT_COLORS);
  });

  it("falls back to the default classes for an empty vocabulary", () => {
    expect(buildEffortBadges([], "light").colors).toBe(EFFORT_COLORS);
  });

  it("generates one style per value of a custom vocabulary", () => {
    const badges = buildEffortBadges(CUSTOM, "light");
    expect(Object.keys(badges.styles)).toEqual(CUSTOM);
    expect(badges.colors).toEqual({});
  });

  it("gives each value of a custom vocabulary a distinct color", () => {
    const badges = buildEffortBadges(CUSTOM, "light");
    const backgrounds = CUSTOM.map((e) => badges.styles[e].backgroundColor);
    expect(new Set(backgrounds).size).toBe(CUSTOM.length);
  });

  it("sets background, text and ring on every generated badge", () => {
    const style = buildEffortBadges(CUSTOM, "light").styles.m;
    expect(style.backgroundColor).toMatch(/^#[0-9a-f]{6}$/);
    expect(style.color).toMatch(/^#[0-9a-f]{6}$/);
    expect(style.boxShadow).toContain("inset");
  });

  it("differs between light and dark themes", () => {
    const light = buildEffortBadges(CUSTOM, "light").styles.m;
    const dark = buildEffortBadges(CUSTOM, "dark").styles.m;
    expect(dark.backgroundColor).not.toBe(light.backgroundColor);
    expect(dark.color).not.toBe(light.color);
  });

  it("handles a single-value vocabulary", () => {
    const badges = buildEffortBadges(["only"], "light");
    expect(Object.keys(badges.styles)).toEqual(["only"]);
  });

  it("handles a vocabulary longer than the anchor set", () => {
    const long = ["a", "b", "c", "d", "e", "f", "g", "h"];
    const badges = buildEffortBadges(long, "dark");
    expect(Object.keys(badges.styles)).toEqual(long);
  });
});
