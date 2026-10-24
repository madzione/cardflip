import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/profiles_model.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/dialogs.dart';
import 'package:cardflip/l10n/l10n.dart';

enum _Action { rename, delete }

/// The profiles: open, add, rename and delete them. Reached from Settings,
/// and shown on start when "Show profile list at startup" is on.
class ProfileListScreen extends StatefulWidget {
  const ProfileListScreen({
    super.key,
    required this.profiles,
    required this.onOpen,
    this.atStartup = false,
  });

  final ProfilesModel profiles;

  /// Opens profile `id` in place of the open collection.
  final Future<void> Function(String id) onOpen;

  /// Shown before any profile is open, with nothing to go back to.
  final bool atStartup;

  @override
  State<ProfileListScreen> createState() => _ProfileListScreenState();
}

class _ProfileListScreenState extends State<ProfileListScreen> {
  bool _busy = false;

  ProfilesModel get _profiles => widget.profiles;

  Future<void> _open(golang.Profile profile) async {
    if (profile.id == _profiles.list.current && !widget.atStartup) {
      Navigator.pop(context);
      return;
    }
    setState(() => _busy = true);
    await guard(() => widget.onOpen(profile.id));
    // Opening another profile replaces the app, this screen included.
    if (mounted) setState(() => _busy = false);
  }

  Future<void> _add() => askAndSave(
    context,
    title: context.l10n.addProfile,
    label: context.l10n.name,
    helper: context.l10n.addProfileHelp,
    action: context.l10n.add,
    fieldKey: const Key('profile-name'),
    save: _profiles.add,
  );

  Future<void> _rename(golang.Profile profile) => askAndSave(
    context,
    title: context.l10n.renameProfile,
    initial: profile.name,
    label: context.l10n.name,
    action: context.l10n.rename,
    fieldKey: const Key('profile-name'),
    save: (name) => _profiles.rename(profile.id, name),
  );

  Future<void> _delete(golang.Profile profile) async {
    int? cards;
    try {
      cards = await golang.app.profileCardCount(profile.id);
    } catch (_) {
      // Unreadable: the confirmation says "all its cards".
    }
    if (!mounted) return;
    final l = context.l10n;
    final ok = await confirm(
      context,
      title: l.deleteNamedTitle(profile.name),
      message: cards == null
          ? l.deleteProfileBodyUncounted
          : l.deleteProfileBody(cards),
      action: l.delete,
      destructive: true,
    );
    if (!ok) return;
    setState(() => _busy = true);
    if (await guard(() => _profiles.delete(profile.id))) {
      showInfo(appL10n.profileDeleted(profile.name));
    }
    if (mounted) setState(() => _busy = false);
  }

  @override
  Widget build(BuildContext context) {
    final l = context.l10n;
    return Scaffold(
      appBar: AppBar(
        title: Text(l.profiles),
        automaticallyImplyLeading: !widget.atStartup,
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _busy ? null : _add,
        icon: const Icon(Icons.person_add_alt),
        label: Text(l.addProfile),
      ),
      body: ListenableBuilder(
        listenable: _profiles,
        builder: (context, _) => ListView(
          key: const Key('profile-list'),
          children: [
            if (_busy) const LinearProgressIndicator(),
            for (final p in _profiles.profiles)
              _tile(context, p, p.id == _profiles.list.current),
          ],
        ),
      ),
    );
  }

  Widget _tile(BuildContext context, golang.Profile profile, bool open) {
    final l = context.l10n;
    return ListTile(
      key: Key('profile-${profile.id}'),
      leading: Icon(
        open ? Icons.account_circle : Icons.account_circle_outlined,
      ),
      title: Text(profile.name),
      subtitle: open ? Text(l.profileIsOpen) : null,
      selected: open,
      enabled: !_busy,
      onTap: () => _open(profile),
      trailing: PopupMenuButton<_Action>(
        key: Key('profile-menu-${profile.id}'),
        popUpAnimationStyle: popupAnimation(context),
        enabled: !_busy,
        onSelected: (action) => switch (action) {
          _Action.rename => _rename(profile),
          _Action.delete => _delete(profile),
        },
        itemBuilder: (_) => [
          PopupMenuItem(value: _Action.rename, child: Text(l.renameProfile)),
          // The open profile cannot be deleted, nor the only one.
          if (!open && _profiles.several)
            PopupMenuItem(value: _Action.delete, child: Text(l.deleteProfile)),
        ],
      ),
    );
  }
}
