import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Asks before deleting the notes of [cardIds], counting every card that goes
/// with them, including the ones the user did not pick. [impact] skips asking
/// Go when the caller already knows it. Returns what is deleted when the
/// user confirmed, else null.
Future<golang.DeleteImpact?> confirmDeleteCards(
  BuildContext context,
  List<int> cardIds, {
  golang.DeleteImpact? impact,
}) async {
  if (impact == null) {
    try {
      impact = await golang.browse.deleteImpact(cardIds);
    } catch (e) {
      showError(e);
      return null;
    }
    if (!context.mounted) return null;
  }
  final others = impact.cards - cardIds.length;
  final l = context.l10n;
  final including = others > 0
      ? impact.notes == 1
            ? l.includingOtherCardsNote(others)
            : l.includingOtherCardsNotes(others)
      : '';
  final ok = await confirm(
    context,
    title: l.deleteCardsTitle(impact.cards),
    message: '$including${l.canUndoUntilClose}',
    action: l.delete,
    destructive: true,
  );
  return ok ? impact : null;
}
