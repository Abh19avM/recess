package rooms

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// Service defines game room business logic operations.
type Service interface {
	CreateRoom(ctx context.Context, hostID, hostUsername, hostAvatar string, hostRating int, req CreateRoomRequest) (*Room, error)
	GetRoomByCode(ctx context.Context, code string) (*Room, error)
	JoinRoom(ctx context.Context, code, userID, username, avatar string, rating int, passcode string) (*Room, error)
	ListActiveRooms(ctx context.Context) ([]*Room, error)
}

type service struct {
	repo Repository
}

// NewService creates a new room service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateRoom(ctx context.Context, hostID, hostUsername, hostAvatar string, hostRating int, req CreateRoomRequest) (*Room, error) {
	if req.GameType == "" {
		return nil, fmt.Errorf("game_type is required")
	}

	maxPlayers := req.MaxPlayers
	if maxPlayers <= 0 {
		maxPlayers = 2
	}

	title := req.Title
	if title == "" {
		title = fmt.Sprintf("%s's Game", hostUsername)
	}

	roomCode := generateRoomCode()

	newRoom := &Room{
		ID:         "rm_" + strings.ToLower(roomCode),
		Code:       roomCode,
		GameType:   req.GameType,
		Title:      title,
		HostID:     hostID,
		Status:     StatusWaiting,
		MaxPlayers: maxPlayers,
		IsPrivate:  req.IsPrivate,
		Passcode:   req.Passcode,
		Players: []PlayerSlot{
			{
				UserID:       hostID,
				Username:     hostUsername,
				AvatarPreset: hostAvatar,
				Rating:       hostRating,
				IsHost:       true,
				IsReady:      true,
				JoinedAt:     time.Now(),
			},
		},
		Spectators: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, newRoom); err != nil {
		return nil, err
	}

	return newRoom, nil
}

func (s *service) GetRoomByCode(ctx context.Context, code string) (*Room, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	return s.repo.GetByCode(ctx, code)
}

func (s *service) JoinRoom(ctx context.Context, code, userID, username, avatar string, rating int, passcode string) (*Room, error) {
	room, err := s.GetRoomByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if room.IsPrivate && room.Passcode != passcode {
		return nil, ErrInvalidPasscode
	}

	// Check if already in room
	for _, p := range room.Players {
		if p.UserID == userID {
			return room, nil
		}
	}

	if len(room.Players) >= room.MaxPlayers {
		return nil, ErrRoomFull
	}

	room.Players = append(room.Players, PlayerSlot{
		UserID:       userID,
		Username:     username,
		AvatarPreset: avatar,
		Rating:       rating,
		IsHost:       false,
		IsReady:      false,
		JoinedAt:     time.Now(),
	})
	room.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *service) ListActiveRooms(ctx context.Context) ([]*Room, error) {
	return s.repo.ListActive(ctx)
}

func generateRoomCode() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return "RECESS-" + string(b)
}
