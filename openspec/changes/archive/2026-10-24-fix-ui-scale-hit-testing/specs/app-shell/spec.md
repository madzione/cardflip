# Spec Delta

## MODIFIED Requirements

### Requirement: UI size
Settings SHALL include a UI size from 75 % to 200 % (default 100 %) that scales text, icons and spacing throughout the app, combined with the operating system's text scale. Changes SHALL apply immediately and persist. At every UI size, every visible element SHALL respond to taps, clicks, hovers, drags and scrolling wherever it appears in the window, including near its right and bottom edges.

#### Scenario: Larger interface
- **WHEN** the user sets the UI size to 150 %
- **THEN** text and icons in every screen are 1.5 times their previous size without restarting

#### Scenario: Smaller interface stays usable
- **WHEN** the UI size is 75 % and the user taps a control at the right or bottom edge of the window
- **THEN** the control responds as it does at 100 %

#### Scenario: Switching back from a smaller size
- **WHEN** the UI size is 75 % and the user opens Settings and picks 100 % in the UI size control
- **THEN** the UI size changes to 100 % at once and stays 100 % after a restart
