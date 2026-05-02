## ADDED Requirements

### Requirement: Feature inventory coverage
The system MUST identify user-facing feature candidates from both frontend and backend surfaces and build a unified inventory for review.

#### Scenario: Inventory includes both sides
- **WHEN** an audit run is executed
- **THEN** the system MUST collect feature candidates from frontend routes, pages, and components, as well as backend handlers, services, and APIs

### Requirement: Feature parity classification
The system MUST classify each inventoried feature as matched, frontend-only, backend-only, or ambiguous.

#### Scenario: Missing backend counterpart
- **WHEN** a frontend feature has no identified backend implementation
- **THEN** the system MUST mark the feature as frontend-only

#### Scenario: Missing frontend counterpart
- **WHEN** a backend capability has no identified frontend implementation
- **THEN** the system MUST mark the feature as backend-only

### Requirement: Gap reporting
The system MUST emit a report that lists coverage gaps and suspected implementation issues with enough evidence for follow-up.

#### Scenario: Report contains evidence
- **WHEN** a gap or mismatch is found
- **THEN** the report MUST include the relevant frontend and backend locations, the classification, and a short reason

### Requirement: Repeatable audit criteria
The system MUST use explicit audit criteria so repeated runs can be compared consistently.

#### Scenario: Same scope produces comparable results
- **WHEN** the audit is re-run against the same codebase scope and criteria
- **THEN** the resulting findings MUST use the same classification rules and report structure