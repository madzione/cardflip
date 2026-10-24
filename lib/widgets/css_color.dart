import 'dart:ui';

const _named = <String, int>{
  'black': 0xFF000000,
  'white': 0xFFFFFFFF,
  'red': 0xFFFF0000,
  'green': 0xFF008000,
  'blue': 0xFF0000FF,
  'yellow': 0xFFFFFF00,
  'orange': 0xFFFFA500,
  'purple': 0xFF800080,
  'gray': 0xFF808080,
  'grey': 0xFF808080,
  'silver': 0xFFC0C0C0,
  'navy': 0xFF000080,
  'teal': 0xFF008080,
  'maroon': 0xFF800000,
  'olive': 0xFF808000,
  'lime': 0xFF00FF00,
  'aqua': 0xFF00FFFF,
  'cyan': 0xFF00FFFF,
  'fuchsia': 0xFFFF00FF,
  'magenta': 0xFFFF00FF,
  'brown': 0xFFA52A2A,
  'pink': 0xFFFFC0CB,
  'beige': 0xFFF5F5DC,
  'ivory': 0xFFFFFFF0,
  'lightgray': 0xFFD3D3D3,
  'lightgrey': 0xFFD3D3D3,
  'darkgray': 0xFFA9A9A9,
  'darkgrey': 0xFFA9A9A9,
  'lightblue': 0xFFADD8E6,
  'darkblue': 0xFF00008B,
  'transparent': 0x00000000,
};

/// Parses a CSS colour (#rgb, #rrggbb, #rrggbbaa, rgb(), rgba() or a common
/// name). Returns null for anything else.
Color? parseCssColor(String? value) {
  if (value == null) return null;
  final v = value.trim().toLowerCase();
  if (v.isEmpty) return null;
  final named = _named[v];
  if (named != null) return Color(named);
  if (v.startsWith('#')) {
    var hex = v.substring(1);
    if (hex.length == 3 || hex.length == 4) {
      hex = hex.split('').map((c) => '$c$c').join();
    }
    if (hex.length == 6) hex = '${hex}ff';
    if (hex.length != 8) return null;
    final n = int.tryParse(hex, radix: 16);
    if (n == null) return null;
    // #rrggbbaa → 0xaarrggbb
    return Color(((n & 0xff) << 24) | (n >> 8));
  }
  final m = RegExp(r'^rgba?\(([^)]*)\)$').firstMatch(v);
  if (m != null) {
    final parts = m
        .group(1)!
        .split(RegExp(r'[,\s/]+'))
        .where((p) => p.isNotEmpty)
        .toList();
    if (parts.length < 3) return null;
    int channel(String p) => p.endsWith('%')
        ? ((double.tryParse(p.substring(0, p.length - 1)) ?? 0) * 2.55)
              .round()
              .clamp(0, 255)
        : (double.tryParse(p) ?? 0).round().clamp(0, 255);
    var alpha = 1.0;
    if (parts.length > 3) {
      final a = parts[3];
      alpha = a.endsWith('%')
          ? (double.tryParse(a.substring(0, a.length - 1)) ?? 100) / 100
          : double.tryParse(a) ?? 1;
    }
    return Color.fromARGB(
      (alpha.clamp(0, 1) * 255).round(),
      channel(parts[0]),
      channel(parts[1]),
      channel(parts[2]),
    );
  }
  return null;
}

/// The WCAG 2 contrast ratio of [a] and [b]: 1 for the same colour, 21 for
/// black and white. Alpha is ignored.
double contrastRatio(Color a, Color b) {
  final la = a.computeLuminance();
  final lb = b.computeLuminance();
  return la > lb ? (la + 0.05) / (lb + 0.05) : (lb + 0.05) / (la + 0.05);
}
