import 'dart:async';

import 'package:flutter/material.dart';

import 'app/bootstrap.dart';
import 'app/errors.dart';
import 'app/licenses.dart';
import 'widgets/media_files.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  // Errors nobody handled are shown, never swallowed.
  FlutterError.onError = (details) {
    FlutterError.presentError(details);
    showError(details.exception);
  };
  WidgetsBinding.instance.platformDispatcher.onError = (error, stack) {
    showError(error);
    return true;
  };
  // Sound files materialised for the player in an earlier run.
  unawaited(StoredMediaFiles.clearCache().catchError((_) {}));
  registerLicenses();
  runApp(const CardflipBootstrap());
}
