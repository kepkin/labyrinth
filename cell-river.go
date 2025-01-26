package labyrinth

import "fmt"

type RiverCell struct {
	Dir     MoveDirection
	isMouth bool
}

const RiverCellDirectionAttr = "river_dir"
const RiverCellMouthAttr = "river_mouth"

func (c RiverCell) Type() CellType {
	name := "river"
	mouthAttr := ""
	if c.isMouth {
		name = "river mouth"
		mouthAttr = "true"
	}
	return CellType{
		Class:      "river",
		Name:       name,
		Attributes: map[string]string{RiverCellDirectionAttr: c.Dir.Utf8Arrow(), RiverCellMouthAttr: mouthAttr},
	}
}

func IsRiverMouth(c Cell) bool {
	if c.Class != "river" {
		return false
	}

	return MoveDirectionFromUtf8Arrow(c.Attributes[RiverCellDirectionAttr]) == MoveNil
}

func BuildRiver(cellMap CellMap, p Position) error {
	currCell := cellMap.Get(p)

	if currCell.Class != "river" {
		return fmt.Errorf("logical error, appeared on non river cell")
	}

	// TODO: optimize - check sidenss
	nextDirection, _ := findNextRiver(cellMap, p, MoveNil)
	return FindRiverLastCell(cellMap, p.Next(nextDirection), nextDirection.TurnBack())
}

func findNextRiver(cf CellMap, pos Position, from MoveDirection) (MoveDirection, int) {
	adjacentCells := map[MoveDirection]Cell{}
	adjacentCells[North] = cf.Get(pos.Next(North))
	adjacentCells[South] = cf.Get(pos.Next(South))
	adjacentCells[East] = cf.Get(pos.Next(East))
	adjacentCells[West] = cf.Get(pos.Next(West))

	nextDirection := MoveNil
	numberOfAdjacentRivers := 0
	for k, v := range adjacentCells {
		if k == from {
			continue
		}
		if v.Class == "river" {
			nextDirection = k
			numberOfAdjacentRivers++
		}
	}

	return nextDirection, numberOfAdjacentRivers
}

// TODO detect LOOPS
func FindRiverLastCell(cellMap CellMap, pos Position, from MoveDirection) error {
	nextDirection, numberOfAdjacentRivers := findNextRiver(cellMap, pos, from)
	if numberOfAdjacentRivers == 0 {
		currCell := cellMap.Get(pos)

		rc := currCell.Custom.(*RiverCell)
		if rc.isMouth {
			return BuildRiverFromMouth(cellMap, pos.Next(nextDirection), nextDirection)
		} else {
			return BuildRiverFromSource(cellMap, pos, from.TurnBack())
		}
	}

	return FindRiverLastCell(cellMap, pos.Next(nextDirection), nextDirection.TurnBack())
}

func BuildRiverFromSource(cellMap CellMap, pos Position, from MoveDirection) error {
	nextDirection, _ := findNextRiver(cellMap, pos, from)

	if nextDirection != MoveNil {
		c := cellMap.Get(pos)
		rc := c.Custom.(*RiverCell)
		rc.Dir = nextDirection

		return BuildRiverFromSource(cellMap, pos.Next(nextDirection), nextDirection.TurnBack())
	}

	return nil
}

func BuildRiverFromMouth(cellMap CellMap, pos Position, from MoveDirection) error {
	nextDirection, _ := findNextRiver(cellMap, pos, from)

	if nextDirection != MoveNil {
		c := cellMap.Get(pos)
		rc := c.Custom.(*RiverCell)
		rc.Dir = from

		return BuildRiverFromSource(cellMap, pos.Next(nextDirection), nextDirection)
	}

	return nil
}

var RiverStringFactoryKeys = []string{"←", "↑", "→", "↓", "r", "R", "RM"}

type RiverStringCellFactory struct {
}

func (rscf RiverStringCellFactory) Make(key string, pos Position) (Cell, error) {
	switch key {
	case "←", "↑", "→", "↓", "r", "R":
		return &CellType{Class: "river", Custom: &RiverCell{Dir: MoveDirectionFromUtf8Arrow(key)}}, nil
	case "RM":
		return &CellType{Class: "river", Custom: &RiverCell{Dir: MoveDirectionFromUtf8Arrow(key), isMouth: true}}, nil
	}

	return nil, fmt.Errorf("can not build river cell from %v", key)
}

func (rscf RiverStringCellFactory) Finish(cm CellMap) {
	for p, c := range cm.All() {
		if c.Class == "river" {
			cr := c.Custom.(*RiverCell)
			if cr.Dir == MoveNil {
				//TODO: simplify this by triggering only on mouth of the river
				BuildRiver(cm, p)
			}
		}
	}
}
