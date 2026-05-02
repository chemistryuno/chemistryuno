## Context

The repository contains a mixed frontend and backend application with many user-facing flows spread across routes, pages, handlers, services, and supporting modules. There is no single authoritative audit artifact that states whether a feature is implemented on both sides, so gaps and inconsistencies are easy to miss during development and review.

## Goals / Non-Goals

**Goals:**
- Establish a repeatable audit process for mapping user-facing features across frontend and backend surfaces.
- Produce a structured report that highlights missing counterparts, inconsistent behavior, and obvious integration risks.
- Keep the audit criteria explicit so future runs can be compared consistently.

**Non-Goals:**
- Automatically fixing implementation gaps.
- Redesigning feature boundaries in the product itself.
- Replacing manual engineering review for ambiguous cases.

## Decisions

- Use a feature inventory as the central unit of analysis. This is preferable to scanning individual files in isolation because it allows frontend routes and backend handlers to be compared against the same feature label.
- Treat the audit as a read-only review workflow. This avoids introducing runtime dependencies or coupling the audit to production behavior.
- Require explicit evidence for each finding, including the frontend and backend locations involved. This is better than summary-only output because it makes the results actionable.
- Keep the initial scope heuristic-driven rather than fully automatic. A perfect cross-repo feature classifier would be brittle; a documented audit pass with clear criteria is easier to maintain and refine.

## Risks / Trade-offs

- [False positives] → Mitigate by allowing manual confirmation of ambiguous feature boundaries and by attaching evidence to each finding.
- [False negatives] → Mitigate by expanding the inventory sources to include routes, handlers, services, and shared modules, then reviewing the audit against known major flows.
- [Scope drift] → Mitigate by keeping the feature definition and audit criteria versioned so the report can be compared across runs.
- [Manual overhead] → Mitigate by keeping the report structure concise and focused on only the gaps that require follow-up.

## Migration Plan

1. Define the initial feature inventory and audit criteria.
2. Run the first audit pass and capture the baseline report.
3. Review the report with engineering and QA, then refine feature boundaries where needed.
4. Re-run the audit after each substantial change to keep coverage current.

## Open Questions

- Which feature sources should be treated as authoritative when frontend and backend naming diverge?
- How should ambiguous shared utilities be classified when they support more than one feature?
- What severity levels should be used for missing implementation versus behavior mismatch findings?