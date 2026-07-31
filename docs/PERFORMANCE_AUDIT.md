# Donna Performance & Network Audit

**Date:** 2026-07-31  
**Scope:** Full monorepo inspection (`apps/web`, `services/api`)  
**Trigger:** Neon monthly network transfer ~92% (4.6GB / 5GB) within a few days  
**Mode:** Read-only audit — **no code changes, no optimizations applied**

---

## Executive summary

Neon transfer is almost certainly dominated by **calendar sync pipelines** (provider → Postgres upserts of wide event rows over a **1 year back / 2 year forward** window), triggered far more often than the 15‑minute background job alone:

1. **Dashboard home** calls `POST /api/v1/calendar/sync` via `useCalendarFreshness` (stale **60s**, refetch on mount + window focus).
2. **Every authenticated page remount** calls `POST /api/v1/calendar/sync/ensure` (`RequireAuth`), which re-runs the full pipeline when last success is older than **2 minutes**.
3. **Platform scheduler** still runs full source+event sync ~**every 15 minutes per connected account**, with ICS always re-downloading the full feed.
4. Continuous **chat (15s, including background)** and **notifications (30s)** polling add steady read traffic but are smaller than sync writes/egress.

A single active developer with the dashboard open can easily drive **hundreds of HTTP calls/hour** into the API and **megabytes–gigabytes/day** of Neon traffic if calendars are large and syncs keep re-touching `calendar_events` (including `provider_payload` jsonb).

---

## 1. API Audit

Assumptions for “average expected call frequency”: one authenticated user, dashboard open and focused unless noted; React Query defaults `staleTime=30s`, `refetchOnWindowFocus=true`, `retry=1`.

### 1.1 Core read endpoints

| Endpoint | Method | Avg frequency (idle dashboard) | Triggered by | Reads (tables) | Writes | Can poll? | Background? | Cacheable? |
|----------|--------|--------------------------------|--------------|----------------|--------|-----------|-------------|------------|
| `/api/v1/me` | GET | ~1 / session | `AuthProvider` mount | `users` | — | No need | No | Short TTL / session |
| `/api/v1/chat/summary` | GET | **~240 / hour** | `useDonnaThreadSummary` (`DonnaPhoneFab` on all `/dashboard/*`) | `conversations`, `messages` (LIMIT 1) | — | **Yes (15s)** | **Yes (`refetchIntervalInBackground`)** | Yes, 10–30s |
| `/api/v1/chat/messages` | GET | **~300 / hour** when Donna thread open; +1 on open | Live `useChatSession` poll 12s; history on mount | `conversations`, `messages` (≤200) | optional `ClearUnread` | **Yes (12s)** | Only while thread mounted | Yes |
| `/api/v1/notifications` | GET | **~120 / hour** focused | `useNotificationsCenter` 30s; BottomBar/Sidebar/TopBar | `notifications` (**unbounded** by status) | — | **Yes (30s)** | Pauses when tab hidden | Yes |
| `/api/v1/timeline` | GET | **~13–50 / hour** | Dashboard greeting + timeline card; Calendar page on range change/focus | `calendar_events` **×2**, `donna_events`, `donna_reminders` | — | Today via RQ focus/mount | No dedicated poll | Yes (range key) |
| `/api/v1/calendar/sources` | GET | **~13–25+ / hour** on home/calendar | `staleTime: 0`, `refetchOnMount: "always"` | `calendar_sources`, account join | — | Over-fetched | — | Yes |
| `/api/v1/calendar/events` | GET | **~13–25+ / hour** on home | Day events widget | `calendar_events`, `donna_events` | — | Over-fetched | — | Yes |
| `/api/v1/tasks/day/:date` | GET | **~13 / hour** | Home greeting + QuickTasks + Tasks page | `task_occurrences`, `tasks`, tags, notes | may carry-forward | RQ defaults | — | Yes |
| `/api/v1/tasks/history` | GET | **~13 / hour** on Tasks | Tasks page | occurrence summaries | may purge CF | RQ | — | Yes |
| `/api/v1/task-tags` | GET | **~13 / hour** on Tasks | Tasks page | `task_tags` | — | RQ | — | Yes |
| `/api/v1/notes` | GET | **~13 / hour** on Notes | Notes page | `notes` (**unbounded**) | — | RQ | — | Yes |
| `/api/v1/integrations` | GET | **~13 / hour** on Integrations | Integrations page | `connected_accounts` | — | RQ | — | Yes |
| `/api/v1/integrations/ics` | GET | **~13 / hour** on Integrations | Integrations page | ICS accounts | — | RQ | — | Yes |
| `/api/v1/push/vapid-public-key` | GET | ~1 / session | `WebPushRegister` | — | — | No | Once | Long TTL |

### 1.2 Sync / write-heavy endpoints

| Endpoint | Method | Avg frequency | Triggered by | Reads | Writes | Ext network | Notes |
|----------|--------|---------------|--------------|-------|--------|-------------|-------|
| `/api/v1/calendar/sync` | POST | **~1–13+ / hour** on home alone | `useCalendarFreshness` (stale 60s, mount + focus) + manual Sync | accounts, secrets, sources, events, sync_runs, jobs | **Heavy** upserts/soft-deletes | Google / MS / ICS | **Top Neon driver from UI** |
| `/api/v1/calendar/sync/ensure` | POST | **1 per RequireAuth remount**; pipeline if stale > **2m** | Every `/dashboard/*` page’s `RequireAuth` | same | Heavy when not skipped | same | Navigation thrash → sync thrash |
| `/api/v1/calendar/events/sync` | POST | Manual / rare | Explicit events sync | sources, events | Heavy | Providers | |
| `/api/v1/integrations/ics/:id/sync` | POST | Manual | Integrations UI | pipeline | Heavy | ICS HTTP full body | |
| `/api/v1/chat/command` | POST | User-driven | Chat send | conversations, messages + action tables | Inserts | — | Side effects may touch tasks/reminders |
| `/api/v1/push/subscribe` | POST | ~1 / device | After auth | — | `push_subscriptions` | — | |

### 1.3 Example: `GET /timeline`

```text
GET /api/v1/timeline?from=&to=

Called by
  Dashboard greeting (civil day range)
  DashboardTimeline card (zoned day range — may be a second query key)
  Calendar page (view window)
  NotificationScheduler (internal Timeline.List, ~35m window × every minute × all users)

Reads
  calendar_events (+ provider join) — TWICE (Google provider + Microsoft/ICS provider)
  donna_events (range; recurring series with start_at < to)
  donna_reminders (same pattern)
  In-memory RRULE expand (Donna series only, cap 2000/series)

Writes
  None (HTTP path)

Polling
  No dedicated interval; React Query remount/focus (~13–25/h on home)
  Scheduler: 60 Timeline.List / user / hour

Response size
  Day view: small–medium
  Calendar week/month: larger
  Default handler window if params omitted: now-7d … now+30d (handler)
  Rows may include bulky jsonb when loaded via calendar event mapping
```

### 1.4 Example: `POST /calendar/sync`

```text
POST /api/v1/calendar/sync

Called by
  useCalendarFreshness (dashboard home day-events path)
  Calendar “Sync” button
  (Related) sync/ensure when stale; scheduler job every ~15m / account

Reads / Writes
  connected_accounts, credential_secrets (token refresh may write)
  calendar_sources upsert / soft-delete
  calendar_events Get+Create/Update/SoftDelete per remote event (N+1 style)
  calendar_sync_runs insert/update
  scheduler_jobs ensure/reschedule

External
  Google Calendar list + events pages (maxResults 250)
  Microsoft Graph calendars + calendarView/delta
  ICS: full feed download every sync (cursor forced empty)

Window (initial / recovery)
  Lookback 365d, lookahead 730d

Can it be cached?
  Results should be DB-backed; UI should call ensure/sync far less often
```

---

## 2. Frontend Polling Audit

### 2.1 QueryClient defaults

**File:** `apps/web/src/providers/QueryProvider.tsx`

| Option | Value |
|--------|--------|
| `staleTime` | `30_000` |
| `refetchOnWindowFocus` | `true` |
| `retry` | `1` |
| `gcTime` | unset → TanStack v5 default **300_000** |
| `refetchInterval` | unset → false |
| `refetchOnMount` | unset → true (if stale) |
| `refetchOnReconnect` | unset → true |

### 2.2 Continuous polls (always-on while dashboard shell alive)

| Source | Interval | Background | Endpoint | Est. calls/hour |
|--------|----------|------------|----------|-----------------|
| `useDonnaThreadSummary` | **15s** | **Yes** | `GET /chat/summary` (fallback `GET /chat/messages?mark_read=false`) | **~240** (up to ~480 if fallback every time) |
| `useNotificationsCenter` | **30s** | No (RQ pauses when hidden) | `GET /notifications` | **~120** focused |
| `useChatSession` live | **12s** | While thread mounted | `GET /chat/messages?mark_read=false` | **~300** |

**Important:** `DonnaPhoneFab` mounts `useDonnaThreadSummary` on **every** `/dashboard/*` route.  
**Important:** Home phone column opens Donna by default when unread=0 → **12s message polling on `/dashboard`**.

### 2.3 Page → API waterfall (home `/dashboard`)

```text
/dashboard
  AuthProvider → GET /me (once/session)
  RequireAuth mount → POST /calendar/sync/ensure
       └─ if last sync > 2m → full calendar pipeline (Neon + Google/MS/ICS)
  DonnaPhoneFab → GET /chat/summary every 15s (also in background)
  BottomBar → GET /notifications every 30s
  useCalendarDayEvents → useCalendarFreshness
       └─ POST /calendar/sync (stale 60s; mount + focus)
       └─ invalidate ["calendar",…] → GET sources + GET events
  Greeting → GET /tasks/day/{date}, GET /timeline (civil day)
  DashboardTimeline → GET /timeline (zoned day; possibly second key)
  Phone Donna thread (unread=0) → GET /chat/messages every 12s
```

### 2.4 React Query inventory (every `useQuery`)

| queryKey | File | staleTime | refetchInterval | refetchOnWindowFocus | refetchOnMount | enabled | retry | Est. calls/h (page open) |
|----------|------|-----------|-----------------|----------------------|----------------|---------|-------|--------------------------|
| `["notifications","list"]` | `Notifications.logic.ts` | 30s (default) | **30s** | true | default | auth+hydrated | 1 | ~120 |
| `["chat","summary"]` | `useDonnaThreadSummary.ts` | **8s** | **15s** + **inBackground** | true | default | always (FAB) | 1 | ~240 |
| `["calendar","freshness"]` | `useCalendarFreshness.ts` | **60s** | — | **true** | **true** | home day-events | 1 | ~1–13 (**POST sync**) |
| `["calendar","sources","v2"]` | `useCalendarDayEvents` / `Calendar.logic` | **0** | — | true | **"always"** | when mounted | 1 | ~13–25+ |
| `["calendar","events","v2", from, to]` | `useCalendarDayEvents.ts` | **0** | — | true | **"always"** | home | 1 | ~13–25+ |
| `["calendar","timeline", from, to]` | `Calendar.logic.ts` | 30s | — | true | default | calendar page | 1 | ~13+ |
| `["tasks","day", date]` | greeting / QuickTasks / Tasks | 30s | — | true | default | when mounted | 1 | ~13 |
| `["timeline","items", from, to]` | greeting / DashboardTimeline / Timeline | 30s / always remount on DT | — | true | DT: **always** | when mounted | 1 | ~13–25 |
| `["tasks","tags"]` | `Tasks.logic.ts` | 30s | — | true | default | tasks | 1 | ~13 |
| `["tasks","history", from, to]` | `Tasks.logic.ts` | 30s | — | true | default | tasks | 1 | ~13 |
| `["notes","list"]` | `Notes.logic.ts` | 30s | — | true | default | notes | 1 | ~13 |
| `["integrations"]` | `Integrations.logic.ts` | 30s | — | true | default | integrations | 1 | ~13 |
| `["integrations","sources"]` | same | 30s | — | true | default | integrations | 1 | ~13 |
| `["integrations","ics"]` | same | 30s | — | true | default | integrations | 1 | ~13 |

`gcTime`: never overridden → **5 minutes** everywhere.  
No `useInfiniteQuery` / WebSocket / SSE found.  
Service worker (`sw.ts`): push display only — **no** API polling (SW disabled in `next dev`).

### 2.5 Non-RQ timers that hit API

| Location | Timer | API |
|----------|-------|-----|
| `Chat.logic.ts` | `setInterval` 12s | chat messages poll |
| UI clocks (`DayView`, `StatusBarTime`, greeting hour) | 15–60s | **none** |

---

## 3. Backend Scheduler Audit

Three in-process `time.Ticker` loops (no cron library). Wired in `services/api/internal/app/app.go`.

| Ticker | Interval | Purpose | DB reads | DB writes | Timeline? | Google/MS/ICS? | Notifications? | Overlap risk |
|--------|----------|---------|----------|-----------|-----------|----------------|----------------|--------------|
| **NotificationScheduler** | **1 min** | Enqueue PENDING notifs for **all active users** | `users` list; per user: Timeline (events×2 + donna events/reminders); EXISTS on `notifications` | INSERT `notifications` (idempotent) | **Yes** (35m window) | No | Creates rows | **High if multiple API replicas** (no lock) |
| **NotificationDispatcher** | **1 min** | PENDING→SENT; chat bubble; Web Push | due `notifications`; 7d chat backfill; `push_subscriptions`; conversations/messages | UPDATE notifications; INSERT messages; soft-delete push | No | Push HTTP only | **Yes** | **High multi-replica** (no row claim) |
| **Platform scheduler Runner** | **Poll 30s**; job cadence **~15 min / account** | `calendar_sync` jobs | `scheduler_jobs` ListDue/Claim; then full pipeline | sync_runs, sources, events, accounts, secrets, jobs | No | **Yes** full pipeline | No | Claim helps same job; **manual sync can race** |

### 3.1 NotificationScheduler detail

- Interval: `NotificationSchedulerInterval = 1m`
- Window: lookahead 20m + max policy lead 15m ≈ **[now, now+35m)**
- Per tick: `ListActiveIDs` then `EnqueueForUser` for **each** active user (not “online only”)
- Each enqueue → full `Timeline.List` → **duplicate** `calendar_events` range queries

**Est. DB ops/hour (1 user):** ~60 user lists + ~120 calendar_events range + ~60 donna_events + ~60 donna_reminders + EXISTS probes.  
**× N users** linearly.

### 3.2 NotificationDispatcher detail

- Batch limit 100/tick
- Idle: ~60 `ListDuePending` + ~60 `ListRecentChatDelivered` / hour (global)
- Chat backfill scans SENT/READ up to **7 days**

### 3.3 Platform calendar_sync job detail

- Runner poll: **30s**, batch 10
- Job interval payload: **15 minutes** (`CalendarSyncIntervalMinutes`)
- Despite some payload naming, job runs **`SyncPipelineForAccount` = sources + events**
- ICS: **always full HTTP re-fetch** (sync cursor cleared)

---

## 4. Google Sync Audit

| Trigger | Frequency | Pipeline | Neon | External |
|---------|-----------|----------|------|----------|
| **Login / Google connect callback** | Rare | Link account + bootstrap job (`Immediate`) | Account/secret/source writes | OAuth + first sync soon |
| **Scheduler `calendar_sync`** | ~**4 / hour / account** | Full sources+events | Heavy upserts | Google Calendar API pages |
| **`POST /calendar/sync`** (freshness) | Up to **every 60s+focus** on home | Full pipeline | Heavy | Google |
| **`POST /calendar/sync/ensure`** | Each RequireAuth remount; sync if age > **2m** | Full if stale | Heavy when run | Google |
| **Manual Sync button** | User | Full | Heavy | Google |
| **Webhook** | **Not implemented** | — | — | — |
| **Timeline read** | Frequent | **DB only** (no live Google) | Reads | — |

**Sync window constants** (`constant/calendar.go`):

- `CalendarSyncStaleAfter = 2 * time.Minute` ← extremely aggressive for ensure/freshness
- `CalendarEventSyncLookback = 365d`
- `CalendarEventSyncLookahead = 730d`

**Per-event DB pattern:** `upsertEvent` does Get-by-provider-id then Create/Update — **N+1 round-trips** across large calendars.

**Microsoft / ICS:** same pipeline; ICS never incremental.

---

## 5. Timeline Audit

| Question | Finding |
|----------|---------|
| How often is Timeline HTTP called? | Home: ~13–25+/h (possibly **two** different `from/to` keys). Calendar: on view/cursor + focus. Scheduler: **60/user/h** internal. |
| How much data? | HTTP: typically day or week/month range. Scheduler: 35m. Handler default if missing params: **7d back + 30d forward**. |
| Recurrence expansion? | Donna events/reminders expanded in memory (cap 2000/series). Provider events stored **pre-expanded** in `calendar_events`. |
| Over-fetch? | SQL for recurring Donna rows uses `start_at < to` / `trigger_at < to` → can load **old open-ended series**. |
| Duplicate work? | Google + MS/ICS providers each call **`ListByUserInRangeWithProvider`** with the **same range**, then filter in Go → **2× Neon read** every Timeline.List. |
| Payload bulk? | Calendar event selects commonly include **`provider_payload` jsonb** — large egress per row. |

---

## 6. Notification Audit

| Layer | Frequency | Notes |
|-------|-----------|-------|
| Scheduler enqueue | **1 / min** × all active users | Timeline expansion + INSERT pending |
| Dispatcher | **1 / min** | SENT + chat + Web Push |
| Notification Center UI | **30s** poll while focused | Shared RQ key across shell |
| Browser / SW | Event-driven push | No SW poll of API |
| Chat summary (related) | **15s** incl. background | Used for unread badge / local notify sound |

Channels default: `["CHAT","WEB_PUSH"]`.  
In-app list endpoint can return **unbounded** history for filtered statuses.

---

## 7. Chat Audit

| Call | When | Neon impact |
|------|------|-------------|
| `GET /chat/summary` | Every 15s globally on dashboard | Light (1 message) but **very frequent** |
| `GET /chat/messages` history | Thread open | Up to 200 messages; may mark read |
| `GET /chat/messages?mark_read=false` | Every 12s while live | Repeated list of ≤200 |
| `POST /chat/command` | Send | Writes + Action Layer (tasks/reminders/events…) |
| Invalidates summary after poll/send | Extra summary fetches | Amplifies summary traffic |

Chat does **not** directly call Timeline HTTP, but commands that create reminders/events cause later scheduler/timeline load.

Live Donna thread on **home by default** (unread=0) means message polling is not limited to `/dashboard/chat`.

---

## 8. Database Read Audit

### Repeated / hot reads

| Pattern | Frequency driver | Tables |
|---------|------------------|--------|
| Timeline dual `calendar_events` range | Timeline HTTP + notif scheduler | `calendar_events` |
| Chat summary latest message | 15s poll | `messages`, `conversations` |
| Chat history LIMIT 200 ASC | 12s poll when live | `messages` |
| Notifications list unbounded | 30s poll | `notifications` |
| Calendar sources (staleTime 0, always remount) | Home + calendar | `calendar_sources` |
| Active users full list | Notif scheduler 1/min | `users` |
| `scheduler_jobs` ListDue | Runner 30s | `scheduler_jobs` |
| Dispatcher due + 7d backfill | 1/min | `notifications` |
| Per-event Get during sync | Each sync × event count | `calendar_events` |

### Expensive characteristics

- Wide rows (`provider_payload`)
- Unbounded lists: notifications, notes, donna event/reminder ListByUser
- Recurring series over-select (`start_at < $to`)
- No evidence of `SELECT *`, but column lists are “all fields including jsonb”

---

## 9. Database Write Audit

| Write | Driver | Tables |
|-------|--------|--------|
| Calendar event upsert/soft-delete | Sync pipeline (UI freshness, ensure, scheduler, ICS) | `calendar_events` |
| Source upsert/soft-delete | Sync | `calendar_sources` |
| Sync run rows | Every pipeline | `calendar_sync_runs` |
| Account sync status / tokens | Sync + OAuth refresh | `connected_accounts`, `credential_secrets` |
| Scheduler job reschedule | After sync | `scheduler_jobs` |
| Notification INSERT | Scheduler 1/min/user window | `notifications` |
| Notification UPDATE delivery | Dispatcher | `notifications` |
| Chat message INSERT + unread | Dispatcher chat + user chat | `messages`, `conversations` |
| Task carry-forward | Occasional on day/history read | `task_occurrences` |
| Push subscribe upsert | Login once | `push_subscriptions` |

**Neon transfer note:** UPDATE/INSERT of large jsonb event rows and replication/egress of those changes dominates transfer far more than tiny chat summary SELECTs.

---

## 10. Hotspots

### Top frequency reads (code-inferred)

1. `messages` / `conversations` — chat summary 15s  
2. `notifications` — UI 30s + dispatcher lists  
3. `calendar_events` range — timeline + notif scheduler (**×2**)  
4. `scheduler_jobs` — 30s poll  
5. `calendar_sources` — staleTime 0 remounts  
6. `users` active IDs — 1/min  
7. `donna_events` / `donna_reminders` — timeline/scheduler  
8. Chat messages list ≤200 — 12s when live  
9. `tasks` day queries — home/tasks focus  
10. Per-event Get during sync — bursty  

### Top writes

1. `calendar_events` upserts during sync (**critical**)  
2. `calendar_sources` upserts  
3. `calendar_sync_runs`  
4. `connected_accounts` / secrets during sync  
5. `notifications` insert/update  
6. `messages` from dispatcher + chat  
7. `scheduler_jobs` claim/finish/reschedule  
8. Task occurrence mutations (user)  
9. Notes / donna events (user)  
10. Push subscription upsert (rare)  

### Repeated SELECT / UPDATE / INSERT classes

- **SELECT:** dual calendar_events; chat summary; notifications list; sources always-remount  
- **UPDATE:** notification delivery; calendar event rows on every sync; account sync cursors  
- **INSERT:** notifications pending; sync_runs; chat messages; scheduler next jobs  
- **DELETE:** soft-deletes on sources/events/push (sync & gone endpoints)

---

## 11. Estimated Calls/Hour

### Single active user — idle on `/dashboard` (focused, unread=0, 1 Google account)

| Endpoint | Est. HTTP/h | Reads | Writes | Typical payload | Notes |
|----------|-------------|-------|--------|-----------------|-------|
| `GET /chat/summary` | ~240 | Light | — | Tiny | Background too |
| `GET /chat/messages` | ~300 | Medium (≤200 msgs) | rare unread clear | Medium | Home Donna open |
| `GET /notifications` | ~120 | Grows with history | — | Small–medium | |
| `POST /calendar/sync` | ~1–13 | Heavy | **Heavy** | Sync result small; **DB churn large** | Freshness 60s |
| `POST /calendar/sync/ensure` | ~nav-dependent | Heavy if stale | Heavy if run | — | 2m stale window |
| `GET /calendar/sources` | ~13–25 | Medium | — | Small | staleTime 0 |
| `GET /calendar/events` | ~13–25 | Medium | — | Medium | Day range |
| `GET /timeline` | ~13–25 | Medium–heavy | — | Medium | Possibly 2 keys |
| `GET /tasks/day/…` | ~13 | Medium | rare CF write | Small | |
| Scheduler calendar sync | ~4 pipelines | Heavy | Heavy | — | Independent of UI |
| Notif schedule+dispatch | ~60 ticks each | Medium | Low–medium | — | Always on |

**Rough total HTTP to API:** often **700–900+/hour** on home alone.  
**Rough Neon transfer:** dominated by sync pipelines, not by chat JSON size.

### Multiplier risks

| Factor | Effect |
|--------|--------|
| Multiple browser tabs | Multiplies RQ polls (summary/notifications); sync may dedupe per key per client |
| Multiple API replicas (Railway scale-out) | **×R** notification schedule/dispatch work |
| Multiple connected accounts | **×A** calendar sync jobs |
| Large Google calendars | Sync N+1 × event count × payload jsonb size |
| ICS feeds | Full download every sync |

---

## 12. Critical Findings

Ranked by likely Neon transfer impact.

### CRITICAL

1. **Dashboard `useCalendarFreshness` → `POST /calendar/sync` every ~60s of staleness + focus/mount**  
   **Why:** Full provider→DB pipeline (365d/730d window, per-event upserts, jsonb). A few days of leaving home open can burn gigabytes.

2. **`CalendarSyncStaleAfter = 2 minutes` + `RequireAuth` ensure on every page remount**  
   **Why:** Normal navigation across dashboard pages repeatedly re-enters the sync pipeline.

3. **Background scheduler still full-syncs each account ~every 15 minutes (ICS always full)**  
   **Why:** Steady baseline Neon + provider traffic even with UI closed.

4. **Multi-replica duplicate notification scheduler/dispatcher (no distributed lock / row claim)**  
   **Why:** Horizontal scale multiplies DB work and may duplicate chat/push side effects.

### HIGH

5. **Chat summary polled every 15s with `refetchIntervalInBackground: true` on all dashboard routes**  
   **Why:** Constant Neon reads 24/7 while any dashboard tab exists (even hidden).

6. **Live chat messages polled every 12s whenever Donna thread is open (default on home)**  
   **Why:** Re-reads up to 200 messages repeatedly.

7. **Timeline loads `calendar_events` twice per List**  
   **Why:** Doubles read egress for every timeline/scheduler expansion.

8. **Calendar sources/events queries use `staleTime: 0` + `refetchOnMount: "always"`**  
   **Why:** Unnecessary remount/focus read amplification after every sync invalidate.

9. **Sync upsert N+1 + `provider_payload` storage/egress**  
   **Why:** Transfer scales with calendar size, not with UI need.

### MEDIUM

10. **Notifications list polled every 30s; query may be unbounded**  
11. **Possible duplicate timeline query keys on home (civil vs zoned day)**  
12. **Dispatcher 7-day chat backfill query every minute**  
13. **Notification scheduler scans all active users, not recently active**  
14. **Unbounded notes / donna list endpoints as data grows**

### LOW

15. React Query default focus refetch (30s stale) on tasks/notes — modest  
16. Push subscribe once per session — negligible  
17. UI-only clocks (no API)

---

## 13. Recommendations

**Do not implement here — ranked by expected Neon traffic reduction.**

| Rank | Recommendation | Est. transfer reduction | Impact |
|------|----------------|-------------------------|--------|
| 1 | Stop calling full `POST /calendar/sync` from `useCalendarFreshness`; use ensure with a **much larger** stale window (e.g. 15–60m) or rely on scheduler only | **50–80%** | Critical |
| 2 | Raise `CalendarSyncStaleAfter` from 2m → align with job cadence (15m+); debounce `RequireAuth` ensure (once/session) | **20–40%** | Critical |
| 3 | Narrow initial sync window (e.g. 30–90d back / 180d forward) + incremental sync; strip/omit `provider_payload` from hot paths | **20–50%** for large calendars | Critical |
| 4 | ICS: conditional GET / ETag / hash; do not full-rewrite events if unchanged | **High for ICS users** | High |
| 5 | Batch calendar event upserts; avoid per-event Get round-trips | Sync CPU/time + Neon ops | High |
| 6 | Single `calendar_events` query in Timeline; filter providers in one pass | ~**2×** fewer timeline/scheduler event reads | High |
| 7 | Chat summary: 30–60s poll; **disable background refetch**; use push for unread | **~50–80%** of summary traffic | High |
| 8 | Chat messages: poll only when thread visible; prefer `after` cursor / ETag; don’t default-open Donna poll on home | **~300 calls/h** saved when idle home | High |
| 9 | Notifications: 60–120s poll or push-driven invalidate; paginate list | Medium | Medium |
| 10 | Give sources/events `staleTime` (30–60s); drop `refetchOnMount: "always"` | Medium read reduction | Medium |
| 11 | Leader election / job claim for notification schedule+dispatch | Prevents ×replicas burn | Critical if scaled |
| 12 | Scheduler: only users with upcoming items / recent activity | Scales enqueue cost | Medium |
| 13 | Unifyuplicate home timeline query keys (one range helper) | Small | Low |
| 14 | Cap/paginate notifications & notes lists | Growth control | Medium |

### Suggested measurement (still no code change in this audit)

- Neon: top queries by data transfer / rows (pg_stat_statements + Neon metrics)
- Correlate spikes with `calendar_sync_runs` timestamps
- Count `POST /calendar/sync` vs `/sync/ensure` in API access logs per hour

---

## Appendix A — File index

| Area | Paths |
|------|--------|
| RQ defaults | `apps/web/src/providers/QueryProvider.tsx` |
| Auth ensure | `apps/web/src/features/auth/RequireAuth.tsx` |
| Freshness sync | `apps/web/src/features/calendar/useCalendarFreshness.ts` |
| Chat poll | `apps/web/src/features/chat/Chat.logic.ts`, `useDonnaThreadSummary.ts` |
| Notifications poll | `apps/web/src/features/notifications/Notifications.logic.ts` |
| App wiring | `services/api/internal/app/app.go` |
| Sync constants | `services/api/internal/constant/calendar.go` |
| Pipeline | `services/api/internal/business/calendar_pipeline.go`, `calendar_events.go` |
| Timeline providers | `services/api/internal/business/timeline_providers.go` |
| Notif schedule/dispatch | `services/api/internal/business/notification_scheduler.go`, `notification_dispatcher.go` |
| Platform jobs | `services/api/internal/scheduler/runner.go`, `calendarsync/job.go` |

---

## Appendix B — What this audit is not

- Not a live Neon `pg_stat_statements` capture (code inspection only)
- Not a production access-log histogram
- Not an implemented fix list

If needed, a follow-up pass can attach measured bytes/hour per endpoint from Neon + Railway logs to validate the rankings above.
