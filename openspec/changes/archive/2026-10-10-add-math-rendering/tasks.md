# Tasks

## 1. Go rendering

- [x] 1.1 Add `golang/render/math.go` that finds balanced `\(…\)` and `\[…\]` outside tags in rendered HTML and replaces them with `<cf-math display=…>` holding the escaped TeX, with inner tags removed and entities decoded. Allow `cf-math` in `sanitize.go`. Add `math_test.go` cases: inline, display, `<br>` and `&nbsp;` inside a formula, unbalanced delimiters left as text. Verify `go test ./render/...` passes.
- [x] 1.2 Make cloze rendering inside math produce `\text{[...]}` on the question and a coloured group on the answer. Add `math_test.go` cases for the cloze scenario. Verify they pass.
- [x] 1.3 Add `golang/render/mhchem.go` translating the `\ce{}` subset from design decision 3 and `\pu{}`. Add `mhchem_test.go` with water, sulfate ion, a reaction with states and an equilibrium arrow with text. Verify `go test ./render/...` passes.

## 2. Flutter display

- [x] 2.1 Add `flutter_math_fork` to `pubspec.yaml` and verify `fvm flutter pub get` and `fluffy build web` succeed.
- [x] 2.2 Map `cf-math` to `Math.tex` in `lib/widgets/card_view.dart` (inline vs block, theme text colour, horizontal scroll for block), with the error fallback. Add `test/card_view_test.dart` tests: a fraction renders a `Math` widget, an unknown command shows the source text. Verify they pass.

## 3. Editor menu

- [x] 3.1 Add the Fx menu and the Ctrl/Cmd chord shortcuts to `note_editor_screen.dart`, wrapping the selection in the focused field. Add `test/editor_test.dart` tests for wrapping a selection and for an empty insert placing the cursor inside. Verify they pass.

## 4. Integration check

- [x] 4.1 Run `go test ./...` in `golang/`, `fvm flutter analyze` and `fvm flutter test`, and verify all pass. Import a chemistry/physics `.apkg` and check formulas on desktop, web and Android.
