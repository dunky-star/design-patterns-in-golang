# Development Progress

## Project Overview

Design Patterns in Go — implementations of creational, structural, and behavioral patterns.

---

## Completed Features

- Web server with stdlib `net/http`, middleware, templates, static assets
- Factory and Abstract Factory patterns (`pets/`, test page, API routes)
- Database connectivity via `internal/driver` and `.env` DSN
- Repository pattern (`internal/repository`: interface, mysql + test implementations)
- Singleton pattern (`configuration/`: single `Application` with `DB repository.Repository`)
- Builder pattern (`pets/builder.go`, test page UI, `/api/dog-from-builder`, `/api/cat-from-builder`)

---

## In Progress

None currently.

---

## Backlog

- Cat breeds list (still stub in go-breeders reference)
- Adapter Pattern
- Decorator Pattern
- Worker Pool Pattern

---

**Last Updated:** July 5, 2026  
**Status:** Singleton pattern implemented
