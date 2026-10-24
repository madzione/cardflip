# Spec Delta

## ADDED Requirements

### Requirement: Attach media in the editor
The editor SHALL let the user pick image, audio and video files and insert them at the cursor of the focused field: images as `<img>` and audio and video as `[sound:…]`. Dropping files onto a field on desktop and web SHALL do the same. The files SHALL be stored in the collection's media.

#### Scenario: Attach an image
- **WHEN** the user picks `cat.jpg` with the Attach button while the Back field has focus, and adds the card
- **THEN** the Back field contains an `<img>` referencing the stored file and the card shows the image

#### Scenario: Attach audio
- **WHEN** the user attaches `hola.mp3` to the Front field
- **THEN** the Front field contains `[sound:hola.mp3]`

### Requirement: Paste images
Pasting an image from the clipboard into a field SHALL store it as a media file and insert it as an `<img>`, in JPEG format unless the setting "Paste images as PNG" is on.

#### Scenario: Paste a screenshot
- **WHEN** the user copies a screenshot and pastes it into a field
- **THEN** the field shows the image and the media contains a new JPEG file

### Requirement: Record audio in the editor
The editor SHALL let the user record from the microphone and insert the recording into the focused field as `[sound:…]`. If microphone access is refused, the editor SHALL explain this and keep the user's input.

#### Scenario: Record pronunciation
- **WHEN** the user records three seconds of audio and confirms
- **THEN** the focused field contains a `[sound:…]` tag for the new recording

### Requirement: Media file naming
A media file added in the editor SHALL keep its extension. Adding a file with the same content as an existing media file SHALL reuse the existing file. Adding a different file whose name is already taken SHALL store it under a new name.

#### Scenario: Same image twice
- **WHEN** the user attaches the same `cat.jpg` to two notes
- **THEN** the media contains one copy, referenced by both notes

#### Scenario: Name collision
- **WHEN** the media already has a different `cat.jpg` and the user attaches another file named `cat.jpg`
- **THEN** the new file is stored under a different name and the existing one is unchanged
