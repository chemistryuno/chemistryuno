## ADDED Requirements

### Requirement: Modals center in the user's visible viewport
The system SHALL display every modal dialog and overlay centered in the user's current browser viewport, not centered relative to the full document, page content, or any scrollable container.

#### Scenario: User opens a modal after scrolling
- **WHEN** the user scrolls a page and opens any modal dialog
- **THEN** the modal content appears centered within the visible viewport
- **AND** the modal is immediately visible without requiring the user to scroll to the page center

#### Scenario: User opens a modal inside a nested layout
- **WHEN** a modal is opened from inside a nested component, panel, or scrollable page section
- **THEN** the modal overlay covers the viewport
- **AND** the modal content is centered relative to the viewport rather than the component's parent container

#### Scenario: User opens a modal on a small screen
- **WHEN** the viewport is narrow or short
- **THEN** the modal remains centered within the visible viewport as far as available space allows
- **AND** overflowing modal content is accessible without moving the modal origin to the document center

### Requirement: Shared and local dialogs use the same centering contract
The system SHALL apply viewport-centered positioning consistently to shared alert, confirm, prompt dialogs and component-local modal overlays.

#### Scenario: Shared dialog opens
- **WHEN** application code opens a shared alert, confirm, or prompt dialog
- **THEN** the dialog uses viewport-fixed overlay positioning
- **AND** the dialog content is centered in the user's visible screen

#### Scenario: Component-local modal opens
- **WHEN** a Vue component opens its own modal overlay
- **THEN** the overlay uses the same viewport-centered behavior as shared dialogs
- **AND** it does not rely on page-height centering or absolute positioning against the document
