# Testing And Quality Dependency - User App Frontend

## 1. What Is This Dependency?

Testing and quality tooling code ko verify karta hai.

This service uses:

- Vitest for unit tests.
- jsdom for browser-like test environment.
- Testing Library for React component tests.
- ESLint and TypeScript ESLint for static checks.
- TypeScript build for type safety.

## 2. Why This Service Uses It

User App me auth, checkout, cart, wishlist, gRPC, schema validation, and UI components hain. In flows me regressions costly ho sakte hain, so tests and linting important hain.

## 3. Required Or Optional

Required for development and CI. Production runtime ke liye optional because test tools bundle me nahi jaate.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `frontend/user-app/vitest.config.ts` | Vitest config. |
| `frontend/user-app/eslint.config.js` | ESLint config. |
| `frontend/user-app/package.json` | `test`, `lint`, `typecheck`, `build` scripts. |
| `frontend/user-app/src/**/*.test.ts` | Unit tests. |
| `frontend/user-app/src/**/*.test.tsx` | Component tests. |
| `frontend/user-app/src/test/create-test-query-wrapper.tsx` | React Query test wrapper. |
| `frontend/user-app/src/test/grpc/mock-transport.ts` | gRPC test transport. |

## 5. Installation Steps

Current dev dependencies already exist.

```bash
cd frontend
pnpm --filter user-app add -D vitest jsdom @testing-library/react @testing-library/user-event
pnpm --filter user-app add -D eslint typescript-eslint eslint-plugin-react-hooks eslint-plugin-react-refresh
```

Suggested command based on project structure.

## 6. Docker Setup, If Possible

CI Docker build should run:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

Actual CI/Docker config was not clearly found in project files.

## 7. Local Setup Without Docker

```bash
cd frontend
pnpm install
```

Run quality checks:

```bash
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

## 8. Required Environment Variables

Tests generally do not require real backend env because they should mock APIs/transports.

Build/typecheck uses defaults from `src/lib/env.ts` if no env is set.

## 9. Start Commands

Testing:

```bash
cd frontend
pnpm --filter user-app test
```

Lint:

```bash
cd frontend
pnpm --filter user-app lint
```

Typecheck:

```bash
cd frontend
pnpm --filter user-app typecheck
```

## 10. Verify Running Commands

Full local verification:

```bash
cd frontend
pnpm --filter user-app typecheck
pnpm --filter user-app lint
pnpm --filter user-app test
pnpm --filter user-app build
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| jsdom APIs missing | Test environment misconfigured | Keep `environment: 'jsdom'` in `vitest.config.ts`. |
| React Query tests share cache | QueryClient reused | Use test wrapper/new client per test. |
| gRPC tests hit real backend | Transport not mocked | Use `src/test/grpc/mock-transport.ts`. |
| ESLint type-aware config slow/failing | TS project service issue | Run from `frontend/user-app` or ensure tsconfig exists. |
| Build passes but tests fail | Mock data out of sync | Align fixtures with API/proto types. |

## 12. Security Notes

- Tests should not require real secrets.
- Avoid committing recorded real API responses with PII.
- Add tests for auth redirects, token expiry behavior, checkout idempotency, and payment failure states.
- CI should fail on lint/typecheck/test failures.

## 13. Final Checklist

| Item | Status |
|------|--------|
| Vitest config found | Completed |
| ESLint config found | Completed |
| Test scripts found | Completed |
| Existing tests found | Completed |
| CI config found | Missing |
