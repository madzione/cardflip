import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/l10n/l10n.dart';

/// The math menu's entries and their shortcuts, as in Anki.
enum MathInsert {
  inline(r'\(', r'\)', LogicalKeyboardKey.keyM, LogicalKeyboardKey.keyM),
  block(r'\[', r'\]', LogicalKeyboardKey.keyM, LogicalKeyboardKey.keyE),
  chemistry(
    r'\(\ce{',
    r'}\)',
    LogicalKeyboardKey.keyM,
    LogicalKeyboardKey.keyC,
  ),
  latex(
    '[latex]',
    '[/latex]',
    LogicalKeyboardKey.keyT,
    LogicalKeyboardKey.keyT,
  ),
  latexEquation(
    r'[$]',
    r'[/$]',
    LogicalKeyboardKey.keyT,
    LogicalKeyboardKey.keyE,
  ),
  latexMath(
    r'[$$]',
    r'[/$$]',
    LogicalKeyboardKey.keyT,
    LogicalKeyboardKey.keyM,
  );

  const MathInsert(this.open, this.close, this.chord, this.second);

  final String open;
  final String close;
  final LogicalKeyboardKey chord;
  final LogicalKeyboardKey second;

  /// The entry's name in the menu.
  String label(AppLocalizations l) => switch (this) {
    inline => l.mathInline,
    block => l.mathBlock,
    chemistry => l.mathChemistry,
    latex => l.mathLatex,
    latexEquation => l.mathLatexEquation,
    latexMath => l.mathLatexMath,
  };

  String get shortcut {
    final mod = defaultTargetPlatform == TargetPlatform.macOS
        ? 'Cmd'
        : appL10n.keyCtrl;
    return '$mod+${chord.keyLabel.toUpperCase()}, '
        '${second.keyLabel.toUpperCase()}';
  }
}

/// Wraps the selection of [c] in [open] and [close], leaving the cursor
/// before [close]. Without a selection the pair goes at the cursor, or at
/// the end of the text.
void wrapSelection(TextEditingController c, String open, String close) {
  final sel = c.selection.isValid
      ? c.selection
      : TextSelection.collapsed(offset: c.text.length);
  final inner = sel.textInside(c.text);
  c.value = TextEditingValue(
    text: '${sel.textBefore(c.text)}$open$inner$close${sel.textAfter(c.text)}',
    selection: TextSelection.collapsed(
      offset: sel.start + open.length + inner.length,
    ),
  );
}

/// The Fx button: a menu of [MathInsert] entries with their shortcuts.
class MathMenuButton extends StatelessWidget {
  const MathMenuButton({super.key, required this.onSelected});

  final ValueChanged<MathInsert> onSelected;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<MathInsert>(
      popUpAnimationStyle: popupAnimation(context),
      tooltip: context.l10n.math,
      icon: const Icon(Icons.functions),
      onSelected: onSelected,
      itemBuilder: (context) => [
        for (final m in MathInsert.values)
          PopupMenuItem(
            value: m,
            child: MenuEntryLabel(m.label(context.l10n), m.shortcut),
          ),
      ],
    );
  }
}

/// A menu entry's [label] with its [shortcut] after it, dimmed, and an
/// optional [leading] widget such as an icon before it. The shortcut is
/// left out where [showShortcutHints] is false.
class MenuEntryLabel extends StatelessWidget {
  const MenuEntryLabel(this.label, this.shortcut, {super.key, this.leading});

  final String label;
  final String shortcut;
  final Widget? leading;

  @override
  Widget build(BuildContext context) => Row(
    children: [
      if (leading case final leading?) ...[leading, const SizedBox(width: 12)],
      Expanded(child: Text(label)),
      if (showShortcutHints) ...[
        const SizedBox(width: 24),
        Text(
          shortcut,
          style: TextStyle(color: Theme.of(context).colorScheme.outline),
        ),
      ],
    ],
  );
}

/// Handles the math chords Ctrl+M M/E/C and Ctrl+T T/E/M (Cmd on macOS)
/// for the fields below it.
class MathShortcuts extends StatefulWidget {
  const MathShortcuts({super.key, required this.onInsert, required this.child});

  final ValueChanged<MathInsert> onInsert;
  final Widget child;

  @override
  State<MathShortcuts> createState() => _MathShortcutsState();
}

class _MathShortcutsState extends State<MathShortcuts> {
  /// The first key of a chord (M or T) while waiting for the second.
  LogicalKeyboardKey? _chord;
  Timer? _timer;

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  KeyEventResult _onKey(FocusNode node, KeyEvent event) {
    if (event is! KeyDownEvent) return KeyEventResult.ignored;
    final key = event.logicalKey;
    final first = _chord;
    if (first != null) {
      _chord = null;
      _timer?.cancel();
      for (final m in MathInsert.values) {
        if (m.chord == first && m.second == key) {
          widget.onInsert(m);
          return KeyEventResult.handled;
        }
      }
      return KeyEventResult.ignored;
    }
    final keyboard = HardwareKeyboard.instance;
    final command = defaultTargetPlatform == TargetPlatform.macOS
        ? keyboard.isMetaPressed
        : keyboard.isControlPressed;
    if (command &&
        (key == LogicalKeyboardKey.keyM || key == LogicalKeyboardKey.keyT)) {
      _chord = key;
      _timer?.cancel();
      _timer = Timer(const Duration(seconds: 2), () => _chord = null);
      return KeyEventResult.handled;
    }
    return KeyEventResult.ignored;
  }

  @override
  Widget build(BuildContext context) =>
      Focus(onKeyEvent: _onKey, child: widget.child);
}
