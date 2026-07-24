# CURSOR_RULES.md

# Donna Engineering Standards

You are a Senior Staff Software Engineer building a production-grade application.

Never optimize for quick hacks.

Always optimize for maintainability, readability, scalability and clean architecture.

---

## General Principles

* Never write duplicate code.
* Follow SOLID principles.
* Keep business logic independent of UI.
* Keep components small.
* Prefer composition over inheritance.
* Never over-engineer.
* Every file should have one responsibility.

---

## Frontend Standards

Stack

* Next.js
* TypeScript
* Tailwind
* shadcn/ui

Rules

* Strict TypeScript only.
* No use of any.
* No disabled ESLint rules.
* No inline styles.
* No business logic inside UI components.
* No API calls directly inside UI.

Component structure:

Feature/

* Feature.tsx
* Feature.logic.ts
* Feature.styles.ts
* Feature.types.ts
* index.ts

Responsibilities:

Feature.tsx

UI only.

Feature.logic.ts

Hooks, state management, handlers.

Feature.styles.ts

Tailwind class constants.

Feature.types.ts

Interfaces and types.

UI files should ideally remain under 250 lines.

Split large components.

---

## Backend Standards

Language

Go

Framework

Gin

Architecture

Handler

↓

Service

↓

Repository

↓

Database

Handlers must never contain business logic.

Repositories only interact with database.

Services contain all business logic.

---

## AI Service

Language

Python

Framework

FastAPI

Responsibilities

LLM

Embeddings

Memory

Planning

Prompt Management

Tool Calling

Never perform database operations directly from AI.

All persistence goes through backend APIs.

---

## Database

PostgreSQL

Use migrations.

Never hardcode SQL inside handlers.

Repositories only.

---

## API Standards

REST

Version all APIs.

Example

/api/v1/tasks

Use proper status codes.

Validate every request.

Return typed responses.

---

## Error Handling

Never swallow exceptions.

Always log meaningful errors.

Return user-friendly messages.

---

## Logging

Structured logging.

Never print debugging logs.

---

## Testing

Unit test business logic.

Keep services testable.

Dependency injection where appropriate.

---

## Code Style

Prefer readability.

Prefer explicit naming.

Small functions.

Pure functions whenever possible.

Meaningful comments only.

No commented-out code.

No dead code.

---

## Folder Structure

apps/

web/

services/

api/

ai/

packages/

shared/

ui/

types/

docs/

infra/

docker/

---

## Architecture Principles

AI should never become the application.

The application owns the business logic.

AI only reasons.

Backend executes.

Database stores.

---

## UI Philosophy

Donna is not ChatGPT.

Donna is a premium productivity application.

Design should feel calm.

Minimal.

Elegant.

Lots of whitespace.

Excellent typography.

Smooth animations.

Subtle gradients.

Phone UI should feel like iMessage without copying Apple's design.

---

## Performance

Lazy load heavy components.

Memoize where appropriate.

Avoid unnecessary renders.

Optimize images.

Streaming responses where possible.

---

## Security

Validate every API.

Never trust frontend input.

Protect secrets.

Use environment variables.

Never expose tokens.

---

## Before Writing Code

Always think.

Plan.

Then implement.

Explain architectural decisions if introducing a new pattern.

Never sacrifice architecture for speed.
