package labyrinth

import (
	"fmt"
	"strconv"
	"strings"
)

const WormholeCellSystemNameAttr = "wormhole_name"

type WormholeCell struct {
	NextPos Position
	Name    string
	Idx     int
}

func (c WormholeCell) Type() CellType {
	return CellType{
		Class:      "wormhole",
		Name:       "wormhole",
		Attributes: map[string]string{WormholeCellSystemNameAttr: c.Name},
	}
}

type WormholeStringCellFactory struct {
	whormholeSystems map[string]map[int]Position
}

func (tscf *WormholeStringCellFactory) Make(key string, pos Position) (Cell, error) {
	whormholeVals := strings.Split(key, ":")
	if len(whormholeVals) != 3 {
		panic(fmt.Sprintf("invalid whormhole cell: `%v`", key))
	}

	whormholeSystemName := strings.TrimSpace(whormholeVals[1])
	whormholeIdx, err := strconv.Atoi(whormholeVals[2])
	if err != nil {
		panic(fmt.Sprintf("invalid whormhole cell: `%v`", key))
	}

	if tscf.whormholeSystems == nil {
		tscf.whormholeSystems = map[string]map[int]Position{}
	}
	if tscf.whormholeSystems[whormholeSystemName] == nil {
		tscf.whormholeSystems[whormholeSystemName] = map[int]Position{}
	}
	tscf.whormholeSystems[whormholeSystemName][whormholeIdx] = pos

	return &CellType{Class: "wormhole", Custom: &WormholeCell{Name: whormholeSystemName, Idx: whormholeIdx}}, nil
}

func (tscf *WormholeStringCellFactory) Finish(cm CellMap) {
	for _, c := range cm.All() {
		if c.Class == "wormhole" {
			tc, ok := c.Custom.(*WormholeCell)
			if !ok {
				panic("type conversion failes")
			}
			nextPos := tscf.whormholeSystems[tc.Name][tc.Idx+1]
			if nextPos == NewPosition(0, 0) {
				nextPos = tscf.whormholeSystems[tc.Name][0]
			}

			tc.NextPos = nextPos
			// cm.Insert(c, p)
		}
	}
}
