# Tasks

## 1. Environment in Go

- [x] 1.1 Add `render.Env` and thread it through `RenderCard`; add `AppService.SetRenderEnvironment(env)` storing it on the app and broadcasting a settings-like change. Add `svc_test.go` coverage that rendering after setting night mode differs for a `.nightMode` rule. Run `fluffy generate` and verify `--check` is clean.
- [x] 1.2 Extend `golang/render/css.go` to match root-class compounds on the card and as ancestor prefixes, with class-count specificity (design decision 2). Add `css_test.go` cases for `.card2`, `.card.nightMode`, `.nightMode .cloze`, `.mobile .card` and a non-matching platform. Verify `go test ./render/...` passes.
- [x] 1.3 Evaluate `@media` blocks per design decision 3. Add `css_test.go` cases for `prefers-color-scheme`, `max-width` above and below the width, `orientation`, a comma list, and an unknown feature. Verify they pass.

## 2. Flutter

- [x] 2.1 Compute the environment (brightness from the effective theme, platform per design decision 4, width bucket from the card view's `LayoutBuilder`) and call `app.setRenderEnvironment` at startup and on change. Add the call to `test/support/fake_go.dart`. Add a `test/card_view_test.dart` test that switching theme triggers a re-request and swaps the HTML without flashing empty. Verify it passes.

## 3. Integration check

- [x] 3.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a shared deck that styles `.nightMode` and `.card2`, and check light and dark mode on desktop and a phone.
