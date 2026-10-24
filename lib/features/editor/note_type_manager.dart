import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/builtin_names.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'rich_field.dart';
import 'template_preview.dart';

/// [base] if no note type in [types] has that name (ignoring case), else
/// the first free "[base] 2", "[base] 3", ….
String freeNoteTypeName(String base, Iterable<golang.NoteType> types) {
  final taken = {for (final t in types) t.name.toLowerCase()};
  if (!taken.contains(base.toLowerCase())) return base;
  for (var i = 2; ; i++) {
    if (!taken.contains('$base $i'.toLowerCase())) return '$base $i';
  }
}

enum _TypeAction { clone, delete }

/// Lists the note types, and adds, clones and deletes them.
class NoteTypeListScreen extends StatefulWidget {
  const NoteTypeListScreen({super.key});

  @override
  State<NoteTypeListScreen> createState() => _NoteTypeListScreenState();
}

class _NoteTypeListScreenState extends State<NoteTypeListScreen> {
  late Future<List<golang.NoteType>> _types = golang.notes.types();
  StreamSubscription<golang.Change>? _changes;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Undo and redo bring note types back or remove them again.
    _changes ??= AppScope.maybeRead(context)?.changes.collection
        .where((c) => c.noteTypes || c.notes || c.all)
        .listen((_) => _reload());
  }

  @override
  void dispose() {
    _changes?.cancel();
    super.dispose();
  }

  void _reload() {
    if (!mounted) return;
    setState(() {
      _types = golang.notes.types();
    });
  }

  Future<void> _add(List<golang.NoteType> types) async {
    final List<String> stock;
    try {
      stock = await golang.notes.stockTypes();
    } catch (e) {
      showError(e);
      return;
    }
    if (!mounted) return;
    final added = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => _NoteTypeNameDialog(
        title: context.l10n.addNoteType,
        action: context.l10n.add,
        stock: stock,
        suggest: (s) =>
            freeNoteTypeName(stockTypeName(context.l10n, s!), types),
        submit: (s, name) => golang.notes.addType(s!, name),
      ),
    );
    if (added == true) _reload();
  }

  Future<void> _clone(golang.NoteType t, List<golang.NoteType> types) async {
    final cloned = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => _NoteTypeNameDialog(
        title: context.l10n.cloneNamed(t.name),
        action: context.l10n.clone,
        suggest: (_) => freeNoteTypeName(context.l10n.copyName(t.name), types),
        submit: (_, name) => golang.notes.cloneType(t.id, name),
      ),
    );
    if (cloned == true) _reload();
  }

  Future<void> _delete(golang.NoteType t) async {
    final golang.NoteTypeDeletionInfo info;
    try {
      info = await golang.notes.typeDeletionInfo(t.id);
    } catch (e) {
      showError(e);
      return;
    }
    if (!mounted) return;
    final ok = await confirm(
      context,
      title: context.l10n.deleteNamedTitle(t.name),
      message:
          '${info.notes == 0 ? context.l10n.noNotesUseType : context.l10n.deleteNoteTypeBody(info.notes, info.cards)} '
          '${context.l10n.canUndoUntilClose}',
      action: context.l10n.delete,
      destructive: true,
    );
    if (!ok || !mounted) return;
    final undo = AppScope.maybeRead(context)?.undo;
    final deleted = context.l10n.deletedNamed(t.name);
    if (await guard(() => golang.notes.deleteType(t.id))) {
      undo?.offer(deleted);
      _reload();
    }
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<golang.NoteType>>(
      future: _types,
      builder: (context, snap) {
        final types = snap.data;
        return Scaffold(
          appBar: AppBar(
            title: Text(context.l10n.noteTypes),
            actions: [
              IconButton(
                tooltip: context.l10n.addNoteType,
                icon: const Icon(Icons.add),
                onPressed: types == null ? null : () => _add(types),
              ),
            ],
          ),
          body: snap.hasError
              ? Center(child: Text(errorMessage(snap.error!)))
              : types == null
              ? const Center(child: CircularProgressIndicator())
              : ListView(
                  children: [
                    for (final t in types)
                      ListTile(
                        title: Text(t.name),
                        subtitle: Text(
                          context.l10n.noteTypeSummary(
                            t.fields.length,
                            t.templates.length,
                            t.noteCount,
                          ),
                        ),
                        trailing: PopupMenuButton<_TypeAction>(
                          popUpAnimationStyle: popupAnimation(context),
                          tooltip: context.l10n.noteTypeActions,
                          onSelected: (a) => switch (a) {
                            _TypeAction.clone => _clone(t, types),
                            _TypeAction.delete => _delete(t),
                          },
                          itemBuilder: (_) => [
                            PopupMenuItem(
                              value: _TypeAction.clone,
                              child: Text(context.l10n.clone),
                            ),
                            PopupMenuItem(
                              value: _TypeAction.delete,
                              // A collection always has a note type.
                              enabled: types.length > 1,
                              child: Text(context.l10n.delete),
                            ),
                          ],
                        ),
                        onTap: () async {
                          await Navigator.of(context).push(
                            MaterialPageRoute<void>(
                              builder: (_) => NoteTypeScreen(typeId: t.id),
                            ),
                          );
                          _reload();
                        },
                      ),
                  ],
                ),
        );
      },
    );
  }
}

/// Asks for the name of a new note type, and for [stock] the stock type it
/// starts from, then calls [submit]. Errors such as a taken name are shown
/// in the dialog. Pops true once [submit] succeeded.
class _NoteTypeNameDialog extends StatefulWidget {
  const _NoteTypeNameDialog({
    required this.title,
    required this.action,
    required this.suggest,
    required this.submit,
    this.stock,
  });

  final String title;
  final String action;

  /// The stock types to choose from; null asks only for the name.
  final List<String>? stock;

  /// The name proposed for a stock type (null without [stock]).
  final String Function(String? stock) suggest;
  final Future<void> Function(String? stock, String name) submit;

  @override
  State<_NoteTypeNameDialog> createState() => _NoteTypeNameDialogState();
}

class _NoteTypeNameDialogState extends State<_NoteTypeNameDialog> {
  late String? _stock = widget.stock?.first;
  late final _name = TextEditingController(text: widget.suggest(_stock));
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _name.dispose();
    super.dispose();
  }

  void _pick(String? stock) {
    if (stock == null) return;
    setState(() {
      // A name the user has not changed follows the stock type.
      if (_name.text == widget.suggest(_stock)) {
        _name.text = widget.suggest(stock);
      }
      _stock = stock;
    });
  }

  Future<void> _submit() async {
    setState(() => _busy = true);
    try {
      await widget.submit(_stock, _name.text);
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          _busy = false;
          _error = errorMessage(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final stock = widget.stock;
    return AlertDialog(
      title: Text(widget.title),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (stock != null)
              RadioGroup<String>(
                groupValue: _stock,
                onChanged: _pick,
                child: Column(
                  children: [
                    for (final s in stock)
                      RadioListTile<String>(
                        value: s,
                        title: Text(stockTypeName(context.l10n, s)),
                        dense: true,
                        contentPadding: EdgeInsets.zero,
                      ),
                  ],
                ),
              ),
            TextField(
              key: const Key('note-type-name'),
              controller: _name,
              autofocus: true,
              decoration: InputDecoration(
                labelText: context.l10n.name,
                errorText: _error,
              ),
              onChanged: (_) {
                if (_error != null) setState(() => _error = null);
              },
              onSubmitted: (_) => _submit(),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          onPressed: _busy ? null : _submit,
          child: Text(widget.action),
        ),
      ],
    );
  }
}

/// Edits one note type: its name, fields, card types and styling.
class NoteTypeScreen extends StatefulWidget {
  const NoteTypeScreen({super.key, required this.typeId});

  final int typeId;

  @override
  State<NoteTypeScreen> createState() => _NoteTypeScreenState();
}

class _NoteTypeScreenState extends State<NoteTypeScreen> {
  golang.NoteType? _type;

  @override
  void initState() {
    super.initState();
    _reload();
  }

  Future<void> _reload() async {
    try {
      final t = await golang.notes.type_(widget.typeId);
      if (mounted) setState(() => _type = t);
    } catch (e) {
      showError(e);
    }
  }

  Future<void> _change(Future<void> Function() action) async {
    if (await guard(action)) await _reload();
  }

  Future<void> _rename() async {
    final name = await askText(
      context,
      title: context.l10n.renameNoteType,
      initial: _type!.name,
      label: context.l10n.name,
    );
    if (name != null) {
      await _change(() => golang.notes.renameType(widget.typeId, name));
    }
  }

  Future<void> _renameField(int i) async {
    final name = await askText(
      context,
      title: context.l10n.renameField,
      initial: _type!.fields[i],
      label: context.l10n.name,
    );
    if (name != null) {
      await _change(() => golang.notes.renameField(widget.typeId, i, name));
    }
  }

  Future<void> _addField() async {
    final name = await askText(
      context,
      title: context.l10n.addField,
      label: context.l10n.name,
      action: context.l10n.add,
    );
    if (name != null) {
      await _change(() => golang.notes.addField(widget.typeId, name));
    }
  }

  Future<void> _removeField(int i) async {
    final t = _type!;
    final ok = await confirm(
      context,
      title: context.l10n.removeFieldTitle(t.fields[i]),
      message: context.l10n.removeFieldBody(t.noteCount),
      action: context.l10n.remove,
      destructive: true,
    );
    if (ok) await _change(() => golang.notes.removeField(widget.typeId, i));
  }

  Future<void> _editTemplate(int? i) async {
    final t = _type!;
    await Navigator.of(context).push(
      MaterialPageRoute<void>(
        builder: (_) => TemplateEditorScreen(
          noteType: t,
          index: i,
          template: i == null
              ? golang.CardTemplate(
                  name: context.l10n.defaultCardTypeName(
                    t.templates.length + 1,
                  ),
                  qFmt: '{{${t.fields.last}}}',
                  aFmt:
                      '{{FrontSide}}\n\n<hr id=answer>\n\n{{${t.fields.first}}}',
                )
              : t.templates[i],
        ),
      ),
    );
    await _reload();
  }

  Future<void> _removeTemplate(int i) async {
    final t = _type!;
    final int cards;
    try {
      cards = await golang.notes.templateCardCount(widget.typeId, i);
    } catch (e) {
      showError(e);
      return;
    }
    if (!mounted) return;
    final ok = await confirm(
      context,
      title: context.l10n.removeCardTypeTitle(t.templates[i].name),
      message: context.l10n.removeCardTypeBody(cards),
      action: context.l10n.remove,
      destructive: true,
    );
    if (ok) await _change(() => golang.notes.removeTemplate(widget.typeId, i));
  }

  /// [t] with the given parts replaced.
  static golang.NoteType _copy(
    golang.NoteType t, {
    List<String>? fields,
    List<golang.FieldOptions>? fieldOptions,
    int? sortField,
    List<golang.CardTemplate>? templates,
  }) => golang.NoteType(
    id: t.id,
    name: t.name,
    cloze: t.cloze,
    imageOcclusion: t.imageOcclusion,
    fields: fields ?? t.fields,
    fieldOptions: fieldOptions ?? t.fieldOptions,
    sortField: sortField ?? t.sortField,
    templates: templates ?? t.templates,
    css: t.css,
    noteCount: t.noteCount,
  );

  /// A copy of [list] with item [from] moved to index [to], as Go moves it.
  static List<T> _moved<T>(List<T> list, int from, int to) {
    final out = [...list];
    return out..insert(to, out.removeAt(from));
  }

  Future<void> _moveField(int from, int to) async {
    if (to == from) return;
    final t = _type!;
    // Show the new order right away; the reload confirms it. The sort
    // field stays on its field.
    final sorted = t.fields[t.sortField];
    final fields = _moved(t.fields, from, to);
    setState(() {
      _type = _copy(
        t,
        fields: fields,
        fieldOptions: t.fieldOptions.length == t.fields.length
            ? _moved(t.fieldOptions, from, to)
            : null,
        sortField: fields.indexOf(sorted),
      );
    });
    await _change(() => golang.notes.moveField(widget.typeId, from, to));
  }

  Future<void> _fieldOptions(int i) async {
    final saved = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => FieldOptionsDialog(noteType: _type!, index: i),
    );
    if (saved == true) await _reload();
  }

  Future<void> _moveTemplate(int from, int to) async {
    if (to == from) return;
    final t = _type!;
    // Show the new order right away; the reload confirms it.
    setState(() {
      _type = _copy(t, templates: _moved(t.templates, from, to));
    });
    await _change(() => golang.notes.moveTemplate(widget.typeId, from, to));
  }

  Future<void> _templateOptions(int i) async {
    final saved = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (_) => CardTypeOptionsDialog(
        typeId: widget.typeId,
        index: i,
        template: _type!.templates[i],
      ),
    );
    if (saved == true) await _reload();
  }

  Future<void> _editCSS() async {
    final css = await Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => CssEditorScreen(noteType: _type!)),
    );
    if (css != null) {
      await _change(() => golang.notes.setCSS(widget.typeId, css));
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = _type;
    if (t == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    final heading = Theme.of(context).textTheme.titleSmall
        ?.copyWith(color: Theme.of(context).colorScheme.primary);
    final l = context.l10n;
    return Scaffold(
      appBar: AppBar(
        title: Text(t.name),
        actions: [
          IconButton(
            tooltip: l.rename,
            icon: const Icon(Icons.edit),
            onPressed: _rename,
          ),
        ],
      ),
      body: ListView(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text(l.fields, style: heading),
          ),
          ReorderableListView(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            buildDefaultDragHandles: false,
            onReorderItem: _moveField,
            children: [
              for (var i = 0; i < t.fields.length; i++)
                ListTile(
                  key: ValueKey('field-$i-${t.fields[i]}'),
                  title: Text(t.fields[i]),
                  subtitle: switch (_fieldSummary(t, i)) {
                    '' => null,
                    final summary => Text(
                      summary,
                      key: Key('field-summary-$i'),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  },
                  onTap: () => _fieldOptions(i),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        tooltip: l.fieldOptions,
                        icon: const Icon(Icons.tune),
                        onPressed: () => _fieldOptions(i),
                      ),
                      IconButton(
                        tooltip: l.renameField,
                        icon: const Icon(Icons.edit_outlined),
                        onPressed: () => _renameField(i),
                      ),
                      IconButton(
                        tooltip: l.removeField,
                        icon: const Icon(Icons.delete_outline),
                        onPressed: () => _removeField(i),
                      ),
                      if (t.fields.length > 1)
                        ReorderableDragStartListener(
                          key: Key('move-field-$i'),
                          index: i,
                          child: Tooltip(
                            message: l.dragToReorder,
                            child: const Padding(
                              padding: EdgeInsets.all(8),
                              child: Icon(Icons.drag_handle),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
            ],
          ),
          ListTile(
            leading: const Icon(Icons.add),
            title: Text(l.addField),
            onTap: _addField,
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text(l.cardTypes, style: heading),
          ),
          ReorderableListView(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            buildDefaultDragHandles: false,
            onReorderItem: _moveTemplate,
            children: [
              for (var i = 0; i < t.templates.length; i++)
                ListTile(
                  key: ValueKey('template-$i-${t.templates[i].name}'),
                  title: Text(t.templates[i].name),
                  subtitle: Text(
                    t.templates[i].qFmt,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  onTap: () => _editTemplate(i),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        tooltip: l.cardTypeOptions,
                        icon: const Icon(Icons.tune),
                        onPressed: () => _templateOptions(i),
                      ),
                      if (!t.cloze) ...[
                        IconButton(
                          tooltip: l.removeCardType,
                          icon: const Icon(Icons.delete_outline),
                          onPressed: () => _removeTemplate(i),
                        ),
                        if (t.templates.length > 1)
                          ReorderableDragStartListener(
                            key: Key('move-template-$i'),
                            index: i,
                            child: Tooltip(
                              message: l.dragToReorder,
                              child: const Padding(
                                padding: EdgeInsets.all(8),
                                child: Icon(Icons.drag_handle),
                              ),
                            ),
                          ),
                      ],
                    ],
                  ),
                ),
            ],
          ),
          if (!t.cloze)
            ListTile(
              leading: const Icon(Icons.add),
              title: Text(l.addCardType),
              onTap: () => _editTemplate(null),
            ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text(l.styling, style: heading),
          ),
          ListTile(
            leading: const Icon(Icons.palette_outlined),
            title: Text(l.editCss),
            onTap: _editCSS,
          ),
        ],
      ),
    );
  }
}

/// The options of field [i] of [t] that differ from the defaults, for the
/// field's list entry; empty when there are none.
String _fieldSummary(golang.NoteType t, int i) {
  final o = t.fieldOption(i);
  final l = appL10n;
  return [
    if (t.sortField == i) l.sortField,
    if (o.font.isNotEmpty || o.size > 0)
      [if (o.font.isNotEmpty) o.font, if (o.size > 0) '${o.size}'].join(' '),
    if (o.rtl) l.rightToLeft,
    if (o.collapsed) l.collapsed,
    if (o.excludeFromSearch) l.notInUnqualifiedSearches,
    if (o.htmlByDefault) l.htmlEditor,
    if (o.sticky) l.pinned,
  ].join(' · ');
}

/// The options of field [index] of [noteType]: how the editor shows it,
/// whether the browser sorts by it and whether unqualified searches look in
/// it. Pops true once saved.
class FieldOptionsDialog extends StatefulWidget {
  const FieldOptionsDialog({
    super.key,
    required this.noteType,
    required this.index,
  });

  final golang.NoteType noteType;
  final int index;

  @override
  State<FieldOptionsDialog> createState() => _FieldOptionsDialogState();
}

class _FieldOptionsDialogState extends State<FieldOptionsDialog> {
  late final golang.FieldOptions _old = widget.noteType.fieldOption(
    widget.index,
  );
  late final _font = TextEditingController(text: _old.font);
  late final _size = TextEditingController(
    text: _old.size > 0 ? '${_old.size}' : '',
  );
  late bool _sort = widget.noteType.sortField == widget.index;
  late bool _rtl = _old.rtl;
  late bool _collapsed = _old.collapsed;
  late bool _exclude = _old.excludeFromSearch;
  late bool _html = _old.htmlByDefault;
  late bool _sticky = _old.sticky;
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _font.dispose();
    _size.dispose();
    super.dispose();
  }

  /// The size typed in: 0 for empty (the default), null when invalid.
  int? get _sizeValue {
    final text = _size.text.trim();
    if (text.isEmpty) return 0;
    final n = int.tryParse(text);
    return n != null && n > 0 && n <= 200 ? n : null;
  }

  Future<void> _save() async {
    final size = _sizeValue;
    if (size == null) {
      setState(() => _error = context.l10n.fieldSizeError);
      return;
    }
    final options = golang.FieldOptions(
      htmlByDefault: _html,
      sticky: _sticky,
      font: _font.text.trim(),
      size: size,
      rtl: _rtl,
      collapsed: _collapsed,
      excludeFromSearch: _exclude,
    );
    final o = _old;
    final changed =
        options.htmlByDefault != o.htmlByDefault ||
        options.sticky != o.sticky ||
        options.font != o.font ||
        options.size != o.size ||
        options.rtl != o.rtl ||
        options.collapsed != o.collapsed ||
        options.excludeFromSearch != o.excludeFromSearch;
    final t = widget.noteType;
    setState(() => _busy = true);
    try {
      if (changed) {
        await golang.notes.setFieldOptions(t.id, widget.index, options);
      }
      if (_sort && t.sortField != widget.index) {
        await golang.notes.setSortField(t.id, widget.index);
      }
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          _busy = false;
          _error = errorMessage(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    final t = widget.noteType;
    final isSortField = t.sortField == widget.index;
    Widget toggle(
      String key,
      String title,
      String subtitle,
      bool value,
      ValueChanged<bool>? onChanged,
    ) => SwitchListTile(
      key: Key(key),
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      subtitle: Text(subtitle),
      value: value,
      onChanged: onChanged,
    );
    return AlertDialog(
      title: Text(l.optionsForNamed(t.fields[widget.index])),
      content: SizedBox(
        width: 480,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l.editingFont,
                style: Theme.of(context).textTheme.titleSmall,
              ),
              const SizedBox(height: 4),
              Text(
                l.editingFontHelp,
                style: Theme.of(context).textTheme.bodySmall,
              ),
              const SizedBox(height: 8),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    flex: 3,
                    child: TextField(
                      key: const Key('field-font'),
                      controller: _font,
                      decoration: InputDecoration(
                        labelText: l.font,
                        hintText: l.defaultValue,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: TextField(
                      key: const Key('field-size'),
                      controller: _size,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        labelText: l.size,
                        hintText: l.defaultValue,
                      ),
                      onChanged: (_) {
                        if (_error != null) setState(() => _error = null);
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              // One field always sorts: another field takes over by being
              // chosen.
              toggle(
                'field-sort',
                l.sortByField,
                isSortField ? l.sortByFieldIsHelp : l.sortByFieldHelp,
                _sort,
                isSortField ? null : (on) => setState(() => _sort = on),
              ),
              toggle(
                'field-rtl',
                l.rightToLeft,
                l.rightToLeftHelp,
                _rtl,
                (on) => setState(() => _rtl = on),
              ),
              toggle(
                'field-collapsed',
                l.collapseByDefault,
                l.collapseByDefaultHelp,
                _collapsed,
                (on) => setState(() => _collapsed = on),
              ),
              toggle(
                'field-exclude',
                l.excludeFromSearch,
                l.excludeFromSearchHelp(t.fields[widget.index].toLowerCase()),
                _exclude,
                (on) => setState(() => _exclude = on),
              ),
              toggle(
                'html-by-default',
                l.htmlByDefault,
                l.htmlByDefaultHelp,
                _html,
                (on) => setState(() => _html = on),
              ),
              toggle(
                'field-sticky',
                l.keepAfterAdding,
                l.keepAfterAddingHelp,
                _sticky,
                (on) => setState(() => _sticky = on),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(
                  _error!,
                  key: const Key('field-options-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
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
        FilledButton(onPressed: _busy ? null : _save, child: Text(l.save)),
      ],
    );
  }
}

/// The options of a card type: the deck its new cards go to, and the
/// simplified templates the browser shows. Pops true once saved.
class CardTypeOptionsDialog extends StatefulWidget {
  const CardTypeOptionsDialog({
    super.key,
    required this.typeId,
    required this.index,
    required this.template,
  });

  final int typeId;
  final int index;
  final golang.CardTemplate template;

  @override
  State<CardTypeOptionsDialog> createState() => _CardTypeOptionsDialogState();
}

class _CardTypeOptionsDialogState extends State<CardTypeOptionsDialog> {
  late final _browserQ = TextEditingController(
    text: widget.template.browserQFmt ?? '',
  );
  late final _browserA = TextEditingController(
    text: widget.template.browserAFmt ?? '',
  );
  late int _deck = widget.template.deckOverride ?? 0;
  List<golang.DeckNode>? _decks;
  String? _error;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    unawaited(_loadDecks());
  }

  Future<void> _loadDecks() async {
    try {
      final decks = await golang.decks.tree();
      if (mounted) setState(() => _decks = decks);
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    }
  }

  @override
  void dispose() {
    _browserQ.dispose();
    _browserA.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final t = widget.template;
    String? text(TextEditingController c) =>
        c.text.trim().isEmpty ? null : c.text;
    setState(() => _busy = true);
    try {
      await golang.notes.updateTemplate(
        widget.typeId,
        widget.index,
        golang.CardTemplate(
          name: t.name,
          qFmt: t.qFmt,
          aFmt: t.aFmt,
          deckOverride: _deck == 0 ? null : _deck,
          browserQFmt: text(_browserQ),
          browserAFmt: text(_browserA),
        ),
      );
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          _busy = false;
          _error = errorMessage(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    const mono = TextStyle(fontFamily: 'monospace');
    final decks = _decks;
    // New cards cannot go to filtered decks; an override deck that is gone
    // shows as none.
    final choices = [
      for (final d in decks ?? const <golang.DeckNode>[])
        if (!d.filtered) d,
    ];
    final deck = choices.any((d) => d.id == _deck) ? _deck : 0;
    return AlertDialog(
      title: Text(l.optionsForNamed(widget.template.name)),
      content: SizedBox(
        width: 480,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (decks == null && _error == null)
                const Center(child: CircularProgressIndicator())
              else
                DropdownButtonFormField<int>(
                  key: const Key('deck-override'),
                  isExpanded: true,
                  initialValue: deck,
                  decoration: InputDecoration(
                    labelText: l.deckOverride,
                    helperText: l.deckOverrideHelp,
                    helperMaxLines: 3,
                  ),
                  items: [
                    DropdownMenuItem(value: 0, child: Text(l.deckOverrideNone)),
                    for (final d in choices)
                      DropdownMenuItem(
                        value: d.id,
                        child: Text(
                          d.fullName,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                  ],
                  onChanged: (v) => setState(() => _deck = v ?? 0),
                ),
              const SizedBox(height: 24),
              Text(
                l.browserAppearance,
                style: Theme.of(context).textTheme.titleSmall,
              ),
              const SizedBox(height: 4),
              Text(
                l.browserAppearanceHelp,
                style: Theme.of(context).textTheme.bodySmall,
              ),
              const SizedBox(height: 12),
              TextField(
                key: const Key('browser-front-template'),
                controller: _browserQ,
                style: mono,
                maxLines: 3,
                minLines: 1,
                decoration: InputDecoration(
                  labelText: l.browserQuestionTemplate,
                  border: const OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                key: const Key('browser-back-template'),
                controller: _browserA,
                style: mono,
                maxLines: 3,
                minLines: 1,
                decoration: InputDecoration(
                  labelText: l.browserAnswerTemplate,
                  border: const OutlineInputBorder(),
                ),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(
                  _error!,
                  key: const Key('card-type-options-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
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
        FilledButton(onPressed: _busy ? null : _save, child: Text(l.save)),
      ],
    );
  }
}

/// Shows [editor] and a [TemplatePreview] side by side on wide screens and
/// as two tabs on narrow ones.
class _EditorWithPreview extends StatelessWidget {
  const _EditorWithPreview({
    required this.title,
    required this.actions,
    required this.editor,
    required this.preview,
  });

  /// From this window width the editor and the preview sit side by side.
  static const sideBySideWidth = 840.0;

  final String title;
  final List<Widget> actions;
  final Widget editor;
  final Widget preview;

  @override
  Widget build(BuildContext context) {
    if (MediaQuery.sizeOf(context).width >= sideBySideWidth) {
      return Scaffold(
        appBar: AppBar(title: Text(title), actions: actions),
        body: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(child: editor),
            const VerticalDivider(width: 1),
            Expanded(child: preview),
          ],
        ),
      );
    }
    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: Text(title),
          actions: actions,
          bottom: TabBar(
            tabs: [
              Tab(text: context.l10n.edit),
              Tab(text: context.l10n.preview),
            ],
          ),
        ),
        body: TabBarView(children: [editor, preview]),
      ),
    );
  }
}

/// Edits a card type's name and question/answer templates, with a live
/// preview.
class TemplateEditorScreen extends StatefulWidget {
  const TemplateEditorScreen({
    super.key,
    required this.noteType,
    required this.index,
    required this.template,
  });

  final golang.NoteType noteType;

  /// The card type to change; null adds a new one.
  final int? index;
  final golang.CardTemplate template;

  /// How long after the last keystroke the templates are checked.
  static const checkDebounce = Duration(milliseconds: 300);

  @override
  State<TemplateEditorScreen> createState() => _TemplateEditorScreenState();
}

class _TemplateEditorScreenState extends State<TemplateEditorScreen> {
  late final _name = TextEditingController(text: widget.template.name);
  late final _q = TextEditingController(text: widget.template.qFmt);
  late final _a = TextEditingController(text: widget.template.aFmt);
  String? _error;

  /// What would refuse saving the templates as they are; Save is disabled
  /// while there is any.
  List<String> _problems = const [];
  Timer? _checkTimer;

  /// Numbers the checks, so an answer to an outdated one is dropped.
  int _checkRequest = 0;

  @override
  void initState() {
    super.initState();
    unawaited(_check());
  }

  void _templateChanged() {
    setState(() {});
    _checkTimer?.cancel();
    _checkTimer = Timer(TemplateEditorScreen.checkDebounce, _check);
  }

  Future<void> _check() async {
    final request = ++_checkRequest;
    List<String> problems;
    try {
      problems = await golang.notes.checkTemplate(
        widget.noteType.id,
        widget.index ?? -1,
        _template,
      );
    } catch (_) {
      // The check is only a hint; saving checks again.
      problems = const [];
    }
    if (mounted && request == _checkRequest) {
      setState(() => _problems = problems);
    }
  }

  @override
  void dispose() {
    _checkTimer?.cancel();
    _name.dispose();
    _q.dispose();
    _a.dispose();
    super.dispose();
  }

  /// The edited card type; its options (deck override, browser
  /// appearance) stay as they are.
  golang.CardTemplate get _template => golang.CardTemplate(
    name: _name.text,
    qFmt: _q.text,
    aFmt: _a.text,
    deckOverride: widget.template.deckOverride,
    browserQFmt: widget.template.browserQFmt,
    browserAFmt: widget.template.browserAFmt,
  );

  Future<void> _save() async {
    final typeId = widget.noteType.id;
    try {
      if (widget.index == null) {
        await golang.notes.addTemplate(typeId, _template);
      } else {
        await golang.notes.updateTemplate(typeId, widget.index!, _template);
      }
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = errorMessage(e));
    }
  }

  @override
  Widget build(BuildContext context) {
    const mono = TextStyle(fontFamily: 'monospace');
    final errorColor = Theme.of(context).colorScheme.error;
    final l = context.l10n;
    final editor = ListView(
      padding: const EdgeInsets.all(16),
      children: [
        TextField(
          controller: _name,
          decoration: InputDecoration(labelText: l.name),
          onChanged: (_) => setState(() {}),
        ),
        const SizedBox(height: 16),
        TextField(
          key: const Key('front-template'),
          controller: _q,
          style: mono,
          maxLines: 6,
          decoration: InputDecoration(
            labelText: l.frontTemplate,
            border: const OutlineInputBorder(),
          ),
          onChanged: (_) => _templateChanged(),
        ),
        const SizedBox(height: 16),
        TextField(
          key: const Key('back-template'),
          controller: _a,
          style: mono,
          maxLines: 6,
          decoration: InputDecoration(
            labelText: l.backTemplate,
            border: const OutlineInputBorder(),
          ),
          onChanged: (_) => _templateChanged(),
        ),
        for (final problem in _problems) ...[
          const SizedBox(height: 8),
          Row(
            key: const Key('template-problem'),
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(Icons.error_outline, size: 18, color: errorColor),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  goText(problem),
                  style: TextStyle(color: errorColor),
                ),
              ),
            ],
          ),
        ],
        const SizedBox(height: 8),
        Text(l.templateHelp, style: Theme.of(context).textTheme.bodySmall),
        if (_error != null) ...[
          const SizedBox(height: 12),
          Text(_error!, style: TextStyle(color: errorColor)),
        ],
      ],
    );
    return _EditorWithPreview(
      title: widget.index == null ? l.newCardType : l.editCardType,
      actions: [
        TextButton(
          onPressed: _problems.isEmpty ? _save : null,
          child: Text(l.save),
        ),
      ],
      editor: editor,
      preview: TemplatePreview(
        noteType: widget.noteType,
        ord: widget.index ?? widget.noteType.templates.length,
        template: _template,
        css: widget.noteType.css,
      ),
    );
  }
}

/// Edits a note type's CSS, with a live preview of one of its card types.
/// Pops the CSS on save.
class CssEditorScreen extends StatefulWidget {
  const CssEditorScreen({super.key, required this.noteType});

  final golang.NoteType noteType;

  @override
  State<CssEditorScreen> createState() => _CssEditorScreenState();
}

class _CssEditorScreenState extends State<CssEditorScreen> {
  late final _css = TextEditingController(text: widget.noteType.css);
  int _ord = 0;

  @override
  void dispose() {
    _css.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final templates = widget.noteType.templates;
    final editor = ListView(
      padding: const EdgeInsets.all(16),
      children: [
        TextField(
          key: const Key('css-input'),
          controller: _css,
          maxLines: 18,
          style: const TextStyle(fontFamily: 'monospace'),
          decoration: const InputDecoration(border: OutlineInputBorder()),
          onChanged: (_) => setState(() {}),
        ),
      ],
    );
    final preview = Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (templates.length > 1)
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
            child: DropdownButton<int>(
              key: const Key('preview-card-type'),
              value: _ord,
              onChanged: (i) => setState(() => _ord = i!),
              items: [
                for (var i = 0; i < templates.length; i++)
                  DropdownMenuItem(value: i, child: Text(templates[i].name)),
              ],
            ),
          ),
        Expanded(
          child: TemplatePreview(
            noteType: widget.noteType,
            ord: _ord,
            template: templates[_ord],
            css: _css.text,
          ),
        ),
      ],
    );
    return _EditorWithPreview(
      title: context.l10n.stylingCss,
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, _css.text),
          child: Text(context.l10n.save),
        ),
      ],
      editor: editor,
      preview: preview,
    );
  }
}
