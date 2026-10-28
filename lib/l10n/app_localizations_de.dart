// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for German (`de`).
class AppLocalizationsDe extends AppLocalizations {
  AppLocalizationsDe([String locale = 'de']) : super(locale);

  @override
  String get appTitle => 'Cardflip';

  @override
  String get errorUnknown => 'Etwas ist schiefgelaufen.';

  @override
  String errorUnexpected(String error) {
    return 'Etwas ist schiefgelaufen: $error';
  }

  @override
  String get undo => 'Rückgängig';

  @override
  String goApkgCollectionDamaged(String error) {
    return 'die Sammlung ist beschädigt ($error)';
  }

  @override
  String goApkgDatabaseDamaged(String error) {
    return 'die Datenbank der Sammlung ist beschädigt ($error)';
  }

  @override
  String goApkgDecksUnreadable(String error) {
    return 'ihre Stapel konnten nicht gelesen werden ($error)';
  }

  @override
  String goApkgEmptyNoteType(String name) {
    return 'Notiztyp „$name“ hat keine Felder oder Vorlagen; nicht unterstützte Version des Anki-Pakets';
  }

  @override
  String goApkgFileDamaged(String file, String error) {
    return '$file ist beschädigt ($error)';
  }

  @override
  String goApkgFileError(String file, String error) {
    return '$file: $error';
  }

  @override
  String goApkgFileMissing(String file) {
    return '$file fehlt';
  }

  @override
  String goApkgMediaDamaged(String file, String error) {
    return 'Mediendatei $file ist beschädigt ($error)';
  }

  @override
  String goApkgMediaListDamaged(String error) {
    return 'die Medienliste ist beschädigt ($error)';
  }

  @override
  String goApkgMediaSkipped(int count, String files) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Audio-/Videodateien wurden',
      one: '1 Audio-/Videodatei wurde',
    );
    return '$_temp0 übersprungen, weil Cardflip ihr Format nicht abspielen kann: $files';
  }

  @override
  String get goApkgMissingNoteType =>
      'eine Notiz verwendet einen fehlenden Notiztyp';

  @override
  String goApkgNewerVersion(String version) {
    return 'es wurde mit einer neueren Anki-Version erstellt (Paketformat $version), die Cardflip noch nicht lesen kann';
  }

  @override
  String get goApkgNoCollection => 'es enthält keine Anki-Sammlung';

  @override
  String goApkgNotZip(String error) {
    return 'die Datei ist beschädigt oder kein ZIP-Archiv ($error)';
  }

  @override
  String goApkgNoteTypesUnreadable(String error) {
    return 'ihre Notiztypen konnten nicht gelesen werden ($error)';
  }

  @override
  String get goApkgUnreadable => 'kein lesbares Anki-Paket';

  @override
  String get goBackupDamaged => 'diese Sicherung ist beschädigt';

  @override
  String goBackupFirstFailed(String error) {
    return 'die Sammlung konnte vorher nicht gesichert werden, daher wurde nichts geändert: $error';
  }

  @override
  String get goBackupGone => 'diese Sicherung existiert nicht mehr';

  @override
  String goBackupIntervalRange(int max) {
    return 'Der Sicherungsabstand muss zwischen 0 und $max Minuten liegen.';
  }

  @override
  String get goBackupNewerFormat =>
      'diese Sicherung wurde mit einer neueren Version von Cardflip erstellt; aktualisiere Cardflip, um sie wiederherzustellen';

  @override
  String get goBackupNotImportable =>
      'eine Sicherung kann nicht importiert werden; stelle sie stattdessen in den Einstellungen wieder her';

  @override
  String goBackupsKeptRange(int max) {
    return 'Die Anzahl der aufbewahrten Sicherungen muss zwischen 0 und $max liegen.';
  }

  @override
  String goBrowserColumnTwice(String column) {
    return 'Die Spalte „$column“ wird im Browser doppelt angezeigt.';
  }

  @override
  String get goBrowserNeedsColumn =>
      'der Browser braucht mindestens eine Spalte';

  @override
  String get goBrowserSplitRange =>
      'Die Aufteilung des Browsers muss zwischen 0 und 1 liegen.';

  @override
  String goCardCountRange(int max) {
    return 'Die Anzahl der Karten muss zwischen 1 und $max liegen.';
  }

  @override
  String get goCardGone => 'Diese Karte existiert nicht mehr.';

  @override
  String get goCardTypeGone => 'Dieser Kartentyp existiert nicht mehr.';

  @override
  String get goCardTypeNameEmpty =>
      'Bitte gib einen Namen für den Kartentyp ein.';

  @override
  String get goCardTypeNameInvalid =>
      'Der Name eines Kartentyps darf weder : { } \" enthalten noch mit # ^ / beginnen.';

  @override
  String goCardTypeNameTaken(String name) {
    return 'Der Notiztyp hat schon einen Kartentyp namens „$name“.';
  }

  @override
  String goCardTypeOneTarget(String name) {
    return 'Kartentyp „$name“ kann nur einem Kartentyp zugeordnet werden.';
  }

  @override
  String goChangeLeavesNotesEmpty(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Diese Änderung würde $_temp0 ohne Karten zurücklassen und wurde daher nicht vorgenommen.';
  }

  @override
  String goCheckCardsWithoutNote(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten ohne Notiz wurden',
      one: '1 Karte ohne Notiz wurde',
    );
    return '$_temp0 entfernt.';
  }

  @override
  String goCheckFieldCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen mit der falschen Anzahl an Feldern wurden',
      one: '1 Notiz mit der falschen Anzahl an Feldern wurde',
    );
    return '$_temp0 repariert.';
  }

  @override
  String get goCheckFieldCountTrimmed =>
      'Der entfernte Feldinhalt steht im Löschprotokoll.';

  @override
  String goCheckInvalidCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten nicht mehr existierender Kartentypen wurden',
      one: '1 Karte eines nicht mehr existierenden Kartentyps wurde',
    );
    return '$_temp0 entfernt.';
  }

  @override
  String goCheckMissingCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString fehlende Karten wurden',
      one: '1 fehlende Karte wurde',
    );
    return '$_temp0 erzeugt.';
  }

  @override
  String goCheckMissingDeck(int count, String deck) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten in fehlenden Stapeln wurden',
      one: '1 Karte in einem fehlenden Stapel wurde',
    );
    return '$_temp0 nach $deck verschoben.';
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
      other:
          '$countString Notizen, deren Notiztyp nicht mehr existiert, wurden',
      one: '1 Notiz, deren Notiztyp nicht mehr existiert, wurde',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString Karten',
      one: '1 Karte',
    );
    return '$_temp0 mit $_temp1 entfernt. Die Notizen stehen im Löschprotokoll.';
  }

  @override
  String goCheckScheduling(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten mit ungültiger Planung wurden',
      one: '1 Karte mit ungültiger Planung wurde',
    );
    return '$_temp0 repariert.';
  }

  @override
  String goCheckSchedulingFromHistory(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Die Gedächtniszustände von $_temp0 wurden aus dem Wiederholungsverlauf neu berechnet.';
  }

  @override
  String goCheckTags(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other:
          '$countString Notizen mit doppelten oder ungültigen Schlagwörtern wurden',
      one: '1 Notiz mit doppelten oder ungültigen Schlagwörtern wurde',
    );
    return '$_temp0 repariert.';
  }

  @override
  String goChooseCardTypeFor(String noteType) {
    return 'Wähle für jeden Kartentyp von „$noteType“ einen Kartentyp.';
  }

  @override
  String goChooseFieldFor(String noteType) {
    return 'Wähle für jedes Feld von „$noteType“ ein Feld.';
  }

  @override
  String get goClipboardImageFormat =>
      'das Bild in der Zwischenablage ist kein PNG-, JPEG- oder GIF-Bild';

  @override
  String get goClozeOneCardType =>
      'Ein Lückentext-Notiztyp hat genau einen Kartentyp.';

  @override
  String goCollectionUnreadable(String error) {
    return 'Deine Sammlung konnte nicht gelesen werden: $error';
  }

  @override
  String get goCsvChooseTypeAndDeck => 'wähle einen Notiztyp und einen Stapel';

  @override
  String goCsvExtraFields(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Zeilen hatten',
      one: '1 Zeile hatte',
    );
    return '$_temp0 mehr Felder als ihr Notiztyp; die zusätzlichen Felder wurden nicht importiert.';
  }

  @override
  String goCsvFieldOutOfRange(int field, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Spalte ist Feld $field zugeordnet, aber der Notiztyp hat $countString Felder';
  }

  @override
  String get goCsvMapAField => 'ordne mindestens eine Spalte einem Feld zu';

  @override
  String get goCsvNoRows => 'die Datei hat keine Zeilen';

  @override
  String get goCsvNotUtf8 =>
      'die Datei ist kein UTF-8-Text; speichere sie als UTF-8 und versuche es erneut';

  @override
  String goCsvNoteTypeMissing(String name) {
    return 'Der Notiztyp „$name“ aus der Datei existiert nicht; Zeilen ohne eigenen Notiztyp verwenden den hier gewählten.';
  }

  @override
  String get goCsvOneDeckColumn => 'ordne nur eine Spalte dem Stapel zu';

  @override
  String get goCsvOneGuidColumn => 'ordne nur eine Spalte der GUID zu';

  @override
  String get goCsvOneNoteTypeColumn => 'ordne nur eine Spalte dem Notiztyp zu';

  @override
  String get goCsvOneTagsColumn => 'ordne nur eine Spalte den Schlagwörtern zu';

  @override
  String goCsvSkippedUnknownType(int count, String name) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Zeilen',
      one: '1 Zeile',
    );
    return '$_temp0 übersprungen: Es gibt keinen Notiztyp namens „$name“.';
  }

  @override
  String goCustomStudyNameTaken(String name) {
    return 'Ein normaler Stapel heißt „$name“. Benenne ihn um, um benutzerdefiniertes Lernen zu verwenden.';
  }

  @override
  String goCustomStudyNeedsRegular(String deck) {
    return 'Benutzerdefiniertes Lernen braucht einen normalen Stapel; „$deck“ ist ein gefilterter Stapel.';
  }

  @override
  String goDailyGoalRange(int max) {
    return 'Das Tagesziel muss zwischen 1 und $max liegen.';
  }

  @override
  String goRestDaysRange(int max) {
    return 'Die Ruhetage pro Woche müssen zwischen 0 und $max liegen.';
  }

  @override
  String goDayCountRange(int max) {
    return 'Die Anzahl der Tage muss zwischen 1 und $max liegen.';
  }

  @override
  String get goDeckGone => 'Dieser Stapel existiert nicht mehr.';

  @override
  String get goDeckMoveIntoItself =>
      'Ein Stapel kann nicht in sich selbst oder einen seiner Unterstapel verschoben werden.';

  @override
  String get goDeckNameEmpty => 'Bitte gib einen Stapelnamen ein.';

  @override
  String goDeckNameEmptyPart(String name) {
    return 'Der Stapelname „$name“ hat einen leeren Teil zwischen „::“.';
  }

  @override
  String goDeckNameTaken(String name) {
    return 'Ein Stapel namens „$name“ existiert bereits.';
  }

  @override
  String goDeckNameTakenByOther(String name) {
    return 'Der Name „$name“ wird bereits von einem anderen Stapel verwendet.';
  }

  @override
  String get goDefaultPresetUndeletable =>
      'Die Standard-Voreinstellung kann nicht gelöscht werden.';

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

    return 'Die gewünschte Behaltensrate muss zwischen $minString und $maxString liegen.';
  }

  @override
  String get goDueDateHelp =>
      'Gib Tage ab heute ein: 0 für heute, 1 für morgen, 60-90 für einen zufälligen Tag in diesem Bereich, mit ! um auch das Intervall zu setzen.';

  @override
  String goDueDateTooFar(int max) {
    return 'Ein Fälligkeitsdatum kann höchstens $max Tage entfernt sein.';
  }

  @override
  String goDueRangeReversed(String from, String to) {
    return 'Der Bereich $from-$to endet, bevor er beginnt; setze die kleinere Zahl an den Anfang.';
  }

  @override
  String get goFieldGone => 'Dieses Feld existiert nicht mehr.';

  @override
  String get goFieldNameEmpty => 'Bitte gib einen Feldnamen ein.';

  @override
  String get goFieldNameInvalid =>
      'Ein Feldname darf weder : { } \" enthalten noch mit # ^ / beginnen.';

  @override
  String goFieldNameTaken(String name) {
    return 'Der Notiztyp hat schon ein Feld namens „$name“.';
  }

  @override
  String get goFileGone =>
      'die Datei ist nicht mehr verfügbar; bitte wähle sie erneut';

  @override
  String goFileUnreadable(String error) {
    return 'die Datei konnte nicht gelesen werden: $error';
  }

  @override
  String get goFilteredFilterCount =>
      'Ein gefilterter Stapel hat einen oder zwei Filter.';

  @override
  String get goFilteredNoMatch =>
      'Keine Karten passen zur Suche. Ausgesetzte und zurückgestellte Karten sowie Karten in anderen gefilterten Stapeln werden nicht berücksichtigt.';

  @override
  String goFilteredNoNotes(String deck) {
    return '„$deck“ ist ein gefilterter Stapel, daher können keine Notizen hinzugefügt werden.';
  }

  @override
  String goFilteredNoSubdecks(String deck) {
    return '„$deck“ ist ein gefilterter Stapel und kann keine Unterstapel haben.';
  }

  @override
  String get goFindTextEmpty => 'Gib den zu suchenden Text ein.';

  @override
  String goFsrsParamRange(int index, String min, String max) {
    return 'FSRS-Parameter $index muss zwischen $min und $max liegen.';
  }

  @override
  String goFsrsParamsCount(int expected, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'FSRS-Parameter müssen $expected Zahlen sein, nicht $countString.';
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

    return 'Die bisherige Behaltensrate muss zwischen $minString und $maxString liegen.';
  }

  @override
  String goImportNotUpdatedNoCards(String note) {
    return 'Nicht aktualisiert, da sie keine Karten erzeugen würde: $note';
  }

  @override
  String goImportOtherNoteType(String note) {
    return 'Eine Notiz mit einem anderen Notiztyp existiert bereits und wurde nicht aktualisiert: $note';
  }

  @override
  String goImportPresetOptions(String preset, String error) {
    return 'Die Optionen der Voreinstellung „$preset“ konnten nicht importiert werden ($error); sie hat die Standardoptionen.';
  }

  @override
  String goImportSkipped(String note, String error) {
    return '„$note“ übersprungen: $error';
  }

  @override
  String goImportSkippedNoCards(String note) {
    return 'Übersprungen, da sie keine Karten erzeugen würde: „$note“';
  }

  @override
  String goInvalidDate(String value) {
    return '„$value“ ist kein Datum wie 2026-01-31.';
  }

  @override
  String goInvalidRating(String rating) {
    return 'Ungültige Bewertung $rating.';
  }

  @override
  String goInvalidRegex(String error) {
    return 'Ungültiger regulärer Ausdruck: $error';
  }

  @override
  String get goLastDeck =>
      'Der letzte Stapel kann nicht gelöscht werden. Eine Sammlung hat immer mindestens einen Stapel.';

  @override
  String get goLastNoteType =>
      'Der letzte Notiztyp kann nicht gelöscht werden. Eine Sammlung hat immer mindestens einen Notiztyp.';

  @override
  String goLastPositionTooHigh(int last, int max) {
    return 'Die letzte Position wäre $last; Positionen dürfen höchstens $max sein.';
  }

  @override
  String goLearnAheadRange(int max) {
    return 'das Vorauslernlimit muss zwischen 0 und $max Minuten liegen';
  }

  @override
  String get goLearningStepsRange =>
      'Jeder Lernschritt muss länger als 0 und höchstens 365 Tage sein.';

  @override
  String goLearningStepsTooMany(int max) {
    return 'Es sind höchstens $max Lernschritte möglich.';
  }

  @override
  String goLeechThresholdRange(int max) {
    return 'Die Schwelle für Problemkarten muss zwischen 1 und $max Fehlern liegen.';
  }

  @override
  String goMaxAnswerSecondsRange(int max) {
    return 'Die maximale Antwortzeit muss zwischen 1 und $max Sekunden liegen.';
  }

  @override
  String goMaxIntervalRange(int max) {
    return 'Das maximale Intervall muss zwischen 1 und $max Tagen liegen.';
  }

  @override
  String goMcpListenFailed(String error) {
    return 'Der Server konnte nicht starten: $error';
  }

  @override
  String goMcpPortInUse(String port) {
    return 'Port $port wird bereits verwendet.';
  }

  @override
  String goMcpPortRange(String min, String max) {
    return 'Der Port muss zwischen $min und $max liegen.';
  }

  @override
  String goMcpSettingsNotSaved(String error) {
    return 'Die Agenten-Einstellungen konnten nicht gespeichert werden: $error';
  }

  @override
  String goMcpSettingsUnreadable(String error) {
    return 'Die Agenten-Einstellungen konnten nicht gelesen werden: $error';
  }

  @override
  String goMediaEmptyTrashFailed(String error) {
    return 'Der Papierkorb konnte nicht geleert werden: $error';
  }

  @override
  String goMediaMissing(String file) {
    return 'Die Mediendatei „$file“ fehlt.';
  }

  @override
  String goMediaRestoreFailed(String error) {
    return 'Die Dateien konnten nicht wiederhergestellt werden: $error';
  }

  @override
  String goMediaSaveFailed(String error) {
    return 'Die Datei konnte nicht gespeichert werden: $error';
  }

  @override
  String goMediaTrashFailed(String error) {
    return 'Die ungenutzten Dateien konnten nicht in den Papierkorb verschoben werden: $error';
  }

  @override
  String goMinIntervalAboveMax(int min, int max) {
    return 'Das minimale Intervall ($min Tage) darf nicht über dem maximalen Intervall ($max Tage) liegen.';
  }

  @override
  String goMinIntervalRange(int max) {
    return 'Das minimale Intervall muss zwischen 1 und $max Tagen liegen.';
  }

  @override
  String get goMoveIntoFiltered =>
      'Karten können nicht in einen gefilterten Stapel verschoben werden; er sammelt seine Karten über seine Suche.';

  @override
  String get goNeedsCardType =>
      'Ein Notiztyp braucht mindestens einen Kartentyp.';

  @override
  String get goNeedsField => 'Ein Notiztyp braucht mindestens ein Feld.';

  @override
  String get goNewPerDayRange =>
      'Neue Karten pro Tag müssen zwischen 0 und 9999 liegen.';

  @override
  String get goNoCardsCloze =>
      'Dabei würden keine Karten entstehen. Füge eine Lücke wie {{c1::text}} hinzu.';

  @override
  String goNoCardsFillInField(String field) {
    return 'Dabei würden keine Karten entstehen. Fülle das Feld $field aus.';
  }

  @override
  String goNoCardsFillInFields(String fields, String last) {
    return 'Dabei würden keine Karten entstehen. Fülle eines der Felder $fields oder $last aus.';
  }

  @override
  String get goNoCardsNoField =>
      'Dabei würden keine Karten entstehen: Keine Frage einer Karte verwendet ein Feld.';

  @override
  String get goNoCardsSelected => 'Keine Karten ausgewählt.';

  @override
  String get goNoNotesSelected => 'Keine Notizen ausgewählt.';

  @override
  String get goNoStudySession => 'keine Lernsitzung; starte zuerst eine';

  @override
  String goNotAutomaticBackup(String name) {
    return '„$name“ ist keine automatische Sicherung';
  }

  @override
  String get goNotBackup => 'diese Datei ist keine Cardflip-Sicherung';

  @override
  String goNotFilteredDeck(String deck) {
    return '„$deck“ ist kein gefilterter Stapel.';
  }

  @override
  String get goNotPackageOrBackup =>
      'diese Datei ist weder ein Anki-Paket noch eine Cardflip-Sicherung';

  @override
  String get goNotTextFile => 'diese Datei ist keine Textdatei';

  @override
  String goNoteFieldCount(int count, String noteType, int expected) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Die Notiz hat $countString Felder, aber Notiztyp „$noteType“ hat $expected.';
  }

  @override
  String get goNoteGone => 'Diese Notiz existiert nicht mehr.';

  @override
  String get goNoteTypeGone => 'Dieser Notiztyp existiert nicht mehr.';

  @override
  String get goNoteTypeNameEmpty =>
      'Bitte gib einen Namen für den Notiztyp ein.';

  @override
  String goNoteTypeNameTaken(String name) {
    return 'Ein Notiztyp namens „$name“ existiert bereits.';
  }

  @override
  String goNotesWouldHaveNoCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen hätten',
      one: '1 Notiz hätte',
    );
    return '$_temp0 keine Karten. Ordne die Felder zu, die ihre Kartentypen verwenden, oder behalte einen Kartentyp.';
  }

  @override
  String get goNothingToRedo => 'Es gibt nichts zu wiederholen.';

  @override
  String get goNothingToReplace =>
      'die Sammlung ist lesbar; es gibt nichts zu ersetzen';

  @override
  String get goNothingToUndo => 'Es gibt nichts rückgängig zu machen.';

  @override
  String goOptimizeSearchInvalid(String error) {
    return 'Die Suche für die Optimierung ist ungültig: $error';
  }

  @override
  String goPositionStepRange(int max) {
    return 'Die Schrittweite muss zwischen 1 und $max liegen.';
  }

  @override
  String get goPresetGone => 'Diese Voreinstellung existiert nicht mehr.';

  @override
  String get goPresetNameEmpty =>
      'Bitte gib einen Namen für die Voreinstellung ein.';

  @override
  String goPresetNameTaken(String name) {
    return 'Eine Voreinstellung namens „$name“ existiert bereits.';
  }

  @override
  String goPreviewDelayRange(int max) {
    return 'Verzögerungen in der Vorschau müssen zwischen 0 und $max Sekunden liegen.';
  }

  @override
  String get goProfileGone => 'Dieses Profil existiert nicht mehr.';

  @override
  String get goProfileNameEmpty => 'Ein Profil braucht einen Namen.';

  @override
  String goProfileNameTaken(String name) {
    return 'Ein Profil namens „$name“ existiert bereits.';
  }

  @override
  String get goProfileOnly => 'Das einzige Profil kann nicht gelöscht werden.';

  @override
  String get goProfileOpen =>
      'Das geöffnete Profil kann nicht gelöscht werden. Öffne zuerst ein anderes Profil.';

  @override
  String goProfilesNotSaved(String error) {
    return 'Die Profilliste konnte nicht gespeichert werden: $error';
  }

  @override
  String goProfilesUnreadable(String error) {
    return 'Die Profilliste konnte nicht gelesen werden: $error';
  }

  @override
  String goQuestionFadeRange(int max) {
    return 'Das Einblenden der Frage muss zwischen 0 und $max Millisekunden liegen.';
  }

  @override
  String get goQuestionTemplateEmpty =>
      'Die Vorlage der Frage darf nicht leer sein.';

  @override
  String get goRelearningStepsRange =>
      'Jeder Neulernschritt muss länger als 0 und höchstens 365 Tage sein.';

  @override
  String goRelearningStepsTooMany(int max) {
    return 'Es sind höchstens $max Neulernschritte möglich.';
  }

  @override
  String goReservedFieldName(String name, String tag) {
    return '„$name“ ist ein reservierter Name: Vorlagen verwenden {{$tag}} für ein besonderes Feld. Bitte wähle einen anderen Namen.';
  }

  @override
  String get goReviewsPerDayRange =>
      'Die maximalen Wiederholungen pro Tag müssen zwischen 0 und 99999 liegen.';

  @override
  String get goRolloverHourRange =>
      'Die Stunde des Tageswechsels muss zwischen 0 und 23 liegen.';

  @override
  String get goSameNoteType =>
      'Alle ausgewählten Karten müssen denselben Notiztyp verwenden.';

  @override
  String goSavedSearchMissing(String name) {
    return 'es gibt keine gespeicherte Suche namens „$name“';
  }

  @override
  String get goSavedSearchNameEmpty =>
      'Eine gespeicherte Suche braucht einen Namen.';

  @override
  String goSavedSearchNameTaken(String name) {
    return 'Eine gespeicherte Suche namens „$name“ existiert bereits.';
  }

  @override
  String get goSearchCardTypeNumber =>
      'Kartentyp-Nummern beginnen bei 1, z. B. card:2';

  @override
  String get goSearchEmptyTerm => 'leerer Suchbegriff';

  @override
  String goSearchErrorAt(String message, String position) {
    return '$message (Zeichen $position)';
  }

  @override
  String goSearchErrorIn(String term, String message) {
    return '$term: $message';
  }

  @override
  String goSearchInvalidRegex(String error) {
    return 'ungültiger regulärer Ausdruck: $error';
  }

  @override
  String get goSearchNeedsCardType =>
      'braucht den Namen oder die Nummer eines Kartentyps, z. B. card:Rückseite oder card:2';

  @override
  String goSearchNeedsComparison(String property, String example) {
    return 'braucht einen Vergleich (=, !=, <, >, <= oder >=) mit einer Zahl, z. B. prop:$property$example';
  }

  @override
  String goSearchNeedsDayComparison(String property) {
    return 'braucht einen Vergleich mit einer Anzahl von Tagen, z. B. prop:$property=-1 (0 ist heute, -1 gestern)';
  }

  @override
  String goSearchNeedsDays(String key) {
    return 'braucht eine Anzahl von Tagen, z. B. $key:7';
  }

  @override
  String goSearchNeedsDaysToday(String key) {
    return 'braucht eine Anzahl von Tagen, z. B. $key:7 (1 ist heute)';
  }

  @override
  String goSearchNeedsIds(String key) {
    return 'braucht eine ID oder eine kommagetrennte Liste von IDs, z. B. $key:1712345678901,1712345678902';
  }

  @override
  String goSearchNeedsName(String key) {
    return 'braucht einen Namen, z. B. $key:Spanisch';
  }

  @override
  String get goSearchNeedsRegex =>
      'braucht einen regulären Ausdruck, z. B. re:^\\d+\$';

  @override
  String get goSearchNeedsWord => 'braucht ein Wort, z. B. w:Hund';

  @override
  String get goSearchNoEase =>
      'Cardflip plant mit FSRS, das keinen Leichtigkeitsfaktor hat; suche stattdessen die Schwierigkeit mit prop:d, z. B. prop:d>0.7';

  @override
  String goSearchOperatorAlone(String operator) {
    return '„$operator“ braucht auf beiden Seiten einen Suchbegriff';
  }

  @override
  String get goSearchParenEmpty => 'leere Klammern';

  @override
  String get goSearchParenNotClosed => 'eine Klammer wird nicht geschlossen';

  @override
  String get goSearchParenNotOpened =>
      'eine schließende Klammer hat keine öffnende';

  @override
  String get goSearchRating =>
      'die Bewertung muss 1 (Nochmal), 2 (Schwer), 3 (Gut) oder 4 (Einfach) sein, z. B. rated:7:1';

  @override
  String goSearchTooManyDays(int max) {
    return 'die Anzahl der Tage darf höchstens $max sein';
  }

  @override
  String get goSearchUnknownFlag =>
      'unbekannte Flagge; verwende flag:0 (keine Flagge) bis flag:7';

  @override
  String get goSearchUnknownProperty =>
      'unbekannte Eigenschaft; verwende prop: mit ivl, due, reps, lapses, pos, s, d, r, rated oder resched, z. B. prop:ivl>=30';

  @override
  String get goSearchUnknownState =>
      'unbekannter Zustand; verwende is:new, is:due, is:review, is:learn, is:suspended, is:buried, is:buried-manually oder is:buried-sibling';

  @override
  String goSecondsAnswerRange(int max) {
    return 'Die Sekunden, die die Antwort gezeigt wird, müssen zwischen 0 und $max liegen.';
  }

  @override
  String goSecondsQuestionRange(int max) {
    return 'Die Sekunden, die die Frage gezeigt wird, müssen zwischen 0 und $max liegen.';
  }

  @override
  String goSectionNotClosed(String tag) {
    return '$tag wird nicht geschlossen';
  }

  @override
  String goSectionNotOpened(String closing, String opening) {
    return 'zu $closing gibt es kein passendes $opening';
  }

  @override
  String goSectionsCrossed(String closing, String outer, String inner) {
    return '$closing schließt $outer, bevor $inner geschlossen ist';
  }

  @override
  String get goSelectCardsToChange =>
      'Wähle die Karten der zu ändernden Notizen aus.';

  @override
  String goSimulationDaysRange(int max) {
    return 'Die zu simulierenden Tage müssen zwischen 1 und $max liegen.';
  }

  @override
  String goStartPositionRange(int max) {
    return 'Die Startposition muss zwischen 0 und $max liegen.';
  }

  @override
  String get goStepAddField => 'Feld hinzufügen';

  @override
  String get goStepAddNote => 'Notiz hinzufügen';

  @override
  String get goStepAddNoteType => 'Notiztyp hinzufügen';

  @override
  String goStepAddNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 hinzufügen';
  }

  @override
  String get goStepAddPreset => 'Voreinstellung hinzufügen';

  @override
  String get goStepAddTags => 'Schlagwörter hinzufügen';

  @override
  String get goStepAddTemplate => 'Kartentyp hinzufügen';

  @override
  String get goStepAnswer => 'Karte beantworten';

  @override
  String get goStepBuildFiltered => 'Gefilterten Stapel erstellen';

  @override
  String goStepBury(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 zurückstellen';
  }

  @override
  String get goStepChange => 'Ändern';

  @override
  String get goStepChangeNoteType => 'Notiztyp ändern';

  @override
  String get goStepCheckDatabase => 'Datenbank prüfen';

  @override
  String get goStepCloneNoteType => 'Notiztyp klonen';

  @override
  String get goStepClonePreset => 'Voreinstellung klonen';

  @override
  String get goStepCreateDeck => 'Stapel erstellen';

  @override
  String get goStepCustomStudy => 'Benutzerdefiniertes Lernen';

  @override
  String get goStepDeckDescription => 'Stapelbeschreibung bearbeiten';

  @override
  String get goStepDeckLimits => 'Stapellimits ändern';

  @override
  String get goStepDeckOptions => 'Stapeloptionen ändern';

  @override
  String get goStepDeckPreset => 'Voreinstellung ändern';

  @override
  String get goStepDeleteDeck => 'Stapel löschen';

  @override
  String goStepDeleteEmptyCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString leere Karten',
      one: '1 leere Karte',
    );
    return '$_temp0 löschen';
  }

  @override
  String get goStepDeleteNoteType => 'Notiztyp löschen';

  @override
  String goStepDeleteNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 löschen';
  }

  @override
  String get goStepDeletePreset => 'Voreinstellung löschen';

  @override
  String get goStepDeleteTags => 'Schlagwörter löschen';

  @override
  String get goStepEmptyFiltered => 'Gefilterten Stapel leeren';

  @override
  String get goStepFieldOptions => 'Feldoptionen ändern';

  @override
  String get goStepFindReplace => 'Suchen und ersetzen';

  @override
  String goStepFlag(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 mit Flagge versehen';
  }

  @override
  String get goStepImport => 'Importieren';

  @override
  String goStepMark(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 markieren';
  }

  @override
  String goStepMoveCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 verschieben';
  }

  @override
  String get goStepMoveField => 'Feld verschieben';

  @override
  String get goStepMoveTags => 'Schlagwörter verschieben';

  @override
  String get goStepMoveTemplate => 'Kartentyp verschieben';

  @override
  String get goStepNoteTypeCSS => 'Stil bearbeiten';

  @override
  String get goStepOptimize => 'Alle Voreinstellungen optimieren';

  @override
  String get goStepPresetToSubdecks => 'Für alle Unterstapel speichern';

  @override
  String get goStepRemoveField => 'Feld entfernen';

  @override
  String get goStepRemoveTags => 'Schlagwörter entfernen';

  @override
  String get goStepRemoveTemplate => 'Kartentyp entfernen';

  @override
  String get goStepRenameDeck => 'Stapel umbenennen';

  @override
  String get goStepRenameField => 'Feld umbenennen';

  @override
  String get goStepRenameNoteType => 'Notiztyp umbenennen';

  @override
  String get goStepRenamePreset => 'Voreinstellung umbenennen';

  @override
  String get goStepRenameTag => 'Schlagwort umbenennen';

  @override
  String get goStepReposition => 'Neu positionieren';

  @override
  String goStepResetCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 zurücksetzen';
  }

  @override
  String goStepSetDueDate(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Fälligkeitsdatum von $_temp0 setzen';
  }

  @override
  String get goStepSortField => 'Sortierfeld ändern';

  @override
  String goStepSuspend(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 aussetzen';
  }

  @override
  String get goStepTodayLimits => 'Benutzerdefiniertes Lernen';

  @override
  String get goStepUnbury => 'Zurückstellung aufheben';

  @override
  String get goStepUndoAnswer => 'Antwort rückgängig machen';

  @override
  String goStepUnflag(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Flagge von $_temp0 entfernen';
  }

  @override
  String goStepUnmark(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Markierung von $_temp0 entfernen';
  }

  @override
  String goStepUnsuspend(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Aussetzung von $_temp0 aufheben';
  }

  @override
  String get goStepUpdateNote => 'Notiz bearbeiten';

  @override
  String get goStepUpdateTemplate => 'Kartentyp bearbeiten';

  @override
  String goStorageOpenFailed(String error) {
    return 'Der Speicher der Sammlung konnte nicht geöffnet werden: $error';
  }

  @override
  String goSyncAlreadyLinked(String profile, String device) {
    return 'Das Profil „$profile“ ist bereits mit einem Profil auf $device verknüpft.';
  }

  @override
  String get goSyncAlreadyLinkedHere =>
      'Dieses Profil ist bereits mit einem Profil auf diesem Gerät verknüpft.';

  @override
  String get goSyncBadAddress =>
      'Gib die Adresse und den Port ein, die auf dem anderen Gerät angezeigt werden.';

  @override
  String goSyncBusy(String device) {
    return '$device synchronisiert gerade mit einem anderen Gerät. Versuche es gleich noch einmal.';
  }

  @override
  String goSyncClocksDifferDays(String device, int days) {
    return 'Die Uhren dieses Geräts und von $device weichen um $days Tage voneinander ab. Stelle auf beiden die richtige Zeit ein und versuche es noch einmal.';
  }

  @override
  String goSyncClocksDifferHours(String device, int hours) {
    return 'Die Uhren dieses Geräts und von $device weichen um $hours Stunden voneinander ab. Stelle auf beiden die richtige Zeit ein und versuche es noch einmal.';
  }

  @override
  String goSyncClocksDifferMinutes(String device, int minutes) {
    return 'Die Uhren dieses Geräts und von $device weichen um $minutes Minuten voneinander ab. Stelle auf beiden die richtige Zeit ein und versuche es noch einmal.';
  }

  @override
  String goSyncConnectFailed(String device, String error) {
    return '$device ist nicht erreichbar: $error';
  }

  @override
  String goSyncConnectTimeout(String device, String address, String port) {
    return '$device antwortet nicht unter $address. Wenn das Gerät an ist, blockiert dort vielleicht eine Firewall die Synchronisierung: Erlaube dort eingehende TCP-Verbindungen auf Port $port.';
  }

  @override
  String goSyncDeviceNameLength(int max) {
    return 'Der Gerätename muss 1 bis $max Zeichen lang sein.';
  }

  @override
  String get goSyncLinkGone => 'Diese Verknüpfung existiert nicht mehr.';

  @override
  String goSyncLinksUnreadable(String error) {
    return 'Die Verknüpfungen dieses Profils konnten nicht gelesen werden: $error';
  }

  @override
  String goSyncListenFailed(String error) {
    return 'Die Synchronisierung konnte nicht starten: $error';
  }

  @override
  String goSyncNeedsChoice(String device) {
    return 'Wähle, wie dieses Profil mit dem auf $device zusammengeführt werden soll.';
  }

  @override
  String get goSyncNoNetwork =>
      'Dieses Gerät ist mit keinem Netzwerk verbunden, dem vertraut werden kann.';

  @override
  String get goSyncNotPaired => 'Dieses Gerät ist nicht gekoppelt.';

  @override
  String goSyncNotReachable(String device) {
    return '$device ist in diesem Netzwerk nicht erreichbar.';
  }

  @override
  String get goSyncNotSetUp => 'Öffne zuerst Einstellungen → Synchronisierung.';

  @override
  String goSyncPairConnectFailed(String address, String error) {
    return 'Das Gerät unter $address ist nicht erreichbar: $error';
  }

  @override
  String goSyncPairDeclined(String device) {
    return '$device hat die Kopplungsanfrage abgelehnt.';
  }

  @override
  String goSyncPairFailed(String device) {
    return 'Die Kopplung mit $device ist fehlgeschlagen: Die Codes wurden nicht auf beiden Geräten bestätigt.';
  }

  @override
  String goSyncPairNotVisible(String device) {
    return '$device ist nicht zum Koppeln sichtbar. Öffne dort „Gerät koppeln“ und versuche es noch einmal.';
  }

  @override
  String goSyncPairTimeout(String address, String port) {
    return 'Das Gerät unter $address antwortet nicht. Dort blockiert vielleicht eine Firewall die Synchronisierung: Erlaube dort eingehende TCP-Verbindungen auf Port $port, oder starte das Koppeln auf diesem Gerät.';
  }

  @override
  String goSyncStateNotSaved(String error) {
    return 'Die Synchronisierungseinstellungen konnten nicht gespeichert werden: $error';
  }

  @override
  String goSyncStateUnreadable(String error) {
    return 'Die Synchronisierungseinstellungen konnten nicht gelesen werden: $error';
  }

  @override
  String goSyncUnknownLink(String device) {
    return '$device hat diese Verknüpfung nicht mehr. Verknüpfe die Profile erneut.';
  }

  @override
  String get goSyncUntrusted => 'Vertraue zuerst diesem Netzwerk.';

  @override
  String goSyncUpdateOther(String device) {
    return 'Aktualisiere Cardflip auf $device, um damit zu synchronisieren.';
  }

  @override
  String goSyncUpdateThis(String device) {
    return 'Aktualisiere Cardflip auf diesem Gerät, um mit $device zu synchronisieren.';
  }

  @override
  String get goTagMoveIntoItself =>
      'Ein Schlagwort kann nicht in sich selbst oder eines seiner Unterschlagwörter verschoben werden.';

  @override
  String get goTagNameEmpty => 'Gib einen Namen für das Schlagwort ein.';

  @override
  String get goTagNameSpaces =>
      'Schlagwörter dürfen keine Leerzeichen enthalten.';

  @override
  String goTagNoFieldName(String tag) {
    return '$tag hat keinen Feldnamen';
  }

  @override
  String goTagNotClosed(String tag) {
    return '$tag hat kein schließendes }}';
  }

  @override
  String goTemplateBackMistake(String problem) {
    return 'In der Vorlage der Rückseite: $problem.';
  }

  @override
  String goTemplateBackUnknownField(String tag, String name) {
    return 'Die Vorlage der Rückseite verwendet $tag, aber es gibt kein Feld „$name“.';
  }

  @override
  String goTemplateFrontMistake(String problem) {
    return 'In der Vorlage der Vorderseite: $problem.';
  }

  @override
  String goTemplateFrontUnknownField(String tag, String name) {
    return 'Die Vorlage der Vorderseite verwendet $tag, aber es gibt kein Feld „$name“.';
  }

  @override
  String goTemplateNeedsCloze(String tag) {
    return 'Die Vorlage der Vorderseite muss $tag oder ein anderes Lückentextfeld verwenden.';
  }

  @override
  String get goTemplateNoField =>
      'Die Vorlage der Vorderseite verwendet kein Feld und würde daher nie eine Karte erzeugen.';

  @override
  String goTimeboxRange(int max) {
    return 'Das Zeitlimit der Lerneinheit muss zwischen 0 und $max Minuten liegen.';
  }

  @override
  String goTodayIncreaseRange(int max) {
    return 'Die Erhöhung muss zwischen 0 und $max Karten liegen.';
  }

  @override
  String goUiScaleRange(int min, int max, int step) {
    return 'Die Oberflächengröße muss zwischen $min und $max % in Schritten von $step % liegen.';
  }

  @override
  String get goUndoOnlyCardflip =>
      'Nur Antworten, die in Cardflip gegeben wurden, können rückgängig gemacht werden.';

  @override
  String get goUndoOnlyLatest =>
      'Nur die letzte Antwort auf eine Karte kann rückgängig gemacht werden.';

  @override
  String goUnknownAnswerAction(String value) {
    return 'Unbekannte Aktion für die Antwort „$value“.';
  }

  @override
  String goUnknownBrowserColumn(String value) {
    return 'Unbekannte Browserspalte „$value“.';
  }

  @override
  String goUnknownBrowserMode(String value) {
    return 'Unbekannter Browsermodus „$value“.';
  }

  @override
  String goUnknownCardState(String value) {
    return 'Unbekannter Kartenzustand „$value“.';
  }

  @override
  String goUnknownCustomStudy(String value) {
    return 'Unbekannte Auswahl für benutzerdefiniertes Lernen „$value“.';
  }

  @override
  String goUnknownDailyGoal(String value) {
    return 'Unbekanntes Tagesziel „$value“.';
  }

  @override
  String goUnknownEasyDaysLevel(String value) {
    return 'Unbekannte Stufe für leichte Tage $value.';
  }

  @override
  String goUnknownFilteredOrder(String value) {
    return 'Unbekannte Reihenfolge für gefilterte Stapel „$value“.';
  }

  @override
  String goUnknownFlag(String flag) {
    return 'Es gibt keine Flagge $flag.';
  }

  @override
  String goUnknownGatherOrder(String value) {
    return 'Unbekannte Sammelreihenfolge für neue Karten „$value“.';
  }

  @override
  String goUnknownInsertionOrder(String value) {
    return 'Unbekannte Einfügereihenfolge „$value“.';
  }

  @override
  String goUnknownInterdayOrder(String value) {
    return 'Unbekannte Reihenfolge für Lernen über Tage und Wiederholungen „$value“.';
  }

  @override
  String goUnknownLanguage(String value) {
    return 'Unbekannte Sprache „$value“.';
  }

  @override
  String goUnknownLeechAction(String value) {
    return 'Unbekannte Aktion für Problemkarten „$value“.';
  }

  @override
  String goUnknownNewReviewOrder(String value) {
    return 'Unbekannte Reihenfolge für neue Karten und Wiederholungen „$value“.';
  }

  @override
  String goUnknownNewSortOrder(String value) {
    return 'Unbekannte Sortierreihenfolge für neue Karten „$value“.';
  }

  @override
  String goUnknownNoteSort(String value) {
    return 'Notizen können nicht nach „$value“ sortiert werden.';
  }

  @override
  String goUnknownQuestionAction(String value) {
    return 'Unbekannte Aktion für die Frage „$value“.';
  }

  @override
  String goUnknownReduceMotion(String value) {
    return 'Unbekannte Auswahl für „Bewegung reduzieren“ „$value“.';
  }

  @override
  String goUnknownReviewSortOrder(String value) {
    return 'Unbekannte Sortierreihenfolge für Wiederholungen „$value“.';
  }

  @override
  String goUnknownSeparator(String value) {
    return 'unbekanntes Trennzeichen „$value“';
  }

  @override
  String goUnknownStockType(String value) {
    return 'Unbekannter mitgelieferter Notiztyp „$value“.';
  }

  @override
  String goUnknownTheme(String value) {
    return 'Unbekanntes Design „$value“.';
  }

  @override
  String get datePattern => 'dd.MM.yyyy';

  @override
  String dueAt(String time) {
    return 'um $time';
  }

  @override
  String get dueTomorrow => 'morgen';

  @override
  String dueInDays(int days) {
    final intl.NumberFormat daysNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String daysString = daysNumberFormat.format(days);

    String _temp0 = intl.Intl.pluralLogic(
      days,
      locale: localeName,
      other: '$daysString Tagen',
      one: '1 Tag',
    );
    return 'in $_temp0';
  }

  @override
  String dueOn(String date) {
    return 'am $date';
  }

  @override
  String spentSeconds(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Sekunden',
      one: '1 Sekunde',
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
      other: '$countString Minuten',
      one: '1 Minute',
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

    return '$hoursString Std. $minutesString Min.';
  }

  @override
  String intervalMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Minuten',
      one: '1 Minute',
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
      other: '$countString Stunden',
      one: '1 Stunde',
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
      other: '$countString Tage',
      one: '1 Tag',
    );
    return '$_temp0';
  }

  @override
  String intervalMonths(String months) {
    return '$months Monate';
  }

  @override
  String intervalYears(String years) {
    return '$years Jahre';
  }

  @override
  String intervalLabelMinutes(int minutes) {
    final intl.NumberFormat minutesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String minutesString = minutesNumberFormat.format(minutes);

    return '<$minutesString Min.';
  }

  @override
  String intervalLabelHours(String hours) {
    return '$hours Std.';
  }

  @override
  String intervalLabelDays(int days) {
    final intl.NumberFormat daysNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String daysString = daysNumberFormat.format(days);

    return '$daysString T';
  }

  @override
  String intervalLabelMonths(String months) {
    return '$months Mon.';
  }

  @override
  String intervalLabelYears(String years) {
    return '$years J';
  }

  @override
  String sizeBytes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Bytes',
      one: '1 Byte',
    );
    return '$_temp0';
  }

  @override
  String get cancel => 'Abbrechen';

  @override
  String get save => 'Speichern';

  @override
  String get typeTheAnswer => 'Antwort eingeben';

  @override
  String get missingImage => 'Fehlendes Bild';

  @override
  String missingImageNamed(String alt) {
    return 'Fehlendes Bild: $alt';
  }

  @override
  String get stop => 'Stopp';

  @override
  String playFile(String file) {
    return '$file abspielen';
  }

  @override
  String get speak => 'Vorlesen';

  @override
  String get noVoicesInstalled =>
      'Auf diesem Gerät sind keine Stimmen installiert.';

  @override
  String fileMissing(String file) {
    return '$file fehlt';
  }

  @override
  String fileCannotPlay(String file) {
    return '$file kann auf diesem Gerät nicht abgespielt werden';
  }

  @override
  String noVoiceForLanguage(String language) {
    return 'Für $language ist keine Stimme installiert';
  }

  @override
  String get noVoiceForThisLanguage =>
      'Für diese Sprache ist keine Stimme installiert';

  @override
  String get microphoneRefused =>
      'Cardflip darf das Mikrofon nicht verwenden. Erlaube den Zugriff auf das Mikrofon in deinen Systemeinstellungen, um deine Stimme aufzunehmen.';

  @override
  String get studyDeck => 'Stapel lernen';

  @override
  String get searchDecks => 'Stapel suchen';

  @override
  String get noMatchingDecks => 'Keine passenden Stapel';

  @override
  String unburiedDeck(String deck) {
    return 'Zurückstellung von „$deck“ aufgehoben';
  }

  @override
  String get editDescription => 'Beschreibung bearbeiten';

  @override
  String get tryAgain => 'Erneut versuchen';

  @override
  String countBuriedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString zurückgestellte Karten',
      one: '1 zurückgestellte Karte',
    );
    return '$_temp0';
  }

  @override
  String get unbury => 'Zurückstellung aufheben';

  @override
  String get studyNow => 'Jetzt lernen';

  @override
  String get options => 'Optionen';

  @override
  String get rebuild => 'Neu erstellen';

  @override
  String get emptyDeck => 'Leeren';

  @override
  String get countNew => 'Neu';

  @override
  String get countLearning => 'Lernen';

  @override
  String get countToReview => 'Wiederholen';

  @override
  String get nothingDueNow => 'Gerade ist nichts fällig.';

  @override
  String nextCardDue(String when) {
    return 'Die nächste Karte ist $when fällig.';
  }

  @override
  String get deckHasNoCards =>
      'Dieser Stapel hat keine Karten zum Lernen. Füge Karten hinzu oder importiere welche, um loszulegen.';

  @override
  String get description => 'Beschreibung';

  @override
  String get descriptionHelper =>
      'Einfacher Text oder HTML wie <b>fett</b> und Links';

  @override
  String enterWholeNumber(int max) {
    return 'Gib eine ganze Zahl von 1 bis $max ein.';
  }

  @override
  String moreAvailable(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString weitere verfügbar';
  }

  @override
  String get increaseBy => 'Erhöhen um';

  @override
  String get unitCards => 'Karten';

  @override
  String get unitDays => 'Tage';

  @override
  String get increaseHelper =>
      'Nur für heute. Die Stapeloptionen bleiben unverändert.';

  @override
  String get cards => 'Karten';

  @override
  String get newCardsOnly => 'Nur neue Karten';

  @override
  String get dueCardsOnly => 'Nur fällige Karten';

  @override
  String get allCardsRandom =>
      'Alle Karten in zufälliger Reihenfolge (ohne Neuplanung)';

  @override
  String get onlyWithTags => 'Nur mit diesen Schlagwörtern (optional)';

  @override
  String get onlyWithTagsHelper =>
      'Karten mit einem davon; Schlagwörter durch Leerzeichen trennen';

  @override
  String get select => 'Auswählen';

  @override
  String get next => 'Nächste';

  @override
  String get last => 'Letzte';

  @override
  String customStudySessionHelper(String deck) {
    return 'Gelernt im gefilterten Stapel „$deck“, der die vorherige Sitzung ersetzt.';
  }

  @override
  String customStudyTitle(String deck) {
    return 'Benutzerdefiniertes Lernen: $deck';
  }

  @override
  String get increaseNewLimit => 'Heutiges Limit für neue Karten erhöhen';

  @override
  String get increaseReviewLimit => 'Heutiges Limit für Wiederholungen erhöhen';

  @override
  String get reviewForgotten => 'Vergessene Karten wiederholen';

  @override
  String get reviewAhead => 'Vorauswiederholen';

  @override
  String get previewNewCards => 'Vorschau neuer Karten';

  @override
  String get studyByStateOrTag => 'Nach Kartenzustand oder Schlagwort lernen';

  @override
  String get ok => 'OK';

  @override
  String get newDeck => 'Neuer Stapel';

  @override
  String get name => 'Name';

  @override
  String get newDeckHelper =>
      'Verwende :: für Unterstapel, z. B. Sprachen::Spanisch';

  @override
  String get create => 'Erstellen';

  @override
  String get decks => 'Stapel';

  @override
  String get statistics => 'Statistiken';

  @override
  String get createDeck => 'Stapel erstellen';

  @override
  String get createFilteredDeck => 'Gefilterten Stapel erstellen';

  @override
  String get collectionEmpty => 'Deine Sammlung ist leer';

  @override
  String get collectionEmptyHint =>
      'Füge deine erste Karte hinzu oder importiere Karten aus einer Textdatei (.csv, .tsv oder .txt), einem Anki-Stapel (.apkg) oder einer ganzen Anki-Sammlung (.colpkg).';

  @override
  String get addACard => 'Karte hinzufügen';

  @override
  String get importAFile => 'Datei importieren';

  @override
  String get renameDeck => 'Stapel umbenennen';

  @override
  String get renameDeckHelper =>
      'Verwende ::, um ihn in einen anderen Stapel zu verschieben';

  @override
  String get rename => 'Umbenennen';

  @override
  String deleteNamedTitle(String name) {
    return '„$name“ löschen?';
  }

  @override
  String deleteDeckBody(String cards, String inFiltered, String subdecks) {
    return 'Dadurch werden $cards$inFiltered$subdecks gelöscht. Du kannst das rückgängig machen, bis du Cardflip schließt.';
  }

  @override
  String deleteDeckInFiltered(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return ' (davon $countString gerade in gefilterten Stapeln)';
  }

  @override
  String deleteDeckSubdecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Unterstapel',
      one: '1 Unterstapel',
    );
    return ' und $_temp0';
  }

  @override
  String countCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0';
  }

  @override
  String get delete => 'Löschen';

  @override
  String deletedNamed(String name) {
    return '„$name“ gelöscht';
  }

  @override
  String deletedDeckReturned(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten sind',
      one: '1 Karte ist',
    );
    return '„$deck“ gelöscht; $_temp0 in ihre Heimatstapel zurückgekehrt';
  }

  @override
  String get expand => 'Aufklappen';

  @override
  String get collapse => 'Zuklappen';

  @override
  String get filteredDeck => 'Gefilterter Stapel';

  @override
  String get deckActions => 'Stapelaktionen';

  @override
  String get study => 'Lernen';

  @override
  String get addCard => 'Karte hinzufügen';

  @override
  String get customStudy => 'Benutzerdefiniertes Lernen';

  @override
  String get export => 'Exportieren';

  @override
  String dueNewTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString neu';
  }

  @override
  String dueLearningTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString im Lernen';
  }

  @override
  String dueReviewTooltip(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString zu wiederholen';
  }

  @override
  String get orderOldestSeen => 'Am längsten nicht gesehen zuerst';

  @override
  String get orderRandom => 'Zufällig';

  @override
  String get orderIntervalsAscending => 'Aufsteigende Intervalle';

  @override
  String get orderIntervalsDescending => 'Absteigende Intervalle';

  @override
  String get orderMostLapses => 'Meiste Fehler';

  @override
  String get orderAdded => 'Reihenfolge des Hinzufügens';

  @override
  String get orderDue => 'Reihenfolge der Fälligkeit';

  @override
  String get orderLatestAdded => 'Zuletzt hinzugefügte zuerst';

  @override
  String get orderRetrievability => 'Niedrigste Abrufbarkeit zuerst';

  @override
  String filteredDeckHolds(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '„$deck“ enthält $_temp0';
  }

  @override
  String rebuiltDeck(String deck, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '„$deck“ neu erstellt: $_temp0';
  }

  @override
  String emptiedDeck(String deck) {
    return '„$deck“ geleert';
  }

  @override
  String filterLimitError(String which) {
    String _temp0 = intl.Intl.selectLogic(which, {
      'first': 'ersten',
      'other': 'zweiten',
    });
    return 'Gib für den $_temp0 Filter eine Anzahl von Karten von 1 bis 99999 ein.';
  }

  @override
  String get previewDelaysError =>
      'Gib Verzögerungen für die Vorschau von 0 bis 86400 Sekunden ein.';

  @override
  String get search => 'Suchen';

  @override
  String get filterSearchHelper => 'Wie im Browser, z. B. deck:Spanisch is:due';

  @override
  String get limitTo => 'Begrenzen auf';

  @override
  String get cardsSelectedBy => 'Karten ausgewählt nach';

  @override
  String get build => 'Erstellen';

  @override
  String get newFilteredDeck => 'Neuer gefilterter Stapel';

  @override
  String get filteredDeckOptions => 'Optionen des gefilterten Stapels';

  @override
  String get filter => 'Filter';

  @override
  String get secondFilter => 'Zweiter Filter';

  @override
  String get secondFilterHelper =>
      'Fügt nach dem ersten die Karten einer weiteren Suche hinzu';

  @override
  String get reschedule =>
      'Karten anhand meiner Antworten in diesem Stapel neu planen';

  @override
  String get rescheduleOn => 'Antworten zählen als normale Wiederholungen';

  @override
  String get rescheduleOff =>
      'Vorschau: Antworten ändern nicht, wann Karten fällig sind';

  @override
  String get previewDelays =>
      'Karte erneut zeigen nach (Sekunden; 0 schickt sie in ihren Heimatstapel zurück)';

  @override
  String get again => 'Nochmal';

  @override
  String get hard => 'Schwer';

  @override
  String get good => 'Gut';

  @override
  String get easy => 'Einfach';

  @override
  String get unitSecondsShort => 's';

  @override
  String get filteredDeckNote =>
      'Ausgesetzte und zurückgestellte Karten sowie Karten in anderen gefilterten Stapeln werden nicht berücksichtigt. Wenn du den Stapel leerst oder löschst, kehren seine Karten in ihre Heimatstapel zurück.';

  @override
  String get simDaysError =>
      'Gib zwischen 1 und 3650 zu simulierende Tage ein.';

  @override
  String get simLimitsError => 'Gib ganze Zahlen für die Tageslimits ein.';

  @override
  String useRetentionTitle(String retention) {
    return 'Gewünschte Behaltensrate von $retention verwenden?';
  }

  @override
  String get useRetentionBody =>
      'Die Stapeloptionen erhalten diese gewünschte Behaltensrate. Sie wird gespeichert, wenn du die Optionen speicherst.';

  @override
  String get use => 'Verwenden';

  @override
  String simulateDeck(String deck) {
    return '$deck simulieren';
  }

  @override
  String get simulatorIntro =>
      'Erwartete Wiederholungen und gemerkte Karten dieses Stapels und seiner Unterstapel, ausgehend von ihren heutigen Karten. Antwortzeiten und Bewertungen stammen aus deinem Wiederholungsverlauf. Es wird nichts geändert.';

  @override
  String get daysToSimulate => 'Zu simulierende Tage';

  @override
  String get newCardsPerDayShort => 'Neue Karten/Tag';

  @override
  String get maxReviewsPerDayShort => 'Maximale Wiederholungen/Tag';

  @override
  String desiredRetentionValue(String retention) {
    return 'Gewünschte Behaltensrate: $retention';
  }

  @override
  String get compareRetention => 'Mit einer anderen Behaltensrate vergleichen';

  @override
  String comparedRetentionValue(String retention) {
    return 'Verglichene Behaltensrate: $retention';
  }

  @override
  String get simulate => 'Simulieren';

  @override
  String get helpMeDecide => 'Hilf mir bei der Entscheidung';

  @override
  String suggestedRetention(String retention) {
    return 'Vorgeschlagene gewünschte Behaltensrate: $retention';
  }

  @override
  String get suggestedRetentionHelp =>
      'Die Behaltensrate mit der geringsten Wiederholungszeit pro gemerkter Karte am Ende der Simulation. Eine niedrigere kostet weniger Zeit, aber du behältst weniger.';

  @override
  String get secondsPerMemorisedCard => 'Sekunden pro gemerkter Karte';

  @override
  String useValue(String value) {
    return '$value verwenden';
  }

  @override
  String dayNumber(int day) {
    return 'Tag $day';
  }

  @override
  String get reviewsPerDay => 'Wiederholungen pro Tag';

  @override
  String get memorisedCards => 'Gemerkte Karten';

  @override
  String atRetention(String retention) {
    return 'Bei $retention';
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
      other: '$reviewsString Wiederholungen',
      one: '1 Wiederholung',
    );
    String _temp1 = intl.Intl.pluralLogic(
      newCards,
      locale: localeName,
      other: '$newCardsString neue Karten',
      one: '1 neue Karte',
    );
    return '$_temp0 ($perDay pro Tag), $_temp1 gelernt';
  }

  @override
  String simTime(String total, String perDay) {
    return 'Zeit: $total ($perDay pro Tag)';
  }

  @override
  String simMemorisedAtEnd(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Am Ende gemerkt: $_temp0';
  }

  @override
  String get readingDeck => 'Stapel wird gelesen…';

  @override
  String simulatingPercent(String percent) {
    return 'Simuliere… $percent';
  }

  @override
  String get simulating => 'Simuliere…';

  @override
  String get enterDaysError =>
      'Gib eine ganze Zahl von Tagen von 1 bis 36500 ein.';

  @override
  String get notAboveMaxInterval =>
      'Darf nicht über dem maximalen Intervall liegen.';

  @override
  String enterSecondsError(int min, int max) {
    return 'Gib eine ganze Zahl von Sekunden von $min bis $max ein.';
  }

  @override
  String get leechThresholdError =>
      'Gib eine ganze Zahl für die Schwelle für Problemkarten ein.';

  @override
  String get correctSteps => 'Korrigiere zuerst die Schritte.';

  @override
  String get correctIntervals => 'Korrigiere zuerst die Intervalle.';

  @override
  String get correctMaxAnswer => 'Korrigiere zuerst die maximale Antwortzeit.';

  @override
  String get correctAutoAdvance =>
      'Korrigiere zuerst die Sekunden für das automatische Weiterblättern.';

  @override
  String get correctFsrs => 'Korrigiere zuerst die FSRS-Parameter.';

  @override
  String get deckLimitsError =>
      'Gib ganze Zahlen für die Limits dieses Stapels ein.';

  @override
  String get rescheduleTitle => 'Karten neu planen?';

  @override
  String rescheduleBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other:
          '$countString Wiederholungskarten erhalten aus ihrem Wiederholungsverlauf neue Fälligkeitsdaten',
      one: '1 Wiederholungskarte erhält aus ihrem Wiederholungsverlauf ein neues Fälligkeitsdatum',
    );
    return '$_temp0, mit der neuen gewünschten Behaltensrate und den neuen FSRS-Parametern.';
  }

  @override
  String get rescheduleAction => 'Neu planen';

  @override
  String get optimizeAllTitle => 'Alle Voreinstellungen optimieren?';

  @override
  String get optimizeAllBody =>
      'Jede Voreinstellung erhält an ihre Wiederholungen angepasste FSRS-Parameter, wo diese besser passen als ihre aktuellen.';

  @override
  String optimizeAllReschedule(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungskarten',
      one: '1 Wiederholungskarte',
    );
    return ' Bis zu $_temp0 von Voreinstellungen, die Karten bei Änderungen neu planen, erhalten neue Fälligkeitsdaten.';
  }

  @override
  String get optimize => 'Optimieren';

  @override
  String optimizeAllResult(int applied) {
    final intl.NumberFormat appliedNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String appliedString = appliedNumberFormat.format(applied);

    String _temp0 = intl.Intl.pluralLogic(
      applied,
      locale: localeName,
      other: '$appliedString Voreinstellungen haben',
      one: '1 Voreinstellung hat',
      zero: 'Keine Voreinstellung hat',
    );
    return '$_temp0 neue Parameter erhalten.';
  }

  @override
  String optimizeAllFew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Voreinstellungen haben',
      one: '1 Voreinstellung hat',
    );
    return ' $_temp0 zu wenige Wiederholungen zum Optimieren.';
  }

  @override
  String optimizePreset(int number, int total) {
    return 'Voreinstellung $number von $total: ';
  }

  @override
  String get optimizeCancelled =>
      'Das Optimieren wurde abgebrochen. Es wurde nichts geändert.';

  @override
  String get ignoreReviewedBefore =>
      'Karten ignorieren, die wiederholt wurden vor';

  @override
  String get discardChangesTitle => 'Änderungen verwerfen?';

  @override
  String presetNotSaved(String preset) {
    return 'Deine Änderungen an der Voreinstellung „$preset“ sind nicht gespeichert.';
  }

  @override
  String get discard => 'Verwerfen';

  @override
  String get addPreset => 'Voreinstellung hinzufügen';

  @override
  String get addPresetHelper => 'Beginnt mit den Standardoptionen';

  @override
  String get add => 'Hinzufügen';

  @override
  String get clonePreset => 'Voreinstellung klonen';

  @override
  String presetCopyName(String name) {
    return '$name Kopie';
  }

  @override
  String get clonePresetHelper => 'Beginnt mit den angezeigten Optionen';

  @override
  String get clone => 'Klonen';

  @override
  String get renamePreset => 'Voreinstellung umbenennen';

  @override
  String get deletePresetTitle => 'Voreinstellung löschen?';

  @override
  String presetUnused(String preset) {
    return 'Die Voreinstellung „$preset“ wird von keinem Stapel verwendet.';
  }

  @override
  String presetUsedBy(String preset, int count, String fallback) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stapeln verwendet. Sie verwenden',
      one: '1 Stapel verwendet. Er verwendet',
    );
    return 'Die Voreinstellung „$preset“ wird von $_temp0 stattdessen $fallback.';
  }

  @override
  String allSubdecksUsePreset(String preset) {
    return 'Alle Unterstapel verwenden jetzt die Voreinstellung „$preset“.';
  }

  @override
  String optionsFor(String deck) {
    return 'Optionen für $deck';
  }

  @override
  String get preset => 'Voreinstellung';

  @override
  String get presetActions => 'Aktionen der Voreinstellung';

  @override
  String get deletePreset => 'Voreinstellung löschen';

  @override
  String get saveToAllSubdecks => 'Für alle Unterstapel speichern';

  @override
  String usedByDecksShared(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stapeln',
      one: '1 Stapel',
    );
    return 'Verwendet von $_temp0. Änderungen gelten für alle.';
  }

  @override
  String usedByDecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stapeln',
      one: '1 Stapel',
    );
    return 'Verwendet von $_temp0.';
  }

  @override
  String get dailyLimits => 'Tageslimits';

  @override
  String get newCardsPerDay => 'Neue Karten pro Tag';

  @override
  String get maxReviewsPerDay => 'Maximale Wiederholungen pro Tag';

  @override
  String get thisDeckOnly => 'Nur dieser Stapel';

  @override
  String get thisDeckOnlyHelp =>
      'Limits für diesen Stapel, die die der Voreinstellung ersetzen. Leer lassen, um die der Voreinstellung zu verwenden.';

  @override
  String get enterWholeNumberShort => 'Gib eine ganze Zahl ein';

  @override
  String get newIgnoresReviewLimit =>
      'Neue Karten ignorieren das Wiederholungslimit';

  @override
  String get newIgnoresReviewLimitHelp =>
      'Wenn aus, werden neue Karten nur gezeigt, solange das Wiederholungslimit noch Platz hat.';

  @override
  String get limitsStartFromTop => 'Limits gelten von oben';

  @override
  String get limitsStartFromTopHelp =>
      'Beim Lernen dieses Stapels gelten auch die Limits seiner übergeordneten Stapel.';

  @override
  String get learningSteps => 'Lernschritte';

  @override
  String get learningStepsHelp =>
      'Verzögerungen, bevor eine neue Karte aufsteigt, z. B. 1m 10m 1d (s, m, h, d). Leer lassen, damit FSRS entscheidet.';

  @override
  String get relearningSteps => 'Neulernschritte';

  @override
  String get relearningStepsHelp =>
      'Verzögerungen, nachdem eine Wiederholungskarte vergessen wurde, z. B. 10m. Leer lassen, damit FSRS entscheidet.';

  @override
  String get insertionOrder => 'Einfügereihenfolge';

  @override
  String get insertionOrderHelp =>
      'Wo hinzugefügte Karten unter den neuen Karten landen. Eine Änderung ordnet die neuen Karten des Stapels neu.';

  @override
  String get insertSequential => 'Der Reihe nach (älteste Karten zuerst)';

  @override
  String get displayOrder => 'Anzeigereihenfolge';

  @override
  String get newGatherOrder => 'Sammelreihenfolge für neue Karten';

  @override
  String get newGatherOrderHelp =>
      'Welche neuen Karten das heutige Limit füllen.';

  @override
  String get gatherDeck => 'Stapel';

  @override
  String get gatherDeckRandomNotes => 'Stapel, dann zufällige Notizen';

  @override
  String get gatherAscending => 'Aufsteigende Position';

  @override
  String get gatherDescending => 'Absteigende Position';

  @override
  String get gatherRandomNotes => 'Zufällige Notizen';

  @override
  String get gatherRandomCards => 'Zufällige Karten';

  @override
  String get newSortOrder => 'Sortierreihenfolge für neue Karten';

  @override
  String get newSortOrderHelp =>
      'Die Reihenfolge, in der die heutigen neuen Karten gezeigt werden.';

  @override
  String get sortGathered => 'Reihenfolge des Sammelns';

  @override
  String get sortTypeThenGathered => 'Kartentyp, dann Reihenfolge des Sammelns';

  @override
  String get sortTypeThenRandom => 'Kartentyp, dann zufällig';

  @override
  String get sortRandomNoteThenType => 'Zufällige Notiz, dann Kartentyp';

  @override
  String get newReviewOrder => 'Reihenfolge neu/Wiederholung';

  @override
  String get mixWithReviews => 'Mit Wiederholungen mischen';

  @override
  String get showBeforeReviews => 'Vor Wiederholungen zeigen';

  @override
  String get showAfterReviews => 'Nach Wiederholungen zeigen';

  @override
  String get interdayOrder => 'Reihenfolge Lernen über Tage/Wiederholung';

  @override
  String get interdayOrderHelp =>
      'Wo Karten mit einem Schritt von einem Tag oder mehr gezeigt werden.';

  @override
  String get reviewSortOrder => 'Sortierreihenfolge für Wiederholungen';

  @override
  String get reviewSortOrderHelp =>
      'Entscheidet auch, welche Wiederholungen das Tageslimit füllen.';

  @override
  String get reviewDueRandom => 'Fälligkeitsdatum, dann zufällig';

  @override
  String get reviewDueDeck => 'Fälligkeitsdatum, dann Stapel';

  @override
  String get reviewDeckDue => 'Stapel, dann Fälligkeitsdatum';

  @override
  String get reviewIntervalsAscending => 'Aufsteigende Intervalle';

  @override
  String get reviewIntervalsDescending => 'Absteigende Intervalle';

  @override
  String get reviewDifficultyAscending => 'Aufsteigende Schwierigkeit';

  @override
  String get reviewDifficultyDescending => 'Absteigende Schwierigkeit';

  @override
  String get reviewRetrievabilityAscending => 'Aufsteigende Abrufbarkeit';

  @override
  String get reviewRelativeOverdueness => 'Relative Überfälligkeit';

  @override
  String get burying => 'Zurückstellen';

  @override
  String get buryingHelp =>
      'Sobald eine Karte einer Notiz gelernt wurde, warten ihre anderen Karten dieser Arten bis zum nächsten Tag.';

  @override
  String get buryNew => 'Neue Geschwisterkarten zurückstellen';

  @override
  String get buryReview => 'Geschwisterkarten in Wiederholung zurückstellen';

  @override
  String get buryInterday =>
      'Geschwisterkarten im Lernen über Tage zurückstellen';

  @override
  String get maxIntervalDays => 'Maximales Intervall (Tage)';

  @override
  String get maxIntervalHelp =>
      'Keine Wiederholung wird weiter in der Zukunft geplant.';

  @override
  String get minIntervalDays => 'Minimales Intervall (Tage)';

  @override
  String get minIntervalHelp =>
      'Das kürzeste Intervall, das eine vergessene Karte erhält, sobald sie neu gelernt ist.';

  @override
  String get leechThreshold => 'Schwelle für Problemkarten';

  @override
  String get leechThresholdHelp =>
      'Fehler, nach denen eine Karte eine Problemkarte ist (1 bis 99). Sie wird nach jeweils halb so vielen weiteren Fehlern erneut gekennzeichnet.';

  @override
  String get leechAction => 'Aktion für Problemkarten';

  @override
  String get leechTagOnly => 'Nur Schlagwort';

  @override
  String get leechSuspend => 'Karte aussetzen und Schlagwort';

  @override
  String get retentionHelp =>
      'Eine höhere Behaltensrate bedeutet, dass du weniger vergisst, aber öfter wiederholst.';

  @override
  String get disableAutoplay => 'Audio nicht automatisch abspielen';

  @override
  String get skipQuestionOnReplay =>
      'Frage beim erneuten Abspielen der Antwort überspringen';

  @override
  String get timer => 'Timer';

  @override
  String get maxAnswerSeconds => 'Maximale Antwortzeit in Sekunden';

  @override
  String get maxAnswerSecondsHelp =>
      'Längere Antworten werden mit so vielen Sekunden erfasst, damit eine Pause nicht als Antwortzeit zählt.';

  @override
  String get showTimer => 'Antworttimer anzeigen';

  @override
  String get showTimerHelp =>
      'Zeigt beim Lernen die Zeit, die auf der Karte verbracht wurde. Das ändert nicht, was erfasst wird.';

  @override
  String get stopTimerOnAnswer => 'Timer bei der Antwort anhalten';

  @override
  String get stopTimerOnAnswerHelp =>
      'Hält den Timer an, wenn die Antwort gezeigt wird.';

  @override
  String get autoAdvance => 'Automatisch weiterblättern';

  @override
  String get autoAdvanceHelp =>
      'Schalte das automatische Weiterblättern beim Lernen über das Menü oder mit Umschalt+A ein. 0 Sekunden schaltet es für diese Seite aus.';

  @override
  String get secondsToShowQuestion => 'Sekunden, die die Frage gezeigt wird';

  @override
  String get secondsToShowAnswer => 'Sekunden, die die Antwort gezeigt wird';

  @override
  String get questionAction => 'Aktion für die Frage';

  @override
  String get showAnswer => 'Antwort zeigen';

  @override
  String get showReminder => 'Erinnerung zeigen';

  @override
  String get answerAction => 'Aktion für die Antwort';

  @override
  String get buryCard => 'Karte zurückstellen';

  @override
  String get answerAgain => 'Mit Nochmal beantworten';

  @override
  String get answerHard => 'Mit Schwer beantworten';

  @override
  String get answerGood => 'Mit Gut beantworten';

  @override
  String get fsrs => 'FSRS';

  @override
  String get fsrsParameters => 'FSRS-Parameter';

  @override
  String get fsrsParametersHelp =>
      'Werden durch Optimieren an deine Wiederholungen angepasst. Leer lassen für die Standardparameter.';

  @override
  String get optimizeSearch => 'Suche für die Optimierung';

  @override
  String get optimizeSearchHelp =>
      'Die Karten, aus deren Wiederholungen das Optimieren lernt. Leer lassen für diesen Stapel und seine Unterstapel.';

  @override
  String get notSet => 'Nicht gesetzt';

  @override
  String get clearDate => 'Datum löschen';

  @override
  String historicalRetentionValue(String retention) {
    return 'Bisherige Behaltensrate: $retention';
  }

  @override
  String get historicalRetentionHelp =>
      'Wie viel du bei Wiederholungen behalten hast, die im Verlauf fehlen, z. B. von Karten, die vor dem Import gelernt wurden.';

  @override
  String get healthCheck => 'Zuverlässigkeitsprüfung';

  @override
  String get healthCheckHelp =>
      'Prüft vor der Verwendung, ob die neuen Parameter zuverlässig sind. Dauert länger.';

  @override
  String get rescheduleOnChange => 'Karten bei Änderungen neu planen';

  @override
  String get rescheduleOnChangeHelp =>
      'Wenn sich die gewünschte Behaltensrate oder die Parameter ändern, erhalten Wiederholungskarten neue Fälligkeitsdaten. Sonst verwenden nur künftige Antworten sie.';

  @override
  String get optimizeAllPresets => 'Alle Voreinstellungen optimieren';

  @override
  String get easyDays => 'Leichte Tage';

  @override
  String get easyDaysHelp =>
      'Tage mit weniger Wiederholungen. Wiederholungen verschieben sich auf Tage in der Nähe, höchstens um die wenigen Tage, um die Intervalle ohnehin schwanken; ein Tag mit „Minimum“ erhält nur Karten, die nirgendwo anders hinkönnen.';

  @override
  String notANumber(String text) {
    return '„$text“ ist keine Zahl.';
  }

  @override
  String fsrsCountError(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Gib $countString Zahlen ein oder lass das Feld leer für die Standardparameter.';
  }

  @override
  String get starting => 'Starte…';

  @override
  String get readingReviewHistory => 'Wiederholungsverlauf wird gelesen…';

  @override
  String optimizingPercent(String percent) {
    return 'Optimiere… $percent';
  }

  @override
  String checkingParametersPercent(String percent) {
    return 'Parameter werden geprüft… $percent';
  }

  @override
  String get optimizing => 'Optimiere…';

  @override
  String countReviews(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
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
      other: 'gibt nur $countString Wiederholungen',
      one: 'gibt nur 1 Wiederholung',
    );
    return 'Das Optimieren braucht mindestens $min Wiederholungen, aber es $_temp0, aus denen gelernt werden kann. Die Parameter bleiben unverändert.';
  }

  @override
  String newParametersFitted(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
    );
    return 'Neue Parameter an $_temp0 angepasst. Speichere, um sie zu verwenden.';
  }

  @override
  String get healthCheckFailed =>
      'Die Zuverlässigkeitsprüfung hält die neuen Parameter für unzuverlässig: Es gibt vielleicht zu wenige oder zu uneinheitliche Wiederholungen. Die Parameter bleiben unverändert.';

  @override
  String parametersAlreadyBest(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
    );
    return 'Die aktuellen Parameter passen bereits am besten zu deinen $_temp0. Sie bleiben unverändert.';
  }

  @override
  String get easyNormal => 'Normal';

  @override
  String get easyReduced => 'Reduziert';

  @override
  String get easyMinimum => 'Minimum';

  @override
  String notADelay(String text) {
    return '„$text“ ist keine Verzögerung. Schreibe eine Zahl und eine Einheit (s, m, h oder d), z. B. 10m.';
  }

  @override
  String get stepTooShort => 'Jeder Schritt muss länger als 0 sein.';

  @override
  String get stepTooLong => 'Ein Schritt kann höchstens 365 Tage lang sein.';

  @override
  String get tooManySteps => 'Es sind höchstens 20 Schritte möglich.';

  @override
  String get leechSuspended =>
      'Die Karte ist eine Problemkarte und wurde ausgesetzt.';

  @override
  String get leech => 'Die Karte ist eine Problemkarte.';

  @override
  String get autoAdvanceOff =>
      'Automatisches Weiterblättern ist an, aber der Stapel dieser Karte zeigt beide Seiten 0 Sekunden lang, was es ausschaltet. Stelle die Sekunden in den Stapeloptionen ein.';

  @override
  String get timeToShowAnswer => 'Zeit, die Antwort zu zeigen.';

  @override
  String get timeToAnswer => 'Zeit zu antworten.';

  @override
  String get recordingUnavailable => 'Aufnehmen ist nicht verfügbar';

  @override
  String recordingFailed(String error) {
    return 'Deine Stimme konnte nicht aufgenommen werden: $error';
  }

  @override
  String get timeboxReached => 'Zeitlimit erreicht';

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
      other: '$cardsString Karten',
      one: '1 Karte',
    );
    String _temp1 = intl.Intl.pluralLogic(
      minutes,
      locale: localeName,
      other: '$minutesString Minuten',
      one: '1 Minute',
    );
    return 'Du hast $_temp0 in $_temp1 gelernt.';
  }

  @override
  String get finish => 'Beenden';

  @override
  String get continueAction => 'Weiter';

  @override
  String get noCardAnsweredYet => 'Es wurde noch keine Karte beantwortet.';

  @override
  String get removeFlag => 'Flagge entfernen';

  @override
  String get stopRecording => 'Aufnahme beenden';

  @override
  String get recordVoice => 'Eigene Stimme aufnehmen';

  @override
  String get cardActions => 'Kartenaktionen';

  @override
  String get createCopy => 'Kopie erstellen';

  @override
  String get replayAudio => 'Audio erneut abspielen';

  @override
  String get playVoice => 'Eigene Stimme abspielen';

  @override
  String get flag => 'Flagge';

  @override
  String get unmarkNote => 'Markierung der Notiz entfernen';

  @override
  String get markNote => 'Notiz markieren';

  @override
  String get previousCardInfo => 'Info zur vorherigen Karte';

  @override
  String get resetCard => 'Karte zurücksetzen';

  @override
  String get buryCardUntilTomorrow => 'Karte bis morgen zurückstellen';

  @override
  String get buryNoteUntilTomorrow => 'Notiz bis morgen zurückstellen';

  @override
  String get suspendCard => 'Karte aussetzen';

  @override
  String get suspendNote => 'Notiz aussetzen';

  @override
  String get deleteNote => 'Notiz löschen';

  @override
  String get stopAutoAdvance => 'Automatisches Weiterblättern beenden';

  @override
  String get deckOptions => 'Stapeloptionen';

  @override
  String get deckCouldNotOpen => 'Dieser Stapel konnte nicht geöffnet werden';

  @override
  String get autoAdvanceIsOn => 'Automatisches Weiterblättern ist an. Beenden';

  @override
  String get showAnswerButton => 'Antwort zeigen';

  @override
  String get previewEnd => 'Ende';

  @override
  String leaveDeck(String deck) {
    return '$deck verlassen';
  }

  @override
  String get leave => 'Verlassen';

  @override
  String flagNamed(String name) {
    return 'Flagge: $name';
  }

  @override
  String get marked => 'Markiert';

  @override
  String get timeOnCard => 'Zeit auf dieser Karte';

  @override
  String get emptyCardNotice =>
      'Diese Karte ist leer: Ihre Frage zeigt nichts mehr an. Entferne leere Karten unter Einstellungen › Leere Karten.';

  @override
  String shownAgainEarly(String time) {
    return 'Diese Karte noch einmal: fällig in $time, früher gezeigt, weil sonst nichts fällig ist.';
  }

  @override
  String shownEarly(String time) {
    return 'Früher gezeigt: fällig in $time, weil sonst nichts fällig ist.';
  }

  @override
  String get shownAgainDueNow =>
      'Diese Karte noch einmal: Sie ist jetzt wieder fällig.';

  @override
  String get learningDueNow => 'Lernkarten sind jetzt fällig';

  @override
  String get backToDecks => 'Zurück zu den Stapeln';

  @override
  String get congratulationsNothingDue =>
      'Glückwunsch! Gerade ist nichts fällig.';

  @override
  String studiedInTime(int count, String time) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Du hast $_temp0 in $time gelernt.';
  }

  @override
  String get sessionComplete => 'Sitzung abgeschlossen';

  @override
  String streakDaysLabel(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tage',
      one: '1 Tag',
    );
    return '$_temp0';
  }

  @override
  String streakAtRisk(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tagen',
      one: '1 Tag',
    );
    return 'Lerne heute, um deine Serie von $_temp0 zu halten';
  }

  @override
  String get streakHeaderTooltip => 'Serien und Tagesziel';

  @override
  String get levelHeaderTooltip => 'Level und Erfolge';

  @override
  String levelHeaderSemantics(int level, int done, int need, int next) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);
    final intl.NumberFormat needNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String needString = needNumberFormat.format(need);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return 'Level $levelString, $doneString von $needString XP bis Level $nextString';
  }

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
      other: '$goalString Karten',
      one: '1 Karte',
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
      other: '$goalString Minuten',
      one: '1 Minute',
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
      other: '$countString Tage',
      one: '1 Tag',
    );
    return 'Serie auf $_temp0 verlängert';
  }

  @override
  String streakCurrent(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tage',
      one: '1 Tag',
    );
    return 'Aktuelle Serie: $_temp0';
  }

  @override
  String celebrationStreak(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tage in Folge',
      one: '1 Tag in Folge',
    );
    return '$_temp0';
  }

  @override
  String get celebrationBest => 'Neuer persönlicher Rekord';

  @override
  String get celebrationNewAchievement => 'Neuer Erfolg';

  @override
  String get celebrationNewAchievements => 'Neue Erfolge';

  @override
  String celebrationTier(String name, String tier) {
    return '$name · $tier';
  }

  @override
  String xpGained(int xp) {
    final intl.NumberFormat xpNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String xpString = xpNumberFormat.format(xp);

    return '+$xpString XP';
  }

  @override
  String xpToLevel(int xp, int next) {
    final intl.NumberFormat xpNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String xpString = xpNumberFormat.format(xp);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return '$xpString XP bis Level $nextString';
  }

  @override
  String xpThisSession(int xp) {
    final intl.NumberFormat xpNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String xpString = xpNumberFormat.format(xp);

    return '+$xpString XP in dieser Sitzung';
  }

  @override
  String get goalReached => 'Tagesziel erreicht';

  @override
  String levelBanner(int level) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);

    return 'Level $levelString';
  }

  @override
  String levelReached(int level) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);

    return 'Level $levelString erreicht';
  }

  @override
  String levelXpLeft(int level, int xp, int next) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);
    final intl.NumberFormat xpNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String xpString = xpNumberFormat.format(xp);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return 'Level $levelString · $xpString XP bis Level $nextString';
  }

  @override
  String goalCardsLeft(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Noch $_temp0 bis zum heutigen Ziel';
  }

  @override
  String goalMinutesLeft(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Minuten',
      one: '1 Minute',
    );
    return 'Noch $_temp0 bis zum heutigen Ziel';
  }

  @override
  String get goalSuggestion =>
      'Ein Tagesziel zeigt dir jeden Tag, wie weit du bist.';

  @override
  String get goalSuggestionSet => 'Tagesziel festlegen';

  @override
  String get goalSuggestionDismiss => 'Nein danke';

  @override
  String moreLearningDue(String when) {
    return 'Weitere Lernkarten sind $when fällig.';
  }

  @override
  String get checkAgain => 'Erneut prüfen';

  @override
  String get browserMayClear =>
      'Dein Browser löscht Cardflip-Daten möglicherweise';

  @override
  String get browserMayClearBody =>
      'Dieser Browser hat keinen dauerhaften Speicher gewährt und löscht deine Sammlung daher möglicherweise, wenn der Platz knapp wird. Exportiere regelmäßig eine Sicherung in den Einstellungen, um deine Karten zu schützen.';

  @override
  String couldNotStart(String error) {
    return 'Cardflip konnte nicht starten: $error';
  }

  @override
  String get collectionUnreadableShort =>
      'Deine Sammlung konnte nicht gelesen werden.';

  @override
  String get startNewCollectionTitle => 'Neue Sammlung beginnen?';

  @override
  String get startNewCollectionBody =>
      'Dadurch wird die gespeicherte Sammlung, die nicht gelesen werden konnte, endgültig gelöscht. Stelle stattdessen eine Sicherung wieder her, falls du eine hast.';

  @override
  String get deleteAndStartOver => 'Löschen und neu beginnen';

  @override
  String get collectionCouldNotOpen =>
      'Deine Sammlung konnte nicht geöffnet werden';

  @override
  String get nothingDeleted => 'Es wurde nichts gelöscht.';

  @override
  String get restoreFromBackup => 'Aus Sicherung wiederherstellen';

  @override
  String get startNewCollection => 'Neue Sammlung beginnen';

  @override
  String get navDecks => 'Stapel';

  @override
  String get navBrowse => 'Browser';

  @override
  String get navAdd => 'Hinzufügen';

  @override
  String get navSettings => 'Einstellungen';

  @override
  String learningDueInDeck(String deck) {
    return 'In $deck sind Lernkarten fällig';
  }

  @override
  String learningDueInDecks(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'In $countString Stapeln sind Lernkarten fällig';
  }

  @override
  String get showDecks => 'Stapel zeigen';

  @override
  String undoStep(String step) {
    return '$step rückgängig machen';
  }

  @override
  String redoStep(String step) {
    return '$step wiederholen';
  }

  @override
  String get redo => 'Wiederholen';

  @override
  String get more => 'Mehr';

  @override
  String undone(String step) {
    return 'Rückgängig gemacht: $step';
  }

  @override
  String redone(String step) {
    return 'Wiederholt: $step';
  }

  @override
  String get keyCtrl => 'Strg';

  @override
  String get keyShift => 'Umschalt';

  @override
  String get keyAlt => 'Alt';

  @override
  String get keyEnter => 'Eingabe';

  @override
  String get keyDelete => 'Entf';

  @override
  String get keyEsc => 'Esc';

  @override
  String labelWithShortcut(String label, String shortcut) {
    return '$label ($shortcut)';
  }

  @override
  String get flagRed => 'Rot';

  @override
  String get flagOrange => 'Orange';

  @override
  String get flagGreen => 'Grün';

  @override
  String get flagBlue => 'Blau';

  @override
  String get flagPink => 'Pink';

  @override
  String get flagTurquoise => 'Türkis';

  @override
  String get flagPurple => 'Lila';

  @override
  String backupSavedAs(String file) {
    return 'Sicherung gespeichert als $file';
  }

  @override
  String get replaceCollectionTitle => 'Deine Sammlung ersetzen?';

  @override
  String restoreBackupBody(String date) {
    return 'Das Wiederherstellen der Sicherung vom $date ersetzt deine aktuelle Sammlung vollständig: Alle Stapel, Notizen, Karten, der Wiederholungsverlauf und die Einstellungen werden überschrieben. Das kann nicht rückgängig gemacht werden.';
  }

  @override
  String get restore => 'Wiederherstellen';

  @override
  String get backupRestored => 'Sicherung wiederhergestellt';

  @override
  String get deletionLogEmpty => 'Das Löschprotokoll ist leer';

  @override
  String deletionLogSaved(int count, String file) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Löschprotokoll mit $_temp0 gespeichert als $file';
  }

  @override
  String get clearDeletionLogTitle => 'Löschprotokoll leeren?';

  @override
  String get clearDeletionLogBody =>
      'Das Löschprotokoll bewahrt jede gelöschte Notiz auf, damit sie exportiert und wieder importiert werden kann. Beim Leeren werden diese Einträge endgültig entfernt. Das kann nicht rückgängig gemacht werden.';

  @override
  String get clear => 'Leeren';

  @override
  String get deletionLogCleared => 'Löschprotokoll geleert';

  @override
  String renameFlag(String flag) {
    return 'Flagge „$flag“ umbenennen';
  }

  @override
  String renameFlagHelper(String flag) {
    return 'Leer lassen für „$flag“';
  }

  @override
  String get defaultBrowserSearch => 'Standardsuche im Browser';

  @override
  String get defaultSearchHint => 'z. B. deck:Spanisch is:due';

  @override
  String get defaultSearchHelper => 'Leer lassen, um alle Karten aufzulisten';

  @override
  String get off => 'Aus';

  @override
  String minutesShort(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString Min.';
  }

  @override
  String countHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stunden',
      one: '1 Stunde',
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
  String get settings => 'Einstellungen';

  @override
  String get appearance => 'Darstellung';

  @override
  String get theme => 'Design';

  @override
  String get themeSystem => 'System';

  @override
  String get themeLight => 'Hell';

  @override
  String get themeDark => 'Dunkel';

  @override
  String get language => 'Sprache';

  @override
  String get languageSystem => 'Systemstandard';

  @override
  String get languageName => 'Deutsch';

  @override
  String get uiSize => 'Oberflächengröße';

  @override
  String get uiSizeHelp => 'Skaliert Text, Symbole und Abstände.';

  @override
  String get minimalistMode => 'Minimalistischer Modus';

  @override
  String get minimalistModeHelp =>
      'Flache Flächen ohne Schatten oder getönte Hintergründe.';

  @override
  String get reduceMotion => 'Bewegung reduzieren';

  @override
  String get reduceMotionHelp =>
      'Seiten, Dialoge und Menüs sofort zeigen, ohne Animationen.';

  @override
  String get reduceMotionSystem => 'Wie im System';

  @override
  String get on => 'An';

  @override
  String get studying => 'Lernen';

  @override
  String get nextDayStarts => 'Nächster Tag beginnt um';

  @override
  String get nextDayStartsHelp =>
      'Tageslimits und fällige Karten werden zu dieser Stunde zurückgesetzt.';

  @override
  String get learnAheadLimit => 'Vorauslernlimit';

  @override
  String get learnAheadLimitHelp =>
      'Wenn sonst nichts fällig ist, Lernkarten zeigen, die innerhalb dieser Zeit fällig werden, statt zu warten.';

  @override
  String get timeboxLimit => 'Zeitlimit der Lerneinheit';

  @override
  String get timeboxLimitHelp =>
      'Jedes Mal, wenn beim Lernen so viel Zeit vergangen ist, zeigen, wie viele Karten du gelernt hast.';

  @override
  String get autoContinueLearning =>
      'Automatisch fortfahren, wenn Lernkarten fällig sind';

  @override
  String get autoContinueLearningHelp =>
      'Sonst fragt eine beendete Sitzung, bevor sie gezeigt werden.';

  @override
  String get showPlayButtons => 'Abspielknöpfe auf Karten mit Audio zeigen';

  @override
  String get interruptAudio => 'Laufendes Audio beim Antworten unterbrechen';

  @override
  String get interruptAudioHelp =>
      'Sonst wird das Audio der Antwort nach dem der Frage abgespielt.';

  @override
  String get stripPasteFormatting => 'Formatierung beim Einfügen entfernen';

  @override
  String stripPasteFormattingHelp(String key) {
    return 'Fügt reinen Text ein; $key+Umschalt+V fügt mit Formatierung ein.';
  }

  @override
  String get pasteImagesAsPng => 'Bilder als PNG einfügen';

  @override
  String get pasteImagesAsPngHelp =>
      'Verlustfrei, aber größer; sonst werden eingefügte Bilder als JPEG gespeichert.';

  @override
  String get deckForNewCards => 'Stapel für neue Karten';

  @override
  String get currentDeck => 'Aktueller Stapel';

  @override
  String get deckByNoteType => 'Stapel je nach Notiztyp wechseln';

  @override
  String get goalsAndStreaks => 'Ziele und Serien';

  @override
  String get showStreaksAndGoals => 'Serien und Ziele anzeigen';

  @override
  String get showStreaksAndGoalsHelp =>
      'In der Stapelliste und nach dem Lernen. Nur auf diesem Gerät.';

  @override
  String get showStudyRewards => 'Belohnungen beim Lernen anzeigen';

  @override
  String get showStudyRewardsHelp =>
      'Ein kurzes Banner, wenn du dein Ziel, ein Level oder einen Erfolg erreichst. Nur auf diesem Gerät.';

  @override
  String get celebrationSounds => 'Töne und Vibration für Erfolge';

  @override
  String get celebrationSoundsHelp =>
      'Ein kurzer Ton und auf Handys eine Vibration bei Belohnungen und erreichten Erfolgen. Nur auf diesem Gerät.';

  @override
  String get showLevelUpsWhileStudying => 'Neue Level beim Lernen anzeigen';

  @override
  String get showLevelUpsWhileStudyingHelp =>
      'Eine ganze Seite, sobald du ein neues Level erreichst. Aus: ein Banner nennt es, und die Seite kommt am Ende der Sitzung. Nur auf diesem Gerät.';

  @override
  String get deckProgressInList =>
      'Stapelfortschritt in der Stapelliste anzeigen';

  @override
  String get deckProgressInListHelp =>
      'Ein schmaler Balken aus neuen, lernenden, jungen und ausgereiften Karten unter jedem Stapel. Nur auf diesem Gerät.';

  @override
  String get deckProgress => 'Fortschritt';

  @override
  String deckProgressMature(String percent) {
    return '$percent ausgereift';
  }

  @override
  String deckProgressLegend(String state, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$state: $countString';
  }

  @override
  String get recapTitle => 'Letzte Woche';

  @override
  String get recapTooltip =>
      'Letzte Woche im Vergleich zur Woche davor. Öffnet die Statistik.';

  @override
  String get recapDismiss => 'Bis nächste Woche ausblenden';

  @override
  String recapStudyDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Lerntage',
      one: '1 Lerntag',
    );
    return '$_temp0';
  }

  @override
  String recapAnswers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Antworten',
      one: '1 Antwort',
    );
    return '$_temp0';
  }

  @override
  String recapMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Minuten',
      one: '1 Minute',
    );
    return '$_temp0';
  }

  @override
  String recapRetention(String percent) {
    return '$percent Behaltensrate';
  }

  @override
  String recapWeekBefore(String value) {
    return 'Woche davor: $value';
  }

  @override
  String recapPoints(String change) {
    return '$change Pkt.';
  }

  @override
  String get recapSame => 'gleich';

  @override
  String get dailyGoal => 'Tagesziel';

  @override
  String get dailyGoalNone => 'Keins';

  @override
  String get dailyGoalCards => 'Karten';

  @override
  String get dailyGoalMinutes => 'Minuten';

  @override
  String dailyGoalCardsPerDay(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 pro Tag';
  }

  @override
  String dailyGoalMinutesPerDay(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Minuten',
      one: '1 Minute',
    );
    return '$_temp0 pro Tag';
  }

  @override
  String get dailyGoalCardsLabel => 'Karten pro Tag';

  @override
  String get dailyGoalMinutesLabel => 'Minuten pro Tag';

  @override
  String get dailyGoalError => 'Gib eine ganze Zahl von 1 bis 9999 ein.';

  @override
  String get dailyGoalHelp =>
      'Jeder Lerntag hält deine Serie, ob du das Ziel erreichst oder nicht.';

  @override
  String get restDaysPerWeek => 'Ruhetage pro Woche';

  @override
  String get restDaysPerWeekHelp =>
      'Tage ohne Lernen, die eine Serie in beliebigen 7 Tagen nicht unterbrechen.';

  @override
  String get achievements => 'Erfolge';

  @override
  String levelTitle(int level) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);

    return 'Level $levelString';
  }

  @override
  String levelXpProgress(int done, int need, int next) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);
    final intl.NumberFormat needNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String needString = needNumberFormat.format(need);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return '$doneString / $needString XP bis Level $nextString';
  }

  @override
  String xpAmount(int xp) {
    final intl.NumberFormat xpNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String xpString = xpNumberFormat.format(xp);

    return '$xpString XP';
  }

  @override
  String get xpFromAnswers => 'Antworten';

  @override
  String get xpFromStudyDays => 'Lerntage';

  @override
  String get xpFromGoalDays => 'Tagesziel erreicht';

  @override
  String get xpFromLongTerm => 'Langzeitgedächtnis';

  @override
  String get xpFromAchievements => 'Erfolge';

  @override
  String get xpExplanation =>
      'Jede Antwort bringt 1 XP, jeder Lerntag 10, jeder Tag mit erreichtem Tagesziel 20, jede Karte im Langzeitgedächtnis 5 und jede Erfolgsstufe 25 bis 200.';

  @override
  String get xpNoGoalHint =>
      'Setze ein Tagesziel, um an jedem Tag, an dem du es erreichst, 20 XP zu bekommen.';

  @override
  String get achievementAnswers => 'Antworten';

  @override
  String get achievementStreak => 'Serie';

  @override
  String get achievementStudyDays => 'Lerntage';

  @override
  String get achievementStudyTime => 'Lernzeit';

  @override
  String get achievementDailyGoal => 'Tagesziel';

  @override
  String get achievementLongTerm => 'Langzeitgedächtnis';

  @override
  String get achievementLevel => 'Level';

  @override
  String get achievementComeback => 'Comeback';

  @override
  String get tierBronze => 'Bronze';

  @override
  String get tierSilver => 'Silber';

  @override
  String get tierGold => 'Gold';

  @override
  String get tierPlatinum => 'Platin';

  @override
  String tierAnswers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Antworten',
      one: '1 Antwort',
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
      other: '$countString Tage',
      one: '1 Tag',
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
      other: '$countString Tagen',
      one: '1 Tag',
    );
    return 'an $_temp0 erreicht';
  }

  @override
  String tierCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0';
  }

  @override
  String tierHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stunden',
      one: '1 Stunde',
    );
    return '$_temp0';
  }

  @override
  String tierLevel(int level) {
    final intl.NumberFormat levelNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String levelString = levelNumberFormat.format(level);

    return 'Level $levelString';
  }

  @override
  String get tierOneTime => 'Einmalig';

  @override
  String tierWithThreshold(String tier, String threshold) {
    return '$tier · $threshold';
  }

  @override
  String achievementEarnedOn(String date) {
    return 'Erreicht am $date';
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
  String achievementProgressHours(int progress, int next) {
    final intl.NumberFormat progressNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String progressString = progressNumberFormat.format(progress);
    final intl.NumberFormat nextNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String nextString = nextNumberFormat.format(next);

    return '$progressString / $nextString Stunden';
  }

  @override
  String get achievementAllEarned => 'Alle Stufen erreicht';

  @override
  String get achievementNeedsGoal => 'Lege ein Tagesziel fest, um zu beginnen';

  @override
  String get achievementComebackHint =>
      'Lerne wieder nach 14 oder mehr Tagen ohne Lernen';

  @override
  String achievementLocked(String name) {
    return '$name, noch nicht erreicht';
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
      other: '$countString Antworten',
      one: '1 Antwort',
    );
    return '$_temp0';
  }

  @override
  String newStreak(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Serie von $countString Tagen';
  }

  @override
  String newStudyDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Lerntage',
      one: '1 Lerntag',
    );
    return '$_temp0';
  }

  @override
  String newStudyTime(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Stunden',
      one: '1 Stunde',
    );
    return '$_temp0 Lernzeit';
  }

  @override
  String newDailyGoal(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tagen',
      one: '1 Tag',
    );
    return 'Tagesziel an $_temp0 erreicht';
  }

  @override
  String newLongTerm(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 im Langzeitgedächtnis';
  }

  @override
  String newAchievementsLine(String names) {
    return 'Neu: $names';
  }

  @override
  String newAchievementsCountLine(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Neu: $countString Erfolge';
  }

  @override
  String newAchievement(String name) {
    return 'Neuer Erfolg: $name';
  }

  @override
  String newAchievements(String names) {
    return 'Neue Erfolge: $names';
  }

  @override
  String achievementsFromHistory(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString Erfolge aus deinem Verlauf';
  }

  @override
  String get showAchievements => 'Ansehen';

  @override
  String get review => 'Abfrage';

  @override
  String get showRemainingCounts => 'Anzahl verbleibender Karten zeigen';

  @override
  String get showIntervals =>
      'Nächste Wiederholung über den Antwortknöpfen zeigen';

  @override
  String get spaceAnswersGood => 'Leertaste/Eingabe antwortet auch mit Gut';

  @override
  String get spaceAnswersGoodHelp => 'Sonst decken sie nur die Antwort auf.';

  @override
  String get hideBarsDuringReview => 'Leisten beim Lernen ausblenden';

  @override
  String get hideBarsDuringReviewHelp =>
      'Blendet die App-Leiste und die Navigation beim Lernen aus; Escape oder der obere Rand beenden das.';

  @override
  String get questionFade => 'Einblenden der Frage';

  @override
  String get questionFadeHelp =>
      'Wie lange jede neue Frage zum Einblenden braucht.';

  @override
  String get browsing => 'Browser';

  @override
  String get ignoreAccents => 'Akzente bei der Suche ignorieren';

  @override
  String get ignoreAccentsHelp => 'Zum Beispiel findet „cafe“ auch „café“.';

  @override
  String get allCards => 'Alle Karten';

  @override
  String get collection => 'Sammlung';

  @override
  String get import => 'Importieren';

  @override
  String get importHelp =>
      'Text (.csv, .tsv, .txt), Anki-Paket (.apkg) oder -Sammlung (.colpkg)';

  @override
  String get exportBackup => 'Sicherung exportieren';

  @override
  String get downloadsBackupFile => 'Lädt eine .cardflip-Datei herunter';

  @override
  String get savesBackupFile => 'Speichert eine .cardflip-Datei';

  @override
  String get exportCollection => 'Sammlung exportieren';

  @override
  String get exportCollectionHelp => 'Anki-Sammlungspaket oder reiner Text';

  @override
  String get restoreBackup => 'Sicherung wiederherstellen';

  @override
  String get restoreBackupHelp => 'Ersetzt die aktuelle Sammlung';

  @override
  String get emptyCards => 'Leere Karten';

  @override
  String get emptyCardsHelp => 'Karten mit leerer Frage finden und löschen';

  @override
  String get checkMedia => 'Medien prüfen';

  @override
  String get checkMediaHelp =>
      'Ungenutzte und fehlende Mediendateien finden, gelöschte wiederherstellen';

  @override
  String get checkDatabase => 'Datenbank prüfen';

  @override
  String get checkDatabaseHelp =>
      'Probleme in der Sammlung finden und reparieren';

  @override
  String get exportDeletionLog => 'Löschprotokoll exportieren';

  @override
  String get exportDeletionLogHelp =>
      'Gelöschte Notizen als Textdatei, die der Import lesen kann';

  @override
  String get clearDeletionLog => 'Löschprotokoll leeren';

  @override
  String get clearDeletionLogHelp => 'Vergisst die bisher gelöschten Notizen';

  @override
  String get backups => 'Sicherungen';

  @override
  String get flags => 'Flaggen';

  @override
  String get agentsMcp => 'Agenten (MCP)';

  @override
  String get allowAgents => 'Agenten erlauben (MCP)';

  @override
  String get allowAgentsHelp =>
      'KI-Agenten wie Claude Code können die geöffnete Sammlung lesen und ändern, solange Cardflip läuft.';

  @override
  String get allowAgentsTitle => 'Agenten erlauben?';

  @override
  String get allowAgentsBody =>
      'Jedes Programm auf diesem Computer, das den Token kennt, kann deine Sammlung lesen und ändern, solange Cardflip geöffnet ist. Gib den Token nur Agenten, denen du vertraust.';

  @override
  String get allow => 'Erlauben';

  @override
  String get mcpServerRunning => 'Server läuft';

  @override
  String get mcpServerStopped => 'Server läuft nicht';

  @override
  String get mcpPort => 'Port';

  @override
  String get mcpPortHelper => '1024 bis 65535';

  @override
  String get mcpToken => 'Token';

  @override
  String get copyToken => 'Token kopieren';

  @override
  String get tokenCopied => 'Token kopiert';

  @override
  String get regenerateToken => 'Token neu erzeugen';

  @override
  String get regenerateTokenTitle => 'Token neu erzeugen?';

  @override
  String get regenerateTokenBody =>
      'Agenten mit dem bisherigen Token können sich erst wieder verbinden, wenn du ihnen den neuen gibst.';

  @override
  String get regenerate => 'Neu erzeugen';

  @override
  String get copyClientConfig => 'Client-Konfiguration kopieren';

  @override
  String get copyClientConfigHelp =>
      'Server-URL und Token für Claude Code, Claude Desktop und andere MCP-Clients';

  @override
  String get clientConfigCopied => 'Client-Konfiguration kopiert';

  @override
  String get agentActivity => 'Letzte Änderungen durch Agenten';

  @override
  String get noAgentActivity => 'Keine Änderungen in dieser Sitzung';

  @override
  String get storage => 'Speicher';

  @override
  String get storagePersistent =>
      'Dauerhaft: Der Browser behält deine Sammlung.';

  @override
  String get storageNotPersistent =>
      'Der Browser löscht deine Sammlung möglicherweise, wenn der Platz knapp wird. Exportiere regelmäßig Sicherungen.';

  @override
  String get about => 'Über';

  @override
  String aboutCardflip(String version) {
    return 'Version $version\nKarteikarten mit verteilter Wiederholung. Deine Daten bleiben auf diesem Gerät.\nCopyright © 2026 madzione. Freie Software unter der GNU GPL Version 3 oder neuer, ohne jede Gewährleistung.';
  }

  @override
  String get openSourceLicenses => 'Open-Source-Lizenzen';

  @override
  String get sourceCode => 'Quellcode';

  @override
  String get licenseLegalese =>
      'Copyright © 2026 madzione\n\nCardflip ist freie Software: Du kannst sie unter den Bedingungen der GNU General Public License, Version 3 oder neuer, weitergeben und verändern. Sie wird ohne jede Gewährleistung bereitgestellt.';

  @override
  String backupEveryMinutes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Alle $countString Min.';
  }

  @override
  String get backupEveryHour => 'Jede Stunde';

  @override
  String backupEveryHours(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Alle $countString Stunden';
  }

  @override
  String restoreAutoBackupBody(String date) {
    return 'Das Wiederherstellen der automatischen Sicherung vom $date ersetzt deine aktuelle Sammlung vollständig: Alle Stapel, Notizen, Karten, der Wiederholungsverlauf und die Einstellungen werden überschrieben. Mediendateien bleiben unverändert. Deine aktuelle Sammlung wird vorher gesichert, damit du zu ihr zurückkehren kannst.';
  }

  @override
  String get backUpAutomatically => 'Automatisch sichern';

  @override
  String get backUpAutomaticallyHelp =>
      'Während sich die Sammlung ändert, ohne Medien. Vor einer Wiederherstellung wird immer eine Sicherung erstellt.';

  @override
  String get keepBackups => 'Neben den heutigen Sicherungen aufbewahren';

  @override
  String get keepDaily => 'Täglich';

  @override
  String get keepWeekly => 'Wöchentlich';

  @override
  String get keepMonthly => 'Monatlich';

  @override
  String get createBackupNow => 'Jetzt sichern';

  @override
  String get noAutoBackups => 'Noch keine automatischen Sicherungen';

  @override
  String countAutoBackups(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString automatische Sicherungen',
      one: '1 automatische Sicherung',
    );
    return '$_temp0';
  }

  @override
  String usedOnDevice(String size) {
    return '$size auf diesem Gerät belegt';
  }

  @override
  String get showFewer => 'Weniger zeigen';

  @override
  String showAllCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Alle $countString zeigen';
  }

  @override
  String deletedEmptyCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString leere Karten',
      one: '1 leere Karte',
    );
    return '$_temp0 gelöscht';
  }

  @override
  String get close => 'Schließen';

  @override
  String get noEmptyCards => 'Es gibt keine leeren Karten.';

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
      other: '$cardsString leere Karten',
      one: '1 leere Karte',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 in $_temp1. Ihre Frage zeigt nichts mehr an.';
  }

  @override
  String emptyCardsKept(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen haben',
      one: '1 Notiz hat',
    );
    return '$_temp0 nur leere Karten. Von jeder bleibt eine Karte erhalten, sodass keine Notiz gelöscht wird.';
  }

  @override
  String get emptyText => '(leer)';

  @override
  String get oneCardKept => 'eine Karte bleibt erhalten';

  @override
  String deleteCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 löschen';
  }

  @override
  String filesWithSize(int count, String size) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Dateien',
      one: '1 Datei',
    );
    return '$_temp0 ($size)';
  }

  @override
  String get deleteUnusedTitle => 'Ungenutzte Dateien löschen?';

  @override
  String deleteUnusedBody(String files) {
    return 'Verschiebt $files in den Medienpapierkorb. „Gelöschte wiederherstellen“ holt sie zurück, bis der Papierkorb geleert wird; Dateien werden nach 30 Tagen aus dem Papierkorb entfernt.';
  }

  @override
  String movedToTrash(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Dateien',
      one: '1 Datei',
    );
    return '$_temp0 in den Medienpapierkorb verschoben';
  }

  @override
  String get tagMissingTitle => 'Notizen mit fehlenden Dateien kennzeichnen?';

  @override
  String tagMissingBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Fügt $_temp0 das Schlagwort missing-media hinzu, damit du sie im Browser mit tag:missing-media findest.';
  }

  @override
  String get tag => 'Kennzeichnen';

  @override
  String taggedMissing(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 mit missing-media gekennzeichnet';
  }

  @override
  String get restoreDeletedTitle => 'Gelöschte Dateien wiederherstellen?';

  @override
  String restoreDeletedBody(String files) {
    return 'Verschiebt $files zurück in deine Medien.';
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
      other: '$countString Dateien',
      one: '1 Datei',
    );
    String _temp1 = intl.Intl.pluralLogic(
      kept,
      locale: localeName,
      other: '$keptString Dateien sind',
      one: '1 Datei ist',
    );
    return '$_temp0 wiederhergestellt; $_temp1 im Papierkorb geblieben, weil jetzt eine andere Datei den Namen hat';
  }

  @override
  String restoredFiles(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Dateien',
      one: '1 Datei',
    );
    return '$_temp0 wiederhergestellt';
  }

  @override
  String get emptyTrashTitle => 'Medienpapierkorb leeren?';

  @override
  String emptyTrashBody(String files) {
    return 'Löscht $files endgültig. Das kann nicht rückgängig gemacht werden.';
  }

  @override
  String get emptyTrash => 'Papierkorb leeren';

  @override
  String get trashEmptied => 'Medienpapierkorb geleert';

  @override
  String get unusedFiles => 'Ungenutzte Dateien';

  @override
  String get allMediaUsed =>
      'Jede Mediendatei wird von einer Notiz, einer Vorlage oder einem Stil verwendet.';

  @override
  String unusedFilesBody(String files) {
    return '$files, auf die keine Notiz, Vorlage und kein Stil verweist. Dateien, deren Name mit _ beginnt, werden immer behalten.';
  }

  @override
  String get deleteUnused => 'Ungenutzte löschen';

  @override
  String get missingFiles => 'Fehlende Dateien';

  @override
  String get noMissingFiles =>
      'Jede Datei, auf die Notizen verweisen, ist vorhanden.';

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
      other: '$countString Dateien',
      one: '1 Datei',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen verweisen',
      one: '1 Notiz verweist',
    );
    String _temp2 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'fehlen',
      one: 'fehlt',
    );
    return '$_temp0, auf die $_temp1, $_temp2 in deinen Medien.';
  }

  @override
  String get tagMissing => 'Fehlende kennzeichnen';

  @override
  String get mediaTrash => 'Medienpapierkorb';

  @override
  String get trashIsEmpty => 'Der Papierkorb ist leer.';

  @override
  String trashBody(String files) {
    return '$files. Dateien werden nach 30 Tagen aus dem Papierkorb entfernt.';
  }

  @override
  String get restoreDeleted => 'Gelöschte wiederherstellen';

  @override
  String get checkingNotes => 'Notizen werden geprüft';

  @override
  String get checkingCards => 'Karten werden geprüft';

  @override
  String get readingHistory => 'Wiederholungsverlauf wird gelesen';

  @override
  String get savingRepairs => 'Reparaturen werden gespeichert';

  @override
  String get backingUp => 'Sammlung wird gesichert';

  @override
  String get noProblemsFound => 'Es wurden keine Probleme gefunden.';

  @override
  String orphanedReviews(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Einträge im Wiederholungsverlauf gehören',
      one: '1 Eintrag im Wiederholungsverlauf gehört',
    );
    return '$_temp0 zu gelöschten Karten. Sie bleiben für die Statistiken erhalten.';
  }

  @override
  String get updateIfNewer => 'Wenn die Version im Paket neuer ist';

  @override
  String get updateAlways => 'Immer';

  @override
  String get updateNever => 'Nie';

  @override
  String get separatorComma => 'Komma';

  @override
  String get separatorSemicolon => 'Semikolon';

  @override
  String get separatorTab => 'Tabulator';

  @override
  String get separatorSpace => 'Leerzeichen';

  @override
  String get separatorPipe => 'Senkrechter Strich';

  @override
  String get separatorColon => 'Doppelpunkt';

  @override
  String get scopeNoteType => 'In ihrem Notiztyp';

  @override
  String get scopeDeck => 'In ihrem Notiztyp und Stapel';

  @override
  String importedNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 importiert';
  }

  @override
  String get importIntro =>
      'Importiere Notizen aus einer Textdatei (.csv, .tsv oder .txt) oder Stapel aus einem Anki-Paket (.apkg). Ein Anki-Sammlungspaket (.colpkg) ersetzt deine Sammlung. Die Datei wird auf diesem Gerät gelesen und nie hochgeladen.';

  @override
  String get chooseAFile => 'Datei wählen';

  @override
  String get importFailed => 'Der Import ist fehlgeschlagen';

  @override
  String get collectionNotChanged => 'Deine Sammlung wurde nicht geändert.';

  @override
  String get chooseAnotherFile => 'Andere Datei wählen';

  @override
  String fileIsBackup(String file) {
    return '$file ist eine Cardflip-Sicherung. Stelle sie in den Einstellungen wieder her, um deine Sammlung zu ersetzen.';
  }

  @override
  String columnNumber(int number) {
    return 'Spalte $number';
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
      other: '$rowsString Zeilen',
      one: '1 Zeile',
    );
    String _temp1 = intl.Intl.pluralLogic(
      columns,
      locale: localeName,
      other: '$columnsString Spalten',
      one: '1 Spalte',
    );
    return '$file: $_temp0, $_temp1';
  }

  @override
  String tagsForEveryNote(String tags) {
    return 'Schlagwörter, die jeder Notiz hinzugefügt werden: $tags';
  }

  @override
  String rowsSkippedUnknownType(int count, String names) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Zeilen werden',
      one: '1 Zeile wird',
    );
    return '$_temp0 übersprungen: Es gibt keinen Notiztyp namens $names.';
  }

  @override
  String get noteTypeOfRowsWithout => 'Notiztyp für Zeilen ohne eigenen';

  @override
  String get noteType => 'Notiztyp';

  @override
  String get deckOfRowsWithout => 'Stapel für Zeilen ohne eigenen';

  @override
  String get deck => 'Stapel';

  @override
  String newDeckName(String deck) {
    return '$deck (neu)';
  }

  @override
  String get separator => 'Trennzeichen';

  @override
  String get allowHtml => 'HTML in Feldern erlauben';

  @override
  String get allowHtmlHelp =>
      'Wenn aus, wird Text wie „a < b & c“ so importiert, wie er geschrieben ist.';

  @override
  String get firstRowNames => 'Die erste Zeile enthält Spaltennamen';

  @override
  String get tags => 'Schlagwörter';

  @override
  String get guid => 'GUID';

  @override
  String get ignore => 'Ignorieren';

  @override
  String get whenFirstFieldMatches =>
      'Wenn das erste Feld zu einer vorhandenen Notiz passt';

  @override
  String get guidUpdates =>
      'Eine Zeile, deren GUID zu einer Notiz ihres Notiztyps passt, aktualisiert diese Notiz.';

  @override
  String get skipRow => 'Zeile überspringen';

  @override
  String get updateExistingNote => 'Vorhandene Notiz aktualisieren';

  @override
  String get addAsDuplicate => 'Als Duplikat hinzufügen';

  @override
  String get lookForMatchingNotes => 'Nach passenden Notizen suchen';

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
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString Karten',
      one: '1 Karte',
    );
    String _temp2 = intl.Intl.pluralLogic(
      media,
      locale: localeName,
      other: '$mediaString Mediendateien',
      one: '1 Mediendatei',
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
      other: '$countString Audio- oder Videodateien werden',
      one: '1 Audio- oder Videodatei wird',
    );
    return '$_temp0 übersprungen: Cardflip kann ihr Format nicht abspielen.';
  }

  @override
  String get scheduling => 'Planung';

  @override
  String get keepScheduling => 'Planung beibehalten';

  @override
  String get keepSchedulingHelp =>
      'Karten behalten ihre Fälligkeitsdaten und ihren Fortschritt aus Anki.';

  @override
  String get importAsNew => 'Alle Karten als neu importieren';

  @override
  String get importAsNewHelp => 'Alles von vorn lernen.';

  @override
  String get notesAlreadyThere => 'Notizen, die schon in deiner Sammlung sind';

  @override
  String get notesAlreadyThereHelp =>
      'Notizen und Notiztypen aus einem früheren Import dieses Stapels. Ihre Planung bleibt in jedem Fall erhalten.';

  @override
  String get updateNotes => 'Notizen aktualisieren';

  @override
  String get updateNoteTypes => 'Notiztypen aktualisieren';

  @override
  String get mergeNoteTypes => 'Notiztypen zusammenführen';

  @override
  String get mergeNoteTypesHelp =>
      'Wenn das Paket Felder oder Kartentypen hinzufügt, werden sie deinem Notiztyp hinzugefügt. Sonst wird die Version aus dem Paket als eigener Notiztyp importiert.';

  @override
  String replaceByImportBody(String file) {
    return 'Der Import von $file ersetzt deine aktuelle Sammlung vollständig: Alle Stapel, Notizen, Karten, der Wiederholungsverlauf und die Mediendateien werden überschrieben. Deine Einstellungen bleiben erhalten. Das kann nicht rückgängig gemacht werden.';
  }

  @override
  String get replace => 'Ersetzen';

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
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString Karten',
      one: '1 Karte',
    );
    String _temp2 = intl.Intl.pluralLogic(
      media,
      locale: localeName,
      other: '$mediaString Mediendateien',
      one: '1 Mediendatei',
    );
    String _temp3 = intl.Intl.pluralLogic(
      decks,
      locale: localeName,
      other: '$decksString Stapeln',
      one: '1 Stapel',
    );
    return 'Anki-Sammlung: $_temp0, $_temp1, $_temp2 in $_temp3';
  }

  @override
  String get replaceWarning =>
      'Dadurch wird deine aktuelle Sammlung durch die Anki-Sammlung mit ihrer Planung und ihrem Wiederholungsverlauf ersetzt. Alles, was du jetzt in Cardflip hast, wird überschrieben; deine Einstellungen bleiben erhalten. Cardflip erstellt vorher eine automatische Sicherung, aber du möchtest vielleicht eine Sicherung in eine Datei exportieren, bevor du fortfährst.';

  @override
  String get exportBackupFirst => 'Zuerst eine Sicherung exportieren';

  @override
  String get replaceCollection => 'Sammlung ersetzen';

  @override
  String get addingNotes => 'Notizen werden hinzugefügt';

  @override
  String get addingHistory => 'Wiederholungsverlauf wird hinzugefügt';

  @override
  String get savingImages => 'Bilder werden gespeichert';

  @override
  String get saving => 'Wird gespeichert';

  @override
  String get readingFile => 'Datei wird gelesen';

  @override
  String progressOf(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return '$doneString von $total';
  }

  @override
  String notesAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 hinzugefügt';
  }

  @override
  String notesUpdated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 aktualisiert';
  }

  @override
  String notesSkipped(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 übersprungen';
  }

  @override
  String notesNotUpdatedNewer(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 nicht aktualisiert: Deine Version ist neuer';
  }

  @override
  String cardsAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 hinzugefügt';
  }

  @override
  String reviewsImported(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
    );
    return '$_temp0 aus dem Verlauf importiert';
  }

  @override
  String mediaImported(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Mediendateien',
      one: '1 Mediendatei',
    );
    return '$_temp0 importiert';
  }

  @override
  String commentLinesSkipped(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Kommentarzeilen',
      one: '1 Kommentarzeile',
    );
    return '$_temp0 übersprungen';
  }

  @override
  String noteTypesAdded(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notiztypen',
      one: '1 Notiztyp',
    );
    return '$_temp0 hinzugefügt';
  }

  @override
  String noteTypesUpdated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notiztypen',
      one: '1 Notiztyp',
    );
    return '$_temp0 aktualisiert';
  }

  @override
  String noteTypesMerged(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notiztypen',
      one: '1 Notiztyp',
    );
    return '$_temp0 zusammengeführt';
  }

  @override
  String get collectionReplaced => 'Sammlung ersetzt';

  @override
  String get importComplete => 'Import abgeschlossen';

  @override
  String get warnings => 'Warnungen';

  @override
  String get done => 'Fertig';

  @override
  String copyName(String name) {
    return '$name Kopie';
  }

  @override
  String get addNoteType => 'Notiztyp hinzufügen';

  @override
  String cloneNamed(String name) {
    return '„$name“ klonen';
  }

  @override
  String get noNotesUseType => 'Keine Notizen verwenden diesen Notiztyp.';

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
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cardsString Karten',
      one: '1 Karte',
    );
    return 'Dadurch werden $_temp0 und $_temp1 gelöscht.';
  }

  @override
  String get canUndoUntilClose =>
      'Du kannst das rückgängig machen, bis du Cardflip schließt.';

  @override
  String get noteTypes => 'Notiztypen';

  @override
  String get builtinDeckName => 'Standard';

  @override
  String get builtinPresetName => 'Standard';

  @override
  String get stockBasic => 'Einfach';

  @override
  String get stockReversed => 'Einfach (mit umgekehrter Karte)';

  @override
  String get stockOptionalReversed => 'Einfach (optional umgekehrte Karte)';

  @override
  String get stockTypeIn => 'Einfach (Antwort eintippen)';

  @override
  String get stockCloze => 'Lückentext';

  @override
  String get stockImageOcclusion => 'Bildverdeckung';

  @override
  String get stockFieldFront => 'Vorderseite';

  @override
  String get stockFieldBack => 'Rückseite';

  @override
  String get stockFieldAddReverse => 'Umgekehrte Karte hinzufügen';

  @override
  String get stockFieldText => 'Text';

  @override
  String get stockFieldBackExtra => 'Rückseite Extra';

  @override
  String get stockFieldOcclusion => 'Verdeckung';

  @override
  String get stockFieldImage => 'Bild';

  @override
  String get stockFieldHeader => 'Kopfzeile';

  @override
  String get stockFieldComments => 'Kommentare';

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
      other: '$fieldsString Felder',
      one: '1 Feld',
    );
    String _temp1 = intl.Intl.pluralLogic(
      cardTypes,
      locale: localeName,
      other: '$cardTypesString Kartentypen',
      one: '1 Kartentyp',
    );
    String _temp2 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 · $_temp1 · $_temp2';
  }

  @override
  String get noteTypeActions => 'Aktionen des Notiztyps';

  @override
  String get renameNoteType => 'Notiztyp umbenennen';

  @override
  String get renameField => 'Feld umbenennen';

  @override
  String get addField => 'Feld hinzufügen';

  @override
  String removeFieldTitle(String field) {
    return 'Feld „$field“ entfernen?';
  }

  @override
  String removeFieldBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Das Feld und sein Inhalt werden aus $_temp0 entfernt. Du kannst das rückgängig machen, bis du Cardflip schließt.';
  }

  @override
  String get remove => 'Entfernen';

  @override
  String defaultCardTypeName(int number) {
    return 'Karte $number';
  }

  @override
  String removeCardTypeTitle(String name) {
    return 'Kartentyp „$name“ entfernen?';
  }

  @override
  String removeCardTypeBody(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Dadurch werden $_temp0 gelöscht, und ihr Wiederholungsverlauf kann nicht mehr verwendet werden.';
  }

  @override
  String get fields => 'Felder';

  @override
  String get fieldOptions => 'Feldoptionen';

  @override
  String get removeField => 'Feld entfernen';

  @override
  String get dragToReorder => 'Zum Umsortieren ziehen';

  @override
  String get cardTypes => 'Kartentypen';

  @override
  String get cardTypeOptions => 'Optionen des Kartentyps';

  @override
  String get removeCardType => 'Kartentyp entfernen';

  @override
  String get addCardType => 'Kartentyp hinzufügen';

  @override
  String get styling => 'Stil';

  @override
  String get editCss => 'CSS bearbeiten';

  @override
  String get sortField => 'Sortierfeld';

  @override
  String get rightToLeft => 'Rechts nach links';

  @override
  String get collapsed => 'Zugeklappt';

  @override
  String get notInUnqualifiedSearches => 'Nicht in einfachen Suchen';

  @override
  String get htmlEditor => 'HTML-Editor';

  @override
  String get pinned => 'Angeheftet';

  @override
  String get fieldSizeError =>
      'Gib eine Größe von 1 bis 200 ein oder lass das Feld leer.';

  @override
  String optionsForNamed(String name) {
    return 'Optionen für „$name“';
  }

  @override
  String get editingFont => 'Schrift im Editor';

  @override
  String get editingFontHelp =>
      'Wird nur im Editor für dieses Feld verwendet, nicht auf Karten.';

  @override
  String get font => 'Schrift';

  @override
  String get defaultValue => 'Standard';

  @override
  String get size => 'Größe';

  @override
  String get sortByField => 'Nach diesem Feld sortieren';

  @override
  String get sortByFieldIsHelp =>
      'Die Spalte „Sortierfeld“ im Browser zeigt dieses Feld. Wähle ein anderes Feld, um das zu ändern.';

  @override
  String get sortByFieldHelp =>
      'Die Spalte „Sortierfeld“ im Browser zeigt dieses Feld.';

  @override
  String get rightToLeftHelp =>
      'Für Arabisch, Hebräisch und andere Schriften, die von rechts nach links laufen.';

  @override
  String get collapseByDefault => 'Standardmäßig zuklappen';

  @override
  String get collapseByDefaultHelp =>
      'Der Editor zeigt nur den Feldnamen, bis du darauf tippst.';

  @override
  String get excludeFromSearch => 'Von einfachen Suchen ausschließen';

  @override
  String excludeFromSearchHelp(String field) {
    return 'Einfache Suchbegriffe überspringen dieses Feld; $field:… findet es trotzdem.';
  }

  @override
  String get htmlByDefault => 'Standardmäßig den HTML-Editor verwenden';

  @override
  String get htmlByDefaultHelp =>
      'Das Feld als HTML-Quelltext öffnen statt im visuellen Editor.';

  @override
  String get keepAfterAdding => 'Nach dem Hinzufügen behalten';

  @override
  String get keepAfterAddingHelp =>
      'Heftet das Feld an, damit sein Inhalt nach dem Hinzufügen einer Notiz im Hinzufügen-Bildschirm bleibt.';

  @override
  String get deckOverride => 'Abweichender Stapel';

  @override
  String get deckOverrideHelp =>
      'Neue Karten dieses Kartentyps kommen in diesen Stapel. Vorhandene Karten werden nicht verschoben.';

  @override
  String get deckOverrideNone => 'Keiner (der beim Hinzufügen gewählte Stapel)';

  @override
  String get browserAppearance => 'Darstellung im Browser';

  @override
  String get browserAppearanceHelp =>
      'Vereinfachte Vorlagen für die Frage- und Antwortspalten des Browsers. Leer lassen, um die normalen Vorlagen zu verwenden.';

  @override
  String get browserQuestionTemplate => 'Fragevorlage im Browser';

  @override
  String get browserAnswerTemplate => 'Antwortvorlage im Browser';

  @override
  String get edit => 'Bearbeiten';

  @override
  String get preview => 'Vorschau';

  @override
  String get frontTemplate => 'Vorlage der Vorderseite';

  @override
  String get backTemplate => 'Vorlage der Rückseite';

  @override
  String get templateHelp =>
      'Füge ein Feld mit {{Field}} ein. Setze einen Filter vor den Namen, z. B. {{type}} für ein Eingabefeld für die Antwort, {{hint}}, {{text}} oder {{cloze}}. Außerdem {{FrontSide}} und Abschnitte {{#Field}}…{{/Field}}.';

  @override
  String get newCardType => 'Neuer Kartentyp';

  @override
  String get editCardType => 'Kartentyp bearbeiten';

  @override
  String get stylingCss => 'Stil (CSS)';

  @override
  String get lossyNote =>
      'Dieses Feld enthält Formatierungen, die der visuelle Editor nicht erhalten kann. Es wird als HTML bearbeitet.';

  @override
  String get switchToVisualTitle => 'Zum visuellen Editor wechseln?';

  @override
  String get switchToVisualBody =>
      'Dieses Feld enthält Formatierungen, die der visuelle Editor nicht erhalten kann. Wenn du es dort bearbeitest, können diese Formatierungen verloren gehen.';

  @override
  String get switchAction => 'Wechseln';

  @override
  String get expandField => 'Feld aufklappen';

  @override
  String get collapseField => 'Feld zuklappen';

  @override
  String get visualEditor => 'Visueller Editor';

  @override
  String get emptyField => 'Leer';

  @override
  String get bold => 'Fett';

  @override
  String get italic => 'Kursiv';

  @override
  String get underline => 'Unterstrichen';

  @override
  String get subscript => 'Tiefgestellt';

  @override
  String get superscript => 'Hochgestellt';

  @override
  String get textColour => 'Textfarbe';

  @override
  String get highlightColour => 'Hervorhebungsfarbe';

  @override
  String get removeFormatting => 'Formatierung entfernen';

  @override
  String get bulletedList => 'Aufzählung';

  @override
  String get numberedList => 'Nummerierte Liste';

  @override
  String get alignment => 'Ausrichtung';

  @override
  String get alignLeft => 'Linksbündig';

  @override
  String get alignCenter => 'Zentriert';

  @override
  String get alignRight => 'Rechtsbündig';

  @override
  String get justify => 'Blocksatz';

  @override
  String get outdent => 'Einzug verkleinern';

  @override
  String get indent => 'Einzug vergrößern';

  @override
  String get colourBlack => 'Schwarz';

  @override
  String get colourGray => 'Grau';

  @override
  String get colourRed => 'Rot';

  @override
  String get colourOrange => 'Orange';

  @override
  String get colourYellow => 'Gelb';

  @override
  String get colourGreen => 'Grün';

  @override
  String get colourBlue => 'Blau';

  @override
  String get colourPurple => 'Lila';

  @override
  String get noColour => 'Keine Farbe';

  @override
  String get mathInline => 'MathJax im Text';

  @override
  String get mathBlock => 'MathJax als Block';

  @override
  String get mathChemistry => 'MathJax Chemie';

  @override
  String get mathLatex => 'LaTeX';

  @override
  String get mathLatexEquation => 'LaTeX-Gleichung';

  @override
  String get mathLatexMath => 'LaTeX-Mathe-Umgebung';

  @override
  String get math => 'Mathe';

  @override
  String get clozeDeletion => 'Lücke';

  @override
  String get nextClozeNumber => 'Nächste Lückennummer';

  @override
  String get sameClozeNumber => 'Gleiche Lückennummer';

  @override
  String get attachMedia => 'Bild, Audio oder Video anhängen';

  @override
  String get recordAudio => 'Audio aufnehmen';

  @override
  String get onlyMediaInFields =>
      'Nur Bilder, Audio und Video können einem Feld hinzugefügt werden.';

  @override
  String get noNotesYet => 'Noch keine Notizen';

  @override
  String clozeNumber(int number) {
    return 'Lücke $number';
  }

  @override
  String get previewLightTheme => 'Vorschau im hellen Design';

  @override
  String get previewDarkTheme => 'Vorschau im dunklen Design';

  @override
  String get front => 'Vorderseite';

  @override
  String get back => 'Rückseite';

  @override
  String get recording => 'Aufnahme läuft…';

  @override
  String get insert => 'Einfügen';

  @override
  String get addTags => 'Schlagwörter hinzufügen';

  @override
  String imageMissing(String file) {
    return 'Das Bild $file fehlt.';
  }

  @override
  String get choosePictureToDrawOn =>
      'Wähle ein Bild, auf das du zeichnen möchtest.';

  @override
  String get chooseImageToDrawOn =>
      'Wähle ein Bild, auf das du zeichnen möchtest.';

  @override
  String get drawShapes => 'Zeichne über jeden abzufragenden Teil eine Form.';

  @override
  String get cardAdded => 'Karte hinzugefügt';

  @override
  String get editCard => 'Karte bearbeiten';

  @override
  String get historyAddedToday => 'Verlauf: heute hinzugefügte Karten';

  @override
  String get image => 'Bild';

  @override
  String get changeImage => 'Bild ändern';

  @override
  String get loadingImage => 'Bild wird geladen…';

  @override
  String get chooseImage => 'Bild wählen';

  @override
  String duplicateOne(String field) {
    return 'Dies ist ein Duplikat: Eine andere Notiz hat dasselbe Feld $field.';
  }

  @override
  String duplicateMany(int count, String field) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return 'Dies ist ein Duplikat: $countString andere Notizen haben dasselbe Feld $field.';
  }

  @override
  String get showDuplicates => 'Duplikate zeigen';

  @override
  String get unpinField => 'Lösen: nach dem Hinzufügen leeren';

  @override
  String get pinField => 'Anheften: nach dem Hinzufügen behalten';

  @override
  String get hideAllGuessOne => 'Alle verdecken, eine erraten';

  @override
  String get hideOneGuessOne => 'Eine verdecken, eine erraten';

  @override
  String get selectAndMove => 'Auswählen und verschieben';

  @override
  String get rectangle => 'Rechteck';

  @override
  String get ellipse => 'Ellipse';

  @override
  String get polygonTool =>
      'Polygon: Tippe auf die Ecken, dann noch einmal auf die erste';

  @override
  String get textLabel => 'Textbeschriftung';

  @override
  String get finishPolygon => 'Polygon abschließen';

  @override
  String get groupShapes =>
      'Gruppieren: die ausgewählten Formen gemeinsam abfragen';

  @override
  String get ungroup => 'Gruppierung aufheben';

  @override
  String get deleteSelectedShapes => 'Ausgewählte Formen löschen';

  @override
  String get zoomOut => 'Verkleinern';

  @override
  String get zoomIn => 'Vergrößern';

  @override
  String get fitImage => 'Bild einpassen';

  @override
  String get imageCannotShow => 'Dieses Bild kann nicht angezeigt werden.';

  @override
  String get shownOnEveryCard => 'Auf jeder Karte gezeigt';

  @override
  String get repositionNewCards => 'Neue Karten neu positionieren';

  @override
  String get startPosition => 'Startposition';

  @override
  String get step => 'Schrittweite';

  @override
  String get randomizeOrder => 'Reihenfolge zufällig';

  @override
  String get shiftExisting => 'Position vorhandener Karten verschieben';

  @override
  String get repositionHelp =>
      'Nur neue Karten werden neu positioniert; andere bleiben unverändert.';

  @override
  String get reposition => 'Neu positionieren';

  @override
  String wholeNumberRange(int min, int max) {
    return 'Eine ganze Zahl von $min bis $max';
  }

  @override
  String get setDueDate => 'Fälligkeitsdatum setzen';

  @override
  String setDueDateOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Fälligkeitsdatum von $_temp0 setzen';
  }

  @override
  String get daysFromToday => 'Tage ab heute';

  @override
  String get dueDateHint => 'z. B. 0, 7, 60-90 oder 30!';

  @override
  String get dueDateHelp =>
      '0 = heute, 1 = morgen, 7 = in einer Woche\n60-90 = für jede Karte ein zufälliger Tag in 60 bis 90 Tagen\nHänge ! an, um das Intervall auf dieselbe Anzahl Tage zu setzen, z. B. 30!';

  @override
  String deleteCardsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 löschen?';
  }

  @override
  String resetCardsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 auf neu zurücksetzen?';
  }

  @override
  String get resetCardsBody =>
      'Ihre Planung wird gelöscht, und sie werden als neue Karten gelernt. Ihr Wiederholungsverlauf bleibt erhalten.';

  @override
  String get restorePosition =>
      'Ursprüngliche Position wenn möglich wiederherstellen';

  @override
  String get resetCounts => 'Wiederholungs- und Fehlerzähler zurücksetzen';

  @override
  String get reset => 'Zurücksetzen';

  @override
  String get noFlag => 'Keine Flagge';

  @override
  String taggedNotes(int count, String tag) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 mit $tag gekennzeichnet';
  }

  @override
  String get findDuplicates => 'Duplikate finden';

  @override
  String get tagDuplicates => 'Duplikate kennzeichnen';

  @override
  String get field => 'Feld';

  @override
  String get optionalSearch => 'Optionale Suche';

  @override
  String get optionalSearchHint => 'z. B. deck:Spanisch';

  @override
  String get find => 'Finden';

  @override
  String get noDuplicatesFound => 'Keine Duplikate gefunden.';

  @override
  String countNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
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
      other: '$countString weiteren Karten',
      one: '1 weiteren Karte',
    );
    return 'Einschließlich $_temp0 derselben Notiz. ';
  }

  @override
  String includingOtherCardsNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString weiteren Karten',
      one: '1 weiteren Karte',
    );
    return 'Einschließlich $_temp0 derselben Notizen. ';
  }

  @override
  String newPosition(int position) {
    return 'Neu #$position';
  }

  @override
  String get colQuestion => 'Frage';

  @override
  String get colAnswer => 'Antwort';

  @override
  String get colSortField => 'Sortierfeld';

  @override
  String get colCardType => 'Kartentyp';

  @override
  String get colNoteType => 'Notiztyp';

  @override
  String get colDeck => 'Stapel';

  @override
  String get colDue => 'Fällig';

  @override
  String get colInterval => 'Intervall';

  @override
  String get colDifficulty => 'Schwierigkeit';

  @override
  String get colStability => 'Stabilität';

  @override
  String get colRetrievability => 'Abrufbarkeit';

  @override
  String get colReviews => 'Wiederholungen';

  @override
  String get colLapses => 'Fehler';

  @override
  String get colTags => 'Schlagwörter';

  @override
  String get colCreated => 'Erstellt';

  @override
  String get colCardModified => 'Karte geändert';

  @override
  String get colNoteModified => 'Notiz geändert';

  @override
  String get colCards => 'Karten';

  @override
  String get emptyQuestion => '(leer)';

  @override
  String get stateSuspended => 'ausgesetzt';

  @override
  String get stateNew => 'neu';

  @override
  String dueWhen(String when) {
    return 'fällig $when';
  }

  @override
  String intervalOfState(String state) {
    String _temp0 = intl.Intl.selectLogic(state, {
      'new': 'neu',
      'learning': 'im Lernen',
      'relearning': 'im Neulernen',
      'other': '$state',
    });
    return '($_temp0)';
  }

  @override
  String get multipleDecks => '(mehrere)';

  @override
  String get moveLeft => 'Nach links verschieben';

  @override
  String get moveRight => 'Nach rechts verschieben';

  @override
  String get hideColumn => 'Spalte ausblenden';

  @override
  String get columnsEllipsis => 'Spalten…';

  @override
  String get columns => 'Spalten';

  @override
  String cardKind(String kind) {
    String _temp0 = intl.Intl.selectLogic(kind, {
      'new': 'Neu',
      'learning': 'Lernen',
      'review': 'Wiederholung',
      'relearning': 'Neulernen',
      'manual': 'Manuell',
      'other': '$kind',
    });
    return '$_temp0';
  }

  @override
  String secondsDecimal(String seconds) {
    return '$seconds s';
  }

  @override
  String get cardInfo => 'Karteninfo';

  @override
  String get reviewHistory => 'Wiederholungsverlauf';

  @override
  String get noReviewsYet => 'Noch keine Wiederholungen.';

  @override
  String get infoAdded => 'Hinzugefügt';

  @override
  String get infoFirstReview => 'Erste Wiederholung';

  @override
  String get infoLatestReview => 'Letzte Wiederholung';

  @override
  String get infoDue => 'Fällig';

  @override
  String get infoInterval => 'Intervall';

  @override
  String get infoStability => 'Stabilität';

  @override
  String get infoDifficulty => 'Schwierigkeit';

  @override
  String get infoRetrievability => 'Abrufbarkeit';

  @override
  String get infoReviews => 'Wiederholungen';

  @override
  String get infoLapses => 'Fehler';

  @override
  String get averageTime => 'Durchschnittliche Zeit';

  @override
  String get totalTime => 'Gesamtzeit';

  @override
  String get infoState => 'Zustand';

  @override
  String get cardType => 'Kartentyp';

  @override
  String get cardId => 'Karten-ID';

  @override
  String get noteId => 'Notiz-ID';

  @override
  String get historyDate => 'Datum';

  @override
  String get historyKind => 'Art';

  @override
  String get historyRating => 'Bewertung';

  @override
  String get historyInterval => 'Intervall';

  @override
  String get historyTime => 'Zeit';

  @override
  String deletedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 gelöscht';
  }

  @override
  String setDueDateOfCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Fälligkeitsdatum von $_temp0 gesetzt';
  }

  @override
  String resetCardsToNew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 auf neu zurückgesetzt';
  }

  @override
  String get editAction => 'Bearbeiten';

  @override
  String get resetToNew => 'Auf neu zurücksetzen';

  @override
  String get unsuspend => 'Aussetzung aufheben';

  @override
  String get suspend => 'Aussetzen';

  @override
  String stateWithSuspended(String state) {
    return '$state (ausgesetzt)';
  }

  @override
  String get newCard => 'neue Karte';

  @override
  String countDays(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Tage',
      one: '1 Tag',
    );
    return '$_temp0';
  }

  @override
  String daysDecimal(String days) {
    return '$days Tage';
  }

  @override
  String get card => 'Karte';

  @override
  String get question => 'Frage';

  @override
  String get answer => 'Antwort';

  @override
  String changesApplyToOthers(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString weitere Karten',
      one: '1 weitere Karte',
    );
    return 'Änderungen gelten auch für $_temp0 dieser Notiz.';
  }

  @override
  String get cardSaved => 'Karte gespeichert';

  @override
  String changeCardsDeleted(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten werden',
      one: '1 Karte wird',
    );
    return '$_temp0 gelöscht.';
  }

  @override
  String changeCardsCreated(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten werden',
      one: '1 Karte wird',
    );
    return '$_temp0 erstellt.';
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
      other: '$deletedString Karten werden',
      one: '1 Karte wird',
    );
    String _temp1 = intl.Intl.pluralLogic(
      created,
      locale: localeName,
      other: '$createdString Karten werden',
      one: '1 Karte wird',
    );
    return '$_temp0 gelöscht und $_temp1 erstellt.';
  }

  @override
  String get changeNoCards => 'Es werden keine Karten gelöscht oder erstellt.';

  @override
  String changeLostFields(String fields) {
    return 'Der Inhalt von $fields geht verloren.';
  }

  @override
  String listAnd(String items, String last) {
    return '$items und $last';
  }

  @override
  String changeNotesTitle(int count, String noteType) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 in $noteType ändern?';
  }

  @override
  String get change => 'Ändern';

  @override
  String get changeNoteType => 'Notiztyp ändern';

  @override
  String currentNoteType(String noteType) {
    return 'Aktueller Notiztyp: $noteType';
  }

  @override
  String get newNoteType => 'Neuer Notiztyp';

  @override
  String get changeFieldsHelp =>
      'Jedes neue Feld übernimmt den Inhalt des Feldes, das du wählst.';

  @override
  String get changeCardsHelp =>
      'Jeder neue Kartentyp übernimmt die Karten des Kartentyps, den du wählst, mit ihrer Planung. Karten nicht gewählter Kartentypen werden gelöscht.';

  @override
  String get changeClozeHelp =>
      'Karten behalten ihre Planung, wenn die geänderte Notiz eine Karte mit derselben Nummer hat. Andere Karten werden gelöscht und fehlende erstellt.';

  @override
  String get nothing => '(nichts)';

  @override
  String get saveSearch => 'Suche speichern';

  @override
  String savedSearchAs(String name) {
    return 'Die Suche wurde als „$name“ gespeichert.';
  }

  @override
  String get savedSearches => 'Gespeicherte Suchen';

  @override
  String get saveCurrentSearch => 'Aktuelle Suche speichern';

  @override
  String get manageSavedSearches =>
      'Gespeicherte Suchen umbenennen oder löschen';

  @override
  String get renameSavedSearch => 'Gespeicherte Suche umbenennen';

  @override
  String deleteQuoted(String name) {
    return '„$name“ löschen?';
  }

  @override
  String deleteSavedSearchBody(String query) {
    return 'Die gespeicherte Suche $query wird entfernt. Karten werden nicht geändert.';
  }

  @override
  String get noSavedSearches => 'Keine gespeicherten Suchen.';

  @override
  String get findTextPrompt => 'Gib den zu suchenden Text ein.';

  @override
  String get counting => 'Zähle…';

  @override
  String get noNotesWillChange => 'Keine Notizen werden geändert.';

  @override
  String notesWillChange(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen werden',
      one: '1 Notiz wird',
    );
    return '$_temp0 geändert.';
  }

  @override
  String get findAndReplace => 'Suchen und ersetzen';

  @override
  String get replaceWith => 'Ersetzen durch';

  @override
  String get regexGroupsHelp => 'Verwende \$1 oder \${name} für Gruppen';

  @override
  String get inField => 'In';

  @override
  String get allFields => 'Allen Feldern';

  @override
  String get selectedNotesOnly => 'Nur ausgewählte Notizen';

  @override
  String get treatAsRegex => 'Eingabe als regulären Ausdruck behandeln';

  @override
  String get ignoreCase => 'Groß-/Kleinschreibung ignorieren';

  @override
  String movedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 verschoben';
  }

  @override
  String get tagsSeparated => 'Schlagwörter, durch Leerzeichen getrennt';

  @override
  String get removeTags => 'Schlagwörter entfernen';

  @override
  String cardsNotNew(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten waren nicht neu und blieben',
      one: '1 Karte war nicht neu und blieb',
    );
    return ' $_temp0 unverändert.';
  }

  @override
  String get noNewCardsSelected => 'Keine neuen Karten ausgewählt.';

  @override
  String repositionedCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 neu positioniert.';
  }

  @override
  String changedNoteTypeOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Notiztyp von $_temp0 geändert';
  }

  @override
  String get noNotesChanged => 'Keine Notizen geändert.';

  @override
  String changedNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 geändert';
  }

  @override
  String get moveToDeck => 'In Stapel verschieben';

  @override
  String get clearSelection => 'Auswahl aufheben';

  @override
  String countSelected(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    return '$countString ausgewählt';
  }

  @override
  String get selectCards => 'Karten auswählen';

  @override
  String get selectNotes => 'Notizen auswählen';

  @override
  String get selectAll => 'Alle auswählen';

  @override
  String get selection => 'Auswahl';

  @override
  String get invertSelection => 'Auswahl umkehren';

  @override
  String get bulkActions => 'Sammelaktionen';

  @override
  String get toggleMark => 'Markierung umschalten';

  @override
  String get deleteCardsAction => 'Karten löschen';

  @override
  String get deleteNotesAction => 'Notizen löschen';

  @override
  String get browse => 'Browser';

  @override
  String get sidebar => 'Seitenleiste';

  @override
  String get sort => 'Sortieren';

  @override
  String get noCardsYet => 'Noch keine Karten.';

  @override
  String get noCardsMatch => 'Keine Karten passen zu deiner Suche.';

  @override
  String get noNotesMatch => 'Keine Notizen passen zu deiner Suche.';

  @override
  String get searchHint => 'Suchen, z. B. deck:Spanisch tag:Verb -is:suspended';

  @override
  String onlyNotes(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Nur $_temp0';
  }

  @override
  String get showAllCards => 'Alle Karten zeigen';

  @override
  String get showAllNotes => 'Alle Notizen zeigen';

  @override
  String get modeCards => 'Karten';

  @override
  String get modeNotes => 'Notizen';

  @override
  String get oneRowPerCard => 'Eine Zeile pro Karte';

  @override
  String get oneRowPerNote => 'Eine Zeile pro Notiz';

  @override
  String get selectCardForDetails =>
      'Wähle eine Karte, um ihre Details zu sehen.';

  @override
  String get selectNoteForDetails =>
      'Wähle eine Notiz, um ihre Details zu sehen.';

  @override
  String get stateLearning => 'Lernen';

  @override
  String get stateReview => 'Wiederholung';

  @override
  String get stateDue => 'Fällig';

  @override
  String get stateSuspendedTitle => 'Ausgesetzt';

  @override
  String get renameTag => 'Schlagwort umbenennen';

  @override
  String get renameTagHelper =>
      'Verwende ::, um es unter ein anderes Schlagwort zu verschieben';

  @override
  String renamedTagOn(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Notizen',
      one: '1 Notiz',
    );
    return 'Schlagwort bei $_temp0 umbenannt';
  }

  @override
  String deleteTagTitle(String tag) {
    return 'Das Schlagwort „$tag“ löschen?';
  }

  @override
  String deleteTagsTitle(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    return '$_temp0 löschen?';
  }

  @override
  String deleteTagBody(int notes) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    return 'Dadurch wird es mit seinen Unterschlagwörtern aus $_temp0 entfernt. Die Notizen bleiben erhalten.';
  }

  @override
  String deleteTagsBody(int notes) {
    final intl.NumberFormat notesNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String notesString = notesNumberFormat.format(notes);

    String _temp0 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    return 'Dadurch werden sie mit ihren Unterschlagwörtern aus $_temp0 entfernt. Die Notizen bleiben erhalten.';
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
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    String _temp1 = intl.Intl.pluralLogic(
      notes,
      locale: localeName,
      other: '$notesString Notizen',
      one: '1 Notiz',
    );
    return '$_temp0 aus $_temp1 entfernt';
  }

  @override
  String movedTagsTop(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    return '$_temp0 auf die oberste Ebene verschoben';
  }

  @override
  String movedTagsUnder(int count, String parent) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    return '$_temp0 unter „$parent“ verschoben';
  }

  @override
  String moveTagsTo(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    return '$_temp0 verschieben nach';
  }

  @override
  String get topLevel => 'Oberste Ebene';

  @override
  String get noUnusedTags => 'Es gibt keine ungenutzten Schlagwörter.';

  @override
  String clearedUnusedTags(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString ungenutzte Schlagwörter',
      one: '1 ungenutztes Schlagwort',
    );
    return '$_temp0 entfernt.';
  }

  @override
  String movedDeckTop(String deck) {
    return '„$deck“ auf die oberste Ebene verschoben';
  }

  @override
  String movedDeckInto(String deck, String parent) {
    return '„$deck“ in „$parent“ verschoben';
  }

  @override
  String moveDeckTo(String deck) {
    return '„$deck“ verschieben nach';
  }

  @override
  String get addToSearch => 'Zur Suche hinzufügen';

  @override
  String get addToSearchOr => 'Mit ODER zur Suche hinzufügen';

  @override
  String get excludeFromSearchAction => 'Von der Suche ausschließen';

  @override
  String get renameEllipsis => 'Umbenennen…';

  @override
  String get moveToEllipsis => 'Verschieben nach…';

  @override
  String get selectTags => 'Schlagwörter auswählen';

  @override
  String get deleteEllipsis => 'Löschen…';

  @override
  String get filterHint => 'Filtern';

  @override
  String get clearFilter => 'Filter löschen';

  @override
  String get cardStates => 'Kartenzustände';

  @override
  String get tagActions => 'Schlagwortaktionen';

  @override
  String get clearUnusedTags => 'Ungenutzte Schlagwörter entfernen';

  @override
  String get noTagsYet => 'Noch keine Schlagwörter.';

  @override
  String tagCount(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
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
      other: '$countString Schlagwörter',
      one: '1 Schlagwort',
    );
    return '$_temp0 ausgewählt';
  }

  @override
  String get moveTagsToTooltip => 'Schlagwörter verschieben nach';

  @override
  String get deleteTags => 'Schlagwörter löschen';

  @override
  String exportNamed(String name) {
    return '„$name“ exportieren';
  }

  @override
  String exportCards(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$_temp0 exportieren';
  }

  @override
  String exportNotesOf(int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Die Notizen von $_temp0 exportieren';
  }

  @override
  String get exportStopped => 'Der Export wurde unerwartet beendet.';

  @override
  String exportedFile(String file) {
    return '$file exportiert';
  }

  @override
  String get preparing => 'Wird vorbereitet…';

  @override
  String renderingCards(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return 'Karten werden dargestellt ($doneString von $total)…';
  }

  @override
  String get writingFile => 'Datei wird geschrieben…';

  @override
  String addingMedia(int done, int total) {
    final intl.NumberFormat doneNumberFormat = intl.NumberFormat.decimalPattern(
      localeName,
    );
    final String doneString = doneNumberFormat.format(done);

    return 'Medien werden hinzugefügt ($doneString von $total)…';
  }

  @override
  String get packingFile => 'Datei wird gepackt…';

  @override
  String get writingCards => 'Karten werden geschrieben…';

  @override
  String get exportColpkgHelp =>
      'Ein Anki-Sammlungspaket (.colpkg) mit allen Stapeln, Notizen und Karten, ihrer Planung, dem Wiederholungsverlauf und den Stapelvoreinstellungen. Wenn du es in Anki öffnest, ersetzt es die Sammlung dort.';

  @override
  String get exportApkgHelp =>
      'Ein Anki-Stapelpaket (.apkg), das Anki, AnkiDroid und Cardflip importieren können.';

  @override
  String get exportNotesTextHelp =>
      'Eine Notiz pro Zeile, ihre Felder durch Tabulatoren getrennt, mit HTML. Cardflip und Anki können die Datei erneut importieren, um die Notizen zu aktualisieren.';

  @override
  String get exportCardsTextHelp =>
      'Eine Karte pro Zeile mit Frage und Antwort als reiner Text, zum Drucken oder Lesen an anderer Stelle.';

  @override
  String get format => 'Format';

  @override
  String get formatColpkg => 'Anki-Sammlungspaket (.colpkg)';

  @override
  String get formatApkg => 'Anki-Stapelpaket (.apkg)';

  @override
  String get formatNotesText => 'Notizen als reiner Text (.txt)';

  @override
  String get formatCardsText => 'Karten als reiner Text (.txt)';

  @override
  String get includeTags => 'Schlagwörter einschließen';

  @override
  String get includeDeck => 'Stapelnamen einschließen';

  @override
  String get includeNoteType => 'Namen des Notiztyps einschließen';

  @override
  String get includeGuid => 'Eindeutige Kennung einschließen';

  @override
  String get includeGuidHelp =>
      'Damit kann Anki diese Notizen aktualisieren, wenn es die Datei erneut importiert.';

  @override
  String get includeScheduling => 'Planungsdaten einschließen';

  @override
  String get includeSchedulingHelp =>
      'Fälligkeitsdaten, Intervalle und Wiederholungsverlauf. Ohne sie ist jede Karte neu.';

  @override
  String get includePresets => 'Stapelvoreinstellungen einschließen';

  @override
  String get includePresetsHelp => 'Die Stapeloptionen';

  @override
  String get includeMedia => 'Medien einschließen';

  @override
  String get includeMediaHelp => 'Bilder, Audio und Video';

  @override
  String get legacyAnki => 'Ältere Anki-Versionen unterstützen';

  @override
  String get legacyAnkiHelp =>
      'Für Anki 2.1.49 und älter. Der Import ist langsamer.';

  @override
  String get secToday => 'Heute';

  @override
  String get secStreaks => 'Serien';

  @override
  String get secFutureDue => 'Künftig fällig';

  @override
  String get secCalendar => 'Kalender';

  @override
  String get secReviews => 'Wiederholungen';

  @override
  String get secReviewTime => 'Wiederholungszeit';

  @override
  String get secCounts => 'Kartenanzahl';

  @override
  String get secIntervals => 'Wiederholungsintervalle';

  @override
  String get secStability => 'Stabilität';

  @override
  String get secDifficulty => 'Schwierigkeit';

  @override
  String get secRetrievability => 'Abrufbarkeit';

  @override
  String get secHourly => 'Nach Stunden';

  @override
  String get secButtons => 'Antwortknöpfe';

  @override
  String get secTrueRetention => 'Tatsächliche Behaltensrate';

  @override
  String get period1Month => '1 Monat';

  @override
  String get period3Months => '3 Monate';

  @override
  String get period1Year => '1 Jahr';

  @override
  String get periodAll => 'Gesamter Verlauf';

  @override
  String get periodAllShort => 'Alle';

  @override
  String get noCardsStudiedToday => 'Heute wurden noch keine Karten gelernt.';

  @override
  String studiedToday(int count, String time) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return 'Heute $_temp0 in $time gelernt.';
  }

  @override
  String againCount(String count, String percent) {
    return 'Anzahl Nochmal: $count ($percent richtig)';
  }

  @override
  String todayKinds(
    String learning,
    String review,
    String relearning,
    String filtered,
  ) {
    return 'Lernen: $learning, Wiederholung: $review, Neulernen: $relearning, Gefiltert: $filtered';
  }

  @override
  String get today => 'Heute';

  @override
  String get tomorrow => 'Morgen';

  @override
  String get overdue => 'Überfällig';

  @override
  String get dueTomorrowLabel => 'Morgen fällig';

  @override
  String get total => 'Gesamt';

  @override
  String get average => 'Durchschnitt';

  @override
  String cardsPerDay(String count) {
    return '$count Karten/Tag';
  }

  @override
  String forecastTip(String day, int count, String total) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Karten',
      one: '1 Karte',
    );
    return '$day: $_temp0\nSumme bis hier: $total';
  }

  @override
  String get runningTotal => 'Laufende Summe';

  @override
  String get daysStudiedYear => 'Lerntage im letzten Jahr';

  @override
  String countOf(String count, String total) {
    return '$count von $total';
  }

  @override
  String get mostReviewsDay => 'Meiste Wiederholungen an einem Tag';

  @override
  String get currentStreak => 'Aktuelle Serie';

  @override
  String get longestStreak => 'Längste Serie';

  @override
  String longestStreakValue(String days, String first, String last) {
    return '$days, vom $first bis $last';
  }

  @override
  String get studyDaysInPeriod => 'Lerntage';

  @override
  String get goalDaysMet => 'Tage mit erreichtem Ziel';

  @override
  String minutesAmount(String minutes) {
    return '$minutes Min.';
  }

  @override
  String minutesAxis(String minutes) {
    return '$minutes Min.';
  }

  @override
  String get daysStudied => 'Lerntage';

  @override
  String get averageDaysStudied => 'Durchschnitt an Lerntagen';

  @override
  String get averageOverPeriod => 'Durchschnitt über den Zeitraum';

  @override
  String get kindLearning => 'Lernen';

  @override
  String get kindYoung => 'Jung';

  @override
  String get kindMature => 'Ausgereift';

  @override
  String get kindRelearning => 'Neulernen';

  @override
  String get kindFiltered => 'Gefiltert';

  @override
  String get kindSuspended => 'Ausgesetzt';

  @override
  String get kindBuried => 'Zurückgestellt';

  @override
  String bucketDays(String days) {
    return '$days T';
  }

  @override
  String get reviewCards => 'Wiederholungskarten';

  @override
  String get averageInterval => 'Durchschnittliches Intervall';

  @override
  String get averageStability => 'Durchschnittliche Stabilität';

  @override
  String get averageDifficulty => 'Durchschnittliche Schwierigkeit';

  @override
  String get averageRetrievability => 'Durchschnittliche Abrufbarkeit';

  @override
  String get estimatedKnowledge => 'Geschätztes Gesamtwissen';

  @override
  String get noCardsInfo => 'Noch keine Karten mit diesen Angaben.';

  @override
  String get noReviewsPeriod => 'Keine Wiederholungen in diesem Zeitraum.';

  @override
  String hourTip(String hours, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
    );
    return '$hours\n$_temp0';
  }

  @override
  String hourTipCorrect(String percent) {
    return ', $percent richtig';
  }

  @override
  String get successRate => 'Erfolgsquote';

  @override
  String get noAnswers => 'Keine Antworten';

  @override
  String percentCorrect(String percent, String count) {
    return '$percent richtig ($count)';
  }

  @override
  String get eachGroupShows =>
      'Jede Gruppe zeigt Nochmal, Schwer, Gut und Einfach.';

  @override
  String get yesterday => 'Gestern';

  @override
  String get lastWeek => 'Letzte Woche';

  @override
  String get lastMonth => 'Letzter Monat';

  @override
  String get lastYear => 'Letztes Jahr';

  @override
  String get allTime => 'Gesamte Zeit';

  @override
  String get desiredRetention => 'Gewünschte Behaltensrate';

  @override
  String calendarTip(String date, int count) {
    final intl.NumberFormat countNumberFormat =
        intl.NumberFormat.decimalPattern(localeName);
    final String countString = countNumberFormat.format(count);

    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$countString Wiederholungen',
      one: '1 Wiederholung',
    );
    return '$date: $_temp0';
  }

  @override
  String get statsDeckFallback => 'Stapel';

  @override
  String statsSearchScope(String query) {
    return 'Suche: $query';
  }

  @override
  String get wholeCollection => 'Gesamte Sammlung';

  @override
  String savedFile(String file) {
    return '$file gespeichert';
  }

  @override
  String get saveAsPdf => 'Als PDF speichern';

  @override
  String get collectionScope => 'Sammlung';

  @override
  String get searchScope => 'Suche';

  @override
  String get deckWithSubdecks => 'Stapel (mit Unterstapeln)';

  @override
  String get statsSearchHint => 'z. B. deck:Spanisch tag:Verben';

  @override
  String get showStatistics => 'Statistiken zeigen';

  @override
  String statsTitleScope(String scope) {
    return 'Statistiken - $scope';
  }

  @override
  String statsPdfName(String scope, String date) {
    return 'Statistiken - $scope - $date.pdf';
  }

  @override
  String statsTimeBased(String scope, String period) {
    return '$scope - zeitbezogene Abschnitte: $period';
  }

  @override
  String createdAt(String date) {
    return 'Erstellt am $date';
  }

  @override
  String get hour => 'Stunde';

  @override
  String get correct => 'Richtig';

  @override
  String get profile => 'Profil';

  @override
  String get profiles => 'Profile';

  @override
  String get switchProfile => 'Profil wechseln';

  @override
  String get showProfilesAtStartup => 'Profilliste beim Start anzeigen';

  @override
  String get showProfilesAtStartupHelp =>
      'Bei jedem Start von Cardflip ein Profil wählen, statt das zuletzt verwendete zu öffnen.';

  @override
  String get addProfile => 'Profil hinzufügen';

  @override
  String get addProfileHelp =>
      'Ein Profil hat eigene Karten, Medien, Einstellungen und Sicherungen.';

  @override
  String get renameProfile => 'Profil umbenennen';

  @override
  String get deleteProfile => 'Profil löschen';

  @override
  String deleteProfileBody(int cards) {
    String _temp0 = intl.Intl.pluralLogic(
      cards,
      locale: localeName,
      other: '$cards Karten',
      one: '1 Karte',
      zero: 'keinen Karten',
    );
    return 'Dadurch wird das Profil mit $_temp0 und allen Medien, Einstellungen und Sicherungen gelöscht. Das kann nicht rückgängig gemacht werden.';
  }

  @override
  String get deleteProfileBodyUncounted =>
      'Dadurch wird das Profil mit allen Karten, Medien, Einstellungen und Sicherungen gelöscht. Das kann nicht rückgängig gemacht werden.';

  @override
  String get profileIsOpen => 'Gerade geöffnet';

  @override
  String profileDeleted(String name) {
    return 'Profil „$name“ gelöscht';
  }

  @override
  String get sync => 'Synchronisierung';

  @override
  String get syncDeviceName => 'Gerätename';

  @override
  String get syncDeviceNameHelper =>
      '1 bis 40 Zeichen. Deine gekoppelten Geräte sehen diesen Namen.';

  @override
  String get syncRenameDevice => 'Dieses Gerät umbenennen';

  @override
  String get syncCurrentNetwork => 'Aktuelles Netzwerk';

  @override
  String syncWiredNetwork(String subnet) {
    return 'Kabelnetzwerk $subnet';
  }

  @override
  String syncSubnetNetwork(String subnet) {
    return 'Netzwerk $subnet';
  }

  @override
  String get syncMobileNetwork => 'Mobile Daten';

  @override
  String get syncNoNetwork => 'Kein Netzwerk';

  @override
  String get syncNetworkTrusted =>
      'Vertrauenswürdig: Cardflip synchronisiert hier mit deinen gekoppelten Geräten.';

  @override
  String get syncNetworkNotTrusted =>
      'Nicht vertrauenswürdig: Cardflip synchronisiert hier nicht und macht dieses Gerät nicht sichtbar.';

  @override
  String get syncMobileNeverTrusted =>
      'Cardflip synchronisiert nie über mobile Daten.';

  @override
  String get trustThisNetwork => 'Diesem Netzwerk vertrauen';

  @override
  String get stopTrusting => 'Nicht mehr vertrauen';

  @override
  String get trustNetworkTitle => 'Diesem Netzwerk vertrauen?';

  @override
  String get trustNetworkBody =>
      'In einem vertrauenswürdigen Netzwerk macht Cardflip dieses Gerät unter seinem Namen sichtbar und synchronisiert mit deinen gekoppelten Geräten. Andere Geräte im Netzwerk sehen den Namen, aber nur gekoppelte Geräte können Karten mit ihm austauschen.\n\nVertraue nur privaten Netzwerken wie deinem WLAN zu Hause: Einem anderen Netzwerk mit demselben Namen wird ebenfalls vertraut.';

  @override
  String get trustNetworkLocation =>
      'Um das Netzwerk an seinem WLAN-Namen zu erkennen, fragt Cardflip als Nächstes nach der Standortberechtigung. Cardflip verwendet deinen Standort nicht. Ohne die Berechtigung wird das Netzwerk an seinem Adressbereich erkannt.';

  @override
  String get trust => 'Vertrauen';

  @override
  String get trustedNetworks => 'Vertrauenswürdige Netzwerke';

  @override
  String get noTrustedNetworks => 'Noch keine vertrauenswürdigen Netzwerke';

  @override
  String get pairedDevices => 'Gekoppelte Geräte';

  @override
  String get noPairedDevices => 'Noch keine gekoppelten Geräte';

  @override
  String get deviceReachable => 'Erreichbar';

  @override
  String get deviceNotReachable => 'Nicht erreichbar';

  @override
  String lastSyncedAt(String time) {
    return 'Zuletzt synchronisiert: $time';
  }

  @override
  String get neverSynced => 'Noch nicht synchronisiert';

  @override
  String get unpair => 'Entkoppeln';

  @override
  String unpairTitle(String device) {
    return '$device entkoppeln?';
  }

  @override
  String unpairBody(String device) {
    return '$device kann nicht mehr mit diesem Gerät synchronisieren, und die Verknüpfungen deiner Profile damit werden entfernt. Die Karten auf beiden Geräten bleiben, wie sie sind.';
  }

  @override
  String get pairDevice => 'Gerät koppeln';

  @override
  String get pairDeviceHelp =>
      'Ein weiteres deiner Geräte in diesem Netzwerk verbinden';

  @override
  String get pairDeviceIntro =>
      'Öffne „Gerät koppeln“ auch auf dem anderen Gerät, im selben Netzwerk. Wähle es dann unten aus oder gib die Adresse ein, die es anzeigt.';

  @override
  String visibleForPairing(String time) {
    return 'Noch $time zum Koppeln sichtbar';
  }

  @override
  String get notVisibleForPairing => 'Nicht zum Koppeln sichtbar';

  @override
  String get makeVisible => 'Sichtbar machen';

  @override
  String get thisDeviceAddress => 'Adresse dieses Geräts';

  @override
  String get devicesNearby => 'Geräte in der Nähe';

  @override
  String get lookingForDevices => 'Suche nach Geräten …';

  @override
  String get enterAddress => 'Adresse des anderen Geräts';

  @override
  String get enterAddressHelper =>
      'Wie auf dem anderen Gerät angezeigt, etwa 192.168.1.20:47812';

  @override
  String get pair => 'Koppeln';

  @override
  String pairingRequestTitle(String device) {
    return 'Mit $device koppeln?';
  }

  @override
  String pairingRequestBody(String device) {
    return '$device möchte sich mit diesem Gerät koppeln. Nimm nur an, wenn es eines deiner eigenen Geräte ist.';
  }

  @override
  String get accept => 'Annehmen';

  @override
  String get decline => 'Ablehnen';

  @override
  String get confirmCodeTitle => 'Codes vergleichen';

  @override
  String confirmCodeBody(String device) {
    return 'Prüfe, ob $device denselben Code anzeigt:';
  }

  @override
  String get codesMatch => 'Codes stimmen überein';

  @override
  String get codesDiffer => 'Codes unterscheiden sich';

  @override
  String waitingForConfirmation(String device) {
    return 'Warte auf die Bestätigung von $device …';
  }

  @override
  String pairedWith(String device) {
    return 'Mit $device gekoppelt';
  }

  @override
  String pairingFailed(String device) {
    return 'Kopplung mit $device fehlgeschlagen';
  }

  @override
  String get profileLinks => 'Verknüpfungen dieses Profils';

  @override
  String get noProfileLinks =>
      'Dieses Profil wird noch mit keinem anderen Gerät synchronisiert.';

  @override
  String get linkProfile => 'Dieses Profil verknüpfen';

  @override
  String get linkProfileHelp =>
      'Mit einem Profil auf einem gekoppelten Gerät synchronisieren';

  @override
  String linkTarget(String profile, String device) {
    return '$profile auf $device';
  }

  @override
  String get chooseDevice => 'Gerät wählen';

  @override
  String chooseRemoteProfile(String device) {
    return 'Profil auf $device';
  }

  @override
  String newProfileNamed(String name) {
    return 'Neues Profil „$name“';
  }

  @override
  String get linkNeedsChoice =>
      'Deine Wahl ist nötig: zusammenführen oder ersetzen';

  @override
  String get unlink => 'Trennen';

  @override
  String unlinkTitle(String device) {
    return 'Von $device trennen?';
  }

  @override
  String unlinkBody(String profile, String device) {
    return 'Dieses Profil wird nicht mehr mit „$profile“ auf $device synchronisiert. Beide behalten ihre Karten.';
  }

  @override
  String get syncSyncing => 'Synchronisiere …';

  @override
  String syncProgress(int received, int sent) {
    return 'Synchronisiere: $received empfangen, $sent gesendet';
  }

  @override
  String syncedAt(String time) {
    return 'Synchronisiert $time';
  }

  @override
  String get syncNotTrusted => 'Kein vertrauenswürdiges Netzwerk';

  @override
  String get syncNeedsChoice => 'Synchronisierung braucht deine Wahl';

  @override
  String get syncFailed => 'Synchronisierung fehlgeschlagen';

  @override
  String get syncAgain => 'Erneut synchronisieren';

  @override
  String get syncNow => 'Jetzt synchronisieren';

  @override
  String syncSummary(int received, int sent) {
    return 'Synchronisiert: $received empfangen, $sent gesendet';
  }

  @override
  String syncSummaryMedia(int received, int sent, int media) {
    return 'Synchronisiert: $received empfangen, $sent gesendet, $media Mediendateien';
  }

  @override
  String get syncNothingReachable =>
      'Kein verknüpftes Gerät ist in diesem Netzwerk erreichbar.';

  @override
  String firstSyncTitle(String device) {
    return 'Mit $device synchronisieren';
  }

  @override
  String firstSyncBody(String profile, String device) {
    return 'Dieses Profil und „$profile“ auf $device wurden noch nicht synchronisiert, oder eines davon wurde seitdem ersetzt. Wie sollen sie zusammengeführt werden? Vor jeder Änderung wird eine Sicherung erstellt.';
  }

  @override
  String get firstSyncMerge => 'Zusammenführen';

  @override
  String get firstSyncMergeHelp =>
      'Die Karten und den Lernverlauf beider behalten.';

  @override
  String firstSyncReplaceOther(String device) {
    return 'Das Profil auf $device ersetzen';
  }

  @override
  String firstSyncReplaceOtherHelp(String device) {
    return '$device erhält genau die Karten und den Verlauf dieses Profils.';
  }

  @override
  String get firstSyncReplaceThis => 'Dieses Profil ersetzen';

  @override
  String firstSyncReplaceThisHelp(String profile, String device) {
    return 'Dieses Gerät erhält genau die Karten und den Verlauf von „$profile“ auf $device.';
  }

  @override
  String get decideLater => 'Später';
}
