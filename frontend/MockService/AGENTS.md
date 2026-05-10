# Repository Guidelines

## Project Structure & Module Organization
This package is a React + TypeScript frontend built with Vite.
- `src/`: application code (`main.tsx`, `App.tsx`, component styles in `App.css`).
- `src/assets/`: static assets imported by code.
- `public/`: public static files served as-is (for example `favicon.svg`).
- `dist/`: production build output (generated; do not edit).
- Root config: `vite.config.ts`, `eslint.config.js`, `tsconfig*.json`, `package.json`.

## Build, Test, and Development Commands
Run commands from `/frontend/MockService`:
- `npm install`: install dependencies.
- `npm run dev`: start local dev server on `0.0.0.0:5173` with HMR.
- `npm run build`: run TypeScript project build (`tsc -b`) then create Vite production bundle.
- `npm run preview`: serve the built app locally for verification.
- `npm run lint`: run ESLint across the project.

## Coding Style & Naming Conventions
- Language: TypeScript + TSX with ES modules.
- Indentation: 2 spaces; keep semicolon and quote style consistent with surrounding files.
- Components: PascalCase file/component names (for example `UserPanel.tsx`).
- Functions/variables: camelCase; constants: UPPER_SNAKE_CASE only when truly constant.
- Keep feature assets near usage in `src/assets/` when practical.
- Linting rules come from `@eslint/js`, `typescript-eslint`, `react-hooks`, and `react-refresh` via `eslint.config.js`.

## Testing Guidelines
No test runner is currently configured in this package (`npm test` is not defined).
- Minimum check before PR: `npm run lint` and `npm run build`.
- When adding tests, prefer colocated `*.test.ts` / `*.test.tsx` files under `src/` and add a `test` script in `package.json`.

## Commit & Pull Request Guidelines
Recent history uses short, imperative messages (for example `optimize transport function`, `unit test for raft`).
- Commit format: `<scope> <action>` or concise imperative summary.
- Keep commits focused; avoid bundling unrelated refactors.
- PRs should include: purpose, key changes, validation steps run, and screenshots/GIFs for UI changes.
- Link related issues/tasks and note any follow-up work.

## Security & Configuration Tips
- Do not commit secrets; use environment variables for runtime configuration.
- Treat `dist/` as generated output; regenerate instead of editing manually.
