---
id: "web-025"
title: "Add ESLint configuration with file length limits"
status: completed
priority: medium
effort: small
dependencies: []
tags:
  - code-quality
  - linting
  - tooling
created: 2026-02-08
---

# Add ESLint Configuration with File Length Limits

## Objective

Set up ESLint for the web project with custom rules to enforce code quality standards, including a maximum file length of 200 lines to promote modular, focused components.

## Tasks

- [ ] Install ESLint and required plugins:
  - `eslint`
  - `@typescript-eslint/eslint-plugin`
  - `@typescript-eslint/parser`
  - `eslint-plugin-react`
  - `eslint-plugin-react-hooks`
- [ ] Create `.eslintrc.json` configuration file
- [ ] Configure ESLint rules:
  - Enable recommended TypeScript rules
  - Enable React and React Hooks rules
  - Add `max-lines` rule: `{ "max": 200, "skipBlankLines": true, "skipComments": true }`
- [ ] Add ESLint scripts to `package.json`:
  - `"lint": "eslint . --ext .ts,.tsx"`
  - `"lint:fix": "eslint . --ext .ts,.tsx --fix"`
- [ ] Create `.eslintignore` file to exclude:
  - `node_modules/`
  - `dist/`
  - `.vite/`
  - `coverage/`
- [ ] Run linter on existing codebase and fix any violations
- [ ] Update documentation if needed

## Acceptance Criteria

- ESLint runs successfully with `pnpm lint`
- The `max-lines` rule is enforced at 200 lines per file
- All existing files pass linting or have been refactored to comply
- ESLint auto-fix works with `pnpm lint:fix`
- Configuration follows TypeScript and React best practices

## Notes

- The 200-line limit aligns with project guidelines for keeping files focused and manageable
- `skipBlankLines` and `skipComments` options ensure only actual code is counted
- Consider adding pre-commit hooks in a future task to enforce linting automatically
