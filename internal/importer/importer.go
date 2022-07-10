package importer

import (
	"os"
	"strings"

	"github.com/go-faster/errors"

	"github.com/xenking/vilingvum/pkg/csv"
)

func ReadRows(f *os.File) ([]*Row, error) {
	var rows []*Row
	if err := csv.NewDecoder(f).Decode(&rows); err != nil {
		return nil, err
	}

	for i, row := range rows {
		if row.Day == "" {
			return rows[:i], nil
		}
	}

	return rows, nil
}

type Row struct {
	Level      string
	Day        string
	Title      string
	Topic      string
	Exercise1  exercise1
	Exercise2  exercise2
	Exercise3  exercise3
	Dictionary dictionary
	VideoURLs  []string
}

func (r *Row) UnmarshalCSV(_, values []string) error {
	if len(values) < 31 {
		return errors.New("invalid row")
	}

	r.Level = values[0]
	r.Day = values[1]
	r.Title = values[2]
	r.Topic = values[3]
	r.Exercise1 = exercise1{
		Title:       strings.TrimSpace(values[4]),
		Description: strings.TrimSpace(values[5]),
		Question:    values[6],
		Correct:     values[7],
	}
	for _, value := range values[8:11] {
		if value != "" {
			r.Exercise1.Incorrect = append(r.Exercise1.Incorrect, value)
		}
	}
	r.Exercise2 = exercise2{
		Title:       strings.TrimSpace(values[11]),
		Description: strings.TrimSpace(values[12]),
		Correct:     values[13],
		Question:    values[21],
	}
	for _, value := range values[14:21] {
		if value != "" {
			r.Exercise2.Incorrect = append(r.Exercise2.Incorrect, value)
		}
	}
	r.Exercise3 = exercise3{
		Title:       strings.TrimSpace(values[22]),
		Description: strings.TrimSpace(values[23]),
		Question:    values[24],
		Correct:     []string{values[25], values[27]},
		Incorrect:   values[26],
	}
	r.Dictionary = dictionary{
		Word:    strings.TrimSpace(values[28]),
		Meaning: strings.TrimSpace(values[29]),
	}
	r.VideoURLs = strings.Split(values[30], ",")
	if len(r.VideoURLs) == 1 && r.VideoURLs[0] == "" {
		r.VideoURLs = nil
	}
	for i, url := range r.VideoURLs {
		r.VideoURLs[i] = strings.TrimSpace(url)
	}

	return nil
}

type exercise1 struct {
	Title       string
	Description string
	Question    string
	Correct     string
	Incorrect   []string
}

type exercise2 struct {
	Title       string
	Description string
	Correct     string
	Incorrect   []string
	Question    string
}

type exercise3 struct {
	Title       string
	Description string
	Question    string
	Correct     []string
	Incorrect   string
}

type dictionary struct {
	Word    string
	Meaning string
}
