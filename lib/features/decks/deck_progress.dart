import 'package:flutter/material.dart';

import 'package:cardflip/features/stats/stats_charts.dart';
import 'package:cardflip/features/stats/stats_sections.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// The share of mature cards in percent, rounded down; 0 for no cards.
int maturePercent(golang.DeckProgress p) {
  final total = p.new_ + p.learning + p.young + p.mature;
  return total == 0 ? 0 : p.mature * 100 ~/ total;
}

/// "62% mature".
String matureLabel(AppLocalizations l, golang.DeckProgress p) =>
    l.deckProgressMature(percent(maturePercent(p) / 100));

/// A deck's new, learning, young and mature cards as a bar of parts in the
/// statistics' colours (app-shell spec, "Deck overview").
class DeckProgressBar extends StatelessWidget {
  const DeckProgressBar({super.key, required this.progress, this.height = 8});

  final golang.DeckProgress progress;
  final double height;

  @override
  Widget build(BuildContext context) {
    final colors = StatsColors.of(context);
    final p = progress;
    final parts = [
      (p.new_, colors.newCards),
      (p.learning, colors.learning),
      (p.young, colors.young),
      (p.mature, colors.mature),
    ];
    return Semantics(
      label: matureLabel(context.l10n, p),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(height / 2),
        child: ColoredBox(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          child: SizedBox(
            height: height,
            child: Row(
              children: [
                for (final (n, color) in parts)
                  if (n > 0)
                    Expanded(
                      flex: n,
                      child: ColoredBox(color: color),
                    ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// The deck overview's progress: the bar, the share of mature cards and
/// the number of cards of each part.
class DeckProgressCard extends StatelessWidget {
  const DeckProgressCard({super.key, required this.progress});

  final golang.DeckProgress progress;

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final theme = Theme.of(context);
    final colors = StatsColors.of(context);
    final p = progress;
    return Card(
      key: const Key('deck-progress'),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    l.deckProgress,
                    style: theme.textTheme.titleSmall,
                  ),
                ),
                Text(
                  matureLabel(l, p),
                  key: const Key('deck-progress-mature'),
                  style: theme.textTheme.titleSmall,
                ),
              ],
            ),
            const SizedBox(height: 8),
            ExcludeSemantics(child: DeckProgressBar(progress: p, height: 12)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 16,
              runSpacing: 4,
              children: [
                for (final (state, n, color) in [
                  (l.countNew, p.new_, colors.newCards),
                  (l.kindLearning, p.learning, colors.learning),
                  (l.kindYoung, p.young, colors.young),
                  (l.kindMature, p.mature, colors.mature),
                ])
                  LegendItem(
                    color: color,
                    label: l.deckProgressLegend(state, n),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
