import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/profiles_model.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/editor/note_editor_screen.dart';
import 'package:cardflip/features/export/export_dialog.dart';
import 'package:cardflip/features/import/import_screen.dart';
import 'package:cardflip/features/progress/achievements_screen.dart';
import 'package:cardflip/features/progress/new_achievement_notice.dart';
import 'package:cardflip/features/stats/stats_screen.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'custom_study_dialog.dart';
import 'deck_overview_screen.dart';
import 'deck_progress.dart';
import 'filtered_deck_screen.dart';
import 'progress_header.dart';
import 'sync_indicator.dart';
import 'weekly_recap.dart';

/// "Decks", and below it the open profile's name when there are several.
class _Title extends StatelessWidget {
  const _Title({required this.profiles});

  final ProfilesModel profiles;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: profiles,
    builder: (context, _) {
      final decks = Text(context.l10n.decks);
      final name = profiles.currentName;
      if (!profiles.several || name == null) return decks;
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          decks,
          Text(
            name,
            key: const Key('profile-name'),
            style: Theme.of(context).textTheme.bodySmall,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      );
    },
  );
}

/// The home screen: the deck tree with due counts.
class DeckListScreen extends StatelessWidget {
  const DeckListScreen({super.key});

  Future<void> _createDeck(BuildContext context) async {
    final name = await askText(
      context,
      title: context.l10n.newDeck,
      label: context.l10n.name,
      helper: context.l10n.newDeckHelper,
      action: context.l10n.create,
    );
    if (name == null) return;
    await guard(() => golang.decks.create(name));
  }

  @override
  Widget build(BuildContext context) {
    final model = AppScope.of(context).deckList;
    final settings = AppScope.of(context).settings;
    final l = context.l10n;
    final narrow = narrowScreen(context);
    return NewAchievementNotice(
      child: Scaffold(
        appBar: AppBar(
          title: _Title(profiles: AppScope.of(context).profiles),
          actions: [
            const SyncIndicator(),
            IconButton(
              tooltip: withShortcut(l.studyDeck, pickDeckShortcutLabel),
              icon: const Icon(Icons.manage_search),
              onPressed: () => AdaptiveScaffold.pickDeck(context),
            ),
            // The menu has Achievements while streaks and goals are shown.
            ListenableBuilder(
              listenable: settings,
              builder: (context, _) {
                // Shown as icons, or in the menu where the app bar is narrow.
                final more = [
                  (
                    l.statistics,
                    Icons.bar_chart,
                    () => showStats(context, collectionScope),
                  ),
                  (
                    l.createDeck,
                    Icons.create_new_folder_outlined,
                    () => _createDeck(context),
                  ),
                  (
                    l.createFilteredDeck,
                    Icons.filter_alt_outlined,
                    () => createFilteredDeck(
                      context,
                      search: defaultFilterSearch(context),
                    ),
                  ),
                ];
                return Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (!narrow)
                      for (final (label, icon, action) in more)
                        IconButton(
                          tooltip: label,
                          icon: Icon(icon),
                          onPressed: action,
                        ),
                    UndoMenuButton(
                      actions: [
                        if (narrow)
                          for (final (label, _, action) in more)
                            (label, action),
                        // Always in the menu, to keep the app bar narrow.
                        if (settings.showProgress)
                          (l.achievements, () => showAchievements(context)),
                      ],
                    ),
                  ],
                );
              },
            ),
          ],
        ),
        body: ListenableBuilder(
          listenable: Listenable.merge([model, settings]),
          builder: (context, _) {
            if (model.loading) {
              return const Center(child: CircularProgressIndicator());
            }
            if (model.error != null) {
              return Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(errorMessage(model.error!)),
                    const SizedBox(height: 12),
                    FilledButton(
                      onPressed: model.load,
                      child: Text(context.l10n.tryAgain),
                    ),
                  ],
                ),
              );
            }
            final empty = model.decks.every((d) => d.depth > 0 || d.cards == 0);
            final progress =
                settings.showProgress && settings.deckProgressInList;
            return ListView(
              padding: const EdgeInsets.only(bottom: 24),
              children: [
                const WeeklyRecap(),
                const ProgressHeader(),
                for (final deck in model.visible)
                  DeckRow(deck: deck, showProgress: progress),
                if (empty) const _EmptyHint(),
              ],
            );
          },
        ),
      ),
    );
  }
}

/// Shown on first run, while the collection has no cards.
class _EmptyHint extends StatelessWidget {
  const _EmptyHint();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                context.l10n.collectionEmpty,
                style: theme.textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              Text(context.l10n.collectionEmptyHint),
              const SizedBox(height: 16),
              Wrap(
                spacing: 12,
                runSpacing: 8,
                children: [
                  FilledButton.icon(
                    onPressed: () => AdaptiveScaffold.select(context, 2),
                    icon: const Icon(Icons.add),
                    label: Text(context.l10n.addACard),
                  ),
                  OutlinedButton.icon(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute<void>(
                        builder: (_) => const ImportScreen(),
                      ),
                    ),
                    icon: const Icon(Icons.file_open_outlined),
                    label: Text(context.l10n.importAFile),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Asks for a new name for [deck] and renames it; `::` moves it.
Future<void> renameDeck(BuildContext context, golang.DeckNode deck) async {
  final name = await askText(
    context,
    title: context.l10n.renameDeck,
    initial: deck.fullName,
    label: context.l10n.name,
    helper: context.l10n.renameDeckHelper,
    action: context.l10n.rename,
  );
  if (name != null) await guard(() => golang.decks.rename(deck.id, name));
}

/// Deletes [deck] with its subdecks and cards, after a confirmation when it
/// has cards; cards in a deleted filtered deck return to their home decks.
Future<void> deleteDeck(BuildContext context, golang.DeckNode deck) async {
  final golang.DeletionInfo info;
  try {
    info = await golang.decks.deletionInfo(deck.id);
  } catch (e) {
    showError(e);
    return;
  }
  if (info.cards > 0 && context.mounted) {
    final subdecks = info.decks - 1;
    final ok = await confirm(
      context,
      title: context.l10n.deleteNamedTitle(deck.fullName),
      message: context.l10n.deleteDeckBody(
        context.l10n.countCards(info.cards),
        info.inFiltered > 0
            ? context.l10n.deleteDeckInFiltered(info.inFiltered)
            : '',
        subdecks > 0 ? context.l10n.deleteDeckSubdecks(subdecks) : '',
      ),
      action: context.l10n.delete,
      destructive: true,
    );
    if (!ok) return;
  }
  if (!context.mounted) return;
  final undo = AppScope.read(context).undo;
  final l = context.l10n;
  if (await guard(() => golang.decks.delete(deck.id))) {
    undo.offer(
      info.returned > 0
          ? l.deletedDeckReturned(deck.fullName, info.returned)
          : l.deletedNamed(deck.fullName),
    );
  }
}

enum _DeckAction {
  study,
  add,
  customStudy,
  rebuild,
  empty,
  statistics,
  options,
  export,
  rename,
  delete,
}

/// One deck: name, due counts and actions, and with [showProgress] a thin
/// progress bar under the name. Tapping it opens the deck's overview; the
/// menu's Study starts studying directly.
class DeckRow extends StatelessWidget {
  const DeckRow({super.key, required this.deck, this.showProgress = false});

  final golang.DeckNode deck;

  /// "Show deck progress in the deck list", while streaks and goals are
  /// shown (app-shell spec, "Deck progress in the deck list").
  final bool showProgress;

  /// Narrower rows show the counts below the deck's name.
  static const _sideBySideWidth = 280.0;

  void _overview(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(builder: (_) => DeckOverviewScreen(deck: deck)),
    );
  }

  void _study(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => StudyScreen(deckId: deck.id, deckName: deck.fullName),
      ),
    );
  }

  Future<void> _action(BuildContext context, _DeckAction action) async {
    switch (action) {
      case _DeckAction.study:
        _study(context);
      case _DeckAction.add:
        await Navigator.of(context).push(
          MaterialPageRoute<void>(
            builder: (_) => NoteEditorScreen(deckId: deck.id),
          ),
        );
      case _DeckAction.customStudy:
        await showDialog<bool>(
          context: context,
          animationStyle: popupAnimation(context),
          builder: (_) => CustomStudyDialog(deck: deck),
        );
      case _DeckAction.rebuild:
        await rebuildFilteredDeck(context, deck);
      case _DeckAction.empty:
        await emptyFilteredDeck(context, deck);
      case _DeckAction.statistics:
        await showStats(context, deckScope(deck.id));
      case _DeckAction.options:
        await openDeckOptions(context, deck);
      case _DeckAction.export:
        await showExportDialog(context, DeckExport(deck.id, deck.name));
      case _DeckAction.rename:
        await renameDeck(context, deck);
      case _DeckAction.delete:
        await deleteDeck(context, deck);
    }
  }

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => _overview(context),
      child: Padding(
        padding: EdgeInsets.only(left: 8.0 + deck.depth * 24, right: 4),
        child: SizedBox(
          height: 56,
          child: LayoutBuilder(
            builder: (context, constraints) {
              final name = Text(
                deck.name,
                overflow: TextOverflow.ellipsis,
                style: deck.filtered
                    ? Theme.of(context).textTheme.bodyLarge?.copyWith(
                        color: Theme.of(context).colorScheme.primary,
                      )
                    : Theme.of(context).textTheme.bodyLarge,
              );
              final counts = DueCounts(
                key: Key('counts-${deck.id}'),
                counts: deck.counts,
              );
              final bar = showProgress
                  ? Padding(
                      padding: const EdgeInsets.only(top: 4, right: 8),
                      child: DeckProgressBar(
                        key: Key('deck-progress-${deck.id}'),
                        progress: deck.progress,
                        height: 4,
                      ),
                    )
                  : null;
              // Too narrow for both side by side, as on a phone with a large
              // UI size: the counts go below the name.
              final stacked = constraints.maxWidth < _sideBySideWidth;
              return Row(
                children: [
                  SizedBox(
                    width: 40,
                    child: deck.hasChildren
                        ? IconButton(
                            tooltip: deck.collapsed
                                ? context.l10n.expand
                                : context.l10n.collapse,
                            icon: Icon(
                              deck.collapsed
                                  ? Icons.chevron_right
                                  : Icons.expand_more,
                            ),
                            onPressed: () => guard(
                              () => golang.decks.setCollapsed(
                                deck.id,
                                !deck.collapsed,
                              ),
                            ),
                          )
                        : null,
                  ),
                  if (deck.filtered)
                    Padding(
                      padding: const EdgeInsets.only(right: 6),
                      child: Tooltip(
                        message: context.l10n.filteredDeck,
                        child: Icon(
                          Icons.filter_alt_outlined,
                          key: Key('filtered-${deck.id}'),
                          size: 18,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ),
                  Expanded(
                    child: stacked
                        ? Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              name,
                              FittedBox(
                                fit: BoxFit.scaleDown,
                                alignment: AlignmentDirectional.centerStart,
                                child: counts,
                              ),
                              ?bar,
                            ],
                          )
                        : bar == null
                        ? name
                        : Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [name, bar],
                          ),
                  ),
                  if (!stacked) counts,
                  PopupMenuButton<_DeckAction>(
                    popUpAnimationStyle: popupAnimation(context),
                    tooltip: context.l10n.deckActions,
                    onSelected: (a) => _action(context, a),
                    itemBuilder: (_) => [
                      PopupMenuItem(
                        value: _DeckAction.study,
                        child: Text(context.l10n.study),
                      ),
                      // Filtered decks gather their cards with their search:
                      // no notes are added to them, and custom study needs a
                      // regular deck.
                      if (deck.filtered) ...[
                        PopupMenuItem(
                          value: _DeckAction.rebuild,
                          child: Text(context.l10n.rebuild),
                        ),
                        PopupMenuItem(
                          value: _DeckAction.empty,
                          child: Text(context.l10n.emptyDeck),
                        ),
                      ] else ...[
                        PopupMenuItem(
                          value: _DeckAction.add,
                          child: Text(context.l10n.addCard),
                        ),
                        PopupMenuItem(
                          value: _DeckAction.customStudy,
                          child: Text(context.l10n.customStudy),
                        ),
                      ],
                      PopupMenuItem(
                        value: _DeckAction.statistics,
                        child: Text(context.l10n.statistics),
                      ),
                      PopupMenuItem(
                        value: _DeckAction.options,
                        child: Text(context.l10n.options),
                      ),
                      PopupMenuItem(
                        value: _DeckAction.export,
                        child: Text(context.l10n.export),
                      ),
                      PopupMenuItem(
                        value: _DeckAction.rename,
                        child: Text(context.l10n.rename),
                      ),
                      PopupMenuItem(
                        value: _DeckAction.delete,
                        child: Text(context.l10n.delete),
                      ),
                    ],
                  ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}

/// New, learning and review counts in their colours; zero counts are dimmed.
class DueCounts extends StatelessWidget {
  const DueCounts({super.key, required this.counts});

  final golang.Counts counts;

  @override
  Widget build(BuildContext context) {
    final colors = CountColors.of(context);
    final dim = Theme.of(context).colorScheme.outline;
    Widget count(int n, Color color, String tooltip) => SizedBox(
      width: 44,
      child: Tooltip(
        message: tooltip,
        child: Text(
          '$n',
          textAlign: TextAlign.end,
          style: TextStyle(
            color: n == 0 ? dim : color,
            fontWeight: n == 0 ? FontWeight.normal : FontWeight.w600,
            fontFeatures: const [FontFeature.tabularFigures()],
          ),
        ),
      ),
    );
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        count(
          counts.new_,
          colors.newCards,
          context.l10n.dueNewTooltip(counts.new_),
        ),
        count(
          counts.learning,
          colors.learning,
          context.l10n.dueLearningTooltip(counts.learning),
        ),
        count(
          counts.review,
          colors.review,
          context.l10n.dueReviewTooltip(counts.review),
        ),
      ],
    );
  }
}
