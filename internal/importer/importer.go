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
	Day        string
	Video      string
	Title      string
	Exercise1  exercise1
	Exercise2  exercise2
	Exercise3  exercise3
	Dictionary dictionary
	VideoURLs  []string
}

func (r *Row) UnmarshalCSV(_, values []string) error {
	if len(values) < 30 {
		return errors.New("invalid row")
	}

	r.Day = values[0]
	r.Video = values[1]
	r.Title = values[2]
	r.Exercise1 = exercise1{
		Title:       values[3],
		Description: values[4],
		Question:    values[5],
		Correct:     values[6],
	}
	for _, value := range values[7:10] {
		if value != "" {
			r.Exercise1.Incorrect = append(r.Exercise1.Incorrect, value)
		}
	}
	r.Exercise2 = exercise2{
		Title:       values[10],
		Description: values[11],
		Correct:     values[12],
		Question:    values[20],
	}
	for _, value := range values[13:20] {
		if value != "" {
			r.Exercise2.Incorrect = append(r.Exercise2.Incorrect, value)
		}
	}
	r.Exercise3 = exercise3{
		Title:       values[21],
		Description: values[22],
		Question:    values[23],
		Correct:     []string{values[24], values[26]},
		Incorrect:   values[25],
	}
	r.Dictionary = dictionary{
		Word:    values[27],
		Meaning: values[28],
	}
	r.VideoURLs = strings.Split(values[29], ",")

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
