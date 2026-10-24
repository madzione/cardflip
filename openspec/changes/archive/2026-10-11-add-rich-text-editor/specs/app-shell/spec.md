# Spec Delta

## ADDED Requirements

### Requirement: Paste formatting setting
Settings SHALL include "Strip formatting when pasting", off by default. When on, a normal paste SHALL insert plain text and Ctrl+Shift+V (Cmd+Shift+V on macOS) SHALL keep formatting.

#### Scenario: Setting inverts paste
- **WHEN** the setting is on and the user pastes bold text with Ctrl+V
- **THEN** the field contains the text without bold
