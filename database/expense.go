package database

import (
	"database/sql"
	"time"

	"github.com/kyawphyothu/sana/types"
)

// ListExpenses returns all expenses from the database, ordered by date descending.
func ListExpenses(db *sql.DB) ([]types.Expense, error) {
	rows, err := db.Query(`
		SELECT id, date, amount, description, expense_type, created_at, updated_at
		FROM expenses
		WHERE strftime('%Y-%m', date) = strftime('%Y-%m', 'now')
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
		SELECT expense_type, SUM(amount) as total, COUNT(*) as count
		FROM expenses
		WHERE strftime('%Y-%m', date) = strftime('%Y-%m', 'now')
		GROUP BY expense_type
		ORDER BY total DESC, count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []types.CategorySummary
	for rows.Next() {
		var typStr string
		var total float64
		var count int
		if err := rows.Scan(&typStr, &total, &count); err != nil {
			return nil, err
		}

		// Convert expense_type string to display name
		expType := types.ExpenseType(typStr)
		summaries = append(summaries, types.CategorySummary{
			Category: expType.String(),
			Total:    total,
			Count:    count,
		})
	}
	return summaries, rows.Err()
}

// GetMonthlyReport returns expenses grouped by month with totals
func GetMonthlyReport(db *sql.DB) ([]types.MonthlyReport, error) {
	rows, err := db.Query(`
		SELECT strftime('%Y-%m', date) as month, SUM(amount) as total
		FROM expenses
		GROUP BY strftime('%Y-%m', date)
		ORDER BY month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monthlyReport []types.MonthlyReport
	for rows.Next() {
		var monthStr string
		var total float64
		if err := rows.Scan(&monthStr, &total); err != nil {
			return nil, err
		}
		month, err := time.Parse("2006-01", monthStr)
		if err != nil {
			return nil, err
		}
		monthlyReport = append(monthlyReport, types.MonthlyReport{
			Month: month,
			Total: total,
		})
	}
	return monthlyReport, rows.Err()
}

// GetTotalExpenses returns the sum of all expenses
func GetTotalExpenses(db *sql.DB) (float64, error) {
	var total float64
	err := db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE strftime('%Y-%m', date) = strftime('%Y-%m', 'now')`).Scan(&total)
	return total, err
}

// CreateExpense inserts a new expense and returns the new ID.
func CreateExpense(db *sql.DB, date time.Time, amount float64, description string, expenseType types.ExpenseType) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO expenses (date, amount, description, expense_type)
		VALUES (?, ?, ?, ?)
	`, date, amount, description, string(expenseType))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DeleteExpense removes an expense by ID.
func DeleteExpense(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	return err
}
