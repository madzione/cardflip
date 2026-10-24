# Spec Delta

## ADDED Requirements

### Requirement: Check media
The user SHALL be able to check media and see the files no note, template or styling refers to, and the file names notes refer to that are missing, with the space used. Files whose names start with `_` SHALL NOT be reported as unused.

#### Scenario: Unused and missing
- **WHEN** the media contains `old.jpg`, which nothing refers to, and a note refers to `gone.mp3`, which is not in the media
- **THEN** the report lists `old.jpg` as unused and `gone.mp3` as missing

#### Scenario: Static template file
- **WHEN** the media contains `_logo.png`, which no note refers to
- **THEN** `_logo.png` is not listed as unused

### Requirement: Media trash
Deleting unused media from the report SHALL move the files to a media trash. The user SHALL be able to restore the trash or empty it. Storage compaction SHALL move unreferenced media to the trash instead of deleting them, and SHALL delete files that have been in the trash for more than 30 days.

#### Scenario: Restore deleted files
- **WHEN** the user deletes unused files and then chooses Restore deleted
- **THEN** the files are back in the media

### Requirement: Tag notes with missing media
From the report, the user SHALL be able to add the tag `missing-media` to every note that refers to a missing file.

#### Scenario: Find notes to fix
- **WHEN** 3 notes refer to missing files and the user chooses Tag missing
- **THEN** those 3 notes have the tag `missing-media`
