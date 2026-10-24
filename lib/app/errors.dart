import 'package:flutter/material.dart';

import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/go_text.dart';
import 'package:cardflip/l10n/l10n.dart';

/// The messenger of the whole app, so errors can be shown from anywhere.
final messengerKey = GlobalKey<ScaffoldMessengerState>();

/// A readable message for an error from Go or Dart, in the interface
/// language.
String errorMessage(Object error) {
  if (error is golang.GolangException) {
    final m = goText(error.message);
    // Validation errors from Go are already sentences.
    return m.isEmpty
        ? appL10n.errorUnknown
        : m[0].toUpperCase() + m.substring(1);
  }
  return appL10n.errorUnexpected('$error');
}

/// Shows an error in a snackbar.
void showError(Object error) {
  messengerKey.currentState
    ?..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text(errorMessage(error)),
        behavior: SnackBarBehavior.floating,
        showCloseIcon: true,
        duration: const Duration(seconds: 6),
      ),
    );
}

/// Shows a short confirmation in a snackbar.
void showInfo(String message) {
  messengerKey.currentState
    ?..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(content: Text(message), behavior: SnackBarBehavior.floating),
    );
}

/// Shows the result of a change with an Undo action, which calls [undo].
void showUndoable(String message, VoidCallback undo) {
  messengerKey.currentState
    ?..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text(message),
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 8),
        action: SnackBarAction(label: appL10n.undo, onPressed: undo),
      ),
    );
}

/// Runs [action] and shows its error, if any. Returns whether it succeeded.
Future<bool> guard(Future<void> Function() action) async {
  try {
    await action();
    return true;
  } catch (e) {
    showError(e);
    return false;
  }
}
