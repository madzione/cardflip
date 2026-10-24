import 'package:flutter/material.dart';

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'browse_screen.dart';

/// The tag Tag duplicates adds.
const duplicateTag = 'duplicate';

/// Finds notes whose chosen field has the same content, optionally among
/// the notes a search matches, and lists them as sets. Selecting a set
/// browses its notes; Tag duplicates tags every listed note.
class FindDuplicatesScreen extends StatefulWidget {
  const FindDuplicatesScreen({super.key});

  @override
  State<FindDuplicatesScreen> createState() => _FindDuplicatesScreenState();
}

class _FindDuplicatesScreenState extends State<FindDuplicatesScreen> {
  final _query = TextEditingController();
  List<String> _fields = const [];
  String? _field;

  /// The sets of the last search; null before the first.
  List<golang.DuplicateSet>? _sets;
  String? _searchError;
  bool _finding = false;

  @override
  void initState() {
    super.initState();
    _loadFields();
  }

  @override
  void dispose() {
    _query.dispose();
    super.dispose();
  }

  /// The field names of all note types, each once, in note type and field
  /// order.
  Future<void> _loadFields() async {
    try {
      final types = await golang.notes.types();
      final fields = <String>[];
      for (final t in types) {
        for (final f in t.fields) {
          if (!fields.contains(f)) fields.add(f);
        }
      }
      if (!mounted) return;
      setState(() {
        _fields = fields;
        _field = fields.firstOrNull;
      });
    } catch (e) {
      showError(e);
    }
  }

  Future<void> _find() async {
    final field = _field;
    if (field == null) return;
    setState(() {
      _finding = true;
      _searchError = null;
    });
    try {
      final sets = await golang.browse.findDuplicates(field, _query.text);
      if (!mounted) return;
      setState(() {
        _sets = sets;
        _finding = false;
      });
    } on golang.GolangException catch (e) {
      // An invalid term: show why next to the search.
      if (!mounted) return;
      setState(() {
        _searchError = goText(e.message);
        _finding = false;
      });
    }
  }

  void _show(golang.DuplicateSet set) => Navigator.of(context).push(
    MaterialPageRoute<void>(
      builder: (_) => BrowseScreen(noteIds: set.noteIDs, query: ''),
    ),
  );

  Future<void> _tag() async {
    final undo = AppScope.read(context).undo;
    final notes = [for (final s in _sets!) ...s.noteIDs];
    final ok = await guard(() async {
      final cards = await golang.browse.noteCards(notes);
      await golang.browse.addTags(cards, [duplicateTag]);
    });
    if (ok) {
      undo.offer(appL10n.taggedNotes(notes.length, duplicateTag));
    }
  }

  @override
  Widget build(BuildContext context) {
    final sets = _sets;
    return Scaffold(
      appBar: AppBar(
        title: Text(context.l10n.findDuplicates),
        actions: [
          TextButton(
            key: const Key('find-duplicates-tag'),
            onPressed: sets == null || sets.isEmpty ? null : _tag,
            child: Text(context.l10n.tagDuplicates),
          ),
        ],
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Rebuilt when the fields arrive.
                KeyedSubtree(
                  key: ValueKey(_fields.join('\u0000')),
                  child: DropdownButtonFormField<String>(
                    key: const Key('find-duplicates-field'),
                    initialValue: _field,
                    decoration: InputDecoration(labelText: context.l10n.field),
                    items: [
                      for (final f in _fields)
                        DropdownMenuItem(value: f, child: Text(f)),
                    ],
                    onChanged: (v) => setState(() => _field = v),
                  ),
                ),
                const SizedBox(height: 8),
                TextField(
                  key: const Key('find-duplicates-search'),
                  controller: _query,
                  decoration: InputDecoration(
                    labelText: context.l10n.optionalSearch,
                    hintText: context.l10n.optionalSearchHint,
                    errorText: _searchError,
                    errorMaxLines: 3,
                  ),
                  onSubmitted: (_) => _find(),
                ),
                const SizedBox(height: 12),
                FilledButton(
                  key: const Key('find-duplicates-find'),
                  onPressed: _field == null || _finding ? null : _find,
                  child: Text(context.l10n.find),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(child: _results(sets)),
        ],
      ),
    );
  }

  Widget _results(List<golang.DuplicateSet>? sets) {
    if (_finding) return const Center(child: CircularProgressIndicator());
    if (sets == null) return const SizedBox();
    if (sets.isEmpty) {
      return Center(child: Text(context.l10n.noDuplicatesFound));
    }
    return ListView.builder(
      key: const Key('find-duplicates-results'),
      itemCount: sets.length,
      itemBuilder: (context, i) {
        final set = sets[i];
        return ListTile(
          key: Key('duplicate-set-$i'),
          title: Text(
            set.content,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          subtitle: Text(context.l10n.countNotes(set.noteIDs.length)),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => _show(set),
        );
      },
    );
  }
}
