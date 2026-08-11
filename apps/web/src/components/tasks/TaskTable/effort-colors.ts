import { useMemo } from "react";
import type { CSSProperties } from "react";
import kleur, { Palette } from "@driangle/kleur";
import { useTheme } from "../../../hooks/use-theme.ts";
import { DEFAULT_EFFORTS, EFFORT_COLORS } from "./constants.ts";

/**
 * Effort badge colors for a project's configured vocabulary.
 *
 * `colors` holds Tailwind classes, `styles` holds inline styles. Tailwind cannot
 * generate classes for values known only at runtime, so a custom vocabulary is
 * styled inline; the default vocabulary keeps its hand-picked classes so
 * projects without an `effort:` config look exactly as they always have.
 */
export interface EffortBadges {
  colors: Record<string, string>;
  styles: Record<string, CSSProperties>;
}

/**
 * Hues anchoring the generated ramp, matching the default small / medium / large
 * badges (emerald-500, amber-500, purple-500). Effort is an ordinal scale, so a
 * vocabulary of any length is interpolated across these rather than assigned
 * arbitrary colors.
 */
const ANCHORS = ["#10b981", "#f59e0b", "#a855f7"];

const DEFAULT_BADGES: EffortBadges = { colors: EFFORT_COLORS, styles: {} };

function isDefaultVocabulary(efforts: string[]): boolean {
  return (
    efforts.length === DEFAULT_EFFORTS.length &&
    efforts.every((e, i) => e === DEFAULT_EFFORTS[i])
  );
}

/** Mix ratios approximating the Tailwind shades the default badges use. */
const LIGHT_BG_MIX = 0.8; // ≈ -100
const LIGHT_TEXT_MIX = 0.3; // ≈ -700
const LIGHT_RING_MIX = 0.5; // ≈ -300
const DARK_BG_ALPHA = 0.18; // ≈ -900/30
const DARK_TEXT_MIX = 0.3; // ≈ -400
const DARK_RING_MIX = 0.3; // ≈ -700

function badgeStyle(hex: string, theme: "light" | "dark"): CSSProperties {
  const base = kleur.hex(hex);
  if (theme === "dark") {
    return {
      backgroundColor: base.withAlpha(DARK_BG_ALPHA).toCss(),
      color: kleur.mix(base, kleur.white, DARK_TEXT_MIX).toHex(),
      boxShadow: `inset 0 0 0 1px ${kleur.mix(base, kleur.black, DARK_RING_MIX).toHex()}`,
    };
  }
  return {
    backgroundColor: kleur.mix(base, kleur.white, LIGHT_BG_MIX).toHex(),
    color: kleur.mix(base, kleur.black, LIGHT_TEXT_MIX).toHex(),
    boxShadow: `inset 0 0 0 1px ${kleur.mix(base, kleur.white, LIGHT_RING_MIX).toHex()}`,
  };
}

/**
 * Builds badge colors for an effort vocabulary. The default vocabulary is
 * returned untouched; anything else gets a generated ramp.
 */
export function buildEffortBadges(
  efforts: string[],
  theme: "light" | "dark",
): EffortBadges {
  if (efforts.length === 0 || isDefaultVocabulary(efforts)) {
    return DEFAULT_BADGES;
  }

  const ramp = new Palette(ANCHORS.map((hex) => kleur.hex(hex)))
    .interpolate(efforts.length)
    .toArray();

  const styles: Record<string, CSSProperties> = {};
  efforts.forEach((value, i) => {
    styles[value] = badgeStyle(ramp[i].toHex(), theme);
  });
  return { colors: {}, styles };
}

/** Badge colors for the current vocabulary and theme. */
export function useEffortBadges(efforts: string[]): EffortBadges {
  const theme = useTheme().theme;
  return useMemo(() => buildEffortBadges(efforts, theme), [efforts, theme]);
}
