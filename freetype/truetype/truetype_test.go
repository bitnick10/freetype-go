// Copyright 2012 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

package truetype

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestParse tests that the luxisr.ttf metrics and glyphs are parsed correctly.
// The numerical values can be manually verified by examining luxisr.ttx.
func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("../../luxi-fonts/luxisr.ttf")
	if err != nil {
		t.Fatal(err)
	}
	font, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := font.FUnitsPerEm(), int32(2048); got != want {
		t.Errorf("FUnitsPerEm: got %v, want %v", got, want)
	}
	fupe := font.FUnitsPerEm()
	if got, want := font.Bounds(fupe), (Bounds{-441, -432, 2024, 2033}); got != want {
		t.Errorf("Bounds: got %v, want %v", got, want)
	}

	i0 := font.Index('A')
	i1 := font.Index('V')
	if i0 != 36 || i1 != 57 {
		t.Fatalf("Index: i0, i1 = %d, %d, want 36, 57", i0, i1)
	}
	if got, want := font.HMetric(fupe, i0), (HMetric{1366, 19}); got != want {
		t.Errorf("HMetric: got %v, want %v", got, want)
	}
	if got, want := font.Kerning(fupe, i0, i1), int32(-144); got != want {
		t.Errorf("Kerning: got %v, want %v", got, want)
	}

	g0 := NewGlyphBuf()
	err = g0.Load(font, fupe, i0, nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	g1 := &GlyphBuf{
		B: Bounds{19, 0, 1342, 1480},
		Point: []Point{
			{19, 0, 51},
			{581, 1480, 1},
			{789, 1480, 51},
			{1342, 0, 1},
			{1116, 0, 35},
			{962, 410, 3},
			{368, 410, 33},
			{214, 0, 3},
			{428, 566, 19},
			{904, 566, 33},
			{667, 1200, 3},
		},
		End: []int{8, 11},
	}
	if got, want := fmt.Sprint(g0), fmt.Sprint(g1); got != want {
		t.Errorf("GlyphBuf:\ngot  %v\nwant %v", got, want)
	}
}

func testScaling(t *testing.T, filename string, hinter *Hinter) {
	b, err := ioutil.ReadFile("../../luxi-fonts/luxisr.ttf")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	font, err := Parse(b)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	f, err := os.Open("../../luxi-fonts/" + filename)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	wants := [][]Point{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			wants = append(wants, []Point{})
			continue
		}
		ss := strings.Split(text, ",")
		points := make([]Point, len(ss))
		for i, s := range ss {
			p := &points[i]
			if _, err := fmt.Sscanf(s, "%d %d %d", &p.X, &p.Y, &p.Flags); err != nil {
				t.Fatalf("Sscanf: %v", err)
			}
		}
		wants = append(wants, points)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		t.Fatalf("Scanner: %v", err)
	}

	const fontSize = 12
	glyphBuf := NewGlyphBuf()
	for i, want := range wants {
		// TODO: completely implement hinting. For now, only the first N glyphs
		// of luxisr.ttf are correctly hinted.
		const N = 1
		if hinter != nil && i == N {
			break
		}

		if err = glyphBuf.Load(font, fontSize*64, Index(i), hinter); err != nil {
			t.Fatalf("Load: %v", err)
		}
		got := glyphBuf.Point
		for i := range got {
			got[i].Flags &= 0x01
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("glyph #%d:\ngot  %v\nwant %v\n", i, got, want)
		}
	}
}

func TestScalingSansHinting(t *testing.T) {
	testScaling(t, "luxisr-12pt-sans-hinting.txt", nil)
}

func TestScalingWithHinting(t *testing.T) {
	testScaling(t, "luxisr-12pt-with-hinting.txt", &Hinter{})
}
