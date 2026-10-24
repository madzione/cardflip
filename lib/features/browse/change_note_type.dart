import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Shows Change note type for the notes of [cardIds], which must share one
/// note type; otherwise it says so and returns null. It asks for the target
/// note type and the field and card type mapping, confirms with the counts
/// and applies the change itself. Returns the change made, or null when
/// cancelled.
Future<golang.NoteTypeChange?> showChangeNoteType(
  BuildContext context,
  List<int> cardIds,
) async {
  late final int source;
  late final List<golang.NoteType> types;
  final ok = await guard(() async {
    source = await golang.browse.noteTypeOfCards(cardIds);
    types = await golang.notes.types();
  });
  if (!ok || !context.mounted) return null;
  return showDialog<golang.NoteTypeChange>(
    context: context,
    animationStyle: popupAnimation(context),
    builder: (_) => _ChangeNoteTypeDialog(cardIds, types, source),
  );
}

/// For each name in [to], the index in [from] it takes by default: the same
/// name, else the name at the same position unless another takes it by
/// name, else -1. Each index of [from] is used at most once.
List<int> defaultMapping(List<String> from, List<String> to) {
  final byName = [for (final n in to) from.indexOf(n)];
  final taken = byName.where((i) => i >= 0).toSet();
  return [
    for (var i = 0; i < to.length; i++)
      if (byName[i] >= 0)
        byName[i]
      else if (i < from.length && !taken.contains(i))
        i
      else
        -1,
  ];
}

/// Says what a change does, for the confirmation.
String changeSummary(golang.NoteTypeChange change) {
  final l = appL10n;
  final (deleted, created) = (change.cardsDeleted, change.cardsCreated);
  final lines = [
    if (deleted > 0 && created > 0)
      l.changeCardsBoth(deleted, created)
    else if (deleted > 0)
      l.changeCardsDeleted(deleted)
    else if (created > 0)
      l.changeCardsCreated(created)
    else
      l.changeNoCards,
    if (change.lostFields.isNotEmpty)
      l.changeLostFields(_joinAnd(change.lostFields)),
  ];
  return lines.join('\n\n');
}

String _joinAnd(List<String> names) => names.length == 1
    ? names.single
    : appL10n.listAnd(
        names.sublist(0, names.length - 1).join(', '),
        names.last,
      );

class _ChangeNoteTypeDialog extends StatefulWidget {
  const _ChangeNoteTypeDialog(this.cardIds, this.types, this.sourceId);

  final List<int> cardIds;
  final List<golang.NoteType> types;
  final int sourceId;

  @override
  State<_ChangeNoteTypeDialog> createState() => _ChangeNoteTypeDialogState();
}

class _ChangeNoteTypeDialogState extends State<_ChangeNoteTypeDialog> {
  late final golang.NoteType _source = _type(widget.sourceId);

  /// The target starts as the first other note type: changing to the same
  /// one only reorders fields.
  late golang.NoteType _target = widget.types.firstWhere(
    (t) => t.id != _source.id,
    orElse: () => _source,
  );
  late List<int> _fieldMap;
  late List<int> _templateMap;
  String? _error;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _resetMaps();
  }

  golang.NoteType _type(int id) => widget.types.firstWhere((t) => t.id == id);

  /// Card types are mapped only between standard note types; otherwise
  /// cards keep their number.
  bool get _mapsTemplates => !_source.cloze && !_target.cloze;

  List<String> _names(golang.NoteType t) => [
    for (final tmpl in t.templates) tmpl.name,
  ];

  void _resetMaps() {
    _fieldMap = defaultMapping(_source.fields, _target.fields);
    _templateMap = _mapsTemplates
        ? defaultMapping(_names(_source), _names(_target))
        : const [];
  }

  void _setTemplate(int target, int source) => setState(() {
    _error = null;
    // A card type's cards can go to only one card type.
    _templateMap = [
      for (var i = 0; i < _templateMap.length; i++)
        i == target
            ? source
            : (source >= 0 && _templateMap[i] == source ? -1 : _templateMap[i]),
    ];
  });

  Future<void> _apply() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    final golang.NoteTypeChange preview;
    try {
      preview = await golang.browse.changeNoteTypePreview(
        widget.cardIds,
        _target.id,
        _fieldMap,
        _templateMap,
      );
    } on golang.GolangException catch (e) {
      // A mapping that cannot be applied: say why in the dialog.
      if (mounted) {
        setState(() {
          _busy = false;
          _error = goText(e.message);
        });
      }
      return;
    }
    if (!mounted) return;
    final ok = await confirm(
      context,
      title: context.l10n.changeNotesTitle(preview.notes, _target.name),
      message: changeSummary(preview),
      action: context.l10n.change,
      destructive: preview.cardsDeleted > 0 || preview.lostFields.isNotEmpty,
    );
    if (!mounted) return;
    if (!ok) {
      setState(() => _busy = false);
      return;
    }
    golang.NoteTypeChange? done;
    await guard(() async {
      done = await golang.browse.changeNoteType(
        widget.cardIds,
        _target.id,
        _fieldMap,
        _templateMap,
      );
    });
    if (!mounted) return;
    if (done != null) {
      Navigator.pop(context, done);
    } else {
      setState(() => _busy = false);
    }
  }

  Widget _mapping({
    required Key key,
    required String label,
    required List<String> sources,
    required int value,
    required ValueChanged<int> onChanged,
  }) => Padding(
    padding: const EdgeInsets.only(top: 8),
    // Rebuilt when the value changes from outside (a new target, or a card
    // type taken by another), as initialValue is read only once.
    child: KeyedSubtree(
      key: ValueKey('$key-${_target.id}-$value'),
      child: DropdownButtonFormField<int>(
        key: key,
        initialValue: value,
        isExpanded: true,
        decoration: InputDecoration(labelText: label),
        items: [
          for (var i = 0; i < sources.length; i++)
            DropdownMenuItem(value: i, child: Text(sources[i])),
          DropdownMenuItem(value: -1, child: Text(context.l10n.nothing)),
        ],
        onChanged: (v) => onChanged(v!),
      ),
    ),
  );

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final l = context.l10n;
    return AlertDialog(
      title: Text(l.changeNoteType),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(l.currentNoteType(_source.name)),
              const SizedBox(height: 8),
              DropdownButtonFormField<int>(
                key: const Key('change-note-type-target'),
                initialValue: _target.id,
                isExpanded: true,
                decoration: InputDecoration(labelText: l.newNoteType),
                items: [
                  for (final t in widget.types)
                    DropdownMenuItem(value: t.id, child: Text(t.name)),
                ],
                onChanged: (id) => setState(() {
                  _target = _type(id!);
                  _error = null;
                  _resetMaps();
                }),
              ),
              const SizedBox(height: 16),
              Text(l.fields, style: text.titleSmall),
              Text(l.changeFieldsHelp, style: text.bodySmall),
              for (var i = 0; i < _target.fields.length; i++)
                _mapping(
                  key: Key('change-field-$i'),
                  label: _target.fields[i],
                  sources: _source.fields,
                  value: _fieldMap[i],
                  onChanged: (v) => setState(() {
                    _error = null;
                    _fieldMap = [..._fieldMap]..[i] = v;
                  }),
                ),
              const SizedBox(height: 16),
              Text(l.cards, style: text.titleSmall),
              if (_mapsTemplates) ...[
                Text(l.changeCardsHelp, style: text.bodySmall),
                for (var i = 0; i < _target.templates.length; i++)
                  _mapping(
                    key: Key('change-template-$i'),
                    label: _target.templates[i].name,
                    sources: _names(_source),
                    value: _templateMap[i],
                    onChanged: (v) => _setTemplate(i, v),
                  ),
              ] else
                Text(l.changeClozeHelp, style: text.bodySmall),
              if (_error case final error?) ...[
                const SizedBox(height: 12),
                Text(
                  error,
                  key: const Key('change-note-type-error'),
                  style: text.bodyMedium?.copyWith(
                    color: Theme.of(context).colorScheme.error,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(l.cancel),
        ),
        FilledButton(
          key: const Key('change-note-type-apply'),
          onPressed: _busy ? null : _apply,
          child: Text(l.change),
        ),
      ],
    );
  }
}
