# Spec Delta

## ADDED Requirements

### Requirement: Playback stops when switching destinations
Switching to another top-level destination (Decks, Browse, Add or Settings), by the user or by the app, SHALL stop all audio and video playback, in study and in the browser's card preview. Returning to the destination SHALL NOT resume or replay it; the card on screen SHALL stay as it was, and the replay controls SHALL play it again on request.

#### Scenario: Study audio stops on switching to Browse
- **WHEN** a sound is playing during study on the Decks destination and the user selects Browse
- **THEN** the sound stops

#### Scenario: Video stops on switching destinations
- **WHEN** a video is playing during study and the user selects Settings
- **THEN** the video stops

#### Scenario: Browser preview stops on switching to Decks
- **WHEN** a sound is playing in the browser's card preview and the user selects Decks
- **THEN** the sound stops

#### Scenario: Nothing resumes on return
- **WHEN** a sound was stopped by switching from study to Browse, and the user selects Decks again
- **THEN** the same card and side are shown, nothing plays, and pressing R plays the side's sounds
