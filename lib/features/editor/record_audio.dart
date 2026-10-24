import 'dart:async';

import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/widgets/voice_recorder.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Records from the microphone in a dialog until the user inserts or
/// cancels the recording. Returns the recording, or null when cancelled.
/// Fails with [MicrophoneRefused] without microphone access.
Future<RecordedAudio?> recordAudio(BuildContext context) async {
  final recorder = VoiceRecorder.create();
  String? uri;
  try {
    await recorder.start();
    if (!context.mounted) {
      uri = await recorder.stop();
      return null;
    }
    final insert = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      barrierDismissible: false,
      builder: (_) => const _RecordingDialog(),
    );
    uri = await recorder.stop();
    if (insert != true || uri == null) return null;
    return await recorder.read(uri);
  } finally {
    if (uri != null) unawaited(recorder.discard(uri));
    recorder.dispose();
  }
}

class _RecordingDialog extends StatefulWidget {
  const _RecordingDialog();

  @override
  State<_RecordingDialog> createState() => _RecordingDialogState();
}

class _RecordingDialogState extends State<_RecordingDialog> {
  late final Timer _tick;
  int _seconds = 0;

  @override
  void initState() {
    super.initState();
    _tick = Timer.periodic(
      const Duration(seconds: 1),
      (_) => setState(() => _seconds++),
    );
  }

  @override
  void dispose() {
    _tick.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final s = _seconds;
    final elapsed = '${s ~/ 60}:${(s % 60).toString().padLeft(2, '0')}';
    return AlertDialog(
      title: Text(context.l10n.recording),
      content: Row(
        children: [
          Icon(Icons.mic, color: Theme.of(context).colorScheme.error),
          const SizedBox(width: 12),
          Text(elapsed, key: const Key('record-elapsed')),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: Text(context.l10n.cancel),
        ),
        FilledButton(
          key: const Key('record-insert'),
          onPressed: () => Navigator.pop(context, true),
          child: Text(context.l10n.insert),
        ),
      ],
    );
  }
}
