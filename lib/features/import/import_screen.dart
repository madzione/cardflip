import 'dart:async';
import 'dart:ui' show PointerDeviceKind;

import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/widgets/font_cache.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/media_files.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// File types the importer accepts.
const importExtensions = ['csv', 'tsv', 'txt', 'apkg', 'colpkg'];

/// Update policies for notes and note types that already exist, with their
/// labels (see Go's ImportOptions).
Map<String, String> _updatePolicies(AppLocalizations l) => {
  'ifNewer': l.updateIfNewer,
  'always': l.updateAlways,
  'never': l.updateNever,
};

/// Separators of delimited text files, with their labels (see Go's
/// importer/csv).
Map<String, String> _separators(AppLocalizations l) => {
  'comma': l.separatorComma,
  'semicolon': l.separatorSemicolon,
  'tab': l.separatorTab,
  'space': l.separatorSpace,
  'pipe': l.separatorPipe,
  'colon': l.separatorColon,
};

/// Column targets besides field indexes (see Go's importer/csv).
const _ignore = -1;
const _tags = -2;
const _noteType = -3;
const _deck = -4;
const _guid = -5;

/// Where a row's first field is looked for, with labels (see Go's
/// ImportOptions.DuplicateScope).
Map<String, String> _scopes(AppLocalizations l) => {
  'noteType': l.scopeNoteType,
  'deck': l.scopeDeck,
};

/// The deck dropdown's value for the deck the file names, which does not
/// exist yet.
const _newDeck = -1;

enum _Step { choose, options, running, done, failed }

/// Imports a CSV/TSV file or an Anki package: choose the file, set the
/// options, follow the progress, read the summary. An Anki collection
/// package (.colpkg) replaces the whole collection, after a confirmation.
class ImportScreen extends StatefulWidget {
  const ImportScreen({super.key});

  @override
  State<ImportScreen> createState() => _ImportScreenState();
}

class _ImportScreenState extends State<ImportScreen> {
  _Step _step = _Step.choose;
  golang.Inspection? _inspection;
  String? _error;

  // CSV options. The preview changes with the separator.
  golang.CsvPreview? _csv;
  String _separator = 'comma';
  bool _allowHTML = true;
  List<golang.NoteType> _types = const [];
  int? _typeId;
  int? _deckId;
  List<int> _mapping = const [];
  bool _hasHeader = false;
  String _duplicates = 'skip';
  String _scope = 'noteType';
  // The note types of the rows, while a column holds them.
  golang.CsvNoteTypes? _rowTypes;

  // Anki package options. Scheduling is null until the user chose.
  bool? _keepScheduling;
  String _updateNotes = 'ifNewer';
  String _updateNoteTypes = 'ifNewer';
  bool _mergeNoteTypes = false;

  // Progress and result.
  golang.ImportEvent? _progress;
  golang.ImportSummary? _summary;
  StreamSubscription<golang.ImportEvent>? _import;

  // Scrolls the CSV preview sideways when it is wider than the window.
  final _previewScroll = ScrollController();

  @override
  void dispose() {
    _import?.cancel();
    _previewScroll.dispose();
    super.dispose();
  }

  /// Whether the chosen file is an Anki collection package.
  bool get _replaces => _inspection?.stage.kind == 'colpkg';

  Future<void> _choose() async {
    final picker = AppScope.of(context).picker;
    try {
      final staged = await picker.pick(importExtensions);
      if (staged == null) return;
      final inspection = await golang.transfer.inspect(staged.id);
      final csv = inspection.csv;
      if (csv != null) {
        _types = await golang.notes.types();
        _typeId = csv.typeID;
        _deckId = csv.deckName.isNotEmpty ? _newDeck : csv.deckID;
        _mapping = [...csv.mapping];
        _hasHeader = csv.headerRow;
        _separator = csv.delimiter;
        _allowHTML = csv.allowHTML;
        _duplicates = 'skip';
        _scope = 'noteType';
        _rowTypes = _mapping.contains(_noteType) ? csv.noteTypes : null;
      }
      _csv = csv;
      _keepScheduling = null;
      _updateNotes = 'ifNewer';
      _updateNoteTypes = 'ifNewer';
      _mergeNoteTypes = false;
      setState(() {
        _inspection = inspection;
        _step = _Step.options;
        _error = null;
      });
    } catch (e) {
      setState(() {
        _error = errorMessage(e);
        _step = _Step.failed;
      });
    }
  }

  golang.NoteType? get _type {
    for (final t in _types) {
      if (t.id == _typeId) return t;
    }
    return null;
  }

  /// The names of the fields columns can map to: the chosen note type's,
  /// or, while a column holds each row's note type, the rows' ones.
  List<String> get _fields =>
      _rowTypes?.fields ?? _type?.fields ?? const <String>[];

  /// Maps the columns that are not tags, note type, deck or GUID to the
  /// first [fields] fields in order.
  void _mapFields(int fields) {
    var next = 0;
    _mapping = [
      for (final m in _mapping)
        if (m < _ignore) m else if (next < fields) next++ else _ignore,
    ];
  }

  void _setType(int? id) {
    setState(() {
      _typeId = id;
      if (_rowTypes == null) _mapFields(_fields.length);
    });
    _loadRowTypes();
  }

  /// Maps column [col] to [target]. Only one column holds the note type,
  /// the deck or the GUID.
  void _setMapping(int col, int target) {
    final hadTypes = _mapping.contains(_noteType);
    setState(() {
      if (target == _noteType || target == _deck || target == _guid) {
        for (var i = 0; i < _mapping.length; i++) {
          if (_mapping[i] == target) _mapping[i] = _ignore;
        }
      }
      _mapping[col] = target;
    });
    if (hadTypes || _mapping.contains(_noteType)) _loadRowTypes();
  }

  /// Reads the note types of the rows again after the note type column,
  /// the header row, the separator or the default note type changed.
  /// Columns mapped to fields the rows do not have are then ignored.
  Future<void> _loadRowTypes() async {
    if (!_mapping.contains(_noteType)) {
      if (_rowTypes == null) return;
      setState(() {
        _rowTypes = null;
        _clampMapping();
      });
      return;
    }
    await guard(() async {
      final types = await golang.transfer.csvNoteTypes(
        _inspection!.stage.id,
        _options(),
      );
      if (!mounted || !_mapping.contains(_noteType)) return;
      setState(() {
        _rowTypes = types;
        _clampMapping();
      });
    });
  }

  void _clampMapping() {
    final fields = _fields.length;
    _mapping = [for (final m in _mapping) m >= fields ? _ignore : m];
  }

  /// Splits the file at [separator] and shows the new columns, mapped to
  /// the fields of the chosen note type.
  Future<void> _setSeparator(String? separator) async {
    if (separator == null || separator == _separator) return;
    final id = _inspection!.stage.id;
    await guard(() async {
      final csv = await golang.transfer.previewCsv(id, separator);
      if (!mounted) return;
      setState(() {
        _csv = csv;
        _separator = separator;
        _hasHeader = csv.headerRow;
        // The roles the file's header lines give columns stay; the other
        // columns map to the fields of the chosen note type.
        _mapping = [for (final m in csv.mapping) m < _ignore ? m : _ignore];
        _rowTypes = null;
        _mapFields(_fields.length);
      });
      await _loadRowTypes();
    });
  }

  golang.ImportOptions _options() {
    final newDeck = _deckId == _newDeck;
    return golang.ImportOptions(
      typeID: _typeId ?? 0,
      deckID: newDeck ? 0 : _deckId ?? 0,
      deckName: newDeck ? _csv?.deckName ?? '' : '',
      mapping: _mapping,
      hasHeader: _hasHeader,
      duplicates: _duplicates,
      duplicateScope: _scope,
      separator: _separator,
      allowHTML: _allowHTML,
      tags: _csv?.tags ?? const [],
      keepScheduling: _keepScheduling ?? false,
      updateNotes: _updateNotes,
      updateNoteTypes: _updateNoteTypes,
      mergeNoteTypes: _mergeNoteTypes,
    );
  }

  void _start() {
    final stage = _inspection!.stage;
    final options = _options();
    setState(() {
      _step = _Step.running;
      _progress = null;
    });
    _import = golang.transfer
        .import_(stage.id, options)
        .listen(
          (e) {
            if (!mounted) return;
            setState(() {
              if (e.error.isNotEmpty) {
                _error = e.error;
                _step = _Step.failed;
              } else if (e.summary case final s?) {
                _summary = s;
                _step = _Step.done;
                if (_replaces) {
                  // A new collection: its media replace the old ones, and
                  // the undo history is gone.
                  MediaCache.instance.clear();
                  FontCache.instance.clear();
                  StoredMediaFiles.forget();
                } else if (s.notesAdded +
                        s.notesUpdated +
                        s.cardsAdded +
                        s.noteTypesAdded +
                        s.noteTypesUpdated +
                        s.noteTypesMerged >
                    0) {
                  // An import that changed nothing is no undo step.
                  AppScope.read(context).undo.offer(
                    context.l10n.importedNotes(s.notesAdded + s.notesUpdated),
                  );
                }
              } else {
                _progress = e;
              }
            });
          },
          onError: (Object e) {
            if (mounted) {
              setState(() {
                _error = errorMessage(e);
                _step = _Step.failed;
              });
            }
          },
        );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.import)),
      body: switch (_step) {
        _Step.choose => _chooseView(),
        _Step.options => _optionsView(),
        _Step.running => _progressView(),
        _Step.done => _summaryView(),
        _Step.failed => _failedView(),
      },
    );
  }

  Widget _centered(List<Widget> children) => Center(
    child: SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: children,
        ),
      ),
    ),
  );

  Widget _chooseView() => _centered([
    const Icon(Icons.file_open_outlined, size: 56),
    const SizedBox(height: 16),
    Text(context.l10n.importIntro, textAlign: TextAlign.center),
    const SizedBox(height: 24),
    FilledButton.icon(
      onPressed: _choose,
      icon: const Icon(Icons.folder_open),
      label: Text(context.l10n.chooseAFile),
    ),
  ]);

  Widget _failedView() => _centered([
    Icon(
      Icons.error_outline,
      size: 56,
      color: Theme.of(context).colorScheme.error,
    ),
    const SizedBox(height: 16),
    Text(
      context.l10n.importFailed,
      style: Theme.of(context).textTheme.titleLarge,
      textAlign: TextAlign.center,
    ),
    const SizedBox(height: 8),
    Text(
      _error ?? '',
      key: const Key('import-error'),
      textAlign: TextAlign.center,
    ),
    const SizedBox(height: 8),
    Text(context.l10n.collectionNotChanged, textAlign: TextAlign.center),
    const SizedBox(height: 24),
    FilledButton(
      onPressed: _choose,
      child: Text(context.l10n.chooseAnotherFile),
    ),
  ]);

  Widget _optionsView() {
    final inspection = _inspection!;
    if (_csv != null) return _csvOptions(_csv!);
    if (inspection.apkg != null && _replaces) {
      return _colpkgOptions(inspection.apkg!);
    }
    if (inspection.apkg != null) return _apkgOptions(inspection.apkg!);
    return _centered([
      const Icon(Icons.restore, size: 56),
      const SizedBox(height: 16),
      Text(
        context.l10n.fileIsBackup(inspection.stage.name),
        textAlign: TextAlign.center,
      ),
      const SizedBox(height: 24),
      OutlinedButton(
        onPressed: _choose,
        child: Text(context.l10n.chooseAnotherFile),
      ),
    ]);
  }

  Widget _csvOptions(golang.CsvPreview csv) {
    final decks = AppScope.of(context).deckList.decks;
    final fields = _fields;
    final typeColumn = _mapping.contains(_noteType);
    final deckColumn = _mapping.contains(_deck);
    final unknown = _rowTypes?.unknown ?? const <String>[];
    final error = Theme.of(context).colorScheme.error;
    final rows = csv.rows.skip(_hasHeader ? 1 : 0).take(5).toList();
    final l = context.l10n;
    String header(int col) => csv.columnNames.length > col
        ? csv.columnNames[col]
        : _hasHeader && csv.rows.isNotEmpty && csv.rows.first.length > col
        ? csv.rows.first[col]
        : l.columnNumber(col + 1);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          l.csvSummary(_inspection!.stage.name, csv.totalRows, csv.columns),
          style: Theme.of(context).textTheme.titleSmall,
        ),
        if (csv.tags.isNotEmpty) ...[
          const SizedBox(height: 4),
          Text(
            l.tagsForEveryNote(csv.tags.join(' ')),
            key: const Key('import-header-tags'),
          ),
        ],
        for (final w in csv.warnings) ...[
          const SizedBox(height: 4),
          Text(goText(w), style: TextStyle(color: error)),
        ],
        if (unknown.isNotEmpty) ...[
          const SizedBox(height: 4),
          Text(
            l.rowsSkippedUnknownType(
              _rowTypes!.unknownRows,
              unknown.map((n) => '"$n"').join(', '),
            ),
            key: const Key('import-unknown-types'),
            style: TextStyle(color: error),
          ),
        ],
        const SizedBox(height: 12),
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: [
            SizedBox(
              width: 260,
              child: DropdownButtonFormField<int>(
                key: const Key('import-type'),
                isExpanded: true,
                initialValue: _typeId,
                decoration: InputDecoration(
                  labelText: typeColumn ? l.noteTypeOfRowsWithout : l.noteType,
                ),
                items: [
                  for (final t in _types)
                    DropdownMenuItem(
                      value: t.id,
                      child: Text(t.name, overflow: TextOverflow.ellipsis),
                    ),
                ],
                onChanged: _setType,
              ),
            ),
            SizedBox(
              width: 260,
              child: DropdownButtonFormField<int>(
                key: const Key('import-deck'),
                isExpanded: true,
                initialValue:
                    decks.any((d) => d.id == _deckId && !d.filtered) ||
                        _deckId == _newDeck && csv.deckName.isNotEmpty
                    ? _deckId
                    : null,
                decoration: InputDecoration(
                  labelText: deckColumn ? l.deckOfRowsWithout : l.deck,
                ),
                items: [
                  if (csv.deckName.isNotEmpty)
                    DropdownMenuItem(
                      value: _newDeck,
                      child: Text(
                        l.newDeckName(csv.deckName),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  // Filtered decks cannot get new notes.
                  for (final d in decks)
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
            SizedBox(
              width: 260,
              child: DropdownButtonFormField<String>(
                key: const Key('import-separator'),
                isExpanded: true,
                initialValue: _separator,
                decoration: InputDecoration(labelText: l.separator),
                items: [
                  for (final e in {
                    ..._separators(l),
                    // A character without a name, as the file declares it.
                    if (!_separators(l).containsKey(_separator))
                      _separator: '"$_separator"',
                  }.entries)
                    DropdownMenuItem(value: e.key, child: Text(e.value)),
                ],
                onChanged: _setSeparator,
              ),
            ),
          ],
        ),
        SwitchListTile(
          key: const Key('import-allow-html'),
          contentPadding: EdgeInsets.zero,
          value: _allowHTML,
          onChanged: (v) => setState(() => _allowHTML = v),
          title: Text(l.allowHtml),
          subtitle: Text(l.allowHtmlHelp),
        ),
        if (csv.columnNames.isEmpty)
          CheckboxListTile(
            contentPadding: EdgeInsets.zero,
            value: _hasHeader,
            onChanged: (v) {
              setState(() => _hasHeader = v ?? false);
              _loadRowTypes();
            },
            title: Text(l.firstRowNames),
          ),
        const SizedBox(height: 8),
        // A wide preview scrolls sideways: with the scrollbar, always shown
        // so it can be found, or by dragging, also with a mouse.
        Scrollbar(
          key: const Key('csv-preview-scrollbar'),
          controller: _previewScroll,
          thumbVisibility: true,
          child: ScrollConfiguration(
            behavior: ScrollConfiguration.of(context).copyWith(
              scrollbars: false,
              dragDevices: PointerDeviceKind.values.toSet(),
            ),
            child: SingleChildScrollView(
              controller: _previewScroll,
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.only(bottom: 12),
              child: DataTable(
                key: const Key('csv-preview'),
                columns: [
                  for (var col = 0; col < csv.columns; col++)
                    DataColumn(
                      label: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            header(col),
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                          DropdownButton<int>(
                            key: Key('map-$col'),
                            isDense: true,
                            value: col < _mapping.length
                                ? _mapping[col]
                                : _ignore,
                            items: [
                              for (var f = 0; f < fields.length; f++)
                                DropdownMenuItem(
                                  value: f,
                                  child: Text(fields[f]),
                                ),
                              DropdownMenuItem(
                                value: _tags,
                                child: Text(l.tags),
                              ),
                              DropdownMenuItem(
                                value: _noteType,
                                child: Text(l.noteType),
                              ),
                              DropdownMenuItem(
                                value: _deck,
                                child: Text(l.deck),
                              ),
                              DropdownMenuItem(
                                value: _guid,
                                child: Text(l.guid),
                              ),
                              DropdownMenuItem(
                                value: _ignore,
                                child: Text(l.ignore),
                              ),
                            ],
                            onChanged: (v) => _setMapping(col, v ?? _ignore),
                          ),
                        ],
                      ),
                    ),
                ],
                dataRowMinHeight: 40,
                headingRowHeight: 72,
                rows: [
                  for (final r in rows)
                    DataRow(
                      cells: [
                        for (var col = 0; col < csv.columns; col++)
                          DataCell(
                            ConstrainedBox(
                              constraints: const BoxConstraints(maxWidth: 240),
                              child: Text(
                                col < r.length ? r[col] : '',
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ),
                      ],
                    ),
                ],
              ),
            ),
          ),
        ),
        const SizedBox(height: 16),
        Text(
          l.whenFirstFieldMatches,
          style: Theme.of(context).textTheme.titleSmall,
        ),
        if (_mapping.contains(_guid)) Text(l.guidUpdates),
        RadioGroup<String>(
          groupValue: _duplicates,
          onChanged: (v) => setState(() => _duplicates = v ?? 'skip'),
          child: Column(
            children: [
              RadioListTile(value: 'skip', title: Text(l.skipRow)),
              RadioListTile(value: 'update', title: Text(l.updateExistingNote)),
              RadioListTile(value: 'duplicate', title: Text(l.addAsDuplicate)),
            ],
          ),
        ),
        SizedBox(
          width: 260,
          child: DropdownButtonFormField<String>(
            key: const Key('import-scope'),
            isExpanded: true,
            initialValue: _scope,
            decoration: InputDecoration(labelText: l.lookForMatchingNotes),
            items: [
              for (final e in _scopes(l).entries)
                DropdownMenuItem(value: e.key, child: Text(e.value)),
            ],
            onChanged: _duplicates == 'duplicate'
                ? null
                : (v) => setState(() => _scope = v ?? 'noteType'),
          ),
        ),
        const SizedBox(height: 16),
        Align(
          alignment: Alignment.centerRight,
          child: FilledButton(
            onPressed:
                _typeId != null &&
                    _deckId != null &&
                    _mapping.any((m) => m >= 0)
                ? _start
                : null,
            child: Text(l.import),
          ),
        ),
      ],
    );
  }

  Widget _apkgOptions(golang.ApkgInfo apkg) {
    final l = context.l10n;
    return _centered([
      Text(
        _inspection!.stage.name,
        style: Theme.of(context).textTheme.titleLarge,
      ),
      const SizedBox(height: 8),
      Text(
        l.apkgSummary(apkg.notes, apkg.cards, apkg.media) +
            (apkg.decks.isEmpty ? '' : l.inDecks(apkg.decks.join(', '))),
      ),
      if (apkg.unsupported > 0) ...[
        const SizedBox(height: 8),
        Text(
          l.unsupportedMedia(apkg.unsupported),
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      ],
      const SizedBox(height: 16),
      Text(l.scheduling, style: Theme.of(context).textTheme.titleSmall),
      RadioGroup<bool>(
        groupValue: _keepScheduling,
        onChanged: (v) => setState(() => _keepScheduling = v),
        child: Column(
          children: [
            RadioListTile(
              value: true,
              title: Text(l.keepScheduling),
              subtitle: Text(l.keepSchedulingHelp),
            ),
            RadioListTile(
              value: false,
              title: Text(l.importAsNew),
              subtitle: Text(l.importAsNewHelp),
            ),
          ],
        ),
      ),
      const SizedBox(height: 16),
      Text(l.notesAlreadyThere, style: Theme.of(context).textTheme.titleSmall),
      const SizedBox(height: 4),
      Text(l.notesAlreadyThereHelp),
      const SizedBox(height: 12),
      _policyField(
        key: const Key('update-notes'),
        label: l.updateNotes,
        value: _updateNotes,
        onChanged: (v) => _updateNotes = v,
      ),
      const SizedBox(height: 12),
      _policyField(
        key: const Key('update-note-types'),
        label: l.updateNoteTypes,
        value: _updateNoteTypes,
        onChanged: (v) => _updateNoteTypes = v,
      ),
      CheckboxListTile(
        key: const Key('merge-note-types'),
        contentPadding: EdgeInsets.zero,
        value: _mergeNoteTypes,
        onChanged: (v) => setState(() => _mergeNoteTypes = v ?? false),
        title: Text(l.mergeNoteTypes),
        subtitle: Text(l.mergeNoteTypesHelp),
      ),
      const SizedBox(height: 16),
      FilledButton(
        onPressed: _keepScheduling == null ? null : _start,
        child: Text(l.import),
      ),
    ]);
  }

  Widget _policyField({
    required Key key,
    required String label,
    required String value,
    required void Function(String) onChanged,
  }) => DropdownButtonFormField<String>(
    key: key,
    isExpanded: true,
    initialValue: value,
    decoration: InputDecoration(labelText: label),
    items: [
      for (final e in _updatePolicies(context.l10n).entries)
        DropdownMenuItem(
          value: e.key,
          child: Text(e.value, overflow: TextOverflow.ellipsis),
        ),
    ],
    onChanged: (v) => setState(() => onChanged(v ?? 'ifNewer')),
  );

  Future<void> _exportBackup() async {
    final picker = AppScope.of(context).picker;
    await guard(() async {
      final file = await golang.transfer.exportBackup();
      if (await picker.save(file.name, file.data)) {
        showInfo(appL10n.backupSavedAs(file.name));
      }
    });
  }

  Future<void> _confirmReplace() async {
    final ok = await confirm(
      context,
      title: context.l10n.replaceCollectionTitle,
      message: context.l10n.replaceByImportBody(_inspection!.stage.name),
      action: context.l10n.replace,
      destructive: true,
    );
    if (ok && mounted) _start();
  }

  Widget _colpkgOptions(golang.ApkgInfo apkg) {
    final scheme = Theme.of(context).colorScheme;
    final l = context.l10n;
    return _centered([
      Text(
        _inspection!.stage.name,
        style: Theme.of(context).textTheme.titleLarge,
      ),
      const SizedBox(height: 8),
      Text(
        l.colpkgSummary(apkg.notes, apkg.cards, apkg.media, apkg.decks.length),
      ),
      if (apkg.unsupported > 0) ...[
        const SizedBox(height: 8),
        Text(
          l.unsupportedMedia(apkg.unsupported),
          style: TextStyle(color: scheme.error),
        ),
      ],
      const SizedBox(height: 16),
      Card(
        key: const Key('replace-warning'),
        color: scheme.errorContainer,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(Icons.warning_amber, color: scheme.onErrorContainer),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  l.replaceWarning,
                  style: TextStyle(color: scheme.onErrorContainer),
                ),
              ),
            ],
          ),
        ),
      ),
      const SizedBox(height: 16),
      OutlinedButton.icon(
        onPressed: _exportBackup,
        icon: const Icon(Icons.save_alt),
        label: Text(l.exportBackupFirst),
      ),
      const SizedBox(height: 8),
      FilledButton(
        style: FilledButton.styleFrom(
          backgroundColor: scheme.error,
          foregroundColor: scheme.onError,
        ),
        onPressed: _confirmReplace,
        child: Text(l.replaceCollection),
      ),
    ]);
  }

  Widget _progressView() {
    final p = _progress;
    final l = context.l10n;
    final stage = switch (p?.stage) {
      'notes' => l.addingNotes,
      'history' => l.addingHistory,
      'media' => l.savingImages,
      'saving' => l.saving,
      _ => l.readingFile,
    };
    final value = p == null || p.total == 0 ? null : p.done / p.total;
    return _centered([
      Text(stage, key: const Key('import-stage'), textAlign: TextAlign.center),
      const SizedBox(height: 16),
      LinearProgressIndicator(key: const Key('import-progress'), value: value),
      if (p != null && p.total > 1) ...[
        const SizedBox(height: 8),
        Text(l.progressOf(p.done, p.total), textAlign: TextAlign.center),
      ],
    ]);
  }

  Widget _summaryView() {
    final s = _summary!;
    final l = context.l10n;
    final lines = [
      if (s.notesAdded > 0) l.notesAdded(s.notesAdded),
      if (s.notesUpdated > 0) l.notesUpdated(s.notesUpdated),
      if (s.notesSkipped > 0) l.notesSkipped(s.notesSkipped),
      if (s.notesSkippedOlder > 0) l.notesNotUpdatedNewer(s.notesSkippedOlder),
      l.cardsAdded(s.cardsAdded),
      if (s.reviews > 0) l.reviewsImported(s.reviews),
      if (s.media > 0) l.mediaImported(s.media),
      if (s.commentLines > 0) l.commentLinesSkipped(s.commentLines),
      if (s.noteTypesAdded > 0 && !_replaces)
        l.noteTypesAdded(s.noteTypesAdded),
      if (s.noteTypesUpdated > 0) l.noteTypesUpdated(s.noteTypesUpdated),
      if (s.noteTypesMerged > 0) l.noteTypesMerged(s.noteTypesMerged),
    ];
    return _centered([
      const Icon(Icons.check_circle_outline, size: 56),
      const SizedBox(height: 16),
      Text(
        _replaces ? l.collectionReplaced : l.importComplete,
        style: Theme.of(context).textTheme.titleLarge,
        textAlign: TextAlign.center,
      ),
      const SizedBox(height: 12),
      for (final line in lines) Text(line, textAlign: TextAlign.center),
      if (s.warnings.isNotEmpty) ...[
        const SizedBox(height: 16),
        Text(l.warnings, style: Theme.of(context).textTheme.titleSmall),
        for (final w in s.warnings)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(
                  Icons.warning_amber,
                  size: 18,
                  color: Theme.of(context).colorScheme.error,
                ),
                const SizedBox(width: 8),
                Expanded(child: Text(goText(w))),
              ],
            ),
          ),
      ],
      const SizedBox(height: 24),
      FilledButton(
        onPressed: () => Navigator.of(context).maybePop(),
        child: Text(l.done),
      ),
    ]);
  }
}
