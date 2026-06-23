---
name: db-schema
description: |
  Guides the developer or agent in managing database schemas and connections within the effort-tracker project.

  Trigger when:
  - Editing schema.sql.
  - Adding tables, fields, or updating constraints in the database schema.
---

# DB Schema Skill

This skill guides how database schemas are declared, migrated, and verified in this project.

## Schema Configuration Rules

1. **Embedded Schema**:
   - The application does not use external migration frameworks (like `migrate` or `golang-migrate`).
   - The database schema is defined in `internal/infra/database/schema.sql`.
   - On startup, the file is embedded via `//go:embed` and applied automatically.

2. **Idempotence**:
   - Every statement in `schema.sql` MUST be idempotent. Use `CREATE TABLE IF NOT EXISTS`.
   - Do not include direct `DROP TABLE` statements as they will destroy user data.
   - Example table definition:
     ```sql
     CREATE TABLE IF NOT EXISTS tasks (
         id BIGINT AUTO_INCREMENT PRIMARY KEY,
         project_id BIGINT NOT NULL,
         title VARCHAR(255) NOT NULL,
         description TEXT,
         status VARCHAR(50) DEFAULT 'todo' NOT NULL,
         assignee_id BIGINT,
         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
         FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
     );
     ```

3. **Constraints & Indexing**:
   - MySQL 8 automatically generates indexes on Foreign Key constraints. There is NO need to write explicit `CREATE INDEX` statements.
   - Keep constraints defined in-line or at the table level in `schema.sql`.
