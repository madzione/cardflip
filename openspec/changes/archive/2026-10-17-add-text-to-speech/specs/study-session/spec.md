# Spec Delta

## ADDED Requirements

### Requirement: Speaking during study
When a question or answer with utterances is shown, the study screen SHALL speak them automatically in order of appearance. Showing the answer or leaving the card SHALL stop speech from the previous side. The user SHALL be able to replay a single utterance with its button, and all utterances of the current side with R or F5. When the device has no voice for an utterance's language, its button SHALL be disabled with an explanation, and nothing else SHALL be affected.

#### Scenario: Autoplay on question
- **WHEN** a card whose question has `{{tts es_ES:Front}}` is shown
- **THEN** the Front text is spoken without user action

#### Scenario: Replay with keyboard
- **WHEN** the answer is shown and the user presses R
- **THEN** the answer side's utterances are spoken again

#### Scenario: No voice
- **WHEN** an utterance is in a language with no voice on the device
- **THEN** its replay button is disabled and explains that no voice is installed for that language
