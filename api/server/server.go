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

type Server struct {
	*fiber.App
	db *database.DB
}

func New(cfg config.ServerConfig, db *database.DB) *Server {

	s := &Server{
		db: db,
		App: fiber.New(fiber.Config{
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
	s.App.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	s.App.Use(limiter.New(limiter.Config{
		Max:                    cfg.Limiter.MaxRequests,
		Expiration:             cfg.Limiter.Expiration,
		SkipFailedRequests:     true,
		SkipSuccessfulRequests: false,
		Storage:                memory.New(),
	}))
	s.App.Use(cache.New(cache.Config{
		Expiration:   cfg.Cache.Expiration,
		CacheControl: true,
		Storage:      memory.New(),
	}))
	s.App.Use(fiberlogger.SetLogger(fiberlogger.Config{
		Logger:  logger.Log,
		Context: log.NewContext(nil).Str("module", "site").Value(),
	}))

	s.App.Static(cfg.StaticPrefix, cfg.StaticFolder, fiber.Static{
		Compress:      true,
		Index:         cfg.TemplatesPrefix + "/index.html",
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
	return s
}

func (s *Server) Serve(listener net.Listener) error {
	return s.App.Listener(listener)
}
