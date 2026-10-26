import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_de.dart';
import 'app_localizations_en.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('de'),
    Locale('en'),
  ];

  /// The app's name.
  ///
  /// In en, this message translates to:
  /// **'Cardflip'**
  String get appTitle;

  /// An error without a message.
  ///
  /// In en, this message translates to:
  /// **'Something went wrong.'**
  String get errorUnknown;

  /// An unexpected error, with its technical description.
  ///
  /// In en, this message translates to:
  /// **'Something went wrong: {error}'**
  String errorUnexpected(String error);

  /// Button that reverts the last change.
  ///
  /// In en, this message translates to:
  /// **'Undo'**
  String get undo;

  /// A message from Go (code apkgCollectionDamaged).
  ///
  /// In en, this message translates to:
  /// **'the collection is damaged ({error})'**
  String goApkgCollectionDamaged(String error);

  /// A message from Go (code apkgDatabaseDamaged).
  ///
  /// In en, this message translates to:
  /// **'the collection database is damaged ({error})'**
  String goApkgDatabaseDamaged(String error);

  /// A message from Go (code apkgDecksUnreadable).
  ///
  /// In en, this message translates to:
  /// **'its decks could not be read ({error})'**
  String goApkgDecksUnreadable(String error);

  /// A message from Go (code apkgEmptyNoteType).
  ///
  /// In en, this message translates to:
  /// **'note type \"{name}\" has no fields or templates; unsupported Anki package version'**
  String goApkgEmptyNoteType(String name);

  /// A message from Go (code apkgFileDamaged).
  ///
  /// In en, this message translates to:
  /// **'{file} is damaged ({error})'**
  String goApkgFileDamaged(String file, String error);

  /// A message from Go (code apkgFileError).
  ///
  /// In en, this message translates to:
  /// **'{file}: {error}'**
  String goApkgFileError(String file, String error);

  /// A message from Go (code apkgFileMissing).
  ///
  /// In en, this message translates to:
  /// **'{file} is missing'**
  String goApkgFileMissing(String file);

  /// A message from Go (code apkgMediaDamaged).
  ///
  /// In en, this message translates to:
  /// **'media file {file} is damaged ({error})'**
  String goApkgMediaDamaged(String file, String error);

  /// A message from Go (code apkgMediaListDamaged).
  ///
  /// In en, this message translates to:
  /// **'the media list is damaged ({error})'**
  String goApkgMediaListDamaged(String error);

  /// A message from Go (code apkgMediaSkipped).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 audio/video file was} other{{count} audio/video files were}} skipped, because Cardflip cannot play their format: {files}'**
  String goApkgMediaSkipped(int count, String files);

  /// A message from Go (code apkgMissingNoteType).
  ///
  /// In en, this message translates to:
  /// **'a note uses a missing note type'**
  String get goApkgMissingNoteType;

  /// A message from Go (code apkgNewerVersion).
  ///
  /// In en, this message translates to:
  /// **'it was made by a newer Anki version (package format {version}) that Cardflip cannot read yet'**
  String goApkgNewerVersion(String version);

  /// A message from Go (code apkgNoCollection).
  ///
  /// In en, this message translates to:
  /// **'it contains no Anki collection'**
  String get goApkgNoCollection;

  /// A message from Go (code apkgNotZip).
  ///
  /// In en, this message translates to:
  /// **'the file is damaged or not a zip archive ({error})'**
  String goApkgNotZip(String error);

  /// A message from Go (code apkgNoteTypesUnreadable).
  ///
  /// In en, this message translates to:
  /// **'its note types could not be read ({error})'**
  String goApkgNoteTypesUnreadable(String error);

  /// A message from Go (code apkgUnreadable).
  ///
  /// In en, this message translates to:
  /// **'not a readable Anki package'**
  String get goApkgUnreadable;

  /// A message from Go (code backupDamaged).
  ///
  /// In en, this message translates to:
  /// **'this backup is damaged'**
  String get goBackupDamaged;

  /// A message from Go (code backupFirstFailed).
  ///
  /// In en, this message translates to:
  /// **'the collection could not be backed up first, so nothing was changed: {error}'**
  String goBackupFirstFailed(String error);

  /// A message from Go (code backupGone).
  ///
  /// In en, this message translates to:
  /// **'this backup no longer exists'**
  String get goBackupGone;

  /// A message from Go (code backupIntervalRange).
  ///
  /// In en, this message translates to:
  /// **'The backup interval must be between 0 and {max} minutes.'**
  String goBackupIntervalRange(int max);

  /// A message from Go (code backupNewerFormat).
  ///
  /// In en, this message translates to:
  /// **'this backup was made by a newer version of Cardflip; update Cardflip to restore it'**
  String get goBackupNewerFormat;

  /// A message from Go (code backupNotImportable).
  ///
  /// In en, this message translates to:
  /// **'a backup cannot be imported; restore it from Settings instead'**
  String get goBackupNotImportable;

  /// A message from Go (code backupsKeptRange).
  ///
  /// In en, this message translates to:
  /// **'The number of backups to keep must be between 0 and {max}.'**
  String goBackupsKeptRange(int max);

  /// A message from Go (code browserColumnTwice).
  ///
  /// In en, this message translates to:
  /// **'The browser column \"{column}\" is shown twice.'**
  String goBrowserColumnTwice(String column);

  /// A message from Go (code browserNeedsColumn).
  ///
  /// In en, this message translates to:
  /// **'the browser needs at least one column'**
  String get goBrowserNeedsColumn;

  /// A message from Go (code browserSplitRange).
  ///
  /// In en, this message translates to:
  /// **'The browser split must be between 0 and 1.'**
  String get goBrowserSplitRange;

  /// A message from Go (code cardCountRange).
  ///
  /// In en, this message translates to:
  /// **'The number of cards must be between 1 and {max}.'**
  String goCardCountRange(int max);

  /// A message from Go (code cardGone).
  ///
  /// In en, this message translates to:
  /// **'That card no longer exists.'**
  String get goCardGone;

  /// A message from Go (code cardTypeGone).
  ///
  /// In en, this message translates to:
  /// **'That card type no longer exists.'**
  String get goCardTypeGone;

  /// A message from Go (code cardTypeNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Please enter a card type name.'**
  String get goCardTypeNameEmpty;

  /// A message from Go (code cardTypeNameInvalid).
  ///
  /// In en, this message translates to:
  /// **'A card type name cannot contain : \'{\' \'}\' \" or start with # ^ /.'**
  String get goCardTypeNameInvalid;

  /// A message from Go (code cardTypeNameTaken).
  ///
  /// In en, this message translates to:
  /// **'The note type already has a card type named \"{name}\".'**
  String goCardTypeNameTaken(String name);

  /// A message from Go (code cardTypeOneTarget).
  ///
  /// In en, this message translates to:
  /// **'Card type \"{name}\" can only go to one card type.'**
  String goCardTypeOneTarget(String name);

  /// A message from Go (code changeLeavesNotesEmpty).
  ///
  /// In en, this message translates to:
  /// **'This change would leave {count, plural, =1{1 note} other{{count} notes}} without cards, so it was not made.'**
  String goChangeLeavesNotesEmpty(int count);

  /// A message from Go (code checkCardsWithoutNote).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card without a note was} other{{count} cards without a note were}} removed.'**
  String goCheckCardsWithoutNote(int count);

  /// A message from Go (code checkFieldCount).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note with the wrong number of fields was} other{{count} notes with the wrong number of fields were}} fixed.'**
  String goCheckFieldCount(int count);

  /// A message from Go (code checkFieldCountTrimmed).
  ///
  /// In en, this message translates to:
  /// **'The removed field content is in the deletion log.'**
  String get goCheckFieldCountTrimmed;

  /// A message from Go (code checkInvalidCards).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card for a card type that no longer exists was} other{{count} cards for card types that no longer exist were}} removed.'**
  String goCheckInvalidCards(int count);

  /// A message from Go (code checkMissingCards).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 missing card was} other{{count} missing cards were}} generated.'**
  String goCheckMissingCards(int count);

  /// A message from Go (code checkMissingDeck).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card in a missing deck was} other{{count} cards in missing decks were}} moved to {deck}.'**
  String goCheckMissingDeck(int count, String deck);

  /// A message from Go (code checkNotesWithoutType).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note whose note type no longer exists was} other{{count} notes whose note type no longer exists were}} removed, with {cards, plural, =1{1 card} other{{cards} cards}}. The notes are in the deletion log.'**
  String goCheckNotesWithoutType(int count, int cards);

  /// A message from Go (code checkScheduling).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card with invalid scheduling was} other{{count} cards with invalid scheduling were}} repaired.'**
  String goCheckScheduling(int count);

  /// A message from Go (code checkSchedulingFromHistory).
  ///
  /// In en, this message translates to:
  /// **'Memory states of {count, plural, =1{1 card} other{{count} cards}} were recomputed from the review history.'**
  String goCheckSchedulingFromHistory(int count);

  /// A message from Go (code checkTags).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note with duplicate or invalid tags was} other{{count} notes with duplicate or invalid tags were}} fixed.'**
  String goCheckTags(int count);

  /// A message from Go (code chooseCardTypeFor).
  ///
  /// In en, this message translates to:
  /// **'Choose a card type for each card type of \"{noteType}\".'**
  String goChooseCardTypeFor(String noteType);

  /// A message from Go (code chooseFieldFor).
  ///
  /// In en, this message translates to:
  /// **'Choose a field for each field of \"{noteType}\".'**
  String goChooseFieldFor(String noteType);

  /// A message from Go (code clipboardImageFormat).
  ///
  /// In en, this message translates to:
  /// **'the clipboard image is not a PNG, JPEG or GIF image'**
  String get goClipboardImageFormat;

  /// A message from Go (code clozeOneCardType).
  ///
  /// In en, this message translates to:
  /// **'A cloze note type has exactly one card type.'**
  String get goClozeOneCardType;

  /// A message from Go (code collectionUnreadable).
  ///
  /// In en, this message translates to:
  /// **'Your collection could not be read: {error}'**
  String goCollectionUnreadable(String error);

  /// A message from Go (code csvChooseTypeAndDeck).
  ///
  /// In en, this message translates to:
  /// **'choose a note type and a deck'**
  String get goCsvChooseTypeAndDeck;

  /// A message from Go (code csvExtraFields).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 row} other{{count} rows}} had more fields than their note type; the extra fields were not imported.'**
  String goCsvExtraFields(int count);

  /// A message from Go (code csvFieldOutOfRange).
  ///
  /// In en, this message translates to:
  /// **'column mapped to field {field}, but the note type has {count} fields'**
  String goCsvFieldOutOfRange(int field, int count);

  /// A message from Go (code csvMapAField).
  ///
  /// In en, this message translates to:
  /// **'map at least one column to a field'**
  String get goCsvMapAField;

  /// A message from Go (code csvNoRows).
  ///
  /// In en, this message translates to:
  /// **'the file has no rows'**
  String get goCsvNoRows;

  /// A message from Go (code csvNotUtf8).
  ///
  /// In en, this message translates to:
  /// **'the file is not UTF-8 text; save it as UTF-8 and try again'**
  String get goCsvNotUtf8;

  /// A message from Go (code csvNoteTypeMissing).
  ///
  /// In en, this message translates to:
  /// **'The file\'\'s note type \"{name}\" does not exist; rows without a note type of their own use the one chosen here.'**
  String goCsvNoteTypeMissing(String name);

  /// A message from Go (code csvOneDeckColumn).
  ///
  /// In en, this message translates to:
  /// **'map only one column to the deck'**
  String get goCsvOneDeckColumn;

  /// A message from Go (code csvOneGuidColumn).
  ///
  /// In en, this message translates to:
  /// **'map only one column to the GUID'**
  String get goCsvOneGuidColumn;

  /// A message from Go (code csvOneNoteTypeColumn).
  ///
  /// In en, this message translates to:
  /// **'map only one column to the note type'**
  String get goCsvOneNoteTypeColumn;

  /// A message from Go (code csvOneTagsColumn).
  ///
  /// In en, this message translates to:
  /// **'map only one column to the tags'**
  String get goCsvOneTagsColumn;

  /// A message from Go (code csvSkippedUnknownType).
  ///
  /// In en, this message translates to:
  /// **'Skipped {count, plural, =1{1 row} other{{count} rows}}: there is no note type named \"{name}\".'**
  String goCsvSkippedUnknownType(int count, String name);

  /// A message from Go (code customStudyNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A regular deck is named \"{name}\". Rename it to use custom study.'**
  String goCustomStudyNameTaken(String name);

  /// A message from Go (code customStudyNeedsRegular).
  ///
  /// In en, this message translates to:
  /// **'Custom study needs a regular deck; \"{deck}\" is a filtered deck.'**
  String goCustomStudyNeedsRegular(String deck);

  /// A message from Go (code dailyGoalRange).
  ///
  /// In en, this message translates to:
  /// **'The daily goal must be between 1 and {max}.'**
  String goDailyGoalRange(int max);

  /// A message from Go (code restDaysRange).
  ///
  /// In en, this message translates to:
  /// **'The rest days per week must be between 0 and {max}.'**
  String goRestDaysRange(int max);

  /// A message from Go (code dayCountRange).
  ///
  /// In en, this message translates to:
  /// **'The number of days must be between 1 and {max}.'**
  String goDayCountRange(int max);

  /// A message from Go (code deckGone).
  ///
  /// In en, this message translates to:
  /// **'That deck no longer exists.'**
  String get goDeckGone;

  /// A message from Go (code deckMoveIntoItself).
  ///
  /// In en, this message translates to:
  /// **'A deck cannot be moved into itself or one of its subdecks.'**
  String get goDeckMoveIntoItself;

  /// A message from Go (code deckNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Please enter a deck name.'**
  String get goDeckNameEmpty;

  /// A message from Go (code deckNameEmptyPart).
  ///
  /// In en, this message translates to:
  /// **'Deck name \"{name}\" has an empty part between \'\'::\'\'.'**
  String goDeckNameEmptyPart(String name);

  /// A message from Go (code deckNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A deck named \"{name}\" already exists.'**
  String goDeckNameTaken(String name);

  /// A message from Go (code deckNameTakenByOther).
  ///
  /// In en, this message translates to:
  /// **'The name \"{name}\" is already taken by another deck.'**
  String goDeckNameTakenByOther(String name);

  /// A message from Go (code defaultPresetUndeletable).
  ///
  /// In en, this message translates to:
  /// **'The Default preset cannot be deleted.'**
  String get goDefaultPresetUndeletable;

  /// A message from Go (code desiredRetentionRange).
  ///
  /// In en, this message translates to:
  /// **'Desired retention must be between {min} and {max}.'**
  String goDesiredRetentionRange(double min, double max);

  /// A message from Go (code dueDateHelp).
  ///
  /// In en, this message translates to:
  /// **'Enter days from today: 0 for today, 1 for tomorrow, 60-90 for a random day in that range, with ! to set the interval too.'**
  String get goDueDateHelp;

  /// A message from Go (code dueDateTooFar).
  ///
  /// In en, this message translates to:
  /// **'A due date can be at most {max} days away.'**
  String goDueDateTooFar(int max);

  /// A message from Go (code dueRangeReversed).
  ///
  /// In en, this message translates to:
  /// **'The range {from}-{to} ends before it starts; put the smaller number first.'**
  String goDueRangeReversed(String from, String to);

  /// A message from Go (code fieldGone).
  ///
  /// In en, this message translates to:
  /// **'That field no longer exists.'**
  String get goFieldGone;

  /// A message from Go (code fieldNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Please enter a field name.'**
  String get goFieldNameEmpty;

  /// A message from Go (code fieldNameInvalid).
  ///
  /// In en, this message translates to:
  /// **'A field name cannot contain : \'{\' \'}\' \" or start with # ^ /.'**
  String get goFieldNameInvalid;

  /// A message from Go (code fieldNameTaken).
  ///
  /// In en, this message translates to:
  /// **'The note type already has a field named \"{name}\".'**
  String goFieldNameTaken(String name);

  /// A message from Go (code fileGone).
  ///
  /// In en, this message translates to:
  /// **'the file is no longer available; please choose it again'**
  String get goFileGone;

  /// A message from Go (code fileUnreadable).
  ///
  /// In en, this message translates to:
  /// **'the file could not be read: {error}'**
  String goFileUnreadable(String error);

  /// A message from Go (code filteredFilterCount).
  ///
  /// In en, this message translates to:
  /// **'A filtered deck has one or two filters.'**
  String get goFilteredFilterCount;

  /// A message from Go (code filteredNoMatch).
  ///
  /// In en, this message translates to:
  /// **'No cards matched the search. Suspended and buried cards and cards in other filtered decks are left out.'**
  String get goFilteredNoMatch;

  /// A message from Go (code filteredNoNotes).
  ///
  /// In en, this message translates to:
  /// **'\"{deck}\" is a filtered deck, so notes cannot be added to it.'**
  String goFilteredNoNotes(String deck);

  /// A message from Go (code filteredNoSubdecks).
  ///
  /// In en, this message translates to:
  /// **'\"{deck}\" is a filtered deck, which cannot have subdecks.'**
  String goFilteredNoSubdecks(String deck);

  /// A message from Go (code findTextEmpty).
  ///
  /// In en, this message translates to:
  /// **'Enter the text to find.'**
  String get goFindTextEmpty;

  /// A message from Go (code fsrsParamRange).
  ///
  /// In en, this message translates to:
  /// **'FSRS parameter {index} must be between {min} and {max}.'**
  String goFsrsParamRange(int index, String min, String max);

  /// A message from Go (code fsrsParamsCount).
  ///
  /// In en, this message translates to:
  /// **'FSRS parameters must be {expected} numbers, not {count}.'**
  String goFsrsParamsCount(int expected, int count);

  /// A message from Go (code historicalRetentionRange).
  ///
  /// In en, this message translates to:
  /// **'Historical retention must be between {min} and {max}.'**
  String goHistoricalRetentionRange(double min, double max);

  /// A message from Go (code importNotUpdatedNoCards).
  ///
  /// In en, this message translates to:
  /// **'Not updated, it would produce no cards: {note}'**
  String goImportNotUpdatedNoCards(String note);

  /// A message from Go (code importOtherNoteType).
  ///
  /// In en, this message translates to:
  /// **'A note already exists with a different note type and was not updated: {note}'**
  String goImportOtherNoteType(String note);

  /// A message from Go (code importPresetOptions).
  ///
  /// In en, this message translates to:
  /// **'The options of preset \"{preset}\" could not be imported ({error}); it has the default options.'**
  String goImportPresetOptions(String preset, String error);

  /// A message from Go (code importSkipped).
  ///
  /// In en, this message translates to:
  /// **'Skipped \"{note}\": {error}'**
  String goImportSkipped(String note, String error);

  /// A message from Go (code importSkippedNoCards).
  ///
  /// In en, this message translates to:
  /// **'Skipped, it would produce no cards: \"{note}\"'**
  String goImportSkippedNoCards(String note);

  /// A message from Go (code invalidDate).
  ///
  /// In en, this message translates to:
  /// **'\"{value}\" is not a date like 2026-01-31.'**
  String goInvalidDate(String value);

  /// A message from Go (code invalidRating).
  ///
  /// In en, this message translates to:
  /// **'Invalid rating {rating}.'**
  String goInvalidRating(String rating);

  /// A message from Go (code invalidRegex).
  ///
  /// In en, this message translates to:
  /// **'Invalid regular expression: {error}'**
  String goInvalidRegex(String error);

  /// A message from Go (code lastDeck).
  ///
  /// In en, this message translates to:
  /// **'The last deck cannot be deleted. A collection always has at least one deck.'**
  String get goLastDeck;

  /// A message from Go (code lastNoteType).
  ///
  /// In en, this message translates to:
  /// **'The last note type cannot be deleted. A collection always has at least one note type.'**
  String get goLastNoteType;

  /// A message from Go (code lastPositionTooHigh).
  ///
  /// In en, this message translates to:
  /// **'The last position would be {last}; positions must be at most {max}.'**
  String goLastPositionTooHigh(int last, int max);

  /// A message from Go (code learnAheadRange).
  ///
  /// In en, this message translates to:
  /// **'the learn ahead limit must be between 0 and {max} minutes'**
  String goLearnAheadRange(int max);

  /// A message from Go (code learningStepsRange).
  ///
  /// In en, this message translates to:
  /// **'Learning steps must each be longer than 0 and at most 365 days.'**
  String get goLearningStepsRange;

  /// A message from Go (code learningStepsTooMany).
  ///
  /// In en, this message translates to:
  /// **'Learning steps can have at most {max} steps.'**
  String goLearningStepsTooMany(int max);

  /// A message from Go (code leechThresholdRange).
  ///
  /// In en, this message translates to:
  /// **'Leech threshold must be between 1 and {max} lapses.'**
  String goLeechThresholdRange(int max);

  /// A message from Go (code maxAnswerSecondsRange).
  ///
  /// In en, this message translates to:
  /// **'Maximum answer seconds must be between 1 and {max}.'**
  String goMaxAnswerSecondsRange(int max);

  /// A message from Go (code maxIntervalRange).
  ///
  /// In en, this message translates to:
  /// **'Maximum interval must be between 1 and {max} days.'**
  String goMaxIntervalRange(int max);

  /// A message from Go (code mcpListenFailed).
  ///
  /// In en, this message translates to:
  /// **'The server could not start: {error}'**
  String goMcpListenFailed(String error);

  /// A message from Go (code mcpPortInUse).
  ///
  /// In en, this message translates to:
  /// **'Port {port} is already in use.'**
  String goMcpPortInUse(String port);

  /// A message from Go (code mcpPortRange).
  ///
  /// In en, this message translates to:
  /// **'The port must be between {min} and {max}.'**
  String goMcpPortRange(String min, String max);

  /// A message from Go (code mcpSettingsNotSaved).
  ///
  /// In en, this message translates to:
  /// **'The agent settings could not be saved: {error}'**
  String goMcpSettingsNotSaved(String error);

  /// A message from Go (code mcpSettingsUnreadable).
  ///
  /// In en, this message translates to:
  /// **'The agent settings could not be read: {error}'**
  String goMcpSettingsUnreadable(String error);

  /// A message from Go (code mediaEmptyTrashFailed).
  ///
  /// In en, this message translates to:
  /// **'The trash could not be emptied: {error}'**
  String goMediaEmptyTrashFailed(String error);

  /// A message from Go (code mediaMissing).
  ///
  /// In en, this message translates to:
  /// **'Media file \"{file}\" is missing.'**
  String goMediaMissing(String file);

  /// A message from Go (code mediaRestoreFailed).
  ///
  /// In en, this message translates to:
  /// **'The files could not be restored: {error}'**
  String goMediaRestoreFailed(String error);

  /// A message from Go (code mediaSaveFailed).
  ///
  /// In en, this message translates to:
  /// **'The file could not be saved: {error}'**
  String goMediaSaveFailed(String error);

  /// A message from Go (code mediaTrashFailed).
  ///
  /// In en, this message translates to:
  /// **'The unused files could not be moved to the trash: {error}'**
  String goMediaTrashFailed(String error);

  /// A message from Go (code minIntervalAboveMax).
  ///
  /// In en, this message translates to:
  /// **'Minimum interval ({min} days) must not be above the maximum interval ({max} days).'**
  String goMinIntervalAboveMax(int min, int max);

  /// A message from Go (code minIntervalRange).
  ///
  /// In en, this message translates to:
  /// **'Minimum interval must be between 1 and {max} days.'**
  String goMinIntervalRange(int max);

  /// A message from Go (code moveIntoFiltered).
  ///
  /// In en, this message translates to:
  /// **'Cards cannot be moved into a filtered deck; it gathers its cards with its search.'**
  String get goMoveIntoFiltered;

  /// A message from Go (code needsCardType).
  ///
  /// In en, this message translates to:
  /// **'A note type needs at least one card type.'**
  String get goNeedsCardType;

  /// A message from Go (code needsField).
  ///
  /// In en, this message translates to:
  /// **'A note type needs at least one field.'**
  String get goNeedsField;

  /// A message from Go (code newPerDayRange).
  ///
  /// In en, this message translates to:
  /// **'New cards per day must be between 0 and 9999.'**
  String get goNewPerDayRange;

  /// A message from Go (code noCardsCloze).
  ///
  /// In en, this message translates to:
  /// **'This would produce no cards. Add a cloze deletion such as \'{{c1::text}}\'.'**
  String get goNoCardsCloze;

  /// A message from Go (code noCardsFillInField).
  ///
  /// In en, this message translates to:
  /// **'This would produce no cards. Fill in the {field} field.'**
  String goNoCardsFillInField(String field);

  /// A message from Go (code noCardsFillInFields).
  ///
  /// In en, this message translates to:
  /// **'This would produce no cards. Fill in the {fields} or the {last} field.'**
  String goNoCardsFillInFields(String fields, String last);

  /// A message from Go (code noCardsNoField).
  ///
  /// In en, this message translates to:
  /// **'This would produce no cards: no card\'\'s question uses a field.'**
  String get goNoCardsNoField;

  /// A message from Go (code noCardsSelected).
  ///
  /// In en, this message translates to:
  /// **'No cards are selected.'**
  String get goNoCardsSelected;

  /// A message from Go (code noNotesSelected).
  ///
  /// In en, this message translates to:
  /// **'No notes are selected.'**
  String get goNoNotesSelected;

  /// A message from Go (code noStudySession).
  ///
  /// In en, this message translates to:
  /// **'no study session; start one first'**
  String get goNoStudySession;

  /// A message from Go (code notAutomaticBackup).
  ///
  /// In en, this message translates to:
  /// **'\"{name}\" is not an automatic backup'**
  String goNotAutomaticBackup(String name);

  /// A message from Go (code notBackup).
  ///
  /// In en, this message translates to:
  /// **'this file is not a Cardflip backup'**
  String get goNotBackup;

  /// A message from Go (code notFilteredDeck).
  ///
  /// In en, this message translates to:
  /// **'\"{deck}\" is not a filtered deck.'**
  String goNotFilteredDeck(String deck);

  /// A message from Go (code notPackageOrBackup).
  ///
  /// In en, this message translates to:
  /// **'this file is neither an Anki package nor a Cardflip backup'**
  String get goNotPackageOrBackup;

  /// A message from Go (code notTextFile).
  ///
  /// In en, this message translates to:
  /// **'this file is not a text file'**
  String get goNotTextFile;

  /// A message from Go (code noteFieldCount).
  ///
  /// In en, this message translates to:
  /// **'The note has {count} fields, but note type \"{noteType}\" has {expected}.'**
  String goNoteFieldCount(int count, String noteType, int expected);

  /// A message from Go (code noteGone).
  ///
  /// In en, this message translates to:
  /// **'That note no longer exists.'**
  String get goNoteGone;

  /// A message from Go (code noteTypeGone).
  ///
  /// In en, this message translates to:
  /// **'That note type no longer exists.'**
  String get goNoteTypeGone;

  /// A message from Go (code noteTypeNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Please enter a note type name.'**
  String get goNoteTypeNameEmpty;

  /// A message from Go (code noteTypeNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A note type named \"{name}\" already exists.'**
  String goNoteTypeNameTaken(String name);

  /// A message from Go (code notesWouldHaveNoCards).
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} would have no cards. Map the fields its card types use, or keep a card type.'**
  String goNotesWouldHaveNoCards(int count);

  /// A message from Go (code nothingToRedo).
  ///
  /// In en, this message translates to:
  /// **'There is nothing to redo.'**
  String get goNothingToRedo;

  /// A message from Go (code nothingToReplace).
  ///
  /// In en, this message translates to:
  /// **'the collection is readable; there is nothing to replace'**
  String get goNothingToReplace;

  /// A message from Go (code nothingToUndo).
  ///
  /// In en, this message translates to:
  /// **'There is nothing to undo.'**
  String get goNothingToUndo;

  /// A message from Go (code optimizeSearchInvalid).
  ///
  /// In en, this message translates to:
  /// **'The optimize search is invalid: {error}'**
  String goOptimizeSearchInvalid(String error);

  /// A message from Go (code positionStepRange).
  ///
  /// In en, this message translates to:
  /// **'The step must be between 1 and {max}.'**
  String goPositionStepRange(int max);

  /// A message from Go (code presetGone).
  ///
  /// In en, this message translates to:
  /// **'That preset no longer exists.'**
  String get goPresetGone;

  /// A message from Go (code presetNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Please enter a preset name.'**
  String get goPresetNameEmpty;

  /// A message from Go (code presetNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A preset named \"{name}\" already exists.'**
  String goPresetNameTaken(String name);

  /// A message from Go (code previewDelayRange).
  ///
  /// In en, this message translates to:
  /// **'Preview delays must be between 0 and {max} seconds.'**
  String goPreviewDelayRange(int max);

  /// A message from Go (code profileGone).
  ///
  /// In en, this message translates to:
  /// **'This profile no longer exists.'**
  String get goProfileGone;

  /// A message from Go (code profileNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'A profile needs a name.'**
  String get goProfileNameEmpty;

  /// A message from Go (code profileNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A profile named \"{name}\" already exists.'**
  String goProfileNameTaken(String name);

  /// A message from Go (code profileOnly).
  ///
  /// In en, this message translates to:
  /// **'The only profile cannot be deleted.'**
  String get goProfileOnly;

  /// A message from Go (code profileOpen).
  ///
  /// In en, this message translates to:
  /// **'The open profile cannot be deleted. Open another profile first.'**
  String get goProfileOpen;

  /// A message from Go (code profilesNotSaved).
  ///
  /// In en, this message translates to:
  /// **'The list of profiles could not be saved: {error}'**
  String goProfilesNotSaved(String error);

  /// A message from Go (code profilesUnreadable).
  ///
  /// In en, this message translates to:
  /// **'The list of profiles could not be read: {error}'**
  String goProfilesUnreadable(String error);

  /// A message from Go (code questionFadeRange).
  ///
  /// In en, this message translates to:
  /// **'The question fade must be between 0 and {max} milliseconds.'**
  String goQuestionFadeRange(int max);

  /// A message from Go (code questionTemplateEmpty).
  ///
  /// In en, this message translates to:
  /// **'The question template cannot be empty.'**
  String get goQuestionTemplateEmpty;

  /// A message from Go (code relearningStepsRange).
  ///
  /// In en, this message translates to:
  /// **'Relearning steps must each be longer than 0 and at most 365 days.'**
  String get goRelearningStepsRange;

  /// A message from Go (code relearningStepsTooMany).
  ///
  /// In en, this message translates to:
  /// **'Relearning steps can have at most {max} steps.'**
  String goRelearningStepsTooMany(int max);

  /// A message from Go (code reservedFieldName).
  ///
  /// In en, this message translates to:
  /// **'\"{name}\" is a reserved name: templates use \'{{\'{tag}\'}}\' for a special field. Please choose another name.'**
  String goReservedFieldName(String name, String tag);

  /// A message from Go (code reviewsPerDayRange).
  ///
  /// In en, this message translates to:
  /// **'Maximum reviews per day must be between 0 and 99999.'**
  String get goReviewsPerDayRange;

  /// A message from Go (code rolloverHourRange).
  ///
  /// In en, this message translates to:
  /// **'The day rollover hour must be between 0 and 23.'**
  String get goRolloverHourRange;

  /// A message from Go (code sameNoteType).
  ///
  /// In en, this message translates to:
  /// **'All selected cards must use the same note type.'**
  String get goSameNoteType;

  /// A message from Go (code savedSearchMissing).
  ///
  /// In en, this message translates to:
  /// **'there is no saved search named \"{name}\"'**
  String goSavedSearchMissing(String name);

  /// A message from Go (code savedSearchNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'A saved search needs a name.'**
  String get goSavedSearchNameEmpty;

  /// A message from Go (code savedSearchNameTaken).
  ///
  /// In en, this message translates to:
  /// **'A saved search named \"{name}\" already exists.'**
  String goSavedSearchNameTaken(String name);

  /// A message from Go (code searchCardTypeNumber).
  ///
  /// In en, this message translates to:
  /// **'card type numbers start at 1, e.g. card:2'**
  String get goSearchCardTypeNumber;

  /// A message from Go (code searchEmptyTerm).
  ///
  /// In en, this message translates to:
  /// **'empty search term'**
  String get goSearchEmptyTerm;

  /// A message from Go (code searchErrorAt).
  ///
  /// In en, this message translates to:
  /// **'{message} (character {position})'**
  String goSearchErrorAt(String message, String position);

  /// A message from Go (code searchErrorIn).
  ///
  /// In en, this message translates to:
  /// **'{term}: {message}'**
  String goSearchErrorIn(String term, String message);

  /// A message from Go (code searchInvalidRegex).
  ///
  /// In en, this message translates to:
  /// **'invalid regular expression: {error}'**
  String goSearchInvalidRegex(String error);

  /// A message from Go (code searchNeedsCardType).
  ///
  /// In en, this message translates to:
  /// **'needs a card type name or number, e.g. card:Reverse or card:2'**
  String get goSearchNeedsCardType;

  /// A message from Go (code searchNeedsComparison).
  ///
  /// In en, this message translates to:
  /// **'needs a comparison (=, !=, <, >, <= or >=) with a number, e.g. prop:{property}{example}'**
  String goSearchNeedsComparison(String property, String example);

  /// A message from Go (code searchNeedsDayComparison).
  ///
  /// In en, this message translates to:
  /// **'needs a comparison with a number of days, e.g. prop:{property}=-1 (0 is today, -1 yesterday)'**
  String goSearchNeedsDayComparison(String property);

  /// A message from Go (code searchNeedsDays).
  ///
  /// In en, this message translates to:
  /// **'needs a number of days, e.g. {key}:7'**
  String goSearchNeedsDays(String key);

  /// A message from Go (code searchNeedsDaysToday).
  ///
  /// In en, this message translates to:
  /// **'needs a number of days, e.g. {key}:7 (1 is today)'**
  String goSearchNeedsDaysToday(String key);

  /// A message from Go (code searchNeedsIds).
  ///
  /// In en, this message translates to:
  /// **'needs an ID or a comma-separated list of IDs, e.g. {key}:1712345678901,1712345678902'**
  String goSearchNeedsIds(String key);

  /// A message from Go (code searchNeedsName).
  ///
  /// In en, this message translates to:
  /// **'needs a name, e.g. {key}:Spanish'**
  String goSearchNeedsName(String key);

  /// A message from Go (code searchNeedsRegex).
  ///
  /// In en, this message translates to:
  /// **'needs a regular expression, e.g. re:^\\d+\$'**
  String get goSearchNeedsRegex;

  /// A message from Go (code searchNeedsWord).
  ///
  /// In en, this message translates to:
  /// **'needs a word, e.g. w:dog'**
  String get goSearchNeedsWord;

  /// A message from Go (code searchNoEase).
  ///
  /// In en, this message translates to:
  /// **'Cardflip schedules with FSRS, which has no ease; search the difficulty with prop:d instead, e.g. prop:d>0.7'**
  String get goSearchNoEase;

  /// A message from Go (code searchOperatorAlone).
  ///
  /// In en, this message translates to:
  /// **'\"{operator}\" needs a search term on both sides'**
  String goSearchOperatorAlone(String operator);

  /// A message from Go (code searchParenEmpty).
  ///
  /// In en, this message translates to:
  /// **'empty parentheses'**
  String get goSearchParenEmpty;

  /// A message from Go (code searchParenNotClosed).
  ///
  /// In en, this message translates to:
  /// **'a parenthesis is not closed'**
  String get goSearchParenNotClosed;

  /// A message from Go (code searchParenNotOpened).
  ///
  /// In en, this message translates to:
  /// **'a closing parenthesis has no opening one'**
  String get goSearchParenNotOpened;

  /// A message from Go (code searchRating).
  ///
  /// In en, this message translates to:
  /// **'the rating must be 1 (Again), 2 (Hard), 3 (Good) or 4 (Easy), e.g. rated:7:1'**
  String get goSearchRating;

  /// A message from Go (code searchTooManyDays).
  ///
  /// In en, this message translates to:
  /// **'the number of days must be at most {max}'**
  String goSearchTooManyDays(int max);

  /// A message from Go (code searchUnknownFlag).
  ///
  /// In en, this message translates to:
  /// **'unknown flag; use flag:0 (no flag) to flag:7'**
  String get goSearchUnknownFlag;

  /// A message from Go (code searchUnknownProperty).
  ///
  /// In en, this message translates to:
  /// **'unknown property; use prop: with ivl, due, reps, lapses, pos, s, d, r, rated or resched, e.g. prop:ivl>=30'**
  String get goSearchUnknownProperty;

  /// A message from Go (code searchUnknownState).
  ///
  /// In en, this message translates to:
  /// **'unknown state; use is:new, is:due, is:review, is:learn, is:suspended, is:buried, is:buried-manually or is:buried-sibling'**
  String get goSearchUnknownState;

  /// A message from Go (code secondsAnswerRange).
  ///
  /// In en, this message translates to:
  /// **'Seconds to show answer for must be between 0 and {max}.'**
  String goSecondsAnswerRange(int max);

  /// A message from Go (code secondsQuestionRange).
  ///
  /// In en, this message translates to:
  /// **'Seconds to show question for must be between 0 and {max}.'**
  String goSecondsQuestionRange(int max);

  /// A message from Go (code sectionNotClosed).
  ///
  /// In en, this message translates to:
  /// **'{tag} is not closed'**
  String goSectionNotClosed(String tag);

  /// A message from Go (code sectionNotOpened).
  ///
  /// In en, this message translates to:
  /// **'{closing} has no matching {opening}'**
  String goSectionNotOpened(String closing, String opening);

  /// A message from Go (code sectionsCrossed).
  ///
  /// In en, this message translates to:
  /// **'{closing} closes {outer} before {inner} is closed'**
  String goSectionsCrossed(String closing, String outer, String inner);

  /// A message from Go (code selectCardsToChange).
  ///
  /// In en, this message translates to:
  /// **'Select the cards of the notes to change.'**
  String get goSelectCardsToChange;

  /// A message from Go (code simulationDaysRange).
  ///
  /// In en, this message translates to:
  /// **'Days to simulate must be between 1 and {max}.'**
  String goSimulationDaysRange(int max);

  /// A message from Go (code startPositionRange).
  ///
  /// In en, this message translates to:
  /// **'The start position must be between 0 and {max}.'**
  String goStartPositionRange(int max);

  /// A message from Go (code stepAddField).
  ///
  /// In en, this message translates to:
  /// **'Add field'**
  String get goStepAddField;

  /// A message from Go (code stepAddNote).
  ///
  /// In en, this message translates to:
  /// **'Add note'**
  String get goStepAddNote;

  /// A message from Go (code stepAddNoteType).
  ///
  /// In en, this message translates to:
  /// **'Add note type'**
  String get goStepAddNoteType;

  /// A message from Go (code stepAddNotes).
  ///
  /// In en, this message translates to:
  /// **'Add {count, plural, =1{1 note} other{{count} notes}}'**
  String goStepAddNotes(int count);

  /// A message from Go (code stepAddPreset).
  ///
  /// In en, this message translates to:
  /// **'Add preset'**
  String get goStepAddPreset;

  /// A message from Go (code stepAddTags).
  ///
  /// In en, this message translates to:
  /// **'Add tags'**
  String get goStepAddTags;

  /// A message from Go (code stepAddTemplate).
  ///
  /// In en, this message translates to:
  /// **'Add card type'**
  String get goStepAddTemplate;

  /// A message from Go (code stepAnswer).
  ///
  /// In en, this message translates to:
  /// **'Answer card'**
  String get goStepAnswer;

  /// A message from Go (code stepBuildFiltered).
  ///
  /// In en, this message translates to:
  /// **'Build filtered deck'**
  String get goStepBuildFiltered;

  /// A message from Go (code stepBury).
  ///
  /// In en, this message translates to:
  /// **'Bury {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepBury(int count);

  /// A message from Go (code stepChange).
  ///
  /// In en, this message translates to:
  /// **'Change'**
  String get goStepChange;

  /// A message from Go (code stepChangeNoteType).
  ///
  /// In en, this message translates to:
  /// **'Change note type'**
  String get goStepChangeNoteType;

  /// A message from Go (code stepCheckDatabase).
  ///
  /// In en, this message translates to:
  /// **'Check database'**
  String get goStepCheckDatabase;

  /// A message from Go (code stepCloneNoteType).
  ///
  /// In en, this message translates to:
  /// **'Clone note type'**
  String get goStepCloneNoteType;

  /// A message from Go (code stepClonePreset).
  ///
  /// In en, this message translates to:
  /// **'Clone preset'**
  String get goStepClonePreset;

  /// A message from Go (code stepCreateDeck).
  ///
  /// In en, this message translates to:
  /// **'Create deck'**
  String get goStepCreateDeck;

  /// A message from Go (code stepCustomStudy).
  ///
  /// In en, this message translates to:
  /// **'Custom study'**
  String get goStepCustomStudy;

  /// A message from Go (code stepDeckDescription).
  ///
  /// In en, this message translates to:
  /// **'Edit deck description'**
  String get goStepDeckDescription;

  /// A message from Go (code stepDeckLimits).
  ///
  /// In en, this message translates to:
  /// **'Change deck limits'**
  String get goStepDeckLimits;

  /// A message from Go (code stepDeckOptions).
  ///
  /// In en, this message translates to:
  /// **'Change deck options'**
  String get goStepDeckOptions;

  /// A message from Go (code stepDeckPreset).
  ///
  /// In en, this message translates to:
  /// **'Change preset'**
  String get goStepDeckPreset;

  /// A message from Go (code stepDeleteDeck).
  ///
  /// In en, this message translates to:
  /// **'Delete deck'**
  String get goStepDeleteDeck;

  /// A message from Go (code stepDeleteEmptyCards).
  ///
  /// In en, this message translates to:
  /// **'Delete {count, plural, =1{1 empty card} other{{count} empty cards}}'**
  String goStepDeleteEmptyCards(int count);

  /// A message from Go (code stepDeleteNoteType).
  ///
  /// In en, this message translates to:
  /// **'Delete note type'**
  String get goStepDeleteNoteType;

  /// A message from Go (code stepDeleteNotes).
  ///
  /// In en, this message translates to:
  /// **'Delete {count, plural, =1{1 note} other{{count} notes}}'**
  String goStepDeleteNotes(int count);

  /// A message from Go (code stepDeletePreset).
  ///
  /// In en, this message translates to:
  /// **'Delete preset'**
  String get goStepDeletePreset;

  /// A message from Go (code stepDeleteTags).
  ///
  /// In en, this message translates to:
  /// **'Delete tags'**
  String get goStepDeleteTags;

  /// A message from Go (code stepEmptyFiltered).
  ///
  /// In en, this message translates to:
  /// **'Empty filtered deck'**
  String get goStepEmptyFiltered;

  /// A message from Go (code stepFieldOptions).
  ///
  /// In en, this message translates to:
  /// **'Change field options'**
  String get goStepFieldOptions;

  /// A message from Go (code stepFindReplace).
  ///
  /// In en, this message translates to:
  /// **'Find and replace'**
  String get goStepFindReplace;

  /// A message from Go (code stepFlag).
  ///
  /// In en, this message translates to:
  /// **'Flag {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepFlag(int count);

  /// A message from Go (code stepImport).
  ///
  /// In en, this message translates to:
  /// **'Import'**
  String get goStepImport;

  /// A message from Go (code stepMark).
  ///
  /// In en, this message translates to:
  /// **'Mark {count, plural, =1{1 note} other{{count} notes}}'**
  String goStepMark(int count);

  /// A message from Go (code stepMoveCards).
  ///
  /// In en, this message translates to:
  /// **'Move {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepMoveCards(int count);

  /// A message from Go (code stepMoveField).
  ///
  /// In en, this message translates to:
  /// **'Move field'**
  String get goStepMoveField;

  /// A message from Go (code stepMoveTags).
  ///
  /// In en, this message translates to:
  /// **'Move tags'**
  String get goStepMoveTags;

  /// A message from Go (code stepMoveTemplate).
  ///
  /// In en, this message translates to:
  /// **'Move card type'**
  String get goStepMoveTemplate;

  /// A message from Go (code stepNoteTypeCSS).
  ///
  /// In en, this message translates to:
  /// **'Edit styling'**
  String get goStepNoteTypeCSS;

  /// A message from Go (code stepOptimize).
  ///
  /// In en, this message translates to:
  /// **'Optimize all presets'**
  String get goStepOptimize;

  /// A message from Go (code stepPresetToSubdecks).
  ///
  /// In en, this message translates to:
  /// **'Save to all subdecks'**
  String get goStepPresetToSubdecks;

  /// A message from Go (code stepRemoveField).
  ///
  /// In en, this message translates to:
  /// **'Remove field'**
  String get goStepRemoveField;

  /// A message from Go (code stepRemoveTags).
  ///
  /// In en, this message translates to:
  /// **'Remove tags'**
  String get goStepRemoveTags;

  /// A message from Go (code stepRemoveTemplate).
  ///
  /// In en, this message translates to:
  /// **'Remove card type'**
  String get goStepRemoveTemplate;

  /// A message from Go (code stepRenameDeck).
  ///
  /// In en, this message translates to:
  /// **'Rename deck'**
  String get goStepRenameDeck;

  /// A message from Go (code stepRenameField).
  ///
  /// In en, this message translates to:
  /// **'Rename field'**
  String get goStepRenameField;

  /// A message from Go (code stepRenameNoteType).
  ///
  /// In en, this message translates to:
  /// **'Rename note type'**
  String get goStepRenameNoteType;

  /// A message from Go (code stepRenamePreset).
  ///
  /// In en, this message translates to:
  /// **'Rename preset'**
  String get goStepRenamePreset;

  /// A message from Go (code stepRenameTag).
  ///
  /// In en, this message translates to:
  /// **'Rename tag'**
  String get goStepRenameTag;

  /// A message from Go (code stepReposition).
  ///
  /// In en, this message translates to:
  /// **'Reposition'**
  String get goStepReposition;

  /// A message from Go (code stepResetCards).
  ///
  /// In en, this message translates to:
  /// **'Reset {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepResetCards(int count);

  /// A message from Go (code stepSetDueDate).
  ///
  /// In en, this message translates to:
  /// **'Set due date of {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepSetDueDate(int count);

  /// A message from Go (code stepSortField).
  ///
  /// In en, this message translates to:
  /// **'Change sort field'**
  String get goStepSortField;

  /// A message from Go (code stepSuspend).
  ///
  /// In en, this message translates to:
  /// **'Suspend {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepSuspend(int count);

  /// A message from Go (code stepTodayLimits).
  ///
  /// In en, this message translates to:
  /// **'Custom study'**
  String get goStepTodayLimits;

  /// A message from Go (code stepUnbury).
  ///
  /// In en, this message translates to:
  /// **'Unbury'**
  String get goStepUnbury;

  /// A message from Go (code stepUndoAnswer).
  ///
  /// In en, this message translates to:
  /// **'Undo answer'**
  String get goStepUndoAnswer;

  /// A message from Go (code stepUnflag).
  ///
  /// In en, this message translates to:
  /// **'Remove flag from {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepUnflag(int count);

  /// A message from Go (code stepUnmark).
  ///
  /// In en, this message translates to:
  /// **'Unmark {count, plural, =1{1 note} other{{count} notes}}'**
  String goStepUnmark(int count);

  /// A message from Go (code stepUnsuspend).
  ///
  /// In en, this message translates to:
  /// **'Unsuspend {count, plural, =1{1 card} other{{count} cards}}'**
  String goStepUnsuspend(int count);

  /// A message from Go (code stepUpdateNote).
  ///
  /// In en, this message translates to:
  /// **'Edit note'**
  String get goStepUpdateNote;

  /// A message from Go (code stepUpdateTemplate).
  ///
  /// In en, this message translates to:
  /// **'Edit card type'**
  String get goStepUpdateTemplate;

  /// A message from Go (code storageOpenFailed).
  ///
  /// In en, this message translates to:
  /// **'The collection storage could not be opened: {error}'**
  String goStorageOpenFailed(String error);

  /// A message from Go (code syncAlreadyLinked).
  ///
  /// In en, this message translates to:
  /// **'Profile \"{profile}\" is already linked with a profile on {device}.'**
  String goSyncAlreadyLinked(String profile, String device);

  /// A message from Go (code syncAlreadyLinkedHere).
  ///
  /// In en, this message translates to:
  /// **'This profile is already linked with a profile on that device.'**
  String get goSyncAlreadyLinkedHere;

  /// A message from Go (code syncBadAddress).
  ///
  /// In en, this message translates to:
  /// **'Enter the address and port shown on the other device.'**
  String get goSyncBadAddress;

  /// A message from Go (code syncBusy).
  ///
  /// In en, this message translates to:
  /// **'{device} is syncing with another device. Try again in a moment.'**
  String goSyncBusy(String device);

  /// A message from Go (code syncClocksDifferDays).
  ///
  /// In en, this message translates to:
  /// **'The clocks of this device and {device} differ by {days} days. Set the correct time on both and try again.'**
  String goSyncClocksDifferDays(String device, int days);

  /// A message from Go (code syncClocksDifferHours).
  ///
  /// In en, this message translates to:
  /// **'The clocks of this device and {device} differ by {hours} hours. Set the correct time on both and try again.'**
  String goSyncClocksDifferHours(String device, int hours);

  /// A message from Go (code syncClocksDifferMinutes).
  ///
  /// In en, this message translates to:
  /// **'The clocks of this device and {device} differ by {minutes} minutes. Set the correct time on both and try again.'**
  String goSyncClocksDifferMinutes(String device, int minutes);

  /// A message from Go (code syncConnectFailed).
  ///
  /// In en, this message translates to:
  /// **'{device} could not be reached: {error}'**
  String goSyncConnectFailed(String device, String error);

  /// A message from Go (code syncConnectTimeout).
  ///
  /// In en, this message translates to:
  /// **'{device} did not answer at {address}. If it is on, a firewall on it may block sync: allow incoming TCP connections on port {port} there.'**
  String goSyncConnectTimeout(String device, String address, String port);

  /// A message from Go (code syncDeviceNameLength).
  ///
  /// In en, this message translates to:
  /// **'The device name must be 1 to {max} characters long.'**
  String goSyncDeviceNameLength(int max);

  /// A message from Go (code syncLinkGone).
  ///
  /// In en, this message translates to:
  /// **'That link no longer exists.'**
  String get goSyncLinkGone;

  /// A message from Go (code syncLinksUnreadable).
  ///
  /// In en, this message translates to:
  /// **'The links of this profile could not be read: {error}'**
  String goSyncLinksUnreadable(String error);

  /// A message from Go (code syncListenFailed).
  ///
  /// In en, this message translates to:
  /// **'Sync could not start: {error}'**
  String goSyncListenFailed(String error);

  /// A message from Go (code syncNeedsChoice).
  ///
  /// In en, this message translates to:
  /// **'Choose how to combine this profile with the one on {device}.'**
  String goSyncNeedsChoice(String device);

  /// A message from Go (code syncNoNetwork).
  ///
  /// In en, this message translates to:
  /// **'This device is not on a network that can be trusted.'**
  String get goSyncNoNetwork;

  /// A message from Go (code syncNotPaired).
  ///
  /// In en, this message translates to:
  /// **'That device is not paired.'**
  String get goSyncNotPaired;

  /// A message from Go (code syncNotReachable).
  ///
  /// In en, this message translates to:
  /// **'{device} is not reachable on this network.'**
  String goSyncNotReachable(String device);

  /// A message from Go (code syncNotSetUp).
  ///
  /// In en, this message translates to:
  /// **'Open Settings → Sync first.'**
  String get goSyncNotSetUp;

  /// A message from Go (code syncPairConnectFailed).
  ///
  /// In en, this message translates to:
  /// **'Could not reach the device at {address}: {error}'**
  String goSyncPairConnectFailed(String address, String error);

  /// A message from Go (code syncPairDeclined).
  ///
  /// In en, this message translates to:
  /// **'{device} declined the pairing request.'**
  String goSyncPairDeclined(String device);

  /// A message from Go (code syncPairFailed).
  ///
  /// In en, this message translates to:
  /// **'Pairing with {device} failed: the codes were not confirmed on both devices.'**
  String goSyncPairFailed(String device);

  /// A message from Go (code syncPairNotVisible).
  ///
  /// In en, this message translates to:
  /// **'{device} is not visible for pairing. Open “Pair a device” on it and try again.'**
  String goSyncPairNotVisible(String device);

  /// A message from Go (code syncPairTimeout).
  ///
  /// In en, this message translates to:
  /// **'The device at {address} did not answer. A firewall on it may block sync: allow incoming TCP connections on port {port} there, or start the pairing on that device.'**
  String goSyncPairTimeout(String address, String port);

  /// A message from Go (code syncStateNotSaved).
  ///
  /// In en, this message translates to:
  /// **'The sync settings could not be saved: {error}'**
  String goSyncStateNotSaved(String error);

  /// A message from Go (code syncStateUnreadable).
  ///
  /// In en, this message translates to:
  /// **'The sync settings could not be read: {error}'**
  String goSyncStateUnreadable(String error);

  /// A message from Go (code syncUnknownLink).
  ///
  /// In en, this message translates to:
  /// **'{device} no longer has this link. Link the profiles again.'**
  String goSyncUnknownLink(String device);

  /// A message from Go (code syncUntrusted).
  ///
  /// In en, this message translates to:
  /// **'Trust this network first.'**
  String get goSyncUntrusted;

  /// A message from Go (code syncUpdateOther).
  ///
  /// In en, this message translates to:
  /// **'Update Cardflip on {device} to sync with it.'**
  String goSyncUpdateOther(String device);

  /// A message from Go (code syncUpdateThis).
  ///
  /// In en, this message translates to:
  /// **'Update Cardflip on this device to sync with {device}.'**
  String goSyncUpdateThis(String device);

  /// A message from Go (code tagMoveIntoItself).
  ///
  /// In en, this message translates to:
  /// **'A tag cannot be moved into itself or one of its child tags.'**
  String get goTagMoveIntoItself;

  /// A message from Go (code tagNameEmpty).
  ///
  /// In en, this message translates to:
  /// **'Enter a tag name.'**
  String get goTagNameEmpty;

  /// A message from Go (code tagNameSpaces).
  ///
  /// In en, this message translates to:
  /// **'Tag names cannot contain spaces.'**
  String get goTagNameSpaces;

  /// A message from Go (code tagNoFieldName).
  ///
  /// In en, this message translates to:
  /// **'{tag} has no field name'**
  String goTagNoFieldName(String tag);

  /// A message from Go (code tagNotClosed).
  ///
  /// In en, this message translates to:
  /// **'{tag} has no closing \'}}\''**
  String goTagNotClosed(String tag);

  /// A message from Go (code templateBackMistake).
  ///
  /// In en, this message translates to:
  /// **'In the back template, {problem}.'**
  String goTemplateBackMistake(String problem);

  /// A message from Go (code templateBackUnknownField).
  ///
  /// In en, this message translates to:
  /// **'The back template uses {tag}, but there is no field \"{name}\".'**
  String goTemplateBackUnknownField(String tag, String name);

  /// A message from Go (code templateFrontMistake).
  ///
  /// In en, this message translates to:
  /// **'In the front template, {problem}.'**
  String goTemplateFrontMistake(String problem);

  /// A message from Go (code templateFrontUnknownField).
  ///
  /// In en, this message translates to:
  /// **'The front template uses {tag}, but there is no field \"{name}\".'**
  String goTemplateFrontUnknownField(String tag, String name);

  /// A message from Go (code templateNeedsCloze).
  ///
  /// In en, this message translates to:
  /// **'The front template must use {tag} or another cloze field.'**
  String goTemplateNeedsCloze(String tag);

  /// A message from Go (code templateNoField).
  ///
  /// In en, this message translates to:
  /// **'The front template uses no field, so it would never produce a card.'**
  String get goTemplateNoField;

  /// A message from Go (code timeboxRange).
  ///
  /// In en, this message translates to:
  /// **'The timebox time limit must be between 0 and {max} minutes.'**
  String goTimeboxRange(int max);

  /// A message from Go (code todayIncreaseRange).
  ///
  /// In en, this message translates to:
  /// **'The increase must be between 0 and {max} cards.'**
  String goTodayIncreaseRange(int max);

  /// A message from Go (code uiScaleRange).
  ///
  /// In en, this message translates to:
  /// **'The UI size must be between {min} and {max} % in steps of {step} %.'**
  String goUiScaleRange(int min, int max, int step);

  /// A message from Go (code undoOnlyCardflip).
  ///
  /// In en, this message translates to:
  /// **'Only answers given in Cardflip can be undone.'**
  String get goUndoOnlyCardflip;

  /// A message from Go (code undoOnlyLatest).
  ///
  /// In en, this message translates to:
  /// **'Only the most recent answer to a card can be undone.'**
  String get goUndoOnlyLatest;

  /// A message from Go (code unknownAnswerAction).
  ///
  /// In en, this message translates to:
  /// **'Unknown answer action \"{value}\".'**
  String goUnknownAnswerAction(String value);

  /// A message from Go (code unknownBrowserColumn).
  ///
  /// In en, this message translates to:
  /// **'Unknown browser column \"{value}\".'**
  String goUnknownBrowserColumn(String value);

  /// A message from Go (code unknownBrowserMode).
  ///
  /// In en, this message translates to:
  /// **'Unknown browser mode \"{value}\".'**
  String goUnknownBrowserMode(String value);

  /// A message from Go (code unknownCardState).
  ///
  /// In en, this message translates to:
  /// **'Unknown card state \"{value}\".'**
  String goUnknownCardState(String value);

  /// A message from Go (code unknownCustomStudy).
  ///
  /// In en, this message translates to:
  /// **'Unknown custom study choice \"{value}\".'**
  String goUnknownCustomStudy(String value);

  /// A message from Go (code unknownDailyGoal).
  ///
  /// In en, this message translates to:
  /// **'Unknown daily goal \"{value}\".'**
  String goUnknownDailyGoal(String value);

  /// A message from Go (code unknownEasyDaysLevel).
  ///
  /// In en, this message translates to:
  /// **'Unknown Easy Days level {value}.'**
  String goUnknownEasyDaysLevel(String value);

  /// A message from Go (code unknownFilteredOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown filtered deck order \"{value}\".'**
  String goUnknownFilteredOrder(String value);

  /// A message from Go (code unknownFlag).
  ///
  /// In en, this message translates to:
  /// **'There is no flag {flag}.'**
  String goUnknownFlag(String flag);

  /// A message from Go (code unknownGatherOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown new card gather order \"{value}\".'**
  String goUnknownGatherOrder(String value);

  /// A message from Go (code unknownInsertionOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown insertion order \"{value}\".'**
  String goUnknownInsertionOrder(String value);

  /// A message from Go (code unknownInterdayOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown interday learning/review order \"{value}\".'**
  String goUnknownInterdayOrder(String value);

  /// A message from Go (code unknownLanguage).
  ///
  /// In en, this message translates to:
  /// **'Unknown language \"{value}\".'**
  String goUnknownLanguage(String value);

  /// A message from Go (code unknownLeechAction).
  ///
  /// In en, this message translates to:
  /// **'Unknown leech action \"{value}\".'**
  String goUnknownLeechAction(String value);

  /// A message from Go (code unknownNewReviewOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown new/review order \"{value}\".'**
  String goUnknownNewReviewOrder(String value);

  /// A message from Go (code unknownNewSortOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown new card sort order \"{value}\".'**
  String goUnknownNewSortOrder(String value);

  /// A message from Go (code unknownNoteSort).
  ///
  /// In en, this message translates to:
  /// **'Notes can\'\'t be sorted by \"{value}\".'**
  String goUnknownNoteSort(String value);

  /// A message from Go (code unknownQuestionAction).
  ///
  /// In en, this message translates to:
  /// **'Unknown question action \"{value}\".'**
  String goUnknownQuestionAction(String value);

  /// A message from Go (code unknownReduceMotion).
  ///
  /// In en, this message translates to:
  /// **'Unknown reduce motion choice \"{value}\".'**
  String goUnknownReduceMotion(String value);

  /// A message from Go (code unknownReviewSortOrder).
  ///
  /// In en, this message translates to:
  /// **'Unknown review sort order \"{value}\".'**
  String goUnknownReviewSortOrder(String value);

  /// A message from Go (code unknownSeparator).
  ///
  /// In en, this message translates to:
  /// **'unknown separator \"{value}\"'**
  String goUnknownSeparator(String value);

  /// A message from Go (code unknownStockType).
  ///
  /// In en, this message translates to:
  /// **'Unknown stock note type \"{value}\".'**
  String goUnknownStockType(String value);

  /// A message from Go (code unknownTheme).
  ///
  /// In en, this message translates to:
  /// **'Unknown theme \"{value}\".'**
  String goUnknownTheme(String value);

  /// The date format, as an intl DateFormat pattern.
  ///
  /// In en, this message translates to:
  /// **'yyyy-MM-dd'**
  String get datePattern;

  /// When something is due today, as in "The next card is due at 14:05."
  ///
  /// In en, this message translates to:
  /// **'at {time}'**
  String dueAt(String time);

  /// When something is due tomorrow, as in "The next card is due tomorrow."
  ///
  /// In en, this message translates to:
  /// **'tomorrow'**
  String get dueTomorrow;

  /// When something is due within a week, as in "The next card is due in 3 days."
  ///
  /// In en, this message translates to:
  /// **'in {days, plural, =1{1 day} other{{days} days}}'**
  String dueInDays(int days);

  /// When something is due on a later date, as in "The next card is due on 2026-11-02."
  ///
  /// In en, this message translates to:
  /// **'on {date}'**
  String dueOn(String date);

  /// No description provided for @spentSeconds.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 second} other{{count} seconds}}'**
  String spentSeconds(int count);

  /// No description provided for @spentMinutes.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 minute} other{{count} minutes}}'**
  String spentMinutes(int count);

  /// No description provided for @spentHours.
  ///
  /// In en, this message translates to:
  /// **'{hours} h {minutes} min'**
  String spentHours(int hours, int minutes);

  /// No description provided for @intervalMinutes.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 minute} other{{count} minutes}}'**
  String intervalMinutes(int count);

  /// No description provided for @intervalHours.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 hour} other{{count} hours}}'**
  String intervalHours(int count);

  /// No description provided for @intervalDays.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 day} other{{count} days}}'**
  String intervalDays(int count);

  /// A scheduling interval in months with one decimal, e.g. "2.5 months".
  ///
  /// In en, this message translates to:
  /// **'{months} months'**
  String intervalMonths(String months);

  /// A scheduling interval in years with one decimal, e.g. "1.3 years".
  ///
  /// In en, this message translates to:
  /// **'{years} years'**
  String intervalYears(String years);

  /// A rating button's interval under an hour, short: "<10m".
  ///
  /// In en, this message translates to:
  /// **'<{minutes}m'**
  String intervalLabelMinutes(int minutes);

  /// A rating button's interval in hours, short: "5h", "1.5h".
  ///
  /// In en, this message translates to:
  /// **'{hours}h'**
  String intervalLabelHours(String hours);

  /// A rating button's interval in days, short: "3d".
  ///
  /// In en, this message translates to:
  /// **'{days}d'**
  String intervalLabelDays(int days);

  /// A rating button's interval in months, short: "1.2mo".
  ///
  /// In en, this message translates to:
  /// **'{months}mo'**
  String intervalLabelMonths(String months);

  /// A rating button's interval in years, short: "2y".
  ///
  /// In en, this message translates to:
  /// **'{years}y'**
  String intervalLabelYears(String years);

  /// No description provided for @sizeBytes.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 byte} other{{count} bytes}}'**
  String sizeBytes(int count);

  /// No description provided for @cancel.
  ///
  /// In en, this message translates to:
  /// **'Cancel'**
  String get cancel;

  /// No description provided for @save.
  ///
  /// In en, this message translates to:
  /// **'Save'**
  String get save;

  /// Hint in the box for typing the answer on a card.
  ///
  /// In en, this message translates to:
  /// **'Type the answer'**
  String get typeTheAnswer;

  /// Screen reader label of an image that is not in the collection.
  ///
  /// In en, this message translates to:
  /// **'Missing image'**
  String get missingImage;

  /// Screen reader label of a missing image, with its alt text.
  ///
  /// In en, this message translates to:
  /// **'Missing image: {alt}'**
  String missingImageNamed(String alt);

  /// No description provided for @stop.
  ///
  /// In en, this message translates to:
  /// **'Stop'**
  String get stop;

  /// Tooltip of a play button for an audio or video file.
  ///
  /// In en, this message translates to:
  /// **'Play {file}'**
  String playFile(String file);

  /// Tooltip of a button that reads text aloud.
  ///
  /// In en, this message translates to:
  /// **'Speak'**
  String get speak;

  /// No description provided for @noVoicesInstalled.
  ///
  /// In en, this message translates to:
  /// **'No voices are installed on this device.'**
  String get noVoicesInstalled;

  /// A media file that is not in the collection.
  ///
  /// In en, this message translates to:
  /// **'{file} is missing'**
  String fileMissing(String file);

  /// No description provided for @fileCannotPlay.
  ///
  /// In en, this message translates to:
  /// **'{file} cannot be played on this device'**
  String fileCannotPlay(String file);

  /// Tooltip of a disabled text-to-speech button.
  ///
  /// In en, this message translates to:
  /// **'No voice is installed for {language}'**
  String noVoiceForLanguage(String language);

  /// No description provided for @noVoiceForThisLanguage.
  ///
  /// In en, this message translates to:
  /// **'No voice is installed for this language'**
  String get noVoiceForThisLanguage;

  /// No description provided for @microphoneRefused.
  ///
  /// In en, this message translates to:
  /// **'Cardflip may not use the microphone. Allow microphone access in your system settings to record your voice.'**
  String get microphoneRefused;

  /// Title of the deck picker (the / shortcut).
  ///
  /// In en, this message translates to:
  /// **'Study deck'**
  String get studyDeck;

  /// No description provided for @searchDecks.
  ///
  /// In en, this message translates to:
  /// **'Search decks'**
  String get searchDecks;

  /// No description provided for @noMatchingDecks.
  ///
  /// In en, this message translates to:
  /// **'No matching decks'**
  String get noMatchingDecks;

  /// Snackbar after unburying the cards of a deck.
  ///
  /// In en, this message translates to:
  /// **'Unburied “{deck}”'**
  String unburiedDeck(String deck);

  /// No description provided for @editDescription.
  ///
  /// In en, this message translates to:
  /// **'Edit description'**
  String get editDescription;

  /// No description provided for @tryAgain.
  ///
  /// In en, this message translates to:
  /// **'Try again'**
  String get tryAgain;

  /// No description provided for @countBuriedCards.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 buried card} other{{count} buried cards}}'**
  String countBuriedCards(int count);

  /// Button that makes buried cards available again.
  ///
  /// In en, this message translates to:
  /// **'Unbury'**
  String get unbury;

  /// No description provided for @studyNow.
  ///
  /// In en, this message translates to:
  /// **'Study now'**
  String get studyNow;

  /// No description provided for @options.
  ///
  /// In en, this message translates to:
  /// **'Options'**
  String get options;

  /// Button that gathers a filtered deck's cards again.
  ///
  /// In en, this message translates to:
  /// **'Rebuild'**
  String get rebuild;

  /// Button that returns a filtered deck's cards to their decks.
  ///
  /// In en, this message translates to:
  /// **'Empty'**
  String get emptyDeck;

  /// Label of the count of new cards to study.
  ///
  /// In en, this message translates to:
  /// **'New'**
  String get countNew;

  /// Label of the count of learning cards to study.
  ///
  /// In en, this message translates to:
  /// **'Learning'**
  String get countLearning;

  /// Label of the count of review cards to study.
  ///
  /// In en, this message translates to:
  /// **'To review'**
  String get countToReview;

  /// No description provided for @nothingDueNow.
  ///
  /// In en, this message translates to:
  /// **'Nothing is due right now.'**
  String get nothingDueNow;

  /// {when} is like "at 14:05", "tomorrow", "in 3 days" or "on 2026-11-02".
  ///
  /// In en, this message translates to:
  /// **'The next card is due {when}.'**
  String nextCardDue(String when);

  /// No description provided for @deckHasNoCards.
  ///
  /// In en, this message translates to:
  /// **'This deck has no cards to study. Add cards or import some to get started.'**
  String get deckHasNoCards;

  /// Title of the dialog that edits a deck's description.
  ///
  /// In en, this message translates to:
  /// **'Description'**
  String get description;

  /// No description provided for @descriptionHelper.
  ///
  /// In en, this message translates to:
  /// **'Plain text, or HTML such as <b>bold</b> and links'**
  String get descriptionHelper;

  /// Error of a number field.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number from 1 to {max}.'**
  String enterWholeNumber(int max);

  /// How many more cards today's limit could be increased by.
  ///
  /// In en, this message translates to:
  /// **'{count} more available'**
  String moreAvailable(int count);

  /// No description provided for @increaseBy.
  ///
  /// In en, this message translates to:
  /// **'Increase by'**
  String get increaseBy;

  /// Unit after a number field, as in "Increase by [10] cards".
  ///
  /// In en, this message translates to:
  /// **'cards'**
  String get unitCards;

  /// Unit after a number field, as in "Last [7] days".
  ///
  /// In en, this message translates to:
  /// **'days'**
  String get unitDays;

  /// No description provided for @increaseHelper.
  ///
  /// In en, this message translates to:
  /// **'For today only. The deck options stay as they are.'**
  String get increaseHelper;

  /// Label of a choice of which cards to study.
  ///
  /// In en, this message translates to:
  /// **'Cards'**
  String get cards;

  /// No description provided for @newCardsOnly.
  ///
  /// In en, this message translates to:
  /// **'New cards only'**
  String get newCardsOnly;

  /// No description provided for @dueCardsOnly.
  ///
  /// In en, this message translates to:
  /// **'Due cards only'**
  String get dueCardsOnly;

  /// No description provided for @allCardsRandom.
  ///
  /// In en, this message translates to:
  /// **'All cards in random order (no rescheduling)'**
  String get allCardsRandom;

  /// No description provided for @onlyWithTags.
  ///
  /// In en, this message translates to:
  /// **'Only with these tags (optional)'**
  String get onlyWithTags;

  /// No description provided for @onlyWithTagsHelper.
  ///
  /// In en, this message translates to:
  /// **'Cards with any of them; separate tags by spaces'**
  String get onlyWithTagsHelper;

  /// Tooltip of the button that starts selecting rows.
  ///
  /// In en, this message translates to:
  /// **'Select'**
  String get select;

  /// As in "Next [7] days", to review ahead.
  ///
  /// In en, this message translates to:
  /// **'Next'**
  String get next;

  /// As in "Last [7] days", to review forgotten cards.
  ///
  /// In en, this message translates to:
  /// **'Last'**
  String get last;

  /// {deck} is the name of the filtered deck, Custom Study Session.
  ///
  /// In en, this message translates to:
  /// **'Studied in the filtered deck “{deck}”, which replaces the previous session.'**
  String customStudySessionHelper(String deck);

  /// Title of the custom study dialog.
  ///
  /// In en, this message translates to:
  /// **'Custom study: {deck}'**
  String customStudyTitle(String deck);

  /// No description provided for @increaseNewLimit.
  ///
  /// In en, this message translates to:
  /// **'Increase today\'\'s new card limit'**
  String get increaseNewLimit;

  /// No description provided for @increaseReviewLimit.
  ///
  /// In en, this message translates to:
  /// **'Increase today\'\'s review card limit'**
  String get increaseReviewLimit;

  /// No description provided for @reviewForgotten.
  ///
  /// In en, this message translates to:
  /// **'Review forgotten cards'**
  String get reviewForgotten;

  /// No description provided for @reviewAhead.
  ///
  /// In en, this message translates to:
  /// **'Review ahead'**
  String get reviewAhead;

  /// No description provided for @previewNewCards.
  ///
  /// In en, this message translates to:
  /// **'Preview new cards'**
  String get previewNewCards;

  /// No description provided for @studyByStateOrTag.
  ///
  /// In en, this message translates to:
  /// **'Study by card state or tag'**
  String get studyByStateOrTag;

  /// No description provided for @ok.
  ///
  /// In en, this message translates to:
  /// **'OK'**
  String get ok;

  /// Title of the dialog that creates a deck.
  ///
  /// In en, this message translates to:
  /// **'New deck'**
  String get newDeck;

  /// Label of a name field.
  ///
  /// In en, this message translates to:
  /// **'Name'**
  String get name;

  /// No description provided for @newDeckHelper.
  ///
  /// In en, this message translates to:
  /// **'Use :: for subdecks, e.g. Languages::Spanish'**
  String get newDeckHelper;

  /// Button that creates something.
  ///
  /// In en, this message translates to:
  /// **'Create'**
  String get create;

  /// Title of the deck list, and its navigation destination.
  ///
  /// In en, this message translates to:
  /// **'Decks'**
  String get decks;

  /// No description provided for @statistics.
  ///
  /// In en, this message translates to:
  /// **'Statistics'**
  String get statistics;

  /// No description provided for @createDeck.
  ///
  /// In en, this message translates to:
  /// **'Create deck'**
  String get createDeck;

  /// No description provided for @createFilteredDeck.
  ///
  /// In en, this message translates to:
  /// **'Create filtered deck'**
  String get createFilteredDeck;

  /// No description provided for @collectionEmpty.
  ///
  /// In en, this message translates to:
  /// **'Your collection is empty'**
  String get collectionEmpty;

  /// No description provided for @collectionEmptyHint.
  ///
  /// In en, this message translates to:
  /// **'Add your first card, or import cards from a text file (.csv, .tsv or .txt), an Anki deck (.apkg) or a whole Anki collection (.colpkg).'**
  String get collectionEmptyHint;

  /// No description provided for @addACard.
  ///
  /// In en, this message translates to:
  /// **'Add a card'**
  String get addACard;

  /// No description provided for @importAFile.
  ///
  /// In en, this message translates to:
  /// **'Import a file'**
  String get importAFile;

  /// No description provided for @renameDeck.
  ///
  /// In en, this message translates to:
  /// **'Rename deck'**
  String get renameDeck;

  /// No description provided for @renameDeckHelper.
  ///
  /// In en, this message translates to:
  /// **'Use :: to move it into another deck'**
  String get renameDeckHelper;

  /// Button or menu item that renames something.
  ///
  /// In en, this message translates to:
  /// **'Rename'**
  String get rename;

  /// Title of a confirmation to delete something named.
  ///
  /// In en, this message translates to:
  /// **'Delete “{name}”?'**
  String deleteNamedTitle(String name);

  /// {cards} is like "12 cards"; {inFiltered} and {subdecks} are empty or the texts of deleteDeckInFiltered and deleteDeckSubdecks.
  ///
  /// In en, this message translates to:
  /// **'This deletes {cards}{inFiltered}{subdecks}. You can undo this until you close Cardflip.'**
  String deleteDeckBody(String cards, String inFiltered, String subdecks);

  /// Part of deleteDeckBody, starting with a space.
  ///
  /// In en, this message translates to:
  /// **' ({count} of them now in filtered decks)'**
  String deleteDeckInFiltered(int count);

  /// Part of deleteDeckBody, starting with a space.
  ///
  /// In en, this message translates to:
  /// **' and {count, plural, =1{1 subdeck} other{{count} subdecks}}'**
  String deleteDeckSubdecks(int count);

  /// No description provided for @countCards.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}}'**
  String countCards(int count);

  /// Button or menu item that deletes something.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get delete;

  /// Snackbar after deleting something named.
  ///
  /// In en, this message translates to:
  /// **'Deleted “{name}”'**
  String deletedNamed(String name);

  /// No description provided for @deletedDeckReturned.
  ///
  /// In en, this message translates to:
  /// **'Deleted “{deck}”; {count, plural, =1{1 card} other{{count} cards}} returned to their home decks'**
  String deletedDeckReturned(String deck, int count);

  /// No description provided for @expand.
  ///
  /// In en, this message translates to:
  /// **'Expand'**
  String get expand;

  /// No description provided for @collapse.
  ///
  /// In en, this message translates to:
  /// **'Collapse'**
  String get collapse;

  /// No description provided for @filteredDeck.
  ///
  /// In en, this message translates to:
  /// **'Filtered deck'**
  String get filteredDeck;

  /// No description provided for @deckActions.
  ///
  /// In en, this message translates to:
  /// **'Deck actions'**
  String get deckActions;

  /// Menu item that starts studying a deck.
  ///
  /// In en, this message translates to:
  /// **'Study'**
  String get study;

  /// Menu item that adds a card to a deck.
  ///
  /// In en, this message translates to:
  /// **'Add card'**
  String get addCard;

  /// No description provided for @customStudy.
  ///
  /// In en, this message translates to:
  /// **'Custom study'**
  String get customStudy;

  /// Menu item or button that exports.
  ///
  /// In en, this message translates to:
  /// **'Export'**
  String get export;

  /// Tooltip of the count of new cards due.
  ///
  /// In en, this message translates to:
  /// **'{count} new'**
  String dueNewTooltip(int count);

  /// Tooltip of the count of learning cards due.
  ///
  /// In en, this message translates to:
  /// **'{count} learning'**
  String dueLearningTooltip(int count);

  /// Tooltip of the count of review cards due.
  ///
  /// In en, this message translates to:
  /// **'{count} to review'**
  String dueReviewTooltip(int count);

  /// No description provided for @orderOldestSeen.
  ///
  /// In en, this message translates to:
  /// **'Oldest seen first'**
  String get orderOldestSeen;

  /// No description provided for @orderRandom.
  ///
  /// In en, this message translates to:
  /// **'Random'**
  String get orderRandom;

  /// No description provided for @orderIntervalsAscending.
  ///
  /// In en, this message translates to:
  /// **'Increasing intervals'**
  String get orderIntervalsAscending;

  /// No description provided for @orderIntervalsDescending.
  ///
  /// In en, this message translates to:
  /// **'Decreasing intervals'**
  String get orderIntervalsDescending;

  /// No description provided for @orderMostLapses.
  ///
  /// In en, this message translates to:
  /// **'Most lapses'**
  String get orderMostLapses;

  /// No description provided for @orderAdded.
  ///
  /// In en, this message translates to:
  /// **'Order added'**
  String get orderAdded;

  /// No description provided for @orderDue.
  ///
  /// In en, this message translates to:
  /// **'Order due'**
  String get orderDue;

  /// No description provided for @orderLatestAdded.
  ///
  /// In en, this message translates to:
  /// **'Latest added first'**
  String get orderLatestAdded;

  /// No description provided for @orderRetrievability.
  ///
  /// In en, this message translates to:
  /// **'Lowest retrievability first'**
  String get orderRetrievability;

  /// Snackbar after building a filtered deck.
  ///
  /// In en, this message translates to:
  /// **'“{deck}” holds {count, plural, =1{1 card} other{{count} cards}}'**
  String filteredDeckHolds(String deck, int count);

  /// No description provided for @rebuiltDeck.
  ///
  /// In en, this message translates to:
  /// **'Rebuilt “{deck}”: {count, plural, =1{1 card} other{{count} cards}}'**
  String rebuiltDeck(String deck, int count);

  /// No description provided for @emptiedDeck.
  ///
  /// In en, this message translates to:
  /// **'Emptied “{deck}”'**
  String emptiedDeck(String deck);

  /// {which} is first or second.
  ///
  /// In en, this message translates to:
  /// **'Enter a number of cards from 1 to 99999 for the {which, select, first{first} other{second}} filter.'**
  String filterLimitError(String which);

  /// No description provided for @previewDelaysError.
  ///
  /// In en, this message translates to:
  /// **'Enter preview delays from 0 to 86400 seconds.'**
  String get previewDelaysError;

  /// No description provided for @search.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get search;

  /// No description provided for @filterSearchHelper.
  ///
  /// In en, this message translates to:
  /// **'As in the browser, e.g. deck:Spanish is:due'**
  String get filterSearchHelper;

  /// Label of the maximum number of cards a filter gathers.
  ///
  /// In en, this message translates to:
  /// **'Limit to'**
  String get limitTo;

  /// No description provided for @cardsSelectedBy.
  ///
  /// In en, this message translates to:
  /// **'Cards selected by'**
  String get cardsSelectedBy;

  /// Button that creates a filtered deck and gathers its cards.
  ///
  /// In en, this message translates to:
  /// **'Build'**
  String get build;

  /// No description provided for @newFilteredDeck.
  ///
  /// In en, this message translates to:
  /// **'New filtered deck'**
  String get newFilteredDeck;

  /// No description provided for @filteredDeckOptions.
  ///
  /// In en, this message translates to:
  /// **'Filtered deck options'**
  String get filteredDeckOptions;

  /// Heading of a filtered deck's search options.
  ///
  /// In en, this message translates to:
  /// **'Filter'**
  String get filter;

  /// No description provided for @secondFilter.
  ///
  /// In en, this message translates to:
  /// **'Second filter'**
  String get secondFilter;

  /// No description provided for @secondFilterHelper.
  ///
  /// In en, this message translates to:
  /// **'Adds the cards of another search after the first'**
  String get secondFilterHelper;

  /// No description provided for @reschedule.
  ///
  /// In en, this message translates to:
  /// **'Reschedule cards based on my answers in this deck'**
  String get reschedule;

  /// No description provided for @rescheduleOn.
  ///
  /// In en, this message translates to:
  /// **'Answers count as normal reviews'**
  String get rescheduleOn;

  /// No description provided for @rescheduleOff.
  ///
  /// In en, this message translates to:
  /// **'Preview: answers do not change when cards are due'**
  String get rescheduleOff;

  /// No description provided for @previewDelays.
  ///
  /// In en, this message translates to:
  /// **'Show the card again after (seconds; 0 returns it to its home deck)'**
  String get previewDelays;

  /// Rating button.
  ///
  /// In en, this message translates to:
  /// **'Again'**
  String get again;

  /// Rating button.
  ///
  /// In en, this message translates to:
  /// **'Hard'**
  String get hard;

  /// Rating button.
  ///
  /// In en, this message translates to:
  /// **'Good'**
  String get good;

  /// Rating button.
  ///
  /// In en, this message translates to:
  /// **'Easy'**
  String get easy;

  /// Unit after a number of seconds.
  ///
  /// In en, this message translates to:
  /// **'s'**
  String get unitSecondsShort;

  /// No description provided for @filteredDeckNote.
  ///
  /// In en, this message translates to:
  /// **'Suspended and buried cards and cards in other filtered decks are left out. Emptying or deleting the deck returns its cards to their home decks.'**
  String get filteredDeckNote;

  /// No description provided for @simDaysError.
  ///
  /// In en, this message translates to:
  /// **'Enter between 1 and 3650 days to simulate.'**
  String get simDaysError;

  /// No description provided for @simLimitsError.
  ///
  /// In en, this message translates to:
  /// **'Enter whole numbers for the daily limits.'**
  String get simLimitsError;

  /// {retention} is a percentage, e.g. 90%.
  ///
  /// In en, this message translates to:
  /// **'Use {retention} desired retention?'**
  String useRetentionTitle(String retention);

  /// No description provided for @useRetentionBody.
  ///
  /// In en, this message translates to:
  /// **'The deck options get this desired retention. It is saved when you save the options.'**
  String get useRetentionBody;

  /// Button that applies a suggested value.
  ///
  /// In en, this message translates to:
  /// **'Use'**
  String get use;

  /// Title of the FSRS simulator.
  ///
  /// In en, this message translates to:
  /// **'Simulate {deck}'**
  String simulateDeck(String deck);

  /// No description provided for @simulatorIntro.
  ///
  /// In en, this message translates to:
  /// **'Expected reviews and memorised cards of this deck and its subdecks, starting from their cards today. Answer times and ratings come from your review history. Nothing is changed.'**
  String get simulatorIntro;

  /// No description provided for @daysToSimulate.
  ///
  /// In en, this message translates to:
  /// **'Days to simulate'**
  String get daysToSimulate;

  /// No description provided for @newCardsPerDayShort.
  ///
  /// In en, this message translates to:
  /// **'New cards/day'**
  String get newCardsPerDayShort;

  /// No description provided for @maxReviewsPerDayShort.
  ///
  /// In en, this message translates to:
  /// **'Maximum reviews/day'**
  String get maxReviewsPerDayShort;

  /// No description provided for @desiredRetentionValue.
  ///
  /// In en, this message translates to:
  /// **'Desired retention: {retention}'**
  String desiredRetentionValue(String retention);

  /// No description provided for @compareRetention.
  ///
  /// In en, this message translates to:
  /// **'Compare with another retention'**
  String get compareRetention;

  /// No description provided for @comparedRetentionValue.
  ///
  /// In en, this message translates to:
  /// **'Compared retention: {retention}'**
  String comparedRetentionValue(String retention);

  /// Button that runs the simulation.
  ///
  /// In en, this message translates to:
  /// **'Simulate'**
  String get simulate;

  /// No description provided for @helpMeDecide.
  ///
  /// In en, this message translates to:
  /// **'Help me decide'**
  String get helpMeDecide;

  /// No description provided for @suggestedRetention.
  ///
  /// In en, this message translates to:
  /// **'Suggested desired retention: {retention}'**
  String suggestedRetention(String retention);

  /// No description provided for @suggestedRetentionHelp.
  ///
  /// In en, this message translates to:
  /// **'The retention with the least review time per memorised card at the end of the simulation. Lower costs less time but you remember less.'**
  String get suggestedRetentionHelp;

  /// No description provided for @secondsPerMemorisedCard.
  ///
  /// In en, this message translates to:
  /// **'Seconds per memorised card'**
  String get secondsPerMemorisedCard;

  /// Button that applies a suggested value, e.g. "Use 87%".
  ///
  /// In en, this message translates to:
  /// **'Use {value}'**
  String useValue(String value);

  /// A day of a simulation, as an axis label.
  ///
  /// In en, this message translates to:
  /// **'Day {day}'**
  String dayNumber(int day);

  /// No description provided for @reviewsPerDay.
  ///
  /// In en, this message translates to:
  /// **'Reviews per day'**
  String get reviewsPerDay;

  /// No description provided for @memorisedCards.
  ///
  /// In en, this message translates to:
  /// **'Memorised cards'**
  String get memorisedCards;

  /// Heading of a simulation's totals, e.g. "At 90%".
  ///
  /// In en, this message translates to:
  /// **'At {retention}'**
  String atRetention(String retention);

  /// No description provided for @simReviewTotals.
  ///
  /// In en, this message translates to:
  /// **'{reviews, plural, =1{1 review} other{{reviews} reviews}} ({perDay} a day), {newCards, plural, =1{1 new card} other{{newCards} new cards}} learned'**
  String simReviewTotals(int reviews, String perDay, int newCards);

  /// {total} and {perDay} are durations like "1 h 5 min".
  ///
  /// In en, this message translates to:
  /// **'Time: {total} ({perDay} a day)'**
  String simTime(String total, String perDay);

  /// No description provided for @simMemorisedAtEnd.
  ///
  /// In en, this message translates to:
  /// **'Memorised at the end: {count, plural, =1{1 card} other{{count} cards}}'**
  String simMemorisedAtEnd(int count);

  /// No description provided for @readingDeck.
  ///
  /// In en, this message translates to:
  /// **'Reading the deck…'**
  String get readingDeck;

  /// {percent} is like 40%.
  ///
  /// In en, this message translates to:
  /// **'Simulating… {percent}'**
  String simulatingPercent(String percent);

  /// No description provided for @simulating.
  ///
  /// In en, this message translates to:
  /// **'Simulating…'**
  String get simulating;

  /// No description provided for @enterDaysError.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number of days from 1 to 36500.'**
  String get enterDaysError;

  /// No description provided for @notAboveMaxInterval.
  ///
  /// In en, this message translates to:
  /// **'Must not be above the maximum interval.'**
  String get notAboveMaxInterval;

  /// Error of a field of seconds.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number of seconds from {min} to {max}.'**
  String enterSecondsError(int min, int max);

  /// No description provided for @leechThresholdError.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number for the leech threshold.'**
  String get leechThresholdError;

  /// No description provided for @correctSteps.
  ///
  /// In en, this message translates to:
  /// **'Correct the steps first.'**
  String get correctSteps;

  /// No description provided for @correctIntervals.
  ///
  /// In en, this message translates to:
  /// **'Correct the intervals first.'**
  String get correctIntervals;

  /// No description provided for @correctMaxAnswer.
  ///
  /// In en, this message translates to:
  /// **'Correct the maximum answer seconds first.'**
  String get correctMaxAnswer;

  /// No description provided for @correctAutoAdvance.
  ///
  /// In en, this message translates to:
  /// **'Correct the auto advance seconds first.'**
  String get correctAutoAdvance;

  /// No description provided for @correctFsrs.
  ///
  /// In en, this message translates to:
  /// **'Correct the FSRS parameters first.'**
  String get correctFsrs;

  /// No description provided for @deckLimitsError.
  ///
  /// In en, this message translates to:
  /// **'Enter whole numbers for the limits of this deck.'**
  String get deckLimitsError;

  /// No description provided for @rescheduleTitle.
  ///
  /// In en, this message translates to:
  /// **'Reschedule cards?'**
  String get rescheduleTitle;

  /// No description provided for @rescheduleBody.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 review card will get a new due date from its review history} other{{count} review cards will get new due dates from their review history}}, with the new desired retention and FSRS parameters.'**
  String rescheduleBody(int count);

  /// Button that gives cards new due dates.
  ///
  /// In en, this message translates to:
  /// **'Reschedule'**
  String get rescheduleAction;

  /// No description provided for @optimizeAllTitle.
  ///
  /// In en, this message translates to:
  /// **'Optimize all presets?'**
  String get optimizeAllTitle;

  /// No description provided for @optimizeAllBody.
  ///
  /// In en, this message translates to:
  /// **'Each preset gets FSRS parameters fitted to its reviews, where they fit better than its current ones.'**
  String get optimizeAllBody;

  /// Added to optimizeAllBody; starts with a space.
  ///
  /// In en, this message translates to:
  /// **' Up to {count, plural, =1{1 review card} other{{count} review cards}} of presets that reschedule cards on change will get new due dates.'**
  String optimizeAllReschedule(int count);

  /// Button that fits FSRS parameters to the review history.
  ///
  /// In en, this message translates to:
  /// **'Optimize'**
  String get optimize;

  /// No description provided for @optimizeAllResult.
  ///
  /// In en, this message translates to:
  /// **'{applied, plural, =0{No preset} =1{1 preset} other{{applied} presets}} got new parameters.'**
  String optimizeAllResult(int applied);

  /// Added to optimizeAllResult; starts with a space.
  ///
  /// In en, this message translates to:
  /// **' {count, plural, =1{1 preset has} other{{count} presets have}} too few reviews to optimize.'**
  String optimizeAllFew(int count);

  /// Progress of Optimize all, before the stage, e.g. "Preset 2 of 3: Optimizing… 40%".
  ///
  /// In en, this message translates to:
  /// **'Preset {number} of {total}: '**
  String optimizePreset(int number, int total);

  /// No description provided for @optimizeCancelled.
  ///
  /// In en, this message translates to:
  /// **'Optimizing was cancelled. Nothing changed.'**
  String get optimizeCancelled;

  /// No description provided for @ignoreReviewedBefore.
  ///
  /// In en, this message translates to:
  /// **'Ignore cards reviewed before'**
  String get ignoreReviewedBefore;

  /// No description provided for @discardChangesTitle.
  ///
  /// In en, this message translates to:
  /// **'Discard changes?'**
  String get discardChangesTitle;

  /// No description provided for @presetNotSaved.
  ///
  /// In en, this message translates to:
  /// **'Your changes to the preset \"{preset}\" are not saved.'**
  String presetNotSaved(String preset);

  /// Button that drops unsaved changes.
  ///
  /// In en, this message translates to:
  /// **'Discard'**
  String get discard;

  /// No description provided for @addPreset.
  ///
  /// In en, this message translates to:
  /// **'Add preset'**
  String get addPreset;

  /// No description provided for @addPresetHelper.
  ///
  /// In en, this message translates to:
  /// **'Starts with the default options'**
  String get addPresetHelper;

  /// Button that adds something.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get add;

  /// No description provided for @clonePreset.
  ///
  /// In en, this message translates to:
  /// **'Clone preset'**
  String get clonePreset;

  /// The suggested name of a cloned preset.
  ///
  /// In en, this message translates to:
  /// **'{name} copy'**
  String presetCopyName(String name);

  /// No description provided for @clonePresetHelper.
  ///
  /// In en, this message translates to:
  /// **'Starts with the options shown'**
  String get clonePresetHelper;

  /// Button that copies something.
  ///
  /// In en, this message translates to:
  /// **'Clone'**
  String get clone;

  /// No description provided for @renamePreset.
  ///
  /// In en, this message translates to:
  /// **'Rename preset'**
  String get renamePreset;

  /// No description provided for @deletePresetTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete preset?'**
  String get deletePresetTitle;

  /// No description provided for @presetUnused.
  ///
  /// In en, this message translates to:
  /// **'The preset \"{preset}\" is not used by any deck.'**
  String presetUnused(String preset);

  /// No description provided for @presetUsedBy.
  ///
  /// In en, this message translates to:
  /// **'The preset \"{preset}\" is used by {count, plural, =1{1 deck. It uses} other{{count} decks. They use}} {fallback} instead.'**
  String presetUsedBy(String preset, int count, String fallback);

  /// No description provided for @allSubdecksUsePreset.
  ///
  /// In en, this message translates to:
  /// **'All subdecks use the preset \"{preset}\" now.'**
  String allSubdecksUsePreset(String preset);

  /// Title of the deck options screen.
  ///
  /// In en, this message translates to:
  /// **'Options for {deck}'**
  String optionsFor(String deck);

  /// No description provided for @preset.
  ///
  /// In en, this message translates to:
  /// **'Preset'**
  String get preset;

  /// No description provided for @presetActions.
  ///
  /// In en, this message translates to:
  /// **'Preset actions'**
  String get presetActions;

  /// No description provided for @deletePreset.
  ///
  /// In en, this message translates to:
  /// **'Delete preset'**
  String get deletePreset;

  /// No description provided for @saveToAllSubdecks.
  ///
  /// In en, this message translates to:
  /// **'Save to all subdecks'**
  String get saveToAllSubdecks;

  /// No description provided for @usedByDecksShared.
  ///
  /// In en, this message translates to:
  /// **'Used by {count, plural, =1{1 deck} other{{count} decks}}. Changes apply to all of them.'**
  String usedByDecksShared(int count);

  /// No description provided for @usedByDecks.
  ///
  /// In en, this message translates to:
  /// **'Used by {count, plural, =1{1 deck} other{{count} decks}}.'**
  String usedByDecks(int count);

  /// No description provided for @dailyLimits.
  ///
  /// In en, this message translates to:
  /// **'Daily limits'**
  String get dailyLimits;

  /// No description provided for @newCardsPerDay.
  ///
  /// In en, this message translates to:
  /// **'New cards per day'**
  String get newCardsPerDay;

  /// No description provided for @maxReviewsPerDay.
  ///
  /// In en, this message translates to:
  /// **'Maximum reviews per day'**
  String get maxReviewsPerDay;

  /// No description provided for @thisDeckOnly.
  ///
  /// In en, this message translates to:
  /// **'This deck only'**
  String get thisDeckOnly;

  /// No description provided for @thisDeckOnlyHelp.
  ///
  /// In en, this message translates to:
  /// **'Limits for this deck that replace the preset\'\'s. Leave empty to use the preset\'\'s.'**
  String get thisDeckOnlyHelp;

  /// No description provided for @enterWholeNumberShort.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number'**
  String get enterWholeNumberShort;

  /// No description provided for @newIgnoresReviewLimit.
  ///
  /// In en, this message translates to:
  /// **'New cards ignore review limit'**
  String get newIgnoresReviewLimit;

  /// No description provided for @newIgnoresReviewLimitHelp.
  ///
  /// In en, this message translates to:
  /// **'When off, new cards are only shown while the review limit has room left.'**
  String get newIgnoresReviewLimitHelp;

  /// No description provided for @limitsStartFromTop.
  ///
  /// In en, this message translates to:
  /// **'Limits start from top'**
  String get limitsStartFromTop;

  /// No description provided for @limitsStartFromTopHelp.
  ///
  /// In en, this message translates to:
  /// **'When studying this deck, the limits of its parent decks apply too.'**
  String get limitsStartFromTopHelp;

  /// No description provided for @learningSteps.
  ///
  /// In en, this message translates to:
  /// **'Learning steps'**
  String get learningSteps;

  /// No description provided for @learningStepsHelp.
  ///
  /// In en, this message translates to:
  /// **'Delays before a new card graduates, e.g. 1m 10m 1d (s, m, h, d). Leave empty to let FSRS decide.'**
  String get learningStepsHelp;

  /// No description provided for @relearningSteps.
  ///
  /// In en, this message translates to:
  /// **'Relearning steps'**
  String get relearningSteps;

  /// No description provided for @relearningStepsHelp.
  ///
  /// In en, this message translates to:
  /// **'Delays after a review card is forgotten, e.g. 10m. Leave empty to let FSRS decide.'**
  String get relearningStepsHelp;

  /// No description provided for @insertionOrder.
  ///
  /// In en, this message translates to:
  /// **'Insertion order'**
  String get insertionOrder;

  /// No description provided for @insertionOrderHelp.
  ///
  /// In en, this message translates to:
  /// **'Where added cards go among new cards. Changing it reorders the deck\'\'s new cards.'**
  String get insertionOrderHelp;

  /// No description provided for @insertSequential.
  ///
  /// In en, this message translates to:
  /// **'Sequential (oldest cards first)'**
  String get insertSequential;

  /// No description provided for @displayOrder.
  ///
  /// In en, this message translates to:
  /// **'Display order'**
  String get displayOrder;

  /// No description provided for @newGatherOrder.
  ///
  /// In en, this message translates to:
  /// **'New card gather order'**
  String get newGatherOrder;

  /// No description provided for @newGatherOrderHelp.
  ///
  /// In en, this message translates to:
  /// **'Which new cards fill today\'\'s limit.'**
  String get newGatherOrderHelp;

  /// No description provided for @gatherDeck.
  ///
  /// In en, this message translates to:
  /// **'Deck'**
  String get gatherDeck;

  /// No description provided for @gatherDeckRandomNotes.
  ///
  /// In en, this message translates to:
  /// **'Deck, then random notes'**
  String get gatherDeckRandomNotes;

  /// No description provided for @gatherAscending.
  ///
  /// In en, this message translates to:
  /// **'Ascending position'**
  String get gatherAscending;

  /// No description provided for @gatherDescending.
  ///
  /// In en, this message translates to:
  /// **'Descending position'**
  String get gatherDescending;

  /// No description provided for @gatherRandomNotes.
  ///
  /// In en, this message translates to:
  /// **'Random notes'**
  String get gatherRandomNotes;

  /// No description provided for @gatherRandomCards.
  ///
  /// In en, this message translates to:
  /// **'Random cards'**
  String get gatherRandomCards;

  /// No description provided for @newSortOrder.
  ///
  /// In en, this message translates to:
  /// **'New card sort order'**
  String get newSortOrder;

  /// No description provided for @newSortOrderHelp.
  ///
  /// In en, this message translates to:
  /// **'The order in which today\'\'s new cards are shown.'**
  String get newSortOrderHelp;

  /// No description provided for @sortGathered.
  ///
  /// In en, this message translates to:
  /// **'Order gathered'**
  String get sortGathered;

  /// No description provided for @sortTypeThenGathered.
  ///
  /// In en, this message translates to:
  /// **'Card type, then order gathered'**
  String get sortTypeThenGathered;

  /// No description provided for @sortTypeThenRandom.
  ///
  /// In en, this message translates to:
  /// **'Card type, then random'**
  String get sortTypeThenRandom;

  /// No description provided for @sortRandomNoteThenType.
  ///
  /// In en, this message translates to:
  /// **'Random note, then card type'**
  String get sortRandomNoteThenType;

  /// No description provided for @newReviewOrder.
  ///
  /// In en, this message translates to:
  /// **'New/review order'**
  String get newReviewOrder;

  /// No description provided for @mixWithReviews.
  ///
  /// In en, this message translates to:
  /// **'Mix with reviews'**
  String get mixWithReviews;

  /// No description provided for @showBeforeReviews.
  ///
  /// In en, this message translates to:
  /// **'Show before reviews'**
  String get showBeforeReviews;

  /// No description provided for @showAfterReviews.
  ///
  /// In en, this message translates to:
  /// **'Show after reviews'**
  String get showAfterReviews;

  /// No description provided for @interdayOrder.
  ///
  /// In en, this message translates to:
  /// **'Interday learning/review order'**
  String get interdayOrder;

  /// No description provided for @interdayOrderHelp.
  ///
  /// In en, this message translates to:
  /// **'Where cards with a step of a day or more are shown.'**
  String get interdayOrderHelp;

  /// No description provided for @reviewSortOrder.
  ///
  /// In en, this message translates to:
  /// **'Review sort order'**
  String get reviewSortOrder;

  /// No description provided for @reviewSortOrderHelp.
  ///
  /// In en, this message translates to:
  /// **'Also decides which reviews fill the daily limit.'**
  String get reviewSortOrderHelp;

  /// No description provided for @reviewDueRandom.
  ///
  /// In en, this message translates to:
  /// **'Due date, then random'**
  String get reviewDueRandom;

  /// No description provided for @reviewDueDeck.
  ///
  /// In en, this message translates to:
  /// **'Due date, then deck'**
  String get reviewDueDeck;

  /// No description provided for @reviewDeckDue.
  ///
  /// In en, this message translates to:
  /// **'Deck, then due date'**
  String get reviewDeckDue;

  /// No description provided for @reviewIntervalsAscending.
  ///
  /// In en, this message translates to:
  /// **'Ascending intervals'**
  String get reviewIntervalsAscending;

  /// No description provided for @reviewIntervalsDescending.
  ///
  /// In en, this message translates to:
  /// **'Descending intervals'**
  String get reviewIntervalsDescending;

  /// No description provided for @reviewDifficultyAscending.
  ///
  /// In en, this message translates to:
  /// **'Ascending difficulty'**
  String get reviewDifficultyAscending;

  /// No description provided for @reviewDifficultyDescending.
  ///
  /// In en, this message translates to:
  /// **'Descending difficulty'**
  String get reviewDifficultyDescending;

  /// No description provided for @reviewRetrievabilityAscending.
  ///
  /// In en, this message translates to:
  /// **'Ascending retrievability'**
  String get reviewRetrievabilityAscending;

  /// No description provided for @reviewRelativeOverdueness.
  ///
  /// In en, this message translates to:
  /// **'Relative overdueness'**
  String get reviewRelativeOverdueness;

  /// No description provided for @burying.
  ///
  /// In en, this message translates to:
  /// **'Burying'**
  String get burying;

  /// No description provided for @buryingHelp.
  ///
  /// In en, this message translates to:
  /// **'Once a card of a note is studied, its other cards of these kinds wait until the next day.'**
  String get buryingHelp;

  /// No description provided for @buryNew.
  ///
  /// In en, this message translates to:
  /// **'Bury new siblings'**
  String get buryNew;

  /// No description provided for @buryReview.
  ///
  /// In en, this message translates to:
  /// **'Bury review siblings'**
  String get buryReview;

  /// No description provided for @buryInterday.
  ///
  /// In en, this message translates to:
  /// **'Bury interday learning siblings'**
  String get buryInterday;

  /// No description provided for @maxIntervalDays.
  ///
  /// In en, this message translates to:
  /// **'Maximum interval (days)'**
  String get maxIntervalDays;

  /// No description provided for @maxIntervalHelp.
  ///
  /// In en, this message translates to:
  /// **'No review is scheduled further away than this.'**
  String get maxIntervalHelp;

  /// No description provided for @minIntervalDays.
  ///
  /// In en, this message translates to:
  /// **'Minimum interval (days)'**
  String get minIntervalDays;

  /// No description provided for @minIntervalHelp.
  ///
  /// In en, this message translates to:
  /// **'The shortest interval a forgotten card gets once it is relearned.'**
  String get minIntervalHelp;

  /// No description provided for @leechThreshold.
  ///
  /// In en, this message translates to:
  /// **'Leech threshold'**
  String get leechThreshold;

  /// No description provided for @leechThresholdHelp.
  ///
  /// In en, this message translates to:
  /// **'Lapses after which a card is a leech (1 to 99). It is flagged again every half that many lapses.'**
  String get leechThresholdHelp;

  /// No description provided for @leechAction.
  ///
  /// In en, this message translates to:
  /// **'Leech action'**
  String get leechAction;

  /// No description provided for @leechTagOnly.
  ///
  /// In en, this message translates to:
  /// **'Tag only'**
  String get leechTagOnly;

  /// No description provided for @leechSuspend.
  ///
  /// In en, this message translates to:
  /// **'Suspend card and tag'**
  String get leechSuspend;

  /// No description provided for @retentionHelp.
  ///
  /// In en, this message translates to:
  /// **'Higher retention means you forget less but review more often.'**
  String get retentionHelp;

  /// No description provided for @disableAutoplay.
  ///
  /// In en, this message translates to:
  /// **'Don\'\'t play audio automatically'**
  String get disableAutoplay;

  /// No description provided for @skipQuestionOnReplay.
  ///
  /// In en, this message translates to:
  /// **'Skip question when replaying answer'**
  String get skipQuestionOnReplay;

  /// Heading of the answer timer options.
  ///
  /// In en, this message translates to:
  /// **'Timer'**
  String get timer;

  /// No description provided for @maxAnswerSeconds.
  ///
  /// In en, this message translates to:
  /// **'Maximum answer seconds'**
  String get maxAnswerSeconds;

  /// No description provided for @maxAnswerSecondsHelp.
  ///
  /// In en, this message translates to:
  /// **'Longer answers are recorded as this many seconds, so a break does not count as answer time.'**
  String get maxAnswerSecondsHelp;

  /// No description provided for @showTimer.
  ///
  /// In en, this message translates to:
  /// **'Show answer timer'**
  String get showTimer;

  /// No description provided for @showTimerHelp.
  ///
  /// In en, this message translates to:
  /// **'Shows the time spent on the card while studying. It does not change what is recorded.'**
  String get showTimerHelp;

  /// No description provided for @stopTimerOnAnswer.
  ///
  /// In en, this message translates to:
  /// **'Stop timer on answer'**
  String get stopTimerOnAnswer;

  /// No description provided for @stopTimerOnAnswerHelp.
  ///
  /// In en, this message translates to:
  /// **'Stops the timer when the answer is shown.'**
  String get stopTimerOnAnswerHelp;

  /// No description provided for @autoAdvance.
  ///
  /// In en, this message translates to:
  /// **'Auto advance'**
  String get autoAdvance;

  /// No description provided for @autoAdvanceHelp.
  ///
  /// In en, this message translates to:
  /// **'Turn auto advance on while studying from the menu or with Shift+A. 0 seconds turns it off for that side.'**
  String get autoAdvanceHelp;

  /// No description provided for @secondsToShowQuestion.
  ///
  /// In en, this message translates to:
  /// **'Seconds to show question for'**
  String get secondsToShowQuestion;

  /// No description provided for @secondsToShowAnswer.
  ///
  /// In en, this message translates to:
  /// **'Seconds to show answer for'**
  String get secondsToShowAnswer;

  /// No description provided for @questionAction.
  ///
  /// In en, this message translates to:
  /// **'Question action'**
  String get questionAction;

  /// No description provided for @showAnswer.
  ///
  /// In en, this message translates to:
  /// **'Show answer'**
  String get showAnswer;

  /// No description provided for @showReminder.
  ///
  /// In en, this message translates to:
  /// **'Show reminder'**
  String get showReminder;

  /// No description provided for @answerAction.
  ///
  /// In en, this message translates to:
  /// **'Answer action'**
  String get answerAction;

  /// No description provided for @buryCard.
  ///
  /// In en, this message translates to:
  /// **'Bury card'**
  String get buryCard;

  /// No description provided for @answerAgain.
  ///
  /// In en, this message translates to:
  /// **'Answer Again'**
  String get answerAgain;

  /// No description provided for @answerHard.
  ///
  /// In en, this message translates to:
  /// **'Answer Hard'**
  String get answerHard;

  /// No description provided for @answerGood.
  ///
  /// In en, this message translates to:
  /// **'Answer Good'**
  String get answerGood;

  /// Heading of the FSRS scheduler options; the name of the algorithm.
  ///
  /// In en, this message translates to:
  /// **'FSRS'**
  String get fsrs;

  /// No description provided for @fsrsParameters.
  ///
  /// In en, this message translates to:
  /// **'FSRS parameters'**
  String get fsrsParameters;

  /// No description provided for @fsrsParametersHelp.
  ///
  /// In en, this message translates to:
  /// **'Fitted to your reviews by Optimize. Leave empty for the default parameters.'**
  String get fsrsParametersHelp;

  /// No description provided for @optimizeSearch.
  ///
  /// In en, this message translates to:
  /// **'Optimize search'**
  String get optimizeSearch;

  /// No description provided for @optimizeSearchHelp.
  ///
  /// In en, this message translates to:
  /// **'The cards whose reviews Optimize learns from. Leave empty for this deck and its subdecks.'**
  String get optimizeSearchHelp;

  /// No description provided for @notSet.
  ///
  /// In en, this message translates to:
  /// **'Not set'**
  String get notSet;

  /// No description provided for @clearDate.
  ///
  /// In en, this message translates to:
  /// **'Clear date'**
  String get clearDate;

  /// No description provided for @historicalRetentionValue.
  ///
  /// In en, this message translates to:
  /// **'Historical retention: {retention}'**
  String historicalRetentionValue(String retention);

  /// No description provided for @historicalRetentionHelp.
  ///
  /// In en, this message translates to:
  /// **'How much you remembered in reviews missing from the history, e.g. of cards studied before they were imported.'**
  String get historicalRetentionHelp;

  /// No description provided for @healthCheck.
  ///
  /// In en, this message translates to:
  /// **'Health check'**
  String get healthCheck;

  /// No description provided for @healthCheckHelp.
  ///
  /// In en, this message translates to:
  /// **'Checks that the new parameters are reliable before using them. Takes longer.'**
  String get healthCheckHelp;

  /// No description provided for @rescheduleOnChange.
  ///
  /// In en, this message translates to:
  /// **'Reschedule cards on change'**
  String get rescheduleOnChange;

  /// No description provided for @rescheduleOnChangeHelp.
  ///
  /// In en, this message translates to:
  /// **'When the desired retention or the parameters change, review cards get new due dates. Otherwise only future answers use them.'**
  String get rescheduleOnChangeHelp;

  /// No description provided for @optimizeAllPresets.
  ///
  /// In en, this message translates to:
  /// **'Optimize all presets'**
  String get optimizeAllPresets;

  /// No description provided for @easyDays.
  ///
  /// In en, this message translates to:
  /// **'Easy Days'**
  String get easyDays;

  /// No description provided for @easyDaysHelp.
  ///
  /// In en, this message translates to:
  /// **'Days with fewer reviews. Reviews move to nearby days, by at most the few days intervals already vary by; a Minimum day only gets cards that cannot go elsewhere.'**
  String get easyDaysHelp;

  /// No description provided for @notANumber.
  ///
  /// In en, this message translates to:
  /// **'\"{text}\" is not a number.'**
  String notANumber(String text);

  /// No description provided for @fsrsCountError.
  ///
  /// In en, this message translates to:
  /// **'Enter {count} numbers, or leave empty for the default parameters.'**
  String fsrsCountError(int count);

  /// No description provided for @starting.
  ///
  /// In en, this message translates to:
  /// **'Starting…'**
  String get starting;

  /// No description provided for @readingReviewHistory.
  ///
  /// In en, this message translates to:
  /// **'Reading the review history…'**
  String get readingReviewHistory;

  /// No description provided for @optimizingPercent.
  ///
  /// In en, this message translates to:
  /// **'Optimizing… {percent}'**
  String optimizingPercent(String percent);

  /// No description provided for @checkingParametersPercent.
  ///
  /// In en, this message translates to:
  /// **'Checking the parameters… {percent}'**
  String checkingParametersPercent(String percent);

  /// No description provided for @optimizing.
  ///
  /// In en, this message translates to:
  /// **'Optimizing…'**
  String get optimizing;

  /// No description provided for @countReviews.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 review} other{{count} reviews}}'**
  String countReviews(int count);

  /// No description provided for @notEnoughReviews.
  ///
  /// In en, this message translates to:
  /// **'Optimizing needs at least {min} reviews, but there {count, plural, =1{is only 1 review} other{are only {count} reviews}} to learn from. The parameters are unchanged.'**
  String notEnoughReviews(int min, int count);

  /// No description provided for @newParametersFitted.
  ///
  /// In en, this message translates to:
  /// **'New parameters fitted to {count, plural, =1{1 review} other{{count} reviews}}. Save to use them.'**
  String newParametersFitted(int count);

  /// No description provided for @healthCheckFailed.
  ///
  /// In en, this message translates to:
  /// **'The health check found the new parameters unreliable: the reviews may be too few or too inconsistent. The parameters are unchanged.'**
  String get healthCheckFailed;

  /// No description provided for @parametersAlreadyBest.
  ///
  /// In en, this message translates to:
  /// **'The current parameters already fit your {count, plural, =1{1 review} other{{count} reviews}} best. They are unchanged.'**
  String parametersAlreadyBest(int count);

  /// Easy Days level: the usual number of reviews.
  ///
  /// In en, this message translates to:
  /// **'Normal'**
  String get easyNormal;

  /// Easy Days level: fewer reviews.
  ///
  /// In en, this message translates to:
  /// **'Reduced'**
  String get easyReduced;

  /// Easy Days level: only reviews that cannot go elsewhere.
  ///
  /// In en, this message translates to:
  /// **'Minimum'**
  String get easyMinimum;

  /// No description provided for @notADelay.
  ///
  /// In en, this message translates to:
  /// **'\"{text}\" is not a delay. Write a number and a unit (s, m, h or d), e.g. 10m.'**
  String notADelay(String text);

  /// No description provided for @stepTooShort.
  ///
  /// In en, this message translates to:
  /// **'Each step must be longer than 0.'**
  String get stepTooShort;

  /// No description provided for @stepTooLong.
  ///
  /// In en, this message translates to:
  /// **'A step can be at most 365 days.'**
  String get stepTooLong;

  /// No description provided for @tooManySteps.
  ///
  /// In en, this message translates to:
  /// **'There can be at most 20 steps.'**
  String get tooManySteps;

  /// No description provided for @leechSuspended.
  ///
  /// In en, this message translates to:
  /// **'The card is a leech and was suspended.'**
  String get leechSuspended;

  /// No description provided for @leech.
  ///
  /// In en, this message translates to:
  /// **'The card is a leech.'**
  String get leech;

  /// No description provided for @autoAdvanceOff.
  ///
  /// In en, this message translates to:
  /// **'Auto advance is on, but this card\'\'s deck shows both sides for 0 seconds, which turns it off. Set the seconds in the deck options.'**
  String get autoAdvanceOff;

  /// No description provided for @timeToShowAnswer.
  ///
  /// In en, this message translates to:
  /// **'Time to show the answer.'**
  String get timeToShowAnswer;

  /// No description provided for @timeToAnswer.
  ///
  /// In en, this message translates to:
  /// **'Time to answer.'**
  String get timeToAnswer;

  /// No description provided for @recordingUnavailable.
  ///
  /// In en, this message translates to:
  /// **'Recording is not available'**
  String get recordingUnavailable;

  /// No description provided for @recordingFailed.
  ///
  /// In en, this message translates to:
  /// **'Your voice could not be recorded: {error}'**
  String recordingFailed(String error);

  /// No description provided for @timeboxReached.
  ///
  /// In en, this message translates to:
  /// **'Timebox reached'**
  String get timeboxReached;

  /// No description provided for @timeboxStudied.
  ///
  /// In en, this message translates to:
  /// **'You studied {cards, plural, =1{1 card} other{{cards} cards}} in {minutes, plural, =1{1 minute} other{{minutes} minutes}}.'**
  String timeboxStudied(int cards, int minutes);

  /// Button that ends a study session.
  ///
  /// In en, this message translates to:
  /// **'Finish'**
  String get finish;

  /// Button that goes on studying.
  ///
  /// In en, this message translates to:
  /// **'Continue'**
  String get continueAction;

  /// No description provided for @noCardAnsweredYet.
  ///
  /// In en, this message translates to:
  /// **'No card has been answered yet.'**
  String get noCardAnsweredYet;

  /// No description provided for @removeFlag.
  ///
  /// In en, this message translates to:
  /// **'Remove flag'**
  String get removeFlag;

  /// No description provided for @stopRecording.
  ///
  /// In en, this message translates to:
  /// **'Stop recording'**
  String get stopRecording;

  /// No description provided for @recordVoice.
  ///
  /// In en, this message translates to:
  /// **'Record own voice'**
  String get recordVoice;

  /// No description provided for @cardActions.
  ///
  /// In en, this message translates to:
  /// **'Card actions'**
  String get cardActions;

  /// No description provided for @createCopy.
  ///
  /// In en, this message translates to:
  /// **'Create copy'**
  String get createCopy;

  /// No description provided for @replayAudio.
  ///
  /// In en, this message translates to:
  /// **'Replay audio'**
  String get replayAudio;

  /// No description provided for @playVoice.
  ///
  /// In en, this message translates to:
  /// **'Play own voice'**
  String get playVoice;

  /// Menu that sets a card's flag.
  ///
  /// In en, this message translates to:
  /// **'Flag'**
  String get flag;

  /// No description provided for @unmarkNote.
  ///
  /// In en, this message translates to:
  /// **'Unmark note'**
  String get unmarkNote;

  /// No description provided for @markNote.
  ///
  /// In en, this message translates to:
  /// **'Mark note'**
  String get markNote;

  /// No description provided for @previousCardInfo.
  ///
  /// In en, this message translates to:
  /// **'Previous card info'**
  String get previousCardInfo;

  /// No description provided for @resetCard.
  ///
  /// In en, this message translates to:
  /// **'Reset card'**
  String get resetCard;

  /// No description provided for @buryCardUntilTomorrow.
  ///
  /// In en, this message translates to:
  /// **'Bury card until tomorrow'**
  String get buryCardUntilTomorrow;

  /// No description provided for @buryNoteUntilTomorrow.
  ///
  /// In en, this message translates to:
  /// **'Bury note until tomorrow'**
  String get buryNoteUntilTomorrow;

  /// No description provided for @suspendCard.
  ///
  /// In en, this message translates to:
  /// **'Suspend card'**
  String get suspendCard;

  /// No description provided for @suspendNote.
  ///
  /// In en, this message translates to:
  /// **'Suspend note'**
  String get suspendNote;

  /// No description provided for @deleteNote.
  ///
  /// In en, this message translates to:
  /// **'Delete note'**
  String get deleteNote;

  /// No description provided for @stopAutoAdvance.
  ///
  /// In en, this message translates to:
  /// **'Stop auto advance'**
  String get stopAutoAdvance;

  /// No description provided for @deckOptions.
  ///
  /// In en, this message translates to:
  /// **'Deck options'**
  String get deckOptions;

  /// No description provided for @deckCouldNotOpen.
  ///
  /// In en, this message translates to:
  /// **'This deck could not be opened'**
  String get deckCouldNotOpen;

  /// No description provided for @autoAdvanceIsOn.
  ///
  /// In en, this message translates to:
  /// **'Auto advance is on. Stop it'**
  String get autoAdvanceIsOn;

  /// Button that reveals the answer of a card.
  ///
  /// In en, this message translates to:
  /// **'Show answer'**
  String get showAnswerButton;

  /// Rating button label in a filtered deck's preview: the card leaves the preview.
  ///
  /// In en, this message translates to:
  /// **'End'**
  String get previewEnd;

  /// Screen reader label of the button that ends studying a deck.
  ///
  /// In en, this message translates to:
  /// **'Leave {deck}'**
  String leaveDeck(String deck);

  /// No description provided for @leave.
  ///
  /// In en, this message translates to:
  /// **'Leave'**
  String get leave;

  /// Tooltip of a card's flag, e.g. "Flag: Red".
  ///
  /// In en, this message translates to:
  /// **'Flag: {name}'**
  String flagNamed(String name);

  /// Tooltip of the star on a card whose note is marked.
  ///
  /// In en, this message translates to:
  /// **'Marked'**
  String get marked;

  /// No description provided for @timeOnCard.
  ///
  /// In en, this message translates to:
  /// **'Time on this card'**
  String get timeOnCard;

  /// No description provided for @emptyCardNotice.
  ///
  /// In en, this message translates to:
  /// **'This card is empty: its question no longer shows anything. Remove empty cards with Settings › Empty cards.'**
  String get emptyCardNotice;

  /// Shown on the question of the card just answered when it comes straight back before it is due. {time} is like "11 minutes" or "1 h 5 min".
  ///
  /// In en, this message translates to:
  /// **'This card again: due in {time}, shown early because nothing else is due.'**
  String shownAgainEarly(String time);

  /// Shown on the question of a learning card shown before it is due, because nothing else is due. {time} is like "11 minutes" or "1 h 5 min".
  ///
  /// In en, this message translates to:
  /// **'Shown early: due in {time}, because nothing else is due.'**
  String shownEarly(String time);

  /// Shown on the question of the card just answered when it comes straight back because it is already due.
  ///
  /// In en, this message translates to:
  /// **'This card again: it is due again now.'**
  String get shownAgainDueNow;

  /// No description provided for @learningDueNow.
  ///
  /// In en, this message translates to:
  /// **'Learning cards are due now'**
  String get learningDueNow;

  /// No description provided for @backToDecks.
  ///
  /// In en, this message translates to:
  /// **'Back to decks'**
  String get backToDecks;

  /// No description provided for @congratulationsNothingDue.
  ///
  /// In en, this message translates to:
  /// **'Congratulations! Nothing is due right now.'**
  String get congratulationsNothingDue;

  /// {time} is like "12 minutes".
  ///
  /// In en, this message translates to:
  /// **'You studied {count, plural, =1{1 card} other{{count} cards}} in {time}.'**
  String studiedInTime(int count, String time);

  /// No description provided for @sessionComplete.
  ///
  /// In en, this message translates to:
  /// **'Session complete'**
  String get sessionComplete;

  /// The current streak on the deck list header, e.g. "12 days".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 day} other{{count} days}}'**
  String streakDaysLabel(int count);

  /// Deck list header when today has no answers yet and the streak is alive.
  ///
  /// In en, this message translates to:
  /// **'Study today to keep your {count}-day streak'**
  String streakAtRisk(int count);

  /// No description provided for @streakHeaderTooltip.
  ///
  /// In en, this message translates to:
  /// **'Streaks and daily goal'**
  String get streakHeaderTooltip;

  /// Tooltip of the level on the deck list header, which opens the achievements.
  ///
  /// In en, this message translates to:
  /// **'Level and achievements'**
  String get levelHeaderTooltip;

  /// Screen reader label of the level on the deck list header, e.g. "Level 12, 460 of 600 XP to level 13".
  ///
  /// In en, this message translates to:
  /// **'Level {level}, {done} of {need} XP to level {next}'**
  String levelHeaderSemantics(int level, int done, int need, int next);

  /// Today's progress toward a card goal, e.g. "18 / 30 cards".
  ///
  /// In en, this message translates to:
  /// **'{done} / {goal, plural, =1{1 card} other{{goal} cards}}'**
  String goalProgressCards(int done, int goal);

  /// Today's progress toward a minute goal, e.g. "12 / 20 minutes".
  ///
  /// In en, this message translates to:
  /// **'{done} / {goal, plural, =1{1 minute} other{{goal} minutes}}'**
  String goalProgressMinutes(int done, int goal);

  /// Study summary, when this session made today a study day.
  ///
  /// In en, this message translates to:
  /// **'Streak extended to {count, plural, =1{1 day} other{{count} days}}'**
  String streakExtended(int count);

  /// Study summary.
  ///
  /// In en, this message translates to:
  /// **'Current streak: {count, plural, =1{1 day} other{{count} days}}'**
  String streakCurrent(int count);

  /// No description provided for @goalReached.
  ///
  /// In en, this message translates to:
  /// **'Daily goal reached'**
  String get goalReached;

  /// Reward banner while studying when an answer raised the level, e.g. "Level 13".
  ///
  /// In en, this message translates to:
  /// **'Level {level}'**
  String levelBanner(int level);

  /// Summary line when the session raised the level, e.g. "Reached level 13".
  ///
  /// In en, this message translates to:
  /// **'Reached level {level}'**
  String levelReached(int level);

  /// Summary line: the level and the experience points (XP) left to the next one, e.g. "Level 12 · 140 XP to level 13".
  ///
  /// In en, this message translates to:
  /// **'Level {level} · {xp} XP to level {next}'**
  String levelXpLeft(int level, int xp, int next);

  /// Study summary.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 more card} other{{count} more cards}} to reach today\'\'s goal'**
  String goalCardsLeft(int count);

  /// Study summary.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 more minute} other{{count} more minutes}} to reach today\'\'s goal'**
  String goalMinutesLeft(int count);

  /// No description provided for @goalSuggestion.
  ///
  /// In en, this message translates to:
  /// **'A daily goal shows how far you are each day.'**
  String get goalSuggestion;

  /// No description provided for @goalSuggestionSet.
  ///
  /// In en, this message translates to:
  /// **'Set a daily goal'**
  String get goalSuggestionSet;

  /// No description provided for @goalSuggestionDismiss.
  ///
  /// In en, this message translates to:
  /// **'No thanks'**
  String get goalSuggestionDismiss;

  /// {when} is like "at 14:05" or "tomorrow".
  ///
  /// In en, this message translates to:
  /// **'More learning cards will be due {when}.'**
  String moreLearningDue(String when);

  /// No description provided for @checkAgain.
  ///
  /// In en, this message translates to:
  /// **'Check again'**
  String get checkAgain;

  /// No description provided for @browserMayClear.
  ///
  /// In en, this message translates to:
  /// **'Your browser may clear Cardflip data'**
  String get browserMayClear;

  /// No description provided for @browserMayClearBody.
  ///
  /// In en, this message translates to:
  /// **'This browser did not grant persistent storage, so it may delete your collection when it runs low on space. Export a backup regularly from Settings to keep your cards safe.'**
  String get browserMayClearBody;

  /// No description provided for @couldNotStart.
  ///
  /// In en, this message translates to:
  /// **'Cardflip could not start: {error}'**
  String couldNotStart(String error);

  /// No description provided for @collectionUnreadableShort.
  ///
  /// In en, this message translates to:
  /// **'Your collection could not be read.'**
  String get collectionUnreadableShort;

  /// No description provided for @startNewCollectionTitle.
  ///
  /// In en, this message translates to:
  /// **'Start a new collection?'**
  String get startNewCollectionTitle;

  /// No description provided for @startNewCollectionBody.
  ///
  /// In en, this message translates to:
  /// **'This permanently deletes the stored collection that could not be read. Restore a backup instead if you have one.'**
  String get startNewCollectionBody;

  /// No description provided for @deleteAndStartOver.
  ///
  /// In en, this message translates to:
  /// **'Delete and start over'**
  String get deleteAndStartOver;

  /// No description provided for @collectionCouldNotOpen.
  ///
  /// In en, this message translates to:
  /// **'Your collection could not be opened'**
  String get collectionCouldNotOpen;

  /// No description provided for @nothingDeleted.
  ///
  /// In en, this message translates to:
  /// **'Nothing has been deleted.'**
  String get nothingDeleted;

  /// No description provided for @restoreFromBackup.
  ///
  /// In en, this message translates to:
  /// **'Restore from backup'**
  String get restoreFromBackup;

  /// No description provided for @startNewCollection.
  ///
  /// In en, this message translates to:
  /// **'Start a new collection'**
  String get startNewCollection;

  /// Navigation destination: the deck list.
  ///
  /// In en, this message translates to:
  /// **'Decks'**
  String get navDecks;

  /// Navigation destination: the card browser.
  ///
  /// In en, this message translates to:
  /// **'Browse'**
  String get navBrowse;

  /// Navigation destination: the screen that adds notes.
  ///
  /// In en, this message translates to:
  /// **'Add'**
  String get navAdd;

  /// Navigation destination: the settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get navSettings;

  /// No description provided for @learningDueInDeck.
  ///
  /// In en, this message translates to:
  /// **'Learning cards are due in {deck}'**
  String learningDueInDeck(String deck);

  /// No description provided for @learningDueInDecks.
  ///
  /// In en, this message translates to:
  /// **'Learning cards are due in {count} decks'**
  String learningDueInDecks(int count);

  /// No description provided for @showDecks.
  ///
  /// In en, this message translates to:
  /// **'Show decks'**
  String get showDecks;

  /// Menu item that reverts a change, e.g. "Undo Delete 3 notes".
  ///
  /// In en, this message translates to:
  /// **'Undo {step}'**
  String undoStep(String step);

  /// Menu item that re-applies an undone change, e.g. "Redo Delete 3 notes".
  ///
  /// In en, this message translates to:
  /// **'Redo {step}'**
  String redoStep(String step);

  /// No description provided for @redo.
  ///
  /// In en, this message translates to:
  /// **'Redo'**
  String get redo;

  /// Tooltip of a menu with more actions.
  ///
  /// In en, this message translates to:
  /// **'More'**
  String get more;

  /// Snackbar after Undo, e.g. "Undone: Delete 3 notes".
  ///
  /// In en, this message translates to:
  /// **'Undone: {step}'**
  String undone(String step);

  /// Snackbar after Redo.
  ///
  /// In en, this message translates to:
  /// **'Redone: {step}'**
  String redone(String step);

  /// The Control key in shortcuts, e.g. Ctrl+Z.
  ///
  /// In en, this message translates to:
  /// **'Ctrl'**
  String get keyCtrl;

  /// The Shift key in shortcuts, e.g. Ctrl+Shift+Z.
  ///
  /// In en, this message translates to:
  /// **'Shift'**
  String get keyShift;

  /// The Alt key in shortcuts.
  ///
  /// In en, this message translates to:
  /// **'Alt'**
  String get keyAlt;

  /// The Enter key in shortcuts, e.g. Ctrl+Enter.
  ///
  /// In en, this message translates to:
  /// **'Enter'**
  String get keyEnter;

  /// The Delete key in shortcuts, e.g. Ctrl+Delete.
  ///
  /// In en, this message translates to:
  /// **'Delete'**
  String get keyDelete;

  /// The Escape key in shortcuts.
  ///
  /// In en, this message translates to:
  /// **'Esc'**
  String get keyEsc;

  /// A menu entry or tooltip with its keyboard shortcut, e.g. "Edit card (E)" or "Bold (Ctrl+B)".
  ///
  /// In en, this message translates to:
  /// **'{label} ({shortcut})'**
  String labelWithShortcut(String label, String shortcut);

  /// Name of flag 1.
  ///
  /// In en, this message translates to:
  /// **'Red'**
  String get flagRed;

  /// Name of flag 2.
  ///
  /// In en, this message translates to:
  /// **'Orange'**
  String get flagOrange;

  /// Name of flag 3.
  ///
  /// In en, this message translates to:
  /// **'Green'**
  String get flagGreen;

  /// Name of flag 4.
  ///
  /// In en, this message translates to:
  /// **'Blue'**
  String get flagBlue;

  /// Name of flag 5.
  ///
  /// In en, this message translates to:
  /// **'Pink'**
  String get flagPink;

  /// Name of flag 6.
  ///
  /// In en, this message translates to:
  /// **'Turquoise'**
  String get flagTurquoise;

  /// Name of flag 7.
  ///
  /// In en, this message translates to:
  /// **'Purple'**
  String get flagPurple;

  /// No description provided for @backupSavedAs.
  ///
  /// In en, this message translates to:
  /// **'Backup saved as {file}'**
  String backupSavedAs(String file);

  /// No description provided for @replaceCollectionTitle.
  ///
  /// In en, this message translates to:
  /// **'Replace your collection?'**
  String get replaceCollectionTitle;

  /// No description provided for @restoreBackupBody.
  ///
  /// In en, this message translates to:
  /// **'Restoring the backup from {date} replaces your current collection entirely: all decks, notes, cards, review history and settings are overwritten. This cannot be undone.'**
  String restoreBackupBody(String date);

  /// Button that restores a backup or deleted files.
  ///
  /// In en, this message translates to:
  /// **'Restore'**
  String get restore;

  /// No description provided for @backupRestored.
  ///
  /// In en, this message translates to:
  /// **'Backup restored'**
  String get backupRestored;

  /// No description provided for @deletionLogEmpty.
  ///
  /// In en, this message translates to:
  /// **'The deletion log is empty'**
  String get deletionLogEmpty;

  /// No description provided for @deletionLogSaved.
  ///
  /// In en, this message translates to:
  /// **'Deletion log with {count, plural, =1{1 note} other{{count} notes}} saved as {file}'**
  String deletionLogSaved(int count, String file);

  /// No description provided for @clearDeletionLogTitle.
  ///
  /// In en, this message translates to:
  /// **'Clear the deletion log?'**
  String get clearDeletionLogTitle;

  /// No description provided for @clearDeletionLogBody.
  ///
  /// In en, this message translates to:
  /// **'The deletion log keeps every note you deleted, so that it can be exported and imported again. Clearing it removes these records for good. This cannot be undone.'**
  String get clearDeletionLogBody;

  /// Button that clears something.
  ///
  /// In en, this message translates to:
  /// **'Clear'**
  String get clear;

  /// No description provided for @deletionLogCleared.
  ///
  /// In en, this message translates to:
  /// **'Deletion log cleared'**
  String get deletionLogCleared;

  /// Title of the dialog that renames a flag; {flag} is its colour name, e.g. Red.
  ///
  /// In en, this message translates to:
  /// **'Rename flag “{flag}”'**
  String renameFlag(String flag);

  /// No description provided for @renameFlagHelper.
  ///
  /// In en, this message translates to:
  /// **'Leave empty for \"{flag}\"'**
  String renameFlagHelper(String flag);

  /// No description provided for @defaultBrowserSearch.
  ///
  /// In en, this message translates to:
  /// **'Default browser search'**
  String get defaultBrowserSearch;

  /// No description provided for @defaultSearchHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. deck:Spanish is:due'**
  String get defaultSearchHint;

  /// No description provided for @defaultSearchHelper.
  ///
  /// In en, this message translates to:
  /// **'Leave empty to list all cards'**
  String get defaultSearchHelper;

  /// A setting that is turned off.
  ///
  /// In en, this message translates to:
  /// **'Off'**
  String get off;

  /// A number of minutes, short.
  ///
  /// In en, this message translates to:
  /// **'{count} min'**
  String minutesShort(int count);

  /// No description provided for @countHours.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 hour} other{{count} hours}}'**
  String countHours(int count);

  /// A number of milliseconds, short.
  ///
  /// In en, this message translates to:
  /// **'{count} ms'**
  String millisecondsShort(int count);

  /// No description provided for @settings.
  ///
  /// In en, this message translates to:
  /// **'Settings'**
  String get settings;

  /// No description provided for @appearance.
  ///
  /// In en, this message translates to:
  /// **'Appearance'**
  String get appearance;

  /// No description provided for @theme.
  ///
  /// In en, this message translates to:
  /// **'Theme'**
  String get theme;

  /// Theme that follows the device.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get themeSystem;

  /// The light theme.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// The dark theme.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// Setting for the interface language.
  ///
  /// In en, this message translates to:
  /// **'Language'**
  String get language;

  /// Interface language that follows the device's languages.
  ///
  /// In en, this message translates to:
  /// **'System default'**
  String get languageSystem;

  /// The name of this file's language in that language, e.g. Deutsch for German; shown in the Language setting.
  ///
  /// In en, this message translates to:
  /// **'English'**
  String get languageName;

  /// Setting that scales text, icons and spacing of the whole app.
  ///
  /// In en, this message translates to:
  /// **'UI size'**
  String get uiSize;

  /// No description provided for @uiSizeHelp.
  ///
  /// In en, this message translates to:
  /// **'Scales text, icons and spacing.'**
  String get uiSizeHelp;

  /// No description provided for @minimalistMode.
  ///
  /// In en, this message translates to:
  /// **'Minimalist mode'**
  String get minimalistMode;

  /// No description provided for @minimalistModeHelp.
  ///
  /// In en, this message translates to:
  /// **'Flat surfaces without shadows or tinted backgrounds.'**
  String get minimalistModeHelp;

  /// No description provided for @reduceMotion.
  ///
  /// In en, this message translates to:
  /// **'Reduce motion'**
  String get reduceMotion;

  /// No description provided for @reduceMotionHelp.
  ///
  /// In en, this message translates to:
  /// **'Show pages, dialogs and menus at once, without animations.'**
  String get reduceMotionHelp;

  /// Reduce motion follows the device's setting.
  ///
  /// In en, this message translates to:
  /// **'Follow system'**
  String get reduceMotionSystem;

  /// A setting that is turned on.
  ///
  /// In en, this message translates to:
  /// **'On'**
  String get on;

  /// No description provided for @studying.
  ///
  /// In en, this message translates to:
  /// **'Studying'**
  String get studying;

  /// No description provided for @nextDayStarts.
  ///
  /// In en, this message translates to:
  /// **'Next day starts at'**
  String get nextDayStarts;

  /// No description provided for @nextDayStartsHelp.
  ///
  /// In en, this message translates to:
  /// **'Daily limits and due cards reset at this hour.'**
  String get nextDayStartsHelp;

  /// No description provided for @learnAheadLimit.
  ///
  /// In en, this message translates to:
  /// **'Learn ahead limit'**
  String get learnAheadLimit;

  /// No description provided for @learnAheadLimitHelp.
  ///
  /// In en, this message translates to:
  /// **'When nothing else is due, show learning cards due within this time instead of waiting.'**
  String get learnAheadLimitHelp;

  /// No description provided for @timeboxLimit.
  ///
  /// In en, this message translates to:
  /// **'Timebox time limit'**
  String get timeboxLimit;

  /// No description provided for @timeboxLimitHelp.
  ///
  /// In en, this message translates to:
  /// **'Each time this much time passes while studying, show how many cards you studied.'**
  String get timeboxLimitHelp;

  /// No description provided for @autoContinueLearning.
  ///
  /// In en, this message translates to:
  /// **'Continue automatically when learning cards are due'**
  String get autoContinueLearning;

  /// No description provided for @autoContinueLearningHelp.
  ///
  /// In en, this message translates to:
  /// **'Otherwise a finished session asks before showing them.'**
  String get autoContinueLearningHelp;

  /// No description provided for @showPlayButtons.
  ///
  /// In en, this message translates to:
  /// **'Show play buttons on cards with audio'**
  String get showPlayButtons;

  /// No description provided for @interruptAudio.
  ///
  /// In en, this message translates to:
  /// **'Interrupt current audio when answering'**
  String get interruptAudio;

  /// No description provided for @interruptAudioHelp.
  ///
  /// In en, this message translates to:
  /// **'Otherwise the answer\'\'s audio plays after the question\'\'s.'**
  String get interruptAudioHelp;

  /// No description provided for @stripPasteFormatting.
  ///
  /// In en, this message translates to:
  /// **'Strip formatting when pasting'**
  String get stripPasteFormatting;

  /// {key} is Ctrl or ⌘.
  ///
  /// In en, this message translates to:
  /// **'Paste plain text; {key}+Shift+V pastes with formatting.'**
  String stripPasteFormattingHelp(String key);

  /// No description provided for @pasteImagesAsPng.
  ///
  /// In en, this message translates to:
  /// **'Paste images as PNG'**
  String get pasteImagesAsPng;

  /// No description provided for @pasteImagesAsPngHelp.
  ///
  /// In en, this message translates to:
  /// **'Lossless but larger; otherwise pasted images are JPEG.'**
  String get pasteImagesAsPngHelp;

  /// No description provided for @deckForNewCards.
  ///
  /// In en, this message translates to:
  /// **'Deck for new cards'**
  String get deckForNewCards;

  /// No description provided for @currentDeck.
  ///
  /// In en, this message translates to:
  /// **'Current deck'**
  String get currentDeck;

  /// No description provided for @deckByNoteType.
  ///
  /// In en, this message translates to:
  /// **'Change deck depending on note type'**
  String get deckByNoteType;

  /// Settings section.
  ///
  /// In en, this message translates to:
  /// **'Goals and streaks'**
  String get goalsAndStreaks;

  /// No description provided for @showStreaksAndGoals.
  ///
  /// In en, this message translates to:
  /// **'Show streaks and goals'**
  String get showStreaksAndGoals;

  /// No description provided for @showStreaksAndGoalsHelp.
  ///
  /// In en, this message translates to:
  /// **'On the deck list and after studying. Only on this device.'**
  String get showStreaksAndGoalsHelp;

  /// No description provided for @showStudyRewards.
  ///
  /// In en, this message translates to:
  /// **'Show rewards while studying'**
  String get showStudyRewards;

  /// No description provided for @showStudyRewardsHelp.
  ///
  /// In en, this message translates to:
  /// **'A short banner when you reach your goal, a level or an achievement. Only on this device.'**
  String get showStudyRewardsHelp;

  /// No description provided for @deckProgressInList.
  ///
  /// In en, this message translates to:
  /// **'Show deck progress in the deck list'**
  String get deckProgressInList;

  /// No description provided for @deckProgressInListHelp.
  ///
  /// In en, this message translates to:
  /// **'A thin bar of new, learning, young and mature cards under each deck. Only on this device.'**
  String get deckProgressInListHelp;

  /// Deck overview: heading of the bar of the deck's new, learning, young and mature cards.
  ///
  /// In en, this message translates to:
  /// **'Progress'**
  String get deckProgress;

  /// Deck overview and deck list: the share of the deck's cards (without suspended ones) that are mature, such as "62% mature".
  ///
  /// In en, this message translates to:
  /// **'{percent} mature'**
  String deckProgressMature(String percent);

  /// Deck overview: one part of the progress bar, such as "Young: 12".
  ///
  /// In en, this message translates to:
  /// **'{state}: {count}'**
  String deckProgressLegend(String state, int count);

  /// Deck list: heading of the weekly recap, which compares the previous week with the week before.
  ///
  /// In en, this message translates to:
  /// **'Last week'**
  String get recapTitle;

  /// No description provided for @recapTooltip.
  ///
  /// In en, this message translates to:
  /// **'Last week compared with the week before. Opens the statistics.'**
  String get recapTooltip;

  /// No description provided for @recapDismiss.
  ///
  /// In en, this message translates to:
  /// **'Hide until next week'**
  String get recapDismiss;

  /// Weekly recap.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 study day} other{{count} study days}}'**
  String recapStudyDays(int count);

  /// Weekly recap.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 answer} other{{count} answers}}'**
  String recapAnswers(int count);

  /// Weekly recap: time spent studying.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 minute} other{{count} minutes}}'**
  String recapMinutes(int count);

  /// Weekly recap: the true retention of the week's reviews, such as "91% retention".
  ///
  /// In en, this message translates to:
  /// **'{percent} retention'**
  String recapRetention(String percent);

  /// Weekly recap: tooltip of a number's change, with the number of the week before.
  ///
  /// In en, this message translates to:
  /// **'The week before: {value}'**
  String recapWeekBefore(String value);

  /// Weekly recap: the change of the retention in percentage points, such as "+3 pts".
  ///
  /// In en, this message translates to:
  /// **'{change} pts'**
  String recapPoints(String change);

  /// Weekly recap: a number did not change from the week before.
  ///
  /// In en, this message translates to:
  /// **'same'**
  String get recapSame;

  /// No description provided for @dailyGoal.
  ///
  /// In en, this message translates to:
  /// **'Daily goal'**
  String get dailyGoal;

  /// No daily goal.
  ///
  /// In en, this message translates to:
  /// **'None'**
  String get dailyGoalNone;

  /// A daily goal of a number of cards.
  ///
  /// In en, this message translates to:
  /// **'Cards'**
  String get dailyGoalCards;

  /// A daily goal of a number of minutes.
  ///
  /// In en, this message translates to:
  /// **'Minutes'**
  String get dailyGoalMinutes;

  /// No description provided for @dailyGoalCardsPerDay.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}} a day'**
  String dailyGoalCardsPerDay(int count);

  /// No description provided for @dailyGoalMinutesPerDay.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 minute} other{{count} minutes}} a day'**
  String dailyGoalMinutesPerDay(int count);

  /// No description provided for @dailyGoalCardsLabel.
  ///
  /// In en, this message translates to:
  /// **'Cards per day'**
  String get dailyGoalCardsLabel;

  /// No description provided for @dailyGoalMinutesLabel.
  ///
  /// In en, this message translates to:
  /// **'Minutes per day'**
  String get dailyGoalMinutesLabel;

  /// No description provided for @dailyGoalError.
  ///
  /// In en, this message translates to:
  /// **'Enter a whole number from 1 to 9999.'**
  String get dailyGoalError;

  /// No description provided for @dailyGoalHelp.
  ///
  /// In en, this message translates to:
  /// **'Studying on any day keeps your streak, whether or not you reach the goal.'**
  String get dailyGoalHelp;

  /// Setting: how many days without study a streak survives in any 7 days.
  ///
  /// In en, this message translates to:
  /// **'Rest days per week'**
  String get restDaysPerWeek;

  /// No description provided for @restDaysPerWeekHelp.
  ///
  /// In en, this message translates to:
  /// **'Days without study that keep a streak going, in any 7 days in a row.'**
  String get restDaysPerWeekHelp;

  /// Screen and menu entry listing the achievement badges.
  ///
  /// In en, this message translates to:
  /// **'Achievements'**
  String get achievements;

  /// The level on the achievements screen's level card, e.g. "Level 12".
  ///
  /// In en, this message translates to:
  /// **'Level {level}'**
  String levelTitle(int level);

  /// XP toward the next level on the level card, e.g. "460 / 600 XP to level 13".
  ///
  /// In en, this message translates to:
  /// **'{done} / {need} XP to level {next}'**
  String levelXpProgress(int done, int need, int next);

  /// An amount of experience points, e.g. "2,900 XP".
  ///
  /// In en, this message translates to:
  /// **'{xp} XP'**
  String xpAmount(int xp);

  /// Level card: XP earned from answers.
  ///
  /// In en, this message translates to:
  /// **'Answers'**
  String get xpFromAnswers;

  /// Level card: XP earned from study days.
  ///
  /// In en, this message translates to:
  /// **'Study days'**
  String get xpFromStudyDays;

  /// Level card: XP earned from days that met the daily goal.
  ///
  /// In en, this message translates to:
  /// **'Daily goal met'**
  String get xpFromGoalDays;

  /// Level card: XP earned from cards reaching long-term memory.
  ///
  /// In en, this message translates to:
  /// **'Long-term memory'**
  String get xpFromLongTerm;

  /// Level card: XP earned from achievement tiers.
  ///
  /// In en, this message translates to:
  /// **'Achievements'**
  String get xpFromAchievements;

  /// Level card: how experience points are earned.
  ///
  /// In en, this message translates to:
  /// **'Each answer earns 1 XP, each study day 10, each day the daily goal is met 20, each card reaching long-term memory 5, and each achievement tier 25 to 200.'**
  String get xpExplanation;

  /// Level card hint while no daily goal is set.
  ///
  /// In en, this message translates to:
  /// **'Set a daily goal to earn 20 XP on each day you meet it.'**
  String get xpNoGoalHint;

  /// Achievement name: number of answers.
  ///
  /// In en, this message translates to:
  /// **'Answers'**
  String get achievementAnswers;

  /// Achievement name: longest streak.
  ///
  /// In en, this message translates to:
  /// **'Streak'**
  String get achievementStreak;

  /// Achievement name: number of days with study.
  ///
  /// In en, this message translates to:
  /// **'Study days'**
  String get achievementStudyDays;

  /// Achievement name: hours spent answering.
  ///
  /// In en, this message translates to:
  /// **'Study time'**
  String get achievementStudyTime;

  /// Achievement name: days the daily goal was met.
  ///
  /// In en, this message translates to:
  /// **'Daily goal'**
  String get achievementDailyGoal;

  /// Achievement name: cards whose interval reached 21 days.
  ///
  /// In en, this message translates to:
  /// **'Long-term memory'**
  String get achievementLongTerm;

  /// Achievement name: the level reached.
  ///
  /// In en, this message translates to:
  /// **'Level'**
  String get achievementLevel;

  /// Achievement name: studying again after 14 or more days without study.
  ///
  /// In en, this message translates to:
  /// **'Comeback'**
  String get achievementComeback;

  /// First achievement tier.
  ///
  /// In en, this message translates to:
  /// **'Bronze'**
  String get tierBronze;

  /// Second achievement tier.
  ///
  /// In en, this message translates to:
  /// **'Silver'**
  String get tierSilver;

  /// Third achievement tier.
  ///
  /// In en, this message translates to:
  /// **'Gold'**
  String get tierGold;

  /// Fourth achievement tier.
  ///
  /// In en, this message translates to:
  /// **'Platinum'**
  String get tierPlatinum;

  /// The threshold of an Answers tier, e.g. "1,000 answers".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 answer} other{{count} answers}}'**
  String tierAnswers(int count);

  /// The threshold of a Streak or Study days tier, e.g. "100 days".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 day} other{{count} days}}'**
  String tierDays(int count);

  /// The threshold of a Daily goal tier, e.g. "met on 100 days".
  ///
  /// In en, this message translates to:
  /// **'met on {count, plural, =1{1 day} other{{count} days}}'**
  String tierGoalDays(int count);

  /// The threshold of a Long-term memory tier, e.g. "1,000 cards".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}}'**
  String tierCards(int count);

  /// The threshold of a Study time tier, e.g. "100 hours".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 hour} other{{count} hours}}'**
  String tierHours(int count);

  /// The threshold of a Level tier, also naming it in a notice, e.g. "Level 25".
  ///
  /// In en, this message translates to:
  /// **'Level {level}'**
  String tierLevel(int level);

  /// The only tier of the Comeback achievement.
  ///
  /// In en, this message translates to:
  /// **'One-time'**
  String get tierOneTime;

  /// An achievement tier, e.g. "Gold · 100 days".
  ///
  /// In en, this message translates to:
  /// **'{tier} · {threshold}'**
  String tierWithThreshold(String tier, String threshold);

  /// When an achievement tier was earned.
  ///
  /// In en, this message translates to:
  /// **'Earned {date}'**
  String achievementEarnedOn(String date);

  /// Progress toward the next tier, e.g. "640 / 1,000".
  ///
  /// In en, this message translates to:
  /// **'{progress} / {next}'**
  String achievementProgress(int progress, int next);

  /// Progress toward the next Study time tier, e.g. "10 / 100 hours".
  ///
  /// In en, this message translates to:
  /// **'{progress} / {next} hours'**
  String achievementProgressHours(int progress, int next);

  /// No description provided for @achievementAllEarned.
  ///
  /// In en, this message translates to:
  /// **'All tiers earned'**
  String get achievementAllEarned;

  /// No description provided for @achievementNeedsGoal.
  ///
  /// In en, this message translates to:
  /// **'Set a daily goal to start'**
  String get achievementNeedsGoal;

  /// No description provided for @achievementComebackHint.
  ///
  /// In en, this message translates to:
  /// **'Study again after 14 days or more without study'**
  String get achievementComebackHint;

  /// Screen reader label of a locked badge.
  ///
  /// In en, this message translates to:
  /// **'{name}, not earned yet'**
  String achievementLocked(String name);

  /// Screen reader label of an earned badge, e.g. "Streak, Gold".
  ///
  /// In en, this message translates to:
  /// **'{name}, {tier}'**
  String achievementEarned(String name, String tier);

  /// A new Answers tier in a notice, e.g. "New: 1,000 answers".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 answer} other{{count} answers}}'**
  String newAnswers(int count);

  /// A new Streak tier in a notice.
  ///
  /// In en, this message translates to:
  /// **'{count}-day streak'**
  String newStreak(int count);

  /// A new Study days tier in a notice.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 study day} other{{count} study days}}'**
  String newStudyDays(int count);

  /// A new Study time tier in a notice, e.g. "100 hours of study".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 hour} other{{count} hours}} of study'**
  String newStudyTime(int count);

  /// A new Daily goal tier in a notice.
  ///
  /// In en, this message translates to:
  /// **'Daily goal met on {count, plural, =1{1 day} other{{count} days}}'**
  String newDailyGoal(int count);

  /// A new Long-term memory tier in a notice.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}} in long-term memory'**
  String newLongTerm(int count);

  /// Summary line naming the achievement tiers earned in the session, e.g. "New: 1,000 answers".
  ///
  /// In en, this message translates to:
  /// **'New: {names}'**
  String newAchievementsLine(String names);

  /// Summary line when more than three tiers were earned in the session.
  ///
  /// In en, this message translates to:
  /// **'New: {count} achievements'**
  String newAchievementsCountLine(int count);

  /// Deck list message for one tier earned through sync or import.
  ///
  /// In en, this message translates to:
  /// **'New achievement: {name}'**
  String newAchievement(String name);

  /// Deck list message for two or three tiers.
  ///
  /// In en, this message translates to:
  /// **'New achievements: {names}'**
  String newAchievements(String names);

  /// Deck list message when more than three tiers are new at once, e.g. after an import.
  ///
  /// In en, this message translates to:
  /// **'{count} achievements from your history'**
  String achievementsFromHistory(int count);

  /// Action of the new achievement message that opens the achievements screen.
  ///
  /// In en, this message translates to:
  /// **'Show'**
  String get showAchievements;

  /// Settings section about the study screen.
  ///
  /// In en, this message translates to:
  /// **'Review'**
  String get review;

  /// No description provided for @showRemainingCounts.
  ///
  /// In en, this message translates to:
  /// **'Show remaining card count'**
  String get showRemainingCounts;

  /// No description provided for @showIntervals.
  ///
  /// In en, this message translates to:
  /// **'Show next review time above answer buttons'**
  String get showIntervals;

  /// No description provided for @spaceAnswersGood.
  ///
  /// In en, this message translates to:
  /// **'Space/Enter also answers Good'**
  String get spaceAnswersGood;

  /// No description provided for @spaceAnswersGoodHelp.
  ///
  /// In en, this message translates to:
  /// **'Otherwise they only reveal the answer.'**
  String get spaceAnswersGoodHelp;

  /// No description provided for @hideBarsDuringReview.
  ///
  /// In en, this message translates to:
  /// **'Hide bars during review'**
  String get hideBarsDuringReview;

  /// No description provided for @hideBarsDuringReviewHelp.
  ///
  /// In en, this message translates to:
  /// **'Hide the app bar and navigation while studying; Escape or the top edge leaves.'**
  String get hideBarsDuringReviewHelp;

  /// No description provided for @questionFade.
  ///
  /// In en, this message translates to:
  /// **'Question fade'**
  String get questionFade;

  /// No description provided for @questionFadeHelp.
  ///
  /// In en, this message translates to:
  /// **'How long each new question fades in.'**
  String get questionFadeHelp;

  /// No description provided for @browsing.
  ///
  /// In en, this message translates to:
  /// **'Browsing'**
  String get browsing;

  /// No description provided for @ignoreAccents.
  ///
  /// In en, this message translates to:
  /// **'Ignore accents in search'**
  String get ignoreAccents;

  /// No description provided for @ignoreAccentsHelp.
  ///
  /// In en, this message translates to:
  /// **'For example, \"cafe\" also finds \"café\".'**
  String get ignoreAccentsHelp;

  /// No description provided for @allCards.
  ///
  /// In en, this message translates to:
  /// **'All cards'**
  String get allCards;

  /// Settings section about the collection's data.
  ///
  /// In en, this message translates to:
  /// **'Collection'**
  String get collection;

  /// Settings item and screen that imports files.
  ///
  /// In en, this message translates to:
  /// **'Import'**
  String get import;

  /// No description provided for @importHelp.
  ///
  /// In en, this message translates to:
  /// **'Text (.csv, .tsv, .txt), Anki package (.apkg) or collection (.colpkg)'**
  String get importHelp;

  /// No description provided for @exportBackup.
  ///
  /// In en, this message translates to:
  /// **'Export backup'**
  String get exportBackup;

  /// No description provided for @downloadsBackupFile.
  ///
  /// In en, this message translates to:
  /// **'Downloads a .cardflip file'**
  String get downloadsBackupFile;

  /// No description provided for @savesBackupFile.
  ///
  /// In en, this message translates to:
  /// **'Saves a .cardflip file'**
  String get savesBackupFile;

  /// No description provided for @exportCollection.
  ///
  /// In en, this message translates to:
  /// **'Export collection'**
  String get exportCollection;

  /// No description provided for @exportCollectionHelp.
  ///
  /// In en, this message translates to:
  /// **'Anki collection package or plain text'**
  String get exportCollectionHelp;

  /// No description provided for @restoreBackup.
  ///
  /// In en, this message translates to:
  /// **'Restore backup'**
  String get restoreBackup;

  /// No description provided for @restoreBackupHelp.
  ///
  /// In en, this message translates to:
  /// **'Replaces the current collection'**
  String get restoreBackupHelp;

  /// No description provided for @emptyCards.
  ///
  /// In en, this message translates to:
  /// **'Empty cards'**
  String get emptyCards;

  /// No description provided for @emptyCardsHelp.
  ///
  /// In en, this message translates to:
  /// **'Find and delete cards whose question is empty'**
  String get emptyCardsHelp;

  /// No description provided for @checkMedia.
  ///
  /// In en, this message translates to:
  /// **'Check media'**
  String get checkMedia;

  /// No description provided for @checkMediaHelp.
  ///
  /// In en, this message translates to:
  /// **'Find unused and missing media files, restore deleted ones'**
  String get checkMediaHelp;

  /// No description provided for @checkDatabase.
  ///
  /// In en, this message translates to:
  /// **'Check database'**
  String get checkDatabase;

  /// No description provided for @checkDatabaseHelp.
  ///
  /// In en, this message translates to:
  /// **'Find and repair problems in the collection'**
  String get checkDatabaseHelp;

  /// No description provided for @exportDeletionLog.
  ///
  /// In en, this message translates to:
  /// **'Export deletion log'**
  String get exportDeletionLog;

  /// No description provided for @exportDeletionLogHelp.
  ///
  /// In en, this message translates to:
  /// **'Deleted notes as a text file that Import can read'**
  String get exportDeletionLogHelp;

  /// No description provided for @clearDeletionLog.
  ///
  /// In en, this message translates to:
  /// **'Clear deletion log'**
  String get clearDeletionLog;

  /// No description provided for @clearDeletionLogHelp.
  ///
  /// In en, this message translates to:
  /// **'Forgets the notes deleted so far'**
  String get clearDeletionLogHelp;

  /// No description provided for @backups.
  ///
  /// In en, this message translates to:
  /// **'Backups'**
  String get backups;

  /// No description provided for @flags.
  ///
  /// In en, this message translates to:
  /// **'Flags'**
  String get flags;

  /// No description provided for @agentsMcp.
  ///
  /// In en, this message translates to:
  /// **'Agents (MCP)'**
  String get agentsMcp;

  /// No description provided for @allowAgents.
  ///
  /// In en, this message translates to:
  /// **'Allow agents (MCP)'**
  String get allowAgents;

  /// No description provided for @allowAgentsHelp.
  ///
  /// In en, this message translates to:
  /// **'AI agents such as Claude Code can read and change the open collection while Cardflip runs.'**
  String get allowAgentsHelp;

  /// No description provided for @allowAgentsTitle.
  ///
  /// In en, this message translates to:
  /// **'Allow agents?'**
  String get allowAgentsTitle;

  /// No description provided for @allowAgentsBody.
  ///
  /// In en, this message translates to:
  /// **'Any program on this computer that has the token can read and change your collection while Cardflip is open. Give the token only to agents you trust.'**
  String get allowAgentsBody;

  /// No description provided for @allow.
  ///
  /// In en, this message translates to:
  /// **'Allow'**
  String get allow;

  /// No description provided for @mcpServerRunning.
  ///
  /// In en, this message translates to:
  /// **'Server running'**
  String get mcpServerRunning;

  /// No description provided for @mcpServerStopped.
  ///
  /// In en, this message translates to:
  /// **'Server not running'**
  String get mcpServerStopped;

  /// No description provided for @mcpPort.
  ///
  /// In en, this message translates to:
  /// **'Port'**
  String get mcpPort;

  /// No description provided for @mcpPortHelper.
  ///
  /// In en, this message translates to:
  /// **'1024 to 65535'**
  String get mcpPortHelper;

  /// No description provided for @mcpToken.
  ///
  /// In en, this message translates to:
  /// **'Token'**
  String get mcpToken;

  /// No description provided for @copyToken.
  ///
  /// In en, this message translates to:
  /// **'Copy token'**
  String get copyToken;

  /// No description provided for @tokenCopied.
  ///
  /// In en, this message translates to:
  /// **'Token copied'**
  String get tokenCopied;

  /// No description provided for @regenerateToken.
  ///
  /// In en, this message translates to:
  /// **'Regenerate token'**
  String get regenerateToken;

  /// No description provided for @regenerateTokenTitle.
  ///
  /// In en, this message translates to:
  /// **'Regenerate the token?'**
  String get regenerateTokenTitle;

  /// No description provided for @regenerateTokenBody.
  ///
  /// In en, this message translates to:
  /// **'Agents that use the current token can no longer connect until you give them the new one.'**
  String get regenerateTokenBody;

  /// No description provided for @regenerate.
  ///
  /// In en, this message translates to:
  /// **'Regenerate'**
  String get regenerate;

  /// No description provided for @copyClientConfig.
  ///
  /// In en, this message translates to:
  /// **'Copy client configuration'**
  String get copyClientConfig;

  /// No description provided for @copyClientConfigHelp.
  ///
  /// In en, this message translates to:
  /// **'The server URL and token for Claude Code, Claude Desktop and other MCP clients'**
  String get copyClientConfigHelp;

  /// No description provided for @clientConfigCopied.
  ///
  /// In en, this message translates to:
  /// **'Client configuration copied'**
  String get clientConfigCopied;

  /// No description provided for @agentActivity.
  ///
  /// In en, this message translates to:
  /// **'Recent agent changes'**
  String get agentActivity;

  /// No description provided for @noAgentActivity.
  ///
  /// In en, this message translates to:
  /// **'No changes in this session'**
  String get noAgentActivity;

  /// No description provided for @storage.
  ///
  /// In en, this message translates to:
  /// **'Storage'**
  String get storage;

  /// No description provided for @storagePersistent.
  ///
  /// In en, this message translates to:
  /// **'Persistent: the browser keeps your collection.'**
  String get storagePersistent;

  /// No description provided for @storageNotPersistent.
  ///
  /// In en, this message translates to:
  /// **'The browser may clear your collection when space runs low. Export backups regularly.'**
  String get storageNotPersistent;

  /// No description provided for @about.
  ///
  /// In en, this message translates to:
  /// **'About'**
  String get about;

  /// No description provided for @aboutCardflip.
  ///
  /// In en, this message translates to:
  /// **'Version {version}\nSpaced repetition flash cards. Your data stays on this device.\nCopyright © 2026 madzione. Free software under the GNU GPL version 3 or later, with absolutely no warranty.'**
  String aboutCardflip(String version);

  /// No description provided for @openSourceLicenses.
  ///
  /// In en, this message translates to:
  /// **'Open-source licenses'**
  String get openSourceLicenses;

  /// No description provided for @sourceCode.
  ///
  /// In en, this message translates to:
  /// **'Source code'**
  String get sourceCode;

  /// No description provided for @licenseLegalese.
  ///
  /// In en, this message translates to:
  /// **'Copyright © 2026 madzione\n\nCardflip is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License, version 3 or later. It comes with absolutely no warranty.'**
  String get licenseLegalese;

  /// No description provided for @backupEveryMinutes.
  ///
  /// In en, this message translates to:
  /// **'Every {count} min'**
  String backupEveryMinutes(int count);

  /// No description provided for @backupEveryHour.
  ///
  /// In en, this message translates to:
  /// **'Every hour'**
  String get backupEveryHour;

  /// No description provided for @backupEveryHours.
  ///
  /// In en, this message translates to:
  /// **'Every {count} hours'**
  String backupEveryHours(int count);

  /// No description provided for @restoreAutoBackupBody.
  ///
  /// In en, this message translates to:
  /// **'Restoring the automatic backup from {date} replaces your current collection entirely: all decks, notes, cards, review history and settings are overwritten. Media files stay as they are. Your current collection is backed up first, so you can go back to it.'**
  String restoreAutoBackupBody(String date);

  /// No description provided for @backUpAutomatically.
  ///
  /// In en, this message translates to:
  /// **'Back up automatically'**
  String get backUpAutomatically;

  /// No description provided for @backUpAutomaticallyHelp.
  ///
  /// In en, this message translates to:
  /// **'While the collection changes, without media. Before a restore a backup is always made.'**
  String get backUpAutomaticallyHelp;

  /// No description provided for @keepBackups.
  ///
  /// In en, this message translates to:
  /// **'Keep besides today\'\'s backups'**
  String get keepBackups;

  /// How many daily backups to keep.
  ///
  /// In en, this message translates to:
  /// **'Daily'**
  String get keepDaily;

  /// How many weekly backups to keep.
  ///
  /// In en, this message translates to:
  /// **'Weekly'**
  String get keepWeekly;

  /// How many monthly backups to keep.
  ///
  /// In en, this message translates to:
  /// **'Monthly'**
  String get keepMonthly;

  /// No description provided for @createBackupNow.
  ///
  /// In en, this message translates to:
  /// **'Create backup now'**
  String get createBackupNow;

  /// No description provided for @noAutoBackups.
  ///
  /// In en, this message translates to:
  /// **'No automatic backups yet'**
  String get noAutoBackups;

  /// No description provided for @countAutoBackups.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 automatic backup} other{{count} automatic backups}}'**
  String countAutoBackups(int count);

  /// No description provided for @usedOnDevice.
  ///
  /// In en, this message translates to:
  /// **'{size} used on this device'**
  String usedOnDevice(String size);

  /// No description provided for @showFewer.
  ///
  /// In en, this message translates to:
  /// **'Show fewer'**
  String get showFewer;

  /// No description provided for @showAllCount.
  ///
  /// In en, this message translates to:
  /// **'Show all {count}'**
  String showAllCount(int count);

  /// No description provided for @deletedEmptyCards.
  ///
  /// In en, this message translates to:
  /// **'Deleted {count, plural, =1{1 empty card} other{{count} empty cards}}'**
  String deletedEmptyCards(int count);

  /// No description provided for @close.
  ///
  /// In en, this message translates to:
  /// **'Close'**
  String get close;

  /// No description provided for @noEmptyCards.
  ///
  /// In en, this message translates to:
  /// **'There are no empty cards.'**
  String get noEmptyCards;

  /// No description provided for @emptyCardsSummary.
  ///
  /// In en, this message translates to:
  /// **'{cards, plural, =1{1 empty card} other{{cards} empty cards}} in {notes, plural, =1{1 note} other{{notes} notes}}. Their question no longer shows anything.'**
  String emptyCardsSummary(int cards, int notes);

  /// No description provided for @emptyCardsKept.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note has} other{{count} notes have}} only empty cards. One card of each is kept, so no note is deleted.'**
  String emptyCardsKept(int count);

  /// Shown for a note whose text is empty.
  ///
  /// In en, this message translates to:
  /// **'(empty)'**
  String get emptyText;

  /// No description provided for @oneCardKept.
  ///
  /// In en, this message translates to:
  /// **'one card is kept'**
  String get oneCardKept;

  /// No description provided for @deleteCards.
  ///
  /// In en, this message translates to:
  /// **'Delete {count, plural, =1{1 card} other{{count} cards}}'**
  String deleteCards(int count);

  /// {size} is like "1.5 MB".
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 file} other{{count} files}} ({size})'**
  String filesWithSize(int count, String size);

  /// No description provided for @deleteUnusedTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete unused files?'**
  String get deleteUnusedTitle;

  /// No description provided for @deleteUnusedBody.
  ///
  /// In en, this message translates to:
  /// **'Moves {files} to the media trash. Restore deleted brings them back until the trash is emptied; files are removed from the trash after 30 days.'**
  String deleteUnusedBody(String files);

  /// No description provided for @movedToTrash.
  ///
  /// In en, this message translates to:
  /// **'Moved {count, plural, =1{1 file} other{{count} files}} to the media trash'**
  String movedToTrash(int count);

  /// No description provided for @tagMissingTitle.
  ///
  /// In en, this message translates to:
  /// **'Tag notes with missing files?'**
  String get tagMissingTitle;

  /// No description provided for @tagMissingBody.
  ///
  /// In en, this message translates to:
  /// **'Adds the tag missing-media to {count, plural, =1{1 note} other{{count} notes}}, so that you can find them in the browser with tag:missing-media.'**
  String tagMissingBody(int count);

  /// Button that adds a tag to notes.
  ///
  /// In en, this message translates to:
  /// **'Tag'**
  String get tag;

  /// No description provided for @taggedMissing.
  ///
  /// In en, this message translates to:
  /// **'Tagged {count, plural, =1{1 note} other{{count} notes}} missing-media'**
  String taggedMissing(int count);

  /// No description provided for @restoreDeletedTitle.
  ///
  /// In en, this message translates to:
  /// **'Restore deleted files?'**
  String get restoreDeletedTitle;

  /// No description provided for @restoreDeletedBody.
  ///
  /// In en, this message translates to:
  /// **'Moves {files} back into your media.'**
  String restoreDeletedBody(String files);

  /// No description provided for @restoredFilesKept.
  ///
  /// In en, this message translates to:
  /// **'Restored {count, plural, =1{1 file} other{{count} files}}; {kept, plural, =1{1 file} other{{kept} files}} stayed in the trash because a different file now has the name'**
  String restoredFilesKept(int count, int kept);

  /// No description provided for @restoredFiles.
  ///
  /// In en, this message translates to:
  /// **'Restored {count, plural, =1{1 file} other{{count} files}}'**
  String restoredFiles(int count);

  /// No description provided for @emptyTrashTitle.
  ///
  /// In en, this message translates to:
  /// **'Empty the media trash?'**
  String get emptyTrashTitle;

  /// No description provided for @emptyTrashBody.
  ///
  /// In en, this message translates to:
  /// **'Deletes {files} for good. This cannot be undone.'**
  String emptyTrashBody(String files);

  /// No description provided for @emptyTrash.
  ///
  /// In en, this message translates to:
  /// **'Empty trash'**
  String get emptyTrash;

  /// No description provided for @trashEmptied.
  ///
  /// In en, this message translates to:
  /// **'Media trash emptied'**
  String get trashEmptied;

  /// No description provided for @unusedFiles.
  ///
  /// In en, this message translates to:
  /// **'Unused files'**
  String get unusedFiles;

  /// No description provided for @allMediaUsed.
  ///
  /// In en, this message translates to:
  /// **'Every media file is used by a note, a template or styling.'**
  String get allMediaUsed;

  /// No description provided for @unusedFilesBody.
  ///
  /// In en, this message translates to:
  /// **'{files} that no note, template or styling refers to. Files whose names start with _ are always kept.'**
  String unusedFilesBody(String files);

  /// No description provided for @deleteUnused.
  ///
  /// In en, this message translates to:
  /// **'Delete unused'**
  String get deleteUnused;

  /// No description provided for @missingFiles.
  ///
  /// In en, this message translates to:
  /// **'Missing files'**
  String get missingFiles;

  /// No description provided for @noMissingFiles.
  ///
  /// In en, this message translates to:
  /// **'Every file that notes refer to is there.'**
  String get noMissingFiles;

  /// No description provided for @missingFilesBody.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 file} other{{count} files}} that {notes, plural, =1{1 note refers} other{{notes} notes refer}} to {count, plural, =1{is} other{are}} not in your media.'**
  String missingFilesBody(int count, int notes);

  /// No description provided for @tagMissing.
  ///
  /// In en, this message translates to:
  /// **'Tag missing'**
  String get tagMissing;

  /// No description provided for @mediaTrash.
  ///
  /// In en, this message translates to:
  /// **'Media trash'**
  String get mediaTrash;

  /// No description provided for @trashIsEmpty.
  ///
  /// In en, this message translates to:
  /// **'The trash is empty.'**
  String get trashIsEmpty;

  /// No description provided for @trashBody.
  ///
  /// In en, this message translates to:
  /// **'{files}. Files are removed from the trash after 30 days.'**
  String trashBody(String files);

  /// No description provided for @restoreDeleted.
  ///
  /// In en, this message translates to:
  /// **'Restore deleted'**
  String get restoreDeleted;

  /// No description provided for @checkingNotes.
  ///
  /// In en, this message translates to:
  /// **'Checking notes'**
  String get checkingNotes;

  /// No description provided for @checkingCards.
  ///
  /// In en, this message translates to:
  /// **'Checking cards'**
  String get checkingCards;

  /// No description provided for @readingHistory.
  ///
  /// In en, this message translates to:
  /// **'Reading the review history'**
  String get readingHistory;

  /// No description provided for @savingRepairs.
  ///
  /// In en, this message translates to:
  /// **'Saving repairs'**
  String get savingRepairs;

  /// No description provided for @backingUp.
  ///
  /// In en, this message translates to:
  /// **'Backing up the collection'**
  String get backingUp;

  /// No description provided for @noProblemsFound.
  ///
  /// In en, this message translates to:
  /// **'No problems were found.'**
  String get noProblemsFound;

  /// No description provided for @orphanedReviews.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 review log entry belongs} other{{count} review log entries belong}} to deleted cards. They are kept for statistics.'**
  String orphanedReviews(int count);

  /// No description provided for @updateIfNewer.
  ///
  /// In en, this message translates to:
  /// **'If the package version is newer'**
  String get updateIfNewer;

  /// No description provided for @updateAlways.
  ///
  /// In en, this message translates to:
  /// **'Always'**
  String get updateAlways;

  /// No description provided for @updateNever.
  ///
  /// In en, this message translates to:
  /// **'Never'**
  String get updateNever;

  /// No description provided for @separatorComma.
  ///
  /// In en, this message translates to:
  /// **'Comma'**
  String get separatorComma;

  /// No description provided for @separatorSemicolon.
  ///
  /// In en, this message translates to:
  /// **'Semicolon'**
  String get separatorSemicolon;

  /// No description provided for @separatorTab.
  ///
  /// In en, this message translates to:
  /// **'Tab'**
  String get separatorTab;

  /// No description provided for @separatorSpace.
  ///
  /// In en, this message translates to:
  /// **'Space'**
  String get separatorSpace;

  /// No description provided for @separatorPipe.
  ///
  /// In en, this message translates to:
  /// **'Pipe'**
  String get separatorPipe;

  /// No description provided for @separatorColon.
  ///
  /// In en, this message translates to:
  /// **'Colon'**
  String get separatorColon;

  /// No description provided for @scopeNoteType.
  ///
  /// In en, this message translates to:
  /// **'In its note type'**
  String get scopeNoteType;

  /// No description provided for @scopeDeck.
  ///
  /// In en, this message translates to:
  /// **'In its note type and deck'**
  String get scopeDeck;

  /// No description provided for @importedNotes.
  ///
  /// In en, this message translates to:
  /// **'Imported {count, plural, =1{1 note} other{{count} notes}}'**
  String importedNotes(int count);

  /// No description provided for @importIntro.
  ///
  /// In en, this message translates to:
  /// **'Import notes from a text file (.csv, .tsv or .txt), or decks from an Anki package (.apkg). An Anki collection package (.colpkg) replaces your collection. The file is read on this device and never uploaded.'**
  String get importIntro;

  /// No description provided for @chooseAFile.
  ///
  /// In en, this message translates to:
  /// **'Choose a file'**
  String get chooseAFile;

  /// No description provided for @importFailed.
  ///
  /// In en, this message translates to:
  /// **'The import failed'**
  String get importFailed;

  /// No description provided for @collectionNotChanged.
  ///
  /// In en, this message translates to:
  /// **'Your collection was not changed.'**
  String get collectionNotChanged;

  /// No description provided for @chooseAnotherFile.
  ///
  /// In en, this message translates to:
  /// **'Choose another file'**
  String get chooseAnotherFile;

  /// No description provided for @fileIsBackup.
  ///
  /// In en, this message translates to:
  /// **'{file} is a Cardflip backup. Restore it from Settings to replace your collection.'**
  String fileIsBackup(String file);

  /// A column of a text file without column names.
  ///
  /// In en, this message translates to:
  /// **'Column {number}'**
  String columnNumber(int number);

  /// The file name and the size of a text file.
  ///
  /// In en, this message translates to:
  /// **'{file}: {rows, plural, =1{1 row} other{{rows} rows}}, {columns, plural, =1{1 column} other{{columns} columns}}'**
  String csvSummary(String file, int rows, int columns);

  /// No description provided for @tagsForEveryNote.
  ///
  /// In en, this message translates to:
  /// **'Tags added to every note: {tags}'**
  String tagsForEveryNote(String tags);

  /// No description provided for @rowsSkippedUnknownType.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 row} other{{count} rows}} will be skipped: there is no note type named {names}.'**
  String rowsSkippedUnknownType(int count, String names);

  /// No description provided for @noteTypeOfRowsWithout.
  ///
  /// In en, this message translates to:
  /// **'Note type of rows without one'**
  String get noteTypeOfRowsWithout;

  /// No description provided for @noteType.
  ///
  /// In en, this message translates to:
  /// **'Note type'**
  String get noteType;

  /// No description provided for @deckOfRowsWithout.
  ///
  /// In en, this message translates to:
  /// **'Deck of rows without one'**
  String get deckOfRowsWithout;

  /// No description provided for @deck.
  ///
  /// In en, this message translates to:
  /// **'Deck'**
  String get deck;

  /// A deck the import creates.
  ///
  /// In en, this message translates to:
  /// **'{deck} (new)'**
  String newDeckName(String deck);

  /// No description provided for @separator.
  ///
  /// In en, this message translates to:
  /// **'Separator'**
  String get separator;

  /// No description provided for @allowHtml.
  ///
  /// In en, this message translates to:
  /// **'Allow HTML in fields'**
  String get allowHtml;

  /// No description provided for @allowHtmlHelp.
  ///
  /// In en, this message translates to:
  /// **'When off, text like \"a < b & c\" is imported as typed.'**
  String get allowHtmlHelp;

  /// No description provided for @firstRowNames.
  ///
  /// In en, this message translates to:
  /// **'The first row contains column names'**
  String get firstRowNames;

  /// No description provided for @tags.
  ///
  /// In en, this message translates to:
  /// **'Tags'**
  String get tags;

  /// A note's globally unique ID.
  ///
  /// In en, this message translates to:
  /// **'GUID'**
  String get guid;

  /// Choice for a column that is not imported.
  ///
  /// In en, this message translates to:
  /// **'Ignore'**
  String get ignore;

  /// No description provided for @whenFirstFieldMatches.
  ///
  /// In en, this message translates to:
  /// **'When the first field matches an existing note'**
  String get whenFirstFieldMatches;

  /// No description provided for @guidUpdates.
  ///
  /// In en, this message translates to:
  /// **'A row whose GUID matches a note of its note type updates that note.'**
  String get guidUpdates;

  /// No description provided for @skipRow.
  ///
  /// In en, this message translates to:
  /// **'Skip the row'**
  String get skipRow;

  /// No description provided for @updateExistingNote.
  ///
  /// In en, this message translates to:
  /// **'Update the existing note'**
  String get updateExistingNote;

  /// No description provided for @addAsDuplicate.
  ///
  /// In en, this message translates to:
  /// **'Add it as a duplicate'**
  String get addAsDuplicate;

  /// No description provided for @lookForMatchingNotes.
  ///
  /// In en, this message translates to:
  /// **'Look for matching notes'**
  String get lookForMatchingNotes;

  /// The contents of an Anki package.
  ///
  /// In en, this message translates to:
  /// **'{notes, plural, =1{1 note} other{{notes} notes}}, {cards, plural, =1{1 card} other{{cards} cards}}, {media, plural, =1{1 media file} other{{media} media files}}'**
  String apkgSummary(int notes, int cards, int media);

  /// Added to apkgSummary: the package's decks; starts with a space.
  ///
  /// In en, this message translates to:
  /// **' in {decks}'**
  String inDecks(String decks);

  /// No description provided for @unsupportedMedia.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 audio or video file} other{{count} audio or video files}} will be skipped: Cardflip cannot play their format.'**
  String unsupportedMedia(int count);

  /// No description provided for @scheduling.
  ///
  /// In en, this message translates to:
  /// **'Scheduling'**
  String get scheduling;

  /// No description provided for @keepScheduling.
  ///
  /// In en, this message translates to:
  /// **'Keep scheduling'**
  String get keepScheduling;

  /// No description provided for @keepSchedulingHelp.
  ///
  /// In en, this message translates to:
  /// **'Cards keep their due dates and progress from Anki.'**
  String get keepSchedulingHelp;

  /// No description provided for @importAsNew.
  ///
  /// In en, this message translates to:
  /// **'Import all cards as new'**
  String get importAsNew;

  /// No description provided for @importAsNewHelp.
  ///
  /// In en, this message translates to:
  /// **'Start learning everything from scratch.'**
  String get importAsNewHelp;

  /// No description provided for @notesAlreadyThere.
  ///
  /// In en, this message translates to:
  /// **'Notes already in your collection'**
  String get notesAlreadyThere;

  /// No description provided for @notesAlreadyThereHelp.
  ///
  /// In en, this message translates to:
  /// **'Notes and note types from an earlier import of this deck. Their scheduling is kept either way.'**
  String get notesAlreadyThereHelp;

  /// No description provided for @updateNotes.
  ///
  /// In en, this message translates to:
  /// **'Update notes'**
  String get updateNotes;

  /// No description provided for @updateNoteTypes.
  ///
  /// In en, this message translates to:
  /// **'Update note types'**
  String get updateNoteTypes;

  /// No description provided for @mergeNoteTypes.
  ///
  /// In en, this message translates to:
  /// **'Merge note types'**
  String get mergeNoteTypes;

  /// No description provided for @mergeNoteTypesHelp.
  ///
  /// In en, this message translates to:
  /// **'When the package adds fields or card types, add them to your note type. Otherwise the package\'\'s version is imported as a separate note type.'**
  String get mergeNoteTypesHelp;

  /// No description provided for @replaceByImportBody.
  ///
  /// In en, this message translates to:
  /// **'Importing {file} replaces your current collection entirely: all decks, notes, cards, review history and media files are overwritten. Your settings are kept. This cannot be undone.'**
  String replaceByImportBody(String file);

  /// Button that replaces the collection.
  ///
  /// In en, this message translates to:
  /// **'Replace'**
  String get replace;

  /// No description provided for @colpkgSummary.
  ///
  /// In en, this message translates to:
  /// **'Anki collection: {notes, plural, =1{1 note} other{{notes} notes}}, {cards, plural, =1{1 card} other{{cards} cards}}, {media, plural, =1{1 media file} other{{media} media files}} in {decks, plural, =1{1 deck} other{{decks} decks}}'**
  String colpkgSummary(int notes, int cards, int media, int decks);

  /// No description provided for @replaceWarning.
  ///
  /// In en, this message translates to:
  /// **'This replaces your current collection with the Anki collection, with its scheduling and review history. Everything you have in Cardflip now is overwritten; your settings are kept. Cardflip makes an automatic backup first, but you may want to export a backup to a file before you continue.'**
  String get replaceWarning;

  /// No description provided for @exportBackupFirst.
  ///
  /// In en, this message translates to:
  /// **'Export a backup first'**
  String get exportBackupFirst;

  /// No description provided for @replaceCollection.
  ///
  /// In en, this message translates to:
  /// **'Replace collection'**
  String get replaceCollection;

  /// No description provided for @addingNotes.
  ///
  /// In en, this message translates to:
  /// **'Adding notes'**
  String get addingNotes;

  /// No description provided for @addingHistory.
  ///
  /// In en, this message translates to:
  /// **'Adding review history'**
  String get addingHistory;

  /// No description provided for @savingImages.
  ///
  /// In en, this message translates to:
  /// **'Saving images'**
  String get savingImages;

  /// No description provided for @saving.
  ///
  /// In en, this message translates to:
  /// **'Saving'**
  String get saving;

  /// No description provided for @readingFile.
  ///
  /// In en, this message translates to:
  /// **'Reading the file'**
  String get readingFile;

  /// Progress, e.g. "12 of 340".
  ///
  /// In en, this message translates to:
  /// **'{done} of {total}'**
  String progressOf(int done, int total);

  /// No description provided for @notesAdded.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} added'**
  String notesAdded(int count);

  /// No description provided for @notesUpdated.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} updated'**
  String notesUpdated(int count);

  /// No description provided for @notesSkipped.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} skipped'**
  String notesSkipped(int count);

  /// No description provided for @notesNotUpdatedNewer.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} not updated: your version is newer'**
  String notesNotUpdatedNewer(int count);

  /// No description provided for @cardsAdded.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}} added'**
  String cardsAdded(int count);

  /// No description provided for @reviewsImported.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 review} other{{count} reviews}} of history imported'**
  String reviewsImported(int count);

  /// No description provided for @mediaImported.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 media file} other{{count} media files}} imported'**
  String mediaImported(int count);

  /// No description provided for @commentLinesSkipped.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 comment line} other{{count} comment lines}} skipped'**
  String commentLinesSkipped(int count);

  /// No description provided for @noteTypesAdded.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note type} other{{count} note types}} added'**
  String noteTypesAdded(int count);

  /// No description provided for @noteTypesUpdated.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note type} other{{count} note types}} updated'**
  String noteTypesUpdated(int count);

  /// No description provided for @noteTypesMerged.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note type} other{{count} note types}} merged'**
  String noteTypesMerged(int count);

  /// No description provided for @collectionReplaced.
  ///
  /// In en, this message translates to:
  /// **'Collection replaced'**
  String get collectionReplaced;

  /// No description provided for @importComplete.
  ///
  /// In en, this message translates to:
  /// **'Import complete'**
  String get importComplete;

  /// No description provided for @warnings.
  ///
  /// In en, this message translates to:
  /// **'Warnings'**
  String get warnings;

  /// No description provided for @done.
  ///
  /// In en, this message translates to:
  /// **'Done'**
  String get done;

  /// The suggested name of a copy, e.g. of a note type.
  ///
  /// In en, this message translates to:
  /// **'{name} copy'**
  String copyName(String name);

  /// No description provided for @addNoteType.
  ///
  /// In en, this message translates to:
  /// **'Add note type'**
  String get addNoteType;

  /// Title of the dialog that copies a note type.
  ///
  /// In en, this message translates to:
  /// **'Clone “{name}”'**
  String cloneNamed(String name);

  /// No description provided for @noNotesUseType.
  ///
  /// In en, this message translates to:
  /// **'No notes use this note type.'**
  String get noNotesUseType;

  /// No description provided for @deleteNoteTypeBody.
  ///
  /// In en, this message translates to:
  /// **'This deletes {notes, plural, =1{1 note} other{{notes} notes}} and {cards, plural, =1{1 card} other{{cards} cards}}.'**
  String deleteNoteTypeBody(int notes, int cards);

  /// No description provided for @canUndoUntilClose.
  ///
  /// In en, this message translates to:
  /// **'You can undo this until you close Cardflip.'**
  String get canUndoUntilClose;

  /// No description provided for @noteTypes.
  ///
  /// In en, this message translates to:
  /// **'Note types'**
  String get noteTypes;

  /// The name of a new collection's first deck.
  ///
  /// In en, this message translates to:
  /// **'Default'**
  String get builtinDeckName;

  /// The name of a new collection's deck options preset.
  ///
  /// In en, this message translates to:
  /// **'Default'**
  String get builtinPresetName;

  /// A stock note type: one card with front and back.
  ///
  /// In en, this message translates to:
  /// **'Basic'**
  String get stockBasic;

  /// A stock note type: a card each way.
  ///
  /// In en, this message translates to:
  /// **'Basic (and reversed card)'**
  String get stockReversed;

  /// A stock note type: the reverse card only when its field is filled in.
  ///
  /// In en, this message translates to:
  /// **'Basic (optional reversed card)'**
  String get stockOptionalReversed;

  /// A stock note type whose card asks to type the answer.
  ///
  /// In en, this message translates to:
  /// **'Basic (type in the answer)'**
  String get stockTypeIn;

  /// The stock cloze note type, and its card type.
  ///
  /// In en, this message translates to:
  /// **'Cloze'**
  String get stockCloze;

  /// The stock Image Occlusion note type, and its card type.
  ///
  /// In en, this message translates to:
  /// **'Image Occlusion'**
  String get stockImageOcclusion;

  /// A field of the stock Basic note types.
  ///
  /// In en, this message translates to:
  /// **'Front'**
  String get stockFieldFront;

  /// A field of the stock Basic note types.
  ///
  /// In en, this message translates to:
  /// **'Back'**
  String get stockFieldBack;

  /// The field of Basic (optional reversed card) that adds the reverse card when filled in.
  ///
  /// In en, this message translates to:
  /// **'Add Reverse'**
  String get stockFieldAddReverse;

  /// The field of the stock Cloze note type with the cloze deletions.
  ///
  /// In en, this message translates to:
  /// **'Text'**
  String get stockFieldText;

  /// A field of the stock Cloze and Image Occlusion note types shown on the answer.
  ///
  /// In en, this message translates to:
  /// **'Back Extra'**
  String get stockFieldBackExtra;

  /// The field of the stock Image Occlusion note type with the shapes.
  ///
  /// In en, this message translates to:
  /// **'Occlusion'**
  String get stockFieldOcclusion;

  /// The field of the stock Image Occlusion note type with the image.
  ///
  /// In en, this message translates to:
  /// **'Image'**
  String get stockFieldImage;

  /// A field of the stock Image Occlusion note type shown above the image.
  ///
  /// In en, this message translates to:
  /// **'Header'**
  String get stockFieldHeader;

  /// A field of the stock Image Occlusion note type for notes to oneself.
  ///
  /// In en, this message translates to:
  /// **'Comments'**
  String get stockFieldComments;

  /// No description provided for @noteTypeSummary.
  ///
  /// In en, this message translates to:
  /// **'{fields, plural, =1{1 field} other{{fields} fields}} · {cardTypes, plural, =1{1 card type} other{{cardTypes} card types}} · {notes, plural, =1{1 note} other{{notes} notes}}'**
  String noteTypeSummary(int fields, int cardTypes, int notes);

  /// No description provided for @noteTypeActions.
  ///
  /// In en, this message translates to:
  /// **'Note type actions'**
  String get noteTypeActions;

  /// No description provided for @renameNoteType.
  ///
  /// In en, this message translates to:
  /// **'Rename note type'**
  String get renameNoteType;

  /// No description provided for @renameField.
  ///
  /// In en, this message translates to:
  /// **'Rename field'**
  String get renameField;

  /// No description provided for @addField.
  ///
  /// In en, this message translates to:
  /// **'Add field'**
  String get addField;

  /// No description provided for @removeFieldTitle.
  ///
  /// In en, this message translates to:
  /// **'Remove field “{field}”?'**
  String removeFieldTitle(String field);

  /// No description provided for @removeFieldBody.
  ///
  /// In en, this message translates to:
  /// **'The field and its content will be removed from {count, plural, =1{1 note} other{{count} notes}}. You can undo this until you close Cardflip.'**
  String removeFieldBody(int count);

  /// Button that removes something.
  ///
  /// In en, this message translates to:
  /// **'Remove'**
  String get remove;

  /// The name of a new card type, e.g. Card 2.
  ///
  /// In en, this message translates to:
  /// **'Card {number}'**
  String defaultCardTypeName(int number);

  /// No description provided for @removeCardTypeTitle.
  ///
  /// In en, this message translates to:
  /// **'Remove card type “{name}”?'**
  String removeCardTypeTitle(String name);

  /// No description provided for @removeCardTypeBody.
  ///
  /// In en, this message translates to:
  /// **'This deletes {count, plural, =1{1 card} other{{count} cards}} and their review history cannot be used anymore.'**
  String removeCardTypeBody(int count);

  /// No description provided for @fields.
  ///
  /// In en, this message translates to:
  /// **'Fields'**
  String get fields;

  /// No description provided for @fieldOptions.
  ///
  /// In en, this message translates to:
  /// **'Field options'**
  String get fieldOptions;

  /// No description provided for @removeField.
  ///
  /// In en, this message translates to:
  /// **'Remove field'**
  String get removeField;

  /// No description provided for @dragToReorder.
  ///
  /// In en, this message translates to:
  /// **'Drag to reorder'**
  String get dragToReorder;

  /// No description provided for @cardTypes.
  ///
  /// In en, this message translates to:
  /// **'Card types'**
  String get cardTypes;

  /// No description provided for @cardTypeOptions.
  ///
  /// In en, this message translates to:
  /// **'Card type options'**
  String get cardTypeOptions;

  /// No description provided for @removeCardType.
  ///
  /// In en, this message translates to:
  /// **'Remove card type'**
  String get removeCardType;

  /// No description provided for @addCardType.
  ///
  /// In en, this message translates to:
  /// **'Add card type'**
  String get addCardType;

  /// Heading of a note type's CSS.
  ///
  /// In en, this message translates to:
  /// **'Styling'**
  String get styling;

  /// No description provided for @editCss.
  ///
  /// In en, this message translates to:
  /// **'Edit CSS'**
  String get editCss;

  /// The field the browser's Sort field column shows.
  ///
  /// In en, this message translates to:
  /// **'Sort field'**
  String get sortField;

  /// No description provided for @rightToLeft.
  ///
  /// In en, this message translates to:
  /// **'Right to left'**
  String get rightToLeft;

  /// No description provided for @collapsed.
  ///
  /// In en, this message translates to:
  /// **'Collapsed'**
  String get collapsed;

  /// No description provided for @notInUnqualifiedSearches.
  ///
  /// In en, this message translates to:
  /// **'Not in unqualified searches'**
  String get notInUnqualifiedSearches;

  /// No description provided for @htmlEditor.
  ///
  /// In en, this message translates to:
  /// **'HTML editor'**
  String get htmlEditor;

  /// A field whose content stays after adding a note.
  ///
  /// In en, this message translates to:
  /// **'Pinned'**
  String get pinned;

  /// No description provided for @fieldSizeError.
  ///
  /// In en, this message translates to:
  /// **'Enter a size from 1 to 200, or leave it empty.'**
  String get fieldSizeError;

  /// Title of the options of a field or card type.
  ///
  /// In en, this message translates to:
  /// **'Options for “{name}”'**
  String optionsForNamed(String name);

  /// No description provided for @editingFont.
  ///
  /// In en, this message translates to:
  /// **'Editing font'**
  String get editingFont;

  /// No description provided for @editingFontHelp.
  ///
  /// In en, this message translates to:
  /// **'Used for this field in the editor only, not on cards.'**
  String get editingFontHelp;

  /// No description provided for @font.
  ///
  /// In en, this message translates to:
  /// **'Font'**
  String get font;

  /// Hint of an empty field that uses the default.
  ///
  /// In en, this message translates to:
  /// **'Default'**
  String get defaultValue;

  /// No description provided for @size.
  ///
  /// In en, this message translates to:
  /// **'Size'**
  String get size;

  /// No description provided for @sortByField.
  ///
  /// In en, this message translates to:
  /// **'Sort by this field'**
  String get sortByField;

  /// No description provided for @sortByFieldIsHelp.
  ///
  /// In en, this message translates to:
  /// **'The browser’s Sort field column shows this field. To change it, choose another field.'**
  String get sortByFieldIsHelp;

  /// No description provided for @sortByFieldHelp.
  ///
  /// In en, this message translates to:
  /// **'The browser’s Sort field column shows this field.'**
  String get sortByFieldHelp;

  /// No description provided for @rightToLeftHelp.
  ///
  /// In en, this message translates to:
  /// **'For Arabic, Hebrew and other right-to-left scripts.'**
  String get rightToLeftHelp;

  /// No description provided for @collapseByDefault.
  ///
  /// In en, this message translates to:
  /// **'Collapse by default'**
  String get collapseByDefault;

  /// No description provided for @collapseByDefaultHelp.
  ///
  /// In en, this message translates to:
  /// **'The editor shows only the field name until you tap it.'**
  String get collapseByDefaultHelp;

  /// No description provided for @excludeFromSearch.
  ///
  /// In en, this message translates to:
  /// **'Exclude from unqualified searches'**
  String get excludeFromSearch;

  /// {field} is the field name in lower case.
  ///
  /// In en, this message translates to:
  /// **'Plain search terms skip this field; {field}:… still finds it.'**
  String excludeFromSearchHelp(String field);

  /// No description provided for @htmlByDefault.
  ///
  /// In en, this message translates to:
  /// **'Use HTML editor by default'**
  String get htmlByDefault;

  /// No description provided for @htmlByDefaultHelp.
  ///
  /// In en, this message translates to:
  /// **'Open the field as HTML source instead of the visual editor.'**
  String get htmlByDefaultHelp;

  /// No description provided for @keepAfterAdding.
  ///
  /// In en, this message translates to:
  /// **'Keep after adding'**
  String get keepAfterAdding;

  /// No description provided for @keepAfterAddingHelp.
  ///
  /// In en, this message translates to:
  /// **'Pin the field, so its content stays in the add screen after adding a note.'**
  String get keepAfterAddingHelp;

  /// No description provided for @deckOverride.
  ///
  /// In en, this message translates to:
  /// **'Deck override'**
  String get deckOverride;

  /// No description provided for @deckOverrideHelp.
  ///
  /// In en, this message translates to:
  /// **'New cards of this card type go to this deck. Existing cards are not moved.'**
  String get deckOverrideHelp;

  /// No description provided for @deckOverrideNone.
  ///
  /// In en, this message translates to:
  /// **'None (the deck chosen when adding)'**
  String get deckOverrideNone;

  /// No description provided for @browserAppearance.
  ///
  /// In en, this message translates to:
  /// **'Browser appearance'**
  String get browserAppearance;

  /// No description provided for @browserAppearanceHelp.
  ///
  /// In en, this message translates to:
  /// **'Simplified templates for the browser\'\'s question and answer columns. Leave empty to use the normal templates.'**
  String get browserAppearanceHelp;

  /// No description provided for @browserQuestionTemplate.
  ///
  /// In en, this message translates to:
  /// **'Browser question template'**
  String get browserQuestionTemplate;

  /// No description provided for @browserAnswerTemplate.
  ///
  /// In en, this message translates to:
  /// **'Browser answer template'**
  String get browserAnswerTemplate;

  /// Tab that edits a template.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get edit;

  /// Tab that previews a template.
  ///
  /// In en, this message translates to:
  /// **'Preview'**
  String get preview;

  /// No description provided for @frontTemplate.
  ///
  /// In en, this message translates to:
  /// **'Front template'**
  String get frontTemplate;

  /// No description provided for @backTemplate.
  ///
  /// In en, this message translates to:
  /// **'Back template'**
  String get backTemplate;

  /// No description provided for @templateHelp.
  ///
  /// In en, this message translates to:
  /// **'Insert a field with \'{{Field}}\'. Add a filter in front of the name, e.g. \'{{type}}\' for a type-in answer box, \'{{hint}}\', \'{{text}}\' or \'{{cloze}}\'. Also \'{{FrontSide}}\' and sections \'{{#Field}}…{{/Field}}\'.'**
  String get templateHelp;

  /// No description provided for @newCardType.
  ///
  /// In en, this message translates to:
  /// **'New card type'**
  String get newCardType;

  /// No description provided for @editCardType.
  ///
  /// In en, this message translates to:
  /// **'Edit card type'**
  String get editCardType;

  /// No description provided for @stylingCss.
  ///
  /// In en, this message translates to:
  /// **'Styling (CSS)'**
  String get stylingCss;

  /// No description provided for @lossyNote.
  ///
  /// In en, this message translates to:
  /// **'This field contains formatting the visual editor can\'\'t keep. Editing as HTML.'**
  String get lossyNote;

  /// No description provided for @switchToVisualTitle.
  ///
  /// In en, this message translates to:
  /// **'Switch to the visual editor?'**
  String get switchToVisualTitle;

  /// No description provided for @switchToVisualBody.
  ///
  /// In en, this message translates to:
  /// **'This field contains formatting the visual editor can\'\'t keep. If you edit it there, that formatting may be lost.'**
  String get switchToVisualBody;

  /// Button that switches to the visual editor.
  ///
  /// In en, this message translates to:
  /// **'Switch'**
  String get switchAction;

  /// No description provided for @expandField.
  ///
  /// In en, this message translates to:
  /// **'Expand field'**
  String get expandField;

  /// No description provided for @collapseField.
  ///
  /// In en, this message translates to:
  /// **'Collapse field'**
  String get collapseField;

  /// No description provided for @visualEditor.
  ///
  /// In en, this message translates to:
  /// **'Visual editor'**
  String get visualEditor;

  /// A collapsed field without content.
  ///
  /// In en, this message translates to:
  /// **'Empty'**
  String get emptyField;

  /// No description provided for @bold.
  ///
  /// In en, this message translates to:
  /// **'Bold'**
  String get bold;

  /// No description provided for @italic.
  ///
  /// In en, this message translates to:
  /// **'Italic'**
  String get italic;

  /// No description provided for @underline.
  ///
  /// In en, this message translates to:
  /// **'Underline'**
  String get underline;

  /// No description provided for @subscript.
  ///
  /// In en, this message translates to:
  /// **'Subscript'**
  String get subscript;

  /// No description provided for @superscript.
  ///
  /// In en, this message translates to:
  /// **'Superscript'**
  String get superscript;

  /// No description provided for @textColour.
  ///
  /// In en, this message translates to:
  /// **'Text colour'**
  String get textColour;

  /// No description provided for @highlightColour.
  ///
  /// In en, this message translates to:
  /// **'Highlight colour'**
  String get highlightColour;

  /// No description provided for @removeFormatting.
  ///
  /// In en, this message translates to:
  /// **'Remove formatting'**
  String get removeFormatting;

  /// No description provided for @bulletedList.
  ///
  /// In en, this message translates to:
  /// **'Bulleted list'**
  String get bulletedList;

  /// No description provided for @numberedList.
  ///
  /// In en, this message translates to:
  /// **'Numbered list'**
  String get numberedList;

  /// No description provided for @alignment.
  ///
  /// In en, this message translates to:
  /// **'Alignment'**
  String get alignment;

  /// No description provided for @alignLeft.
  ///
  /// In en, this message translates to:
  /// **'Align left'**
  String get alignLeft;

  /// No description provided for @alignCenter.
  ///
  /// In en, this message translates to:
  /// **'Center'**
  String get alignCenter;

  /// No description provided for @alignRight.
  ///
  /// In en, this message translates to:
  /// **'Align right'**
  String get alignRight;

  /// No description provided for @justify.
  ///
  /// In en, this message translates to:
  /// **'Justify'**
  String get justify;

  /// No description provided for @outdent.
  ///
  /// In en, this message translates to:
  /// **'Outdent'**
  String get outdent;

  /// No description provided for @indent.
  ///
  /// In en, this message translates to:
  /// **'Indent'**
  String get indent;

  /// No description provided for @colourBlack.
  ///
  /// In en, this message translates to:
  /// **'Black'**
  String get colourBlack;

  /// No description provided for @colourGray.
  ///
  /// In en, this message translates to:
  /// **'Gray'**
  String get colourGray;

  /// No description provided for @colourRed.
  ///
  /// In en, this message translates to:
  /// **'Red'**
  String get colourRed;

  /// No description provided for @colourOrange.
  ///
  /// In en, this message translates to:
  /// **'Orange'**
  String get colourOrange;

  /// No description provided for @colourYellow.
  ///
  /// In en, this message translates to:
  /// **'Yellow'**
  String get colourYellow;

  /// No description provided for @colourGreen.
  ///
  /// In en, this message translates to:
  /// **'Green'**
  String get colourGreen;

  /// No description provided for @colourBlue.
  ///
  /// In en, this message translates to:
  /// **'Blue'**
  String get colourBlue;

  /// No description provided for @colourPurple.
  ///
  /// In en, this message translates to:
  /// **'Purple'**
  String get colourPurple;

  /// No description provided for @noColour.
  ///
  /// In en, this message translates to:
  /// **'No colour'**
  String get noColour;

  /// No description provided for @mathInline.
  ///
  /// In en, this message translates to:
  /// **'MathJax inline'**
  String get mathInline;

  /// No description provided for @mathBlock.
  ///
  /// In en, this message translates to:
  /// **'MathJax block'**
  String get mathBlock;

  /// No description provided for @mathChemistry.
  ///
  /// In en, this message translates to:
  /// **'MathJax chemistry'**
  String get mathChemistry;

  /// No description provided for @mathLatex.
  ///
  /// In en, this message translates to:
  /// **'LaTeX'**
  String get mathLatex;

  /// No description provided for @mathLatexEquation.
  ///
  /// In en, this message translates to:
  /// **'LaTeX equation'**
  String get mathLatexEquation;

  /// No description provided for @mathLatexMath.
  ///
  /// In en, this message translates to:
  /// **'LaTeX math env.'**
  String get mathLatexMath;

  /// Tooltip of the menu that inserts math.
  ///
  /// In en, this message translates to:
  /// **'Math'**
  String get math;

  /// No description provided for @clozeDeletion.
  ///
  /// In en, this message translates to:
  /// **'Cloze deletion'**
  String get clozeDeletion;

  /// No description provided for @nextClozeNumber.
  ///
  /// In en, this message translates to:
  /// **'Next cloze number'**
  String get nextClozeNumber;

  /// No description provided for @sameClozeNumber.
  ///
  /// In en, this message translates to:
  /// **'Same cloze number'**
  String get sameClozeNumber;

  /// No description provided for @attachMedia.
  ///
  /// In en, this message translates to:
  /// **'Attach picture, audio or video'**
  String get attachMedia;

  /// No description provided for @recordAudio.
  ///
  /// In en, this message translates to:
  /// **'Record audio'**
  String get recordAudio;

  /// No description provided for @onlyMediaInFields.
  ///
  /// In en, this message translates to:
  /// **'Only pictures, audio and video can be added to a field.'**
  String get onlyMediaInFields;

  /// No description provided for @noNotesYet.
  ///
  /// In en, this message translates to:
  /// **'No notes yet'**
  String get noNotesYet;

  /// A cloze deletion to preview, e.g. Cloze 2.
  ///
  /// In en, this message translates to:
  /// **'Cloze {number}'**
  String clozeNumber(int number);

  /// No description provided for @previewLightTheme.
  ///
  /// In en, this message translates to:
  /// **'Preview light theme'**
  String get previewLightTheme;

  /// No description provided for @previewDarkTheme.
  ///
  /// In en, this message translates to:
  /// **'Preview dark theme'**
  String get previewDarkTheme;

  /// The question side of a card.
  ///
  /// In en, this message translates to:
  /// **'Front'**
  String get front;

  /// The answer side of a card.
  ///
  /// In en, this message translates to:
  /// **'Back'**
  String get back;

  /// No description provided for @recording.
  ///
  /// In en, this message translates to:
  /// **'Recording…'**
  String get recording;

  /// Button that inserts a recording.
  ///
  /// In en, this message translates to:
  /// **'Insert'**
  String get insert;

  /// No description provided for @addTags.
  ///
  /// In en, this message translates to:
  /// **'Add tags'**
  String get addTags;

  /// No description provided for @imageMissing.
  ///
  /// In en, this message translates to:
  /// **'The image {file} is missing.'**
  String imageMissing(String file);

  /// No description provided for @choosePictureToDrawOn.
  ///
  /// In en, this message translates to:
  /// **'Choose a picture to draw on.'**
  String get choosePictureToDrawOn;

  /// No description provided for @chooseImageToDrawOn.
  ///
  /// In en, this message translates to:
  /// **'Choose an image to draw on.'**
  String get chooseImageToDrawOn;

  /// No description provided for @drawShapes.
  ///
  /// In en, this message translates to:
  /// **'Draw a shape over each part to ask.'**
  String get drawShapes;

  /// No description provided for @cardAdded.
  ///
  /// In en, this message translates to:
  /// **'Card added'**
  String get cardAdded;

  /// No description provided for @editCard.
  ///
  /// In en, this message translates to:
  /// **'Edit card'**
  String get editCard;

  /// No description provided for @historyAddedToday.
  ///
  /// In en, this message translates to:
  /// **'History: cards added today'**
  String get historyAddedToday;

  /// Heading of an Image Occlusion note's image.
  ///
  /// In en, this message translates to:
  /// **'Image'**
  String get image;

  /// No description provided for @changeImage.
  ///
  /// In en, this message translates to:
  /// **'Change image'**
  String get changeImage;

  /// No description provided for @loadingImage.
  ///
  /// In en, this message translates to:
  /// **'Loading the image…'**
  String get loadingImage;

  /// No description provided for @chooseImage.
  ///
  /// In en, this message translates to:
  /// **'Choose image'**
  String get chooseImage;

  /// No description provided for @duplicateOne.
  ///
  /// In en, this message translates to:
  /// **'This is a duplicate: another note has the same {field} field.'**
  String duplicateOne(String field);

  /// No description provided for @duplicateMany.
  ///
  /// In en, this message translates to:
  /// **'This is a duplicate: {count} other notes have the same {field} field.'**
  String duplicateMany(int count, String field);

  /// No description provided for @showDuplicates.
  ///
  /// In en, this message translates to:
  /// **'Show duplicates'**
  String get showDuplicates;

  /// No description provided for @unpinField.
  ///
  /// In en, this message translates to:
  /// **'Unpin: clear after adding'**
  String get unpinField;

  /// No description provided for @pinField.
  ///
  /// In en, this message translates to:
  /// **'Pin: keep after adding'**
  String get pinField;

  /// No description provided for @hideAllGuessOne.
  ///
  /// In en, this message translates to:
  /// **'Hide all, guess one'**
  String get hideAllGuessOne;

  /// No description provided for @hideOneGuessOne.
  ///
  /// In en, this message translates to:
  /// **'Hide one, guess one'**
  String get hideOneGuessOne;

  /// No description provided for @selectAndMove.
  ///
  /// In en, this message translates to:
  /// **'Select and move'**
  String get selectAndMove;

  /// No description provided for @rectangle.
  ///
  /// In en, this message translates to:
  /// **'Rectangle'**
  String get rectangle;

  /// No description provided for @ellipse.
  ///
  /// In en, this message translates to:
  /// **'Ellipse'**
  String get ellipse;

  /// No description provided for @polygonTool.
  ///
  /// In en, this message translates to:
  /// **'Polygon: tap the corners, then the first one again'**
  String get polygonTool;

  /// No description provided for @textLabel.
  ///
  /// In en, this message translates to:
  /// **'Text label'**
  String get textLabel;

  /// No description provided for @finishPolygon.
  ///
  /// In en, this message translates to:
  /// **'Finish the polygon'**
  String get finishPolygon;

  /// No description provided for @groupShapes.
  ///
  /// In en, this message translates to:
  /// **'Group: ask the selected shapes together'**
  String get groupShapes;

  /// No description provided for @ungroup.
  ///
  /// In en, this message translates to:
  /// **'Ungroup'**
  String get ungroup;

  /// No description provided for @deleteSelectedShapes.
  ///
  /// In en, this message translates to:
  /// **'Delete the selected shapes'**
  String get deleteSelectedShapes;

  /// No description provided for @zoomOut.
  ///
  /// In en, this message translates to:
  /// **'Zoom out'**
  String get zoomOut;

  /// No description provided for @zoomIn.
  ///
  /// In en, this message translates to:
  /// **'Zoom in'**
  String get zoomIn;

  /// No description provided for @fitImage.
  ///
  /// In en, this message translates to:
  /// **'Fit the image'**
  String get fitImage;

  /// No description provided for @imageCannotShow.
  ///
  /// In en, this message translates to:
  /// **'This image cannot be shown.'**
  String get imageCannotShow;

  /// No description provided for @shownOnEveryCard.
  ///
  /// In en, this message translates to:
  /// **'Shown on every card'**
  String get shownOnEveryCard;

  /// No description provided for @repositionNewCards.
  ///
  /// In en, this message translates to:
  /// **'Reposition new cards'**
  String get repositionNewCards;

  /// No description provided for @startPosition.
  ///
  /// In en, this message translates to:
  /// **'Start position'**
  String get startPosition;

  /// How far apart repositioned cards are.
  ///
  /// In en, this message translates to:
  /// **'Step'**
  String get step;

  /// No description provided for @randomizeOrder.
  ///
  /// In en, this message translates to:
  /// **'Randomize order'**
  String get randomizeOrder;

  /// No description provided for @shiftExisting.
  ///
  /// In en, this message translates to:
  /// **'Shift position of existing cards'**
  String get shiftExisting;

  /// No description provided for @repositionHelp.
  ///
  /// In en, this message translates to:
  /// **'Only new cards are repositioned; others are left unchanged.'**
  String get repositionHelp;

  /// Button that moves new cards in the queue.
  ///
  /// In en, this message translates to:
  /// **'Reposition'**
  String get reposition;

  /// Error of a number field.
  ///
  /// In en, this message translates to:
  /// **'A whole number from {min} to {max}'**
  String wholeNumberRange(int min, int max);

  /// No description provided for @setDueDate.
  ///
  /// In en, this message translates to:
  /// **'Set due date'**
  String get setDueDate;

  /// No description provided for @setDueDateOf.
  ///
  /// In en, this message translates to:
  /// **'Set due date of {count, plural, =1{1 card} other{{count} cards}}'**
  String setDueDateOf(int count);

  /// No description provided for @daysFromToday.
  ///
  /// In en, this message translates to:
  /// **'Days from today'**
  String get daysFromToday;

  /// No description provided for @dueDateHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. 0, 7, 60-90 or 30!'**
  String get dueDateHint;

  /// No description provided for @dueDateHelp.
  ///
  /// In en, this message translates to:
  /// **'0 = today, 1 = tomorrow, 7 = in a week\n60-90 = a random day 60 to 90 days from now, for each card\nAdd ! to set the interval to the same number of days, e.g. 30!'**
  String get dueDateHelp;

  /// No description provided for @deleteCardsTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete {count, plural, =1{1 card} other{{count} cards}}?'**
  String deleteCardsTitle(int count);

  /// No description provided for @resetCardsTitle.
  ///
  /// In en, this message translates to:
  /// **'Reset {count, plural, =1{1 card} other{{count} cards}} to new?'**
  String resetCardsTitle(int count);

  /// No description provided for @resetCardsBody.
  ///
  /// In en, this message translates to:
  /// **'Their scheduling is cleared and they will be studied as new cards. Their review history is kept.'**
  String get resetCardsBody;

  /// No description provided for @restorePosition.
  ///
  /// In en, this message translates to:
  /// **'Restore original position where possible'**
  String get restorePosition;

  /// No description provided for @resetCounts.
  ///
  /// In en, this message translates to:
  /// **'Reset repetition and lapse counts'**
  String get resetCounts;

  /// Button that resets cards to new.
  ///
  /// In en, this message translates to:
  /// **'Reset'**
  String get reset;

  /// No description provided for @noFlag.
  ///
  /// In en, this message translates to:
  /// **'No flag'**
  String get noFlag;

  /// Snackbar after tagging notes, e.g. "Tagged 3 notes duplicate".
  ///
  /// In en, this message translates to:
  /// **'Tagged {count, plural, =1{1 note} other{{count} notes}} {tag}'**
  String taggedNotes(int count, String tag);

  /// No description provided for @findDuplicates.
  ///
  /// In en, this message translates to:
  /// **'Find duplicates'**
  String get findDuplicates;

  /// No description provided for @tagDuplicates.
  ///
  /// In en, this message translates to:
  /// **'Tag duplicates'**
  String get tagDuplicates;

  /// No description provided for @field.
  ///
  /// In en, this message translates to:
  /// **'Field'**
  String get field;

  /// No description provided for @optionalSearch.
  ///
  /// In en, this message translates to:
  /// **'Optional search'**
  String get optionalSearch;

  /// No description provided for @optionalSearchHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. deck:Spanish'**
  String get optionalSearchHint;

  /// Button that starts a search.
  ///
  /// In en, this message translates to:
  /// **'Find'**
  String get find;

  /// No description provided for @noDuplicatesFound.
  ///
  /// In en, this message translates to:
  /// **'No duplicates found.'**
  String get noDuplicatesFound;

  /// No description provided for @countNotes.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}}'**
  String countNotes(int count);

  /// No description provided for @includingOtherCardsNote.
  ///
  /// In en, this message translates to:
  /// **'Including {count, plural, =1{1 other card} other{{count} other cards}} from the same note. '**
  String includingOtherCardsNote(int count);

  /// No description provided for @includingOtherCardsNotes.
  ///
  /// In en, this message translates to:
  /// **'Including {count, plural, =1{1 other card} other{{count} other cards}} from the same notes. '**
  String includingOtherCardsNotes(int count);

  /// A new card's position in the new card queue.
  ///
  /// In en, this message translates to:
  /// **'New #{position}'**
  String newPosition(int position);

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Question'**
  String get colQuestion;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Answer'**
  String get colAnswer;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Sort field'**
  String get colSortField;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Card type'**
  String get colCardType;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Note type'**
  String get colNoteType;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Deck'**
  String get colDeck;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Due'**
  String get colDue;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Interval'**
  String get colInterval;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Difficulty'**
  String get colDifficulty;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Stability'**
  String get colStability;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Retrievability'**
  String get colRetrievability;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Reviews'**
  String get colReviews;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Lapses'**
  String get colLapses;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Tags'**
  String get colTags;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Created'**
  String get colCreated;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Card modified'**
  String get colCardModified;

  /// Browser column.
  ///
  /// In en, this message translates to:
  /// **'Note modified'**
  String get colNoteModified;

  /// Browser column of Notes mode.
  ///
  /// In en, this message translates to:
  /// **'Cards'**
  String get colCards;

  /// A browser row whose question is empty.
  ///
  /// In en, this message translates to:
  /// **'(empty)'**
  String get emptyQuestion;

  /// Due column of a suspended card.
  ///
  /// In en, this message translates to:
  /// **'suspended'**
  String get stateSuspended;

  /// Due column of a new card without a position.
  ///
  /// In en, this message translates to:
  /// **'new'**
  String get stateNew;

  /// Due column, e.g. "due tomorrow".
  ///
  /// In en, this message translates to:
  /// **'due {when}'**
  String dueWhen(String when);

  /// Interval column of a card without an interval.
  ///
  /// In en, this message translates to:
  /// **'({state, select, new{new} learning{learning} relearning{relearning} other{{state}}})'**
  String intervalOfState(String state);

  /// Deck column of a note whose cards are in several decks.
  ///
  /// In en, this message translates to:
  /// **'(multiple)'**
  String get multipleDecks;

  /// No description provided for @moveLeft.
  ///
  /// In en, this message translates to:
  /// **'Move left'**
  String get moveLeft;

  /// No description provided for @moveRight.
  ///
  /// In en, this message translates to:
  /// **'Move right'**
  String get moveRight;

  /// No description provided for @hideColumn.
  ///
  /// In en, this message translates to:
  /// **'Hide column'**
  String get hideColumn;

  /// No description provided for @columnsEllipsis.
  ///
  /// In en, this message translates to:
  /// **'Columns…'**
  String get columnsEllipsis;

  /// No description provided for @columns.
  ///
  /// In en, this message translates to:
  /// **'Columns'**
  String get columns;

  /// A card's state, or the kind of a review: new, learning, review, relearning or manual (a change made by hand).
  ///
  /// In en, this message translates to:
  /// **'{kind, select, new{New} learning{Learning} review{Review} relearning{Relearning} manual{Manual} other{{kind}}}'**
  String cardKind(String kind);

  /// An answer time under a minute, e.g. "4.2 s".
  ///
  /// In en, this message translates to:
  /// **'{seconds} s'**
  String secondsDecimal(String seconds);

  /// No description provided for @cardInfo.
  ///
  /// In en, this message translates to:
  /// **'Card info'**
  String get cardInfo;

  /// No description provided for @reviewHistory.
  ///
  /// In en, this message translates to:
  /// **'Review history'**
  String get reviewHistory;

  /// No description provided for @noReviewsYet.
  ///
  /// In en, this message translates to:
  /// **'No reviews yet.'**
  String get noReviewsYet;

  /// Card info: when the card was added.
  ///
  /// In en, this message translates to:
  /// **'Added'**
  String get infoAdded;

  /// No description provided for @infoFirstReview.
  ///
  /// In en, this message translates to:
  /// **'First review'**
  String get infoFirstReview;

  /// No description provided for @infoLatestReview.
  ///
  /// In en, this message translates to:
  /// **'Latest review'**
  String get infoLatestReview;

  /// Card info: when the card is due.
  ///
  /// In en, this message translates to:
  /// **'Due'**
  String get infoDue;

  /// Card info.
  ///
  /// In en, this message translates to:
  /// **'Interval'**
  String get infoInterval;

  /// Card info.
  ///
  /// In en, this message translates to:
  /// **'Stability'**
  String get infoStability;

  /// Card info.
  ///
  /// In en, this message translates to:
  /// **'Difficulty'**
  String get infoDifficulty;

  /// Card info.
  ///
  /// In en, this message translates to:
  /// **'Retrievability'**
  String get infoRetrievability;

  /// Card info: how many times the card was reviewed.
  ///
  /// In en, this message translates to:
  /// **'Reviews'**
  String get infoReviews;

  /// Card info: how many times the card was forgotten.
  ///
  /// In en, this message translates to:
  /// **'Lapses'**
  String get infoLapses;

  /// No description provided for @averageTime.
  ///
  /// In en, this message translates to:
  /// **'Average time'**
  String get averageTime;

  /// No description provided for @totalTime.
  ///
  /// In en, this message translates to:
  /// **'Total time'**
  String get totalTime;

  /// Card info: new, learning, review or relearning.
  ///
  /// In en, this message translates to:
  /// **'State'**
  String get infoState;

  /// No description provided for @cardType.
  ///
  /// In en, this message translates to:
  /// **'Card type'**
  String get cardType;

  /// No description provided for @cardId.
  ///
  /// In en, this message translates to:
  /// **'Card ID'**
  String get cardId;

  /// No description provided for @noteId.
  ///
  /// In en, this message translates to:
  /// **'Note ID'**
  String get noteId;

  /// Column of the review history.
  ///
  /// In en, this message translates to:
  /// **'Date'**
  String get historyDate;

  /// Column of the review history: learning, review, relearning or manual.
  ///
  /// In en, this message translates to:
  /// **'Kind'**
  String get historyKind;

  /// Column of the review history.
  ///
  /// In en, this message translates to:
  /// **'Rating'**
  String get historyRating;

  /// Column of the review history.
  ///
  /// In en, this message translates to:
  /// **'Interval'**
  String get historyInterval;

  /// Column of the review history: the answer time.
  ///
  /// In en, this message translates to:
  /// **'Time'**
  String get historyTime;

  /// No description provided for @deletedCards.
  ///
  /// In en, this message translates to:
  /// **'Deleted {count, plural, =1{1 card} other{{count} cards}}'**
  String deletedCards(int count);

  /// No description provided for @setDueDateOfCards.
  ///
  /// In en, this message translates to:
  /// **'Set due date of {count, plural, =1{1 card} other{{count} cards}}'**
  String setDueDateOfCards(int count);

  /// No description provided for @resetCardsToNew.
  ///
  /// In en, this message translates to:
  /// **'Reset {count, plural, =1{1 card} other{{count} cards}} to new'**
  String resetCardsToNew(int count);

  /// Tooltip of the button that edits a card.
  ///
  /// In en, this message translates to:
  /// **'Edit'**
  String get editAction;

  /// No description provided for @resetToNew.
  ///
  /// In en, this message translates to:
  /// **'Reset to new'**
  String get resetToNew;

  /// No description provided for @unsuspend.
  ///
  /// In en, this message translates to:
  /// **'Unsuspend'**
  String get unsuspend;

  /// No description provided for @suspend.
  ///
  /// In en, this message translates to:
  /// **'Suspend'**
  String get suspend;

  /// A suspended card's state, e.g. "Review (suspended)".
  ///
  /// In en, this message translates to:
  /// **'{state} (suspended)'**
  String stateWithSuspended(String state);

  /// Due of a new card without a position.
  ///
  /// In en, this message translates to:
  /// **'new card'**
  String get newCard;

  /// No description provided for @countDays.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 day} other{{count} days}}'**
  String countDays(int count);

  /// A stability in days with one decimal, e.g. "12.3 days".
  ///
  /// In en, this message translates to:
  /// **'{days} days'**
  String daysDecimal(String days);

  /// Heading of a card's details; also the card's note type and card type.
  ///
  /// In en, this message translates to:
  /// **'Card'**
  String get card;

  /// Heading of a card's question.
  ///
  /// In en, this message translates to:
  /// **'Question'**
  String get question;

  /// Heading of a card's answer.
  ///
  /// In en, this message translates to:
  /// **'Answer'**
  String get answer;

  /// No description provided for @changesApplyToOthers.
  ///
  /// In en, this message translates to:
  /// **'Changes also apply to {count, plural, =1{1 other card} other{{count} other cards}} from this note.'**
  String changesApplyToOthers(int count);

  /// No description provided for @cardSaved.
  ///
  /// In en, this message translates to:
  /// **'Card saved'**
  String get cardSaved;

  /// No description provided for @changeCardsDeleted.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}} will be deleted.'**
  String changeCardsDeleted(int count);

  /// No description provided for @changeCardsCreated.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 card} other{{count} cards}} will be created.'**
  String changeCardsCreated(int count);

  /// No description provided for @changeCardsBoth.
  ///
  /// In en, this message translates to:
  /// **'{deleted, plural, =1{1 card} other{{deleted} cards}} will be deleted and {created, plural, =1{1 card} other{{created} cards}} will be created.'**
  String changeCardsBoth(int deleted, int created);

  /// No description provided for @changeNoCards.
  ///
  /// In en, this message translates to:
  /// **'No cards will be deleted or created.'**
  String get changeNoCards;

  /// No description provided for @changeLostFields.
  ///
  /// In en, this message translates to:
  /// **'The content of {fields} will be lost.'**
  String changeLostFields(String fields);

  /// The end of a list, e.g. "Front, Back and Extra"; {items} is the rest joined with commas.
  ///
  /// In en, this message translates to:
  /// **'{items} and {last}'**
  String listAnd(String items, String last);

  /// No description provided for @changeNotesTitle.
  ///
  /// In en, this message translates to:
  /// **'Change {count, plural, =1{1 note} other{{count} notes}} to {noteType}?'**
  String changeNotesTitle(int count, String noteType);

  /// Button that changes the note type.
  ///
  /// In en, this message translates to:
  /// **'Change'**
  String get change;

  /// No description provided for @changeNoteType.
  ///
  /// In en, this message translates to:
  /// **'Change note type'**
  String get changeNoteType;

  /// No description provided for @currentNoteType.
  ///
  /// In en, this message translates to:
  /// **'Current note type: {noteType}'**
  String currentNoteType(String noteType);

  /// No description provided for @newNoteType.
  ///
  /// In en, this message translates to:
  /// **'New note type'**
  String get newNoteType;

  /// No description provided for @changeFieldsHelp.
  ///
  /// In en, this message translates to:
  /// **'Each new field takes the content of the field you choose.'**
  String get changeFieldsHelp;

  /// No description provided for @changeCardsHelp.
  ///
  /// In en, this message translates to:
  /// **'Each new card type takes the cards of the card type you choose, with their scheduling. Cards of card types not chosen are deleted.'**
  String get changeCardsHelp;

  /// No description provided for @changeClozeHelp.
  ///
  /// In en, this message translates to:
  /// **'Cards keep their scheduling when the changed note has a card with the same number. Other cards are deleted, and missing ones are created.'**
  String get changeClozeHelp;

  /// A field or card type that takes nothing.
  ///
  /// In en, this message translates to:
  /// **'(nothing)'**
  String get nothing;

  /// No description provided for @saveSearch.
  ///
  /// In en, this message translates to:
  /// **'Save search'**
  String get saveSearch;

  /// No description provided for @savedSearchAs.
  ///
  /// In en, this message translates to:
  /// **'Saved the search as \"{name}\".'**
  String savedSearchAs(String name);

  /// No description provided for @savedSearches.
  ///
  /// In en, this message translates to:
  /// **'Saved searches'**
  String get savedSearches;

  /// No description provided for @saveCurrentSearch.
  ///
  /// In en, this message translates to:
  /// **'Save current search'**
  String get saveCurrentSearch;

  /// No description provided for @manageSavedSearches.
  ///
  /// In en, this message translates to:
  /// **'Rename or delete saved searches'**
  String get manageSavedSearches;

  /// No description provided for @renameSavedSearch.
  ///
  /// In en, this message translates to:
  /// **'Rename saved search'**
  String get renameSavedSearch;

  /// Title of a confirmation to delete a saved search.
  ///
  /// In en, this message translates to:
  /// **'Delete \"{name}\"?'**
  String deleteQuoted(String name);

  /// No description provided for @deleteSavedSearchBody.
  ///
  /// In en, this message translates to:
  /// **'The saved search {query} is removed. Cards are not changed.'**
  String deleteSavedSearchBody(String query);

  /// No description provided for @noSavedSearches.
  ///
  /// In en, this message translates to:
  /// **'No saved searches.'**
  String get noSavedSearches;

  /// No description provided for @findTextPrompt.
  ///
  /// In en, this message translates to:
  /// **'Enter the text to find.'**
  String get findTextPrompt;

  /// No description provided for @counting.
  ///
  /// In en, this message translates to:
  /// **'Counting…'**
  String get counting;

  /// No description provided for @noNotesWillChange.
  ///
  /// In en, this message translates to:
  /// **'No notes will change.'**
  String get noNotesWillChange;

  /// No description provided for @notesWillChange.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 note} other{{count} notes}} will change.'**
  String notesWillChange(int count);

  /// No description provided for @findAndReplace.
  ///
  /// In en, this message translates to:
  /// **'Find and replace'**
  String get findAndReplace;

  /// No description provided for @replaceWith.
  ///
  /// In en, this message translates to:
  /// **'Replace with'**
  String get replaceWith;

  /// No description provided for @regexGroupsHelp.
  ///
  /// In en, this message translates to:
  /// **'Use \'\$1\' or \'\${name}\' for groups'**
  String get regexGroupsHelp;

  /// Label of where find and replace looks: a field, all fields or tags.
  ///
  /// In en, this message translates to:
  /// **'In'**
  String get inField;

  /// No description provided for @allFields.
  ///
  /// In en, this message translates to:
  /// **'All fields'**
  String get allFields;

  /// No description provided for @selectedNotesOnly.
  ///
  /// In en, this message translates to:
  /// **'Selected notes only'**
  String get selectedNotesOnly;

  /// No description provided for @treatAsRegex.
  ///
  /// In en, this message translates to:
  /// **'Treat input as regular expression'**
  String get treatAsRegex;

  /// No description provided for @ignoreCase.
  ///
  /// In en, this message translates to:
  /// **'Ignore case'**
  String get ignoreCase;

  /// No description provided for @movedCards.
  ///
  /// In en, this message translates to:
  /// **'Moved {count, plural, =1{1 card} other{{count} cards}}'**
  String movedCards(int count);

  /// No description provided for @tagsSeparated.
  ///
  /// In en, this message translates to:
  /// **'Tags, separated by spaces'**
  String get tagsSeparated;

  /// No description provided for @removeTags.
  ///
  /// In en, this message translates to:
  /// **'Remove tags'**
  String get removeTags;

  /// Added to the reposition result; starts with a space.
  ///
  /// In en, this message translates to:
  /// **' {count, plural, =1{1 card was} other{{count} cards were}} not new and left unchanged.'**
  String cardsNotNew(int count);

  /// No description provided for @noNewCardsSelected.
  ///
  /// In en, this message translates to:
  /// **'No new cards selected.'**
  String get noNewCardsSelected;

  /// No description provided for @repositionedCards.
  ///
  /// In en, this message translates to:
  /// **'Repositioned {count, plural, =1{1 card} other{{count} cards}}.'**
  String repositionedCards(int count);

  /// No description provided for @changedNoteTypeOf.
  ///
  /// In en, this message translates to:
  /// **'Changed note type of {count, plural, =1{1 note} other{{count} notes}}'**
  String changedNoteTypeOf(int count);

  /// No description provided for @noNotesChanged.
  ///
  /// In en, this message translates to:
  /// **'No notes changed.'**
  String get noNotesChanged;

  /// No description provided for @changedNotes.
  ///
  /// In en, this message translates to:
  /// **'Changed {count, plural, =1{1 note} other{{count} notes}}'**
  String changedNotes(int count);

  /// No description provided for @moveToDeck.
  ///
  /// In en, this message translates to:
  /// **'Move to deck'**
  String get moveToDeck;

  /// No description provided for @clearSelection.
  ///
  /// In en, this message translates to:
  /// **'Clear selection'**
  String get clearSelection;

  /// How many rows are selected.
  ///
  /// In en, this message translates to:
  /// **'{count} selected'**
  String countSelected(int count);

  /// No description provided for @selectCards.
  ///
  /// In en, this message translates to:
  /// **'Select cards'**
  String get selectCards;

  /// No description provided for @selectNotes.
  ///
  /// In en, this message translates to:
  /// **'Select notes'**
  String get selectNotes;

  /// No description provided for @selectAll.
  ///
  /// In en, this message translates to:
  /// **'Select all'**
  String get selectAll;

  /// Menu of selection actions.
  ///
  /// In en, this message translates to:
  /// **'Selection'**
  String get selection;

  /// No description provided for @invertSelection.
  ///
  /// In en, this message translates to:
  /// **'Invert selection'**
  String get invertSelection;

  /// No description provided for @bulkActions.
  ///
  /// In en, this message translates to:
  /// **'Bulk actions'**
  String get bulkActions;

  /// No description provided for @toggleMark.
  ///
  /// In en, this message translates to:
  /// **'Toggle mark'**
  String get toggleMark;

  /// No description provided for @deleteCardsAction.
  ///
  /// In en, this message translates to:
  /// **'Delete cards'**
  String get deleteCardsAction;

  /// No description provided for @deleteNotesAction.
  ///
  /// In en, this message translates to:
  /// **'Delete notes'**
  String get deleteNotesAction;

  /// Title of the card browser.
  ///
  /// In en, this message translates to:
  /// **'Browse'**
  String get browse;

  /// No description provided for @sidebar.
  ///
  /// In en, this message translates to:
  /// **'Sidebar'**
  String get sidebar;

  /// No description provided for @sort.
  ///
  /// In en, this message translates to:
  /// **'Sort'**
  String get sort;

  /// No description provided for @noCardsYet.
  ///
  /// In en, this message translates to:
  /// **'No cards yet.'**
  String get noCardsYet;

  /// No description provided for @noCardsMatch.
  ///
  /// In en, this message translates to:
  /// **'No cards match your search.'**
  String get noCardsMatch;

  /// No description provided for @noNotesMatch.
  ///
  /// In en, this message translates to:
  /// **'No notes match your search.'**
  String get noNotesMatch;

  /// No description provided for @searchHint.
  ///
  /// In en, this message translates to:
  /// **'Search, e.g. deck:Spanish tag:verb -is:suspended'**
  String get searchHint;

  /// Filter that shows only some notes.
  ///
  /// In en, this message translates to:
  /// **'Only {count, plural, =1{1 note} other{{count} notes}}'**
  String onlyNotes(int count);

  /// No description provided for @showAllCards.
  ///
  /// In en, this message translates to:
  /// **'Show all cards'**
  String get showAllCards;

  /// No description provided for @showAllNotes.
  ///
  /// In en, this message translates to:
  /// **'Show all notes'**
  String get showAllNotes;

  /// Browser mode: one row per card.
  ///
  /// In en, this message translates to:
  /// **'Cards'**
  String get modeCards;

  /// Browser mode: one row per note.
  ///
  /// In en, this message translates to:
  /// **'Notes'**
  String get modeNotes;

  /// No description provided for @oneRowPerCard.
  ///
  /// In en, this message translates to:
  /// **'One row per card'**
  String get oneRowPerCard;

  /// No description provided for @oneRowPerNote.
  ///
  /// In en, this message translates to:
  /// **'One row per note'**
  String get oneRowPerNote;

  /// No description provided for @selectCardForDetails.
  ///
  /// In en, this message translates to:
  /// **'Select a card to see its details.'**
  String get selectCardForDetails;

  /// No description provided for @selectNoteForDetails.
  ///
  /// In en, this message translates to:
  /// **'Select a note to see its details.'**
  String get selectNoteForDetails;

  /// Sidebar: cards in learning.
  ///
  /// In en, this message translates to:
  /// **'Learning'**
  String get stateLearning;

  /// Sidebar: review cards.
  ///
  /// In en, this message translates to:
  /// **'Review'**
  String get stateReview;

  /// Sidebar: due cards.
  ///
  /// In en, this message translates to:
  /// **'Due'**
  String get stateDue;

  /// Sidebar: suspended cards.
  ///
  /// In en, this message translates to:
  /// **'Suspended'**
  String get stateSuspendedTitle;

  /// No description provided for @renameTag.
  ///
  /// In en, this message translates to:
  /// **'Rename tag'**
  String get renameTag;

  /// No description provided for @renameTagHelper.
  ///
  /// In en, this message translates to:
  /// **'Use :: to move it under another tag'**
  String get renameTagHelper;

  /// No description provided for @renamedTagOn.
  ///
  /// In en, this message translates to:
  /// **'Renamed the tag on {count, plural, =1{1 note} other{{count} notes}}'**
  String renamedTagOn(int count);

  /// No description provided for @deleteTagTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete the tag “{tag}”?'**
  String deleteTagTitle(String tag);

  /// No description provided for @deleteTagsTitle.
  ///
  /// In en, this message translates to:
  /// **'Delete {count, plural, =1{1 tag} other{{count} tags}}?'**
  String deleteTagsTitle(int count);

  /// No description provided for @deleteTagBody.
  ///
  /// In en, this message translates to:
  /// **'This removes it and its child tags from {notes, plural, =1{1 note} other{{notes} notes}}. The notes are kept.'**
  String deleteTagBody(int notes);

  /// No description provided for @deleteTagsBody.
  ///
  /// In en, this message translates to:
  /// **'This removes them and their child tags from {notes, plural, =1{1 note} other{{notes} notes}}. The notes are kept.'**
  String deleteTagsBody(int notes);

  /// No description provided for @removedTagsFrom.
  ///
  /// In en, this message translates to:
  /// **'Removed {count, plural, =1{1 tag} other{{count} tags}} from {notes, plural, =1{1 note} other{{notes} notes}}'**
  String removedTagsFrom(int count, int notes);

  /// No description provided for @movedTagsTop.
  ///
  /// In en, this message translates to:
  /// **'Moved {count, plural, =1{1 tag} other{{count} tags}} to the top level'**
  String movedTagsTop(int count);

  /// No description provided for @movedTagsUnder.
  ///
  /// In en, this message translates to:
  /// **'Moved {count, plural, =1{1 tag} other{{count} tags}} under “{parent}”'**
  String movedTagsUnder(int count, String parent);

  /// No description provided for @moveTagsTo.
  ///
  /// In en, this message translates to:
  /// **'Move {count, plural, =1{1 tag} other{{count} tags}} to'**
  String moveTagsTo(int count);

  /// No description provided for @topLevel.
  ///
  /// In en, this message translates to:
  /// **'Top level'**
  String get topLevel;

  /// No description provided for @noUnusedTags.
  ///
  /// In en, this message translates to:
  /// **'There are no unused tags.'**
  String get noUnusedTags;

  /// No description provided for @clearedUnusedTags.
  ///
  /// In en, this message translates to:
  /// **'Cleared {count, plural, =1{1 unused tag} other{{count} unused tags}}.'**
  String clearedUnusedTags(int count);

  /// No description provided for @movedDeckTop.
  ///
  /// In en, this message translates to:
  /// **'Moved “{deck}” to the top level'**
  String movedDeckTop(String deck);

  /// No description provided for @movedDeckInto.
  ///
  /// In en, this message translates to:
  /// **'Moved “{deck}” into “{parent}”'**
  String movedDeckInto(String deck, String parent);

  /// No description provided for @moveDeckTo.
  ///
  /// In en, this message translates to:
  /// **'Move “{deck}” to'**
  String moveDeckTo(String deck);

  /// No description provided for @addToSearch.
  ///
  /// In en, this message translates to:
  /// **'Add to search'**
  String get addToSearch;

  /// No description provided for @addToSearchOr.
  ///
  /// In en, this message translates to:
  /// **'Add to search with OR'**
  String get addToSearchOr;

  /// No description provided for @excludeFromSearchAction.
  ///
  /// In en, this message translates to:
  /// **'Exclude from search'**
  String get excludeFromSearchAction;

  /// No description provided for @renameEllipsis.
  ///
  /// In en, this message translates to:
  /// **'Rename…'**
  String get renameEllipsis;

  /// No description provided for @moveToEllipsis.
  ///
  /// In en, this message translates to:
  /// **'Move to…'**
  String get moveToEllipsis;

  /// No description provided for @selectTags.
  ///
  /// In en, this message translates to:
  /// **'Select tags'**
  String get selectTags;

  /// No description provided for @deleteEllipsis.
  ///
  /// In en, this message translates to:
  /// **'Delete…'**
  String get deleteEllipsis;

  /// Hint of the sidebar's filter field.
  ///
  /// In en, this message translates to:
  /// **'Filter'**
  String get filterHint;

  /// No description provided for @clearFilter.
  ///
  /// In en, this message translates to:
  /// **'Clear filter'**
  String get clearFilter;

  /// No description provided for @cardStates.
  ///
  /// In en, this message translates to:
  /// **'Card states'**
  String get cardStates;

  /// No description provided for @tagActions.
  ///
  /// In en, this message translates to:
  /// **'Tag actions'**
  String get tagActions;

  /// No description provided for @clearUnusedTags.
  ///
  /// In en, this message translates to:
  /// **'Clear unused tags'**
  String get clearUnusedTags;

  /// No description provided for @noTagsYet.
  ///
  /// In en, this message translates to:
  /// **'No tags yet.'**
  String get noTagsYet;

  /// Shown while dragging several tags in the browser sidebar.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 tag} other{{count} tags}}'**
  String tagCount(int count);

  /// No description provided for @tagsSelected.
  ///
  /// In en, this message translates to:
  /// **'{count, plural, =1{1 tag} other{{count} tags}} selected'**
  String tagsSelected(int count);

  /// No description provided for @moveTagsToTooltip.
  ///
  /// In en, this message translates to:
  /// **'Move tags to'**
  String get moveTagsToTooltip;

  /// No description provided for @deleteTags.
  ///
  /// In en, this message translates to:
  /// **'Delete tags'**
  String get deleteTags;

  /// Title of the export dialog for a deck.
  ///
  /// In en, this message translates to:
  /// **'Export “{name}”'**
  String exportNamed(String name);

  /// No description provided for @exportCards.
  ///
  /// In en, this message translates to:
  /// **'Export {count, plural, =1{1 card} other{{count} cards}}'**
  String exportCards(int count);

  /// No description provided for @exportNotesOf.
  ///
  /// In en, this message translates to:
  /// **'Export the notes of {count, plural, =1{1 card} other{{count} cards}}'**
  String exportNotesOf(int count);

  /// No description provided for @exportStopped.
  ///
  /// In en, this message translates to:
  /// **'The export stopped unexpectedly.'**
  String get exportStopped;

  /// No description provided for @exportedFile.
  ///
  /// In en, this message translates to:
  /// **'Exported {file}'**
  String exportedFile(String file);

  /// No description provided for @preparing.
  ///
  /// In en, this message translates to:
  /// **'Preparing…'**
  String get preparing;

  /// No description provided for @renderingCards.
  ///
  /// In en, this message translates to:
  /// **'Rendering the cards ({done} of {total})…'**
  String renderingCards(int done, int total);

  /// No description provided for @writingFile.
  ///
  /// In en, this message translates to:
  /// **'Writing the file…'**
  String get writingFile;

  /// No description provided for @addingMedia.
  ///
  /// In en, this message translates to:
  /// **'Adding media ({done} of {total})…'**
  String addingMedia(int done, int total);

  /// No description provided for @packingFile.
  ///
  /// In en, this message translates to:
  /// **'Packing the file…'**
  String get packingFile;

  /// No description provided for @writingCards.
  ///
  /// In en, this message translates to:
  /// **'Writing the cards…'**
  String get writingCards;

  /// No description provided for @exportColpkgHelp.
  ///
  /// In en, this message translates to:
  /// **'An Anki collection package (.colpkg) with every deck, note and card, their scheduling, review history and deck presets. Opening it in Anki replaces the collection there.'**
  String get exportColpkgHelp;

  /// No description provided for @exportApkgHelp.
  ///
  /// In en, this message translates to:
  /// **'An Anki deck package (.apkg), which Anki, AnkiDroid and Cardflip can import.'**
  String get exportApkgHelp;

  /// No description provided for @exportNotesTextHelp.
  ///
  /// In en, this message translates to:
  /// **'One note per line with its fields separated by tabs, HTML kept. Cardflip and Anki can import the file again to update the notes.'**
  String get exportNotesTextHelp;

  /// No description provided for @exportCardsTextHelp.
  ///
  /// In en, this message translates to:
  /// **'One card per line with its question and answer as plain text, for printing or reading elsewhere.'**
  String get exportCardsTextHelp;

  /// No description provided for @format.
  ///
  /// In en, this message translates to:
  /// **'Format'**
  String get format;

  /// No description provided for @formatColpkg.
  ///
  /// In en, this message translates to:
  /// **'Anki collection package (.colpkg)'**
  String get formatColpkg;

  /// No description provided for @formatApkg.
  ///
  /// In en, this message translates to:
  /// **'Anki deck package (.apkg)'**
  String get formatApkg;

  /// No description provided for @formatNotesText.
  ///
  /// In en, this message translates to:
  /// **'Notes in plain text (.txt)'**
  String get formatNotesText;

  /// No description provided for @formatCardsText.
  ///
  /// In en, this message translates to:
  /// **'Cards in plain text (.txt)'**
  String get formatCardsText;

  /// No description provided for @includeTags.
  ///
  /// In en, this message translates to:
  /// **'Include tags'**
  String get includeTags;

  /// No description provided for @includeDeck.
  ///
  /// In en, this message translates to:
  /// **'Include deck name'**
  String get includeDeck;

  /// No description provided for @includeNoteType.
  ///
  /// In en, this message translates to:
  /// **'Include note type name'**
  String get includeNoteType;

  /// No description provided for @includeGuid.
  ///
  /// In en, this message translates to:
  /// **'Include unique identifier'**
  String get includeGuid;

  /// No description provided for @includeGuidHelp.
  ///
  /// In en, this message translates to:
  /// **'Lets Anki update these notes when it imports the file again.'**
  String get includeGuidHelp;

  /// No description provided for @includeScheduling.
  ///
  /// In en, this message translates to:
  /// **'Include scheduling information'**
  String get includeScheduling;

  /// No description provided for @includeSchedulingHelp.
  ///
  /// In en, this message translates to:
  /// **'Due dates, intervals and review history. Without it, every card is new.'**
  String get includeSchedulingHelp;

  /// No description provided for @includePresets.
  ///
  /// In en, this message translates to:
  /// **'Include deck presets'**
  String get includePresets;

  /// No description provided for @includePresetsHelp.
  ///
  /// In en, this message translates to:
  /// **'The deck options'**
  String get includePresetsHelp;

  /// No description provided for @includeMedia.
  ///
  /// In en, this message translates to:
  /// **'Include media'**
  String get includeMedia;

  /// No description provided for @includeMediaHelp.
  ///
  /// In en, this message translates to:
  /// **'Images, audio and video'**
  String get includeMediaHelp;

  /// No description provided for @legacyAnki.
  ///
  /// In en, this message translates to:
  /// **'Support older Anki versions'**
  String get legacyAnki;

  /// No description provided for @legacyAnkiHelp.
  ///
  /// In en, this message translates to:
  /// **'For Anki 2.1.49 and earlier. Slower to import.'**
  String get legacyAnkiHelp;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Today'**
  String get secToday;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Streaks'**
  String get secStreaks;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Future due'**
  String get secFutureDue;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Calendar'**
  String get secCalendar;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Reviews'**
  String get secReviews;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Review time'**
  String get secReviewTime;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Card counts'**
  String get secCounts;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Review intervals'**
  String get secIntervals;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Stability'**
  String get secStability;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Difficulty'**
  String get secDifficulty;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Retrievability'**
  String get secRetrievability;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Hourly breakdown'**
  String get secHourly;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'Answer buttons'**
  String get secButtons;

  /// Statistics section.
  ///
  /// In en, this message translates to:
  /// **'True retention'**
  String get secTrueRetention;

  /// No description provided for @period1Month.
  ///
  /// In en, this message translates to:
  /// **'1 month'**
  String get period1Month;

  /// No description provided for @period3Months.
  ///
  /// In en, this message translates to:
  /// **'3 months'**
  String get period3Months;

  /// No description provided for @period1Year.
  ///
  /// In en, this message translates to:
  /// **'1 year'**
  String get period1Year;

  /// No description provided for @periodAll.
  ///
  /// In en, this message translates to:
  /// **'All history'**
  String get periodAll;

  /// Short label of the period All history.
  ///
  /// In en, this message translates to:
  /// **'All'**
  String get periodAllShort;

  /// No description provided for @noCardsStudiedToday.
  ///
  /// In en, this message translates to:
  /// **'No cards have been studied today.'**
  String get noCardsStudiedToday;

  /// No description provided for @studiedToday.
  ///
  /// In en, this message translates to:
  /// **'Studied {count, plural, =1{1 card} other{{count} cards}} in {time} today.'**
  String studiedToday(int count, String time);

  /// No description provided for @againCount.
  ///
  /// In en, this message translates to:
  /// **'Again count: {count} ({percent} correct)'**
  String againCount(String count, String percent);

  /// No description provided for @todayKinds.
  ///
  /// In en, this message translates to:
  /// **'Learning: {learning}, Review: {review}, Relearning: {relearning}, Filtered: {filtered}'**
  String todayKinds(
    String learning,
    String review,
    String relearning,
    String filtered,
  );

  /// A day: today.
  ///
  /// In en, this message translates to:
  /// **'Today'**
  String get today;

  /// A day: tomorrow.
  ///
  /// In en, this message translates to:
  /// **'Tomorrow'**
  String get tomorrow;

  /// No description provided for @overdue.
  ///
  /// In en, this message translates to:
  /// **'Overdue'**
  String get overdue;

  /// No description provided for @dueTomorrowLabel.
  ///
  /// In en, this message translates to:
  /// **'Due tomorrow'**
  String get dueTomorrowLabel;

  /// No description provided for @total.
  ///
  /// In en, this message translates to:
  /// **'Total'**
  String get total;

  /// No description provided for @average.
  ///
  /// In en, this message translates to:
  /// **'Average'**
  String get average;

  /// An average with one decimal, e.g. "12.5 cards/day".
  ///
  /// In en, this message translates to:
  /// **'{count} cards/day'**
  String cardsPerDay(String count);

  /// No description provided for @forecastTip.
  ///
  /// In en, this message translates to:
  /// **'{day}: {count, plural, =1{1 card} other{{count} cards}}\nRunning total: {total}'**
  String forecastTip(String day, int count, String total);

  /// No description provided for @runningTotal.
  ///
  /// In en, this message translates to:
  /// **'Running total'**
  String get runningTotal;

  /// No description provided for @daysStudiedYear.
  ///
  /// In en, this message translates to:
  /// **'Days studied in the last year'**
  String get daysStudiedYear;

  /// e.g. "120 of 365".
  ///
  /// In en, this message translates to:
  /// **'{count} of {total}'**
  String countOf(String count, String total);

  /// No description provided for @mostReviewsDay.
  ///
  /// In en, this message translates to:
  /// **'Most reviews on one day'**
  String get mostReviewsDay;

  /// No description provided for @currentStreak.
  ///
  /// In en, this message translates to:
  /// **'Current streak'**
  String get currentStreak;

  /// No description provided for @longestStreak.
  ///
  /// In en, this message translates to:
  /// **'Longest streak'**
  String get longestStreak;

  /// The longest streak with its dates, e.g. "41 days, from Mar 2 to Apr 11"; {days} is like "41 days".
  ///
  /// In en, this message translates to:
  /// **'{days}, from {first} to {last}'**
  String longestStreakValue(String days, String first, String last);

  /// No description provided for @studyDaysInPeriod.
  ///
  /// In en, this message translates to:
  /// **'Study days'**
  String get studyDaysInPeriod;

  /// No description provided for @goalDaysMet.
  ///
  /// In en, this message translates to:
  /// **'Days the goal was met'**
  String get goalDaysMet;

  /// Review time in minutes, e.g. "12 min".
  ///
  /// In en, this message translates to:
  /// **'{minutes} min'**
  String minutesAmount(String minutes);

  /// Review time in minutes on a chart axis, e.g. "30m".
  ///
  /// In en, this message translates to:
  /// **'{minutes}m'**
  String minutesAxis(String minutes);

  /// No description provided for @daysStudied.
  ///
  /// In en, this message translates to:
  /// **'Days studied'**
  String get daysStudied;

  /// No description provided for @averageDaysStudied.
  ///
  /// In en, this message translates to:
  /// **'Average for days studied'**
  String get averageDaysStudied;

  /// No description provided for @averageOverPeriod.
  ///
  /// In en, this message translates to:
  /// **'Average over period'**
  String get averageOverPeriod;

  /// Kind of review or card in statistics.
  ///
  /// In en, this message translates to:
  /// **'Learning'**
  String get kindLearning;

  /// Review cards with an interval under 21 days.
  ///
  /// In en, this message translates to:
  /// **'Young'**
  String get kindYoung;

  /// Review cards with an interval of 21 days or more.
  ///
  /// In en, this message translates to:
  /// **'Mature'**
  String get kindMature;

  /// Kind of review or card in statistics.
  ///
  /// In en, this message translates to:
  /// **'Relearning'**
  String get kindRelearning;

  /// Reviews in filtered decks.
  ///
  /// In en, this message translates to:
  /// **'Filtered'**
  String get kindFiltered;

  /// Card count in statistics.
  ///
  /// In en, this message translates to:
  /// **'Suspended'**
  String get kindSuspended;

  /// Card count in statistics.
  ///
  /// In en, this message translates to:
  /// **'Buried'**
  String get kindBuried;

  /// A histogram bucket of days, e.g. "7d" or "1-7d".
  ///
  /// In en, this message translates to:
  /// **'{days}d'**
  String bucketDays(String days);

  /// No description provided for @reviewCards.
  ///
  /// In en, this message translates to:
  /// **'Review cards'**
  String get reviewCards;

  /// No description provided for @averageInterval.
  ///
  /// In en, this message translates to:
  /// **'Average interval'**
  String get averageInterval;

  /// No description provided for @averageStability.
  ///
  /// In en, this message translates to:
  /// **'Average stability'**
  String get averageStability;

  /// No description provided for @averageDifficulty.
  ///
  /// In en, this message translates to:
  /// **'Average difficulty'**
  String get averageDifficulty;

  /// No description provided for @averageRetrievability.
  ///
  /// In en, this message translates to:
  /// **'Average retrievability'**
  String get averageRetrievability;

  /// No description provided for @estimatedKnowledge.
  ///
  /// In en, this message translates to:
  /// **'Estimated total knowledge'**
  String get estimatedKnowledge;

  /// No description provided for @noCardsInfo.
  ///
  /// In en, this message translates to:
  /// **'No cards with this information yet.'**
  String get noCardsInfo;

  /// No description provided for @noReviewsPeriod.
  ///
  /// In en, this message translates to:
  /// **'No reviews in this period.'**
  String get noReviewsPeriod;

  /// No description provided for @hourTip.
  ///
  /// In en, this message translates to:
  /// **'{hours}\n{count, plural, =1{1 review} other{{count} reviews}}'**
  String hourTip(String hours, int count);

  /// No description provided for @hourTipCorrect.
  ///
  /// In en, this message translates to:
  /// **', {percent} correct'**
  String hourTipCorrect(String percent);

  /// No description provided for @successRate.
  ///
  /// In en, this message translates to:
  /// **'Success rate'**
  String get successRate;

  /// No description provided for @noAnswers.
  ///
  /// In en, this message translates to:
  /// **'No answers'**
  String get noAnswers;

  /// Answer buttons: share of correct answers and how many there were.
  ///
  /// In en, this message translates to:
  /// **'{percent} correct ({count})'**
  String percentCorrect(String percent, String count);

  /// No description provided for @eachGroupShows.
  ///
  /// In en, this message translates to:
  /// **'Each group shows Again, Hard, Good and Easy.'**
  String get eachGroupShows;

  /// No description provided for @yesterday.
  ///
  /// In en, this message translates to:
  /// **'Yesterday'**
  String get yesterday;

  /// No description provided for @lastWeek.
  ///
  /// In en, this message translates to:
  /// **'Last week'**
  String get lastWeek;

  /// No description provided for @lastMonth.
  ///
  /// In en, this message translates to:
  /// **'Last month'**
  String get lastMonth;

  /// No description provided for @lastYear.
  ///
  /// In en, this message translates to:
  /// **'Last year'**
  String get lastYear;

  /// No description provided for @allTime.
  ///
  /// In en, this message translates to:
  /// **'All time'**
  String get allTime;

  /// No description provided for @desiredRetention.
  ///
  /// In en, this message translates to:
  /// **'Desired retention'**
  String get desiredRetention;

  /// No description provided for @calendarTip.
  ///
  /// In en, this message translates to:
  /// **'{date}: {count, plural, =1{1 review} other{{count} reviews}}'**
  String calendarTip(String date, int count);

  /// Statistics scope when the deck no longer exists.
  ///
  /// In en, this message translates to:
  /// **'Deck'**
  String get statsDeckFallback;

  /// No description provided for @statsSearchScope.
  ///
  /// In en, this message translates to:
  /// **'Search: {query}'**
  String statsSearchScope(String query);

  /// No description provided for @wholeCollection.
  ///
  /// In en, this message translates to:
  /// **'Whole collection'**
  String get wholeCollection;

  /// No description provided for @savedFile.
  ///
  /// In en, this message translates to:
  /// **'Saved {file}'**
  String savedFile(String file);

  /// No description provided for @saveAsPdf.
  ///
  /// In en, this message translates to:
  /// **'Save as PDF'**
  String get saveAsPdf;

  /// Statistics scope: the whole collection.
  ///
  /// In en, this message translates to:
  /// **'Collection'**
  String get collectionScope;

  /// Statistics scope: the cards of a search.
  ///
  /// In en, this message translates to:
  /// **'Search'**
  String get searchScope;

  /// No description provided for @deckWithSubdecks.
  ///
  /// In en, this message translates to:
  /// **'Deck (with its subdecks)'**
  String get deckWithSubdecks;

  /// No description provided for @statsSearchHint.
  ///
  /// In en, this message translates to:
  /// **'e.g. deck:Spanish tag:verbs'**
  String get statsSearchHint;

  /// No description provided for @showStatistics.
  ///
  /// In en, this message translates to:
  /// **'Show statistics'**
  String get showStatistics;

  /// Title of a statistics PDF.
  ///
  /// In en, this message translates to:
  /// **'Statistics - {scope}'**
  String statsTitleScope(String scope);

  /// File name of a statistics PDF.
  ///
  /// In en, this message translates to:
  /// **'Statistics - {scope} - {date}.pdf'**
  String statsPdfName(String scope, String date);

  /// No description provided for @statsTimeBased.
  ///
  /// In en, this message translates to:
  /// **'{scope} - time-based sections: {period}'**
  String statsTimeBased(String scope, String period);

  /// No description provided for @createdAt.
  ///
  /// In en, this message translates to:
  /// **'Created {date}'**
  String createdAt(String date);

  /// No description provided for @hour.
  ///
  /// In en, this message translates to:
  /// **'Hour'**
  String get hour;

  /// Column of answer statistics: the share of correct answers.
  ///
  /// In en, this message translates to:
  /// **'Correct'**
  String get correct;

  /// Settings section about profiles, each a separate collection.
  ///
  /// In en, this message translates to:
  /// **'Profile'**
  String get profile;

  /// Title of the profile list.
  ///
  /// In en, this message translates to:
  /// **'Profiles'**
  String get profiles;

  /// Settings entry and button that open the profile list.
  ///
  /// In en, this message translates to:
  /// **'Switch profile'**
  String get switchProfile;

  /// Setting: show the profile list when the app starts.
  ///
  /// In en, this message translates to:
  /// **'Show profile list at startup'**
  String get showProfilesAtStartup;

  /// Explains the setting showProfilesAtStartup.
  ///
  /// In en, this message translates to:
  /// **'Choose a profile each time Cardflip starts, instead of opening the last one used.'**
  String get showProfilesAtStartupHelp;

  /// Button and dialog title for adding a profile.
  ///
  /// In en, this message translates to:
  /// **'Add profile'**
  String get addProfile;

  /// Helper text in the add profile dialog.
  ///
  /// In en, this message translates to:
  /// **'A profile has its own cards, media, settings and backups.'**
  String get addProfileHelp;

  /// Menu item and dialog title for renaming a profile.
  ///
  /// In en, this message translates to:
  /// **'Rename profile'**
  String get renameProfile;

  /// Menu item for deleting a profile.
  ///
  /// In en, this message translates to:
  /// **'Delete profile'**
  String get deleteProfile;

  /// Confirmation before deleting a profile; the title names it.
  ///
  /// In en, this message translates to:
  /// **'This deletes the profile with {cards, plural, =0{no cards} =1{1 card} other{{cards} cards}} and all its media, settings and backups. This cannot be undone.'**
  String deleteProfileBody(int cards);

  /// Confirmation before deleting a profile whose cards could not be counted; the title names it.
  ///
  /// In en, this message translates to:
  /// **'This deletes the profile with all its cards, media, settings and backups. This cannot be undone.'**
  String get deleteProfileBodyUncounted;

  /// Marks the open profile in the profile list.
  ///
  /// In en, this message translates to:
  /// **'Open now'**
  String get profileIsOpen;

  /// Shown after deleting a profile.
  ///
  /// In en, this message translates to:
  /// **'Profile “{name}” deleted'**
  String profileDeleted(String name);

  /// Settings section for syncing with paired devices.
  ///
  /// In en, this message translates to:
  /// **'Sync'**
  String get sync;

  /// Settings tile with the name of this device for sync.
  ///
  /// In en, this message translates to:
  /// **'Device name'**
  String get syncDeviceName;

  /// Helper text when renaming this device.
  ///
  /// In en, this message translates to:
  /// **'1 to 40 characters. Your paired devices see this name.'**
  String get syncDeviceNameHelper;

  /// Dialog title for renaming this device.
  ///
  /// In en, this message translates to:
  /// **'Rename this device'**
  String get syncRenameDevice;

  /// Settings tile with the network the device is on.
  ///
  /// In en, this message translates to:
  /// **'Current network'**
  String get syncCurrentNetwork;

  /// A cable network, identified by its subnet such as 192.168.178.0/24.
  ///
  /// In en, this message translates to:
  /// **'Wired network {subnet}'**
  String syncWiredNetwork(String subnet);

  /// A network identified by its subnet such as 192.168.1.0/24, because its Wi-Fi name is unknown.
  ///
  /// In en, this message translates to:
  /// **'Network {subnet}'**
  String syncSubnetNetwork(String subnet);

  /// The current network is a mobile (cellular) network.
  ///
  /// In en, this message translates to:
  /// **'Mobile data'**
  String get syncMobileNetwork;

  /// The device is not connected to a network.
  ///
  /// In en, this message translates to:
  /// **'No network'**
  String get syncNoNetwork;

  /// Below the current network when it is trusted.
  ///
  /// In en, this message translates to:
  /// **'Trusted: Cardflip syncs here with your paired devices.'**
  String get syncNetworkTrusted;

  /// Below the current network when it is not trusted.
  ///
  /// In en, this message translates to:
  /// **'Not trusted: Cardflip does not sync or announce this device here.'**
  String get syncNetworkNotTrusted;

  /// Below the current network when it is a mobile network.
  ///
  /// In en, this message translates to:
  /// **'Cardflip never syncs over mobile data.'**
  String get syncMobileNeverTrusted;

  /// Button that marks the current network as trusted.
  ///
  /// In en, this message translates to:
  /// **'Trust this network'**
  String get trustThisNetwork;

  /// Button and tooltip that removes a network from the trusted networks.
  ///
  /// In en, this message translates to:
  /// **'Stop trusting'**
  String get stopTrusting;

  /// Title of the explanation shown before trusting a network.
  ///
  /// In en, this message translates to:
  /// **'Trust this network?'**
  String get trustNetworkTitle;

  /// Explanation shown before trusting a network.
  ///
  /// In en, this message translates to:
  /// **'On a trusted network, Cardflip makes this device visible by its name and syncs with your paired devices. Other devices on the network can see the name, but only devices you paired can exchange cards with it.\n\nTrust only private networks, such as your home Wi-Fi: another network with the same name is trusted too.'**
  String get trustNetworkBody;

  /// Added to the trust explanation where reading the Wi-Fi name needs the location permission (Android, iOS).
  ///
  /// In en, this message translates to:
  /// **'To recognise the network by its Wi-Fi name, Cardflip asks for the location permission next. Cardflip does not use your location. Without the permission, the network is recognised by its address range.'**
  String get trustNetworkLocation;

  /// Button that confirms trusting a network.
  ///
  /// In en, this message translates to:
  /// **'Trust'**
  String get trust;

  /// Settings tile above the list of trusted networks.
  ///
  /// In en, this message translates to:
  /// **'Trusted networks'**
  String get trustedNetworks;

  /// Shown when no network is trusted.
  ///
  /// In en, this message translates to:
  /// **'No trusted networks yet'**
  String get noTrustedNetworks;

  /// Settings tile above the list of paired devices.
  ///
  /// In en, this message translates to:
  /// **'Paired devices'**
  String get pairedDevices;

  /// Shown when no device is paired.
  ///
  /// In en, this message translates to:
  /// **'No paired devices yet'**
  String get noPairedDevices;

  /// A paired device is on the network now.
  ///
  /// In en, this message translates to:
  /// **'Reachable'**
  String get deviceReachable;

  /// A paired device is not on the network now.
  ///
  /// In en, this message translates to:
  /// **'Not reachable'**
  String get deviceNotReachable;

  /// When this profile last synced with a device; time is a date and time.
  ///
  /// In en, this message translates to:
  /// **'Last synced {time}'**
  String lastSyncedAt(String time);

  /// This profile has not synced with a device yet.
  ///
  /// In en, this message translates to:
  /// **'Not synced yet'**
  String get neverSynced;

  /// Button that unpairs a device.
  ///
  /// In en, this message translates to:
  /// **'Unpair'**
  String get unpair;

  /// Confirmation title before unpairing a device.
  ///
  /// In en, this message translates to:
  /// **'Unpair {device}?'**
  String unpairTitle(String device);

  /// Confirmation before unpairing a device.
  ///
  /// In en, this message translates to:
  /// **'{device} can no longer sync with this device, and the links of your profiles with it are removed. The cards on both devices stay as they are.'**
  String unpairBody(String device);

  /// Settings tile and screen title for pairing with another device.
  ///
  /// In en, this message translates to:
  /// **'Pair a device'**
  String get pairDevice;

  /// Below "Pair a device" when the network is trusted.
  ///
  /// In en, this message translates to:
  /// **'Connect another of your devices on this network'**
  String get pairDeviceHelp;

  /// Explanation at the top of the pairing screen.
  ///
  /// In en, this message translates to:
  /// **'Open “Pair a device” on the other device too, on the same network. Then choose it below, or enter the address it shows.'**
  String get pairDeviceIntro;

  /// How long this device stays visible for pairing; time is minutes and seconds such as 4:59.
  ///
  /// In en, this message translates to:
  /// **'Visible for pairing for {time}'**
  String visibleForPairing(String time);

  /// This device is no longer visible for pairing.
  ///
  /// In en, this message translates to:
  /// **'Not visible for pairing'**
  String get notVisibleForPairing;

  /// Button that makes this device visible for pairing again.
  ///
  /// In en, this message translates to:
  /// **'Make visible'**
  String get makeVisible;

  /// Tile with this device's address and port.
  ///
  /// In en, this message translates to:
  /// **'Address of this device'**
  String get thisDeviceAddress;

  /// Heading of the devices found for pairing.
  ///
  /// In en, this message translates to:
  /// **'Devices nearby'**
  String get devicesNearby;

  /// Shown while no device was found for pairing.
  ///
  /// In en, this message translates to:
  /// **'Looking for devices…'**
  String get lookingForDevices;

  /// Label of the field for pairing by address.
  ///
  /// In en, this message translates to:
  /// **'Address of the other device'**
  String get enterAddress;

  /// Helper text of the address field.
  ///
  /// In en, this message translates to:
  /// **'As shown on the other device, such as 192.168.1.20:47812'**
  String get enterAddressHelper;

  /// Button that sends a pairing request.
  ///
  /// In en, this message translates to:
  /// **'Pair'**
  String get pair;

  /// Title of an incoming pairing request.
  ///
  /// In en, this message translates to:
  /// **'Pair with {device}?'**
  String pairingRequestTitle(String device);

  /// An incoming pairing request.
  ///
  /// In en, this message translates to:
  /// **'{device} wants to pair with this device. Accept only if it is one of your own devices.'**
  String pairingRequestBody(String device);

  /// Button that accepts a request.
  ///
  /// In en, this message translates to:
  /// **'Accept'**
  String get accept;

  /// Button that declines a request.
  ///
  /// In en, this message translates to:
  /// **'Decline'**
  String get decline;

  /// Title of the pairing code dialog.
  ///
  /// In en, this message translates to:
  /// **'Compare the codes'**
  String get confirmCodeTitle;

  /// Above the pairing code.
  ///
  /// In en, this message translates to:
  /// **'Check that {device} shows the same code:'**
  String confirmCodeBody(String device);

  /// Button that confirms the pairing codes are the same.
  ///
  /// In en, this message translates to:
  /// **'Codes match'**
  String get codesMatch;

  /// Button that says the pairing codes are not the same.
  ///
  /// In en, this message translates to:
  /// **'Codes differ'**
  String get codesDiffer;

  /// Shown after confirming the code until the other device confirms.
  ///
  /// In en, this message translates to:
  /// **'Waiting for {device} to confirm…'**
  String waitingForConfirmation(String device);

  /// Shown when pairing succeeded.
  ///
  /// In en, this message translates to:
  /// **'Paired with {device}'**
  String pairedWith(String device);

  /// Shown when pairing failed.
  ///
  /// In en, this message translates to:
  /// **'Pairing with {device} failed'**
  String pairingFailed(String device);

  /// Settings tile above the open profile's links.
  ///
  /// In en, this message translates to:
  /// **'Links of this profile'**
  String get profileLinks;

  /// Shown when the open profile has no links.
  ///
  /// In en, this message translates to:
  /// **'This profile does not sync with another device yet.'**
  String get noProfileLinks;

  /// Settings tile that links the open profile with a profile on a paired device.
  ///
  /// In en, this message translates to:
  /// **'Link this profile'**
  String get linkProfile;

  /// Below "Link this profile".
  ///
  /// In en, this message translates to:
  /// **'Sync it with a profile on a paired device'**
  String get linkProfileHelp;

  /// A link: the profile on the other device, and the device.
  ///
  /// In en, this message translates to:
  /// **'{profile} on {device}'**
  String linkTarget(String profile, String device);

  /// Title of the dialog choosing a paired device to link with.
  ///
  /// In en, this message translates to:
  /// **'Choose a device'**
  String get chooseDevice;

  /// Title of the dialog choosing the profile to link with.
  ///
  /// In en, this message translates to:
  /// **'Profile on {device}'**
  String chooseRemoteProfile(String device);

  /// Option that creates a profile with the open profile's name on the other device.
  ///
  /// In en, this message translates to:
  /// **'New profile “{name}”'**
  String newProfileNamed(String name);

  /// A link whose first sync waits for the user to choose how to combine the profiles.
  ///
  /// In en, this message translates to:
  /// **'Needs your choice: merge or replace'**
  String get linkNeedsChoice;

  /// Button that removes a profile link.
  ///
  /// In en, this message translates to:
  /// **'Unlink'**
  String get unlink;

  /// Confirmation title before removing a link.
  ///
  /// In en, this message translates to:
  /// **'Unlink {device}?'**
  String unlinkTitle(String device);

  /// Confirmation before removing a link.
  ///
  /// In en, this message translates to:
  /// **'This profile stops syncing with “{profile}” on {device}. Both keep their cards.'**
  String unlinkBody(String profile, String device);

  /// Sync indicator while the profile syncs.
  ///
  /// In en, this message translates to:
  /// **'Syncing…'**
  String get syncSyncing;

  /// Sync indicator while the profile syncs, with the changes so far.
  ///
  /// In en, this message translates to:
  /// **'Syncing: {received} received, {sent} sent'**
  String syncProgress(int received, int sent);

  /// Sync indicator after the last sync; time is a time of day such as 14:05, or a date and time.
  ///
  /// In en, this message translates to:
  /// **'Synced {time}'**
  String syncedAt(String time);

  /// Sync indicator when the device is not on a trusted network.
  ///
  /// In en, this message translates to:
  /// **'Not on a trusted network'**
  String get syncNotTrusted;

  /// Sync indicator when a first sync waits for the user to choose merge or replace.
  ///
  /// In en, this message translates to:
  /// **'Sync needs your choice'**
  String get syncNeedsChoice;

  /// Sync indicator and dialog title after a failed sync.
  ///
  /// In en, this message translates to:
  /// **'Sync failed'**
  String get syncFailed;

  /// Button that retries a failed sync.
  ///
  /// In en, this message translates to:
  /// **'Sync again'**
  String get syncAgain;

  /// Tooltip of the sync indicator.
  ///
  /// In en, this message translates to:
  /// **'Sync now'**
  String get syncNow;

  /// Shown after Sync now: how many changes were received and sent.
  ///
  /// In en, this message translates to:
  /// **'Synced: {received} received, {sent} sent'**
  String syncSummary(int received, int sent);

  /// Shown after Sync now when media files were copied too.
  ///
  /// In en, this message translates to:
  /// **'Synced: {received} received, {sent} sent, {media} media files'**
  String syncSummaryMedia(int received, int sent, int media);

  /// Shown after Sync now when no linked device was reachable.
  ///
  /// In en, this message translates to:
  /// **'No linked device is reachable on this network.'**
  String get syncNothingReachable;

  /// Title of the first-sync choice.
  ///
  /// In en, this message translates to:
  /// **'Sync with {device}'**
  String firstSyncTitle(String device);

  /// The first-sync choice.
  ///
  /// In en, this message translates to:
  /// **'This profile and “{profile}” on {device} have not synced yet, or one of them was replaced since. How should they be combined? A backup is made before anything changes.'**
  String firstSyncBody(String profile, String device);

  /// First-sync choice that merges both profiles.
  ///
  /// In en, this message translates to:
  /// **'Merge'**
  String get firstSyncMerge;

  /// Explains the Merge choice.
  ///
  /// In en, this message translates to:
  /// **'Keep the cards and review history of both.'**
  String get firstSyncMergeHelp;

  /// First-sync choice that replaces the other device's profile with this one.
  ///
  /// In en, this message translates to:
  /// **'Replace the profile on {device}'**
  String firstSyncReplaceOther(String device);

  /// Explains the Replace other choice.
  ///
  /// In en, this message translates to:
  /// **'{device} gets exactly the cards and history of this profile.'**
  String firstSyncReplaceOtherHelp(String device);

  /// First-sync choice that replaces this profile with the other device's.
  ///
  /// In en, this message translates to:
  /// **'Replace this profile'**
  String get firstSyncReplaceThis;

  /// Explains the Replace this choice.
  ///
  /// In en, this message translates to:
  /// **'This device gets exactly the cards and history of “{profile}” on {device}.'**
  String firstSyncReplaceThisHelp(String profile, String device);

  /// Button that closes the first-sync choice without choosing.
  ///
  /// In en, this message translates to:
  /// **'Later'**
  String get decideLater;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['de', 'en'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'de':
      return AppLocalizationsDe();
    case 'en':
      return AppLocalizationsEn();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
