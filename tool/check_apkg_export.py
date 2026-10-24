#!/usr/bin/env python3
"""Imports Cardflip's exported Anki packages into current Anki and checks them.

Writes sample packages of the importer's fixture collection (modern.apkg
imported with scheduling) with the Go test TestWriteSamples, then imports each
into a fresh collection with the official `anki` package (pip install anki):
deck packages with Anki's package importer, collection packages as Anki opens
them. It compares notes, cards, scheduling, review history, presets and media
with Anki's own legacy.apkg fixture, and runs Check Database, which must find
nothing to fix.

It runs with current Anki and with Anki 2.1.66, the last version before
FSRS, whose import has no options (it always keeps scheduling) and which has
no memory states to compare:

    tool/check_apkg_export.py   exit 0 when every package imports correctly

    uv venv --python 3.11 /tmp/anki2166
    VIRTUAL_ENV=/tmp/anki2166 uv pip install anki==2.1.66
    /tmp/anki2166/bin/python tool/check_apkg_export.py
"""

from __future__ import annotations

import datetime
import json
import os
import sqlite3
import subprocess
import sys
import tempfile
import zipfile
from pathlib import Path

import anki.collection  # noqa: F401  (must load before anki.cards)
from anki import deck_config_pb2
from anki._backend import RustBackend
from anki.buildinfo import version as ANKI_VERSION
from anki.collection import Collection

# Anki 23.10 added FSRS, and options to the package import.
FSRS = hasattr(anki.collection, "ImportAnkiPackageOptions")
# Anki 23.12 added auto advance.
AUTO_ADVANCE = "seconds_to_show_question" in deck_config_pb2.DeckConfig.Config.DESCRIPTOR.fields_by_name

ROOT = Path(__file__).resolve().parent.parent
REFERENCE = ROOT / "golang" / "importer" / "apkg" / "testdata" / "legacy.apkg"

failures: list[str] = []


def check(ok: bool, what: str) -> None:
    if not ok:
        failures.append(what)
        print(f"  FAIL {what}")


def reference(tmp: Path) -> dict:
    """Notes, cards, review log, presets and media of Anki's legacy.apkg."""
    out = tmp / "reference"
    with zipfile.ZipFile(REFERENCE) as z:
        z.extractall(out)
    db = sqlite3.connect(out / "collection.anki21")
    guids = dict(db.execute("select id, guid from notes"))
    cards = {}
    card_key = {}
    for cid, nid, ord_, typ, queue, ivl, flags, data in db.execute(
        "select id, nid, ord, type, queue, ivl, flags, data from cards"
    ):
        key = (guids[nid], ord_)
        card_key[cid] = key
        cards[key] = dict(type=typ, queue=queue, ivl=ivl, flag=flags & 7, data=json.loads(data or "{}"))
    revlog: dict = {}
    for rid, cid, ease, ivl in db.execute("select id, cid, ease, ivl from revlog order by id"):
        revlog.setdefault(card_key[cid], []).append((rid, ease, ivl))
    dconf = json.loads(db.execute("select dconf from col").fetchone()[0])
    decks = json.loads(db.execute("select decks from col").fetchone()[0])
    media_map = json.loads((out / "media").read_text())
    media = {name: (out / key).read_bytes() for key, name in media_map.items()}
    spanish = next(d for d in decks.values() if d["name"] == "Fixture::Spanish")
    return dict(notes=set(guids.values()), cards=cards, revlog=revlog, preset=dconf[str(spanish["conf"])], media=media)


def write_samples(tmp: Path) -> Path:
    out = tmp / "samples"
    out.mkdir()
    subprocess.run(
        ["go", "test", "./exporter/apkg", "-run", "TestWriteSamples", "-count=1"],
        cwd=ROOT / "golang",
        env={**os.environ, "CARDFLIP_EXPORT_SAMPLES": str(out)},
        check=True,
        stdout=subprocess.DEVNULL,
    )
    return out


def collection_view(col: Collection) -> dict:
    cards = {}
    for cid in col.find_cards(""):
        c = col.get_card(cid)
        key = (c.note().guid, c.ord)
        cards[key] = c
    revlog: dict = {}
    for rid, cid, ease, ivl in col.db.all("select id, cid, ease, ivl from revlog order by id"):
        c = col.get_card(cid)
        revlog.setdefault((c.note().guid, c.ord), []).append((rid, ease, ivl))
    return dict(cards=cards, revlog=revlog)


def check_content(col: Collection, ref: dict, scheduling: bool) -> None:
    v = collection_view(col)
    guids = {col.get_note(nid).guid for nid in col.find_notes("")}
    check(guids == ref["notes"], f"notes {sorted(guids)}")
    check(set(v["cards"]) == set(ref["cards"]), f"cards {sorted(v['cards'])}")
    names = {d.name for d in col.decks.all_names_and_ids()}
    check({"Fixture", "Fixture::Spanish", "Fixture::Geo"} <= names, f"decks {sorted(names)}")
    for name, data in ref["media"].items():
        path = Path(col.media.dir()) / name
        check(path.exists() and path.read_bytes() == data, f"media file {name}")

    if not scheduling:
        check(all(c.type == 0 and c.queue == 0 for c in v["cards"].values()), "every card is new")
        check(not v["revlog"], "no review history")
        return

    for key, want in ref["cards"].items():
        c = v["cards"].get(key)
        if c is None:
            continue
        check(c.type == want["type"] and c.queue == want["queue"], f"{key} type/queue {c.type}/{c.queue}, want {want['type']}/{want['queue']}")
        check(c.flags & 7 == want["flag"], f"{key} flag {c.flags & 7}")
        if want["type"] == 2:
            check(c.ivl == want["ivl"], f"{key} ivl {c.ivl}, want {want['ivl']}")
        if FSRS and "s" in want["data"]:
            m = c.memory_state
            check(
                m is not None and abs(m.stability - want["data"]["s"]) < 0.01 and abs(m.difficulty - want["data"]["d"]) < 0.01,
                f"{key} memory state {m}",
            )
    # Review due on the fixture's due day: crt 2026-01-01 + 300 days.
    hola = v["cards"][("cf-basic-hola", 0)]
    due_date = datetime.date.fromtimestamp(col.crt + 4 * 3600) + datetime.timedelta(days=hola.due)
    check(due_date == datetime.date(2026, 10, 28), f"cf-basic-hola due on {due_date}")
    learning = v["cards"][("cf-cloze-paris", 0)]
    check(learning.due == 1790000000, f"learning card due {learning.due}")

    for key, want in ref["revlog"].items():
        got = v["revlog"].get(key, [])
        check([r[:2] for r in got] == [r[:2] for r in want], f"{key} review log ids/eases {got}")
        check(all(g[2] == w[2] for g, w in zip(got, want) if w[2] > 0), f"{key} review intervals")


AUTO_ADVANCE_KEYS = [("secondsToShowQuestion",), ("secondsToShowAnswer",), ("questionAction",), ("answerAction",)]


def check_preset(col: Collection, ref: dict) -> None:
    deck = col.decks.by_name("Fixture::Spanish")
    conf = col.decks.config_dict_for_deck_id(deck["id"])
    want = ref["preset"]
    check(conf["name"] == want["name"], f"preset name {conf['name']}")
    for path in [("new", "perDay"), ("new", "delays"), ("rev", "perDay"), ("rev", "maxIvl"), ("lapse", "delays"),
                 ("lapse", "minInt"), ("lapse", "leechFails"), ("lapse", "leechAction"), ("desiredRetention",),
                 ("newMix",), ("interdayLearningMix",), ("reviewOrder",), ("newSortOrder",), ("newGatherPriority",),
                 ("buryInterdayLearning",), ("autoplay",), ("replayq",)] + (AUTO_ADVANCE_KEYS if AUTO_ADVANCE else []):
        g, w = conf, want
        for p in path:
            g, w = g[p], w[p]
        check(g == w if not isinstance(w, float) else abs(g - w) < 1e-6, f"preset {'.'.join(path)} = {g}, want {w}")


def check_database(col: Collection) -> None:
    problems, ok = col.fix_integrity()
    lines = [line for line in problems.splitlines() if line and "rebuilt and optimized" not in line]
    check(ok and not lines, f"Check Database: {problems!r}")


def import_apkg(tmp: Path, name: str, path: Path, ref: dict, scheduling: bool) -> None:
    print(name)
    col = Collection(str(tmp / f"{name}.anki2"))
    try:
        if FSRS:
            from anki.collection import ImportAnkiPackageOptions, ImportAnkiPackageRequest

            request = ImportAnkiPackageRequest(
                package_path=str(path),
                options=ImportAnkiPackageOptions(with_scheduling=True, with_deck_configs=True),
            )
            log = col.import_anki_package(request).log
        else:
            log = col.import_anki_package(str(path)).log
        check(len(log.new) == len(ref["notes"]), f"{len(log.new)} notes imported")
        check(not (log.conflicting or log.first_field_match or log.missing_notetype), "import conflicts")
        check_content(col, ref, scheduling)
        if scheduling:
            check_preset(col, ref)
        check_database(col)
    finally:
        col.close()


def import_colpkg(tmp: Path, name: str, path: Path, ref: dict) -> None:
    print(name)
    target = tmp / name
    target.mkdir()
    col_path = target / "collection.anki2"
    media = target / "collection.media"
    RustBackend().import_collection_package(
        col_path=str(col_path), backup_path=str(path), media_folder=str(media), media_db=str(target / "collection.media.db2")
    )
    col = Collection(str(col_path))
    try:
        check_content(col, ref, True)
        check_preset(col, ref)
        if FSRS:
            check(bool(col.get_config("fsrs", False)), "FSRS is on")
        check_database(col)
    finally:
        col.close()


def main() -> int:
    print(f"Anki {ANKI_VERSION}")
    with tempfile.TemporaryDirectory() as t:
        tmp = Path(t)
        ref = reference(tmp)
        samples = write_samples(tmp)
        import_apkg(tmp, "deck.apkg", samples / "deck.apkg", ref, scheduling=False)
        import_apkg(tmp, "deck-scheduling.apkg", samples / "deck-scheduling.apkg", ref, scheduling=True)
        import_apkg(tmp, "deck-legacy.apkg", samples / "deck-legacy.apkg", ref, scheduling=True)
        import_colpkg(tmp, "collection.colpkg", samples / "collection.colpkg", ref)
        import_colpkg(tmp, "collection-legacy.colpkg", samples / "collection-legacy.colpkg", ref)
    if failures:
        print(f"\n{len(failures)} check(s) failed.")
        return 1
    print("\nEvery package imports into Anki correctly.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
