package application

import (
	"context"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	tele "gopkg.in/telebot.v3"

	"tgbot/database"
	"tgbot/pkg/utils"
)

// Register

func (b *Bot) onStartCmd(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		inviteCode := c.Message().Payload

		if u, exists := b.users.Get(c.Sender().ID); exists {
			user := u.(*User)

			return c.Send(strings.TrimSpace("Hello "+user.Name+"!"), mainMenu)
		}

		if inviteCode == "" {
			return c.Send("You need to register with an invite code", authMenu)
		}

		return b.onRegisterCmd(ctx)(c)
	}
}

func (b *Bot) onRegisterCmd(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := c.Sender()

		exists, err := b.db.IsUserExists(ctx, user.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return c.Send(Error{Err: err})
		}

		if exists {
			return c.Send("You are already registered", mainMenu)
		}

		inviteCode := c.Message().Payload
		if inviteCode == "" {
			return c.Send("Can't register new user Invite code is empty")
		}

		err = b.registerUser(ctx, user, inviteCode)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		return c.Send("You are successfully registered", mainMenu)
	}
}

func (b *Bot) onRegisterBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		userID := c.Sender().ID
		cnt, ok := b.auth.Get(userID)
		var count int
		if ok {
			count = cnt.(int)
		}
		b.auth.Set(userID, count+1)

		return c.Send("Please enter your invite code")
	}
}

func (b *Bot) onText(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		userID := c.Sender().ID
		if cnt, exist := b.auth.Get(userID); exist {
			count := cnt.(int)
			if count > 5 {
				return c.Send("You have exceeded the limit of attempts")
			}
			err := b.registerUser(ctx, c.Sender(), c.Text())
			if err != nil {
				b.auth.Set(userID, count+1)

				return c.Send(err.Error())
			}
			b.auth.Del(userID)

			return c.Send("You are successfully registered", mainMenu)
		}

		user, err := b.getUser(ctx, userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		} else if err != nil {
			return err
		}

		switch user.CurrentAction {
		case ActionRegister:
			err = b.registerUser(ctx, c.Sender(), c.Text())
			if err != nil {
				return c.Send(Error{Err: err})
			}
			user.CurrentAction = ActionNone

			return c.Send("You are successfully registered", mainMenu)
		case ActionCreatePost:
			postID, err := b.createPost(ctx, &Post{Data: c.Text()})
			if err != nil {
				return c.Send(Error{Err: err})
			}
			user.CurrentAction = ActionNone

			return c.Send("Post created successfully\nPost id: "+utils.WriteUint(postID), adminMenu)
		case ActionCreateInvite:
			invite := c.Text()
			if invite == "" {
				return c.Send("Enter invite code")
			}

			err = b.createInvite(ctx, user.ID, invite)
			user.CurrentAction = ActionNone
			if err != nil {
				return c.Send(Error{Err: err})
			}

			return c.Send("Invite created successfully", adminMenu)
		}

		return nil
	}
}

// User

func (b *Bot) onHelp(_ context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("Here is some help:", mainMenu)
	}
}

func (b *Bot) onSettingsBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		var sb strings.Builder

		user, err := b.getUser(ctx, c.Sender().ID)
		if err != nil {
			return err
		}
		sb.WriteString("User settings:\n")
		sb.WriteString("Next post will be sent in: ")
		sb.WriteString(user.Settings.NextPost.Format(time.RFC1123))
		sb.WriteString("\nPost sending interval: ")
		sb.WriteString(user.Settings.PostInterval.String())

		return c.Send(sb.String())
	}
}

// Post

func (b *Bot) onGetPostBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		userID := c.Sender().ID
		postID, err := b.getLastUserPostID(ctx, userID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		post, err := b.db.GetPost(ctx, postID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		return b.sendPost(ctx, c, post.ID, post.Content.Bytes)
	}
}

func (b *Bot) onReadPost(ctx context.Context, menu *tele.ReplyMarkup) tele.HandlerFunc {
	return func(c tele.Context) error {
		fields := strings.Split(c.Data(), "|")
		if len(fields) == 0 {
			return nil
		}

		postID, err := utils.ParseUint(fields[len(fields)-1])
		if err != nil {
			return c.Send(Error{Err: err})
		}

		err = b.db.UpdatePostEntry(ctx, &database.UpdatePostEntryParams{
			PostID: postID,
			UserID: c.Sender().ID,
			Status: PostEntryRead,
		})
		if err != nil {
			return c.Send(Error{Err: err})
		}

		menu.InlineKeyboard[0][0].Text = "❤️ Read"

		return c.Edit(menu)
	}
}

func (b *Bot) onNextPost(ctx context.Context, menu *tele.ReplyMarkup) tele.HandlerFunc {
	return func(c tele.Context) error {
		fields := strings.Split(c.Data(), "|")
		if len(fields) == 0 {
			return nil
		}

		postID, err := utils.ParseUint(fields[len(fields)-1])
		if err != nil {
			return c.Send(Error{Err: err})
		}

		user, err := b.getUser(ctx, c.Sender().ID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		post, err := b.db.GetNextPost(ctx, postID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				user.LastPostID = post.ID

				return c.Send("No more posts")
			}

			return c.Send(Error{Err: err})
		}

		user.LastPostID = post.ID

		return b.sendPost(ctx, c, post.ID, post.Content.Bytes)
	}
}

// Admin

func (b *Bot) onCreateInviteBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := b.getUser(ctx, c.Sender().ID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		user.CurrentAction = ActionCreateInvite

		randomInvite := &tele.ReplyMarkup{ResizeKeyboard: true}
		randomBtn := randomInvite.Data("🎲 Random", "create_invite_random")
		randomInvite.Inline(randomInvite.Row(randomBtn))
		b.Handle(&randomBtn, b.onCreateInviteRandomBtn(ctx))

		return c.Send("Enter invite code", randomInvite)
	}
}

func (b *Bot) onCreateInviteRandomBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := b.getUser(ctx, c.Sender().ID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		invite := ksuid.New().String()

		err = b.createInvite(ctx, user.ID, invite)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		user.CurrentAction = ActionNone

		return c.Send("Invite created: " + invite)
	}
}

func (b *Bot) onCreatePostBtn(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := b.getUser(ctx, c.Sender().ID)
		if err != nil {
			return c.Send(Error{Err: err})
		}

		user.CurrentAction = ActionCreatePost

		return nil
	}
}
