import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'deck_list_screen.dart';

/// Asks which deck to study ("Study deck", `/`): a search over the decks'
/// full names. Returns the chosen deck, or null when cancelled.
Future<golang.DeckNode?> showDeckPicker(BuildContext context) {
  final app = AppScope.read(context);
  return showDialog<golang.DeckNode>(
    context: context,
    animationStyle: popupAnimation(context),
    builder: (_) =>
        DeckPicker(decks: app.deckList.decks, recent: app.settings.recentDecks),
  );
}

/// The decks whose full names contain every word of [query], ignoring case
/// and word order, the recently studied ones first (in [recent]'s order),
/// the others in [decks]' order.
List<golang.DeckNode> filterDecks(
  List<golang.DeckNode> decks,
  List<int> recent,
  String query,
) {
  final words = query
      .toLowerCase()
      .split(RegExp(r'\s+'))
      .where((w) => w.isNotEmpty)
      .toList();
  final matches = {
    for (final d in decks)
      if (words.every(d.fullName.toLowerCase().contains)) d.id: d,
  };
  final first = recent.toSet();
  return [
    for (final id in recent) ?matches[id],
    for (final d in matches.values)
      if (!first.contains(d.id)) d,
  ];
}

/// The "Study deck" dialog: a search field over the decks' full names and
/// the matching decks. The arrow keys move the selection, Enter or a tap
/// chooses a deck.
class DeckPicker extends StatefulWidget {
  const DeckPicker({super.key, required this.decks, required this.recent});

  /// All decks, in the deck tree's order.
  final List<golang.DeckNode> decks;

  /// The IDs of the decks most recently studied, latest first.
  final List<int> recent;

  @override
  State<DeckPicker> createState() => _DeckPickerState();
}

class _DeckPickerState extends State<DeckPicker> {
  static const _rowHeight = 48.0;

  final _scroll = ScrollController();
  late List<golang.DeckNode> _matches = filterDecks(
    widget.decks,
    widget.recent,
    '',
  );
  int _selected = 0;

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  void _search(String query) {
    setState(() {
      _matches = filterDecks(widget.decks, widget.recent, query);
      _selected = 0;
    });
    if (_scroll.hasClients) _scroll.jumpTo(0);
  }

  /// Moves the selection by [delta] rows and scrolls it into view.
  void _move(int delta) {
    if (_matches.isEmpty) return;
    setState(
      () => _selected = (_selected + delta).clamp(0, _matches.length - 1),
    );
    if (!_scroll.hasClients) return;
    final top = _selected * _rowHeight;
    final view = _scroll.position;
    if (top < view.pixels) {
      _scroll.jumpTo(top);
    } else if (top + _rowHeight > view.pixels + view.viewportDimension) {
      _scroll.jumpTo(top + _rowHeight - view.viewportDimension);
    }
  }

  void _choose(golang.DeckNode deck) => Navigator.of(context).pop(deck);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return AlertDialog(
      title: Text(context.l10n.studyDeck),
      contentPadding: const EdgeInsets.fromLTRB(24, 16, 24, 0),
      content: SizedBox(
        width: 480,
        height: 360,
        child: Column(
          children: [
            CallbackShortcuts(
              bindings: {
                const SingleActivator(LogicalKeyboardKey.arrowDown): () =>
                    _move(1),
                const SingleActivator(LogicalKeyboardKey.arrowUp): () =>
                    _move(-1),
              },
              child: TextField(
                autofocus: true,
                decoration: InputDecoration(
                  hintText: context.l10n.searchDecks,
                  prefixIcon: const Icon(Icons.search),
                ),
                onChanged: _search,
                onSubmitted: (_) {
                  if (_matches.isNotEmpty) _choose(_matches[_selected]);
                },
              ),
            ),
            const SizedBox(height: 8),
            Expanded(
              child: _matches.isEmpty
                  ? Center(
                      child: Text(
                        context.l10n.noMatchingDecks,
                        style: TextStyle(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    )
                  : ListView.builder(
                      controller: _scroll,
                      itemExtent: _rowHeight,
                      itemCount: _matches.length,
                      itemBuilder: (context, i) {
                        final deck = _matches[i];
                        return ListTile(
                          key: ValueKey('pick-${deck.id}'),
                          dense: true,
                          selected: i == _selected,
                          selectedTileColor:
                              theme.colorScheme.secondaryContainer,
                          title: Text(
                            deck.fullName,
                            overflow: TextOverflow.ellipsis,
                          ),
                          trailing: DueCounts(counts: deck.counts),
                          onTap: () => _choose(deck),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(context.l10n.cancel),
        ),
      ],
    );
  }
}
