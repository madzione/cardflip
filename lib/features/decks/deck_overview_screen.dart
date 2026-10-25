import 'package:flutter/material.dart';
import 'package:flutter_widget_from_html_core/flutter_widget_from_html_core.dart';
import 'package:url_launcher/url_launcher.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'deck_progress.dart';
import 'filtered_deck_screen.dart';

/// What choosing a deck on the deck list opens: today's counts, the buried
/// cards, the deck's progress, its description, and Study now, Options and
/// Unbury. It reloads whenever the deck list does, so it follows every
/// change and the passing of time.
class DeckOverviewScreen extends StatefulWidget {
  const DeckOverviewScreen({super.key, required this.deck});

  /// The deck as the deck list showed it when chosen.
  final golang.DeckNode deck;

  /// Opens a link of a description in the system browser.
  @visibleForTesting
  static Future<bool> Function(Uri url) openUrl = (url) =>
      launchUrl(url, mode: LaunchMode.externalApplication);

  @override
  State<DeckOverviewScreen> createState() => _DeckOverviewScreenState();
}

class _DeckOverviewScreenState extends State<DeckOverviewScreen> {
  late golang.DeckNode _deck = widget.deck;
  golang.DeckOverview? _overview;
  Object? _error;
  Listenable? _deckList;

  /// Set while a load runs; another request during it loads again after.
  bool _loading = false;
  bool _reload = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final deckList = AppScope.of(context).deckList;
    if (!identical(deckList, _deckList)) {
      _deckList?.removeListener(_onDecksChanged);
      _deckList = deckList..addListener(_onDecksChanged);
      _load();
    }
  }

  @override
  void dispose() {
    _deckList?.removeListener(_onDecksChanged);
    super.dispose();
  }

  void _onDecksChanged() {
    final decks = AppScope.read(context).deckList.decks;
    for (final d in decks) {
      if (d.id == _deck.id) {
        setState(() => _deck = d);
        _load();
        return;
      }
    }
    // The deck was deleted (or the collection replaced).
    if (decks.isNotEmpty) Navigator.of(context).maybePop();
  }

  Future<void> _load() async {
    if (_loading) {
      _reload = true;
      return;
    }
    _loading = true;
    try {
      do {
        _reload = false;
        try {
          final o = await golang.decks.overview(_deck.id);
          if (!mounted) return;
          setState(() {
            _overview = o;
            _error = null;
          });
        } catch (e) {
          if (!mounted) return;
          setState(() => _error = e);
        }
      } while (_reload);
    } finally {
      _loading = false;
    }
  }

  void _study() {
    Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => StudyScreen(deckId: _deck.id, deckName: _deck.fullName),
      ),
    );
  }

  Future<void> _options() => openDeckOptions(context, _deck);

  Future<void> _unbury() async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    if (await guard(() => golang.decks.unbury(_deck.id))) {
      undo.offer(l.unburiedDeck(_deck.fullName));
    }
  }

  Future<void> _editDescription() async {
    final text = await showDialog<String>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => _DescriptionDialog(initial: _overview?.description ?? ''),
    );
    if (text == null) return;
    await guard(() => golang.decks.setDescription(_deck.id, text));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_deck.fullName),
        actions: [
          IconButton(
            tooltip: context.l10n.editDescription,
            icon: const Icon(Icons.edit_note),
            onPressed: _overview == null ? null : _editDescription,
          ),
          const UndoMenuButton(),
        ],
      ),
      body: _body(context),
    );
  }

  Widget _body(BuildContext context) {
    final overview = _overview;
    if (overview == null) {
      if (_error case final error?) {
        return Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(errorMessage(error)),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: _load,
                child: Text(context.l10n.tryAgain),
              ),
            ],
          ),
        );
      }
      return const Center(child: CircularProgressIndicator());
    }
    final theme = Theme.of(context);
    final counts = overview.counts;
    final due = counts.new_ + counts.learning + counts.review;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 560),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (due > 0)
                _Counts(counts: counts)
              else
                _NothingDue(nextDue: overview.nextDue),
              if (overview.buried > 0) ...[
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      context.l10n.countBuriedCards(overview.buried),
                      key: const Key('buried'),
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(width: 8),
                    TextButton(
                      onPressed: _unbury,
                      child: Text(context.l10n.unbury),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 16),
              Wrap(
                alignment: WrapAlignment.center,
                spacing: 12,
                runSpacing: 8,
                children: [
                  // Focused, so that Enter starts studying a deck chosen
                  // with the keyboard (the deck picker).
                  FilledButton.icon(
                    autofocus: true,
                    onPressed: due > 0 ? _study : null,
                    icon: const Icon(Icons.play_arrow),
                    label: Text(context.l10n.studyNow),
                  ),
                  OutlinedButton.icon(
                    onPressed: _options,
                    icon: const Icon(Icons.tune),
                    label: Text(context.l10n.options),
                  ),
                  if (_deck.filtered) ...[
                    OutlinedButton.icon(
                      onPressed: () => rebuildFilteredDeck(context, _deck),
                      icon: const Icon(Icons.refresh),
                      label: Text(context.l10n.rebuild),
                    ),
                    OutlinedButton.icon(
                      onPressed: () => emptyFilteredDeck(context, _deck),
                      icon: const Icon(Icons.remove_done),
                      label: Text(context.l10n.emptyDeck),
                    ),
                  ],
                ],
              ),
              const SizedBox(height: 16),
              DeckProgressCard(progress: _deck.progress),
              if (overview.descriptionHTML.isNotEmpty) ...[
                const SizedBox(height: 24),
                const Divider(),
                const SizedBox(height: 8),
                _Description(html: overview.descriptionHTML),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

/// New, learning and review counts, each with its label.
class _Counts extends StatelessWidget {
  const _Counts({required this.counts});

  final golang.Counts counts;

  @override
  Widget build(BuildContext context) {
    final colors = CountColors.of(context);
    final theme = Theme.of(context);
    Widget count(String label, int n, Color color, String key) => Expanded(
      child: Column(
        children: [
          Text(
            '$n',
            key: Key(key),
            style: theme.textTheme.headlineMedium?.copyWith(
              color: n == 0 ? theme.colorScheme.outline : color,
              fontWeight: FontWeight.w600,
              fontFeatures: const [FontFeature.tabularFigures()],
            ),
          ),
          Text(label, style: theme.textTheme.bodyMedium),
        ],
      ),
    );
    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 8),
        child: Row(
          children: [
            count(
              context.l10n.countNew,
              counts.new_,
              colors.newCards,
              'overview-new',
            ),
            count(
              context.l10n.countLearning,
              counts.learning,
              colors.learning,
              'overview-learning',
            ),
            count(
              context.l10n.countToReview,
              counts.review,
              colors.review,
              'overview-review',
            ),
          ],
        ),
      ),
    );
  }
}

/// Says that nothing is due, and when the next card is.
class _NothingDue extends StatelessWidget {
  const _NothingDue({required this.nextDue});

  final DateTime? nextDue;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final next = nextDue;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            Icon(
              Icons.celebration_outlined,
              size: 40,
              color: theme.colorScheme.primary,
            ),
            const SizedBox(height: 12),
            Text(
              context.l10n.nothingDueNow,
              style: theme.textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 6),
            Text(
              next != null
                  ? context.l10n.nextCardDue(formatDue(next))
                  : context.l10n.deckHasNoCards,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

/// The deck's description: sanitized HTML from Go, shown like a card's
/// content, with images from the collection's media. Links open in the
/// system browser.
class _Description extends StatelessWidget {
  const _Description({required this.html});

  final String html;

  static Future<bool> _open(String url) async {
    final uri = Uri.tryParse(url);
    if (uri == null ||
        !const {'http', 'https', 'mailto'}.contains(uri.scheme)) {
      return true; // shown, not followed
    }
    try {
      await DeckOverviewScreen.openUrl(uri);
    } catch (e) {
      showError(e);
    }
    return true;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return HtmlWidget(
      html,
      key: const Key('deck-description'),
      textStyle: theme.textTheme.bodyLarge,
      customWidgetBuilder: (element) {
        if (element.localName != 'img') return null;
        final src = (element.attributes['src'] ?? '').trim();
        if (src.isEmpty || src.contains(':')) return const SizedBox.shrink();
        return MediaImage(
          name: src,
          cache: MediaCache.instance,
          alt: element.attributes['alt'],
        );
      },
      onTapUrl: _open,
    );
  }
}

/// Edits a deck's description: plain text, or simple HTML.
class _DescriptionDialog extends StatefulWidget {
  const _DescriptionDialog({required this.initial});

  final String initial;

  @override
  State<_DescriptionDialog> createState() => _DescriptionDialogState();
}

class _DescriptionDialogState extends State<_DescriptionDialog> {
  late final _controller = TextEditingController(text: widget.initial);

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(context.l10n.description),
      content: SizedBox(
        width: 480,
        child: TextField(
          controller: _controller,
          autofocus: true,
          minLines: 4,
          maxLines: 12,
          keyboardType: TextInputType.multiline,
          decoration: InputDecoration(
            border: const OutlineInputBorder(),
            helperText: context.l10n.descriptionHelper,
            helperMaxLines: 2,
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.of(context).pop(_controller.text),
          child: Text(context.l10n.save),
        ),
      ],
    );
  }
}
