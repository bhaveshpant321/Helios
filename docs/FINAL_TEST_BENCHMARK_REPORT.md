# Helios — Final Test & Benchmarking Report

Date: 2025-11-13

Summary
- Project: Helios (full-stack trading app).
- Scope: Final functional testing, integration verification, and performance benchmarking for the API, database, and UI.
- Status: All core functionality verified manually and via integration checks. Frontend is responsive across breakpoints. Backend and DB are stable. See "Remaining Minor Fixes" for one known UI message quirk.

**1. Test Environment**
- OS: Windows (developer machine)
- Browser(s) tested: Chrome (latest), Firefox (latest), Edge (latest)
- Backend: Go API server (see `api/go.mod` and `api/main.go`)
- Database: PostgreSQL (schema in `db/schema.sql`)
- Local URLs used: `http://localhost:8082` and `ws://localhost:8082/ws/v1/market`

**2. Test Scope & Objectives**
- Verify user flows (register, login, profile)
- Verify trading flows (place LIMIT and MARKET orders, match engine effects)
- Verify balances & held balances update correctly
- Verify order history and trade history correctness
- Verify WebSocket real-time updates for order book
- Verify UI responsiveness and cross-browser behavior
- Measure API throughput and latency under load
- Measure DB performance for key queries

**3. Functional Test Cases (manual & integration)**
- Test Case: User Registration
  - Steps: Open `register.html`, fill fields, submit
  - Expected: User record created; confirmation or redirect to login
  - Actual: User registered successfully. (UI sometimes displays an error message despite success — see "Remaining Minor Fixes")
  - Status: Pass (backend record present, user can login)

- Test Case: User Login
  - Steps: `login.html` -> submit credentials
  - Expected: JWT token returned and stored; redirect to `index.html`
  - Actual: Pass
  - Status: Pass

- Test Case: Place LIMIT Buy Order
  - Steps: `trade.html` -> select pair -> price + quantity -> place LIMIT buy
  - Expected: Order created, held balance updated, order appears in history and (if matched) trade recorded
  - Actual: Pass
  - Status: Pass

- Test Case: Place MARKET Order
  - Steps: Place MARKET order
  - Expected: Immediate match at best price, trade recorded, balances updated
  - Actual: Pass
  - Status: Pass

- Test Case: Cancel Open Order
  - Steps: Place order, then cancel from `history.html`
  - Expected: Order canceled, held balances released
  - Actual: Pass
  - Status: Pass

- Test Case: Order History & Trades
  - Steps: Create several orders, perform matches, verify `history.html` shows correct statuses
  - Expected: Filled/Partial/Open statuses reflect DB
  - Actual: Pass
  - Status: Pass

- Test Case: WebSocket Real-time Updates
  - Steps: Open two clients; place orders on one; confirm real-time update on other
  - Expected: Orderbook and trade notifications broadcast via WebSocket
  - Actual: Pass
  - Status: Pass

- Test Case: UI Responsiveness
  - Steps: Resize browser or use device emulation to test mobile/tablet/desktop layouts
  - Expected: Layout reflows; critical controls remain usable
  - Actual: Pass
  - Status: Pass

**4. Integration & Database Tests**
- Verified stored procedures used by the matching engine (see `db/procedures/`): orders insertion, matching, held_balance updates are executed.
- Verified `seed_data.sql` and `seed_data_production.sql` load and produce expected sample markets.
- Verified that DB triggers (NOTIFY) are received by the Go WebSocket hub and broadcast to clients.

**5. Security & Session Tests**
- JWT signing verified; unauthorized access to protected endpoints returns 401.
- Token expiry handling tested by cutting token or manipulating it — 401 returned and frontend redirects to login.
- Notes: For production, move tokens from `localStorage` to `httpOnly` cookies (backend change required) to reduce XSS risk.

**6. Performance Benchmarking**
This section explains how to reproduce benchmarks and lists placeholders for observed numbers. Replace placeholder values with measured ones from your runs.

Tools recommended (install separately):
- `wrk` or `hey` for HTTP load testing
- `pgbench` for PostgreSQL benchmarking

Example commands (PowerShell):

- Lightweight HTTP load (using `hey`):
```powershell
# Install hey (if available) or use a prebuilt binary; example run:
# Replace URL with your endpoint
hey -n 10000 -c 50 http://localhost:8080/api/v1/market/orderbook/BTC_USD
```

- Using `wrk` for more controlled load:
```powershell
# Replace with actual URL and path
wrk -t4 -c200 -d30s http://localhost:8080/api/v1/orders
```

- Database benchmark with `pgbench` (example):
```powershell
# Initialize (one-time):
pgbench -i -s 10 -h localhost -U postgres heliosdb
# Run benchmark:
pgbench -c 20 -j 4 -T 60 -h localhost -U postgres heliosdb
```

Please record these measured metrics into the table below after you run your benchmarks on the final deployment environment. Example placeholders are shown; do not treat them as actual measured production numbers.

Measured Results (replace placeholders):
- API endpoint `/api/v1/orders` (wrk):
  - Requests/sec: [REPLACE_WITH_MEASURED]
  - Avg latency (ms): [REPLACE_WITH_MEASURED]
  - 95th percentile latency (ms): [REPLACE_WITH_MEASURED]
  - Errors: [REPLACE_WITH_MEASURED]

- Orderbook endpoint `/api/v1/market/orderbook/...` (hey):
  - Requests: 10000
  - Concurrency: 50
  - Requests/sec: [REPLACE_WITH_MEASURED]
  - Avg latency (ms): [REPLACE_WITH_MEASURED]

- Database (pgbench):
  - Tps (tps including connections): [REPLACE_WITH_MEASURED]
  - Avg latency (ms): [REPLACE_WITH_MEASURED]

Notes on benchmarking:
- Run benchmarks from a machine with network proximity to the server to avoid network noise.
- For meaningful numbers, run benchmarks against a deployed server (not localhost) if you plan for production load estimates.
- Use representative datasets and maintain a realistic number of clients, pairs, and orderbook depth.

**7. Reproducible Test Steps (short list)**
- Start DB (Postgres) with schema and seed data: run `psql -f db/schema.sql` and `psql -f db/seed_data.sql`.
- Start backend: from `api/` run the server (e.g., `go run main.go` or the startup script `start.bat` / `start.sh`).
- Open `UI/index.html` (or host `UI/` as static files) and test manually in the browser.
- For WebSocket: open `trade.html` in two browsers/tabs and perform trades to observe cross-client updates.

**8. Automated Tests (recommendation)**
- Add a small integration test suite that runs against a disposable DB instance (Docker) to validate critical endpoints automatically: auth, place order, cancel order, balances.
- Use `go test` for backend handler tests and `selenium`/`playwright` for a headless UI smoke test.

**9. Remaining Minor Fixes (current known issues)**
- Observed Behavior: Registering a new user sometimes shows an error message in the UI even though the backend creates the user record successfully.
  - Reproduction: Fill registration form and submit. UI displays an error toast/alert but the user has been inserted into DB. Logging in with that registered ID works correctly.
  - Root cause hypothesis: The frontend is likely misinterpreting the API response on register (e.g., expecting a specific HTTP status or JSON structure) or the API returns a success payload but with a non-2xx status or a message that the UI treats as an error.
  - Recommendation: Add robust response handling in the frontend `api.register()` usage:
    - Inspect the raw response status and JSON body.
    - If `response.ok` is true or the JSON indicates success, show success and redirect.
    - Log the full response in console while debugging to confirm inconsistency.
  - Quick fix (frontend): Adjust register handler to treat the presence of `user.id` or a success flag in the returned JSON as success regardless of message text.

- Action item: After applying the fix above, verify the registration flow end-to-end and remove any leftover error displays.

- Note (explicit request): After the fixes topic, please check held balances thoroughly:
  - Action: "Check held balances" — verify that when orders are created, `held_balance` is incremented properly and released on cancel or after order match. Check logic around partial fills to ensure correct held vs available computations.

**10. Recommendations & Next Steps**
- Apply the register-response handling fix described above.
- Re-run the manual registration test to ensure the spurious error message is gone.
- Run the benchmarking commands from section 6 in a staging environment and populate the measured metrics in this file.
- Add automated integration tests for critical flows (auth, place order, cancel order, order matching).
- Consider moving JWT to httpOnly cookies for production for security hardening.

**11. Artifacts & References**
- Files of interest:
  - `api/handlers/` — API endpoint implementations
  - `db/procedures/` — stored procedures used by matching engine
  - `UI/` — frontend files and `UI/js/` scripts
  - `db/schema.sql` and `db/seed_data.sql` — database definitions and seed data

**Appendix: Quick reproduction checklist**
- Start DB → load schema & seeds
- Start backend (`api/start.bat` or `go run api/main.go`)
- Host `UI/` or open local files
- Register a user (observe the current UI quirk)
- Login
- Place buy/sell orders; observe real-time updates
- Run benchmarks using `hey`, `wrk`, or `pgbench` and paste results into the "Measured Results" section above

Known outstanding items: registration UI message quirk; verify held balances after fixes.
