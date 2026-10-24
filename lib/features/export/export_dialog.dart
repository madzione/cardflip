import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

/// What an export writes: a deck with its subdecks, the notes of selected
/// cards, or the whole collection.
sealed class ExportTarget {
  const ExportTarget();
}

class DeckExport extends ExportTarget {
  const DeckExport(this.deckId, this.name);

  final int deckId;
  final String name;
}

class NotesExport extends ExportTarget {
  const NotesExport(this.cardIds);

  final List<int> cardIds;
}

class CollectionExport extends ExportTarget {
  const CollectionExport();
}

/// The file an export writes: an Anki package (.apkg, or .colpkg for the
/// collection), notes as tab-separated text, or cards as plain text.
enum ExportFormat { package, notesText, cardsText }

/// Shows the export options for [target], then exports and saves the file.
Future<void> showExportDialog(BuildContext context, ExportTarget target) =>
    showDialog<void>(
      context: context,
      animationStyle: popupAnimation(context),
      barrierDismissible: false,
      builder: (_) => ExportDialog(target: target),
    );

/// Export format, options and progress (data-export spec). The file is
/// saved with the picker: a save dialog on native platforms, a download on
/// the web.
class ExportDialog extends StatefulWidget {
  const ExportDialog({super.key, required this.target});

  final ExportTarget target;

  @override
  State<ExportDialog> createState() => _ExportDialogState();
}

class _ExportDialogState extends State<ExportDialog> {
  ExportFormat _format = ExportFormat.package;
  bool _scheduling = false;
  bool _presets = false;
  bool _media = true;
  bool _legacy = false;
  // Columns of a notes text export.
  bool _tags = true;
  bool _deck = false;
  bool _noteType = false;
  bool _guid = false;

  StreamSubscription<golang.ExportEvent>? _export;
  golang.ExportEvent? _progress;
  String? _error;

  bool get _collection => widget.target is CollectionExport;
  bool get _running => _export != null;

  @override
  void dispose() {
    _export?.cancel();
    super.dispose();
  }

  String get _title => switch (widget.target) {
    DeckExport(:final name) => context.l10n.exportNamed(name),
    NotesExport(:final cardIds) when _format == ExportFormat.cardsText =>
      context.l10n.exportCards(cardIds.length),
    NotesExport(:final cardIds) => context.l10n.exportNotesOf(cardIds.length),
    CollectionExport() => context.l10n.exportCollection,
  };

  golang.TextScope get _textScope => switch (widget.target) {
    DeckExport(:final deckId) => golang.TextScope(
      deckID: deckId,
      cardIDs: const [],
      all: false,
    ),
    NotesExport(:final cardIds) => golang.TextScope(
      deckID: 0,
      cardIDs: cardIds,
      all: false,
    ),
    CollectionExport() => const golang.TextScope(
      deckID: 0,
      cardIDs: [],
      all: true,
    ),
  };

  Stream<golang.ExportEvent> _package() {
    final options = golang.ExportOptions(
      scheduling: _scheduling,
      presets: _presets,
      media: _media,
      legacy: _legacy,
    );
    return switch (widget.target) {
      DeckExport(:final deckId) => golang.transfer.exportDeck(deckId, options),
      NotesExport(:final cardIds) => golang.transfer.exportNotes(
        cardIds,
        options,
      ),
      CollectionExport() => golang.transfer.exportCollection(options),
    };
  }

  void _start() {
    final events = switch (_format) {
      ExportFormat.package => _package(),
      ExportFormat.notesText => golang.transfer.exportNotesText(
        _textScope,
        golang.TextOptions(
          tags: _tags,
          deck: _deck,
          noteType: _noteType,
          guid: _guid,
        ),
      ),
      ExportFormat.cardsText => golang.transfer.exportCardsText(_textScope),
    };
    setState(() {
      _error = null;
      _progress = null;
      _export = events.listen(
        _onEvent,
        onError: (Object e) => _fail(errorMessage(e)),
        onDone: () {
          if (mounted && _running) _fail(context.l10n.exportStopped);
        },
      );
    });
  }

  Future<void> _onEvent(golang.ExportEvent e) async {
    if (e.error.isNotEmpty) {
      _fail(e.error[0].toUpperCase() + e.error.substring(1));
      return;
    }
    final file = e.file;
    if (file == null) {
      setState(() => _progress = e);
      return;
    }
    // The file is here; the stream's end does not matter.
    unawaited(_export?.cancel());
    _export = null;
    if (!mounted) return;
    final picker = AppScope.read(context).picker;
    final navigator = Navigator.of(context);
    try {
      if (await picker.save(file.name, file.data)) {
        showInfo(appL10n.exportedFile(file.name));
      }
      navigator.pop();
    } catch (err) {
      _fail(errorMessage(err));
    }
  }

  void _fail(String message) {
    unawaited(_export?.cancel());
    if (!mounted) return;
    setState(() {
      _export = null;
      _error = message;
    });
  }

  void _cancel() {
    unawaited(_export?.cancel());
    _export = null;
    Navigator.pop(context);
  }

  String _stage(golang.ExportEvent? e) => switch (e) {
    null => context.l10n.preparing,
    golang.ExportEvent(stage: 'cards', :final done, :final total) =>
      context.l10n.renderingCards(done, total),
    _ when _format != ExportFormat.package => context.l10n.writingFile,
    golang.ExportEvent(stage: 'media', :final done, :final total) =>
      context.l10n.addingMedia(done, total),
    golang.ExportEvent(stage: 'saving') => context.l10n.packingFile,
    _ => context.l10n.writingCards,
  };

  String get _description => switch (_format) {
    ExportFormat.package when _collection => context.l10n.exportColpkgHelp,
    ExportFormat.package => context.l10n.exportApkgHelp,
    ExportFormat.notesText => context.l10n.exportNotesTextHelp,
    ExportFormat.cardsText => context.l10n.exportCardsTextHelp,
  };

  Widget _option({
    required Key key,
    required String title,
    String? subtitle,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) => CheckboxListTile(
    key: key,
    contentPadding: EdgeInsets.zero,
    controlAffinity: ListTileControlAffinity.leading,
    title: Text(title),
    subtitle: subtitle == null ? null : Text(subtitle),
    value: value,
    onChanged: _running ? null : (v) => setState(() => onChanged(v ?? false)),
  );

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final progress = _progress;
    final counted =
        (progress?.stage == 'media' || progress?.stage == 'cards') &&
        progress!.total > 0;
    final fraction = counted ? progress.done / progress.total : null;
    final l = context.l10n;
    return AlertDialog(
      title: Text(_title),
      content: SizedBox(
        width: 440,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              DropdownButtonFormField<ExportFormat>(
                key: const Key('export-format'),
                initialValue: _format,
                isExpanded: true,
                decoration: InputDecoration(labelText: l.format),
                items: [
                  DropdownMenuItem(
                    value: ExportFormat.package,
                    child: Text(_collection ? l.formatColpkg : l.formatApkg),
                  ),
                  DropdownMenuItem(
                    value: ExportFormat.notesText,
                    child: Text(l.formatNotesText),
                  ),
                  DropdownMenuItem(
                    value: ExportFormat.cardsText,
                    child: Text(l.formatCardsText),
                  ),
                ],
                onChanged: _running
                    ? null
                    : (v) => setState(() => _format = v ?? _format),
              ),
              const SizedBox(height: 12),
              Text(_description, style: theme.textTheme.bodyMedium),
              const SizedBox(height: 8),
              if (_format == ExportFormat.notesText) ...[
                _option(
                  key: const Key('export-tags'),
                  title: l.includeTags,
                  value: _tags,
                  onChanged: (v) => _tags = v,
                ),
                _option(
                  key: const Key('export-deck'),
                  title: l.includeDeck,
                  value: _deck,
                  onChanged: (v) => _deck = v,
                ),
                _option(
                  key: const Key('export-note-type'),
                  title: l.includeNoteType,
                  value: _noteType,
                  onChanged: (v) => _noteType = v,
                ),
                _option(
                  key: const Key('export-guid'),
                  title: l.includeGuid,
                  subtitle: l.includeGuidHelp,
                  value: _guid,
                  onChanged: (v) => _guid = v,
                ),
              ],
              if (_format == ExportFormat.package && !_collection) ...[
                _option(
                  key: const Key('export-scheduling'),
                  title: l.includeScheduling,
                  subtitle: l.includeSchedulingHelp,
                  value: _scheduling,
                  onChanged: (v) => _scheduling = v,
                ),
                _option(
                  key: const Key('export-presets'),
                  title: l.includePresets,
                  subtitle: l.includePresetsHelp,
                  value: _presets,
                  onChanged: (v) => _presets = v,
                ),
              ],
              if (_format == ExportFormat.package) ...[
                _option(
                  key: const Key('export-media'),
                  title: l.includeMedia,
                  subtitle: l.includeMediaHelp,
                  value: _media,
                  onChanged: (v) => _media = v,
                ),
                _option(
                  key: const Key('export-legacy'),
                  title: l.legacyAnki,
                  subtitle: l.legacyAnkiHelp,
                  value: _legacy,
                  onChanged: (v) => _legacy = v,
                ),
              ],
              if (_running) ...[
                const SizedBox(height: 16),
                LinearProgressIndicator(
                  key: const Key('export-progress'),
                  value: fraction,
                ),
                const SizedBox(height: 8),
                Text(_stage(progress), style: theme.textTheme.bodySmall),
              ],
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: _cancel, child: Text(l.cancel)),
        FilledButton(
          onPressed: _running ? null : _start,
          child: Text(l.export),
        ),
      ],
    );
  }
}
