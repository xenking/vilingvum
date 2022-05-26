package calendar

import (
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/pkg/utils"
)

const (
	BtnPrev = "<"
	BtnNext = ">"
)

func GenerateCalendar(year int64, month time.Month) *tele.ReplyMarkup {
	keyboard := &tele.ReplyMarkup{}

	rows := []tele.Row{
		addMonthYearRow(year, month, keyboard),
		addDaysNamesRow(keyboard),
	}

	rows = append(rows, generateMonth(year, int64(month), keyboard)...)
	rows = append(rows, keyboard.Row(keyboard.Data(BtnPrev, BtnPrev), keyboard.Data(BtnNext, BtnNext)))
	keyboard.Inline(rows...)

	return keyboard
}

func HandlerPrevButton(year int64, month time.Month) (*tele.ReplyMarkup, int64, time.Month) {
	if month != 1 {
		month--
	} else {
		month = 12
		year--
	}

	return GenerateCalendar(year, month), year, month
}

func HandlerNextButton(year int64, month time.Month) (*tele.ReplyMarkup, int64, time.Month) {
	if month != 12 {
		month++
	} else {
		year++
	}

	return GenerateCalendar(year, month), year, month
}

func addMonthYearRow(year int64, month time.Month, keyboard *tele.ReplyMarkup) tele.Row {
	return keyboard.Row(
		keyboard.Data(
			strings.Join([]string{month.String(), utils.WriteUint(year)}, " "),
			"1"),
	)
}

func addDaysNamesRow(keyboard *tele.ReplyMarkup) tele.Row {
	days := [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	buttons := make([]tele.Btn, len(days))

	for i, day := range days {
		buttons[i] = keyboard.Data(day, day)
	}

	return buttons
}

func generateMonth(year, month int64, keyboard *tele.ReplyMarkup) []tele.Row {
	firstDay := getDate(year, month, 0)
	amountDaysInMonth := getDate(year, month+1, 0).Day()
	weekday := int64(firstDay.Weekday())

	var rows []tele.Row
	var row []tele.Btn

	for i := int64(1); i <= weekday; i++ {
		btn := keyboard.Data(" ", utils.WriteUint(i))
		row = append(row, btn)
	}

	amountWeek := weekday
	for day := int64(1); day <= int64(amountDaysInMonth); day++ {
		if amountWeek == 7 {
			rows = append(rows, row)
			amountWeek = 0
			row = nil
		}

		uniqueDate := strings.Join([]string{
			utils.WriteUint(year),
			utils.WriteUint(month),
			utils.WriteUint(day),
		}, "-")

		btn := keyboard.Data(utils.WriteUint(day), uniqueDate)
		row = append(row, btn)
		amountWeek++
	}

	for i := int64(1); i <= 7-amountWeek; i++ {
		btn := keyboard.Data(" ", utils.WriteUint(i))
		row = append(row, btn)
	}

	rows = append(rows, row)

	return rows
}

func getDate(year, month, day int64) time.Time {
	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
}
