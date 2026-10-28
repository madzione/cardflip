import 'dart:async';

import 'package:cardflip/features/progress/celebration_feedback.dart';

import 'support/fake_celebration_feedback.dart';

/// No test plays celebration sounds through media_kit: they all start with
/// a fake, which a test can replace with useFakeCelebrationFeedback to
/// check what was played.
Future<void> testExecutable(FutureOr<void> Function() testMain) async {
  CelebrationFeedback.instance = FakeCelebrationFeedback();
  await testMain();
}
