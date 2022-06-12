package application

import (
	"context"
	"time"

	"github.com/cornelk/hashmap"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/xenking/vilingvum/config"
	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/application/menu"
	"github.com/xenking/vilingvum/internal/application/users"
	"github.com/xenking/vilingvum/pkg/logger"
)

const (
	PostEntryUnread = "unread"
	PostEntryRead   = "read"
)

type Bot struct {
	*tele.Bot
	PaymentToken string
	forwardIDs   []domain.ForwardID
	db           *database.DB
	users        *users.Store
	actions      *hashmap.HashMap // map[int64]*domain.Action
	retryTopics  *hashmap.HashMap // map[int64]map[int64]*domain.Topic
}

func New(ctx context.Context, cfg config.BotConfig, db *database.DB) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	client, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	store, err := users.InitStore(ctx, db)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		PaymentToken: cfg.PaymentToken,
		Bot:          client,
		db:           db,
		users:        store,
		actions:      &hashmap.HashMap{},
		retryTopics:  &hashmap.HashMap{},
	}

	return bot.RegisterHandlers(ctx)
}

func (b *Bot) RegisterHandlers(ctx context.Context) (*Bot, error) {
	b.Use(LoggerMiddleware(logger.NewModule("bot")))
	b.Use(middleware.AutoRespond())
	b.Use(middleware.IgnoreVia())

	b.Handle("/start", b.HandleStart(ctx))
	b.Handle(tele.OnText, b.OnAction(ctx))
	b.Handle(tele.OnVideo, b.OnAction(ctx))
	b.Handle(tele.OnAudio, b.OnAction(ctx))
	b.Handle(tele.OnVoice, b.OnAction(ctx))
	b.Handle(tele.OnCheckout, b.OnCheckout(ctx))
	b.Handle(tele.OnPayment, b.OnPaymentSuccess(ctx))
	b.OnError = b.HandleError(ctx)

	usersGroup := b.Group()
	usersGroup.Use(IsUserMiddleware(b.users))

	b.InitMenus(ctx)

	adminIDs, err := b.db.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}

	b.forwardIDs = make([]domain.ForwardID, len(adminIDs))
	for i, id := range adminIDs {
		b.forwardIDs[i] = domain.ForwardID(id)
	}

	admin := b.Group()
	admin.Use(middleware.Whitelist(adminIDs...))

	return b, nil
}

func (b *Bot) InitMenus(ctx context.Context) {
	menu.Main = &tele.ReplyMarkup{ResizeKeyboard: true}
	menu.Guest = &tele.ReplyMarkup{ResizeKeyboard: true}

	btnCurrentTopic := menu.Main.Text("🎓 Current topic")
	btnDict := menu.Main.Text("📔 Dictionary")
	btnPrevTopics := menu.Main.Text("🔄 Previous topics")
	btnAbout := menu.Main.Text("ℹ About me")
	btnFeedback := menu.Main.Text("📝 Feedback")
	btnSubscribe := menu.Main.Text("💳 Subscribe")

	menu.Main.Reply(
		menu.Main.Row(btnCurrentTopic),
		menu.Main.Row(btnDict, btnPrevTopics),
		menu.Main.Row(btnAbout, btnFeedback),
	)

	menu.Guest.Reply(
		menu.Main.Row(btnCurrentTopic),
		menu.Main.Row(btnDict, btnPrevTopics),
		menu.Main.Row(btnAbout, btnSubscribe),
	)

	b.Handle(&btnCurrentTopic, b.HandleGetCurrentTopic(ctx))
	b.Handle(&btnDict, b.GetDict(ctx))
	b.Handle(&btnPrevTopics, b.HandleGetPrevTopics(ctx))
	b.Handle(&btnAbout, b.HandleAbout(ctx))
	b.Handle(&btnFeedback, b.HandleFeedback(ctx))
	b.Handle(&btnSubscribe, b.HandleSubscribe(ctx))

	menu.Admin = &tele.ReplyMarkup{ResizeKeyboard: true}
	btnUsers := menu.Admin.Text("🔎 Users")

	menu.Admin.Reply(
		menu.Admin.Row(btnUsers),
	)

	b.Handle(&btnUsers, b.GetUserInfo(ctx))
}
