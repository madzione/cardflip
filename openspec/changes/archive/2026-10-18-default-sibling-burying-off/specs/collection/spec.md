# Spec Delta

## MODIFIED Requirements

### Requirement: Burying options
Each deck SHALL have the options bury new siblings and bury review siblings, both off by default, as in Anki. Collections created before these defaults SHALL have both options turned off once in their Default preset; presets the user added and presets imported from Anki SHALL keep their values.

#### Scenario: Defaults
- **WHEN** a deck is created
- **THEN** both burying options are off

#### Scenario: New preset
- **WHEN** the user adds a preset that starts with the default options
- **THEN** both burying options of the preset are off

#### Scenario: Existing collection
- **WHEN** a collection whose Default preset buries new and review siblings, and whose preset `Exam` (added by the user) buries them too, is opened after the update
- **THEN** the Default preset buries neither, and `Exam` still buries both

#### Scenario: Older backup
- **WHEN** the user restores a backup made before these defaults
- **THEN** its Default preset buries neither new nor review siblings, and its other presets keep their burying options
