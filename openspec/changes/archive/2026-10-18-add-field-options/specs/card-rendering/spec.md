# Spec Delta

## ADDED Requirements

### Requirement: Right-to-left text
Elements with `dir="rtl"` in templates or fields SHALL be laid out right to left, with right alignment by default.

#### Scenario: RTL template
- **WHEN** the question template is `<div dir="rtl">{{Arabic}}</div>`
- **THEN** the Arabic text is laid out right to left and aligned right
