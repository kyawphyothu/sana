package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kyawphyothu/sana/cli"
	"github.com/kyawphyothu/sana/config"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/program"
	"github.com/mattn/go-isatty"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}
	db, err := database.NewDB(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating database:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Fprintln(os.Stderr, "Error running migrations:", err)
		os.Exit(1)
	}

	// CLI: if a subcommand was given, run it and exit
	if handled, code := cli.Run(db, os.Args); handled {
		os.Exit(code)
	}

	// TUI
	if isatty.IsTerminal(os.Stdout.Fd()) && os.Getenv("COLORTERM") == "" {
		os.Setenv("COLORTERM", "truecolor")
	}
	m := program.InitialModel(db)
	p := tea.NewProgram(m)
	_, err = p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error running program:", err)
		os.Exit(1)
	}
}
