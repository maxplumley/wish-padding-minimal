package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

var (
	blue = gloss.Color("#6ea3ff")
	pink = gloss.Color("#ff83fc")
)

type style struct {
	background gloss.Style
	base       gloss.Style
}

type model struct {
	width  int
	height int
	style  style
	term   ssh.Pty
}

func initModel(theme theme, renderer *gloss.Renderer, term ssh.Pty) model {
	background := renderer.NewStyle().
		MarginBackground(blue).
		Background(blue)

	base := renderer.NewStyle().
		Inherit(background).
		Foreground(pink)

	return model{
		width:  0,
		height: 0,
		style: style{
			background: background,
			base:       base,
		},
		term: term,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}
	return m, nil
}

func (m model) View() string {
	document := strings.Builder{}
	document.WriteString(gloss.PlaceHorizontal(20, gloss.Center, m.style.base.Render("hello"), gloss.WithWhitespaceBackground(blue), gloss.WithWhitespaceChars("/")))
	return m.style.background.Render(gloss.Place(m.width, m.height, gloss.Center, gloss.Top, document.String(), gloss.WithWhitespaceBackground(blue), gloss.WithWhitespaceChars(".")))
}

type theme int

const (
	Dark theme = iota
	Light
)

func teaHandler(session ssh.Session) (tea.Model, []tea.ProgramOption) {
	// should never fail as we are using the activeTerm middleware
	pty, _, _ := session.Pty()

	renderer := bubbletea.MakeRenderer(session)
	// set the TrueColor profile so that we get some pretty colors
	renderer.SetColorProfile(termenv.TrueColor)

	theme := Light
	if renderer.HasDarkBackground() {
		theme = Dark
	}

	return initModel(theme, renderer, pty), []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithInput(pty.Slave),
		tea.WithOutput(pty.Slave)}
}

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
