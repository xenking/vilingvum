package application

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/go-faster/errors"
	"github.com/goccy/go-json"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"tgbot/config"
	"tgbot/database"
	"tgbot/internal/application/middlewares"
	"tgbot/pkg/logger"
	"tgbot/pkg/utils"
)

type Action int8

const (
	ActionNone Action = iota
	ActionRegister
	ActionCreatePost
	ActionCreateInvite
)

type User struct {
	ID            int64
	LastPostID    int64
	CurrentAction Action
	IsAdmin       bool
	Name          string
	Settings      UserSetting
}

type UserSetting struct {
	NextPostID   int64         `json:"next_post_id"`
	PostInterval time.Duration `json:"post_interval"`
	NextPost     time.Time     `json:"next_post"`
}

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

const (
	PostEntryUnread = "unread"
	PostEntryRead   = "read"
)

type Bot struct {
	*tele.Bot
	db    *database.DB
	auth  *hashmap.HashMap // map[int64]int
	users *hashmap.HashMap // map[int64]*User
}

func New(ctx context.Context, cfg config.BotConfig, db *database.DB) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot:   b,
		db:    db,
		auth:  &hashmap.HashMap{},
		users: &hashmap.HashMap{},
	}

	return bot.RegisterHandlers(ctx)
}

func (b *Bot) RegisterHandlers(ctx context.Context) (*Bot, error) {
	b.Use(middlewares.Logger(logger.NewModule("bot")))
	b.Use(middleware.AutoRespond())
	b.Use(middleware.IgnoreVia())

	b.Handle(tele.OnText, b.onText(ctx))

	b.Handle("/start", b.onStartCmd(ctx))
	b.Handle("/register", b.onRegisterCmd(ctx))

	dbUsers, err := b.db.ListActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range dbUsers {
		u := &User{
			ID:      user.ID,
			Name:    user.Name,
			IsAdmin: user.IsAdmin,
		}

		_ = json.Unmarshal(user.Settings.Bytes, &u.Settings)

		b.users.Set(user.ID, u)
	}

	users := b.Group()
	users.Use(middlewares.IsUser(b.users))

	b.InitMenus(ctx)

	admins, err := b.db.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}

	admin := b.Group()
	admin.Use(middleware.Whitelist(admins...))

	return b, nil
}

var (
	authMenu  *tele.ReplyMarkup
	mainMenu  *tele.ReplyMarkup
	adminMenu *tele.ReplyMarkup
)

func (b *Bot) InitMenus(ctx context.Context) {
	authMenu = &tele.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
	btnRegister := authMenu.Data("📝 Register", "register")

	authMenu.Inline(mainMenu.Row(btnRegister))

	b.Handle(&btnRegister, b.onRegisterBtn(ctx))

	mainMenu = &tele.ReplyMarkup{ResizeKeyboard: true}

	btnHelp := mainMenu.Text("ℹ Help")
	btnSettings := mainMenu.Text("⚙ Settings")
	btnGetPost := mainMenu.Text("🔄 Get post")

	mainMenu.Reply(
		mainMenu.Row(btnHelp, btnSettings),
		mainMenu.Row(btnGetPost),
	)

	b.Handle(&btnHelp, b.onHelp(ctx))
	b.Handle(&btnSettings, b.onSettingsBtn(ctx))
	b.Handle(&btnGetPost, b.onGetPostBtn(ctx))

	adminMenu = &tele.ReplyMarkup{ResizeKeyboard: true}
	btnCreateInvite := adminMenu.Text("📩 Invite create")
	btnCreatePost := adminMenu.Text("📝 Create post")

	adminMenu.Reply(
		adminMenu.Row(btnCreateInvite),
		adminMenu.Row(btnCreatePost),
	)

	b.Handle(&btnCreateInvite, b.onCreateInviteBtn(ctx))
	b.Handle(&btnCreatePost, b.onCreatePostBtn(ctx))
}

func (b *Bot) getUser(ctx context.Context, id int64) (*User, error) {
	u, exists := b.users.Get(id)
	if exists {
		return u.(*User), nil
	}

	dbUser, err := b.db.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:      dbUser.ID,
		Name:    dbUser.Name,
		IsAdmin: dbUser.IsAdmin,
	}

	_ = json.Unmarshal(dbUser.Settings.Bytes, &user.Settings)

	b.users.Set(id, user)

	return user, nil
}

func (b *Bot) getLastUserPostID(ctx context.Context, id int64) (int64, error) {
	user, err := b.getUser(ctx, id)
	if err != nil {
		return 0, err
	}

	if user.LastPostID == 0 {
		postID, err := b.db.GetLastPostID(ctx, id)
		if err != nil {
			return 0, err
		}

		user.LastPostID = postID
	}

	return user.LastPostID, nil
}

func (b *Bot) sendPost(ctx context.Context, c tele.Context, postID int64, content []byte) error {
	post := &Post{}
	err := json.Unmarshal(content, post)
	if err != nil {
		return c.Send(Error{Err: err})
	}

	err = b.db.CreatePostEntry(ctx, &database.CreatePostEntryParams{
		UserID: c.Sender().ID,
		PostID: postID,
		Status: PostEntryUnread,
	})
	if err != nil {
		return c.Send(Error{Err: err})
	}

	postMenu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnReadPost := postMenu.Data("✅ I read", "read_post", utils.WriteUint(postID))
	btnNextPost := postMenu.Data("➡ Next post", "next_post", utils.WriteUint(postID))

	postMenu.Inline(
		postMenu.Row(btnReadPost, btnNextPost),
	)

	b.Handle(&btnReadPost, b.onReadPost(ctx, postMenu))
	b.Handle(&btnNextPost, b.onNextPost(ctx, postMenu))

	return c.Send(post, postMenu)
}

func (b *Bot) registerUser(ctx context.Context, user *tele.User, invite string) error {
	usedAt, err := b.db.GetInviteCode(ctx, invite)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("Invite code is not valid")
		}

		return err
	}

	if usedAt != nil && !usedAt.IsZero() {
		return errors.New("Invite code is already used")
	}

	userName := strings.Join([]string{user.FirstName, user.LastName}, " ")
	req := &database.CreateUserParams{
		ID:          user.ID,
		Name:        userName,
		Username:    user.Username,
		InviteCode:  invite,
		State:       UserStatusActive,
		ActiveUntil: time.Now().AddDate(0, 1, 0),
	}

	dbUser, err := b.db.CreateUser(ctx, req)
	if err != nil {
		return err
	}

	b.users.Set(user.ID, &User{
		ID:   user.ID,
		Name: userName,
	})

	err = b.db.ActivateInviteCode(ctx, &database.ActivateInviteCodeParams{
		Code: invite,
		UsedBy: sql.NullInt64{
			Int64: dbUser.ID,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	err = b.db.CreatePostEntry(ctx, &database.CreatePostEntryParams{
		UserID: user.ID,
		PostID: 1,
		Status: PostEntryUnread,
	})

	return err
}

func (b *Bot) createPost(ctx context.Context, p *Post) (int64, error) {
	buf, err := json.Marshal(p)

	postID, err := b.db.CreatePost(ctx, pgtype.JSONB{
		Bytes:  buf,
		Status: pgtype.Present,
	})
	if err != nil {
		return 0, err
	}

	return postID, nil
}

func (b *Bot) createInvite(ctx context.Context, userID int64, invite string) error {
	return b.db.CreateInviteCode(ctx, &database.CreateInviteCodeParams{
		Code: invite,
		CreatedBy: sql.NullInt64{
			Int64: userID,
			Valid: true,
		},
	})
}

type Error struct {
	Err     error
	Message string
}

func (e Error) Error() string {
	return e.Err.Error()
}

func (e Error) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	if e.Message == "" {
		e.Message = "Unexpected error occurred. Contact the bot owner.\nError: " + e.Err.Error()
	}

	return bot.Send(recipient, e.Message, options)
}

type Post struct {
	Data string `json:"data"`
}

func (p *Post) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	return bot.Send(recipient, p.Data, options)
}
