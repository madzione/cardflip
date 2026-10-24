# Spec Delta

## ADDED Requirements

### Requirement: Image Occlusion import
Image Occlusion notes in an `.apkg` SHALL be imported with their note type and image, and their cards SHALL render with masks.

#### Scenario: Imported occlusion deck
- **WHEN** the user imports an anatomy deck made with Anki's Image Occlusion
- **THEN** its cards show the diagram with masks, one card per shape group
