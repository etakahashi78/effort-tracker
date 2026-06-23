---
name: infra-terraform
description: |
  Guides the developer or agent in managing Terraform configuration code in the infra/ directory.

  Trigger when:
  - Modifying *.tf files under infra/.
  - Creating or configuring AWS resources (such as S3 state buckets, ECS, or EKS).
---

# Infra Terraform Skill

This skill guides the design, structure, and verification of AWS Infrastructure as Code using Terraform.

## Terraform Guidelines

1. **State Management**:
   - Use S3 for remote state storage and DynamoDB for state locking (`aws_dynamodb_table`).
   - For state storage S3 buckets, always enable versioning.
   - Enforce deletion protection: add `prevent_destroy = true` in the bucket's `lifecycle` block.

2. **Code Structure**:
   - Organize resource setups inside logical directories (e.g., `infra/bootstrap` for state tables, `infra/app` for app resources).
   - Use a prefix (such as `var.name_prefix`) for naming resources to avoid naming collisions in sharing AWS accounts.
   - Restrict variables to `variables.tf` and outputs to `outputs.tf`. Do not inline inputs.

3. **Local Verification**:
   - Format configuration files using `terraform fmt`.
   - Run `terraform validate` to ensure configuration code syntax is correct.
