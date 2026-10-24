# media-playback Specification

## Purpose

Playing a card's audio and video during study and in previews, the controls for replaying and seeking, the audio options, and recording one's own voice for comparison.

## Requirements

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

### Requirement: Closing a screen during playback
Closing a screen while its audio or video is playing, such as leaving study with the back button or closing the browser's card preview, SHALL stop the playback and close the screen without crashing or freezing the app.

#### Scenario: Back during a video
- **WHEN** a video is playing during study and the user presses the back button
- **THEN** the video stops, the deck list is shown, and the app keeps running

#### Scenario: Back during a sound
- **WHEN** a sound is playing during study and the user presses the back button
- **THEN** the sound stops, the deck list is shown, and the app keeps running

#### Scenario: Study again after leaving during a video
- **WHEN** the user left study while a video was playing and then opens the same deck again
- **THEN** the session opens and its video can be played

### Requirement: Record own voice
During study the user SHALL be able to record from the microphone (Shift+V or a menu item) and play the recording back (V). The recording SHALL be discarded when the card changes and SHALL never be stored in the collection. If microphone access is refused, the app SHALL explain this and continue the session.

#### Scenario: Compare pronunciation
- **WHEN** the user presses Shift+V, says a word, stops the recording and presses V
- **THEN** the recording plays back

#### Scenario: Recording not kept
- **WHEN** the user records on one card and then answers it
- **THEN** V on the next card has nothing to play, and no new media file exists

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
