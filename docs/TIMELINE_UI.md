# Calendar / Timeline UI (Phase 3.1)

Donna’s planning surface is the **existing Calendar UI**
(`/dashboard/calendar`). Same Day / Week / Month / Agenda chrome —
create, edit, delete, and reminders layered on top.

## Data

| Purpose | API |
| --- | --- |
| Read (events + reminders + RRULE) | `GET /api/v1/timeline?from=&to=` |
| Create / update / delete events | `/api/v1/donna/events` |
| Create / update / delete reminders | `/api/v1/donna/reminders` |
| Sync providers | `POST /api/v1/calendar/sync` |

No provider APIs from the browser.

## Routes

| Path | Role |
| --- | --- |
| `/dashboard/calendar` | Main Calendar |
| `/donna/timeline?occurrence=` | Redirect → `?event=` |
| `/dashboard/timeline` | Redirect → calendar |

## Sidebar

Connected Google / Microsoft / ICS accounts (unchanged), plus:

- Donna Events (purple)
- Donna Reminders (orange)

## CRUD

- Toolbar **Create** → Event or Reminder
- Details drawer → Edit / Delete for Donna-owned items
- Google & ICS stay read-only

## Phase 3.1B (deferred)

Drag & drop · resize · bulk actions · agenda virtualization polish
