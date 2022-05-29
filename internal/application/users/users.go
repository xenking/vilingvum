package users

import (
	"context"
	"strings"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/go-faster/errors"
	"github.com/goccy/go-json"
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

const StartTopicID int64 = 1

func InitStore(ctx context.Context, db *database.DB) (*Store, error) {
	cache := &hashmap.HashMap{}
	currentTopic := &hashmap.HashMap{}

	dbUsers, err := db.ListActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, dbUser := range dbUsers {
		u := &domain.User{
			ID:      dbUser.ID,
			Name:    dbUser.Name,
			IsAdmin: dbUser.IsAdmin,
		}

		err = json.Unmarshal(dbUser.Settings.Bytes, &u.Settings)
		if err != nil {
			log.Debug().Err(err).Int64("user_id", u.ID).
				Str("data", string(dbUser.Settings.Bytes)).Msg("Unmarshal settings")
		}

		err = json.Unmarshal(dbUser.Settings.Bytes, &u.Settings)
		if err != nil {
			log.Debug().Err(err).Int64("user_id", u.ID).
				Str("data", string(dbUser.Settings.Bytes)).Msg("Unmarshal settings")
		}

		topicID, dbErr := db.GetLastTopicID(ctx, u.ID)
		if dbErr != nil {
			if errors.Is(dbErr, pgx.ErrNoRows) {
				topicID = StartTopicID
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

	err = json.Unmarshal(dbUser.Settings.Bytes, &usr.Settings)
	if err != nil {
		log.Debug().Err(err).Int64("user_id", usr.ID).
			Str("data", string(dbUser.Settings.Bytes)).Msg("Unmarshal settings")
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
	s.currentTopic.Set(usr.ID, StartTopicID)

	return usr, nil
}

func (s *Store) UpdateLicense(id int64, active time.Time) error {
	err := s.db.SetActiveUser(s.globalCtx, &database.SetActiveUserParams{
		ID:          id,
		ActiveUntil: &active,
	})
	if err != nil {
		return err
	}

	u, ok := s.cache.Get(id)
	if ok {
		user := u.(*database.User)
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
