# Spec Delta

## MODIFIED Requirements

### Requirement: Card details and editing
Selecting a card SHALL show its rendered question and answer, its scheduling information (state, due, interval, stability, difficulty, reviews, lapses), and an editor for its fields and tags. The editor SHALL come directly after the question and answer, before the scheduling information. Saves SHALL regenerate cards according to the collection rules. When the card's note has other cards, the editor SHALL say that changes also apply to them. For an Image Occlusion note, the editor SHALL be the mask editor in place of the Occlusion and Image fields, followed by the note's other fields.

#### Scenario: Edit from browser
- **WHEN** the user changes a field in the browser editor and saves
- **THEN** the list row and the preview show the new content

#### Scenario: Shared note is disclosed
- **WHEN** the user opens a card whose `Basic (and reversed card)` note has two cards
- **THEN** the editor says that changes also apply to 1 other card

#### Scenario: Image Occlusion card in the browser
- **WHEN** the user selects a card of an Image Occlusion note in the browser
- **THEN** the editor shows the note's image with its shapes and the shape tools, and Header, Back Extra and Comments as fields; the raw Occlusion and Image fields are not shown

#### Scenario: Edit action on an Image Occlusion card
- **WHEN** the user chooses Edit on a card of an Image Occlusion note
- **THEN** the editor scrolls into view and the Header field gets the focus

### Requirement: Mask editor
When adding or editing an Image Occlusion note, in the note editor or in the browser's card details, the editor SHALL let the user choose an image, draw rectangles, ellipses, polygons and text labels on it, select, move, resize and delete shapes, group and ungroup them, choose between "Hide all, guess one" and "Hide one, guess one", and fill the other fields. It SHALL work with mouse and touch. Shapes of an existing note SHALL keep their numbers when the note is edited. A note without an image or without shapes SHALL NOT be saved.

#### Scenario: Create from an image
- **WHEN** the user picks a diagram, draws two rectangles and adds the note
- **THEN** a note with two cards is created, and its image is stored in the collection's media

#### Scenario: Edit masks
- **WHEN** the user opens an existing Image Occlusion note in the editor, deletes one shape and saves
- **THEN** the card for that shape is kept as an empty card until the user removes it with Empty Cards, and the other cards keep their scheduling

#### Scenario: Numbers not reused
- **WHEN** the user later adds a shape to that note
- **THEN** it gets a number no card of the note has, so it does not take over the empty card's scheduling

#### Scenario: Edit masks in the browser
- **WHEN** the user selects a card of an Image Occlusion note with shapes 1 and 2 in the browser, draws a rectangle, and saves
- **THEN** the note gets a third card for the new shape, shapes 1 and 2 keep their numbers, and the list and preview show the change

#### Scenario: Change image in the browser
- **WHEN** the user chooses Change image in the browser's editor and picks another picture
- **THEN** the mask editor shows the new picture with the shapes where they were, and saving stores it in the media and in the Image field

#### Scenario: No shapes left
- **WHEN** the user deletes every shape of an Image Occlusion note in the browser and saves
- **THEN** the note is not saved and the editor asks the user to draw shapes
