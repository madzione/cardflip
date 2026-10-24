# card-rendering Specification

## Purpose

Turns a note and one of its note type's templates into the question and answer the user sees, using a subset of Anki's template language, and displays the result as styled HTML with images.

## Requirements

### Requirement: Field substitution
A template SHALL replace `{{FieldName}}` with that field's content. Field names SHALL match exactly, except that surrounding whitespace inside the braces SHALL be ignored. A reference to an unknown field SHALL render an inline error marker naming the field, instead of failing the card.

#### Scenario: Basic substitution
- **WHEN** the question template is `{{Front}}` and the note's Front is `hola`
- **THEN** the question renders as `hola`

#### Scenario: Unknown field
- **WHEN** a template references `{{Missing}}`
- **THEN** the card renders, with a visible marker saying field `Missing` was not found

### Requirement: FrontSide
In an answer template, `{{FrontSide}}` SHALL be replaced by the rendered question, excluding any audio.

#### Scenario: Answer includes front
- **WHEN** the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`
- **THEN** the answer shows the question, a divider, and then the Back field

### Requirement: Conditional sections
`{{#Field}}…{{/Field}}` SHALL render its content only when the field is non-empty after stripping HTML and whitespace. `{{^Field}}…{{/Field}}` SHALL render its content only when the field is empty.

#### Scenario: Optional hint omitted
- **WHEN** the template contains `{{#Hint}}Hint: {{Hint}}{{/Hint}}` and the Hint field is empty
- **THEN** no "Hint:" text appears

### Requirement: Field filters
The renderer SHALL support the filters `text:` (strip HTML), `hint:` (collapsed content revealed by tapping), `type:` (an answer input on the question side and a comparison on the answer side, as described in "Type-in-the-answer comparison"), and `cloze:`. Unknown filters SHALL be ignored, and the field content SHALL render unfiltered.

#### Scenario: Text filter
- **WHEN** the template has `{{text:Front}}` and Front is `<b>bold</b>`
- **THEN** the output is `bold` without markup

### Requirement: Furigana filters
The renderer SHALL support the filters `furigana:`, `kana:` and `kanji:` on field text that marks readings as `base[reading]`. The base SHALL be the characters directly before `[` back to the previous space, HTML tag or start of text, and one space directly before the base SHALL be removed. `furigana:` SHALL render each base with its reading as ruby text, `kana:` SHALL render only the readings in place of their bases, and `kanji:` SHALL render only the bases. Text without brackets SHALL pass through unchanged, and `[sound:…]` SHALL NOT be treated as a reading.

#### Scenario: Ruby text
- **WHEN** the template has `{{furigana:Expression}}` and Expression is `日本語[にほんご]を 勉強[べんきょう]する`
- **THEN** the card shows 日本語 with にほんご above it, followed by を, 勉強 with べんきょう above it, and する, with no brackets and no space before 勉強

#### Scenario: Kana only
- **WHEN** the template has `{{kana:Expression}}` for the same field
- **THEN** the card shows `にほんごをべんきょうする`

#### Scenario: Kanji only
- **WHEN** the template has `{{kanji:Expression}}` for the same field
- **THEN** the card shows `日本語を勉強する`

#### Scenario: Sound tag untouched
- **WHEN** the field is `猫[ねこ] [sound:neko.mp3]` and the template has `{{furigana:Expression}}`
- **THEN** only 猫 gets a reading and the sound tag is handled as a sound tag

### Requirement: Cloze deletions
For cloze card number N, `{{cloze:Field}}` SHALL show cloze N as `[...]` (or `[hint]` when written `{{cN::answer::hint}}`) on the question side, and highlight its answer on the answer side. Other cloze numbers SHALL show their answer text unhighlighted on both sides.

#### Scenario: Cloze question and answer
- **WHEN** card 1 of `{{c1::Paris}} is in {{c2::France}}` is rendered
- **THEN** the question shows `[...] is in France` and the answer shows `Paris is in France` with `Paris` highlighted

#### Scenario: Cloze hint
- **WHEN** card 1 of `{{c1::Paris::city}} is in France` is rendered
- **THEN** the question shows `[city] is in France`

### Requirement: Multi-number cloze deletions
A cloze deletion SHALL accept a comma-separated list of cloze numbers, written `{{c1,2::answer}}` or `{{c1,2::answer::hint}}`, with optional spaces after commas. For each listed number N, the deletion SHALL render on card N as a single-number deletion `{{cN::…}}` would. On cards whose number is not listed, it SHALL render as its plain answer text.

#### Scenario: Hidden on every listed card
- **WHEN** cards 1, 2 and 3 of `{{c1,2::Paris}} is in {{c3::France}}` are rendered
- **THEN** the questions of cards 1 and 2 show `[...] is in France`, and the question of card 3 shows `Paris is in [...]`

#### Scenario: Hint on a multi-number deletion
- **WHEN** card 2 of `{{c1,2::Paris::city}}` is rendered
- **THEN** the question shows `[city]`

#### Scenario: Malformed list is literal text
- **WHEN** a field contains `{{c1,,2::Paris}}`
- **THEN** the text is shown as written and no cloze is detected

### Requirement: Cloze-number conditionals
When rendering a card of a cloze note type, a section named `cN`, where N is a positive number and the note type has no field named `cN`, SHALL be a cloze-number condition: `{{#cN}}…{{/cN}}` SHALL render its content only on cloze card N, and `{{^cN}}…{{/cN}}` SHALL render its content on every other cloze card. This SHALL apply on both the question and the answer side. When the note type has a field named `cN`, the section SHALL test that field as usual.

#### Scenario: Content only on card 1
- **WHEN** the question template is `{{cloze:Text}}{{#c1}} (capital){{/c1}}` and cards 1 and 2 are rendered
- **THEN** card 1's question ends with ` (capital)` and card 2's question does not

#### Scenario: Inverted condition
- **WHEN** the template contains `{{^c2}}Not card 2{{/c2}}`
- **THEN** cards 1 and 3 show `Not card 2` and card 2 does not

#### Scenario: Field named c1 wins
- **WHEN** a cloze note type has a field `c1` that is empty and the template contains `{{#c1}}X{{/c1}}`
- **THEN** no card shows `X`

### Requirement: HTML display with styling
Rendered cards SHALL be displayed as HTML, supporting at least: text formatting, lists, tables, line breaks, links, colours, and the note type's CSS for font size, colour, alignment and background. Scripts SHALL NOT be executed.

#### Scenario: Script ignored
- **WHEN** a field contains `<script>alert(1)</script>Text`
- **THEN** only `Text` is displayed and no script runs

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

### Requirement: Theme-aware card display
Cards SHALL be readable in both light and dark app themes. In dark mode, the default card background and text colour SHALL follow the theme unless the note type's CSS sets them explicitly.

#### Scenario: Dark mode default
- **WHEN** the app is in dark mode and the note type CSS sets no colours
- **THEN** the card shows light text on a dark background

### Requirement: MathJax typesetting
Text between `\(` and `\)` SHALL be typeset as inline math, and text between `\[` and `\]` as display math on its own centred line. Typesetting SHALL work offline on every supported platform. HTML tags and entities inside the delimiters SHALL be ignored as markup.

#### Scenario: Inline formula
- **WHEN** a field contains `Area is \(\pi r^2\)`
- **THEN** the card shows "Area is" followed by a typeset π r² in the line

#### Scenario: Display formula
- **WHEN** a field contains `\[\frac{a}{b}\]`
- **THEN** the card shows a typeset fraction centred on its own line

### Requirement: Chemistry formulas
Inside math, `\ce{…}` SHALL be typeset as a chemical formula, supporting element symbols with subscript counts, charges, stoichiometric numbers, states such as `(aq)`, `+` between species and reaction arrows.

#### Scenario: Reaction
- **WHEN** a field contains `\(\ce{2H2 + O2 -> 2H2O}\)`
- **THEN** the card shows the reaction with subscript 2s, upright element symbols and a right arrow

### Requirement: Math errors and clozes
A formula that cannot be typeset SHALL be shown as its TeX source with an error style, and the rest of the card SHALL render. A cloze deletion inside math SHALL show as `[...]` within the formula on the question side and highlighted on the answer side.

#### Scenario: Unknown command
- **WHEN** a field contains `\(\unknowncommand{x}\)` and `text`
- **THEN** the card shows `\unknowncommand{x}` with an error style and `text` normally

#### Scenario: Cloze in a formula
- **WHEN** card 1 of `\(E = {{c1::mc^2}}\)` is rendered
- **THEN** the question shows `E = [...]` typeset and the answer shows `E = mc²` with `mc²` highlighted

### Requirement: Type-in-the-answer comparison
On the answer side, `{{type:Field}}` SHALL show the typed text compared with the field's text, with HTML stripped and surrounding whitespace ignored. Matching characters, wrong typed characters and missing expected characters SHALL be marked distinguishably. When nothing was typed, it SHALL show the expected text. Only the first `type:` filter of a card SHALL be active. When the answer template has no `type:` filter of its own, the comparison SHALL appear where the question's `type:` filter sits inside `{{FrontSide}}`, and an `<hr id=answer>` line on the answer SHALL move to just before it. When the answer template has its own, `{{FrontSide}}` SHALL show no input or comparison.

#### Scenario: Exact match
- **WHEN** the expected text is `perro` and the user typed `perro`
- **THEN** the answer shows `perro` marked as correct

#### Scenario: Typo
- **WHEN** the expected text is `perro` and the user typed `pero`
- **THEN** the answer shows the typed text and the expected text, with the missing `r` marked

#### Scenario: Preview without input
- **WHEN** a card with `{{type:Back}}` is shown in the browser details
- **THEN** no input is shown and the answer shows the expected text

#### Scenario: Type filter only on the question
- **WHEN** the question template is `{{Front}} {{type:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`, Front is `dog`, Back is `perro` and the user typed `pero`
- **THEN** the answer shows `dog`, then the answer line, then the comparison of `pero` with `perro`, then `perro`

#### Scenario: Question-only type filter keeps its modifiers
- **WHEN** the question template has `{{type:nc:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{Back}}`, Back is `über` and the user typed `uber`
- **THEN** the comparison shows the answer as correct

#### Scenario: Question-only type filter in a preview
- **WHEN** a card whose question template is `{{Front}} {{type:Back}}` and whose answer template is `{{FrontSide}}<hr id=answer>{{Back}}` is shown in the browser details
- **THEN** the answer shows the expected text below the answer line, and the browser's answer column includes it

#### Scenario: Type filter on both sides
- **WHEN** the question template is `{{Front}} {{type:Back}}`, the answer template is `{{FrontSide}}<hr id=answer>{{type:Back}}` and the user typed `perro`
- **THEN** the answer shows the comparison once, below the answer line

### Requirement: Type-in variants
`{{type:cloze:Field}}` SHALL compare against the answer of the active cloze number, joining several deletions of that number with `, `. `{{type:nc:Field}}` SHALL ignore accents and other combining marks on both sides when comparing.

#### Scenario: Cloze answer
- **WHEN** card 1 of `{{c1::Paris}} is in {{c2::France}}` uses `{{type:cloze:Text}}` and the user typed `Paris`
- **THEN** the comparison shows `Paris` as correct

#### Scenario: Ignore accents
- **WHEN** the template uses `{{type:nc:Back}}`, Back is `über` and the user typed `uber`
- **THEN** the comparison shows the answer as correct

### Requirement: Readable type-answer comparison
Every part of the type-in-the-answer comparison SHALL be readable in light and dark mode. When the colour a part gets from the note type's CSS has too little contrast with the background it is drawn on, the part SHALL use the card's text colour instead. If that colour is also too faint there, it SHALL use black or white, whichever contrasts more. Colours with enough contrast SHALL be kept.

#### Scenario: Black mark without a valid highlight in dark mode
- **WHEN** the app is in dark mode, the note type CSS has `.typeGood { color: black; background-color: light green; }`, the expected text is `ka` and the user typed `ka`
- **THEN** `ka` shows in the card's text colour, readable on the dark card

#### Scenario: Black mark on its own highlight is kept
- **WHEN** the app is in dark mode, the note type CSS has `.typeBad { color: black; background-color: red; }` and the user typed a wrong character
- **THEN** the wrong character shows black on red

#### Scenario: Unchanged in light mode
- **WHEN** the app is in light mode, the note type CSS has `.typeGood { color: black; }` and the user typed the expected text
- **THEN** the comparison shows black on the light card

#### Scenario: Card colour also too faint
- **WHEN** the note type CSS has `.typeGood { color: #ddd; background-color: #eee; }` and `.card { color: white; }`, and the user typed the expected text
- **THEN** the comparison shows black on the `#eee` highlight

#### Scenario: Rest of the card untouched
- **WHEN** the app is in dark mode and the note type CSS gives a `.hint` element outside the comparison `color: black`
- **THEN** that element still shows black

### Requirement: Card root classes
The rendered card SHALL carry the classes `card` and `cardN`, where N is the card type number starting at 1 (the cloze number for cloze cards), `nightMode` and `night_mode` while the app shows the dark theme, and the class or classes of the current platform: `win`, `mac`, `linux`, `android`, `iphone`, `ipad`, and `mobile` on phones and tablets. CSS rules selecting these classes, on their own or followed by a descendant selector, SHALL apply.

#### Scenario: Reverse card styling
- **WHEN** the CSS contains `.card2 { color: red }` and the reverse card of a `Basic (and reversed card)` note is shown
- **THEN** the reverse card's text is red and the forward card's is not

#### Scenario: Night mode rule
- **WHEN** the CSS contains `.nightMode .cloze { color: lightblue }` and the app is in dark mode
- **THEN** cloze highlights are light blue, and in light mode they use the other rules

#### Scenario: Phone rule
- **WHEN** the CSS contains `.mobile .card { font-size: 30px }` and the card is shown on an Android phone
- **THEN** the card text is 30 px

### Requirement: Environment media queries
`@media` rules SHALL apply when their conditions match: `prefers-color-scheme` SHALL follow the app's current theme, and `min-width`, `max-width` and `orientation` SHALL be evaluated against the card's display area. Unsupported conditions SHALL make the rule not apply. When the theme or display width changes, the shown card SHALL update to the matching rules.

#### Scenario: Dark scheme query
- **WHEN** the CSS contains `@media (prefers-color-scheme: dark) { .card { background: black } }` and the user switches the app to dark mode while a card is shown
- **THEN** the card background turns black without leaving the card

### Requirement: CardFlag special field
`{{CardFlag}}` SHALL render `flag1` to `flag7` for a card with that flag, and nothing for a card without a flag.

#### Scenario: Flag as CSS class
- **WHEN** the question template is `<div class="{{CardFlag}}">{{Front}}</div>` and the card has flag 3
- **THEN** the question renders as `<div class="flag3">` around the Front field

### Requirement: Text-to-speech tags
`{{tts <lang>:Field}}` and `[anki:tts lang=<lang>]text[/anki:tts]` SHALL produce an utterance of the text, with HTML stripped, in language `<lang>` (written like `en_US`), shown as a replay button where the tag is. The options `voices=` (a comma-separated preference list) and `speed=` SHALL be honoured when the device's speech engine supports them. `{{tts <lang>:cloze:Field}}` SHALL read the active deletion as "blank" on the question side, and `{{tts <lang>:cloze-only:Field}}` SHALL read only the active deletion's text.

#### Scenario: Field spoken
- **WHEN** the question template is `{{Front}}{{tts es_ES:Front}}` and Front is `perro`
- **THEN** the question shows `perro` and a replay button, and the utterance is "perro" in Spanish

#### Scenario: Cloze only
- **WHEN** card 1 of `{{c1::Paris}} is in France` has `{{tts en_US:cloze-only:Text}}` on the answer side
- **THEN** the answer's utterance is "Paris"

### Requirement: Voice list
`{{tts-voices:}}` SHALL render the list of speech voices available on the device, with their language codes.

#### Scenario: Show voices
- **WHEN** a template contains `{{tts-voices:}}`
- **THEN** the card lists the device's voices, for example `en_US Samantha`

### Requirement: Fonts from media
An `@font-face` rule in the note type's CSS whose `src` names a `.ttf` or `.otf` file in the collection's media SHALL make that font available under the rule's `font-family` name, and declarations using that name SHALL display text in that font. Fonts in other formats, remote URLs and missing files SHALL be ignored, and text SHALL use the next listed family or the default font.

#### Scenario: Deck font applied
- **WHEN** the CSS contains `@font-face { font-family: jp; src: url("_jp.ttf") }` and `.card { font-family: jp }`, and `_jp.ttf` is in the media
- **THEN** the card text is displayed in that font

#### Scenario: Missing font file
- **WHEN** the CSS names `_missing.ttf` in `@font-face` and the file is not in the media
- **THEN** the card is shown in the default font without an error

#### Scenario: Unsupported format
- **WHEN** the `src` is `url("_font.woff2")`
- **THEN** the rule is ignored and the default font is used

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

### Requirement: Pre-generated LaTeX images
`[latex]…[/latex]`, `[$]…[/$]` and `[$$]…[/$$]` SHALL be replaced by the media image Anki would have generated for that source, found by the same file name Anki uses (PNG, or SVG when the note type is set to SVG). When that image is not in the collection's media, the LaTeX source SHALL be shown as preformatted text. SVG images SHALL display on every platform.

#### Scenario: Image shipped with the deck
- **WHEN** a card from an imported deck contains `[$]x^2[/$]` and the deck's media includes the image Anki generated for it
- **THEN** the card shows that image

#### Scenario: Image missing
- **WHEN** a card contains `[latex]\LaTeX[/latex]` and no matching image exists
- **THEN** the card shows `\LaTeX` as preformatted text and no broken image

### Requirement: Occlusion masks
A card whose cloze field contains image occlusion shapes SHALL show the note's image with the shapes drawn over it at their relative positions, scaling with the image, without running template scripts. In "Hide all, guess one" mode, every shape SHALL be masked and the asked shape SHALL be highlighted on the question. In "Hide one, guess one" mode, only the asked shape SHALL be masked. On the answer, the asked shape SHALL be revealed and outlined.

#### Scenario: Hide all, guess one
- **WHEN** card 2 of a note with three shapes in "Hide all, guess one" mode is shown
- **THEN** the question shows all three areas masked, with shape 2 in the highlight colour

#### Scenario: Answer reveal
- **WHEN** the answer of that card is shown
- **THEN** the area under shape 2 is visible with an outline, and shapes 1 and 3 stay masked

### Requirement: Right-to-left text
Elements with `dir="rtl"` in templates or fields SHALL be laid out right to left, with right alignment by default.

#### Scenario: RTL template
- **WHEN** the question template is `<div dir="rtl">{{Arabic}}</div>`
- **THEN** the Arabic text is laid out right to left and aligned right
