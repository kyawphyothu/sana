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

// GetExpensesSummary returns expenses grouped by category with totals, ordered alphabetically
func GetExpensesSummary(db *sql.DB) ([]types.CategorySummary, error) {
	rows, err := db.Query(`
		SELECT expense_type, SUM(amount) as total
		FROM expenses
		GROUP BY expense_type
		ORDER BY total DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []types.CategorySummary
	for rows.Next() {
		var typStr string
		var total float64
		if err := rows.Scan(&typStr, &total); err != nil {
			return nil, err
		}

		// Convert expense_type string to display name
		expType := types.ExpenseType(typStr)
		summaries = append(summaries, types.CategorySummary{
			Category: expType.String(),
			Total:    total,
		})
	}
	return summaries, rows.Err()
}

// GetTotalExpenses returns the sum of all expenses
func GetTotalExpenses(db *sql.DB) (float64, error) {
	var total float64
	err := db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM expenses`).Scan(&total)
	return total, err
}
