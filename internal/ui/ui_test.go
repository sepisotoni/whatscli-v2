package ui

import (
	"strings"
	"testing"

	"github.com/rivo/tview"
)

func TestQRRenderDimensions(t *testing.T) {
	qr, width, height, err := renderQR("whatshell-qr-test-data")
	if err != nil {
		t.Fatal(err)
	}
	if width < 20 || height < 10 || height != (width+1)/2 {
		t.Fatalf("unexpected QR dimensions %dx%d", width, height)
	}
	if strings.Count(qr, "\n") != height || !strings.ContainsAny(qr, "▀▄█") {
		t.Fatalf("QR rendering lost rows or pixels: %q", qr)
	}
}

func TestCommandPaletteFiltersAndCompletes(t *testing.T) {
	u := &UI{input: tview.NewInputField(), palette: tview.NewList(), main: tview.NewFlex()}
	u.main.AddItem(u.palette, 0, 0, false)
	u.updatePalette("/se")
	if u.palette.GetItemCount() != 1 || !u.completeCommand() {
		t.Fatalf("expected /search suggestion, got %d", u.palette.GetItemCount())
	}
	if u.input.GetText() != "/search " {
		t.Fatalf("completion = %q", u.input.GetText())
	}
	u.updatePalette("/not-a-command")
	if u.palette.GetItemCount() != 0 || u.completeCommand() {
		t.Fatal("unknown command offered autocomplete")
	}
	u.updatePalette("hello")
	if u.palette.GetItemCount() != 0 {
		t.Fatal("normal chat text unexpectedly opens command palette")
	}
}