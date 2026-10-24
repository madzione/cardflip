import 'package:flutter/material.dart';
import 'package:flutter_quill/flutter_quill.dart';
import 'package:flutter_quill/quill_delta.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/features/editor/rich_field.dart';
import 'package:cardflip/features/editor/rich_html.dart';

/// HTML → document → HTML, also through a Quill [Document], as the editor
/// does.
String roundTrip(String html) {
  final doc = htmlToDocument(html);
  final quill = Document.fromDelta(doc.delta);
  return documentToHtml(quill.toDelta(), divLines: doc.divLines);
}

void main() {
  group('converters keep Anki-style HTML unchanged', () {
    // Every element of the subset, written as Anki's editor writes it.
    const exact = {
      'empty field': '',
      'plain text': 'perro',
      'bold': '<b>perro</b>',
      'italic': 'el <i>perro</i>',
      'underline': '<u>dog</u>',
      'strike': '<s>dog</s>',
      'subscript': 'H<sub>2</sub>O',
      'superscript': 'x<sup>2</sup>',
      'text colour': '<span style="color: #ff0000">red</span>',
      'highlight': '<span style="background-color: yellow">marked</span>',
      'link': '<a href="https://example.com/a?b=1&amp;c=2">link</a>',
      'line breaks': 'one<br>two<br><br>four',
      'trailing break': 'perro<br>',
      'entities': 'a &lt; b &amp;&amp; c &gt; d&nbsp;e',
      'bulleted list': '<ul><li>uno</li><li>dos</li></ul>',
      'numbered list': '<ol><li>first</li><li>second</li></ol>',
      'nested list':
          '<ul><li>a<ul><li>a1</li><li>a2<ol><li>x</li></ol></li></ul></li>'
          '<li>b</li></ul>',
      'empty list item': '<ul><li><br></li></ul>',
      'list around text': 'before<ul><li>item</li></ul>after',
      'centered line': 'a<div style="text-align: center">centered</div>b',
      // Explicit, because card CSS often centres text (Anki's default).
      'left-aligned line': 'a<div style="text-align: left">left</div>b',
      'left-aligned item': '<ul><li style="text-align: left">l</li></ul>',
      'right-aligned item': '<ul><li style="text-align: right">r</li></ul>',
      'indented line': '<div style="margin-left: 80px">deep</div>',
      'image': 'perro<br><img src="dog.png">',
      'resized image': '<img src="cat.jpg" width="120"> <b>cat</b>',
      'div lines': '<div>one</div><div><br></div><div>three</div>',
      'div lines with list': '<div>a</div><ul><li>b</li></ul><div>c</div>',
      'coloured underline': '<span style="color: red"><u>x</u> y</span>',
      'nested inline formats':
          '<span style="color: blue"><b>bold <i>both</i></b> plain</span>',
      'cloze and template text': '{{c1::Paris::city}} is in {{c2::France}}',
      'math': r'\(x^2\) and \[\frac{1}{2}\]',
      'sound tag': '[sound:hola.mp3] hola',
    };
    for (final MapEntry(key: name, value: html) in exact.entries) {
      test(name, () {
        expect(roundTrip(html), html);
        expect(isLossless(html), isTrue);
      });
    }
  });

  group('equivalent HTML is written the Anki way', () {
    const canonical = {
      '<strong>a</strong>': '<b>a</b>',
      '<em>a</em>': '<i>a</i>',
      '<strike>a</strike>': '<s>a</s>',
      '<font color="red">a</font>': '<span style="color: red">a</span>',
      '<i><b>a</b></i>': '<b><i>a</i></b>',
      '<span style="color:red;background-color:yellow">a</span>':
          '<span style="color: red"><span style="background-color: yellow">'
          'a</span></span>',
      '<b>a</b><b>b</b>': '<b>ab</b>',
      'a\n  b': 'a b',
    };
    for (final MapEntry(key: input, value: output) in canonical.entries) {
      test(input, () {
        expect(roundTrip(input), output);
        expect(isLossless(input), isTrue, reason: 'only the spelling changes');
      });
    }
  });

  test('the document has the expected attributes', () {
    final doc = htmlToDocument(
      '<b>a</b><span style="color: red">b</span><ul><li>c<ul><li>d</li></ul>'
      '</li></ul><img src="x.png" width="50">',
    );
    expect(doc.delta.toJson(), [
      {
        'insert': 'a',
        'attributes': {'bold': true},
      },
      {
        'insert': 'b',
        'attributes': {'color': 'red'},
      },
      {'insert': '\nc'},
      {
        'insert': '\n',
        'attributes': {'list': 'bullet'},
      },
      {'insert': 'd'},
      {
        'insert': '\n',
        'attributes': {'list': 'bullet', 'indent': 1},
      },
      {
        'insert': {'image': 'x.png'},
        'attributes': {'width': '50'},
      },
      {'insert': '\n'},
    ]);
  });

  test('editing the document writes Anki-style HTML', () {
    final doc = Document.fromDelta(htmlToDocument('perro gato').delta);
    doc.format(0, 5, Attribute.bold);
    doc.format(6, 4, const ColorAttribute('#0000ff'));
    doc.format(10, 0, Attribute.centerAlignment);
    expect(
      documentToHtml(doc.toDelta()),
      '<div style="text-align: center"><b>perro</b> '
      '<span style="color: #0000ff">gato</span></div>',
    );
  });

  // Fields of the fixture packages (made with Anki) and fields as they
  // appear in widely shared decks: Spanish vocabulary written by older
  // Anki versions with one <div> per line, and a geography deck with
  // coloured, nested formatting and lists edited in Chrome.
  group('fields from real decks', () {
    const fixtures = [
      'hola',
      'perro<br><img src="dog.png">',
      '<img src="cat.jpg"> <b>cat</b>',
      '{{c1::Paris::city}} is the capital of {{c2::France}}',
      '[sound:hola.mp3] hola (audio)',
      'adiós [sound:adios.mp3]',
      '{{c1,2::Paris}} is in {{c3::France}}',
    ];
    const spanishVocab = [
      '<div>el perro</div><div><i>(m)</i></div>',
      '<div><b>comer</b>&nbsp;– to eat</div><div><br></div>'
          '<div><span style="color: rgb(128, 128, 128)">yo como, tú comes'
          '</span></div>',
      'la casa<br><span style="color: rgb(0, 0, 255);">house</span>',
      '<b>ser</b> vs <b>estar</b><br><ul><li>permanent</li><li>temporary'
          '</li></ul>',
    ];
    const geography = [
      '<b><span style="color: #c00000">Canberra</span></b>',
      'Capital of <i>Australia</i><br><br>Pop. ~ 460&nbsp;000',
      '<ol><li>Lake Baikal</li><li>Tanganyika</li></ol>',
      '<div style="text-align: center;"><img src="map-au.png" width="300">'
          '</div>',
      'H<sub>2</sub>O &amp; CO<sub>2</sub>',
    ];
    for (final html in [...fixtures, ...spanishVocab, ...geography]) {
      test(html, () {
        expect(isLossless(html), isTrue);
        // Converting the converted HTML again changes nothing.
        final once = roundTrip(html);
        expect(roundTrip(once), once);
      });
    }
  });

  group('lossless guard', () {
    test('a table cannot be kept', () {
      const html = '<table><tr><td>perro</td><td>dog</td></tr></table>';
      expect(isLossless(html), isFalse);
      // Its text is still there when the user switches anyway.
      expect(roundTrip(html), contains('perro'));
    });

    test('a div with a class cannot be kept', () {
      expect(isLossless('<div class="note">x</div>'), isFalse);
    });

    test('a field that round-trips cleanly can be kept', () {
      expect(isLossless('<b>perro</b><br><i>dog</i>'), isTrue);
    });

    test('other markup outside the subset cannot be kept', () {
      for (final html in [
        '<span class="hint">x</span>',
        '<span style="font-size: 20px">x</span>',
        '<img src="a.png" style="max-width: 50%">',
        '<a href="x" target="_blank">x</a>',
        '<!-- comment -->x',
        '<h1>title</h1>',
        '<code>x</code>',
        '<div>a<br>b</div>',
        '<span style="color: hsl(0, 100%, 50%)">x</span>',
        // The underline is black on cards; the visual editor would make it
        // red.
        '<u><span style="color: red">x</span></u>',
      ]) {
        expect(isLossless(html), isFalse, reason: html);
      }
    });
  });

  group('paste', () {
    String pasted(String html) => documentToHtml(
      Document.fromDelta(Delta.from(pasteDelta(html))..insert('\n')).toDelta(),
    );

    test('keeps bold, italic, underline, colours, lists and line breaks', () {
      expect(
        pasted(
          '<b>b</b><i>i</i><u>u</u><span style="color: red">c</span>'
          '<span style="background-color: yellow">h</span><br>x',
        ),
        '<b>b</b><i>i</i><u>u</u><span style="color: red">c</span>'
        '<span style="background-color: yellow">h</span><br>x',
      );
      expect(
        pasted('<ul><li>a<ul><li>b</li></ul></li></ul>'),
        '<ul><li>a<ul><li>b</li></ul></li></ul>',
      );
    });

    test('drops other formatting, images and links', () {
      expect(
        pasted(
          'H<sub>2</sub>O <s>old</s> <a href="x">link</a><img src="y.png">'
          '<div style="text-align: center; margin-left: 40px">c</div>',
        ),
        'H2O old link<br>c',
      );
    });

    test('reads word processor styles and skips hidden content', () {
      expect(
        pasted(
          '<html><head><style>p { color: red }</style><title>t</title>'
          '</head><body><!--StartFragment--><span style="font-weight: 700">'
          'bold</span> <span style="font-style: italic">it</span>'
          '<!--EndFragment--></body></html>',
        ),
        '<b>bold</b> <i>it</i>',
      );
    });
  });

  testWidgets('trailing toolbar buttons stay enabled in HTML mode', (
    tester,
  ) async {
    final field = RichFieldController(html: 'perro')..showHtml();
    addTearDown(field.dispose);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: RichToolbar(
            field: field,
            trailing: [
              IconButton(
                tooltip: 'Tool',
                icon: const Icon(Icons.build),
                onPressed: () {},
              ),
            ],
          ),
        ),
      ),
    );
    IconButton button(String tooltip) => tester.widget<IconButton>(
      find.byWidgetPredicate((w) => w is IconButton && w.tooltip == tooltip),
    );
    expect(button('Bold').onPressed, isNull);
    expect(button('Tool').onPressed, isNotNull);

    field.showVisual();
    await tester.pump();
    expect(button('Bold').onPressed, isNotNull);
    expect(button('Tool').onPressed, isNotNull);
  });
}
