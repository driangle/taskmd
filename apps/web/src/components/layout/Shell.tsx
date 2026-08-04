import { useState, useEffect, useCallback } from "react";
import type { ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { useConfig } from "../../hooks/use-config.ts";
import { useProject } from "../../hooks/use-project.ts";
import { useTheme } from "../../hooks/use-theme.ts";
import { SearchDialog } from "../search/SearchDialog.tsx";
import { DesktopNav, HeaderActions, MobileMenu } from "./NavTabs.tsx";
import { PhaseSelector } from "./PhaseSelector.tsx";
import { ProjectSelector } from "./ProjectSelector.tsx";

interface ShellProps {
  children: ReactNode;
}

export function Shell({ children }: ShellProps) {
  const { project } = useProject();
  const { readonly, version } = useConfig(project);
  const { theme, toggle } = useTheme();
  const [menuOpen, setMenuOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const location = useLocation();
  const isGraphPage = location.pathname === "/graph";

  // Close mobile menu on route change
  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { setMenuOpen(false); }, [location.pathname]);

  // Global keyboard shortcuts
  useEffect(() => {
    const FOCUSABLE = 'a, button, [tabindex="0"], input, select, textarea';

    function handleKeyDown(e: KeyboardEvent) {
      // Cmd+K / Ctrl+K - open search
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
        return;
      }

      // "/" key - open search (only when not in input/textarea)
      if (
        e.key === "/" &&
        !e.metaKey &&
        !e.ctrlKey &&
        !e.altKey
      ) {
        const tag = (e.target as HTMLElement)?.tagName;
        if (tag !== "INPUT" && tag !== "TEXTAREA" && tag !== "SELECT") {
          e.preventDefault();
          setSearchOpen(true);
        }
      }

      // Escape - close mobile menu
      if (e.key === "Escape" && menuOpen) {
        setMenuOpen(false);
      }

      // Skip arrow-nav handling for text inputs
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

      // ArrowLeft/Right — navigate within any [data-arrow-nav] container
      if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
        const container = (e.target as HTMLElement).closest?.("[data-arrow-nav]");
        if (!container) return;
        const items = Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE));
        const idx = items.indexOf(e.target as HTMLElement);
        if (idx < 0) return;
        e.preventDefault();
        const next = e.key === "ArrowRight"
          ? items[(idx + 1) % items.length]
          : items[(idx - 1 + items.length) % items.length];
        next.focus();
        return;
      }

      // ArrowUp/Down — move between sibling [data-arrow-nav] rows, or to next/prev focusable
      if (e.key === "ArrowDown" || e.key === "ArrowUp") {
        const container = (e.target as HTMLElement).closest?.("[data-arrow-nav]");
        if (container) {
          // Try sibling [data-arrow-nav] containers first
          const siblings = Array.from(
            (container.parentElement?.querySelectorAll(":scope > [data-arrow-nav]") ?? []),
          ) as HTMLElement[];
          if (siblings.length > 1) {
            const colIdx = siblings.indexOf(container as HTMLElement);
            const items = Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE));
            const posIdx = items.indexOf(e.target as HTMLElement);
            const nextRowIdx = e.key === "ArrowDown"
              ? (colIdx + 1) % siblings.length
              : (colIdx - 1 + siblings.length) % siblings.length;
            const nextRow = siblings[nextRowIdx];
            const nextItems = Array.from(nextRow.querySelectorAll<HTMLElement>(FOCUSABLE));
            if (nextItems.length > 0) {
              e.preventDefault();
              nextItems[Math.min(posIdx, nextItems.length - 1)].focus();
              return;
            }
          }

          // No sibling row — jump to next/prev focusable element outside the container
          const all = Array.from(document.querySelectorAll<HTMLElement>(FOCUSABLE));
          const containerItems = new Set(container.querySelectorAll<HTMLElement>(FOCUSABLE));
          if (e.key === "ArrowDown") {
            const lastInContainer = all.findIndex((el) => containerItems.has(el)) + containerItems.size;
            const next = all.slice(lastInContainer).find((el) => !containerItems.has(el));
            if (next) { e.preventDefault(); next.focus(); return; }
          } else {
            const firstInContainer = all.findIndex((el) => containerItems.has(el));
            for (let i = firstInContainer - 1; i >= 0; i--) {
              if (!containerItems.has(all[i])) { e.preventDefault(); all[i].focus(); return; }
            }
          }
        }

        // ArrowDown from header — jump to main content
        if (e.key === "ArrowDown" && (e.target as HTMLElement).closest?.("header")) {
          e.preventDefault();
          const main = document.getElementById("main-content");
          if (!main) return;
          const focusable = main.querySelector<HTMLElement>(FOCUSABLE);
          (focusable ?? main).focus();
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [menuOpen]);

  const closeSearch = useCallback(() => setSearchOpen(false), []);

  return (
    <div className={`overflow-x-hidden bg-gray-50 text-gray-900 dark:bg-gray-900 dark:text-gray-100 ${isGraphPage ? "h-screen flex flex-col overflow-hidden" : "min-h-screen"}`}>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-[100] focus:px-4 focus:py-2 focus:bg-blue-600 focus:text-white focus:rounded"
      >
        Skip to main content
      </a>
      <header className="bg-white border-b border-gray-200 dark:bg-gray-800 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          {/* Row 1: Branding + selectors + utilities */}
          <div className="flex items-center justify-between h-12">
            <div className="flex items-center gap-2">
              <Link
                to="/tasks"
                className="text-lg font-semibold tracking-tight"
              >
                taskmd
              </Link>
              {version && (
                <span className="text-xs text-gray-400 dark:text-gray-500">
                  {version}
                </span>
              )}
              {readonly && (
                <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
                  Read Only
                </span>
              )}
              <ProjectSelector />
              <PhaseSelector />
            </div>
            <div className="flex items-center gap-1">
              <HeaderActions onSearchOpen={() => setSearchOpen(true)} />
              {/* Theme toggle - always visible */}
              <button
                onClick={toggle}
                className="ml-1 p-2.5 sm:p-1.5 min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 inline-flex items-center justify-center rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700 transition-colors"
                aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
              >
                {theme === "dark" ? (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                ) : (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                  </svg>
                )}
              </button>
              {/* Hamburger button - mobile only */}
              <button
                onClick={() => setMenuOpen((o) => !o)}
                className="md:hidden ml-1 p-2.5 sm:p-1.5 min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 inline-flex items-center justify-center rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700 transition-colors"
                aria-label="Toggle navigation menu"
              >
                {menuOpen ? (
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
                  </svg>
                )}
              </button>
            </div>
          </div>
          {/* Row 2: Navigation */}
          <div className="hidden md:flex items-center h-10 mb-1">
            <DesktopNav />
          </div>
        </div>
        {menuOpen && <MobileMenu />}
      </header>
      <main
        id="main-content"
        tabIndex={-1}
        data-layout={isGraphPage ? "graph" : "default"}
        className={isGraphPage ? "px-2 py-2 flex-1 min-h-0" : "max-w-7xl mx-auto px-4 sm:px-6 py-4 md:py-6"}
        onKeyDown={(e) => {
          if (e.key === "ArrowUp" && e.target === e.currentTarget) {
            e.preventDefault();
            const nav = document.querySelector<HTMLElement>("header nav");
            const link = nav?.querySelector<HTMLElement>("a, button");
            (link ?? nav)?.focus();
          }
        }}
      >{children}</main>
      <SearchDialog open={searchOpen} onClose={closeSearch} />
    </div>
  );
}
