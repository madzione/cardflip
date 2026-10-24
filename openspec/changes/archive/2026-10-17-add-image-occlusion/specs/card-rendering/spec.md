# Spec Delta

## ADDED Requirements

### Requirement: Occlusion masks
A card whose cloze field contains image occlusion shapes SHALL show the note's image with the shapes drawn over it at their relative positions, scaling with the image, without running template scripts. In "Hide all, guess one" mode, every shape SHALL be masked and the asked shape SHALL be highlighted on the question. In "Hide one, guess one" mode, only the asked shape SHALL be masked. On the answer, the asked shape SHALL be revealed and outlined.

#### Scenario: Hide all, guess one
- **WHEN** card 2 of a note with three shapes in "Hide all, guess one" mode is shown
- **THEN** the question shows all three areas masked, with shape 2 in the highlight colour

#### Scenario: Answer reveal
- **WHEN** the answer of that card is shown
- **THEN** the area under shape 2 is visible with an outline, and shapes 1 and 3 stay masked
