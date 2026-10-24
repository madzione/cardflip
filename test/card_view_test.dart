import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_math_fork/flutter_math.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_widget_from_html_core/flutter_widget_from_html_core.dart';

import 'package:cardflip/app/theme.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/card_view.dart';
import 'package:cardflip/widgets/css_color.dart';
import 'package:cardflip/widgets/font_cache.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/widgets/media_player.dart';
import 'package:cardflip/widgets/occlusion_view.dart';
import 'package:cardflip/widgets/speech.dart';

import 'support/fake_go.dart';
import 'support/fake_media.dart';
import 'support/fake_speech.dart';

// A 1×1 PNG.
final pixel = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
);

/// A grey PNG of [width] × [height] pixels.
Future<Uint8List> _png(int width, int height) async {
  final recorder = ui.PictureRecorder();
  Canvas(recorder).drawRect(
    Rect.fromLTWH(0, 0, width.toDouble(), height.toDouble()),
    Paint()..color = const Color(0xFF888888),
  );
  final image = await recorder.endRecording().toImage(width, height);
  final data = await image.toByteData(format: ui.ImageByteFormat.png);
  return data!.buffer.asUint8List();
}

Future<void> pumpCard(
  WidgetTester tester,
  String html, {
  ThemeData? theme,
  String? background,
  String? color,
  MediaCache? media,
}) async {
  await tester.pumpWidget(
    MaterialApp(
      theme: theme ?? lightTheme(),
      home: Scaffold(
        body: CardView(
          html: html,
          background: background,
          color: color,
          media: media,
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

/// Pumps a card side with sounds, played by a player on fake media.
Future<MediaPlayer> pumpSounds(
  WidgetTester tester,
  String html,
  List<MediaItem> sounds, {
  FakeBackend? backend,
  Iterable<String> available = const [],
  bool showPlayButtons = true,
  ThemeData? theme,
}) async {
  final player = MediaPlayer(
    backend: () => backend ?? FakeBackend(),
    files: FakeMediaFiles(available),
  );
  addTearDown(player.dispose);
  await tester.pumpWidget(
    MaterialApp(
      theme: theme,
      home: Scaffold(
        body: CardView(
          html: html,
          sounds: sounds,
          player: player,
          showPlayButtons: showPlayButtons,
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return player;
}

const word = MediaItem(CardSide.question, 0, 'word.mp3');

Color surfaceOf(WidgetTester tester) =>
    tester.widget<Container>(find.byKey(const Key('card-surface'))).color!;

/// The style of the text span reading [text].
TextStyle styleOf(WidgetTester tester, String text) {
  TextStyle? found;
  for (final p in tester.renderObjectList<RenderParagraph>(
    find.byType(RichText),
  )) {
    p.text.visitChildren((span) {
      if (span is TextSpan && span.text == text) found = span.style;
      return found == null;
    });
  }
  return found!;
}

void main() {
  testWidgets('renders card HTML with inlined styles', (tester) async {
    await pumpCard(
      tester,
      '<div class="card" style="text-align:center"><b style="color:#aa0000">hola</b><br><ul><li>uno</li></ul></div>',
    );
    expect(find.textContaining('hola', findRichText: true), findsOneWidget);
    expect(find.textContaining('uno', findRichText: true), findsOneWidget);
  });

  group('right-to-left text (spec)', () {
    /// The paragraph showing [text] and where it ends up on screen.
    (RenderParagraph, Rect) paragraph(WidgetTester tester, String text) {
      final finder = find.textContaining(text, findRichText: true).last;
      final p = tester.renderObject<RenderParagraph>(finder);
      final box = tester.getRect(finder);
      final width = p.getMaxIntrinsicWidth(double.infinity);
      // The line's text, wherever the paragraph puts it.
      final start = p.localToGlobal(
        p.getOffsetForCaret(const TextPosition(offset: 0), Rect.zero),
      );
      return (p, Rect.fromLTWH(start.dx - width, box.top, width, box.height));
    }

    testWidgets('an RTL template lays out and aligns right', (tester) async {
      await pumpCard(
        tester,
        '<div class="card"><div dir="rtl">كتاب جديد</div><div>new book</div></div>',
      );
      final (rtl, rtlRect) = paragraph(tester, 'كتاب');
      expect(rtl.textDirection, TextDirection.rtl);
      final card = tester.getRect(find.byType(CardView));
      // The first character is at the right edge.
      expect(rtlRect.right, greaterThan(card.center.dx));
      final (ltr, _) = paragraph(tester, 'new book');
      expect(ltr.textDirection, TextDirection.ltr);
      expect(
        tester
            .getTopLeft(find.textContaining('new book', findRichText: true))
            .dx,
        lessThan(card.center.dx),
      );
    });

    testWidgets('dir in a field and an explicit alignment', (tester) async {
      await pumpCard(
        tester,
        '<div class="card" style="text-align:center">'
        '<div dir="rtl">مرحبا</div>'
        '<p dir="RTL" style="text-align:left">שלום</p>'
        'see <span dir="rtl">כתב</span></div>',
      );
      // An explicit alignment wins over the direction's default; dir
      // ignores case, as in HTML.
      final (centered, _) = paragraph(tester, 'مرحبا');
      expect(centered.textDirection, TextDirection.rtl);
      expect(centered.textAlign, TextAlign.center);
      final (left, _) = paragraph(tester, 'שלום');
      expect(left.textDirection, TextDirection.rtl);
      expect(left.textAlign, TextAlign.left);
      // As in browsers, an inline dir orders its own text only; the line
      // keeps its direction and alignment.
      final (line, _) = paragraph(tester, 'כתב');
      expect(line.textDirection, TextDirection.ltr);
      expect(line.textAlign, TextAlign.center);
    });
  });

  group('underlines and strike-throughs take the colour of their element', () {
    const red = Color(0xFFFF0000);
    const blue = Color(0xFF0000FF);

    testWidgets('coloured underlined text, as the editor writes it', (
      tester,
    ) async {
      await pumpCard(
        tester,
        '<span style="color: #ff0000"><u>under</u> <s>struck</s></span>',
      );
      for (final text in ['under', 'struck']) {
        expect(styleOf(tester, text).decorationColor, red, reason: text);
      }
    });

    testWidgets('a colour inside the underline keeps the outer one', (
      tester,
    ) async {
      await pumpCard(
        tester,
        '<u style="color: blue">a<span style="color: red">b</span></u>',
        color: '#0000ff',
      );
      expect(styleOf(tester, 'b').color, red);
      expect(styleOf(tester, 'b').decorationColor, blue);
    });

    testWidgets('without a colour, the card colour', (tester) async {
      await pumpCard(tester, '<u>plain</u>', color: '#0000ff');
      expect(styleOf(tester, 'plain').decorationColor, blue);
    });

    testWidgets('a text-decoration style and the note type colour', (
      tester,
    ) async {
      await pumpCard(
        tester,
        '<div style="color: red"><span style="text-decoration: underline">'
        'css</span></div>',
      );
      expect(styleOf(tester, 'css').decorationColor, red);
    });
  });

  testWidgets('shows an SVG image from the collection media', (tester) async {
    // A LaTeX image as Anki names it, set to SVG.
    const name = 'latex-3e4efedf3b1b80c4b04fd17126b5f3df859565c1.svg';
    final cache = MediaCache(
      load: (file) async => switch (file) {
        name => Uint8List.fromList(
          utf8.encode(
            '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20">'
            '<rect width="40" height="20" fill="#3060c0"/></svg>',
          ),
        ),
        _ => Uint8List.fromList(
          utf8.encode('<svg xmlns="http://www.w3.org/2000/svg"><rect'),
        ),
      },
    );
    await pumpCard(
      tester,
      '<div class="card"><img class="latex" alt="\$E = mc^2\$" src="$name"></div>',
      media: cache,
    );
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pumpAndSettle();
    expect(find.byType(SvgPicture), findsOneWidget);
    expect(find.byType(Image), findsNothing);
    expect(find.byKey(const Key('missing-image')), findsNothing);
    expect(find.bySemanticsLabel(r'$E = mc^2$'), findsOneWidget);
    expect(tester.getSize(find.byType(SvgPicture)), const Size(40, 20));

    // A file that is no SVG shows the missing image placeholder, also in
    // place of a picture shown before.
    await pumpCard(
      tester,
      '<div class="card"><img alt="broken" src="broken.svg"></div>',
      media: cache,
    );
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('missing-image')), findsOneWidget);
  });

  testWidgets('shows an image from the collection media', (tester) async {
    final requested = <String>[];
    final cache = MediaCache(
      load: (name) async {
        requested.add(name);
        return Uint8List.fromList(pixel);
      },
    );
    await pumpCard(
      tester,
      '<div class="card">perro<br><img src="dog.png"></div>',
      media: cache,
    );
    expect(find.byType(Image), findsOneWidget);
    expect(find.byKey(const Key('missing-image')), findsNothing);
    expect(requested, ['dog.png']);
    // A second view of the same image comes from the cache.
    await pumpCard(tester, '<div><img src="dog.png"></div>', media: cache);
    expect(requested, ['dog.png']);
  });

  group('image size', () {
    Future<Size> imageSize(WidgetTester tester, String img) async {
      final cache = MediaCache(load: (_) async => Uint8List.fromList(pixel));
      await pumpCard(tester, '<div class="card">$img</div>', media: cache);
      // Images decode outside the test's fake clock.
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      // On screen, after any scaling down.
      return tester.getRect(find.byType(Image)).size;
    }

    testWidgets('a width set in the editor is kept', (tester) async {
      expect(
        await imageSize(tester, '<img src="dog.png" width="120">'),
        const Size(120, 120),
      );
    });

    testWidgets('width and height attributes, also with px', (tester) async {
      expect(
        await imageSize(tester, '<img src="dog.png" width="80px" height="40">'),
        const Size(80, 40),
      );
      expect(
        await imageSize(tester, '<img src="dog.png" height="60">'),
        const Size(60, 60),
      );
    });

    testWidgets('a width in the style attribute wins', (tester) async {
      expect(
        await imageSize(
          tester,
          '<img src="dog.png" width="300" style="border: 0; width: 90px">',
        ),
        const Size(90, 90),
      );
    });

    testWidgets('never wider than the card', (tester) async {
      final card =
          tester.view.physicalSize.width / tester.view.devicePixelRatio;
      final size = await imageSize(
        tester,
        '<img src="dog.png" width="5000" height="100">',
      );
      expect(size.width, lessThanOrEqualTo(card));
      // The height shrinks with the width, keeping the proportions.
      expect(size.height, closeTo(100 * size.width / 5000, 0.01));
    });

    testWidgets('a percentage is of the card width', (tester) async {
      final half = await imageSize(tester, '<img src="dog.png" width="50%">');
      final full = await imageSize(tester, '<img src="dog.png" width="100%">');
      expect(half.width, closeTo(full.width / 2, 0.01));
    });

    testWidgets('without a size the image keeps its own size', (tester) async {
      expect(await imageSize(tester, '<img src="dog.png">'), const Size(1, 1));
    });

    testWidgets('a width from the note type CSS (spec: fixed width)', (
      tester,
    ) async {
      // Go inlines `img { width: 100px }` as the image's style.
      expect(
        await imageSize(tester, '<img src="dog.png" style="width:100px">'),
        const Size(100, 100),
      );
    });

    testWidgets('max-width and max-height in pixels', (tester) async {
      expect(
        await imageSize(
          tester,
          '<img src="dog.png" width="120" style="max-width:40px">',
        ),
        const Size(40, 40),
      );
      expect(
        await imageSize(
          tester,
          '<img src="dog.png" width="120" height="60" style="max-height:30px">',
        ),
        const Size(60, 30),
      );
    });

    testWidgets('a percentage max-width is of the card width', (tester) async {
      final full = await imageSize(tester, '<img src="dog.png" width="100%">');
      final half = await imageSize(
        tester,
        '<img src="dog.png" width="5000" height="100" style="max-width:50%">',
      );
      expect(half.width, closeTo(full.width / 2, 0.01));
      expect(half.height, closeTo(100 * half.width / 5000, 0.01));
    });

    testWidgets('max-width: none keeps the natural width and scrolls '
        '(spec: no max width)', (tester) async {
      final wide = await tester.runAsync(() => _png(2000, 10));
      final cache = MediaCache(load: (_) async => wide!);
      await pumpCard(
        tester,
        '<div class="card"><img src="wide.png" style="max-width:none"></div>',
        media: cache,
      );
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      expect(tester.getSize(find.byType(Image)), const Size(2000, 10));
      final scroll = find.byKey(const Key('image-scroll'));
      final card = tester.getRect(find.byKey(const Key('card-surface')));
      expect(tester.getSize(scroll).width, lessThanOrEqualTo(card.width));
      final before = tester.getRect(find.byType(Image)).left;
      await tester.drag(scroll, const Offset(-500, 0));
      await tester.pumpAndSettle();
      expect(tester.getRect(find.byType(Image)).left, lessThan(before - 400));
    });

    testWidgets('without max-width a wide image is scaled to the card', (
      tester,
    ) async {
      final wide = await tester.runAsync(() => _png(2000, 10));
      final cache = MediaCache(load: (_) async => wide!);
      await pumpCard(
        tester,
        '<div class="card"><img src="wide.png"></div>',
        media: cache,
      );
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      final card = tester.getRect(find.byKey(const Key('card-surface')));
      expect(
        tester.getSize(find.byType(Image)).width,
        lessThanOrEqualTo(card.width),
      );
      expect(find.byKey(const Key('image-scroll')), findsNothing);
    });

    testWidgets('alone in its block it starts at the edge, or follows '
        'text-align', (tester) async {
      final cache = MediaCache(load: (_) async => Uint8List.fromList(pixel));
      Future<Rect> place(String card) async {
        await pumpCard(tester, card, media: cache);
        await tester.runAsync(
          () => Future<void>.delayed(const Duration(milliseconds: 200)),
        );
        await tester.pumpAndSettle();
        return tester.getRect(find.byType(Image));
      }

      final start = await place(
        '<div class="card"><img src="dog.png" width="120"></div>',
      );
      final surface = tester.getRect(find.byKey(const Key('card-surface')));
      expect(start.size, const Size(120, 120));
      expect(start.left, closeTo(surface.left + 24, 0.5));
      final centered = await place(
        '<div class="card" style="text-align: center">'
        '<img src="dog.png" width="120"></div>',
      );
      expect(centered.center.dx, closeTo(surface.center.dx, 0.5));
    });
  });

  testWidgets('a missing image shows a placeholder, not an error', (
    tester,
  ) async {
    final cache = MediaCache(
      load: (name) async => throw 'Media file "$name" is missing.',
    );
    await pumpCard(tester, '<div>gato <img src="cat.jpg"></div>', media: cache);
    expect(find.byKey(const Key('missing-image')), findsOneWidget);
    expect(find.textContaining('gato', findRichText: true), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('remote images are never loaded', (tester) async {
    var loads = 0;
    final cache = MediaCache(
      load: (_) async {
        loads++;
        return Uint8List.fromList(pixel);
      },
    );
    await pumpCard(
      tester,
      '<img src="https://example.org/x.png">',
      media: cache,
    );
    expect(loads, 0);
    expect(find.byKey(const Key('missing-image')), findsOneWidget);
  });

  group('background images', () {
    List<DecorationImage> backgrounds(WidgetTester tester) => [
      for (final w in tester.widgetList<DecoratedBox>(
        find.byType(DecoratedBox),
      ))
        if (w.decoration case BoxDecoration(:final image?)) image,
    ];

    testWidgets('a card background from the media (spec)', (tester) async {
      final requested = <String>[];
      final cache = MediaCache(
        load: (name) async {
          requested.add(name);
          return Uint8List.fromList(pixel);
        },
      );
      // As Go inlines `.card { background-image: url("_paper.png") }`.
      await pumpCard(
        tester,
        '<div class="card card1" style="background-image:url(&quot;_paper.png&quot;)">'
        'hola</div>',
        media: cache,
      );
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      expect(backgrounds(tester), hasLength(1));
      expect(requested, ['_paper.png']);
      expect(find.textContaining('hola', findRichText: true), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('a missing file leaves the background empty', (tester) async {
      final cache = MediaCache(
        load: (name) async => throw 'Media file "$name" is missing.',
      );
      await pumpCard(
        tester,
        '<div style="background-image:url(_gone.png)">hola</div>',
        media: cache,
      );
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      expect(find.textContaining('hola', findRichText: true), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('remote and file URLs are never loaded', (tester) async {
      var loads = 0;
      final cache = MediaCache(
        load: (_) async {
          loads++;
          return Uint8List.fromList(pixel);
        },
      );
      for (final url in [
        'https://example.org/p.png',
        '//example.org/p.png',
        'file:///etc/p.png',
      ]) {
        await pumpCard(
          tester,
          '<div style="background-image:url($url)">hola</div>'
          '<p style="background: url($url)">b</p>',
          media: cache,
        );
        expect(backgrounds(tester), isEmpty, reason: url);
      }
      expect(loads, 0);
    });
  });

  testWidgets('dark mode: no note type colours means theme colours', (
    tester,
  ) async {
    final dark = darkTheme();
    await pumpCard(tester, '<div class="card">texto</div>', theme: dark);
    expect(surfaceOf(tester), dark.colorScheme.surface);
    expect(dark.colorScheme.surface.computeLuminance(), lessThan(0.2));
    final text = tester.widget<RichText>(find.byType(RichText).first);
    final fg =
        text.text.style?.color ??
        (text.text as TextSpan).children?.first.style?.color;
    expect(fg, isNotNull);
    expect(fg!.computeLuminance(), greaterThan(0.5));
  });

  testWidgets('explicit note type colours win in dark mode', (tester) async {
    await pumpCard(
      tester,
      '<div>x</div>',
      theme: darkTheme(),
      background: '#fdf6e3',
      color: '#222',
    );
    expect(surfaceOf(tester), const Color(0xFFFDF6E3));
  });

  test('CSS colours', () {
    expect(parseCssColor('#fff'), const Color(0xFFFFFFFF));
    expect(parseCssColor('#fdf6e3'), const Color(0xFFFDF6E3));
    expect(parseCssColor('#11223380'), const Color(0x80112233));
    expect(parseCssColor('rgb(255, 0, 0)'), const Color(0xFFFF0000));
    expect(parseCssColor('rgba(0,0,255,0.5)')?.toARGB32(), 0x800000FF);
    expect(parseCssColor('Navy'), const Color(0xFF000080));
    expect(parseCssColor('url(x)'), isNull);
    expect(parseCssColor(null), isNull);
    const black = Color(0xFF000000);
    const white = Color(0xFFFFFFFF);
    expect(contrastRatio(black, white), closeTo(21, 0.01));
    expect(contrastRatio(white, black), closeTo(21, 0.01));
    expect(contrastRatio(black, const Color(0xFFFF0000)), closeTo(5.25, 0.01));
    expect(contrastRatio(const Color(0xFF336699), const Color(0xFF336699)), 1);
  });

  testWidgets('a sound becomes a play button in place (spec)', (tester) async {
    final backend = FakeBackend();
    await pumpSounds(
      tester,
      '<div class="card"><cf-sound idx="0"></cf-sound> palabra</div>',
      const [word],
      backend: backend,
      available: ['word.mp3'],
    );
    expect(find.textContaining('palabra', findRichText: true), findsOneWidget);
    expect(find.byTooltip('Play word.mp3'), findsOneWidget);
    await tester.tap(find.byTooltip('Play word.mp3'));
    await tester.pump();
    expect(backend.played, ['word.mp3']);
    expect(find.byTooltip('Stop'), findsOneWidget);
    backend.finish();
    await tester.pumpAndSettle();
    expect(find.byTooltip('Play word.mp3'), findsOneWidget);
  });

  testWidgets('a missing sound file shows a disabled control (spec)', (
    tester,
  ) async {
    await pumpSounds(
      tester,
      '<div class="card"><cf-sound idx="0"></cf-sound> palabra</div>',
      const [word],
    );
    expect(find.textContaining('palabra', findRichText: true), findsOneWidget);
    expect(find.text('word.mp3 is missing'), findsOneWidget);
    final button = tester.widget<TextButton>(
      find.byKey(const Key('sound-question-0')),
    );
    expect(button.onPressed, isNull);
  });

  testWidgets('a sound the device cannot play is disabled after trying', (
    tester,
  ) async {
    final backend = FakeBackend()..unplayable.add('word.mp3');
    await pumpSounds(
      tester,
      '<cf-sound idx="0"></cf-sound>',
      const [word],
      backend: backend,
      available: ['word.mp3'],
    );
    await tester.tap(find.byTooltip('Play word.mp3'));
    await tester.pumpAndSettle();
    expect(
      find.text('word.mp3 cannot be played on this device'),
      findsOneWidget,
    );
  });

  testWidgets('play buttons are hidden when the setting is off', (
    tester,
  ) async {
    await pumpSounds(
      tester,
      '<cf-sound idx="0"></cf-sound> palabra',
      const [word],
      available: ['word.mp3'],
      showPlayButtons: false,
    );
    expect(find.textContaining('palabra', findRichText: true), findsOneWidget);
    expect(find.byTooltip('Play word.mp3'), findsNothing);
  });

  testWidgets('a video shows an inline player (spec)', (tester) async {
    const clip = MediaItem(CardSide.answer, 0, 'clip.mp4', video: true);
    final backend = FakeBackend();
    await pumpSounds(
      tester,
      '<cf-video idx="0"></cf-video>',
      const [clip],
      backend: backend,
      available: ['clip.mp4'],
    );
    expect(find.byKey(const Key('video-answer-0')), findsOneWidget);
    await tester.tap(find.byTooltip('Play clip.mp4'));
    await tester.pump();
    expect(backend.played, ['clip.mp4']);
    expect(find.byKey(const Key('fake-video')), findsOneWidget);
  });

  testWidgets('placeholders without a player or out of range show nothing', (
    tester,
  ) async {
    await pumpCard(tester, '<cf-sound idx="0"></cf-sound>ok');
    expect(find.byType(SoundButton), findsNothing);
    await pumpSounds(tester, '<cf-sound idx="3"></cf-sound>ok', const [word]);
    expect(find.byType(SoundButton), findsNothing);
  });

  group('text to speech', () {
    MediaItem tts(String lang, String text, {int idx = 0}) => MediaItem.tts(
      CardSide.question,
      idx,
      golang.Utterance(text: text, lang: lang, voices: const [], speed: 1),
    );

    testWidgets('a tts tag becomes a speak button in place (spec)', (
      tester,
    ) async {
      final engine = FakeSpeechEngine();
      useFakeSpeech(engine);
      await pumpSounds(tester, 'perro<cf-tts idx="0"></cf-tts>', [
        tts('es_ES', 'perro'),
      ]);
      expect(find.textContaining('perro', findRichText: true), findsOneWidget);
      await tester.tap(find.byTooltip('Speak'));
      await tester.pump();
      expect(engine.spoken, ['es_ES:perro']);
      expect(find.byTooltip('Stop'), findsOneWidget);
      engine.finish();
      await tester.pumpAndSettle();
      expect(find.byTooltip('Speak'), findsOneWidget);
    });

    testWidgets('no voice for the language: disabled, with the reason '
        '(spec)', (tester) async {
      useFakeSpeech(FakeSpeechEngine());
      await pumpSounds(tester, '<cf-tts idx="0"></cf-tts>', [
        tts('ja_JP', '猫'),
      ]);
      final button = tester.widget<IconButton>(
        find.byKey(const Key('tts-question-0')),
      );
      expect(button.onPressed, isNull);
      expect(button.tooltip, 'No voice is installed for ja_JP');
    });

    testWidgets('a sound and an utterance with the same index', (tester) async {
      useFakeSpeech(FakeSpeechEngine());
      await pumpSounds(
        tester,
        '<cf-sound idx="0"></cf-sound><cf-tts idx="0"></cf-tts>',
        [word, tts('es_ES', 'palabra')],
        available: ['word.mp3'],
      );
      expect(find.byTooltip('Play word.mp3'), findsOneWidget);
      expect(find.byTooltip('Speak'), findsOneWidget);
    });

    testWidgets('hidden with the play buttons setting off', (tester) async {
      useFakeSpeech(FakeSpeechEngine());
      await pumpSounds(tester, 'perro<cf-tts idx="0"></cf-tts>', [
        tts('es_ES', 'perro'),
      ], showPlayButtons: false);
      expect(find.byKey(const Key('tts-question-0')), findsNothing);
    });

    testWidgets('{{tts-voices:}} lists the device voices (spec)', (
      tester,
    ) async {
      useFakeSpeech(
        FakeSpeechEngine(const [
          Voice('Samantha', 'en_US'),
          Voice('Mónica', 'es_ES'),
        ]),
      );
      await pumpCard(tester, '<cf-tts-voices></cf-tts-voices>');
      expect(find.text('en_US Samantha'), findsOneWidget);
      expect(find.text('es_ES Mónica'), findsOneWidget);
    });

    testWidgets('{{tts-voices:}} without voices says so', (tester) async {
      useFakeSpeech(FakeSpeechEngine(const []));
      await pumpCard(tester, '<cf-tts-voices></cf-tts-voices>');
      expect(
        find.text('No voices are installed on this device.'),
        findsOneWidget,
      );
    });
  });

  testWidgets('a formula is typeset inline (spec)', (tester) async {
    await pumpCard(
      tester,
      r'Area is <cf-math display="inline">\pi r^2</cf-math>',
    );
    expect(find.byType(Math), findsOneWidget);
    expect(find.textContaining(r'\pi'), findsNothing);
    expect(find.byType(SingleChildScrollView), findsNothing);
  });

  testWidgets('a display formula is a fraction on its own line (spec)', (
    tester,
  ) async {
    await pumpCard(tester, r'<cf-math display="block">\frac{a}{b}</cf-math>');
    final math = tester.widget<Math>(find.byType(Math));
    expect(math.parseError, isNull);
    expect(math.mathStyle, MathStyle.display);
    // Wide formulas scroll horizontally in their own box.
    expect(
      find.ancestor(
        of: find.byType(Math),
        matching: find.byType(SingleChildScrollView),
      ),
      findsOneWidget,
    );
  });

  testWidgets('an unknown command shows the source, the rest renders (spec)', (
    tester,
  ) async {
    await pumpCard(
      tester,
      r'<cf-math display="inline">\unknowncommand{x}</cf-math> text',
    );
    final error = tester.widget<Text>(find.byKey(const Key('math-error')));
    expect(error.data, r'\unknowncommand{x}');
    expect(error.style!.color, lightTheme().colorScheme.error);
    expect(find.textContaining(' text', findRichText: true), findsOneWidget);
  });

  testWidgets('a cloze in a formula gets the highlight colour', (tester) async {
    await pumpCard(
      tester,
      r'<cf-math display="inline">E = \cfcloze{mc^2}</cf-math>',
    );
    final math = tester.widget<Math>(find.byType(Math));
    expect(math.parseError, isNull);
    final colours = tester
        .widgetList<RichText>(
          find.descendant(
            of: find.byType(Math),
            matching: find.byType(RichText),
          ),
        )
        .map((t) => t.text.style?.color)
        .toSet();
    expect(colours, contains(lightTheme().colorScheme.primary));
  });

  testWidgets('the answer placeholder becomes an input where it is', (
    tester,
  ) async {
    final controller = TextEditingController();
    final focus = FocusNode();
    addTearDown(controller.dispose);
    addTearDown(focus.dispose);
    var submitted = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: CardView(
            html:
                '<div class="card"><p>dog</p><cf-typeans></cf-typeans>'
                '<p>below</p></div>',
            typeAnswer: TypeAnswerField(
              controller: controller,
              focusNode: focus,
              onSubmitted: () => submitted++,
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    final field = find.byKey(const Key('type-answer'));
    expect(field, findsOneWidget);
    // Between the text before and after the placeholder.
    final y = tester.getCenter(field).dy;
    expect(
      tester.getCenter(find.text('dog', findRichText: true)).dy,
      lessThan(y),
    );
    expect(
      tester.getCenter(find.text('below', findRichText: true)).dy,
      greaterThan(y),
    );

    await tester.enterText(field, 'perro');
    expect(controller.text, 'perro');
    await tester.testTextInput.receiveAction(TextInputAction.done);
    expect(submitted, 1);
    expect(focus.hasFocus, isTrue, reason: 'the caller decides on focus');
  });

  testWidgets('without an input (previews) the placeholder shows nothing', (
    tester,
  ) async {
    await pumpCard(tester, '<p>dog</p><cf-typeans></cf-typeans>');
    expect(find.text('dog', findRichText: true), findsOneWidget);
    expect(find.byType(TextField), findsNothing);
  });

  test('media cache evicts the least recently used', () async {
    final loads = <String>[];
    final cache = MediaCache(
      load: (name) async {
        loads.add(name);
        return Uint8List(10);
      },
      maxBytes: 25,
    );
    await cache.get('a');
    await cache.get('b');
    await cache.get('a'); // cached; a is now the most recent
    await cache.get('c'); // over 25 bytes: evicts b
    await cache.get('a'); // still cached
    await cache.get('b'); // loaded again
    expect(loads, ['a', 'b', 'c', 'b']);
  });

  group('the type-answer comparison stays readable (spec)', () {
    const black = Color(0xFF000000);

    testWidgets('a black mark without a valid highlight in dark mode', (
      tester,
    ) async {
      final dark = darkTheme();
      await pumpCard(
        tester,
        'か <hr id="answer"><code id="typeans"><span class="typeGood" '
        'style="color:black;background-color:light green">ka</span></code>',
        theme: dark,
      );
      expect(styleOf(tester, 'ka').color, dark.colorScheme.onSurface);
    });

    testWidgets('a black mark on its own highlight is kept', (tester) async {
      await pumpCard(
        tester,
        '<code id="typeans"><span class="typeGood" '
        'style="color:black;background-color:light green">k</span>'
        '<span class="typeBad" style="color:black;background-color:red">i'
        '</span></code>',
        theme: darkTheme(),
      );
      expect(styleOf(tester, 'i').color, black);
      expect(styleOf(tester, 'k').color, isNot(black));
    });

    testWidgets('unchanged in light mode', (tester) async {
      await pumpCard(
        tester,
        '<code id="typeans"><span class="typeGood" style="color:black">ka'
        '</span></code>',
      );
      expect(styleOf(tester, 'ka').color, black);
    });

    testWidgets('the card colour is too faint as well', (tester) async {
      await pumpCard(
        tester,
        '<code id="typeans"><span class="typeGood" '
        'style="color:#ddd;background-color:#eee">ka</span></code>',
        color: 'white',
      );
      expect(styleOf(tester, 'ka').color, black);
    });

    testWidgets('the rest of the card is untouched', (tester) async {
      await pumpCard(
        tester,
        '<span class="hint" style="color:black">hint</span>'
        '<code id="typeans"><span class="typeGood" style="color:black">ka'
        '</span></code>',
        theme: darkTheme(),
      );
      expect(styleOf(tester, 'hint').color, black);
      expect(styleOf(tester, 'ka').color, isNot(black));
    });

    testWidgets('the rest of the answer renders as before', (tester) async {
      const answer =
          '<div class="card">か<hr id="answer"><code id="typeans">'
          '<span class="typeGood" style="color:black">k</span>'
          '<span class="typeBad" style="color:black;background-color:red">i'
          '</span><br><span id="typearrow">&darr;</span><br>'
          '<span class="typeGood" style="color:black">k</span>'
          '<span class="typeMissed" style="color:black;background-color:grey">'
          'a</span></code><cf-sound idx="0"></cf-sound> palabra</div>';
      await pumpSounds(
        tester,
        answer,
        const [word],
        available: ['word.mp3'],
        theme: darkTheme(),
      );
      final widget = tester.widget<HtmlWidget>(find.byType(HtmlWidget));
      expect(widget.html, isNot(answer));
      expect(find.byTooltip('Play word.mp3'), findsOneWidget);
      expect(widget.html, contains('<hr id="answer">'));
      expect(widget.html, contains('<cf-sound idx="0"></cf-sound>'));
      expect(
        find.textContaining('palabra', findRichText: true),
        findsOneWidget,
      );
      expect(find.textContaining('↓', findRichText: true), findsOneWidget);
      expect(styleOf(tester, 'i').color, black);
      expect(styleOf(tester, 'a').color, black);
    });

    testWidgets('HTML without a comparison is passed on unchanged', (
      tester,
    ) async {
      const question =
          '<div class="card" style="color:black">か<cf-typeans></cf-typeans>'
          '</div>';
      await pumpCard(tester, question, theme: darkTheme());
      expect(tester.widget<HtmlWidget>(find.byType(HtmlWidget)).html, question);
      const plain = '<b style="color:black">x</b>';
      await pumpCard(tester, plain, theme: darkTheme());
      expect(tester.widget<HtmlWidget>(find.byType(HtmlWidget)).html, plain);
    });
  });

  testWidgets('ruby text shows the reading above the base (spec)', (
    tester,
  ) async {
    await pumpCard(
      tester,
      '<ruby>日本語<rp>(</rp><rt>にほんご</rt><rp>)</rp></ruby>を',
    );
    RichText richText(String text) => tester.widget<RichText>(
      find.byWidgetPredicate(
        (w) => w is RichText && w.text.toPlainText() == text,
      ),
    );
    double fontSize(String text) {
      double? size;
      richText(text).text.visitChildren((span) {
        size ??= span.style?.fontSize;
        return size == null;
      });
      return size!;
    }

    final base = richText('日本語');
    final reading = richText('にほんご');
    final baseRect = tester.getRect(find.byWidget(base));
    final readingRect = tester.getRect(find.byWidget(reading));
    expect(readingRect.bottom, lessThanOrEqualTo(baseRect.top + 1));
    expect(readingRect.center.dx, closeTo(baseRect.center.dx, 1));
    expect(fontSize('にほんご'), lessThan(fontSize('日本語')));
    // <rp> is the fallback for displays without ruby and stays hidden.
    expect(find.textContaining('(', findRichText: true), findsNothing);
    expect(find.textContaining('を', findRichText: true), findsOneWidget);
  });

  // Card root classes and environment media queries (card-rendering spec):
  // the shown card follows a theme switch without leaving it.
  testWidgets('switching the theme renders the shown card again', (
    tester,
  ) async {
    final go = FakeGo()..settings['theme'] = 'light';
    final night = <bool>[];
    Map<String, Object?> state() {
      final dark = go.renderEnv?['night'] == true;
      night.add(dark);
      final html = dark
          ? '<div class="card card1 nightMode">hola <b>noche</b></div>'
          : '<div class="card card1">hola <b>día</b></div>';
      return {
        'deckId': 1,
        'deckName': 'Default',
        'card': {
          'cardId': 7,
          'noteId': 70,
          'question': html,
          'answer': html,
          'kind': 'new',
          'empty': false,
          'previews': <Object?>[],
          'autoplay': true,
          'replayQuestion': true,
          'showTimer': false,
          'stopTimerOnAnswer': true,
          'maxAnswerSeconds': 60,
          'secondsToShowQuestion': 0,
          'secondsToShowAnswer': 0,
          'questionAction': 'showAnswer',
          'answerAction': 'buryCard',
          'typeAnswer': false,
          'flag': 0,
          'marked': false,
        },
        'counts': {'new': 1, 'learning': 0, 'review': 0},
        'canUndo': false,
        'studied': 0,
        'timeSpent': 0,
        'nextDue': null,
        'learningDue': null,
        'leech': false,
        'leechSuspended': false,
      };
    }

    go.handle('study.Start', (_) => state());
    go.handle('study.Current', (_) => state());
    await pumpApp(tester);
    await studyDeck(tester, 'Default');
    final hola = find.textContaining('hola', findRichText: true);
    expect(find.textContaining('día', findRichText: true), findsOneWidget);
    expect(go.renderEnv, {
      'night': false,
      'platform': 'android mobile',
      'widthPx': 400,
      'heightPx': 800,
    });

    go.settings = {...go.settings, 'theme': 'dark'};
    go.emitChange(settings: true, decks: false);
    // Every frame shows a card: the old one until the new one arrives.
    for (var i = 0; i < 30; i++) {
      await tester.pump(const Duration(milliseconds: 20));
      expect(hola, findsOneWidget, reason: 'frame $i');
    }
    await tester.pumpAndSettle();
    expect(go.renderEnv?['night'], isTrue);
    expect(night.last, isTrue);
    expect(go.callsTo('study.Current'), isNotEmpty);
    expect(find.textContaining('noche', findRichText: true), findsOneWidget);
    expect(find.textContaining('día', findRichText: true), findsNothing);
  });

  group('answer anchor', () {
    Future<List<BuildContext>> pumpAnchored(
      WidgetTester tester,
      String html,
    ) async {
      final reported = <BuildContext>[];
      await tester.pumpWidget(
        MaterialApp(
          theme: lightTheme(),
          home: Scaffold(
            body: CardView(html: html, onAnswerAnchor: reported.add),
          ),
        ),
      );
      await tester.pumpAndSettle();
      return reported;
    }

    final anchor = find.byKey(const Key('answer-anchor'));

    testWidgets('is reported once, right above the divider', (tester) async {
      final reported = await pumpAnchored(
        tester,
        '<div>front<hr id="answer">back</div>',
      );
      expect(reported, hasLength(1));
      expect(anchor, findsOneWidget);
      final front = tester.getRect(
        find.textContaining('front', findRichText: true),
      );
      final back = tester.getRect(
        find.textContaining('back', findRichText: true),
      );
      final top = tester.getTopLeft(anchor).dy;
      expect(top, greaterThanOrEqualTo(front.bottom));
      expect(top, lessThan(back.top));
    });

    testWidgets('keeps an inline answer on its line', (tester) async {
      await pumpAnchored(tester, '<div>one <span id="answer">two</span></div>');
      expect(anchor, findsOneWidget);
      // One paragraph of one line.
      final one = find.textContaining('one', findRichText: true);
      final two = find.textContaining('two', findRichText: true);
      expect(tester.widget(one), same(tester.widget(two)));
      expect(tester.getSize(one).height, lessThan(40));
    });

    testWidgets('nothing without an answer element', (tester) async {
      final reported = await pumpAnchored(tester, '<div>front<hr>back</div>');
      expect(reported, isEmpty);
      expect(anchor, findsNothing);
    });
  });

  group('fonts from media (spec)', () {
    const jp = golang.FontFace(family: 'jp', file: '_jp.ttf');
    late List<String> loads;
    late List<String> registered;

    FontCache fontCache({
      Set<String> files = const {'_jp.ttf'},
      Completer<void>? wait,
    }) {
      loads = [];
      registered = [];
      return FontCache(
        load: (name) async {
          loads.add(name);
          await wait?.future;
          if (!files.contains(name)) throw Exception('missing $name');
          return Uint8List.fromList([0, 1, 0, 0]);
        },
        register: (family, fonts) async => registered.add(family),
      );
    }

    Future<void> pumpFonts(
      WidgetTester tester,
      FontCache cache,
      List<String> sides, {
      List<golang.FontFace> fonts = const [jp],
      bool settle = true,
    }) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: [
                for (final html in sides)
                  CardView(html: html, fonts: fonts, fontCache: cache),
              ],
            ),
          ),
        ),
      );
      if (settle) await tester.pumpAndSettle();
    }

    testWidgets('a deck font is applied', (tester) async {
      final cache = fontCache();
      await pumpFonts(tester, cache, [
        '<div class="card" style="font-family:&quot;JP&quot;, serif">hola</div>',
      ]);
      final style = styleOf(tester, 'hola');
      expect(registered, ['cf-font-0']);
      expect(style.fontFamily, 'cf-font-0');
      expect(style.fontFamilyFallback, ['serif']);
    });

    testWidgets('a font file is loaded once for two cards', (tester) async {
      final cache = fontCache();
      await pumpFonts(tester, cache, [
        '<div style="font-family:jp">one</div>',
        '<div style="font-family:jp">two</div>',
      ]);
      expect(loads, ['_jp.ttf']);
      expect(registered, hasLength(1));
      expect(styleOf(tester, 'one').fontFamily, 'cf-font-0');
      expect(styleOf(tester, 'two').fontFamily, 'cf-font-0');
      // The next card has the font at once.
      await pumpFonts(tester, cache, [
        '<div style="font-family:jp">three</div>',
      ], settle: false);
      expect(styleOf(tester, 'three').fontFamily, 'cf-font-0');
      expect(loads, hasLength(1));
    });

    testWidgets('a missing file renders with the default font', (tester) async {
      final cache = fontCache(files: {});
      await pumpFonts(
        tester,
        cache,
        ['<div style="font-family:missing">hola</div>'],
        fonts: const [golang.FontFace(family: 'missing', file: '_missing.ttf')],
      );
      expect(registered, isEmpty);
      // An unregistered family: the engine uses the default font.
      expect(styleOf(tester, 'hola').fontFamily, 'missing');
      expect(tester.takeException(), isNull);
    });

    testWidgets('fonts the card does not use are not loaded', (tester) async {
      final cache = fontCache();
      await pumpFonts(tester, cache, ['<div>hola</div>']);
      expect(loads, isEmpty);
      expect(find.text('hola', findRichText: true), findsOneWidget);
    });

    testWidgets('the card waits for its fonts for at most a second', (
      tester,
    ) async {
      final wait = Completer<void>();
      final cache = fontCache(wait: wait);
      await pumpFonts(tester, cache, [
        '<div style="font-family:jp">hola</div>',
      ], settle: false);
      await tester.pump(const Duration(milliseconds: 900));
      expect(find.text('hola', findRichText: true), findsNothing);
      await tester.pump(const Duration(milliseconds: 100));
      expect(styleOf(tester, 'hola').fontFamily, 'jp');
      // A font that arrives later is still used.
      wait.complete();
      await tester.pumpAndSettle();
      expect(styleOf(tester, 'hola').fontFamily, 'cf-font-0');
    });
  });

  group('image occlusion', () {
    // Anki's default colours.
    const inactive = Color(0xFFFFEBA2);
    const active = Color(0xFFFF8E8E);
    const border = Color(0xFF212121);
    const highlight = Color(0x00FF8E8E);
    const white = Color(0xFFFFFFFF);

    /// Matches a drawOval call in [color] (and of [rect]).
    PaintPatternPredicate oval(Color color, [Rect? rect]) =>
        (method, args) =>
            method == #drawOval &&
            (rect == null || args[0] == rect) &&
            (args[1] as Paint).color.toARGB32() == color.toARGB32();

    /// A card side as Go renders it: three shapes, card [ord], on a 200×100
    /// image, and a label.
    String side(int ord, String side, {required bool oi, String style = ''}) {
      final shapes = [
        {
          'kind': 'rect',
          'ord': 1,
          'left': .1,
          'top': .2,
          'width': .3,
          'height': .1,
          'oi': oi,
        },
        {
          'kind': 'rect',
          'ord': 2,
          'left': .5,
          'top': .5,
          'width': .2,
          'height': .2,
          'oi': oi,
        },
        {
          'kind': 'ellipse',
          'ord': 3,
          'left': .0,
          'top': .0,
          'rx': .1,
          'ry': .1,
          'oi': oi,
        },
        {
          'kind': 'text',
          'ord': 0,
          'left': .6,
          'top': .1,
          'text': 'Aorta',
          'scale': 1,
          'fontSize': .1,
        },
      ];
      return '<div class="card"><div>Heart</div><cf-occlusion image="heart.png" '
          'id="image-occlusion-canvas" shapes="${htmlEscape.convert(jsonEncode(shapes))}" '
          'active="$ord" side="$side" style="$style"></cf-occlusion></div>';
    }

    Future<RenderObject> pumpMasks(WidgetTester tester, String html) async {
      final image = await tester.runAsync(() => _png(200, 100));
      final requested = <String>[];
      final cache = MediaCache(
        load: (name) async {
          requested.add(name);
          return image!;
        },
      );
      await pumpCard(tester, html, media: cache);
      await tester.runAsync(
        () => Future<void>.delayed(const Duration(milliseconds: 200)),
      );
      await tester.pumpAndSettle();
      expect(requested, ['heart.png']);
      expect(find.text('Heart', findRichText: true), findsOneWidget);
      expect(
        tester.getSize(find.byKey(const Key('occlusion-image'))),
        const Size(200, 100),
      );
      return tester.renderObject(find.byKey(const Key('occlusion-masks')));
    }

    testWidgets('hide all, guess one: question (spec)', (tester) async {
      // All three areas masked, shape 2 in the active colour.
      final masks = await pumpMasks(tester, side(2, 'q', oi: true));
      expect(
        masks,
        paints
          ..translate(x: 20, y: 20)
          ..rect(rect: const Rect.fromLTWH(0, 0, 60, 10), color: inactive)
          ..rect(
            rect: const Rect.fromLTWH(0, 0, 60, 10),
            color: border,
            style: PaintingStyle.stroke,
            strokeWidth: 1,
          )
          ..translate(x: 100, y: 50)
          ..rect(rect: const Rect.fromLTWH(0, 0, 40, 20), color: active)
          ..rect(rect: const Rect.fromLTWH(0, 0, 40, 20), color: border)
          ..translate(x: 0, y: 0)
          ..something(oval(inactive, const Rect.fromLTWH(0, 0, 40, 20)))
          ..something(oval(border))
          // The label, on top.
          ..translate(x: 120, y: 10)
          ..rect(color: white)
          ..paragraph(),
      );
    });

    testWidgets('hide all, guess one: answer reveal (spec)', (tester) async {
      // Shape 2 is shown with an outline, shapes 1 and 3 stay masked.
      final masks = await pumpMasks(tester, side(2, 'a', oi: true));
      expect(
        masks,
        paints
          ..rect(rect: const Rect.fromLTWH(0, 0, 60, 10), color: inactive)
          ..rect(color: border)
          ..rect(rect: const Rect.fromLTWH(0, 0, 40, 20), color: highlight)
          ..rect(
            rect: const Rect.fromLTWH(0, 0, 40, 20),
            color: active,
            style: PaintingStyle.stroke,
          )
          ..something(oval(inactive)),
      );
    });

    testWidgets('hide one, guess one: only the asked shape', (tester) async {
      final question = await pumpMasks(tester, side(2, 'q', oi: false));
      // Shape 2 (fill and border) and the label's box.
      expect(question, paintsExactlyCountTimes(#drawRect, 3));
      expect(question, paintsExactlyCountTimes(#drawOval, 0));
      expect(
        question,
        paints
          ..translate(x: 100, y: 50)
          ..rect(rect: const Rect.fromLTWH(0, 0, 40, 20), color: active),
      );

      final answer = await pumpMasks(tester, side(2, 'a', oi: false));
      expect(answer, paintsExactlyCountTimes(#drawRect, 3));
      expect(answer, paintsExactlyCountTimes(#drawOval, 0));
      expect(
        answer,
        paints
          ..rect(rect: const Rect.fromLTWH(0, 0, 40, 20), color: highlight)
          ..rect(color: active, style: PaintingStyle.stroke),
      );
    });

    testWidgets("the note type's colours", (tester) async {
      const style =
          '--active-shape-color:#00ff00;--highlight-shape-border:3px #0000ff;'
          '--inactive-shape-color:rgb(10, 20, 30)';
      final question = await pumpMasks(
        tester,
        side(1, 'q', oi: true, style: style),
      );
      expect(
        question,
        paints
          ..rect(color: const Color(0xFF00FF00))
          ..rect(color: border)
          ..rect(color: const Color(0xFF0A141E)),
      );
      final answer = await pumpMasks(
        tester,
        side(1, 'a', oi: true, style: style),
      );
      expect(
        answer,
        paints
          ..rect(color: highlight)
          ..rect(color: const Color(0xFF0000FF), strokeWidth: 3),
      );
    });

    testWidgets('a missing image shows the placeholder, without masks', (
      tester,
    ) async {
      await pumpCard(
        tester,
        side(1, 'q', oi: true),
        media: MediaCache(load: (_) async => throw StateError('missing')),
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('missing-image')), findsOneWidget);
      expect(find.byKey(const Key('occlusion-masks')), findsNothing);
    });

    test('colours from CSS variables, Anki defaults otherwise', () {
      expect(OcclusionColors.fromStyle(null), const OcclusionColors());
      final c = OcclusionColors.fromStyle(
        'color: red; --inactive-shape-border: 2px solid #ff0000; '
        '--active-shape-color: nonsense',
      );
      expect(c.inactiveBorder, (width: 2.0, color: const Color(0xFFFF0000)));
      expect(c.active, active);
    });

    test('which shapes are drawn how', () {
      const all = OcclusionShape(kind: 'rect', ord: 2, occludeInactive: true);
      const one = OcclusionShape(kind: 'rect', ord: 2);
      const label = OcclusionShape(kind: 'text', text: 'x');
      expect(occlusionLook(all, 2, true), OcclusionLook.active);
      expect(occlusionLook(all, 2, false), OcclusionLook.highlight);
      expect(occlusionLook(all, 1, true), OcclusionLook.inactive);
      expect(occlusionLook(all, 1, false), OcclusionLook.inactive);
      expect(occlusionLook(one, 1, true), OcclusionLook.hidden);
      expect(occlusionLook(one, 1, false), OcclusionLook.hidden);
      expect(occlusionLook(label, 1, true), OcclusionLook.label);
    });
  });
}
