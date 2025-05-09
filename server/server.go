package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/titpetric/task-ui/server/internal"
)

const (
	indexTemplate = "templates/index.tpl"
	taskTemplate  = "templates/task.tpl"
)

type Server struct {
	files     *embed.FS
	config    *flags
	semaphore *internal.Semaphore
	template  internal.TemplateRendererFunc
}

func New(files *embed.FS) (*Server, error) {
	flags := new(flags)
	flags.Bind()

	if err := flags.Validate(); err != nil {
		return nil, err
	}

	if flags.PrintBuildInfo {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			return nil, errors.New("error when reading build info")
		}
		fmt.Println(buildInfo)
		os.Exit(0)
	}

	return &Server{
		config:    flags,
		semaphore: internal.NewSemaphore(),
		files:     files,
		template:  internal.NewTemplateRenderer(files),
	}, nil
}

func (svc *Server) Start(ctx context.Context) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// app routes
	r.Get("/", svc.Index)
	r.Get("/{id}", svc.Task)

	r.Handle("/ws/{id}", svc.Launch())

	if svc.config.EnableHistory {
		r.Handle("/ws/history/{id}", svc.Play(false))
		r.Handle("/ws/history/{id}/replay", svc.Play(true))
	}

	if svc.config.EnableHistory {
		r.HandleFunc("/api/history/{id}", svc.History)
		r.HandleFunc("/api/history", svc.History)
	}

	// Only serve /static from filesystem directly
	static := fs.FS(svc.files)
	files, err := fs.Sub(static, "public_html/static")
	if err != nil {
		return err
	}
	staticFS := http.FileServer(http.FS(files))
	r.Handle("/static/*", http.StripPrefix("/static/", staticFS))

	_, port, err := net.SplitHostPort(svc.config.ListenAddress)
	if err != nil {
		return fmt.Errorf("invalid input for listenAddress: %w", err)
	}

	fmt.Printf("Listening on http://localhost:%s\n", port)

	return http.ListenAndServe(svc.config.ListenAddress, r)
}
