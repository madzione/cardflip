# Spec Delta

## Purpose

Playing a card's audio and video during study and in previews, the controls for replaying and seeking, the audio options, and recording one's own voice for comparison.

## ADDED Requirements

### Requirement: Play sounds on study
When a question or answer is shown during study, its sounds SHALL play automatically in order of appearance, unless the deck's "Don't play audio automatically" option is on. Showing the answer SHALL stop the question's sounds, unless the setting "Interrupt current audio when answering" is off. Leaving the card SHALL stop all playback.

#### Scenario: Autoplay question audio
- **WHEN** a card whose question contains `[sound:hola.mp3]` is shown
- **THEN** `hola.mp3` plays without user action

#### Scenario: Autoplay off
- **WHEN** the deck's "Don't play audio automatically" option is on and the same card is shown
- **THEN** nothing plays until the user presses the replay button

### Requirement: Replay controls
Each sound SHALL have a replay button where it appears on the card, unless the setting "Show play buttons on cards with audio" is off. R or F5 SHALL replay the current side's sounds; on the answer side, the question's sounds SHALL play first unless the deck's "Skip question when replaying answer" option is on. 5 SHALL pause or resume, and 6 and 7 SHALL seek back and forward 5 seconds.

#### Scenario: Replay answer side
- **WHEN** the answer is shown, both sides have one sound, "Skip question when replaying answer" is off, and the user presses R
- **THEN** the question's sound plays, followed by the answer's sound

#### Scenario: Pause
- **WHEN** a sound is playing and the user presses 5
- **THEN** playback pauses, and pressing 5 again resumes it

### Requirement: Video
Video references SHALL show an inline player on the card with play, pause and seek controls, and SHALL autoplay with the side's other sounds under the same rules.

#### Scenario: Inline video
- **WHEN** a card contains `[sound:clip.mp4]`
- **THEN** an inline video player shows and plays `clip.mp4`

### Requirement: Playback failures
A sound or video that cannot be played on the device, because of its format or a missing file, SHALL show a disabled control with an explanation, and SHALL NOT interrupt studying or the playback of other sounds.

#### Scenario: Missing file
- **WHEN** a card references `[sound:missing.mp3]` that is not in the media
- **THEN** its button is disabled and says the file is missing, and the card's other sounds still play

### Requirement: Record own voice
During study the user SHALL be able to record from the microphone (Shift+V or a menu item) and play the recording back (V). The recording SHALL be discarded when the card changes and SHALL never be stored in the collection. If microphone access is refused, the app SHALL explain this and continue the session.

#### Scenario: Compare pronunciation
- **WHEN** the user presses Shift+V, says a word, stops the recording and presses V
- **THEN** the recording plays back

#### Scenario: Recording not kept
- **WHEN** the user records on one card and then answers it
- **THEN** V on the next card has nothing to play, and no new media file exists
