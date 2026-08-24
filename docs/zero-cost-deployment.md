# Zero-Cost & Free-Tier Deployment Guide (Railway, Vercel, Supabase, Upstash)

If you have **$0 in cloud credits** or want to host Recess permanently on **100% free platforms**, this guide walks through zero-cost setup in under 5 minutes.

---

## 1. Zero-Cost Architecture Overview

```
                                  [ Free Internet Users ]
                                             │
                       ┌─────────────────────┴─────────────────────┐
                       ▼                                           ▼
             ┌───────────────────┐                       ┌───────────────────┐
             │ Vercel / Netlify  │                       │ Railway / Render  │
             │ (React SPA Host)  │                       │ (Go Backend App)  │
             │ FREE TIER (100GB) │                       │ FREE / TRIAL TIER │
             └───────────────────┘                       └─────────┬─────────┘
                                                                   │
                                                ┌──────────────────┴──────────────────┐
                                                ▼                                     ▼
                                      ┌───────────────────┐                 ┌───────────────────┐
                                      │ Supabase / Neon   │                 │ Upstash / Redis   │
                                      │ (Free PostgreSQL) │                 │ (Free Serverless) │
                                      │ 500 MB FREE       │                 │ 10,000 req/day    │
                                      └───────────────────┘                 └───────────────────┘
```

---

## 2. Platform Breakdown & Sizing

| Component | Recommended Provider | Free Tier Limits | Cost | Configuration File |
| :--- | :--- | :--- | :--- | :--- |
| **Frontend** | [Vercel](https://vercel.com) or [Netlify](https://netlify.com) | Unlimited builds, 100 GB bandwidth, SSL, edge CDN | **$0.00** | `frontend/vercel.json` / `frontend/netlify.toml` |
| **Go Backend** | [Railway](https://railway.app) or [Render](https://render.com) | 500 hours/mo, WebSocket support, auto-SSL | **$0.00** | `railway.json` / `render.yaml` / `Procfile` |
| **PostgreSQL** | [Supabase](https://supabase.com) or [Neon](https://neon.tech) | 500 MB storage, pooling connection string | **$0.00** | Standard Postgres URL |
| **Redis Cache**| [Upstash](https://upstash.com) or [Railway Redis](https://railway.app) | 10,000 commands/day, TLS, Pub/Sub support | **$0.00** | Standard Redis URL (`redis://...` or `rediss://...`) |

---

## 3. Step-by-Step 1-Click Deployment

### Step 1: Provision Free Database & Redis
1. **PostgreSQL**:
   - Go to [Supabase](https://database.new) or [Neon](https://neon.tech) and create a free project.
   - Copy your connection string: `postgres://postgres:[PASSWORD]@[HOST]:5432/postgres?sslmode=require`.
2. **Redis**:
   - Go to [Upstash](https://upstash.com) -> **Create Database** (Select primary region).
   - Copy the Redis URL: `rediss://default:[PASSWORD]@[ENDPOINT]:6379`.

---

### Step 2: Deploy Go Backend on Railway (or Render)

#### Option A: Railway (Fastest)
1. Go to [Railway](https://railway.app/new).
2. Select **Deploy from GitHub repo** -> choose `Abh19avM/recess`.
3. Set **Root Directory** to `/` or configure via `railway.json`.
4. Add the following Environment Variables in the Railway Dashboard:
   ```env
   PORT=8080
   ENVIRONMENT=production
   LOG_LEVEL=info
   DATABASE_URL=postgres://... (from Supabase/Neon)
   REDIS_URL=rediss://... (from Upstash/Railway)
   JWT_SECRET=generate-a-32-character-random-secret
   CORS_ALLOWED_ORIGINS=https://*.vercel.app,https://*.netlify.app
   ```
5. Click **Deploy**. Railway will build the Docker container and give you a public URL (e.g. `https://recess-production.up.railway.app`).

#### Option B: Render
1. Go to [Render Dashboard](https://dashboard.render.com/blueprints).
2. Connect your repository — Render will automatically read `render.yaml`.
3. Fill in your environment variables and click **Apply Blueprint**.

---

### Step 3: Deploy Frontend on Vercel

1. Go to [Vercel](https://vercel.com/new).
2. Import the `Abh19avM/recess` repository.
3. Set **Root Directory** to `frontend`.
4. Framework Preset: **Vite**.
5. Add Environment Variables:
   ```env
   VITE_API_URL=https://recess-production.up.railway.app
   VITE_WS_URL=wss://recess-production.up.railway.app/ws
   ```
6. Click **Deploy**. Vercel will build the React SPA and output your public URL (e.g. `https://recess.vercel.app`).

---

### Step 4: Run Initial Database Migrations

Run database migrations against the Supabase/Neon PostgreSQL database from your local machine:
```bash
cd backend
DATABASE_URL="postgres://postgres:[PASSWORD]@[HOST]:5432/postgres?sslmode=require" go run ./cmd/migrate up
```

---

## 4. Verification

1. Open `https://recess.vercel.app`.
2. Click **"Quick Guest Play"** to test instant authentication.
3. Open two browser windows and start a live **Hand Cricket** or **Dots & Boxes** match to verify WebSocket synchronization over Railway.
4. Check Supabase/Neon tables (`users`, `matches`, `user_stats`) to confirm data persistence.
