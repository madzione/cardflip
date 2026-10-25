import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/data/changes.dart';
import 'package:cardflip/data/deck_list_model.dart';
import 'package:cardflip/data/progress_model.dart';
import 'package:cardflip/data/settings_model.dart';

import 'support/fake_go.dart';

void main() {
  /// Loads the settings, deck list and progress models against the fake.
  /// Call the returned close function before the test ends, to cancel the
  /// deck list's timer.
  Future<(ProgressModel, void Function())> loaded(WidgetTester tester) async {
    final hub = ChangeHub();
    final settings = SettingsModel(hub);
    final deckList = DeckListModel(hub);
    final model = ProgressModel(deckList, settings);
    await settings.load();
    await deckList.load();
    await tester.pump();
    return (
      model,
      () {
        model.dispose();
        deckList.dispose();
        settings.dispose();
        hub.dispose();
      },
    );
  }

  testWidgets('loads with the deck list and after changes', (tester) async {
    final go = FakeGo()..progress = progressJson(current: 12);
    final (model, close) = await loaded(tester);
    expect(model.progress?.current, 12);
    expect(model.progress?.atRisk, isTrue);

    // After a session: the first answers of the day extend the streak.
    go.progress = progressJson(
      current: 13,
      cards: 12,
      goalKind: 'cards',
      goal: 30,
    );
    go.emitChange(kind: 'answer');
    await tester.pump();
    await tester.pump();
    expect(model.progress?.current, 13);
    expect(model.progress?.todayCards, 12);
    expect(model.progress?.atRisk, isFalse);
    close();
  });

  testWidgets('reloads when a new study day starts', (tester) async {
    final go = FakeGo()..progress = progressJson(current: 5, cards: 20);
    // The next study day starts in a second.
    go.nextChange = () => DateTime.now().add(const Duration(seconds: 1));
    final (model, close) = await loaded(tester);
    expect(model.progress?.todayStudied, isTrue);

    // Day change (spec): today's progress is 0 and the streak at risk.
    go.progress = progressJson(current: 5);
    await tester.pump(const Duration(seconds: 3));
    await tester.pump();
    expect(model.progress?.todayCards, 0);
    expect(model.progress?.atRisk, isTrue);
    close();
  });

  testWidgets('does not read the history while hidden', (tester) async {
    final go = FakeGo()..progress = progressJson(current: 3);
    go.settings = {...go.settings, 'showProgress': false};
    final (model, close) = await loaded(tester);
    expect(go.callsTo('progress.Today'), isEmpty);
    expect(model.progress, isNull);

    go.settings = {...go.settings, 'showProgress': true};
    go.emitChange(kind: 'settings', settings: true);
    await tester.pump();
    await tester.pump();
    expect(model.progress?.current, 3);
    close();
  });
}
