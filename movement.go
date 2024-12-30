package labyrinth

type SimpleMoveCommand struct{}

func (c SimpleMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	nextCoo := p.Pos.Next(direction)
	p.Pos = nextCoo
	nextCell := w.Cells.Get(nextCoo)

	e := NewEventf(nextCell.Class)
	w.Emmit(e)
	return []Event{e}
}

type WallMoveCommand struct{}

func (c WallMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	e := NewEventf("it's a wall")
	w.Emmit(e)
	return []Event{e}
}

type ExitMoveCommand struct{}

func (c ExitMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	se := SimpleMoveCommand{}.Do(w, p, direction)
	if p.Hand == Treasure {
		e := NewEventf("your tresure is genuine")
		w.Emmit(e)
		return append(se, e)
	} else if p.Hand == FakeTreasure {
		e := NewEventf("your tresure is Fake")
		w.Emmit(e)
		return append(se, e)
	}

	return se
}

var moveRouting = map[string]map[string]MoveCommandType{
	"earth": {
		"river":    &RiverMoveCommand{},
		"exit":     &ExitMoveCommand{},
		"wall":     &WallMoveCommand{},
		"wormhole": &WormholeMoveCommand{},
	},
	"exit": {
		"river":    &RiverMoveCommand{},
		"wall":     &WallMoveCommand{},
		"wormhole": &WormholeMoveCommand{},
	},
	"river": {
		"wall":     &WallMoveCommand{},
		"wormhole": &WormholeMoveCommand{},
	},
	"wormhole": {
		"wall":  &WormholeMoveCommand{},
		"river": &RiverMoveCommand{},
	},
}

type MoveCommand struct {
	Direction MoveDirection
}

type MoveCommandType interface {
	Do(w *World, p *Player, direction MoveDirection) []Event
}

func (c *MoveCommand) Do(w *World, p *Player) []Event {
	cell := w.Cells.Get(p.Pos)
	nextCoo := p.Pos.Next(c.Direction)
	nextCell := w.Cells.Get(nextCoo)
	if nextCell == nil {
		//TODO error crash
		e := Event("there is no cell there")
		w.Emmit(e)
		return []Event{e}
	}

	routeFromMap := moveRouting[cell.Class]
	if routeFromMap == nil {
		return SimpleMoveCommand{}.Do(w, p, c.Direction)

	}

	mvCmd := routeFromMap[nextCell.Class]
	if mvCmd == nil {
		return SimpleMoveCommand{}.Do(w, p, c.Direction)
	}
	return mvCmd.Do(w, p, c.Direction)
}

type RiverMoveCommand struct{}

func (rm RiverMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	nextPos := p.Pos.Next(direction)
	nextCell := w.Cells.Get(nextPos)

	nextRiverCell, ok := nextCell.Custom.(*RiverCell)
	if !ok {
		panic("assertion failed")
	}

	return rm.interact(w, p, nextRiverCell, nextPos)
}

func (rm RiverMoveCommand) interact(w *World, p *Player, recCtxRiverCell *RiverCell, recCtxPos Position) []Event {

	var recCtxCounter int
	var recEvents []Event

	for {
		p.Pos = recCtxPos

		if recCtxRiverCell.isMouth {
			e := NewEventf("mouth of the river")
			recEvents = append(recEvents, e)
			w.Emmit(e)
			break
		}

		if recCtxCounter > 1 {
			break
		}

		if recCtxCounter == 0 {
			if p.Hand != Nothing {
				e := NewEventf("You've fallen into a River - and dropped a Treasure")
				recEvents = append(recEvents, e)
				w.Emmit(e)
			} else {
				e := NewEventf("You've fallen into a River and dragged downstream")
				recEvents = append(recEvents, e)
				w.Emmit(e)
			}
		} else {
			e := NewEventf("Dragged downstream")
			recEvents = append(recEvents, e)
			w.Emmit(e)
		}

		recCtxPos = recCtxPos.Next(recCtxRiverCell.Dir)
		nextRiverCell := w.Cells.Get(recCtxPos)
		nextRiver, ok := nextRiverCell.Custom.(*RiverCell)
		if !ok {
			e := NewEventf("ERROR: can't cast to river " + nextRiverCell.Class)
			recEvents = append(recEvents, e)
			w.Emmit(e)
			break
		}

		recCtxRiverCell = nextRiver
		recCtxCounter++
		continue
	}

	return recEvents
}

type WormholeMoveCommand struct{}

func (rm WormholeMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	nextPos := p.Pos.Next(direction)
	nextCell := w.Cells.Get(nextPos)

	wormholeCell, ok := nextCell.Custom.(*WormholeCell)
	if !ok {
		if nextCell.Class == "wall" {
			wormholeCell, ok := w.Cells.Get(p.Pos).Custom.(*WormholeCell)
			if !ok {
				panic("can not handle this")
			}
			p.Pos = wormholeCell.NextPos
			e := NewEventf("you meet the wall, and teleported again")
			w.Emmit(e)
			return []Event{e}
		} else {
			//TODO: rethink this place
			panic("unexpected state")
		}
	}

	p.Pos = wormholeCell.NextPos
	e := NewEventf("teleported")
	w.Emmit(e)
	return []Event{e}
}
