import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter/services.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Ctrl+Alt+B, or Cmd+Option+B on macOS: create a backup now.
SingleActivator get backupShortcut =>
    defaultTargetPlatform == TargetPlatform.macOS
    ? const SingleActivator(LogicalKeyboardKey.keyB, meta: true, alt: true)
    : const SingleActivator(LogicalKeyboardKey.keyB, control: true, alt: true);

/// How [backupShortcut] is written on this platform.
String get backupShortcutLabel =>
    defaultTargetPlatform == TargetPlatform.macOS ? '⌘⌥B' : '$_ctrl+$_alt+B';

bool get _mac => defaultTargetPlatform == TargetPlatform.macOS;

/// Whether menus, tooltips and settings show keyboard shortcuts next to
/// their labels: not on Android and iOS, where a hardware keyboard is rare.
/// On the web the platform follows the browser, so a phone browser counts
/// as mobile. The shortcuts work everywhere.
bool get showShortcutHints =>
    defaultTargetPlatform != TargetPlatform.android &&
    defaultTargetPlatform != TargetPlatform.iOS;

/// [label] followed by its [shortcut] in parentheses, or [label] alone
/// where [showShortcutHints] is false.
String withShortcut(String label, String shortcut) =>
    showShortcutHints ? appL10n.labelWithShortcut(label, shortcut) : label;

// The names of keys in shortcut labels, in the interface language (German
// keyboards say Strg for Ctrl). The Mac labels use the key symbols.
String get _ctrl => appL10n.keyCtrl;
String get _shift => appL10n.keyShift;
String get _alt => appL10n.keyAlt;

/// Ctrl+Z, or Cmd+Z on macOS: undo.
List<SingleActivator> get undoShortcuts => [
  SingleActivator(LogicalKeyboardKey.keyZ, control: !_mac, meta: _mac),
];

/// Ctrl+Shift+Z or Ctrl+Y, or Cmd+Shift+Z or Cmd+Y on macOS: redo.
List<SingleActivator> get redoShortcuts => [
  SingleActivator(
    LogicalKeyboardKey.keyZ,
    control: !_mac,
    meta: _mac,
    shift: true,
  ),
  SingleActivator(LogicalKeyboardKey.keyY, control: !_mac, meta: _mac),
];

/// Ctrl+Enter, or Cmd+Enter on macOS: add the card in the add screen, save
/// the note in the edit screen.
List<SingleActivator> get submitShortcuts => [
  SingleActivator(LogicalKeyboardKey.enter, control: !_mac, meta: _mac),
  SingleActivator(LogicalKeyboardKey.numpadEnter, control: !_mac, meta: _mac),
];

/// How [submitShortcuts] are written on this platform.
String get submitShortcutLabel => _mac ? '⌘↩' : '$_ctrl+${appL10n.keyEnter}';

/// Ctrl+Shift+C, or Cmd+Shift+C on macOS: wrap the selection in the next
/// cloze number.
SingleActivator get clozeShortcut => SingleActivator(
  LogicalKeyboardKey.keyC,
  control: !_mac,
  meta: _mac,
  shift: true,
);

/// Ctrl+Alt+Shift+C, or Cmd+Option+Shift+C on macOS: wrap the selection in
/// the highest cloze number already in the note.
SingleActivator get clozeSameShortcut => SingleActivator(
  LogicalKeyboardKey.keyC,
  control: !_mac,
  meta: _mac,
  shift: true,
  alt: true,
);

/// How [clozeShortcut] and [clozeSameShortcut] are written on this
/// platform.
String get clozeShortcutLabel => _mac ? '⇧⌘C' : '$_ctrl+$_shift+C';
String get clozeSameShortcutLabel => _mac ? '⌥⇧⌘C' : '$_ctrl+$_alt+$_shift+C';

/// Ctrl+Alt+I, or Cmd+Option+I on macOS: card info of the previously
/// answered card.
SingleActivator get previousCardInfoShortcut => SingleActivator(
  LogicalKeyboardKey.keyI,
  control: !_mac,
  meta: _mac,
  alt: true,
);

/// How [previousCardInfoShortcut] is written on this platform.
String get previousCardInfoShortcutLabel => _mac ? '⌘⌥I' : '$_ctrl+$_alt+I';

/// How the shortcut of flag 1 to 7 (Ctrl+N, or Cmd+N on macOS) is written
/// on this platform.
String flagShortcutLabel(int flag) => _mac ? '⌘$flag' : '$_ctrl+$flag';

/// How undo and redo shortcuts are written on this platform.
String get undoShortcutLabel => _mac ? '⌘Z' : '$_ctrl+Z';
String get redoShortcutLabel => _mac ? '⇧⌘Z' : '$_ctrl+$_shift+Z';

/// Ctrl+Alt+S, or Cmd+Option+S on macOS: invert the browser's selection.
SingleActivator get invertSelectionShortcut => SingleActivator(
  LogicalKeyboardKey.keyS,
  control: !_mac,
  meta: _mac,
  alt: true,
);

/// How [invertSelectionShortcut] is written on this platform.
String get invertSelectionShortcutLabel => _mac ? '⌘⌥S' : '$_ctrl+$_alt+S';

/// Ctrl+F, or Cmd+F on macOS: focus the browser's search field.
SingleActivator get findShortcut =>
    SingleActivator(LogicalKeyboardKey.keyF, control: !_mac, meta: _mac);

/// Ctrl+Shift+F, or Cmd+Shift+F on macOS: focus the browser's list.
SingleActivator get focusListShortcut => SingleActivator(
  LogicalKeyboardKey.keyF,
  control: !_mac,
  meta: _mac,
  shift: true,
);

/// Delete, and on macOS also the Delete key above Return (Backspace):
/// delete the browser's selected notes.
List<SingleActivator> get deleteRowsShortcuts => [
  const SingleActivator(LogicalKeyboardKey.delete),
  if (_mac) const SingleActivator(LogicalKeyboardKey.backspace),
];

/// Ctrl+Shift+D, or Cmd+Shift+D on macOS: set the due date of the card
/// being studied or the browser's selection.
SingleActivator get setDueDateShortcut => SingleActivator(
  LogicalKeyboardKey.keyD,
  control: !_mac,
  meta: _mac,
  shift: true,
);

/// How [setDueDateShortcut] is written on this platform.
String get setDueDateShortcutLabel => _mac ? '⇧⌘D' : '$_ctrl+$_shift+D';

/// Ctrl+Delete, or Cmd+Delete on macOS (also with the Delete key above
/// Return): delete the note of the card being studied.
List<SingleActivator> get deleteNoteShortcuts => [
  SingleActivator(LogicalKeyboardKey.delete, control: !_mac, meta: _mac),
  if (_mac) const SingleActivator(LogicalKeyboardKey.backspace, meta: true),
];

/// How [deleteNoteShortcuts] are written on this platform.
String get deleteNoteShortcutLabel =>
    _mac ? '⌘⌫' : '$_ctrl+${appL10n.keyDelete}';

/// `/`: open the deck picker ("Study deck"). On layouts where `/` needs
/// Shift (German), the character counts, not the key.
const pickDeckShortcut = CharacterActivator('/');

// How the study screen's single-key shortcuts are written. Shift+V and
// Shift+A use the interface language's name for Shift.
const editCardShortcutLabel = 'E';
const replayAudioShortcutLabel = 'R';
String get recordVoiceShortcutLabel => '$_shift+V';
const playVoiceShortcutLabel = 'V';
const markNoteShortcutLabel = '*';
const cardInfoShortcutLabel = 'I';
const buryCardShortcutLabel = '-';
const buryNoteShortcutLabel = '=';
const suspendCardShortcutLabel = '@';
const suspendNoteShortcutLabel = '!';
String get autoAdvanceShortcutLabel => '$_shift+A';
const deckOptionsShortcutLabel = 'O';

/// How [pickDeckShortcut] is written.
const pickDeckShortcutLabel = '/';

/// How Esc, which leaves the study screen, is written.
String get escShortcutLabel => appL10n.keyEsc;

/// "Ctrl" or "Cmd", for the field editor's shortcuts.
String get commandKey => _mac ? 'Cmd' : _ctrl;

// How the field editor's formatting shortcuts (Ctrl, or Cmd on macOS) are
// written.
String get boldShortcutLabel => '$commandKey+B';
String get italicShortcutLabel => '$commandKey+I';
String get underlineShortcutLabel => '$commandKey+U';
String get subscriptShortcutLabel => '$commandKey+=';
String get superscriptShortcutLabel => '$commandKey+$_shift+=';
String get removeFormattingShortcutLabel => '$commandKey+R';

/// How the field editor's HTML/visual toggle (Ctrl/Cmd+Shift+X) is written.
String get htmlToggleShortcutLabel => '$commandKey+$_shift+X';
