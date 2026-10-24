# Design

## Context

- `golang/render/css.go` inlines the note type's CSS because the HTML widget ignores `<style>`. It supports `.card`, element, class and element-with-class selectors, optionally prefixed with `.card `. At-rules are skipped entirely.
- Dark mode currently only changes the default card colours when the CSS sets none.
- Rendering happens in Go and is cached per card for the browser question column.

## Goals / Non-Goals

**Goals:**
- Same matching rules on every platform; environment passed explicitly so rendering stays a pure function of note, template, CSS and environment.

**Non-Goals:**
- General descendant combinators (see `broaden-css-support`); only the root classes act as ancestors here.

## Decisions

### 1. A render environment set by Dart

`render.Env{Night bool, Platform string, WidthPx int}`. Dart calls `AppService.SetRenderEnvironment(env)` at startup and whenever theme brightness, or the card view's width bucket, changes; Go stores it on the app and uses it for every render; the change is broadcast like a settings change so open card views re-request their HTML. Width is bucketed (for example to 50 px) to avoid re-rendering on every resize frame.

Alternative: return light and dark variants on every render. Rejected: doubles rendering work and does not cover platform or width.

### 2. Root classes as part of the `.card` match

A selector whose first compound consists only of root classes (`card`, `cardN`, `nightMode`, `night_mode`, platform classes) matches the root when all its classes are present in the environment; when followed by a descendant part, the rest is matched as today's `.card X` prefix. Specificity counts every class.

### 3. `@media`

Evaluate `prefers-color-scheme`, `min-width`, `max-width`, `orientation`, `screen`, `all` and `not`/`and`/`,` combinations against the environment; rules inside a matching block are added in source order; unknown features make the query false.

### 4. Platform names

Linux desktop `linux`, Windows `win`, macOS `mac`, Android `android mobile`, iPhone `iphone mobile`, iPad `ipad mobile`, web: none of the OS classes except `mobile` when the browser reports a mobile user agent. Web gets no OS class because the same deck on a desktop browser should look like the desktop app, not like a specific OS.

## Risks / Trade-offs

- [Question cache in `core/search.go` would hold HTML for one environment] → The cache stores plain text without CSS, so it is unaffected; only full renders use the environment.
- [Re-render flicker when the theme changes] → The card view keeps the old HTML until the new one arrives.
