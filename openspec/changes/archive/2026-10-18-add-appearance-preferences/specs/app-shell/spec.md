# Spec Delta

## ADDED Requirements

### Requirement: UI size
Settings SHALL include a UI size from 75 % to 200 % (default 100 %) that scales text, icons and spacing throughout the app, combined with the operating system's text scale. Changes SHALL apply immediately and persist.

#### Scenario: Larger interface
- **WHEN** the user sets the UI size to 150 %
- **THEN** text and icons in every screen are 1.5 times their previous size without restarting

### Requirement: Minimalist mode
Settings SHALL include a minimalist mode, off by default, which removes shadows, elevation tints and decorative fills from the interface in light and dark themes.

#### Scenario: Flat interface
- **WHEN** the user turns on minimalist mode
- **THEN** cards, dialogs and bars are shown without shadows

### Requirement: Reduce motion
Settings SHALL include reduce motion with the choices follow system (default), on and off. When reduce motion is in effect, page transitions, expand and collapse animations and fades SHALL be replaced by immediate changes.

#### Scenario: System preference
- **WHEN** the OS requests reduced motion and the setting is "follow system"
- **THEN** opening a study session shows it without a transition animation
