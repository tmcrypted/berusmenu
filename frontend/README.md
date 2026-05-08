# frontend

Frontend workspace for the Berusmenu MVP.

## Responsibility

- public vacancy-style home with mocked shifts
- registration / login / email verification UI
- post-verify profile completion for `worker` and `owner`
- owner PVZ flow through `inventory-service`
- moderator PVZ moderation panel
- worker entry points for later domain flows
- responsive web UI for mobile, tablet, and desktop

## Current stack

- `Next.js 15`
- `React 19`
- `Tailwind CSS 4`
- `TypeScript`
- `Zustand`
- `next-themes`
- `react-hook-form`
- `zod`

## Frontend auth rules

- browser auth uses HttpOnly cookies from `api-gateway`
- auth tokens are never stored in Zustand or localStorage
- unsafe API calls send `X-CSRF-Token` from the readable `berusmenu_csrf` cookie
- first paint auth bootstrap happens on the Next.js server via request cookies and read-only `GET /v1/auth/session`
- Zustand is hydrated from that server snapshot before client components render
- client auth requests reuse the same snapshot resolver to avoid duplicated state logic

## UX rules

- registration is minimal: `email + password + role`
- unverified user is redirected to `/verify-email`
- verified but incomplete user is redirected to a role-specific completion screen
- worker cannot proceed to responses until profile completion
- owner completes profile first and then proceeds to the PVZ flow
- owner can create/list/edit PVZ through inventory endpoints
- moderator accounts are assigned outside the UI and can process pending PVZ requests, including timed owner spam block with PVZ cleanup
- both light and dark themes are first-class and switched manually
- new UI must remain usable across phones, tablets, and desktop screens
- public home can use mocked vacancies until matching endpoints are ready
- verification screen must keep logout, resend code, and verify code input
- withdraw response must keep strong confirmation UI when response flows are added
- replies under review stay collapsed by default with a "show replies" affordance
