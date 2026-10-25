// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Cardflip';

  @override
  String get errorUnknown => 'Something went wrong.';

  @override
  String errorUnexpected(String error) {
    return 'Something went wrong: $error';
  }

  @override
  String get undo => 'Undo';

  @override
  String goApkgCollectionDamaged(String error) {
    return 'the collection is damaged ($error)';
  }

  @override
  String goApkgDatabaseDamaged(String error) {
    return 'the collection database is damaged ($error)';
  }

  @override
  String goApkgDecksUnreadable(String error) {
    return 'its decks could not be read ($error)';
  }

  @override
  String goApkgEmptyNoteType(String name) {
    return 'note type \"$name\" has no fields or templates; unsupported Anki package version';
  }

  @override
  String goApkgFileDamaged(String file, String error) {
    return '$file is damaged ($error)';
  }

  @override
  String goApkgFileError(String file, String error) {
    return '$file: $error';
  }

  @override
  String goApkgFileMissing(String file) {
    return '$file is missing';
  }

  @override
  String goApkgMediaDamaged(String file, String error) {
    return 'media file $file is damaged ($error)';
  }

  @override
  String goApkgMediaListDamaged(String error) {
    return 'the media list is damaged ($error)';
  }

  @override
  String goApkgMediaSkipped(int count, String files) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString audio/video files were',
      one: '1 audio/video file was',
    );
    return '$_temp0 skipped, because Cardflip cannot play their format: $files';
  }

  @override
  String get goApkgMissingNoteType => 'a note uses a missing note type';

  @override
  String goApkgNewerVersion(String version) {
    return 'it was made by a newer Anki version (package format $version) that Cardflip cannot read yet';
  }

  @override
  String get goApkgNoCollection => 'it contains no Anki collection';

  @override
  String goApkgNotZip(String error) {
    return 'the file is damaged or not a zip archive ($error)';
  }

  @override
  String goApkgNoteTypesUnreadable(String error) {
    return 'its note types could not be read ($error)';
  }

  @override
  String get goApkgUnreadable => 'not a readable Anki package';

  @override
  String get goBackupDamaged => 'this backup is damaged';

  @override
  String goBackupFirstFailed(String error) {
    return 'the collection could not be backed up first, so nothing was changed: $error';
  }

  @override
  String get goBackupGone => 'this backup no longer exists';

  @override
  String goBackupIntervalRange(int max) {
    return 'The backup interval must be between 0 and $max minutes.';
  }

  @override
  String get goBackupNewerFormat =>
      'this backup was made by a newer version of Cardflip; update Cardflip to restore it';

  @override
  String get goBackupNotImportable =>
      'a backup cannot be imported; restore it from Settings instead';

  @override
  String goBackupsKeptRange(int max) {
    return 'The number of backups to keep must be between 0 and $max.';
  }

  @override
  String goBrowserColumnTwice(String column) {
    return 'The browser column \"$column\" is shown twice.';
  }

  @override
  String get goBrowserNeedsColumn => 'the browser needs at least one column';

  @override
  String get goBrowserSplitRange =>
      'The browser split must be between 0 and 1.';

  @override
  String goCardCountRange(int max) {
    return 'The number of cards must be between 1 and $max.';
  }

  @override
  String get goCardGone => 'That card no longer exists.';

  @override
  String get goCardTypeGone => 'That card type no longer exists.';

  @override
  String get goCardTypeNameEmpty => 'Please enter a card type name.';

  @override
  String get goCardTypeNameInvalid =>
      'A card type name cannot contain : { } \" or start with # ^ /.';

  @override
  String goCardTypeNameTaken(String name) {
    return 'The note type already has a card type named \"$name\".';
  }

  @override
  String goCardTypeOneTarget(String name) {
    return 'Card type \"$name\" can only go to one card type.';
  }

  @override
  String goChangeLeavesNotesEmpty(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'This change would leave $_temp0 without cards, so it was not made.';
  }

  @override
  String goCheckCardsWithoutNote(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards without a note were',
      one: '1 card without a note was',
    );
    return '$_temp0 removed.';
  }

  @override
  String goCheckFieldCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes with the wrong number of fields were',
      one: '1 note with the wrong number of fields was',
    );
    return '$_temp0 fixed.';
  }

  @override
  String get goCheckFieldCountTrimmed =>
      'The removed field content is in the deletion log.';

  @override
  String goCheckInvalidCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards for card types that no longer exist were',
      one: '1 card for a card type that no longer exists was',
    );
    return '$_temp0 removed.';
  }

  @override
  String goCheckMissingCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString missing cards were',
      one: '1 missing card was',
    );
    return '$_temp0 generated.';
  }

  @override
  String goCheckMissingDeck(int count, String deck) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards in missing decks were',
      one: '1 card in a missing deck was',
    );
    return '$_temp0 moved to $deck.';
  }

  @override
  String goCheckNotesWithoutType(int count, int cards) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes whose note type no longer exists were',
      one: '1 note whose note type no longer exists was',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString cards',
      one: '1 card',
    );
    return '$_temp0 removed, with $_temp1. The notes are in the deletion log.';
  }

  @override
  String goCheckScheduling(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards with invalid scheduling were',
      one: '1 card with invalid scheduling was',
    );
    return '$_temp0 repaired.';
  }

  @override
  String goCheckSchedulingFromHistory(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Memory states of $_temp0 were recomputed from the review history.';
  }

  @override
  String goCheckTags(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes with duplicate or invalid tags were',
      one: '1 note with duplicate or invalid tags was',
    );
    return '$_temp0 fixed.';
  }

  @override
  String goChooseCardTypeFor(String noteType) {
    return 'Choose a card type for each card type of \"$noteType\".';
  }

  @override
  String goChooseFieldFor(String noteType) {
    return 'Choose a field for each field of \"$noteType\".';
  }

  @override
  String get goClipboardImageFormat =>
      'the clipboard image is not a PNG, JPEG or GIF image';

  @override
  String get goClozeOneCardType =>
      'A cloze note type has exactly one card type.';

  @override
  String goCollectionUnreadable(String error) {
    return 'Your collection could not be read: $error';
  }

  @override
  String get goCsvChooseTypeAndDeck => 'choose a note type and a deck';

  @override
  String goCsvExtraFields(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString rows',
      one: '1 row',
    );
    return '$_temp0 had more fields than their note type; the extra fields were not imported.';
  }

  @override
  String goCsvFieldOutOfRange(int field, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'column mapped to field $field, but the note type has $countString fields';
  }

  @override
  String get goCsvMapAField => 'map at least one column to a field';

  @override
  String get goCsvNoRows => 'the file has no rows';

  @override
  String get goCsvNotUtf8 =>
      'the file is not UTF-8 text; save it as UTF-8 and try again';

  @override
  String goCsvNoteTypeMissing(String name) {
    return 'The file\'s note type \"$name\" does not exist; rows without a note type of their own use the one chosen here.';
  }

  @override
  String get goCsvOneDeckColumn => 'map only one column to the deck';

  @override
  String get goCsvOneGuidColumn => 'map only one column to the GUID';

  @override
  String get goCsvOneNoteTypeColumn => 'map only one column to the note type';

  @override
  String get goCsvOneTagsColumn => 'map only one column to the tags';

  @override
  String goCsvSkippedUnknownType(int count, String name) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString rows',
      one: '1 row',
    );
    return 'Skipped $_temp0: there is no note type named \"$name\".';
  }

  @override
  String goCustomStudyNameTaken(String name) {
    return 'A regular deck is named \"$name\". Rename it to use custom study.';
  }

  @override
  String goCustomStudyNeedsRegular(String deck) {
    return 'Custom study needs a regular deck; \"$deck\" is a filtered deck.';
  }

  @override
  String goDailyGoalRange(int max) {
    return 'The daily goal must be between 1 and $max.';
  }

  @override
  String goRestDaysRange(int max) {
    return 'The rest days per week must be between 0 and $max.';
  }

  @override
  String goDayCountRange(int max) {
    return 'The number of days must be between 1 and $max.';
  }

  @override
  String get goDeckGone => 'That deck no longer exists.';

  @override
  String get goDeckMoveIntoItself =>
      'A deck cannot be moved into itself or one of its subdecks.';

  @override
  String get goDeckNameEmpty => 'Please enter a deck name.';

  @override
  String goDeckNameEmptyPart(String name) {
    return 'Deck name \"$name\" has an empty part between \'::\'.';
  }

  @override
  String goDeckNameTaken(String name) {
    return 'A deck named \"$name\" already exists.';
  }

  @override
  String goDeckNameTakenByOther(String name) {
    return 'The name \"$name\" is already taken by another deck.';
  }

  @override
  String get goDefaultPresetUndeletable =>
      'The Default preset cannot be deleted.';

  @override
  String goDesiredRetentionRange(double min, double max) {
    final intl.NumberFormat minNumberFormat =
        intl.NumberFormat.decimalPatternDigits(
          locale: localeName,
          decimalDigits: 2,
        );
    final String minString = minNumberFormat.format(min);
    final intl.NumberFormat maxNumberFormat =
        intl.NumberFormat.decimalPatternDigits(
          locale: localeName,
          decimalDigits: 2,
        );
    final String maxString = maxNumberFormat.format(max);

    return 'Desired retention must be between $minString and $maxString.';
  }

  @override
  String get goDueDateHelp =>
      'Enter days from today: 0 for today, 1 for tomorrow, 60-90 for a random day in that range, with ! to set the interval too.';

  @override
  String goDueDateTooFar(int max) {
    return 'A due date can be at most $max days away.';
  }

  @override
  String goDueRangeReversed(String from, String to) {
    return 'The range $from-$to ends before it starts; put the smaller number first.';
  }

  @override
  String get goFieldGone => 'That field no longer exists.';

  @override
  String get goFieldNameEmpty => 'Please enter a field name.';

  @override
  String get goFieldNameInvalid =>
      'A field name cannot contain : { } \" or start with # ^ /.';

  @override
  String goFieldNameTaken(String name) {
    return 'The note type already has a field named \"$name\".';
  }

  @override
  String get goFileGone =>
      'the file is no longer available; please choose it again';

  @override
  String goFileUnreadable(String error) {
    return 'the file could not be read: $error';
  }

  @override
  String get goFilteredFilterCount => 'A filtered deck has one or two filters.';

  @override
  String get goFilteredNoMatch =>
      'No cards matched the search. Suspended and buried cards and cards in other filtered decks are left out.';

  @override
  String goFilteredNoNotes(String deck) {
    return '\"$deck\" is a filtered deck, so notes cannot be added to it.';
  }

  @override
  String goFilteredNoSubdecks(String deck) {
    return '\"$deck\" is a filtered deck, which cannot have subdecks.';
  }

  @override
  String get goFindTextEmpty => 'Enter the text to find.';

  @override
  String goFsrsParamRange(int index, String min, String max) {
    return 'FSRS parameter $index must be between $min and $max.';
  }

  @override
  String goFsrsParamsCount(int expected, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'FSRS parameters must be $expected numbers, not $countString.';
  }

  @override
  String goHistoricalRetentionRange(double min, double max) {
    final intl.NumberFormat minNumberFormat =
        intl.NumberFormat.decimalPatternDigits(
          locale: localeName,
          decimalDigits: 2,
        );
    final String minString = minNumberFormat.format(min);
    final intl.NumberFormat maxNumberFormat =
        intl.NumberFormat.decimalPatternDigits(
          locale: localeName,
          decimalDigits: 2,
        );
    final String maxString = maxNumberFormat.format(max);

    return 'Historical retention must be between $minString and $maxString.';
  }

  @override
  String goImportNotUpdatedNoCards(String note) {
    return 'Not updated, it would produce no cards: $note';
  }

  @override
  String goImportOtherNoteType(String note) {
    return 'A note already exists with a different note type and was not updated: $note';
  }

  @override
  String goImportPresetOptions(String preset, String error) {
    return 'The options of preset \"$preset\" could not be imported ($error); it has the default options.';
  }

  @override
  String goImportSkipped(String note, String error) {
    return 'Skipped \"$note\": $error';
  }

  @override
  String goImportSkippedNoCards(String note) {
    return 'Skipped, it would produce no cards: \"$note\"';
  }

  @override
  String goInvalidDate(String value) {
    return '\"$value\" is not a date like 2026-01-31.';
  }

  @override
  String goInvalidRating(String rating) {
    return 'Invalid rating $rating.';
  }

  @override
  String goInvalidRegex(String error) {
    return 'Invalid regular expression: $error';
  }

  @override
  String get goLastDeck =>
      'The last deck cannot be deleted. A collection always has at least one deck.';

  @override
  String get goLastNoteType =>
      'The last note type cannot be deleted. A collection always has at least one note type.';

  @override
  String goLastPositionTooHigh(int last, int max) {
    return 'The last position would be $last; positions must be at most $max.';
  }

  @override
  String goLearnAheadRange(int max) {
    return 'the learn ahead limit must be between 0 and $max minutes';
  }

  @override
  String get goLearningStepsRange =>
      'Learning steps must each be longer than 0 and at most 365 days.';

  @override
  String goLearningStepsTooMany(int max) {
    return 'Learning steps can have at most $max steps.';
  }

  @override
  String goLeechThresholdRange(int max) {
    return 'Leech threshold must be between 1 and $max lapses.';
  }

  @override
  String goMaxAnswerSecondsRange(int max) {
    return 'Maximum answer seconds must be between 1 and $max.';
  }

  @override
  String goMaxIntervalRange(int max) {
    return 'Maximum interval must be between 1 and $max days.';
  }

  @override
  String goMcpListenFailed(String error) {
    return 'The server could not start: $error';
  }

  @override
  String goMcpPortInUse(String port) {
    return 'Port $port is already in use.';
  }

  @override
  String goMcpPortRange(String min, String max) {
    return 'The port must be between $min and $max.';
  }

  @override
  String goMcpSettingsNotSaved(String error) {
    return 'The agent settings could not be saved: $error';
  }

  @override
  String goMcpSettingsUnreadable(String error) {
    return 'The agent settings could not be read: $error';
  }

  @override
  String goMediaEmptyTrashFailed(String error) {
    return 'The trash could not be emptied: $error';
  }

  @override
  String goMediaMissing(String file) {
    return 'Media file \"$file\" is missing.';
  }

  @override
  String goMediaRestoreFailed(String error) {
    return 'The files could not be restored: $error';
  }

  @override
  String goMediaSaveFailed(String error) {
    return 'The file could not be saved: $error';
  }

  @override
  String goMediaTrashFailed(String error) {
    return 'The unused files could not be moved to the trash: $error';
  }

  @override
  String goMinIntervalAboveMax(int min, int max) {
    return 'Minimum interval ($min days) must not be above the maximum interval ($max days).';
  }

  @override
  String goMinIntervalRange(int max) {
    return 'Minimum interval must be between 1 and $max days.';
  }

  @override
  String get goMoveIntoFiltered =>
      'Cards cannot be moved into a filtered deck; it gathers its cards with its search.';

  @override
  String get goNeedsCardType => 'A note type needs at least one card type.';

  @override
  String get goNeedsField => 'A note type needs at least one field.';

  @override
  String get goNewPerDayRange =>
      'New cards per day must be between 0 and 9999.';

  @override
  String get goNoCardsCloze =>
      'This would produce no cards. Add a cloze deletion such as {{c1::text}}.';

  @override
  String goNoCardsFillInField(String field) {
    return 'This would produce no cards. Fill in the $field field.';
  }

  @override
  String goNoCardsFillInFields(String fields, String last) {
    return 'This would produce no cards. Fill in the $fields or the $last field.';
  }

  @override
  String get goNoCardsNoField =>
      'This would produce no cards: no card\'s question uses a field.';

  @override
  String get goNoCardsSelected => 'No cards are selected.';

  @override
  String get goNoNotesSelected => 'No notes are selected.';

  @override
  String get goNoStudySession => 'no study session; start one first';

  @override
  String goNotAutomaticBackup(String name) {
    return '\"$name\" is not an automatic backup';
  }

  @override
  String get goNotBackup => 'this file is not a Cardflip backup';

  @override
  String goNotFilteredDeck(String deck) {
    return '\"$deck\" is not a filtered deck.';
  }

  @override
  String get goNotPackageOrBackup =>
      'this file is neither an Anki package nor a Cardflip backup';

  @override
  String get goNotTextFile => 'this file is not a text file';

  @override
  String goNoteFieldCount(int count, String noteType, int expected) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'The note has $countString fields, but note type \"$noteType\" has $expected.';
  }

  @override
  String get goNoteGone => 'That note no longer exists.';

  @override
  String get goNoteTypeGone => 'That note type no longer exists.';

  @override
  String get goNoteTypeNameEmpty => 'Please enter a note type name.';

  @override
  String goNoteTypeNameTaken(String name) {
    return 'A note type named \"$name\" already exists.';
  }

  @override
  String goNotesWouldHaveNoCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 would have no cards. Map the fields its card types use, or keep a card type.';
  }

  @override
  String get goNothingToRedo => 'There is nothing to redo.';

  @override
  String get goNothingToReplace =>
      'the collection is readable; there is nothing to replace';

  @override
  String get goNothingToUndo => 'There is nothing to undo.';

  @override
  String goOptimizeSearchInvalid(String error) {
    return 'The optimize search is invalid: $error';
  }

  @override
  String goPositionStepRange(int max) {
    return 'The step must be between 1 and $max.';
  }

  @override
  String get goPresetGone => 'That preset no longer exists.';

  @override
  String get goPresetNameEmpty => 'Please enter a preset name.';

  @override
  String goPresetNameTaken(String name) {
    return 'A preset named \"$name\" already exists.';
  }

  @override
  String goPreviewDelayRange(int max) {
    return 'Preview delays must be between 0 and $max seconds.';
  }

  @override
  String get goProfileGone => 'This profile no longer exists.';

  @override
  String get goProfileNameEmpty => 'A profile needs a name.';

  @override
  String goProfileNameTaken(String name) {
    return 'A profile named \"$name\" already exists.';
  }

  @override
  String get goProfileOnly => 'The only profile cannot be deleted.';

  @override
  String get goProfileOpen =>
      'The open profile cannot be deleted. Open another profile first.';

  @override
  String goProfilesNotSaved(String error) {
    return 'The list of profiles could not be saved: $error';
  }

  @override
  String goProfilesUnreadable(String error) {
    return 'The list of profiles could not be read: $error';
  }

  @override
  String goQuestionFadeRange(int max) {
    return 'The question fade must be between 0 and $max milliseconds.';
  }

  @override
  String get goQuestionTemplateEmpty =>
      'The question template cannot be empty.';

  @override
  String get goRelearningStepsRange =>
      'Relearning steps must each be longer than 0 and at most 365 days.';

  @override
  String goRelearningStepsTooMany(int max) {
    return 'Relearning steps can have at most $max steps.';
  }

  @override
  String goReservedFieldName(String name, String tag) {
    return '\"$name\" is a reserved name: templates use {{$tag}} for a special field. Please choose another name.';
  }

  @override
  String get goReviewsPerDayRange =>
      'Maximum reviews per day must be between 0 and 99999.';

  @override
  String get goRolloverHourRange =>
      'The day rollover hour must be between 0 and 23.';

  @override
  String get goSameNoteType =>
      'All selected cards must use the same note type.';

  @override
  String goSavedSearchMissing(String name) {
    return 'there is no saved search named \"$name\"';
  }

  @override
  String get goSavedSearchNameEmpty => 'A saved search needs a name.';

  @override
  String goSavedSearchNameTaken(String name) {
    return 'A saved search named \"$name\" already exists.';
  }

  @override
  String get goSearchCardTypeNumber =>
      'card type numbers start at 1, e.g. card:2';

  @override
  String get goSearchEmptyTerm => 'empty search term';

  @override
  String goSearchErrorAt(String message, String position) {
    return '$message (character $position)';
  }

  @override
  String goSearchErrorIn(String term, String message) {
    return '$term: $message';
  }

  @override
  String goSearchInvalidRegex(String error) {
    return 'invalid regular expression: $error';
  }

  @override
  String get goSearchNeedsCardType =>
      'needs a card type name or number, e.g. card:Reverse or card:2';

  @override
  String goSearchNeedsComparison(String property, String example) {
    return 'needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:$property$example';
  }

  @override
  String goSearchNeedsDayComparison(String property) {
    return 'needs a comparison with a number of days, e.g. prop:$property=-1 (0 is today, -1 yesterday)';
  }

  @override
  String goSearchNeedsDays(String key) {
    return 'needs a number of days, e.g. $key:7';
  }

  @override
  String goSearchNeedsDaysToday(String key) {
    return 'needs a number of days, e.g. $key:7 (1 is today)';
  }

  @override
  String goSearchNeedsIds(String key) {
    return 'needs an ID or a comma-separated list of IDs, e.g. $key:1712345678901,1712345678902';
  }

  @override
  String goSearchNeedsName(String key) {
    return 'needs a name, e.g. $key:Spanish';
  }

  @override
  String get goSearchNeedsRegex =>
      'needs a regular expression, e.g. re:^\\d+\$';

  @override
  String get goSearchNeedsWord => 'needs a word, e.g. w:dog';

  @override
  String get goSearchNoEase =>
      'Cardflip schedules with FSRS, which has no ease; search the difficulty with prop:d instead, e.g. prop:d>0.7';

  @override
  String goSearchOperatorAlone(String operator) {
    return '\"$operator\" needs a search term on both sides';
  }

  @override
  String get goSearchParenEmpty => 'empty parentheses';

  @override
  String get goSearchParenNotClosed => 'a parenthesis is not closed';

  @override
  String get goSearchParenNotOpened =>
      'a closing parenthesis has no opening one';

  @override
  String get goSearchRating =>
      'the rating must be 1 (Again), 2 (Hard), 3 (Good) or 4 (Easy), e.g. rated:7:1';

  @override
  String goSearchTooManyDays(int max) {
    return 'the number of days must be at most $max';
  }

  @override
  String get goSearchUnknownFlag =>
      'unknown flag; use flag:0 (no flag) to flag:7';

  @override
  String get goSearchUnknownProperty =>
      'unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30';

  @override
  String get goSearchUnknownState =>
      'unknown state; use is:new, is:due, is:review, is:learn, is:suspended, is:buried, is:buried-manually or is:buried-sibling';

  @override
  String goSecondsAnswerRange(int max) {
    return 'Seconds to show answer for must be between 0 and $max.';
  }

  @override
  String goSecondsQuestionRange(int max) {
    return 'Seconds to show question for must be between 0 and $max.';
  }

  @override
  String goSectionNotClosed(String tag) {
    return '$tag is not closed';
  }

  @override
  String goSectionNotOpened(String closing, String opening) {
    return '$closing has no matching $opening';
  }

  @override
  String goSectionsCrossed(String closing, String outer, String inner) {
    return '$closing closes $outer before $inner is closed';
  }

  @override
  String get goSelectCardsToChange =>
      'Select the cards of the notes to change.';

  @override
  String goSimulationDaysRange(int max) {
    return 'Days to simulate must be between 1 and $max.';
  }

  @override
  String goStartPositionRange(int max) {
    return 'The start position must be between 0 and $max.';
  }

  @override
  String get goStepAddField => 'Add field';

  @override
  String get goStepAddNote => 'Add note';

  @override
  String get goStepAddNoteType => 'Add note type';

  @override
  String goStepAddNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Add $_temp0';
  }

  @override
  String get goStepAddPreset => 'Add preset';

  @override
  String get goStepAddTags => 'Add tags';

  @override
  String get goStepAddTemplate => 'Add card type';

  @override
  String get goStepAnswer => 'Answer card';

  @override
  String get goStepBuildFiltered => 'Build filtered deck';

  @override
  String goStepBury(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Bury $_temp0';
  }

  @override
  String get goStepChange => 'Change';

  @override
  String get goStepChangeNoteType => 'Change note type';

  @override
  String get goStepCheckDatabase => 'Check database';

  @override
  String get goStepCloneNoteType => 'Clone note type';

  @override
  String get goStepClonePreset => 'Clone preset';

  @override
  String get goStepCreateDeck => 'Create deck';

  @override
  String get goStepCustomStudy => 'Custom study';

  @override
  String get goStepDeckDescription => 'Edit deck description';

  @override
  String get goStepDeckLimits => 'Change deck limits';

  @override
  String get goStepDeckOptions => 'Change deck options';

  @override
  String get goStepDeckPreset => 'Change preset';

  @override
  String get goStepDeleteDeck => 'Delete deck';

  @override
  String goStepDeleteEmptyCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString empty cards',
      one: '1 empty card',
    );
    return 'Delete $_temp0';
  }

  @override
  String get goStepDeleteNoteType => 'Delete note type';

  @override
  String goStepDeleteNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Delete $_temp0';
  }

  @override
  String get goStepDeletePreset => 'Delete preset';

  @override
  String get goStepDeleteTags => 'Delete tags';

  @override
  String get goStepEmptyFiltered => 'Empty filtered deck';

  @override
  String get goStepFieldOptions => 'Change field options';

  @override
  String get goStepFindReplace => 'Find and replace';

  @override
  String goStepFlag(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Flag $_temp0';
  }

  @override
  String get goStepImport => 'Import';

  @override
  String goStepMark(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Mark $_temp0';
  }

  @override
  String goStepMoveCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Move $_temp0';
  }

  @override
  String get goStepMoveField => 'Move field';

  @override
  String get goStepMoveTags => 'Move tags';

  @override
  String get goStepMoveTemplate => 'Move card type';

  @override
  String get goStepNoteTypeCSS => 'Edit styling';

  @override
  String get goStepOptimize => 'Optimize all presets';

  @override
  String get goStepPresetToSubdecks => 'Save to all subdecks';

  @override
  String get goStepRemoveField => 'Remove field';

  @override
  String get goStepRemoveTags => 'Remove tags';

  @override
  String get goStepRemoveTemplate => 'Remove card type';

  @override
  String get goStepRenameDeck => 'Rename deck';

  @override
  String get goStepRenameField => 'Rename field';

  @override
  String get goStepRenameNoteType => 'Rename note type';

  @override
  String get goStepRenamePreset => 'Rename preset';

  @override
  String get goStepRenameTag => 'Rename tag';

  @override
  String get goStepReposition => 'Reposition';

  @override
  String goStepResetCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Reset $_temp0';
  }

  @override
  String goStepSetDueDate(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Set due date of $_temp0';
  }

  @override
  String get goStepSortField => 'Change sort field';

  @override
  String goStepSuspend(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Suspend $_temp0';
  }

  @override
  String get goStepTodayLimits => 'Custom study';

  @override
  String get goStepUnbury => 'Unbury';

  @override
  String get goStepUndoAnswer => 'Undo answer';

  @override
  String goStepUnflag(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Remove flag from $_temp0';
  }

  @override
  String goStepUnmark(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Unmark $_temp0';
  }

  @override
  String goStepUnsuspend(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Unsuspend $_temp0';
  }

  @override
  String get goStepUpdateNote => 'Edit note';

  @override
  String get goStepUpdateTemplate => 'Edit card type';

  @override
  String goStorageOpenFailed(String error) {
    return 'The collection storage could not be opened: $error';
  }

  @override
  String goSyncAlreadyLinked(String profile, String device) {
    return 'Profile \"$profile\" is already linked with a profile on $device.';
  }

  @override
  String get goSyncAlreadyLinkedHere =>
      'This profile is already linked with a profile on that device.';

  @override
  String get goSyncBadAddress =>
      'Enter the address and port shown on the other device.';

  @override
  String goSyncBusy(String device) {
    return '$device is syncing with another device. Try again in a moment.';
  }

  @override
  String goSyncClocksDifferDays(String device, int days) {
    return 'The clocks of this device and $device differ by $days days. Set the correct time on both and try again.';
  }

  @override
  String goSyncClocksDifferHours(String device, int hours) {
    return 'The clocks of this device and $device differ by $hours hours. Set the correct time on both and try again.';
  }

  @override
  String goSyncClocksDifferMinutes(String device, int minutes) {
    return 'The clocks of this device and $device differ by $minutes minutes. Set the correct time on both and try again.';
  }

  @override
  String goSyncConnectFailed(String device, String error) {
    return '$device could not be reached: $error';
  }

  @override
  String goSyncConnectTimeout(String device, String address, String port) {
    return '$device did not answer at $address. If it is on, a firewall on it may block sync: allow incoming TCP connections on port $port there.';
  }

  @override
  String goSyncDeviceNameLength(int max) {
    return 'The device name must be 1 to $max characters long.';
  }

  @override
  String get goSyncLinkGone => 'That link no longer exists.';

  @override
  String goSyncLinksUnreadable(String error) {
    return 'The links of this profile could not be read: $error';
  }

  @override
  String goSyncListenFailed(String error) {
    return 'Sync could not start: $error';
  }

  @override
  String goSyncNeedsChoice(String device) {
    return 'Choose how to combine this profile with the one on $device.';
  }

  @override
  String get goSyncNoNetwork =>
      'This device is not on a network that can be trusted.';

  @override
  String get goSyncNotPaired => 'That device is not paired.';

  @override
  String goSyncNotReachable(String device) {
    return '$device is not reachable on this network.';
  }

  @override
  String get goSyncNotSetUp => 'Open Settings → Sync first.';

  @override
  String goSyncPairConnectFailed(String address, String error) {
    return 'Could not reach the device at $address: $error';
  }

  @override
  String goSyncPairDeclined(String device) {
    return '$device declined the pairing request.';
  }

  @override
  String goSyncPairFailed(String device) {
    return 'Pairing with $device failed: the codes were not confirmed on both devices.';
  }

  @override
  String goSyncPairNotVisible(String device) {
    return '$device is not visible for pairing. Open “Pair a device” on it and try again.';
  }

  @override
  String goSyncPairTimeout(String address, String port) {
    return 'The device at $address did not answer. A firewall on it may block sync: allow incoming TCP connections on port $port there, or start the pairing on that device.';
  }

  @override
  String goSyncStateNotSaved(String error) {
    return 'The sync settings could not be saved: $error';
  }

  @override
  String goSyncStateUnreadable(String error) {
    return 'The sync settings could not be read: $error';
  }

  @override
  String goSyncUnknownLink(String device) {
    return '$device no longer has this link. Link the profiles again.';
  }

  @override
  String get goSyncUntrusted => 'Trust this network first.';

  @override
  String goSyncUpdateOther(String device) {
    return 'Update Cardflip on $device to sync with it.';
  }

  @override
  String goSyncUpdateThis(String device) {
    return 'Update Cardflip on this device to sync with $device.';
  }

  @override
  String get goTagMoveIntoItself =>
      'A tag cannot be moved into itself or one of its child tags.';

  @override
  String get goTagNameEmpty => 'Enter a tag name.';

  @override
  String get goTagNameSpaces => 'Tag names cannot contain spaces.';

  @override
  String goTagNoFieldName(String tag) {
    return '$tag has no field name';
  }

  @override
  String goTagNotClosed(String tag) {
    return '$tag has no closing }}';
  }

  @override
  String goTemplateBackMistake(String problem) {
    return 'In the back template, $problem.';
  }

  @override
  String goTemplateBackUnknownField(String tag, String name) {
    return 'The back template uses $tag, but there is no field \"$name\".';
  }

  @override
  String goTemplateFrontMistake(String problem) {
    return 'In the front template, $problem.';
  }

  @override
  String goTemplateFrontUnknownField(String tag, String name) {
    return 'The front template uses $tag, but there is no field \"$name\".';
  }

  @override
  String goTemplateNeedsCloze(String tag) {
    return 'The front template must use $tag or another cloze field.';
  }

  @override
  String get goTemplateNoField =>
      'The front template uses no field, so it would never produce a card.';

  @override
  String goTimeboxRange(int max) {
    return 'The timebox time limit must be between 0 and $max minutes.';
  }

  @override
  String goTodayIncreaseRange(int max) {
    return 'The increase must be between 0 and $max cards.';
  }

  @override
  String goUiScaleRange(int min, int max, int step) {
    return 'The UI size must be between $min and $max % in steps of $step %.';
  }

  @override
  String get goUndoOnlyCardflip =>
      'Only answers given in Cardflip can be undone.';

  @override
  String get goUndoOnlyLatest =>
      'Only the most recent answer to a card can be undone.';

  @override
  String goUnknownAnswerAction(String value) {
    return 'Unknown answer action \"$value\".';
  }

  @override
  String goUnknownBrowserColumn(String value) {
    return 'Unknown browser column \"$value\".';
  }

  @override
  String goUnknownBrowserMode(String value) {
    return 'Unknown browser mode \"$value\".';
  }

  @override
  String goUnknownCardState(String value) {
    return 'Unknown card state \"$value\".';
  }

  @override
  String goUnknownCustomStudy(String value) {
    return 'Unknown custom study choice \"$value\".';
  }

  @override
  String goUnknownDailyGoal(String value) {
    return 'Unknown daily goal \"$value\".';
  }

  @override
  String goUnknownEasyDaysLevel(String value) {
    return 'Unknown Easy Days level $value.';
  }

  @override
  String goUnknownFilteredOrder(String value) {
    return 'Unknown filtered deck order \"$value\".';
  }

  @override
  String goUnknownFlag(String flag) {
    return 'There is no flag $flag.';
  }

  @override
  String goUnknownGatherOrder(String value) {
    return 'Unknown new card gather order \"$value\".';
  }

  @override
  String goUnknownInsertionOrder(String value) {
    return 'Unknown insertion order \"$value\".';
  }

  @override
  String goUnknownInterdayOrder(String value) {
    return 'Unknown interday learning/review order \"$value\".';
  }

  @override
  String goUnknownLanguage(String value) {
    return 'Unknown language \"$value\".';
  }

  @override
  String goUnknownLeechAction(String value) {
    return 'Unknown leech action \"$value\".';
  }

  @override
  String goUnknownNewReviewOrder(String value) {
    return 'Unknown new/review order \"$value\".';
  }

  @override
  String goUnknownNewSortOrder(String value) {
    return 'Unknown new card sort order \"$value\".';
  }

  @override
  String goUnknownNoteSort(String value) {
    return 'Notes can\'t be sorted by \"$value\".';
  }

  @override
  String goUnknownQuestionAction(String value) {
    return 'Unknown question action \"$value\".';
  }

  @override
  String goUnknownReduceMotion(String value) {
    return 'Unknown reduce motion choice \"$value\".';
  }

  @override
  String goUnknownReviewSortOrder(String value) {
    return 'Unknown review sort order \"$value\".';
  }

  @override
  String goUnknownSeparator(String value) {
    return 'unknown separator \"$value\"';
  }

  @override
  String goUnknownStockType(String value) {
    return 'Unknown stock note type \"$value\".';
  }

  @override
  String goUnknownTheme(String value) {
    return 'Unknown theme \"$value\".';
  }

  @override
  String get datePattern => 'yyyy-MM-dd';

  @override
  String dueAt(String time) {
    return 'at $time';
  }

  @override
  String get dueTomorrow => 'tomorrow';

  @override
  String dueInDays(int days) {
    final intl.NumberFormat daysNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String daysString = daysNumberFormat.format(days);

    String _temp0 = intl.Intl.pluralLogic(
      days,
      locale: localeName,
      other: '$daysString days',
      one: '1 day',
    );
    return 'in $_temp0';
  }

  @override
  String dueOn(String date) {
    return 'on $date';
  }

  @override
  String spentSeconds(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString seconds',
      one: '1 second',
    );
    return '$_temp0';
  }

  @override
  String spentMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString minutes',
      one: '1 minute',
    );
    return '$_temp0';
  }

  @override
  String spentHours(int hours, int minutes) {
    final intl.NumberFormat hoursNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String hoursString = hoursNumberFormat.format(hours);
    final intl.NumberFormat minutesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String minutesString = minutesNumberFormat.format(minutes);

    return '$hoursString h $minutesString min';
  }

  @override
  String intervalMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString minutes',
      one: '1 minute',
    );
    return '$_temp0';
  }

  @override
  String intervalHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString hours',
      one: '1 hour',
    );
    return '$_temp0';
  }

  @override
  String intervalDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return '$_temp0';
  }

  @override
  String intervalMonths(String months) {
    return '$months months';
  }

  @override
  String intervalYears(String years) {
    return '$years years';
  }

  @override
  String intervalLabelMinutes(int minutes) {
    final intl.NumberFormat minutesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String minutesString = minutesNumberFormat.format(minutes);

    return '<${minutesString}m';
  }

  @override
  String intervalLabelHours(String hours) {
    return '${hours}h';
  }

  @override
  String intervalLabelDays(int days) {
    final intl.NumberFormat daysNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String daysString = daysNumberFormat.format(days);

    return '${daysString}d';
  }

  @override
  String intervalLabelMonths(String months) {
    return '${months}mo';
  }

  @override
  String intervalLabelYears(String years) {
    return '${years}y';
  }

  @override
  String sizeBytes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString bytes',
      one: '1 byte',
    );
    return '$_temp0';
  }

  @override
  String get cancel => 'Cancel';

  @override
  String get save => 'Save';

  @override
  String get typeTheAnswer => 'Type the answer';

  @override
  String get missingImage => 'Missing image';

  @override
  String missingImageNamed(String alt) {
    return 'Missing image: $alt';
  }

  @override
  String get stop => 'Stop';

  @override
  String playFile(String file) {
    return 'Play $file';
  }

  @override
  String get speak => 'Speak';

  @override
  String get noVoicesInstalled => 'No voices are installed on this device.';

  @override
  String fileMissing(String file) {
    return '$file is missing';
  }

  @override
  String fileCannotPlay(String file) {
    return '$file cannot be played on this device';
  }

  @override
  String noVoiceForLanguage(String language) {
    return 'No voice is installed for $language';
  }

  @override
  String get noVoiceForThisLanguage =>
      'No voice is installed for this language';

  @override
  String get microphoneRefused =>
      'Cardflip may not use the microphone. Allow microphone access in your system settings to record your voice.';

  @override
  String get studyDeck => 'Study deck';

  @override
  String get searchDecks => 'Search decks';

  @override
  String get noMatchingDecks => 'No matching decks';

  @override
  String unburiedDeck(String deck) {
    return 'Unburied “$deck”';
  }

  @override
  String get editDescription => 'Edit description';

  @override
  String get tryAgain => 'Try again';

  @override
  String countBuriedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString buried cards',
      one: '1 buried card',
    );
    return '$_temp0';
  }

  @override
  String get unbury => 'Unbury';

  @override
  String get studyNow => 'Study now';

  @override
  String get options => 'Options';

  @override
  String get rebuild => 'Rebuild';

  @override
  String get emptyDeck => 'Empty';

  @override
  String get countNew => 'New';

  @override
  String get countLearning => 'Learning';

  @override
  String get countToReview => 'To review';

  @override
  String get nothingDueNow => 'Nothing is due right now.';

  @override
  String nextCardDue(String when) {
    return 'The next card is due $when.';
  }

  @override
  String get deckHasNoCards =>
      'This deck has no cards to study. Add cards or import some to get started.';

  @override
  String get description => 'Description';

  @override
  String get descriptionHelper =>
      'Plain text, or HTML such as <b>bold</b> and links';

  @override
  String enterWholeNumber(int max) {
    return 'Enter a whole number from 1 to $max.';
  }

  @override
  String moreAvailable(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString more available';
  }

  @override
  String get increaseBy => 'Increase by';

  @override
  String get unitCards => 'cards';

  @override
  String get unitDays => 'days';

  @override
  String get increaseHelper =>
      'For today only. The deck options stay as they are.';

  @override
  String get cards => 'Cards';

  @override
  String get newCardsOnly => 'New cards only';

  @override
  String get dueCardsOnly => 'Due cards only';

  @override
  String get allCardsRandom => 'All cards in random order (no rescheduling)';

  @override
  String get onlyWithTags => 'Only with these tags (optional)';

  @override
  String get onlyWithTagsHelper =>
      'Cards with any of them; separate tags by spaces';

  @override
  String get select => 'Select';

  @override
  String get next => 'Next';

  @override
  String get last => 'Last';

  @override
  String customStudySessionHelper(String deck) {
    return 'Studied in the filtered deck “$deck”, which replaces the previous session.';
  }

  @override
  String customStudyTitle(String deck) {
    return 'Custom study: $deck';
  }

  @override
  String get increaseNewLimit => 'Increase today\'s new card limit';

  @override
  String get increaseReviewLimit => 'Increase today\'s review card limit';

  @override
  String get reviewForgotten => 'Review forgotten cards';

  @override
  String get reviewAhead => 'Review ahead';

  @override
  String get previewNewCards => 'Preview new cards';

  @override
  String get studyByStateOrTag => 'Study by card state or tag';

  @override
  String get ok => 'OK';

  @override
  String get newDeck => 'New deck';

  @override
  String get name => 'Name';

  @override
  String get newDeckHelper => 'Use :: for subdecks, e.g. Languages::Spanish';

  @override
  String get create => 'Create';

  @override
  String get decks => 'Decks';

  @override
  String get statistics => 'Statistics';

  @override
  String get createDeck => 'Create deck';

  @override
  String get createFilteredDeck => 'Create filtered deck';

  @override
  String get collectionEmpty => 'Your collection is empty';

  @override
  String get collectionEmptyHint =>
      'Add your first card, or import cards from a text file (.csv, .tsv or .txt), an Anki deck (.apkg) or a whole Anki collection (.colpkg).';

  @override
  String get addACard => 'Add a card';

  @override
  String get importAFile => 'Import a file';

  @override
  String get renameDeck => 'Rename deck';

  @override
  String get renameDeckHelper => 'Use :: to move it into another deck';

  @override
  String get rename => 'Rename';

  @override
  String deleteNamedTitle(String name) {
    return 'Delete “$name”?';
  }

  @override
  String deleteDeckBody(String cards, String inFiltered, String subdecks) {
    return 'This deletes $cards$inFiltered$subdecks. You can undo this until you close Cardflip.';
  }

  @override
  String deleteDeckInFiltered(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return ' ($countString of them now in filtered decks)';
  }

  @override
  String deleteDeckSubdecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString subdecks',
      one: '1 subdeck',
    );
    return ' and $_temp0';
  }

  @override
  String countCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0';
  }

  @override
  String get delete => 'Delete';

  @override
  String deletedNamed(String name) {
    return 'Deleted “$name”';
  }

  @override
  String deletedDeckReturned(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Deleted “$deck”; $_temp0 returned to their home decks';
  }

  @override
  String get expand => 'Expand';

  @override
  String get collapse => 'Collapse';

  @override
  String get filteredDeck => 'Filtered deck';

  @override
  String get deckActions => 'Deck actions';

  @override
  String get study => 'Study';

  @override
  String get addCard => 'Add card';

  @override
  String get customStudy => 'Custom study';

  @override
  String get export => 'Export';

  @override
  String dueNewTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString new';
  }

  @override
  String dueLearningTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString learning';
  }

  @override
  String dueReviewTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString to review';
  }

  @override
  String get orderOldestSeen => 'Oldest seen first';

  @override
  String get orderRandom => 'Random';

  @override
  String get orderIntervalsAscending => 'Increasing intervals';

  @override
  String get orderIntervalsDescending => 'Decreasing intervals';

  @override
  String get orderMostLapses => 'Most lapses';

  @override
  String get orderAdded => 'Order added';

  @override
  String get orderDue => 'Order due';

  @override
  String get orderLatestAdded => 'Latest added first';

  @override
  String get orderRetrievability => 'Lowest retrievability first';

  @override
  String filteredDeckHolds(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '“$deck” holds $_temp0';
  }

  @override
  String rebuiltDeck(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Rebuilt “$deck”: $_temp0';
  }

  @override
  String emptiedDeck(String deck) {
    return 'Emptied “$deck”';
  }

  @override
  String filterLimitError(String which) {
    String _temp0 = intl.Intl.selectLogic(which, {
      'first': 'first',
      'other': 'second',
    });
    return 'Enter a number of cards from 1 to 99999 for the $_temp0 filter.';
  }

  @override
  String get previewDelaysError =>
      'Enter preview delays from 0 to 86400 seconds.';

  @override
  String get search => 'Search';

  @override
  String get filterSearchHelper =>
      'As in the browser, e.g. deck:Spanish is:due';

  @override
  String get limitTo => 'Limit to';

  @override
  String get cardsSelectedBy => 'Cards selected by';

  @override
  String get build => 'Build';

  @override
  String get newFilteredDeck => 'New filtered deck';

  @override
  String get filteredDeckOptions => 'Filtered deck options';

  @override
  String get filter => 'Filter';

  @override
  String get secondFilter => 'Second filter';

  @override
  String get secondFilterHelper =>
      'Adds the cards of another search after the first';

  @override
  String get reschedule => 'Reschedule cards based on my answers in this deck';

  @override
  String get rescheduleOn => 'Answers count as normal reviews';

  @override
  String get rescheduleOff =>
      'Preview: answers do not change when cards are due';

  @override
  String get previewDelays =>
      'Show the card again after (seconds; 0 returns it to its home deck)';

  @override
  String get again => 'Again';

  @override
  String get hard => 'Hard';

  @override
  String get good => 'Good';

  @override
  String get easy => 'Easy';

  @override
  String get unitSecondsShort => 's';

  @override
  String get filteredDeckNote =>
      'Suspended and buried cards and cards in other filtered decks are left out. Emptying or deleting the deck returns its cards to their home decks.';

  @override
  String get simDaysError => 'Enter between 1 and 3650 days to simulate.';

  @override
  String get simLimitsError => 'Enter whole numbers for the daily limits.';

  @override
  String useRetentionTitle(String retention) {
    return 'Use $retention desired retention?';
  }

  @override
  String get useRetentionBody =>
      'The deck options get this desired retention. It is saved when you save the options.';

  @override
  String get use => 'Use';

  @override
  String simulateDeck(String deck) {
    return 'Simulate $deck';
  }

  @override
  String get simulatorIntro =>
      'Expected reviews and memorised cards of this deck and its subdecks, starting from their cards today. Answer times and ratings come from your review history. Nothing is changed.';

  @override
  String get daysToSimulate => 'Days to simulate';

  @override
  String get newCardsPerDayShort => 'New cards/day';

  @override
  String get maxReviewsPerDayShort => 'Maximum reviews/day';

  @override
  String desiredRetentionValue(String retention) {
    return 'Desired retention: $retention';
  }

  @override
  String get compareRetention => 'Compare with another retention';

  @override
  String comparedRetentionValue(String retention) {
    return 'Compared retention: $retention';
  }

  @override
  String get simulate => 'Simulate';

  @override
  String get helpMeDecide => 'Help me decide';

  @override
  String suggestedRetention(String retention) {
    return 'Suggested desired retention: $retention';
  }

  @override
  String get suggestedRetentionHelp =>
      'The retention with the least review time per memorised card at the end of the simulation. Lower costs less time but you remember less.';

  @override
  String get secondsPerMemorisedCard => 'Seconds per memorised card';

  @override
  String useValue(String value) {
    return 'Use $value';
  }

  @override
  String dayNumber(int day) {
    return 'Day $day';
  }

  @override
  String get reviewsPerDay => 'Reviews per day';

  @override
  String get memorisedCards => 'Memorised cards';

  @override
  String atRetention(String retention) {
    return 'At $retention';
  }

  @override
  String simReviewTotals(int reviews, String perDay, int newCards) {
    final intl.NumberFormat reviewsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String reviewsString = reviewsNumberFormat.format(reviews);
    final intl.NumberFormat newCardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String newCardsString = newCardsNumberFormat.format(newCards);

    String _temp0 = intl.Intl.pluralLogic(
      reviews,
      locale: localeName,
      other: '$reviewsString reviews',
      one: '1 review',
    );
    String _temp1 = intl.Intl.pluralLogic(
      newCards,
      locale: localeName,
      other: '$newCardsString new cards',
      one: '1 new card',
    );
    return '$_temp0 ($perDay a day), $_temp1 learned';
  }

  @override
  String simTime(String total, String perDay) {
    return 'Time: $total ($perDay a day)';
  }

  @override
  String simMemorisedAtEnd(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Memorised at the end: $_temp0';
  }

  @override
  String get readingDeck => 'Reading the deck…';

  @override
  String simulatingPercent(String percent) {
    return 'Simulating… $percent';
  }

  @override
  String get simulating => 'Simulating…';

  @override
  String get enterDaysError => 'Enter a whole number of days from 1 to 36500.';

  @override
  String get notAboveMaxInterval => 'Must not be above the maximum interval.';

  @override
  String enterSecondsError(int min, int max) {
    return 'Enter a whole number of seconds from $min to $max.';
  }

  @override
  String get leechThresholdError =>
      'Enter a whole number for the leech threshold.';

  @override
  String get correctSteps => 'Correct the steps first.';

  @override
  String get correctIntervals => 'Correct the intervals first.';

  @override
  String get correctMaxAnswer => 'Correct the maximum answer seconds first.';

  @override
  String get correctAutoAdvance => 'Correct the auto advance seconds first.';

  @override
  String get correctFsrs => 'Correct the FSRS parameters first.';

  @override
  String get deckLimitsError =>
      'Enter whole numbers for the limits of this deck.';

  @override
  String get rescheduleTitle => 'Reschedule cards?';

  @override
  String rescheduleBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other:
          '$countString review cards will get new due dates from their review history',
      one: '1 review card will get a new due date from its review history',
    );
    return '$_temp0, with the new desired retention and FSRS parameters.';
  }

  @override
  String get rescheduleAction => 'Reschedule';

  @override
  String get optimizeAllTitle => 'Optimize all presets?';

  @override
  String get optimizeAllBody =>
      'Each preset gets FSRS parameters fitted to its reviews, where they fit better than its current ones.';

  @override
  String optimizeAllReschedule(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString review cards',
      one: '1 review card',
    );
    return ' Up to $_temp0 of presets that reschedule cards on change will get new due dates.';
  }

  @override
  String get optimize => 'Optimize';

  @override
  String optimizeAllResult(int applied) {
    final intl.NumberFormat appliedNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String appliedString = appliedNumberFormat.format(applied);

    String _temp0 = intl.Intl.pluralLogic(
      applied,
      locale: localeName,
      other: '$appliedString presets',
      one: '1 preset',
      zero: 'No preset',
    );
    return '$_temp0 got new parameters.';
  }

  @override
  String optimizeAllFew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString presets have',
      one: '1 preset has',
    );
    return ' $_temp0 too few reviews to optimize.';
  }

  @override
  String optimizePreset(int number, int total) {
    return 'Preset $number of $total: ';
  }

  @override
  String get optimizeCancelled => 'Optimizing was cancelled. Nothing changed.';

  @override
  String get ignoreReviewedBefore => 'Ignore cards reviewed before';

  @override
  String get discardChangesTitle => 'Discard changes?';

  @override
  String presetNotSaved(String preset) {
    return 'Your changes to the preset \"$preset\" are not saved.';
  }

  @override
  String get discard => 'Discard';

  @override
  String get addPreset => 'Add preset';

  @override
  String get addPresetHelper => 'Starts with the default options';

  @override
  String get add => 'Add';

  @override
  String get clonePreset => 'Clone preset';

  @override
  String presetCopyName(String name) {
    return '$name copy';
  }

  @override
  String get clonePresetHelper => 'Starts with the options shown';

  @override
  String get clone => 'Clone';

  @override
  String get renamePreset => 'Rename preset';

  @override
  String get deletePresetTitle => 'Delete preset?';

  @override
  String presetUnused(String preset) {
    return 'The preset \"$preset\" is not used by any deck.';
  }

  @override
  String presetUsedBy(String preset, int count, String fallback) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString decks. They use',
      one: '1 deck. It uses',
    );
    return 'The preset \"$preset\" is used by $_temp0 $fallback instead.';
  }

  @override
  String allSubdecksUsePreset(String preset) {
    return 'All subdecks use the preset \"$preset\" now.';
  }

  @override
  String optionsFor(String deck) {
    return 'Options for $deck';
  }

  @override
  String get preset => 'Preset';

  @override
  String get presetActions => 'Preset actions';

  @override
  String get deletePreset => 'Delete preset';

  @override
  String get saveToAllSubdecks => 'Save to all subdecks';

  @override
  String usedByDecksShared(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString decks',
      one: '1 deck',
    );
    return 'Used by $_temp0. Changes apply to all of them.';
  }

  @override
  String usedByDecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString decks',
      one: '1 deck',
    );
    return 'Used by $_temp0.';
  }

  @override
  String get dailyLimits => 'Daily limits';

  @override
  String get newCardsPerDay => 'New cards per day';

  @override
  String get maxReviewsPerDay => 'Maximum reviews per day';

  @override
  String get thisDeckOnly => 'This deck only';

  @override
  String get thisDeckOnlyHelp =>
      'Limits for this deck that replace the preset\'s. Leave empty to use the preset\'s.';

  @override
  String get enterWholeNumberShort => 'Enter a whole number';

  @override
  String get newIgnoresReviewLimit => 'New cards ignore review limit';

  @override
  String get newIgnoresReviewLimitHelp =>
      'When off, new cards are only shown while the review limit has room left.';

  @override
  String get limitsStartFromTop => 'Limits start from top';

  @override
  String get limitsStartFromTopHelp =>
      'When studying this deck, the limits of its parent decks apply too.';

  @override
  String get learningSteps => 'Learning steps';

  @override
  String get learningStepsHelp =>
      'Delays before a new card graduates, e.g. 1m 10m 1d (s, m, h, d). Leave empty to let FSRS decide.';

  @override
  String get relearningSteps => 'Relearning steps';

  @override
  String get relearningStepsHelp =>
      'Delays after a review card is forgotten, e.g. 10m. Leave empty to let FSRS decide.';

  @override
  String get insertionOrder => 'Insertion order';

  @override
  String get insertionOrderHelp =>
      'Where added cards go among new cards. Changing it reorders the deck\'s new cards.';

  @override
  String get insertSequential => 'Sequential (oldest cards first)';

  @override
  String get displayOrder => 'Display order';

  @override
  String get newGatherOrder => 'New card gather order';

  @override
  String get newGatherOrderHelp => 'Which new cards fill today\'s limit.';

  @override
  String get gatherDeck => 'Deck';

  @override
  String get gatherDeckRandomNotes => 'Deck, then random notes';

  @override
  String get gatherAscending => 'Ascending position';

  @override
  String get gatherDescending => 'Descending position';

  @override
  String get gatherRandomNotes => 'Random notes';

  @override
  String get gatherRandomCards => 'Random cards';

  @override
  String get newSortOrder => 'New card sort order';

  @override
  String get newSortOrderHelp =>
      'The order in which today\'s new cards are shown.';

  @override
  String get sortGathered => 'Order gathered';

  @override
  String get sortTypeThenGathered => 'Card type, then order gathered';

  @override
  String get sortTypeThenRandom => 'Card type, then random';

  @override
  String get sortRandomNoteThenType => 'Random note, then card type';

  @override
  String get newReviewOrder => 'New/review order';

  @override
  String get mixWithReviews => 'Mix with reviews';

  @override
  String get showBeforeReviews => 'Show before reviews';

  @override
  String get showAfterReviews => 'Show after reviews';

  @override
  String get interdayOrder => 'Interday learning/review order';

  @override
  String get interdayOrderHelp =>
      'Where cards with a step of a day or more are shown.';

  @override
  String get reviewSortOrder => 'Review sort order';

  @override
  String get reviewSortOrderHelp =>
      'Also decides which reviews fill the daily limit.';

  @override
  String get reviewDueRandom => 'Due date, then random';

  @override
  String get reviewDueDeck => 'Due date, then deck';

  @override
  String get reviewDeckDue => 'Deck, then due date';

  @override
  String get reviewIntervalsAscending => 'Ascending intervals';

  @override
  String get reviewIntervalsDescending => 'Descending intervals';

  @override
  String get reviewDifficultyAscending => 'Ascending difficulty';

  @override
  String get reviewDifficultyDescending => 'Descending difficulty';

  @override
  String get reviewRetrievabilityAscending => 'Ascending retrievability';

  @override
  String get reviewRelativeOverdueness => 'Relative overdueness';

  @override
  String get burying => 'Burying';

  @override
  String get buryingHelp =>
      'Once a card of a note is studied, its other cards of these kinds wait until the next day.';

  @override
  String get buryNew => 'Bury new siblings';

  @override
  String get buryReview => 'Bury review siblings';

  @override
  String get buryInterday => 'Bury interday learning siblings';

  @override
  String get maxIntervalDays => 'Maximum interval (days)';

  @override
  String get maxIntervalHelp =>
      'No review is scheduled further away than this.';

  @override
  String get minIntervalDays => 'Minimum interval (days)';

  @override
  String get minIntervalHelp =>
      'The shortest interval a forgotten card gets once it is relearned.';

  @override
  String get leechThreshold => 'Leech threshold';

  @override
  String get leechThresholdHelp =>
      'Lapses after which a card is a leech (1 to 99). It is flagged again every half that many lapses.';

  @override
  String get leechAction => 'Leech action';

  @override
  String get leechTagOnly => 'Tag only';

  @override
  String get leechSuspend => 'Suspend card and tag';

  @override
  String get retentionHelp =>
      'Higher retention means you forget less but review more often.';

  @override
  String get disableAutoplay => 'Don\'t play audio automatically';

  @override
  String get skipQuestionOnReplay => 'Skip question when replaying answer';

  @override
  String get timer => 'Timer';

  @override
  String get maxAnswerSeconds => 'Maximum answer seconds';

  @override
  String get maxAnswerSecondsHelp =>
      'Longer answers are recorded as this many seconds, so a break does not count as answer time.';

  @override
  String get showTimer => 'Show answer timer';

  @override
  String get showTimerHelp =>
      'Shows the time spent on the card while studying. It does not change what is recorded.';

  @override
  String get stopTimerOnAnswer => 'Stop timer on answer';

  @override
  String get stopTimerOnAnswerHelp =>
      'Stops the timer when the answer is shown.';

  @override
  String get autoAdvance => 'Auto advance';

  @override
  String get autoAdvanceHelp =>
      'Turn auto advance on while studying from the menu or with Shift+A. 0 seconds turns it off for that side.';

  @override
  String get secondsToShowQuestion => 'Seconds to show question for';

  @override
  String get secondsToShowAnswer => 'Seconds to show answer for';

  @override
  String get questionAction => 'Question action';

  @override
  String get showAnswer => 'Show answer';

  @override
  String get showReminder => 'Show reminder';

  @override
  String get answerAction => 'Answer action';

  @override
  String get buryCard => 'Bury card';

  @override
  String get answerAgain => 'Answer Again';

  @override
  String get answerHard => 'Answer Hard';

  @override
  String get answerGood => 'Answer Good';

  @override
  String get fsrs => 'FSRS';

  @override
  String get fsrsParameters => 'FSRS parameters';

  @override
  String get fsrsParametersHelp =>
      'Fitted to your reviews by Optimize. Leave empty for the default parameters.';

  @override
  String get optimizeSearch => 'Optimize search';

  @override
  String get optimizeSearchHelp =>
      'The cards whose reviews Optimize learns from. Leave empty for this deck and its subdecks.';

  @override
  String get notSet => 'Not set';

  @override
  String get clearDate => 'Clear date';

  @override
  String historicalRetentionValue(String retention) {
    return 'Historical retention: $retention';
  }

  @override
  String get historicalRetentionHelp =>
      'How much you remembered in reviews missing from the history, e.g. of cards studied before they were imported.';

  @override
  String get healthCheck => 'Health check';

  @override
  String get healthCheckHelp =>
      'Checks that the new parameters are reliable before using them. Takes longer.';

  @override
  String get rescheduleOnChange => 'Reschedule cards on change';

  @override
  String get rescheduleOnChangeHelp =>
      'When the desired retention or the parameters change, review cards get new due dates. Otherwise only future answers use them.';

  @override
  String get optimizeAllPresets => 'Optimize all presets';

  @override
  String get easyDays => 'Easy Days';

  @override
  String get easyDaysHelp =>
      'Days with fewer reviews. Reviews move to nearby days, by at most the few days intervals already vary by; a Minimum day only gets cards that cannot go elsewhere.';

  @override
  String notANumber(String text) {
    return '\"$text\" is not a number.';
  }

  @override
  String fsrsCountError(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Enter $countString numbers, or leave empty for the default parameters.';
  }

  @override
  String get starting => 'Starting…';

  @override
  String get readingReviewHistory => 'Reading the review history…';

  @override
  String optimizingPercent(String percent) {
    return 'Optimizing… $percent';
  }

  @override
  String checkingParametersPercent(String percent) {
    return 'Checking the parameters… $percent';
  }

  @override
  String get optimizing => 'Optimizing…';

  @override
  String countReviews(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return '$_temp0';
  }

  @override
  String notEnoughReviews(int min, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'are only $countString reviews',
      one: 'is only 1 review',
    );
    return 'Optimizing needs at least $min reviews, but there $_temp0 to learn from. The parameters are unchanged.';
  }

  @override
  String newParametersFitted(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return 'New parameters fitted to $_temp0. Save to use them.';
  }

  @override
  String get healthCheckFailed =>
      'The health check found the new parameters unreliable: the reviews may be too few or too inconsistent. The parameters are unchanged.';

  @override
  String parametersAlreadyBest(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return 'The current parameters already fit your $_temp0 best. They are unchanged.';
  }

  @override
  String get easyNormal => 'Normal';

  @override
  String get easyReduced => 'Reduced';

  @override
  String get easyMinimum => 'Minimum';

  @override
  String notADelay(String text) {
    return '\"$text\" is not a delay. Write a number and a unit (s, m, h or d), e.g. 10m.';
  }

  @override
  String get stepTooShort => 'Each step must be longer than 0.';

  @override
  String get stepTooLong => 'A step can be at most 365 days.';

  @override
  String get tooManySteps => 'There can be at most 20 steps.';

  @override
  String get leechSuspended => 'The card is a leech and was suspended.';

  @override
  String get leech => 'The card is a leech.';

  @override
  String get autoAdvanceOff =>
      'Auto advance is on, but this card\'s deck shows both sides for 0 seconds, which turns it off. Set the seconds in the deck options.';

  @override
  String get timeToShowAnswer => 'Time to show the answer.';

  @override
  String get timeToAnswer => 'Time to answer.';

  @override
  String get recordingUnavailable => 'Recording is not available';

  @override
  String recordingFailed(String error) {
    return 'Your voice could not be recorded: $error';
  }

  @override
  String get timeboxReached => 'Timebox reached';

  @override
  String timeboxStudied(int cards, int minutes) {
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);
    final intl.NumberFormat minutesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String minutesString = minutesNumberFormat.format(minutes);

    String _temp0 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString cards',
      one: '1 card',
    );
    String _temp1 = intl.Intl.pluralLogic(
      minutes,
      locale: localeName,
      other: '$minutesString minutes',
      one: '1 minute',
    );
    return 'You studied $_temp0 in $_temp1.';
  }

  @override
  String get finish => 'Finish';

  @override
  String get continueAction => 'Continue';

  @override
  String get noCardAnsweredYet => 'No card has been answered yet.';

  @override
  String get removeFlag => 'Remove flag';

  @override
  String get stopRecording => 'Stop recording';

  @override
  String get recordVoice => 'Record own voice';

  @override
  String get cardActions => 'Card actions';

  @override
  String get createCopy => 'Create copy';

  @override
  String get replayAudio => 'Replay audio';

  @override
  String get playVoice => 'Play own voice';

  @override
  String get flag => 'Flag';

  @override
  String get unmarkNote => 'Unmark note';

  @override
  String get markNote => 'Mark note';

  @override
  String get previousCardInfo => 'Previous card info';

  @override
  String get resetCard => 'Reset card';

  @override
  String get buryCardUntilTomorrow => 'Bury card until tomorrow';

  @override
  String get buryNoteUntilTomorrow => 'Bury note until tomorrow';

  @override
  String get suspendCard => 'Suspend card';

  @override
  String get suspendNote => 'Suspend note';

  @override
  String get deleteNote => 'Delete note';

  @override
  String get stopAutoAdvance => 'Stop auto advance';

  @override
  String get deckOptions => 'Deck options';

  @override
  String get deckCouldNotOpen => 'This deck could not be opened';

  @override
  String get autoAdvanceIsOn => 'Auto advance is on. Stop it';

  @override
  String get showAnswerButton => 'Show answer';

  @override
  String get previewEnd => 'End';

  @override
  String leaveDeck(String deck) {
    return 'Leave $deck';
  }

  @override
  String get leave => 'Leave';

  @override
  String flagNamed(String name) {
    return 'Flag: $name';
  }

  @override
  String get marked => 'Marked';

  @override
  String get timeOnCard => 'Time on this card';

  @override
  String get emptyCardNotice =>
      'This card is empty: its question no longer shows anything. Remove empty cards with Settings › Empty cards.';

  @override
  String shownAgainEarly(String time) {
    return 'This card again: due in $time, shown early because nothing else is due.';
  }

  @override
  String shownEarly(String time) {
    return 'Shown early: due in $time, because nothing else is due.';
  }

  @override
  String get shownAgainDueNow => 'This card again: it is due again now.';

  @override
  String get learningDueNow => 'Learning cards are due now';

  @override
  String get backToDecks => 'Back to decks';

  @override
  String get congratulationsNothingDue =>
      'Congratulations! Nothing is due right now.';

  @override
  String studiedInTime(int count, String time) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'You studied $_temp0 in $time.';
  }

  @override
  String get sessionComplete => 'Session complete';

  @override
  String streakDaysLabel(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return '$_temp0';
  }

  @override
  String streakAtRisk(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Study today to keep your $countString-day streak';
  }

  @override
  String get streakHeaderTooltip => 'Streaks and daily goal';

  @override
  String goalProgressCards(int done, int goal) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);
    final intl.NumberFormat goalNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String goalString = goalNumberFormat.format(goal);

    String _temp0 = intl.Intl.pluralLogic(
      goal,
      locale: localeName,
      other: '$goalString cards',
      one: '1 card',
    );
    return '$doneString / $_temp0';
  }

  @override
  String goalProgressMinutes(int done, int goal) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);
    final intl.NumberFormat goalNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String goalString = goalNumberFormat.format(goal);

    String _temp0 = intl.Intl.pluralLogic(
      goal,
      locale: localeName,
      other: '$goalString minutes',
      one: '1 minute',
    );
    return '$doneString / $_temp0';
  }

  @override
  String streakExtended(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return 'Streak extended to $_temp0';
  }

  @override
  String streakCurrent(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return 'Current streak: $_temp0';
  }

  @override
  String get goalReached => 'Daily goal reached';

  @override
  String goalCardsLeft(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString more cards',
      one: '1 more card',
    );
    return '$_temp0 to reach today\'s goal';
  }

  @override
  String goalMinutesLeft(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString more minutes',
      one: '1 more minute',
    );
    return '$_temp0 to reach today\'s goal';
  }

  @override
  String get goalSuggestion => 'A daily goal shows how far you are each day.';

  @override
  String get goalSuggestionSet => 'Set a daily goal';

  @override
  String get goalSuggestionDismiss => 'No thanks';

  @override
  String moreLearningDue(String when) {
    return 'More learning cards will be due $when.';
  }

  @override
  String get checkAgain => 'Check again';

  @override
  String get browserMayClear => 'Your browser may clear Cardflip data';

  @override
  String get browserMayClearBody =>
      'This browser did not grant persistent storage, so it may delete your collection when it runs low on space. Export a backup regularly from Settings to keep your cards safe.';

  @override
  String couldNotStart(String error) {
    return 'Cardflip could not start: $error';
  }

  @override
  String get collectionUnreadableShort => 'Your collection could not be read.';

  @override
  String get startNewCollectionTitle => 'Start a new collection?';

  @override
  String get startNewCollectionBody =>
      'This permanently deletes the stored collection that could not be read. Restore a backup instead if you have one.';

  @override
  String get deleteAndStartOver => 'Delete and start over';

  @override
  String get collectionCouldNotOpen => 'Your collection could not be opened';

  @override
  String get nothingDeleted => 'Nothing has been deleted.';

  @override
  String get restoreFromBackup => 'Restore from backup';

  @override
  String get startNewCollection => 'Start a new collection';

  @override
  String get navDecks => 'Decks';

  @override
  String get navBrowse => 'Browse';

  @override
  String get navAdd => 'Add';

  @override
  String get navSettings => 'Settings';

  @override
  String learningDueInDeck(String deck) {
    return 'Learning cards are due in $deck';
  }

  @override
  String learningDueInDecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Learning cards are due in $countString decks';
  }

  @override
  String get showDecks => 'Show decks';

  @override
  String undoStep(String step) {
    return 'Undo $step';
  }

  @override
  String redoStep(String step) {
    return 'Redo $step';
  }

  @override
  String get redo => 'Redo';

  @override
  String get more => 'More';

  @override
  String undone(String step) {
    return 'Undone: $step';
  }

  @override
  String redone(String step) {
    return 'Redone: $step';
  }

  @override
  String get keyCtrl => 'Ctrl';

  @override
  String get keyShift => 'Shift';

  @override
  String get keyAlt => 'Alt';

  @override
  String get keyEnter => 'Enter';

  @override
  String get keyDelete => 'Delete';

  @override
  String get keyEsc => 'Esc';

  @override
  String labelWithShortcut(String label, String shortcut) {
    return '$label ($shortcut)';
  }

  @override
  String get flagRed => 'Red';

  @override
  String get flagOrange => 'Orange';

  @override
  String get flagGreen => 'Green';

  @override
  String get flagBlue => 'Blue';

  @override
  String get flagPink => 'Pink';

  @override
  String get flagTurquoise => 'Turquoise';

  @override
  String get flagPurple => 'Purple';

  @override
  String backupSavedAs(String file) {
    return 'Backup saved as $file';
  }

  @override
  String get replaceCollectionTitle => 'Replace your collection?';

  @override
  String restoreBackupBody(String date) {
    return 'Restoring the backup from $date replaces your current collection entirely: all decks, notes, cards, review history and settings are overwritten. This cannot be undone.';
  }

  @override
  String get restore => 'Restore';

  @override
  String get backupRestored => 'Backup restored';

  @override
  String get deletionLogEmpty => 'The deletion log is empty';

  @override
  String deletionLogSaved(int count, String file) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Deletion log with $_temp0 saved as $file';
  }

  @override
  String get clearDeletionLogTitle => 'Clear the deletion log?';

  @override
  String get clearDeletionLogBody =>
      'The deletion log keeps every note you deleted, so that it can be exported and imported again. Clearing it removes these records for good. This cannot be undone.';

  @override
  String get clear => 'Clear';

  @override
  String get deletionLogCleared => 'Deletion log cleared';

  @override
  String renameFlag(String flag) {
    return 'Rename flag “$flag”';
  }

  @override
  String renameFlagHelper(String flag) {
    return 'Leave empty for \"$flag\"';
  }

  @override
  String get defaultBrowserSearch => 'Default browser search';

  @override
  String get defaultSearchHint => 'e.g. deck:Spanish is:due';

  @override
  String get defaultSearchHelper => 'Leave empty to list all cards';

  @override
  String get off => 'Off';

  @override
  String minutesShort(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString min';
  }

  @override
  String countHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString hours',
      one: '1 hour',
    );
    return '$_temp0';
  }

  @override
  String millisecondsShort(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString ms';
  }

  @override
  String get settings => 'Settings';

  @override
  String get appearance => 'Appearance';

  @override
  String get theme => 'Theme';

  @override
  String get themeSystem => 'System';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get language => 'Language';

  @override
  String get languageSystem => 'System default';

  @override
  String get languageName => 'English';

  @override
  String get uiSize => 'UI size';

  @override
  String get uiSizeHelp => 'Scales text, icons and spacing.';

  @override
  String get minimalistMode => 'Minimalist mode';

  @override
  String get minimalistModeHelp =>
      'Flat surfaces without shadows or tinted backgrounds.';

  @override
  String get reduceMotion => 'Reduce motion';

  @override
  String get reduceMotionHelp =>
      'Show pages, dialogs and menus at once, without animations.';

  @override
  String get reduceMotionSystem => 'Follow system';

  @override
  String get on => 'On';

  @override
  String get studying => 'Studying';

  @override
  String get nextDayStarts => 'Next day starts at';

  @override
  String get nextDayStartsHelp =>
      'Daily limits and due cards reset at this hour.';

  @override
  String get learnAheadLimit => 'Learn ahead limit';

  @override
  String get learnAheadLimitHelp =>
      'When nothing else is due, show learning cards due within this time instead of waiting.';

  @override
  String get timeboxLimit => 'Timebox time limit';

  @override
  String get timeboxLimitHelp =>
      'Each time this much time passes while studying, show how many cards you studied.';

  @override
  String get autoContinueLearning =>
      'Continue automatically when learning cards are due';

  @override
  String get autoContinueLearningHelp =>
      'Otherwise a finished session asks before showing them.';

  @override
  String get showPlayButtons => 'Show play buttons on cards with audio';

  @override
  String get interruptAudio => 'Interrupt current audio when answering';

  @override
  String get interruptAudioHelp =>
      'Otherwise the answer\'s audio plays after the question\'s.';

  @override
  String get stripPasteFormatting => 'Strip formatting when pasting';

  @override
  String stripPasteFormattingHelp(String key) {
    return 'Paste plain text; $key+Shift+V pastes with formatting.';
  }

  @override
  String get pasteImagesAsPng => 'Paste images as PNG';

  @override
  String get pasteImagesAsPngHelp =>
      'Lossless but larger; otherwise pasted images are JPEG.';

  @override
  String get deckForNewCards => 'Deck for new cards';

  @override
  String get currentDeck => 'Current deck';

  @override
  String get deckByNoteType => 'Change deck depending on note type';

  @override
  String get goalsAndStreaks => 'Goals and streaks';

  @override
  String get showStreaksAndGoals => 'Show streaks and goals';

  @override
  String get showStreaksAndGoalsHelp =>
      'On the deck list and after studying. Only on this device.';

  @override
  String get dailyGoal => 'Daily goal';

  @override
  String get dailyGoalNone => 'None';

  @override
  String get dailyGoalCards => 'Cards';

  @override
  String get dailyGoalMinutes => 'Minutes';

  @override
  String dailyGoalCardsPerDay(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0 a day';
  }

  @override
  String dailyGoalMinutesPerDay(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString minutes',
      one: '1 minute',
    );
    return '$_temp0 a day';
  }

  @override
  String get dailyGoalCardsLabel => 'Cards per day';

  @override
  String get dailyGoalMinutesLabel => 'Minutes per day';

  @override
  String get dailyGoalError => 'Enter a whole number from 1 to 9999.';

  @override
  String get dailyGoalHelp =>
      'Studying on any day keeps your streak, whether or not you reach the goal.';

  @override
  String get restDaysPerWeek => 'Rest days per week';

  @override
  String get restDaysPerWeekHelp =>
      'Days without study that keep a streak going, in any 7 days in a row.';

  @override
  String get achievements => 'Achievements';

  @override
  String get achievementAnswers => 'Answers';

  @override
  String get achievementStreak => 'Streak';

  @override
  String get achievementStudyDays => 'Study days';

  @override
  String get achievementDailyGoal => 'Daily goal';

  @override
  String get achievementLongTerm => 'Long-term memory';

  @override
  String get achievementComeback => 'Comeback';

  @override
  String get tierBronze => 'Bronze';

  @override
  String get tierSilver => 'Silver';

  @override
  String get tierGold => 'Gold';

  @override
  String get tierPlatinum => 'Platinum';

  @override
  String tierAnswers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString answers',
      one: '1 answer',
    );
    return '$_temp0';
  }

  @override
  String tierDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return '$_temp0';
  }

  @override
  String tierGoalDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return 'met on $_temp0';
  }

  @override
  String tierCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0';
  }

  @override
  String get tierOneTime => 'One-time';

  @override
  String tierWithThreshold(String tier, String threshold) {
    return '$tier · $threshold';
  }

  @override
  String achievementEarnedOn(String date) {
    return 'Earned $date';
  }

  @override
  String achievementProgress(int progress, int next) {
    final intl.NumberFormat progressNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String progressString = progressNumberFormat.format(progress);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return '$progressString / $nextString';
  }

  @override
  String get achievementAllEarned => 'All tiers earned';

  @override
  String get achievementNeedsGoal => 'Set a daily goal to start';

  @override
  String get achievementComebackHint =>
      'Study again after 14 days or more without study';

  @override
  String achievementLocked(String name) {
    return '$name, not earned yet';
  }

  @override
  String achievementEarned(String name, String tier) {
    return '$name, $tier';
  }

  @override
  String newAnswers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString answers',
      one: '1 answer',
    );
    return '$_temp0';
  }

  @override
  String newStreak(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString-day streak';
  }

  @override
  String newStudyDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString study days',
      one: '1 study day',
    );
    return '$_temp0';
  }

  @override
  String newDailyGoal(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return 'Daily goal met on $_temp0';
  }

  @override
  String newLongTerm(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0 in long-term memory';
  }

  @override
  String newAchievementsLine(String names) {
    return 'New: $names';
  }

  @override
  String newAchievementsCountLine(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'New: $countString achievements';
  }

  @override
  String newAchievement(String name) {
    return 'New achievement: $name';
  }

  @override
  String newAchievements(String names) {
    return 'New achievements: $names';
  }

  @override
  String achievementsFromHistory(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString achievements from your history';
  }

  @override
  String get showAchievements => 'Show';

  @override
  String get review => 'Review';

  @override
  String get showRemainingCounts => 'Show remaining card count';

  @override
  String get showIntervals => 'Show next review time above answer buttons';

  @override
  String get spaceAnswersGood => 'Space/Enter also answers Good';

  @override
  String get spaceAnswersGoodHelp => 'Otherwise they only reveal the answer.';

  @override
  String get hideBarsDuringReview => 'Hide bars during review';

  @override
  String get hideBarsDuringReviewHelp =>
      'Hide the app bar and navigation while studying; Escape or the top edge leaves.';

  @override
  String get questionFade => 'Question fade';

  @override
  String get questionFadeHelp => 'How long each new question fades in.';

  @override
  String get browsing => 'Browsing';

  @override
  String get ignoreAccents => 'Ignore accents in search';

  @override
  String get ignoreAccentsHelp => 'For example, \"cafe\" also finds \"café\".';

  @override
  String get allCards => 'All cards';

  @override
  String get collection => 'Collection';

  @override
  String get import => 'Import';

  @override
  String get importHelp =>
      'Text (.csv, .tsv, .txt), Anki package (.apkg) or collection (.colpkg)';

  @override
  String get exportBackup => 'Export backup';

  @override
  String get downloadsBackupFile => 'Downloads a .cardflip file';

  @override
  String get savesBackupFile => 'Saves a .cardflip file';

  @override
  String get exportCollection => 'Export collection';

  @override
  String get exportCollectionHelp => 'Anki collection package or plain text';

  @override
  String get restoreBackup => 'Restore backup';

  @override
  String get restoreBackupHelp => 'Replaces the current collection';

  @override
  String get emptyCards => 'Empty cards';

  @override
  String get emptyCardsHelp => 'Find and delete cards whose question is empty';

  @override
  String get checkMedia => 'Check media';

  @override
  String get checkMediaHelp =>
      'Find unused and missing media files, restore deleted ones';

  @override
  String get checkDatabase => 'Check database';

  @override
  String get checkDatabaseHelp => 'Find and repair problems in the collection';

  @override
  String get exportDeletionLog => 'Export deletion log';

  @override
  String get exportDeletionLogHelp =>
      'Deleted notes as a text file that Import can read';

  @override
  String get clearDeletionLog => 'Clear deletion log';

  @override
  String get clearDeletionLogHelp => 'Forgets the notes deleted so far';

  @override
  String get backups => 'Backups';

  @override
  String get flags => 'Flags';

  @override
  String get agentsMcp => 'Agents (MCP)';

  @override
  String get allowAgents => 'Allow agents (MCP)';

  @override
  String get allowAgentsHelp =>
      'AI agents such as Claude Code can read and change the open collection while Cardflip runs.';

  @override
  String get allowAgentsTitle => 'Allow agents?';

  @override
  String get allowAgentsBody =>
      'Any program on this computer that has the token can read and change your collection while Cardflip is open. Give the token only to agents you trust.';

  @override
  String get allow => 'Allow';

  @override
  String get mcpServerRunning => 'Server running';

  @override
  String get mcpServerStopped => 'Server not running';

  @override
  String get mcpPort => 'Port';

  @override
  String get mcpPortHelper => '1024 to 65535';

  @override
  String get mcpToken => 'Token';

  @override
  String get copyToken => 'Copy token';

  @override
  String get tokenCopied => 'Token copied';

  @override
  String get regenerateToken => 'Regenerate token';

  @override
  String get regenerateTokenTitle => 'Regenerate the token?';

  @override
  String get regenerateTokenBody =>
      'Agents that use the current token can no longer connect until you give them the new one.';

  @override
  String get regenerate => 'Regenerate';

  @override
  String get copyClientConfig => 'Copy client configuration';

  @override
  String get copyClientConfigHelp =>
      'The server URL and token for Claude Code, Claude Desktop and other MCP clients';

  @override
  String get clientConfigCopied => 'Client configuration copied';

  @override
  String get agentActivity => 'Recent agent changes';

  @override
  String get noAgentActivity => 'No changes in this session';

  @override
  String get storage => 'Storage';

  @override
  String get storagePersistent =>
      'Persistent: the browser keeps your collection.';

  @override
  String get storageNotPersistent =>
      'The browser may clear your collection when space runs low. Export backups regularly.';

  @override
  String get about => 'About';

  @override
  String aboutCardflip(String version) {
    return 'Version $version\nSpaced repetition flash cards. Your data stays on this device.\nCopyright © 2026 madzione. Free software under the GNU GPL version 3 or later, with absolutely no warranty.';
  }

  @override
  String get openSourceLicenses => 'Open-source licenses';

  @override
  String get sourceCode => 'Source code';

  @override
  String get licenseLegalese =>
      'Copyright © 2026 madzione\n\nCardflip is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License, version 3 or later. It comes with absolutely no warranty.';

  @override
  String backupEveryMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Every $countString min';
  }

  @override
  String get backupEveryHour => 'Every hour';

  @override
  String backupEveryHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Every $countString hours';
  }

  @override
  String restoreAutoBackupBody(String date) {
    return 'Restoring the automatic backup from $date replaces your current collection entirely: all decks, notes, cards, review history and settings are overwritten. Media files stay as they are. Your current collection is backed up first, so you can go back to it.';
  }

  @override
  String get backUpAutomatically => 'Back up automatically';

  @override
  String get backUpAutomaticallyHelp =>
      'While the collection changes, without media. Before a restore a backup is always made.';

  @override
  String get keepBackups => 'Keep besides today\'s backups';

  @override
  String get keepDaily => 'Daily';

  @override
  String get keepWeekly => 'Weekly';

  @override
  String get keepMonthly => 'Monthly';

  @override
  String get createBackupNow => 'Create backup now';

  @override
  String get noAutoBackups => 'No automatic backups yet';

  @override
  String countAutoBackups(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString automatic backups',
      one: '1 automatic backup',
    );
    return '$_temp0';
  }

  @override
  String usedOnDevice(String size) {
    return '$size used on this device';
  }

  @override
  String get showFewer => 'Show fewer';

  @override
  String showAllCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Show all $countString';
  }

  @override
  String deletedEmptyCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString empty cards',
      one: '1 empty card',
    );
    return 'Deleted $_temp0';
  }

  @override
  String get close => 'Close';

  @override
  String get noEmptyCards => 'There are no empty cards.';

  @override
  String emptyCardsSummary(int cards, int notes) {
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString empty cards',
      one: '1 empty card',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    return '$_temp0 in $_temp1. Their question no longer shows anything.';
  }

  @override
  String emptyCardsKept(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes have',
      one: '1 note has',
    );
    return '$_temp0 only empty cards. One card of each is kept, so no note is deleted.';
  }

  @override
  String get emptyText => '(empty)';

  @override
  String get oneCardKept => 'one card is kept';

  @override
  String deleteCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Delete $_temp0';
  }

  @override
  String filesWithSize(int count, String size) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString files',
      one: '1 file',
    );
    return '$_temp0 ($size)';
  }

  @override
  String get deleteUnusedTitle => 'Delete unused files?';

  @override
  String deleteUnusedBody(String files) {
    return 'Moves $files to the media trash. Restore deleted brings them back until the trash is emptied; files are removed from the trash after 30 days.';
  }

  @override
  String movedToTrash(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString files',
      one: '1 file',
    );
    return 'Moved $_temp0 to the media trash';
  }

  @override
  String get tagMissingTitle => 'Tag notes with missing files?';

  @override
  String tagMissingBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Adds the tag missing-media to $_temp0, so that you can find them in the browser with tag:missing-media.';
  }

  @override
  String get tag => 'Tag';

  @override
  String taggedMissing(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Tagged $_temp0 missing-media';
  }

  @override
  String get restoreDeletedTitle => 'Restore deleted files?';

  @override
  String restoreDeletedBody(String files) {
    return 'Moves $files back into your media.';
  }

  @override
  String restoredFilesKept(int count, int kept) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);
    final intl.NumberFormat keptNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String keptString = keptNumberFormat.format(kept);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString files',
      one: '1 file',
    );
    String _temp1 = intl.Intl.pluralLogic(
      kept,
      locale: localeName,
      other: '$keptString files',
      one: '1 file',
    );
    return 'Restored $_temp0; $_temp1 stayed in the trash because a different file now has the name';
  }

  @override
  String restoredFiles(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString files',
      one: '1 file',
    );
    return 'Restored $_temp0';
  }

  @override
  String get emptyTrashTitle => 'Empty the media trash?';

  @override
  String emptyTrashBody(String files) {
    return 'Deletes $files for good. This cannot be undone.';
  }

  @override
  String get emptyTrash => 'Empty trash';

  @override
  String get trashEmptied => 'Media trash emptied';

  @override
  String get unusedFiles => 'Unused files';

  @override
  String get allMediaUsed =>
      'Every media file is used by a note, a template or styling.';

  @override
  String unusedFilesBody(String files) {
    return '$files that no note, template or styling refers to. Files whose names start with _ are always kept.';
  }

  @override
  String get deleteUnused => 'Delete unused';

  @override
  String get missingFiles => 'Missing files';

  @override
  String get noMissingFiles => 'Every file that notes refer to is there.';

  @override
  String missingFilesBody(int count, int notes) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString files',
      one: '1 file',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes refer',
      one: '1 note refers',
    );
    String _temp2 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'are',
      one: 'is',
    );
    return '$_temp0 that $_temp1 to $_temp2 not in your media.';
  }

  @override
  String get tagMissing => 'Tag missing';

  @override
  String get mediaTrash => 'Media trash';

  @override
  String get trashIsEmpty => 'The trash is empty.';

  @override
  String trashBody(String files) {
    return '$files. Files are removed from the trash after 30 days.';
  }

  @override
  String get restoreDeleted => 'Restore deleted';

  @override
  String get checkingNotes => 'Checking notes';

  @override
  String get checkingCards => 'Checking cards';

  @override
  String get readingHistory => 'Reading the review history';

  @override
  String get savingRepairs => 'Saving repairs';

  @override
  String get backingUp => 'Backing up the collection';

  @override
  String get noProblemsFound => 'No problems were found.';

  @override
  String orphanedReviews(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString review log entries belong',
      one: '1 review log entry belongs',
    );
    return '$_temp0 to deleted cards. They are kept for statistics.';
  }

  @override
  String get updateIfNewer => 'If the package version is newer';

  @override
  String get updateAlways => 'Always';

  @override
  String get updateNever => 'Never';

  @override
  String get separatorComma => 'Comma';

  @override
  String get separatorSemicolon => 'Semicolon';

  @override
  String get separatorTab => 'Tab';

  @override
  String get separatorSpace => 'Space';

  @override
  String get separatorPipe => 'Pipe';

  @override
  String get separatorColon => 'Colon';

  @override
  String get scopeNoteType => 'In its note type';

  @override
  String get scopeDeck => 'In its note type and deck';

  @override
  String importedNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Imported $_temp0';
  }

  @override
  String get importIntro =>
      'Import notes from a text file (.csv, .tsv or .txt), or decks from an Anki package (.apkg). An Anki collection package (.colpkg) replaces your collection. The file is read on this device and never uploaded.';

  @override
  String get chooseAFile => 'Choose a file';

  @override
  String get importFailed => 'The import failed';

  @override
  String get collectionNotChanged => 'Your collection was not changed.';

  @override
  String get chooseAnotherFile => 'Choose another file';

  @override
  String fileIsBackup(String file) {
    return '$file is a Cardflip backup. Restore it from Settings to replace your collection.';
  }

  @override
  String columnNumber(int number) {
    return 'Column $number';
  }

  @override
  String csvSummary(String file, int rows, int columns) {
    final intl.NumberFormat rowsNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String rowsString = rowsNumberFormat.format(rows);
    final intl.NumberFormat columnsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String columnsString = columnsNumberFormat.format(columns);

    String _temp0 = intl.Intl.pluralLogic(
      rows,
      locale: localeName,
      other: '$rowsString rows',
      one: '1 row',
    );
    String _temp1 = intl.Intl.pluralLogic(
      columns,
      locale: localeName,
      other: '$columnsString columns',
      one: '1 column',
    );
    return '$file: $_temp0, $_temp1';
  }

  @override
  String tagsForEveryNote(String tags) {
    return 'Tags added to every note: $tags';
  }

  @override
  String rowsSkippedUnknownType(int count, String names) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString rows',
      one: '1 row',
    );
    return '$_temp0 will be skipped: there is no note type named $names.';
  }

  @override
  String get noteTypeOfRowsWithout => 'Note type of rows without one';

  @override
  String get noteType => 'Note type';

  @override
  String get deckOfRowsWithout => 'Deck of rows without one';

  @override
  String get deck => 'Deck';

  @override
  String newDeckName(String deck) {
    return '$deck (new)';
  }

  @override
  String get separator => 'Separator';

  @override
  String get allowHtml => 'Allow HTML in fields';

  @override
  String get allowHtmlHelp =>
      'When off, text like \"a < b & c\" is imported as typed.';

  @override
  String get firstRowNames => 'The first row contains column names';

  @override
  String get tags => 'Tags';

  @override
  String get guid => 'GUID';

  @override
  String get ignore => 'Ignore';

  @override
  String get whenFirstFieldMatches =>
      'When the first field matches an existing note';

  @override
  String get guidUpdates =>
      'A row whose GUID matches a note of its note type updates that note.';

  @override
  String get skipRow => 'Skip the row';

  @override
  String get updateExistingNote => 'Update the existing note';

  @override
  String get addAsDuplicate => 'Add it as a duplicate';

  @override
  String get lookForMatchingNotes => 'Look for matching notes';

  @override
  String apkgSummary(int notes, int cards, int media) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);
    final intl.NumberFormat mediaNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String mediaString = mediaNumberFormat.format(media);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString cards',
      one: '1 card',
    );
    String _temp2 = intl.Intl.pluralLogic(
      media,
      locale: localeName,
      other: '$mediaString media files',
      one: '1 media file',
    );
    return '$_temp0, $_temp1, $_temp2';
  }

  @override
  String inDecks(String decks) {
    return ' in $decks';
  }

  @override
  String unsupportedMedia(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString audio or video files',
      one: '1 audio or video file',
    );
    return '$_temp0 will be skipped: Cardflip cannot play their format.';
  }

  @override
  String get scheduling => 'Scheduling';

  @override
  String get keepScheduling => 'Keep scheduling';

  @override
  String get keepSchedulingHelp =>
      'Cards keep their due dates and progress from Anki.';

  @override
  String get importAsNew => 'Import all cards as new';

  @override
  String get importAsNewHelp => 'Start learning everything from scratch.';

  @override
  String get notesAlreadyThere => 'Notes already in your collection';

  @override
  String get notesAlreadyThereHelp =>
      'Notes and note types from an earlier import of this deck. Their scheduling is kept either way.';

  @override
  String get updateNotes => 'Update notes';

  @override
  String get updateNoteTypes => 'Update note types';

  @override
  String get mergeNoteTypes => 'Merge note types';

  @override
  String get mergeNoteTypesHelp =>
      'When the package adds fields or card types, add them to your note type. Otherwise the package\'s version is imported as a separate note type.';

  @override
  String replaceByImportBody(String file) {
    return 'Importing $file replaces your current collection entirely: all decks, notes, cards, review history and media files are overwritten. Your settings are kept. This cannot be undone.';
  }

  @override
  String get replace => 'Replace';

  @override
  String colpkgSummary(int notes, int cards, int media, int decks) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);
    final intl.NumberFormat mediaNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String mediaString = mediaNumberFormat.format(media);
    final intl.NumberFormat decksNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String decksString = decksNumberFormat.format(decks);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString cards',
      one: '1 card',
    );
    String _temp2 = intl.Intl.pluralLogic(
      media,
      locale: localeName,
      other: '$mediaString media files',
      one: '1 media file',
    );
    String _temp3 = intl.Intl.pluralLogic(
      decks,
      locale: localeName,
      other: '$decksString decks',
      one: '1 deck',
    );
    return 'Anki collection: $_temp0, $_temp1, $_temp2 in $_temp3';
  }

  @override
  String get replaceWarning =>
      'This replaces your current collection with the Anki collection, with its scheduling and review history. Everything you have in Cardflip now is overwritten; your settings are kept. Cardflip makes an automatic backup first, but you may want to export a backup to a file before you continue.';

  @override
  String get exportBackupFirst => 'Export a backup first';

  @override
  String get replaceCollection => 'Replace collection';

  @override
  String get addingNotes => 'Adding notes';

  @override
  String get addingHistory => 'Adding review history';

  @override
  String get savingImages => 'Saving images';

  @override
  String get saving => 'Saving';

  @override
  String get readingFile => 'Reading the file';

  @override
  String progressOf(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return '$doneString of $total';
  }

  @override
  String notesAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 added';
  }

  @override
  String notesUpdated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 updated';
  }

  @override
  String notesSkipped(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 skipped';
  }

  @override
  String notesNotUpdatedNewer(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 not updated: your version is newer';
  }

  @override
  String cardsAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0 added';
  }

  @override
  String reviewsImported(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return '$_temp0 of history imported';
  }

  @override
  String mediaImported(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString media files',
      one: '1 media file',
    );
    return '$_temp0 imported';
  }

  @override
  String commentLinesSkipped(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString comment lines',
      one: '1 comment line',
    );
    return '$_temp0 skipped';
  }

  @override
  String noteTypesAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString note types',
      one: '1 note type',
    );
    return '$_temp0 added';
  }

  @override
  String noteTypesUpdated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString note types',
      one: '1 note type',
    );
    return '$_temp0 updated';
  }

  @override
  String noteTypesMerged(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString note types',
      one: '1 note type',
    );
    return '$_temp0 merged';
  }

  @override
  String get collectionReplaced => 'Collection replaced';

  @override
  String get importComplete => 'Import complete';

  @override
  String get warnings => 'Warnings';

  @override
  String get done => 'Done';

  @override
  String copyName(String name) {
    return '$name copy';
  }

  @override
  String get addNoteType => 'Add note type';

  @override
  String cloneNamed(String name) {
    return 'Clone “$name”';
  }

  @override
  String get noNotesUseType => 'No notes use this note type.';

  @override
  String deleteNoteTypeBody(int notes, int cards) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);
    final intl.NumberFormat cardsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardsString = cardsNumberFormat.format(cards);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString cards',
      one: '1 card',
    );
    return 'This deletes $_temp0 and $_temp1.';
  }

  @override
  String get canUndoUntilClose => 'You can undo this until you close Cardflip.';

  @override
  String get noteTypes => 'Note types';

  @override
  String get builtinDeckName => 'Default';

  @override
  String get builtinPresetName => 'Default';

  @override
  String get stockBasic => 'Basic';

  @override
  String get stockReversed => 'Basic (and reversed card)';

  @override
  String get stockOptionalReversed => 'Basic (optional reversed card)';

  @override
  String get stockTypeIn => 'Basic (type in the answer)';

  @override
  String get stockCloze => 'Cloze';

  @override
  String get stockImageOcclusion => 'Image Occlusion';

  @override
  String get stockFieldFront => 'Front';

  @override
  String get stockFieldBack => 'Back';

  @override
  String get stockFieldAddReverse => 'Add Reverse';

  @override
  String get stockFieldText => 'Text';

  @override
  String get stockFieldBackExtra => 'Back Extra';

  @override
  String get stockFieldOcclusion => 'Occlusion';

  @override
  String get stockFieldImage => 'Image';

  @override
  String get stockFieldHeader => 'Header';

  @override
  String get stockFieldComments => 'Comments';

  @override
  String noteTypeSummary(int fields, int cardTypes, int notes) {
    final intl.NumberFormat fieldsNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String fieldsString = fieldsNumberFormat.format(fields);
    final intl.NumberFormat cardTypesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String cardTypesString = cardTypesNumberFormat.format(cardTypes);
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      fields,
      locale: localeName,
      other: '$fieldsString fields',
      one: '1 field',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cardTypes,
      locale: localeName,
      other: '$cardTypesString card types',
      one: '1 card type',
    );
    String _temp2 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    return '$_temp0 · $_temp1 · $_temp2';
  }

  @override
  String get noteTypeActions => 'Note type actions';

  @override
  String get renameNoteType => 'Rename note type';

  @override
  String get renameField => 'Rename field';

  @override
  String get addField => 'Add field';

  @override
  String removeFieldTitle(String field) {
    return 'Remove field “$field”?';
  }

  @override
  String removeFieldBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'The field and its content will be removed from $_temp0. You can undo this until you close Cardflip.';
  }

  @override
  String get remove => 'Remove';

  @override
  String defaultCardTypeName(int number) {
    return 'Card $number';
  }

  @override
  String removeCardTypeTitle(String name) {
    return 'Remove card type “$name”?';
  }

  @override
  String removeCardTypeBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'This deletes $_temp0 and their review history cannot be used anymore.';
  }

  @override
  String get fields => 'Fields';

  @override
  String get fieldOptions => 'Field options';

  @override
  String get removeField => 'Remove field';

  @override
  String get dragToReorder => 'Drag to reorder';

  @override
  String get cardTypes => 'Card types';

  @override
  String get cardTypeOptions => 'Card type options';

  @override
  String get removeCardType => 'Remove card type';

  @override
  String get addCardType => 'Add card type';

  @override
  String get styling => 'Styling';

  @override
  String get editCss => 'Edit CSS';

  @override
  String get sortField => 'Sort field';

  @override
  String get rightToLeft => 'Right to left';

  @override
  String get collapsed => 'Collapsed';

  @override
  String get notInUnqualifiedSearches => 'Not in unqualified searches';

  @override
  String get htmlEditor => 'HTML editor';

  @override
  String get pinned => 'Pinned';

  @override
  String get fieldSizeError => 'Enter a size from 1 to 200, or leave it empty.';

  @override
  String optionsForNamed(String name) {
    return 'Options for “$name”';
  }

  @override
  String get editingFont => 'Editing font';

  @override
  String get editingFontHelp =>
      'Used for this field in the editor only, not on cards.';

  @override
  String get font => 'Font';

  @override
  String get defaultValue => 'Default';

  @override
  String get size => 'Size';

  @override
  String get sortByField => 'Sort by this field';

  @override
  String get sortByFieldIsHelp =>
      'The browser’s Sort field column shows this field. To change it, choose another field.';

  @override
  String get sortByFieldHelp =>
      'The browser’s Sort field column shows this field.';

  @override
  String get rightToLeftHelp =>
      'For Arabic, Hebrew and other right-to-left scripts.';

  @override
  String get collapseByDefault => 'Collapse by default';

  @override
  String get collapseByDefaultHelp =>
      'The editor shows only the field name until you tap it.';

  @override
  String get excludeFromSearch => 'Exclude from unqualified searches';

  @override
  String excludeFromSearchHelp(String field) {
    return 'Plain search terms skip this field; $field:… still finds it.';
  }

  @override
  String get htmlByDefault => 'Use HTML editor by default';

  @override
  String get htmlByDefaultHelp =>
      'Open the field as HTML source instead of the visual editor.';

  @override
  String get keepAfterAdding => 'Keep after adding';

  @override
  String get keepAfterAddingHelp =>
      'Pin the field, so its content stays in the add screen after adding a note.';

  @override
  String get deckOverride => 'Deck override';

  @override
  String get deckOverrideHelp =>
      'New cards of this card type go to this deck. Existing cards are not moved.';

  @override
  String get deckOverrideNone => 'None (the deck chosen when adding)';

  @override
  String get browserAppearance => 'Browser appearance';

  @override
  String get browserAppearanceHelp =>
      'Simplified templates for the browser\'s question and answer columns. Leave empty to use the normal templates.';

  @override
  String get browserQuestionTemplate => 'Browser question template';

  @override
  String get browserAnswerTemplate => 'Browser answer template';

  @override
  String get edit => 'Edit';

  @override
  String get preview => 'Preview';

  @override
  String get frontTemplate => 'Front template';

  @override
  String get backTemplate => 'Back template';

  @override
  String get templateHelp =>
      'Insert a field with {{Field}}. Add a filter in front of the name, e.g. {{type}} for a type-in answer box, {{hint}}, {{text}} or {{cloze}}. Also {{FrontSide}} and sections {{#Field}}…{{/Field}}.';

  @override
  String get newCardType => 'New card type';

  @override
  String get editCardType => 'Edit card type';

  @override
  String get stylingCss => 'Styling (CSS)';

  @override
  String get lossyNote =>
      'This field contains formatting the visual editor can\'t keep. Editing as HTML.';

  @override
  String get switchToVisualTitle => 'Switch to the visual editor?';

  @override
  String get switchToVisualBody =>
      'This field contains formatting the visual editor can\'t keep. If you edit it there, that formatting may be lost.';

  @override
  String get switchAction => 'Switch';

  @override
  String get expandField => 'Expand field';

  @override
  String get collapseField => 'Collapse field';

  @override
  String get visualEditor => 'Visual editor';

  @override
  String get emptyField => 'Empty';

  @override
  String get bold => 'Bold';

  @override
  String get italic => 'Italic';

  @override
  String get underline => 'Underline';

  @override
  String get subscript => 'Subscript';

  @override
  String get superscript => 'Superscript';

  @override
  String get textColour => 'Text colour';

  @override
  String get highlightColour => 'Highlight colour';

  @override
  String get removeFormatting => 'Remove formatting';

  @override
  String get bulletedList => 'Bulleted list';

  @override
  String get numberedList => 'Numbered list';

  @override
  String get alignment => 'Alignment';

  @override
  String get alignLeft => 'Align left';

  @override
  String get alignCenter => 'Center';

  @override
  String get alignRight => 'Align right';

  @override
  String get justify => 'Justify';

  @override
  String get outdent => 'Outdent';

  @override
  String get indent => 'Indent';

  @override
  String get colourBlack => 'Black';

  @override
  String get colourGray => 'Gray';

  @override
  String get colourRed => 'Red';

  @override
  String get colourOrange => 'Orange';

  @override
  String get colourYellow => 'Yellow';

  @override
  String get colourGreen => 'Green';

  @override
  String get colourBlue => 'Blue';

  @override
  String get colourPurple => 'Purple';

  @override
  String get noColour => 'No colour';

  @override
  String get mathInline => 'MathJax inline';

  @override
  String get mathBlock => 'MathJax block';

  @override
  String get mathChemistry => 'MathJax chemistry';

  @override
  String get mathLatex => 'LaTeX';

  @override
  String get mathLatexEquation => 'LaTeX equation';

  @override
  String get mathLatexMath => 'LaTeX math env.';

  @override
  String get math => 'Math';

  @override
  String get clozeDeletion => 'Cloze deletion';

  @override
  String get nextClozeNumber => 'Next cloze number';

  @override
  String get sameClozeNumber => 'Same cloze number';

  @override
  String get attachMedia => 'Attach picture, audio or video';

  @override
  String get recordAudio => 'Record audio';

  @override
  String get onlyMediaInFields =>
      'Only pictures, audio and video can be added to a field.';

  @override
  String get noNotesYet => 'No notes yet';

  @override
  String clozeNumber(int number) {
    return 'Cloze $number';
  }

  @override
  String get previewLightTheme => 'Preview light theme';

  @override
  String get previewDarkTheme => 'Preview dark theme';

  @override
  String get front => 'Front';

  @override
  String get back => 'Back';

  @override
  String get recording => 'Recording…';

  @override
  String get insert => 'Insert';

  @override
  String get addTags => 'Add tags';

  @override
  String imageMissing(String file) {
    return 'The image $file is missing.';
  }

  @override
  String get choosePictureToDrawOn => 'Choose a picture to draw on.';

  @override
  String get chooseImageToDrawOn => 'Choose an image to draw on.';

  @override
  String get drawShapes => 'Draw a shape over each part to ask.';

  @override
  String get cardAdded => 'Card added';

  @override
  String get editCard => 'Edit card';

  @override
  String get historyAddedToday => 'History: cards added today';

  @override
  String get image => 'Image';

  @override
  String get changeImage => 'Change image';

  @override
  String get loadingImage => 'Loading the image…';

  @override
  String get chooseImage => 'Choose image';

  @override
  String duplicateOne(String field) {
    return 'This is a duplicate: another note has the same $field field.';
  }

  @override
  String duplicateMany(int count, String field) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'This is a duplicate: $countString other notes have the same $field field.';
  }

  @override
  String get showDuplicates => 'Show duplicates';

  @override
  String get unpinField => 'Unpin: clear after adding';

  @override
  String get pinField => 'Pin: keep after adding';

  @override
  String get hideAllGuessOne => 'Hide all, guess one';

  @override
  String get hideOneGuessOne => 'Hide one, guess one';

  @override
  String get selectAndMove => 'Select and move';

  @override
  String get rectangle => 'Rectangle';

  @override
  String get ellipse => 'Ellipse';

  @override
  String get polygonTool =>
      'Polygon: tap the corners, then the first one again';

  @override
  String get textLabel => 'Text label';

  @override
  String get finishPolygon => 'Finish the polygon';

  @override
  String get groupShapes => 'Group: ask the selected shapes together';

  @override
  String get ungroup => 'Ungroup';

  @override
  String get deleteSelectedShapes => 'Delete the selected shapes';

  @override
  String get zoomOut => 'Zoom out';

  @override
  String get zoomIn => 'Zoom in';

  @override
  String get fitImage => 'Fit the image';

  @override
  String get imageCannotShow => 'This image cannot be shown.';

  @override
  String get shownOnEveryCard => 'Shown on every card';

  @override
  String get repositionNewCards => 'Reposition new cards';

  @override
  String get startPosition => 'Start position';

  @override
  String get step => 'Step';

  @override
  String get randomizeOrder => 'Randomize order';

  @override
  String get shiftExisting => 'Shift position of existing cards';

  @override
  String get repositionHelp =>
      'Only new cards are repositioned; others are left unchanged.';

  @override
  String get reposition => 'Reposition';

  @override
  String wholeNumberRange(int min, int max) {
    return 'A whole number from $min to $max';
  }

  @override
  String get setDueDate => 'Set due date';

  @override
  String setDueDateOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Set due date of $_temp0';
  }

  @override
  String get daysFromToday => 'Days from today';

  @override
  String get dueDateHint => 'e.g. 0, 7, 60-90 or 30!';

  @override
  String get dueDateHelp =>
      '0 = today, 1 = tomorrow, 7 = in a week\n60-90 = a random day 60 to 90 days from now, for each card\nAdd ! to set the interval to the same number of days, e.g. 30!';

  @override
  String deleteCardsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Delete $_temp0?';
  }

  @override
  String resetCardsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Reset $_temp0 to new?';
  }

  @override
  String get resetCardsBody =>
      'Their scheduling is cleared and they will be studied as new cards. Their review history is kept.';

  @override
  String get restorePosition => 'Restore original position where possible';

  @override
  String get resetCounts => 'Reset repetition and lapse counts';

  @override
  String get reset => 'Reset';

  @override
  String get noFlag => 'No flag';

  @override
  String taggedNotes(int count, String tag) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Tagged $_temp0 $tag';
  }

  @override
  String get findDuplicates => 'Find duplicates';

  @override
  String get tagDuplicates => 'Tag duplicates';

  @override
  String get field => 'Field';

  @override
  String get optionalSearch => 'Optional search';

  @override
  String get optionalSearchHint => 'e.g. deck:Spanish';

  @override
  String get find => 'Find';

  @override
  String get noDuplicatesFound => 'No duplicates found.';

  @override
  String countNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0';
  }

  @override
  String includingOtherCardsNote(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString other cards',
      one: '1 other card',
    );
    return 'Including $_temp0 from the same note. ';
  }

  @override
  String includingOtherCardsNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString other cards',
      one: '1 other card',
    );
    return 'Including $_temp0 from the same notes. ';
  }

  @override
  String newPosition(int position) {
    return 'New #$position';
  }

  @override
  String get colQuestion => 'Question';

  @override
  String get colAnswer => 'Answer';

  @override
  String get colSortField => 'Sort field';

  @override
  String get colCardType => 'Card type';

  @override
  String get colNoteType => 'Note type';

  @override
  String get colDeck => 'Deck';

  @override
  String get colDue => 'Due';

  @override
  String get colInterval => 'Interval';

  @override
  String get colDifficulty => 'Difficulty';

  @override
  String get colStability => 'Stability';

  @override
  String get colRetrievability => 'Retrievability';

  @override
  String get colReviews => 'Reviews';

  @override
  String get colLapses => 'Lapses';

  @override
  String get colTags => 'Tags';

  @override
  String get colCreated => 'Created';

  @override
  String get colCardModified => 'Card modified';

  @override
  String get colNoteModified => 'Note modified';

  @override
  String get colCards => 'Cards';

  @override
  String get emptyQuestion => '(empty)';

  @override
  String get stateSuspended => 'suspended';

  @override
  String get stateNew => 'new';

  @override
  String dueWhen(String when) {
    return 'due $when';
  }

  @override
  String intervalOfState(String state) {
    String _temp0 = intl.Intl.selectLogic(state, {
      'new': 'new',
      'learning': 'learning',
      'relearning': 'relearning',
      'other': '$state',
    });
    return '($_temp0)';
  }

  @override
  String get multipleDecks => '(multiple)';

  @override
  String get moveLeft => 'Move left';

  @override
  String get moveRight => 'Move right';

  @override
  String get hideColumn => 'Hide column';

  @override
  String get columnsEllipsis => 'Columns…';

  @override
  String get columns => 'Columns';

  @override
  String cardKind(String kind) {
    String _temp0 = intl.Intl.selectLogic(kind, {
      'new': 'New',
      'learning': 'Learning',
      'review': 'Review',
      'relearning': 'Relearning',
      'manual': 'Manual',
      'other': '$kind',
    });
    return '$_temp0';
  }

  @override
  String secondsDecimal(String seconds) {
    return '$seconds s';
  }

  @override
  String get cardInfo => 'Card info';

  @override
  String get reviewHistory => 'Review history';

  @override
  String get noReviewsYet => 'No reviews yet.';

  @override
  String get infoAdded => 'Added';

  @override
  String get infoFirstReview => 'First review';

  @override
  String get infoLatestReview => 'Latest review';

  @override
  String get infoDue => 'Due';

  @override
  String get infoInterval => 'Interval';

  @override
  String get infoStability => 'Stability';

  @override
  String get infoDifficulty => 'Difficulty';

  @override
  String get infoRetrievability => 'Retrievability';

  @override
  String get infoReviews => 'Reviews';

  @override
  String get infoLapses => 'Lapses';

  @override
  String get averageTime => 'Average time';

  @override
  String get totalTime => 'Total time';

  @override
  String get infoState => 'State';

  @override
  String get cardType => 'Card type';

  @override
  String get cardId => 'Card ID';

  @override
  String get noteId => 'Note ID';

  @override
  String get historyDate => 'Date';

  @override
  String get historyKind => 'Kind';

  @override
  String get historyRating => 'Rating';

  @override
  String get historyInterval => 'Interval';

  @override
  String get historyTime => 'Time';

  @override
  String deletedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Deleted $_temp0';
  }

  @override
  String setDueDateOfCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Set due date of $_temp0';
  }

  @override
  String resetCardsToNew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Reset $_temp0 to new';
  }

  @override
  String get editAction => 'Edit';

  @override
  String get resetToNew => 'Reset to new';

  @override
  String get unsuspend => 'Unsuspend';

  @override
  String get suspend => 'Suspend';

  @override
  String stateWithSuspended(String state) {
    return '$state (suspended)';
  }

  @override
  String get newCard => 'new card';

  @override
  String countDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString days',
      one: '1 day',
    );
    return '$_temp0';
  }

  @override
  String daysDecimal(String days) {
    return '$days days';
  }

  @override
  String get card => 'Card';

  @override
  String get question => 'Question';

  @override
  String get answer => 'Answer';

  @override
  String changesApplyToOthers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString other cards',
      one: '1 other card',
    );
    return 'Changes also apply to $_temp0 from this note.';
  }

  @override
  String get cardSaved => 'Card saved';

  @override
  String changeCardsDeleted(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0 will be deleted.';
  }

  @override
  String changeCardsCreated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$_temp0 will be created.';
  }

  @override
  String changeCardsBoth(int deleted, int created) {
    final intl.NumberFormat deletedNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String deletedString = deletedNumberFormat.format(deleted);
    final intl.NumberFormat createdNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String createdString = createdNumberFormat.format(created);

    String _temp0 = intl.Intl.pluralLogic(
      deleted,
      locale: localeName,
      other: '$deletedString cards',
      one: '1 card',
    );
    String _temp1 = intl.Intl.pluralLogic(
      created,
      locale: localeName,
      other: '$createdString cards',
      one: '1 card',
    );
    return '$_temp0 will be deleted and $_temp1 will be created.';
  }

  @override
  String get changeNoCards => 'No cards will be deleted or created.';

  @override
  String changeLostFields(String fields) {
    return 'The content of $fields will be lost.';
  }

  @override
  String listAnd(String items, String last) {
    return '$items and $last';
  }

  @override
  String changeNotesTitle(int count, String noteType) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Change $_temp0 to $noteType?';
  }

  @override
  String get change => 'Change';

  @override
  String get changeNoteType => 'Change note type';

  @override
  String currentNoteType(String noteType) {
    return 'Current note type: $noteType';
  }

  @override
  String get newNoteType => 'New note type';

  @override
  String get changeFieldsHelp =>
      'Each new field takes the content of the field you choose.';

  @override
  String get changeCardsHelp =>
      'Each new card type takes the cards of the card type you choose, with their scheduling. Cards of card types not chosen are deleted.';

  @override
  String get changeClozeHelp =>
      'Cards keep their scheduling when the changed note has a card with the same number. Other cards are deleted, and missing ones are created.';

  @override
  String get nothing => '(nothing)';

  @override
  String get saveSearch => 'Save search';

  @override
  String savedSearchAs(String name) {
    return 'Saved the search as \"$name\".';
  }

  @override
  String get savedSearches => 'Saved searches';

  @override
  String get saveCurrentSearch => 'Save current search';

  @override
  String get manageSavedSearches => 'Rename or delete saved searches';

  @override
  String get renameSavedSearch => 'Rename saved search';

  @override
  String deleteQuoted(String name) {
    return 'Delete \"$name\"?';
  }

  @override
  String deleteSavedSearchBody(String query) {
    return 'The saved search $query is removed. Cards are not changed.';
  }

  @override
  String get noSavedSearches => 'No saved searches.';

  @override
  String get findTextPrompt => 'Enter the text to find.';

  @override
  String get counting => 'Counting…';

  @override
  String get noNotesWillChange => 'No notes will change.';

  @override
  String notesWillChange(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return '$_temp0 will change.';
  }

  @override
  String get findAndReplace => 'Find and replace';

  @override
  String get replaceWith => 'Replace with';

  @override
  String get regexGroupsHelp => 'Use \$1 or \${name} for groups';

  @override
  String get inField => 'In';

  @override
  String get allFields => 'All fields';

  @override
  String get selectedNotesOnly => 'Selected notes only';

  @override
  String get treatAsRegex => 'Treat input as regular expression';

  @override
  String get ignoreCase => 'Ignore case';

  @override
  String movedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Moved $_temp0';
  }

  @override
  String get tagsSeparated => 'Tags, separated by spaces';

  @override
  String get removeTags => 'Remove tags';

  @override
  String cardsNotNew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards were',
      one: '1 card was',
    );
    return ' $_temp0 not new and left unchanged.';
  }

  @override
  String get noNewCardsSelected => 'No new cards selected.';

  @override
  String repositionedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Repositioned $_temp0.';
  }

  @override
  String changedNoteTypeOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Changed note type of $_temp0';
  }

  @override
  String get noNotesChanged => 'No notes changed.';

  @override
  String changedNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Changed $_temp0';
  }

  @override
  String get moveToDeck => 'Move to deck';

  @override
  String get clearSelection => 'Clear selection';

  @override
  String countSelected(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString selected';
  }

  @override
  String get selectCards => 'Select cards';

  @override
  String get selectNotes => 'Select notes';

  @override
  String get selectAll => 'Select all';

  @override
  String get selection => 'Selection';

  @override
  String get invertSelection => 'Invert selection';

  @override
  String get bulkActions => 'Bulk actions';

  @override
  String get toggleMark => 'Toggle mark';

  @override
  String get deleteCardsAction => 'Delete cards';

  @override
  String get deleteNotesAction => 'Delete notes';

  @override
  String get browse => 'Browse';

  @override
  String get sidebar => 'Sidebar';

  @override
  String get sort => 'Sort';

  @override
  String get noCardsYet => 'No cards yet.';

  @override
  String get noCardsMatch => 'No cards match your search.';

  @override
  String get noNotesMatch => 'No notes match your search.';

  @override
  String get searchHint => 'Search, e.g. deck:Spanish tag:verb -is:suspended';

  @override
  String onlyNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Only $_temp0';
  }

  @override
  String get showAllCards => 'Show all cards';

  @override
  String get showAllNotes => 'Show all notes';

  @override
  String get modeCards => 'Cards';

  @override
  String get modeNotes => 'Notes';

  @override
  String get oneRowPerCard => 'One row per card';

  @override
  String get oneRowPerNote => 'One row per note';

  @override
  String get selectCardForDetails => 'Select a card to see its details.';

  @override
  String get selectNoteForDetails => 'Select a note to see its details.';

  @override
  String get stateLearning => 'Learning';

  @override
  String get stateReview => 'Review';

  @override
  String get stateDue => 'Due';

  @override
  String get stateSuspendedTitle => 'Suspended';

  @override
  String get renameTag => 'Rename tag';

  @override
  String get renameTagHelper => 'Use :: to move it under another tag';

  @override
  String renamedTagOn(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString notes',
      one: '1 note',
    );
    return 'Renamed the tag on $_temp0';
  }

  @override
  String deleteTagTitle(String tag) {
    return 'Delete the tag “$tag”?';
  }

  @override
  String deleteTagsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return 'Delete $_temp0?';
  }

  @override
  String deleteTagBody(int notes) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    return 'This removes it and its child tags from $_temp0. The notes are kept.';
  }

  @override
  String deleteTagsBody(int notes) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    return 'This removes them and their child tags from $_temp0. The notes are kept.';
  }

  @override
  String removedTagsFrom(int count, int notes) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString notes',
      one: '1 note',
    );
    return 'Removed $_temp0 from $_temp1';
  }

  @override
  String movedTagsTop(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return 'Moved $_temp0 to the top level';
  }

  @override
  String movedTagsUnder(int count, String parent) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return 'Moved $_temp0 under “$parent”';
  }

  @override
  String moveTagsTo(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return 'Move $_temp0 to';
  }

  @override
  String get topLevel => 'Top level';

  @override
  String get noUnusedTags => 'There are no unused tags.';

  @override
  String clearedUnusedTags(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString unused tags',
      one: '1 unused tag',
    );
    return 'Cleared $_temp0.';
  }

  @override
  String movedDeckTop(String deck) {
    return 'Moved “$deck” to the top level';
  }

  @override
  String movedDeckInto(String deck, String parent) {
    return 'Moved “$deck” into “$parent”';
  }

  @override
  String moveDeckTo(String deck) {
    return 'Move “$deck” to';
  }

  @override
  String get addToSearch => 'Add to search';

  @override
  String get addToSearchOr => 'Add to search with OR';

  @override
  String get excludeFromSearchAction => 'Exclude from search';

  @override
  String get renameEllipsis => 'Rename…';

  @override
  String get moveToEllipsis => 'Move to…';

  @override
  String get selectTags => 'Select tags';

  @override
  String get deleteEllipsis => 'Delete…';

  @override
  String get filterHint => 'Filter';

  @override
  String get clearFilter => 'Clear filter';

  @override
  String get cardStates => 'Card states';

  @override
  String get tagActions => 'Tag actions';

  @override
  String get clearUnusedTags => 'Clear unused tags';

  @override
  String get noTagsYet => 'No tags yet.';

  @override
  String tagCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return '$_temp0';
  }

  @override
  String tagsSelected(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString tags',
      one: '1 tag',
    );
    return '$_temp0 selected';
  }

  @override
  String get moveTagsToTooltip => 'Move tags to';

  @override
  String get deleteTags => 'Delete tags';

  @override
  String exportNamed(String name) {
    return 'Export “$name”';
  }

  @override
  String exportCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Export $_temp0';
  }

  @override
  String exportNotesOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Export the notes of $_temp0';
  }

  @override
  String get exportStopped => 'The export stopped unexpectedly.';

  @override
  String exportedFile(String file) {
    return 'Exported $file';
  }

  @override
  String get preparing => 'Preparing…';

  @override
  String renderingCards(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return 'Rendering the cards ($doneString of $total)…';
  }

  @override
  String get writingFile => 'Writing the file…';

  @override
  String addingMedia(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return 'Adding media ($doneString of $total)…';
  }

  @override
  String get packingFile => 'Packing the file…';

  @override
  String get writingCards => 'Writing the cards…';

  @override
  String get exportColpkgHelp =>
      'An Anki collection package (.colpkg) with every deck, note and card, their scheduling, review history and deck presets. Opening it in Anki replaces the collection there.';

  @override
  String get exportApkgHelp =>
      'An Anki deck package (.apkg), which Anki, AnkiDroid and Cardflip can import.';

  @override
  String get exportNotesTextHelp =>
      'One note per line with its fields separated by tabs, HTML kept. Cardflip and Anki can import the file again to update the notes.';

  @override
  String get exportCardsTextHelp =>
      'One card per line with its question and answer as plain text, for printing or reading elsewhere.';

  @override
  String get format => 'Format';

  @override
  String get formatColpkg => 'Anki collection package (.colpkg)';

  @override
  String get formatApkg => 'Anki deck package (.apkg)';

  @override
  String get formatNotesText => 'Notes in plain text (.txt)';

  @override
  String get formatCardsText => 'Cards in plain text (.txt)';

  @override
  String get includeTags => 'Include tags';

  @override
  String get includeDeck => 'Include deck name';

  @override
  String get includeNoteType => 'Include note type name';

  @override
  String get includeGuid => 'Include unique identifier';

  @override
  String get includeGuidHelp =>
      'Lets Anki update these notes when it imports the file again.';

  @override
  String get includeScheduling => 'Include scheduling information';

  @override
  String get includeSchedulingHelp =>
      'Due dates, intervals and review history. Without it, every card is new.';

  @override
  String get includePresets => 'Include deck presets';

  @override
  String get includePresetsHelp => 'The deck options';

  @override
  String get includeMedia => 'Include media';

  @override
  String get includeMediaHelp => 'Images, audio and video';

  @override
  String get legacyAnki => 'Support older Anki versions';

  @override
  String get legacyAnkiHelp => 'For Anki 2.1.49 and earlier. Slower to import.';

  @override
  String get secToday => 'Today';

  @override
  String get secStreaks => 'Streaks';

  @override
  String get secFutureDue => 'Future due';

  @override
  String get secCalendar => 'Calendar';

  @override
  String get secReviews => 'Reviews';

  @override
  String get secReviewTime => 'Review time';

  @override
  String get secCounts => 'Card counts';

  @override
  String get secIntervals => 'Review intervals';

  @override
  String get secStability => 'Stability';

  @override
  String get secDifficulty => 'Difficulty';

  @override
  String get secRetrievability => 'Retrievability';

  @override
  String get secHourly => 'Hourly breakdown';

  @override
  String get secButtons => 'Answer buttons';

  @override
  String get secTrueRetention => 'True retention';

  @override
  String get period1Month => '1 month';

  @override
  String get period3Months => '3 months';

  @override
  String get period1Year => '1 year';

  @override
  String get periodAll => 'All history';

  @override
  String get periodAllShort => 'All';

  @override
  String get noCardsStudiedToday => 'No cards have been studied today.';

  @override
  String studiedToday(int count, String time) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return 'Studied $_temp0 in $time today.';
  }

  @override
  String againCount(String count, String percent) {
    return 'Again count: $count ($percent correct)';
  }

  @override
  String todayKinds(
    String learning,
    String review,
    String relearning,
    String filtered,
  ) {
    return 'Learning: $learning, Review: $review, Relearning: $relearning, Filtered: $filtered';
  }

  @override
  String get today => 'Today';

  @override
  String get tomorrow => 'Tomorrow';

  @override
  String get overdue => 'Overdue';

  @override
  String get dueTomorrowLabel => 'Due tomorrow';

  @override
  String get total => 'Total';

  @override
  String get average => 'Average';

  @override
  String cardsPerDay(String count) {
    return '$count cards/day';
  }

  @override
  String forecastTip(String day, int count, String total) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString cards',
      one: '1 card',
    );
    return '$day: $_temp0\nRunning total: $total';
  }

  @override
  String get runningTotal => 'Running total';

  @override
  String get daysStudiedYear => 'Days studied in the last year';

  @override
  String countOf(String count, String total) {
    return '$count of $total';
  }

  @override
  String get mostReviewsDay => 'Most reviews on one day';

  @override
  String get currentStreak => 'Current streak';

  @override
  String get longestStreak => 'Longest streak';

  @override
  String longestStreakValue(String days, String first, String last) {
    return '$days, from $first to $last';
  }

  @override
  String get studyDaysInPeriod => 'Study days';

  @override
  String get goalDaysMet => 'Days the goal was met';

  @override
  String minutesAmount(String minutes) {
    return '$minutes min';
  }

  @override
  String minutesAxis(String minutes) {
    return '${minutes}m';
  }

  @override
  String get daysStudied => 'Days studied';

  @override
  String get averageDaysStudied => 'Average for days studied';

  @override
  String get averageOverPeriod => 'Average over period';

  @override
  String get kindLearning => 'Learning';

  @override
  String get kindYoung => 'Young';

  @override
  String get kindMature => 'Mature';

  @override
  String get kindRelearning => 'Relearning';

  @override
  String get kindFiltered => 'Filtered';

  @override
  String get kindSuspended => 'Suspended';

  @override
  String get kindBuried => 'Buried';

  @override
  String bucketDays(String days) {
    return '${days}d';
  }

  @override
  String get reviewCards => 'Review cards';

  @override
  String get averageInterval => 'Average interval';

  @override
  String get averageStability => 'Average stability';

  @override
  String get averageDifficulty => 'Average difficulty';

  @override
  String get averageRetrievability => 'Average retrievability';

  @override
  String get estimatedKnowledge => 'Estimated total knowledge';

  @override
  String get noCardsInfo => 'No cards with this information yet.';

  @override
  String get noReviewsPeriod => 'No reviews in this period.';

  @override
  String hourTip(String hours, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return '$hours\n$_temp0';
  }

  @override
  String hourTipCorrect(String percent) {
    return ', $percent correct';
  }

  @override
  String get successRate => 'Success rate';

  @override
  String get noAnswers => 'No answers';

  @override
  String percentCorrect(String percent, String count) {
    return '$percent correct ($count)';
  }

  @override
  String get eachGroupShows => 'Each group shows Again, Hard, Good and Easy.';

  @override
  String get yesterday => 'Yesterday';

  @override
  String get lastWeek => 'Last week';

  @override
  String get lastMonth => 'Last month';

  @override
  String get lastYear => 'Last year';

  @override
  String get allTime => 'All time';

  @override
  String get desiredRetention => 'Desired retention';

  @override
  String calendarTip(String date, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString reviews',
      one: '1 review',
    );
    return '$date: $_temp0';
  }

  @override
  String get statsDeckFallback => 'Deck';

  @override
  String statsSearchScope(String query) {
    return 'Search: $query';
  }

  @override
  String get wholeCollection => 'Whole collection';

  @override
  String savedFile(String file) {
    return 'Saved $file';
  }

  @override
  String get saveAsPdf => 'Save as PDF';

  @override
  String get collectionScope => 'Collection';

  @override
  String get searchScope => 'Search';

  @override
  String get deckWithSubdecks => 'Deck (with its subdecks)';

  @override
  String get statsSearchHint => 'e.g. deck:Spanish tag:verbs';

  @override
  String get showStatistics => 'Show statistics';

  @override
  String statsTitleScope(String scope) {
    return 'Statistics - $scope';
  }

  @override
  String statsPdfName(String scope, String date) {
    return 'Statistics - $scope - $date.pdf';
  }

  @override
  String statsTimeBased(String scope, String period) {
    return '$scope - time-based sections: $period';
  }

  @override
  String createdAt(String date) {
    return 'Created $date';
  }

  @override
  String get hour => 'Hour';

  @override
  String get correct => 'Correct';

  @override
  String get profile => 'Profile';

  @override
  String get profiles => 'Profiles';

  @override
  String get switchProfile => 'Switch profile';

  @override
  String get showProfilesAtStartup => 'Show profile list at startup';

  @override
  String get showProfilesAtStartupHelp =>
      'Choose a profile each time Cardflip starts, instead of opening the last one used.';

  @override
  String get addProfile => 'Add profile';

  @override
  String get addProfileHelp =>
      'A profile has its own cards, media, settings and backups.';

  @override
  String get renameProfile => 'Rename profile';

  @override
  String get deleteProfile => 'Delete profile';

  @override
  String deleteProfileBody(int cards) {
    String _temp0 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cards cards',
      one: '1 card',
      zero: 'no cards',
    );
    return 'This deletes the profile with $_temp0 and all its media, settings and backups. This cannot be undone.';
  }

  @override
  String get deleteProfileBodyUncounted =>
      'This deletes the profile with all its cards, media, settings and backups. This cannot be undone.';

  @override
  String get profileIsOpen => 'Open now';

  @override
  String profileDeleted(String name) {
    return 'Profile “$name” deleted';
  }

  @override
  String get sync => 'Sync';

  @override
  String get syncDeviceName => 'Device name';

  @override
  String get syncDeviceNameHelper =>
      '1 to 40 characters. Your paired devices see this name.';

  @override
  String get syncRenameDevice => 'Rename this device';

  @override
  String get syncCurrentNetwork => 'Current network';

  @override
  String syncWiredNetwork(String subnet) {
    return 'Wired network $subnet';
  }

  @override
  String syncSubnetNetwork(String subnet) {
    return 'Network $subnet';
  }

  @override
  String get syncMobileNetwork => 'Mobile data';

  @override
  String get syncNoNetwork => 'No network';

  @override
  String get syncNetworkTrusted =>
      'Trusted: Cardflip syncs here with your paired devices.';

  @override
  String get syncNetworkNotTrusted =>
      'Not trusted: Cardflip does not sync or announce this device here.';

  @override
  String get syncMobileNeverTrusted => 'Cardflip never syncs over mobile data.';

  @override
  String get trustThisNetwork => 'Trust this network';

  @override
  String get stopTrusting => 'Stop trusting';

  @override
  String get trustNetworkTitle => 'Trust this network?';

  @override
  String get trustNetworkBody =>
      'On a trusted network, Cardflip makes this device visible by its name and syncs with your paired devices. Other devices on the network can see the name, but only devices you paired can exchange cards with it.\n\nTrust only private networks, such as your home Wi-Fi: another network with the same name is trusted too.';

  @override
  String get trustNetworkLocation =>
      'To recognise the network by its Wi-Fi name, Cardflip asks for the location permission next. Cardflip does not use your location. Without the permission, the network is recognised by its address range.';

  @override
  String get trust => 'Trust';

  @override
  String get trustedNetworks => 'Trusted networks';

  @override
  String get noTrustedNetworks => 'No trusted networks yet';

  @override
  String get pairedDevices => 'Paired devices';

  @override
  String get noPairedDevices => 'No paired devices yet';

  @override
  String get deviceReachable => 'Reachable';

  @override
  String get deviceNotReachable => 'Not reachable';

  @override
  String lastSyncedAt(String time) {
    return 'Last synced $time';
  }

  @override
  String get neverSynced => 'Not synced yet';

  @override
  String get unpair => 'Unpair';

  @override
  String unpairTitle(String device) {
    return 'Unpair $device?';
  }

  @override
  String unpairBody(String device) {
    return '$device can no longer sync with this device, and the links of your profiles with it are removed. The cards on both devices stay as they are.';
  }

  @override
  String get pairDevice => 'Pair a device';

  @override
  String get pairDeviceHelp =>
      'Connect another of your devices on this network';

  @override
  String get pairDeviceIntro =>
      'Open “Pair a device” on the other device too, on the same network. Then choose it below, or enter the address it shows.';

  @override
  String visibleForPairing(String time) {
    return 'Visible for pairing for $time';
  }

  @override
  String get notVisibleForPairing => 'Not visible for pairing';

  @override
  String get makeVisible => 'Make visible';

  @override
  String get thisDeviceAddress => 'Address of this device';

  @override
  String get devicesNearby => 'Devices nearby';

  @override
  String get lookingForDevices => 'Looking for devices…';

  @override
  String get enterAddress => 'Address of the other device';

  @override
  String get enterAddressHelper =>
      'As shown on the other device, such as 192.168.1.20:47812';

  @override
  String get pair => 'Pair';

  @override
  String pairingRequestTitle(String device) {
    return 'Pair with $device?';
  }

  @override
  String pairingRequestBody(String device) {
    return '$device wants to pair with this device. Accept only if it is one of your own devices.';
  }

  @override
  String get accept => 'Accept';

  @override
  String get decline => 'Decline';

  @override
  String get confirmCodeTitle => 'Compare the codes';

  @override
  String confirmCodeBody(String device) {
    return 'Check that $device shows the same code:';
  }

  @override
  String get codesMatch => 'Codes match';

  @override
  String get codesDiffer => 'Codes differ';

  @override
  String waitingForConfirmation(String device) {
    return 'Waiting for $device to confirm…';
  }

  @override
  String pairedWith(String device) {
    return 'Paired with $device';
  }

  @override
  String pairingFailed(String device) {
    return 'Pairing with $device failed';
  }

  @override
  String get profileLinks => 'Links of this profile';

  @override
  String get noProfileLinks =>
      'This profile does not sync with another device yet.';

  @override
  String get linkProfile => 'Link this profile';

  @override
  String get linkProfileHelp => 'Sync it with a profile on a paired device';

  @override
  String linkTarget(String profile, String device) {
    return '$profile on $device';
  }

  @override
  String get chooseDevice => 'Choose a device';

  @override
  String chooseRemoteProfile(String device) {
    return 'Profile on $device';
  }

  @override
  String newProfileNamed(String name) {
    return 'New profile “$name”';
  }

  @override
  String get linkNeedsChoice => 'Needs your choice: merge or replace';

  @override
  String get unlink => 'Unlink';

  @override
  String unlinkTitle(String device) {
    return 'Unlink $device?';
  }

  @override
  String unlinkBody(String profile, String device) {
    return 'This profile stops syncing with “$profile” on $device. Both keep their cards.';
  }

  @override
  String get syncSyncing => 'Syncing…';

  @override
  String syncProgress(int received, int sent) {
    return 'Syncing: $received received, $sent sent';
  }

  @override
  String syncedAt(String time) {
    return 'Synced $time';
  }

  @override
  String get syncNotTrusted => 'Not on a trusted network';

  @override
  String get syncNeedsChoice => 'Sync needs your choice';

  @override
  String get syncFailed => 'Sync failed';

  @override
  String get syncAgain => 'Sync again';

  @override
  String get syncNow => 'Sync now';

  @override
  String syncSummary(int received, int sent) {
    return 'Synced: $received received, $sent sent';
  }

  @override
  String syncSummaryMedia(int received, int sent, int media) {
    return 'Synced: $received received, $sent sent, $media media files';
  }

  @override
  String get syncNothingReachable =>
      'No linked device is reachable on this network.';

  @override
  String firstSyncTitle(String device) {
    return 'Sync with $device';
  }

  @override
  String firstSyncBody(String profile, String device) {
    return 'This profile and “$profile” on $device have not synced yet, or one of them was replaced since. How should they be combined? A backup is made before anything changes.';
  }

  @override
  String get firstSyncMerge => 'Merge';

  @override
  String get firstSyncMergeHelp => 'Keep the cards and review history of both.';

  @override
  String firstSyncReplaceOther(String device) {
    return 'Replace the profile on $device';
  }

  @override
  String firstSyncReplaceOtherHelp(String device) {
    return '$device gets exactly the cards and history of this profile.';
  }

  @override
  String get firstSyncReplaceThis => 'Replace this profile';

  @override
  String firstSyncReplaceThisHelp(String profile, String device) {
    return 'This device gets exactly the cards and history of “$profile” on $device.';
  }

  @override
  String get decideLater => 'Later';
}
