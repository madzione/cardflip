import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/features/study/study_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// The kinds of custom study.
enum CustomStudyKind {
  moreNew,
  moreReviews,
  forgotten,
  ahead,
  previewNew,
  byState,
}

/// The name of the filtered deck the custom study sessions use.
const customStudySessionName = 'Custom Study Session';

/// Custom study for a deck: increase today's new or review card limit by a
/// number, without changing the deck's options; or study a session of
/// chosen cards in the filtered deck "Custom Study Session", which replaces
/// the previous session and starts at once.
class CustomStudyDialog extends StatefulWidget {
  const CustomStudyDialog({super.key, required this.deck});

  final golang.DeckNode deck;

  @override
  State<CustomStudyDialog> createState() => _CustomStudyDialogState();
}

class _CustomStudyDialogState extends State<CustomStudyDialog> {
  golang.CustomStudyInfo? _info;
  CustomStudyKind _kind = CustomStudyKind.moreNew;
  final _by = TextEditingController(text: '10');
  final _days = TextEditingController(text: '1');
  final _cards = TextEditingController(text: '100');
  final _tags = TextEditingController();
  String _state = 'new';
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final info = await golang.decks.customStudyInfo(widget.deck.id);
      if (mounted) setState(() => _info = info);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  @override
  void dispose() {
    _by.dispose();
    _days.dispose();
    _cards.dispose();
    _tags.dispose();
    super.dispose();
  }

  /// The whole number in [c] from 1 to [max], or null after showing what
  /// is wrong.
  int? _number(TextEditingController c, int max) {
    final n = int.tryParse(c.text.trim());
    if (n == null || n < 1 || n > max) {
      setState(() => _error = context.l10n.enterWholeNumber(max));
      return null;
    }
    return n;
  }

  Future<void> _save() async {
    if (_busy) return;
    switch (_kind) {
      case CustomStudyKind.moreNew || CustomStudyKind.moreReviews:
        final by = _number(_by, 9999);
        if (by == null) return;
        try {
          await golang.decks.increaseToday(
            widget.deck.id,
            _kind == CustomStudyKind.moreNew ? by : 0,
            _kind == CustomStudyKind.moreReviews ? by : 0,
          );
          if (mounted) Navigator.pop(context, true);
        } catch (e) {
          setState(() => _error = errorMessage(e));
        }
      case _:
        final choice = _choice();
        if (choice != null) await _session(choice);
    }
  }

  golang.CustomStudyChoice? _choice() {
    if (_kind == CustomStudyKind.byState) {
      final cards = _number(_cards, 9999);
      if (cards == null) return null;
      return golang.CustomStudyChoice(
        kind: 'byState',
        days: 0,
        state: _state,
        tags: _tags.text
            .split(RegExp(r'[\s,]+'))
            .where((t) => t.isNotEmpty)
            .toList(),
        cards: cards,
      );
    }
    final days = _number(_days, 365);
    if (days == null) return null;
    return golang.CustomStudyChoice(
      kind: switch (_kind) {
        CustomStudyKind.forgotten => 'forgotten',
        CustomStudyKind.ahead => 'ahead',
        _ => 'previewNew',
      },
      days: days,
      state: '',
      tags: const [],
      cards: 0,
    );
  }

  /// Builds the custom study session and starts studying it.
  Future<void> _session(golang.CustomStudyChoice choice) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final built = await golang.decks.customStudy(widget.deck.id, choice);
      if (!mounted) return;
      final nav = Navigator.of(context);
      nav.pop(true);
      nav.push(
        MaterialPageRoute<void>(
          builder: (_) => StudyScreen(
            deckId: built.deckID,
            deckName: customStudySessionName,
          ),
        ),
      );
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = errorMessage(e);
          _busy = false;
        });
      }
    }
  }

  String _available(int? n) => n == null ? '…' : context.l10n.moreAvailable(n);

  Widget _field() {
    switch (_kind) {
      case CustomStudyKind.moreNew || CustomStudyKind.moreReviews:
        return TextField(
          key: const Key('increaseBy'),
          controller: _by,
          keyboardType: TextInputType.number,
          decoration: InputDecoration(
            labelText: context.l10n.increaseBy,
            suffixText: context.l10n.unitCards,
            helperText: context.l10n.increaseHelper,
            helperMaxLines: 2,
          ),
          onSubmitted: (_) => _save(),
        );
      case CustomStudyKind.byState:
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            DropdownButtonFormField<String>(
              key: const Key('cardState'),
              initialValue: _state,
              isExpanded: true,
              decoration: InputDecoration(labelText: context.l10n.cards),
              items: [
                for (final (value, label) in [
                  ('new', context.l10n.newCardsOnly),
                  ('due', context.l10n.dueCardsOnly),
                  ('all', context.l10n.allCardsRandom),
                ])
                  DropdownMenuItem(
                    value: value,
                    child: Text(label, overflow: TextOverflow.ellipsis),
                  ),
              ],
              onChanged: (v) => setState(() => _state = v ?? _state),
            ),
            const SizedBox(height: 12),
            TextField(
              key: const Key('customTags'),
              controller: _tags,
              decoration: InputDecoration(
                labelText: context.l10n.onlyWithTags,
                helperText: context.l10n.onlyWithTagsHelper,
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              key: const Key('customCards'),
              controller: _cards,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(
                labelText: context.l10n.select,
                suffixText: context.l10n.unitCards,
              ),
              onSubmitted: (_) => _save(),
            ),
          ],
        );
      case _:
        return TextField(
          key: const Key('customDays'),
          controller: _days,
          keyboardType: TextInputType.number,
          decoration: InputDecoration(
            labelText: _kind == CustomStudyKind.ahead
                ? context.l10n.next
                : context.l10n.last,
            suffixText: context.l10n.unitDays,
            helperText: context.l10n.customStudySessionHelper(
              customStudySessionName,
            ),
            helperMaxLines: 2,
          ),
          onSubmitted: (_) => _save(),
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final info = _info;
    RadioListTile<CustomStudyKind> choice(
      CustomStudyKind kind,
      String title, [
      String? subtitle,
    ]) => RadioListTile(
      key: Key('custom-${kind.name}'),
      contentPadding: EdgeInsets.zero,
      dense: true,
      value: kind,
      title: Text(title),
      subtitle: subtitle == null ? null : Text(subtitle),
    );
    return AlertDialog(
      title: Text(context.l10n.customStudyTitle(widget.deck.name)),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            RadioGroup<CustomStudyKind>(
              groupValue: _kind,
              onChanged: (v) => setState(() {
                _kind = v ?? _kind;
                _error = null;
              }),
              child: Column(
                children: [
                  choice(
                    CustomStudyKind.moreNew,
                    context.l10n.increaseNewLimit,
                    _available(info?.newAvailable),
                  ),
                  choice(
                    CustomStudyKind.moreReviews,
                    context.l10n.increaseReviewLimit,
                    _available(info?.reviewAvailable),
                  ),
                  choice(
                    CustomStudyKind.forgotten,
                    context.l10n.reviewForgotten,
                  ),
                  choice(CustomStudyKind.ahead, context.l10n.reviewAhead),
                  choice(
                    CustomStudyKind.previewNew,
                    context.l10n.previewNewCards,
                  ),
                  choice(
                    CustomStudyKind.byState,
                    context.l10n.studyByStateOrTag,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            _field(),
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(
                _error!,
                key: const Key('customStudyError'),
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: _busy ? null : _save,
          child: Text(context.l10n.ok),
        ),
      ],
    );
  }
}
