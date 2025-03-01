package labyrinth

type SimpleMoveCommand struct{}

func (c SimpleMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	evs := []Event{}

	nextCoo := p.Pos.Next(direction)
	p.Pos = nextCoo
	nextCell := w.Cells.Get(nextCoo)

	e := NewEventf2(LearnCellEventType, p.Name, nextCell.Class)
	evs = append(evs, e)
	w.Emmit(e)

	for _, v := range nextCell.Items {
		e := NewEventf2(FoundObjectEventType, p.Name, v.Name)
		evs = append(evs, e)
		w.Emmit(e)
	}

	return evs
}

type WallMoveCommand struct{}

func (c WallMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	e := NewEventf2(LearnCellEventType, p.Name, CellWall)
	w.Emmit(e)
	return []Event{e}
}

type ExitMoveCommand struct{}

func (c ExitMoveCommand) Do(w *World, p *Player, direction MoveDirection) []Event {
	se := SimpleMoveCommand{}.Do(w, p, direction)

	if p.Hand == nil {
		return se
	}

	if p.Hand.ID == Treasure {
		e := NewEventf2(RevealObjectEventType, p.Name, "genuine")
		e2 := NewEventf2(WinEventType, p.Name, "")
		w.Emmit(e)
		w.Emmit(e2)
		se = append(se, e)
		se = append(se, e2)
	} else if p.Hand.ID == FakeTreasure {
		e := NewEventf2(RevealObjectEventType, p.Name, "fake")
		se = append(se, e)
		w.Emmit(e)
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
		e := NewEventf2(ErrorEventType, "", "there is no cell there")
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
			e := NewEventf2(LearnCellEventType, p.Name, CellRiverMouth)
			recEvents = append(recEvents, e)
			w.Emmit(e)
			break
		}

		if recCtxCounter > 1 {
			break
		}

		if recCtxCounter == 0 {
			e := NewEventf2(RiverDragEventType, p.Name, CellRiver)
			recEvents = append(recEvents, e)
			w.Emmit(e)

			if p.Hand != nil {
				e2 := NewEventf2(LooseObjectEventType, p.Name, p.Hand.Name)
				recEvents = append(recEvents, e2)
				w.Emmit(e2)
			}
		} else {
			e := NewEventf2(RiverDragEventType, p.Name, CellRiver)
			recEvents = append(recEvents, e)
			w.Emmit(e)
		}

		recCtxPos = recCtxPos.Next(recCtxRiverCell.Dir)
		nextRiverCell := w.Cells.Get(recCtxPos)
		nextRiver, ok := nextRiverCell.Custom.(*RiverCell)
		if !ok {
			e := NewEventf2(ErrorEventType, p.Name, "ERROR: can't cast to river "+nextRiverCell.Class)
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
		if nextCell.Class == CellWall {
			wormholeCell, ok := w.Cells.Get(p.Pos).Custom.(*WormholeCell)
			if !ok {
				panic("can not handle this")
			}
			p.Pos = wormholeCell.NextPos
			e := NewEventf2(LearnCellEventType, p.Name, CellWall)
			e2 := NewEventf2(TeleportEventType, p.Name, "")
			w.Emmit(e)
			w.Emmit(e2)
			return []Event{e, e2}
		} else {
			//TODO: rethink this place
			e := NewEventf2(ErrorEventType, p.Name, "unexepcted state")
			return []Event{e}
		}
	}

	p.Pos = wormholeCell.NextPos
	e := NewEventf2(TeleportEventType, p.Name, "")
	w.Emmit(e)
	return []Event{e}
}
