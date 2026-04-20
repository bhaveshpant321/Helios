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
    -   `DB_SSLMODE`: `require` (Required for Neon).
    -   `CORS_ALLOWED_ORIGINS`: `https://helios-trading.vercel.app` (Match your Vercel URL exactly).
5.  **Deploy**: Click **Create Web Service**. Render will build and deploy your Go API.
6.  **Copy API URL**: Once deployed, copy the URL (e.g., `https://helios-api-ax7p.onrender.com`).

---

## Part 3: Frontend Deployment (Vercel)

1.  **Sign Up**: Go to [Vercel.com](https://vercel.com/) and connect your GitHub account.
2.  **New Project**:
    -   Click **Add New** -> **Project**.
    -   Import your `Helios` repository.
3.  **Configure Project**:
    -   **Framework Preset**: Other (or Plain HTML).
    -   **Root Directory**: `UI` (IMPORTANT).
4.  **Deploy**: Click **Deploy**.

---

## ✅ Post-Deployment Check (Troubleshooting CORS)

If you see "Failed to load markets" in the UI:
1.  **Check Origin**: Ensure `CORS_ALLOWED_ORIGINS` in Render exactly matches your Vercel URL (including `https://`, no trailing slash).
2.  **Health Check**: Visit `https://your-api.onrender.com/`. You should see `Welcome to Helios API`.

---
**Congratulations! Your Helios platform is now live! 🚀**
