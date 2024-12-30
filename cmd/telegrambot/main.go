package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"

	lab "github.com/kepkin/labyrinth"
	md "github.com/kepkin/labyrinth/markdown"
)

// Send any text message to the bot after the bot has been started

type TgUser struct {
	ID       ChatUserID
	Username string
}

func NewTgUserFromUpdate(u *models.Update) TgUser {
	return TgUser{
		ID:       u.Message.From.ID,
		Username: u.Message.From.Username,
	}
}

type ChatUserID = int64

type SessionID struct {
	id int64
}

type SessionRepository interface {
	GetSessionInPrepareMode(id string) (*MemSession, error)
	JoinUserToSession(sessionID string, user TgUser, p lab.Position) (*MemSession, error)
	StartSession(id string) error
	GetActiveSessionForUser(userID int64) (*MemSession, error)
	StopSession(userID int64)
}

type MemSession struct {
	Users []TgUser

	Started     bool
	GameSession lab.Session
}

func (s *MemSession) Join(user TgUser, p lab.Position) error {
	if s.Started {
		return fmt.Errorf("session started already")
	}
	s.Users = append(s.Users, user)

	s.GameSession.AddPlayer(user.Username, p)
	return nil
}

type MemSessionRepository struct {
	store map[string]*MemSession

	userToActiveSessionMap map[ChatUserID]string

	m sync.RWMutex
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

var sessionRepository SessionRepository
var userStateRepository UserStateRepository

var _world *lab.World

func makeWorld() *lab.World {
	if _world != nil {
		return _world
	}

	worldBytes, err := os.ReadFile("lab-map1.md")
	if err != nil {
		panic(err.Error())
	}
	bb := md.WorldBuilder{}
	_world, _, err := bb.Build(string(worldBytes))
	if err != nil {
		panic(err.Error())
	}

	return _world
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		// bot.WithMessageTextHandler("/play", bot.MatchTypeExact, handlerInlineKeyboard),
		// bot.WithMessageTextHandler("/new", bot.MatchTypeExact, handlerNew),
		// bot.WithMessageTextHandler("/join", bot.MatchTypePrefix, handlerJoin),
		// bot.WithMessageTextHandler("/info", bot.MatchTypeExact, handlerInfo),
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	b, err := bot.New(botToken, opts...)
	if err != nil {
		panic(err)
	}
	initInlineKeyboard(b)

	sessionRepository = &MemSessionRepository{
		store:                  map[string]*MemSession{},
		userToActiveSessionMap: map[ChatUserID]string{},
	}
	userStateRepository = UserStateRepository{
		store: map[ChatUserID]UserState{},
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	st := userStateRepository.GetByChatUserID(update.Message.From.ID)
	st.Handle(ctx, b, update)
}

var demoInlineKeyboard *inline.Keyboard

func initInlineKeyboard(b *bot.Bot) {
	demoInlineKeyboard = inline.New(b, inline.WithPrefix("inline")).
		Row().
		Button("North", []byte("north"), onInlineKeyboardSelect).
		Row().
		Button("West", []byte("west"), onInlineKeyboardSelect).
		Button("East", []byte("east"), onInlineKeyboardSelect).
		Row().
		Button("South", []byte("South"), onInlineKeyboardSelect)
}

func handlerInlineKeyboard(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Select the variant",
		ReplyMarkup: demoInlineKeyboard,
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func onInlineKeyboardSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: mes.Message.Chat.ID,
		Text:   "You selected: " + string(data),
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func handlerInfo(ctx context.Context, b *bot.Bot, update *models.Update) {

	sess, _ := sessionRepository.GetActiveSessionForUser(update.Message.From.ID)
	if sess == nil {

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No game",
		})
		log.Print(err.Error())
		return
	}

	msg := strings.Builder{}
	for _, x := range sess.Users {
		msg.WriteString(" - ")
		msg.WriteString(x.Username)
		msg.WriteString("\n")
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg.String(),
	})

	log.Print(err.Error())
}

var words = []string{
	"pen",
	"string",
	"water",
}
