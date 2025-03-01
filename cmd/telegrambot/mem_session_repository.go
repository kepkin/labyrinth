package main

import (
	"fmt"
	"slices"
	"sync"

	lab "github.com/kepkin/labyrinth"
)

type MemSessionRepository struct {
	store map[string]*MemSession

	userToActiveSessionMap map[ChatUserID]string

	m sync.RWMutex
}

func (s *MemSessionRepository) FindSession(id string) (*MemSession, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	res := s.store[id]
	return res, nil
}

func (s *MemSessionRepository) GetSessionInPrepareMode(id string) (*MemSession, error) {
	s.m.RLock()
	res := s.store[id]
	s.m.RUnlock()
	if res == nil {
		s.m.Lock()
		defer s.m.Unlock()
		s.store[id] = &MemSession{}
		return s.store[id], nil
	}

	return res, nil
}

func (s *MemSessionRepository) StartSession(id string) error {
	s.m.Lock()
	defer s.m.Unlock()

	res := s.store[id]
	if res == nil {
		return fmt.Errorf("no such sesssion")
	}

	for _, x := range res.Users {
		s.userToActiveSessionMap[x.ID] = id
	}

	return nil
}

func (s *MemSessionRepository) JoinUserToSession(sessionID string, user TgUser, p lab.Position) (*MemSession, error) {
	sess, err := s.GetSessionInPrepareMode(sessionID)
	if err != nil {
		return nil, err
	}
	s.m.Lock()
	defer s.m.Unlock()

	s.userToActiveSessionMap[user.ID] = sessionID
	err = sess.Join(user, p)

	return sess, err
}

func (s *MemSessionRepository) GetActiveSessionForUser(userID int64) (*MemSession, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if idx, ok := s.userToActiveSessionMap[userID]; ok {
		return s.store[idx], nil
	}

	return nil, fmt.Errorf("user not in a game")
}

func (s *MemSessionRepository) StopSession(userID int64) {
	s.m.Lock()
	defer s.m.Unlock()

	if idx, ok := s.userToActiveSessionMap[userID]; ok {
		gameSession := s.store[idx]
		for _, x := range gameSession.Users {
			delete(s.userToActiveSessionMap, x.ID)
		}

		delete(s.store, idx)
	}
}

func (s *MemSessionRepository) RemoveUserFromSession(userID int64) error {
	s.m.Lock()
	defer s.m.Unlock()

	if idx, ok := s.userToActiveSessionMap[userID]; ok {
		gameSession := s.store[idx]

		gameSession.Users = slices.DeleteFunc(gameSession.Users, func(e TgUser) bool {
			return e.ID == userID
		})

		if len(gameSession.Users) == 0 {
			delete(s.store, idx)
		}
	}

	return nil
}
