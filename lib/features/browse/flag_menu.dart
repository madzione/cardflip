import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/l10n/l10n.dart';

/// A button with a menu of the seven flags, by their names, and "No flag".
/// It calls [onSelected] with 1 to 7, or 0 for "No flag".
class FlagMenuButton extends StatelessWidget {
  const FlagMenuButton({
    super.key,
    required this.onSelected,
    this.current = 0,
    this.enabled = true,
  });

  final ValueChanged<int> onSelected;

  /// The flag of the card the menu is for, shown on the button and checked
  /// in the menu; 0 for none or several cards.
  final int current;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final settings = AppScope.of(context).settings;
    final colors = FlagColors.of(context);
    return PopupMenuButton<int>(
      popUpAnimationStyle: popupAnimation(context),
      tooltip: context.l10n.flag,
      enabled: enabled,
      icon: current == 0
          ? const Icon(Icons.outlined_flag)
          : Icon(Icons.flag, color: colors.flag(current)),
      onSelected: onSelected,
      itemBuilder: (_) => [
        for (var n = 1; n <= flagCount; n++)
          CheckedPopupMenuItem(
            key: Key('flag-item-$n'),
            value: n,
            checked: current == n,
            child: Row(
              children: [
                Icon(Icons.flag, color: colors.flag(n)),
                const SizedBox(width: 12),
                Flexible(child: Text(settings.flagName(n))),
              ],
            ),
          ),
        const PopupMenuDivider(),
        PopupMenuItem(
          key: const Key('flag-item-0'),
          value: 0,
          child: Text(context.l10n.noFlag),
        ),
      ],
    );
  }
}
