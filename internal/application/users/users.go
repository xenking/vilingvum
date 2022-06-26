package users

import (
	"context"
	"strings"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/go-faster/errors"
	"github.com/jackc/pgx/v4"
	"github.com/phuslu/log"
	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
)

type Store struct {
	globalCtx    context.Context
	db           *database.DB
	cache        *hashmap.HashMap // map[int64]*domain.User
	currentTopic *hashmap.HashMap // map[int64]int64
}

func InitStore(ctx context.Context, db *database.DB) (*Store, error) {
	cache := &hashmap.HashMap{}
	currentTopic := &hashmap.HashMap{}

	dbUsers, err := db.ListActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, dbUser := range dbUsers {
		u := &domain.User{
			ID:          dbUser.ID,
			Name:        dbUser.Name,
			IsAdmin:     dbUser.IsAdmin,
			ActiveUntil: dbUser.ActiveUntil,
		}

		topicID, dbErr := db.GetLastTopicID(ctx, u.ID)
		if dbErr != nil {
			if errors.Is(dbErr, pgx.ErrNoRows) {
				topicID = domain.StartTopicID
			} else {
				return nil, dbErr
			}
		}

		currentTopic.Set(u.ID, topicID)
		cache.Set(dbUser.ID, u)
	}

	return &Store{
		globalCtx:    ctx,
		db:           db,
		cache:        cache,
		currentTopic: currentTopic,
	}, nil
}

func (s *Store) Get(id int64) *domain.User {
	if usr, exists := s.cache.Get(id); exists {
		return usr.(*domain.User)
	}

	dbUser, err := s.db.GetUser(s.globalCtx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Debug().Err(err).Int64("user_id", id).Msg("GetUser")
		}

		return nil
	}

	usr := &domain.User{
		ID:      dbUser.ID,
		Name:    dbUser.Name,
		IsAdmin: dbUser.IsAdmin,
	}

	s.cache.Set(dbUser.ID, usr)

	return usr
}

func (s *Store) Add(ctx context.Context, user *tele.User) (*domain.User, error) {
	userName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	req := &database.CreateUserParams{
		ID:       user.ID,
		Name:     userName,
		Username: user.Username,
		State:    string(domain.UserStatusActive),
	}

	dbUser, err := s.db.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	usr := &domain.User{
		ID:      dbUser.ID,
		Name:    dbUser.Username,
		IsAdmin: dbUser.IsAdmin,
	}
	s.cache.Set(usr.ID, usr)
	s.currentTopic.Set(usr.ID, domain.StartTopicID)

	return usr, nil
}

func (s *Store) UpdateLicense(ctx context.Context, id int64, email, phoneNumber string) error {
	active := time.Now().Add(domain.UserSubscriptionDuration)
	args := &database.UpdateUserSubscriptionParams{
		ID:          id,
		ActiveUntil: &active,
	}

	if email != "" {
		args.Email = &email
	}

	if phoneNumber != "" {
		args.PhoneNumber = &phoneNumber
	}

	err := s.db.UpdateUserSubscription(ctx, args)
	if err != nil {
		return err
	}

	u, ok := s.cache.Get(id)
	if user, ok2 := u.(*domain.User); ok && ok2 {
		user.ActiveUntil = &active
	}

	return nil
}

func (s *Store) GetTopicID(userID int64) int64 {
	if id, exists := s.currentTopic.Get(userID); exists {
		return id.(int64)
	}

	topicID, err := s.db.GetLastTopicID(s.globalCtx, userID)
	if err != nil {
		return -1
	}

	return topicID
}

func (s *Store) SetTopicID(userID, topicID int64) {
	s.currentTopic.Set(userID, topicID)
}

func (s *Store) NextTopicID(userID int64) {
	topicID := s.GetTopicID(userID) + 1

	s.currentTopic.Set(userID, topicID)
}
