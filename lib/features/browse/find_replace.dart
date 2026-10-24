import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Shows Find and Replace for the notes of [cardIds], or for every note
/// when there are none or the user turns off "Selected notes only". It
/// applies the replacement itself and returns how many notes changed, or
/// null when cancelled.
Future<int?> showFindReplace(BuildContext context, List<int> cardIds) =>
    showDialog<int>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => _FindReplaceDialog(cardIds),
    );

/// The "in" choices besides a field name.
const _allFields = '';
const _tags = '\u0000tags';

class _FindReplaceDialog extends StatefulWidget {
  const _FindReplaceDialog(this.cardIds);

  final List<int> cardIds;

  @override
  State<_FindReplaceDialog> createState() => _FindReplaceDialogState();
}

class _FindReplaceDialogState extends State<_FindReplaceDialog> {
  final _find = TextEditingController();
  final _replace = TextEditingController();
  late bool _selectedOnly = widget.cardIds.isNotEmpty;
  bool _regex = false;
  bool _ignoreCase = true;
  String _in = _allFields;
  List<String> _fields = const [];

  /// How many notes the replacement changes; null while unknown.
  int? _count;
  String? _error;
  bool _applying = false;
  Timer? _debounce;
  int _generation = 0;

  @override
  void initState() {
    super.initState();
    _loadFields();
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _find.dispose();
    _replace.dispose();
    super.dispose();
  }

  bool get _all => !_selectedOnly;

  golang.FindReplaceOptions get _options => golang.FindReplaceOptions(
    find: _find.text,
    replace: _replace.text,
    field: _in == _tags ? '' : _in,
    tags: _in == _tags,
    regex: _regex,
    ignoreCase: _ignoreCase,
  );

  Future<void> _loadFields() async {
    final all = _all;
    try {
      final fields = await golang.browse.fieldNames(widget.cardIds, all);
      if (!mounted || all != _all) return;
      setState(() {
        _fields = fields;
        if (_in != _allFields && _in != _tags && !fields.contains(_in)) {
          _in = _allFields;
        }
      });
    } catch (e) {
      showError(e);
    }
  }

  /// Recounts the notes that change, shortly after the last edit.
  void _changed({bool now = false}) {
    _debounce?.cancel();
    setState(() {
      _count = null;
      _error = null;
    });
    if (_find.text.isEmpty) return;
    _debounce = Timer(
      now ? Duration.zero : const Duration(milliseconds: 300),
      _preview,
    );
  }

  Future<void> _preview() async {
    final generation = ++_generation;
    try {
      final n = await golang.browse.findReplacePreview(
        widget.cardIds,
        _all,
        _options,
      );
      if (!mounted || generation != _generation) return;
      setState(() => _count = n);
    } on golang.GolangException catch (e) {
      // An invalid regular expression: say why next to the input.
      if (!mounted || generation != _generation) return;
      setState(() => _error = goText(e.message));
    }
  }

  Future<void> _apply() async {
    setState(() => _applying = true);
    var changed = 0;
    final ok = await guard(() async {
      changed = await golang.browse.findReplace(widget.cardIds, _all, _options);
    });
    if (!mounted) return;
    if (ok) {
      Navigator.pop(context, changed);
    } else {
      setState(() => _applying = false);
    }
  }

  String get _countText {
    final l = context.l10n;
    if (_find.text.isEmpty) return l.findTextPrompt;
    final n = _count;
    if (n == null) return l.counting;
    if (n == 0) return l.noNotesWillChange;
    return l.notesWillChange(n);
  }

  @override
  Widget build(BuildContext context) {
    final canApply =
        !_applying && _error == null && _count != null && _count! > 0;
    final l = context.l10n;
    return AlertDialog(
      title: Text(l.findAndReplace),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TextField(
                key: const Key('find-replace-find'),
                controller: _find,
                autofocus: true,
                decoration: InputDecoration(
                  labelText: l.find,
                  errorText: _error,
                  errorMaxLines: 3,
                ),
                onChanged: (_) => _changed(),
              ),
              TextField(
                key: const Key('find-replace-replace'),
                controller: _replace,
                decoration: InputDecoration(
                  labelText: l.replaceWith,
                  helperText: _regex ? l.regexGroupsHelp : null,
                ),
                onChanged: (_) => _changed(),
                onSubmitted: (_) {
                  if (canApply) _apply();
                },
              ),
              const SizedBox(height: 12),
              // Rebuilt when the fields change, which can reset the choice.
              KeyedSubtree(
                key: ValueKey(_fields.join('\u0000')),
                child: DropdownButtonFormField<String>(
                  key: const Key('find-replace-in'),
                  initialValue: _in,
                  decoration: InputDecoration(labelText: l.inField),
                  items: [
                    DropdownMenuItem(
                      value: _allFields,
                      child: Text(l.allFields),
                    ),
                    DropdownMenuItem(value: _tags, child: Text(l.tags)),
                    for (final f in _fields)
                      DropdownMenuItem(value: f, child: Text(f)),
                  ],
                  onChanged: (v) {
                    _in = v!;
                    _changed(now: true);
                  },
                ),
              ),
              const SizedBox(height: 8),
              CheckboxListTile(
                key: const Key('find-replace-selected'),
                contentPadding: EdgeInsets.zero,
                title: Text(l.selectedNotesOnly),
                value: _selectedOnly,
                onChanged: widget.cardIds.isEmpty
                    ? null
                    : (v) {
                        _selectedOnly = v!;
                        _loadFields();
                        _changed(now: true);
                      },
              ),
              CheckboxListTile(
                key: const Key('find-replace-regex'),
                contentPadding: EdgeInsets.zero,
                title: Text(l.treatAsRegex),
                value: _regex,
                onChanged: (v) {
                  _regex = v!;
                  _changed(now: true);
                },
              ),
              CheckboxListTile(
                key: const Key('find-replace-ignore-case'),
                contentPadding: EdgeInsets.zero,
                title: Text(l.ignoreCase),
                value: _ignoreCase,
                onChanged: (v) {
                  _ignoreCase = v!;
                  _changed(now: true);
                },
              ),
              const SizedBox(height: 8),
              if (_error == null)
                Text(
                  _countText,
                  key: const Key('find-replace-count'),
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
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
          key: const Key('find-replace-apply'),
          onPressed: canApply ? _apply : null,
          child: Text(l.replace),
        ),
      ],
    );
  }
}
