package labyrinth

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

func (s *Session) Do(text string) []Event {
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
