import 'package:cardflip/golang/golang.dart' as golang;

import 'l10n.dart';

/// The name stock note type [stock] (its English name, which identifies it
/// in the Go API) has in [l].
String stockTypeName(AppLocalizations l, String stock) => switch (stock) {
  'Basic' => l.stockBasic,
  'Basic (and reversed card)' => l.stockReversed,
  'Basic (optional reversed card)' => l.stockOptionalReversed,
  'Basic (type in the answer)' => l.stockTypeIn,
  'Cloze' => l.stockCloze,
  'Image Occlusion' => l.stockImageOcclusion,
  _ => stock,
};

/// The names that a new collection's deck and preset and the stock note
/// types are created with in [l]'s language. The fields and card types are
/// in the order of Go's English stock note types.
golang.BuiltinNames builtinNames(AppLocalizations l) {
  golang.StockNames stock(
    String stock,
    List<String> fields,
    List<String> cardTypes,
  ) => golang.StockNames(
    stock: stock,
    name: stockTypeName(l, stock),
    fields: fields,
    cardTypes: cardTypes,
  );
  final front = l.stockFieldFront, back = l.stockFieldBack;
  final card1 = l.defaultCardTypeName(1), card2 = l.defaultCardTypeName(2);
  return golang.BuiltinNames(
    deck: l.builtinDeckName,
    preset: l.builtinPresetName,
    noteTypes: [
      stock('Basic', [front, back], [card1]),
      stock('Basic (and reversed card)', [front, back], [card1, card2]),
      stock(
        'Basic (optional reversed card)',
        [front, back, l.stockFieldAddReverse],
        [card1, card2],
      ),
      stock('Basic (type in the answer)', [front, back], [card1]),
      stock('Cloze', [l.stockFieldText, l.stockFieldBackExtra], [l.stockCloze]),
      stock(
        'Image Occlusion',
        [
          l.stockFieldOcclusion,
          l.stockFieldImage,
          l.stockFieldHeader,
          l.stockFieldBackExtra,
          l.stockFieldComments,
        ],
        [l.stockImageOcclusion],
      ),
    ],
  );
}

/// Tells Go the built-in names of [l]'s language.
Future<void> sendBuiltinNames(AppLocalizations l) =>
    golang.app.setBuiltinNames(builtinNames(l));
