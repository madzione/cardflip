import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';

void main() {
  test('the celebration sounds are short mono WAV files', () {
    // Sounds (progress spec, add-celebration-screens design decision 6):
    // rendered from tool/sounds/celebration-sounds-C2.scs as 16-bit PCM,
    // mono, 22,050 Hz, each under 100 KB.
    for (final name in ['chime', 'streak', 'goal', 'level', 'achievement']) {
      final file = File('assets/sounds/$name.wav');
      expect(file.existsSync(), isTrue, reason: name);
      final bytes = file.readAsBytesSync();
      expect(bytes.length, lessThan(100 * 1024), reason: name);
      final header = ByteData.sublistView(bytes, 0, 44);
      String ascii(int at) => String.fromCharCodes(bytes.sublist(at, at + 4));
      expect((ascii(0), ascii(8), ascii(12)), ('RIFF', 'WAVE', 'fmt '));
      expect(header.getUint16(20, Endian.little), 1, reason: '$name: PCM');
      expect(header.getUint16(22, Endian.little), 1, reason: '$name: mono');
      expect(header.getUint32(24, Endian.little), 22050, reason: name);
      expect(header.getUint16(34, Endian.little), 16, reason: name);
    }
  });
}
