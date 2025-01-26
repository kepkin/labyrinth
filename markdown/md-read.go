package labyrinth

import (
	"fmt"
	"strconv"
	"strings"

	lab "github.com/kepkin/labyrinth"
)

type mdTableProccessor interface {
	next(c rune) (bool, error)
}

type headerReader struct {
	wb              *WorldBuilder
	columns         int
	prefix          strings.Builder
	readFirstColumn bool
}

func (h *headerReader) next(c rune) (bool, error) {
	if c == '|' && !h.readFirstColumn && strings.TrimSpace(h.prefix.String()) != "" {
		h.readFirstColumn = true
		h.prefix.Reset()

		h.columns++

		return false, nil
	}

	if c == '|' {
		h.columns++
		return false, nil
	}

	if !h.readFirstColumn {
		h.prefix.WriteRune(c)
	}

	if c == '\n' {
		h.wb.maxX = h.columns
		for i := 0; i < h.columns; i++ {
			h.wb.Cf.MakeCell("w", i, 0)
		}

		return true, nil
	}

	return false, nil
}

type afterHeaderLineReader struct {
}

func (h *afterHeaderLineReader) next(c rune) (bool, error) {
	if c != '|' && c != '-' && c != ' ' && c != '\n' {
		return false, fmt.Errorf("unexpected symbol `%v` after header line", c)
	}

	return c == '\n', nil
}

type rowReader struct {
	wb              *WorldBuilder
	columns         int
	prefix          strings.Builder
	readFirstColumn bool
	y               int
}

func (h *rowReader) next(c rune) (bool, error) {
	if c == '|' && !h.readFirstColumn && strings.TrimSpace(h.prefix.String()) != "" {
		h.readFirstColumn = true
		h.prefix.Reset()

		h.wb.Cf.MakeCell("w", h.columns, h.y)

		h.columns++

		return false, nil
	}

	if c == '|' && h.readFirstColumn {
		h.columns++

		h.wb.Cf.MakeCell(strings.TrimSpace(h.prefix.String()), h.columns-1, h.y)
		h.prefix.Reset()
		return false, nil
	}

	h.prefix.WriteRune(c)

	if c == '\n' && h.columns == 0 {
		for i := 0; i < h.wb.maxX; i++ {
			h.wb.Cf.MakeCell("w", i, h.y)
		}

		return true, nil
	}

	if c == '\n' {
		h.wb.Cf.MakeCell("w", h.columns, h.y)
		h.y++

		h.readFirstColumn = false
		h.prefix.Reset()
		h.columns = 0

		return false, nil
	}

	return false, nil
}

type namePosReader struct {
	wb     *WorldBuilder
	prefix strings.Builder
}

func (h *namePosReader) next(c rune) (bool, error) {
	if c == '\n' {
		lineValue := strings.TrimSpace(h.prefix.String())
		if lineValue == "" {
			return false, nil
		}

		if property, position, ok := strings.Cut(lineValue, ":"); ok {
			vals := strings.Split(position, ":")
			if len(vals) != 2 {
				return false, fmt.Errorf("property %v has incorrection position: `%v`", property, position)
			}
			x, err := strconv.Atoi(strings.TrimSpace(vals[0]))
			if err != nil {
				return false, fmt.Errorf("property %v has incorrection position: `%v`", property, position)
			}
			y, err := strconv.Atoi(strings.TrimSpace(vals[1]))
			if err != nil {
				return false, fmt.Errorf("property %v has incorrection position: `%v`", property, position)
			}

			h.wb.properties[property] = lab.NewPosition(x, y)
		}

		h.prefix.Reset()
		return false, nil
	}

	h.prefix.WriteRune(c)
	return false, nil
}

type WorldBuilder struct {
	Cf      lab.CellWorldBuilder
	Factory lab.StringCellFactory

	maxX       int
	properties map[string]lab.Position
}

func (wb *WorldBuilder) Build(wmap string) (*lab.World, []*lab.Player, error) {
	cf := &(wb.Cf)
	wb.properties = map[string]lab.Position{}
	if wb.Factory == nil {
		wb.Factory = lab.DefaultCellFactory
	}

	mdTableProccessorIdx := 0
	mdTableProccessor := []mdTableProccessor{
		&headerReader{wb: wb},
		&afterHeaderLineReader{},
		&rowReader{wb: wb, y: 1},
		&namePosReader{wb: wb},
	}

	for _, c := range wmap {
		if len(mdTableProccessor) <= mdTableProccessorIdx {
			break
		}

		next, err := mdTableProccessor[mdTableProccessorIdx].next(c)
		if err != nil {
			return nil, nil, err
		}
		if next {
			mdTableProccessorIdx++
		}
	}

	cellMap, err := cf.BuildCellMap()
	if err != nil {
		return nil, nil, err
	}

	var players []*lab.Player
	for propName, pos := range wb.properties {
		if propName == "exit" {
			exitCell, err := wb.Factory.Make("exit", pos)
			if err != nil {
				return nil, nil, err
			}
			cellMap.Insert(exitCell, pos)
		} else if propName == "treasure" {
			c := cellMap.Get(pos)
			c.PutItem(&lab.Item{ID: lab.Treasure, Name: "tresure"})
		} else if propName == "fake_treasure" {
			c := cellMap.Get(pos)
			c.PutItem(&lab.Item{ID: lab.FakeTreasure, Name: "tresure"})
		} else {
			players = append(players, &lab.Player{Name: propName, Pos: pos})
		}
	}

	return &lab.World{Cells: cellMap}, players, nil
}
