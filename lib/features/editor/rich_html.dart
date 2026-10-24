import 'package:flutter_quill/quill_delta.dart';
import 'package:html/dom.dart' as dom;
import 'package:html/parser.dart' as html_parser;

import 'package:cardflip/widgets/css_color.dart';

/// A field's HTML as a Quill document for the visual editor.
class FieldDocument {
  const FieldDocument(this.delta, {this.divLines = false});

  /// The document: text and images with inline attributes (bold, italic,
  /// underline, strike, script, color, background, link, image width), and
  /// line attributes (list, indent, align) on each `\n`.
  final Delta delta;

  /// Whether the field wrote each line as a `<div>`, as older Anki and
  /// browser editors do, instead of separating lines with `<br>`. Converting
  /// back keeps that style, so untouched lines come out as they went in.
  final bool divLines;
}

/// Pixels of `margin-left` per indent level of a line outside lists.
const indentPixels = 40;

/// Converts field HTML to a document for the visual editor. Markup outside
/// the supported subset is dropped (its text is kept); [isLossless] tells
/// whether that happens.
FieldDocument htmlToDocument(String html) {
  final nodes = html_parser.parseFragment(html).nodes;
  final p = _Parser()..block(nodes, const {});
  p.finish();
  final delta = Delta();
  for (final line in p.lines) {
    for (final run in line.runs) {
      delta.insert(run.insert, run.attrs.isEmpty ? null : run.attrs);
    }
    delta.insert('\n', line.attrs.isEmpty ? null : line.attrs);
  }
  return FieldDocument(delta, divLines: _usesDivLines(nodes));
}

/// Whether every top-level node is a block and at least one is a `<div>`.
bool _usesDivLines(List<dom.Node> nodes) {
  var div = false;
  for (final n in nodes) {
    if (n is dom.Text && n.text.trim().isEmpty) continue;
    if (n is! dom.Element) return false;
    switch (n.localName) {
      case 'div':
        div = true;
      case 'ul' || 'ol':
        break;
      default:
        return false;
    }
  }
  return div;
}

/// Converts a document of the visual editor to field HTML as Anki's editor
/// writes it: `<b>`, `<i>`, `<u>`, `<span style="color: …">`, lines
/// separated by `<br>` (or one `<div>` each when [divLines]), and `<div>`s
/// for aligned or indented lines.
String documentToHtml(Delta delta, {bool divLines = false}) {
  final lines = _lines(delta);
  // A single empty line is an empty field.
  if (lines.length == 1 &&
      lines.first.runs.isEmpty &&
      lines.first.attrs.isEmpty) {
    return '';
  }
  final out = StringBuffer();
  var lastPlain = false;
  final lists = <_OpenList>[];
  void closeLists([int depth = 0]) {
    while (lists.length > depth) {
      out.write('</li></${lists.removeLast().tag}>');
    }
  }

  for (final line in lines) {
    final list = line.attrs['list'];
    final indent = (line.attrs['indent'] as int?) ?? 0;
    final align = line.attrs['align'] as String?;
    if (list == 'bullet' || list == 'ordered') {
      final tag = list == 'bullet' ? 'ul' : 'ol';
      closeLists(indent + 1);
      if (lists.length == indent + 1 && lists.last.tag != tag) {
        closeLists(indent);
      }
      if (lists.length == indent + 1) {
        out.write('</li>');
      }
      while (lists.length < indent + 1) {
        // A list nested deeper than the one before it starts inside an
        // empty item.
        if (lists.isNotEmpty && lists.last.empty) out.write('<li>');
        out.write('<$tag>');
        lists.add(_OpenList(tag));
      }
      lists.last.empty = false;
      out.write(align == null ? '<li>' : '<li style="text-align: $align">');
      out.write(line.runs.isEmpty ? '<br>' : _inlineHtml(line.runs));
      lastPlain = false;
      continue;
    }
    closeLists();
    final styles = [
      if (align != null) 'text-align: $align',
      if (indent > 0) 'margin-left: ${indent * indentPixels}px',
    ];
    final content = _inlineHtml(line.runs);
    if (styles.isNotEmpty || divLines) {
      final style = styles.isEmpty ? '' : ' style="${styles.join('; ')}"';
      out.write('<div$style>${content.isEmpty ? '<br>' : content}</div>');
      lastPlain = false;
    } else {
      if (lastPlain) out.write('<br>');
      out.write(content);
      lastPlain = true;
    }
  }
  closeLists();
  return out.toString();
}

/// Whether the visual editor keeps everything in [html]: converting it to
/// a document and back gives HTML that is the same after normalising
/// whitespace, attribute and style order, synonymous tags (`<strong>` is
/// `<b>`) and the nesting order of inline formatting. Colours must also be
/// ones the editor can show.
bool isLossless(String html) {
  final doc = htmlToDocument(html);
  for (final op in doc.delta.operations) {
    for (final key in const ['color', 'background']) {
      final value = op.attributes?[key];
      if (value is String && parseCssColor(value) == null) return false;
    }
  }
  final back = documentToHtml(doc.delta, divLines: doc.divLines);
  return _listEquals(normaliseHtml(html), normaliseHtml(back));
}

/// Inline attributes pasting keeps.
const _pasteInline = {'bold', 'italic', 'underline', 'color', 'background'};

/// Line attributes pasting keeps: lists with their nesting.
const _pasteLine = {'list', 'indent'};

/// Converts clipboard HTML to what pasting inserts: text with bold,
/// italic, underline, colours, lists and line breaks; other formatting,
/// images and links are dropped. A single line has no line break at its end.
Delta pasteDelta(String html) {
  final ops = <(String, Map<String, dynamic>?)>[];
  for (final op in htmlToDocument(html).delta.operations) {
    final data = op.data;
    if (data is! String) continue;
    final keep = data.contains('\n') ? _pasteLine : _pasteInline;
    final attrs = {
      for (final e in (op.attributes ?? const {}).entries)
        if (keep.contains(e.key)) e.key: e.value,
    };
    if (!attrs.containsKey('list')) attrs.remove('indent');
    ops.add((data, attrs.isEmpty ? null : attrs));
  }
  // Drop the document's last line break, unless it ends a list item.

  if (ops.isNotEmpty && ops.last.$2 == null && ops.last.$1.endsWith('\n')) {
    final text = ops.removeLast().$1;
    if (text.length > 1) ops.add((text.substring(0, text.length - 1), null));
  }
  final out = Delta();
  for (final (data, attrs) in ops) {
    out.insert(data, attrs);
  }
  return out;
}

/// The colours (text and background) a document uses, for the editor's
/// palette.
Set<String> documentColors(Delta delta) => {
  for (final op in delta.operations)
    for (final key in const ['color', 'background'])
      if (op.attributes?[key] case final String c) c,
};

// ---------------------------------------------------------------------------
// HTML → document

class _Run {
  _Run(this.insert, this.attrs);
  final Object insert;
  final Map<String, dynamic> attrs;
}

class _Line {
  _Line(this.attrs);
  final runs = <_Run>[];
  final Map<String, dynamic> attrs;
}

class _OpenList {
  _OpenList(this.tag);
  final String tag;

  /// No item of the list has been written yet.
  bool empty = true;
}

/// Elements whose content is not shown (clipboard HTML from browsers
/// brings some).
const _hiddenTags = {
  'head',
  'title',
  'meta',
  'link',
  'style',
  'script',
  'template',
  'noscript',
};

/// Block elements end the line before and after them.
const _blockTags = {
  'div', 'p', 'ul', 'ol', 'li', 'table', 'thead', 'tbody', 'tfoot', 'tr', //
  'td', 'th', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'pre', 'hr',
  'section', 'article', 'header', 'footer', 'nav', 'aside', 'figure',
  'figcaption', 'dl', 'dt', 'dd', 'address', 'center',
};

class _Parser {
  final lines = <_Line>[];

  /// Runs of the line being read.
  var _runs = <_Run>[];

  /// The line being read was started by a `<br>`, so it exists even when
  /// empty.
  var _afterBreak = false;

  /// Lines emitted since the innermost block started.
  var _blockLines = 0;

  void _emit(Map<String, dynamic> attrs) {
    lines.add(_Line(Map.of(attrs))..runs.addAll(_runs));
    _runs = [];
    _blockLines++;
  }

  /// Ends the line being read at the start or end of a block, if it has
  /// content or a `<br>` started it.
  void _endLine(Map<String, dynamic> attrs) {
    if (_runs.isNotEmpty || _afterBreak) _emit(attrs);
    _afterBreak = false;
  }

  void finish() {
    if (_runs.isNotEmpty || _afterBreak || lines.isEmpty) _emit(const {});
  }

  /// Reads [nodes] in a block whose lines get [attrs].
  void block(List<dom.Node> nodes, Map<String, dynamic> attrs) {
    for (final n in nodes) {
      _node(n, attrs, const {}, nodes);
    }
  }

  void _text(String text, Map<String, dynamic> inline) {
    text = collapseWhitespace(text);
    if (text.isEmpty) return;
    if (_runs.isNotEmpty) {
      final last = _runs.last;
      if (last.insert is String && _mapEquals(last.attrs, inline)) {
        _runs[_runs.length - 1] = _Run('${last.insert}$text', last.attrs);
        return;
      }
    }
    _runs.add(_Run(text, Map.of(inline)));
  }

  void _node(
    dom.Node n,
    Map<String, dynamic> line,
    Map<String, dynamic> inline,
    List<dom.Node> siblings,
  ) {
    if (n is dom.Text) {
      if (n.text.trim().isEmpty && _isBlockGap(n, siblings)) return;
      _text(n.text, inline);
      return;
    }
    if (n is! dom.Element) return;
    final tag = n.localName ?? '';
    if (_hiddenTags.contains(tag)) return;
    switch (tag) {
      case 'br':
        _emit(line);
        _afterBreak = true;
        return;
      case 'img':
        final attrs = <String, dynamic>{...inline};
        if (n.attributes['width'] case final w?) attrs['width'] = w;
        _runs.add(_Run({'image': n.attributes['src'] ?? ''}, attrs));
        return;
      case 'div' || 'p':
        _blockElement(n, line, {...line, ..._blockStyle(n)}, inline);
        return;
      case 'ul' || 'ol':
        _list(n, line, inline, (line['indent'] as int?) ?? 0, tag);
        return;
    }
    if (_blockTags.contains(tag)) {
      // Unsupported block: keep its text on lines of its own.
      _blockElement(n, line, line, inline);
      return;
    }
    final next = {...inline, ..._inlineFormat(n)};
    for (final c in n.nodes) {
      _node(c, line, next, n.nodes);
    }
  }

  /// Reads block [e] whose lines get [line]; the line before it ends with
  /// [outer].
  void _blockElement(
    dom.Element e,
    Map<String, dynamic> outer,
    Map<String, dynamic> line,
    Map<String, dynamic> inline,
  ) {
    _endLine(outer);
    final before = _blockLines;
    _blockLines = 0;
    for (final c in e.nodes) {
      _node(c, line, inline, e.nodes);
    }
    // `<div><br></div>` is one empty line; a block's trailing `<br>` adds
    // none.
    if (_runs.isNotEmpty) _emit(line);
    _afterBreak = false;
    _blockLines = before + 1;
  }

  void _list(
    dom.Element list,
    Map<String, dynamic> line,
    Map<String, dynamic> inline,
    int depth,
    String tag,
  ) {
    _endLine(line);
    final kind = tag == 'ul' ? 'bullet' : 'ordered';
    for (final item in list.nodes) {
      if (item is dom.Text && item.text.trim().isEmpty) continue;
      if (item is dom.Element &&
          (item.localName == 'ul' || item.localName == 'ol')) {
        // A list directly in a list, as some browsers write it.
        _list(item, line, inline, depth + 1, item.localName!);
        continue;
      }
      final attrs = <String, dynamic>{
        ...line,
        'list': kind,
        if (depth > 0) 'indent': depth else 'indent': null,
        if (item is dom.Element) ..._blockStyle(item),
      }..removeWhere((k, v) => v == null);
      if (item is! dom.Element || item.localName != 'li') {
        _node(item, attrs, inline, list.nodes);
        _endLine(attrs);
        continue;
      }
      final before = _blockLines;
      _blockLines = 0;
      for (final c in item.nodes) {
        if (c is dom.Element && (c.localName == 'ul' || c.localName == 'ol')) {
          if (_runs.isNotEmpty) _emit(attrs);
          _list(c, line, inline, depth + 1, c.localName!);
          continue;
        }
        _node(c, attrs, inline, item.nodes);
      }
      // Every item is a line, also an empty one.
      if (_runs.isNotEmpty || _blockLines == 0) _emit(attrs);
      _afterBreak = false;
      _blockLines = before + 1;
    }
    _afterBreak = false;
  }
}

/// Whether whitespace-only text [n] sits between blocks (or at the edge of
/// its parent next to one), where it has no effect.
bool _isBlockGap(dom.Text n, List<dom.Node> siblings) {
  final parent = n.parent;
  final tag = parent?.localName;
  if (tag == 'ul' ||
      tag == 'ol' ||
      tag == 'table' ||
      tag == 'tr' ||
      tag == 'tbody') {
    return true;
  }
  final i = siblings.indexOf(n);
  bool isBlock(int j) =>
      j < 0 ||
      j >= siblings.length ||
      (siblings[j] is dom.Element &&
          _blockTags.contains((siblings[j] as dom.Element).localName));
  return isBlock(i - 1) || isBlock(i + 1);
}

/// Line attributes from a block's `style`.
Map<String, dynamic> _blockStyle(dom.Element e) {
  final style = parseStyle(e.attributes['style']);
  return {
    // Left is kept: card CSS often centres text, so it is not the default.
    // Other values (start, end) are not in the subset: the field opens as
    // HTML instead of losing them.
    if (style['text-align'] case final a?
        when const {'left', 'center', 'right', 'justify'}.contains(a))
      'align': a,
    if (style['margin-left'] case final m?)
      if (RegExp(r'^(\d+)px$').firstMatch(m) case final px?)
        if (int.parse(px[1]!) ~/ indentPixels case final level when level > 0)
          'indent': level,
  };
}

/// Inline attributes an element adds.
Map<String, dynamic> _inlineFormat(dom.Element e) {
  switch (e.localName) {
    case 'b' || 'strong':
      return {'bold': true};
    case 'i' || 'em':
      return {'italic': true};
    case 'u':
      return {'underline': true};
    case 's' || 'strike' || 'del':
      return {'strike': true};
    case 'sub':
      return {'script': 'sub'};
    case 'sup':
      return {'script': 'super'};
    case 'a':
      return {'link': ?e.attributes['href']};
    case 'font':
      return {'color': ?e.attributes['color']};
    case 'span':
      final style = parseStyle(e.attributes['style']);
      final weight = style['font-weight'];
      return {
        'color': ?style['color'],
        'background': ?style['background-color'],
        // As word processors and web pages write formatting.
        if (weight == 'bold' || (int.tryParse(weight ?? '') ?? 0) >= 600)
          'bold': true,
        if (style['font-style'] == 'italic') 'italic': true,
        if (style['text-decoration']?.contains('underline') ?? false)
          'underline': true,
        if (style['vertical-align'] case final v?
            when v == 'sub' || v == 'super')
          'script': v,
      };
  }
  return const {};
}

// ---------------------------------------------------------------------------
// Document → HTML

const _lineKeys = {'list', 'indent', 'align'};

List<_Line> _lines(Delta delta) {
  final lines = <_Line>[];
  var runs = <_Run>[];
  for (final op in delta.operations) {
    if (!op.isInsert) continue;
    final all = op.attributes ?? const <String, dynamic>{};
    final inline = {
      for (final e in all.entries)
        if (!_lineKeys.contains(e.key) && e.value != null) e.key: e.value,
    };
    final data = op.data;
    if (data is! String) {
      runs.add(_Run(data!, inline));
      continue;
    }
    final line = {
      for (final e in all.entries)
        if (_lineKeys.contains(e.key) && e.value != null) e.key: e.value,
    };
    final parts = data.split('\n');
    for (var i = 0; i < parts.length; i++) {
      if (parts[i].isNotEmpty) runs.add(_Run(parts[i], inline));
      if (i < parts.length - 1) {
        lines.add(_Line(line)..runs.addAll(runs));
        runs = [];
      }
    }
  }
  if (runs.isNotEmpty) lines.add(_Line(const {})..runs.addAll(runs));
  if (lines.isEmpty) lines.add(_Line(const {}));
  return lines;
}

/// One inline tag of a run, outermost first: its opening HTML and closing
/// tag name.
typedef _Tag = ({String open, String close});

List<_Tag> _tagsOf(Map<String, dynamic> a) => [
  if (a['link'] case final String href)
    (open: '<a href="${escapeAttribute(href)}">', close: 'a'),
  if (a['color'] case final String c)
    (open: '<span style="color: ${escapeAttribute(c)}">', close: 'span'),
  if (a['background'] case final String c)
    (
      open: '<span style="background-color: ${escapeAttribute(c)}">',
      close: 'span',
    ),
  if (a['bold'] == true) (open: '<b>', close: 'b'),
  if (a['italic'] == true) (open: '<i>', close: 'i'),
  if (a['underline'] == true) (open: '<u>', close: 'u'),
  if (a['strike'] == true) (open: '<s>', close: 's'),
  if (a['script'] == 'sub') (open: '<sub>', close: 'sub'),
  if (a['script'] == 'super') (open: '<sup>', close: 'sup'),
];

/// Inline HTML of runs; tags shared by neighbouring runs stay open.
String _inlineHtml(List<_Run> runs) {
  final out = StringBuffer();
  final open = <_Tag>[];
  for (final run in runs) {
    final tags = _tagsOf(run.attrs);
    var keep = 0;
    while (keep < open.length &&
        keep < tags.length &&
        open[keep] == tags[keep]) {
      keep++;
    }
    while (open.length > keep) {
      out.write('</${open.removeLast().close}>');
    }
    for (final t in tags.skip(keep)) {
      out.write(t.open);
      open.add(t);
    }
    final data = run.insert;
    if (data is String) {
      out.write(escapeText(data));
    } else if (data is Map && data['image'] is String) {
      final width = run.attrs['width'];
      out.write('<img src="${escapeAttribute(data['image'] as String)}"');
      if (width != null) out.write(' width="${escapeAttribute('$width')}"');
      out.write('>');
    }
  }
  while (open.isNotEmpty) {
    out.write('</${open.removeLast().close}>');
  }
  return out.toString();
}

String escapeText(String s) => s
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll(' ', '&nbsp;');

String escapeAttribute(String s) => s
    .replaceAll('&', '&amp;')
    .replaceAll('"', '&quot;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');

// ---------------------------------------------------------------------------
// Normalisation for the lossless check

/// Collapses runs of HTML whitespace (not `&nbsp;`) to one space.
String collapseWhitespace(String s) =>
    s.replaceAll(RegExp(r'[ \t\n\r\f]+'), ' ');

/// Parses a `style` attribute into lower-case properties and their values.
Map<String, String> parseStyle(String? style) {
  final out = <String, String>{};
  for (final decl in (style ?? '').split(';')) {
    final i = decl.indexOf(':');
    if (i < 0) continue;
    final prop = decl.substring(0, i).trim().toLowerCase();
    final value = collapseWhitespace(decl.substring(i + 1).trim());
    if (prop.isNotEmpty && value.isNotEmpty) out[prop] = value;
  }
  return out;
}

String _styleKey(Map<String, String> style) =>
    (style.keys.toList()..sort()).map((k) => '$k: ${style[k]}').join('; ');

String _attrsKey(dom.Element e, {Set<String> skip = const {}}) {
  final keys = [
    for (final k in e.attributes.keys)
      if (!skip.contains('$k')) '$k',
  ]..sort();
  return keys
      .map((k) {
        final v = e.attributes[k]!;
        return k == 'style' ? 'style="${_styleKey(parseStyle(v))}"' : '$k="$v"';
      })
      .join(' ');
}

const _formatSynonyms = {'strong': 'b', 'em': 'i', 'strike': 's', 'del': 's'};

/// Tokens describing what [html] shows: text runs with the set of inline
/// formats around them, and block, void and unknown elements with their
/// attributes. Two fields with the same tokens look the same.
List<String> normaliseHtml(String html) {
  final tokens = <String>[];
  void text(String t, Set<String> formats) {
    final key = (formats.toList()..sort()).join(',');
    final prefix = 'text[$key]:';
    if (tokens.isNotEmpty && tokens.last.startsWith(prefix)) {
      tokens[tokens.length - 1] = tokens.last + t;
    } else {
      tokens.add(prefix + t);
    }
  }

  // [color] is the inherited text colour ('' for the card's). Browsers
  // draw an underline or strike-through in the colour of the element that
  // declares it, so the tokens record it: `<u><span red>` and
  // `<span red><u>` look different.
  void walk(List<dom.Node> nodes, Set<String> formats, String color) {
    for (final n in nodes) {
      if (n is dom.Text) {
        if (n.text.trim().isEmpty && _isBlockGap(n, nodes)) continue;
        final t = collapseWhitespace(n.text);
        if (t.isNotEmpty) text(t, formats);
        continue;
      }
      if (n is dom.Comment) {
        tokens.add('comment:${n.data}');
        continue;
      }
      if (n is! dom.Element) continue;
      final tag = n.localName ?? '';
      if (tag == 'br' ||
          tag == 'img' ||
          tag == 'hr' ||
          tag == 'input' ||
          tag == 'wbr') {
        tokens.add('<$tag ${_attrsKey(n)}>');
        continue;
      }
      if (_blockTags.contains(tag)) {
        tokens.add('<$tag ${_attrsKey(n)}>');
        walk(
          n.nodes,
          formats,
          parseStyle(n.attributes['style'])['color'] ?? color,
        );
        tokens.add('</$tag>');
        continue;
      }
      final next = {...formats};
      var nextColor = color;
      switch (tag) {
        case 'span' || 'font':
          final style = parseStyle(n.attributes['style']);
          if (tag == 'font' && n.attributes['color'] != null) {
            style['color'] = n.attributes['color']!;
          }
          nextColor = style['color'] ?? color;
          for (final e in style.entries) {
            next.add('style:${e.key}=${e.value}');
          }
          final rest = _attrsKey(
            n,
            skip: {'style', if (tag == 'font') 'color'},
          );
          if (rest.isNotEmpty) next.add('$tag[$rest]');
        default:
          final attrs = _attrsKey(n);
          next.add(
            '${_formatSynonyms[tag] ?? tag}${attrs.isEmpty ? '' : '[$attrs]'}',
          );
          nextColor = parseStyle(n.attributes['style'])['color'] ?? color;
          if (const {'u', 'ins', 's', 'strike', 'del'}.contains(tag)) {
            next.add('line-color=$nextColor');
          }
      }
      walk(n.nodes, next, nextColor);
    }
  }

  walk(html_parser.parseFragment(html).nodes, const {}, '');
  return tokens;
}

bool _listEquals(List<String> a, List<String> b) {
  if (a.length != b.length) return false;
  for (var i = 0; i < a.length; i++) {
    if (a[i] != b[i]) return false;
  }
  return true;
}

bool _mapEquals(Map<String, dynamic> a, Map<String, dynamic> b) {
  if (a.length != b.length) return false;
  for (final e in a.entries) {
    if (b[e.key] != e.value) return false;
  }
  return true;
}
