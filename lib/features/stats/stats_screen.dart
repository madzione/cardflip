import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'stats_pdf.dart';
import 'stats_sections.dart';

/// Statistics for the whole collection.
const collectionScope = golang.Scope(kind: 'collection');

/// Statistics for a deck with its subdecks.
golang.Scope deckScope(int deckId) =>
    golang.Scope(kind: 'deck', deckID: deckId);

/// Opens statistics for [scope].
Future<void> showStats(BuildContext context, golang.Scope scope) =>
    Navigator.of(
      context,
    ).push(MaterialPageRoute<void>(builder: (_) => StatsScreen(scope: scope)));

/// The statistics screen (statistics spec): scope and period controls and
/// every section, each loading on its own so the screen fills in as Go
/// finishes them.
class StatsScreen extends StatefulWidget {
  const StatsScreen({super.key, required this.scope});

  final golang.Scope scope;

  @override
  State<StatsScreen> createState() => _StatsScreenState();
}

class _StatsScreenState extends State<StatsScreen> {
  late golang.Scope _scope = widget.scope;
  String _period = 'month';
  final _sections = <StatsSection, Future<Object>>{};
  late final _search = TextEditingController(text: widget.scope.query ?? '');

  /// The deck the deck scope shows: the last one chosen.
  late int? _deckId = widget.scope.deckID;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  /// Asks Go for every section, or only the time-based ones.
  void _load({bool timeBasedOnly = false}) {
    for (final s in StatsSection.values) {
      if (timeBasedOnly && !s.timeBased) continue;
      // Sections scrolled out of view have no listener yet; their errors
      // are shown once they are built.
      _sections[s] = s.load(_scope, _period)..ignore();
    }
  }

  void _setScope(golang.Scope scope) {
    setState(() {
      _scope = scope;
      _load();
    });
  }

  void _setPeriod(String period) {
    setState(() {
      _period = period;
      _load(timeBasedOnly: true);
    });
  }

  void _setKind(String kind, List<golang.DeckNode> decks) {
    switch (kind) {
      case 'deck':
        final id = _deckId ?? (decks.isEmpty ? null : decks.first.id);
        if (id == null) return;
        _deckId = id;
        _setScope(deckScope(id));
      case 'search':
        _setScope(golang.Scope(kind: 'search', query: _search.text.trim()));
      default:
        _setScope(collectionScope);
    }
  }

  /// What the statistics cover, for the title and the PDF.
  String _scopeLabel(List<golang.DeckNode> decks) {
    switch (_scope.kind) {
      case 'deck':
        final deck = decks.where((d) => d.id == _scope.deckID).firstOrNull;
        return deck?.fullName ?? context.l10n.statsDeckFallback;
      case 'search':
        final q = _scope.query ?? '';
        return q.isEmpty
            ? context.l10n.allCards
            : context.l10n.statsSearchScope(q);
    }
    return context.l10n.wholeCollection;
  }

  Future<void> _savePdf(String scopeLabel) async {
    final picker = AppScope.read(context).picker;
    setState(() => _saving = true);
    try {
      final data = <StatsSection, Object>{};
      for (final MapEntry(key: s, value: f) in _sections.entries) {
        try {
          data[s] = await f;
        } catch (_) {
          // A failed section is left out of the PDF, as it is on screen.
        }
      }
      final now = DateTime.now();
      final bytes = await buildStatsPdf(
        scope: scopeLabel,
        period: statsPeriods[_period]!,
        date: now,
        data: data,
      );
      final name = statsPdfName(scopeLabel, now);
      if (await picker.save(name, bytes)) showInfo(appL10n.savedFile(name));
    } catch (e) {
      showError(e);
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final decks = AppScope.of(context).deckList.decks;
    final label = _scopeLabel(decks);
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(context.l10n.statistics),
            Text(
              label,
              key: const Key('stats-scope-label'),
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: context.l10n.saveAsPdf,
            icon: _saving
                ? const SizedBox.square(
                    dimension: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.picture_as_pdf_outlined),
            onPressed: _saving ? null : () => _savePdf(label),
          ),
        ],
      ),
      body: LayoutBuilder(
        builder: (context, constraints) {
          final controls = _Controls(
            scope: _scope,
            period: _period,
            decks: decks,
            search: _search,
            onKind: (k) => _setKind(k, decks),
            onDeck: (id) {
              _deckId = id;
              _setScope(deckScope(id));
            },
            onSearch: () => _setKind('search', decks),
            onPeriod: _setPeriod,
          );
          final cards = [
            for (final s in StatsSection.values)
              _SectionCard(
                key: Key('stats-${s.name}'),
                section: s,
                future: _sections[s]!,
              ),
          ];
          if (constraints.maxWidth >= wideLayoutWidth) {
            const gap = 12.0;
            final width = (constraints.maxWidth - 32 - gap) / 2;
            return SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  controls,
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: gap,
                    runSpacing: gap,
                    children: [
                      for (final c in cards) SizedBox(width: width, child: c),
                    ],
                  ),
                ],
              ),
            );
          }
          return ListView(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            children: [
              controls,
              for (final c in cards)
                Padding(padding: const EdgeInsets.only(top: 12), child: c),
            ],
          );
        },
      ),
    );
  }
}

/// The scope (deck, collection or search) and period controls.
class _Controls extends StatelessWidget {
  const _Controls({
    required this.scope,
    required this.period,
    required this.decks,
    required this.search,
    required this.onKind,
    required this.onDeck,
    required this.onSearch,
    required this.onPeriod,
  });

  final golang.Scope scope;
  final String period;
  final List<golang.DeckNode> decks;
  final TextEditingController search;
  final ValueChanged<String> onKind;
  final ValueChanged<int> onDeck;
  final VoidCallback onSearch;
  final ValueChanged<String> onPeriod;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        SegmentedButton<String>(
          key: const Key('stats-scope'),
          showSelectedIcon: false,
          segments: [
            ButtonSegment(
              value: 'deck',
              label: Text(context.l10n.deck),
              enabled: decks.isNotEmpty,
            ),
            ButtonSegment(
              value: 'collection',
              label: Text(context.l10n.collectionScope),
            ),
            ButtonSegment(
              value: 'search',
              label: Text(context.l10n.searchScope),
            ),
          ],
          selected: {scope.kind},
          onSelectionChanged: (s) => onKind(s.first),
        ),
        if (scope.kind == 'deck') ...[
          const SizedBox(height: 8),
          DropdownButtonFormField<int>(
            key: const Key('stats-deck'),
            initialValue: decks.any((d) => d.id == scope.deckID)
                ? scope.deckID
                : null,
            isExpanded: true,
            decoration: InputDecoration(
              labelText: context.l10n.deckWithSubdecks,
              border: const OutlineInputBorder(),
              isDense: true,
            ),
            items: [
              for (final d in decks)
                DropdownMenuItem(
                  value: d.id,
                  child: Text(d.fullName, overflow: TextOverflow.ellipsis),
                ),
            ],
            onChanged: (id) {
              if (id != null) onDeck(id);
            },
          ),
        ],
        if (scope.kind == 'search') ...[
          const SizedBox(height: 8),
          TextField(
            key: const Key('stats-search'),
            controller: search,
            textInputAction: TextInputAction.search,
            decoration: InputDecoration(
              hintText: context.l10n.statsSearchHint,
              border: const OutlineInputBorder(),
              isDense: true,
              prefixIcon: const Icon(Icons.search),
              suffixIcon: IconButton(
                tooltip: context.l10n.showStatistics,
                icon: const Icon(Icons.arrow_forward),
                onPressed: onSearch,
              ),
            ),
            onSubmitted: (_) => onSearch(),
          ),
        ],
        const SizedBox(height: 8),
        SegmentedButton<String>(
          key: const Key('stats-period'),
          showSelectedIcon: false,
          segments: [
            for (final MapEntry(key: p, value: label) in statsPeriods.entries)
              ButtonSegment(
                value: p,
                label: Text(
                  p == 'all' ? context.l10n.periodAllShort : label,
                  overflow: TextOverflow.ellipsis,
                ),
                tooltip: label,
              ),
          ],
          selected: {period},
          onSelectionChanged: (s) => onPeriod(s.first),
        ),
      ],
    );
  }
}

/// One section: its title and its content once Go has computed it.
class _SectionCard extends StatelessWidget {
  const _SectionCard({super.key, required this.section, required this.future});

  final StatsSection section;
  final Future<Object> future;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(section.title, style: theme.textTheme.titleMedium),
            const SizedBox(height: 12),
            FutureBuilder<Object>(
              future: future,
              builder: (context, snap) {
                if (snap.hasError) {
                  return Text(
                    errorMessage(snap.error!),
                    style: TextStyle(color: theme.colorScheme.error),
                  );
                }
                if (!snap.hasData) {
                  return const SizedBox(
                    height: 80,
                    child: Center(child: CircularProgressIndicator()),
                  );
                }
                // A new scope or period is loading: the old content stays
                // until it is ready.
                if (snap.connectionState != ConnectionState.done) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      const LinearProgressIndicator(),
                      Opacity(opacity: 0.5, child: section.content(snap.data!)),
                    ],
                  );
                }
                return section.content(snap.data!);
              },
            ),
          ],
        ),
      ),
    );
  }
}
