import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:cardflip/app/bootstrap.dart';
import 'package:cardflip/data/sync_model.dart';
import 'package:cardflip/features/import/pick.dart';
import 'package:cardflip/golang/golang.dart' as golang;

import 'fake_sync_platform.dart';

/// JSON for the default deck options.
Map<String, Object?> deckOptionsJson() => {
  'newPerDay': 20,
  'reviewsPerDay': 200,
  'desiredRetention': 0.9,
  'disableAutoplay': false,
  'skipQuestionOnReplay': false,
  'learningSteps': [60, 600],
  'relearningSteps': [600],
  'interdayLearningOrder': 'mix',
  'buryNew': false,
  'buryReview': false,
  'buryInterdayLearning': false,
  'leechThreshold': 8,
  'leechAction': 'tag',
  'newIgnoresReviewLimit': true,
  'limitsStartFromTop': false,
  'maxInterval': 36500,
  'minRelearnInterval': 1,
  'newGatherOrder': 'ascending',
  'newSortOrder': 'gathered',
  'newReviewOrder': 'mix',
  'reviewSortOrder': 'dueRandom',
  'insertionOrder': 'sequential',
  'maxAnswerSeconds': 60,
  'showTimer': false,
  'stopTimerOnAnswer': true,
  'secondsToShowQuestion': 0,
  'secondsToShowAnswer': 0,
  'questionAction': 'showAnswer',
  'answerAction': 'buryCard',
  'weights': <double>[],
  'optimizeSearch': '',
  'ignoreBefore': '',
  'historicalRetention': 0.9,
  'healthCheck': false,
  'rescheduleOnChange': false,
  'easyDays': [0, 0, 0, 0, 0, 0, 0],
};

/// Go's default FSRS parameters.
const defaultFsrsWeights = [
  0.40255, 1.18385, 3.173, 15.69105, 7.1949, 0.5345, 1.4604, 0.0046, //
  1.54575, 0.1192, 1.01925, 1.9395, 0.11, 0.29605, 2.2698, 0.2315, //
  2.9898, 0.51655, 0.6621,
];

/// JSON for the result of optimizing: fitted [weights] learned from
/// [reviews] reviews, to be used when [use]; with [notEnough] too few
/// reviews.
Map<String, Object?> optimizeResultJson({
  List<double> weights = defaultFsrsWeights,
  int reviews = 5000,
  bool use = true,
  bool notEnough = false,
  bool? healthPassed,
}) => {
  'weights': notEnough ? <double>[] : weights,
  'use': use && !notEnough && healthPassed != false,
  'better': use && !notEnough,
  'reviews': reviews,
  'logLoss': 0.3,
  'rmse': 0.03,
  'currentLogLoss': 0.32,
  'currentRmse': 0.04,
  'healthChecked': healthPassed != null,
  'healthPassed': healthPassed ?? false,
  'notEnoughReviews': notEnough,
  'minReviews': 400,
};

/// A progress event of `decks.Optimize` or `decks.OptimizeAll`.
Map<String, Object?> optimizeProgress(String stage, int done, int total) => {
  'stage': stage,
  'done': done,
  'total': total,
  'result': null,
  'presets': <Object?>[],
  'error': '',
};

/// The last event of `decks.Optimize`, with [result] (see
/// [optimizeResultJson]).
Map<String, Object?> optimizeDone(Map<String, Object?> result) => {
  ...optimizeProgress('done', 1, 1),
  'result': result,
};

/// A progress event of `decks.Simulate` or `decks.SuggestRetention`.
Map<String, Object?> simulationProgress(String stage, int done, int total) => {
  'stage': stage,
  'done': done,
  'total': total,
  'series': <Object?>[],
  'suggestion': null,
  'error': '',
};

/// JSON for one simulated retention over [days]: reviews rise with the
/// retention, memorised cards grow by one a day.
Map<String, Object?> simulationSeriesJson(double retention, int days) {
  final perDay = (retention * 100 - 60).round();
  return {
    'retention': retention,
    'reviews': List.filled(days, perDay),
    'new': List.filled(days, 1),
    'seconds': List.filled(days, perDay * 10.0),
    'memorized': [for (var d = 0; d < days; d++) d + 1.0],
    'totalReviews': perDay * days,
    'totalNew': days,
    'totalSeconds': perDay * days * 10.0,
  };
}

/// The last event of `decks.SuggestRetention`, suggesting [retention].
Map<String, Object?> suggestionDone(double retention) => {
  ...simulationProgress('done', 1, 1),
  'suggestion': {
    'retention': retention,
    'points': [
      for (var i = 0; i < 30; i++)
        {
          'retention': (70 + i) / 100,
          'secondsPerCard': 50 + ((70 + i) / 100 - retention).abs() * 400,
        },
    ],
  },
};

/// JSON for a deck tree row.
Map<String, Object?> deckNode(
  int id,
  String fullName, {
  int newCount = 0,
  int learning = 0,
  int review = 0,
  int parentId = 0,
  bool hasChildren = false,
  bool collapsed = false,
  int cards = 0,
  bool filtered = false,
  Map<String, int> progress = const {},
}) {
  final parts = fullName.split('::');
  return {
    'id': id,
    'name': parts.last,
    'fullName': fullName,
    'depth': parts.length - 1,
    'parentId': parentId,
    'hasChildren': hasChildren,
    'collapsed': collapsed,
    'counts': {'new': newCount, 'learning': learning, 'review': review},
    'cards': cards,
    'progress': {
      'new': progress['new'] ?? 0,
      'learning': progress['learning'] ?? 0,
      'young': progress['young'] ?? 0,
      'mature': progress['mature'] ?? 0,
    },
    'options': deckOptionsJson(),
    'filtered': filtered,
  };
}

/// JSON for a weekly recap starting on study day [weekStart]; [last] and
/// [before] are (study days, answers, minutes, retention).
Map<String, Object?> recapJson({
  int weekStart = 20750,
  (int, int, int, double?) last = (6, 420, 95, 0.91),
  (int, int, int, double?) before = (4, 300, 70, 0.88),
}) {
  Map<String, Object?> week((int, int, int, double?) w) => {
    'studyDays': w.$1,
    'answers': w.$2,
    'minutes': w.$3,
    'retention': w.$4,
  };
  return {'weekStart': weekStart, 'last': week(last), 'before': week(before)};
}

/// JSON for what deleting a deck removes.
Map<String, Object?> deletionInfoJson({
  int decks = 1,
  int cards = 0,
  int inFiltered = 0,
  int returned = 0,
}) => {
  'decks': decks,
  'cards': cards,
  'inFiltered': inFiltered,
  'returned': returned,
};

/// JSON for a filtered deck's options, as `decks.NewFilteredOptions` makes
/// them.
Map<String, Object?> filteredOptionsJson(
  String name, {
  String search = '',
  bool reschedule = true,
}) => {
  'name': name,
  'terms': [
    {'search': search, 'limit': 100, 'order': 'random'},
  ],
  'reschedule': reschedule,
  'previewDelays': [60, 600, 0],
  'secondLimit': 20,
  'secondOrder': 'due',
};

/// Anki's Image Occlusion note type.
Map<String, Object?> occlusionTypeJson({int id = 13}) => noteTypeJson(
  id,
  'Image Occlusion',
  ['Occlusion', 'Image', 'Header', 'Back Extra', 'Comments'],
  cloze: true,
  imageOcclusion: true,
  templates: const ['Image Occlusion'],
);

/// An `McpStatus` as Go sends it; the URL follows the port.
Map<String, Object?> mcpStatusJson({
  bool enabled = false,
  int port = 47811,
  String token = '',
  bool running = false,
  String? error,
}) => {
  'enabled': enabled,
  'port': port,
  'token': token,
  'url': 'http://127.0.0.1:$port/mcp',
  'running': running,
  'error': ?error,
};

/// A `SyncStatus` as Go sends it.
Map<String, Object?> syncStatusJson({
  bool supported = true,
  bool ready = true,
  String deviceId = 'LAPTOPDEVICEID01',
  String deviceName = 'Laptop',
  String networkKind = 'wifi',
  String wifi = 'Home',
  String subnet = '192.168.1.0/24',
  List<Map<String, Object?>> trusted = const [],
  bool pairing = false,
  List<Map<String, Object?>> paired = const [],
  List<Map<String, Object?>> discovered = const [],
  List<Map<String, Object?>> links = const [],
  bool syncing = false,
  Map<String, Object?>? progress,
  DateTime? lastSynced,
  String? error,
}) {
  final id = networkKind == 'cellular' || networkKind == 'none'
      ? ''
      : networkKind == 'wifi' && wifi.isNotEmpty
      ? 'wifi:$wifi'
      : subnet.isEmpty
      ? ''
      : 'subnet:$subnet';
  final isTrusted = id.isNotEmpty && trusted.any((n) => n['id'] == id);
  return {
    'supported': supported,
    'ready': ready,
    if (ready) 'deviceId': deviceId,
    if (ready) 'deviceName': deviceName,
    'network': {
      'kind': networkKind,
      if (wifi.isNotEmpty) 'wifi': wifi,
      if (subnet.isNotEmpty) 'subnet': subnet,
      if (id.isNotEmpty) 'id': id,
      'trusted': isTrusted,
    },
    'trusted': trusted,
    'listening': ready && isTrusted,
    if (ready && isTrusted) 'port': 47812,
    if (ready && isTrusted) 'address': '192.168.1.20',
    'pairing': ready && isTrusted && pairing,
    'paired': paired,
    'discovered': discovered,
    'links': links,
    'syncing': syncing,
    'progress': ?progress,
    if (lastSynced != null) 'lastSynced': lastSynced.toUtc().toIso8601String(),
    'error': ?error,
  };
}

/// A paired device in a `SyncStatus`.
Map<String, Object?> pairedDeviceJson(
  String id,
  String name, {
  bool reachable = true,
  DateTime? lastSynced,
}) => {
  'id': id,
  'name': name,
  'reachable': reachable,
  if (lastSynced != null) 'lastSynced': lastSynced.toUtc().toIso8601String(),
};

/// A link in a `SyncStatus`.
Map<String, Object?> linkInfoJson(
  String id,
  String device,
  String deviceName, {
  String remoteProfile = 'default',
  String remoteName = 'Default',
  bool needsChoice = false,
  DateTime? lastSynced,
  String? error,
}) => {
  'id': id,
  'device': device,
  'deviceName': deviceName,
  'remoteProfile': remoteProfile,
  'remoteName': remoteName,
  if (lastSynced != null) 'lastSynced': lastSynced.toUtc().toIso8601String(),
  'needsChoice': needsChoice,
  'error': ?error,
};

Map<String, Object?> noteTypeJson(
  int id,
  String name,
  List<String> fields, {
  bool cloze = false,
  bool imageOcclusion = false,
  int noteCount = 0,
  Set<int> htmlByDefault = const {},
  Set<int> sticky = const {},
  Map<int, Map<String, Object?>> options = const {},
  int sortField = 0,
  List<String> templates = const ['Card 1'],
}) => {
  'id': id,
  'name': name,
  'cloze': cloze,
  if (imageOcclusion) ...{
    'imageOcclusion': true,
    // As Go finds them in Anki's template.
    'occlusionField': fields.indexOf('Occlusion'),
    'imageField': fields.indexOf('Image'),
  },
  'fields': fields,
  'fieldOptions': [
    for (var i = 0; i < fields.length; i++)
      {
        'htmlByDefault': htmlByDefault.contains(i),
        'sticky': sticky.contains(i),
        'font': '',
        'size': 0,
        'rtl': false,
        'collapsed': false,
        'excludeFromSearch': false,
        ...?options[i],
      },
  ],
  'sortField': sortField,
  'templates': [
    for (final t in templates)
      {'name': t, 'qfmt': '{{${fields.first}}}', 'afmt': '{{FrontSide}}'},
  ],
  'css': '.card {}',
  'noteCount': noteCount,
};

/// JSON for what changing notes to another note type does.
Map<String, Object?> noteTypeChangeJson({
  int notes = 1,
  int cardsDeleted = 0,
  int cardsCreated = 0,
  List<String> lostFields = const [],
}) => {
  'notes': notes,
  'cardsDeleted': cardsDeleted,
  'cardsCreated': cardsCreated,
  'lostFields': lostFields,
};

/// Plays the Go side in widget tests: in-memory answers for every channel
/// the app uses, and a record of the calls it made.
class FakeGo {
  FakeGo() : bridge = golang.Bridge.fake() {
    _defaults();
  }

  final golang.Bridge bridge;

  /// Calls made per channel, with their arguments.
  final calls = <String, List<List<dynamic>>>{};

  Map<String, Object?> status = {
    'state': 'ok',
    'persistent': true,
    'showPersistenceNotice': false,
    'version': '0.1.0',
  };
  Map<String, Object?> settings = {
    'theme': 'system',
    'language': '',
    'rolloverHour': 4,
    'autoContinueLearning': false,
    'learnAheadMinutes': 20,
    'showPlayButtons': true,
    'interruptAudio': true,
    'flagNames': List.filled(7, ''),
    'stripPasteFormatting': false,
    'pasteImagesAsPng': false,
    'backup': {'intervalMinutes': 30, 'daily': 8, 'weekly': 8, 'monthly': 8},
    'deckByNoteType': false,
    'ignoreAccents': false,
    'browserColumns': ['question', 'deck', 'due', 'noteType'],
    'browserMode': 'cards',
    'browserSplit': 0.6,
    'defaultSearch': '',
    'lastDueDate': '',
    'timeboxMinutes': 0,
    'showRemainingCounts': true,
    'showIntervals': true,
    'spaceAnswersGood': true,
    'hideBarsDuringReview': false,
    'questionFadeMs': 100,
    'recentDecks': <int>[],
    'uiScale': 100,
    'minimalist': false,
    'reduceMotion': 'system',
    'showProgress': true,
    'showStudyRewards': true,
    'dailyGoalKind': '',
    'dailyGoal': 0,
    'restDaysPerWeek': 0,
    'goalSuggestionShown': false,
    'deckProgressInList': false,
    'recapDismissedWeek': 0,
  };

  /// The profiles `app.Profiles` lists, as `{'id', 'name'}` maps. Like Go,
  /// `app.AddProfile` and `app.RenameProfile` keep names unique ignoring
  /// case, the list is sorted by name, and `app.DeleteProfile` refuses the
  /// open profile and the only one. [currentProfile] is the open profile,
  /// null until `app.Open` or `app.SwitchProfile`; `app.Open` opens
  /// [lastProfile]. `app.SwitchProfile` answers [profileStatus] for the
  /// profile, or [status], and sends a change of every kind.
  List<Map<String, Object?>> profiles = [
    {'id': 'default', 'name': 'Default'},
  ];
  String? currentProfile;
  String lastProfile = 'default';
  bool showProfilesAtStartup = false;
  Map<String, int> profileCards = {};
  Map<String, Map<String, Object?>> profileStatus = {};
  int _nextProfile = 2;

  /// The decks `notes.Defaults` answers: [lastDeck], or the one per note
  /// type ID when the setting "Change deck depending on note type" is on
  /// and the note type has one. The default note type is 10.
  int lastDeck = 1;
  Map<int, int> deckByType = {};

  /// The media files `media.Add` stored, by name. Like Go it reuses a file
  /// with the same name and content, numbers a taken name (cat-1.jpg) and
  /// names a file without a stem (".m4a") media-N. `media.AddPastedImage`
  /// stores paste-N.jpg, or paste-N.png when asked for PNG.
  final media = <String, Uint8List>{};

  /// The notes `notes.Duplicates` compares with: like Go it matches the
  /// first field of the same note type with HTML, case and surrounding
  /// whitespace ignored, leaves out the excluded note and never matches an
  /// empty field.
  List<({int id, int typeId, String first})> existingNotes = [];

  /// The tags of each note, for the tag calls. Like Go, `notes.Tags` lists
  /// them once each (ignoring case) sorted, and `notes.RenameTag`,
  /// `notes.MoveTags` and `notes.DeleteTags` change a tag and its child tags
  /// on every note, merging duplicates, and answer how many notes changed;
  /// `notes.TagUsage` counts the notes with any of the tags or their
  /// children. `notes.ClearUnusedTags` answers [unusedTags].
  List<List<String>> noteTags = [];
  int unusedTags = 0;

  /// The saved searches `browse.SavedSearches` lists, as `{'name', 'query'}`
  /// maps. Like Go, `browse.SaveSearch` and `browse.RenameSavedSearch` keep
  /// the names unique ignoring case and the list sorted by name, and saving
  /// refuses a query [searchError] rejects, as `app.SetDefaultSearch` does.
  List<Map<String, Object?>> savedSearches = [];

  /// The sets `browse.FindDuplicates` answers, as `{'content', 'noteIds'}`
  /// maps, whatever the field; like Go it refuses a query [searchError]
  /// rejects.
  List<Map<String, Object?>> duplicateSets = [];

  /// The note type of each card ID for `browse.NoteTypeOfCards`; cards not
  /// listed are Basic (10). Like Go it refuses cards of several note types.
  Map<int, int> cardNoteTypes = {};

  /// What `browse.ChangeNoteTypePreview` and `browse.ChangeNoteType` answer
  /// for their arguments (card IDs, target, field map, card type map), as a
  /// NoteTypeChange map; by default every card's note changes and no card
  /// is deleted or created. `browse.ChangeNoteType` also sends a change.
  Map<String, Object?> Function(List<dynamic> args) noteTypeChange = (args) =>
      noteTypeChangeJson(notes: (args[0] as List).length);

  /// The error Go gives for an invalid search, or null when it is valid.
  String? Function(String query) searchError = (query) =>
      query.contains('is:bogus') ? bogusSearchError : null;

  /// What Go says about `is:bogus`.
  static const bogusSearchError =
      'is:bogus: unknown state; use is:new, is:due, is:review, is:learn or is:suspended';

  /// The automatic backups `app.Backups` lists, newest first.
  List<Map<String, Object?>> backups = [];

  /// The notes with empty cards `app.EmptyCards` lists, as [emptyNote]
  /// builds them. `app.DeleteEmptyCards` removes the deleted cards.
  List<Map<String, Object?>> emptyNotes = [];

  /// The number of notes in the deletion log `app.ExportDeletionLog`
  /// exports; `app.ClearDeletionLog` sets it to 0.
  int deletedNotes = 0;

  /// The media report `app.CheckMedia` returns: unused files and the trash
  /// as `{'name', 'size'}` maps, the missing names and the number of notes
  /// that refer to them. Like Go, `app.DeleteUnusedMedia` moves the unused
  /// files to the trash, `app.RestoreMediaTrash` moves them back (they are
  /// unused again), `app.EmptyMediaTrash` empties it and
  /// `app.TagMissingMedia` reports [missingMediaNotes] tagged notes.
  List<Map<String, Object?>> unusedMedia = [];
  List<String> missingMedia = [];
  int missingMediaNotes = 0;
  List<Map<String, Object?>> mediaTrash = [];

  /// What the `app.CheckDatabase` stream sends; by default a progress event
  /// and then a report without problems. Tests can return their own
  /// controller's stream to drive progress and errors.
  late Stream<Object?> Function() checkDatabaseEvents = () =>
      Stream.fromIterable([
        checkProgress('notes', 0, 1),
        checkDone(checkReport()),
      ]);

  /// When `app.CreateBackup` says it made its backup; it advances by a
  /// minute per backup.
  DateTime backupClock = DateTime.utc(2026, 10, 13, 12);
  List<Map<String, Object?>> tree = [deckNode(1, 'Default')];

  /// The steps `app.UndoInfo` names; `app.Undo` and `app.Redo` move the
  /// step between them and call [onUndo] or [onRedo].
  String undoStep = '';
  String redoStep = '';
  void Function()? onUndo;
  void Function()? onRedo;

  /// The cards beyond each deck's limits that `decks.CustomStudyInfo`
  /// reports, by deck ID (none when missing). Like Go, `decks.IncreaseToday`
  /// moves up to the increase of them into the deck's due counts in [tree]
  /// and sends a change.
  Map<int, ({int newCards, int reviews})> extraCards = {};

  /// How many cards a filtered deck's first search gathers, for
  /// `decks.CreateFiltered`, `decks.UpdateFiltered`, `decks.Rebuild` and
  /// `decks.CustomStudy`; like Go, they fail when it is 0 and change
  /// nothing. They keep [tree] up to date (a filtered deck's cards are its
  /// new cards) and [filteredOptions], the filtered decks' options by ID.
  int Function(String search) filterMatches = (_) => 10;
  Map<int, Map<String, Object?>> filteredOptions = {};
  int _nextDeck = 1000;

  static const noMatchError =
      'No cards matched the search. Suspended and buried cards and cards in '
      'other filtered decks are left out.';

  /// What `decks.Overview` reports besides the counts in [tree], by deck
  /// ID: buried cards, descriptions (shown as given, without scripts) and
  /// when the next card is due, which it reports only while nothing is
  /// due. Like Go, `decks.Unbury` moves the buried cards into the deck's
  /// review count and `decks.SetDescription` trims the text; both send a
  /// change.
  Map<int, int> buried = {};
  Map<int, String> descriptions = {};
  Map<int, DateTime> nextDue = {};

  /// The deck options presets by ID; 1 is Default. Decks use the preset in
  /// [deckPresets] (Default when missing) and have their own limits in
  /// [deckLimits]. Like Go, the preset calls keep names unique ignoring
  /// case, refuse to delete Default, and move the decks of a deleted preset
  /// to Default; `decks.SaveSettings` checks retention and the limits.
  Map<int, Map<String, Object?>> presets = {
    1: {'id': 1, 'name': 'Default', 'options': deckOptionsJson()},
  };
  Map<int, int> deckPresets = {};
  Map<int, Map<String, Object?>> deckLimits = {};
  int _nextPreset = 100;

  /// What the `transfer.ExportDeck`, `transfer.ExportNotes`,
  /// `transfer.ExportCollection`, `transfer.ExportNotesText` and
  /// `transfer.ExportCardsText` streams send, given the channel and its
  /// arguments; by default a finished export named as Go names it. Tests
  /// can return their own controller's stream to drive progress and errors.
  late Stream<Object?> Function(String channel, List<dynamic> args)
  exportEvents = (channel, args) =>
      Stream.value(exportDone(exportName(channel, args)));

  /// The file name Go gives an export on [channel].
  String exportName(String channel, List<dynamic> args) => switch (channel) {
    'transfer.ExportDeck' =>
      '${(tree.firstWhere((n) => n['id'] == args[0])['name'])}.apkg',
    'transfer.ExportNotes' => 'Selected notes.apkg',
    'transfer.ExportNotesText' || 'transfer.ExportCardsText' => _textName(
      channel == 'transfer.ExportNotesText' ? 'notes' : 'cards',
      args[0] as Map<String, dynamic>,
    ),
    _ => 'collection-2026-10-13.colpkg',
  };

  String _textName(String what, Map<String, dynamic> scope) {
    if (scope['all'] == true) return '$what-2026-10-13.txt';
    final deckId = scope['deckId'] as int;
    if (deckId != 0) {
      return '${tree.firstWhere((n) => n['id'] == deckId)['name']}.txt';
    }
    return 'Selected $what.txt';
  }

  /// What the `decks.Optimize` and `decks.OptimizeAll` streams send, given
  /// the channel and its arguments; by default `decks.Optimize` finishes
  /// with [optimizeResultJson] and `decks.OptimizeAll` with every preset
  /// unchanged. Tests can return their own controller's stream to drive
  /// progress and cancelling.
  late Stream<Object?> Function(String channel, List<dynamic> args)
  optimizeEvents = (channel, args) => Stream.value(
    channel == 'decks.Optimize'
        ? optimizeDone(optimizeResultJson())
        : {
            ...optimizeProgress('done', 1, 1),
            'presets': [
              for (final p in presets.values)
                {
                  'presetId': p['id'],
                  'name': p['name'],
                  'result': optimizeResultJson(notEnough: true, reviews: 0),
                  'applied': false,
                },
            ],
          },
  );

  /// What the `decks.Simulate` and `decks.SuggestRetention` streams send,
  /// given the channel and its arguments; by default `decks.Simulate`
  /// finishes with [simulationSeriesJson] for the retention and the
  /// compared one, and `decks.SuggestRetention` with [suggestionDone] at
  /// 0.85. Tests can return their own controller's stream to drive progress.
  late Stream<Object?> Function(String channel, List<dynamic> args)
  simulationEvents = (channel, args) {
    final input = args[2] as Map<String, dynamic>;
    if (channel == 'decks.SuggestRetention') {
      return Stream.fromIterable([
        simulationProgress('simulating', 15, 30),
        suggestionDone(0.85),
      ]);
    }
    final days = input['days'] as int;
    final compare = (input['compareRetention'] as num).toDouble();
    return Stream.fromIterable([
      simulationProgress('reading', 0, 1),
      {
        ...simulationProgress('done', 1, 1),
        'series': [
          simulationSeriesJson((input['retention'] as num).toDouble(), days),
          if (compare != 0) simulationSeriesJson(compare, days),
        ],
      },
    ]);
  };

  /// What `decks.RescheduleCount` and `decks.RescheduleCountAll` answer.
  int rescheduleCount = 0;
  int rescheduleCountAll = 0;

  /// What `decks.NextChange` answers; by default one hour after the call.
  DateTime Function() nextChange = () =>
      DateTime.now().add(const Duration(hours: 1));
  List<Map<String, Object?>> noteTypes = [
    noteTypeJson(10, 'Basic', ['Front', 'Back']),
    noteTypeJson(11, 'Basic (and reversed card)', ['Front', 'Back']),
    noteTypeJson(12, 'Cloze', ['Text', 'Back Extra'], cloze: true),
  ];

  /// What the `stats.*` calls answer, by section (the channel without
  /// `stats.`), whatever the scope and period; [statsJson] builds the
  /// defaults. A section with a completer in [statsGates] answers once it
  /// completes; one with an error in [statsErrors] fails with it.
  Map<String, Object?> stats = statsJson();

  /// What `progress.Today` answers, see [progressJson]; while null it
  /// fails, so the deck list shows no streak. Like Go, `progress.MarkSeen`
  /// takes the tiers it marks out of its `newAchievements`.
  Map<String, Object?>? progress;

  /// What `progress.Level` answers, see [levelJson]; while null, the XP
  /// of [progress], all from answers.
  Map<String, Object?>? level;

  /// What `progress.WeeklyRecap` answers, see [recapJson]; null for no
  /// recap.
  Map<String, Object?>? recap;

  /// What `progress.Achievements` answers, see [achievementJson]; by
  /// default every achievement is locked without progress.
  List<Map<String, Object?>> achievements = [
    for (final id in achievementTiers.keys) achievementJson(id),
  ];
  Map<String, Completer<void>> statsGates = {};
  Map<String, String> statsErrors = {};

  /// The MCP server for agents. `mcp.Supported` answers [mcpSupported] and
  /// `mcp.Status` [mcpStatus]. Like Go, `mcp.SetEnabled` creates the token
  /// the first time and keeps it, and the server runs unless
  /// [mcpPortError] gives a reason for the port, which the status then
  /// shows with the switch on; `mcp.SetPort` refuses ports outside 1024 to
  /// 65535 and moves the server; `mcp.RegenerateToken` makes a new token.
  /// `mcp.Activity` lists [mcpActivity] newest first. `mcp.StatusChanges`
  /// sends the status, and again after every change and [addMcpActivity].
  bool mcpSupported = true;
  Map<String, Object?> mcpStatus = mcpStatusJson();
  String? Function(int port) mcpPortError = (_) => null;
  final mcpActivity = <Map<String, Object?>>[];
  final _mcpChanges = StreamController<Object?>.broadcast();
  int _nextMcpToken = 1;

  /// Records an agent's change named [step] (a Go text), as Go does after
  /// a tool call that changed the collection.
  void addMcpActivity(String step, DateTime time) {
    mcpActivity.add({'time': time.toUtc().toIso8601String(), 'step': step});
    _mcpChanges.add(mcpStatus);
  }

  /// Sets the MCP status to [enabled] on [port], running unless
  /// [mcpPortError] says why not, and sends it on `mcp.StatusChanges`.
  Map<String, Object?> _setMcp({bool? enabled, int? port, String? token}) {
    final on = enabled ?? mcpStatus['enabled'] as bool;
    final p = port ?? mcpStatus['port'] as int;
    final error = on ? mcpPortError(p) : null;
    mcpStatus = mcpStatusJson(
      enabled: on,
      port: p,
      token: token ?? mcpStatus['token'] as String,
      running: on && error == null,
      error: error,
    );
    _mcpChanges.add(mcpStatus);
    return mcpStatus;
  }

  /// Device sync. `sync.Supported` answers [syncSupported], `sync.Status`
  /// and `sync.StatusChanges` [syncStatus], which [setSyncStatus] replaces
  /// and sends. Like Go, `sync.Init` makes the device ready, named after
  /// its argument unless it has a name; `sync.SetNetwork` changes the
  /// network, which is trusted when it is in the trusted list, and
  /// `sync.TrustNetwork` and `sync.UntrustNetwork` change that list;
  /// `sync.StartPairing` and `sync.StopPairing` switch pairing;
  /// `sync.SetDeviceName` refuses names that are empty or longer than 40
  /// characters. `sync.PairWith`, `sync.Unpair`, `sync.Link` and
  /// `sync.Unlink` change the paired devices and links, `sync.PairWith`
  /// failing with [syncPairError] when it is set. `sync.SyncNow` answers
  /// [syncNowResult], after [syncNowGate] if set. `sync.RemoteProfiles`
  /// answers [remoteProfiles]. Every call is recorded in [syncCalls], with
  /// its arguments.
  bool syncSupported = true;
  Map<String, Object?> syncStatus = syncStatusJson(ready: false);
  final syncCalls = <String>[];
  String? syncPairError;
  Map<String, Object?> syncNowResult = {
    'received': 0,
    'sent': 0,
    'media': 0,
    'reachable': true,
  };
  Completer<void>? syncNowGate;
  List<Map<String, Object?>> remoteProfiles = [
    {'id': 'default', 'name': 'Default'},
  ];
  final _syncChanges = StreamController<Object?>.broadcast();

  /// Replaces the sync status and sends it on `sync.StatusChanges`.
  void setSyncStatus(Map<String, Object?> status) {
    syncStatus = status;
    _syncChanges.add(status);
  }

  /// The sync status with [changes] applied.
  Map<String, Object?> _syncWith(Map<String, Object?> changes) {
    final st = {...syncStatus, ...changes};
    final net = st['network'] as Map<String, Object?>;
    final trusted = (st['trusted'] as List).cast<Map<String, Object?>>();
    final id = net['id'] as String? ?? '';
    final isTrusted = id.isNotEmpty && trusted.any((n) => n['id'] == id);
    final ready = st['ready'] as bool;
    st['network'] = {...net, 'trusted': isTrusted};
    st['listening'] = ready && isTrusted;
    if (ready && isTrusted) {
      st['port'] = 47812;
      st['address'] = '192.168.1.20';
    } else {
      st.remove('port');
      st.remove('address');
      st['pairing'] = false;
    }
    return st;
  }

  /// Go asks the Dart side about a pairing request from [name].
  Future<bool> requestPairing(String name) =>
      bridge.emitter.request<bool>('SyncUI.PairingRequest', [name]);

  /// Go asks the Dart side to confirm the pairing [code] with [name].
  Future<bool> confirmPairing(String name, String code) =>
      bridge.emitter.request<bool>('SyncUI.ConfirmPairing', [name, code]);

  /// Go tells the Dart side whether pairing with [name] worked.
  void pairingFinished(String name, bool paired) =>
      bridge.emitter.emit('SyncUI.PairingFinished', [name, paired]);

  /// Go asks the Dart side how to combine the profiles of [link].
  Future<String> chooseFirstSync(Map<String, Object?> link) =>
      bridge.emitter.request<String>('SyncUI.ChooseFirstSync', [link]);

  final _changes = StreamController<Object?>.broadcast();

  /// The render environment `app.SetRenderEnvironment` set last; setting
  /// a different one sends an "environment" change, as Go does.
  Map<String, Object?>? renderEnv;

  /// Pushes a change notification, as Go does after a mutation.
  void emitChange({
    String kind = 'test',
    bool decks = true,
    bool noteTypes = false,
    bool notes = false,
    bool all = false,
    bool settings = false,
  }) {
    _changes.add({
      'kind': kind,
      'settings': settings,
      'decks': decks,
      'noteTypes': noteTypes,
      'notes': notes,
      'cards': false,
      'reviews': false,
      'media': false,
      'all': all,
    });
  }

  /// Answers [channel] with [fn], recording the call.
  void handle(
    String channel,
    FutureOr<Object?> Function(List<dynamic> args) fn,
  ) {
    bridge.receiver.handle(channel, (data) {
      final args = (data as List<dynamic>?) ?? const [];
      calls.putIfAbsent(channel, () => []).add(args);
      return fn(args);
    });
  }

  /// The cards of an empty-card note left after deleting [ids]; a note
  /// whose cards are all empty keeps its first one.
  static List<Object?> _keptCards(Map<String, Object?> note, Set<dynamic> ids) {
    final cards = note['cards'] as List<dynamic>;
    final kept = [
      for (final c in cards)
        if (!ids.contains((c as Map)['cardId'])) c,
    ];
    return note['allEmpty'] == true && kept.isEmpty ? [cards.first] : kept;
  }

  /// Bytes from Go call arguments, however the bridge encoded them.
  static Uint8List _bytes(Object? arg) => switch (arg) {
    final Uint8List bytes => bytes,
    final List<dynamic> bytes => Uint8List.fromList(bytes.cast<int>()),
    final String b64 => base64Decode(b64),
    final other => throw ArgumentError('not bytes: $other'),
  };

  /// Sets option [key] of a field, from the arguments typeId, index, on.
  void _setFieldOption(List<dynamic> args, String key) {
    final [typeId as int, index as int, on as bool] = args;
    noteTypes = [
      for (final t in noteTypes)
        if (t['id'] == typeId)
          {
            ...t,
            'fieldOptions': [
              for (final (i, o) in (t['fieldOptions'] as List).indexed)
                i == index ? {...o as Map, key: on} : o,
            ],
          }
        else
          t,
    ];
  }

  /// Replaces note type [typeId] with [edit]'s result.
  void _editType(
    int typeId,
    Map<String, Object?> Function(Map<String, Object?>) edit,
  ) {
    noteTypes = [
      for (final t in noteTypes)
        if (t['id'] == typeId) edit(t) else t,
    ];
  }

  /// Replaces the card types of note type [typeId] with [edit]'s result.
  void _editTemplates(int typeId, List<Object?> Function(List<Object?>) edit) {
    noteTypes = [
      for (final t in noteTypes)
        if (t['id'] == typeId)
          {
            ...t,
            'templates': edit([...t['templates'] as List]),
          }
        else
          t,
    ];
  }

  /// Whether [tag] is [root] or one of its child tags, ignoring case.
  static bool _inTree(String tag, String root) {
    final t = tag.toLowerCase(), r = root.toLowerCase();
    return t == r || t.startsWith('$r::');
  }

  /// Maps every note's tags with [edit] (null drops a tag), merging
  /// duplicates ignoring case, and sends a change; returns how many notes
  /// changed.
  int _editTags(String? Function(String tag) edit) {
    var changed = 0;
    noteTags = [
      for (final tags in noteTags)
        () {
          final out = <String>[];
          for (final t in tags) {
            final e = edit(t);
            if (e != null &&
                !out.any((o) => o.toLowerCase() == e.toLowerCase())) {
              out.add(e);
            }
          }
          if (out.join(' ') != tags.join(' ')) changed++;
          return out;
        }(),
    ];
    if (changed > 0) emitChange(kind: 'tags', decks: false, notes: true);
    return changed;
  }

  static String _duplicateText(String html) => html
      .replaceAll(RegExp(r'<[^>]*>'), ' ')
      .replaceAll('&nbsp;', ' ')
      .trim()
      .toLowerCase();

  /// The calls made on a channel.
  List<List<dynamic>> callsTo(String channel) => calls[channel] ?? const [];

  /// Builds filtered deck [id] from [options] like Go: adds or updates its
  /// row in [tree], or fails when its search gathers no card.
  Map<String, Object?> _buildFiltered(
    int id,
    Map<String, Object?> options,
    String kind,
  ) {
    final terms = options['terms'] as List;
    final n = filterMatches((terms.first as Map)['search'] as String);
    if (n == 0) throw noMatchError;
    filteredOptions[id] = options;
    final node = {
      ...deckNode(
        id,
        options['name'] as String,
        newCount: n,
        cards: n,
        filtered: true,
      ),
      if (tree.any((n) => n['id'] == id))
        'collapsed': tree.firstWhere((n) => n['id'] == id)['collapsed'],
    };
    final i = tree.indexWhere((n) => n['id'] == id);
    tree = [...tree];
    if (i < 0) {
      tree.add(node);
      tree.sort(
        (a, b) => (a['fullName'] as String).toLowerCase().compareTo(
          (b['fullName'] as String).toLowerCase(),
        ),
      );
    } else {
      tree[i] = node;
    }
    emitChange(kind: kind);
    return {'deckId': id, 'cards': n};
  }

  /// The trimmed [name], or the error Go gives for an empty or taken one.
  String _profileName(String name, {String? except}) {
    final n = name.trim();
    if (n.isEmpty) throw 'A profile needs a name.';
    for (final p in profiles) {
      final other = p['name'] as String;
      if (p['id'] != except && other.toLowerCase() == n.toLowerCase()) {
        throw 'A profile named "$other" already exists.';
      }
    }
    return n;
  }

  Map<String, Object?> _profileList() => {
    'profiles': [...profiles]
      ..sort(
        (a, b) => (a['name'] as String).toLowerCase().compareTo(
          (b['name'] as String).toLowerCase(),
        ),
      ),
    'current': currentProfile ?? '',
    'showAtStartup': showProfilesAtStartup,
  };

  int _profile(Object? id) {
    final i = profiles.indexWhere((p) => p['id'] == id);
    if (i < 0) throw 'This profile no longer exists.';
    return i;
  }

  /// The trimmed [name], or the error Go gives for an empty or taken one.
  String _presetName(String name, {int except = 0}) {
    final n = name.trim();
    if (n.isEmpty) throw 'Please enter a preset name.';
    for (final p in presets.values) {
      final other = p['name'] as String;
      if (p['id'] != except && other.toLowerCase() == n.toLowerCase()) {
        throw 'A preset named "$other" already exists.';
      }
    }
    return n;
  }

  /// The trimmed [name], or the error Go gives for an empty or taken one.
  String _savedSearchName(String name, {String? except}) {
    final n = name.trim();
    if (n.isEmpty) throw 'A saved search needs a name.';
    for (final s in savedSearches) {
      final other = s['name'] as String;
      if (other != except && other.toLowerCase() == n.toLowerCase()) {
        throw 'A saved search named "$n" already exists.';
      }
    }
    return n;
  }

  /// The index of the saved search [name], or the error Go gives.
  int _savedSearch(String name) {
    final i = savedSearches.indexWhere((s) => s['name'] == name);
    if (i < 0) throw 'there is no saved search named "$name"';
    return i;
  }

  void _sortSavedSearches() => savedSearches.sort(
    (a, b) => (a['name'] as String).toLowerCase().compareTo(
      (b['name'] as String).toLowerCase(),
    ),
  );

  int _addPreset(int deck, String name, Object? options) {
    final n = _presetName(name);
    final id = _nextPreset++;
    presets[id] = {'id': id, 'name': n, 'options': options};
    deckPresets[deck] = id;
    emitChange(kind: 'addPreset');
    return id;
  }

  void _defaults() {
    for (final section in statsSections) {
      handle('stats.$section', (_) async {
        await statsGates[section]?.future;
        if (statsErrors[section] case final error?) throw error;
        return stats[section];
      });
    }
    handle('app.Open', (_) {
      currentProfile ??= lastProfile;
      return status;
    });
    handle('app.LoadProfiles', (_) => _profileList());
    handle('app.Profiles', (_) => _profileList());
    handle('app.AddProfile', (args) {
      final p = {'id': 'p${_nextProfile++}', 'name': _profileName(args[0])};
      profiles = [...profiles, p];
      return p;
    });
    handle('app.RenameProfile', (args) {
      final [id as String, name as String] = args;
      final i = _profile(id);
      profiles = [...profiles]
        ..[i] = {'id': id, 'name': _profileName(name, except: id)};
      return null;
    });
    handle('app.ProfileCardCount', (args) {
      _profile(args[0]);
      return profileCards[args[0]] ?? 0;
    });
    handle('app.DeleteProfile', (args) {
      final i = _profile(args[0]);
      if (args[0] == currentProfile) {
        throw 'The open profile cannot be deleted. Open another profile first.';
      }
      if (profiles.length == 1) throw 'The only profile cannot be deleted.';
      profiles = [...profiles]..removeAt(i);
      return null;
    });
    handle('app.SwitchProfile', (args) {
      final id = args[0] as String;
      _profile(id);
      currentProfile = lastProfile = id;
      emitChange(kind: 'replaced', all: true);
      return profileStatus[id] ?? status;
    });
    handle('app.SetShowProfilesAtStartup', (args) {
      showProfilesAtStartup = args[0] as bool;
      return null;
    });
    handle('app.Status', (_) => status);
    handle('app.StartFresh', (_) {
      status = {...status, 'state': 'fresh', 'message': null};
      return status;
    });
    handle('app.Settings', (_) => settings);
    handle('app.SetTheme', (args) {
      settings = {...settings, 'theme': args[0]};
      return null;
    });
    handle('app.SetBuiltinNames', (args) {
      builtinNames = args[0] as Map<String, Object?>;
      return null;
    });
    handle('app.SetLanguage', (args) {
      settings = {...settings, 'language': args[0]};
      return null;
    });
    handle('app.SetRolloverHour', (args) {
      settings = {...settings, 'rolloverHour': args[0]};
      return null;
    });
    handle('app.SetAutoContinueLearning', (args) {
      settings = {...settings, 'autoContinueLearning': args[0]};
      return null;
    });
    handle('app.SetLearnAheadMinutes', (args) {
      settings = {...settings, 'learnAheadMinutes': args[0]};
      return null;
    });
    handle('app.SetTimeboxMinutes', (args) {
      final minutes = args[0] as int;
      if (minutes < 0 || minutes > 1440) {
        throw 'The timebox time limit must be between 0 and 1440 minutes.';
      }
      settings = {...settings, 'timeboxMinutes': minutes};
      return null;
    });
    handle('app.SetShowRemainingCounts', (args) {
      settings = {...settings, 'showRemainingCounts': args[0]};
      return null;
    });
    handle('app.SetShowIntervals', (args) {
      settings = {...settings, 'showIntervals': args[0]};
      return null;
    });
    handle('app.SetSpaceAnswersGood', (args) {
      settings = {...settings, 'spaceAnswersGood': args[0]};
      return null;
    });
    handle('app.SetHideBarsDuringReview', (args) {
      settings = {...settings, 'hideBarsDuringReview': args[0]};
      return null;
    });
    handle('app.SetQuestionFadeMs', (args) {
      final ms = args[0] as int;
      if (ms < 0 || ms > 1000) {
        throw 'The question fade must be between 0 and 1000 milliseconds.';
      }
      settings = {...settings, 'questionFadeMs': ms};
      return null;
    });
    handle('app.SetUIScale', (args) {
      final percent = args[0] as int;
      if (percent < 75 || percent > 200 || percent % 5 != 0) {
        throw 'The UI size must be between 75 and 200 % in steps of 5 %.';
      }
      settings = {...settings, 'uiScale': percent};
      return null;
    });
    handle('app.SetMinimalist', (args) {
      settings = {...settings, 'minimalist': args[0]};
      return null;
    });
    handle('app.SetReduceMotion', (args) {
      final mode = args[0] as String;
      if (!const ['system', 'on', 'off'].contains(mode)) {
        throw 'Unknown reduce motion choice "$mode".';
      }
      settings = {...settings, 'reduceMotion': mode};
      return null;
    });
    handle('progress.Today', (_) {
      if (progress case final p?) return p;
      throw 'no progress';
    });
    handle('progress.WeeklyRecap', (_) => recap);
    handle('app.SetDeckProgressInList', (args) {
      settings = {...settings, 'deckProgressInList': args[0]};
      return null;
    });
    handle('app.DismissRecap', (args) {
      settings = {...settings, 'recapDismissedWeek': args[0]};
      return null;
    });
    handle('app.SetShowProgress', (args) {
      settings = {...settings, 'showProgress': args[0]};
      return null;
    });
    handle('app.SetShowStudyRewards', (args) {
      settings = {...settings, 'showStudyRewards': args[0]};
      return null;
    });
    handle('progress.Achievements', (_) => achievements);
    handle(
      'progress.Level',
      (_) => level ?? levelJson(answers: progress?['xp'] as int? ?? 0),
    );
    handle('progress.MarkSeen', (args) {
      final keys = List<String>.from(args[0] as List);
      if (progress case final p?) {
        progress = {
          ...p,
          'newAchievements': [
            for (final a in p['newAchievements'] as List? ?? const [])
              if (!keys.contains((a as Map)['key'])) a,
          ],
        };
      }
      return null;
    });
    handle('app.SetRestDaysPerWeek', (args) {
      final n = args[0] as int;
      if (n < 0 || n > 2) {
        throw 'The rest days per week must be between 0 and 2.';
      }
      settings = {...settings, 'restDaysPerWeek': n};
      return null;
    });
    handle('app.SetDailyGoal', (args) {
      final [kind as String, value as int] = args;
      if (!const ['', 'cards', 'minutes'].contains(kind)) {
        throw 'Unknown daily goal "$kind".';
      }
      if (kind.isNotEmpty && (value < 1 || value > 9999)) {
        throw 'The daily goal must be between 1 and 9999.';
      }
      settings = {
        ...settings,
        'dailyGoalKind': kind,
        'dailyGoal': kind.isEmpty ? 0 : value,
      };
      return null;
    });
    handle('app.GoalSuggestionShown', (_) {
      settings = {...settings, 'goalSuggestionShown': true};
      return null;
    });
    handle('app.SetShowPlayButtons', (args) {
      settings = {...settings, 'showPlayButtons': args[0]};
      return null;
    });
    handle('app.SetInterruptAudio', (args) {
      settings = {...settings, 'interruptAudio': args[0]};
      return null;
    });
    handle('app.SetFlagName', (args) {
      final names = List<String>.from(settings['flagNames'] as List);
      names[(args[0] as int) - 1] = (args[1] as String).trim();
      settings = {...settings, 'flagNames': names};
      return null;
    });
    handle('app.SetStripPasteFormatting', (args) {
      settings = {...settings, 'stripPasteFormatting': args[0]};
      return null;
    });
    handle('app.SetPasteImagesAsPng', (args) {
      settings = {...settings, 'pasteImagesAsPng': args[0]};
      return null;
    });
    handle('app.SetIgnoreAccents', (args) {
      settings = {...settings, 'ignoreAccents': args[0]};
      return null;
    });
    handle('app.SetDeckByNoteType', (args) {
      settings = {...settings, 'deckByNoteType': args[0]};
      return null;
    });
    handle('app.SetBrowserColumns', (args) {
      settings = {...settings, 'browserColumns': args[0]};
      return null;
    });
    handle('app.SetBrowserMode', (args) {
      settings = {...settings, 'browserMode': args[0]};
      return null;
    });
    handle('app.SetBrowserSplit', (args) {
      settings = {...settings, 'browserSplit': args[0]};
      return null;
    });
    handle('app.SetDefaultSearch', (args) {
      final query = args[0] as String;
      if (searchError(query) case final error?) throw error;
      settings = {...settings, 'defaultSearch': query.trim()};
      return null;
    });
    handle('browse.NoteTypeOfCards', (args) {
      final types = {
        for (final id in (args[0] as List).cast<int>()) cardNoteTypes[id] ?? 10,
      };
      if (types.length > 1) {
        throw 'All selected cards must use the same note type.';
      }
      return types.single;
    });
    handle('browse.ChangeNoteTypePreview', (args) => noteTypeChange(args));
    handle('browse.ChangeNoteType', (args) {
      final result = noteTypeChange(args);
      emitChange(notes: true);
      return result;
    });
    handle('browse.SavedSearches', (_) => savedSearches);
    handle('browse.FindDuplicates', (args) {
      if (searchError(args[1] as String) case final error?) throw error;
      return duplicateSets;
    });
    handle('browse.SaveSearch', (args) {
      final [name as String, query as String] = args;
      if (searchError(query) case final error?) throw error;
      savedSearches = [
        ...savedSearches,
        {'name': _savedSearchName(name), 'query': query.trim()},
      ];
      _sortSavedSearches();
      return null;
    });
    handle('browse.RenameSavedSearch', (args) {
      final [name as String, newName as String] = args;
      final i = _savedSearch(name);
      final n = _savedSearchName(newName, except: name);
      savedSearches = [...savedSearches];
      savedSearches[i] = {...savedSearches[i], 'name': n};
      _sortSavedSearches();
      return null;
    });
    handle('browse.DeleteSavedSearch', (args) {
      final i = _savedSearch(args[0] as String);
      savedSearches = [...savedSearches]..removeAt(i);
      return null;
    });
    handle('app.SetBackupSettings', (args) {
      settings = {...settings, 'backup': args[0]};
      return null;
    });
    handle('app.PersistenceNoticeShown', (_) => null);
    handle('app.SetRenderEnvironment', (args) {
      final env = Map<String, Object?>.from(args[0] as Map);
      if (!mapEquals(env, renderEnv)) {
        renderEnv = env;
        emitChange(kind: 'environment', decks: false);
      }
      return null;
    });
    handle(
      'app.Backups',
      (_) => {
        'backups': backups,
        'bytes': backups.fold<int>(0, (sum, b) => sum + (b['size'] as int)),
      },
    );
    handle('app.CreateBackup', (_) {
      backupClock = backupClock.add(const Duration(minutes: 1));
      final made = autoBackup(backupClock);
      backups = [made, ...backups];
      return made;
    });
    handle('app.RestoreBackup', (_) => null);
    handle('app.EmptyCards', (_) => emptyCardList(emptyNotes));
    handle('app.DeleteEmptyCards', (args) {
      final ids = (args[0] as List<dynamic>).toSet();
      emptyNotes = [
        for (final n in emptyNotes)
          if (_keptCards(n, ids) case final cards when cards.isNotEmpty)
            {...n, 'cards': cards},
      ];
      return null;
    });
    handle(
      'app.ExportDeletionLog',
      (_) => {
        'name': 'cardflip-deleted-2026-10-13.txt',
        'data': base64Encode(utf8.encode('#separator:tab\n')),
        'notes': deletedNotes,
      },
    );
    handle('app.ClearDeletionLog', (_) {
      deletedNotes = 0;
      return null;
    });
    int bytes(List<Map<String, Object?>> files) =>
        files.fold<int>(0, (sum, f) => sum + (f['size'] as int));
    handle(
      'app.CheckMedia',
      (_) => {
        'unused': unusedMedia,
        'unusedBytes': bytes(unusedMedia),
        'missing': missingMedia,
        'missingNotes': missingMediaNotes,
        'trash': mediaTrash,
        'trashBytes': bytes(mediaTrash),
      },
    );
    handle('app.DeleteUnusedMedia', (_) {
      final n = unusedMedia.length;
      mediaTrash = [...mediaTrash, ...unusedMedia]
        ..sort((a, b) => (a['name'] as String).compareTo(b['name'] as String));
      unusedMedia = [];
      return n;
    });
    handle('app.RestoreMediaTrash', (_) {
      final n = mediaTrash.length;
      unusedMedia = [...unusedMedia, ...mediaTrash]
        ..sort((a, b) => (a['name'] as String).compareTo(b['name'] as String));
      mediaTrash = [];
      return n;
    });
    handle('app.EmptyMediaTrash', (_) {
      mediaTrash = [];
      return null;
    });
    handle('app.TagMissingMedia', (_) => missingMediaNotes);
    handle('app.UndoInfo', (_) => {'undo': undoStep, 'redo': redoStep});
    handle('app.Undo', (_) {
      if (undoStep.isEmpty) throw 'There is nothing to undo.';
      final name = undoStep;
      redoStep = name;
      undoStep = '';
      onUndo?.call();
      return name;
    });
    handle('app.Redo', (_) {
      if (redoStep.isEmpty) throw 'There is nothing to redo.';
      final name = redoStep;
      undoStep = name;
      redoStep = '';
      onRedo?.call();
      return name;
    });
    bridge.receiver.stream('app.Changes', (_) => _changes.stream);
    bridge.receiver.stream('app.CheckDatabase', (data) {
      calls.putIfAbsent('app.CheckDatabase', () => []).add(const []);
      return checkDatabaseEvents();
    });
    for (final channel in [
      'transfer.ExportDeck',
      'transfer.ExportNotes',
      'transfer.ExportCollection',
      'transfer.ExportNotesText',
      'transfer.ExportCardsText',
    ]) {
      bridge.receiver.stream(channel, (data) {
        final args = (data as List<dynamic>?) ?? const [];
        calls.putIfAbsent(channel, () => []).add(args);
        return exportEvents(channel, args);
      });
    }
    for (final channel in ['decks.Optimize', 'decks.OptimizeAll']) {
      bridge.receiver.stream(channel, (data) {
        final args = (data as List<dynamic>?) ?? const [];
        calls.putIfAbsent(channel, () => []).add(args);
        return optimizeEvents(channel, args);
      });
    }
    for (final channel in ['decks.Simulate', 'decks.SuggestRetention']) {
      bridge.receiver.stream(channel, (data) {
        final args = (data as List<dynamic>?) ?? const [];
        calls.putIfAbsent(channel, () => []).add(args);
        return simulationEvents(channel, args);
      });
    }
    handle('decks.RescheduleCount', (_) => rescheduleCount);
    handle('decks.RescheduleCountAll', (_) => rescheduleCountAll);
    handle('media.Add', (args) {
      final name = args[0] as String;
      final data = _bytes(args[1]);
      final dot = name.lastIndexOf('.');
      final ext = dot < 0 ? '' : name.substring(dot);
      final stem = dot < 0 ? name : name.substring(0, dot);
      var stored = stem.isEmpty ? 'media-${media.length + 1}$ext' : name;
      for (var i = 1; media.containsKey(stored); i++) {
        if (listEquals(media[stored], data)) return stored;
        stored = '$stem-$i$ext';
      }
      media[stored] = data;
      return stored;
    });
    handle('media.AddPastedImage', (args) {
      final ext = args[1] == true ? '.png' : '.jpg';
      final name = 'paste-${media.length + 1}$ext';
      media[name] = _bytes(args[0]);
      return name;
    });
    handle('media.Get', (args) {
      final data = media[args[0]];
      if (data == null) throw 'Media file "${args[0]}" is missing.';
      return base64Encode(data);
    });
    handle('decks.Tree', (_) => tree);
    handle('decks.Settings', (args) {
      final deck = args[0] as int;
      final sorted = presets.values.toList()
        ..sort(
          (a, b) => (a['name'] as String).toLowerCase().compareTo(
            (b['name'] as String).toLowerCase(),
          ),
        );
      return {
        'presetId': deckPresets[deck] ?? 1,
        'defaultPresetId': 1,
        'limits':
            deckLimits[deck] ?? {'newPerDay': null, 'reviewsPerDay': null},
        'presets': [
          for (final p in sorted)
            {
              ...p,
              'decks': tree
                  .where((n) => (deckPresets[n['id']] ?? 1) == p['id'])
                  .length,
            },
        ],
      };
    });
    handle('decks.SaveSettings', (args) {
      final [deck as int, preset as int, options as Map, limits as Map] = args;
      final o = options.cast<String, Object?>();
      final retention = o['desiredRetention'] as double;
      if (retention < 0.7 || retention > 0.99) {
        throw 'Desired retention must be between 0.70 and 0.99.';
      }
      for (final n in [o['newPerDay'], limits['newPerDay']]) {
        if (n is int && (n < 0 || n > 9999)) {
          throw 'New cards per day must be between 0 and 9999.';
        }
      }
      final threshold = o['leechThreshold'] as int;
      if (threshold < 1 || threshold > 99) {
        throw 'Leech threshold must be between 1 and 99 lapses.';
      }
      final maxAnswer = o['maxAnswerSeconds'] as int;
      if (maxAnswer < 1 || maxAnswer > 9999) {
        throw 'Maximum answer seconds must be between 1 and 9999.';
      }
      for (final (name, key) in [
        ('Seconds to show question for', 'secondsToShowQuestion'),
        ('Seconds to show answer for', 'secondsToShowAnswer'),
      ]) {
        final seconds = o[key] as int;
        if (seconds < 0 || seconds > 9999) {
          throw '$name must be between 0 and 9999.';
        }
      }
      final historical = o['historicalRetention'] as double;
      if (historical < 0.5 || historical > 0.99) {
        throw 'Historical retention must be between 0.50 and 0.99.';
      }
      final weights = o['weights'] as List;
      if (weights.isNotEmpty && weights.length != 19) {
        throw 'FSRS parameters must be 19 numbers, not ${weights.length}.';
      }
      if (!presets.containsKey(preset)) throw 'That preset no longer exists.';
      presets[preset] = {...presets[preset]!, 'options': o};
      deckPresets[deck] = preset;
      deckLimits[deck] = limits.cast<String, Object?>();
      emitChange(kind: 'deckOptions');
      return null;
    });
    handle('decks.AddPreset', (args) {
      final [deck as int, name as String] = args;
      return _addPreset(deck, name, deckOptionsJson());
    });
    handle('decks.ClonePreset', (args) {
      final [deck as int, name as String] = args;
      final options = presets[deckPresets[deck] ?? 1]!['options'];
      return _addPreset(deck, name, options);
    });
    handle('decks.RenamePreset', (args) {
      final [id as int, name as String] = args;
      presets[id] = {...presets[id]!, 'name': _presetName(name, except: id)};
      emitChange(kind: 'renamePreset');
      return null;
    });
    handle('decks.DeletePreset', (args) {
      final id = args[0] as int;
      if (id == 1) throw 'The Default preset cannot be deleted.';
      presets.remove(id);
      deckPresets.removeWhere((_, p) => p == id);
      emitChange(kind: 'deletePreset');
      return null;
    });
    handle('decks.SavePresetToSubdecks', (args) {
      final deck = args[0] as int;
      final name = tree.firstWhere((n) => n['id'] == deck)['fullName'];
      for (final n in tree) {
        if ((n['fullName'] as String).startsWith('$name::')) {
          deckPresets[n['id'] as int] = deckPresets[deck] ?? 1;
        }
      }
      emitChange(kind: 'presetToSubdecks');
      return null;
    });
    handle('decks.NextChange', (_) => nextChange().toUtc().toIso8601String());
    handle('decks.Overview', (args) {
      final id = args[0] as int;
      final node = tree.firstWhere(
        (n) => n['id'] == id,
        orElse: () => throw 'That deck no longer exists.',
      );
      final counts = node['counts'] as Map;
      final total =
          (counts['new'] as int) +
          (counts['learning'] as int) +
          (counts['review'] as int);
      final description = descriptions[id] ?? '';
      return {
        'counts': counts,
        'buried': buried[id] ?? 0,
        'description': description,
        'descriptionHtml': description.replaceAll(
          RegExp(r'<script>.*?</script>'),
          '',
        ),
        'nextDue': total == 0 ? nextDue[id]?.toUtc().toIso8601String() : null,
      };
    });
    handle('decks.SetDescription', (args) {
      final [id as int, text as String] = args;
      descriptions[id] = text.trim();
      emitChange(kind: 'deckDescription');
      return null;
    });
    handle('decks.Unbury', (args) {
      final id = args[0] as int;
      final n = buried.remove(id) ?? 0;
      tree = [
        for (final node in tree)
          if (node['id'] == id)
            {
              ...node,
              'counts': {
                ...(node['counts'] as Map),
                'review': ((node['counts'] as Map)['review'] as int) + n,
              },
            }
          else
            node,
      ];
      emitChange(kind: 'unbury');
      return null;
    });
    handle('decks.NewFilteredOptions', (args) {
      var i = 1;
      while (tree.any((n) => n['fullName'] == 'Filtered Deck $i')) {
        i++;
      }
      return filteredOptionsJson('Filtered Deck $i', search: args[0] as String);
    });
    handle('decks.FilteredOptions', (args) {
      final options = filteredOptions[args[0]];
      if (options == null) throw 'That deck is not a filtered deck.';
      return options;
    });
    handle('decks.CreateFiltered', (args) {
      final options = Map<String, Object?>.from(args[0] as Map);
      final name = (options['name'] as String).trim();
      if (name.isEmpty) throw 'Please enter a deck name.';
      if (tree.any((n) => n['fullName'] == name)) {
        throw 'A deck named "$name" already exists.';
      }
      final id = _nextDeck++;
      return _buildFiltered(id, {...options, 'name': name}, 'buildFiltered');
    });
    handle('decks.UpdateFiltered', (args) {
      final [id as int, options as Map] = args;
      return _buildFiltered(
        id,
        Map<String, Object?>.from(options),
        'buildFiltered',
      );
    });
    handle('decks.Rebuild', (args) {
      final id = args[0] as int;
      return _buildFiltered(id, filteredOptions[id]!, 'buildFiltered');
    });
    handle('decks.Empty', (args) {
      final id = args[0] as int;
      tree = [
        for (final n in tree)
          if (n['id'] == id)
            {
              ...n,
              'cards': 0,
              'counts': {'new': 0, 'learning': 0, 'review': 0},
            }
          else
            n,
      ];
      emitChange(kind: 'emptyFiltered');
      return null;
    });
    handle('decks.CustomStudy', (args) {
      final [deck as int, choice as Map] = args;
      final deckName =
          tree.firstWhere((n) => n['id'] == deck)['fullName'] as String;
      final search = switch (choice['kind']) {
        'forgotten' => 'deck:"$deckName" rated:${choice['days']}:1',
        'ahead' => 'deck:"$deckName" prop:due<=${choice['days']} prop:due>0',
        'previewNew' => 'deck:"$deckName" is:new added:${choice['days']}',
        _ => 'deck:"$deckName" ${choice['state']}',
      };
      final existing = tree.where(
        (n) => n['fullName'] == 'Custom Study Session',
      );
      final id = existing.isEmpty ? _nextDeck++ : existing.first['id'] as int;
      return _buildFiltered(
        id,
        filteredOptionsJson(
          'Custom Study Session',
          search: search,
          reschedule: choice['kind'] != 'previewNew',
        ),
        'customStudy',
      );
    });
    handle('decks.CustomStudyInfo', (args) {
      final extra = extraCards[args[0]] ?? (newCards: 0, reviews: 0);
      return {'newAvailable': extra.newCards, 'reviewAvailable': extra.reviews};
    });
    handle('decks.IncreaseToday', (args) {
      final [id as int, newBy as int, reviewBy as int] = args;
      if (newBy < 0 || reviewBy < 0) {
        throw 'The increase must be between 0 and 9999 cards.';
      }
      final extra = extraCards[id] ?? (newCards: 0, reviews: 0);
      final addNew = newBy < extra.newCards ? newBy : extra.newCards;
      final addReview = reviewBy < extra.reviews ? reviewBy : extra.reviews;
      extraCards[id] = (
        newCards: extra.newCards - addNew,
        reviews: extra.reviews - addReview,
      );
      tree = [
        for (final n in tree)
          if (n['id'] == id)
            {
              ...n,
              'counts': {
                ...(n['counts'] as Map<String, Object?>),
                'new': ((n['counts'] as Map)['new'] as int) + addNew,
                'review': ((n['counts'] as Map)['review'] as int) + addReview,
              },
            }
          else
            n,
      ];
      emitChange(kind: 'todayLimits');
      return null;
    });
    handle('notes.Types', (_) => noteTypes);
    // Adds Image Occlusion as id 13, as in collections from before it was
    // built in.
    handle('notes.ImageOcclusionType', (_) {
      for (final t in noteTypes) {
        if (t['imageOcclusion'] == true) return t['id'];
      }
      noteTypes = [...noteTypes, occlusionTypeJson()];
      return 13;
    });
    // Like Go: a note type or deck that is gone falls back to Basic (or
    // the first note type) and the first regular deck.
    handle('notes.Defaults', (args) {
      var typeId = args[0] == 0 ? 10 : args[0] as int;
      if (!noteTypes.any((t) => t['id'] == typeId) && noteTypes.isNotEmpty) {
        typeId =
            (noteTypes.firstWhere(
                  (t) => t['name'] == 'Basic',
                  orElse: () => noteTypes.first,
                )['id'])
                as int;
      }
      final follows = settings['deckByNoteType'] == true;
      var deckId = follows ? deckByType[typeId] ?? lastDeck : lastDeck;
      if (!tree.any((d) => d['id'] == deckId && d['filtered'] != true)) {
        deckId =
            tree.firstWhere(
                  (d) => d['filtered'] != true,
                  orElse: () => {'id': deckId},
                )['id']
                as int;
      }
      return {'typeId': typeId, 'deckId': deckId};
    });
    handle('notes.Duplicates', (args) {
      final [typeId as int, first as String, exclude as int] = args;
      final key = _duplicateText(first);
      return [
        if (key.isNotEmpty)
          for (final n in existingNotes)
            if (n.typeId == typeId &&
                n.id != exclude &&
                _duplicateText(n.first) == key)
              n.id,
      ];
    });
    // Every row of the file has the default note type, unless a test says
    // otherwise.
    handle('transfer.CsvNoteTypes', (args) {
      final options = args[1] as Map<String, dynamic>;
      final type = noteTypes.firstWhere(
        (t) => t['id'] == options['typeId'],
        orElse: () => noteTypes.first,
      );
      return {
        'unknown': <String>[],
        'unknownRows': 0,
        'fields': type['fields'],
      };
    });
    handle('notes.Tags', (_) {
      final seen = <String, String>{};
      for (final tags in noteTags) {
        for (final t in tags) {
          seen.putIfAbsent(t.toLowerCase(), () => t);
        }
      }
      return seen.values.toList()
        ..sort((a, b) => a.toLowerCase().compareTo(b.toLowerCase()));
    });
    handle('notes.TagUsage', (args) {
      final roots = (args[0] as List).cast<String>();
      return noteTags
          .where((tags) => tags.any((t) => roots.any((r) => _inTree(t, r))))
          .length;
    });
    handle('notes.RenameTag', (args) {
      final [old as String, name as String] = args;
      if (name.trim().isEmpty) throw 'Enter a tag name.';
      return _editTags(
        (t) => _inTree(t, old) ? name.trim() + t.substring(old.length) : t,
      );
    });
    handle('notes.MoveTags', (args) {
      final roots = (args[0] as List).cast<String>();
      final parent = args[1] as String;
      if (roots.any((r) => parent.isNotEmpty && _inTree(parent, r))) {
        throw 'A tag cannot be moved into itself or one of its child tags.';
      }
      return _editTags((t) {
        for (final r in roots) {
          if (!_inTree(t, r)) continue;
          final base = r.split('::').last;
          final moved = parent.isEmpty ? base : '$parent::$base';
          return moved + t.substring(r.length);
        }
        return t;
      });
    });
    handle('notes.DeleteTags', (args) {
      final roots = (args[0] as List).cast<String>();
      return _editTags((t) => roots.any((r) => _inTree(t, r)) ? null : t);
    });
    handle('notes.ClearUnusedTags', (_) {
      final n = unusedTags;
      unusedTags = 0;
      return n;
    });
    handle('notes.StockTypes', (_) => stockTypes);
    handle('notes.AddType', (args) {
      final [stock as String, name as String] = args;
      final base = noteTypes.firstWhere(
        (t) => t['name'] == stock,
        orElse: () => noteTypeJson(0, stock, ['Front', 'Back']),
      );
      return _addType({...base, 'name': name});
    });
    handle('notes.CloneType', (args) {
      final [id as int, name as String] = args;
      final base = noteTypes.firstWhere((t) => t['id'] == id);
      return _addType({...base, 'name': name});
    });
    handle('notes.TypeDeletionInfo', (args) {
      final type = noteTypes.firstWhere((t) => t['id'] == args[0]);
      final notes = type['noteCount'] as int;
      return {
        'notes': notes,
        'cards': notes * (type['templates'] as List).length,
      };
    });
    handle('notes.DeleteType', (args) {
      if (noteTypes.length == 1) {
        throw 'The last note type cannot be deleted.';
      }
      noteTypes = [
        for (final t in noteTypes)
          if (t['id'] != args[0]) t,
      ];
      return null;
    });
    handle('notes.SetFieldSticky', (args) {
      _setFieldOption(args, 'sticky');
      return null;
    });
    handle('notes.SetFieldOptions', (args) {
      final [typeId as int, index as int, options as Map] = args;
      _editType(
        typeId,
        (t) => {
          ...t,
          'fieldOptions': [
            for (final (i, o) in (t['fieldOptions'] as List).indexed)
              i == index ? {...options} : o,
          ],
        },
      );
      return null;
    });
    handle('notes.SetSortField', (args) {
      final [typeId as int, index as int] = args;
      _editType(typeId, (t) => {...t, 'sortField': index});
      return null;
    });
    handle('notes.MoveField', (args) {
      final [typeId as int, from as int, to as int] = args;
      List<Object?> move(List<Object?> l) {
        final out = [...l];
        return out..insert(to, out.removeAt(from));
      }

      _editType(typeId, (t) {
        // Like Go: the sort field stays on its field.
        final sorted = (t['fields'] as List)[t['sortField'] as int? ?? 0];
        final fields = move(t['fields'] as List);
        final options = move(t['fieldOptions'] as List);
        return {
          ...t,
          'fields': fields,
          'fieldOptions': options,
          'sortField': fields.indexOf(sorted),
        };
      });
      return null;
    });
    handle('notes.UpdateTemplate', (args) {
      final [typeId as int, index as int, template as Map] = args;
      _editTemplates(typeId, (ts) => ts..[index] = {...template});
      return null;
    });
    handle('notes.MoveTemplate', (args) {
      final [typeId as int, from as int, to as int] = args;
      _editTemplates(typeId, (ts) => ts..insert(to, ts.removeAt(from)));
      return null;
    });
    handle(
      'notes.Type',
      (args) => noteTypes.firstWhere((t) => t['id'] == args[0]),
    );
    handle(
      'notes.PreviewNotes',
      (args) => [
        for (final n in previewNotes[args[0]] ?? const <PreviewNote>[])
          {'id': n.id, 'label': n.fields.values.first},
      ],
    );
    handle('notes.PreviewTemplate', (args) {
      final [
        typeId as int,
        template as Map,
        css as String,
        noteId as int,
        ord as int,
        night as bool,
      ] = args;
      final type = noteTypes.firstWhere((t) => t['id'] == typeId);
      final note = previewNotes[typeId]?.firstWhere((n) => n.id == noteId);
      final fields = {
        for (final f in type['fields'] as List)
          f as String: note?.fields[f] ?? '($f)',
      };
      final cloze = <int>[
        if (type['cloze'] == true) ...note?.clozeNumbers ?? const [1],
      ];
      final card = cloze.contains(ord) ? ord : (cloze.firstOrNull ?? ord + 1);
      String fill(String tmpl, String front) => tmpl
          .replaceAll('{{FrontSide}}', front)
          .replaceAllMapped(
            RegExp(r'\{\{(?:cloze:)?([^}]+)\}\}'),
            (m) => fields[m[1]] ?? '⚠ unknown field ${m[1]}',
          );
      final classes = 'card card$card${night ? ' nightMode' : ''}';
      final q = fill(template['qfmt'] as String, '');
      final a = fill(template['afmt'] as String, q);
      return {
        'question': '<div class="$classes">$q</div>',
        'answer': '<div class="$classes">$a</div>',
        'color': RegExp(r'color:\s*([^;}\s]+)').firstMatch(css)?[1],
        'questionMedia': const [],
        'answerMedia': const [],
        'fonts': const [],
        'clozeNumbers': cloze,
      };
    });
    handle('notes.CheckTemplate', (args) {
      final [typeId as int, _, template as Map] = args;
      final type = noteTypes.firstWhere((t) => t['id'] == typeId);
      final known = {...type['fields'] as List, ..._specialFields};
      final problems = <String>[];
      for (final (side, key) in [('front', 'qfmt'), ('back', 'afmt')]) {
        final tags = RegExp(r'\{\{([^}]*)\}\}')
            .allMatches(template[key] as String)
            .map((m) => m[1]!.trim());
        for (final tag in tags.toSet()) {
          final name = tag.replaceFirst(RegExp('^[#^/]'), '').split(':').last;
          if (!known.contains(name)) {
            problems.add(
              'The $side template uses {{$tag}}, but there is no field "$name".',
            );
          }
        }
      }
      return problems;
    });
    handle('mcp.Supported', (_) => mcpSupported);
    handle('mcp.Status', (_) => mcpStatus);
    handle('mcp.SetEnabled', (args) {
      final [on as bool] = args;
      var token = mcpStatus['token'] as String;
      if (on && token.isEmpty) token = 'token-${_nextMcpToken++}';
      return _setMcp(enabled: on, token: token);
    });
    handle('mcp.SetPort', (args) {
      final [port as int] = args;
      if (port < 1024 || port > 65535) {
        throw 'The port must be between 1024 and 65535.';
      }
      return _setMcp(port: port);
    });
    handle(
      'mcp.RegenerateToken',
      (_) => _setMcp(token: 'token-${_nextMcpToken++}'),
    );
    handle('mcp.Activity', (_) => mcpActivity.reversed.toList());
    bridge.receiver.stream('mcp.StatusChanges', (_) async* {
      yield mcpStatus;
      yield* _mcpChanges.stream;
    });
    _handleSync();
  }

  void _handleSync() {
    void sync(String name, Object? Function(List<dynamic> args) fn) {
      handle('sync.$name', (args) {
        syncCalls.add([name, ...args].join(' '));
        return fn(args);
      });
    }

    void update(Map<String, Object?> changes) =>
        setSyncStatus(_syncWith(changes));
    sync('Supported', (_) => syncSupported);
    sync('Status', (_) => syncStatus);
    bridge.receiver.stream('sync.StatusChanges', (_) async* {
      yield syncStatus;
      yield* _syncChanges.stream;
    });
    sync('Init', (args) {
      final [name as String] = args;
      if (syncStatus['ready'] != true) {
        update({
          'ready': true,
          'deviceId': 'LAPTOPDEVICEID01',
          'deviceName': name.isEmpty ? 'Laptop' : name,
        });
      }
      return syncStatus;
    });
    sync('SetDeviceName', (args) {
      final [name as String] = args;
      final n = name.trim();
      if (n.isEmpty || n.runes.length > 40) {
        throw 'The device name must be 1 to 40 characters long.';
      }
      update({'deviceName': n});
      return null;
    });
    sync('SetNetwork', (args) {
      final [kind as String, wifi as String, subnet as String] = args;
      final id = kind == 'cellular' || kind == 'none'
          ? ''
          : kind == 'wifi' && wifi.isNotEmpty
          ? 'wifi:$wifi'
          : subnet.isEmpty
          ? ''
          : 'subnet:$subnet';
      update({
        'network': {
          'kind': kind,
          if (wifi.isNotEmpty) 'wifi': wifi,
          if (subnet.isNotEmpty) 'subnet': subnet,
          if (id.isNotEmpty) 'id': id,
          'trusted': false,
        },
      });
      return null;
    });
    sync('TrustNetwork', (_) {
      final net = syncStatus['network'] as Map<String, Object?>;
      final id = net['id'] as String?;
      if (id == null || id.isEmpty) {
        throw 'This device is not on a network that can be trusted.';
      }
      update({
        'trusted': [
          ...(syncStatus['trusted'] as List),
          {
            'id': id,
            if (id.startsWith('wifi:')) 'wifi': net['wifi'],
            if (!id.startsWith('wifi:')) 'subnet': net['subnet'],
            if (net['kind'] == 'ethernet') 'wired': true,
          },
        ],
      });
      return null;
    });
    sync('UntrustNetwork', (args) {
      final [id as String] = args;
      update({
        'trusted': [
          for (final n in syncStatus['trusted'] as List)
            if ((n as Map)['id'] != id) n,
        ],
      });
      return null;
    });
    sync('StartPairing', (_) {
      if (syncStatus['listening'] != true) throw 'Trust this network first.';
      update({'pairing': true});
      return null;
    });
    sync('StopPairing', (_) {
      update({'pairing': false});
      return null;
    });
    sync('PairWith', (args) {
      final [host as String, port as int] = args;
      if (syncPairError case final error?) throw error;
      final found = (syncStatus['discovered'] as List)
          .cast<Map<String, Object?>>()
          .where((d) => d['host'] == host && d['port'] == port);
      final name = found.isEmpty ? 'Phone' : found.first['name'] as String;
      final id = found.isEmpty
          ? 'PHONEDEVICEID001'
          : found.first['id'] as String;
      update({
        'paired': [
          for (final d in syncStatus['paired'] as List)
            if ((d as Map)['id'] != id) d,
          pairedDeviceJson(id, name),
        ],
        'discovered': [
          for (final d in syncStatus['discovered'] as List)
            if ((d as Map)['id'] != id) d,
        ],
      });
      return null;
    });
    sync('Unpair', (args) {
      final [id as String] = args;
      update({
        'paired': [
          for (final d in syncStatus['paired'] as List)
            if ((d as Map)['id'] != id) d,
        ],
        'links': [
          for (final l in syncStatus['links'] as List)
            if ((l as Map)['device'] != id) l,
        ],
      });
      return null;
    });
    sync('Seen', (_) => null);
    sync('Lost', (_) => null);
    sync('Foreground', (_) => null);
    sync('Background', (_) => null);
    sync('RemoteProfiles', (_) => remoteProfiles);
    sync('Link', (args) {
      final [device as String, remote as String] = args;
      final peer = (syncStatus['paired'] as List)
          .cast<Map<String, Object?>>()
          .firstWhere((d) => d['id'] == device);
      final profile = remote.isEmpty
          ? {'id': 'p9', 'name': 'Default'}
          : remoteProfiles.firstWhere((p) => p['id'] == remote);
      final link = linkInfoJson(
        'link-${(syncStatus['links'] as List).length + 1}',
        device,
        peer['name'] as String,
        remoteProfile: profile['id'] as String,
        remoteName: profile['name'] as String,
        needsChoice: true,
      );
      update({
        'links': [...(syncStatus['links'] as List), link],
      });
      return link;
    });
    sync('Unlink', (args) {
      final [id as String] = args;
      update({
        'links': [
          for (final l in syncStatus['links'] as List)
            if ((l as Map)['id'] != id) l,
        ],
      });
      return null;
    });
    handle('sync.SyncNow', (_) async {
      syncCalls.add('SyncNow');
      await syncNowGate?.future;
      return syncNowResult;
    });
  }

  /// The names `notes.CheckTemplate` accepts besides the note type's
  /// fields; it reports every other name in `{{…}}` as an unknown field.
  static const _specialFields = [
    'FrontSide',
    'Tags',
    'Type',
    'Deck',
    'Subdeck',
    'Card',
    'CardFlag',
  ];

  /// The notes `notes.PreviewNotes` lists per note type ID, which
  /// `notes.PreviewTemplate` fills templates from; without notes it uses
  /// "(Field)". It substitutes `{{Field}}`, `{{cloze:Field}}` and
  /// `{{FrontSide}}`, and takes the text colour from `color:` in the CSS.
  Map<int, List<PreviewNote>> previewNotes = {};

  /// The names `notes.StockTypes` lists. `notes.AddType` copies the note
  /// type of the same name from [noteTypes], or makes a Front/Back one.
  /// The names last sent with `app.SetBuiltinNames`.
  Map<String, Object?>? builtinNames;

  List<String> stockTypes = [
    'Basic',
    'Basic (and reversed card)',
    'Basic (optional reversed card)',
    'Basic (type in the answer)',
    'Cloze',
  ];

  /// Adds a copy of [type] without notes to [noteTypes], rejecting a taken
  /// name as Go does, and returns its new ID.
  int _addType(Map<String, Object?> type) {
    final name = (type['name'] as String).trim();
    for (final t in noteTypes) {
      if ((t['name'] as String).toLowerCase() == name.toLowerCase()) {
        throw 'A note type named "${t['name']}" already exists.';
      }
    }
    final id =
        noteTypes.fold<int>(
          0,
          (m, t) => t['id'] as int > m ? t['id'] as int : m,
        ) +
        1;
    noteTypes = [
      ...noteTypes,
      {...type, 'id': id, 'name': name, 'noteCount': 0},
    ];
    return id;
  }
}

/// JSON for a note with empty cards: card ids with their names.
Map<String, Object?> emptyNote(
  int noteId,
  String text,
  Map<int, String> cards, {
  bool allEmpty = false,
}) => {
  'noteId': noteId,
  'text': text,
  'cards': [
    for (final e in cards.entries) {'cardId': e.key, 'name': e.value},
  ],
  'allEmpty': allEmpty,
};

/// JSON for the empty card list of [notes].
Map<String, Object?> emptyCardList(List<Map<String, Object?>> notes) {
  var total = 0, deletable = 0;
  for (final n in notes) {
    final count = (n['cards'] as List).length;
    total += count;
    deletable += n['allEmpty'] == true ? count - 1 : count;
  }
  return {'notes': notes, 'total': total, 'deletable': deletable};
}

/// An export progress event.
Map<String, Object?> exportProgress(String stage, int done, int total) => {
  'stage': stage,
  'done': done,
  'total': total,
  'file': null,
  'error': '',
};

/// The last event of a finished export: the file [name] with a zip header.
Map<String, Object?> exportDone(String name) => {
  'stage': 'done',
  'done': 1,
  'total': 1,
  'file': {'name': name, 'data': 'UEsDBA=='},
  'error': '',
};

/// The last event of a failed export.
Map<String, Object?> exportFailed(String error) => {
  'stage': '',
  'done': 0,
  'total': 0,
  'file': null,
  'error': error,
};

/// A check database progress event.
Map<String, Object?> checkProgress(String stage, int done, int total) => {
  'stage': stage,
  'done': done,
  'total': total,
  'report': null,
  'error': '',
};

/// The last event of a finished check, with [report].
Map<String, Object?> checkDone(Map<String, Object?> report) => {
  'stage': 'done',
  'done': 1,
  'total': 1,
  'report': report,
  'error': '',
};

/// A check database report with [problems] as (kind, count, message).
Map<String, Object?> checkReport({
  List<(String, int, String)> problems = const [],
  int orphanedReviews = 0,
}) => {
  'problems': [
    for (final (kind, count, message) in problems)
      {'kind': kind, 'count': count, 'message': message},
  ],
  'orphanedReviews': orphanedReviews,
};

/// JSON for an automatic backup made at [at].
Map<String, Object?> autoBackup(DateTime at, {int size = 2048}) {
  final u = at.toUtc();
  String two(int n) => n.toString().padLeft(2, '0');
  return {
    'name':
        'backup-${u.year}-${two(u.month)}-${two(u.day)}-'
        '${two(u.hour)}${two(u.minute)}${two(u.second)}-000.cardflip',
    'createdAt': u.toIso8601String(),
    'size': size,
  };
}

/// JSON for a history entry of card info. Durations are in nanoseconds, as
/// Go encodes them.
Map<String, Object?> historyEntry(
  DateTime time,
  int rating, {
  String kind = 'review',
  Duration? interval,
  Duration answerTime = const Duration(seconds: 4),
}) => {
  'time': time.toUtc().toIso8601String(),
  'kind': kind,
  'rating': rating,
  'interval': _nanos(interval),
  'answerTime': _nanos(answerTime),
};

int? _nanos(Duration? d) => d == null ? null : d.inMicroseconds * 1000;

/// JSON for card info: a New card at [position], or with [due] a card in
/// [state] with a memory state and [history] (newest first).
Map<String, Object?> cardInfoJson(
  int id, {
  String state = 'new',
  int? position = 1,
  DateTime? due,
  Duration? interval,
  int reviews = 0,
  int lapses = 0,
  List<Map<String, Object?>> history = const [],
}) {
  final added = DateTime.utc(2026, 9, 1, 10);
  final answers = history.where((h) => h['rating'] != 0).toList();
  final total = answers.fold<int>(0, (t, h) => t + (h['answerTime'] as int));
  return {
    'cardId': id,
    'noteId': id + 100,
    'added': added.toIso8601String(),
    'firstReview': answers.isEmpty ? null : answers.last['time'],
    'latestReview': answers.isEmpty ? null : answers.first['time'],
    'state': due == null ? 'new' : state,
    'due': due?.toUtc().toIso8601String(),
    'position': due == null ? position : null,
    'interval': _nanos(interval),
    'stability': due == null ? null : 12.5,
    'difficulty': due == null ? null : 5.2,
    'retrievability': due == null ? null : 0.93,
    'reviews': reviews,
    'lapses': lapses,
    'averageTime': answers.isEmpty ? null : total ~/ answers.length,
    'totalTime': answers.isEmpty ? null : total,
    'cardType': 'Card 1',
    'noteType': 'Basic',
    'deck': 'Default',
    'history': history,
  };
}

/// A picker that returns prepared staged files.
class FakePicker extends FilePickerService {
  FakePicker();

  golang.StageInfo? next;
  final saved = <String, Uint8List>{};
  List<String>? lastExtensions;

  /// What `pickMedia` returns; empty is a cancelled picker.
  List<PickedFile> media = [];
  int mediaPicks = 0;

  @override
  Future<List<PickedFile>> pickMedia() async {
    mediaPicks++;
    return media;
  }

  @override
  Future<golang.StageInfo?> pick(List<String> extensions) async {
    lastExtensions = extensions;
    return next;
  }

  @override
  Future<bool> save(String fileName, Uint8List bytes) async {
    saved[fileName] = bytes;
    return true;
  }
}

/// Sets the test window to a logical size.
void setSize(WidgetTester tester, Size size) {
  tester.view.physicalSize = size * tester.view.devicePixelRatio;
  addTearDown(tester.view.resetPhysicalSize);
}

const phone = Size(400, 800);
const desktop = Size(1280, 800);

/// Starts the whole app against [go].
Future<void> pumpApp(
  WidgetTester tester, {
  FakePicker? picker,
  Size size = phone,
  bool isWeb = false,
  SyncPlatform? syncPlatform,
}) async {
  setSize(tester, size);
  await tester.pumpWidget(
    CardflipBootstrap(
      dataDir: () async => '/data',
      isWeb: isWeb,
      picker: picker ?? FakePicker(),
      syncPlatform: syncPlatform ?? FakeSyncPlatform(),
    ),
  );
  await tester.pumpAndSettle();
}

/// A staged file, as the picker returns it.
golang.StageInfo stageInfo(int id, String name, String kind) =>
    golang.StageInfo(id: id, name: name, kind: kind, size: 100);

/// Presses [keys] together and releases them.
Future<void> pressKeys(
  WidgetTester tester,
  List<LogicalKeyboardKey> keys,
) async {
  for (final k in keys) {
    await tester.sendKeyDownEvent(k);
  }
  for (final k in keys.reversed) {
    await tester.sendKeyUpEvent(k);
  }
  await tester.pumpAndSettle();
}

/// A note of [FakeGo.previewNotes]: its fields by name, the first one
/// naming it in the note picker.
class PreviewNote {
  const PreviewNote(this.id, this.fields, {this.clozeNumbers = const [1]});

  final int id;
  final Map<String, String> fields;
  final List<int> clozeNumbers;
}

/// The sections of the statistics screen, as `StatsService` names them.
const statsSections = [
  'Today',
  'Streaks',
  'FutureDue',
  'Calendar',
  'Reviews',
  'ReviewTime',
  'Counts',
  'Intervals',
  'Stability',
  'Difficulty',
  'Retrievability',
  'Hourly',
  'Buttons',
  'TrueRetention',
];

/// JSON for `progress.Today`: a [current] streak, [cards] and [minutes]
/// today, a goal of [goal] [goalKind] ("cards" or "minutes"; "" for none)
/// and [xp] XP. Today is studied when cards were answered; the streak is at
/// risk otherwise.
Map<String, Object?> progressJson({
  int current = 0,
  int cards = 0,
  int minutes = 0,
  String goalKind = '',
  int goal = 0,
  int? longest,
  int xp = 0,
  List<Map<String, Object?>> newAchievements = const [],
}) => {
  'newAchievements': newAchievements,
  ...levelFields(xp),
  'current': current,
  'atRisk': current > 0 && cards == 0,
  'todayStudied': cards > 0,
  'longest': longest ?? current,
  'todayCards': cards,
  'todayMinutes': minutes,
  'goalKind': goalKind,
  'goal': goal,
  'goalReached': switch (goalKind) {
    'cards' => cards >= goal,
    'minutes' => minutes >= goal,
    _ => false,
  },
};

/// The level of [xp] XP as Go reports it: 50 × n XP from level n to n + 1.
Map<String, Object?> levelFields(int xp) {
  var level = 1;
  while (25 * (level + 1) * level <= xp) {
    level++;
  }
  return {
    'level': level,
    'xp': xp,
    'levelStart': 25 * level * (level - 1),
    'nextLevelXp': 25 * (level + 1) * level,
  };
}

/// JSON for `progress.Level`: the XP from each source and their level.
Map<String, Object?> levelJson({
  int answers = 0,
  int studyDays = 0,
  int goalDays = 0,
  int longTerm = 0,
  int achievements = 0,
}) => {
  ...levelFields(answers + studyDays + goalDays + longTerm + achievements),
  'fromAnswers': answers,
  'fromStudyDays': studyDays,
  'fromGoalDays': goalDays,
  'fromLongTerm': longTerm,
  'fromAchievements': achievements,
};

/// The tiers of each achievement, as Go defines them.
const achievementTiers = {
  'answers': [100, 1000, 10000, 100000],
  'streak': [7, 30, 100, 365],
  'study-days': [30, 100, 365, 1000],
  'study-time': [10, 100, 500, 1000],
  'daily-goal': [7, 30, 100],
  'long-term': [100, 1000, 5000],
  'level': [5, 10, 25, 50],
  'comeback': [1],
};

/// JSON for one entry of `progress.Achievements`: achievement [id] with
/// [progress] (by default none, level 1 for Level) and the dates its first
/// tiers were [earned] ("2006-01-02").
Map<String, Object?> achievementJson(
  String id, {
  int? progress,
  List<String> earned = const [],
  bool needsGoal = false,
}) {
  final tiers = achievementTiers[id]!;
  return {
    'id': id,
    'tiers': tiers,
    'tierNames': tiers.length == 1
        ? ['gold']
        : ['bronze', 'silver', 'gold', 'platinum'].sublist(0, tiers.length),
    'progress': progress ?? (id == 'level' ? 1 : 0),
    'earned': earned,
    'needsGoal': needsGoal,
  };
}

/// JSON for a tier of `Progress.newAchievements`: achievement [id]'s tier
/// of [threshold].
Map<String, Object?> newAchievementJson(String id, int threshold) {
  final tiers = achievementTiers[id]!;
  final i = tiers.indexOf(threshold);
  return {
    'id': id,
    'threshold': threshold,
    'tierName': tiers.length == 1
        ? 'gold'
        : ['bronze', 'silver', 'gold', 'platinum'][i],
    'key': '$id:$threshold',
  };
}

/// JSON for the Streaks section: a 9-day streak, a 41-day longest streak,
/// 24 of 30 study days, and [goalDays] days that met the goal (null without
/// a goal).
Map<String, Object?> streakSectionJson({int? goalDays}) => {
  'current': 9,
  'atRisk': false,
  'todayStudied': true,
  'longest': 41,
  'longestFirst': '2026-03-02',
  'longestLast': '2026-04-11',
  'studyDays': 24,
  'periodDays': 30,
  'goalDays': ?goalDays,
};

/// JSON for a histogram with [counts] in buckets of [width] from [start].
Map<String, Object?> histogramJson(
  List<int> counts, {
  double start = 0,
  double width = 1,
  double average = 0,
  double sum = 0,
}) => {
  'buckets': [
    for (final (i, c) in counts.indexed)
      {'lo': start + i * width, 'hi': start + (i + 1) * width, 'count': c},
  ],
  'count': counts.fold<int>(0, (a, b) => a + b),
  'average': average,
  'sum': sum,
};

/// JSON for a review history with one bucket per [days] entry, each
/// `[learning, young, mature, relearning, filtered]`, the last being today.
Map<String, Object?> reviewHistoryJson(
  List<List<int>> days, {
  String unit = 'day',
}) {
  final step = switch (unit) {
    'week' => 7,
    'month' => 30,
    _ => 1,
  };
  final total = days.expand((d) => d).fold<int>(0, (a, b) => a + b);
  return {
    'unit': unit,
    'today': '2026-10-13',
    'buckets': [
      for (final (i, d) in days.indexed)
        {
          'start': (i - days.length + 1) * step,
          'learning': d[0],
          'young': d[1],
          'mature': d[2],
          'relearning': d[3],
          'filtered': d[4],
        },
    ],
    'total': total,
    'daysStudied': days.where((d) => d.any((v) => v > 0)).length,
    'periodDays': days.length * step,
  };
}

Map<String, Object?> _retentionCount(int passed, int failed) => {
  'passed': passed,
  'failed': failed,
  'retention': passed + failed == 0 ? null : passed / (passed + failed),
};

/// JSON for every statistics section: 40 cards today, a month of reviews,
/// a small forecast and a few memory states.
Map<String, Object?> statsJson() => {
  'Streaks': streakSectionJson(),
  'Today': {
    'cards': 40,
    'time': const Duration(minutes: 10).inMicroseconds * 1000,
    'again': 4,
    'correct': 0.9,
    'learning': 6,
    'review': 30,
    'relearning': 4,
    'filtered': 0,
  },
  'FutureDue': {
    'days': [5, 30, 12, ...List.filled(27, 0)],
    'cumulative': [5, 35, 47, ...List.filled(27, 47)],
    'overdue': 3,
    'total': 47,
    'average': 47 / 30,
    'dueTomorrow': 30,
  },
  'Calendar': {
    'today': '2026-10-13',
    'days': [
      {'day': -2, 'count': 10},
      {'day': 0, 'count': 40},
    ],
    'max': 40,
  },
  'Reviews': reviewHistoryJson([
    for (var i = 0; i < 29; i++) [0, 0, 0, 0, 0],
    [6, 20, 10, 4, 0],
  ]),
  'ReviewTime': reviewHistoryJson([
    for (var i = 0; i < 29; i++) [0, 0, 0, 0, 0],
    [60000, 300000, 180000, 60000, 0],
  ]),
  'Counts': {
    'new': 100,
    'learning': 5,
    'relearning': 2,
    'young': 40,
    'mature': 60,
    'suspended': 3,
    'buried': 1,
    'total': 211,
  },
  'Intervals': histogramJson([10, 20, 30], start: 1, width: 7, average: 9),
  'Stability': histogramJson([15, 25, 20], width: 10, average: 14),
  'Difficulty': histogramJson(
    [0, 0, 5, 10, 20, 15, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    width: 0.05,
    average: 0.22,
  ),
  'Retrievability': histogramJson(
    [...List.filled(17, 0), 10, 30, 20],
    width: 0.05,
    average: 0.92,
    sum: 55.2,
  ),
  'Hourly': [
    for (var h = 0; h < 24; h++)
      {'reviews': h == 9 ? 40 : 0, 'correct': h == 9 ? 0.9 : 0},
  ],
  'Buttons': {
    'learning': {
      'counts': [2, 1, 3, 0],
      'total': 6,
      'correct': 4 / 6,
    },
    'young': {
      'counts': [1, 2, 15, 2],
      'total': 20,
      'correct': 0.95,
    },
    'mature': {
      'counts': [1, 1, 7, 1],
      'total': 10,
      'correct': 0.9,
    },
  },
  'TrueRetention': {
    'desired': 0.9,
    'rows': [
      for (final (period, young, mature) in [
        ('today', (19, 1), (9, 1)),
        ('yesterday', (0, 0), (0, 0)),
        ('week', (30, 3), (40, 4)),
        ('month', (100, 10), (180, 20)),
        ('year', (300, 30), (500, 50)),
        ('all', (300, 30), (500, 50)),
      ])
        {
          'period': period,
          'young': _retentionCount(young.$1, young.$2),
          'mature': _retentionCount(mature.$1, mature.$2),
          'total': _retentionCount(young.$1 + mature.$1, young.$2 + mature.$2),
        },
    ],
  },
};

/// Starts studying [deck] from the deck list, with the deck menu's Study,
/// which skips the deck overview.
Future<void> studyDeck(WidgetTester tester, String deck) async {
  final row = find
      .ancestor(of: find.text(deck), matching: find.byType(InkWell))
      .first;
  await tester.tap(
    find.descendant(of: row, matching: find.byTooltip('Deck actions')),
  );
  await tester.pumpAndSettle();
  await tester.tap(find.text('Study').last);
  await tester.pumpAndSettle();
}
