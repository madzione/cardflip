import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/changes.dart';
import 'package:cardflip/data/settings_model.dart';
import 'package:cardflip/features/progress/celebration_feedback.dart';

import 'support/fake_celebration_feedback.dart';
import 'support/fake_go.dart';

void main() {
  /// The settings, loaded against the fake with [sounds] for *Celebration
  /// sounds and vibration*.
  Future<SettingsModel> loaded({bool sounds = true}) async {
    final go = FakeGo();
    go.settings = {...go.settings, 'celebrationSounds': sounds};
    final hub = ChangeHub();
    final settings = SettingsModel(hub);
    addTearDown(() {
      settings.dispose();
      hub.dispose();
    });
    await settings.load();
    return settings;
  }

  testWidgets('sounds and vibration follow the setting', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.android;
    final fake = useFakeCelebrationFeedback();
    final on = await loaded();
    fake
      ..prepare(on)
      ..play(on, CelebrationSound.level)
      ..vibrate(on, CelebrationHaptic.medium);
    expect(fake.prepared, 1);
    expect(fake.sounds, [CelebrationSound.level]);
    expect(fake.haptics, [CelebrationHaptic.medium]);

    // Off (spec): nothing is played or vibrated.
    final off = await loaded(sounds: false);
    fake
      ..prepare(off)
      ..play(off, CelebrationSound.chime)
      ..vibrate(off, CelebrationHaptic.light);
    expect(fake.prepared, 1);
    expect(fake.sounds, [CelebrationSound.level]);
    expect(fake.haptics, [CelebrationHaptic.medium]);

    // Nor with Show streaks and goals off.
    await on.setShowProgress(false);
    fake.play(on, CelebrationSound.goal);
    expect(fake.sounds, [CelebrationSound.level]);
    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets('only phones vibrate', (tester) async {
    final fake = useFakeCelebrationFeedback();
    final settings = await loaded();

    // Desktop (spec): the sound plays and nothing vibrates.
    debugDefaultTargetPlatformOverride = TargetPlatform.linux;
    fake
      ..play(settings, CelebrationSound.level)
      ..vibrate(settings, CelebrationHaptic.medium);
    expect(fake.sounds, [CelebrationSound.level]);
    expect(fake.haptics, isEmpty);
    expect(CelebrationFeedback.vibrates(), isFalse);

    debugDefaultTargetPlatformOverride = TargetPlatform.iOS;
    expect(CelebrationFeedback.vibrates(), isTrue);
    // A phone's browser does not vibrate.
    debugDefaultTargetPlatformOverride = TargetPlatform.android;
    expect(CelebrationFeedback.vibrates(), isTrue);
    expect(CelebrationFeedback.vibrates(web: true), isFalse);
    debugDefaultTargetPlatformOverride = null;
  });
}
