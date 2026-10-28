import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:media_kit/media_kit.dart';

import 'package:cardflip/data/settings_model.dart';

/// The sounds of the celebrations, in assets/sounds/ (rendered from the
/// SoundCraft session tool/sounds/celebration-sounds-C2.scs).
enum CelebrationSound { chime, streak, goal, level, achievement }

/// The vibrations of the celebrations: [medium] for a celebration page,
/// [light] for a reward banner.
enum CelebrationHaptic { medium, light }

/// Plays the celebration sounds and vibrations when *Celebration sounds and
/// vibration* is on. Tests replace [instance].
abstract class CelebrationFeedback {
  static CelebrationFeedback instance = MediaKitCelebrationFeedback();

  /// Whether this device vibrates: Android and iOS, not in a browser.
  static bool vibrates({bool web = kIsWeb}) =>
      !web &&
      (defaultTargetPlatform == TargetPlatform.android ||
          defaultTargetPlatform == TargetPlatform.iOS);

  /// Gets ready to play without delay, when [settings] have sounds on: at
  /// the start of a session.
  void prepare(SettingsModel settings) {
    if (_on(settings)) prepareSounds();
  }

  /// Plays [sound] when [settings] have sounds on.
  void play(SettingsModel settings, CelebrationSound sound) {
    if (_on(settings)) playSound(sound);
  }

  /// Vibrates with [haptic] when [settings] have sounds on and the device
  /// [vibrates].
  void vibrate(SettingsModel settings, CelebrationHaptic haptic) {
    if (_on(settings) && vibrates()) vibrateDevice(haptic);
  }

  /// Releases what [prepare] took, when the study screen closes.
  void release();

  bool _on(SettingsModel settings) =>
      settings.showProgress && settings.celebrationSounds;

  @protected
  void prepareSounds();

  @protected
  void playSound(CelebrationSound sound);

  @protected
  void vibrateDevice(CelebrationHaptic haptic);
}

/// Plays the sounds with a media_kit player of its own, apart from card
/// audio (MediaPlayer), which it never queues behind, stops or pauses.
class MediaKitCelebrationFeedback extends CelebrationFeedback {
  static const _volume = 70.0;

  Player? _player;

  static Media _media(CelebrationSound sound) =>
      Media('asset:///assets/sounds/${sound.name}.wav');

  Player _create() {
    MediaKit.ensureInitialized();
    final p = Player();
    // Load the chime without playing it, so the first sound has no
    // decoder warm-up.
    _ignore(p.open(_media(CelebrationSound.chime), play: false));
    return p;
  }

  @override
  void prepareSounds() => _player ??= _create();

  @override
  void playSound(CelebrationSound sound) {
    final p = _player ??= _create();
    _ignore(() async {
      await p.setVolume(_volume);
      await p.open(_media(sound));
    }());
  }

  @override
  void vibrateDevice(CelebrationHaptic haptic) => _ignore(switch (haptic) {
    CelebrationHaptic.medium => HapticFeedback.mediumImpact(),
    CelebrationHaptic.light => HapticFeedback.lightImpact(),
  });

  @override
  void release() {
    final p = _player;
    _player = null;
    if (p != null) _ignore(p.dispose());
  }

  /// A sound or vibration that fails is left out: it is decoration.
  static void _ignore(Future<void> f) => unawaited(f.catchError((_) {}));
}
