# services/ai

Python + FastAPI reasoning service.

## Responsibilities

- LLM calls
- Embeddings and memory retrieval
- Planning and prompt management
- Tool-call decisions

## Hard rule

Never touch the database. Persist only by calling `services/api`.

Scaffolded in **M6**.
