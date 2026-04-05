package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maiconpml/heylisten/internal/audio"
	"github.com/maiconpml/heylisten/internal/config"
	"github.com/maiconpml/heylisten/internal/logger"
	"github.com/maiconpml/heylisten/internal/tui"
	"github.com/maiconpml/heylisten/internal/ytdlp"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

func main() {
	if err := logger.Init(); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	if err := ytdlp.CheckDependencies(); err != nil {
		slog.Error("Dependency error", "err", err)
		os.Exit(1)
	}

	if err := audio.Init(); err != nil {
		slog.Warn("Could not initialize audio system", "err", err)
	} else {
		defer audio.Quit()
	}

	cookiePath, err := config.GetCookiePath()
	if err != nil {
		slog.Error("Error getting config path", "err", err)
		os.Exit(1)
	}

	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		fmt.Printf("No authentication cookie found. Please paste your YouTube Music cookie string:\n> ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			slog.Error("Error reading input", "err", err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)
		if input == "" {
			slog.Error("Cookie cannot be empty")
			os.Exit(1)
		}
		if err := config.SaveCookie(input); err != nil {
			slog.Error("Error saving cookie", "err", err)
			os.Exit(1)
		}
		fmt.Println("Cookie saved successfully!")
	}

	cookieString, err := config.LoadCookie()
	if err != nil {
		slog.Error("Error loading cookie", "err", err)
		os.Exit(1)
	}

	if err := ytdlp.Init(cookiePath); err != nil {
		slog.Error("Error on ytdlp initializing", "err", err)
		os.Exit(1)
	}
	defer ytdlp.CleanCache()

	client := goytmusic.NewClient(&http.Client{}).WithAuthCookie(cookieString)

	// Configura o programa com log em arquivo
	p := tea.NewProgram(
		tui.NewModel(client),
		tea.WithAltScreen(),
	)

	if f, err := tea.LogToFile("app.log", "tea"); err == nil {
		defer f.Close()
	}

	if _, err := p.Run(); err != nil {
		slog.Error("Execution error", "err", err)
		os.Exit(1)
	}
}
