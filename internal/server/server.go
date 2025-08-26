package server

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"geodb-example/internal/controllers/api"
	"geodb-example/internal/database"
	"geodb-example/internal/embedfs"
	"geodb-example/web"
)

type Server struct {
	cfg     *viper.Viper
	db      database.DB
	logger  *logrus.Logger
	httpSrv *http.Server
	apiCtrl *api.Controller
}

// NewServer
func NewServer(cfg *viper.Viper, db database.DB, logger *logrus.Logger) *Server {
	srv := &Server{}
	srv.cfg = cfg
	srv.db = db
	srv.logger = logger

	// Создаём HTTP-сервер
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logger.Out

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Заголовок X-Forwarded-For можно подделать и поэтому необходимо указывать, каким прокси-серверам ты доверяешь
	router.SetTrustedProxies(srv.cfg.GetStringSlice("server.trusted_proxies"))
	router.ForwardedByClientIP = srv.cfg.GetBool("server.forwarded_for")

	// Некоторые настройки, связанные с безопасностью
	router.Use(secure.Secure(secure.Options{
		FrameDeny:          true, // Запрещает показывать сайт во фрейме
		ContentTypeNosniff: true, //
		BrowserXssFilter:   true, //
	}))

	// Шаблоны
	templ := template.Must(template.New("").ParseFS(web.FS, "templates/*.gohtml"))
	router.SetHTMLTemplate(templ)

	router.GET("/", func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		geo, err := srv.db.Lookup(ip)
		if err != nil {
			geo = &database.GeoRecord{}
		}
		ctx.HTML(http.StatusOK, "index.gohtml", gin.H{"ip": ip, "geo": geo, "version": srv.db.Version()})
	})

	// Статические файлы
	router.Use(static.Serve("/", embedfs.EmbedFolder(web.FS, "static")))

	// API
	srv.apiCtrl = api.NewController(srv)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/lookup", srv.apiCtrl.Lookup())
	}

	srv.httpSrv = &http.Server{
		Handler:      router,
		WriteTimeout: 5 * time.Minute, // Таймаут ответа от сервера
	}

	return srv
}

// Run запускает HTTP-сервер
func (s *Server) Run(addr string) {
	s.httpSrv.Addr = addr
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatalf("Error: %s\n", err)
		}
	}()
}

// Shutdown останавливает HTTP-сервер
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpSrv.Shutdown(ctx)
}

// Config возвращает указатель на конфиг
func (s *Server) Config() *viper.Viper {
	return s.cfg
}

// DB возвращает указатель на БД
func (s *Server) DB() database.DB {
	return s.db
}

// Logger возвращает указатель на логгер
func (s *Server) Logger() *logrus.Logger {
	return s.logger
}
