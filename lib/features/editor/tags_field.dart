import 'package:flutter/material.dart';
import 'package:cardflip/l10n/l10n.dart';

/// Edits a list of tags as chips. Typing a space, comma or Enter turns the
/// text into tags; tags never contain spaces.
class TagsField extends StatefulWidget {
  const TagsField({super.key, required this.tags, required this.onChanged});

  final List<String> tags;
  final ValueChanged<List<String>> onChanged;

  @override
  State<TagsField> createState() => TagsFieldState();
}

class TagsFieldState extends State<TagsField> {
  final _controller = TextEditingController();
  final _focus = FocusNode();

  /// Turns typed text that was not committed yet into tags.
  List<String> commit() {
    final words = _controller.text
        .split(RegExp(r'[\s,]+'))
        .where((w) => w.isNotEmpty);
    if (words.isEmpty) return widget.tags;
    final tags = [...widget.tags];
    for (final w in words) {
      if (!tags.any((t) => t.toLowerCase() == w.toLowerCase())) tags.add(w);
    }
    _controller.clear();
    widget.onChanged(tags);
    return tags;
  }

  void _onChanged(String text) {
    if (text.contains(RegExp(r'[\s,]'))) commit();
  }

  @override
  void dispose() {
    _controller.dispose();
    _focus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return InputDecorator(
      decoration: InputDecoration(
        labelText: context.l10n.tags,
        border: const OutlineInputBorder(),
      ),
      child: Wrap(
        spacing: 6,
        runSpacing: 6,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          for (final t in widget.tags)
            InputChip(
              label: Text(t),
              onDeleted: () => widget.onChanged([...widget.tags]..remove(t)),
            ),
          ConstrainedBox(
            constraints: const BoxConstraints(minWidth: 120, maxWidth: 240),
            child: TextField(
              key: const Key('tags-input'),
              controller: _controller,
              focusNode: _focus,
              decoration: InputDecoration.collapsed(
                hintText: context.l10n.addTags,
              ),
              onChanged: _onChanged,
              onSubmitted: (_) {
                commit();
                _focus.requestFocus();
              },
            ),
          ),
        ],
      ),
    );
  }
}
