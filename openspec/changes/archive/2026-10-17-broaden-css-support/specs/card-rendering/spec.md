# Spec Delta

## MODIFIED Requirements

### Requirement: Images
`<img src="name">` referencing a file in the collection's media SHALL display that image. Its size SHALL follow `width`, `height`, `max-width` and `max-height` from the note type's CSS, inline styles or the `width` and `height` attributes; without such sizing, it SHALL be scaled to fit the card width. A missing media file SHALL show a placeholder instead of failing.

#### Scenario: Imported image shown
- **WHEN** a card from an imported `.apkg` references `cat.jpg`, which was imported with the deck
- **THEN** the image is displayed on the card

#### Scenario: Fixed width
- **WHEN** the CSS contains `img { width: 100px }`
- **THEN** images on the card are 100 logical pixels wide

#### Scenario: No max width
- **WHEN** the CSS contains `img { max-width: none }` and the image is wider than the card
- **THEN** the image keeps its natural width and the card can be scrolled horizontally

## ADDED Requirements

### Requirement: CSS selectors
Rules in the note type's CSS SHALL apply when their selector uses element, class, id, attribute and universal selectors, the descendant, child, adjacent sibling and general sibling combinators, and the pseudo-classes `:first-child`, `:last-child`, `:nth-child()`, `:only-child` and `:not()` with a simple selector. Conflicts SHALL be resolved by importance, specificity and source order as in web browsers. Rules with other selectors SHALL be ignored.

#### Scenario: Descendant selector
- **WHEN** the CSS contains `.front b { color: green }` and a field has `<div class="front"><b>x</b></div>`
- **THEN** `x` is green

#### Scenario: Id and sibling
- **WHEN** the CSS contains `#answer + div { display: none }` and the answer template is `{{FrontSide}}<hr id=answer><div>Hidden</div>`
- **THEN** `Hidden` is not shown

#### Scenario: Specificity
- **WHEN** the CSS contains `#x { color: red }` and then `.y { color: blue }` and an element has id `x` and class `y`
- **THEN** the element is red

### Requirement: Background images
`background-image: url(name)` naming a media file SHALL display that image as the element's background. Remote URLs SHALL be ignored.

#### Scenario: Card background
- **WHEN** the CSS contains `.card { background-image: url("_paper.png") }` and `_paper.png` is in the media
- **THEN** the card shows that image behind its content
