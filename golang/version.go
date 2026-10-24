package main

// appVersion is the Cardflip version, shown in About and written to backups.
const appVersion = "0.1.0"

// Version returns the Cardflip version.
//
//fluffy:export
func Version() string { return appVersion }
