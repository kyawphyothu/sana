package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kyawphyothu/sana/config"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/program"
	"github.com/mattn/go-isatty"
)

func main() {
	// Force true color when attached to a TTY (e.g. Docker + iTerm2) so hex colors
	if isatty.IsTerminal(os.Stdout.Fd()) && os.Getenv("COLORTERM") == "" {
		os.Setenv("COLORTERM", "truecolor")
	}
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
	db, err := database.NewDB(config)
	if err != nil {
		fmt.Println("Error creating database:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Println("Error running migrations:", err)
		os.Exit(1)
	}

	m := program.InitialModel(db)
	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
