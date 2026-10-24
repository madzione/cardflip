import 'package:desktop_drop/desktop_drop.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'package:cardflip/features/import/pick.dart';

/// Whether files can be dropped onto the app: on desktop and the web.
bool get canDropFiles =>
    kIsWeb ||
    switch (defaultTargetPlatform) {
      TargetPlatform.linux ||
      TargetPlatform.macOS ||
      TargetPlatform.windows => true,
      _ => false,
    };

/// Accepts files dropped onto a field (desktop and web) and outlines the
/// field while files are dragged over it.
class FieldDropTarget extends StatefulWidget {
  const FieldDropTarget({super.key, required this.onDrop, required this.child});

  /// Called with the dropped files (folders are left out).
  final ValueChanged<List<PickedFile>> onDrop;
  final Widget child;

  @override
  State<FieldDropTarget> createState() => _FieldDropTargetState();
}

class _FieldDropTargetState extends State<FieldDropTarget> {
  bool _over = false;

  Future<void> _done(DropDoneDetails details) async {
    setState(() => _over = false);
    final files = <PickedFile>[
      for (final f in details.files)
        if (f is! DropItemDirectory)
          (name: f.name, bytes: await f.readAsBytes()),
    ];
    if (files.isNotEmpty) widget.onDrop(files);
  }

  @override
  Widget build(BuildContext context) {
    if (!canDropFiles) return widget.child;
    return DropTarget(
      onDragEntered: (_) => setState(() => _over = true),
      onDragExited: (_) => setState(() => _over = false),
      onDragDone: _done,
      child: DecoratedBox(
        position: DecorationPosition.foreground,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(4),
          border: _over
              ? Border.all(
                  color: Theme.of(context).colorScheme.primary,
                  width: 2,
                )
              : null,
        ),
        child: widget.child,
      ),
    );
  }
}
