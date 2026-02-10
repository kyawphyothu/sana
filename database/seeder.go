package database

import (
	"database/sql"
	"time"

	"github.com/kyawphyothu/sana/types"
)

func Seed(db *sql.DB) error {
	expenses := []types.Expense{
		{
			Date:        time.Now().UTC(),
			Amount:      100,
			Description: "Groceries",
			Type:        types.ExpenseTypeFood,
		},
		{
			Date:        time.Now().UTC(),
			Amount:      200,
			Description: "Rent",
			Type:        types.ExpenseTypeBills,
		},
		{
			Date:        time.Now().UTC(),
			Amount:      300,
			Description: "Utilities",
			Type:        types.ExpenseTypeBills,
		},
		{
			Date:        time.Now().UTC(),
			Amount:      400,
			Description: "Entertainment",
			Type:        types.ExpenseTypeOther,
		},
		{
			Date:        time.Now().UTC(),
			Amount:      500,
			Description: "Other",
			Type:        types.ExpenseTypeOther,
		},
		{
			Date:        time.Now().UTC(),
			Amount:      600,
			Description: "Other",
			Type:        types.ExpenseTypeOther,
		},
	}

	for _, e := range expenses {
		_, err := db.Exec(
			`INSERT INTO expenses (date, amount, description, expense_type) VALUES (?, ?, ?, ?)`,
			e.Date, e.Amount, e.Description, string(e.Type),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
