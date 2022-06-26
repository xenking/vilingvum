package application

import (
	"context"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/pkg/utils"
)

type PaginationCallback func(ctx context.Context, page, lastIdx int64) tele.HandlerFunc

const PaginationPageSize = 20

// · 1 ·,      ,      ,      ,
// « 1  , · 2 ·,      ,      ,
// « 1  ,   2  , · 3 ·,      ,
// « 1  ,   2  ,   3  , · 4 ·,
// « 1  ,   2  ,   3  , · 4 ·,
// « 1  ,   2  ,   3  ,   4  , · 5 ·

// · 1 ·,   2  ,   3  ,   4  ,   5 »
// « 1  , · 2 ·,   3  ,   4  ,   5 »
// « 1  ,   2  , · 3 ·,   4  ,   5 »
// « 1  ,   2  ,   3  , · 4 ·,   5 »
// « 1  ,   2  ,   3  ,   4  , · 5 ·

// · 1 ·,   2  ,   3  ,   4 ›,   6 »
// « 1  , · 2 ·,   3  ,   4 ›,   6 »
// « 1  ,   2  , · 3 ·,   4 ›,   6 »
// « 1  , ‹ 3  , · 4 ·,   5  ,   6 »
// « 1  , ‹ 3  ,   4  , · 5 ·,   6 »
// « 1  , ‹ 3  ,   4  ,   5  , · 6 ·

// « 1  , ‹ 3  , · 4 ·,   5 ›,   7 »

func (b *Bot) paginationMenu(ctx context.Context, page, lastIdx int64, cb PaginationCallback) *tele.ReplyMarkup {
	pagination := &tele.ReplyMarkup{ResizeKeyboard: true}
	row := make(tele.Row, 5)
	maxNum := lastIdx / PaginationPageSize

	if maxNum == 0 {
		return nil
	}

	if lastIdx%PaginationPageSize != 0 {
		maxNum++
	}

	if maxNum <= 5 {
		for i := int64(0); i < maxNum; i++ {
			num := utils.WriteUint(i + 1)
			row[i] = pagination.Data(num, num)

			if i == page-1 {
				row[i].Text = "· " + utils.WriteUint(page) + " ·"
			}
		}
	} else {
		row[0] = pagination.Data("« 1", "1")

		num := utils.WriteUint(page - 1)
		row[1] = pagination.Data("‹ "+num, num)

		num = utils.WriteUint(page)
		row[2] = pagination.Data("· "+num+" ·", num)

		num = utils.WriteUint(page + 1)
		row[3] = pagination.Data(num+" ›", num)

		maxPages := utils.WriteUint(maxNum)
		row[4] = pagination.Data(maxPages+" »", maxPages)

		switch page {
		case 1:
			row[0].Text = "· 1 ·"

			row[1].Text = "2"
			row[1].Unique = "2"

			row[2].Text = "3"
			row[2].Unique = "3"

			row[3].Text = "4 ›"
			row[3].Unique = "4"
		case 2:
			row[1].Text = "· 2 ·"
			row[1].Unique = "2"

			row[2].Text = "3"
			row[2].Unique = "3"

			row[3].Text = "4 ›"
			row[3].Unique = "4"
		case maxNum - 1:
			row[1].Text = "‹ " + utils.WriteUint(maxNum-3)
			row[1].Unique = utils.WriteUint(maxNum - 3)

			row[2].Text = utils.WriteUint(maxNum - 2)
			row[2].Unique = row[2].Text

			row[3].Text = "· " + utils.WriteUint(maxNum-1) + " ·"
			row[3].Unique = utils.WriteUint(maxNum - 1)
		case maxNum:
			row[1].Text = "‹ " + utils.WriteUint(maxNum-3)
			row[1].Unique = utils.WriteUint(maxNum - 3)

			row[2].Text = utils.WriteUint(maxNum - 2)
			row[2].Unique = row[2].Text

			row[3].Text = utils.WriteUint(maxNum - 1)
			row[3].Unique = row[3].Text

			row[4].Text = "· " + maxPages + " ·"
			row[4].Unique = maxPages
		}
	}

	for i := range row {
		if row[i].Text != " " {
			row[i].Data = row[i].Unique
			row[i].Unique += "_" + utils.WriteUint(page)

			if !strings.HasPrefix(row[i].Text, "· ") {
				currentPage, _ := utils.ParseUint(row[i].Data)
				b.Handle(&row[i], cb(ctx, currentPage, lastIdx))
			}
		}
	}

	pagination.Inline(row)

	return pagination
}
