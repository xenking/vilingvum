package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/application/menu"
	"github.com/xenking/vilingvum/pkg/utils"
)

func (b *Bot) HandleSubscribe(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user != nil && user.ActiveUntil != nil {
			return c.Send("You are already subscribed", menu.Main)
		}

		var sb strings.Builder

		invoiceUUID := uuid.New()
		payload := domain.PaymentData{
			Receipt: domain.Receipt{
				Items: []domain.ReceiptItem{
					{
						Description: "Bot subscription",
						Quantity:    "1",
						Amount: domain.ReceiptItemAmount{
							Value:    utils.WriteUint(100),
							Currency: "RUB",
						},
						VAT: domain.VATNo,
					},
				},
			},
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		sb.Write(data)

		invoice := &tele.Invoice{
			Title:       "Subscription to bot",
			Description: "Please pay for subscription to bot",
			Payload:     invoiceUUID.String(),
			Currency:    "RUB",
			Data:        sb.String(),
			Prices: []tele.Price{{
				Label:  "Bot subscription",
				Amount: 10000,
			}},
			Token:     b.PaymentToken,
			Total:     10000,
			NeedEmail: true,
			SendEmail: true,
		}

		err = b.db.CreateInvoice(ctx, &database.CreateInvoiceParams{
			Uuid:   invoiceUUID,
			UserID: user.ID,
			Payload: pgtype.JSONB{
				Bytes:  data,
				Status: pgtype.Present,
			},
		})
		if err != nil {
			return err
		}

		//nolint:gocritic
		//markup := &tele.ReplyMarkup{
		//	ResizeKeyboard:  true,
		//}
		//
		//btnPay := markup.Data("Pay","pay")
		//markup.Inline(markup.Row(btnPay))
		//
		//return c.Send(invoice, markup)

		return c.Send(invoice)
	}
}

func (b *Bot) OnCheckout(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return c.Accept("You are not authorized")
		}
		if user.ActiveUntil != nil {
			if err := c.Accept("You are already subscribed"); err != nil {
				return err
			}

			return c.Send(fmt.Sprintf("Hey %s!", user.Name), menu.Main)
		}

		defer c.DeleteAfter(domain.PaymentDeleteDelay)

		query := c.PreCheckoutQuery()
		invoiceUUID, err := uuid.Parse(query.Payload)
		if err != nil {
			return c.Accept("Invalid invoice id")
		}
		if query.Order.Email == "" {
			return c.Accept("Invalid email")
		}

		err = b.db.AddInvoiceDetails(ctx, &database.AddInvoiceDetailsParams{
			Uuid:  invoiceUUID,
			Email: &query.Order.Email,
		})
		if err != nil {
			return c.Accept(err.Error())
		}

		return c.Accept()
	}
}

func (b *Bot) OnPaymentSuccess(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		payment := c.Message().Payment

		invoiceUUID, err := uuid.Parse(payment.Payload)
		if err != nil {
			return errors.New("Invalid invoice id")
		}

		err = b.db.CompleteInvoice(ctx, &database.CompleteInvoiceParams{
			Uuid:     invoiceUUID,
			ChargeID: &payment.ProviderChargeID,
		})
		if err != nil {
			return err
		}

		err = b.users.UpdateLicense(ctx, c.Sender().ID, payment.Order.Email, payment.Order.PhoneNumber)
		if err != nil {
			return err
		}

		return c.Send("Thank you for your subscription!", menu.Main)
	}
}
