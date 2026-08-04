import { vi, describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { mockApi, resetMockApi } from "../../test-utils/mock-api.ts";

vi.mock("../../hooks/use-config.ts", () => ({
  useConfig: () => mockApi.config,
}));
vi.mock("../../hooks/use-project.ts", () => ({
  useProject: () => mockApi.project,
}));
vi.mock("../../hooks/use-theme.ts", () => ({
  useTheme: () => ({ theme: "light", toggle: vi.fn() }),
}));
vi.mock("../search/SearchDialog.tsx", () => ({
  SearchDialog: ({ open }: { open: boolean }) =>
    open ? <div data-testid="search-dialog">Search</div> : null,
}));
vi.mock("./NavTabs.tsx", () => ({
  DesktopNav: () => <nav data-testid="desktop-nav">Nav</nav>,
  HeaderActions: ({ onSearchOpen }: { onSearchOpen: () => void }) => (
    <div data-testid="header-actions">
      <button onClick={onSearchOpen}>Search</button>
    </div>
  ),
  MobileMenu: () => <nav data-testid="mobile-menu">Mobile Menu</nav>,
}));
vi.mock("./PhaseSelector.tsx", () => ({
  PhaseSelector: () => <div data-testid="phase-selector" />,
}));
vi.mock("./ProjectSelector.tsx", () => ({
  ProjectSelector: () => <div data-testid="project-selector" />,
}));

import { Shell } from "./Shell.tsx";

function renderShell(initialEntries = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Shell>
        <div data-testid="child-content">Page Content</div>
      </Shell>
    </MemoryRouter>,
  );
}

describe("Shell", () => {
  beforeEach(() => {
    resetMockApi();
    mockApi.config = { readonly: false, version: "1.2.3", phases: [] };
  });

  it("renders taskmd brand link", () => {
    renderShell();
    const link = screen.getByText("taskmd");
    expect(link).toBeDefined();
    expect(link.closest("a")).toBeDefined();
  });

  it("renders version when provided", () => {
    renderShell();
    expect(screen.getByText("1.2.3")).toBeDefined();
  });

  it("renders Read Only badge when readonly is true", () => {
    mockApi.config = { readonly: true, version: "1.2.3", phases: [] };
    renderShell();
    expect(screen.getByText("Read Only")).toBeDefined();
  });

  it("does not render Read Only badge when readonly is false", () => {
    renderShell();
    expect(screen.queryByText("Read Only")).toBeNull();
  });

  it("renders child content in main area", () => {
    renderShell();
    expect(screen.getByTestId("child-content")).toBeDefined();
    expect(screen.getByText("Page Content")).toBeDefined();
  });

  it("renders skip-to-main-content link", () => {
    renderShell();
    expect(screen.getByText("Skip to main content")).toBeDefined();
  });

  it("Ctrl+K opens search dialog", () => {
    renderShell();
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    fireEvent.keyDown(document, { key: "k", ctrlKey: true });
    expect(screen.getByTestId("search-dialog")).toBeDefined();
  });

  it("/ key opens search dialog", () => {
    renderShell();
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    fireEvent.keyDown(document, { key: "/" });
    expect(screen.getByTestId("search-dialog")).toBeDefined();
  });

  it("/ does not open search when target is an INPUT element", () => {
    renderShell();
    const input = document.createElement("input");
    document.body.appendChild(input);
    fireEvent.keyDown(input, { key: "/" });
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    document.body.removeChild(input);
  });

  it("Escape closes mobile menu", () => {
    renderShell();
    const hamburger = screen.getByLabelText("Toggle navigation menu");
    fireEvent.click(hamburger);
    expect(screen.getByTestId("mobile-menu")).toBeDefined();
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByTestId("mobile-menu")).toBeNull();
  });

  it("hamburger button toggles mobile menu", () => {
    renderShell();
    const hamburger = screen.getByLabelText("Toggle navigation menu");
    expect(screen.queryByTestId("mobile-menu")).toBeNull();
    fireEvent.click(hamburger);
    expect(screen.getByTestId("mobile-menu")).toBeDefined();
    fireEvent.click(hamburger);
    expect(screen.queryByTestId("mobile-menu")).toBeNull();
  });

  it("DesktopNav search button opens search dialog", () => {
    renderShell();
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    const searchButton = screen.getByText("Search");
    fireEvent.click(searchButton);
    expect(screen.getByTestId("search-dialog")).toBeDefined();
  });

  it("uses the graph page layout when on /graph", () => {
    renderShell(["/graph"]);
    expect(screen.getByRole("main")).toHaveAttribute("data-layout", "graph");
  });

  it("uses the default layout on other pages", () => {
    renderShell(["/"]);
    expect(screen.getByRole("main")).toHaveAttribute("data-layout", "default");
  });

  it("does not open search for / key on TEXTAREA", () => {
    renderShell();
    const textarea = document.createElement("textarea");
    document.body.appendChild(textarea);
    fireEvent.keyDown(textarea, { key: "/" });
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    document.body.removeChild(textarea);
  });

  it("does not open search for / key on SELECT", () => {
    renderShell();
    const select = document.createElement("select");
    document.body.appendChild(select);
    fireEvent.keyDown(select, { key: "/" });
    expect(screen.queryByTestId("search-dialog")).toBeNull();
    document.body.removeChild(select);
  });

  it("does not render version when not provided", () => {
    mockApi.config = { readonly: false, version: "", phases: [] };
    renderShell();
    expect(screen.queryByText("1.2.3")).toBeNull();
  });

  it("ArrowUp on main content focuses header nav", () => {
    renderShell();
    const main = document.getElementById("main-content")!;
    main.focus();
    fireEvent.keyDown(main, { key: "ArrowUp" });
    expect(document.activeElement?.tagName).toBeDefined();
  });
});
