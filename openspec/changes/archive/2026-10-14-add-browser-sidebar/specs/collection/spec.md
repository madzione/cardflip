# Spec Delta

## ADDED Requirements

### Requirement: Rename tags
The user SHALL be able to rename a tag on all notes that have it, including its child tags. Renaming into an existing tag SHALL merge them. Moving a tag under another tag SHALL be a rename to `parent::tag`.

#### Scenario: Rename with children
- **WHEN** the user renames `verb` to `verbs`
- **THEN** notes tagged `verb::irregular` are tagged `verbs::irregular`

#### Scenario: Merge
- **WHEN** the user renames `noun` to `nouns` and some notes already have `nouns`
- **THEN** those notes have `nouns` once and no note has `noun`

### Requirement: Delete tags
The user SHALL be able to delete a tag and its child tags from all notes, after a confirmation stating how many notes are affected. Notes SHALL NOT be deleted.

#### Scenario: Delete a tag
- **WHEN** the user deletes the tag `old` used by 12 notes and confirms
- **THEN** no note has `old`, and the 12 notes still exist

### Requirement: Clear unused tags
The user SHALL be able to clear tags that no note uses from every tag list.

#### Scenario: Tag left behind
- **WHEN** the tag `draft` remains in the tag list but no note uses it, and the user clears unused tags
- **THEN** `draft` no longer appears in the sidebar or the tag suggestions
