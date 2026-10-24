import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/features/study/study_screen.dart' show ratingNames;
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'reposition_dialog.dart';

/// Opens card info for a card.
Future<void> showCardInfo(BuildContext context, int cardId) => Navigator.of(
  context,
).push(MaterialPageRoute<void>(builder: (_) => CardInfoScreen(cardId: cardId)));

/// A card's identity, scheduling and memory state, and its review history,
/// newest first. Wide layouts show the summary and the history side by
/// side.
class CardInfoScreen extends StatefulWidget {
  const CardInfoScreen({super.key, required this.cardId});

  final int cardId;

  @override
  State<CardInfoScreen> createState() => _CardInfoScreenState();
}

class _CardInfoScreenState extends State<CardInfoScreen> {
  golang.CardInfo? _info;
  Object? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final info = await golang.browse.cardInfo(widget.cardId);
      if (mounted) setState(() => _info = info);
    } catch (e) {
      if (mounted) setState(() => _error = e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.cardInfo)),
      body: _body(context),
    );
  }

  Widget _body(BuildContext context) {
    final info = _info;
    if (info == null) {
      return Center(
        child: _error != null
            ? Padding(
                padding: const EdgeInsets.all(16),
                child: Text(errorMessage(_error!)),
              )
            : const CircularProgressIndicator(),
      );
    }
    final theme = Theme.of(context);
    final heading = Text(
      context.l10n.reviewHistory,
      style: theme.textTheme.titleSmall,
    );
    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth >= wideLayoutWidth) {
          return Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              SizedBox(
                width: 360,
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: _Summary(info: info),
                ),
              ),
              const VerticalDivider(width: 1),
              Expanded(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      heading,
                      const SizedBox(height: 8),
                      _HistoryTable(history: info.history),
                    ],
                  ),
                ),
              ),
            ],
          );
        }
        return ListView(
          padding: const EdgeInsets.symmetric(vertical: 16),
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: _Summary(info: info),
            ),
            const SizedBox(height: 24),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: heading,
            ),
            if (info.history.isEmpty)
              const _NoHistory()
            else
              for (final (i, h) in info.history.indexed)
                _HistoryTile(key: Key('history-$i'), entry: h),
          ],
        );
      },
    );
  }
}

String _rating(int rating) =>
    rating >= 1 && rating <= 4 ? ratingNames(appL10n)[rating - 1] : '—';

String _kind(String kind) => kind.isEmpty ? kind : appL10n.cardKind(kind);

/// An answer time: "4.2 s" under a minute, otherwise as [formatSpent].
String _answerTime(Duration d) => d.inSeconds < 60
    ? appL10n.secondsDecimal(_oneDecimal(d.inMilliseconds / 1000))
    : formatSpent(d);

String _oneDecimal(double v) =>
    NumberFormat('0.0', appL10n.localeName).format(v);

Color? _ratingColor(BuildContext context, int rating) =>
    rating == 1 ? Theme.of(context).colorScheme.error : null;

class _Summary extends StatelessWidget {
  const _Summary({required this.info});

  final golang.CardInfo info;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final d = info;
    final l = context.l10n;
    final number = NumberFormat.decimalPattern(l.localeName);
    String opt<T>(T? v, String Function(T) f) => v == null ? '—' : f(v);
    String date(DateTime t) => formatDate(t);
    String dateTime(DateTime t) => formatDateTime(t);
    String interval(Duration d) => formatInterval(d);
    final rows = <(String, String)>[
      (l.infoAdded, formatDate(d.added)),
      (l.infoFirstReview, opt(d.firstReview, date)),
      (l.infoLatestReview, opt(d.latestReview, date)),
      (
        l.infoDue,
        d.state == 'new'
            ? newCardText(d.position, l.cardKind('new'))
            : opt(d.due, dateTime),
      ),
      (l.infoInterval, opt(d.interval, interval)),
      (l.infoStability, opt(d.stability, (s) => formatInterval(_days(s)))),
      (l.infoDifficulty, opt(d.difficulty, _oneDecimal)),
      (
        l.infoRetrievability,
        opt(
          d.retrievability,
          (r) => NumberFormat.percentPattern(l.localeName).format(r),
        ),
      ),
      (l.infoReviews, number.format(d.reviews)),
      (l.infoLapses, number.format(d.lapses)),
      (l.averageTime, opt(d.averageTime, _answerTime)),
      (l.totalTime, opt(d.totalTime, _answerTime)),
      (l.infoState, _kind(d.state)),
      (l.cardType, d.cardType),
      (l.noteType, d.noteType),
      (l.deck, d.deck),
      (l.cardId, '${d.cardID}'),
      (l.noteId, '${d.noteID}'),
    ];
    return Table(
      key: const Key('card-info-summary'),
      columnWidths: const {0: IntrinsicColumnWidth()},
      children: [
        for (final (k, v) in rows)
          TableRow(
            children: [
              Padding(
                padding: const EdgeInsets.only(right: 16, bottom: 6),
                child: Text(
                  k,
                  style: TextStyle(color: theme.colorScheme.onSurfaceVariant),
                ),
              ),
              Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: SelectableText(v),
              ),
            ],
          ),
      ],
    );
  }

  /// Stability is in days.
  static Duration _days(double days) =>
      Duration(minutes: (days * 24 * 60).round());
}

class _NoHistory extends StatelessWidget {
  const _NoHistory();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Text(
        context.l10n.noReviewsYet,
        style: TextStyle(color: Theme.of(context).colorScheme.onSurfaceVariant),
      ),
    );
  }
}

/// The history as a table with one column per value, for wide layouts.
class _HistoryTable extends StatelessWidget {
  const _HistoryTable({required this.history});

  final List<golang.HistoryEntry> history;

  @override
  Widget build(BuildContext context) {
    if (history.isEmpty) return const _NoHistory();
    final theme = Theme.of(context);
    Widget cell(String text, {TextStyle? style}) => Padding(
      padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 8),
      child: Text(text, style: style),
    );
    final header = theme.textTheme.labelMedium?.copyWith(
      color: theme.colorScheme.onSurfaceVariant,
    );
    return Table(
      defaultColumnWidth: const IntrinsicColumnWidth(),
      children: [
        TableRow(
          decoration: BoxDecoration(
            border: Border(bottom: BorderSide(color: theme.dividerColor)),
          ),
          children: [
            for (final h in [
              context.l10n.historyDate,
              context.l10n.historyKind,
              context.l10n.historyRating,
              context.l10n.historyInterval,
              context.l10n.historyTime,
            ])
              cell(h, style: header),
          ],
        ),
        for (final (i, h) in history.indexed)
          TableRow(
            key: ValueKey('history-$i'),
            children: [
              cell(formatDateTime(h.time)),
              cell(_kind(h.kind)),
              cell(
                _rating(h.rating),
                style: TextStyle(color: _ratingColor(context, h.rating)),
              ),
              cell(h.interval == null ? '—' : formatInterval(h.interval!)),
              cell(_answerTime(h.answerTime)),
            ],
          ),
      ],
    );
  }
}

/// One history entry as a two-line row, for narrow layouts.
class _HistoryTile extends StatelessWidget {
  const _HistoryTile({super.key, required this.entry});

  final golang.HistoryEntry entry;

  @override
  Widget build(BuildContext context) {
    final h = entry;
    final muted = TextStyle(
      color: Theme.of(context).colorScheme.onSurfaceVariant,
    );
    return ListTile(
      dense: true,
      title: Text.rich(
        TextSpan(
          children: [
            TextSpan(
              text: _rating(h.rating),
              style: TextStyle(color: _ratingColor(context, h.rating)),
            ),
            TextSpan(text: ' · ${_kind(h.kind)}'),
          ],
        ),
      ),
      subtitle: Text(formatDateTime(h.time)),
      trailing: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Text(h.interval == null ? '—' : formatInterval(h.interval!)),
          Text(_answerTime(h.answerTime), style: muted),
        ],
      ),
    );
  }
}
