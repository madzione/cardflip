import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// Asks how to reset [cards] cards to new, with "Restore original position
/// where possible" on and "Reset repetition and lapse counts" off. Returns
/// null when cancelled.
Future<golang.ResetOptions?> askReset(
  BuildContext context, {
  required int cards,
}) => showDialog<golang.ResetOptions>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => _ResetDialog(cards: cards),
);

class _ResetDialog extends StatefulWidget {
  const _ResetDialog({required this.cards});

  final int cards;

  @override
  State<_ResetDialog> createState() => _ResetDialogState();
}

class _ResetDialogState extends State<_ResetDialog> {
  bool _restorePosition = true;
  bool _resetCounts = false;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return AlertDialog(
      title: Text(context.l10n.resetCardsTitle(widget.cards)),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(context.l10n.resetCardsBody),
          const SizedBox(height: 8),
          CheckboxListTile(
            key: const Key('reset-restore-position'),
            contentPadding: EdgeInsets.zero,
            title: Text(context.l10n.restorePosition),
            value: _restorePosition,
            onChanged: (v) => setState(() => _restorePosition = v!),
          ),
          CheckboxListTile(
            key: const Key('reset-counts'),
            contentPadding: EdgeInsets.zero,
            title: Text(context.l10n.resetCounts),
            value: _resetCounts,
            onChanged: (v) => setState(() => _resetCounts = v!),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          style: FilledButton.styleFrom(
            backgroundColor: scheme.error,
            foregroundColor: scheme.onError,
          ),
          onPressed: () => Navigator.pop(
            context,
            golang.ResetOptions(
              restorePosition: _restorePosition,
              resetCounts: _resetCounts,
            ),
          ),
          child: Text(context.l10n.reset),
        ),
      ],
    );
  }
}
