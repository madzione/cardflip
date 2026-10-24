import 'dart:async';
import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/scaffold.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/features/decks/filtered_deck_screen.dart';
import 'package:cardflip/features/editor/math_insert.dart' show MenuEntryLabel;
import 'package:cardflip/features/editor/note_editor_screen.dart';
import 'package:cardflip/features/export/export_dialog.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'browse_sidebar.dart';
import 'card_details.dart';
import 'change_note_type.dart';
import 'columns.dart';
import 'delete_cards.dart';
import 'find_duplicates.dart';
import 'find_replace.dart';
import 'flag_menu.dart';
import 'reposition_dialog.dart';
import 'reset_dialog.dart';
import 'saved_searches.dart';
import 'set_due_date_dialog.dart';

/// Rows are fetched from Go in pages of this size as the list scrolls.
const browsePageSize = 100;

/// The height of a row in the table (wide screens) and in the list (narrow
/// screens).
const _tableRowHeight = 40.0;
const _tileRowHeight = 72.0;

/// The narrowest and widest share of the width the list can be dragged to
/// next to the details.
const _minSplit = 0.2;
const _maxSplit = 0.8;

/// Finds, inspects and bulk-edits cards, or notes in Notes mode.
class BrowseScreen extends StatefulWidget {
  const BrowseScreen({super.key, this.noteIds, this.query});

  /// Shows only the cards of these notes (the editor's duplicates), until
  /// the user removes the filter. The search still applies within them.
  final List<int>? noteIds;

  /// The search the screen starts with, e.g. `added:1`; without it and
  /// [noteIds] it starts with the default browser search.
  final String? query;

  @override
  State<BrowseScreen> createState() => _BrowseScreenState();
}

/// How the list shows rows of type [R]: their columns, and the flag, mark
/// and suspension that tint them.
class _RowView<R> {
  const _RowView(
    this.columns,
    this.row,
    this.flag,
    this.marked,
    this.suspended,
  );

  final List<BrowseColumn<R>> columns;

  /// The row at index i, or null until its page is loaded.
  final R? Function(int i) row;
  final int Function(R row) flag;
  final bool Function(R row) marked;
  final bool Function(R row) suspended;

  /// The row's tint: flag, then mark, then suspension.
  Color? tint(BuildContext context, R row) => FlagColors.of(context)
      .rowTint(flag: flag(row), marked: marked(row), suspended: suspended(row));
}

class _BrowseScreenState extends State<BrowseScreen> {
  final _query = TextEditingController();
  String _sort = 'created';
  bool _descending = false;

  /// Whether the list has one row per note (Notes mode) instead of one per
  /// card. [_ids], [_selected] and [_details] are then note IDs.
  bool _notes = false;
  List<int> _ids = const [];
  final _rows = <int, golang.Row>{};
  final _noteRows = <int, golang.NoteRow>{};
  final _loadingPages = <int>{};
  final _selected = <int>{};
  final _sideways = ScrollController();
  final _vertical = ScrollController();
  final _scaffold = GlobalKey<ScaffoldState>();
  final _pane = GlobalKey<CardDetailsPaneState>();
  final _searchFocus = FocusNode();
  final _listFocus = FocusNode();

  /// Whether the last build had the wide layout.
  bool _wide = false;

  /// The share of the width the list takes next to the details.
  double _split = 0.6;

  /// Where Shift+Up/Down started extending the selection, and the selection
  /// before that; null when the next Shift+Up/Down starts afresh.
  int? _anchor;
  Set<int> _anchorBase = const {};

  /// Whether rows are being picked for bulk actions. It stays on when the
  /// last row is deselected, until the user leaves it.
  bool _selecting = false;

  /// The row whose card the details pane shows, and that card: the note's
  /// first card in Notes mode. The arrow keys move this current row.
  int? _details;
  int? _detailsCard;
  String? _searchError;
  bool _searching = true;
  Timer? _debounce;
  StreamSubscription<golang.Change>? _changes;
  int _generation = 0;
  SettingsModel? _settings;

  /// The columns the rows in [_rows] were loaded with.
  List<String> _rowColumns = SettingsModel.defaultBrowserColumns;

  /// See [BrowseScreen.noteIds]; null once the filter is removed.
  List<int>? _onlyNotes;

  @override
  void initState() {
    super.initState();
    _onlyNotes = widget.noteIds;
    _query.text = widget.query ?? '';
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _changes ??= AppScope.of(context).changes.collection
        .listen((_) => _search());
    if (_settings == null) {
      _settings = AppScope.of(context).settings..addListener(_onSettings);
      if (widget.query == null && widget.noteIds == null) {
        _query.text = _settings!.defaultSearch;
      }
      _rowColumns = _settings!.browserColumns;
      _notes = _settings!.browserMode == 'notes';
      _split = _settings!.browserSplit.clamp(_minSplit, _maxSplit);
      _search();
    }
  }

  /// Reloads the rows when the columns or the mode change.
  void _onSettings() {
    final notes = _settings!.browserMode == 'notes';
    if (notes != _notes) {
      _switchMode(notes);
      return;
    }
    final columns = _settings!.browserColumns;
    if (listEquals(columns, _rowColumns)) return;
    setState(() {
      _rowColumns = columns;
      _rows.clear();
      _loadingPages.clear();
    });
  }

  void _setColumns(List<String>? columns) {
    if (columns != null) _settings!.setBrowserColumns(columns);
  }

  /// Lists notes or cards; the choice is remembered.
  void _setMode(bool notes) {
    if (notes == _notes) return;
    _switchMode(notes);
    _settings!.setBrowserMode(notes ? 'notes' : 'cards');
  }

  void _switchMode(bool notes) {
    setState(() {
      _notes = notes;
      // Notes sort by fewer columns.
      if (notes && !noteSortColumns.containsKey(_sort)) {
        _sort = 'created';
        _descending = false;
      }
      _ids = const [];
      _rows.clear();
      _noteRows.clear();
      _loadingPages.clear();
      _selected.clear();
      _anchor = null;
      _details = null;
      _detailsCard = null;
      _searching = true;
    });
    _search();
  }

  /// Sorts by [column], or turns the order around when it is the sort
  /// column already.
  void _sortBy(String column) {
    setState(() {
      if (column == _sort) {
        _descending = !_descending;
      } else {
        _sort = column;
        _descending = false;
      }
    });
    _search();
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _changes?.cancel();
    _settings?.removeListener(_onSettings);
    _query.dispose();
    _sideways.dispose();
    _vertical.dispose();
    _searchFocus.dispose();
    _listFocus.dispose();
    super.dispose();
  }

  Future<void> _search() async {
    final generation = ++_generation;
    final notes = _notes;
    try {
      var ids = notes
          ? await golang.browse.searchNotes(_query.text, _sort, _descending)
          : await golang.browse.search(_query.text, _sort, _descending);
      if (_onlyNotes case final only?) {
        final keep = notes ? only.toSet() : await _cardsOf(only);
        ids = [
          for (final id in ids)
            if (keep.contains(id)) id,
        ];
      }
      if (!mounted || generation != _generation) return;
      setState(() {
        _ids = ids;
        _rows.clear();
        _noteRows.clear();
        _loadingPages.clear();
        _selected.retainAll(ids.toSet());
        _anchor = null;
        if (_details != null && !ids.contains(_details)) {
          _details = _detailsCard = null;
        }
        _searchError = null;
        _searching = false;
      });
    } on golang.GolangException catch (e) {
      // An invalid term: show why instead of an empty list.
      if (!mounted || generation != _generation) return;
      setState(() {
        // Go names the term; keep its spelling (e.g. "is:bogus: …").
        _searchError = goText(e.message);
        _searching = false;
      });
    }
  }

  /// The cards of [notes]; notes deleted meanwhile have none.
  static Future<Set<int>> _cardsOf(List<int> notes) async {
    final cards = await Future.wait([
      for (final id in notes)
        golang.notes
            .get_(id)
            .then((n) => n.cardIDs, onError: (Object _) => const <int>[]),
    ]);
    return {for (final c in cards) ...c};
  }

  void _removeNoteFilter() {
    setState(() => _onlyNotes = null);
    _search();
  }

  /// Searches for a sidebar item, combined with the search as [how] says.
  void _sidebarSearch(String term, SidebarSearch how) {
    _query.text = composeSearch(_query.text, term, how);
    _debounce?.cancel();
    _scaffold.currentState?.closeDrawer();
    _search();
  }

  void _runSavedSearch(String query) {
    _query.text = query;
    _debounce?.cancel();
    _search();
  }

  void _onQueryChanged(String _) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 300), _search);
  }

  void _loadPage(int page) {
    if (_loadingPages.contains(page)) return;
    _loadingPages.add(page);
    final generation = _generation;
    final columns = _rowColumns;
    final start = page * browsePageSize;
    final ids = _ids.sublist(start, min(start + browsePageSize, _ids.length));
    void failed(Object e) {
      _loadingPages.remove(page);
      showError(e);
    }

    if (_notes) {
      golang.browse.noteRows(ids).then((rows) {
        if (!mounted || generation != _generation) return;
        setState(() {
          for (final r in rows) {
            _noteRows[r.noteID] = r;
          }
          // The keys may have moved to a note before its row was there.
          if (_details case final id? when _detailsCard == null) {
            _detailsCard = _cardOf(id);
          }
        });
      }, onError: failed);
      return;
    }
    golang.browse.rows(ids, columns).then((rows) {
      if (!mounted ||
          generation != _generation ||
          !identical(columns, _rowColumns)) {
        return;
      }
      setState(() {
        for (final r in rows) {
          _rows[r.cardID] = r;
        }
      });
    }, onError: failed);
  }

  /// The card whose details row [id] opens: the card itself, or a note's
  /// first card.
  int? _cardOf(int id) => _notes ? _noteRows[id]?.cardID : id;

  void _open(int id, bool wide) {
    _listFocus.requestFocus();
    if (_selecting) {
      _toggle(id);
      return;
    }
    final card = _cardOf(id);
    if (card == null) return;
    if (wide) {
      setState(() {
        _anchor = null;
        _details = id;
        _detailsCard = card;
      });
    } else {
      _showDetails(card);
    }
  }

  void _showDetails(int card) => Navigator.of(context).push(
    MaterialPageRoute<void>(builder: (_) => CardDetailsScreen(cardId: card)),
  );

  void _toggle(int id) => setState(() {
    _selecting = true;
    _anchor = null;
    _selected.contains(id) ? _selected.remove(id) : _selected.add(id);
  });

  void _endSelecting() => setState(() {
    _selecting = false;
    _anchor = null;
    _selected.clear();
  });

  void _selectAll() => setState(() {
    _anchor = null;
    _selected.addAll(_ids);
  });

  /// Selects exactly the listed rows that were not selected.
  void _invertSelection() => setState(() {
    _selecting = true;
    _anchor = null;
    final before = _selected.toSet();
    _selected
      ..clear()
      ..addAll(_ids.where((id) => !before.contains(id)));
  });

  /// Adds every listed card of the notes of the selected cards to the
  /// selection. In Notes mode the rows are notes already.
  Future<void> _selectNotes() async {
    if (_notes || _selected.isEmpty) return;
    final generation = _generation;
    late final List<int> cards;
    if (!await guard(
      () async => cards = await golang.browse.cardsOfNotes(_selected.toList()),
    )) {
      return;
    }
    if (!mounted || generation != _generation || _notes) return;
    final listed = _ids.toSet();
    setState(() {
      _anchor = null;
      _selected.addAll(cards.where(listed.contains));
    });
  }

  /// The cards of [rows] (the selected rows by default): in Notes mode, all
  /// cards of those notes.
  Future<List<int>> _selectedCards([Iterable<int>? rows]) async {
    final ids = (rows ?? _selected).toList();
    return _notes ? golang.browse.noteCards(ids) : ids;
  }

  /// The index of the current row in the list, or null.
  int? get _currentIndex {
    final i = _details == null ? -1 : _ids.indexOf(_details!);
    return i < 0 ? null : i;
  }

  /// Moves the current row by [delta] rows, from the first row when there
  /// is none.
  void _move(int delta, {bool extend = false}) {
    final i = _currentIndex;
    _goTo(i == null ? 0 : i + delta, extend: extend);
  }

  /// Makes the row at [index] current and shows its details. With [extend]
  /// the selection becomes what it was when Shift+Up/Down started, plus the
  /// rows from there to [index].
  void _goTo(int index, {bool extend = false}) {
    if (_searching || _ids.isEmpty) return;
    index = index.clamp(0, _ids.length - 1);
    final current = _currentIndex;
    setState(() {
      if (!extend) {
        _anchor = null;
      } else {
        if (_anchor == null || !_ids.contains(_anchor)) {
          _anchor = _ids[current ?? index];
          _anchorBase = _selected.toSet();
        }
        final from = _ids.indexOf(_anchor!);
        _selecting = true;
        _selected
          ..clear()
          ..addAll(_anchorBase)
          ..addAll(_ids.sublist(min(from, index), max(from, index) + 1));
      }
      _details = _ids[index];
      _detailsCard = _cardOf(_ids[index]);
    });
    _reveal(index);
  }

  /// Scrolls the list just enough to show the row at [index].
  void _reveal(int index) {
    if (!_vertical.hasClients) return;
    final height = _wide ? _tableRowHeight : _tileRowHeight;
    final p = _vertical.position;
    final top = index * height;
    if (top < p.pixels) {
      _vertical.jumpTo(top);
    } else if (top + height > p.pixels + p.viewportDimension) {
      _vertical.jumpTo(top + height - p.viewportDimension);
    }
  }

  /// Opens the editor of the current card: the details' editor on wide
  /// screens, the details screen on narrow ones.
  void _editCurrent() {
    final card = _detailsCard;
    if (card == null) return;
    if (_wide) {
      _pane.currentState?.edit();
    } else {
      _showDetails(card);
    }
  }

  /// Deletes the notes of the selection, or of the current row when nothing
  /// is selected, after confirmation.
  void _deleteFromKeys() {
    if (_selected.isNotEmpty) {
      _bulk('delete');
    } else if (_details case final id?) {
      _bulk('delete', [id]);
    }
  }

  /// Sets the due date of the selection, or of the current row when
  /// nothing is selected.
  void _setDueDateFromKeys() {
    if (_selected.isNotEmpty) {
      _bulk('setDueDate');
    } else if (_details case final id?) {
      _bulk('setDueDate', [id]);
    }
  }

  void _focusSearch() {
    _searchFocus.requestFocus();
    _query.selection = TextSelection(
      baseOffset: 0,
      extentOffset: _query.text.length,
    );
  }

  Future<void> _setFlag(int flag) async {
    await guard(
      () async => golang.browse.setFlag(await _selectedCards(), flag),
    );
  }

  /// Runs a bulk [action] on the cards of [rows], the selected rows by
  /// default.
  Future<void> _bulk(String action, [Iterable<int>? rows]) async {
    final undo = AppScope.read(context).undo;
    late final List<int> ids;
    if (!await guard(() async => ids = await _selectedCards(rows))) return;
    if (!mounted) return;
    final l = context.l10n;
    switch (action) {
      case 'suspend':
        await guard(() => golang.browse.setSuspended(ids, true));
      case 'unsuspend':
        await guard(() => golang.browse.setSuspended(ids, false));
      case 'move':
        final deck = await _pickDeck();
        if (deck != null && await guard(() => golang.browse.move(ids, deck))) {
          undo.offer(l.movedCards(ids.length));
        }
      case 'toggleMark':
        await guard(() => golang.browse.toggleMark(ids));
      case 'addTags':
        final tags = await askText(
          context,
          title: l.addTags,
          label: l.tagsSeparated,
          action: l.add,
        );
        if (tags != null) await guard(() => golang.browse.addTags(ids, [tags]));
      case 'removeTags':
        final tags = await askText(
          context,
          title: l.removeTags,
          label: l.tagsSeparated,
          action: l.remove,
        );
        if (tags != null) {
          await guard(() => golang.browse.removeTags(ids, [tags]));
        }
      case 'reset':
        final opts = await askReset(context, cards: ids.length);
        if (opts != null &&
            await guard(() => golang.browse.resetToNew(ids, opts))) {
          undo.offer(l.resetCardsToNew(ids.length));
        }
      case 'setDueDate':
        final done = await askSetDueDate(
          context,
          cards: ids.length,
          apply: (spec) async {
            await golang.browse.setDueDate(ids, spec);
            return true;
          },
        );
        if (done == true) {
          undo.offer(l.setDueDateOfCards(ids.length));
        }
      case 'reposition':
        final r = await askReposition(context);
        if (r == null) return;
        var skipped = 0;
        final ok = await guard(() async {
          skipped = await golang.browse.reposition(
            ids,
            r.start,
            r.step,
            r.randomize,
            r.shift,
          );
        });
        if (!ok) return;
        final moved = ids.length - skipped;
        final notNew = skipped == 0 ? '' : l.cardsNotNew(skipped);
        if (moved == 0) {
          showInfo('${l.noNewCardsSelected}$notNew');
        } else {
          undo.offer('${l.repositionedCards(moved)}$notNew');
        }
      case 'findReplace':
        await _findReplace(ids);
      case 'changeNoteType':
        final done = await showChangeNoteType(context, ids);
        if (done != null) {
          _endSelecting();
          undo.offer(l.changedNoteTypeOf(done.notes));
        }
      case 'export':
        await showExportDialog(context, NotesExport(ids));
      case 'delete':
        final impact = await confirmDeleteCards(context, ids);
        if (impact != null &&
            await guard(() => golang.browse.deleteNotes(ids))) {
          _endSelecting();
          undo.offer(l.deletedCards(impact.cards));
        }
    }
  }

  /// Find and Replace in the notes of [ids], or in all notes.
  Future<void> _findReplace(List<int> ids) async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final changed = await showFindReplace(context, ids);
    if (changed == null) return;
    if (changed == 0) {
      showInfo(l.noNotesChanged);
    } else {
      undo.offer(l.changedNotes(changed));
    }
  }

  Future<int?> _pickDeck() {
    final decks = AppScope.of(context).deckList.decks;
    return showDialog<int>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => SimpleDialog(
        title: Text(context.l10n.moveToDeck),
        children: [
          // Filtered decks gather their cards with their search.
          for (final d in decks)
            if (!d.filtered)
              SimpleDialogOption(
                onPressed: () => Navigator.pop(context, d.id),
                child: Padding(
                  padding: EdgeInsets.only(left: d.depth * 16.0),
                  child: Text(d.name),
                ),
              ),
        ],
      ),
    );
  }

  PreferredSizeWidget _appBar(bool wide) {
    final l = context.l10n;
    final narrow = narrowScreen(context);
    if (_selecting) {
      final picked = _selected.isNotEmpty;
      return AppBar(
        leading: IconButton(
          tooltip: l.clearSelection,
          icon: const Icon(Icons.close),
          onPressed: _endSelecting,
        ),
        title: Text(
          picked
              ? l.countSelected(_selected.length)
              : _notes
              ? l.selectNotes
              : l.selectCards,
        ),
        // Where the app bar is narrow, Select all and Move to deck are in
        // the menus.
        actions: [
          if (!narrow)
            IconButton(
              tooltip: l.selectAll,
              icon: const Icon(Icons.select_all),
              onPressed: _selectAll,
            ),
          PopupMenuButton<VoidCallback>(
            popUpAnimationStyle: popupAnimation(context),
            tooltip: l.selection,
            icon: const Icon(Icons.rule),
            onSelected: (run) => run(),
            itemBuilder: (context) => [
              if (narrow)
                PopupMenuItem(value: _selectAll, child: Text(l.selectAll)),
              PopupMenuItem(
                value: _invertSelection,
                child: MenuEntryLabel(
                  l.invertSelection,
                  invertSelectionShortcutLabel,
                ),
              ),
              if (!_notes)
                PopupMenuItem(
                  value: _selectNotes,
                  enabled: picked,
                  child: Text(l.selectNotes),
                ),
            ],
          ),
          if (!narrow)
            IconButton(
              tooltip: l.moveToDeck,
              icon: const Icon(Icons.drive_file_move_outline),
              onPressed: picked ? () => _bulk('move') : null,
            ),
          FlagMenuButton(enabled: picked, onSelected: _setFlag),
          PopupMenuButton<String>(
            popUpAnimationStyle: popupAnimation(context),
            tooltip: l.bulkActions,
            enabled: picked,
            onSelected: _bulk,
            itemBuilder: (_) => [
              if (narrow)
                PopupMenuItem(value: 'move', child: Text(l.moveToDeck)),
              PopupMenuItem(value: 'suspend', child: Text(l.suspend)),
              PopupMenuItem(value: 'unsuspend', child: Text(l.unsuspend)),
              PopupMenuItem(value: 'toggleMark', child: Text(l.toggleMark)),
              PopupMenuItem(value: 'addTags', child: Text(l.addTags)),
              PopupMenuItem(value: 'removeTags', child: Text(l.removeTags)),
              PopupMenuItem(
                value: 'findReplace',
                child: Text(l.findAndReplace),
              ),
              PopupMenuItem(
                value: 'changeNoteType',
                child: Text(l.changeNoteType),
              ),
              PopupMenuItem(
                value: 'setDueDate',
                child: Text(
                  withShortcut(l.setDueDate, setDueDateShortcutLabel),
                ),
              ),
              PopupMenuItem(value: 'reposition', child: Text(l.reposition)),
              PopupMenuItem(value: 'reset', child: Text(l.resetToNew)),
              PopupMenuItem(value: 'export', child: Text(l.export)),
              PopupMenuItem(
                value: 'delete',
                child: Text(_notes ? l.deleteNotesAction : l.deleteCardsAction),
              ),
            ],
          ),
        ],
      );
    }
    // On narrow screens the sidebar is a drawer, opened by the leading
    // button, or by a Sidebar action where that is the back button.
    final canPop = Navigator.of(context).canPop();
    // Shown as icons, or in the menu where the app bar is narrow.
    final more = [
      if (!_notes)
        (
          l.columns,
          Icons.view_column_outlined,
          () async =>
              _setColumns(await showColumnsDialog(context, _rowColumns)),
        ),
      (l.select, Icons.checklist, () => setState(() => _selecting = true)),
      (
        l.addCard,
        Icons.add,
        () => Navigator.of(context).push(
          MaterialPageRoute<void>(builder: (_) => const NoteEditorScreen()),
        ),
      ),
    ];
    return AppBar(
      leading: !wide && canPop ? const BackButton() : null,
      title: Text(l.browse),
      actions: [
        if (!wide && canPop)
          IconButton(
            tooltip: l.sidebar,
            icon: const Icon(Icons.menu),
            onPressed: () => _scaffold.currentState?.openDrawer(),
          ),
        PopupMenuButton<String>(
          popUpAnimationStyle: popupAnimation(context),
          tooltip: l.sort,
          icon: const Icon(Icons.sort),
          onSelected: _sortBy,
          itemBuilder: (_) => [
            for (final (key, label)
                in _notes
                    ? [
                        for (final e in noteSortColumns.entries)
                          (e.key, e.value),
                      ]
                    : [for (final c in browseColumns) (c.key, c.label)])
              CheckedPopupMenuItem(
                key: Key('sort-$key'),
                value: key,
                checked: key == _sort,
                child: Text(label),
              ),
          ],
        ),
        if (!narrow)
          for (final (label, icon, action) in more)
            IconButton(tooltip: label, icon: Icon(icon), onPressed: action),
        UndoMenuButton(
          actions: [
            if (narrow)
              for (final (label, _, action) in more) (label, action),
            (l.findAndReplace, () => _findReplace(const [])),
            (
              l.createFilteredDeck,
              () => createFilteredDeck(context, search: _query.text.trim()),
            ),
            (
              l.findDuplicates,
              () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => const FindDuplicatesScreen(),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _list(bool wide) {
    if (_searching) return const Center(child: CircularProgressIndicator());
    if (_ids.isEmpty) {
      return Center(
        child: Text(
          _query.text.isEmpty && _onlyNotes == null
              ? context.l10n.noCardsYet
              : _notes
              ? context.l10n.noNotesMatch
              : context.l10n.noCardsMatch,
        ),
      );
    }
    if (_notes) {
      final view = _RowView<golang.NoteRow>(
        noteColumns,
        (i) => _row(_noteRows, i),
        (_) => 0,
        (r) => r.marked,
        (r) => r.suspended,
      );
      return wide ? _table(view) : _tiles(view);
    }
    final view = _RowView<golang.Row>(
      [for (final key in _rowColumns) browseColumn(key)],
      (i) => _row(_rows, i),
      (r) => r.flag,
      (r) => r.marked,
      (r) => r.suspended,
    );
    return wide ? _table(view) : _tiles(view);
  }

  /// The row at index [i] from [rows], loading its page when needed; null
  /// until it is there.
  R? _row<R>(Map<int, R> rows, int i) {
    final row = rows[_ids[i]];
    if (row == null) _loadPage(i ~/ browsePageSize);
    return row;
  }

  Widget? _flagMarker(int id, int flag) {
    if (flag == 0) return null;
    return Tooltip(
      message: context.l10n.flagNamed(
        AppScope.of(context).settings.flagName(flag),
      ),
      child: Icon(
        Icons.flag,
        key: Key('row-flag-$id'),
        color: FlagColors.of(context).flag(flag),
      ),
    );
  }

  /// Narrow screens: the first column is a row's title, the next two its
  /// subtitle.
  Widget _tiles<R>(_RowView<R> view) {
    final columns = view.columns;
    return ListView.builder(
      key: const Key('browse-list'),
      controller: _vertical,
      itemCount: _ids.length,
      itemExtent: _tileRowHeight,
      itemBuilder: (context, i) {
        final id = _ids[i];
        final row = view.row(i);
        if (row == null) return const ListTile(title: Text('…'));
        final selected = _selected.contains(id);
        final subtitle = [for (final c in columns.skip(1).take(2)) c.text(row)]
            .where((t) => t.isNotEmpty)
            .join(' · ');
        return ListTile(
          key: Key('row-$id'),
          selected: selected || _details == id,
          tileColor: view.tint(context, row),
          trailing: _flagMarker(id, view.flag(row)),
          leading: _selecting
              ? Checkbox(value: selected, onChanged: (_) => _toggle(id))
              : null,
          title: Text(
            columns.first.text(row),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          subtitle: subtitle.isEmpty
              ? null
              : Text(subtitle, maxLines: 1, overflow: TextOverflow.ellipsis),
          onTap: () => _open(id, false),
          onLongPress: () => _toggle(id),
        );
      },
    );
  }

  /// Wide screens: a table with a header per column. Clicking a header
  /// sorts by it; right-clicking or long-pressing it opens the column menu
  /// (cards only: Notes mode has fixed columns). Columns share the width;
  /// when they do not fit, the table scrolls sideways.
  Widget _table<R>(_RowView<R> view) {
    final columns = view.columns;
    const flagWidth = 32.0;
    final checkWidth = _selecting ? 48.0 : 0.0;
    double minWidth(BrowseColumn<R> c) => 48.0 + 24 * c.flex;
    return LayoutBuilder(
      builder: (context, constraints) {
        final fixed = flagWidth + checkWidth;
        final needed = columns.fold(fixed, (w, c) => w + minWidth(c));
        final width = max(needed, constraints.maxWidth);
        final flexes = columns.fold(0, (n, c) => n + c.flex);
        final widths = [
          for (final c in columns)
            needed >= constraints.maxWidth
                ? minWidth(c)
                : (width - fixed) * c.flex / flexes,
        ];
        final scheme = Theme.of(context).colorScheme;
        final textTheme = Theme.of(context).textTheme;

        Widget cell(BrowseColumn<R> c, double w, Widget child) => SizedBox(
          width: w,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: Align(
              alignment: c.numeric
                  ? Alignment.centerRight
                  : Alignment.centerLeft,
              child: child,
            ),
          ),
        );

        final header = Material(
          color: scheme.surfaceContainerHigh,
          child: SizedBox(
            height: 40,
            child: Row(
              children: [
                SizedBox(width: fixed),
                for (final (i, c) in columns.indexed)
                  GestureDetector(
                    key: Key('column-${c.key}'),
                    onLongPressStart: _notes
                        ? null
                        : (d) => _columnMenu(d.globalPosition, c.key),
                    child: InkWell(
                      onTap: c.sortable ? () => _sortBy(c.key) : null,
                      onSecondaryTapUp: _notes
                          ? null
                          : (d) => _columnMenu(d.globalPosition, c.key),
                      child: cell(
                        c,
                        widths[i],
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Flexible(
                              child: Text(
                                c.label,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: textTheme.labelLarge,
                              ),
                            ),
                            if (c.key == _sort)
                              Icon(
                                _descending
                                    ? Icons.arrow_downward
                                    : Icons.arrow_upward,
                                key: const Key('sort-indicator'),
                                size: 16,
                              ),
                          ],
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        );

        final list = ListView.builder(
          key: const Key('browse-list'),
          controller: _vertical,
          itemCount: _ids.length,
          itemExtent: _tableRowHeight,
          itemBuilder: (context, i) {
            final id = _ids[i];
            final row = view.row(i);
            if (row == null) {
              return const Align(
                alignment: Alignment.centerLeft,
                child: Padding(
                  padding: EdgeInsets.only(left: flagWidth),
                  child: Text('…'),
                ),
              );
            }
            final selected = _selected.contains(id) || _details == id;
            final flag = view.flag(row);
            return Material(
              key: Key('row-$id'),
              color: selected
                  ? scheme.secondaryContainer
                  : view.tint(context, row) ?? Colors.transparent,
              child: InkWell(
                onTap: () => _open(id, true),
                onLongPress: () => _toggle(id),
                child: Row(
                  children: [
                    if (_selecting)
                      SizedBox(
                        width: checkWidth,
                        child: Checkbox(
                          value: _selected.contains(id),
                          onChanged: (_) => _toggle(id),
                        ),
                      ),
                    SizedBox(
                      width: flagWidth,
                      child: _flagMarker(id, flag) ?? const SizedBox(),
                    ),
                    for (final (i, c) in columns.indexed)
                      cell(
                        c,
                        widths[i],
                        Text(
                          c.text(row),
                          key: Key('cell-$id-${c.key}'),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                  ],
                ),
              ),
            );
          },
        );

        final table = SizedBox(
          width: width,
          child: Column(
            children: [
              header,
              const Divider(height: 1),
              Expanded(child: list),
            ],
          ),
        );
        if (width <= constraints.maxWidth) return table;
        return Scrollbar(
          controller: _sideways,
          child: SingleChildScrollView(
            controller: _sideways,
            scrollDirection: Axis.horizontal,
            child: table,
          ),
        );
      },
    );
  }

  Future<void> _columnMenu(Offset position, String column) async =>
      _setColumns(await showColumnMenu(context, position, _rowColumns, column));

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final wide = _wide = constraints.maxWidth >= wideLayoutWidth;
        final list = Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
              child: TextField(
                key: const Key('search'),
                controller: _query,
                focusNode: _searchFocus,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: SavedSearchesButton(
                    query: () => _query.text,
                    onRun: _runSavedSearch,
                  ),
                  hintText: context.l10n.searchHint,
                  errorText: _searchError,
                  border: const OutlineInputBorder(),
                  isDense: true,
                ),
                onChanged: _onQueryChanged,
                onSubmitted: (_) {
                  _debounce?.cancel();
                  _search();
                },
              ),
            ),
            if (_onlyNotes case final notes?)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: InputChip(
                    key: const Key('browse-note-filter'),
                    label: Text(context.l10n.onlyNotes(notes.length)),
                    deleteButtonTooltipMessage: _notes
                        ? context.l10n.showAllNotes
                        : context.l10n.showAllCards,
                    onDeleted: _removeNoteFilter,
                  ),
                ),
              ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 4, 12, 4),
              child: LayoutBuilder(
                builder: (context, constraints) => Row(
                  children: [
                    Expanded(
                      child: _searchError == null && !_searching
                          ? Text(
                              _notes
                                  ? context.l10n.countNotes(_ids.length)
                                  : context.l10n.countCards(_ids.length),
                              style: Theme.of(context).textTheme.bodySmall,
                            )
                          : const SizedBox(),
                    ),
                    // Scaled down where the row is too narrow, as on a
                    // phone with a large UI size.
                    ConstrainedBox(
                      constraints: BoxConstraints(
                        maxWidth: constraints.maxWidth * 0.7,
                      ),
                      child: FittedBox(
                        fit: BoxFit.scaleDown,
                        child: SegmentedButton<bool>(
                          key: const Key('browse-mode'),
                          showSelectedIcon: false,
                          style: const ButtonStyle(
                            visualDensity: VisualDensity.compact,
                            tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                          ),
                          segments: [
                            ButtonSegment(
                              value: false,
                              label: Text(context.l10n.modeCards),
                              tooltip: context.l10n.oneRowPerCard,
                            ),
                            ButtonSegment(
                              value: true,
                              label: Text(context.l10n.modeNotes),
                              tooltip: context.l10n.oneRowPerNote,
                            ),
                          ],
                          selected: {_notes},
                          onSelectionChanged: (s) => _setMode(s.single),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Expanded(child: _listKeys(_list(wide))),
          ],
        );
        final sidebar = BrowseSidebar(onSearch: _sidebarSearch);
        return Shortcuts(
          shortcuts: {
            findShortcut: const _FocusSearchIntent(),
            focusListShortcut: const _FocusListIntent(),
            invertSelectionShortcut: const _InvertSelectionIntent(),
            setDueDateShortcut: const _SetDueDateIntent(),
          },
          child: Actions(
            actions: {
              _FocusSearchIntent: CallbackAction<_FocusSearchIntent>(
                onInvoke: (_) => _focusSearch(),
              ),
              _FocusListIntent: CallbackAction<_FocusListIntent>(
                onInvoke: (_) => _listFocus.requestFocus(),
              ),
              _InvertSelectionIntent: CallbackAction<_InvertSelectionIntent>(
                onInvoke: (_) => _invertSelection(),
              ),
              _SetDueDateIntent: CallbackAction<_SetDueDateIntent>(
                onInvoke: (_) => _setDueDateFromKeys(),
              ),
            },
            child: Scaffold(
              key: _scaffold,
              appBar: _appBar(wide),
              drawer: wide ? null : Drawer(child: SafeArea(child: sidebar)),
              body: wide
                  ? Row(
                      children: [
                        SizedBox(width: 220, child: sidebar),
                        const VerticalDivider(width: 1),
                        Expanded(child: _splitView(list, _detailsView())),
                      ],
                    )
                  : list,
            ),
          ),
        );
      },
    );
  }

  Widget _detailsView() => _detailsCard == null
      ? Center(
          child: Text(
            _notes
                ? context.l10n.selectNoteForDetails
                : context.l10n.selectCardForDetails,
          ),
        )
      : CardDetailsPane(key: _pane, cardId: _detailsCard!);

  /// The list next to the details, with a divider between them that can be
  /// dragged; where it is dropped is remembered.
  Widget _splitView(Widget list, Widget details) {
    const handle = 9.0;
    return LayoutBuilder(
      builder: (context, constraints) {
        final room = max(constraints.maxWidth - handle, 1.0);
        return Row(
          children: [
            SizedBox(width: room * _split, child: list),
            MouseRegion(
              cursor: SystemMouseCursors.resizeColumn,
              child: GestureDetector(
                key: const Key('browse-split'),
                behavior: HitTestBehavior.opaque,
                onHorizontalDragUpdate: (d) => setState(() {
                  _split = (_split + d.delta.dx / room).clamp(
                    _minSplit,
                    _maxSplit,
                  );
                }),
                onHorizontalDragEnd: (_) => _settings!.setBrowserSplit(_split),
                child: const VerticalDivider(width: handle),
              ),
            ),
            Expanded(child: details),
          ],
        );
      },
    );
  }

  /// The keys that work while the list has focus: Up and Down move the
  /// current row, with Shift extending the selection, Home and End go to
  /// the first and last row, Enter edits the current card and Delete
  /// deletes the selected notes.
  Widget _listKeys(Widget list) {
    return Shortcuts(
      shortcuts: {
        const SingleActivator(LogicalKeyboardKey.arrowUp): const _MoveIntent(
          -1,
        ),
        const SingleActivator(LogicalKeyboardKey.arrowDown): const _MoveIntent(
          1,
        ),
        const SingleActivator(LogicalKeyboardKey.arrowUp, shift: true):
            const _MoveIntent(-1, extend: true),
        const SingleActivator(LogicalKeyboardKey.arrowDown, shift: true):
            const _MoveIntent(1, extend: true),
        const SingleActivator(LogicalKeyboardKey.home): const _JumpIntent(
          last: false,
        ),
        const SingleActivator(LogicalKeyboardKey.end): const _JumpIntent(
          last: true,
        ),
        const SingleActivator(LogicalKeyboardKey.enter): const _EditIntent(),
        const SingleActivator(LogicalKeyboardKey.numpadEnter):
            const _EditIntent(),
        for (final a in deleteRowsShortcuts) a: const _DeleteIntent(),
      },
      child: Actions(
        actions: {
          _MoveIntent: CallbackAction<_MoveIntent>(
            onInvoke: (i) => _move(i.delta, extend: i.extend),
          ),
          _JumpIntent: CallbackAction<_JumpIntent>(
            onInvoke: (i) => _goTo(i.last ? _ids.length - 1 : 0),
          ),
          _EditIntent: CallbackAction<_EditIntent>(
            onInvoke: (_) => _editCurrent(),
          ),
          _DeleteIntent: CallbackAction<_DeleteIntent>(
            onInvoke: (_) => _deleteFromKeys(),
          ),
        },
        child: Focus(focusNode: _listFocus, child: list),
      ),
    );
  }
}

class _FocusSearchIntent extends Intent {
  const _FocusSearchIntent();
}

class _FocusListIntent extends Intent {
  const _FocusListIntent();
}

class _InvertSelectionIntent extends Intent {
  const _InvertSelectionIntent();
}

class _SetDueDateIntent extends Intent {
  const _SetDueDateIntent();
}

class _MoveIntent extends Intent {
  const _MoveIntent(this.delta, {this.extend = false});

  final int delta;
  final bool extend;
}

class _JumpIntent extends Intent {
  const _JumpIntent({required this.last});

  final bool last;
}

class _EditIntent extends Intent {
  const _EditIntent();
}

class _DeleteIntent extends Intent {
  const _DeleteIntent();
}
