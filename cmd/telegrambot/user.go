package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/goombaio/namegenerator"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	lab "github.com/kepkin/labyrinth"
	"github.com/kepkin/labyrinth/image"
	labtv "github.com/kepkin/labyrinth/tview"
)

type UserStateRepository struct {
	store *lru.LRU[ChatUserID, UserState]

	mu sync.RWMutex
}

func (usr *UserStateRepository) GetByChatUserID(id ChatUserID) UserState {
	usr.mu.RLock()
	st, ok := usr.store.Get(id)
	usr.mu.RUnlock()
	if ok && st != nil {
		return st
	}

	usr.mu.Lock()
	defer usr.mu.Unlock()

	val := &BaseRouteState{
		Route: map[string]UserState{
			"":      &UserHelpState{},
			"/new":  &NewState{},
			"/join": &JoinState{},
		},
	}
	usr.store.Add(id, val)
	return val
}

func (usr *UserStateRepository) SetUserState(id ChatUserID, st UserState) {
	usr.mu.Lock()
	defer usr.mu.Unlock()

	if st == nil {
		usr.store.Remove(id)
	} else {
		usr.store.Add(id, st)
	}
}

type UserState interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

type UserHelpState struct {
}

func (s *UserHelpState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("To start a new game write `/new`. If you want to join someone's game, ask for a joining code and type /join <code>"),
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type NewState struct {
}

func (s *NewState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	sessionID := nameGenerator.Generate()

	sess, err := sessionRepository.GetSessionInPrepareMode(sessionID)
	if err != nil {
		log.Default().Println(err)
	}
	if sess == nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "there is no such session. You can start a new one with /new",
			ParseMode: models.ParseModeMarkdown,
		})

		if err != nil {
			log.Print(err.Error())
		}
	}

	userStateRepository.SetUserState(update.Message.From.ID, &BaseRouteState{
		Route: map[string]UserState{
			"":      &ChoosePositionState{SessionID: sessionID},
			"info":  &InfoState{SessionID: sessionID},
			"/exit": &ExitState{},
		},
	})

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      fmt.Sprintf("You game code is `%v`\\. Ask you friends to join the game with this code\\. Now choose you position in labyrinth in format: x:y \\(like 3:5\\)\\.", sessionID),
		ParseMode: models.ParseModeMarkdown,
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type JoinState struct {
}

func (s *JoinState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	ss, _ := strings.CutPrefix(update.Message.Text, "/join")
	sessionID := strings.TrimSpace(ss)
	sess, err := sessionRepository.FindSession(sessionID)
	if err != nil {
		log.Default().Println(err)
	}
	if sess == nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "there is no such session. You can start a new one with /new",
		})

		if err != nil {
			log.Print(err.Error())
		}
		return
	}

	userStateRepository.SetUserState(update.Message.From.ID, &BaseRouteState{
		Route: map[string]UserState{
			"":      &ChoosePositionState{SessionID: sessionID},
			"info":  &InfoState{SessionID: sessionID},
			"/exit": &ExitState{},
		},
	})
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Game name: `%v`. Choose your position. Write two numbers (Example: 1:3)", sessionID),
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type ExitState struct {
}

func (s *ExitState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := NewTgUserFromUpdate(update)
	sess, err := sessionRepository.GetActiveSessionForUser(user.ID)
	if err != nil {
		log.Default().Println(err)
	}
	if sess != nil {
		sess.GameSession.RemovePlayer(user.Username)
		for _, x := range sess.Users {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: x.ID,
				Text:   fmt.Sprintf("%v left the game", user.Username),
			})

			if err != nil {
				log.Print(err.Error())
			}
		}
	}

	err = sessionRepository.RemoveUserFromSession(user.ID)
	if err != nil {
		log.Default().Println(err)
	}

	userStateRepository.SetUserState(update.Message.From.ID, nil)
}

type ChoosePositionState struct {
	SessionID string
}

func (s *ChoosePositionState) handleIncorrectFormat(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Inccorecct format. Please enter your desired position in the format X:Y (Example: 1:3)",
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func (s *ChoosePositionState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := NewTgUserFromUpdate(update)
	posValues := strings.Split(update.Message.Text, ":")
	if len(posValues) != 2 {
		s.handleIncorrectFormat(ctx, b, update)
		return
	}

	x, err := strconv.Atoi(posValues[0])
	if err != nil {
		s.handleIncorrectFormat(ctx, b, update)
		return
	}
	y, err := strconv.Atoi(posValues[1])
	if err != nil {
		s.handleIncorrectFormat(ctx, b, update)
		return
	}

	sess, _ := sessionRepository.JoinUserToSession(s.SessionID, user, lab.NewPosition(x, y))

	userStateRepository.SetUserState(user.ID, &BaseRouteState{
		Route: map[string]UserState{
			"":     &WaitForGameStartState{SessionID: s.SessionID},
			"info": &InfoState{SessionID: s.SessionID},
		},
	})

	for _, x := range sess.Users {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: x.ID,
			Text:   fmt.Sprintf("%v joined", user.Username),
		})

		if err != nil {
			log.Print(err.Error())
		}

	}
}

type WaitForGameStartState struct {
	SessionID string
}

func (s *WaitForGameStartState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Text == "/startGame" {
		user := NewTgUserFromUpdate(update)
		sess, err := sessionRepository.GetActiveSessionForUser(user.ID)
		if err != nil {
			log.Default().Println(err)
		}

		w := makeWorld()
		sess.GameSession.World = w

		go func() {
			labtv.RunDebug(w, &sess.GameSession)
		}()

		pl := sess.GameSession.GetCurrentPlayer()
		nextTgUser := TgUser{}
		for _, x := range sess.Users {
			userStateRepository.SetUserState(x.ID, &BaseRouteState{
				Route: map[string]UserState{
					"":      &InGameCommandState{SessionID: s.SessionID},
					"/exit": &ExitState{},
				},
			})
			pl.NewMap()

			if pl.Name == x.Username {
				nextTgUser = x
			} else {
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: x.ID,
					Text:   fmt.Sprintf("Game started. %v move", pl.Name),
				})

				if err != nil {
					log.Print(err.Error())
				}
			}
		}

		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      nextTgUser.ID,
			Text:        "Your turn",
			ReplyMarkup: getInGameMoveReplyKeyboard(sess.GameSession.GetCurrentPlayerPossibleActions()),
		})

		if err != nil {
			log.Print(err.Error())
		}
	}
}

type BaseRouteState struct {
	Route map[string]UserState
}

func (s *BaseRouteState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	prefix, _, _ := strings.Cut(update.Message.Text, " ")

	if next, ok := s.Route[prefix]; ok {
		next.Handle(ctx, b, update)
		return
	}

	if next, ok := s.Route[""]; ok {
		next.Handle(ctx, b, update)
		return
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "unknow command",
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type InfoState struct {
	SessionID string
}

func (s *InfoState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	sess, _ := sessionRepository.GetActiveSessionForUser(update.Message.From.ID)
	if sess == nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No game",
		})

		if err != nil {
			log.Print(err.Error())
		}
		return
	}

	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("You are currently in a game `%v`. Players are:\n", s.SessionID))
	for _, x := range sess.Users {
		msg.WriteString(" - ")
		msg.WriteString(x.Username)
		msg.WriteString("\n")
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg.String(),
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type StartState struct {
}

func (s *StartState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome. Write any abrakadabra to make a new game and send this abrakdabra other players to join in",
	})

	if err != nil {
		log.Print(err.Error())
	}
}

type InGameCommandState struct {
	SessionID string
}

func (s *InGameCommandState) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	user := NewTgUserFromUpdate(update)
	sess, err := sessionRepository.GetActiveSessionForUser(user.ID)
	if err != nil {
		log.Default().Println(err)
	}

	pl := sess.GameSession.GetCurrentPlayer()
	if pl.Name != user.Username {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("It's a %v's turn. Please wait.", pl.Name),
		})

		if err != nil {
			log.Print(err.Error())
		}
	}

	eventStringer := lab.DefaultEventStringer{}

	move := update.Message.Text
	evs := sess.GameSession.Do(move)
	nextPl := sess.GameSession.GetCurrentPlayer()
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("Player %v made a move %v", pl.Name, move))

	isWin := false
	for _, event := range evs {
		msg.WriteString("\n \\- ")
		msg.WriteString(eventStringer.ToString(event))

		if event.Type == lab.WinEventType {
			isWin = true
		}
	}
	msg.WriteString("\n\n```\n")

	lastRow := 0
	for p, c := range sess.GameSession.World.Cells.Rect(pl.Map.LeftCorner, pl.Map.RightCorner) {
		if p.Y > lastRow {
			msg.WriteString("\n")
			lastRow = p.Y
		}

		_, ok := pl.Map.KnonwnCells[p]
		if !ok {
			msg.WriteString(" ")
			continue
		}

		switch c.Class {
		case lab.CellEarth:
			msg.WriteString("e")

		case lab.CellRiver:
			msg.WriteString("r")

		case lab.CellWall:
			msg.WriteString("w")

		case lab.CellWormHole:
			msg.WriteString("o")
		}

	}
	msg.WriteString("\n```")

	ipm := image.NewPlayerMap(makeCellMapImage(), &(pl.Map))

	f := bytes.NewBuffer(nil)
	if err != nil {
		log.Print(err)
	}

	err = jpeg.Encode(f, ipm, nil)
	if err != nil {
		log.Print(err)
	}

	nextTgUser := TgUser{}
	for _, x := range sess.Users {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    x.ID,
			Text:      msg.String(),
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			log.Print(err.Error())
		}

		params := &bot.SendPhotoParams{
			ChatID: x.ID,
			Photo: &models.InputFileUpload{
				Filename: "example.gif",
				Data:     bytes.NewReader(f.Bytes()),
			},
			Caption: "map",
		}

		_, err = b.SendPhoto(ctx, params)

		if err != nil {
			log.Print(err.Error())
		}

		if isWin {
			userStateRepository.SetUserState(x.ID, &JoinState{})
			continue
		}

		if x.Username == nextPl.Name {
			nextTgUser = x
		} else {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: x.ID,
				Text:   fmt.Sprintf("%v turn", nextPl.Name),
			})

			if err != nil {
				log.Print(err.Error())
			}
		}
	}

	if isWin {
		sessionRepository.StopSession(user.ID)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      nextTgUser.ID,
		Text:        "Your turn",
		ReplyMarkup: getInGameMoveReplyKeyboard(sess.GameSession.GetCurrentPlayerPossibleActions()),
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func getInGameMoveReplyKeyboard(actions []string) models.ReplyKeyboardMarkup {
	markup := [][]models.KeyboardButton{
		[]models.KeyboardButton{},
	}

	addButton := func(text string) {
		markup[len(markup)-1] = append(markup[len(markup)-1], models.KeyboardButton{
			Text: text,
		})
	}

	addRow := func() {
		if len(markup[len(markup)-1]) > 0 {
			markup = append(markup, []models.KeyboardButton{})
		}
	}

	addButton("North")
	addRow()
	addButton("West")
	addButton("East")
	addRow()
	addButton("South")

	for _, action := range actions {
		if action == "north" {
			continue
		}

		if action == "south" {
			continue
		}

		if action == "east" {
			continue
		}

		if action == "west" {
			continue
		}

		addRow()
		addButton(action)
	}

	res := models.ReplyKeyboardMarkup{
		Keyboard: markup,
		// Selective:             kb.selective,
		OneTimeKeyboard:       true,
		ResizeKeyboard:        true,
		InputFieldPlaceholder: "",
	}

	return res
}
