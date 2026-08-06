# EthPhish Redesign + Sprint05 Participant Data Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the admin UI's visual identity with an EthPhish dark/light theme (removing the six inherited Anglerphish themes) and implement the Sprint05.md participant data model: extended target fields, CSV/XLSX import with validation and preview, and segmentation filters.

**Architecture:** Reuse the existing CSS-class-scoped theme system (`applyTheme()` + `localStorage`) for the redesign — new theme files, no template restructuring. Extend the existing `BaseRecipient` struct and `ParseCSV` header-detection pattern for the new fields; XLSX import converts to CSV client-side (SheetJS) and reuses the existing `/api/import/group` endpoint unchanged.

**Tech Stack:** Go 1.25 (backend, GORM legacy), vanilla JS + jQuery + DataTables (frontend, gulp-built), PostgreSQL 17 + SQLite (tests only), Bootstrap 3 CSS.

## Global Constraints

- Follow the existing paired-migration convention: one `.sql` file per driver in `db/db_postgres/migrations/` and `db/db_sqlite3/migrations/`, `-- +goose Up` / `-- +goose Down` markers, timestamp-prefixed filename.
- No new Go dependencies (XLSX handled entirely client-side per spec).
- New JS vendor files go in `static/js/src/vendor/` and must be added to `vendorjs` in `gulpfile.js`; run `npx gulp` (or `yarn gulp`) to rebuild `static/js/dist/`.
- Tags field is a single free-text, comma-separated column — no new normalized table.
- Existing `custom` field pattern is the template for all new text columns (nullable, no default beyond empty string).
- Old theme CSS files and their `<link>`/`<option>` references are deleted outright, not deprecated in place.

---

## Task 1: Database migration — participant profile fields

**Files:**
- Create: `db/db_postgres/migrations/20260806060000_add_participant_profile_fields.sql`
- Create: `db/db_sqlite3/migrations/20260806060000_add_participant_profile_fields.sql`
- Test: `tests/integration/postgres_test.go` (extend)

**Interfaces:**
- Produces: `targets.department`, `targets.company`, `targets.city`, `targets.state`, `targets.country`, `targets.unit`, `targets.tags` — all nullable `TEXT`/`VARCHAR`, no default. Consumed by Task 2 (model).

- [ ] **Step 1: Write the Postgres migration**

```sql
-- +goose Up
ALTER TABLE targets ADD COLUMN department VARCHAR(255);
ALTER TABLE targets ADD COLUMN company VARCHAR(255);
ALTER TABLE targets ADD COLUMN city VARCHAR(255);
ALTER TABLE targets ADD COLUMN state VARCHAR(255);
ALTER TABLE targets ADD COLUMN country VARCHAR(255);
ALTER TABLE targets ADD COLUMN unit VARCHAR(255);
ALTER TABLE targets ADD COLUMN tags TEXT;

-- +goose Down
ALTER TABLE targets DROP COLUMN department;
ALTER TABLE targets DROP COLUMN company;
ALTER TABLE targets DROP COLUMN city;
ALTER TABLE targets DROP COLUMN state;
ALTER TABLE targets DROP COLUMN country;
ALTER TABLE targets DROP COLUMN unit;
ALTER TABLE targets DROP COLUMN tags;
```

Save as `db/db_postgres/migrations/20260806060000_add_participant_profile_fields.sql`.

- [ ] **Step 2: Write the SQLite migration**

```sql
-- +goose Up
ALTER TABLE targets ADD COLUMN department VARCHAR(255);
ALTER TABLE targets ADD COLUMN company VARCHAR(255);
ALTER TABLE targets ADD COLUMN city VARCHAR(255);
ALTER TABLE targets ADD COLUMN state VARCHAR(255);
ALTER TABLE targets ADD COLUMN country VARCHAR(255);
ALTER TABLE targets ADD COLUMN unit VARCHAR(255);
ALTER TABLE targets ADD COLUMN tags TEXT;

-- +goose Down
ALTER TABLE targets DROP COLUMN department;
ALTER TABLE targets DROP COLUMN company;
ALTER TABLE targets DROP COLUMN city;
ALTER TABLE targets DROP COLUMN state;
ALTER TABLE targets DROP COLUMN country;
ALTER TABLE targets DROP COLUMN unit;
ALTER TABLE targets DROP COLUMN tags;
```

Save as `db/db_sqlite3/migrations/20260806060000_add_participant_profile_fields.sql`.

- [ ] **Step 3: Commit**

```bash
git add db/db_postgres/migrations/20260806060000_add_participant_profile_fields.sql \
        db/db_sqlite3/migrations/20260806060000_add_participant_profile_fields.sql
git commit -m "feat: add participant profile fields migration"
```

(Applied automatically by `models.Setup` / the `db-migrate` compose step — no manual apply needed; verified in Task 2's test run.)

---

## Task 2: Model — extend BaseRecipient

**Files:**
- Modify: `models/group.go:61-68` (the `BaseRecipient` struct)
- Test: `models/group_test.go`

**Interfaces:**
- Consumes: migration from Task 1 (columns must exist before this compiles against a real DB in tests).
- Produces: `models.BaseRecipient{Department, Company, City, State, Country, Unit, Tags string}` — consumed by Task 3 (`ParseCSV`), Task 5/6 (frontend via JSON), and automatically by `Target`/campaign `Result` (both embed `BaseRecipient`).

- [ ] **Step 1: Write the failing test**

Add to `models/group_test.go` (near existing `Target`/group creation tests — check the file for the exact helper used to build a test group, e.g. `buildTarget` or literal struct construction, and match it):

```go
func TestTargetProfileFields(t *testing.T) {
	setupTest(t)
	g := Group{
		Name: "Profile Fields Group",
		Targets: []Target{
			{
				BaseRecipient: BaseRecipient{
					Email:      "profile@example.com",
					FirstName:  "Ada",
					LastName:   "Lovelace",
					Department: "Engineering",
					Company:    "Acme Corp",
					City:       "London",
					State:      "England",
					Country:    "UK",
					Unit:       "Platform",
					Tags:       "vip,executive",
				},
			},
		},
	}
	err := PutGroup(&g)
	if err != nil {
		t.Fatalf("error putting group: %v", err)
	}
	got, err := GetGroup(g.Id, g.UserId)
	if err != nil {
		t.Fatalf("error getting group: %v", err)
	}
	if len(got.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(got.Targets))
	}
	target := got.Targets[0]
	if target.Department != "Engineering" || target.Company != "Acme Corp" ||
		target.City != "London" || target.State != "England" ||
		target.Country != "UK" || target.Unit != "Platform" ||
		target.Tags != "vip,executive" {
		t.Fatalf("profile fields did not round-trip: %+v", target)
	}
}
```

Check `models/group_test.go` first for the real signature of `PutGroup`/`GetGroup` (tenant-scoped variants may exist post-Sprint04 — e.g. `PutGroupForTenant`) and use whichever the other tests in that file already use, so this test matches the file's established pattern instead of inventing a new one.

- [ ] **Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./models/... -run TestTargetProfileFields -v`
Expected: FAIL — compile error, `BaseRecipient` has no field `Department` (etc).

- [ ] **Step 3: Add the fields**

In `models/group.go`, change:

```go
type BaseRecipient struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	Custom    string `json:"custom"`
}
```

to:

```go
type BaseRecipient struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Position   string `json:"position"`
	Custom     string `json:"custom"`
	Department string `json:"department"`
	Company    string `json:"company"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Unit       string `json:"unit"`
	Tags       string `json:"tags"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=1 go test ./models/... -run TestTargetProfileFields -v`
Expected: PASS

- [ ] **Step 5: Run the full models package test suite**

Run: `CGO_ENABLED=1 go test ./models/...`
Expected: all PASS (confirms nothing existing broke — e.g. any hand-written `INSERT`/column-list SQL touching `targets` that would need the new columns; if something fails here because of raw SQL column lists elsewhere, fix that file too before moving on).

- [ ] **Step 6: Commit**

```bash
git add models/group.go models/group_test.go
git commit -m "feat: add participant profile fields to BaseRecipient"
```

---

## Task 3: CSV import — recognize new columns

**Files:**
- Modify: `util/util.go:28-34` (regex vars), `util/util.go:51-145` (`ParseCSV`)
- Test: `util/util_test.go`

**Interfaces:**
- Consumes: `models.BaseRecipient` fields from Task 2.
- Produces: `ParseCSV(r *http.Request) ([]models.Target, error)` — unchanged signature, now populates the seven new fields when matching headers are present. Consumed by `controllers/api/import.go:ImportGroup` (no change needed there — it already just calls `ParseCSV` and returns the result).

- [ ] **Step 1: Write the failing test**

Check `util/util_test.go` for how existing CSV-parsing tests build a multipart request (likely a helper that writes a CSV string into a `multipart.Writer` and constructs an `*http.Request`). Reuse that helper. Add:

```go
func TestParseCSVProfileFields(t *testing.T) {
	csvContent := "First Name,Last Name,Email,Phone,Position,Custom,Department,Company,City,State,Country,Unit,Tags\n" +
		"Ada,Lovelace,ada@example.com,+15551234567,Engineer,note,Engineering,Acme Corp,London,England,UK,Platform,\"vip,executive\"\n"
	req := newMultipartCSVRequest(t, "targets.csv", csvContent) // use the file's existing helper name
	targets, err := ParseCSV(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	tg := targets[0]
	if tg.Department != "Engineering" || tg.Company != "Acme Corp" ||
		tg.City != "London" || tg.State != "England" ||
		tg.Country != "UK" || tg.Unit != "Platform" ||
		tg.Tags != "vip,executive" {
		t.Fatalf("profile fields not parsed: %+v", tg)
	}
}

func TestParseCSVProfileFieldsOptional(t *testing.T) {
	csvContent := "First Name,Last Name,Email\nAda,Lovelace,ada@example.com\n"
	req := newMultipartCSVRequest(t, "targets.csv", csvContent)
	targets, err := ParseCSV(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Department != "" || targets[0].Tags != "" {
		t.Fatalf("expected empty optional fields, got %+v", targets[0])
	}
}
```

If `util_test.go` has no existing multipart-CSV helper, write one modeled on `ParseCSV`'s own reading of a multipart file part:

```go
func newMultipartCSVRequest(t *testing.T, filename, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("error creating form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("error writing csv content: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("error closing writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/import/group", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}
```

(add `bytes`, `mime/multipart`, `net/http/httptest` to the test file's imports if not already present).

- [ ] **Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test ./util/... -run TestParseCSVProfileFields -v`
Expected: FAIL — new fields come back empty even though headers matched nothing (no such regex yet).

- [ ] **Step 3: Add the new regexes**

In `util/util.go`, extend the `var (...)` block:

```go
var (
	firstNameRegex  = regexp.MustCompile(`(?i)first[\s_-]*name`)
	lastNameRegex   = regexp.MustCompile(`(?i)last[\s_-]*name`)
	emailRegex      = regexp.MustCompile(`(?i)email`)
	phoneRegex      = regexp.MustCompile(`(?i)phone`)
	positionRegex   = regexp.MustCompile(`(?i)position`)
	customRegex     = regexp.MustCompile(`(?i)custom`)
	departmentRegex = regexp.MustCompile(`(?i)depart(?:ment)?`)
	companyRegex    = regexp.MustCompile(`(?i)compan(?:y|ies)`)
	cityRegex       = regexp.MustCompile(`(?i)city`)
	stateRegex      = regexp.MustCompile(`(?i)state`)
	countryRegex    = regexp.MustCompile(`(?i)country`)
	unitRegex       = regexp.MustCompile(`(?i)unit`)
	tagsRegex       = regexp.MustCompile(`(?i)tags?`)
)
```

- [ ] **Step 4: Extend `ParseCSV`'s header detection and row assembly**

In `ParseCSV`, extend the index/value declarations and the header-matching `switch`:

```go
		fi := -1
		li := -1
		ei := -1
		phi := -1
		pi := -1
		ci := -1
		di := -1
		coi := -1
		cti := -1
		sti := -1
		coui := -1
		ui := -1
		tgi := -1
		fn := ""
		ln := ""
		ea := ""
		ph := ""
		ps := ""
		cm := ""
		dp := ""
		co := ""
		ct := ""
		st := ""
		cou := ""
		un := ""
		tg := ""
		for i, v := range record {
			switch {
			case firstNameRegex.MatchString(v):
				fi = i
			case lastNameRegex.MatchString(v):
				li = i
			case emailRegex.MatchString(v):
				ei = i
			case phoneRegex.MatchString(v):
				phi = i
			case positionRegex.MatchString(v):
				pi = i
			case customRegex.MatchString(v):
				ci = i
			case departmentRegex.MatchString(v):
				di = i
			case companyRegex.MatchString(v):
				coi = i
			case cityRegex.MatchString(v):
				cti = i
			case stateRegex.MatchString(v):
				sti = i
			case countryRegex.MatchString(v):
				coui = i
			case unitRegex.MatchString(v):
				ui = i
			case tagsRegex.MatchString(v):
				tgi = i
			}
		}
```

Note the `switch` is first-match-wins (no `fallthrough`), so ordering matters only where two patterns could both match the same header — none of the new patterns overlap with each other or the existing ones (e.g. `custom` won't match `company` — `custom` regex is `custom`, not a substring of `company`). Keep `customRegex` before `companyRegex` in the switch as a defensive habit, but it isn't load-bearing here.

Then in the row-reading loop, after the existing `if ci != -1 { cm = record[ci] }` block, add:

```go
			if di != -1 && len(record) > di {
				dp = record[di]
			}
			if coi != -1 && len(record) > coi {
				co = record[coi]
			}
			if cti != -1 && len(record) > cti {
				ct = record[cti]
			}
			if sti != -1 && len(record) > sti {
				st = record[sti]
			}
			if coui != -1 && len(record) > coui {
				cou = record[coui]
			}
			if ui != -1 && len(record) > ui {
				un = record[ui]
			}
			if tgi != -1 && len(record) > tgi {
				tg = record[tgi]
			}
```

And extend the `models.Target` literal:

```go
			t := models.Target{
				BaseRecipient: models.BaseRecipient{
					FirstName:  fn,
					LastName:   ln,
					Email:      ea,
					Phone:      ph,
					Position:   ps,
					Custom:     cm,
					Department: dp,
					Company:    co,
					City:       ct,
					State:      st,
					Country:    cou,
					Unit:       un,
					Tags:       tg,
				},
			}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `CGO_ENABLED=1 go test ./util/... -run TestParseCSVProfileFields -v`
Expected: PASS (both `TestParseCSVProfileFields` and `TestParseCSVProfileFieldsOptional`)

- [ ] **Step 6: Run the full util package test suite**

Run: `CGO_ENABLED=1 go test ./util/...`
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add util/util.go util/util_test.go
git commit -m "feat: recognize participant profile columns in CSV import"
```

---

## Task 4: Vendor SheetJS and wire into the build

**Files:**
- Create: `static/js/src/vendor/xlsx.full.min.js`
- Modify: `gulpfile.js:20-40` (`vendorjs` file list)

**Interfaces:**
- Produces: global `XLSX` object (from SheetJS, attached to `window`) available in every page after `vendor.min.js` loads — consumed by Task 6 (`groups.js`).

- [ ] **Step 1: Download the vendor file**

```bash
curl -sL -o static/js/src/vendor/xlsx.full.min.js \
  https://cdn.sheetjs.com/xlsx-0.20.3/package/dist/xlsx.full.min.js
```

Verify it downloaded a real JS file (not an error page):

```bash
head -c 200 static/js/src/vendor/xlsx.full.min.js
```

Expected: starts with `/*! xlsx.js` or similar SheetJS banner comment, not HTML.

- [ ] **Step 2: Add it to the gulp vendor bundle**

In `gulpfile.js`, in the `vendorjs` function's `gulp.src([...])` array, add a new line (placement doesn't matter — it's a self-contained UMD bundle with no dependency on the others):

```js
            vendor_directory + 'xlsx.full.min.js',
```

- [ ] **Step 3: Rebuild the JS bundle**

```bash
npx gulp vendorjs
```

Expected: no errors; `static/js/dist/vendor.min.js` is regenerated (check its mtime/size changed).

- [ ] **Step 4: Commit**

```bash
git add static/js/src/vendor/xlsx.full.min.js gulpfile.js static/js/dist/vendor.min.js
git commit -m "feat: vendor SheetJS for client-side XLSX import"
```

---

## Task 5: Theme CSS — EthPhish dark and light

**Files:**
- Create: `static/css/ethphish-dark-theme.css`
- Create: `static/css/ethphish-light-theme.css`
- Delete: `static/css/dark-theme.css`, `static/css/dark-crimson-theme.css`, `static/css/goldphish-theme.css`, `static/css/lagocephalus-theme.css`, `static/css/light-sand-theme.css`, `static/css/matrix-theme.css`

**Interfaces:**
- Produces: CSS classes `.ethphish-dark-theme` and `.ethphish-light-theme`, matching the selector coverage `dark-theme.css` had (so nothing currently reskinned falls back to unstyled Bootstrap). Consumed by Task 7 (base.html links) and Task 8 (JS toggle).

- [ ] **Step 1: Read the file being replaced, to inventory selector coverage**

```bash
grep -c '^\.' static/css/dark-theme.css
grep -oE '\.[a-zA-Z0-9_-]+(-theme)? [a-zA-Z0-9.:_-]*' static/css/dark-theme.css | sort -u | head -80
```

Use this to confirm the new file's selector list (navbar, sidebar, `.main`, `.panel`, `.table`, `.btn-*`, `.form-control`, `.modal`, `.dropdown-menu`, DataTables classes, `.dataTables_wrapper`, `.select2-*`) is at least as complete as the old one — copy `dark-theme.css` as the starting structure (same selectors, scoped to `.dark-theme` today) and do a bulk rename of the scoping class plus a full color-token swap, rather than writing selectors from scratch. This keeps coverage parity by construction.

- [ ] **Step 2: Write `ethphish-dark-theme.css`**

Start from a copy of `static/css/dark-theme.css`:

```bash
cp static/css/dark-theme.css static/css/ethphish-dark-theme.css
```

Then, in `static/css/ethphish-dark-theme.css`:
1. Replace every occurrence of the scoping class `.dark-theme` (including `html.dark-theme`, `body.dark-theme`, `html.dark-theme .foo` forms) with `.ethphish-dark-theme` — this is a mechanical find/replace across the whole file (`dark-theme` → `ethphish-dark-theme`), safe because the source file uses that exact token consistently as the scope class per its own header comment ("Scoped to .dark-theme").
2. Replace the CSS variable block (the file's own `--dark-*` custom properties, near the top) with:

```css
.ethphish-dark-theme,
html.ethphish-dark-theme {
    --dark-bg-primary: #0d1117;
    --dark-bg-secondary: #161b22;
    --dark-bg-tertiary: #1c222b;
    --dark-bg-card: #1c222b;
    --dark-bg-input: #232a35;
    --dark-border: #2a313c;
    --dark-text-primary: #e6e9ef;
    --dark-text-secondary: #9aa4b2;
    --dark-text-muted: #6b7480;
    --dark-accent: #2c5282;
    --dark-accent-hover: #3a6ba5;
    --dark-navbar: #0d1117;
    --dark-sidebar: #12161c;
    --dark-table-stripe: rgba(255, 255, 255, 0.02);
    --dark-table-hover: rgba(44, 82, 130, 0.15);
}
```

(the rest of the file references these same `--dark-*` variable names throughout, so only this block needs new values — do not rename the variables themselves, only their values, to avoid having to touch every rule).

3. Update the file's header comment to `EthPhish Dark Theme`.

- [ ] **Step 3: Write `ethphish-light-theme.css`**

```bash
cp static/css/ethphish-dark-theme.css static/css/ethphish-light-theme.css
```

Then in `static/css/ethphish-light-theme.css`:
1. Replace `.ethphish-dark-theme` → `.ethphish-light-theme` throughout (mechanical, same rationale as Step 2.1).
2. Replace the variable block with light values (same variable names, so the rest of the file needs no further edits):

```css
.ethphish-light-theme,
html.ethphish-light-theme {
    --dark-bg-primary: #ffffff;
    --dark-bg-secondary: #f4f6f8;
    --dark-bg-tertiary: #eceff2;
    --dark-bg-card: #ffffff;
    --dark-bg-input: #ffffff;
    --dark-border: #e2e5e9;
    --dark-text-primary: #1c222b;
    --dark-text-secondary: #5b6470;
    --dark-text-muted: #8a93a0;
    --dark-accent: #1e3a5f;
    --dark-accent-hover: #2c5282;
    --dark-navbar: #1e3a5f;
    --dark-sidebar: #f4f6f8;
    --dark-table-stripe: rgba(30, 58, 95, 0.03);
    --dark-table-hover: rgba(30, 58, 95, 0.06);
}
```

Note `--dark-navbar` stays a dark navy even in the light theme — the navbar is `navbar-inverse` (white-on-dark) in every existing theme including the un-themed default, so keeping it dark here preserves that established pattern instead of inventing a new light navbar look.

3. Update the file's header comment to `EthPhish Light Theme`.

- [ ] **Step 4: Delete the six legacy theme files**

```bash
git rm static/css/dark-theme.css static/css/dark-crimson-theme.css \
       static/css/goldphish-theme.css static/css/lagocephalus-theme.css \
       static/css/light-sand-theme.css static/css/matrix-theme.css
```

(Task 7 removes the remaining references in templates; do this deletion here so Task 7's grep-for-leftover-references check has something to catch if it's done out of order.)

- [ ] **Step 5: Commit**

```bash
git add static/css/ethphish-dark-theme.css static/css/ethphish-light-theme.css
git commit -m "feat: add EthPhish dark and light theme stylesheets"
```

---

## Task 6: Logo assets

**Files:**
- Create: `static/images/ethphish-logo.png` (copy of `IMG/eth-phish-logo.png`)
- Create: `static/images/ethphish-logo-large.png` (copy of `IMG/eth-phish-logo-maior.png`)

**Interfaces:**
- Produces: the two image files at those paths, served at `/images/ethphish-logo.png` and `/images/ethphish-logo-large.png` (the existing `static/images/` directory is already served at `/images/` — confirmed by current templates referencing `/images/logo_inv_small.png`). Consumed by Task 7.

- [ ] **Step 1: Copy the files**

```bash
cp IMG/eth-phish-logo.png static/images/ethphish-logo.png
cp IMG/eth-phish-logo-maior.png static/images/ethphish-logo-large.png
```

- [ ] **Step 2: Verify they're valid images**

```bash
file static/images/ethphish-logo.png static/images/ethphish-logo-large.png
```

Expected: both report as `PNG image data`.

- [ ] **Step 3: Commit**

```bash
git add static/images/ethphish-logo.png static/images/ethphish-logo-large.png
git commit -m "feat: add EthPhish logo assets"
```

---

## Task 7: Templates — theme selector, links, branding, logo

**Files:**
- Modify: `templates/base.html`, `templates/login.html`, `templates/reset_password.html`, `templates/settings.html`, `templates/api_documentation.html`

**Interfaces:**
- Consumes: theme files from Task 5, logo files from Task 6.
- Produces: pages that link only the two new theme stylesheets, offer only the two new theme options, and show the EthPhish logo/name — consumed visually, nothing downstream depends on this programmatically.

- [ ] **Step 1: Find every reference to the old themes and old branding**

```bash
grep -rn "dark-theme.css\|dark-crimson-theme.css\|goldphish-theme.css\|lagocephalus-theme.css\|light-sand-theme.css\|matrix-theme.css" templates/
grep -rln "Anglerphish\|logo_purple\|logo_inv_small\|gophish_purple" templates/*.html
```

Use this exact list to drive the rest of this task — don't rely on memory of which files matched.

- [ ] **Step 2: `templates/base.html`**

Replace the six theme `<link>` tags:

```html
    <link href="/css/dark-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/dark-crimson-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/goldphish-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/lagocephalus-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/light-sand-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/matrix-theme.css" rel="stylesheet" type="text/css">
```

with:

```html
    <link href="/css/ethphish-dark-theme.css" rel="stylesheet" type="text/css">
    <link href="/css/ethphish-light-theme.css" rel="stylesheet" type="text/css">
```

Replace:

```html
                <img class="navbar-logo" src="/images/logo_inv_small.png" />
                <a class="navbar-brand" href="/">&nbsp;&nbsp;Anglerphish</a>
```

with:

```html
                <img class="navbar-logo" src="/images/ethphish-logo.png" />
                <a class="navbar-brand" href="/">&nbsp;&nbsp;EthPhish</a>
```

Also update the inline pre-render theme script (the block reading `localStorage.getItem('gophish.theme')` around line 26 and line 74, per the earlier grep in this codebase) — check both spots for a class-name allowlist or default value referencing old theme names (e.g. `'default'`) and change the default to `'ethphish-light'`. Read the surrounding ~15 lines at each match before editing, since this script runs before `gophish.js` loads and must independently know the new class names to avoid a flash of unstyled theme on page load.

- [ ] **Step 3: `templates/login.html`**

Same theme `<link>` swap as Step 2 (this template has its own `<head>`, confirmed by the earlier grep). Same navbar logo/brand swap. Additionally replace:

```html
            <img id="logo" src="/images/logo_purple.png" />
```

with:

```html
            <img id="logo" src="/images/ethphish-logo-large.png" />
```

Replace the `<title>` `Anglerphish - {{ .Title }}` with `EthPhish - {{ .Title }}`.

- [ ] **Step 4: `templates/reset_password.html`**

Apply the same theme `<link>`, navbar logo/brand, and any `<title>` changes found by the Step 1 grep for this file.

- [ ] **Step 5: `templates/settings.html`**

Replace the theme `<select>` options:

```html
                <select id="theme_selector" class="form-control" style="max-width: 300px;">
                    <option value="default">Default (Light)</option>
                    <option value="dark-teal">Teal (Dark)</option>
                    <option value="dark-crimson">Crimson (Dark)</option>
                    <option value="matrix">Matrix (Dark)</option>
                    <option value="light-sand">Sand (Light)</option>
                    <option value="goldphish">Goldphish</option>
                    <option value="lagocephalus">Lagocephalus</option>
                </select>
```

with:

```html
                <select id="theme_selector" class="form-control" style="max-width: 300px;">
                    <option value="ethphish-light">EthPhish Light</option>
                    <option value="ethphish-dark">EthPhish Dark</option>
                </select>
```

- [ ] **Step 6: `templates/api_documentation.html`**

Apply whatever "Anglerphish"/"Gophish" branding-text replacements the Step 1 grep found in this file, changing them to "EthPhish".

- [ ] **Step 7: Verify no old references remain**

```bash
grep -rn "dark-theme.css\|dark-crimson-theme.css\|goldphish-theme.css\|lagocephalus-theme.css\|light-sand-theme.css\|matrix-theme.css\|Anglerphish\|logo_purple\|logo_inv_small" templates/
```

Expected: no output.

- [ ] **Step 8: Commit**

```bash
git add templates/base.html templates/login.html templates/reset_password.html \
        templates/settings.html templates/api_documentation.html
git commit -m "feat: apply EthPhish branding and theme selector to templates"
```

---

## Task 8: JS — theme toggle logic

**Files:**
- Modify: `static/js/src/app/gophish.js:557-591` (`applyTheme`, `applyDarkTheme`)
- Modify: `static/js/src/app/settings.js:385-445` (duplicate theme logic)
- Modify: `static/js/src/app/campaign_results.js:1495` (default theme fallback)

**Interfaces:**
- Consumes: CSS classes `.ethphish-dark-theme`/`.ethphish-light-theme` from Task 5.
- Produces: `window.applyTheme(theme: 'ethphish-dark' | 'ethphish-light')` — sole implementation; `settings.js` calls this instead of re-implementing the class toggle.

- [ ] **Step 1: Rewrite `applyTheme` in `gophish.js`**

Replace:

```js
// Global function to apply theme (supports: 'default', 'dark-teal', 'dark-crimson', 'goldphish', 'lagocephalus', 'light-sand', 'matrix')
function applyTheme(theme) {
    // Remove all theme classes first
    document.body.classList.remove('dark-theme', 'crimson-theme', 'goldphish-theme', 'lagocephalus-theme', 'light-sand-theme', 'matrix-theme');
    document.documentElement.classList.remove('dark-theme', 'crimson-theme', 'goldphish-theme', 'lagocephalus-theme', 'light-sand-theme', 'matrix-theme');

    // Apply the selected theme
    if (theme === 'dark-teal') {
        document.body.classList.add('dark-theme');
        document.documentElement.classList.add('dark-theme');
    } else if (theme === 'dark-crimson') {
        document.body.classList.add('crimson-theme');
        document.documentElement.classList.add('crimson-theme');
    } else if (theme === 'goldphish') {
        document.body.classList.add('goldphish-theme');
        document.documentElement.classList.add('goldphish-theme');
    } else if (theme === 'lagocephalus') {
        document.body.classList.add('lagocephalus-theme');
        document.documentElement.classList.add('lagocephalus-theme');
    } else if (theme === 'light-sand') {
        document.body.classList.add('light-sand-theme');
        document.documentElement.classList.add('light-sand-theme');
    } else if (theme === 'matrix') {
        document.body.classList.add('matrix-theme');
        document.documentElement.classList.add('matrix-theme');
    }
    // 'default' theme has no classes, so light theme is shown
}
window.applyTheme = applyTheme;

// Legacy function for backward compatibility
function applyDarkTheme(enabled) {
    applyTheme(enabled ? 'dark-teal' : 'default');
}
window.applyDarkTheme = applyDarkTheme;
```

with:

```js
// Global function to apply theme (supports: 'ethphish-light', 'ethphish-dark')
function applyTheme(theme) {
    document.body.classList.remove('ethphish-dark-theme', 'ethphish-light-theme');
    document.documentElement.classList.remove('ethphish-dark-theme', 'ethphish-light-theme');

    if (theme === 'ethphish-dark') {
        document.body.classList.add('ethphish-dark-theme');
        document.documentElement.classList.add('ethphish-dark-theme');
    } else {
        // Unrecognized or 'ethphish-light' values both fall back to light,
        // so a stored legacy theme name (from before this change) doesn't
        // leave the page unstyled.
        document.body.classList.add('ethphish-light-theme');
        document.documentElement.classList.add('ethphish-light-theme');
    }
}
window.applyTheme = applyTheme;

// Legacy function for backward compatibility
function applyDarkTheme(enabled) {
    applyTheme(enabled ? 'ethphish-dark' : 'ethphish-light');
}
window.applyDarkTheme = applyDarkTheme;
```

- [ ] **Step 2: Read `settings.js`'s theme block in full before editing**

```bash
sed -n '380,450p' static/js/src/app/settings.js
```

It duplicates class-toggle logic under its own local `applyTheme` function (shadowing the global one within that scope) — confirm this by reading the output, then replace that entire local function body so it calls the shared `window.applyTheme` instead of re-implementing the toggle, e.g. change:

```js
    function applyTheme(theme) {
        // ...duplicated class toggle logic...
    }
```

to:

```js
    function applyTheme(theme) {
        window.applyTheme(theme);
    }
```

keeping the rest of the surrounding code (the `$("#theme_selector").on('change', ...)` handler, the `localStorage.setItem('gophish.theme', ...)` calls, and the old-key migration block reading `gophish.use_dark_theme`) as-is — only the inner implementation of the local `applyTheme` changes. Also update the old-key migration's mapped value: where it currently sets `currentTheme` to something like `'dark-teal'`/`'default'` based on the legacy boolean, change those two literal strings to `'ethphish-dark'`/`'ethphish-light'`.

- [ ] **Step 3: `campaign_results.js` default fallback**

At line 1495 (`var currentTheme = localStorage.getItem('gophish.theme') || 'default';`), change the fallback literal to `'ethphish-light'`.

- [ ] **Step 4: Rebuild the JS bundle**

```bash
npx gulp scripts
```

(or the equivalent full build task in `gulpfile.js` if `scripts` isn't the exact exported task name — check `gulpfile.js`'s `exports`/`gulp.task` calls at the bottom of the file first and use whichever task name compiles `static/js/src/app/*.js` into `static/js/dist/`.)

- [ ] **Step 5: Commit**

```bash
git add static/js/src/app/gophish.js static/js/src/app/settings.js \
        static/js/src/app/campaign_results.js static/js/dist/
git commit -m "feat: switch theme toggle logic to EthPhish dark/light"
```

---

## Task 9: groups.html — new fields, filters, XLSX accept, validation styling hooks

**Files:**
- Modify: `templates/groups.html`

**Interfaces:**
- Consumes: nothing new structurally.
- Produces: DOM elements Task 10 (`groups.js`) binds to — element IDs listed below are exact and must match Task 10.

- [ ] **Step 1: Read the current target add/edit form and DataTable header**

```bash
grep -n "targetForm\|addTarget\|id=\"targets\"\|<thead>\|<th>" templates/groups.html | head -40
```

Use this to locate the exact form fields for a single target (first name, last name, email, phone, position, custom) and the DataTable `<thead>` column list, so the additions below go in the same place with matching structure.

- [ ] **Step 2: Add form fields for the new columns**

In the target add/edit form (same block as the existing Position/Custom inputs), add matching `<div class="form-group">` blocks for `department`, `company`, `city`, `state`, `country`, `unit`, `tags`, using the same markup pattern as the existing `position` input (label + `<input type="text" class="form-control">` with an `id` following the existing naming convention, e.g. if position uses `id="position"`, use `id="department"`, `id="company"`, `id="city"`, `id="state"`, `id="country"`, `id="unit"`, `id="tags"`).

- [ ] **Step 3: Add columns to the DataTable header**

In the `<thead>` for the targets table, add `<th>Department</th><th>Company</th><th>City</th><th>State</th><th>Country</th><th>Unit</th><th>Tags</th>` immediately before the existing trailing actions column (the trash-icon column), matching where `<th>Custom</th>` (or equivalent last data column) currently sits.

- [ ] **Step 4: Add segmentation filter controls above the targets table**

Immediately above the `<table id="targets" ...>` element, add:

```html
<div class="row" id="targetFilters" style="margin-bottom: 10px;">
    <div class="col-sm-2">
        <select id="filterCompany" class="form-control input-sm">
            <option value="">All Companies</option>
        </select>
    </div>
    <div class="col-sm-2">
        <select id="filterDepartment" class="form-control input-sm">
            <option value="">All Departments</option>
        </select>
    </div>
    <div class="col-sm-2">
        <select id="filterCity" class="form-control input-sm">
            <option value="">All Cities</option>
        </select>
    </div>
    <div class="col-sm-2">
        <select id="filterState" class="form-control input-sm">
            <option value="">All States</option>
        </select>
    </div>
    <div class="col-sm-2">
        <select id="filterCountry" class="form-control input-sm">
            <option value="">All Countries</option>
        </select>
    </div>
    <div class="col-sm-2">
        <input id="filterTag" type="text" class="form-control input-sm" placeholder="Filter by tag" />
    </div>
</div>
<div id="importSummary" class="text-muted small" style="margin-bottom: 6px;"></div>
```

- [ ] **Step 5: Extend the `.fileupload` accept attribute for XLSX**

Find the file input (`<input type="file" id="csvupload" multiple>`) and add an `accept` attribute so the OS file picker offers XLSX too:

```html
<input type="file" id="csvupload" multiple accept=".csv,.txt,.xlsx">
```

- [ ] **Step 6: Commit**

```bash
git add templates/groups.html
git commit -m "feat: add participant profile fields and filters to groups UI"
```

---

## Task 10: groups.js — new fields, XLSX conversion, validation, dedupe, filters, CSV template/export

**Files:**
- Modify: `static/js/src/app/groups.js`

**Interfaces:**
- Consumes: `window.XLSX` (Task 4), DOM elements from Task 9 (`#filterCompany`, `#filterDepartment`, `#filterCity`, `#filterState`, `#filterCountry`, `#filterTag`, `#importSummary`, and the new per-target form inputs `#department #company #city #state #country #unit #tags`), `models.BaseRecipient` JSON fields from Task 2 (`department`, `company`, `city`, `state`, `country`, `unit`, `tags`).
- Produces: nothing consumed elsewhere — this is the leaf of the chain.

- [ ] **Step 1: Read the current `addTarget`, DataTable init, and rows-add code in full**

```bash
grep -n "function addTarget\|targets = \$\|DataTable(\|rows.add\|targetRows.push" static/js/src/app/groups.js
```

Read the surrounding ~20 lines at each hit before editing — the exact column order in `targetRows.push([...])` (seen earlier in this session: first_name, last_name, email, phone, position, custom, then the trash-icon cell) must be extended in the same order as the `<th>` additions from Task 9 Step 3, or columns will misalign.

- [ ] **Step 2: Extend `addTarget` to accept and render the new fields**

Update the function signature and the row array it builds/pushes to include the seven new values in the same order as the new `<th>` columns (department, company, city, state, country, unit, tags), escaping each with the existing `escapeHtml` helper exactly as the current fields are escaped. Update every call site of `addTarget` in this file (the CSV-import `done` handler, and the "add single target from form" handler) to pass the new values, reading them from the new form inputs' `id`s from Task 9 Step 2 (or empty string `""` where a call site — e.g. re-rendering an existing group's targets from `group.targets` — should read `record.department`, `record.company`, etc. from the API object instead).

- [ ] **Step 3: Row validation after import**

After the CSV-import `done` handler's loop that calls `addTarget` for each imported record (and after the single "add target from form" handler), add validation that tags the just-added row:

```js
var emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
var phonePattern = /^\+?[0-9]{7,15}$/;

function validateTargetRows() {
    var seenEmails = {};
    var invalidCount = 0;
    var duplicateCount = 0;
    targets.DataTable().rows().every(function () {
        var data = this.data();
        var node = $(this.node());
        node.removeClass('row-invalid row-duplicate');
        var email = data[2]; // column index of Email — must match the addTarget column order from Step 2
        var phone = data[3]; // column index of Phone
        var isValid = emailPattern.test(email) && (phone === '' || phonePattern.test(phone));
        if (!isValid) {
            node.addClass('row-invalid');
            invalidCount++;
        }
        var emailKey = email.toLowerCase();
        if (emailKey) {
            if (seenEmails[emailKey]) {
                node.addClass('row-duplicate');
                duplicateCount++;
            }
            seenEmails[emailKey] = true;
        }
    });
    var total = targets.DataTable().rows().count();
    var needsAttention = invalidCount + duplicateCount;
    if (needsAttention > 0) {
        $("#importSummary").text(total + " imported, " + needsAttention + " need attention");
    } else if (total > 0) {
        $("#importSummary").text(total + " imported");
    } else {
        $("#importSummary").text("");
    }
}
window.validateTargetRows = validateTargetRows;
```

Confirm the actual column indices for email/phone against the real order from Step 1/2 before hardcoding them (the plan assumes email is index 2 and phone is index 3, matching first_name, last_name, email, phone — verify, don't assume) and adjust the two index constants if the real order differs.

Call `validateTargetRows()`:
- At the end of the CSV-import `done` handler (after all `addTarget` calls for that import).
- At the end of the "add single target from form" handler.
- After any row deletion (find the existing trash-icon click handler for a target row and add a call there too, since removing a duplicate should let its former pair go back to looking valid).

- [ ] **Step 4: XLSX-to-CSV conversion in the upload handler**

In the `$("#csvupload").fileupload({...})` block's `add` callback, before the existing `acceptFileTypes` check, detect and convert `.xlsx` files:

```js
        add: function (e, data) {
            $("#modal\\.flashes").empty()
            var file = data.originalFiles[0];
            var filename = file['name'];
            var isXlsx = /\.xlsx$/i.test(filename);
            if (isXlsx) {
                var reader = new FileReader();
                reader.onload = function (evt) {
                    var workbook = XLSX.read(evt.target.result, { type: 'array' });
                    var firstSheetName = workbook.SheetNames[0];
                    var csvString = XLSX.utils.sheet_to_csv(workbook.Sheets[firstSheetName]);
                    var csvBlob = new Blob([csvString], { type: 'text/csv' });
                    var csvFile = new File([csvBlob], filename.replace(/\.xlsx$/i, '.csv'), { type: 'text/csv' });
                    data.files = [csvFile];
                    data.submit();
                };
                reader.onerror = function () {
                    modalError("Error reading XLSX file");
                };
                reader.readAsArrayBuffer(file);
                return;
            }
            var acceptFileTypes = /(csv|txt)$/i;
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv, .txt, or .xlsx)")
                return false;
            }
            data.submit();
        },
```

Update the `done` handler's `$.each(data.result, ...)` loop (from Step 2) to also call `validateTargetRows()` once after the loop, per Step 3.

- [ ] **Step 5: Populate and wire the segmentation filters**

Add a function that (re)builds the filter `<select>` options from the values currently in the grid, and wires DataTables' column search:

```js
function refreshFilterOptions() {
    var companies = {}, departments = {}, cities = {}, states = {}, countries = {};
    targets.DataTable().rows().every(function () {
        var data = this.data();
        // Column indices per the addTarget order from Step 2: department, company,
        // city, state, country are the columns immediately after tags/custom —
        // confirm and adjust these five indices against the real column order.
        if (data[6]) departments[data[6]] = true;
        if (data[7]) companies[data[7]] = true;
        if (data[8]) cities[data[8]] = true;
        if (data[9]) states[data[9]] = true;
        if (data[10]) countries[data[10]] = true;
    });
    function fillSelect(selector, values, placeholder) {
        var select = $(selector);
        var current = select.val();
        select.empty().append($('<option value="">' + placeholder + '</option>'));
        Object.keys(values).sort().forEach(function (v) {
            select.append($('<option></option>').attr('value', v).text(v));
        });
        select.val(current || "");
    }
    fillSelect('#filterCompany', companies, 'All Companies');
    fillSelect('#filterDepartment', departments, 'All Departments');
    fillSelect('#filterCity', cities, 'All Cities');
    fillSelect('#filterState', states, 'All States');
    fillSelect('#filterCountry', countries, 'All Countries');
}
window.refreshFilterOptions = refreshFilterOptions;

$("#filterCompany, #filterDepartment, #filterCity, #filterState, #filterCountry").on('change', function () {
    applyTargetFilters();
});
$("#filterTag").on('keyup', function () {
    applyTargetFilters();
});

function applyTargetFilters() {
    var dt = targets.DataTable();
    // Column indices — same five as refreshFilterOptions, plus tags; confirm
    // against the real column order before relying on these numbers.
    dt.column(7).search($("#filterCompany").val() ? '^' + $.fn.dataTable.util.escapeRegex($("#filterCompany").val()) + '$' : '', true, false);
    dt.column(6).search($("#filterDepartment").val() ? '^' + $.fn.dataTable.util.escapeRegex($("#filterDepartment").val()) + '$' : '', true, false);
    dt.column(8).search($("#filterCity").val() ? '^' + $.fn.dataTable.util.escapeRegex($("#filterCity").val()) + '$' : '', true, false);
    dt.column(9).search($("#filterState").val() ? '^' + $.fn.dataTable.util.escapeRegex($("#filterState").val()) + '$' : '', true, false);
    dt.column(10).search($("#filterCountry").val() ? '^' + $.fn.dataTable.util.escapeRegex($("#filterCountry").val()) + '$' : '', true, false);
    dt.column(11).search($("#filterTag").val() || '');
    dt.draw();
}
window.applyTargetFilters = applyTargetFilters;
```

Call `refreshFilterOptions()` at the same points `validateTargetRows()` is called from Step 3 (after CSV/XLSX import, after adding a single target, after row deletion), and once on initial page load after the existing group-targets render loop (the block that calls `addTarget` for each `group.targets` entry when editing an existing group).

- [ ] **Step 6: Extend `downloadCSVTemplate` and `downloadGroup`**

In `downloadCSVTemplate`, add the seven new columns to both example rows (empty string or a short example value is fine, e.g. `'Department': 'Engineering'` on the first row, `''` on the second, following the existing pattern of one row with more fields populated than the other) and update the header list implicitly generated by `Papa.unparse` (it derives headers from object keys, so adding the keys is sufficient — no separate header string to edit for this function).

In `downloadGroup`, update the header line:

```js
            var csvContent = "First_Name,Last_Name,Email,Phone,Position,Custom,Department,Company,City,State,Country,Unit,Tags\n";
```

and add the corresponding `escapeCsvField` calls and row-building for the seven new `target.*` properties, in the same order as the header, following the exact pattern already used for `target.position`/`target.custom`.

- [ ] **Step 7: Add validation row CSS**

The `row-invalid`/`row-duplicate` classes from Step 3 need visible styling. Add to both `static/css/ethphish-dark-theme.css` and `static/css/ethphish-light-theme.css` (from Task 5), inside each file's theme-scoped block:

```css
.ethphish-dark-theme table.dataTable tr.row-invalid,
.ethphish-light-theme table.dataTable tr.row-invalid {
    background-color: rgba(220, 53, 69, 0.12) !important;
}
.ethphish-dark-theme table.dataTable tr.row-duplicate,
.ethphish-light-theme table.dataTable tr.row-duplicate {
    background-color: rgba(255, 193, 7, 0.15) !important;
}
```

(one rule per file, each scoped to its own theme class, so this styling only ever appears alongside the corresponding theme — no changes needed to non-themed base CSS since a default/unthemed state no longer exists after Task 7 always applies one of the two theme classes).

- [ ] **Step 8: Rebuild the JS bundle**

```bash
npx gulp scripts
```

- [ ] **Step 9: Commit**

```bash
git add static/js/src/app/groups.js static/css/ethphish-dark-theme.css \
        static/css/ethphish-light-theme.css static/js/dist/
git commit -m "feat: add profile fields, XLSX import, validation and filters to groups UI"
```

---

## Task 11: Full verification pass

**Files:** none (verification only)

**Interfaces:** none

- [ ] **Step 1: Go build and full test suite**

```bash
CGO_ENABLED=1 go build ./...
CGO_ENABLED=1 go test ./...
```

Expected: build succeeds, all tests PASS.

- [ ] **Step 2: Rebuild the full frontend bundle**

```bash
npx gulp
```

(the default task — check `gulpfile.js`'s `exports.default`/`gulp.task('default', ...)` to confirm what it runs; it should cover both `vendorjs` and `scripts` plus any CSS minification task so `static/css/dist/` and `static/js/dist/` are both current.)

- [ ] **Step 3: Rebuild and restart the dev stack**

```bash
docker compose build server reverse-proxy
docker compose up -d
docker compose ps
```

Expected: all services `Up`/`healthy`.

- [ ] **Step 4: Manual verification checklist**

Using `https://localhost:9443` (or `9444` for admin, per the existing reverse-proxy split):
- Login page shows the EthPhish logo and "EthPhish" branding, not Anglerphish/Gophish.
- Settings → Theme shows exactly two options (EthPhish Light, EthPhish Dark); switching between them restyles the whole app (navbar, sidebar, tables, forms, modals) with no unstyled/default-Bootstrap flashes.
- Reload the page after picking Dark — theme persists (via `localStorage`) and there's no flash of the light theme before dark applies.
- Groups → add/edit a group: the target form has Department/Company/City/State/Country/Unit/Tags fields; adding a target manually shows them in the grid.
- Download the CSV template, fill in the new columns, re-import it — new columns populate; an intentionally malformed email row is visibly flagged (`row-invalid` styling) and the summary line above the grid reports it.
- Build a small `.xlsx` with the same columns (any spreadsheet tool) and import it through the same "Bulk Import Users" control — rows appear identically to a CSV import.
- Duplicate an email between two rows — both get the duplicate highlight.
- Use the company/department/city/state/country filters and the tag text filter above the grid — the visible rows narrow accordingly; clearing a filter restores them.
- Export a group's CSV (`downloadGroup`) — new columns appear with correct values.

- [ ] **Step 5: Record completion**

No code changes in this step — this task exists to gate final delivery on the manual checklist above actually being run, not just planned.
