---
name: add-entity
description: |
  Guides the developer or agent in adding a new domain entity to the Go REST API project
  following Clean Architecture principles.

  Trigger when:
  - User requests adding a new entity (e.g., User, TimeEntry).
  - Creating new database tables and mapping them to Go structs.
---

# Add Entity Skill

This skill guides the implementation of a new domain entity using Clean Architecture and vertical slice patterns.

## Vertical Slice Steps

Always implement from the innermost layer (Domain) outward:

1. **Domain Definition**:
   - Add the entity struct to `internal/domain/model.go`.
   - Add the repository interface to `internal/domain/repository.go`.
   - Define domain-specific errors in `internal/domain/errors.go` if needed.

2. **Usecase Layer**:
   - Create `internal/usecase/<entity>.go`.
   - Implement the Usecase struct and its constructor `New<Entity>Usecase`.
   - Implement validations and business rules. Trim whitespace, enforce non-empty requirements, and wrap validation errors with `fmt.Errorf("%w: ...", domain.ErrInvalidInput)`.

3. **Persistence (MySQL) Layer**:
   - Create `internal/adapter/persistence/<entity>.go`.
   - Implement the Repository interface using the MySQL driver (`*sql.DB`).
   - Use `var _ domain.<Entity>Repository = (*<Entity>Repository)(nil)` to verify compliance at compile time.
   - For `Create` & `Update`: MySQL lacks `RETURNING` clauses. Execute, get `LastInsertId` or use existing ID, then query via `Get()` to return the freshly retrieved model.
   - If a record is not found, return `domain.ErrNotFound`.

4. **Handler Layer**:
   - Create `internal/adapter/handler/<entity>.go`.
   - Define request payloads (`*Input` structs) and decode bodies using `json.NewDecoder(r.Body)`. Enable `DisallowUnknownFields()`.
   - Define consumer interface mapping to Usecase in the handler file.
   - Implement routing functions, map errors using `mapError` helper (translating domain errors to HTTP statuses).

5. **Routing & Wiring**:
   - Update `internal/infra/router/router.go` to add `/api` routes and associate them with the new handler.
   - Update `cmd/server/main.go` to wire up the repository, usecase, and handler instances, and inject them into `router.New`.
