# Spec Delta

## ADDED Requirements

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
