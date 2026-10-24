# Spec Delta

## ADDED Requirements

### Requirement: Text-to-speech tags
`{{tts <lang>:Field}}` and `[anki:tts lang=<lang>]text[/anki:tts]` SHALL produce an utterance of the text, with HTML stripped, in language `<lang>` (written like `en_US`), shown as a replay button where the tag is. The options `voices=` (a comma-separated preference list) and `speed=` SHALL be honoured when the device's speech engine supports them. `{{tts <lang>:cloze:Field}}` SHALL read the active deletion as "blank" on the question side, and `{{tts <lang>:cloze-only:Field}}` SHALL read only the active deletion's text.

#### Scenario: Field spoken
- **WHEN** the question template is `{{Front}}{{tts es_ES:Front}}` and Front is `perro`
- **THEN** the question shows `perro` and a replay button, and the utterance is "perro" in Spanish

#### Scenario: Cloze only
- **WHEN** card 1 of `{{c1::Paris}} is in France` has `{{tts en_US:cloze-only:Text}}` on the answer side
- **THEN** the answer's utterance is "Paris"

### Requirement: Voice list
`{{tts-voices:}}` SHALL render the list of speech voices available on the device, with their language codes.

#### Scenario: Show voices
- **WHEN** a template contains `{{tts-voices:}}`
- **THEN** the card lists the device's voices, for example `en_US Samantha`
