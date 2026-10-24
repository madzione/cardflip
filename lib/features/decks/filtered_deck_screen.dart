import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/browse/browse_sidebar.dart' show sidebarTerm;
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'deck_options_screen.dart';

/// The orders a filtered deck's search can pick and show its cards in.
Map<String, String> filterOrders(AppLocalizations l) => {
  'oldestSeen': l.orderOldestSeen,
  'random': l.orderRandom,
  'intervalsAscending': l.orderIntervalsAscending,
  'intervalsDescending': l.orderIntervalsDescending,
  'mostLapses': l.orderMostLapses,
  'added': l.orderAdded,
  'due': l.orderDue,
  'latestAdded': l.orderLatestAdded,
  'retrievability': l.orderRetrievability,
};

/// The search a new filtered deck starts with: the most recently studied
/// regular deck, or every card.
String defaultFilterSearch(BuildContext context) {
  final app = AppScope.read(context);
  final decks = {for (final d in app.deckList.decks) d.id: d};
  for (final id in app.settings.recentDecks) {
    final d = decks[id];
    if (d != null && !d.filtered) return sidebarTerm('deck', d.fullName);
  }
  return '';
}

/// Opens the options of a new filtered deck searching for [search], and
/// builds it when the user chooses Build. Returns the result, or null when
/// cancelled.
Future<golang.FilteredBuild?> createFilteredDeck(
  BuildContext context, {
  required String search,
}) async {
  final golang.FilteredDeckOptions options;
  try {
    options = await golang.decks.newFilteredOptions(search);
  } catch (e) {
    showError(e);
    return null;
  }
  if (!context.mounted) return null;
  return _open(context, null, options);
}

/// Opens filtered deck [deck]'s options; Rebuild saves them and rebuilds
/// it.
Future<golang.FilteredBuild?> editFilteredDeck(
  BuildContext context,
  golang.DeckNode deck,
) async {
  final golang.FilteredDeckOptions options;
  try {
    options = await golang.decks.filteredOptions(deck.id);
  } catch (e) {
    showError(e);
    return null;
  }
  if (!context.mounted) return null;
  return _open(context, deck.id, options);
}

Future<golang.FilteredBuild?> _open(
  BuildContext context,
  int? deckId,
  golang.FilteredDeckOptions options,
) async {
  final built = await Navigator.of(context).push(
    MaterialPageRoute<golang.FilteredBuild>(
      builder: (_) => FilteredDeckScreen(deckId: deckId, options: options),
    ),
  );
  if (built != null && context.mounted) {
    AppScope.read(context).undo
        .offer(context.l10n.filteredDeckHolds(options.name, built.cards));
  }
  return built;
}

/// Opens [deck]'s options: the filtered deck options for a filtered deck,
/// the deck options otherwise.
Future<void> openDeckOptions(BuildContext context, golang.DeckNode deck) async {
  if (deck.filtered) {
    await editFilteredDeck(context, deck);
    return;
  }
  await Navigator.of(context).push(
    MaterialPageRoute<bool>(builder: (_) => DeckOptionsScreen(deck: deck)),
  );
}

/// Returns filtered deck [deck]'s cards home and gathers them again.
Future<void> rebuildFilteredDeck(
  BuildContext context,
  golang.DeckNode deck,
) async {
  final undo = AppScope.read(context).undo;
  final l = context.l10n;
  try {
    final built = await golang.decks.rebuild(deck.id);
    undo.offer(l.rebuiltDeck(deck.fullName, built.cards));
  } catch (e) {
    showError(e);
  }
}

/// Returns filtered deck [deck]'s cards to their home decks.
Future<void> emptyFilteredDeck(
  BuildContext context,
  golang.DeckNode deck,
) async {
  final undo = AppScope.read(context).undo;
  final l = context.l10n;
  if (await guard(() => golang.decks.empty(deck.id))) {
    undo.offer(l.emptiedDeck(deck.fullName));
  }
}

/// The options of a filtered deck: its name, one or two searches with a
/// card limit and an order each, and whether answers reschedule cards or
/// only preview them. Build (Rebuild for an existing deck) saves and
/// builds it in one step and closes with the result; when no card matches,
/// the screen stays open with the error.
class FilteredDeckScreen extends StatefulWidget {
  const FilteredDeckScreen({
    super.key,
    required this.deckId,
    required this.options,
  });

  /// The deck edited, or null for a new one.
  final int? deckId;
  final golang.FilteredDeckOptions options;

  @override
  State<FilteredDeckScreen> createState() => _FilteredDeckScreenState();
}

class _Filter {
  _Filter(golang.FilterTerm t)
    : search = TextEditingController(text: t.search),
      limit = TextEditingController(text: '${t.limit}'),
      order = t.order;

  final TextEditingController search, limit;
  String order;

  void dispose() {
    search.dispose();
    limit.dispose();
  }
}

class _FilteredDeckScreenState extends State<FilteredDeckScreen> {
  late final _name = TextEditingController(text: widget.options.name);
  late final _first = _Filter(widget.options.terms.first);
  late final _second = _Filter(
    widget.options.terms.length > 1
        ? widget.options.terms[1]
        : golang.FilterTerm(
            search: '',
            limit: widget.options.secondLimit,
            order: widget.options.secondOrder,
          ),
  );
  late bool _useSecond = widget.options.terms.length > 1;
  late bool _reschedule = widget.options.reschedule;
  late final _delays = [
    for (var i = 0; i < 3; i++)
      TextEditingController(
        text: i < widget.options.previewDelays.length
            ? '${widget.options.previewDelays[i]}'
            : '0',
      ),
  ];
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _name.dispose();
    _first.dispose();
    _second.dispose();
    for (final c in _delays) {
      c.dispose();
    }
    super.dispose();
  }

  /// The options as entered, or null after showing what is wrong.
  golang.FilteredDeckOptions? _options() {
    golang.FilterTerm? term(_Filter f, String which) {
      final limit = int.tryParse(f.limit.text.trim());
      if (limit == null || limit < 1 || limit > 99999) {
        setState(() => _error = context.l10n.filterLimitError(which));
        return null;
      }
      return golang.FilterTerm(
        search: f.search.text.trim(),
        limit: limit,
        order: f.order,
      );
    }

    final first = term(_first, 'first');
    if (first == null) return null;
    final terms = [first];
    if (_useSecond) {
      final second = term(_second, 'second');
      if (second == null) return null;
      terms.add(second);
    }
    final delays = <int>[];
    for (final c in _delays) {
      final d = int.tryParse(c.text.trim());
      if (d == null || d < 0 || d > 86400) {
        setState(() => _error = context.l10n.previewDelaysError);
        return null;
      }
      delays.add(d);
    }
    return golang.FilteredDeckOptions(
      name: _name.text.trim(),
      terms: terms,
      reschedule: _reschedule,
      previewDelays: delays,
      secondLimit: widget.options.secondLimit,
      secondOrder: widget.options.secondOrder,
    );
  }

  Future<void> _build() async {
    final options = _options();
    if (options == null || _busy) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final id = widget.deckId;
      final built = id == null
          ? await golang.decks.createFiltered(options)
          : await golang.decks.updateFiltered(id, options);
      if (mounted) Navigator.of(context).pop(built);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Widget _filter(_Filter f, String prefix) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          key: Key('$prefix-search'),
          controller: f.search,
          decoration: InputDecoration(
            labelText: context.l10n.search,
            helperText: context.l10n.filterSearchHelper,
          ),
          onSubmitted: (_) => _build(),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            SizedBox(
              width: 120,
              child: TextField(
                key: Key('$prefix-limit'),
                controller: f.limit,
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: InputDecoration(labelText: context.l10n.limitTo),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: DropdownButtonFormField<String>(
                key: Key('$prefix-order'),
                initialValue: f.order,
                isExpanded: true,
                decoration: InputDecoration(
                  labelText: context.l10n.cardsSelectedBy,
                ),
                items: [
                  for (final e in filterOrders(context.l10n).entries)
                    DropdownMenuItem(value: e.key, child: Text(e.value)),
                ],
                onChanged: (v) => setState(() => f.order = v ?? f.order),
              ),
            ),
          ],
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l = context.l10n;
    final action = widget.deckId == null ? l.build : l.rebuild;
    Widget heading(String text) => Padding(
      padding: const EdgeInsets.only(top: 24, bottom: 8),
      child: Text(text, style: theme.textTheme.titleSmall),
    );
    // Keys name the ratings in English.
    const delayKeys = ['again', 'hard', 'good'];
    final delayLabels = [l.again, l.hard, l.good];
    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.deckId == null ? l.newFilteredDeck : l.filteredDeckOptions,
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 8),
            child: FilledButton(
              key: const Key('build-filtered'),
              onPressed: _busy ? null : _build,
              child: Text(action),
            ),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 560),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (_error != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Text(
                        _error!,
                        key: const Key('filtered-error'),
                        style: TextStyle(color: theme.colorScheme.error),
                      ),
                    ),
                  TextField(
                    key: const Key('filtered-name'),
                    controller: _name,
                    decoration: InputDecoration(labelText: l.name),
                  ),
                  heading(l.filter),
                  _filter(_first, 'first'),
                  const SizedBox(height: 8),
                  SwitchListTile(
                    key: const Key('second-filter'),
                    contentPadding: EdgeInsets.zero,
                    title: Text(l.secondFilter),
                    subtitle: Text(l.secondFilterHelper),
                    value: _useSecond,
                    onChanged: (v) => setState(() => _useSecond = v),
                  ),
                  if (_useSecond) _filter(_second, 'second'),
                  heading(l.options),
                  SwitchListTile(
                    key: const Key('reschedule'),
                    contentPadding: EdgeInsets.zero,
                    title: Text(l.reschedule),
                    subtitle: Text(
                      _reschedule ? l.rescheduleOn : l.rescheduleOff,
                    ),
                    value: _reschedule,
                    onChanged: (v) => setState(() => _reschedule = v),
                  ),
                  if (!_reschedule) ...[
                    const SizedBox(height: 8),
                    Text(l.previewDelays, style: theme.textTheme.bodyMedium),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        for (var i = 0; i < 3; i++) ...[
                          if (i > 0) const SizedBox(width: 12),
                          Expanded(
                            child: TextField(
                              key: Key('delay-${delayKeys[i]}'),
                              controller: _delays[i],
                              keyboardType: TextInputType.number,
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                              ],
                              decoration: InputDecoration(
                                labelText: delayLabels[i],
                                suffixText: l.unitSecondsShort,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ],
                  const SizedBox(height: 24),
                  Text(
                    l.filteredDeckNote,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
