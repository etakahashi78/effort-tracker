---
name: generate-test
description: |
  Guides the developer or agent in generating Go unit tests with manual mocks in the effort-tracker project.

  Trigger when:
  - User requests writing tests for usecases, handlers, or repositories.
  - Adding test code to ensure code-coverage (targeting 100% statement coverage for Usecases).
---

# Generate Test Skill

This skill guides the implementation of Go unit tests using the custom manual mock pattern established in this project.

## Project Testing Guidelines

1. **Manual Mocking**:
   - Do NOT use external mocking libraries (such as `gomock` or `testify/mock`).
   - Write simple mock structs containing function fields for mocking interfaces.
   - Example function-field mock definition:
     ```go
     type mockRepository struct {
         createFunc func(ctx context.Context, p *domain.Project) (*domain.Project, error)
     }

     func (m *mockRepository) Create(ctx context.Context, p *domain.Project) (*domain.Project, error) {
         return m.createFunc(ctx, p)
     }
     ```
   - In each test case, dynamically swap function fields to assert parameters and return mock outputs.

2. **Validation and Error Propagation Tests**:
   - Ensure you test BOTH happy path (200/201/204) and failure path behaviors.
   - For Usecase tests: Assert that the repository is NOT called when input validation fails (returns early with `domain.ErrInvalidInput`).
   - For Handler tests: Assert that the Usecase is NOT called when payload JSON is malformed or invalid request parameters (such as non-numeric ID values) are passed.

3. **Code Coverage Goals**:
   - Target 100% statement coverage for Usecase logic.
   - Use `go test -cover ./...` to verify code coverage percentage.
