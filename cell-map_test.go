package labyrinth

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPrintCellMap(t *testing.T) {
	w := WallCell{}

	tests := []struct {
		name    string
		cellMap CellMap
		wantW   string
	}{
		// TODO: Add test cases.
		{
			name: "print one row",
			cellMap: CellMap{
				v: [][]Cell{
					{w, w},
				},
			},

			wantW: "|wall|wall|",
		},
		{
			name:    "print none on empty CellMap",
			cellMap: CellMap{},

			wantW: "",
		},
		{
			name: "print each row on separate line",
			cellMap: CellMap{
				v: [][]Cell{
					{w, w},
					{w, w},
				},
			},

			wantW: "|wall|wall|\n|wall|wall|",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			FPrintCellMap(w, tt.cellMap)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("FPrintCellMap() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestCellMap_Insert(t *testing.T) {
	w := WallCell{}

	tests := []struct {
		name string
		v    [][]Cell

		c Cell
		p Position

		want [][]Cell
	}{
		{
			v: nil,

			c: w,
			p: NewPosition(3, 3),

			want: [][]Cell{
				nil,
				nil,
				nil,
				{WallCell{pos: NewPosition(0, 3)}, WallCell{pos: NewPosition(1, 3)}, WallCell{pos: NewPosition(2, 3)}, WallCell{pos: NewPosition(0, 0)}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &CellMap{
				v: tt.v,
			}
			cm.Insert(tt.c, tt.p)

			assert.Equal(t, tt.want, cm.v)
		})
	}
}
