import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/decks/deck_list_screen.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

/// How a sidebar item's search combines with the current search.
enum SidebarSearch {
  /// Replaces the search (click).
  replace,

  /// Adds the item with AND (Ctrl-click, Cmd-click on macOS).
  and,

  /// Adds the item with OR (Shift-click).
  or,

  /// Adds the item negated (Alt-click, Option-click on macOS).
  not,
}

/// The search term for qualifier [key] and [name]: `deck:Spanish`, or
/// `deck:"My deck"` when the name has spaces, parentheses or quotes.
String sidebarTerm(String key, String name) {
  final escaped = name.replaceAll(r'\', r'\\');
  if (RegExp(r'[\s()"]').hasMatch(name)) {
    return '$key:"${escaped.replaceAll('"', r'\"')}"';
  }
  return '$key:$escaped';
}

/// Combines the [current] search with [term] as [how] says. AND binds more
/// tightly than OR, so a search with a top-level `or` is grouped before an
/// AND or a negated term is added.
String composeSearch(String current, String term, SidebarSearch how) {
  final q = current.trim();
  final t = how == SidebarSearch.not ? '-$term' : term;
  if (how == SidebarSearch.replace || q.isEmpty) return t;
  if (how == SidebarSearch.or) return '$q or $t';
  return _hasTopLevelOr(q) ? '($q) $t' : '$q $t';
}

/// Whether [q] has an `or` outside quotes and parentheses.
bool _hasTopLevelOr(String q) {
  var depth = 0;
  var quoted = false;
  final word = StringBuffer();
  bool endWord() {
    final isOr = depth == 0 && word.toString().toLowerCase() == 'or';
    word.clear();
    return isOr;
  }

  for (var i = 0; i < q.length; i++) {
    final c = q[i];
    if (c == r'\' && i + 1 < q.length) {
      word.write(q.substring(i, i + 2));
      i++;
    } else if (c == '"') {
      quoted = !quoted;
      word.write(c);
    } else if (quoted) {
      word.write(c);
    } else if (c == '(' || c == ')' || c.trim().isEmpty) {
      if (endWord()) return true;
      if (c == '(') depth++;
      if (c == ')') depth--;
    } else {
      word.write(c);
    }
  }
  return endWord();
}

/// A deck or tag in a sidebar tree. Parents that are only part of a longer
/// name (tag `verb` of `verb::irregular`) are nodes too.
class SidebarNode<T> {
  SidebarNode(this.name, this.fullName);

  /// The last part of the name, e.g. `irregular`.
  final String name;

  /// The whole name, e.g. `verb::irregular`.
  final String fullName;

  /// What the node stands for, e.g. its deck; null for a parent that is
  /// only part of a longer name.
  T? value;
  final children = <SidebarNode<T>>[];
}

/// Builds a tree from `::`-separated names, matching parents ignoring case.
/// Children are kept in the order of [names].
List<SidebarNode<T>> sidebarTree<T>(Iterable<(String, T?)> names) {
  final roots = <SidebarNode<T>>[];
  final byName = <String, SidebarNode<T>>{};
  for (final (fullName, value) in names) {
    final parts = fullName.split('::');
    var siblings = roots;
    SidebarNode<T>? node;
    for (var i = 0; i < parts.length; i++) {
      final path = parts.take(i + 1).join('::');
      node = byName.putIfAbsent(path.toLowerCase(), () {
        final n = SidebarNode<T>(parts[i], path);
        siblings.add(n);
        return n;
      });
      siblings = node.children;
    }
    node!.value ??= value;
  }
  return roots;
}

/// What is dragged in the sidebar: tags, or a deck.
sealed class _Drag {
  const _Drag();
}

class _TagDrag extends _Drag {
  const _TagDrag(this.tags);
  final List<String> tags;
}

class _DeckDrag extends _Drag {
  const _DeckDrag(this.deck);
  final golang.DeckNode deck;
}

/// Whether [name] is [root] or below it, ignoring case.
bool _inTree(String name, String root) {
  final n = name.toLowerCase(), r = root.toLowerCase();
  return n == r || n.startsWith('$r::');
}

/// The browser's sidebar: card states, decks, note types and tags. Clicking
/// an item searches for it ([onSearch]); decks and tags can be renamed,
/// deleted and moved by dragging or from their menu.
class BrowseSidebar extends StatefulWidget {
  const BrowseSidebar({super.key, required this.onSearch});

  final void Function(String term, SidebarSearch how) onSearch;

  @override
  State<BrowseSidebar> createState() => _BrowseSidebarState();
}

class _BrowseSidebarState extends State<BrowseSidebar> {
  static List<(String, String, IconData)> _states(AppLocalizations l) => [
    (l.countNew, 'is:new', Icons.fiber_new_outlined),
    (l.stateLearning, 'is:learn', Icons.school_outlined),
    (l.stateReview, 'is:review', Icons.replay),
    (l.stateDue, 'is:due', Icons.schedule),
    (l.stateSuspendedTitle, 'is:suspended', Icons.pause_circle_outline),
  ];

  final _filter = TextEditingController();
  final _mouse = RendererBinding.instance.mouseTracker;
  StreamSubscription<golang.Change>? _changes;
  List<String> _tags = const [];
  List<String> _noteTypes = const [];

  /// Collapsed sections, and expanded deck and tag nodes, by key.
  final _collapsed = <String>{};
  final _expanded = <String>{};

  /// Whether tags are being picked, and the picked ones by lower-case name.
  bool _selectingTags = false;
  final _selectedTags = <String, String>{};

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _changes ??= AppScope.of(context).changes.collection.listen((c) {
      if (c.all || c.notes || c.noteTypes) _load();
    });
  }

  @override
  void dispose() {
    _changes?.cancel();
    _filter.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final tags = await golang.notes.tags();
      final types = await golang.notes.types();
      if (!mounted) return;
      setState(() {
        _tags = tags;
        _noteTypes = [for (final t in types) t.name];
        _selectedTags.removeWhere((k, _) => !tags.any((t) => _inTree(t, k)));
      });
    } catch (e) {
      showError(e);
    }
  }

  /// How a click combines with the search, from the modifier keys held.
  static SidebarSearch _clickSearch() {
    final keys = HardwareKeyboard.instance;
    final mac =
        defaultTargetPlatform == TargetPlatform.macOS ||
        defaultTargetPlatform == TargetPlatform.iOS;
    if (keys.isAltPressed) return SidebarSearch.not;
    if (mac ? keys.isMetaPressed : keys.isControlPressed) {
      return SidebarSearch.and;
    }
    if (keys.isShiftPressed) return SidebarSearch.or;
    return SidebarSearch.replace;
  }

  void _toggleTag(String tag) => setState(() {
    final key = tag.toLowerCase();
    if (_selectedTags.remove(key) == null) _selectedTags[key] = tag;
  });

  void _endSelectingTags() => setState(() {
    _selectingTags = false;
    _selectedTags.clear();
  });

  // Tag and deck actions.

  Future<void> _renameTag(String tag) async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final name = await askText(
      context,
      title: l.renameTag,
      initial: tag,
      label: l.name,
      helper: l.renameTagHelper,
      action: l.rename,
    );
    if (name == null || name.trim() == tag) return;
    var changed = 0;
    if (await guard(
          () async => changed = await golang.notes.renameTag(tag, name),
        ) &&
        changed > 0) {
      undo.offer(l.renamedTagOn(changed));
    }
  }

  Future<void> _deleteTags(List<String> tags) async {
    final undo = AppScope.read(context).undo;
    var count = 0;
    if (!await guard(() async => count = await golang.notes.tagUsage(tags))) {
      return;
    }
    if (!mounted) return;
    final l = context.l10n;
    final ok = await confirm(
      context,
      title: tags.length == 1
          ? l.deleteTagTitle(tags.single)
          : l.deleteTagsTitle(tags.length),
      message: tags.length == 1
          ? l.deleteTagBody(count)
          : l.deleteTagsBody(count),
      action: l.delete,
      destructive: true,
    );
    if (!ok) return;
    var changed = 0;
    if (await guard(
      () async => changed = await golang.notes.deleteTags(tags),
    )) {
      if (_selectingTags) _endSelectingTags();
      if (changed > 0) {
        undo.offer(l.removedTagsFrom(tags.length, changed));
      }
    }
  }

  /// Moves [tags] under [parent] ("" for the top level).
  Future<void> _moveTags(List<String> tags, String parent) async {
    final undo = AppScope.read(context).undo;
    var changed = 0;
    if (!await guard(
      () async => changed = await golang.notes.moveTags(tags, parent),
    )) {
      return;
    }
    if (_selectingTags) _endSelectingTags();
    if (changed > 0) {
      undo.offer(
        parent.isEmpty
            ? appL10n.movedTagsTop(tags.length)
            : appL10n.movedTagsUnder(tags.length, parent),
      );
    }
  }

  /// Asks where to move [tags] and moves them.
  Future<void> _pickTagParent(List<String> tags) async {
    const top = '\u0000';
    final parent = await showDialog<String>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => SimpleDialog(
        title: Text(context.l10n.moveTagsTo(tags.length)),
        children: [
          SimpleDialogOption(
            onPressed: () => Navigator.pop(context, top),
            child: Text(context.l10n.topLevel),
          ),
          for (final t in _allTagPaths())
            if (!tags.any((d) => _inTree(t, d)))
              SimpleDialogOption(
                onPressed: () => Navigator.pop(context, t),
                child: Padding(
                  padding: EdgeInsets.only(
                    left: 16.0 * ('::'.allMatches(t).length),
                  ),
                  child: Text(t.split('::').last),
                ),
              ),
        ],
      ),
    );
    if (parent != null) await _moveTags(tags, parent == top ? '' : parent);
  }

  /// Every tag and parent of a tag, in tree order.
  List<String> _allTagPaths() {
    final out = <String>[];
    void walk(List<SidebarNode<String>> nodes) {
      for (final n in nodes) {
        out.add(n.fullName);
        walk(n.children);
      }
    }

    walk(sidebarTree([for (final t in _tags) (t, t)]));
    return out;
  }

  Future<void> _clearUnusedTags() async {
    var n = 0;
    if (await guard(() async => n = await golang.notes.clearUnusedTags())) {
      showInfo(n == 0 ? appL10n.noUnusedTags : appL10n.clearedUnusedTags(n));
    }
  }

  /// Moves [deck] under [parent], or to the top level when null.
  Future<void> _moveDeck(golang.DeckNode deck, golang.DeckNode? parent) async {
    final undo = AppScope.read(context).undo;
    final l = context.l10n;
    final name = parent == null
        ? deck.name
        : '${parent.fullName}::${deck.name}';
    if (await guard(() => golang.decks.rename(deck.id, name))) {
      undo.offer(
        parent == null
            ? l.movedDeckTop(deck.name)
            : l.movedDeckInto(deck.name, parent.fullName),
      );
    }
  }

  /// Whether [deck] can move under [parent] (null: the top level).
  static bool _canMoveDeck(golang.DeckNode deck, golang.DeckNode? parent) {
    if (parent == null) return deck.depth > 0;
    // Filtered decks have no subdecks.
    return !parent.filtered &&
        parent.id != deck.parentID &&
        !_inTree(parent.fullName, deck.fullName);
  }

  Future<void> _pickDeckParent(golang.DeckNode deck) async {
    final decks = AppScope.read(context).deckList.decks;
    final picked = await showDialog<golang.DeckNode?>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => SimpleDialog(
        title: Text(context.l10n.moveDeckTo(deck.name)),
        children: [
          if (_canMoveDeck(deck, null))
            SimpleDialogOption(
              onPressed: () => Navigator.pop(context, deck),
              child: Text(context.l10n.topLevel),
            ),
          for (final d in decks)
            if (_canMoveDeck(deck, d))
              SimpleDialogOption(
                onPressed: () => Navigator.pop(context, d),
                child: Padding(
                  padding: EdgeInsets.only(left: d.depth * 16.0),
                  child: Text(d.name),
                ),
              ),
        ],
      ),
    );
    if (picked == null) return;
    // The deck itself stands for the top level.
    await _moveDeck(deck, identical(picked, deck) ? null : picked);
  }

  // The item menu.

  Future<void> _menu(
    Offset position, {
    required String term,
    String? tag,
    golang.DeckNode? deck,
  }) async {
    final overlay =
        Overlay.of(context).context.findRenderObject()! as RenderBox;
    final choice = await showMenu<Object>(
      context: context,
      popUpAnimationStyle: popupAnimation(context),
      position: RelativeRect.fromRect(
        // The overlay is scaled with the UI size, the pointer is not.
        overlay.globalToLocal(position) & const Size(1, 1),
        Offset.zero & overlay.size,
      ),
      items: [
        PopupMenuItem(
          value: SidebarSearch.replace,
          child: Text(context.l10n.search),
        ),
        PopupMenuItem(
          value: SidebarSearch.and,
          child: Text(context.l10n.addToSearch),
        ),
        PopupMenuItem(
          value: SidebarSearch.or,
          child: Text(context.l10n.addToSearchOr),
        ),
        PopupMenuItem(
          value: SidebarSearch.not,
          child: Text(context.l10n.excludeFromSearchAction),
        ),
        if (tag != null || deck != null) ...[
          const PopupMenuDivider(),
          PopupMenuItem(
            value: 'rename',
            child: Text(context.l10n.renameEllipsis),
          ),
          PopupMenuItem(
            value: 'move',
            child: Text(context.l10n.moveToEllipsis),
          ),
          if (tag != null)
            PopupMenuItem(
              value: 'select',
              child: Text(context.l10n.selectTags),
            ),
          PopupMenuItem(
            value: 'delete',
            child: Text(context.l10n.deleteEllipsis),
          ),
        ],
      ],
    );
    if (!mounted || choice == null) return;
    switch (choice) {
      case final SidebarSearch how:
        widget.onSearch(term, how);
      case 'rename':
        await (tag != null ? _renameTag(tag) : renameDeck(context, deck!));
      case 'move':
        await (tag != null ? _pickTagParent([tag]) : _pickDeckParent(deck!));
      case 'select':
        setState(() {
          _selectingTags = true;
          _selectedTags[tag!.toLowerCase()] = tag;
        });
      case 'delete':
        await (tag != null ? _deleteTags([tag]) : deleteDeck(context, deck!));
    }
  }

  // Building.

  bool _matches(String label) {
    final f = _filter.text.trim().toLowerCase();
    return f.isEmpty || label.toLowerCase().contains(f);
  }

  /// Whether a tree node is shown under the filter: it matches, or one of
  /// its ancestors ([ancestorMatched]) or descendants does.
  bool _visible(SidebarNode<Object?> node, bool ancestorMatched) =>
      ancestorMatched ||
      _matches(node.name) ||
      node.children.any((c) => _visible(c, false));

  /// A section header; tapping it collapses or expands the section. Tags
  /// or decks dropped on a [dropTarget] header move to the top level.
  Widget _header(
    String key,
    String label, {
    Widget? trailing,
    bool dropTarget = false,
  }) {
    final collapsed = _collapsed.contains(key);
    final header = InkWell(
      key: Key('sidebar-section-$key'),
      onTap: () => setState(
        () => collapsed ? _collapsed.remove(key) : _collapsed.add(key),
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 12, 4, 4),
        child: Row(
          children: [
            Expanded(
              child: Text(label, style: Theme.of(context).textTheme.titleSmall),
            ),
            ?trailing,
            Icon(collapsed ? Icons.expand_more : Icons.expand_less, size: 20),
          ],
        ),
      ),
    );
    return dropTarget ? _dropTarget(header, null, null) : header;
  }

  /// Wraps [child] so tags (onto [tag]) or decks (onto [deck]) can be
  /// dropped on it; both null makes it the section header, which moves
  /// them to the top level.
  Widget _dropTarget(Widget child, String? tag, golang.DeckNode? deck) {
    return DragTarget<_Drag>(
      onWillAcceptWithDetails: (d) => switch (d.data) {
        _TagDrag(:final tags) =>
          deck == null &&
              (tag == null
                  ? tags.any((t) => t.contains('::'))
                  : !tags.any((t) => _inTree(tag, t))),
        _DeckDrag(deck: final dragged) =>
          tag == null && _canMoveDeck(dragged, deck),
      },
      onAcceptWithDetails: (d) => switch (d.data) {
        _TagDrag(:final tags) => _moveTags(tags, tag ?? ''),
        _DeckDrag(deck: final dragged) => _moveDeck(dragged, deck),
      },
      builder: (context, candidates, _) => candidates.isEmpty
          ? child
          : DecoratedBox(
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primaryContainer,
              ),
              child: child,
            ),
    );
  }

  /// One item: clicking searches (with modifiers combining), right-click or
  /// long-press opens its menu. Tags and decks can be dragged with a mouse.
  Widget _item({
    required String key,
    required String label,
    required String term,
    required IconData icon,
    int depth = 0,
    SidebarNode<Object?>? node,
    String? tag,
    golang.DeckNode? deck,
  }) {
    final filtering = _filter.text.trim().isNotEmpty;
    final hasChildren = node != null && node.children.isNotEmpty;
    final expanded = filtering || _expanded.contains(key);
    final selected =
        tag != null && _selectedTags.containsKey(tag.toLowerCase());
    final picking = tag != null && _selectingTags;
    Widget row = GestureDetector(
      onLongPressStart: (d) =>
          _menu(d.globalPosition, term: term, tag: tag, deck: deck),
      child: InkWell(
        key: Key('sidebar-$key'),
        onTap: picking
            ? () => _toggleTag(tag)
            : () => widget.onSearch(term, _clickSearch()),
        onSecondaryTapUp: (d) =>
            _menu(d.globalPosition, term: term, tag: tag, deck: deck),
        child: Container(
          height: 36,
          color: selected
              ? Theme.of(context).colorScheme.secondaryContainer
              : null,
          padding: EdgeInsets.only(left: 4.0 + depth * 16),
          child: Row(
            children: [
              SizedBox(
                width: 28,
                child: hasChildren && !filtering
                    ? IconButton(
                        key: Key('sidebar-expand-$key'),
                        tooltip: expanded
                            ? context.l10n.collapse
                            : context.l10n.expand,
                        padding: EdgeInsets.zero,
                        iconSize: 18,
                        icon: Icon(
                          expanded ? Icons.expand_more : Icons.chevron_right,
                        ),
                        onPressed: () => setState(
                          () => expanded
                              ? _expanded.remove(key)
                              : _expanded.add(key),
                        ),
                      )
                    : null,
              ),
              if (picking)
                Checkbox(
                  value: selected,
                  visualDensity: VisualDensity.compact,
                  onChanged: (_) => _toggleTag(tag),
                )
              else
                Icon(icon, size: 18),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        ),
      ),
    );
    if (tag == null && deck == null) return row;
    final drag = tag != null
        ? _TagDrag(selected ? _selectedTags.values.toList() : [tag])
        : _DeckDrag(deck!);
    final count = drag is _TagDrag ? drag.tags.length : 1;
    row = Draggable<_Drag>(
      data: drag,
      // Without a mouse, drags scroll the list; the menu's Move to… moves
      // instead.
      maxSimultaneousDrags: _mouse.mouseIsConnected ? null : 0,
      feedback: Material(
        elevation: 4,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Text(count == 1 ? label : context.l10n.tagCount(count)),
        ),
      ),
      child: row,
    );
    return _dropTarget(row, tag, deck);
  }

  /// The rows of tree [nodes] at [depth], keyed by [prefix].
  List<Widget> _tree<T>(
    List<SidebarNode<T>> nodes,
    String prefix,
    int depth,
    Widget Function(SidebarNode<T> node, String key, int depth) item, [
    bool ancestorMatched = false,
  ]) {
    final filtering = _filter.text.trim().isNotEmpty;
    return [
      for (final n in nodes)
        if (_visible(n, ancestorMatched)) ...[
          item(n, '$prefix-${n.fullName.toLowerCase()}', depth),
          if (filtering ||
              _expanded.contains('$prefix-${n.fullName.toLowerCase()}'))
            ..._tree(
              n.children,
              prefix,
              depth + 1,
              item,
              ancestorMatched || (filtering && _matches(n.name)),
            ),
        ],
    ];
  }

  List<Widget> _section(String key, Widget header, List<Widget> items) => [
    header,
    if (!_collapsed.contains(key)) ...items,
  ];

  @override
  Widget build(BuildContext context) {
    final deckList = AppScope.of(context).deckList;
    return ListenableBuilder(
      listenable: Listenable.merge([deckList, _mouse]),
      builder: (context, _) {
        final decks = sidebarTree([
          for (final d in deckList.decks) (d.fullName, d),
        ]);
        final tags = sidebarTree([for (final t in _tags) (t, t)]);
        return Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(8, 8, 8, 0),
              child: TextField(
                key: const Key('sidebar-filter'),
                controller: _filter,
                decoration: InputDecoration(
                  prefixIcon: const Icon(Icons.filter_list),
                  hintText: context.l10n.filterHint,
                  isDense: true,
                  border: const OutlineInputBorder(),
                  suffixIcon: _filter.text.isEmpty
                      ? null
                      : IconButton(
                          tooltip: context.l10n.clearFilter,
                          icon: const Icon(Icons.clear),
                          onPressed: () => setState(_filter.clear),
                        ),
                ),
                onChanged: (_) => setState(() {}),
              ),
            ),
            if (_selectingTags) _selectionBar(),
            Expanded(
              child: ListView(
                key: const Key('sidebar-list'),
                children: [
                  ..._section(
                    'states',
                    _header('states', context.l10n.cardStates),
                    [
                      for (final (label, term, icon) in _states(context.l10n))
                        if (_matches(label))
                          _item(
                            key: 'state-$term',
                            label: label,
                            term: term,
                            icon: icon,
                          ),
                    ],
                  ),
                  ..._section(
                    'decks',
                    _header('decks', context.l10n.decks, dropTarget: true),
                    _tree(decks, 'deck', 0, (n, key, depth) {
                      final deck = n.value!;
                      return _item(
                        key: key,
                        label: n.name,
                        term: sidebarTerm('deck', deck.fullName),
                        icon: Icons.folder_outlined,
                        depth: depth,
                        node: n,
                        deck: deck,
                      );
                    }),
                  ),
                  ..._section(
                    'noteTypes',
                    _header('noteTypes', context.l10n.noteTypes),
                    [
                      for (final name in _noteTypes)
                        if (_matches(name))
                          _item(
                            key: 'noteType-${name.toLowerCase()}',
                            label: name,
                            term: sidebarTerm('note', name),
                            icon: Icons.description_outlined,
                          ),
                    ],
                  ),
                  ..._section(
                    'tags',
                    _header(
                      'tags',
                      context.l10n.tags,
                      dropTarget: true,
                      trailing: PopupMenuButton<String>(
                        popUpAnimationStyle: popupAnimation(context),
                        key: const Key('sidebar-tags-menu'),
                        tooltip: context.l10n.tagActions,
                        iconSize: 18,
                        icon: const Icon(Icons.more_vert),
                        onSelected: (v) => switch (v) {
                          'select' => setState(() => _selectingTags = true),
                          _ => _clearUnusedTags(),
                        },
                        itemBuilder: (_) => [
                          PopupMenuItem(
                            value: 'select',
                            child: Text(context.l10n.selectTags),
                          ),
                          PopupMenuItem(
                            value: 'clear',
                            child: Text(context.l10n.clearUnusedTags),
                          ),
                        ],
                      ),
                    ),
                    [
                      if (_tags.isEmpty && !_collapsed.contains('tags'))
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 4, 16, 8),
                          child: Text(context.l10n.noTagsYet),
                        ),
                      ..._tree(tags, 'tag', 0, (n, key, depth) {
                        return _item(
                          key: key,
                          label: n.name,
                          term: sidebarTerm('tag', n.fullName),
                          icon: Icons.label_outline,
                          depth: depth,
                          node: n,
                          tag: n.fullName,
                        );
                      }),
                    ],
                  ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }

  /// The bar shown while tags are picked: the count, Move to, Delete, Done.
  Widget _selectionBar() {
    final picked = _selectedTags.values.toList();
    return Material(
      key: const Key('sidebar-tag-selection'),
      color: Theme.of(context).colorScheme.secondaryContainer,
      child: Row(
        children: [
          IconButton(
            tooltip: context.l10n.done,
            icon: const Icon(Icons.close),
            onPressed: _endSelectingTags,
          ),
          Expanded(child: Text(context.l10n.tagsSelected(picked.length))),
          IconButton(
            tooltip: context.l10n.moveTagsToTooltip,
            icon: const Icon(Icons.drive_file_move_outline),
            onPressed: picked.isEmpty ? null : () => _pickTagParent(picked),
          ),
          IconButton(
            tooltip: context.l10n.deleteTags,
            icon: const Icon(Icons.delete_outline),
            onPressed: picked.isEmpty ? null : () => _deleteTags(picked),
          ),
        ],
      ),
    );
  }
}
