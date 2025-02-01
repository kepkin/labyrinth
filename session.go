package labyrinth

import (
	"fmt"
	"strings"
)

func NewCycledInt(max int64, initialValue int64) CycledInt {
	return CycledInt{
		max:   max,
		value: initialValue,
	}
}

type CycledInt struct {
	value int64
	max   int64
}

func (c *CycledInt) Current() int64 {
	return c.value
}

func (c *CycledInt) Next() int64 {
	c.value++
	if c.value >= c.max {
		c.value = 0
	}

	return c.value
}

func (c *CycledInt) SetMax(v int64) {
	c.max = v
}

type Session struct {
	World                *World
	Players              []*Player
	PlayerHasUncertainty []bool

	currentPlayer CycledInt
}

func (s *Session) AddPlayer(name string, p Position) {
	s.Players = append(s.Players, &Player{Name: name, Pos: p})
	s.PlayerHasUncertainty = append(s.PlayerHasUncertainty, false)
	s.currentPlayer.SetMax(int64(len(s.Players)))
}

func (s *Session) GetCurrentPlayer() *Player {
	if s.currentPlayer.max != int64(len(s.Players)) {
		s.currentPlayer.max = int64(len(s.Players))
	}
	return s.Players[s.currentPlayer.Current()]
}

func (s *Session) HookPreMove() {
	res := s.Players[s.currentPlayer.Current()]
	if s.PlayerHasUncertainty[s.currentPlayer.Current()] {
		res.NewMap()
		s.PlayerHasUncertainty[s.currentPlayer.Current()] = false
	}
}

func (s *Session) SetCurrentPlayerUncertainty(uncertainty bool) {
	s.PlayerHasUncertainty[s.currentPlayer.Current()] = uncertainty
}

// Returns possible actions
func (s *Session) GetCurrentPlayerPossibleActions() []string {
	res := []string{"north", "south", "west", "east"}

	p := s.Players[s.currentPlayer.Current()]

	c := s.World.Cells.Get(p.Pos)

	if p.Hand != nil {
		res = append(res, fmt.Sprintf("drop  %v", p.Hand.Name))
	} else {
		for _, v := range c.Items {
			res = append(res, fmt.Sprintf("pick up %v", v.Name))
		}
	}

	return res
}

func (s *Session) Do(text string) []Event {
	if strings.HasPrefix(text, "pick up") {
		object := strings.TrimPrefix(text, "pick up ")

		p := s.GetCurrentPlayer()
		c := s.World.Cells.Get(p.Pos)
		item := c.TakeItem(object)
		p.Hand = item
		s.World.Emmit(NewEventf2(PickObjectEventType, p.Name, item.Name))

		return nil
	}

	p := s.GetCurrentPlayer()

	dir, err := MoveDirectionFromWord(text)
	if err != nil {
		return []Event{NewEventf2(ErrorEventType, p.Name, "impossible move")}
	}
	s.World.Emmit(NewEventf2(MoveEventType, p.Name, dir.String()))
	nextPlayerPos := s.GetCurrentPlayer().Pos.Next(dir)

	mc := MoveCommand{
		Direction: dir,
	}

	s.HookPreMove()
	ev := mc.Do(s.World, p)
	uncertainty := false
	for _, event := range ev {
		if event.Type == RiverDragEventType || event.Type == TeleportEventType {
			uncertainty = true
			break
		}
	}
	s.SetCurrentPlayerUncertainty(uncertainty)

	p.Map.Learn(nextPlayerPos)
	s.currentPlayer.Next()

	return ev
}
