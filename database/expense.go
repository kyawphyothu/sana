package database

import (
	"database/sql"

	"github.com/kyawphyothu/sana/types"
)

// ListExpenses returns all expenses from the database, ordered by date descending.
func ListExpenses(db *sql.DB) ([]types.Expense, error) {
	rows, err := db.Query(`
		SELECT id, date, amount, description, expense_type, created_at, updated_at
		FROM expenses
		ORDER BY date DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []types.Expense
	for rows.Next() {
		var e types.Expense
		var typ string
		if err := rows.Scan(&e.ID, &e.Date, &e.Amount, &e.Description, &typ, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.Type = types.ExpenseType(typ)
		list = append(list, e)
	}
	return list, rows.Err()
}
