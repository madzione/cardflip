#!/usr/bin/env python3
"""Builds the Anki .apkg and .colpkg fixtures in golang/importer/apkg/testdata/.

Uses the official `anki` Python package (pip install anki) to build one small
collection that covers every case the importer must handle, then exports it as
a modern (collection.anki21b) and a legacy (collection.anki21) package, and
writes a truncated copy of the modern one. It then adds an empty deck and a
filtered deck to the collection and exports the whole collection as a modern
and a legacy collection package (.colpkg), with a truncated copy of the
modern one. Separate small collections give the update packages (one deck at
three points in time), the LaTeX packages (pre-generated LaTeX images) and
the Image Occlusion packages.

Anki assigns note/card IDs and modification times from the clock, so the .apkg
bytes differ between runs. Everything else (GUIDs, crt, due values, memory
states, media) is fixed. fixtures.json records that normalised content, read
straight from the generated packages; the Go tests use it as expected values.

    tool/make_apkg_fixtures.py          regenerate fixtures and fixtures.json
    tool/make_apkg_fixtures.py --check  regenerate into a temp dir and compare
                                        its content with the committed fixtures
"""

from __future__ import annotations

import argparse
import hashlib
import io
import json
import os
import shutil
import sqlite3
import struct
import sys
import tempfile
import zipfile
import zlib
from compression import zstd
from pathlib import Path

import anki.collection  # noqa: F401  (must load before anki.cards)
from anki import deck_config_pb2, notetypes_pb2
from anki.cards import FSRSMemoryState
from anki.collection import Collection, DeckIdLimit, ExportAnkiPackageOptions

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / "golang" / "importer" / "apkg" / "testdata"

# 2026-01-01 04:00:00 UTC. Review due values are days since this time.
CRT = 1767240000
# A fixed absolute time for the learning card (epoch seconds).
LEARN_DUE = 1790000000

CSS = """.card {
    font-family: arial;
    font-size: 24px;
    text-align: center;
    color: #222222;
    background-color: #fdf6e3;
}
.cloze { font-weight: bold; color: blue; }
b { color: #aa0000; }
"""


def png(rgb: tuple[int, int, int], size: int = 8, height: int | None = None) -> bytes:
    """A small solid-colour PNG, built by hand so the bytes never change;
    square unless height is given."""

    def chunk(kind: bytes, data: bytes) -> bytes:
        body = kind + data
        return struct.pack(">I", len(data)) + body + struct.pack(">I", zlib.crc32(body))

    height = size if height is None else height
    row = b"\x00" + bytes(rgb) * size
    raw = row * height
    return (
        b"\x89PNG\r\n\x1a\n"
        + chunk(b"IHDR", struct.pack(">IIBBBBB", size, height, 8, 2, 0, 0, 0))
        + chunk(b"IDAT", zlib.compress(raw, 9))
        + chunk(b"IEND", b"")
    )


def jpeg(rgb: tuple[int, int, int], size: int = 8) -> bytes:
    from PIL import Image

    buf = io.BytesIO()
    Image.new("RGB", (size, size), rgb).save(buf, format="JPEG", quality=90)
    return buf.getvalue()


MEDIA = {
    "dog.png": png((200, 120, 40)),
    "cat.jpg": jpeg((90, 90, 90)),
    # Not real audio: the importer only has to skip these with a warning.
    "hola.mp3": b"ID3\x03\x00\x00\x00\x00\x00\x00fake-mp3-hola",
    "adios.mp3": b"ID3\x03\x00\x00\x00\x00\x00\x00fake-mp3-adios",
}

# (guid, notetype, deck, fields, tags)
NOTES = [
    ("cf-basic-hola", "Basic", "Fixture::Spanish", ["hola", "hello"], ["verb", "spanish"]),
    ("cf-basic-dog", "Basic", "Fixture::Spanish", ['perro<br><img src="dog.png">', "dog"], ["noun"]),
    ("cf-basic-cat", "Basic", "Fixture::Spanish", ["gato", '<img src="cat.jpg"> <b>cat</b>'], ["noun"]),
    ("cf-rev-comer", "Basic (and reversed card)", "Fixture::Spanish", ["comer", "to eat"], ["verb"]),
    ("cf-cloze-paris", "Cloze", "Fixture::Geo", ["{{c1::Paris::city}} is the capital of {{c2::France}}", "Europe"], []),
    ("cf-audio-hola", "Basic", "Fixture::Spanish", ["[sound:hola.mp3] hola (audio)", "hello"], ["audio"]),
    ("cf-audio-adios", "Basic", "Fixture::Spanish", ["adiós [sound:adios.mp3]", "goodbye"], ["audio"]),
    # One deletion on two cards: Anki creates cards 1, 2 and 3.
    ("cf-cloze-multi", "Cloze", "Fixture::Geo", ["{{c1,2::Paris}} is in {{c3::France}}", ""], []),
]

# Scheduling applied to specific cards: (guid, card ord) -> settings.
SCHED = {
    # Review card with an FSRS memory state, due 300 days after crt.
    ("cf-basic-hola", 0): dict(type=2, queue=2, due=300, ivl=15, factor=2500, reps=4, lapses=1,
                               memory=(14.25, 5.5)),
    # Suspended review card without a memory state (estimated from ivl/ease).
    ("cf-basic-cat", 0): dict(type=2, queue=-1, due=310, ivl=20, factor=2300, reps=5, lapses=0),
    # Learning card, due at an absolute epoch time.
    ("cf-cloze-paris", 0): dict(type=1, queue=1, due=LEARN_DUE, ivl=0, factor=0, reps=1, lapses=0, left=1001),
}

# A deck options preset of its own for Fixture::Spanish, with every option
# the importer maps set away from Anki's defaults (legacy dconf keys). The
# other decks use Anki's default preset.
SPANISH_PRESET = "Fixture Spanish"
SPANISH_CONFIG = {
    ("new", "perDay"): 10,
    ("rev", "perDay"): 150,
    ("new", "delays"): [2.0, 15.0],
    ("lapse", "delays"): [5.0],
    ("rev", "maxIvl"): 3650,
    ("lapse", "minInt"): 2,
    ("lapse", "leechFails"): 6,
    ("lapse", "leechAction"): 0,  # suspend
    ("new", "bury"): True,
    ("rev", "bury"): False,
    ("buryInterdayLearning",): True,
    ("autoplay",): False,
    ("replayq",): False,
    ("desiredRetention",): 0.85,
    ("newMix",): 2,  # new cards before reviews
    ("interdayLearningMix",): 1,  # interday learning after reviews
    ("reviewOrder",): 3,  # intervals ascending
    ("newSortOrder",): 3,  # random note, then card type
    ("newGatherPriority",): 5,  # deck, then random notes
    ("secondsToShowQuestion",): 5.0,
    ("secondsToShowAnswer",): 3.0,
    ("questionAction",): 1,  # show reminder
    ("answerAction",): 2,  # answer Good
}

# Flags set on specific cards: (guid, card ord) -> flag (3 is green). The
# reverse card of the note stays unflagged.
FLAGS = {
    ("cf-rev-comer", 0): 3,
}

# Added after the .apkg exports, for the collection packages only: a deck
# without cards, and a filtered deck that pulls the unsuspended nouns of
# Fixture::Spanish (cf-basic-dog) out of their home deck.
EMPTY_DECK = "Inbox::Later"
FILTERED_DECK = "Filtered Nouns"
FILTER_SEARCH = '"deck:Fixture::Spanish" tag:noun'


# Three versions of a small shared deck, for re-import with an update
# policy. Modification times are fixed (epoch seconds, 2026-03-01, -04-01 and
# -05-01 00:00 UTC), so "if newer" comparisons are stable.
UPDATE_DECK = "Update Fixture"
UPDATE_TYPE = "Update Vocab"
UPDATE_T1, UPDATE_T2, UPDATE_T3 = 1772323200, 1775001600, 1777593600
UPDATE_CSS = ".card { font-size: 20px; }\n"
UPDATE_CSS_V2 = ".card { font-size: 28px; color: #333; }\n"
# (guid, fields) of version 1. Version 2 fixes the typo in cf-up-hola and
# restyles the note type; version 3 adds the field Audio (shown on the
# back), gives cf-up-hola an Audio value and adds cf-up-perro.
UPDATE_NOTES = [
    ("cf-up-hola", ["hola", "helo"]),
    ("cf-up-gato", ["gato", "cat"]),
]


# A deck with pre-generated LaTeX images, as shared decks ship them. One note
# type keeps Anki's PNG images, the other is set to SVG. Without a TeX
# installation the images are placeholders, but their names come from Anki,
# which also picks them for the export. The image of cf-latex-missing is
# left out of the media.
LATEX_DECK = "Latex Fixture"
LATEX_PNG_TYPE = "Latex PNG"
LATEX_SVG_TYPE = "Latex SVG"
# (guid, notetype, fields)
LATEX_NOTES = [
    ("cf-latex-png", LATEX_PNG_TYPE, ["[latex]a<br>b &amp; <b>c</b>[/latex]", r"[$]x^2[/$] and [$$]\sum_i i[/$$]"]),
    ("cf-latex-missing", LATEX_PNG_TYPE, [r"[latex]\LaTeX[/latex]", "missing"]),
    ("cf-latex-svg", LATEX_SVG_TYPE, ["[$]E = mc^2[/$]", r"[$$]\frac{1}{2}&nbsp;<i>x</i>[/$$]"]),
]
LATEX_MISSING = "cf-latex-missing"

# A deck of Image Occlusion notes, made with Anki's own Image Occlusion
# note type and add_image_occlusion_note, as its mask editor does. The
# shapes are in Anki's syntax, coordinates relative to the image (0-1).
# cf-io-all is "Hide all, guess one" (oi=1 on every shape): a rectangle, an
# ellipse, a polygon grouped with a rectangle (both c3, asked together) and
# a text label, which is always shown and makes no card (c0). cf-io-one is
# "Hide one, guess one" (no oi): two rectangles.
OCCLUSION_DECK = "Occlusion Fixture"
OCCLUSION_IMAGE = "diagram.png"
# (guid, occlusions, header, back extra, tags)
OCCLUSION_NOTES = [
    (
        "cf-io-all",
        "{{c1::image-occlusion:rect:left=.1:top=.2:width=.3:height=.1:oi=1}}<br>"
        "{{c2::image-occlusion:ellipse:left=.5:top=.5:rx=.1:ry=.05:angle=10:oi=1}}<br>"
        "{{c3::image-occlusion:polygon:points=.1,.6 .3,.6 .2,.8:oi=1}}<br>"
        "{{c3::image-occlusion:rect:left=.7:top=.7:width=.1:height=.1:oi=1}}<br>"
        "{{c0::image-occlusion:text:left=.6:top=.1:text=Label\\: heart:scale=1:fs=.05:oi=1}}",
        "Heart",
        "Anatomy",
        ["anatomy"],
    ),
    (
        "cf-io-one",
        "{{c1::image-occlusion:rect:left=.05:top=.05:width=.2:height=.2}}<br>"
        "{{c2::image-occlusion:rect:left=.6:top=.6:width=.3:height=.25}}",
        "",
        "",
        [],
    ),
]

# Name prefixes of the decks a deck package lists.
FIXTURE_DECKS = ("Fixture", UPDATE_DECK, LATEX_DECK, OCCLUSION_DECK)
LATEX_SVG = b'<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20"><rect width="40" height="20" fill="#3060c0"/></svg>\n'


def rev_id(day: float, minutes: float = 0) -> int:
    """A review log ID (epoch milliseconds): 06:00 UTC on day `day` after crt,
    plus `minutes`."""
    return int((CRT + day * 86400 + 2 * 3600 + minutes * 60) * 1000)


# Review history (Anki's revlog) per card: (guid, card ord) -> entries of
# (id, ease, ivl, lastIvl, factor, time ms, type). ease 0 is a manual entry;
# a negative interval is in seconds; type 0 learning, 1 review, 2 relearning,
# 3 filtered, 4 manual, 5 rescheduled. Anki refuses to export entries dated
# in the future, so all of them lie before 2026-09-22.
REVLOG = {
    # Twelve entries: learning, reviews, a lapse with relearning, a review in
    # a filtered deck that reschedules, and a set due date (rescheduled). The
    # card has an FSRS memory state, which the import keeps.
    ("cf-basic-hola", 0): [
        (rev_id(4), 3, -600, 0, 0, 8000, 0),
        (rev_id(4, 10), 3, 1, -600, 2500, 5000, 0),
        (rev_id(5), 3, 3, 1, 2500, 4000, 1),
        (rev_id(8), 3, 8, 3, 2500, 3500, 1),
        (rev_id(16), 1, -600, 8, 2300, 12000, 1),
        (rev_id(16, 10), 3, 1, -600, 2300, 6000, 2),
        (rev_id(17), 3, 3, 1, 2300, 4000, 1),
        (rev_id(20), 3, 7, 3, 2300, 3000, 3),
        (rev_id(27), 4, 20, 7, 2450, 2000, 1),
        (rev_id(47), 2, 25, 20, 2300, 9000, 1),
        (rev_id(60), 0, 40, 25, 2300, 0, 5),
        (rev_id(200), 3, 15, 40, 2300, 4500, 1),
    ],
    # Eight reviews under SM-2, without an FSRS memory state: the import
    # computes it from this history.
    ("cf-basic-cat", 0): [
        (rev_id(100), 3, -60, 0, 0, 7000, 0),
        (rev_id(100, 1), 3, -600, -60, 0, 5000, 0),
        (rev_id(100, 11), 3, 1, -600, 2500, 4000, 0),
        (rev_id(101), 3, 3, 1, 2500, 3000, 1),
        (rev_id(104), 3, 8, 3, 2500, 3000, 1),
        (rev_id(112), 2, 10, 8, 2350, 8000, 1),
        (rev_id(122), 3, 30, 10, 2350, 4000, 1),
        (rev_id(210), 2, 20, 30, 2300, 9500, 1),
    ],
    # Learning, a cram review in a filtered deck (no rescheduling, factor 0),
    # a reset to new (manual, interval 0), then learning again.
    ("cf-cloze-paris", 0): [
        (rev_id(250), 3, -600, 0, 0, 6000, 0),
        (rev_id(250, 5), 3, 0, 0, 0, 3000, 3),
        (rev_id(251), 0, 0, 0, 0, 0, 4),
        (rev_id(263), 1, -60, 0, 0, 10000, 0),
    ],
}


def build_collection(path: Path) -> Collection:
    col = Collection(str(path))
    col.crt = CRT
    media_dir = Path(col.media.dir())
    for name, data in MEDIA.items():
        (media_dir / name).write_bytes(data)

    for name in ("Basic", "Cloze"):
        nt = col.models.by_name(name)
        nt["css"] = CSS
        col.models.update_dict(nt)

    for guid, nt_name, deck_name, fields, tags in NOTES:
        nt = col.models.by_name(nt_name)
        deck_id = col.decks.id(deck_name)
        note = col.new_note(nt)
        for i, value in enumerate(fields):
            note.fields[i] = value
        note.tags = tags
        note.guid = guid
        col.add_note(note, deck_id)

    conf_id = col.decks.add_config_returning_id(SPANISH_PRESET)
    conf = col.decks.get_config(conf_id)
    for keys, value in SPANISH_CONFIG.items():
        target = conf
        for key in keys[:-1]:
            target = target[key]
        target[keys[-1]] = value
    col.decks.update_config(conf)
    spanish = col.decks.get(col.decks.id("Fixture::Spanish"))
    spanish["conf"] = conf_id
    col.decks.save(spanish)

    for (guid, ord_), s in SCHED.items():
        nid = col.db.scalar("select id from notes where guid = ?", guid)
        cid = col.db.scalar("select id from cards where nid = ? and ord = ?", nid, ord_)
        card = col.get_card(cid)
        card.type = s["type"]
        card.queue = s["queue"]
        card.due = s["due"]
        card.ivl = s["ivl"]
        card.factor = s["factor"]
        card.reps = s["reps"]
        card.lapses = s["lapses"]
        card.left = s.get("left", 0)
        if "memory" in s:
            stability, difficulty = s["memory"]
            card.memory_state = FSRSMemoryState(stability=stability, difficulty=difficulty)
        col.update_card(card)

    for (guid, ord_), flag in FLAGS.items():
        nid = col.db.scalar("select id from notes where guid = ?", guid)
        cid = col.db.scalar("select id from cards where nid = ? and ord = ?", nid, ord_)
        col.set_user_flag_for_cards(flag, [cid])

    for (guid, ord_), entries in REVLOG.items():
        nid = col.db.scalar("select id from notes where guid = ?", guid)
        cid = col.db.scalar("select id from cards where nid = ? and ord = ?", nid, ord_)
        for rid, ease, ivl, last_ivl, factor, time_ms, type_ in entries:
            col.db.execute(
                "insert into revlog (id, cid, usn, ease, ivl, lastIvl, factor, time, type)"
                " values (?, ?, -1, ?, ?, ?, ?, ?, ?)",
                rid, cid, ease, ivl, last_ivl, factor, time_ms, type_,
            )
    return col


def export(col: Collection, out: Path, legacy: bool) -> None:
    col.export_anki_package(
        out_path=str(out),
        options=ExportAnkiPackageOptions(
            with_scheduling=True, with_deck_configs=True, with_media=True, legacy=legacy
        ),
        limit=DeckIdLimit(deck_id=col.decks.id("Fixture")),
    )


def export_updates(path: Path, out: Path) -> None:
    """Writes update-v1.apkg (and its legacy twin update-v1-legacy.apkg),
    update-v2.apkg and update-v3.apkg: one deck at three points in time."""
    col = Collection(str(path))
    col.crt = CRT
    try:
        mm = col.models
        nt = mm.new(UPDATE_TYPE)
        for name in ("Front", "Back"):
            mm.add_field(nt, mm.new_field(name))
        t = mm.new_template("Card 1")
        t["qfmt"] = "{{Front}}"
        t["afmt"] = "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}"
        mm.add_template(nt, t)
        nt["css"] = UPDATE_CSS
        mm.add_dict(nt)
        deck_id = col.decks.id(UPDATE_DECK)
        for guid, fields in UPDATE_NOTES:
            note = col.new_note(mm.by_name(UPDATE_TYPE))
            for i, value in enumerate(fields):
                note.fields[i] = value
            note.guid = guid
            col.add_note(note, deck_id)

        def set_mods(type_mod: int, note_mods: dict[str, int]) -> None:
            # Anki stamps changes with the clock; pin them before exporting.
            col.db.execute("update notetypes set mtime_secs = ? where name = ?", type_mod, UPDATE_TYPE)
            for guid, mod in note_mods.items():
                col.db.execute("update notes set mod = ? where guid = ?", mod, guid)

        def note_of(guid: str):
            return col.get_note(col.db.scalar("select id from notes where guid = ?", guid))

        def export_deck(name: str, legacy: bool = False) -> None:
            col.export_anki_package(
                out_path=str(out / name),
                options=ExportAnkiPackageOptions(
                    with_scheduling=False, with_deck_configs=False, with_media=False, legacy=legacy
                ),
                limit=DeckIdLimit(deck_id=deck_id),
            )

        set_mods(UPDATE_T1, {"cf-up-hola": UPDATE_T1, "cf-up-gato": UPDATE_T1})
        export_deck("update-v1.apkg")
        export_deck("update-v1-legacy.apkg", legacy=True)

        hola = note_of("cf-up-hola")
        hola["Back"] = "hello"
        col.update_note(hola)
        nt = mm.by_name(UPDATE_TYPE)
        nt["css"] = UPDATE_CSS_V2
        mm.update_dict(nt)
        set_mods(UPDATE_T2, {"cf-up-hola": UPDATE_T2})
        export_deck("update-v2.apkg")

        nt = mm.by_name(UPDATE_TYPE)
        mm.add_field(nt, mm.new_field("Audio"))
        nt["tmpls"][0]["afmt"] += "\n{{Audio}}"
        mm.update_dict(nt)
        hola = note_of("cf-up-hola")
        hola["Audio"] = "hola-audio"
        col.update_note(hola)
        perro = col.new_note(mm.by_name(UPDATE_TYPE))
        perro["Front"], perro["Back"], perro["Audio"] = "perro", "dog", "perro-audio"
        perro.guid = "cf-up-perro"
        col.add_note(perro, deck_id)
        # Adding a field rewrites every note of the type.
        set_mods(UPDATE_T3, {"cf-up-hola": UPDATE_T3, "cf-up-gato": UPDATE_T3, "cf-up-perro": UPDATE_T3})
        export_deck("update-v3.apkg")
    finally:
        col.close()


UPDATE_PACKAGES = ("update-v1.apkg", "update-v1-legacy.apkg", "update-v2.apkg", "update-v3.apkg")


def export_latex(path: Path, out: Path) -> dict[str, list[list[str]]]:
    """Writes latex.apkg and its legacy twin latex-legacy.apkg and returns
    the LaTeX image names Anki gives each note's fields, by GUID."""
    col = Collection(str(path))
    col.crt = CRT
    names: dict[str, list[list[str]]] = {}
    try:
        mm = col.models
        for type_name, svg in ((LATEX_PNG_TYPE, False), (LATEX_SVG_TYPE, True)):
            nt = mm.new(type_name)
            for name in ("Front", "Back"):
                mm.add_field(nt, mm.new_field(name))
            t = mm.new_template("Card 1")
            t["qfmt"] = "{{Front}}"
            t["afmt"] = "{{FrontSide}}\n\n<hr id=answer>\n\n{{Back}}"
            mm.add_template(nt, t)
            nt["latexsvg"] = svg
            mm.add_dict(nt)
        deck_id = col.decks.id(LATEX_DECK)
        media_dir = Path(col.media.dir())
        for guid, type_name, fields in LATEX_NOTES:
            nt = mm.by_name(type_name)
            note = col.new_note(nt)
            for i, value in enumerate(fields):
                note.fields[i] = value
            note.guid = guid
            col.add_note(note, deck_id)
            names[guid] = []
            for value in fields:
                refs = col._backend.extract_latex(text=value, svg=nt["latexsvg"], expand_clozes=False).latex
                names[guid].append([r.filename for r in refs])
                if guid == LATEX_MISSING:
                    continue
                for r in refs:
                    data = LATEX_SVG if r.filename.endswith(".svg") else png((48, 96, 192))
                    (media_dir / r.filename).write_bytes(data)
        for name, legacy in (("latex.apkg", False), ("latex-legacy.apkg", True)):
            col.export_anki_package(
                out_path=str(out / name),
                options=ExportAnkiPackageOptions(
                    with_scheduling=False, with_deck_configs=False, with_media=True, legacy=legacy
                ),
                limit=DeckIdLimit(deck_id=deck_id),
            )
    finally:
        col.close()
    return names


LATEX_PACKAGES = ("latex.apkg", "latex-legacy.apkg")


def export_occlusion(path: Path, out: Path) -> None:
    """Writes occlusion.apkg and its legacy twin occlusion-legacy.apkg."""
    col = Collection(str(path))
    col.crt = CRT
    try:
        col.add_image_occlusion_notetype()
        nt = col.models.by_name("Image Occlusion")
        # A 40x24 image, so x and y scale differently.
        image = path.parent / OCCLUSION_IMAGE
        image.write_bytes(png((230, 220, 200), size=40, height=24))
        for guid, occlusions, header, back_extra, tags in OCCLUSION_NOTES:
            col.add_image_occlusion_note(nt["id"], str(image), occlusions, header, back_extra, tags)
            nid = col.db.scalar("select id from notes where guid not like 'cf-%'")
            note = col.get_note(nid)
            note.guid = guid
            col.update_note(note)
        # add_image_occlusion_note adds to the current deck; move the cards.
        deck_id = col.decks.id(OCCLUSION_DECK)
        col.set_deck(col.db.list("select id from cards"), deck_id)
        for name, legacy in (("occlusion.apkg", False), ("occlusion-legacy.apkg", True)):
            col.export_anki_package(
                out_path=str(out / name),
                options=ExportAnkiPackageOptions(
                    with_scheduling=False, with_deck_configs=False, with_media=True, legacy=legacy
                ),
                limit=DeckIdLimit(deck_id=deck_id),
            )
    finally:
        col.close()


OCCLUSION_PACKAGES = ("occlusion.apkg", "occlusion-legacy.apkg")


def add_collection_extras(col: Collection) -> None:
    col.decks.id(EMPTY_DECK)
    did = col.decks.new_filtered(FILTERED_DECK)
    deck = col.decks.get(did)
    deck["terms"] = [[FILTER_SEARCH, 100, 0]]
    deck["resched"] = True
    col.decks.save(deck)
    col.sched.rebuild_filtered_deck(did)


def export_collection(col: Collection, out: Path, legacy: bool) -> None:
    # Anki closes the collection to export it.
    col.export_collection_package(str(out), include_media=True, legacy=legacy)
    col.reopen()


# --- normalised content, read straight from the packages -------------------


def read_package(path: Path, mods: bool = False) -> dict:
    with zipfile.ZipFile(path) as z:
        names = set(z.namelist())
        if "collection.anki21b" in names:
            db_bytes = zstd.decompress(z.read("collection.anki21b"))
            kind = "anki21b"
        elif "collection.anki21" in names:
            db_bytes = z.read("collection.anki21")
            kind = "anki21"
        else:
            db_bytes = z.read("collection.anki2")
            kind = "anki2"
        media = read_media(z, kind)

    with tempfile.TemporaryDirectory() as tmp:
        db_path = Path(tmp) / "c.db"
        db_path.write_bytes(db_bytes)
        db = sqlite3.connect(db_path)
        # Anki declares some text columns with its own collation.
        db.create_collation("unicase", lambda a, b: (a.casefold() > b.casefold()) - (a.casefold() < b.casefold()))
        try:
            content = dump_db(db, kind, whole=path.suffix == ".colpkg", mods=mods)
        finally:
            db.close()
    content["format"] = kind
    content["media"] = media
    return content


def read_media(z: zipfile.ZipFile, kind: str) -> dict[str, str]:
    """Media file name -> sha256 of the uncompressed bytes."""
    out = {}
    if kind == "anki21b":
        # MediaEntries { repeated MediaEntry entries = 1; }
        # MediaEntry { string name = 1; uint32 size = 2; bytes sha1 = 3; }
        raw = zstd.decompress(z.read("media"))
        for i, entry in enumerate(proto_fields(raw, 1)):
            name = proto_fields(entry, 1)[0].decode()
            out[name] = hashlib.sha256(zstd.decompress(z.read(str(i)))).hexdigest()
    else:
        for idx, name in json.loads(z.read("media")).items():
            out[name] = hashlib.sha256(z.read(idx)).hexdigest()
    return dict(sorted(out.items()))


def proto_fields(buf: bytes, want: int) -> list[bytes]:
    """Length-delimited values of field number `want` (tiny protobuf reader)."""
    vals, i = [], 0

    def varint() -> int:
        nonlocal i
        shift = result = 0
        while True:
            b = buf[i]
            i += 1
            result |= (b & 0x7F) << shift
            if not b & 0x80:
                return result
            shift += 7

    while i < len(buf):
        key = varint()
        num, wt = key >> 3, key & 7
        if wt == 0:
            varint()
        elif wt == 2:
            n = varint()
            if num == want:
                vals.append(buf[i : i + n])
            i += n
        elif wt == 1:
            i += 8
        elif wt == 5:
            i += 4
        else:
            raise ValueError(f"wire type {wt}")
    return vals


def preset_of(name: str, cfg: deck_config_pb2.DeckConfig.Config) -> dict:
    """The preset fields the importer maps, with Anki's values and enums."""
    return {
        "name": name,
        "newPerDay": cfg.new_per_day,
        "reviewsPerDay": cfg.reviews_per_day,
        "learnSteps": [round(s, 4) for s in cfg.learn_steps],
        "relearnSteps": [round(s, 4) for s in cfg.relearn_steps],
        "desiredRetention": round(cfg.desired_retention, 4),
        "maxIvl": cfg.maximum_review_interval,
        "minInt": cfg.minimum_lapse_interval,
        "leechThreshold": cfg.leech_threshold,
        "leechAction": cfg.leech_action,
        "insertOrder": cfg.new_card_insert_order,
        "gatherPriority": cfg.new_card_gather_priority,
        "sortOrder": cfg.new_card_sort_order,
        "newMix": cfg.new_mix,
        "interdayMix": cfg.interday_learning_mix,
        "reviewOrder": cfg.review_order,
        "buryNew": cfg.bury_new,
        "buryReviews": cfg.bury_reviews,
        "buryInterdayLearning": cfg.bury_interday_learning,
        "disableAutoplay": cfg.disable_autoplay,
        "skipQuestionWhenReplaying": cfg.skip_question_when_replaying_answer,
        "secondsToShowQuestion": round(cfg.seconds_to_show_question, 4),
        "secondsToShowAnswer": round(cfg.seconds_to_show_answer, 4),
        "questionAction": cfg.question_action,
        "answerAction": cfg.answer_action,
    }


def legacy_config(c: dict) -> deck_config_pb2.DeckConfig.Config:
    """A legacy dconf entry as Anki's protobuf config."""
    return deck_config_pb2.DeckConfig.Config(
        new_per_day=c["new"]["perDay"],
        reviews_per_day=c["rev"]["perDay"],
        learn_steps=c["new"]["delays"],
        relearn_steps=c["lapse"]["delays"],
        desired_retention=c.get("desiredRetention", 0.9),
        maximum_review_interval=c["rev"]["maxIvl"],
        minimum_lapse_interval=c["lapse"]["minInt"],
        leech_threshold=c["lapse"]["leechFails"],
        leech_action=c["lapse"]["leechAction"],
        # Legacy order: 0 random, 1 due; protobuf: 0 due, 1 random.
        new_card_insert_order=1 - c["new"]["order"],
        new_card_gather_priority=c.get("newGatherPriority", 0),
        new_card_sort_order=c.get("newSortOrder", 0),
        new_mix=c.get("newMix", 0),
        interday_learning_mix=c.get("interdayLearningMix", 0),
        review_order=c.get("reviewOrder", 0),
        bury_new=c["new"].get("bury", False),
        bury_reviews=c["rev"].get("bury", False),
        bury_interday_learning=c.get("buryInterdayLearning", False),
        disable_autoplay=not c.get("autoplay", True),
        skip_question_when_replaying_answer=not c.get("replayq", True),
        seconds_to_show_question=c.get("secondsToShowQuestion", 0),
        seconds_to_show_answer=c.get("secondsToShowAnswer", 0),
        question_action=c.get("questionAction", 0),
        answer_action=c.get("answerAction", 0),
    )


def deck_presets(db: sqlite3.Connection, kind: str, decks: dict[int, str], whole: bool) -> dict[str, dict]:
    """Deck name -> the preset it uses, for the Fixture decks (all normal
    decks of a collection package)."""
    by_deck: dict[int, dict] = {}
    if kind == "anki21b":
        configs = {
            cid: preset_of(name, deck_config_pb2.DeckConfig.Config.FromString(blob))
            for cid, name, blob in db.execute("select id, name, config from deck_config")
        }
        for did, kind_blob in db.execute("select id, kind from decks"):
            # Deck.KindContainer { Normal normal = 1; } Normal { int64 config_id = 1; }
            normal = proto_fields(kind_blob, 1)
            if normal:
                by_deck[did] = configs[proto_varint(normal[0], 1) or 1]
    else:
        dconf = json.loads(db.execute("select dconf from col").fetchone()[0])
        configs = {int(k): preset_of(v["name"], legacy_config(v)) for k, v in dconf.items()}
        for k, v in json.loads(db.execute("select decks from col").fetchone()[0]).items():
            if not v.get("dyn"):
                by_deck[int(k)] = configs[v["conf"]]
    return {
        decks[did]: p
        for did, p in sorted(by_deck.items(), key=lambda x: decks[x[0]])
        if whole or decks[did].startswith(FIXTURE_DECKS)
    }


def proto_varint(buf: bytes, want: int) -> int:
    """The value of varint field number `want`, or 0."""
    out, i = 0, 0

    def varint() -> int:
        nonlocal i
        shift = result = 0
        while True:
            b = buf[i]
            i += 1
            result |= (b & 0x7F) << shift
            if not b & 0x80:
                return result
            shift += 7

    while i < len(buf):
        key = varint()
        num, wt = key >> 3, key & 7
        if wt == 0:
            v = varint()
            if num == want:
                out = v
        elif wt == 2:
            i += varint()
        elif wt == 1:
            i += 8
        elif wt == 5:
            i += 4
        else:
            raise ValueError(f"wire type {wt}")
    return out


def dump_db(db: sqlite3.Connection, kind: str, whole: bool, mods: bool = False) -> dict:
    """The package's content. A collection package (whole) lists all decks
    and note types; a deck package only the Fixture decks and the note types
    its notes use. With mods, note types and notes include their
    modification times (only pinned in the update packages)."""
    crt = db.execute("select crt from col").fetchone()[0]
    if kind == "anki21b":
        decks = {did: name.replace("\x1f", "::") for did, name in db.execute("select id, name from decks")}
        notetypes = {}
        for ntid, name, mtime, config in db.execute("select id, name, mtime_secs, config from notetypes"):
            fields = [r[0] for r in db.execute("select name from fields where ntid = ? order by ord", (ntid,))]
            templates = [r[0] for r in db.execute("select name from templates where ntid = ? order by ord", (ntid,))]
            notetypes[ntid] = {"name": name, "fields": fields, "templates": templates}
            if notetypes_pb2.Notetype.Config.FromString(config).latex_svg:
                notetypes[ntid]["latexSvg"] = True
            if mods:
                notetypes[ntid]["mod"] = mtime
    else:
        decks = {int(k): v["name"] for k, v in json.loads(db.execute("select decks from col").fetchone()[0]).items()}
        notetypes = {
            int(k): {
                "name": v["name"],
                "fields": [f["name"] for f in v["flds"]],
                "templates": [t["name"] for t in v["tmpls"]],
                **({"latexSvg": True} if v.get("latexsvg") else {}),
                **({"mod": v["mod"]} if mods else {}),
            }
            for k, v in json.loads(db.execute("select models from col").fetchone()[0]).items()
        }

    notes = {}
    guid_by_id = {}
    for nid, guid, mid, tags, flds, mod in db.execute("select id, guid, mid, tags, flds, mod from notes order by guid"):
        guid_by_id[nid] = guid
        notes[guid] = {
            "notetype": notetypes[mid]["name"],
            "fields": flds.split("\x1f"),
            "tags": tags.split(),
            **({"mod": mod} if mods else {}),
        }

    cards = []
    for nid, did, odid, odue, ord_, type_, queue, due, ivl, factor, reps, lapses, flags, data in db.execute(
        "select nid, did, odid, odue, ord, type, queue, due, ivl, factor, reps, lapses, flags, data from cards"
    ):
        memory = None
        if data:
            d = json.loads(data)
            if "s" in d:
                memory = {"s": d["s"], "d": d["d"]}
        # A card in a filtered deck: its home deck and original due.
        home = {"homeDeck": decks[odid], "homeDue": odue} if odid else {}
        cards.append(
            {
                "note": guid_by_id[nid],
                "ord": ord_,
                "deck": decks[did],
                "type": type_,
                "queue": queue,
                "due": due,
                "ivl": ivl,
                "factor": factor,
                "reps": reps,
                "lapses": lapses,
                "flag": flags & 7,
                "memory": memory,
                **home,
            }
        )
    cards.sort(key=lambda c: (c["note"], c["ord"]))

    # Review history per card, "<guid>/<ord>" -> [id, ease, ivl, lastIvl,
    # factor, time, type] in ID order.
    card_key = {
        cid: f"{guid_by_id[nid]}/{ord_}" for cid, nid, ord_ in db.execute("select id, nid, ord from cards")
    }
    revlog: dict[str, list[list[int]]] = {}
    for row in db.execute("select cid, id, ease, ivl, lastIvl, factor, time, type from revlog order by id"):
        revlog.setdefault(card_key[row[0]], []).append(list(row[1:]))

    used = {notes[g]["notetype"] for g in notes}
    if kind == "anki21b":
        filtered = {did for did, kind_blob in db.execute("select id, kind from decks") if proto_fields(kind_blob, 2)}
    else:
        filtered = {int(k) for k, v in json.loads(db.execute("select decks from col").fetchone()[0]).items() if v.get("dyn")}
    return {
        "crt": crt,
        "decks": sorted(name for name in decks.values() if whole or name.startswith(FIXTURE_DECKS)),
        "filteredDecks": sorted(decks[did] for did in filtered),
        "deckPresets": deck_presets(db, kind, decks, whole),
        "notetypes": sorted(
            (nt for nt in notetypes.values() if whole or nt["name"] in used), key=lambda nt: nt["name"]
        ),
        "notes": notes,
        "cards": cards,
        "revlog": dict(sorted(revlog.items())),
    }


# --- driver ------------------------------------------------------------------


def generate(out: Path) -> dict:
    out.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory() as tmp:
        col = build_collection(Path(tmp) / "fixture.anki2")
        try:
            export(col, out / "modern.apkg", legacy=False)
            export(col, out / "legacy.apkg", legacy=True)
            add_collection_extras(col)
            export_collection(col, out / "modern.colpkg", legacy=False)
            export_collection(col, out / "legacy.colpkg", legacy=True)
        finally:
            col.close()
        export_updates(Path(tmp) / "updates.anki2", out)
        latex_names = export_latex(Path(tmp) / "latex.anki2", out)
        export_occlusion(Path(tmp) / "occlusion.anki2", out)

    for ext in ("apkg", "colpkg"):
        modern = (out / f"modern.{ext}").read_bytes()
        (out / f"truncated.{ext}").write_bytes(modern[: len(modern) // 2])

    manifest = {
        "comment": "Generated by tool/make_apkg_fixtures.py. Do not edit.",
        "crt": CRT,
        "learnDue": LEARN_DUE,
        "packages": {
            name: read_package(out / name)
            for name in ("modern.apkg", "legacy.apkg", "modern.colpkg", "legacy.colpkg")
        },
        "updates": {name: read_package(out / name, mods=True) for name in UPDATE_PACKAGES},
        "latex": {
            "packages": {name: read_package(out / name) for name in LATEX_PACKAGES},
            # The LaTeX image names Anki gives each field, by note GUID.
            "names": latex_names,
            "missing": LATEX_MISSING,
        },
        "occlusion": {name: read_package(out / name) for name in OCCLUSION_PACKAGES},
    }
    (out / "fixtures.json").write_text(json.dumps(manifest, indent=2, ensure_ascii=False) + "\n")
    return manifest


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--check", action="store_true", help="compare a fresh build with the committed fixtures")
    args = parser.parse_args()

    if not args.check:
        generate(OUT)
        print(f"wrote fixtures to {OUT.relative_to(ROOT)}")
        return 0

    committed = json.loads((OUT / "fixtures.json").read_text())
    with tempfile.TemporaryDirectory() as tmp:
        fresh = generate(Path(tmp))
        # The committed packages must still contain what fixtures.json says.
        for name in fresh["packages"]:
            if read_package(OUT / name) != committed["packages"][name]:
                print(f"{name}: committed package does not match fixtures.json")
                return 1
        for name in fresh["updates"]:
            if read_package(OUT / name, mods=True) != committed["updates"][name]:
                print(f"{name}: committed package does not match fixtures.json")
                return 1
        for name in fresh["latex"]["packages"]:
            if read_package(OUT / name) != committed["latex"]["packages"][name]:
                print(f"{name}: committed package does not match fixtures.json")
                return 1
        for name in fresh["occlusion"]:
            if read_package(OUT / name) != committed["occlusion"][name]:
                print(f"{name}: committed package does not match fixtures.json")
                return 1
    if fresh != committed:
        print("regenerated fixture content differs from the committed fixtures.json")
        return 1
    print("fixtures are reproducible: regenerated content is identical")
    return 0


if __name__ == "__main__":
    sys.exit(main())
