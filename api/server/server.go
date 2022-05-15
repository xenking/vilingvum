package server

import (
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/html"
	"github.com/phuslu/log"
	fiberlogger "github.com/phuslu/log-contrib/fiber"

	"tgbot/config"
	"tgbot/database"
	"tgbot/pkg/logger"
	"tgbot/pkg/storage/memory"
)

var appVersion string

type Server struct {
	*fiber.App
	db *database.DB
}

func New(cfg config.ServerConfig, db *database.DB) *Server {
	appVersion = cfg.Version
	cfg.Limiter.Storage = memory.New()
	cfg.Cache.Storage = memory.New()

	s := &Server{
		db: db,
		App: fiber.New(fiber.Config{
			Prefork:           cfg.Prefork,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			AppName:           "vilingvum",
			Views:             html.New(cfg.ViewsFolder, cfg.ViewsExt),
			StreamRequestBody: true,
		}),
	}

	return s.setupMiddlewares(cfg)
}

func (s *Server) setupMiddlewares(cfg config.ServerConfig) *Server {
	s.App.Use(recover.New())
	s.App.Use(requestid.New())

	s.App.Use(csrf.New())
	s.App.Use(limiter.New(cfg.Limiter))
	s.App.Use(cache.New(cfg.Cache))
	s.App.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	s.App.Use(fiberlogger.SetLogger(fiberlogger.Config{
		Logger:  logger.Log,
		Context: log.NewContext(nil).Str("module", "rest").Value(),
	}))

	s.App.Static(cfg.StaticPrefix, cfg.StaticFolder, fiber.Static{
		Compress:      true,
		Index:         cfg.TemplatesPrefix + "/index.html",
		CacheDuration: 10 * time.Hour,
		MaxAge:        int(time.Hour / time.Second),
	})
	s.App.Static(cfg.FilesPrefix, cfg.FilesFolder, fiber.Static{
		Compress:      true,
		CacheDuration: 10 * time.Hour,
		MaxAge:        int(time.Hour / time.Second),
	})
	// s.App.Use(notFoundHandler(cfg))

	return s
}

//func notFoundHandler(cfg config.ServerConfig) fiber.Handler {
//	return func(ctx *fiber.Ctx) error {
//
//		err := ctx.Status(fiber.StatusNotFound).Render("404", fiber.Map{})
//		if err != nil {
//			return ctx.Status(500).SendString("Internal Server Error")
//		}
//		return nil
//	}
//}

func (s *Server) registerRoutes() *Server {
	s.Get("/version", handleVersion)
	return s
}

func handleVersion(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"version":   appVersion,
		"timestamp": time.Now(),
	})
}

func (s *Server) Serve(listener net.Listener) error {
	return s.App.Listener(listener)
}
