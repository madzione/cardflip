import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/theme.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Shows a card type's front and back rendered from unsaved [template] and
/// [css], for the template and CSS editors. It renders again
/// [TemplatePreview.debounce] after the last change. The user picks the note
/// it shows (placeholder text "(Field)" when the note type has none), the
/// cloze card for cloze note types, and the light or dark variant.
class TemplatePreview extends StatefulWidget {
  const TemplatePreview({
    super.key,
    required this.noteType,
    required this.ord,
    required this.template,
    required this.css,
  });

  static const debounce = Duration(milliseconds: 300);

  final golang.NoteType noteType;

  /// The card type's index; ignored for cloze note types.
  final int ord;
  final golang.CardTemplate template;
  final String css;

  @override
  State<TemplatePreview> createState() => _TemplatePreviewState();
}

class _TemplatePreviewState extends State<TemplatePreview> {
  List<golang.NoteChoice> _notes = const [];
  int _noteId = 0;

  /// The chosen cloze number, 0 for the note's first.
  int _cloze = 0;
  bool? _night;
  golang.TemplatePreview? _preview;
  String? _error;
  Timer? _timer;

  /// Numbers the requests, so an answer to an outdated one is dropped.
  int _request = 0;

  @override
  void initState() {
    super.initState();
    unawaited(_loadNotes());
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // The preview starts in the app's theme.
    if (_night == null) {
      _night = Theme.of(context).brightness == Brightness.dark;
      unawaited(_render());
    }
  }

  @override
  void didUpdateWidget(TemplatePreview old) {
    super.didUpdateWidget(old);
    if (old.template.qFmt != widget.template.qFmt ||
        old.template.aFmt != widget.template.aFmt ||
        old.template.name != widget.template.name ||
        old.css != widget.css ||
        old.ord != widget.ord) {
      _timer?.cancel();
      _timer = Timer(TemplatePreview.debounce, _render);
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  Future<void> _loadNotes() async {
    try {
      final notes = await golang.notes.previewNotes(widget.noteType.id);
      if (!mounted || notes.isEmpty) return;
      setState(() {
        _notes = notes;
        _noteId = notes.first.id;
      });
      await _render();
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  Future<void> _render() async {
    final request = ++_request;
    try {
      final p = await golang.notes.previewTemplate(
        widget.noteType.id,
        widget.template,
        widget.css,
        _noteId,
        widget.noteType.cloze ? _cloze : widget.ord,
        _night ?? false,
      );
      if (!mounted || request != _request) return;
      setState(() {
        _preview = p;
        _error = null;
      });
    } catch (e) {
      if (!mounted || request != _request) return;
      setState(() => _error = errorMessage(e));
    }
  }

  void _change(VoidCallback fn) {
    setState(fn);
    unawaited(_render());
  }

  @override
  Widget build(BuildContext context) {
    final p = _preview;
    final night = _night ?? false;
    final theme = Theme.of(context);
    final clozeNumbers = p?.clozeNumbers ?? const <int>[];
    final shownCloze = clozeNumbers.contains(_cloze)
        ? _cloze
        : clozeNumbers.firstOrNull;
    return ListView(
      key: const Key('template-preview'),
      padding: const EdgeInsets.all(16),
      children: [
        Wrap(
          spacing: 16,
          runSpacing: 8,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            if (_notes.isNotEmpty)
              DropdownButton<int>(
                key: const Key('preview-note'),
                value: _noteId,
                onChanged: (id) => _change(() {
                  _noteId = id!;
                  _cloze = 0;
                }),
                items: [
                  for (final n in _notes)
                    DropdownMenuItem(
                      value: n.id,
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 220),
                        child: Text(
                          n.label.isEmpty ? context.l10n.emptyText : n.label,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ),
                ],
              )
            else
              Text(context.l10n.noNotesYet, style: theme.textTheme.bodySmall),
            if (widget.noteType.cloze && shownCloze != null)
              DropdownButton<int>(
                key: const Key('preview-cloze'),
                value: shownCloze,
                onChanged: (n) => _change(() => _cloze = n!),
                items: [
                  for (final n in clozeNumbers)
                    DropdownMenuItem(
                      value: n,
                      child: Text(context.l10n.clozeNumber(n)),
                    ),
                ],
              ),
            IconButton(
              key: const Key('preview-theme'),
              tooltip: night
                  ? context.l10n.previewLightTheme
                  : context.l10n.previewDarkTheme,
              icon: Icon(night ? Icons.light_mode : Icons.dark_mode),
              onPressed: () => _change(() => _night = !night),
            ),
          ],
        ),
        if (_error != null)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Text(
              _error!,
              style: TextStyle(color: theme.colorScheme.error),
            ),
          ),
        if (p != null)
          Theme(
            data: night ? darkTheme() : lightTheme(),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                for (final (label, html) in [
                  (context.l10n.front, p.question),
                  (context.l10n.back, p.answer),
                ]) ...[
                  const SizedBox(height: 12),
                  Text(label, style: theme.textTheme.titleSmall),
                  const SizedBox(height: 4),
                  Card(
                    clipBehavior: Clip.antiAlias,
                    child: CardView(
                      key: Key('preview-${label.toLowerCase()}'),
                      html: html,
                      background: p.background,
                      color: p.color,
                      fonts: p.fonts,
                      padding: const EdgeInsets.all(16),
                    ),
                  ),
                ],
              ],
            ),
          ),
      ],
    );
  }
}
