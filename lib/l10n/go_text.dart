import 'dart:convert';

import 'go_messages.dart';
import 'l10n.dart';

/// A tagged Go text (see golang/internal/usermsg): the code and arguments as
/// JSON, then the English wording.
final _tagged = RegExp('\x02([^\x03]*)\x03([^\x04]*)\x04');

/// The code of the first tagged text in [s], or null if it has none.
String? goCode(String s) {
  final m = _tagged.firstMatch(s);
  if (m == null) return null;
  try {
    return (jsonDecode(m[1]!) as Map<String, Object?>)['code'] as String?;
  } on FormatException {
    return null;
  }
}

/// The text [s] from Go in the interface language: each tagged text in it is
/// replaced with the translation of its code, made with its arguments (which
/// are translated the same way), or with its English wording when there is
/// no message for the code. Text around tagged texts is kept as it is.
String goText(String s, [AppLocalizations? l10n]) {
  if (!s.contains('\x02')) return s;
  final l = l10n ?? appL10n;
  return s.replaceAllMapped(_tagged, (m) {
    try {
      final tag = jsonDecode(m[1]!) as Map<String, Object?>;
      final message = goMessages[tag['code']];
      if (message != null) {
        final args = tag['args'] as List<Object?>? ?? const [];
        return message(l, [for (final a in args) goText(a as String, l)]);
      }
    } on FormatException {
      // A malformed tag or argument: the English wording.
    } on RangeError {
      // Fewer arguments than the message has placeholders.
    }
    return m[2]!;
  });
}
