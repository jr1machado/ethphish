# Training & quiz (Sprint 07, item 7.6) — design

Date: 2026-08-06
Status: approved

## Purpose

Add an awareness-training subsystem: admin-authored lessons with an
optional multiple-choice/true-false quiz, delivered to targets either by
direct assignment to a group or automatically as a "teachable moment"
after a phishing campaign click/submit. Certificates and an aggregate
indicators dashboard (Sprint08 §14.4) are explicitly deferred.

## Scope decisions (from brainstorming)

- **MVP cut**: content + quiz + progress tracking only. No certificate
  generation, no indicators dashboard — those get their own design later.
- **Two delivery triggers, both feeding the same assignment record**:
  1. Direct assignment — admin picks a training and a group from a new
     "Trainings" admin page; every target in the group gets a
     `TrainingAssignment` and an e-mailed unique link, same delivery
     mechanism (tenant SMTP profile) already used for approval e-mails.
  2. Post-click ("teachable moment") — a campaign optionally names a
     training and a trigger (`click`, `submit`, or `both`) in the
     creation/edit wizard, editable after creation. When that event
     fires, an assignment is created (if one doesn't already exist for
     that result) and the target is redirected into the training instead
     of the normal landing page flow.
- **Lesson structure**: ordered list of HTML pages (title + HTML body,
  same editing pattern as landing pages), navigated sequentially. No
  video/SCORM.
- **Quiz**: mixed question types — each question is independently
  multiple-choice (one correct option) or true/false, chosen per question
  in the admin editor. Admin sets a pass percentage and a max-attempts
  count (0 = unlimited) per training.
- **Access**: token-based, not password-based — consistent with the rest
  of the public-facing surface (campaign `rid`, approval magic links,
  portal login links). The token identifies the assignment; no separate
  login step.

## Data model

New tables (tenant-scoped, `tenant_id` column, app-level scoping via
`withTenantTransaction` — matching the existing contracts/approvals
tables, which don't have PostgreSQL RLS enabled either; consistent with
current state, not a regression).

```
trainings
  id, tenant_id, name, description, created_by, created_at, updated_at

training_lessons
  id, training_id, position, title, html

training_quizzes
  id, training_id, pass_percent, max_attempts        -- one quiz per training, optional

quiz_questions
  id, quiz_id, position, type ('multiple_choice'|'true_false'),
  text, options (json array, multiple_choice only), correct_answer

training_assignments
  id, tenant_id, training_id, campaign_id (nullable), result_id (nullable),
  email, name, token_hash, status ('assigned'|'in_progress'|'completed'|'failed'),
  attempts, best_score, created_at, completed_at

training_lesson_views
  id, assignment_id, lesson_id, viewed_at             -- one row per lesson seen

quiz_attempts
  id, assignment_id, attempt_number, score, passed, answers (json), created_at
```

`campaign_id`/`result_id` are set only for assignments spawned by the
post-click trigger — direct assignments leave them null. `token_hash`
follows the same SHA-256-of-opaque-token pattern as
`models/token.go`/`ContractApprover.TokenHash`/`PortalLoginToken`.

`campaigns` gains two nullable columns: `training_id` and
`training_trigger` (`''`/`'click'`/`'submit'`/`'both'`).

## Routes

**Admin (3333/9444, `RequireLogin`)** — new `controllers/training.go`,
mirroring `controllers/route.go`'s Contracts/ApprovalsCenter pattern:

| Route | Behavior |
| --- | --- |
| `/trainings` | Page: list/create/edit trainings, lessons, quiz questions |
| `/api/trainings/*` | CRUD for training, lessons, quiz, questions (mirrors `controllers/api/contract.go` shape) |
| `/api/trainings/{id}/assign` | POST `{group_id}` — creates one assignment per target in the group, e-mails each a link |

Campaign creation modal (`templates/campaigns.html`) gets a "Training
(optional)" select + trigger radio, alongside the existing "Contract
(optional)" select added last sprint — same pattern, same
`static/js/src/app/campaigns.js` wiring.

**Public (8080/9443)** — new `controllers/training_delivery.go`, mounted
next to `registerApprovalPortalRoutes`/`registerClientPortalRoutes`:

| Route | Behavior |
| --- | --- |
| `/training/{token}` | Redirects to the first unseen lesson (or the quiz, if all lessons are seen) |
| `/training/{token}/lessons/{n}` | Renders one lesson, marks it viewed, links to the next |
| `/training/{token}/quiz` | Renders the quiz form |
| `/training/{token}/quiz` (POST) | Grades the attempt, records it, updates assignment status; blocks submission past `max_attempts` |

`PhishHandler`'s existing click/submit event path gains a check: if the
campaign has a `training_id` and the firing event matches
`training_trigger`, get-or-create a `TrainingAssignment` for that result
and redirect to `/training/{token}` instead of continuing the normal
landing-page response.

## Non-goals (explicitly deferred)

- Certificates (PDF or otherwise).
- Aggregate indicators dashboard (start rate, completion rate, average
  score, department rollups, campaign-impact correlation — Sprint08 §14.4).
- SCORM/video content.
- Exposing training progress in the Sprint 7.5 client portal (the
  dashboard already reserves a nav slot for this; wiring it up is a
  follow-on, not part of this design).

## Testing

- Model tests: assignment token round-trip/expiry-equivalent (single-use
  is not applicable here — an assignment is revisited across multiple
  lessons/attempts, so the token stays valid for the assignment's
  lifetime, not single-redemption like the portal login token), quiz
  grading (multiple-choice + true/false, pass/fail threshold, attempt
  limit enforcement).
- Controller test: post-click trigger creates exactly one assignment per
  result even if the matching event fires more than once (e.g. `both`
  trigger with click then submit).
