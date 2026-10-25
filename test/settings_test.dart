import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/licenses.dart';
import 'package:cardflip/data/sync_model.dart';
import 'package:cardflip/features/settings/settings_screen.dart';
import 'package:cardflip/l10n/l10n.dart';
import 'package:cardflip/widgets/time_format.dart';

import 'support/fake_go.dart';
import 'support/fake_sync_platform.dart';

Future<void> openSettings(
  WidgetTester tester, {
  FakePicker? picker,
  bool isWeb = false,
}) async {
  await pumpApp(tester, picker: picker, size: desktop, isWeb: isWeb);
  await tester.tap(find.text('Settings').last);
  await tester.pumpAndSettle();
}

Map<String, Object?> backupInspection() => {
  'stage': {
    'id': 9,
    'name': 'cardflip-2026-10-01.cardflip',
    'kind': 'backup',
    'size': 1000,
  },
  'csv': null,
  'apkg': null,
  'backup': {
    'format': 1,
    'appVersion': '0.1.0',
    'createdAt': '2026-10-01T08:00:00Z',
  },
};

/// Scrolls the settings list until [finder] is visible.
Future<void> scrollTo(WidgetTester tester, Finder finder) async {
  await tester.scrollUntilVisible(
    finder,
    200,
    scrollable: find.byType(Scrollable).last,
  );
  await tester.pumpAndSettle();
}

/// The Dart packages' licenses from the bundle's NOTICES, as the app
/// registers them at startup; the test binding registers none.
Stream<LicenseEntry> dartPackageLicenses() async* {
  final data = await rootBundle.load('NOTICES.Z');
  final notices = utf8.decode(gzip.decode(data.buffer.asUint8List()));
  for (final entry in notices.split('\n${'-' * 80}\n')) {
    final split = entry.indexOf('\n\n');
    yield LicenseEntryWithLineBreaks(
      entry.substring(0, split).split('\n'),
      entry.substring(split + 2),
    );
  }
}

void main() {
  testWidgets('restore asks before overwriting the collection', (tester) async {
    final go = FakeGo();
    final picker = FakePicker()
      ..next = stageInfo(9, 'cardflip-2026-10-01.cardflip', 'backup');
    go.handle('transfer.Inspect', (_) => backupInspection());
    go.handle('transfer.Discard', (_) => null);
    go.handle('transfer.Restore', (_) => null);
    await openSettings(tester, picker: picker);

    await scrollTo(tester, find.text('Restore backup'));
    await tester.tap(find.text('Restore backup'));
    await tester.pumpAndSettle();
    expect(picker.lastExtensions, ['cardflip']);
    expect(find.text('Replace your collection?'), findsOneWidget);
    expect(find.textContaining('backup from 2026-10-01'), findsOneWidget);
    expect(
      find.textContaining('replaces your current collection entirely'),
      findsOneWidget,
    );
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(go.callsTo('transfer.Restore'), isEmpty);
    expect(go.callsTo('transfer.Discard'), [
      [9],
    ]);

    await scrollTo(tester, find.text('Restore backup'));
    await tester.tap(find.text('Restore backup'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Restore'));
    await tester.pumpAndSettle();
    expect(go.callsTo('transfer.Restore'), [
      [9],
    ]);
    expect(find.text('Backup restored'), findsOneWidget);
  });

  testWidgets('restoring a file that is not a backup explains why', (
    tester,
  ) async {
    final go = FakeGo();
    final picker = FakePicker()
      ..next = stageInfo(9, 'photo.cardflip', 'backup');
    go.handle(
      'transfer.Inspect',
      (_) => throw 'this file is not a Cardflip backup',
    );
    await openSettings(tester, picker: picker);
    await scrollTo(tester, find.text('Restore backup'));
    await tester.tap(find.text('Restore backup'));
    await tester.pumpAndSettle();
    expect(find.text('This file is not a Cardflip backup'), findsOneWidget);
    expect(go.callsTo('transfer.Restore'), isEmpty);
  });

  testWidgets('changing the rollover hour is sent to Go', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    expect(find.text('04:00'), findsOneWidget);
    await tester.tap(find.byKey(const Key('rollover')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('06:00').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetRolloverHour'), [
      [6],
    ]);
    expect(find.text('06:00'), findsOneWidget);
  });

  testWidgets('export saves cardflip-<date>.cardflip', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    go.handle(
      'transfer.ExportBackup',
      (_) => {'name': 'cardflip-2026-10-05.cardflip', 'data': 'UEsDBA=='},
    );
    await openSettings(tester, picker: picker);
    await scrollTo(tester, find.text('Export backup'));
    await tester.tap(find.text('Export backup'));
    await tester.pumpAndSettle();
    expect(picker.saved.keys, ['cardflip-2026-10-05.cardflip']);
    expect(
      picker.saved.values.single,
      Uint8List.fromList([0x50, 0x4b, 0x03, 0x04]),
    );
    expect(
      find.text('Backup saved as cardflip-2026-10-05.cardflip'),
      findsOneWidget,
    );
  });

  testWidgets('export collection saves a .colpkg', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    await openSettings(tester, picker: picker);
    await scrollTo(tester, find.text('Export collection'));
    await tester.tap(find.text('Export collection'));
    await tester.pumpAndSettle();
    // Scheduling and presets are always included.
    expect(find.byKey(const Key('export-scheduling')), findsNothing);
    expect(find.byKey(const Key('export-presets')), findsNothing);
    await tester.tap(find.text('Include media'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Export'));
    await tester.pumpAndSettle();
    expect(go.callsTo('transfer.ExportCollection'), [
      [
        {
          'scheduling': false,
          'presets': false,
          'media': false,
          'legacy': false,
        },
      ],
    ]);
    expect(picker.saved.keys, ['collection-2026-10-13.colpkg']);
    expect(find.text('Exported collection-2026-10-13.colpkg'), findsOneWidget);
  });

  testWidgets('export collection as text saves every card', (tester) async {
    final go = FakeGo();
    final picker = FakePicker();
    await openSettings(tester, picker: picker);
    await scrollTo(tester, find.text('Export collection'));
    await tester.tap(find.text('Export collection'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('export-format')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Cards in plain text (.txt)').last);
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Export'));
    await tester.pumpAndSettle();
    expect(go.callsTo('transfer.ExportCardsText'), [
      [
        {'deckId': 0, 'cardIds': <int>[], 'all': true},
      ],
    ]);
    expect(go.callsTo('transfer.ExportCollection'), isEmpty);
    expect(picker.saved.keys, ['cards-2026-10-13.txt']);
  });

  testWidgets('web: one-time notice when storage is not persistent', (
    tester,
  ) async {
    final go = FakeGo();
    go.status = {
      ...go.status,
      'persistent': false,
      'showPersistenceNotice': true,
    };
    await pumpApp(tester, isWeb: true);
    expect(find.text('Your browser may clear Cardflip data'), findsOneWidget);
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.PersistenceNoticeShown'), hasLength(1));
    await tester.tap(find.text('Settings').last);
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.textContaining('may clear your collection'),
      200,
    );
    expect(find.textContaining('may clear your collection'), findsOneWidget);
  });

  testWidgets('About shows the version', (tester) async {
    FakeGo();
    await openSettings(tester);
    await tester.dragUntilVisible(
      find.text('Cardflip'),
      find.byType(ListView),
      const Offset(0, -200),
    );
    expect(find.textContaining('Version 0.1.0'), findsOneWidget);
  });

  group('licenses', () {
    setUpAll(() {
      LicenseRegistry.addLicense(dartPackageLicenses);
      registerLicenses();
    });

    /// The license page's package list.
    Finder packageList() => find
        .descendant(
          of: find.byType(LicensePage),
          matching: find.byType(Scrollable),
        )
        .first;

    /// Pumps until the license page has parsed the licenses, which it does
    /// in isolates, outside the fake clock.
    Future<void> waitForLicenses(WidgetTester tester) async {
      for (var i = 0; i < 200; i++) {
        await tester.runAsync(
          () => Future<void>.delayed(const Duration(milliseconds: 50)),
        );
        await tester.pump();
        if (find.byType(CircularProgressIndicator).evaluate().isEmpty) break;
      }
      await tester.pumpAndSettle();
    }

    /// Opens the license page from Settings.
    Future<void> openLicenses(WidgetTester tester, String entry) async {
      await scrollTo(tester, find.text(entry));
      await tester.tap(find.text(entry));
      await waitForLicenses(tester);
    }

    testWidgets('About names the license (spec: License named)', (
      tester,
    ) async {
      FakeGo();
      await openSettings(tester);
      await scrollTo(tester, find.text('Open-source licenses'));
      expect(find.textContaining('Version 0.1.0'), findsOneWidget);
      expect(
        find.textContaining(
          'Copyright © 2026 madzione. Free software under the GNU GPL version 3 '
          'or later, with absolutely no warranty.',
        ),
        findsOneWidget,
      );
    });

    testWidgets('Source code opens the repository (spec)', (tester) async {
      final opened = <Uri>[];
      final open = SettingsScreen.openUrl;
      SettingsScreen.openUrl = (url) async {
        opened.add(url);
        return true;
      };
      addTearDown(() => SettingsScreen.openUrl = open);
      FakeGo();
      await openSettings(tester);
      await scrollTo(tester, find.text('Source code'));
      await tester.tap(find.text('Source code'));
      await tester.pumpAndSettle();
      expect(opened, [Uri.parse(repositoryUrl)]);
    });

    testWidgets('Open-source licenses opens the page (spec: Open the page)', (
      tester,
    ) async {
      FakeGo();
      await openSettings(tester);
      await openLicenses(tester, 'Open-source licenses');
      expect(find.byType(LicensePage), findsOneWidget);
      expect(find.text('Cardflip'), findsWidgets);
      expect(find.text('0.1.0'), findsOneWidget);
      expect(find.textContaining('Copyright © 2026 madzione'), findsWidgets);
      expect(
        find.textContaining('It comes with absolutely no warranty.'),
        findsOneWidget,
      );
      // In the page's order, which ignores case.
      for (final package in [
        'Anki',
        'cardflip',
        'fluffy',
        'fsrs-rs',
        'github.com/klauspost/compress',
        'Go',
      ]) {
        await tester.scrollUntilVisible(
          find.text(package),
          200,
          scrollable: packageList(),
        );
        expect(find.text(package), findsOneWidget, reason: package);
      }
    });

    testWidgets('selecting flutter_quill shows its MIT license (spec: '
        'Read a license)', (tester) async {
      FakeGo();
      await openSettings(tester);
      await openLicenses(tester, 'Open-source licenses');
      await tester.scrollUntilVisible(
        find.text('flutter_quill'),
        200,
        scrollable: packageList(),
      );
      await tester.tap(find.text('flutter_quill'));
      await waitForLicenses(tester);
      expect(find.text('MIT License'), findsWidgets);
      expect(
        find.textContaining(
          'Flutter Quill project and open source contributors',
        ),
        findsWidgets,
      );
    });

    testWidgets('the entry and the page are German (spec: German interface)', (
      tester,
    ) async {
      addTearDown(() => appL10n = english);
      final go = FakeGo();
      go.settings = {...go.settings, 'language': 'de'};
      await pumpApp(tester, size: desktop);
      await tester.tap(find.text('Einstellungen').last);
      await tester.pumpAndSettle();
      await openLicenses(tester, 'Open-Source-Lizenzen');
      expect(find.text('Lizenzen'), findsWidgets);
      expect(
        find.textContaining('ohne jede Gewährleistung bereitgestellt'),
        findsOneWidget,
      );
    });
  });

  testWidgets('auto-continue for learning cards can be turned on', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final tile = find.widgetWithText(
      SwitchListTile,
      'Continue automatically when learning cards are due',
    );
    expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    await tester.tap(tile);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetAutoContinueLearning'), [
      [true],
    ]);
    expect(tester.widget<SwitchListTile>(tile).value, isTrue);
  });

  testWidgets('the learn ahead limit defaults to 20 minutes and can be off', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final dropdown = find.byKey(const Key('learnAhead'));
    await tester.ensureVisible(dropdown);
    expect(tester.widget<DropdownButton<int>>(dropdown).value, 20);
    await tester.tap(dropdown);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Off').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetLearnAheadMinutes'), [
      [0],
    ]);
    expect(tester.widget<DropdownButton<int>>(dropdown).value, 0);
  });

  testWidgets('the timebox is off by default and can be set', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    final dropdown = find.byKey(const Key('timebox'));
    await tester.ensureVisible(dropdown);
    expect(tester.widget<DropdownButton<int>>(dropdown).value, 0);
    // Enable timebox (spec): 15 minutes is stored.
    await tester.tap(dropdown);
    await tester.pumpAndSettle();
    await tester.tap(find.text('15 min').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetTimeboxMinutes'), [
      [15],
    ]);
    expect(go.settings['timeboxMinutes'], 15);
    expect(tester.widget<DropdownButton<int>>(dropdown).value, 15);
  });

  testWidgets('the review preferences are sent to Go', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    for (final (label, call, on) in [
      ('Show remaining card count', 'app.SetShowRemainingCounts', true),
      (
        'Show next review time above answer buttons',
        'app.SetShowIntervals',
        true,
      ),
      ('Space/Enter also answers Good', 'app.SetSpaceAnswersGood', true),
      ('Hide bars during review', 'app.SetHideBarsDuringReview', false),
    ]) {
      final tile = find.widgetWithText(SwitchListTile, label);
      await scrollTo(tester, tile);
      expect(tester.widget<SwitchListTile>(tile).value, on);
      await tester.tap(tile);
      await tester.pumpAndSettle();
      expect(go.callsTo(call), [
        [!on],
      ]);
      expect(tester.widget<SwitchListTile>(tile).value, !on);
    }
    final fade = find.byKey(const Key('question-fade'));
    await scrollTo(tester, fade);
    expect(tester.widget<DropdownButton<int>>(fade).value, 100);
    await tester.tap(fade);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Off').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetQuestionFadeMs'), [
      [0],
    ]);
    expect(go.settings['questionFadeMs'], 0);
    expect(tester.widget<DropdownButton<int>>(fade).value, 0);
  });

  testWidgets('stored review preferences are shown', (tester) async {
    // Setting persists (spec): the stored value is what the screen shows.
    final go = FakeGo();
    go.settings = {
      ...go.settings,
      'showRemainingCounts': false,
      'questionFadeMs': 300,
    };
    await openSettings(tester);
    final tile = find.widgetWithText(
      SwitchListTile,
      'Show remaining card count',
    );
    await scrollTo(tester, tile);
    expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    final fade = find.byKey(const Key('question-fade'));
    await scrollTo(tester, fade);
    expect(tester.widget<DropdownButton<int>>(fade).value, 300);
  });

  group('goals and streaks', () {
    testWidgets('the values save', (tester) async {
      // Goals and streaks settings (progress spec): shown by default, no
      // goal.
      final go = FakeGo();
      await openSettings(tester);
      final show = find.byKey(const Key('show-progress'));
      await scrollTo(tester, show);
      expect(tester.widget<SwitchListTile>(show).value, isTrue);
      await tester.tap(show);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetShowProgress'), [
        [false],
      ]);
      expect(tester.widget<SwitchListTile>(show).value, isFalse);

      // Show deck progress in the deck list: off by default, and saved.
      final inList = find.byKey(const Key('deck-progress-in-list'));
      await scrollTo(tester, inList);
      expect(tester.widget<SwitchListTile>(inList).value, isFalse);
      await tester.tap(inList);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetDeckProgressInList'), [
        [true],
      ]);
      expect(tester.widget<SwitchListTile>(inList).value, isTrue);

      // Set a goal (spec): 20 minutes.
      final goal = find.byKey(const Key('daily-goal'));
      await scrollTo(tester, goal);
      expect(
        find.descendant(of: goal, matching: find.text('None')),
        findsOneWidget,
      );
      await tester.tap(goal);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Minutes'));
      await tester.pumpAndSettle();
      final value = find.byKey(const Key('daily-goal-value'));
      await tester.enterText(value, '0');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(find.text('Enter a whole number from 1 to 9999.'), findsOneWidget);
      expect(go.callsTo('app.SetDailyGoal'), isEmpty);
      await tester.enterText(value, '20');
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetDailyGoal'), [
        ['minutes', 20],
      ]);
      expect(
        find.descendant(of: goal, matching: find.text('20 minutes a day')),
        findsOneWidget,
      );

      // No goal again.
      await tester.tap(goal);
      await tester.pumpAndSettle();
      expect(tester.widget<TextField>(value).controller!.text, '20');
      await tester.tap(find.text('None').last);
      await tester.pumpAndSettle();
      expect(value, findsNothing);
      await tester.tap(find.text('Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetDailyGoal').last, ['', 0]);
      expect(go.settings['dailyGoalKind'], '');
      expect(
        find.descendant(of: goal, matching: find.text('None')),
        findsOneWidget,
      );
    });

    testWidgets('rest days save and the header follows', (tester) async {
      // Rest days per week (progress spec): 0 by default; the deck list
      // header shows the streak Go computes with them.
      final go = FakeGo()..progress = progressJson(current: 4, cards: 3);
      await pumpApp(tester, size: desktop);
      final header = find.byKey(const Key('progress-header'));
      expect(
        find.descendant(of: header, matching: find.text('4 days')),
        findsOneWidget,
      );
      await tester.tap(find.text('Settings').last);
      await tester.pumpAndSettle();
      final rest = find.byKey(const Key('rest-days'));
      await scrollTo(tester, rest);
      expect(tester.widget<DropdownButton<int>>(rest).value, 0);
      await tester.tap(rest);
      await tester.pumpAndSettle();
      go.progress = progressJson(current: 10, cards: 3);
      await tester.tap(find.text('1').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetRestDaysPerWeek'), [
        [1],
      ]);
      expect(go.settings['restDaysPerWeek'], 1);
      expect(tester.widget<DropdownButton<int>>(rest).value, 1);
      go.emitChange(kind: 'settings', settings: true);
      await tester.tap(find.text('Decks').last);
      await tester.pumpAndSettle();
      expect(
        find.descendant(of: header, matching: find.text('10 days')),
        findsOneWidget,
      );
    });

    testWidgets('hiding removes the deck list header', (tester) async {
      // Hidden (spec).
      final go = FakeGo()..progress = progressJson(current: 12, cards: 3);
      await pumpApp(tester, size: desktop);
      final header = find.byKey(const Key('progress-header'));
      expect(header, findsOneWidget);
      await tester.tap(find.text('Settings').last);
      await tester.pumpAndSettle();
      final show = find.byKey(const Key('show-progress'));
      await scrollTo(tester, show);
      await tester.tap(show);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Decks').last);
      await tester.pumpAndSettle();
      expect(header, findsNothing);
      expect(go.settings['showProgress'], isFalse);
    });
  });

  group('appearance', () {
    testWidgets('UI size, minimalist mode and reduce motion default and are '
        'sent to Go', (tester) async {
      final go = FakeGo();
      await openSettings(tester);
      final size = find.byKey(const Key('ui-size'));
      await scrollTo(tester, size);
      expect(tester.widget<DropdownButton<int>>(size).value, 100);
      await tester.tap(size);
      await tester.pumpAndSettle();
      await tester.tap(find.text('120%').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetUIScale'), [
        [120],
      ]);
      expect(go.settings['uiScale'], 120);
      expect(tester.widget<DropdownButton<int>>(size).value, 120);

      final minimalist = find.widgetWithText(SwitchListTile, 'Minimalist mode');
      await scrollTo(tester, minimalist);
      expect(tester.widget<SwitchListTile>(minimalist).value, isFalse);
      await tester.tap(minimalist);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetMinimalist'), [
        [true],
      ]);
      expect(tester.widget<SwitchListTile>(minimalist).value, isTrue);

      final motion = find.byKey(const Key('reduce-motion'));
      await scrollTo(tester, motion);
      expect(tester.widget<DropdownButton<String>>(motion).value, 'system');
      expect(
        find.descendant(of: motion, matching: find.text('Follow system')),
        findsOneWidget,
      );
      await tester.tap(motion);
      await tester.pumpAndSettle();
      await tester.tap(find.text('On').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetReduceMotion'), [
        ['on'],
      ]);
      expect(tester.widget<DropdownButton<String>>(motion).value, 'on');
    });

    testWidgets('the UI size offers 75 to 200 % in steps of 5 %', (
      tester,
    ) async {
      FakeGo();
      await openSettings(tester);
      final size = find.byKey(const Key('ui-size'));
      await scrollTo(tester, size);
      final items = tester.widget<DropdownButton<int>>(size).items!;
      expect(items.map((i) => i.value), [for (var p = 75; p <= 200; p += 5) p]);
    });

    testWidgets('stored appearance settings are shown and applied', (
      tester,
    ) async {
      // Changes persist (spec): the stored values are what the screen
      // shows.
      final go = FakeGo();
      go.settings = {
        ...go.settings,
        'uiScale': 80,
        'minimalist': true,
        'reduceMotion': 'off',
      };
      await openSettings(tester);
      final size = find.byKey(const Key('ui-size'));
      await scrollTo(tester, size);
      expect(tester.widget<DropdownButton<int>>(size).value, 80);
      expect(
        tester
            .widget<SwitchListTile>(
              find.widgetWithText(SwitchListTile, 'Minimalist mode'),
            )
            .value,
        isTrue,
      );
      expect(
        tester
            .widget<DropdownButton<String>>(
              find.byKey(const Key('reduce-motion')),
            )
            .value,
        'off',
      );
      final context = tester.element(size);
      expect(
        MediaQuery.sizeOf(context).width,
        closeTo(desktop.width / 0.8, 0.01),
      );
      expect(Theme.of(context).cardTheme.elevation, 0);
      expect(MediaQuery.disableAnimationsOf(context), isFalse);
    });

    testWidgets('the UI size switches back from 75 to 100 % (spec)', (
      tester,
    ) async {
      // Every visible element responds at every UI size: a tap that misses
      // its target fails the test.
      WidgetController.hitTestWarningShouldBeFatal = true;
      addTearDown(() => WidgetController.hitTestWarningShouldBeFatal = false);
      final go = FakeGo();
      go.settings = {...go.settings, 'uiScale': 75};
      await openSettings(tester);
      final size = find.byKey(const Key('ui-size'));
      await scrollTo(tester, size);
      expect(tester.widget<DropdownButton<int>>(size).value, 75);
      await tester.tap(size);
      await tester.pumpAndSettle();
      await tester.tap(find.text('100%').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetUIScale'), [
        [100],
      ]);
      expect(go.settings['uiScale'], 100);
      expect(tester.widget<DropdownButton<int>>(size).value, 100);
    });

    testWidgets('a failed save keeps the old UI size', (tester) async {
      final go = FakeGo();
      go.handle('app.SetUIScale', (_) => throw 'could not save: disk full');
      await openSettings(tester);
      final size = find.byKey(const Key('ui-size'));
      await scrollTo(tester, size);
      await tester.tap(size);
      await tester.pumpAndSettle();
      await tester.tap(find.text('150%').last);
      await tester.pumpAndSettle();
      expect(find.text('Could not save: disk full'), findsOneWidget);
      expect(tester.widget<DropdownButton<int>>(size).value, 100);
      expect(
        MediaQuery.sizeOf(tester.element(size)).width,
        closeTo(desktop.width, 0.01),
      );
    });
  });

  testWidgets('the audio preferences are sent to Go', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    for (final (label, call) in [
      ('Show play buttons on cards with audio', 'app.SetShowPlayButtons'),
      ('Interrupt current audio when answering', 'app.SetInterruptAudio'),
    ]) {
      final tile = find.widgetWithText(SwitchListTile, label);
      await scrollTo(tester, tile);
      expect(tester.widget<SwitchListTile>(tile).value, isTrue);
      await tester.tap(tile);
      await tester.pumpAndSettle();
      expect(go.callsTo(call), [
        [false],
      ]);
      expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    }
  });

  testWidgets('"Strip formatting when pasting" is off and sent to Go', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final tile = find.byKey(const Key('strip-paste-formatting'));
    await scrollTo(tester, tile);
    expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    expect(find.text('Strip formatting when pasting'), findsOneWidget);
    await tester.tap(tile);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetStripPasteFormatting'), [
      [true],
    ]);
    expect(tester.widget<SwitchListTile>(tile).value, isTrue);
  });

  testWidgets('"Paste images as PNG" is off and sent to Go', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    final tile = find.byKey(const Key('paste-images-as-png'));
    await scrollTo(tester, tile);
    expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    expect(find.text('Paste images as PNG'), findsOneWidget);
    await tester.tap(tile);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetPasteImagesAsPng'), [
      [true],
    ]);
    expect(tester.widget<SwitchListTile>(tile).value, isTrue);
  });

  testWidgets('"Ignore accents in search" is off and sent to Go', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final tile = find.byKey(const Key('ignore-accents'));
    await scrollTo(tester, tile);
    expect(tester.widget<SwitchListTile>(tile).value, isFalse);
    expect(find.text('Ignore accents in search'), findsOneWidget);
    await tester.tap(tile);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetIgnoreAccents'), [
      [true],
    ]);
    expect(tester.widget<SwitchListTile>(tile).value, isTrue);
  });

  testWidgets('"Default browser search" is checked and sent to Go', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final tile = find.byKey(const Key('default-search'));
    await scrollTo(tester, tile);
    expect(
      find.descendant(of: tile, matching: find.text('All cards')),
      findsOneWidget,
    );

    // An invalid search is refused with Go's error; the dialog stays open.
    await tester.tap(tile);
    await tester.pumpAndSettle();
    final field = find.byKey(const Key('default-search-field'));
    await tester.enterText(field, 'is:bogus');
    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await tester.pumpAndSettle();
    expect(find.text(FakeGo.bogusSearchError), findsOneWidget);
    expect(go.settings['defaultSearch'], '');

    await tester.enterText(field, 'is:due');
    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await tester.pumpAndSettle();
    expect(find.byType(AlertDialog), findsNothing);
    expect(go.callsTo('app.SetDefaultSearch'), [
      ['is:bogus'],
      ['is:due'],
    ]);
    expect(go.settings['defaultSearch'], 'is:due');
    expect(
      find.descendant(of: tile, matching: find.text('is:due')),
      findsOneWidget,
    );
  });

  testWidgets('"Deck for new cards" is "Current deck" and sent to Go', (
    tester,
  ) async {
    final go = FakeGo();
    await openSettings(tester);
    final picker = find.byKey(const Key('deck-for-new-cards'));
    await scrollTo(tester, picker);
    expect(find.text('Deck for new cards'), findsOneWidget);
    expect(tester.widget<DropdownButton<bool>>(picker).value, isFalse);
    expect(find.text('Current deck'), findsOneWidget);
    await tester.tap(picker);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Change deck depending on note type').last);
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetDeckByNoteType'), [
      [true],
    ]);
    expect(tester.widget<DropdownButton<bool>>(picker).value, isTrue);
    expect(go.settings['deckByNoteType'], isTrue);
  });

  group('empty cards', () {
    Future<void> openEmptyCards(WidgetTester tester) async {
      await scrollTo(tester, find.text('Empty cards'));
      await tester.tap(find.text('Empty cards'));
      await tester.pumpAndSettle();
    }

    testWidgets('nothing to remove', (tester) async {
      FakeGo();
      await openSettings(tester);
      await openEmptyCards(tester);
      // Nothing to remove (spec): the dialog says there are no empty cards.
      expect(find.text('There are no empty cards.'), findsOneWidget);
      await tester.tap(find.text('Close'));
      await tester.pumpAndSettle();
      expect(find.text('There are no empty cards.'), findsNothing);
    });

    testWidgets('lists empty cards by note and deletes them when confirmed', (
      tester,
    ) async {
      final go = FakeGo()
        ..emptyNotes = [
          emptyNote(1, 'Paris is the capital of France', {11: 'Cloze 2'}),
          emptyNote(2, 'hola', {21: 'Card 2'}),
          emptyNote(3, 'perro', {31: 'Card 2'}),
          emptyNote(4, 'gato', {41: 'Card 1'}, allEmpty: true),
        ];
      await openSettings(tester);
      await openEmptyCards(tester);
      expect(find.textContaining('4 empty cards in 4 notes'), findsOneWidget);
      expect(find.text('Cloze 2'), findsOneWidget);
      expect(find.text('hola'), findsOneWidget);
      // All cards empty (spec): the list marks the note.
      expect(find.text('Card 1 · one card is kept'), findsOneWidget);
      expect(
        find.textContaining('1 note has only empty cards'),
        findsOneWidget,
      );

      // Cancel deletes nothing.
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteEmptyCards'), isEmpty);

      // Remove empty cards (spec): confirming deletes them in one call.
      await openEmptyCards(tester);
      await tester.tap(find.text('Delete 3 cards'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteEmptyCards').single.single, [
        11,
        21,
        31,
        41,
      ]);
      expect(find.text('Deleted 3 empty cards'), findsOneWidget);
      expect(find.text('Delete 3 cards'), findsNothing);

      // The note with only empty cards keeps its card and stays listed.
      await openEmptyCards(tester);
      expect(find.textContaining('1 empty card in 1 note'), findsOneWidget);
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Delete 0 cards'),
            )
            .onPressed,
        isNull,
      );
    });
  });

  group('deletion log', () {
    testWidgets('export saves the log as a text file', (tester) async {
      FakeGo().deletedNotes = 3;
      final picker = FakePicker();
      await openSettings(tester, picker: picker);
      await scrollTo(tester, find.text('Export deletion log'));
      await tester.tap(find.text('Export deletion log'));
      await tester.pumpAndSettle();
      expect(picker.saved.keys, ['cardflip-deleted-2026-10-13.txt']);
      expect(utf8.decode(picker.saved.values.single), '#separator:tab\n');
      expect(
        find.text(
          'Deletion log with 3 notes saved as cardflip-deleted-2026-10-13.txt',
        ),
        findsOneWidget,
      );
    });

    testWidgets('an empty log is not saved', (tester) async {
      FakeGo();
      final picker = FakePicker();
      await openSettings(tester, picker: picker);
      await scrollTo(tester, find.text('Export deletion log'));
      await tester.tap(find.text('Export deletion log'));
      await tester.pumpAndSettle();
      expect(picker.saved, isEmpty);
      expect(find.text('The deletion log is empty'), findsOneWidget);
    });

    testWidgets('clear asks first', (tester) async {
      final go = FakeGo()..deletedNotes = 2;
      await openSettings(tester);
      await scrollTo(tester, find.text('Clear deletion log'));
      await tester.tap(find.text('Clear deletion log'));
      await tester.pumpAndSettle();
      expect(find.text('Clear the deletion log?'), findsOneWidget);

      // Cancel keeps the log.
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.ClearDeletionLog'), isEmpty);

      await tester.tap(find.text('Clear deletion log'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Clear'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.ClearDeletionLog'), hasLength(1));
      expect(go.deletedNotes, 0);
      expect(find.text('Deletion log cleared'), findsOneWidget);
    });
  });

  group('automatic backups', () {
    final morning = DateTime.utc(2026, 10, 13, 6, 30);
    final yesterday = DateTime.utc(2026, 10, 12, 18);

    FakeGo withBackups() => FakeGo()
      ..backups = [
        autoBackup(morning, size: 3 * 1024 * 1024),
        autoBackup(yesterday, size: 1024 * 1024),
      ];

    testWidgets('lists backups with date, time and space used', (tester) async {
      withBackups();
      await openSettings(tester);
      await scrollTo(tester, find.text(formatDateTime(yesterday)));
      expect(find.text('2 automatic backups'), findsOneWidget);
      expect(find.text('4.0 MB used on this device'), findsOneWidget);
      expect(find.text(formatDateTime(morning)), findsOneWidget);
      expect(find.text('3.0 MB'), findsOneWidget);
      // Newest first.
      expect(
        tester.getTopLeft(find.text(formatDateTime(morning))).dy,
        lessThan(tester.getTopLeft(find.text(formatDateTime(yesterday))).dy),
      );
    });

    testWidgets('restore asks first and keeps media', (tester) async {
      final go = withBackups();
      await openSettings(tester);
      final row = find.ancestor(
        of: find.text(formatDateTime(morning)),
        matching: find.byType(ListTile),
      );
      await scrollTo(tester, row);
      final restore = find.descendant(of: row, matching: find.text('Restore'));
      await tester.tap(restore);
      await tester.pumpAndSettle();
      expect(find.text('Replace your collection?'), findsOneWidget);
      expect(
        find.textContaining('automatic backup from ${formatDateTime(morning)}'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Media files stay as they are'),
        findsOneWidget,
      );
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.RestoreBackup'), isEmpty);

      await tester.tap(restore);
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Restore'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.RestoreBackup'), [
        [autoBackup(morning)['name']],
      ]);
      expect(find.text('Backup restored'), findsOneWidget);
    });

    testWidgets('create backup now adds one at the top', (tester) async {
      final go = withBackups();
      await openSettings(tester);
      await scrollTo(tester, find.text('Create backup now'));
      expect(find.text('Ctrl+Alt+B'), findsOneWidget);
      await tester.tap(find.text('Create backup now'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.CreateBackup'), hasLength(1));
      expect(find.text('Backup created'), findsOneWidget);
      expect(find.text('3 automatic backups'), findsOneWidget);
      final made = DateTime.parse(go.backups.first['createdAt']! as String);
      expect(
        tester.getTopLeft(find.text(formatDateTime(made))).dy,
        lessThan(tester.getTopLeft(find.text(formatDateTime(morning))).dy),
      );
    }, variant: TargetPlatformVariant.only(TargetPlatform.windows));

    testWidgets('backup entry on a phone (spec)', (tester) async {
      withBackups();
      await openSettings(tester);
      await scrollTo(tester, find.text('Create backup now'));
      expect(find.text('Ctrl+Alt+B'), findsNothing);
    }, variant: TargetPlatformVariant.only(TargetPlatform.android));

    testWidgets('Ctrl+Alt+B makes a backup from anywhere', (tester) async {
      final go = withBackups();
      await pumpApp(tester, size: desktop);
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.altLeft,
        LogicalKeyboardKey.keyB,
      ]);
      expect(go.callsTo('app.CreateBackup'), hasLength(1));
      expect(find.text('Backup created'), findsOneWidget);
      // Without Alt it is not the shortcut.
      await pressKeys(tester, [
        LogicalKeyboardKey.controlLeft,
        LogicalKeyboardKey.keyB,
      ]);
      expect(go.callsTo('app.CreateBackup'), hasLength(1));

      await tester.tap(find.text('Settings').last);
      await tester.pumpAndSettle();
      await scrollTo(tester, find.text('3 automatic backups'));
      final made = DateTime.parse(go.backups.first['createdAt']! as String);
      expect(find.text(formatDateTime(made)), findsOneWidget);
    });

    testWidgets('Cmd+Option+B on macOS', (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
      try {
        final go = withBackups();
        await openSettings(tester);
        await pressKeys(tester, [
          LogicalKeyboardKey.metaLeft,
          LogicalKeyboardKey.altLeft,
          LogicalKeyboardKey.keyB,
        ]);
        expect(go.callsTo('app.CreateBackup'), hasLength(1));
        await scrollTo(tester, find.text('Create backup now'));
        expect(find.text('⌘⌥B'), findsOneWidget);
      } finally {
        debugDefaultTargetPlatformOverride = null;
      }
    });

    testWidgets('Settings reloads the list when it is shown', (tester) async {
      final go = withBackups();
      await pumpApp(tester, size: desktop);
      // A periodic backup made by Go while another screen was shown.
      go.backups = [autoBackup(DateTime.utc(2026, 10, 13, 9)), ...go.backups];
      await tester.tap(find.text('Settings').last);
      await tester.pumpAndSettle();
      await scrollTo(tester, find.text('3 automatic backups'));
      expect(find.text('3 automatic backups'), findsOneWidget);
    });

    testWidgets('interval and counts are sent to Go', (tester) async {
      final go = FakeGo();
      await openSettings(tester);
      await scrollTo(tester, find.byKey(const Key('keepMonthly')));
      expect(find.text('Every 30 min'), findsOneWidget);
      await tester.ensureVisible(find.byKey(const Key('backupInterval')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('backupInterval')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Off').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetBackupSettings').last, [
        {'intervalMinutes': 0, 'daily': 8, 'weekly': 8, 'monthly': 8},
      ]);
      expect(find.text('Off'), findsOneWidget);

      await tester.tap(find.byKey(const Key('keepWeekly')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('4').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetBackupSettings').last, [
        {'intervalMinutes': 0, 'daily': 8, 'weekly': 4, 'monthly': 8},
      ]);
    });

    testWidgets('long lists are collapsed', (tester) async {
      final go = FakeGo();
      go.backups = [
        for (var i = 0; i < 8; i++)
          autoBackup(
            DateTime.utc(2026, 10, 13, 10).subtract(Duration(hours: i)),
          ),
      ];
      await openSettings(tester);
      await scrollTo(tester, find.text('Show all 8'));
      expect(find.widgetWithText(TextButton, 'Restore'), findsNWidgets(5));
      await tester.tap(find.text('Show all 8'));
      await tester.pumpAndSettle();
      await scrollTo(tester, find.text('Show fewer'));
      expect(
        find.text(formatDateTime(DateTime.utc(2026, 10, 13, 3))),
        findsOneWidget,
      );
    });
  });

  testWidgets('renaming a flag calls app.SetFlagName', (tester) async {
    final go = FakeGo();
    await openSettings(tester);
    await scrollTo(tester, find.byKey(const Key('flag1')));
    expect(
      find.descendant(
        of: find.byKey(const Key('flag1')),
        matching: find.text('Red'),
      ),
      findsOneWidget,
    );

    // Rename a flag (spec).
    await tester.tap(find.byKey(const Key('flag1')));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'Fix later');
    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetFlagName'), [
      [1, 'Fix later'],
    ]);
    final tile = find.byKey(const Key('flag1'));
    expect(
      find.descendant(of: tile, matching: find.text('Fix later')),
      findsOneWidget,
    );
    expect(
      find.descendant(of: tile, matching: find.text('Red')),
      findsOneWidget,
    );

    // Clearing the name restores the colour name.
    await tester.tap(tile);
    await tester.pumpAndSettle();
    expect(find.widgetWithText(TextField, 'Fix later'), findsOneWidget);
    await tester.enterText(find.byType(TextField), '');
    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await tester.pumpAndSettle();
    expect(go.callsTo('app.SetFlagName').last, [1, '']);
    expect(
      find.descendant(of: tile, matching: find.text('Fix later')),
      findsNothing,
    );
    expect(
      find.descendant(of: tile, matching: find.text('Red')),
      findsOneWidget,
    );
  });

  group('check database', () {
    Future<void> openCheckDatabase(WidgetTester tester) async {
      await scrollTo(tester, find.text('Check database'));
      await tester.tap(find.text('Check database'));
      // The progress bar may be indeterminate, so the dialog never settles.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
    }

    // Run the check (app-shell spec): progress, then the report.
    testWidgets('shows progress, then the report', (tester) async {
      final go = FakeGo();
      final events = StreamController<Object?>();
      go.checkDatabaseEvents = () => events.stream;
      await openSettings(tester);
      await openCheckDatabase(tester);
      expect(go.callsTo('app.CheckDatabase'), hasLength(1));
      expect(find.text('Backing up the collection'), findsOneWidget);

      events.add(checkProgress('cards', 1, 4));
      await tester.pump();
      expect(find.text('Checking cards'), findsOneWidget);
      final bar = tester.widget<LinearProgressIndicator>(
        find.byKey(const Key('check-progress')),
      );
      expect(bar.value, 0.25);
      // It cannot be dismissed while it runs.
      expect(find.text('Close'), findsNothing);
      await tester.tapAt(const Offset(5, 5));
      await tester.pump();
      expect(find.text('Checking cards'), findsOneWidget);

      events.add(
        checkDone(
          checkReport(
            problems: [
              ('cardsWithoutNote', 1, '1 card without a note was removed.'),
              ('tags', 2, '2 notes with duplicate or invalid tags were fixed.'),
            ],
            orphanedReviews: 3,
          ),
        ),
      );
      await events.close();
      await tester.pumpAndSettle();
      expect(find.text('1 card without a note was removed.'), findsOneWidget);
      expect(
        find.text('2 notes with duplicate or invalid tags were fixed.'),
        findsOneWidget,
      );
      expect(
        find.text(
          '3 review log entries belong to deleted cards. They are kept for '
          'statistics.',
        ),
        findsOneWidget,
      );
      expect(find.text('No problems were found.'), findsNothing);
      await tester.tap(find.text('Close'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('check-report')), findsNothing);
    });

    // Healthy collection (persistence spec).
    testWidgets('says when no problems were found', (tester) async {
      FakeGo();
      await openSettings(tester);
      await openCheckDatabase(tester);
      await tester.pumpAndSettle();
      expect(find.text('No problems were found.'), findsOneWidget);
      expect(find.textContaining('review log entr'), findsNothing);
    });

    testWidgets('shows why the check failed', (tester) async {
      final go = FakeGo();
      go.checkDatabaseEvents = () => Stream.fromIterable([
        checkProgress('backup', 0, 1),
        {
          'stage': '',
          'done': 0,
          'total': 0,
          'report': null,
          'error': 'the collection could not be backed up first, so nothing was changed: disk full',
        },
      ]);
      await openSettings(tester);
      await openCheckDatabase(tester);
      await tester.pumpAndSettle();
      expect(
        find.text(
          'The collection could not be backed up first, so nothing was changed: disk full',
        ),
        findsOneWidget,
      );
      await tester.tap(find.text('Close'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('check-error')), findsNothing);
    });
  });

  group('check media', () {
    Future<void> openCheckMedia(WidgetTester tester) async {
      await scrollTo(tester, find.text('Check media'));
      await tester.tap(find.text('Check media'));
      await tester.pumpAndSettle();
    }

    Finder button(String label) => find.widgetWithText(OutlinedButton, label);

    testWidgets('nothing to report', (tester) async {
      FakeGo();
      await openSettings(tester);
      await openCheckMedia(tester);
      expect(
        find.text('Every media file is used by a note, a template or styling.'),
        findsOneWidget,
      );
      expect(
        find.text('Every file that notes refer to is there.'),
        findsOneWidget,
      );
      expect(find.text('The trash is empty.'), findsOneWidget);
      for (final label in [
        'Delete unused',
        'Tag missing',
        'Restore deleted',
        'Empty trash',
      ]) {
        expect(button(label), findsNothing);
      }
      await tester.tap(find.text('Close'));
      await tester.pumpAndSettle();
      expect(find.text('The trash is empty.'), findsNothing);
    });

    // Open the report (app-shell spec), unused and missing (persistence
    // spec).
    testWidgets('shows unused and missing files and the trash size', (
      tester,
    ) async {
      FakeGo()
        ..unusedMedia = [
          {'name': 'old.jpg', 'size': 2048},
        ]
        ..missingMedia = ['gone.mp3']
        ..missingMediaNotes = 3
        ..mediaTrash = [
          {'name': 'a.png', 'size': 1000},
          {'name': 'b.png', 'size': 1048},
        ];
      await openSettings(tester);
      await openCheckMedia(tester);
      expect(find.text('Unused files'), findsOneWidget);
      expect(
        find.textContaining('1 file (2.0 KB) that no note, template'),
        findsOneWidget,
      );
      expect(find.text('old.jpg'), findsOneWidget);
      expect(find.text('2.0 KB'), findsOneWidget);
      expect(
        find.text('1 file that 3 notes refer to is not in your media.'),
        findsOneWidget,
      );
      expect(find.text('gone.mp3'), findsOneWidget);
      expect(
        find.textContaining('2 files (2.0 KB). Files are removed'),
        findsOneWidget,
      );
    });

    // Restore deleted files (persistence spec).
    testWidgets('deletes unused files into the trash and restores them', (
      tester,
    ) async {
      final go = FakeGo()
        ..unusedMedia = [
          {'name': 'old.jpg', 'size': 2048},
          {'name': 'older.png', 'size': 10},
        ];
      await openSettings(tester);
      await openCheckMedia(tester);

      // Cancel deletes nothing.
      await tester.tap(button('Delete unused'));
      await tester.pumpAndSettle();
      expect(find.text('Delete unused files?'), findsOneWidget);
      expect(
        find.textContaining('Moves 2 files (2.0 KB) to the media trash'),
        findsOneWidget,
      );
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteUnusedMedia'), isEmpty);

      await tester.tap(button('Delete unused'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteUnusedMedia'), hasLength(1));
      expect(find.text('Moved 2 files to the media trash'), findsOneWidget);
      // The report is checked again.
      expect(find.text('old.jpg'), findsNothing);
      expect(
        find.textContaining('2 files (2.0 KB). Files are removed'),
        findsOneWidget,
      );

      await tester.tap(button('Restore deleted'));
      await tester.pumpAndSettle();
      expect(find.text('Restore deleted files?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Restore'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.RestoreMediaTrash'), hasLength(1));
      expect(find.text('Restored 2 files'), findsOneWidget);
      expect(find.text('old.jpg'), findsOneWidget);
      expect(find.text('The trash is empty.'), findsOneWidget);
    });

    // Find notes to fix (persistence spec).
    testWidgets('tags notes with missing files when confirmed', (tester) async {
      final go = FakeGo()
        ..missingMedia = ['gone.mp3', 'lost.png']
        ..missingMediaNotes = 3;
      await openSettings(tester);
      await openCheckMedia(tester);
      await tester.tap(button('Tag missing'));
      await tester.pumpAndSettle();
      expect(find.text('Tag notes with missing files?'), findsOneWidget);
      expect(
        find.textContaining('Adds the tag missing-media to 3 notes'),
        findsOneWidget,
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Tag'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.TagMissingMedia'), hasLength(1));
      expect(find.text('Tagged 3 notes missing-media'), findsOneWidget);
    });

    testWidgets('empties the trash only when confirmed', (tester) async {
      final go = FakeGo()
        ..mediaTrash = [
          {'name': 'a.png', 'size': 3 * 1024 * 1024},
        ];
      await openSettings(tester);
      await openCheckMedia(tester);
      await tester.tap(button('Empty trash'));
      await tester.pumpAndSettle();
      expect(find.text('Empty the media trash?'), findsOneWidget);
      expect(
        find.text('Deletes 1 file (3.0 MB) for good. This cannot be undone.'),
        findsOneWidget,
      );
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.EmptyMediaTrash'), isEmpty);

      await tester.tap(button('Empty trash'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Empty trash'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.EmptyMediaTrash'), hasLength(1));
      expect(find.text('Media trash emptied'), findsOneWidget);
      expect(find.text('The trash is empty.'), findsOneWidget);
    });
  });

  group('language', () {
    setUp(() => addTearDown(() => appL10n = english));

    testWidgets('choosing German translates the interface at once', (
      tester,
    ) async {
      final go = FakeGo();
      await openSettings(tester);
      expect(find.text('System default'), findsOneWidget);
      expect(go.builtinNames?['deck'], 'Default');

      await tester.tap(find.byKey(const Key('language')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Deutsch').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetLanguage'), [
        ['de'],
      ]);
      // The settings and the navigation, without a restart.
      expect(find.text('Darstellung'), findsOneWidget);
      expect(find.text('Sprache'), findsOneWidget);
      expect(find.text('Stapel'), findsWidgets);
      expect(find.text('Settings'), findsNothing);
      expect(appL10n.localeName, 'de');
      // New note types and collections are named in German from now on.
      expect(go.builtinNames?['deck'], 'Standard');
      expect(
        (go.builtinNames?['noteTypes'] as List).first,
        containsPair('name', 'Einfach'),
      );

      await tester.tap(find.byKey(const Key('language')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Systemstandard').last);
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetLanguage').last, ['']);
      expect(find.text('Appearance'), findsOneWidget);
    });

    testWidgets('the system default uses a translated device language', (
      tester,
    ) async {
      tester.platformDispatcher.localesTestValue = const [
        Locale('fi'),
        Locale('de', 'AT'),
      ];
      addTearDown(tester.platformDispatcher.clearLocalesTestValue);
      FakeGo();
      await pumpApp(tester, size: desktop);
      await tester.tap(find.text('Einstellungen').last);
      await tester.pumpAndSettle();
      expect(find.text('Systemstandard'), findsOneWidget);
    });

    testWidgets('the system default falls back to English', (tester) async {
      tester.platformDispatcher.localesTestValue = const [Locale('fi')];
      addTearDown(tester.platformDispatcher.clearLocalesTestValue);
      FakeGo();
      await openSettings(tester);
      expect(find.text('System default'), findsOneWidget);
      expect(find.text('Appearance'), findsOneWidget);
    });
  });

  group('profiles', () {
    /// Opens the profile list from Settings.
    Future<void> openProfiles(WidgetTester tester) async {
      await scrollTo(tester, find.text('Switch profile'));
      await tester.tap(find.text('Switch profile'));
      await tester.pumpAndSettle();
    }

    /// Chooses [action] in the menu of profile [id].
    Future<void> profileMenu(
      WidgetTester tester,
      String id,
      String action,
    ) async {
      await tester.tap(find.byKey(Key('profile-menu-$id')));
      await tester.pumpAndSettle();
      await tester.tap(find.text(action).last);
      await tester.pumpAndSettle();
    }

    testWidgets('Settings names the open profile', (tester) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
        ];
      await openSettings(tester);
      await scrollTo(tester, find.text('Switch profile'));
      expect(
        find.descendant(
          of: find.byKey(const Key('switch-profile')),
          matching: find.text('Anna'),
        ),
        findsOneWidget,
      );
      expect(go.callsTo('app.Profiles'), isNotEmpty);
    });

    testWidgets('add, rename and delete profiles', (tester) async {
      final go = FakeGo()..profileCards = {'p2': 300};
      await openSettings(tester);
      await openProfiles(tester);
      expect(find.text('Profiles'), findsOneWidget);
      expect(find.text('Open now'), findsOneWidget);

      await tester.tap(find.text('Add profile'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const Key('profile-name')), 'Ben');
      await tester.tap(find.widgetWithText(FilledButton, 'Add'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.AddProfile'), [
        ['Ben'],
      ]);
      expect(find.byKey(const Key('profile-p2')), findsOneWidget);

      // A taken name keeps the dialog open with Go's message.
      await profileMenu(tester, 'p2', 'Rename profile');
      await tester.enterText(find.byKey(const Key('profile-name')), 'default');
      await tester.tap(find.widgetWithText(FilledButton, 'Rename'));
      await tester.pumpAndSettle();
      expect(
        find.text('A profile named "Default" already exists.'),
        findsOneWidget,
      );
      await tester.enterText(find.byKey(const Key('profile-name')), 'Benjamin');
      await tester.tap(find.widgetWithText(FilledButton, 'Rename'));
      await tester.pumpAndSettle();
      expect(find.text('Benjamin'), findsOneWidget);

      // The open profile cannot be deleted.
      await tester.tap(find.byKey(const Key('profile-menu-default')));
      await tester.pumpAndSettle();
      expect(find.text('Rename profile'), findsOneWidget);
      expect(find.text('Delete profile'), findsNothing);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(find.text('Rename profile'), findsNothing);
      expect(find.text('Profiles'), findsOneWidget);

      // Deleting names the profile and its cards, and asks first.
      await profileMenu(tester, 'p2', 'Delete profile');
      expect(find.text('Delete “Benjamin”?'), findsOneWidget);
      expect(find.textContaining('with 300 cards'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteProfile'), isEmpty);

      await profileMenu(tester, 'p2', 'Delete profile');
      await tester.tap(find.widgetWithText(FilledButton, 'Delete'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.DeleteProfile'), [
        ['p2'],
      ]);
      expect(find.byKey(const Key('profile-p2')), findsNothing);
      expect(find.text('Profile “Benjamin” deleted'), findsOneWidget);
    });

    testWidgets('deleting a profile whose cards cannot be counted', (
      tester,
    ) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Default'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      go.handle('app.ProfileCardCount', (_) => throw 'unreadable');
      await openSettings(tester);
      await openProfiles(tester);
      await profileMenu(tester, 'p2', 'Delete profile');
      expect(find.text('Delete “Ben”?'), findsOneWidget);
      expect(find.textContaining('with all its cards'), findsOneWidget);
    });

    testWidgets('opening another profile starts the app over', (tester) async {
      final go = FakeGo()
        ..profiles = [
          {'id': 'default', 'name': 'Anna'},
          {'id': 'p2', 'name': 'Ben'},
        ];
      await openSettings(tester);
      await openProfiles(tester);
      // The open profile just goes back.
      await tester.tap(find.text('Anna'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), isEmpty);
      expect(find.text('Profiles'), findsNothing);

      await openProfiles(tester);
      await tester.tap(find.text('Ben'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SwitchProfile'), [
        ['p2'],
      ]);
      // A new app: on the deck list, naming the profile.
      expect(find.text('Profiles'), findsNothing);
      expect(find.byKey(const Key('settings-list')), findsNothing);
      expect(
        find.descendant(
          of: find.byType(AppBar),
          matching: find.byKey(const Key('profile-name')),
        ),
        findsOneWidget,
      );
      expect(find.text('Ben'), findsOneWidget);
      expect(go.callsTo('app.Settings').length, greaterThan(1));
    });

    testWidgets('the startup setting is stored outside the collection', (
      tester,
    ) async {
      final go = FakeGo();
      await openSettings(tester);
      await scrollTo(tester, find.text('Show profile list at startup'));
      await tester.tap(find.text('Show profile list at startup'));
      await tester.pumpAndSettle();
      expect(go.callsTo('app.SetShowProfilesAtStartup'), [
        [true],
      ]);
      expect(go.showProfilesAtStartup, isTrue);
      final toggle = tester.widget<SwitchListTile>(
        find.byKey(const Key('profiles-at-startup')),
      );
      expect(toggle.value, isTrue);
    });
  });
  group('agents (MCP)', () {
    /// Opens Settings on Linux and scrolls to the switch.
    Future<void> openOnLinux(WidgetTester tester) async {
      await openSettings(tester);
      await scrollTo(tester, find.text('Allow agents (MCP)'));
    }

    /// Records what the app copies to the clipboard.
    List<String> recordClipboard(WidgetTester tester) {
      final copied = <String>[];
      tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        (call) async {
          if (call.method == 'Clipboard.setData') {
            copied.add((call.arguments as Map)['text'] as String);
          }
          return null;
        },
      );
      addTearDown(
        () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
          SystemChannels.platform,
          null,
        ),
      );
      return copied;
    }

    testWidgets('not offered on Android', (tester) async {
      FakeGo();
      await openSettings(tester);
      await scrollTo(tester, find.text('About'));
      expect(find.text('Agents (MCP)'), findsNothing);
    });

    testWidgets('not offered on the web', (tester) async {
      FakeGo();
      await openSettings(tester, isWeb: true);
      await scrollTo(tester, find.text('About'));
      expect(find.text('Agents (MCP)'), findsNothing);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('not offered when Go has no server', (tester) async {
      FakeGo().mcpSupported = false;
      await openSettings(tester);
      await scrollTo(tester, find.text('About'));
      expect(find.text('Agents (MCP)'), findsNothing);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('turning it on explains it first and shows the URL', (
      tester,
    ) async {
      final go = FakeGo();
      await openOnLinux(tester);
      expect(find.text('Agents (MCP)'), findsOneWidget);
      expect(find.text('Server running'), findsNothing);

      await tester.tap(find.byKey(const Key('mcp-switch')));
      await tester.pumpAndSettle();
      expect(find.text('Allow agents?'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.callsTo('mcp.SetEnabled'), isEmpty);

      await tester.tap(find.byKey(const Key('mcp-switch')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Allow'));
      await tester.pumpAndSettle();
      expect(go.callsTo('mcp.SetEnabled'), [
        [true],
      ]);
      expect(find.text('Server running'), findsOneWidget);
      expect(find.text('http://127.0.0.1:47811/mcp'), findsOneWidget);

      // Off and on again: the token exists, no second explanation.
      await tester.tap(find.byKey(const Key('mcp-switch')));
      await tester.pumpAndSettle();
      expect(find.text('Server running'), findsNothing);
      await tester.tap(find.byKey(const Key('mcp-switch')));
      await tester.pumpAndSettle();
      expect(find.text('Allow agents?'), findsNothing);
      expect(find.text('Server running'), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('a port in use is shown and another port can be set', (
      tester,
    ) async {
      final go = FakeGo()
        ..mcpPortError = ((port) =>
            port == 47811 ? 'Port 47811 is already in use.' : null)
        ..mcpStatus = mcpStatusJson(token: 'secret');
      await openOnLinux(tester);
      await tester.tap(find.byKey(const Key('mcp-switch')));
      await tester.pumpAndSettle();
      expect(find.text('Server not running'), findsOneWidget);
      expect(find.text('Port 47811 is already in use.'), findsOneWidget);

      await scrollTo(tester, find.byKey(const Key('mcp-port')));
      await tester.tap(find.byKey(const Key('mcp-port')));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, '80');
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(
        find.text('The port must be between 1024 and 65535.'),
        findsOneWidget,
      );

      await tester.tap(find.byKey(const Key('mcp-port')));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, '50000');
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(go.callsTo('mcp.SetPort').last, [50000]);
      expect(find.text('Server running'), findsOneWidget);
      expect(find.text('http://127.0.0.1:50000/mcp'), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets('the client configuration and token can be copied', (
      tester,
    ) async {
      final go = FakeGo()
        ..mcpStatus = mcpStatusJson(
          enabled: true,
          token: 'secret',
          running: true,
        );
      final copied = recordClipboard(tester);
      await openOnLinux(tester);
      expect(find.text('secret'), findsNothing);

      await scrollTo(tester, find.text('Copy client configuration'));
      await tester.tap(find.text('Copy client configuration'));
      await tester.pumpAndSettle();
      expect(find.text('Client configuration copied'), findsOneWidget);
      final config = jsonDecode(copied.single) as Map<String, dynamic>;
      expect(config['mcpServers']['cardflip'], {
        'type': 'http',
        'url': 'http://127.0.0.1:47811/mcp',
        'headers': {'Authorization': 'Bearer secret'},
      });

      await tester.tap(find.byKey(const Key('mcp-copy-token')));
      await tester.pumpAndSettle();
      expect(copied.last, 'secret');

      // Regenerating asks first, then the copies use the new token.
      await tester.tap(find.byKey(const Key('mcp-regenerate-token')));
      await tester.pumpAndSettle();
      expect(find.text('Regenerate the token?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Regenerate'));
      await tester.pumpAndSettle();
      expect(go.callsTo('mcp.RegenerateToken'), hasLength(1));
      await tester.tap(find.byKey(const Key('mcp-copy-token')));
      await tester.pumpAndSettle();
      expect(copied.last, isNot('secret'));
      expect(copied.last, go.mcpStatus['token']);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));

    testWidgets("the agents' recent changes are listed newest first", (
      tester,
    ) async {
      final go = FakeGo()
        ..mcpStatus = mcpStatusJson(
          enabled: true,
          token: 'secret',
          running: true,
        );
      await openOnLinux(tester);
      await scrollTo(tester, find.text('Recent agent changes'));
      expect(find.text('No changes in this session'), findsOneWidget);

      final added = DateTime.utc(2026, 10, 20, 9, 30);
      final deleted = DateTime.utc(2026, 10, 20, 9, 45);
      go.addMcpActivity(
        '\x02{"code":"stepAddNotes","args":["3"]}\x03Add 3 notes\x04',
        added,
      );
      go.addMcpActivity(
        '\x02{"code":"stepDeleteNotes","args":["1"]}\x03Delete 1 note\x04',
        deleted,
      );
      await tester.pumpAndSettle();
      await scrollTo(tester, find.text('Add 3 notes'));
      expect(find.text('No changes in this session'), findsNothing);
      final first = tester.getTopLeft(find.text('Delete 1 note')).dy;
      final second = tester.getTopLeft(find.text('Add 3 notes')).dy;
      expect(first, lessThan(second));
      expect(find.text(formatDateTime(deleted.toLocal())), findsOneWidget);
      expect(find.text(formatDateTime(added.toLocal())), findsOneWidget);
    }, variant: TargetPlatformVariant.only(TargetPlatform.linux));
  });

  group('sync', () {
    const home = {'id': 'wifi:Home', 'wifi': 'Home'};

    /// Opens Settings and scrolls to [finder] in the Sync section.
    Future<void> openSync(
      WidgetTester tester, {
      Finder? finder,
      FakeSyncPlatform? platform,
      Size size = desktop,
    }) async {
      await pumpApp(tester, size: size, syncPlatform: platform);
      await tester.tap(find.text('Settings').last);
      await tester.pumpAndSettle();
      await scrollTo(tester, finder ?? find.byKey(const Key('link-profile')));
    }

    /// Scrolls [finder] into view and taps it.
    Future<void> tapShown(WidgetTester tester, Finder finder) async {
      await tester.ensureVisible(finder);
      await tester.pumpAndSettle();
      await tester.tap(finder);
      await tester.pumpAndSettle();
    }

    testWidgets('hidden on the web', (tester) async {
      final go = FakeGo();
      await openSettings(tester, isWeb: true);
      await scrollTo(tester, find.text('About'));
      expect(find.text('Sync'), findsNothing);
      expect(find.text('Pair a device'), findsNothing);
      expect(go.syncCalls, isEmpty);
    });

    testWidgets('shown on Android with the device name and no paired devices', (
      tester,
    ) async {
      final go = FakeGo();
      await openSync(tester);
      expect(find.text('Sync'), findsOneWidget);
      // Opening the section gives the device its identity, named after
      // the phone.
      expect(go.syncCalls, contains('Init Pixel 9'));
      expect(find.text('Pixel 9'), findsOneWidget);
      expect(find.text('No paired devices yet'), findsOneWidget);
      expect(find.text('No trusted networks yet'), findsOneWidget);
      // Pairing needs a trusted network.
      final pair = tester.widget<ListTile>(
        find.byKey(const Key('pair-device')),
      );
      expect(pair.enabled, isFalse);
    });

    testWidgets('the device can be renamed', (tester) async {
      final go = FakeGo();
      await openSync(tester);
      await tapShown(tester, find.byKey(const Key('sync-device-name')));
      await tester.enterText(find.byKey(const Key('device-name-field')), '');
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(
        find.text('The device name must be 1 to 40 characters long.'),
        findsOneWidget,
      );
      await tester.enterText(
        find.byKey(const Key('device-name-field')),
        'Kitchen tablet',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('SetDeviceName Kitchen tablet'));
      expect(find.text('Kitchen tablet'), findsOneWidget);
    });

    testWidgets('trusting the network explains it first', (tester) async {
      final go = FakeGo();
      await openSync(tester);
      expect(find.textContaining('Not trusted'), findsOneWidget);

      await tapShown(tester, find.byKey(const Key('trust-network')));
      expect(find.text('Trust this network?'), findsOneWidget);
      expect(find.textContaining('location permission'), findsNothing);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, isNot(contains('TrustNetwork')));

      await tapShown(tester, find.byKey(const Key('trust-network')));
      await tester.tap(find.widgetWithText(FilledButton, 'Trust'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('TrustNetwork'));
      expect(find.byKey(const ValueKey('trusted-wifi:Home')), findsOneWidget);
      expect(find.textContaining('Trusted:'), findsOneWidget);
      final pair = tester.widget<ListTile>(
        find.byKey(const Key('pair-device')),
      );
      expect(pair.enabled, isTrue);

      // Stop trusting it from the list.
      await tapShown(tester, find.byTooltip('Stop trusting'));
      expect(go.syncCalls, contains('UntrustNetwork wifi:Home'));
      expect(find.text('No trusted networks yet'), findsOneWidget);
    });

    testWidgets('on a small phone the network text keeps its width', (
      tester,
    ) async {
      FakeGo();
      await openSync(
        tester,
        size: const Size(360, 640),
        finder: find.byKey(const Key('trust-network')),
      );
      // The button is below the text, not beside it.
      final text = tester.getRect(find.textContaining('Not trusted'));
      final button = tester.getRect(find.byKey(const Key('trust-network')));
      expect(text.width, greaterThan(220));
      expect(button.top, greaterThanOrEqualTo(text.bottom));
    });

    testWidgets('the trust dialog fits a small phone with large text', (
      tester,
    ) async {
      FakeGo();
      tester.platformDispatcher.textScaleFactorTestValue = 1.3;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
      await openSync(
        tester,
        size: const Size(360, 640),
        finder: find.byKey(const Key('trust-network')),
        platform: FakeSyncPlatform(needsLocation: true),
      );
      await tapShown(tester, find.byKey(const Key('trust-network')));
      expect(tester.takeException(), isNull);
      final trust = find.widgetWithText(FilledButton, 'Trust');
      expect(tester.getRect(trust).bottom, lessThanOrEqualTo(640));
      await tester.tap(trust);
      await tester.pumpAndSettle();
      expect(find.text('Trust this network?'), findsNothing);
    });

    testWidgets('a wired network is trusted by its subnet', (tester) async {
      FakeGo();
      await openSync(
        tester,
        platform: FakeSyncPlatform(
          current: const NetworkState('ethernet', subnet: '192.168.178.0/24'),
        ),
      );
      await tapShown(tester, find.byKey(const Key('trust-network')));
      await tester.tap(find.widgetWithText(FilledButton, 'Trust'));
      await tester.pumpAndSettle();
      final trusted = find.byKey(
        const ValueKey('trusted-subnet:192.168.178.0/24'),
      );
      expect(
        find.descendant(
          of: trusted,
          matching: find.text('Wired network 192.168.178.0/24'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('refusing the location permission trusts the subnet', (
      tester,
    ) async {
      FakeGo();
      await openSync(
        tester,
        platform: FakeSyncPlatform(needsLocation: true, grantLocation: false),
      );
      await tapShown(tester, find.byKey(const Key('trust-network')));
      expect(find.textContaining('location permission'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Trust'));
      await tester.pumpAndSettle();
      expect(
        find.descendant(
          of: find.byKey(const ValueKey('trusted-subnet:192.168.1.0/24')),
          matching: find.text('Network 192.168.1.0/24'),
        ),
        findsOneWidget,
      );
    });

    testWidgets('pairing with a device found nearby and a matching code', (
      tester,
    ) async {
      final go = FakeGo()
        ..syncStatus = syncStatusJson(deviceName: 'Laptop', trusted: [home]);
      await openSync(tester, finder: find.byKey(const Key('pair-device')));
      await tapShown(tester, find.byKey(const Key('pair-device')));
      expect(go.syncCalls, contains('StartPairing'));
      expect(find.textContaining('Visible for pairing for 4:'), findsOneWidget);
      expect(find.text('192.168.1.20:47812'), findsOneWidget);
      expect(find.text('Looking for devices…'), findsOneWidget);

      go.setSyncStatus({
        ...go.syncStatus,
        'discovered': [
          {
            'id': 'PHONEDEVICEID001',
            'name': 'Phone',
            'host': '192.168.1.30',
            'port': 47812,
          },
        ],
      });
      await tester.pumpAndSettle();
      await tester.tap(find.text('Phone'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('PairWith 192.168.1.30 47812'));

      // Go shows the code, both users confirm.
      final answer = go.confirmPairing('Phone', '482913');
      await tester.pumpAndSettle();
      expect(find.text('Compare the codes'), findsOneWidget);
      expect(find.text('482 913'), findsOneWidget);
      await tester.tap(find.text('Codes match'));
      await tester.pumpAndSettle();
      expect(await answer, isTrue);
      expect(find.text('Waiting for Phone to confirm…'), findsOneWidget);
      go.pairingFinished('Phone', true);
      await tester.pumpAndSettle();
      expect(find.text('Paired with Phone'), findsOneWidget);

      // Closing the screen ends the visibility; the phone is listed.
      await tester.pageBack();
      await tester.pumpAndSettle();
      expect(go.syncCalls.last, 'StopPairing');
      expect(find.byKey(const ValueKey('paired-PHONEDEVICEID001')), findsOne);
      expect(find.text('Reachable · Not synced yet'), findsOneWidget);
    });

    testWidgets('pairing by address', (tester) async {
      final go = FakeGo()..syncStatus = syncStatusJson(trusted: [home]);
      await openSync(tester, finder: find.byKey(const Key('pair-device')));
      await tapShown(tester, find.byKey(const Key('pair-device')));
      await tester.enterText(
        find.byKey(const Key('pair-address')),
        '192.168.1.30:47900',
      );
      await tester.tap(find.byKey(const Key('pair-by-address')));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('PairWith 192.168.1.30 47900'));
      await tester.enterText(find.byKey(const Key('pair-address')), 'tablet');
      await tester.tap(find.byKey(const Key('pair-by-address')));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('PairWith tablet 47812'));
    });

    testWidgets('a pairing request is accepted, and a failed code closes', (
      tester,
    ) async {
      final go = FakeGo()..syncStatus = syncStatusJson(trusted: [home]);
      await openSync(tester, finder: find.byKey(const Key('pair-device')));
      await tapShown(tester, find.byKey(const Key('pair-device')));

      final accepted = go.requestPairing('Tablet');
      await tester.pumpAndSettle();
      expect(find.text('Pair with Tablet?'), findsOneWidget);
      await tester.tap(find.text('Accept'));
      await tester.pumpAndSettle();
      expect(await accepted, isTrue);

      // The tablet's user says the codes differ: the dialog closes.
      final answer = go.confirmPairing('Tablet', '111222');
      await tester.pumpAndSettle();
      expect(find.text('111 222'), findsOneWidget);
      go.pairingFinished('Tablet', false);
      await tester.pumpAndSettle();
      expect(await answer, isFalse);
      expect(find.text('111 222'), findsNothing);
      expect(find.text('Pairing with Tablet failed'), findsOneWidget);

      final declined = go.requestPairing('Stranger');
      await tester.pumpAndSettle();
      await tester.tap(find.text('Decline'));
      await tester.pumpAndSettle();
      expect(await declined, isFalse);
    });

    testWidgets('unpairing asks first', (tester) async {
      final synced = DateTime.utc(2026, 10, 20, 9, 30);
      final go = FakeGo()
        ..syncStatus = syncStatusJson(
          trusted: [home],
          paired: [
            pairedDeviceJson(
              'PHONEDEVICEID001',
              'Phone',
              reachable: false,
              lastSynced: synced,
            ),
          ],
        );
      await openSync(tester);
      expect(
        find.text(
          'Not reachable · Last synced ${formatDateTime(synced.toLocal())}',
        ),
        findsOneWidget,
      );

      await tapShown(tester, find.widgetWithText(TextButton, 'Unpair'));
      expect(find.text('Unpair Phone?'), findsOneWidget);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(go.syncCalls.where((c) => c.startsWith('Unpair')), isEmpty);

      await tapShown(tester, find.widgetWithText(TextButton, 'Unpair'));
      await tester.tap(find.widgetWithText(FilledButton, 'Unpair'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('Unpair PHONEDEVICEID001'));
      expect(find.text('No paired devices yet'), findsOneWidget);
    });

    testWidgets('the profile is linked to a new profile and unlinked', (
      tester,
    ) async {
      final go = FakeGo()
        ..syncStatus = syncStatusJson(
          trusted: [home],
          paired: [pairedDeviceJson('PHONEDEVICEID001', 'Phone')],
        );
      await openSync(tester);
      expect(
        find.text('This profile does not sync with another device yet.'),
        findsOneWidget,
      );
      await tapShown(tester, find.byKey(const Key('link-profile')));
      expect(find.text('Profile on Phone'), findsOneWidget);
      expect(find.byKey(const ValueKey('remote-default')), findsOneWidget);
      await tester.tap(find.byKey(const Key('remote-new')));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('Link PHONEDEVICEID001 '));
      expect(find.text('Default on Phone'), findsOneWidget);
      expect(find.text('Needs your choice: merge or replace'), findsOne);
      // One link per device.
      final link = tester.widget<ListTile>(
        find.byKey(const Key('link-profile')),
      );
      expect(link.enabled, isFalse);

      await tapShown(tester, find.widgetWithText(TextButton, 'Unlink'));
      expect(find.text('Unlink Phone?'), findsOneWidget);
      await tester.tap(find.widgetWithText(FilledButton, 'Unlink'));
      await tester.pumpAndSettle();
      expect(go.syncCalls, contains('Unlink link-1'));
      expect(find.text('Default on Phone'), findsNothing);
    });
  });
}
