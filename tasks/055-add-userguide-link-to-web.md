---
id: "055"
title: "Add link to user guides on GitHub from web UI"
status: pending
priority: low
effort: small
dependencies: []
tags:
  - web
  - documentation
  - user-experience
  - enhancement
  - mvp
created: 2026-02-12
---

# Add Link to User Guides on GitHub from Web UI

## Objective

Add a visible link in the taskmd web interface that directs users to the user guides and documentation on GitHub, making it easy for users to access help and documentation while using the web UI.

## Context

Users working in the web interface may need help understanding features, task format, or troubleshooting issues. Currently, there's no easy way to access documentation from the web UI. Adding a prominent link to the GitHub documentation improves discoverability and user experience.

## Tasks

- [ ] Determine best location for documentation link
  - Options: Header navigation, sidebar, footer, help icon
  - Recommended: Header navigation with "Docs" or "Help" link
- [ ] Add documentation link to web UI header/navigation
  - Link text: "Documentation" or "Help" or "User Guide"
  - Target URL: GitHub repository docs or user guides
  - Opens in new tab (`target="_blank"`)
- [ ] Choose appropriate icon (optional)
  - Material Icons: `help_outline`, `menu_book`, or `description`
  - Should be consistent with other UI elements
- [ ] Ensure link is visible on all pages
  - Dashboard view
  - Graph view
  - Board view
- [ ] Test link functionality
  - Verify URL is correct
  - Verify new tab behavior
  - Test on mobile responsive view
- [ ] Update web UI if user guide location changes

## Acceptance Criteria

- Documentation link is visible in the web UI
- Link opens GitHub user guides in a new tab
- Link is accessible from all web UI pages
- Link styling is consistent with the rest of the UI
- Mobile responsive - link visible and clickable on mobile
- URL points to correct documentation location

## Implementation Notes

### Possible Link Locations

**Option 1: Header Navigation (Recommended)**
```tsx
<header>
  <nav>
    <Logo />
    <NavLinks>
      <Link to="/">Dashboard</Link>
      <Link to="/graph">Graph</Link>
      <Link to="/board">Board</Link>
      <ExternalLink
        href="https://github.com/driangle/taskmd/blob/main/README.md"
        target="_blank"
        rel="noopener noreferrer"
      >
        Documentation
      </ExternalLink>
    </NavLinks>
  </nav>
</header>
```

**Option 2: Help Icon Button**
```tsx
<IconButton
  icon={<HelpIcon />}
  onClick={() => window.open('https://github.com/driangle/taskmd/blob/main/README.md', '_blank')}
  tooltip="View Documentation"
/>
```

**Option 3: Footer Link**
```tsx
<footer>
  <Links>
    <a href="https://github.com/driangle/taskmd" target="_blank">GitHub</a>
    <a href="https://github.com/driangle/taskmd/blob/main/README.md" target="_blank">Documentation</a>
  </Links>
</footer>
```

### Target URLs

Depending on documentation structure, link to:
- Main README: `https://github.com/driangle/taskmd/blob/main/README.md`
- User guide (if separate): `https://github.com/driangle/taskmd/blob/main/docs/USER_GUIDE.md`
- Documentation site (future): `https://docs.taskmd.dev`

### Security Considerations

- Always use `rel="noopener noreferrer"` with `target="_blank"` to prevent security vulnerabilities
- Ensure HTTPS links only

## Files to Modify

Likely locations in `apps/web/`:
- `src/components/Header.tsx` or similar navigation component
- `src/components/Layout.tsx` if using a layout wrapper
- `src/App.tsx` if navigation is defined at top level
- Corresponding CSS/styling files

## Related Tasks

- Task 043: User guides and README (provides content to link to)
- Task 046: Documentation site (may change link target in future)

## Future Enhancements

- Add dropdown menu with links to multiple documentation sections
  - User Guide
  - CLI Reference
  - Task Format Specification
  - FAQ
  - GitHub Issues
- Add in-app help tooltips or modal
- Add keyboard shortcut (e.g., `?` key) to open documentation
- Add context-sensitive help (different docs based on current page)

## Success Metrics

- Users can easily find and access documentation from web UI
- Reduced confusion about web UI features
- Improved discoverability of documentation resources
