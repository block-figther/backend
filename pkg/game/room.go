package game

import (
	"errors"

	"github.com/gofrs/uuid"
)

type Player struct {
	nickname string
}

type Games struct {
}

type Room struct {
	players map[string]*Player
}

var room = Room{
	players: make(map[string]*Player),
}

// Leave for future to implement multi-room feature
func NewRoom() *Room {
	return &Room{}
}

func (room *Room) JoinPlayer(nickname string) (string, error) {
	for _, player := range room.players {
		if player.nickname == nickname {
			return "", errors.New("nickname already taken")
		}
	}

	playerId, err := uuid.NewV4()
	if err != nil {
		return "", errors.New("internal error. Unable player to join game")
	}
	room.players[playerId.String()] = &Player{
		nickname: nickname,
	}

	game.AddPlayer(playerId.String())

	return playerId.String(), nil
}

func (room *Room) RemovePlayer(id string) {
	delete(room.players, id)
	game.RemovePlayer(id)
}

func (room *Room) GetPlayersIds() []string {
	ids := []string{}
	for id := range room.players {
		ids = append(ids, id)
	}
	return ids
}

func (room *Room) GetPlayer(id string) *Player {
	player := room.players[id]
	if player == nil {
		return nil
	}
	return player
}
