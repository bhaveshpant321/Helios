# 🚀 Helios Deployment Guide

Follow this step-by-step guide to deploy the Helios Trading Platform to the cloud for free using **Neon** (Database), **Render** (Backend), and **Vercel** (Frontend).

---

## Part 1: Database Setup (Neon.tech)

1.  **Sign Up**: Go to [Neon.tech](https://neon.tech/) and create a free account.
2.  **Create Project**: Create a new project (e.g., "helios-db").
3.  **Get Connection String**: Copy the "Connection string" from the dashboard (it looks like `postgres://alex:abcd@ep-cool-darkness-123.us-east-2.aws.neon.tech/neondb?sslmode=verify-full`).
4.  **Run Migrations**:
    -   Go to the **SQL Editor** in the Neon dashboard.
    -   Copy the contents of `db/schema.sql` and run them.
    -   Copy the contents of `db/procedures/user_auth_procs.sql`, `db/procedures/order_query_procs.sql`, and `db/procedures/matching_engine_procs.sql` and run them.
    -   (Optional) Run `db/seed_data.sql` to populate initial assets and pairs.

---

## Part 2: Backend Deployment (Render.com)

1.  **Sign Up**: Go to [Render.com](https://render.com/) and connect your GitHub account.
2.  **New Web Service**:
    -   Click **New +** -> **Web Service**.
    -   Select your `Helios` repository.
3.  **Configure Service**:
    -   **Name**: `helios-api`.
    -   **Root Directory**: `api` (IMPORTANT).
    -   **Runtime**: `Docker`.
4.  **Environment Variables**:
    -   Click **Advanced** -> **Add Environment Variable**.
    -   `DATABASE_URL`: (Paste your Neon connection string here).
    -   `ENV`: `production`.
    -   `JWT_SECRET`: (Generate a long random string).
    -   `DB_SSLMODE`: `verify-full` (Required for Neon).
    -   `CORS_ALLOWED_ORIGINS`: `https://helios-frontend.vercel.app` (Update this AFTER you deploy the frontend).
5.  **Deploy**: Click **Create Web Service**. Render will build and deploy your Go API.
6.  **Copy API URL**: Once deployed, copy the URL (e.g., `https://helios-api.onrender.com`).

---

## Part 3: Frontend Deployment (Vercel)

1.  **Sign Up**: Go to [Vercel.com](https://vercel.com/) and connect your GitHub account.
2.  **New Project**:
    -   Click **Add New** -> **Project**.
    -   Import your `Helios` repository.
3.  **Configure Project**:
    -   **Framework Preset**: Other (or Plain HTML).
    -   **Root Directory**: `UI` (IMPORTANT).
4.  **Inject API URL**:
    -   Since we are using vanilla JS, we'll inject the API URL in the `index.html` (or a global script).
    -   **Quick Fix**: In `UI/js/config.js`, you can now hardcode the production URL you got from Render, or add this snippet to your `index.html` head:
    ```html
    <script>
      window.__HELIOS_CONFIG__ = {
        API_BASE_URL: 'https://helios-api.onrender.com/api/v1',
        WS_URL: 'wss://helios-api.onrender.com/ws/v1/market'
      };
    </script>
    ```
5.  **Deploy**: Click **Deploy**.
6.  **Update CORS**: Go back to Render and update `CORS_ALLOWED_ORIGINS` with your new Vercel URL.

---

## ✅ Post-Deployment Check

1.  **Health Check**: Visit `https://your-api.onrender.com/health`. It should return `{"status":"healthy"}`.
2.  **WebSocket Test**: Open the Trade page on your Vercel site. Check the browser console to ensure the WebSocket connects successfully (`wss://...`).
3.  **Database**: Register a new user and verify they appear in your Neon database dashboard.

---
**Congratulations! Your Helios platform is now live! 🚀**
