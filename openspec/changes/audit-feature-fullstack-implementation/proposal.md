## Why

The codebase needs a repeatable way to verify that user-facing features are implemented on both the frontend and backend, and to surface gaps or inconsistencies before they become production issues. This change creates that audit capability so coverage and problem discovery can be reviewed in a structured way.

## What Changes

- Introduce a feature-coverage audit flow that maps user-facing functionality across frontend and backend surfaces.
- Produce a structured gap report that highlights missing implementations, mismatched behavior, and other obvious integration issues.
- Define a consistent review scope so future audits can be repeated with the same criteria.
- Surface findings in a format that can be acted on by engineering and QA.

## Capabilities

### New Capabilities
- `feature-coverage-audit`: Audit feature parity between frontend and backend implementations and report gaps or issues.

### Modified Capabilities
- None.

## Impact

This change affects audit and review workflows, supporting documentation, and any code or tooling used to enumerate feature coverage across the app. It may also influence how frontend routes, backend handlers, and shared flows are cataloged for review.