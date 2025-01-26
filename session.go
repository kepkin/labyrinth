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
	World   *World
	Players []*Player

	currentPlayer CycledInt
}

func (s *Session) AddPlayer(name string, p Position) {
	s.Players = append(s.Players, &Player{Name: name, Pos: p})
	s.currentPlayer.SetMax(int64(len(s.Players)))
}

func (s *Session) GetCurrentPlayer() *Player {
	if s.currentPlayer.max != int64(len(s.Players)) {
		s.currentPlayer.max = int64(len(s.Players))
	}
	return s.Players[s.currentPlayer.Current()]
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
		s.World.Emmit(NewEventf("%s picks up %v", p.Name, object))

		return nil
	}

	p := s.GetCurrentPlayer()

	dir, err := MoveDirectionFromWord(text)
	if err != nil {
		return []Event{"impossible move"}
	}
	s.World.Emmit(NewEventf("%s goes %s", p.Name, dir.String()))

	mc := MoveCommand{
		Direction: dir,
	}

	ev := mc.Do(s.World, p)
	s.currentPlayer.Next()

	return ev
}
