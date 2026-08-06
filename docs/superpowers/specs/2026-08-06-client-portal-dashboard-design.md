# Client portal dashboard — design (Sprint 07, item 7.5)

Date: 2026-08-06
Status: approved

## Purpose

Extend the existing client approval portal (magic-link login, `ClientUser`/
`ClientSession`, mounted on the public phishing server at 9443) into an
ongoing dashboard the client can return to on their own: browse the
tenant's campaigns, see aggregate indicators, track historical evolution,
and export an authorized report — not just decide one pending approval.

Training/quiz content (Sprint07.md §7.6) is explicitly out of scope; the
nav reserves a slot for it but it renders nothing yet.

## Scope decisions (from brainstorming)

- **Data scope**: the client sees every campaign belonging to their
  `ClientUser.TenantID` — not just campaigns tied to the contract(s) they
  approve. Being a named approver on any one contract in a tenant grants
  visibility into that tenant's whole campaign program.
- **Detail level**: aggregate only, matching the "Cliente executivo"
  access profile in Sprint07.md §8.4 ("visualiza indicadores agregados;
  não visualiza dados sensíveis individuais"). No per-target rows (name,
  email, IP) are ever rendered or exported through `/portal/*`.
- **Auth model**: no password. Session lifetime, cookie mechanics and
  hashing stay identical to the existing approval portal
  (`ClientSession`, 7-day TTL, `ethphish_client` cookie, `HttpOnly` +
  `Secure` + `SameSite=Lax`).
- **Login self-service**: since the client now wants in without a pending
  approval driving a magic link, add a "request a login link" form. Given
  an e-mail, if it matches a `ContractApprover` in some tenant, a fresh,
  single-use, generically-worded link is e-mailed. The response message is
  identical whether or not the e-mail matched, to avoid confirming which
  addresses are registered approvers.
- **Reports**: reuse `reports.JobService`/`GenerateReport` as-is, scoped by
  tenant ID, same Word/Excel formats already used by the admin side. No
  new metric engine.
- **Routing**: new `/portal/*` prefix, separate from `/approvals/*`. Same
  session cookie and `requireClientSession`-style guard. Cross-links
  between the two areas for a client who already has a session.

## Data model

New table `portal_login_tokens` (one per pending self-service login
request), mirroring the hashing pattern in `models/token.go`:

| Column | Type | Notes |
| --- | --- | --- |
| `id` | bigserial PK | |
| `tenant_id` | bigint not null | resolved from the matching `ContractApprover`'s contract |
| `email` | text not null | the address the link was sent to |
| `token_hash` | text not null, unique | SHA-256 hex of the plaintext token |
| `expires_at` | timestamptz not null | short TTL (15 min — a login link, not an approval window) |
| `used_at` | timestamptz null | set on redemption; a used token is invalid even if not expired |
| `created_at` | timestamptz not null | |

Kept separate from `ContractApprover.TokenHash` deliberately: that field
is a single mutable slot tied to one active `ApprovalRequest` per
approver, and overwriting it for a generic login would break in-flight
approval links. `ClientUser`/`ClientSession` are unchanged and reused
as-is — `GetOrCreateClientUser`/`CreateClientSession` already do exactly
what a self-service login needs.

## Routes (new `controllers/client_portal.go`, registered from
`registerApprovalPortalRoutes`'s sibling on the same phishing-server
router, same CSRF sub-router)

| Route | Method | Behavior |
| --- | --- | --- |
| `/portal/login` | GET | Renders "enter your e-mail" form |
| `/portal/login` | POST | Looks up `ContractApprover` by email; if found, mints a `portal_login_tokens` row and e-mails the link via a new `approvals.SendPortalLoginEmail`; always flashes the same generic "if that address is registered, check your inbox" message |
| `/portal/login/verify` | GET `?token=` | Validates + marks the token used, `GetOrCreateClientUser`, `CreateClientSession`, sets the `ethphish_client` cookie, redirects to `/portal` |
| `/portal` | GET (session required) | Dashboard: campaign list with status + aggregate counts (sent/opened/clicked/submitted/reported), reusing `models.GetCampaignSummariesForTenant` |
| `/portal/campaigns/{id}` | GET (session required) | One campaign's aggregate breakdown and event timeline counts (no named results) |
| `/portal/reports` | GET (session required) | Form to pick campaigns/date range + format, submits to... |
| `/portal/reports` | POST (session required) | ...queues a report via `reports.JobService.QueueReport` scoped to the client's tenant, polls/downloads like the admin reports page already does |

`requireClientSession` (already in `controllers/approval_portal.go`) is
reused unchanged — it only depends on the shared cookie, not on the route
prefix.

## Access control notes

- Every query on the `/portal/*` side takes `cs.ClientUser.TenantID`, never
  a user-supplied tenant identifier.
- `GetCampaignSummariesForTenant`/report generation currently accept a
  `uid` parameter designed for admin ownership checks; the client-portal
  call sites pass the tenant's zero-value/placeholder in a way that keeps
  the tenant filter authoritative — verified during implementation that no
  admin-only join silently re-broadens scope when `uid` is client-supplied
  as 0. If it does, a tenant-scoped variant ignoring `uid` gets added
  instead of relaxing the existing one.
- No route under `/portal/*` ever serializes `Result.Email`, `Result.IP`,
  or any other individually-identifying field.

## Non-goals (explicitly deferred)

- Training/quiz content (7.6).
- Client password/traditional login.
- New indicator/metric types beyond what the admin side already computes.
- Per-target detail views for clients.

## Testing

- Model tests for `portal_login_tokens` issuance/expiry/single-use,
  mirroring `models/contract_approval_test.go`'s token round-trip test.
- Controller test verifying `/portal` never includes target-level fields
  in its rendered payload for a tenant with campaign results.
- Controller test verifying a login request for a non-matching e-mail
  returns the same response as a matching one (no enumeration signal).
