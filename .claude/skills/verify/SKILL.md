---
name: verify
description: Build/run/drive recipe for this repo (Go + chi + htmx server-rendered app), learned from an end-to-end CRUD pass.
---

# Verifying this app

Server-rendered Go app (chi router, htmx). No test suite exists — verification means running the binary against the real local Postgres dev DB and driving HTTP requests through it.

## Build & run

```bash
go build -buildvcs=false -o ./bin/university-agency.exe ./cmd/server
./bin/university-agency.exe > /tmp/run.log 2>&1 &
```

Reads `.env` in the repo root (already configured for local Postgres at 127.0.0.1:5432, db `university_agency`, seed data enabled since `APP_ENV=development`). No flags needed. Check `/tmp/run.log` for `"server started"` and no `error`/`panic` lines.

## Driving it (curl)

- **All mutating (POST) requests need an `Origin` header matching the host**, or they get a 403 from `originGuard` — always pass `-H "Origin: http://127.0.0.1:9001"`.
- Login: `curl -c cookies.txt -X POST http://127.0.0.1:9001/login -H "Origin: ..." -d "username=owner&password=owner12345"` (demo accounts: `owner`/`owner12345`, `staff`/`staff12345`, `bayu`/`bayu12345` — second staff, useful for PIC-gating checks —, `student`/`student12345`).
- File-upload endpoints (expense receipts, documents, payment proof) require real multipart encoding — use `-F field=value` for **every** field, not `-d`. Sending `-d` (urlencoded) makes `r.FormFile` return `http.ErrNotMultipart`, which several handlers don't excuse the way they excuse `http.ErrMissingFile` — you'll get a misleading "file tidak valid" notice that has nothing to do with the actual field values.
- New client accounts are gated behind the signed-agreement redirect: after creating a client and logging in as them, `GET /student` 302s to `/student/agreement` until you `POST /student/agreement/sign -d "agree=1&full_name=..."`.
- Extracting IDs from rendered HTML: table rows don't carry the record ID as a visible attribute — it's embedded in the action URL of the row's own inline form (e.g. `/staff/tasks/{id}/complete`, `/institutions/{id}/update`). Find the row by its unique label text first, then read the ID from the form action immediately following it in the HTML — don't grab "the last ID matching this pattern on the page," it may belong to a different, unrelated row from list ordering.
- Owner/staff CRUD list pages (institutions, templates, service packages, pipeline stages) only render their edit/delete forms with `?manage=1` in the query string — the plain URL shows a read-only view.
- Correct field/route names worth double-checking before assuming a failure is a real bug: mark-paid is `POST /owner/orders/mark-paid` with field `order_code` (not `/owner/invoices/mark-paid` / `code`).
- Grepping a whole rendered page for a status word like "lunas" is unreliable — it matches unrelated orders/filters elsewhere on the same page. Anchor on the specific status badge (`Status Pembayaran</dt><dd><span class="status ...">`) near the record you're checking.

## Known pre-existing gaps (not bugs from any specific change)

- No route creates orders/invoices — only seeded. Don't look for a "create invoice" flow.
- Most entities have no delete endpoint (tasks, expenses, documents, shipments, client accounts) — only institutions/templates/service-packages/pipeline-stages support delete. Test data created via curl during verification accumulates in the dev DB; mention it in the report rather than silently leaving it, since there's no UI path to remove it.
