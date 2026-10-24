import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

/// The menu of saved searches next to the browser's search field: choosing
/// one calls [onRun] with its query. It also saves the current search
/// ([query]) and opens the dialog that renames and deletes them.
class SavedSearchesButton extends StatefulWidget {
  const SavedSearchesButton({
    super.key,
    required this.query,
    required this.onRun,
  });

  final String Function() query;
  final ValueChanged<String> onRun;

  @override
  State<SavedSearchesButton> createState() => _SavedSearchesButtonState();
}

/// A menu choice: run a saved search, save the current one, or manage them.
typedef _Choice = ({String action, golang.SavedSearch? search});

class _SavedSearchesButtonState extends State<SavedSearchesButton> {
  List<golang.SavedSearch> _saved = const [];
  SettingsModel? _settings;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_settings == null) {
      // Saved searches are settings: a restored backup brings others back.
      _settings = AppScope.of(context).settings..addListener(_load);
      _load();
    }
  }

  @override
  void dispose() {
    _settings?.removeListener(_load);
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final saved = await golang.browse.savedSearches();
      if (mounted) setState(() => _saved = saved);
    } catch (e) {
      showError(e);
    }
  }

  Future<void> _onSelected(_Choice choice) async {
    switch (choice.action) {
      case 'run':
        widget.onRun(choice.search!.query);
      case 'save':
        final query = widget.query();
        final saved = await askAndSave(
          context,
          title: context.l10n.saveSearch,
          message: query,
          label: context.l10n.name,
          fieldKey: const Key('saved-search-name'),
          save: (name) => golang.browse.saveSearch(name, query),
        );
        if (saved != null) {
          showInfo(appL10n.savedSearchAs(saved));
          await _load();
        }
      case 'manage':
        await showDialog<void>(
          context: context,
          animationStyle: popupAnimation(context),
          builder: (_) => _ManageDialog(saved: _saved),
        );
        await _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<_Choice>(
      popUpAnimationStyle: popupAnimation(context),
      key: const Key('saved-searches'),
      tooltip: context.l10n.savedSearches,
      icon: const Icon(Icons.bookmark_border),
      onSelected: _onSelected,
      itemBuilder: (_) => [
        for (final s in _saved)
          PopupMenuItem(
            key: Key('saved-search-${s.name}'),
            value: (action: 'run', search: s),
            child: ListTile(
              contentPadding: EdgeInsets.zero,
              title: Text(s.name),
              subtitle: Text(
                s.query,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ),
        if (_saved.isNotEmpty) const PopupMenuDivider(),
        PopupMenuItem(
          key: const Key('save-search'),
          value: (action: 'save', search: null),
          enabled: widget.query().trim().isNotEmpty,
          child: Text(context.l10n.saveCurrentSearch),
        ),
        if (_saved.isNotEmpty)
          PopupMenuItem(
            key: const Key('manage-saved-searches'),
            value: (action: 'manage', search: null),
            child: Text(context.l10n.manageSavedSearches),
          ),
      ],
    );
  }
}

/// Lists the saved searches with buttons to rename and delete them.
class _ManageDialog extends StatefulWidget {
  const _ManageDialog({required this.saved});

  final List<golang.SavedSearch> saved;

  @override
  State<_ManageDialog> createState() => _ManageDialogState();
}

class _ManageDialogState extends State<_ManageDialog> {
  late var _saved = widget.saved;

  Future<void> _reload() async {
    try {
      final saved = await golang.browse.savedSearches();
      if (mounted) setState(() => _saved = saved);
    } catch (e) {
      showError(e);
    }
  }

  Future<void> _rename(golang.SavedSearch s) async {
    final renamed = await askAndSave(
      context,
      title: context.l10n.renameSavedSearch,
      message: s.query,
      label: context.l10n.name,
      initial: s.name,
      fieldKey: const Key('saved-search-name'),
      save: (name) => golang.browse.renameSavedSearch(s.name, name),
    );
    if (renamed != null) await _reload();
  }

  Future<void> _delete(golang.SavedSearch s) async {
    final ok = await confirm(
      context,
      title: context.l10n.deleteQuoted(s.name),
      message: context.l10n.deleteSavedSearchBody(s.query),
      action: context.l10n.delete,
      destructive: true,
    );
    if (ok && await guard(() => golang.browse.deleteSavedSearch(s.name))) {
      await _reload();
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(context.l10n.savedSearches),
      contentPadding: const EdgeInsets.symmetric(vertical: 16),
      content: SizedBox(
        width: 400,
        child: _saved.isEmpty
            ? Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: Text(context.l10n.noSavedSearches),
              )
            : ListView(
                shrinkWrap: true,
                children: [
                  for (final s in _saved)
                    ListTile(
                      key: Key('manage-${s.name}'),
                      title: Text(s.name),
                      subtitle: Text(
                        s.query,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      trailing: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            key: Key('rename-${s.name}'),
                            tooltip: context.l10n.rename,
                            icon: const Icon(Icons.edit_outlined),
                            onPressed: () => _rename(s),
                          ),
                          IconButton(
                            key: Key('delete-${s.name}'),
                            tooltip: context.l10n.delete,
                            icon: const Icon(Icons.delete_outline),
                            onPressed: () => _delete(s),
                          ),
                        ],
                      ),
                    ),
                ],
              ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.done),
        ),
      ],
    );
  }
}
