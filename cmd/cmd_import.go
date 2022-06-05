package main

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/goccy/go-json"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgtype"
	"github.com/phuslu/log"

	"github.com/xenking/vilingvum/config"
	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/importer"
	"github.com/xenking/vilingvum/pkg/logger"
)

func importCmd(ctx context.Context, flags cmdFlags) error {
	cfg, err := config.NewConfig(flags.Config)
	if err != nil {
		return err
	}

	l := logger.New(cfg.Log)
	logger.SetGlobal(l)
	log.DefaultLogger = *logger.NewModule("global")

	// Check if migration needed
	if cfg.MigrationMode {
		err = migrateDatabase(cfg)
		if err != nil {
			return err
		}
	}

	db, err := database.Init(ctx, cfg.Postgres)
	if err != nil {
		return err
	}

	return importDatabase(ctx, db, cfg.Import)
}

func importDatabase(ctx context.Context, db *database.DB, cfg config.ImportConfig) error {
	f, err := os.Open(cfg.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := importer.ReadRows(f)
	if err != nil {
		return err
	}

	urls := make(map[string]struct{})

	var prevTopicID int64 = -1
	var prevTopicType domain.TopicType
	for _, row := range rows {
		topicType := domain.TopicTypeQuestion
		if strings.HasPrefix(row.Video, "Progress test") {
			topicType = domain.TopicTypeTest
		}

		if prevTopicType == domain.TopicTypeTest &&
			topicType != domain.TopicTypeTest {
			t := &domain.Topic{
				Text: "Write a video or audio where you speak new words from your lessons",
				Type: domain.TopicTypeTestReport,
			}
			prevTopicID, err = createTopic(ctx, db, t, prevTopicID)
			if err != nil {
				return err
			}
		}

		for i, url := range row.VideoURLs {
			if _, ok := urls[url]; ok {
				continue
			}
			urls[url] = struct{}{}

			t := &domain.Topic{
				Type:       domain.TopicTypeVideo,
				VideoURL:   url,
				NextButton: "📝 Start exercises",
			}
			if i == 0 {
				t.Text = row.Title
			}
			if i != len(row.VideoURLs)-1 {
				t.NextButton = "▶️ Next video"
			}
			prevTopicID, err = createTopic(ctx, db, t, prevTopicID)
			if err != nil {
				return err
			}
		}

		if row.Exercise1.Correct != "" {
			t := &domain.Topic{
				Type:     topicType,
				Text:     row.Exercise1.Description,
				Question: row.Exercise1.Question,
				Answers: []domain.TopicAnswer{
					{
						Text:    row.Exercise1.Correct,
						Correct: true,
					},
				},
			}
			for _, value := range row.Exercise1.Incorrect {
				t.Answers = append(t.Answers, domain.TopicAnswer{
					Text: value,
				})
			}
			prevTopicID, err = createTopic(ctx, db, t, prevTopicID)
			if err != nil {
				return err
			}
		}

		if row.Exercise2.Correct != "" {
			t := &domain.Topic{
				Type:     topicType,
				Text:     row.Exercise2.Description,
				Question: row.Exercise2.Question,
				Answers: []domain.TopicAnswer{
					{
						Text:    row.Exercise2.Correct,
						Correct: true,
					},
				},
			}
			for _, value := range row.Exercise2.Incorrect {
				t.Answers = append(t.Answers, domain.TopicAnswer{
					Text: value,
				})
			}
			prevTopicID, err = createTopic(ctx, db, t, prevTopicID)
			if err != nil {
				return err
			}
		}

		if row.Exercise3.Incorrect != "" {
			t := &domain.Topic{
				Type:     topicType,
				Text:     row.Exercise3.Description,
				Question: row.Exercise3.Question,
				Answers: []domain.TopicAnswer{
					{
						Text:    row.Exercise3.Incorrect,
						Correct: true,
					},
				},
			}
			for _, value := range row.Exercise3.Correct {
				t.Answers = append(t.Answers, domain.TopicAnswer{
					Text: value,
				})
			}

			prevTopicID, err = createTopic(ctx, db, t, prevTopicID)
			if err != nil {
				return err
			}
		}

		if row.Dictionary.Word != "" {
			err = db.AddDictionaryRecord(ctx, &database.AddDictionaryRecordParams{
				TopicID: prevTopicID,
				Word:    row.Dictionary.Word,
				Meaning: row.Dictionary.Meaning,
			})
			if err != nil {
				return err
			}
		}

		prevTopicType = topicType
	}

	return nil
}

func createTopic(ctx context.Context, db *database.DB, topic *domain.Topic, prevTopicID int64) (int64, error) {
	b, mErr := json.Marshal(topic)
	if mErr != nil {
		return prevTopicID, mErr
	}

	topicID, err := db.CreateTopic(ctx, &database.CreateTopicParams{
		Type: string(topic.Type),
		Content: pgtype.JSONB{
			Bytes:  b,
			Status: pgtype.Present,
		},
	})
	if err != nil {
		return prevTopicID, err
	}

	if prevTopicID < 0 {
		return topicID, nil
	}

	err = db.UpdateTopicRelation(ctx, &database.UpdateTopicRelationParams{
		ID: prevTopicID,
		NextTopicID: sql.NullInt64{
			Valid: true,
			Int64: topicID,
		},
	})
	if err != nil {
		return topicID, err
	}

	return topicID, nil
}
