import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:html/parser.dart' as html_parser;

import 'package:cardflip/app/errors.dart';
import 'package:cardflip/data/scope.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;
import 'package:cardflip/widgets/media_cache.dart';
import 'package:cardflip/l10n/l10n.dart';

import 'occlusion_editor.dart';
import 'rich_field.dart';
import 'rich_html.dart';

/// The mask editor in place of an Image Occlusion note's Occlusion and
/// Image fields: its shapes, the image, picking another one and the check
/// before saving, for the note editor and the browser's edit pane.
mixin OcclusionFields<T extends StatefulWidget> on State<T> {
  /// The mask editor's shapes, for Image Occlusion note types.
  OcclusionController? occlusion;

  /// The Image field, and the image it shows.
  String imageHtml = '';
  Uint8List? occlusionImage;

  int _occlusionField = 0;
  int _imageField = 0;

  /// Shows why the image could not be loaded or picked; null clears it.
  void showOcclusionError(String? message);

  /// The index of the Occlusion field, where the mask editor is shown.
  int get occlusionFieldIndex => _occlusionField;

  /// Makes the mask editor for note type [type] with fields [raw], or none
  /// when it is not an Image Occlusion type. With [editing] the shapes keep
  /// their numbers, and new ones get numbers none of [cardNumbers] has.
  void initOcclusion(
    golang.NoteType? type,
    List<String> raw, {
    required bool editing,
    List<int> cardNumbers = const [],
  }) {
    disposeOcclusion();
    occlusion = null;
    occlusionImage = null;
    imageHtml = '';
    if (!(type?.imageOcclusion ?? false)) return;
    _occlusionField = type!.occlusionField ?? 0;
    _imageField = type.imageField ?? 0;
    String value(int i) => i >= 0 && i < raw.length ? raw[i] : '';
    occlusion = OcclusionController(
      field: value(_occlusionField),
      keepNumbers: editing,
      cardNumbers: cardNumbers,
    );
    imageHtml = value(_imageField);
    loadOcclusionImage();
  }

  void disposeOcclusion() => occlusion?.dispose();

  /// Whether field [i] is edited with a field input; the mask editor
  /// stands for the Occlusion and Image fields.
  bool shownAsField(int i) =>
      occlusion == null || (i != _occlusionField && i != _imageField);

  /// Why the note cannot be saved: it has no image or no shapes.
  String? occlusionProblem() {
    final o = occlusion;
    if (o == null) return null;
    return imageHtml.isEmpty
        ? context.l10n.chooseImageToDrawOn
        : o.cardCount == 0
        ? context.l10n.drawShapes
        : null;
  }

  /// Writes the shapes and the image into the Occlusion and Image fields
  /// of [fields].
  void applyOcclusion(List<String> fields) {
    final o = occlusion;
    if (o == null) return;
    if (_occlusionField >= 0) fields[_occlusionField] = o.field;
    if (_imageField >= 0) fields[_imageField] = imageHtml;
  }

  /// Shows the Image field's image in the mask editor.
  Future<void> loadOcclusionImage() async {
    final html = imageHtml;
    final img = html_parser.parseFragment(html).querySelector('img');
    final src = img?.attributes['src']?.trim() ?? '';
    if (src.isEmpty) return;
    try {
      final bytes = await loadMediaBytes(MediaCache.instance, src);
      if (mounted && imageHtml == html) {
        setState(() => occlusionImage = bytes);
      }
    } catch (e) {
      if (mounted) showOcclusionError(context.l10n.imageMissing(src));
    }
  }

  /// Lets the user pick the image to draw on, which goes into the media
  /// and the Image field. The shapes stay where they are.
  Future<void> pickOcclusionImage() async {
    final List<PickedFile> files;
    try {
      files = await AppScope.read(context).picker.pickMedia();
    } catch (e) {
      if (mounted) showOcclusionError(errorMessage(e));
      return;
    }
    if (files.isEmpty || !mounted) return;
    final image = files.where((f) => isImageFile(f.name)).firstOrNull;
    if (image == null) {
      showOcclusionError(context.l10n.choosePictureToDrawOn);
      return;
    }
    try {
      final name = await golang.media.add(image.name, image.bytes);
      if (!mounted || occlusion == null) return;
      setState(() {
        imageHtml = '<img src="${escapeAttribute(name)}">';
        occlusionImage = image.bytes;
      });
      showOcclusionError(null);
    } catch (e) {
      if (mounted) showOcclusionError(errorMessage(e));
    }
  }

  /// The image and the mask editor, for the Occlusion and Image fields.
  Widget occlusionSection(BuildContext context) {
    final image = occlusionImage;
    final theme = Theme.of(context);
    return Padding(
      key: const Key('occlusion-section'),
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  context.l10n.image,
                  style: theme.textTheme.titleSmall,
                ),
              ),
              if (image != null)
                TextButton.icon(
                  key: const Key('occlusion-pick-image'),
                  onPressed: pickOcclusionImage,
                  icon: const Icon(Icons.image_outlined),
                  label: Text(context.l10n.changeImage),
                ),
            ],
          ),
          const SizedBox(height: 8),
          if (image != null)
            OcclusionEditor(controller: occlusion!, image: image)
          else
            Container(
              height: 160,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                border: Border.all(color: theme.colorScheme.outlineVariant),
                borderRadius: BorderRadius.circular(8),
              ),
              child: imageHtml.isNotEmpty
                  ? Text(context.l10n.loadingImage)
                  : FilledButton.tonalIcon(
                      key: const Key('occlusion-pick-image'),
                      onPressed: pickOcclusionImage,
                      icon: const Icon(Icons.add_photo_alternate_outlined),
                      label: Text(context.l10n.chooseImage),
                    ),
            ),
        ],
      ),
    );
  }
}
