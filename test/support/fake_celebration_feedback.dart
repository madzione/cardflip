import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/progress/celebration_feedback.dart';

/// Celebration feedback for tests: records the sounds and vibrations that
/// passed the setting and platform checks.
class FakeCelebrationFeedback extends CelebrationFeedback {
  final sounds = <CelebrationSound>[];
  final haptics = <CelebrationHaptic>[];
  int prepared = 0;
  int released = 0;

  @override
  void prepareSounds() => prepared++;

  @override
  void playSound(CelebrationSound sound) => sounds.add(sound);

  @override
  void vibrateDevice(CelebrationHaptic haptic) => haptics.add(haptic);

  @override
  void release() => released++;
}

/// Replaces [CelebrationFeedback.instance] with a fake for the current test.
FakeCelebrationFeedback useFakeCelebrationFeedback() {
  final old = CelebrationFeedback.instance;
  final fake = FakeCelebrationFeedback();
  CelebrationFeedback.instance = fake;
  addTearDown(() => CelebrationFeedback.instance = old);
  return fake;
}
