# Phase 1 Build Plan

Milestone order for Donna Phase 1. Each milestone must leave the monorepo buildable and typed.

## M0 — Monorepo foundation

- Repo layout per engineering standards
- Docker Compose: PostgreSQL
- Root tooling docs and env templates
- Shared types package scaffold
- Cursor rules installed

**Exit:** `docker compose up` starts Postgres; docs and rules in place.

## M1 — Go API skeleton

- Gin app with `/api/v1/health`
- Config via env
- Structured logging
- Migration runner (golang-migrate or goose)
- Handler → Service → Repository wiring example

**Exit:** Health endpoint returns 200 against running Postgres.

## M2 — Auth and users

- Google OAuth for Donna account creation
- Profiles, sessions, secure cookies/JWT
- Separate `connections` table for integrations
- Web login/callback pages

**Exit:** User can sign in with Google and hit an authenticated `/api/v1/me`.

## M3 — Core domain schema

Migrations for:

- users, profiles, settings, preferences
- connections, calendar_sources, calendar_events
- tasks, goals, daily_plans, daily_reviews, check_ins
- notes, memories
- conversations, messages
- notifications

**Exit:** Schema migrated; repositories stubbed for tasks and conversations.

## M4 — Web shell + Phone UI

- Next.js App Router + Tailwind + shadcn
- Dashboard layout: left widgets (placeholders), right persistent Phone
- Phone: bubbles, typing indicator, unread, history shell
- Feature folder convention enforced

**Exit:** Authenticated dashboard loads with interactive mock chat UI.

## M5 — Tasks and notes

- CRUD `/api/v1/tasks`, `/api/v1/notes`
- Quick Todo + Backlog widgets wired to API
- Client data layer (no API calls inside UI components)

**Exit:** Create/complete tasks from dashboard without AI.

## M6 — Conversations + AI service

- FastAPI AI service: prompts, streaming, tool schemas
- API orchestration: persist message → call AI → execute tools → persist reply
- Phone UI streams Donna responses

**Exit:** Real chat round-trip with at least `create_task` tool working.

## M7 — Daily planning and accountability

- Morning planning flow via phone
- Persist daily_plans / goals
- Evening review + next-morning follow-up
- Scheduler triggers (API cron or lightweight worker)

**Exit:** Morning goal asked, stored, and followed up next day.

## M8 — Unified calendar (Google)

- `CalendarProvider` interface
- Google Calendar adapter: read/create/update/delete
- Connect Google Calendar as integration (not login)
- Dashboard calendar + meetings widgets
- Defaults: personal / work / reminder calendars

**Exit:** Events sync; Donna can create a meeting via tool call.

## M9 — Memory

- Memory records + embeddings in AI service
- Semantic search tool
- Persist only through API

**Exit:** Donna recalls a prior commitment in a later conversation.

## M10 — Notifications and polish

- Browser push for briefing / meeting / check-in
- Daily summary widgets
- AI insights panel
- Performance: lazy widgets, streaming

**Exit:** Phase 1 success criteria met for daily use.

---

## Explicitly out of Phase 1

Gmail, Drive, GitHub, Voice, Telegram, WhatsApp, OpenClaw, multi-provider calendar implementations beyond Google.

## Current status

M0 in progress (scaffold + docs + rules).
