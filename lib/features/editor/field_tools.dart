import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/errors.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/voice_recorder.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'math_insert.dart';
import 'record_audio.dart';
import 'rich_field.dart';

/// The buttons that change a field's content (Cloze deletion, Math, Attach
/// and Record audio), pasting and dropping media, for the note editor and
/// the browser's edit pane.
mixin FieldTools<T extends StatefulWidget> on State<T> {
  /// The fields the tools act on, and their focus nodes.
  List<RichFieldController> get toolFields;
  List<FocusNode> get toolFocus;

  /// The focused field, or the last focused one.
  int get lastFocusedField;

  /// Whether the note type is a cloze type, which adds Cloze deletion.
  bool get clozeNoteType;

  /// Shows why media could not be added; the input stays.
  void showToolError(String message);

  int get _target => lastFocusedField.clamp(0, toolFields.length - 1);

  /// Makes images pasted into the fields go to their media. Call it
  /// whenever [toolFields] are made.
  void enableImagePaste() {
    for (final (i, c) in toolFields.indexed) {
      c.onPasteImage = (image) => _pasteImage(i, image);
    }
  }

  /// Cloze deletion (cloze note types only), Math, Attach and Record
  /// audio, for the end of the field toolbar.
  List<Widget> toolButtons() => [
    if (clozeNoteType)
      PopupMenuButton<bool>(
        popUpAnimationStyle: popupAnimation(context),
        key: const Key('insert-cloze'),
        tooltip: appL10n.clozeDeletion,
        icon: const Icon(Icons.data_array),
        enabled: toolFields.isNotEmpty,
        onSelected: (same) => insertCloze(sameNumber: same),
        itemBuilder: (context) => [
          PopupMenuItem(
            value: false,
            child: MenuEntryLabel(appL10n.nextClozeNumber, clozeShortcutLabel),
          ),
          PopupMenuItem(
            value: true,
            child: MenuEntryLabel(
              appL10n.sameClozeNumber,
              clozeSameShortcutLabel,
            ),
          ),
        ],
      ),
    MathMenuButton(onSelected: insertMath),
    IconButton(
      key: const Key('attach-media'),
      tooltip: appL10n.attachMedia,
      icon: const Icon(Icons.attach_file),
      onPressed: toolFields.isEmpty ? null : _attach,
    ),
    IconButton(
      key: const Key('record-audio'),
      tooltip: appL10n.recordAudio,
      icon: const Icon(Icons.mic_none),
      onPressed: toolFields.isEmpty ? null : _record,
    ),
  ];

  /// Ctrl+Shift+C and Ctrl+Alt+Shift+C (Cmd on macOS), in cloze note types.
  Map<ShortcutActivator, VoidCallback> get clozeShortcuts =>
      clozeNoteType && toolFields.isNotEmpty
      ? {
          clozeShortcut: insertCloze,
          clozeSameShortcut: () => insertCloze(sameNumber: true),
        }
      : const {};

  /// Wraps the selection of the focused field in the next unused cloze
  /// number of the note, or with [sameNumber] in the highest one already
  /// used (1 when there is none).
  void insertCloze({bool sameNumber = false}) {
    final highest = _highestCloze();
    final n = sameNumber ? max(highest, 1) : highest + 1;
    _wrap('{{c$n::', '}}');
  }

  /// `{{cN::` or `{{cN,M,…::`, as the Go cloze parser reads them.
  static final _clozeOpener = RegExp(r'\{\{[cC](\d+(?:, *\d+)*)::');

  /// The highest cloze number in any field of the note, 0 when none.
  int _highestCloze() {
    var highest = 0;
    for (final f in toolFields) {
      for (final m in _clozeOpener.allMatches(f.html)) {
        final numbers = [
          for (final n in m.group(1)!.split(',')) int.tryParse(n.trim()),
        ];
        // Not an opener for Go either: c0, or a number too large.
        if (numbers.any((n) => n == null || n < 1)) continue;
        highest = max(highest, numbers.map((n) => n!).reduce(max));
      }
    }
    return highest;
  }

  /// Wraps the selection of the focused field in [m]'s delimiters.
  void insertMath(MathInsert m) => _wrap(m.open, m.close);

  /// Wraps the selection of the focused field in [open] and [close].
  void _wrap(String open, String close) {
    if (toolFields.isEmpty) return;
    final index = _target;
    toolFields[index].wrap(open, close);
    focusAfterFrame(toolFocus[index]);
  }

  /// Lets the user pick images, audio and video and inserts them into the
  /// focused field.
  Future<void> _attach() async {
    if (toolFields.isEmpty) return;
    final index = _target;
    final List<PickedFile> files;
    try {
      files = await AppScope.read(context).picker.pickMedia();
    } catch (e) {
      if (mounted) showToolError(errorMessage(e));
      return;
    }
    await _addMedia(index, files);
  }

  /// Records from the microphone and inserts the recording into the
  /// focused field. Refused microphone access is explained.
  Future<void> _record() async {
    if (toolFields.isEmpty) return;
    final index = _target;
    final RecordedAudio audio;
    try {
      final recorded = await recordAudio(context);
      if (recorded == null) return;
      audio = recorded;
    } on MicrophoneRefused catch (e) {
      if (mounted) showToolError(e.toString());
      return;
    } catch (e) {
      if (mounted) showToolError(errorMessage(e));
      return;
    }
    await _store(index, [() => golang.media.add(audio.extension, audio.bytes)]);
  }

  /// Stores an image pasted into field [index]: as JPEG, or PNG when the
  /// setting says so.
  Future<void> _pasteImage(int index, Uint8List image) {
    final png = AppScope.read(context).settings.pasteImagesAsPng;
    return _store(index, [() => golang.media.addPastedImage(image, png)]);
  }

  /// Adds the images, audio and video among files dropped onto field
  /// [index]; other files are left out with a note.
  void dropMedia(int index, List<PickedFile> files) {
    final media = [
      for (final f in files)
        if (mediaExtensions.contains(_extension(f.name))) f,
    ];
    if (media.length < files.length) {
      showInfo(appL10n.onlyMediaInFields);
    }
    _addMedia(index, media);
  }

  static String _extension(String name) {
    final dot = name.lastIndexOf('.');
    return dot < 0 ? '' : name.substring(dot + 1).toLowerCase();
  }

  /// Stores [files] in the collection's media and inserts them at the
  /// cursor of field [index].
  Future<void> _addMedia(int index, List<PickedFile> files) => _store(index, [
    for (final f in files) () => golang.media.add(f.name, f.bytes),
  ]);

  /// Runs [stores], each storing one media file and returning its name,
  /// and inserts the files at the cursor of field [index].
  Future<void> _store(int index, List<Future<String> Function()> stores) async {
    if (stores.isEmpty || index >= toolFields.length) return;
    final field = toolFields[index];
    for (final store in stores) {
      try {
        final name = await store();
        // The fields were replaced meanwhile (another note type).
        if (!mounted ||
            index >= toolFields.length ||
            toolFields[index] != field) {
          return;
        }
        field.insertMedia(name);
      } catch (e) {
        if (mounted) showToolError(errorMessage(e));
        return;
      }
    }
    if (mounted) focusAfterFrame(toolFocus[index]);
  }
}
