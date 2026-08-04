import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LoadingState } from "./LoadingState.tsx";

const variants = ["default", "table", "board", "graph", "cards", "detail"] as const;

describe("LoadingState", () => {
  it("exposes an accessible loading status when no variant provided", () => {
    render(<LoadingState />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });

  it.each(variants)("renders an accessible loading status for variant='%s'", (variant) => {
    render(<LoadingState variant={variant} />);
    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
  });
});
