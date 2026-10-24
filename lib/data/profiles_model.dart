import 'package:flutter/foundation.dart';

import 'package:cardflip/golang/golang.dart' as golang;

/// The profiles, each a separate collection. They are kept outside the
/// collection, so changes to them send no collection change: the model
/// reloads after each of its own.
class ProfilesModel extends ChangeNotifier {
  ProfilesModel([
    this.list = const golang.ProfileList(
      profiles: [],
      current: '',
      showAtStartup: false,
    ),
  ]);

  golang.ProfileList list;

  List<golang.Profile> get profiles => list.profiles;

  /// The open profile's name; null before one is open.
  String? get currentName {
    for (final p in list.profiles) {
      if (p.id == list.current) return p.name;
    }
    return null;
  }

  /// Whether there is more than one profile, so the open one is worth
  /// naming.
  bool get several => list.profiles.length > 1;

  Future<void> load() async {
    list = await golang.app.profiles();
    notifyListeners();
  }

  Future<void> add(String name) async {
    await golang.app.addProfile(name);
    await load();
  }

  Future<void> rename(String id, String name) async {
    await golang.app.renameProfile(id, name);
    await load();
  }

  Future<void> delete(String id) async {
    await golang.app.deleteProfile(id);
    await load();
  }

  Future<void> setShowAtStartup(bool on) async {
    await golang.app.setShowProfilesAtStartup(on);
    await load();
  }
}
