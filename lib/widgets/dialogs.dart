import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Asks for a line of text. Returns null when cancelled.
Future<String?> askText(
  BuildContext context, {
  required String title,
  String initial = '',
  String? label,
  String? helper,
  String? action,
}) {
  final controller = TextEditingController(text: initial);
  return showDialog<String>(
    context: context,
    animationStyle: popupAnimation(context),
    builder: (context) => AlertDialog(
      title: Text(title),
      content: TextField(
        controller: controller,
        autofocus: true,
        decoration: InputDecoration(labelText: label, helperText: helper),
        onSubmitted: (v) => Navigator.pop(context, v),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, controller.text),
          child: Text(action ?? context.l10n.save),
        ),
      ],
    ),
  );
}

/// Asks the user to confirm. Destructive actions get an error-coloured
/// button.
Future<bool> confirm(
  BuildContext context, {
  required String title,
  required String message,
  required String action,
  bool destructive = false,
}) async {
  final result = await showDialog<bool>(
    context: context,
    animationStyle: popupAnimation(context),
    builder: (context) {
      final scheme = Theme.of(context).colorScheme;
      return AlertDialog(
        // Long messages scroll on small screens; the buttons stay.
        scrollable: true,
        title: Text(title),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            style: destructive
                ? FilledButton.styleFrom(
                    backgroundColor: scheme.error,
                    foregroundColor: scheme.onError,
                  )
                : null,
            onPressed: () => Navigator.pop(context, true),
            child: Text(action),
          ),
        ],
      );
    },
  );
  return result ?? false;
}

/// Asks for a line of text and saves it with [save] before closing. When
/// saving fails, the dialog stays open and shows the error under the field,
/// so the text can be corrected. [message] is shown above the field.
/// Returns the trimmed text that was saved, or null when cancelled.
Future<String?> askAndSave(
  BuildContext context, {
  required String title,
  required Future<void> Function(String text) save,
  String? message,
  String initial = '',
  String? label,
  String? helper,
  String? hint,
  String? action,
  Key? fieldKey,
}) => showDialog<String>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => _SaveTextDialog(
    title: title,
    save: save,
    message: message,
    initial: initial,
    label: label,
    helper: helper,
    hint: hint,
    action: action,
    fieldKey: fieldKey,
  ),
);

class _SaveTextDialog extends StatefulWidget {
  const _SaveTextDialog({
    required this.title,
    required this.save,
    required this.message,
    required this.initial,
    required this.label,
    required this.helper,
    required this.hint,
    required this.action,
    required this.fieldKey,
  });

  final String title;
  final Future<void> Function(String text) save;
  final String? message;
  final String initial;
  final String? label;
  final String? helper;
  final String? hint;
  final String? action;
  final Key? fieldKey;

  @override
  State<_SaveTextDialog> createState() => _SaveTextDialogState();
}

class _SaveTextDialogState extends State<_SaveTextDialog> {
  late final _text = TextEditingController(text: widget.initial);
  String? _error;
  bool _saving = false;

  @override
  void dispose() {
    _text.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (_saving) return;
    setState(() => _saving = true);
    try {
      await widget.save(_text.text);
      if (mounted) Navigator.pop(context, _text.text.trim());
    } catch (e) {
      if (!mounted) return;
      setState(() {
        // Go's search errors name the term; keep its spelling (e.g.
        // "is:bogus: …").
        _error = e is golang.GolangException
            ? goText(e.message)
            : errorMessage(e);
        _saving = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.title),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (widget.message case final message?) ...[
            Text(message),
            const SizedBox(height: 12),
          ],
          TextField(
            key: widget.fieldKey,
            controller: _text,
            autofocus: true,
            decoration: InputDecoration(
              labelText: widget.label,
              helperText: widget.helper,
              hintText: widget.hint,
              errorText: _error,
              errorMaxLines: 3,
            ),
            onChanged: (_) {
              if (_error != null) setState(() => _error = null);
            },
            onSubmitted: (_) => _save(),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: _saving ? null : _save,
          child: Text(widget.action ?? context.l10n.save),
        ),
      ],
    );
  }
}
