# MagicStream — System Design

**Overview**
- Full‑stack web app to browse movies, write admin reviews, and deliver personalized recommendations.
- Frontend: React + TypeScript + Vite over HTTPS.
- Backend: Go + Gin over HTTPS with MongoDB persistence.
- Auth via JWT in HttpOnly, Secure cookies validated by middleware.

**Architecture**
- Frontend
  - Runs at `https://localhost:5173` with local TLS certs.
  - HTTP client: public axios instance and private axios hook for authenticated calls.
- Backend
  - Gin router over `https://localhost:8080` via `RunTLS`.
  - MongoDB: collections `users`, `movies`, `genres`, `rankings`.
  - CORS allows `https://localhost:5173` and credentials.
- External
  - Google GenAI (Gemini) classifies admin review sentiment into a ranking name.

**Data Model**
- User: `user_id`, `first_name`, `last_name`, `email`, `password` (hashed), `role`, `token`, `refresh_token`, `favourite_genres`.
- Movie: `imdb_id`, `title`, `poster_path`, `youtube_id`, `genre[]`, `admin_review`, `ranking {name, value}`.
- Genre: `genre_id`, `genre_name`.
- Ranking: `ranking_value`, `ranking_name`.

**API**
- Public: `GET /movies`, `GET /genres`, `POST /register`, `POST /login`, `POST /refresh-token`.
- Protected: `GET /movie/:imdb_id`, `POST /add-movie`, `GET /recommended-movies`, `PATCH /update-review/:imdb_id`, `POST /logout`.
- Middleware: reads `access_token` cookie, validates JWT, injects `user_id` and `role`, aborts 401 on failure.

**Authentication**
- Tokens: access (~1h) and refresh (~24h) via HMAC‑signed JWTs.
- Cookies: HttpOnly, Secure, `SameSite=None`, `Domain=localhost` for local cross‑origin.
- Refresh flow: validates `refresh_token` cookie, issues new pair, updates DB, resets cookies.
- Authorization: `role` claim gates admin endpoints.

**CORS & HTTPS**
- CORS: allow origin `https://localhost:5173`, methods `GET, POST, PATCH, PUT, DELETE, OPTIONS`, headers `Origin, Content-Type, Authorization`, `AllowCredentials=true`.
- HTTPS: required for `Secure` cookies; both frontend and backend run with TLS.

**Recommendation Flow**
- Read `user_id` from context; project `favourite_genres` (`genre_name` only).
- Filter movies by genres using `$in` on `genre.genre_name`.
- Sort by `ranking.ranking_value` ascending; limit by `RECOMMENDED_MOVIE_LIMIT` (default 5).

**Admin Review & Ranking**
- Admin submits `admin_review`.
- Build prompt from ranking names; call Gemini to infer `ranking_name`.
- Update movie with `admin_review` and ranking `{name, value}`.

**Frontend Design**
- Auth context persists user state in `localStorage`.
- HTTP clients:
  - `axiosConfig` for public endpoints (HTTPS + credentials).
  - `useAxiosPrivate` interceptor for protected endpoints (`withCredentials`, `baseURL`).
- Screens: Home, Login, Register, Recommended, Review, Header.

**Error Handling**
- Backend: JSON error responses, 401/403/500 status codes.
- Frontend: user‑friendly messages, logs, and clears user on 401 when needed.
- Dev pitfalls avoided: mixed content and missing `withCredentials` in cross‑origin cookies.

**Environment & Configuration**
- Backend `.env`: `MONGODB_URI`, `DATABASE_NAME`, `ALLOWED_ORIGINS`, `SECRET_KEY`, `SECRET_REFRESH_KEY`, `GEMINI_API_KEY`, `BASE_PROMPT_TEMPLATE`, `RECOMMENDED_MOVIE_LIMIT`.
- Frontend `.env`: `VITE_API_BASE_URL` or `VITE_API_URL` (e.g., `https://localhost:8080`).
- TLS: local certs used by Vite and Gin; paths configured in project.

**Sequence Flows**
- Login: verify credentials → set cookies → update frontend auth.
- Logout: call with credentials → validate cookie → clear tokens → expire cookie → reset auth.

**Demonstrated Knowledge**
- React + TypeScript: component patterns, hooks, context, forms.
- Axios & CORS: interceptors, credentialed requests, secure cross‑origin setup.
- Cookie Security: HttpOnly, Secure, `SameSite=None`, domain scoping.
- Go + Gin: routing, middleware, context, TLS server configuration.
- JWT: claims design, signing/validating, expiry, refresh flow.
- MongoDB: modeling, filters, projection, sorting, limiting.
- External API integration: prompt engineering and response handling with Gemini.
- Configuration & Security: environment management, HTTPS alignment, error strategies.
