import 'dart:async';
import 'dart:ui' as ui;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_math_fork/flutter_math.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:flutter_widget_from_html_core/flutter_widget_from_html_core.dart';
import 'package:html/dom.dart' as dom;
import 'package:html/parser.dart' as html_parser;

import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/l10n/l10n.dart';

import 'css_color.dart';
import 'font_cache.dart';
import 'media_cache.dart';
import 'media_player.dart';
import 'occlusion_view.dart';
import 'speech.dart';

/// Shows one side of a card: the sanitized HTML from Go, with the note
/// type's styles already inlined. The card surface uses the note type's
/// background and text colour when it sets them and the app theme
/// otherwise, so cards stay readable in light and dark mode. Scripts never
/// run; images, also CSS background images, come from the collection's
/// media, and remote ones are never loaded. The `<cf-sound>`,
/// `<cf-video>` and `<cf-tts>` placeholders become replay buttons and
/// inline video players for the side's [sounds], played by [player], and
/// `<cf-tts-voices>` the list of the device's voices. `<cf-math>` holds
/// the TeX of a formula, typeset inline or as a centred block. The
/// `<cf-typeans>` placeholder of `{{type:…}}` becomes the [typeAnswer]
/// input, and nothing when there is none (previews). `<cf-occlusion>` shows
/// an image occlusion note's image with its shapes drawn over it. The view
/// reports its
/// size, so `@media` width rules match the space the card has, and, to
/// [onAnswerAnchor], where the element with id `answer` (the divider of
/// `<hr id=answer>`) is. The note type's [fonts] from the media are loaded
/// before the card is shown, for at most a second, and its `font-family`
/// declarations use them; families that cannot be loaded fall back to the
/// next listed one.
class CardView extends StatelessWidget {
  const CardView({
    super.key,
    required this.html,
    this.background,
    this.color,
    this.media,
    this.fonts = const [],
    this.fontCache,
    this.sounds = const [],
    this.player,
    this.showPlayButtons = true,
    this.typeAnswer,
    this.onAnswerAnchor,
    this.padding = const EdgeInsets.all(24),
  });

  final String html;
  final String? background;
  final String? color;
  final MediaCache? media;

  /// The note type's `@font-face` fonts, loaded through [fontCache]
  /// ([FontCache.instance] when null).
  final List<golang.FontFace> fonts;
  final FontCache? fontCache;

  /// The side's sounds, videos and utterances; placeholders find theirs by
  /// kind and index.
  final List<MediaItem> sounds;
  final MediaPlayer? player;

  /// Off with the setting "Show play buttons on cards with audio".
  final bool showPlayButtons;
  final TypeAnswerField? typeAnswer;

  /// Called once the element with id `answer` is laid out, with a context
  /// at its top, e.g. to scroll it into view; never when there is none.
  final ValueChanged<BuildContext>? onAnswerAnchor;
  final EdgeInsets padding;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final bg = parseCssColor(background) ?? scheme.surface;
    final fg = parseCssColor(color) ?? scheme.onSurface;
    final highlight = scheme.primary;
    final cache = media ?? MediaCache.instance;
    var shown = _readableTypeAnswer(html, bg, fg);
    if (onAnswerAnchor != null) shown = _withAnswerAnchor(shown);
    final families = _usedFonts(shown);
    return LayoutBuilder(
      builder: (context, constraints) {
        _reportSize(context, constraints);
        if (families.isEmpty) return _surface(shown, bg, fg, highlight, cache);
        return _FontGate(
          families: families,
          cache: fontCache ?? FontCache.instance,
          waiting: Container(
            key: const Key('card-surface'),
            color: bg,
            padding: padding,
          ),
          builder: (loaded) =>
              _surface(_withFonts(shown, loaded), bg, fg, highlight, cache),
        );
      },
    );
  }

  /// The files of each font family (in lower case) that [html] may use.
  Map<String, List<String>> _usedFonts(String html) {
    if (fonts.isEmpty) return const {};
    final lower = html.toLowerCase();
    final used = <String, List<String>>{};
    for (final f in fonts) {
      final family = _familyName(f.family);
      if (family.isNotEmpty && lower.contains(family)) {
        (used[family] ??= []).add(f.file);
      }
    }
    return used;
  }

  /// A CSS font family name as it is compared: unquoted, single spaced,
  /// lower case.
  static String _familyName(String name) {
    var n = name.trim();
    if (n.length >= 2 &&
        (n[0] == '"' || n[0] == "'") &&
        n[n.length - 1] == n[0]) {
      n = n.substring(1, n.length - 1);
    }
    return n
        .split(RegExp(r'\s+'))
        .where((w) => w.isNotEmpty)
        .join(' ')
        .toLowerCase();
  }

  static final _fontFamilyDecl = RegExp(
    r'(font-family\s*:)([^;]*)',
    caseSensitive: false,
  );

  /// [html] with the font families in [loaded] (CSS name in lower case to
  /// registered name) replaced in `font-family` styles and `<font face>`.
  static String _withFonts(String html, Map<String, String> loaded) {
    if (loaded.isEmpty) return html;
    String families(String list) =>
        list.split(',').map((f) => loaded[_familyName(f)] ?? f).join(',');
    final root = html_parser.parseFragment(html);
    var changed = false;
    for (final e in root.querySelectorAll('*')) {
      final style = e.attributes['style'];
      if (style != null && style.toLowerCase().contains('font-family')) {
        final mapped = style.replaceAllMapped(
          _fontFamilyDecl,
          (m) => '${m[1]}${families(m[2]!)}',
        );
        if (mapped != style) {
          e.attributes['style'] = mapped;
          changed = true;
        }
      }
      final face = e.localName == 'font' ? e.attributes['face'] : null;
      if (face != null && families(face) != face) {
        e.attributes['face'] = families(face);
        changed = true;
      }
    }
    return changed ? root.outerHtml : html;
  }

  /// Tells Go the card's display size; in a scroll view the height is the
  /// screen's.
  static void _reportSize(BuildContext context, BoxConstraints constraints) {
    final env = AppScope.maybeRead(context)?.renderEnv;
    if (env == null) return;
    final size = Size(
      constraints.maxWidth,
      constraints.hasBoundedHeight
          ? constraints.maxHeight
          : MediaQuery.sizeOf(context).height,
    );
    WidgetsBinding.instance.addPostFrameCallback((_) => env.setCardSize(size));
  }

  Widget _surface(
    String html,
    Color bg,
    Color fg,
    Color highlight,
    MediaCache cache,
  ) {
    return Container(
      key: const Key('card-surface'),
      color: bg,
      padding: padding,
      alignment: Alignment.topCenter,
      child: HtmlWidget(
        html,
        // The factory is made once per widget, with the cache.
        key: ObjectKey(cache),
        factoryBuilder: () => _CardWidgetFactory(cache),
        textStyle: TextStyle(color: fg, fontSize: 20, height: 1.4),
        customStylesBuilder: (element) {
          final styles = <String, String>{};
          // Cloze highlights follow the theme unless the CSS coloured them.
          if (_isThemedCloze(element)) {
            styles['color'] = _hex(highlight);
            styles['font-weight'] = 'bold';
          }
          if (_drawsLine(element)) {
            styles['text-decoration-color'] = _hex(
              _textColor(element, fg, highlight),
            );
          }
          // dir values are case-insensitive in HTML, not in the HTML
          // widget; a direction in the style attribute still wins.
          if (element.attributes['dir']?.trim().toLowerCase() case final dir?
              when dir == 'rtl' || dir == 'ltr') {
            styles['direction'] = dir;
          }
          return styles.isEmpty ? null : styles;
        },
        customWidgetBuilder: (element) => switch (element.localName) {
          'img' => _image(element, cache),
          'cf-sound' || 'cf-video' || 'cf-tts' => _media(element),
          'cf-tts-voices' => VoiceList(speech: player?.speech),
          'cf-math' => _math(element, fg, highlight),
          'cf-typeans' => Builder(builder: _typeAnswer),
          'cf-occlusion' => _occlusion(element, cache),
          'cf-answer-anchor' => _answerAnchor(element),
          _ => null,
        },
        // Links are shown, not followed: cards never trigger network access.
        onTapUrl: (_) => true,
      ),
    );
  }

  /// An image, inline as in a browser, at the size its width and height
  /// attributes or style give (pixels or a percentage of the card width),
  /// never wider than the card.
  static Widget _image(dom.Element element, MediaCache cache) {
    final src = (element.attributes['src'] ?? '').trim();
    final alt = element.attributes['alt'];
    final Widget image;
    if (src.startsWith('data:')) {
      final data = Uri.tryParse(src)?.data;
      if (data == null || !data.mimeType.startsWith('image/')) {
        return _placeholder(alt);
      }
      final bytes = data.contentAsBytes();
      image = _sized(
        element,
        (width, height, fit) => Image.memory(
          bytes,
          width: width,
          height: height,
          fit: fit,
          errorBuilder: (_, _, _) => _placeholder(alt),
        ),
      );
    } else if (src.isEmpty || src.contains('://')) {
      return _placeholder(alt);
    } else {
      image = _sized(
        element,
        (width, height, fit) => MediaImage(
          name: src,
          cache: cache,
          alt: alt,
          width: width,
          height: height,
          fit: fit,
        ),
      );
    }
    // Alone in its block, the image is given the block's full width; it
    // keeps its own size there and starts at the edge, as in a browser
    // (text-align makes such an image inline, so it is aligned anyway).
    return InlineCustomWidget(
      child: Align(
        alignment: AlignmentDirectional.topStart,
        widthFactor: 1,
        heightFactor: 1,
        child: image,
      ),
    );
  }

  /// Sizes the image [build] makes as the element's width, height,
  /// max-width and max-height say (the note type's CSS is inlined). As
  /// with Anki's default `max-width: 100%`, the image may be as wide as the
  /// card and a larger size is scaled down keeping its proportions. A
  /// max-width of `none` or in pixels replaces that limit: a wider image
  /// then scrolls sideways.
  static Widget _sized(
    dom.Element element,
    Widget Function(double? width, double? height, BoxFit fit) build,
  ) {
    final width = _length(element, 'width');
    final height = _length(element, 'height');
    final maxWidth = _length(element, 'max-width', attribute: false);
    final maxHeight = _length(element, 'max-height', attribute: false);
    // A percentage height depends on a container height cards don't have.
    final h = height != null && !height.percent ? height.value : null;
    Widget image;
    if (width != null && width.percent) {
      image = FractionallySizedBox(
        widthFactor: (width.value / 100).clamp(0, 1),
        child: build(null, h, h == null ? BoxFit.contain : BoxFit.fill),
      );
    } else {
      final w = width?.value;
      image = w == null && h == null
          ? build(null, null, BoxFit.scaleDown)
          // Both given: stretched to them, as browsers do.
          : FittedBox(
              fit: BoxFit.scaleDown,
              child: build(
                w,
                h,
                w != null && h != null ? BoxFit.fill : BoxFit.contain,
              ),
            );
    }
    final maxW = maxWidth != null && !maxWidth.percent ? maxWidth.value : null;
    final maxH = maxHeight != null && !maxHeight.percent
        ? maxHeight.value
        : null;
    if (maxW != null || maxH != null) {
      image = ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: maxW ?? double.infinity,
          maxHeight: maxH ?? double.infinity,
        ),
        child: image,
      );
    }
    if (maxWidth != null && maxWidth.percent) {
      image = _MaxWidthFactor(
        factor: (maxWidth.value / 100).clamp(0, 1),
        child: image,
      );
    }
    final overflows =
        !(width?.percent ?? false) &&
        (maxW != null || _style(element, 'max-width')?.trim() == 'none');
    if (!overflows) return image;
    return SingleChildScrollView(
      key: const Key('image-scroll'),
      scrollDirection: Axis.horizontal,
      child: image,
    );
  }

  static final _cssLength = RegExp(r'^\s*([\d.]+)\s*(px|%)?\s*$');

  /// The value of the element's style property [name] (in lower case), the
  /// last one when it is set twice.
  static String? _style(dom.Element element, String name) {
    String? value;
    for (final decl in (element.attributes['style'] ?? '').split(';')) {
      final colon = decl.indexOf(':');
      if (colon > 0 && decl.substring(0, colon).trim().toLowerCase() == name) {
        value = decl.substring(colon + 1).toLowerCase();
      }
    }
    return value;
  }

  /// The element's [name] (width, height or their limits), from its style
  /// or else, when [attribute], its attribute: a number of pixels or a
  /// percentage. Null when absent or in another unit.
  static ({double value, bool percent})? _length(
    dom.Element element,
    String name, {
    bool attribute = true,
  }) {
    final raw =
        _style(element, name) ??
        (attribute ? element.attributes[name]?.toLowerCase() : null);
    final m = raw == null ? null : _cssLength.firstMatch(raw);
    final value = double.tryParse(m?[1] ?? '');
    if (value == null || value <= 0) return null;
    return (value: value, percent: m![2] == '%');
  }

  Widget _media(dom.Element element) {
    final idx = int.tryParse(element.attributes['idx'] ?? '');
    final tts = element.localName == 'cf-tts';
    final player = this.player;
    final item = sounds
        .where((s) => s.idx == idx && s.isTts == tts)
        .firstOrNull;
    if (player == null || item == null) return const SizedBox.shrink();
    if (item.video) return InlineVideo(item: item, player: player);
    if (!showPlayButtons) return const SizedBox.shrink();
    return InlineCustomWidget(
      child: tts
          ? TtsButton(item: item, player: player)
          : SoundButton(item: item, player: player),
    );
  }

  /// The image and shapes of an image occlusion card side, in the colours
  /// of the note type's CSS variables, which Go inlined as its style.
  static Widget _occlusion(dom.Element element, MediaCache cache) {
    final image = (element.attributes['image'] ?? '').trim();
    return OcclusionView(
      image: image,
      shapes: OcclusionShape.listFromJson(element.attributes['shapes']),
      active: int.tryParse(element.attributes['active'] ?? '') ?? 0,
      question: element.attributes['side'] != 'a',
      colors: OcclusionColors.fromStyle(element.attributes['style']),
      cache: cache,
      missing: _placeholder(image.isEmpty ? null : image),
    );
  }

  Widget _typeAnswer(BuildContext context) {
    final field = typeAnswer;
    if (field == null) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: TextField(
        key: const Key('type-answer'),
        controller: field.controller,
        focusNode: field.focusNode,
        decoration: InputDecoration(
          hintText: context.l10n.typeTheAnswer,
          border: OutlineInputBorder(),
          isDense: true,
        ),
        style: const TextStyle(fontSize: 18),
        textAlign: TextAlign.center,
        // Spelling is what is being practised.
        autocorrect: false,
        enableSuggestions: false,
        textInputAction: TextInputAction.done,
        // Keep the focus; the submit callback decides what Enter does.
        onEditingComplete: () {},
        onSubmitted: (_) => field.onSubmitted?.call(),
      ),
    );
  }

  /// A formula from Go. Revealed clozes inside it are marked `\cfcloze{…}`
  /// and get the theme's highlight unless the note type's CSS coloured them.
  /// A formula that cannot be typeset shows its source in the error colour.
  static Widget _math(dom.Element element, Color fg, Color highlight) {
    final block = element.attributes['display'] == 'block';
    final tex = element.text.replaceAll(
      r'\cfcloze{',
      '\\color{${_hex(highlight)}}{',
    );
    final math = Math.tex(
      tex,
      mathStyle: block ? MathStyle.display : MathStyle.text,
      textStyle: TextStyle(color: fg, fontSize: 20),
      onErrorFallback: (_) => Builder(
        builder: (context) => Text(
          element.text,
          key: const Key('math-error'),
          style: TextStyle(
            color: Theme.of(context).colorScheme.error,
            fontFamily: 'monospace',
            fontSize: 16,
          ),
        ),
      ),
    );
    if (!block) return InlineCustomWidget(child: math);
    // Wide formulas scroll inside their own box instead of overflowing.
    return Center(
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: math,
        ),
      ),
    );
  }

  static bool _isThemedCloze(dom.Element e) =>
      e.classes.contains('cloze') &&
      !(e.attributes['style'] ?? '').contains('color');

  /// Whether [e] underlines or strikes through its text, without saying in
  /// which colour.
  static bool _drawsLine(dom.Element e) {
    final style = (e.attributes['style'] ?? '').toLowerCase();
    if (style.contains('text-decoration-color')) return false;
    return switch (e.localName) {
          'u' || 'ins' || 's' || 'strike' || 'del' => true,
          _ => false,
        } ||
        RegExp(
          r'text-decoration(-line)?\s*:[^;]*(underline|line-through|overline)',
        ).hasMatch(style);
  }

  /// The text colour of [e], from its own or its ancestors' style (the note
  /// type's CSS is inlined), else the card's. A browser draws an element's
  /// underline in this colour, also below descendants of another colour.
  static Color _textColor(dom.Element? e, Color fg, Color highlight) {
    for (; e != null; e = e.parent) {
      final style = e.attributes['style'] ?? '';
      for (final decl in style.split(';').reversed) {
        final i = decl.indexOf(':');
        if (i < 0 || decl.substring(0, i).trim().toLowerCase() != 'color') {
          continue;
        }
        final c = parseCssColor(decl.substring(i + 1));
        if (c != null) return c;
      }
      if (e.localName == 'font') {
        final c = parseCssColor(e.attributes['color']);
        if (c != null) return c;
      }
      if (_isThemedCloze(e)) return highlight;
    }
    return fg;
  }

  /// [html] with the type-answer comparison (`code#typeans`) made readable:
  /// an element whose text colour has less than 3:1 contrast with the
  /// background it is drawn on gets the card colour [fg], or black or white
  /// when [fg] is too faint there as well. Note types often colour the marks
  /// for a light card or for a highlight the widget can't parse. Inline
  /// styles win over [HtmlWidget.customStylesBuilder], so the colour is
  /// rewritten in the HTML. Other HTML is returned as it is.
  /// Puts a `<cf-answer-anchor>` right before the element with id
  /// `answer`, which marks where it starts. Before an inline element it is
  /// inline too, so the line does not break.
  static String _withAnswerAnchor(String html) {
    if (!html.contains('answer')) return html;
    final root = html_parser.parseFragment(html);
    final target = root.querySelector('#answer');
    final parent = target?.parentNode;
    if (target == null || parent == null) return html;
    final anchor = dom.Element.tag('cf-answer-anchor');
    if (!_blockTags.contains(target.localName)) {
      anchor.attributes['inline'] = '';
    }
    parent.insertBefore(anchor, target);
    return root.outerHtml;
  }

  static const _blockTags = {
    'address',
    'article',
    'aside',
    'blockquote',
    'details',
    'div',
    'dl',
    'fieldset',
    'figure',
    'footer',
    'form',
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    'header',
    'hr',
    'li',
    'main',
    'nav',
    'ol',
    'p',
    'pre',
    'section',
    'table',
    'ul',
  };

  Widget _answerAnchor(dom.Element element) {
    final anchor = _AnswerAnchor(onLaidOut: onAnswerAnchor!);
    return element.attributes.containsKey('inline')
        ? InlineCustomWidget(child: anchor)
        : anchor;
  }

  static String _readableTypeAnswer(String html, Color bg, Color fg) {
    if (!html.contains('typeans')) return html;
    final root = html_parser.parseFragment(html);
    var changed = false;
    for (final block in root.querySelectorAll('code#typeans')) {
      // Document order: an element is fixed before its descendants, which
      // then inherit the fixed colour.
      for (final e in [block, ...block.querySelectorAll('*')]) {
        final back = _backgroundColor(e, bg);
        final text = Color.alphaBlend(_textColor(e, fg, fg), back);
        if (contrastRatio(text, back) >= 3) continue;
        final Color readable;
        if (contrastRatio(fg, back) >= 3) {
          readable = fg;
        } else {
          readable =
              contrastRatio(Colors.black, back) >=
                  contrastRatio(Colors.white, back)
              ? Colors.black
              : Colors.white;
        }
        final style = (e.attributes['style'] ?? '').trim();
        e.attributes['style'] = [
          if (style.isNotEmpty) style.endsWith(';') ? style : '$style;',
          'color:${_hex(readable)}',
        ].join();
        changed = true;
      }
    }
    return changed ? root.outerHtml : html;
  }

  /// The colour [e] is drawn on: its own or its ancestors' background
  /// colour, composited over [bg] where it is semi-transparent.
  static Color _backgroundColor(dom.Element? e, Color bg) {
    final layers = <Color>[];
    for (; e != null; e = e.parent) {
      final c = _ownBackground(e);
      if (c == null) continue;
      layers.add(c);
      if (c.a == 1) break;
    }
    return layers.reversed.fold(bg, (under, c) => Color.alphaBlend(c, under));
  }

  /// The background colour [e]'s style sets, if it parses and is not fully
  /// transparent. An unparseable value is dropped when drawing, too.
  static Color? _ownBackground(dom.Element e) {
    final style = e.attributes['style'] ?? '';
    for (final decl in style.split(';').reversed) {
      final i = decl.indexOf(':');
      if (i < 0) continue;
      final name = decl.substring(0, i).trim().toLowerCase();
      if (name != 'background-color' && name != 'background') continue;
      final c = parseCssColor(decl.substring(i + 1));
      if (c != null && c.a > 0) return c;
    }
    return null;
  }

  static String _hex(Color c) =>
      '#${(c.toARGB32() & 0xFFFFFF).toRadixString(16).padLeft(6, '0')}';
}

/// Shows [waiting] until the font [families] are loaded, for at most a
/// second, and then [builder] with the families loaded so far; fonts that
/// arrive later show the card again with them.
class _FontGate extends StatefulWidget {
  const _FontGate({
    required this.families,
    required this.cache,
    required this.waiting,
    required this.builder,
  });

  /// The files of each CSS font family, by its name in lower case.
  final Map<String, List<String>> families;
  final FontCache cache;
  final Widget waiting;
  final Widget Function(Map<String, String> loaded) builder;

  @override
  State<_FontGate> createState() => _FontGateState();
}

class _FontGateState extends State<_FontGate> {
  static const _timeout = Duration(seconds: 1);

  // CSS family (lower case) to the name it is registered under.
  var _loaded = <String, String>{};
  var _ready = false;
  Timer? _timer;
  // Bumped when the families change, so late results of the old ones are
  // dropped.
  var _run = 0;

  @override
  void initState() {
    super.initState();
    _start();
  }

  @override
  void didUpdateWidget(_FontGate old) {
    super.didUpdateWidget(old);
    if (old.cache != widget.cache ||
        _signature(old.families) != _signature(widget.families)) {
      _start();
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  static String _signature(Map<String, List<String>> families) => [
    for (final MapEntry(:key, :value) in families.entries)
      '$key\n${value.join('\n')}',
  ].join('\n\n');

  void _start() {
    final run = ++_run;
    _timer?.cancel();
    _loaded = {};
    var pending = 0;
    for (final MapEntry(key: family, value: files) in widget.families.entries) {
      final hit = widget.cache.loaded(files);
      if (hit != null) {
        _loaded[family] = hit;
        continue;
      }
      pending++;
      widget.cache.family(files).then((name) {
        if (!mounted || run != _run) return;
        setState(() {
          if (name != null) _loaded = {..._loaded, family: name};
          if (--pending == 0) {
            _ready = true;
            _timer?.cancel();
          }
        });
      });
    }
    _ready = pending == 0;
    if (!_ready) {
      _timer = Timer(_timeout, () {
        if (mounted && run == _run) setState(() => _ready = true);
      });
    }
  }

  @override
  Widget build(BuildContext context) =>
      _ready ? widget.builder(_loaded) : widget.waiting;
}

/// The input a question's `<cf-typeans>` placeholder becomes. The caller
/// owns the controller and focus node, so the typed text outlives
/// rebuilds; [onSubmitted] runs on Enter.
class TypeAnswerField {
  const TypeAnswerField({
    required this.controller,
    required this.focusNode,
    this.onSubmitted,
  });

  final TextEditingController controller;
  final FocusNode focusNode;
  final VoidCallback? onSubmitted;
}

Widget _placeholder(String? alt) => _MissingImage(alt: alt);

/// The HTML widget's factory for cards: `background-image` shows files
/// from the collection's media, and remote and local file URLs are never
/// loaded.
class _CardWidgetFactory extends WidgetFactory {
  _CardWidgetFactory(this.cache);

  final MediaCache cache;

  static final _scheme = RegExp(r'^[a-z][a-z0-9+.-]*:', caseSensitive: false);

  @override
  DecorationImage? buildDecorationImage(
    BuildTree tree,
    String? url, {
    AlignmentGeometry alignment = Alignment.topLeft,
    BoxFit fit = BoxFit.scaleDown,
    ImageRepeat repeat = ImageRepeat.noRepeat,
  }) {
    final name = url?.trim() ?? '';
    if (name.startsWith('data:image/')) {
      return super.buildDecorationImage(
        tree,
        name,
        alignment: alignment,
        fit: fit,
        repeat: repeat,
      );
    }
    if (name.isEmpty || name.startsWith('//') || _scheme.hasMatch(name)) {
      return null;
    }
    return DecorationImage(
      image: _MediaImageProvider(name, cache),
      alignment: alignment,
      fit: fit,
      repeat: repeat,
      // A missing file leaves the background empty.
      onError: (_, _) {},
    );
  }

  @override
  ImageProvider? imageProviderFromNetwork(String url) => null;

  @override
  ImageProvider? imageProviderFromFileUri(String url) => null;
}

/// An image from the collection's media, for decorations.
@immutable
class _MediaImageProvider extends ImageProvider<_MediaImageProvider> {
  const _MediaImageProvider(this.name, this.cache);

  final String name;
  final MediaCache cache;

  @override
  Future<_MediaImageProvider> obtainKey(ImageConfiguration configuration) =>
      SynchronousFuture(this);

  @override
  ImageStreamCompleter loadImage(
    _MediaImageProvider key,
    ImageDecoderCallback decode,
  ) => MultiFrameImageStreamCompleter(
    codec: _codec(decode),
    scale: 1,
    debugLabel: name,
  );

  Future<ui.Codec> _codec(ImageDecoderCallback decode) async => decode(
    await ui.ImmutableBuffer.fromUint8List(await loadMediaBytes(cache, name)),
  );

  @override
  bool operator ==(Object other) =>
      other is _MediaImageProvider &&
      other.name == name &&
      identical(other.cache, cache);

  @override
  int get hashCode => Object.hash(name, identityHashCode(cache));
}

/// Limits its child's width to a [factor] of the width it is given, for a
/// percentage max-width.
class _MaxWidthFactor extends SingleChildRenderObjectWidget {
  const _MaxWidthFactor({required this.factor, super.child});

  final double factor;

  @override
  RenderObject createRenderObject(BuildContext context) =>
      _RenderMaxWidthFactor(factor);

  @override
  void updateRenderObject(
    BuildContext context,
    _RenderMaxWidthFactor renderObject,
  ) => renderObject.factor = factor;
}

class _RenderMaxWidthFactor extends RenderProxyBox {
  _RenderMaxWidthFactor(this._factor);

  double _factor;
  set factor(double value) {
    if (value == _factor) return;
    _factor = value;
    markNeedsLayout();
  }

  BoxConstraints _limit(BoxConstraints c) => c.hasBoundedWidth
      ? c.copyWith(
          minWidth: c.minWidth.clamp(0, c.maxWidth * _factor),
          maxWidth: c.maxWidth * _factor,
        )
      : c;

  @override
  Size computeDryLayout(BoxConstraints constraints) =>
      child?.getDryLayout(_limit(constraints)) ?? constraints.smallest;

  @override
  void performLayout() {
    final child = this.child;
    if (child == null) {
      size = constraints.smallest;
      return;
    }
    child.layout(_limit(constraints), parentUsesSize: true);
    size = constraints.constrain(child.size);
  }
}

/// An image from the collection's media, or a placeholder when it is missing.
/// SVG files, such as Anki's LaTeX images, are drawn without running scripts
/// or loading anything they refer to.
class MediaImage extends StatefulWidget {
  const MediaImage({
    super.key,
    required this.name,
    required this.cache,
    this.alt,
    this.width,
    this.height,
    this.fit = BoxFit.scaleDown,
  });

  final String name;
  final MediaCache cache;
  final String? alt;

  /// The size to show the image at; its own size when null.
  final double? width;
  final double? height;
  final BoxFit fit;

  @override
  State<MediaImage> createState() => _MediaImageState();
}

class _MediaImageState extends State<MediaImage> {
  late Future<Uint8List> _bytes = _load();

  Future<Uint8List> _load() => loadMediaBytes(widget.cache, widget.name);

  @override
  void didUpdateWidget(MediaImage old) {
    super.didUpdateWidget(old);
    if (old.name != widget.name) _bytes = _load();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<Uint8List>(
      future: _bytes,
      builder: (context, snap) {
        if (snap.hasError) return _MissingImage(alt: widget.alt ?? widget.name);
        if (!snap.hasData) {
          return const SizedBox(width: 48, height: 48);
        }
        if (widget.name.toLowerCase().endsWith('.svg')) {
          return SvgPicture.memory(
            snap.data!,
            // A new file gets a new picture: a reused one would keep showing
            // the previous file when the new one fails to load.
            key: ObjectKey(snap.data),
            width: widget.width,
            height: widget.height,
            fit: widget.fit,
            semanticsLabel: widget.alt,
            errorBuilder: (_, _, _) =>
                _MissingImage(alt: widget.alt ?? widget.name),
          );
        }
        return Image.memory(
          snap.data!,
          width: widget.width,
          height: widget.height,
          fit: widget.fit,
          semanticLabel: widget.alt,
          errorBuilder: (_, _, _) =>
              _MissingImage(alt: widget.alt ?? widget.name),
        );
      },
    );
  }
}

class _MissingImage extends StatelessWidget {
  const _MissingImage({this.alt});

  final String? alt;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Semantics(
      label: alt == null
          ? context.l10n.missingImage
          : context.l10n.missingImageNamed(alt!),
      child: Container(
        key: const Key('missing-image'),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          border: Border.all(color: scheme.outlineVariant),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.broken_image_outlined, color: scheme.outline),
            if (alt != null) ...[
              const SizedBox(width: 8),
              Flexible(
                child: Text(
                  alt!,
                  style: TextStyle(color: scheme.outline, fontSize: 14),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Plays one sound of a card, or says why it cannot be played.
class SoundButton extends StatelessWidget {
  const SoundButton({super.key, required this.item, required this.player});

  final MediaItem item;
  final MediaPlayer player;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: player,
      builder: (context, _) => FutureBuilder<String?>(
        future: player.check(item.file),
        builder: (context, snap) {
          final problem = player.problem(item.file) ?? snap.data;
          if (problem != null) return _Unplayable(item: item, why: problem);
          if (player.current == item) {
            return IconButton(
              key: Key('sound-${item.side.name}-${item.idx}'),
              tooltip: context.l10n.stop,
              icon: const Icon(Icons.stop_circle_outlined),
              onPressed: player.stop,
            );
          }
          return IconButton(
            key: Key('sound-${item.side.name}-${item.idx}'),
            tooltip: context.l10n.playFile(item.file),
            icon: const Icon(Icons.play_circle_outline),
            onPressed: snap.connectionState == ConnectionState.done
                ? () => player.play([item])
                : null,
          );
        },
      ),
    );
  }
}

/// Speaks one utterance of a card, or, when the device has no voice for its
/// language, says so.
class TtsButton extends StatefulWidget {
  const TtsButton({super.key, required this.item, required this.player});

  final MediaItem item;
  final MediaPlayer player;

  @override
  State<TtsButton> createState() => _TtsButtonState();
}

class _TtsButtonState extends State<TtsButton> {
  late Future<Voice?> _voice = _find();

  Future<Voice?> _find() =>
      widget.player.speech.voiceFor(widget.item.utterance!);

  @override
  void didUpdateWidget(TtsButton old) {
    super.didUpdateWidget(old);
    if (old.item != widget.item || old.player != widget.player) {
      _voice = _find();
    }
  }

  @override
  Widget build(BuildContext context) {
    final item = widget.item;
    final player = widget.player;
    final key = Key('tts-${item.side.name}-${item.idx}');
    return ListenableBuilder(
      listenable: player,
      builder: (context, _) => FutureBuilder<Voice?>(
        future: _voice,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.done &&
              snap.data == null) {
            return IconButton(
              key: key,
              tooltip: noVoiceMessage(item.utterance!.lang),
              icon: const Icon(Icons.voice_over_off),
              onPressed: null,
            );
          }
          if (player.current == item) {
            return IconButton(
              key: key,
              tooltip: context.l10n.stop,
              icon: const Icon(Icons.stop_circle_outlined),
              onPressed: player.stop,
            );
          }
          return IconButton(
            key: key,
            tooltip: context.l10n.speak,
            icon: const Icon(Icons.record_voice_over_outlined),
            onPressed: snap.connectionState == ConnectionState.done
                ? () => player.play([item])
                : null,
          );
        },
      ),
    );
  }
}

/// The device's speech voices with their languages, for `{{tts-voices:}}`.
class VoiceList extends StatefulWidget {
  const VoiceList({super.key, this.speech});

  /// Null for the app's speech.
  final Speech? speech;

  @override
  State<VoiceList> createState() => _VoiceListState();
}

class _VoiceListState extends State<VoiceList> {
  late final Future<List<Voice>> _voices = (widget.speech ?? Speech.shared)
      .voices();

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Voice>>(
      future: _voices,
      builder: (context, snap) {
        final voices = snap.data;
        if (voices == null) return const SizedBox.shrink();
        return Column(
          key: const Key('tts-voices'),
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (voices.isEmpty) Text(context.l10n.noVoicesInstalled),
            for (final v in voices) Text('${v.lang} ${v.name}'),
          ],
        );
      },
    );
  }
}

/// A video of a card, played inline with its own controls.
class InlineVideo extends StatelessWidget {
  const InlineVideo({super.key, required this.item, required this.player});

  final MediaItem item;
  final MediaPlayer player;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: player,
      builder: (context, _) => FutureBuilder<String?>(
        future: player.check(item.file),
        builder: (context, snap) {
          final problem = player.problem(item.file) ?? snap.data;
          if (problem != null) return _Unplayable(item: item, why: problem);
          final playing = player.current == item;
          return Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 480),
              child: AspectRatio(
                key: Key('video-${item.side.name}-${item.idx}'),
                aspectRatio: 16 / 9,
                child: ColoredBox(
                  color: Colors.black,
                  child: playing
                      ? player.backend.videoView()
                      : Center(
                          child: IconButton(
                            tooltip: context.l10n.playFile(item.file),
                            iconSize: 48,
                            color: Colors.white,
                            icon: const Icon(Icons.play_arrow),
                            onPressed: () => player.play([item]),
                          ),
                        ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

/// The disabled control of a sound or video that cannot be played.
class _Unplayable extends StatelessWidget {
  const _Unplayable({required this.item, required this.why});

  final MediaItem item;
  final String why;

  @override
  Widget build(BuildContext context) {
    return TextButton.icon(
      key: Key(
        '${item.video ? 'video' : 'sound'}-${item.side.name}-${item.idx}',
      ),
      onPressed: null,
      icon: Icon(item.video ? Icons.videocam_off_outlined : Icons.volume_off),
      label: Text(why, style: const TextStyle(fontSize: 14)),
    );
  }
}

/// The zero-height block before the element with id `answer`; it reports
/// its context once it is laid out.
class _AnswerAnchor extends StatefulWidget {
  const _AnswerAnchor({required this.onLaidOut});

  final ValueChanged<BuildContext> onLaidOut;

  @override
  State<_AnswerAnchor> createState() => _AnswerAnchorState();
}

class _AnswerAnchorState extends State<_AnswerAnchor> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) widget.onLaidOut(context);
    });
  }

  @override
  Widget build(BuildContext context) =>
      const SizedBox.shrink(key: Key('answer-anchor'));
}
