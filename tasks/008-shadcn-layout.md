---
id: "008"
title: "Install shadcn/ui components + layout shell"
status: pending
priority: medium
effort: medium
dependencies:
  - "001"
tags:
  - ui
  - layout
created: 2026-02-08
---

# Install shadcn/ui Components + Layout Shell

## Objective

Set up shadcn/ui in the project, install the required component primitives, and build the application layout shell (sidebar + header + main content area).

## Tasks

- [ ] Initialize shadcn/ui with `npx shadcn-ui@latest init`
  - Choose the default style and colors
  - Ensure it configures `tailwind.config.ts` and `globals.css` correctly
- [ ] Install required shadcn/ui components:
  - `button`, `input`, `label`
  - `select` (for dropdowns)
  - `dialog` (for modals)
  - `badge` (for status/priority)
  - `table` (for data table base)
  - `dropdown-menu` (for actions)
  - `separator`
  - `sidebar` (if available, or build a simple one)
  - `skeleton` (for loading states)
  - `toast` or `sonner` (for notifications)
- [ ] Create `src/components/layout/app-sidebar.tsx`
  - Sidebar with app title/logo area
  - Slot for project switcher
  - Navigation items (just "Tasks" for MVP)
- [ ] Create `src/components/layout/header.tsx`
  - Top bar with page title
  - Slot for action buttons
- [ ] Update `src/app/layout.tsx` to use the layout shell
  - Sidebar + main content area layout
  - Wrap with any required providers (e.g., SWR config, theme)
- [ ] Update `src/app/page.tsx` to render inside the layout

## Acceptance Criteria

- shadcn/ui is initialized and components are importable
- Layout renders with sidebar and main content area
- Layout is responsive (sidebar collapses on mobile or uses sheet)
- All installed components render without errors
- The app looks polished with consistent spacing and typography
