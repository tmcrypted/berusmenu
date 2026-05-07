# frontend

Frontend workspace for Berusmenu MVP inside the monorepo.

## Ответственность

- registration / login / email verification UI
- owner PVZ cabinet
- vacancy creation and management
- worker vacancy feed
- response chat
- reviews and review replies

## Key UX rules

- auth tokens are never stored in Zustand/localStorage; browser auth uses HttpOnly cookies from `api-gateway`
- unsafe API calls send `X-CSRF-Token` from the readable `berusmenu_csrf` cookie
- unverified user is redirected to verification screen
- verification screen has:
  - logout
  - resend code
  - verify code input
- withdraw response must have strong confirmation UI
- replies under review are collapsed by default with "show replies"

## Next implementation steps

1. auth and verification screens
2. owner PVZ flow
3. owner vacancy flow
4. worker response flow
5. chat UI
6. reviews UI
