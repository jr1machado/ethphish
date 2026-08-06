# EthPhish visual redesign + Sprint05 participant data — design

Date: 2026-08-06
Branch: feature/sprint05

## Goal

Two independent changes, both approved for this pass:

1. Replace the admin UI's visual identity with an EthPhish-branded theme
   (dark and light variants, black/gray/dark-blue palette, EthPhish logo),
   removing the six inherited Anglerphish themes.
2. Implement the participant data model and import/segmentation
   functionality described in `INFO/Sprint05.md`: extended target fields,
   CSV/XLSX import with validation and preview, and segmentation filters.

Explicitly out of scope for this pass: landing page / email template
import-export (belongs to the Sprint 17 bundle/library feature in
`INFO/biblia-projeto.md`; needs its own design later).

## A. Visual redesign

### Theming mechanism (unchanged)

Reuse the existing pattern: a body/html CSS class toggled by
`applyTheme(theme)` (`static/js/src/app/gophish.js`), persisted in
`localStorage['gophish.theme']`, driven by a `<select>` in
`templates/settings.html`. All theme CSS files are always linked in
`templates/base.html`; only the active one's class matches anything. No
architecture change — just new theme files and removal of old ones.

### Removed

- `static/css/dark-theme.css`, `dark-crimson-theme.css`,
  `goldphish-theme.css`, `lagocephalus-theme.css`,
  `light-sand-theme.css`, `matrix-theme.css` — unlinked from
  `base.html`, `login.html`, `reset_password.html` and removed as files.
- Their `<option>` entries in `templates/settings.html`.
- The `dark-theme`/`crimson-theme`/etc. branches in `applyTheme()`
  (`gophish.js`) and the duplicate copy in `settings.js`.

### Added

- `static/css/ethphish-dark-theme.css` — class `.ethphish-dark-theme`.
  Palette: page background `#0d1117`, secondary surfaces `#161b22`, cards
  `#1c222b`, borders `#2a313c`, primary text `#e6e9ef`, muted text
  `#9aa4b2`, accent `#1e3a5f` with hover `#2c5282`.
- `static/css/ethphish-light-theme.css` — class `.ethphish-light-theme`.
  Palette: page background `#ffffff`, secondary surfaces `#f4f6f8`, cards
  `#ffffff` with `#e2e5e9` borders, primary text `#1c222b`, muted text
  `#5b6470`, same accent family (`#1e3a5f`/`#2c5282`) so both modes read
  as the same brand.
- Both files follow the existing file's structure (CSS variables scoped
  to the theme class, then rules for navbar, sidebar, tables, buttons,
  modals, forms — copied and re-colored from `dark-theme.css`'s
  selector coverage so nothing currently reskinned goes back to
  Bootstrap defaults).
- `templates/settings.html` theme `<select>` becomes:
  `EthPhish Light` (value `ethphish-light`, default) and
  `EthPhish Dark` (value `ethphish-dark`).
- `applyTheme()` in `gophish.js` (single source of truth; `settings.js`'s
  duplicate calls the shared one instead of reimplementing) toggles
  `ethphish-dark-theme` / `ethphish-light-theme` on `<html>` and `<body>`.

### Logo and branding text

- Copy `IMG/eth-phish-logo.png` → `static/images/ethphish-logo.png`
  (navbar, small) and `IMG/eth-phish-logo-maior.png` →
  `static/images/ethphish-logo-large.png` (login/reset-password hero).
- Replace `<img class="navbar-logo" src="/images/logo_inv_small.png">`
  and `<a class="navbar-brand">Anglerphish</a>` with the new logo and
  "EthPhish" text in `base.html`, `login.html`, `reset_password.html`.
- Replace `<img id="logo" src="/images/logo_purple.png">` on the login
  page with the large logo.
- Replace remaining "Anglerphish"/"Gophish" strings in page `<title>`
  tags and `api_documentation.html` header with "EthPhish".
- Old logo/theme image files (`logo_purple.png`, `logo_inv_small*.png`,
  `gophish_purple*.png`, theme-specific small logos) are removed only if
  nothing else references them after the grep sweep; otherwise left in
  place (unused legacy assets are not a stated goal of this pass).

## B. Sprint05 — participant data, import, segmentation

### Schema

New migration (Postgres and SQLite, matching the existing paired-file
convention under `db/db_postgres/migrations/` and
`db/db_sqlite3/migrations/`) adds nullable text columns to `targets`:
`department`, `company`, `city`, `state`, `country`, `unit`, `tags`.
`tags` is a single free-text column, comma-separated values — consistent
with the existing `custom` field, not a new normalized taxonomy.

### Model

`BaseRecipient` in `models/group.go` gains the seven new fields (JSON
tags `department`, `company`, `city`, `state`, `country`, `unit`,
`tags`). Because `Target` and campaign `Result` both embed
`BaseRecipient`, the API, campaign results, and report export pick up
the new fields with no additional plumbing.

### CSV import (server-side, existing endpoint)

`util.ParseCSV` (`util/util.go`) gains header-regex detection for the
seven new columns, following the existing per-column index pattern
(`firstNameRegex`, `emailRegex`, etc.). Column order in the CSV is
free-form (header-detected), matching current behavior.

### XLSX import (client-side, no backend change)

- Vendor SheetJS's `xlsx.full.min.js` into `static/js/src/vendor/`, add
  it to the `vendorjs` file list in `gulpfile.js`, rebuild
  `static/js/dist/vendor.min.js`.
- In `groups.js`, the file-upload `add` handler detects a `.xlsx`
  extension, reads it with `XLSX.read`, converts the first sheet to a
  CSV string with `XLSX.utils.sheet_to_csv`, wraps it in a `Blob`/`File`
  named `<original>.csv`, and submits that in place of the original file
  through the same `jquery.fileupload` flow to `/api/import/group`. CSV
  files continue to submit unchanged. This keeps `ImportGroup` /
  `ParseCSV` as the single import code path.

### Preview and validation

The existing participant grid in `groups.html` already acts as a preview
(imported rows land in an editable DataTable; nothing is persisted until
the group's Save button is clicked). Add, in `groups.js`:

- Per-row client-side validation after import: email required and
  regex-checked, phone checked against a loose E.164-ish pattern (digits,
  optional leading `+`, 7–15 digits) — invalid rows get a warning icon
  and a `row-invalid` CSS class (styled in both new theme files) rather
  than being silently dropped, so the user can fix or remove them before
  saving.
- Duplicate-email detection across the current grid: a repeated email
  gets a `row-duplicate` highlight. Detection only, no auto-merge.
- A small summary line above the grid after import: "`N` imported, `X`
  need attention" (invalid + duplicate count), so large imports don't
  require scanning the whole table to notice problems.

### Segmentation filters

Above the participant DataTable in `groups.html`, add filter controls
(company, department, city, state, country as `<select>` populated from
distinct values currently in the grid; tag as a free-text contains
filter) wired to DataTables' built-in column search
(`column().search()` + `draw()`). Client-side only, scoped to the
group being edited — no new "saved segment" concept, no schema change
beyond the column additions above.

### CSV template and export

`downloadCSVTemplate()` and `downloadGroup()` in `groups.js` gain the
seven new columns, both in the generated template and in per-group CSV
export, so round-tripping (export → edit → re-import) keeps working.

## Testing

- Go: `models` test for the new `BaseRecipient` fields round-tripping
  through create/read; `util` test for `ParseCSV` recognizing the new
  headers (including a CSV that omits them, to confirm they stay
  optional). Existing PostgreSQL/SQLite migration tests extend to cover
  the new columns.
- Frontend: no existing JS test harness in this repo (confirmed — no
  `*.test.js` / jest config); manual verification via `docker compose
  up` is the existing project convention for UI changes (per
  `CLAUDE.md`/session norms), covering both themes, CSV import, XLSX
  import, validation highlighting, and segmentation filters.

## Risks / explicit non-goals

- No data migration/backfill needed (new columns nullable, existing rows
  unaffected).
- Old theme CSS files are deleted, not deprecated-in-place — no
  migration path for a user with a legacy theme selected; `applyTheme`
  falls back to `ethphish-light` for any unrecognized stored value.
- XLSX parsing trusts the browser-side SheetJS conversion; the server
  still only ever sees CSV text through the existing `ParseCSV`, so no
  new server-side attack surface from XLSX specifically (SheetJS itself
  is a well-established, actively maintained parser).
- Landing page / email template import-export is explicitly deferred.
