# Spec Delta

## ADDED Requirements

### Requirement: Mask editor
When adding or editing an Image Occlusion note, the editor SHALL let the user choose an image, draw rectangles, ellipses, polygons and text labels on it, select, move, resize and delete shapes, group and ungroup them, choose between "Hide all, guess one" and "Hide one, guess one", and fill the other fields. It SHALL work with mouse and touch. Shapes of an existing note SHALL keep their numbers when the note is edited.

#### Scenario: Create from an image
- **WHEN** the user picks a diagram, draws two rectangles and adds the note
- **THEN** a note with two cards is created, and its image is stored in the collection's media

#### Scenario: Edit masks
- **WHEN** the user opens an existing Image Occlusion note in the editor, deletes one shape and saves
- **THEN** the card for that shape is kept as an empty card until the user removes it with Empty Cards, and the other cards keep their scheduling

#### Scenario: Numbers not reused
- **WHEN** the user later adds a shape to that note
- **THEN** it gets a number no card of the note has, so it does not take over the empty card's scheduling
