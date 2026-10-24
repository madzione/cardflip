# Spec Delta

## ADDED Requirements

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
