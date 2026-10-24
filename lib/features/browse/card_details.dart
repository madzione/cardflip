import 'dart:async';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/editor/field_drop_target.dart';
import 'package:cardflip/features/editor/field_tools.dart';
import 'package:cardflip/features/editor/math_insert.dart';
import 'package:cardflip/features/editor/occlusion_fields.dart';
import 'package:cardflip/features/editor/rich_field.dart';
import 'package:cardflip/features/editor/tags_field.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/widgets/time_format.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'card_info_screen.dart';
import 'delete_cards.dart';
import 'flag_menu.dart';
import 'reposition_dialog.dart';
import 'reset_dialog.dart';
import 'set_due_date_dialog.dart';

/// A card's rendered question and answer, an editor for its note, its
/// scheduling information, and actions on the card. It refreshes when the
/// collection changes.
class CardDetailsPane extends StatefulWidget {
  const CardDetailsPane({
    super.key,
    required this.cardId,
    this.showActions = true,
    this.onChanged,
    this.onDeleted,
  });

  final int cardId;

  /// Whether the pane shows the card actions itself. A screen that puts
  /// [CardDetailsPaneState.actionButtons] in its app bar turns this off.
  final bool showActions;

  /// Called when the card actions may have changed.
  final VoidCallback? onChanged;

  /// Called after the card was deleted from the pane.
  final VoidCallback? onDeleted;

  @override
  State<CardDetailsPane> createState() => CardDetailsPaneState();
}

class CardDetailsPaneState extends State<CardDetailsPane> {
  golang.CardDetails? _details;
  Object? _error;
  StreamSubscription<golang.Change>? _changes;
  final _editorKey = GlobalKey();
  final _firstField = FocusNode();

  /// Plays the card's sounds on request; the details never autoplay.
  final _player = MediaPlayer();

  /// Set while the card is being deleted, so the pane does not try to load
  /// it again.
  bool _deleting = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _changes ??= AppScope.of(context).changes.stream.listen((_) => _load());
  }

  @override
  void didUpdateWidget(CardDetailsPane oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.cardId != widget.cardId) {
      _player.leave();
      _details = null;
      _deleting = false;
      _load();
    }
  }

  @override
  void dispose() {
    _changes?.cancel();
    _firstField.dispose();
    _player.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    if (_deleting) return;
    try {
      final d = await golang.browse.card(widget.cardId);
      if (!mounted || _deleting) return;
      setState(() {
        _details = d;
        _error = null;
      });
    } catch (e) {
      if (!mounted || _deleting) return;
      setState(() => _error = e);
    }
    widget.onChanged?.call();
  }

  /// Brings the editor into view and focuses its first field.
  Future<void> edit() async {
    final editor = _editorKey.currentContext;
    if (editor == null) return;
    await Scrollable.ensureVisible(
      editor,
      duration: motionDuration(context, const Duration(milliseconds: 200)),
      curve: Curves.easeOut,
    );
    _firstField.requestFocus();
  }

  Future<void> _setSuspended(bool suspended) =>
      guard(() => golang.browse.setSuspended([widget.cardId], suspended));

  Future<void> _delete(golang.CardDetails d) async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final impact = await confirmDeleteCards(context, [
      d.cardID,
    ], impact: golang.DeleteImpact(notes: 1, cards: d.noteCards));
    if (impact == null || !mounted) return;
    setState(() => _deleting = true);
    widget.onChanged?.call();
    if (await guard(() => golang.browse.deleteNotes([d.cardID]))) {
      undo.offer(l.deletedCards(impact.cards));
      widget.onDeleted?.call();
    } else if (mounted) {
      setState(() => _deleting = false);
      _load();
    }
  }

  Future<void> _setFlag(int flag) =>
      guard(() => golang.browse.setFlag([widget.cardId], flag));

  Future<void> _setDueDate() async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final done = await askSetDueDate(
      context,
      cards: 1,
      apply: (spec) async {
        await golang.browse.setDueDate([widget.cardId], spec);
        return true;
      },
    );
    if (done == true) undo.offer(l.setDueDateOfCards(1));
  }

  Future<void> _reset() async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final opts = await askReset(context, cards: 1);
    if (opts == null) return;
    if (await guard(() => golang.browse.resetToNew([widget.cardId], opts))) {
      undo.offer(l.resetCardsToNew(1));
    }
  }

  /// Edit, Flag, Card info, Set due date, Reset, Suspend or Unsuspend, and
  /// Delete, as icon buttons.
  List<Widget> actionButtons() {
    final d = _details;
    if (d == null) return const [];
    final l = context.l10n;
    return [
      IconButton(
        tooltip: l.editAction,
        icon: const Icon(Icons.edit_outlined),
        onPressed: _deleting ? null : edit,
      ),
      FlagMenuButton(
        current: d.flag,
        enabled: !_deleting,
        onSelected: _setFlag,
      ),
      IconButton(
        tooltip: l.cardInfo,
        icon: const Icon(Icons.info_outline),
        onPressed: _deleting ? null : () => showCardInfo(context, d.cardID),
      ),
      IconButton(
        tooltip: l.setDueDate,
        icon: const Icon(Icons.edit_calendar_outlined),
        onPressed: _deleting ? null : _setDueDate,
      ),
      IconButton(
        tooltip: l.resetToNew,
        icon: const Icon(Icons.restart_alt),
        onPressed: _deleting ? null : _reset,
      ),
      d.suspended
          ? IconButton(
              tooltip: l.unsuspend,
              icon: const Icon(Icons.play_circle_outline),
              onPressed: _deleting ? null : () => _setSuspended(false),
            )
          : IconButton(
              tooltip: l.suspend,
              icon: const Icon(Icons.pause_circle_outline),
              onPressed: _deleting ? null : () => _setSuspended(true),
            ),
      IconButton(
        tooltip: l.delete,
        icon: const Icon(Icons.delete_outline),
        onPressed: _deleting ? null : () => _delete(d),
      ),
    ];
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      // On a phone the pane is its own route: stop as soon as the pop
      // starts, so the video is gone well before the player is freed.
      onPopInvokedWithResult: (didPop, _) {
        if (didPop) _player.stop();
      },
      child: _content(context),
    );
  }

  Widget _content(BuildContext context) {
    final d = _details;
    if (d == null) {
      return Center(
        child: _error != null
            ? Text(errorMessage(_error!))
            : const CircularProgressIndicator(),
      );
    }
    final theme = Theme.of(context);
    final showPlayButtons = AppScope.of(context).settings.showPlayButtons;
    String date(DateTime? t) => t == null ? '—' : formatDue(t);
    final l = context.l10n;
    final oneDecimal = NumberFormat('0.0', l.localeName);
    final number = NumberFormat.decimalPattern(l.localeName);
    final info = <(String, String)>[
      (
        l.infoState,
        d.suspended
            ? l.stateWithSuspended(l.cardKind(d.state))
            : l.cardKind(d.state),
      ),
      if (d.flag != 0) (l.flag, AppScope.of(context).settings.flagName(d.flag)),
      (
        l.infoDue,
        d.state == 'new' ? newCardText(d.position, l.newCard) : date(d.due),
      ),
      (l.infoInterval, d.interval == 0 ? '—' : l.countDays(d.interval)),
      (
        l.infoStability,
        d.stability == null
            ? '—'
            : l.daysDecimal(oneDecimal.format(d.stability)),
      ),
      (
        l.infoDifficulty,
        d.difficulty == null ? '—' : oneDecimal.format(d.difficulty),
      ),
      (l.infoReviews, number.format(d.reviews)),
      (l.infoLapses, number.format(d.lapses)),
      (l.deck, d.deck),
      (l.card, '${d.noteType.name} · ${d.cardName}'),
    ];
    final details = SingleChildScrollView(
      key: const Key('card-details'),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(l.question, style: theme.textTheme.titleSmall),
          const SizedBox(height: 4),
          Card(
            clipBehavior: Clip.antiAlias,
            child: CardView(
              html: d.question,
              background: d.background,
              color: d.color,
              fonts: d.fonts,
              sounds: MediaItem.inOrder(
                CardSide.question,
                d.question,
                d.questionMedia,
                d.questionTTS,
              ),
              player: _player,
              showPlayButtons: showPlayButtons,
              padding: const EdgeInsets.all(16),
            ),
          ),
          const SizedBox(height: 12),
          Text(l.answer, style: theme.textTheme.titleSmall),
          const SizedBox(height: 4),
          Card(
            clipBehavior: Clip.antiAlias,
            child: CardView(
              html: d.answer,
              background: d.background,
              color: d.color,
              fonts: d.fonts,
              sounds: MediaItem.inOrder(
                CardSide.answer,
                d.answer,
                d.answerMedia,
                d.answerTTS,
              ),
              player: _player,
              showPlayButtons: showPlayButtons,
              padding: const EdgeInsets.all(16),
            ),
          ),
          const SizedBox(height: 16),
          KeyedSubtree(
            key: _editorKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(l.editCard, style: theme.textTheme.titleSmall),
                if (d.noteCards > 1)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      l.changesApplyToOthers(d.noteCards - 1),
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ),
                const SizedBox(height: 8),
                NoteFieldsEditor(
                  key: ValueKey(d.note.id),
                  note: d.note,
                  noteType: d.noteType,
                  firstFieldFocusNode: _firstField,
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Text(l.scheduling, style: theme.textTheme.titleSmall),
          const SizedBox(height: 4),
          Table(
            columnWidths: const {0: IntrinsicColumnWidth()},
            children: [
              for (final (k, v) in info)
                TableRow(
                  children: [
                    Padding(
                      padding: const EdgeInsets.only(right: 16, bottom: 4),
                      child: Text(
                        k,
                        style: TextStyle(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ),
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: Text(v),
                    ),
                  ],
                ),
            ],
          ),
        ],
      ),
    );
    if (!widget.showActions) return details;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 8, 8, 0),
          child: Row(
            children: [
              Expanded(child: Text(l.card, style: theme.textTheme.titleMedium)),
              ...actionButtons(),
            ],
          ),
        ),
        Expanded(child: details),
      ],
    );
  }
}

/// Edits a note's fields and tags in place; an Image Occlusion note's
/// shapes and image with the mask editor.
class NoteFieldsEditor extends StatefulWidget {
  const NoteFieldsEditor({
    super.key,
    required this.note,
    required this.noteType,
    this.firstFieldFocusNode,
  });

  final golang.Note note;
  final golang.NoteType noteType;

  /// Focuses the first field shown, e.g. when the user chooses Edit.
  final FocusNode? firstFieldFocusNode;

  @override
  State<NoteFieldsEditor> createState() => _NoteFieldsEditorState();
}

class _NoteFieldsEditorState extends State<NoteFieldsEditor>
    with FieldTools, OcclusionFields {
  late final _fields = [
    for (var i = 0; i < widget.noteType.fields.length; i++)
      RichFieldController(
        html: i < widget.note.fields.length ? widget.note.fields[i] : '',
        htmlByDefault: widget.noteType.fieldOption(i).htmlByDefault,
      ),
  ];
  late final _ownFocus = [for (final _ in _fields) FocusNode()];
  late final _focus = [
    for (var i = 0; i < _fields.length; i++)
      i == _firstShown
          ? widget.firstFieldFocusNode ?? _ownFocus[i]
          : _ownFocus[i],
  ];

  /// The first field edited with a field input; the mask editor stands
  /// for an Image Occlusion note's first ones.
  late final int _firstShown = () {
    for (var i = 0; i < _fields.length; i++) {
      if (shownAsField(i)) return i;
    }
    return 0;
  }();

  /// The field that had focus last; menus take the focus away while open.
  late int _lastField = _firstShown;
  late List<String> _tags = [...widget.note.tags];
  final _tagsKey = GlobalKey<TagsFieldState>();
  String? _error;
  bool _saving = false;

  @override
  List<RichFieldController> get toolFields => _fields;

  @override
  List<FocusNode> get toolFocus => _focus;

  @override
  int get lastFocusedField => _lastField;

  @override
  bool get clozeNoteType => widget.noteType.cloze;

  @override
  void showToolError(String message) => setState(() => _error = message);

  @override
  void showOcclusionError(String? message) => setState(() => _error = message);

  @override
  void initState() {
    super.initState();
    initOcclusion(
      widget.noteType,
      [for (final c in _fields) c.html],
      editing: true,
      cardNumbers: [for (final o in widget.note.cardOrds) o + 1],
    );
    enableImagePaste();
    for (var i = 0; i < _focus.length; i++) {
      final f = _focus[i];
      f.addListener(() {
        if (f.hasFocus && _lastField != i && mounted) {
          setState(() => _lastField = i);
        }
      });
    }
  }

  @override
  void dispose() {
    disposeOcclusion();
    for (final c in _fields) {
      c.dispose();
    }
    for (final f in _ownFocus) {
      f.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    final tags = _tagsKey.currentState?.commit() ?? _tags;
    final fields = [for (final c in _fields) c.html];
    if (occlusionProblem() case final problem?) {
      setState(() => _error = problem);
      return;
    }
    applyOcclusion(fields);
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      await golang.notes.update(widget.note.id, fields, tags);
      showInfo(appL10n.cardSaved);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return CallbackShortcuts(
      bindings: clozeShortcuts,
      child: MathShortcuts(
        onInsert: insertMath,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            RichToolbar(
              field: _fields.isEmpty ? null : _fields[_lastField],
              trailing: toolButtons(),
            ),
            const SizedBox(height: 8),
            for (var i = 0; i < _fields.length; i++)
              if (!shownAsField(i)) ...[
                if (i == occlusionFieldIndex) occlusionSection(context),
              ] else
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: FieldDropTarget(
                    key: Key('details-drop-$i'),
                    onDrop: (files) => dropMedia(i, files),
                    child: RichField(
                      key: Key('details-field-$i'),
                      label: widget.noteType.fields[i],
                      controller: _fields[i],
                      focusNode: _focus[i],
                      options: widget.noteType.fieldOption(i),
                    ),
                  ),
                ),
            TagsField(
              key: _tagsKey,
              tags: _tags,
              onChanged: (t) => setState(() => _tags = t),
            ),
            if (_error != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  _error!,
                  key: const Key('details-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            const SizedBox(height: 12),
            Row(
              key: const Key('details-save-row'),
              children: [
                const Spacer(),
                FilledButton(
                  onPressed: _saving ? null : _save,
                  child: Text(context.l10n.save),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

/// The details pane as a screen, for narrow layouts, with the card actions
/// in its app bar. It closes after the card is deleted.
class CardDetailsScreen extends StatefulWidget {
  const CardDetailsScreen({super.key, required this.cardId});

  final int cardId;

  @override
  State<CardDetailsScreen> createState() => _CardDetailsScreenState();
}

class _CardDetailsScreenState extends State<CardDetailsScreen> {
  final _pane = GlobalKey<CardDetailsPaneState>();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(context.l10n.card),
        actions: _pane.currentState?.actionButtons(),
      ),
      body: CardDetailsPane(
        key: _pane,
        cardId: widget.cardId,
        showActions: false,
        onChanged: () {
          if (mounted) setState(() {});
        },
        onDeleted: () {
          if (mounted) Navigator.of(context).pop();
        },
      ),
    );
  }
}
