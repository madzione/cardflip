import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart' show DragStartBehavior;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_quill/flutter_quill.dart';
import 'package:flutter_quill/quill_delta.dart';
import 'package:quill_native_bridge/quill_native_bridge.dart';

import 'package:cardflip/app/appearance.dart';
import 'package:cardflip/app/shortcuts.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/css_color.dart';
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'math_insert.dart';
import 'rich_html.dart';

/// The note shown above a field that opened as HTML because the visual
/// editor would lose some of its formatting.
String get lossyNote => appL10n.lossyNote;

bool get _isMac => defaultTargetPlatform == TargetPlatform.macOS;

const _imageExtensions = {
  '.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg', '.avif', //
};

/// Whether a media file is an image (by extension); other media files are
/// audio or video.
bool isImageFile(String name) {
  final dot = name.lastIndexOf('.');
  return dot >= 0 &&
      _imageExtensions.contains(name.substring(dot).toLowerCase());
}

SingleActivator _command(LogicalKeyboardKey key, {bool shift = false}) =>
    SingleActivator(key, control: !_isMac, meta: _isMac, shift: shift);

/// The HTML toggle's shortcut, Ctrl/Cmd+Shift+X.
final htmlToggleShortcut = _command(LogicalKeyboardKey.keyX, shift: true);

/// Focuses [node] once the next frame is built: a visual editor whose
/// content just changed must lay it out before it takes the focus.
void focusAfterFrame(FocusNode node) {
  WidgetsBinding.instance.addPostFrameCallback((_) {
    if (node.context != null) node.requestFocus();
  });
  WidgetsBinding.instance.ensureVisualUpdate();
}

/// The clipboard as pasting reads it. Tests replace [instance].
class FieldClipboard {
  static FieldClipboard instance = FieldClipboard();

  /// The clipboard's HTML, where the platform offers it (desktop and web;
  /// null on mobile).
  Future<String?> html() async {
    try {
      final bridge = QuillNativeBridge();
      if (!await bridge.isSupported(
        QuillNativeBridgeFeature.getClipboardHtml,
      )) {
        return null;
      }
      return await bridge.getClipboardHtml();
    } catch (_) {
      return null;
    }
  }

  /// The clipboard's plain text.
  Future<String?> text() async =>
      (await Clipboard.getData(Clipboard.kTextPlain))?.text;

  /// The clipboard's image (usually PNG), or null.
  Future<Uint8List?> image() async {
    try {
      final bridge = QuillNativeBridge();
      if (!await bridge.isSupported(
        QuillNativeBridgeFeature.getClipboardImage,
      )) {
        return null;
      }
      return await bridge.getClipboardImage();
    } catch (_) {
      return null;
    }
  }
}

/// One note field: a visual editor for the supported HTML subset, or the
/// field's HTML source.
class RichFieldController extends ChangeNotifier {
  RichFieldController({String html = '', this.htmlByDefault = false}) {
    quill.addListener(notifyListeners);
    load(html);
  }

  /// The field opens in HTML mode (note type option).
  bool htmlByDefault;

  /// "Strip formatting when pasting": a normal paste inserts plain text,
  /// and the plain-paste shortcut keeps formatting.
  bool stripPasteFormatting = false;

  /// Stores an image pasted into the field and inserts it (see
  /// [insertMedia]); without it images on the clipboard are ignored.
  Future<void> Function(Uint8List image)? onPasteImage;

  late final quill = QuillController(
    document: Document(),
    selection: const TextSelection.collapsed(offset: 0),
    // The clipboard hooks are marked experimental; they are the only way to
    // replace Quill's own HTML paste, which keeps far more formatting.
    config: QuillControllerConfig(
      // ignore: experimental_member_use
      clipboardConfig: QuillClipboardConfig(
        // ignore: experimental_member_use
        onClipboardPaste: () async {
          await paste(formatted: !stripPasteFormatting);
          return true;
        },
      ),
    ),
  );
  final source = TextEditingController();

  /// Whether the field shows its HTML source.
  bool get htmlMode => _htmlMode;
  bool _htmlMode = false;

  /// Whether the field opened as HTML because the visual editor would lose
  /// formatting; the editor shows [lossyNote].
  bool get lossy => _lossy;
  bool _lossy = false;

  /// The HTML the visual editor was loaded from, and its document then:
  /// while the document is unchanged, [html] is this HTML exactly.
  String _loadedHtml = '';
  Delta _loadedDelta = Delta();
  bool _divLines = false;

  /// Colours used in the document, for the editor's palette.
  final colors = <String>{};

  /// Shows [html], in HTML mode when the note type says so or the visual
  /// editor can't keep it.
  void load(String html) {
    _lossy = !htmlByDefault && !isLossless(html);
    if (htmlByDefault || _lossy) {
      _htmlMode = true;
      source.text = html;
      _setDocument('');
    } else {
      _htmlMode = false;
      source.text = '';
      _setDocument(html);
    }
    notifyListeners();
  }

  void _setDocument(String html) {
    final doc = htmlToDocument(html);
    _loadedHtml = html;
    _loadedDelta = doc.delta;
    _divLines = doc.divLines;
    colors.addAll(documentColors(doc.delta));
    // Replace the content in place: swapping the Document of a focused
    // editor leaves its text input out of step.
    final replace = Delta();
    for (final op in doc.delta.operations) {
      replace.push(op);
    }
    replace.delete(quill.document.length);
    // A change made here is not typing: the editor must not take the
    // focus for it (two fields doing so steal it from each other forever).
    quill.ignoreFocusOnTextChange = true;
    try {
      quill.compose(
        replace,
        const TextSelection.collapsed(offset: 0),
        ChangeSource.remote,
      );
      quill.updateSelection(
        const TextSelection.collapsed(offset: 0),
        ChangeSource.remote,
      );
    } finally {
      quill.ignoreFocusOnTextChange = false;
    }
  }

  /// The field's HTML.
  String get html {
    if (_htmlMode) return source.text;
    final delta = quill.document.toDelta();
    if (delta == _loadedDelta) return _loadedHtml;
    return documentToHtml(delta, divLines: _divLines);
  }

  /// Empties the field, keeping its mode.
  void clear() {
    if (_htmlMode) {
      source.clear();
    } else {
      _setDocument('');
    }
    _lossy = false;
    notifyListeners();
  }

  /// Switches to the HTML source.
  void showHtml() {
    if (_htmlMode) return;
    source.text = html;
    _htmlMode = true;
    notifyListeners();
  }

  /// Whether the visual editor keeps everything in the current HTML.
  bool get canShowVisualLosslessly => !_htmlMode || isLossless(source.text);

  /// Switches to the visual editor. Formatting outside the subset is
  /// dropped once the field is edited; check [canShowVisualLosslessly]
  /// first.
  void showVisual() {
    if (!_htmlMode) return;
    _setDocument(source.text);
    _htmlMode = false;
    _lossy = false;
    notifyListeners();
  }

  /// Wraps the selection in [open] and [close] as text, leaving the cursor
  /// before [close]. Without a selection the pair goes at the end.
  void wrap(String open, String close) {
    if (_htmlMode) {
      wrapSelection(source, open, close);
      return;
    }
    final end = quill.document.length - 1;
    var sel = quill.selection;
    if (!sel.isValid || sel.end > end) {
      sel = TextSelection.collapsed(offset: end);
    }
    quill.replaceText(sel.end, 0, close, null);
    quill.replaceText(sel.start, 0, open, null);
    quill.updateSelection(
      TextSelection.collapsed(offset: sel.end + open.length),
      ChangeSource.local,
    );
  }

  /// Inserts a stored media file over the selection: an image as `<img>`,
  /// audio and video as `[sound:…]`. Without a selection it goes at the
  /// end.
  void insertMedia(String name) {
    final image = isImageFile(name);
    if (_htmlMode) {
      final markup = image
          ? '<img src="${escapeAttribute(name)}">'
          : '[sound:$name]';
      final sel = source.selection.isValid
          ? source.selection
          : TextSelection.collapsed(offset: source.text.length);
      source.value = source.value.replaced(sel, markup);
      return;
    }
    final end = quill.document.length - 1;
    var sel = quill.selection;
    if (!sel.isValid || sel.end > end) {
      sel = TextSelection.collapsed(offset: end);
    }
    final Object data = image ? BlockEmbed.image(name) : '[sound:$name]';
    final length = image ? 1 : '[sound:$name]'.length;
    quill.replaceText(
      sel.start,
      sel.end - sel.start,
      data,
      TextSelection.collapsed(offset: sel.start + length),
    );
  }

  /// Pastes the clipboard over the selection: with the toolbar's
  /// formatting when [formatted] and the clipboard has HTML (desktop and
  /// web), otherwise as plain text. In HTML mode it pastes plain text. A
  /// clipboard with an image and no text goes to [onPasteImage].
  Future<void> paste({required bool formatted}) async {
    final html = formatted && !_htmlMode
        ? await FieldClipboard.instance.html()
        : null;
    final Object data;
    final int length;
    final delta = html == null ? null : pasteDelta(html);
    if (delta != null && delta.isNotEmpty) {
      data = delta;
      length = delta.operations.fold(0, (n, op) => n + (op.length ?? 0));
      colors.addAll(documentColors(delta));
    } else {
      final text = (await FieldClipboard.instance.text())?.replaceAll(
        '\r\n',
        '\n',
      );
      if (text == null || text.isEmpty) {
        final onImage = onPasteImage;
        final image = onImage == null
            ? null
            : await FieldClipboard.instance.image();
        if (image != null && image.isNotEmpty) await onImage!(image);
        return;
      }
      if (_htmlMode) {
        final sel = source.selection.isValid
            ? source.selection
            : TextSelection.collapsed(offset: source.text.length);
        source.value = source.value.replaced(sel, text);
        return;
      }
      data = text;
      length = text.length;
    }
    final sel = quill.selection;
    quill.replaceText(
      sel.start,
      sel.end - sel.start,
      data,
      TextSelection.collapsed(offset: sel.start + length),
    );
  }

  // Formatting of the visual editor's selection.

  /// Whether the selection has [attribute] (with its value).
  bool has(Attribute attribute) =>
      quill.getSelectionStyle().attributes[attribute.key]?.value ==
      attribute.value;

  /// Turns [attribute] on for the selection, or off when it has it.
  void toggle(Attribute attribute) {
    if (_htmlMode) return;
    quill.formatSelection(
      has(attribute) ? Attribute.clone(attribute, null) : attribute,
    );
  }

  /// Sets the text colour, or removes it when [color] is null.
  void setColor(String? color) => _setColor(Attribute.color, color);

  /// Sets the highlight colour, or removes it when [color] is null.
  void setBackground(String? color) => _setColor(Attribute.background, color);

  void _setColor(Attribute<String?> attribute, String? color) {
    if (_htmlMode) return;
    if (color != null) colors.add(color);
    quill.formatSelection(Attribute.clone(attribute, color));
  }

  /// Sets the alignment of the selected lines; null removes it, so the
  /// card's CSS decides (often centred).
  void align(String? value) {
    if (_htmlMode) return;
    quill.formatSelection(Attribute.clone(Attribute.align, value));
  }

  void indent(bool increase) {
    if (_htmlMode) return;
    quill.indentSelection(increase);
  }

  /// Removes inline formatting (not lists and alignment) from the
  /// selection.
  void clearFormatting() {
    if (_htmlMode) return;
    for (final a in <Attribute>[
      Attribute.bold,
      Attribute.italic,
      Attribute.underline,
      Attribute.strikeThrough,
      Attribute.script,
      Attribute.color,
      Attribute.background,
      Attribute.link,
    ]) {
      quill.formatSelection(Attribute.clone(a, null));
    }
  }

  @override
  void dispose() {
    quill.dispose();
    source.dispose();
    super.dispose();
  }
}

/// An intent no action handles, so that the key reaches the widgets above
/// the editor (the math chords use Ctrl/Cmd+M).
class _PassIntent extends Intent {
  const _PassIntent();
}

/// Quill shortcuts for things the field doesn't support (headers, code,
/// quotes, check lists, links, search, image dialog) or that are taken
/// (Ctrl/Cmd+M starts a math chord).
final _passShortcuts = <ShortcutActivator, Intent>{
  for (final key in [
    LogicalKeyboardKey.keyM,
    LogicalKeyboardKey.keyK,
    LogicalKeyboardKey.keyG,
    LogicalKeyboardKey.keyF,
    LogicalKeyboardKey.backquote,
    LogicalKeyboardKey.digit0,
    LogicalKeyboardKey.digit1,
    LogicalKeyboardKey.digit2,
    LogicalKeyboardKey.digit3,
    LogicalKeyboardKey.digit4,
    LogicalKeyboardKey.digit5,
    LogicalKeyboardKey.digit6,
  ])
    _command(key): const _PassIntent(),
  for (final key in [
    LogicalKeyboardKey.keyM,
    LogicalKeyboardKey.keyB,
    LogicalKeyboardKey.keyC,
    LogicalKeyboardKey.tilde,
  ])
    _command(key, shift: true): const _PassIntent(),
};

/// The options of a field that has none set.
const defaultFieldOptions = golang.FieldOptions(
  htmlByDefault: false,
  sticky: false,
  font: '',
  size: 0,
  rtl: false,
  collapsed: false,
  excludeFromSearch: false,
);

extension NoteTypeFieldOptions on golang.NoteType {
  /// The options of field [index]; the defaults when Go sent none.
  golang.FieldOptions fieldOption(int index) =>
      index < fieldOptions.length ? fieldOptions[index] : defaultFieldOptions;
}

/// A note field with its label, HTML toggle and editor.
class RichField extends StatefulWidget {
  const RichField({
    super.key,
    required this.label,
    required this.controller,
    required this.focusNode,
    this.autofocus = false,
    this.action,
    this.marked = false,
    this.options = defaultFieldOptions,
  });

  final String label;
  final RichFieldController controller;
  final FocusNode focusNode;
  final bool autofocus;

  /// The field's options: the visual editor uses its editing font and
  /// size and its direction, and a field collapsed by default starts with
  /// only its header shown.
  final golang.FieldOptions options;

  /// A button shown in the field's header, before the HTML toggle.
  final Widget? action;

  /// Draws the border in the error colour, e.g. for a duplicate first
  /// field.
  final bool marked;

  @override
  State<RichField> createState() => RichFieldState();
}

class RichFieldState extends State<RichField> {
  RichFieldController get _c => widget.controller;
  final _scroll = ScrollController();
  late bool _collapsed = widget.options.collapsed;

  @override
  void didUpdateWidget(RichField oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Another note or note type starts as its options say.
    if (oldWidget.controller != widget.controller) {
      _collapsed = widget.options.collapsed;
    }
  }

  /// Shows the editor of a collapsed field and focuses it.
  void expand() {
    if (!_collapsed) return;
    setState(() => _collapsed = false);
    focusAfterFrame(widget.focusNode);
  }

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  /// Switches between the visual editor and the HTML source; switching to
  /// the visual editor asks first when formatting would be lost.
  Future<void> toggleHtml() async {
    if (!_c.htmlMode) {
      _c.showHtml();
    } else {
      if (!_c.canShowVisualLosslessly && !await _confirmVisual()) return;
      _c.showVisual();
    }
    focusAfterFrame(widget.focusNode);
  }

  Future<bool> _confirmVisual() async {
    final ok = await showDialog<bool>(
      context: context,
      animationStyle: popupAnimation(context),
      builder: (context) => AlertDialog(
        title: Text(context.l10n.switchToVisualTitle),
        content: Text(context.l10n.switchToVisualBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            key: const Key('confirm-visual'),
            onPressed: () => Navigator.pop(context, true),
            child: Text(context.l10n.switchAction),
          ),
        ],
      ),
    );
    return ok ?? false;
  }

  Map<ShortcutActivator, VoidCallback> get _bindings => {
    htmlToggleShortcut: toggleHtml,
    // The HTML source pastes through the field too, so images on the
    // clipboard are added (the visual editor has its own paste hook).
    if (_c.htmlMode)
      _command(LogicalKeyboardKey.keyV): () => _c.paste(formatted: false),
    // The other paste: plain text, or formatted when the setting strips
    // formatting from normal pastes.
    _command(LogicalKeyboardKey.keyV, shift: true): () =>
        _c.paste(formatted: _c.stripPasteFormatting),
    _command(LogicalKeyboardKey.equal): () => _c.toggle(Attribute.subscript),
    _command(LogicalKeyboardKey.equal, shift: true): () =>
        _c.toggle(Attribute.superscript),
    _command(LogicalKeyboardKey.keyR): _c.clearFormatting,
  };

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final settings = AppScope.maybeRead(context)?.settings;
    return ListenableBuilder(
      listenable: Listenable.merge([_c, widget.focusNode, settings]),
      builder: (context, _) {
        _c.stripPasteFormatting = settings?.stripPasteFormatting ?? false;
        final focused = widget.focusNode.hasFocus;
        return CallbackShortcuts(
          bindings: _bindings,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                children: [
                  Expanded(
                    child: GestureDetector(
                      behavior: HitTestBehavior.opaque,
                      onTap: _collapsed ? expand : null,
                      child: Text(
                        widget.label,
                        style: theme.textTheme.labelLarge?.copyWith(
                          color: focused ? theme.colorScheme.primary : null,
                        ),
                      ),
                    ),
                  ),
                  if (widget.options.collapsed || _collapsed)
                    IconButton(
                      key: ValueKey('$_keyName-collapse'),
                      tooltip: _collapsed
                          ? context.l10n.expandField
                          : context.l10n.collapseField,
                      visualDensity: VisualDensity.compact,
                      icon: Icon(
                        _collapsed ? Icons.expand_more : Icons.expand_less,
                      ),
                      onPressed: _collapsed
                          ? expand
                          : () => setState(() => _collapsed = true),
                    ),
                  ?widget.action,
                  IconButton(
                    key: ValueKey('$_keyName-html'),
                    tooltip: withShortcut(
                      _c.htmlMode
                          ? context.l10n.visualEditor
                          : context.l10n.htmlEditor,
                      htmlToggleShortcutLabel,
                    ),
                    isSelected: _c.htmlMode,
                    visualDensity: VisualDensity.compact,
                    icon: const Icon(Icons.code),
                    onPressed: toggleHtml,
                  ),
                ],
              ),
              if (_collapsed)
                _collapsedPreview(theme)
              else if (_c.lossy)
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text(
                    lossyNote,
                    key: ValueKey('$_keyName-lossy'),
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.tertiary,
                    ),
                  ),
                ),
              if (!_collapsed)
                _c.htmlMode ? _source(theme) : _visual(theme, focused),
            ],
          ),
        );
      },
    );
  }

  String get _keyName => switch (widget.key) {
    ValueKey<String>(:final value) => value,
    _ => 'field',
  };

  Widget _source(ThemeData theme) => TextField(
    key: ValueKey('$_keyName-source'),
    controller: _c.source,
    focusNode: widget.focusNode,
    autofocus: widget.autofocus,
    minLines: 2,
    maxLines: 12,
    style: theme.textTheme.bodyMedium?.copyWith(fontFamily: 'monospace'),
    decoration: InputDecoration(
      border: const OutlineInputBorder(),
      enabledBorder: widget.marked
          ? OutlineInputBorder(
              borderSide: BorderSide(color: theme.colorScheme.error),
            )
          : null,
      focusedBorder: widget.marked
          ? OutlineInputBorder(
              borderSide: BorderSide(color: theme.colorScheme.error, width: 2),
            )
          : null,
    ),
  );

  /// The start of a collapsed field's text, tapped to expand it.
  Widget _collapsedPreview(ThemeData theme) {
    // Images and other embeds are left out.
    final text = _c.quill.document
        .toPlainText()
        .replaceAll('\uFFFC', '')
        .trim()
        .replaceAll(RegExp(r'\s+'), ' ');
    return InkWell(
      key: ValueKey('$_keyName-collapsed'),
      onTap: expand,
      borderRadius: BorderRadius.circular(4),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        child: Text(
          text.isEmpty ? context.l10n.emptyField : text,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          textDirection: widget.options.rtl ? TextDirection.rtl : null,
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }

  /// The quill styles for the field's editing font and size, merged over
  /// the defaults; null without either.
  DefaultStyles? _fontStyles(BuildContext context) {
    final o = widget.options;
    if (o.font.isEmpty && o.size <= 0) return null;
    final base = DefaultStyles.getInstance(context);
    TextStyle font(TextStyle s) => s.copyWith(
      fontFamily: o.font.isEmpty ? null : o.font,
      fontSize: o.size > 0 ? o.size.toDouble() : null,
    );
    return DefaultStyles(
      paragraph: base.paragraph?.copyWith(style: font(base.paragraph!.style)),
      lists: base.lists?.copyWith(style: font(base.lists!.style)),
    );
  }

  Widget _visual(ThemeData theme, bool focused) {
    final scheme = theme.colorScheme;
    final palette = DefaultStyles(
      palette: {for (final c in _c.colors) c: ?parseCssColor(c)},
    );
    final fontStyles = _fontStyles(context);
    final editor = Container(
      constraints: const BoxConstraints(minHeight: 48),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(4),
        border: Border.all(
          color: widget.marked
              ? scheme.error
              : focused
              ? scheme.primary
              : scheme.outline,
          width: focused ? 2 : 1,
        ),
      ),
      child: QuillEditor(
        key: ValueKey('$_keyName-visual'),
        controller: _c.quill,
        focusNode: widget.focusNode,
        scrollController: _scroll,
        config: QuillEditorConfig(
          autoFocus: widget.autofocus,
          expands: false,
          scrollable: false,
          textCapitalization: TextCapitalization.none,
          embedBuilders: const [ImageEmbedBuilder()],
          customShortcuts: _passShortcuts,
          customStyles: fontStyles?.merge(palette) ?? palette,
        ),
      ),
    );
    // Right-to-left fields lay out and align from the right, like the
    // card renders dir="rtl".
    return widget.options.rtl
        ? Directionality(textDirection: TextDirection.rtl, child: editor)
        : editor;
  }
}

/// The formatting toolbar for the field being edited.
class RichToolbar extends StatelessWidget {
  const RichToolbar({super.key, required this.field, this.trailing = const []});

  /// The field the buttons act on; null disables them.
  final RichFieldController? field;

  /// Buttons after the formatting ones, set apart by a divider and kept on
  /// one line. They are shown as given, also while the field shows its
  /// HTML source.
  final List<Widget> trailing;

  @override
  Widget build(BuildContext context) {
    final field = this.field;
    if (field == null) return _buttons(context, null);
    return ListenableBuilder(
      listenable: field,
      builder: (context, _) => _buttons(context, field.htmlMode ? null : field),
    );
  }

  Widget _buttons(BuildContext context, RichFieldController? f) {
    Widget toggle(String tooltip, IconData icon, Attribute attribute) =>
        IconButton(
          tooltip: tooltip,
          icon: Icon(icon),
          isSelected: f?.has(attribute) ?? false,
          onPressed: f == null ? null : () => f.toggle(attribute),
        );
    Widget action(String tooltip, IconData icon, VoidCallback? onPressed) =>
        IconButton(
          tooltip: tooltip,
          icon: Icon(icon),
          onPressed: f == null ? null : onPressed,
        );
    final l = context.l10n;
    return Wrap(
      key: const Key('rich-toolbar'),
      children: [
        toggle(
          withShortcut(l.bold, boldShortcutLabel),
          Icons.format_bold,
          Attribute.bold,
        ),
        toggle(
          withShortcut(l.italic, italicShortcutLabel),
          Icons.format_italic,
          Attribute.italic,
        ),
        toggle(
          withShortcut(l.underline, underlineShortcutLabel),
          Icons.format_underlined,
          Attribute.underline,
        ),
        toggle(
          withShortcut(l.subscript, subscriptShortcutLabel),
          Icons.subscript,
          Attribute.subscript,
        ),
        toggle(
          withShortcut(l.superscript, superscriptShortcutLabel),
          Icons.superscript,
          Attribute.superscript,
        ),
        _ColorButton(
          tooltip: l.textColour,
          icon: Icons.format_color_text,
          onSelected: f?.setColor,
        ),
        _ColorButton(
          tooltip: l.highlightColour,
          icon: Icons.format_color_fill,
          onSelected: f?.setBackground,
        ),
        action(
          withShortcut(l.removeFormatting, removeFormattingShortcutLabel),
          Icons.format_clear,
          f?.clearFormatting,
        ),
        toggle(l.bulletedList, Icons.format_list_bulleted, Attribute.ul),
        toggle(l.numberedList, Icons.format_list_numbered, Attribute.ol),
        PopupMenuButton<String?>(
          popUpAnimationStyle: popupAnimation(context),
          tooltip: l.alignment,
          enabled: f != null,
          icon: const Icon(Icons.format_align_left),
          // Left is set, not cleared: card CSS often centres text.
          onSelected: (v) => f?.align(v),
          itemBuilder: (context) => [
            PopupMenuItem(value: 'left', child: Text(l.alignLeft)),
            PopupMenuItem(value: 'center', child: Text(l.alignCenter)),
            PopupMenuItem(value: 'right', child: Text(l.alignRight)),
            PopupMenuItem(value: 'justify', child: Text(l.justify)),
          ],
        ),
        action(l.outdent, Icons.format_indent_decrease, () => f?.indent(false)),
        action(l.indent, Icons.format_indent_increase, () => f?.indent(true)),
        if (trailing.isNotEmpty)
          Row(
            key: const Key('rich-toolbar-tools'),
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(height: 24, child: VerticalDivider(width: 16)),
              ...trailing,
            ],
          ),
      ],
    );
  }
}

/// Text and highlight colours offered by the toolbar, by name.
Map<String, String> _swatches(AppLocalizations l) => {
  l.colourBlack: '#000000',
  l.colourGray: '#808080',
  l.colourRed: '#ff0000',
  l.colourOrange: '#ff8c00',
  l.colourYellow: '#ffff00',
  l.colourGreen: '#008000',
  l.colourBlue: '#0000ff',
  l.colourPurple: '#800080',
};

class _ColorButton extends StatelessWidget {
  const _ColorButton({
    required this.tooltip,
    required this.icon,
    required this.onSelected,
  });

  final String tooltip;
  final IconData icon;

  /// Called with a colour, or null to remove it; null disables the button.
  final ValueChanged<String?>? onSelected;

  @override
  Widget build(BuildContext context) {
    final onSelected = this.onSelected;
    return PopupMenuButton<String>(
      popUpAnimationStyle: popupAnimation(context),
      tooltip: tooltip,
      enabled: onSelected != null,
      icon: Icon(icon),
      onSelected: (v) => onSelected?.call(v.isEmpty ? null : v),
      itemBuilder: (context) => [
        for (final MapEntry(key: name, value: hex) in _swatches(
          context.l10n,
        ).entries)
          PopupMenuItem(
            value: hex,
            child: Row(
              children: [
                Icon(Icons.square, color: parseCssColor(hex)),
                const SizedBox(width: 12),
                Text(name),
              ],
            ),
          ),
        PopupMenuItem(value: '', child: Text(context.l10n.noColour)),
      ],
    );
  }
}

/// Shows images of a field inline, with a handle to resize them.
class ImageEmbedBuilder extends EmbedBuilder {
  const ImageEmbedBuilder();

  @override
  String get key => BlockEmbed.imageType;

  @override
  bool get expanded => false;

  @override
  Widget build(BuildContext context, EmbedContext embedContext) {
    final node = embedContext.node;
    final width = double.tryParse(
      '${node.style.attributes[Attribute.width.key]?.value ?? ''}',
    );
    return EditorImage(
      name: node.value.data as String,
      width: width,
      readOnly: embedContext.readOnly,
      // Quill's format rules don't apply to image width, so compose the
      // change directly.
      onResize: (w) {
        final controller = embedContext.controller;
        controller.compose(
          Delta()
            ..retain(node.documentOffset)
            ..retain(1, {Attribute.width.key: '${w.round()}'}),
          controller.selection,
          ChangeSource.local,
        );
      },
    );
  }
}

/// An image in the visual editor at [width] (its own width when null),
/// with a corner handle that resizes it.
class EditorImage extends StatefulWidget {
  const EditorImage({
    super.key,
    required this.name,
    this.width,
    this.readOnly = false,
    required this.onResize,
  });

  final String name;
  final double? width;
  final bool readOnly;

  /// Called with the new width when the user lets go of the handle.
  final ValueChanged<double> onResize;

  @override
  State<EditorImage> createState() => _EditorImageState();
}

class _EditorImageState extends State<EditorImage> {
  late Future<Uint8List> _bytes = _load();
  final _box = GlobalKey();

  /// The width while the handle is being dragged.
  double? _dragWidth;

  Future<Uint8List> _load() async {
    try {
      return await MediaCache.instance.get(widget.name);
    } catch (_) {
      final decoded = Uri.decodeFull(widget.name);
      if (decoded == widget.name) rethrow;
      return MediaCache.instance.get(decoded);
    }
  }

  @override
  void didUpdateWidget(EditorImage old) {
    super.didUpdateWidget(old);
    if (old.name != widget.name) _bytes = _load();
  }

  @override
  Widget build(BuildContext context) {
    final width = _dragWidth ?? widget.width;
    final image = FutureBuilder<Uint8List>(
      future: _bytes,
      builder: (context, snap) {
        if (snap.hasError) {
          return _missing(context);
        }
        if (!snap.hasData) return const SizedBox(width: 48, height: 48);
        return Image.memory(
          snap.data!,
          width: width,
          fit: BoxFit.contain,
          errorBuilder: (context, _, _) => _missing(context),
        );
      },
    );
    return LayoutBuilder(
      builder: (context, constraints) => ConstrainedBox(
        constraints: BoxConstraints(maxWidth: constraints.maxWidth),
        child: Stack(
          key: _box,
          children: [
            SizedBox(width: width, child: image),
            if (!widget.readOnly)
              Positioned(
                right: 0,
                bottom: 0,
                child: GestureDetector(
                  key: ValueKey('resize-${widget.name}'),
                  behavior: HitTestBehavior.opaque,
                  // The whole movement counts, not only what follows the
                  // drag slop.
                  dragStartBehavior: DragStartBehavior.down,
                  onPanStart: (_) => setState(
                    () => _dragWidth =
                        (_box.currentContext?.size?.width ?? width ?? 100),
                  ),
                  onPanUpdate: (d) => setState(
                    () => _dragWidth = (_dragWidth! + d.delta.dx).clamp(
                      16,
                      constraints.maxWidth.isFinite
                          ? constraints.maxWidth
                          : 4096,
                    ),
                  ),
                  onPanEnd: (_) {
                    final w = _dragWidth;
                    setState(() => _dragWidth = null);
                    if (w != null) widget.onResize(w);
                  },
                  child: MouseRegion(
                    cursor: SystemMouseCursors.resizeDownRight,
                    child: Container(
                      width: 16,
                      height: 16,
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.primary,
                        border: Border.all(color: Colors.white),
                      ),
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _missing(BuildContext context) => Container(
    padding: const EdgeInsets.all(8),
    color: Theme.of(context).colorScheme.surfaceContainerHighest,
    child: Text('🖼 ${widget.name}'),
  );
}
