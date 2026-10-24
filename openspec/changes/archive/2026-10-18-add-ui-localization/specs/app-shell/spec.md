# Spec Delta

## ADDED Requirements

### Requirement: Interface language
Settings SHALL include the interface language: "System default" (default), English or German. With "System default", the app SHALL use the first of the device's preferred languages it has a translation for, otherwise English. Changing the language SHALL apply immediately.

#### Scenario: German interface
- **WHEN** the user chooses German
- **THEN** navigation, buttons, dialogs and settings are shown in German without restarting

#### Scenario: Unsupported system language
- **WHEN** the device's only preferred language is Finnish and the setting is "System default"
- **THEN** the interface is in English

### Requirement: Translated text
All interface text, error messages, relative intervals, dates and numbers SHALL follow the selected language. A text without a translation SHALL be shown in English. Card content SHALL NOT be translated.

#### Scenario: Error in German
- **WHEN** the interface is German and the user renames a deck to a name a sibling already uses
- **THEN** the error explains in German that the name is taken

#### Scenario: Interval label
- **WHEN** the interface is German and a rating button's interval is 3 days
- **THEN** the button shows the interval in German formatting

### Requirement: Built-in names
Stock note types (their names, field names and card type names), and the first deck and the default preset of a new collection, SHALL be created with names in the interface language. Note types, decks and presets that already exist SHALL keep their names when the language changes. Notes of an Image Occlusion note type SHALL be edited with the mask editor whatever its field names.

#### Scenario: New collection in German
- **WHEN** a new collection is created while the interface is German
- **THEN** its deck is named "Standard", its preset "Standard", and it has the note type "Einfach" with the fields "Vorderseite" and "Rückseite", whose card type "Karte 1" shows them

#### Scenario: Adding a stock note type in German
- **WHEN** the interface is German and the user adds the stock note type Cloze
- **THEN** the list of stock note types shows "Lückentext", and the new note type has the fields "Text" and "Rückseite Extra"

#### Scenario: Existing note types are kept
- **WHEN** a collection has the note type "Basic" and the user switches the interface to German
- **THEN** the note type is still named "Basic" with the fields "Front" and "Back"

#### Scenario: German Image Occlusion
- **WHEN** the interface is German and the user adds an Image Occlusion note
- **THEN** the note type "Bildverdeckung" is created with the fields "Verdeckung" and "Bild", and the note is edited with the mask editor
