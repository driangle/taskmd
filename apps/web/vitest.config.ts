import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test-setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "json-summary", "lcov"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/test-setup.ts", "src/**/*.test.{ts,tsx}", "src/main.tsx", "src/test-utils/**"],
      // Recalibrated for vitest 4: its AST-aware v8 remapping counts the same
      // covered code differently than v3's line-based mapping, so these numbers
      // moved without any test changing. Measured actuals are ~0.5pt above each.
      thresholds: {
        lines: 92,
        branches: 85,
        functions: 93,
        statements: 90,
      },
    },
  },
});
