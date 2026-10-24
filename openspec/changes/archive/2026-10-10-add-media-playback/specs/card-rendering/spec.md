# Spec Delta

## REMOVED Requirements

### Requirement: Unsupported media degrades gracefully
**Reason**: Audio and video are now played instead of omitted; the graceful fallback is narrowed to media the device cannot play, in "Unplayable media degrades gracefully".
**Migration**: None for users; sound tags that used to disappear now show controls.

## ADDED Requirements

### Requirement: Sound references
`[sound:name]`, `<audio src="name">` and `<video src="name">` referring to media files SHALL be rendered as sound or video controls in place, in order of appearance. `{{FrontSide}}` SHALL NOT include the question's sound controls.

#### Scenario: Sound in a field
- **WHEN** a field contains `[sound:word.mp3] palabra` and `word.mp3` is in the media
- **THEN** the card shows a play control followed by `palabra`

### Requirement: Unplayable media degrades gracefully
Audio and video that the device cannot play, because of its format or because the file is missing, SHALL be shown as a disabled control, without errors and without affecting the rest of the card.

#### Scenario: Missing sound file
- **WHEN** a field contains `[sound:word.mp3] palabra` and `word.mp3` is not in the media
- **THEN** the card shows `palabra` and a disabled sound control
