import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';

import 'app_localizations.dart';
import 'builtin_names.dart';

export 'app_localizations.dart';

/// The English texts, which fill in for missing translations and for code
/// that runs outside the app's localisations (as in widget tests).
final english = lookupAppLocalizations(const Locale('en'));

/// The texts of the current interface language, for code without a
/// [BuildContext], such as errors shown in a snackbar. [L10nTracker] keeps
/// it up to date; widgets use `context.l10n`, which rebuilds them when the
/// language changes.
AppLocalizations appL10n = english;

extension L10nContext on BuildContext {
  /// The texts of the interface language, or [english] outside the app's
  /// localisations.
  AppLocalizations get l10n =>
      Localizations.of<AppLocalizations>(this, AppLocalizations) ?? english;
}

var _dateFormats = false;

/// Makes intl's date formats (DateFormat) work in every language: Flutter's
/// localisations load their data when the app starts, which widget tests
/// without them skip. Loading them is synchronous.
void ensureDateFormats() {
  if (_dateFormats) return;
  _dateFormats = true;
  GlobalMaterialLocalizations.delegate.load(const Locale('en'));
}

/// The locale for a stored language code such as "de" or "pt-BR"; null for
/// an empty code, which follows the device's languages.
Locale? localeOf(String code) {
  if (code.isEmpty) return null;
  final [language, ...rest] = code.split('-');
  return Locale(language, rest.isEmpty ? null : rest.first);
}

/// The interface locale for the device's [preferred] locales: the first
/// one with a translation, else English (app-shell spec: unsupported system
/// language).
Locale resolveLocale(List<Locale>? preferred, Iterable<Locale> supported) {
  for (final p in preferred ?? const <Locale>[]) {
    for (final s in supported) {
      if (s.languageCode == p.languageCode &&
          (s.countryCode == null || s.countryCode == p.countryCode)) {
        return s;
      }
    }
  }
  return const Locale('en');
}

/// Keeps [appL10n] on the language of the localisations above it, and Go's
/// built-in names (see [sendBuiltinNames]) too.
class L10nTracker extends StatefulWidget {
  const L10nTracker({super.key, required this.child});

  final Widget child;

  @override
  State<L10nTracker> createState() => _L10nTrackerState();
}

class _L10nTrackerState extends State<L10nTracker> {
  String? _sent;

  @override
  Widget build(BuildContext context) {
    final l = appL10n = context.l10n;
    if (_sent != l.localeName) {
      _sent = l.localeName;
      // Without them Go keeps the previous names, so a failure only
      // leaves new note types in the earlier language.
      unawaited(sendBuiltinNames(l).catchError((Object _) {}));
    }
    return widget.child;
  }
}
