import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'reposition_dialog.dart';

/// A browser column: its name in Go, its header, and its text in a row of
/// type [R] (a card's `golang.Row` or a note's `golang.NoteRow`).
class BrowseColumn<R> {
  const BrowseColumn(
    this.key,
    this.text, {
    this.numeric = false,
    this.flex = 2,
    this.sortable = true,
  });

  final String key;
  final String Function(R row) text;

  /// The header, in the interface language.
  String get label {
    final l = appL10n;
    return switch (key) {
      'question' => l.colQuestion,
      'answer' => l.colAnswer,
      'sortField' => l.colSortField,
      'cardType' => l.colCardType,
      'noteType' => l.colNoteType,
      'deck' => l.colDeck,
      'due' => l.colDue,
      'interval' => l.colInterval,
      'difficulty' => l.colDifficulty,
      'stability' => l.colStability,
      'retrievability' => l.colRetrievability,
      'reviews' => l.colReviews,
      'lapses' => l.colLapses,
      'tags' => l.colTags,
      'created' => l.colCreated,
      'cardModified' => l.colCardModified,
      'noteModified' => l.colNoteModified,
      'cards' => l.colCards,
      _ => key,
    };
  }

  /// Numbers are right-aligned in the table.
  final bool numeric;

  /// The column's share of the table's width.
  final int flex;

  /// Whether clicking the header sorts by the column.
  final bool sortable;
}

/// Every browser column, in the order the Columns dialog lists them.
const browseColumns = <BrowseColumn<golang.Row>>[
  BrowseColumn('question', _question, flex: 4),
  BrowseColumn('answer', _answer, flex: 4),
  BrowseColumn('sortField', _sortField, flex: 3),
  BrowseColumn('cardType', _cardType),
  BrowseColumn('noteType', _noteType),
  BrowseColumn('deck', _deck),
  BrowseColumn('due', dueText),
  BrowseColumn('interval', _interval, numeric: true),
  BrowseColumn('difficulty', _difficulty, numeric: true),
  BrowseColumn('stability', _stability, numeric: true),
  BrowseColumn('retrievability', _retrievability, numeric: true),
  BrowseColumn('reviews', _reviews, numeric: true, flex: 1),
  BrowseColumn('lapses', _lapses, numeric: true, flex: 1),
  BrowseColumn('tags', _tags, flex: 3),
  BrowseColumn('created', _created),
  BrowseColumn('cardModified', _cardModified),
  BrowseColumn('noteModified', _noteModified),
];

/// The column named [key]; unknown names fall back to the question.
BrowseColumn<golang.Row> browseColumn(String key) => browseColumns.firstWhere(
  (c) => c.key == key,
  orElse: () => browseColumns.first,
);

String _question(golang.Row r) =>
    r.question.isEmpty ? appL10n.emptyQuestion : r.question;
String _answer(golang.Row r) => r.answer;
String _sortField(golang.Row r) => r.sortField;
String _cardType(golang.Row r) => r.cardType;
String _noteType(golang.Row r) => r.noteType;
String _deck(golang.Row r) => r.deck;
String _tags(golang.Row r) => r.tags;
String _reviews(golang.Row r) => _count(r.reviews);
String _lapses(golang.Row r) => _count(r.lapses);
String _count(int n) =>
    NumberFormat.decimalPattern(appL10n.localeName).format(n);

/// "suspended", "New #42", or when a scheduled card is due.
String dueText(golang.Row r) {
  if (r.suspended) return appL10n.stateSuspended;
  if (r.state == 'new' || r.due == null) {
    return newCardText(r.position, appL10n.stateNew);
  }
  return appL10n.dueWhen(formatDue(r.due!));
}

/// A Review card's interval; "(new)", "(learning)" or "(relearning)" for
/// cards without one.
String _interval(golang.Row r) => r.interval == null
    ? appL10n.intervalOfState(r.state)
    : formatInterval(Duration(days: r.interval!));

String _percent(double? v) =>
    v == null ? '' : NumberFormat.percentPattern(appL10n.localeName).format(v);
String _difficulty(golang.Row r) => _percent(r.difficulty);
String _retrievability(golang.Row r) => _percent(r.retrievability);
String _stability(golang.Row r) => r.stability == null
    ? ''
    : formatInterval(Duration(minutes: (r.stability! * 24 * 60).round()));

/// The columns of Notes mode, which are fixed. Notes sort by
/// [noteSortColumns] only.
const noteColumns = <BrowseColumn<golang.NoteRow>>[
  BrowseColumn('sortField', _noteSortField, flex: 4),
  BrowseColumn('cards', _noteCards, flex: 1, sortable: false),
  BrowseColumn('noteType', _noteNoteType, sortable: false),
  BrowseColumn('deck', _noteDeck, sortable: false),
  BrowseColumn('tags', _noteTags, flex: 3, sortable: false),
  BrowseColumn('due', noteDueText),
];

/// What Notes mode sorts by, with the sort menu's labels.
Map<String, String> get noteSortColumns => {
  'created': appL10n.colCreated,
  'sortField': appL10n.colSortField,
  'due': appL10n.colDue,
};

String _noteSortField(golang.NoteRow r) =>
    r.sortField.isEmpty ? appL10n.emptyQuestion : r.sortField;
String _noteCards(golang.NoteRow r) => appL10n.countCards(r.cards);
String _noteNoteType(golang.NoteRow r) => r.noteType;
String _noteDeck(golang.NoteRow r) =>
    r.multipleDecks ? appL10n.multipleDecks : r.deck;
String _noteTags(golang.NoteRow r) => r.tags;

/// "suspended" when all of a note's cards are, else when its first card is
/// due, as [dueText] says it.
String noteDueText(golang.NoteRow r) {
  if (r.suspended) return appL10n.stateSuspended;
  if (r.state == 'new' || r.due == null) {
    return newCardText(r.position, appL10n.stateNew);
  }
  return appL10n.dueWhen(formatDue(r.due!));
}

String _date(DateTime? t) => t == null ? '' : formatDate(t);
String _created(golang.Row r) => _date(r.created);
String _cardModified(golang.Row r) => _date(r.cardModified);
String _noteModified(golang.Row r) => _date(r.noteModified);

/// The column header menu: move or hide [column], show another column next
/// to it, or open the Columns dialog. Returns the new columns, or null when
/// nothing changed.
Future<List<String>?> showColumnMenu(
  BuildContext context,
  Offset position,
  List<String> columns,
  String column,
) async {
  final overlay = Overlay.of(context).context.findRenderObject() as RenderBox;
  final i = columns.indexOf(column);
  final picked = await showMenu<String>(
    context: context,
    popUpAnimationStyle: popupAnimation(context),
    position: RelativeRect.fromRect(
      // The overlay is scaled with the UI size, the pointer is not.
      overlay.globalToLocal(position) & const Size(1, 1),
      Offset.zero & overlay.size,
    ),
    items: [
      PopupMenuItem(
        value: 'left',
        enabled: i > 0,
        child: Text(context.l10n.moveLeft),
      ),
      PopupMenuItem(
        value: 'right',
        enabled: i < columns.length - 1,
        child: Text(context.l10n.moveRight),
      ),
      PopupMenuItem(
        value: 'hide',
        enabled: columns.length > 1,
        child: Text(context.l10n.hideColumn),
      ),
      const PopupMenuDivider(),
      for (final c in browseColumns)
        CheckedPopupMenuItem(
          key: Key('show-column-${c.key}'),
          value: 'toggle:${c.key}',
          checked: columns.contains(c.key),
          // The last shown column cannot be hidden.
          enabled: !(columns.length == 1 && columns.contains(c.key)),
          child: Text(c.label),
        ),
      const PopupMenuDivider(),
      PopupMenuItem(value: 'dialog', child: Text(context.l10n.columnsEllipsis)),
    ],
  );
  if (picked == null || !context.mounted) return null;
  final out = [...columns];
  switch (picked) {
    case 'left':
      out.insert(i - 1, out.removeAt(i));
    case 'right':
      out.insert(i + 1, out.removeAt(i));
    case 'hide':
      out.remove(column);
    case 'dialog':
      return showColumnsDialog(context, columns);
    default:
      final key = picked.substring('toggle:'.length);
      // A new column goes next to the one whose menu it was.
      if (!out.remove(key)) out.insert(i + 1, key);
  }
  return out;
}

/// Lets the user show, hide and reorder the columns: shown columns first,
/// in order, by dragging. Returns the new columns, or null when cancelled.
Future<List<String>?> showColumnsDialog(
  BuildContext context,
  List<String> columns,
) => showDialog<List<String>>(
  context: context,
  animationStyle: popupAnimation(context),
  builder: (_) => _ColumnsDialog(columns),
);

class _ColumnsDialog extends StatefulWidget {
  const _ColumnsDialog(this.columns);

  final List<String> columns;

  @override
  State<_ColumnsDialog> createState() => _ColumnsDialogState();
}

class _ColumnsDialogState extends State<_ColumnsDialog> {
  late final List<String> _order = [
    ...widget.columns,
    for (final c in browseColumns)
      if (!widget.columns.contains(c.key)) c.key,
  ];
  late final Set<String> _shown = widget.columns.toSet();

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(context.l10n.columns),
      content: SizedBox(
        width: 360,
        height: 480,
        child: ReorderableListView(
          key: const Key('columns-list'),
          buildDefaultDragHandles: false,
          onReorderItem: (from, to) =>
              setState(() => _order.insert(to, _order.removeAt(from))),
          children: [
            for (final (i, key) in _order.indexed)
              CheckboxListTile(
                key: Key('columns-$key'),
                value: _shown.contains(key),
                // The last shown column cannot be hidden.
                onChanged: _shown.length == 1 && _shown.contains(key)
                    ? null
                    : (on) => setState(
                        () => on! ? _shown.add(key) : _shown.remove(key),
                      ),
                controlAffinity: ListTileControlAffinity.leading,
                title: Text(browseColumn(key).label),
                secondary: ReorderableDragStartListener(
                  index: i,
                  child: const Icon(Icons.drag_handle),
                ),
              ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        TextButton(
          onPressed: () => Navigator.pop(context, [
            for (final key in _order)
              if (_shown.contains(key)) key,
          ]),
          child: Text(context.l10n.save),
        ),
      ],
    );
  }
}
