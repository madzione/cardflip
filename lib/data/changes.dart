import 'dart:async';

import 'package:cardflip/golang/golang.dart' as golang;

/// Delivers collection changes from Go to every screen. It holds the one
/// subscription to `app.changes()` and re-broadcasts it. Besides changes of
/// the collection it carries those of the render environment (kind
/// [environment]), after which shown cards are rendered again.
class ChangeHub {
  ChangeHub() {
    _subscription = golang.app.changes().listen(
      _controller.add,
      onError: (Object _) {},
    );
  }

  final _controller = StreamController<golang.Change>.broadcast();
  StreamSubscription<golang.Change>? _subscription;

  /// The kind of change sent when the render environment changed.
  static const environment = 'environment';

  /// The kind of the changes received by device sync.
  static const synced = 'sync';

  Stream<golang.Change> get stream => _controller.stream;

  /// The changes of the collection, without those of the render
  /// environment.
  Stream<golang.Change> get collection =>
      stream.where((c) => c.kind != environment);

  void dispose() {
    _subscription?.cancel();
    _controller.close();
  }
}
