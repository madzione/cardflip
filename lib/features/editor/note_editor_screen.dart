import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/browse/browse_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/builtin_names.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'field_drop_target.dart';
import 'field_tools.dart';
import 'math_insert.dart';
import 'note_type_manager.dart';
import 'occlusion_fields.dart';
import 'rich_field.dart';
import 'tags_field.dart';

/// What the add screen starts with instead of the remembered note type and
/// deck and empty fields, e.g. a copy of a note.
class NoteDraft {
  const NoteDraft({
    required this.typeId,
    required this.deckId,
    required this.fields,
    this.tags = const [],
  });

  final int typeId;
  final int deckId;
  final List<String> fields;
  final List<String> tags;
}

/// Adds notes (rapid entry: after each add the fields that are not pinned
/// clear and the deck, note type and tags stay), or edits one note when
/// [noteId] is set. Ctrl/Cmd+Enter adds or saves. A first field that
/// duplicates another note of the note type is marked, without blocking
/// the add. Image Occlusion notes get the mask editor in place of their
/// Occlusion and Image fields; the note type picker offers Image Occlusion
/// also in collections that do not have the note type yet.
class NoteEditorScreen extends StatefulWidget {
  const NoteEditorScreen({super.key, this.deckId, this.noteId, this.initial});

  /// How long after the last change to the first field it is checked for
  /// duplicates.
  static const duplicateDebounce = Duration(milliseconds: 300);

  /// The deck to add to; the remembered deck when null.
  final int? deckId;

  /// The note to edit; null adds new notes.
  final int? noteId;

  /// The values to start adding with; the remembered ones when null.
  final NoteDraft? initial;

  @override
  State<NoteEditorScreen> createState() => _NoteEditorScreenState();
}

class _NoteEditorScreenState extends State<NoteEditorScreen>
    with FieldTools, OcclusionFields {
  List<golang.NoteType> _types = const [];
  int? _typeId;
  int? _deckId;
  List<String> _tags = [];
  List<RichFieldController> _fields = [];
  List<FocusNode> _focus = [];
  final _tagsKey = GlobalKey<TagsFieldState>();

  /// The field that had focus last; menus take the focus away while open.
  int _lastField = 0;
  String? _error;
  bool _loading = true;
  bool _saving = false;

  /// The other notes of the note type whose first field equals this one's.
  List<int> _duplicates = const [];
  Timer? _duplicateTimer;

  /// The note type and first field last checked, so moving the cursor does
  /// not check again.
  (int, String)? _duplicateChecked;

  /// Numbers the checks, so an answer to an outdated one is dropped.
  int _duplicateRequest = 0;

  bool get _editing => widget.noteId != null;

  /// Collection changes, which the add screen follows (see
  /// [_followChanges]).
  StreamSubscription<golang.Change>? _changes;

  /// The cloze numbers of the edited note's cards.
  List<int> _cardNumbers = const [];

  /// The type picker's entry that adds the Image Occlusion note type.
  static const _addOcclusionType = -1;

  bool get _isOcclusion => _type?.imageOcclusion ?? false;

  golang.NoteType? get _type {
    for (final t in _types) {
      if (t.id == _typeId) return t;
    }
    return null;
  }

  @override
  void initState() {
    super.initState();
    _load();
    if (!_editing) {
      _changes = AppScope.maybeRead(context)?.changes.collection
          .listen(_followChanges);
    }
  }

  /// Note types and decks changed elsewhere, such as by device sync, whose
  /// first merge combines the built-in note types and the Default decks of
  /// both devices: the pickers show the current ones, and a note type or
  /// deck that is gone is replaced with the defaults, keeping the text of
  /// fields with the same name.
  Future<void> _followChanges(golang.Change c) async {
    if (_loading || !(c.noteTypes || c.decks || c.all)) return;
    try {
      final types = await golang.notes.types();
      final decks = await golang.decks.tree();
      if (!mounted) return;
      final before = _type;
      final typeGone = !types.any((t) => t.id == _typeId);
      final deckGone = !decks.any((d) => d.id == _deckId && !d.filtered);
      final d = typeGone || deckGone
          ? await golang.notes.defaults(typeGone ? 0 : _typeId!)
          : null;
      if (!mounted) return;
      setState(() {
        _types = types;
        if (d != null && typeGone) _typeId = d.typeID;
        if (d != null && deckGone) _deckId = d.deckID;
        final after = _type;
        if (after?.id != before?.id ||
            !listEquals(after?.fields, before?.fields)) {
          _resetFields();
        }
      });
    } catch (_) {
      // The next change or adding the note shows what is wrong.
    }
  }

  Future<void> _load() async {
    try {
      final types = await golang.notes.types();
      int? typeId;
      int? deckId;
      var values = <String>[];
      if (_editing) {
        final note = await golang.notes.get_(widget.noteId!);
        typeId = note.typeID;
        values = note.fields;
        _tags = [...note.tags];
        _cardNumbers = [for (final o in note.cardOrds) o + 1];
      } else if (widget.initial case final initial?) {
        typeId = initial.typeId;
        deckId = initial.deckId;
        values = initial.fields;
        _tags = [...initial.tags];
      } else {
        final d = await golang.notes.defaults(0);
        typeId = d.typeID;
        deckId = widget.deckId ?? d.deckID;
      }
      if (!mounted) return;
      setState(() {
        _types = types;
        _typeId = typeId;
        _deckId = deckId;
        _resetFields(values);
        _loading = false;
      });
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  /// Makes one input per field of the current note type, keeping text of
  /// fields with the same name.
  void _resetFields([List<String>? values]) {
    final old = {
      for (var i = 0; i < _fields.length && i < (_lastNames.length); i++)
        _lastNames[i]: _fields[i].html,
    };
    for (final c in _fields) {
      c.dispose();
    }
    for (final f in _focus) {
      f.dispose();
    }
    final type = _type;
    final names = type?.fields ?? const <String>[];
    final raw = [
      for (var i = 0; i < names.length; i++)
        values != null && i < values.length ? values[i] : old[names[i]] ?? '',
    ];
    _fields = [
      for (var i = 0; i < names.length; i++)
        RichFieldController(
          html: raw[i],
          htmlByDefault: type!.fieldOption(i).htmlByDefault,
        ),
    ];
    initOcclusion(type, raw, editing: _editing, cardNumbers: _cardNumbers);
    enableImagePaste();
    _focus = [for (final _ in names) FocusNode()];
    for (var i = 0; i < _focus.length; i++) {
      final f = _focus[i];
      f.addListener(() {
        if (f.hasFocus && _lastField != i && mounted) {
          setState(() => _lastField = i);
        }
      });
    }
    _lastField = 0;
    for (var i = 0; i < names.length; i++) {
      if (shownAsField(i)) {
        _lastField = i;
        break;
      }
    }
    _lastNames = names;
    if (_fields.isNotEmpty) {
      _fields.first.addListener(_scheduleDuplicateCheck);
      _fields.first.source.addListener(_scheduleDuplicateCheck);
    }
    _duplicateChecked = null;
    _scheduleDuplicateCheck();
  }

  void _scheduleDuplicateCheck() {
    _duplicateTimer?.cancel();
    _duplicateTimer = Timer(
      NoteEditorScreen.duplicateDebounce,
      _checkDuplicates,
    );
  }

  Future<void> _checkDuplicates() async {
    final type = _type;
    // The first field of an occlusion note holds its shapes.
    if (type == null || _fields.isEmpty || !mounted || _isOcclusion) return;
    final checked = (type.id, _fields.first.html);
    if (checked == _duplicateChecked) return;
    _duplicateChecked = checked;
    final request = ++_duplicateRequest;
    List<int> ids;
    try {
      ids = await golang.notes.duplicates(
        type.id,
        checked.$2,
        widget.noteId ?? 0,
      );
    } catch (_) {
      // The warning is only a hint; without an answer none is shown.
      ids = const [];
    }
    if (mounted && request == _duplicateRequest) {
      setState(() => _duplicates = ids);
    }
  }

  Future<void> _showDuplicates() => Navigator.of(context).push(
    MaterialPageRoute<void>(builder: (_) => BrowseScreen(noteIds: _duplicates)),
  );

  /// Opens Browse with the cards added today.
  Future<void> _showHistory() => Navigator.of(context).push(
    MaterialPageRoute<void>(
      builder: (_) => const BrowseScreen(query: 'added:1'),
    ),
  );

  List<String> _lastNames = const [];

  /// Whether field [index] is pinned: it keeps its content after an add.
  bool _pinned(int index) {
    return _type?.fieldOption(index).sticky ?? false;
  }

  Future<void> _togglePin(int index) async {
    final type = _type;
    if (type == null) return;
    try {
      await golang.notes.setFieldSticky(type.id, index, !_pinned(index));
      final types = await golang.notes.types();
      if (mounted) setState(() => _types = types);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  /// Creates the Image Occlusion note type, for collections from before it
  /// was built in, and switches to it.
  Future<void> _chooseOcclusionType() async {
    try {
      final id = await golang.notes.imageOcclusionType();
      final types = await golang.notes.types();
      if (!mounted) return;
      setState(() => _types = types);
      await _changeType(id);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  /// Switches to note type [typeId]. With "Change deck depending on note
  /// type" the deck switches to the one last used with it.
  Future<void> _changeType(int? typeId) async {
    setState(() {
      _typeId = typeId;
      _resetFields();
    });
    final followType = AppScope.read(context).settings.deckByNoteType;
    if (!followType || typeId == null) return;
    try {
      final d = await golang.notes.defaults(typeId);
      if (mounted && _typeId == typeId) setState(() => _deckId = d.deckID);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  @override
  void dispose() {
    _changes?.cancel();
    _duplicateTimer?.cancel();
    disposeOcclusion();
    for (final c in _fields) {
      c.dispose();
    }
    for (final f in _focus) {
      f.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    if (_saving || _type == null) return;
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
      if (_editing) {
        await golang.notes.update(widget.noteId!, fields, tags);
        if (mounted) Navigator.of(context).pop(true);
        return;
      }
      await golang.notes.add(
        golang.NewNote(
          typeID: _typeId!,
          deckID: _deckId!,
          fields: fields,
          tags: tags,
        ),
      );
      showInfo(appL10n.cardAdded);
      // Rapid entry: clear the fields that are not pinned, keep deck, note
      // type and tags.
      for (final (i, c) in _fields.indexed) {
        if (!_pinned(i)) c.clear();
      }
      // The next occlusion note starts with a new image.
      if (occlusion case final occlusion?) {
        occlusion.load('');
        setState(() {
          imageHtml = '';
          occlusionImage = null;
        });
      }
      // A pinned first field now duplicates the note just added.
      _duplicateChecked = null;
      _scheduleDuplicateCheck();
      if (_focus.isNotEmpty) focusAfterFrame(_focus.first);
    } catch (e) {
      // The input stays, so nothing typed is lost.
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  List<RichFieldController> get toolFields => _fields;

  @override
  List<FocusNode> get toolFocus => _focus;

  @override
  int get lastFocusedField => _lastField;

  @override
  bool get clozeNoteType => _type?.cloze ?? false;

  @override
  void showToolError(String message) => setState(() => _error = message);

  @override
  void showOcclusionError(String? message) => setState(() => _error = message);

  @override
  Widget build(BuildContext context) {
    final decks = AppScope.of(context).deckList;
    final l = context.l10n;
    final submit = FilledButton.icon(
      key: const Key('editor-submit'),
      onPressed: _saving ? null : _save,
      icon: Icon(_editing ? Icons.save : Icons.add),
      label: Text(_editing ? l.save : l.addCard),
    );
    return Scaffold(
      appBar: AppBar(
        title: Text(_editing ? l.editCard : l.addCard),
        actions: [
          if (!_editing)
            IconButton(
              key: const Key('editor-history'),
              tooltip: l.historyAddedToday,
              icon: const Icon(Icons.history),
              onPressed: _showHistory,
            ),
          IconButton(
            tooltip: l.noteTypes,
            icon: const Icon(Icons.tune),
            onPressed: () async {
              await Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => const NoteTypeListScreen(),
                ),
              );
              final types = await golang.notes.types();
              if (mounted) {
                setState(() {
                  _types = types;
                  if (_type == null && types.isNotEmpty) {
                    _typeId = types.first.id;
                  }
                  _resetFields();
                });
              }
            },
          ),
        ],
      ),
      body: _loading
          ? Center(
              child: _error != null
                  ? Text(_error!)
                  : const CircularProgressIndicator(),
            )
          : CallbackShortcuts(
              bindings: {
                for (final s in submitShortcuts) s: _save,
                ...clozeShortcuts,
              },
              child: MathShortcuts(
                onInsert: insertMath,
                child: ListenableBuilder(
                  listenable: decks,
                  builder: (context, _) => ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      if (!_editing)
                        Wrap(
                          spacing: 12,
                          runSpacing: 12,
                          children: [
                            SizedBox(
                              // Keyed by the note type and the list, so a
                              // note type added from the list is shown.
                              key: ValueKey((_typeId, _types.length)),
                              width: 260,
                              child: DropdownButtonFormField<int>(
                                key: const Key('type-picker'),
                                isExpanded: true,
                                initialValue: _typeId,
                                decoration: InputDecoration(
                                  labelText: l.noteType,
                                ),
                                items: [
                                  for (final t in _types)
                                    DropdownMenuItem(
                                      value: t.id,
                                      child: Text(t.name),
                                    ),
                                  if (!_types.any(
                                    (t) => t.imageOcclusion ?? false,
                                  ))
                                    DropdownMenuItem(
                                      value: _addOcclusionType,
                                      child: Text(
                                        stockTypeName(l, 'Image Occlusion'),
                                      ),
                                    ),
                                ],
                                onChanged: (v) => v == _addOcclusionType
                                    ? _chooseOcclusionType()
                                    : _changeType(v),
                              ),
                            ),
                            SizedBox(
                              // Keyed by the deck, so a deck chosen for a note
                              // type replaces the shown one.
                              key: ValueKey(_deckId),
                              width: 260,
                              child: DropdownButtonFormField<int>(
                                key: const Key('deck-picker'),
                                isExpanded: true,
                                // Filtered decks cannot get new notes.
                                initialValue:
                                    decks.decks.any(
                                      (d) => d.id == _deckId && !d.filtered,
                                    )
                                    ? _deckId
                                    : null,
                                decoration: InputDecoration(labelText: l.deck),
                                items: [
                                  for (final d in decks.decks)
                                    if (!d.filtered)
                                      DropdownMenuItem(
                                        value: d.id,
                                        child: Text(
                                          d.fullName,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ),
                                ],
                                onChanged: (v) => setState(() => _deckId = v),
                              ),
                            ),
                          ],
                        ),
                      const SizedBox(height: 8),
                      RichToolbar(
                        field: _fields.isEmpty
                            ? null
                            : _fields[_lastField.clamp(0, _fields.length - 1)],
                        trailing: toolButtons(),
                      ),
                      const SizedBox(height: 8),
                      for (var i = 0; i < _fields.length; i++)
                        if (!shownAsField(i)) ...[
                          if (i == occlusionFieldIndex)
                            occlusionSection(context),
                        ] else ...[
                          Padding(
                            padding: const EdgeInsets.only(bottom: 12),
                            child: FieldDropTarget(
                              key: Key('drop-$i'),
                              onDrop: (files) => dropMedia(i, files),
                              child: RichField(
                                key: Key('field-$i'),
                                label: _type!.fields[i],
                                controller: _fields[i],
                                focusNode: _focus[i],
                                autofocus: i == 0 && !_editing && !_isOcclusion,
                                action: _editing ? null : _pinButton(i),
                                marked: i == 0 && _duplicates.isNotEmpty,
                                options: _type!.fieldOption(i),
                              ),
                            ),
                          ),
                          if (i == 0 && _duplicates.isNotEmpty)
                            _duplicateWarning(context),
                        ],
                      TagsField(
                        key: _tagsKey,
                        tags: _tags,
                        onChanged: (t) => setState(() => _tags = t),
                      ),
                      if (_error != null) ...[
                        const SizedBox(height: 12),
                        Card(
                          color: Theme.of(context).colorScheme.errorContainer,
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Text(
                              _error!,
                              key: const Key('editor-error'),
                              style: TextStyle(
                                color: Theme.of(context)
                                    .colorScheme
                                    .onErrorContainer,
                              ),
                            ),
                          ),
                        ),
                      ],
                      const SizedBox(height: 16),
                      Align(
                        alignment: Alignment.centerRight,
                        child: showShortcutHints
                            ? Tooltip(
                                message: submitShortcutLabel,
                                child: submit,
                              )
                            : submit,
                      ),
                    ],
                  ),
                ),
              ),
            ),
    );
  }

  Widget _duplicateWarning(BuildContext context) {
    final theme = Theme.of(context);
    final n = _duplicates.length;
    return Padding(
      key: const Key('duplicate-warning'),
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Icon(Icons.warning_amber, color: theme.colorScheme.error, size: 20),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              n == 1
                  ? context.l10n.duplicateOne(_type!.fields.first)
                  : context.l10n.duplicateMany(n, _type!.fields.first),
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.error,
              ),
            ),
          ),
          TextButton(
            key: const Key('show-duplicates'),
            onPressed: _showDuplicates,
            child: Text(context.l10n.showDuplicates),
          ),
        ],
      ),
    );
  }

  Widget _pinButton(int index) {
    final pinned = _pinned(index);
    return IconButton(
      key: Key('field-$index-pin'),
      tooltip: pinned ? appL10n.unpinField : appL10n.pinField,
      isSelected: pinned,
      visualDensity: VisualDensity.compact,
      icon: const Icon(Icons.push_pin_outlined),
      selectedIcon: const Icon(Icons.push_pin),
      onPressed: () => _togglePin(index),
    );
  }
}
