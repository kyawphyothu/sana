package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kyawphyothu/sana/config"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/types"
)

func main() {
	expenses := []types.Expense{
		{
			Date:        time.Now().AddDate(0, -1, 0),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now().AddDate(0, -1, 0),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now().AddDate(0, -1, 0),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now().AddDate(0, -1, 0),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now(),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now(),
			Amount:      200,
			Description: "Rent",
			Type:        types.ExpenseTypeBills,
		},
		{
			Date:        time.Now(),
			Amount:      300,
			Description: "Utilities",
			Type:        types.ExpenseTypeBills,
		},
		{
			Date:        time.Now(),
			Amount:      400,
			Description: "Entertainment",
			Type:        types.ExpenseTypeOther,
		},
		{
			Date:        time.Now(),
			Amount:      500,
			Description: "Other",
			Type:        types.ExpenseTypeOther,
		},
		{
			Date:        time.Now(),
			Amount:      600,
			Description: "Other",
			Type:        types.ExpenseTypeOther,
		},
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

	for _, e := range expenses {
		_, err := database.CreateExpense(db, e.Date, e.Amount, e.Description, e.Type)
		if err != nil {
			fmt.Println("Error inserting expense:", err)
			os.Exit(1)
		}
	}
	fmt.Println("Database seeded successfully")
	os.Exit(0)
}
