#!/usr/bin/env python3
"""Writes test.db for the sqlite reader tests. Run from this directory."""

import os
import sqlite3

if os.path.exists("test.db"):
    os.remove("test.db")
db = sqlite3.connect("test.db")
db.execute("pragma page_size = 1024")
db.executescript(
    """
    create table items (
      id integer primary key, -- rowid alias
      name text not null,
      score real,
      data blob
    );
    create table "pairs" (
      label text not null,
      a integer not null,
      b integer not null,
      primary key (b, a)
    ) without rowid;
    create table plain (x, y);
    """
)
# 1500 rows force interior table pages; every 100th row has a blob spanning
# several overflow pages.
for i in range(1, 1501):
    blob = bytes((i + j) % 256 for j in range(5000)) if i % 100 == 0 else None
    db.execute(
        "insert into items values (?, ?, ?, ?)",
        (i * 7, f"item-{i}-" + "x" * (i % 50), i / 4 - 100, blob),
    )
for i in range(800):
    db.execute("insert into pairs values (?, ?, ?)", ("p" * (i % 300) + str(i), i, -i * 1000003))
db.execute("insert into plain values (1, 2)")
db.execute("alter table plain add column z text default 'later'")
db.execute("insert into plain values (3, 4, 'z4')")
db.execute("insert into plain values (-9223372036854775808, 9223372036854775807, null)")
db.commit()
db.execute("vacuum")
db.close()
