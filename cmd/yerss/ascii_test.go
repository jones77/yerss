package main

import (
	"testing"

	"yerss/internal/ui"
)

func TestAsciiFlagForcesAscii(t *testing.T) {
	cfg, st := gateTestEnv(t, "# nothing\n")
	m := ui.New(cfg, st)
	// Simulate config ascii=false and terminal detection=false.
	m.SetAscii(false)

	applyAsciiFlag(m, true)
	if !m.Ascii() {
		t.Error("expected -a to force ASCII fallback on the model")
	}
}

func TestAsciiFlagAbsentLeavesAsciiOff(t *testing.T) {
	cfg, st := gateTestEnv(t, "# nothing\n")
	m := ui.New(cfg, st)
	m.SetAscii(false)

	applyAsciiFlag(m, false)
	if m.Ascii() {
		t.Error("expected ascii to stay off when -a is absent")
	}
}