package cli

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/expense"
	"github.com/kyawphyothu/sana/types"
)

// Run parses CLI args and runs the appropriate command if a subcommand is given.
// Returns (true, exitCode) if a CLI command was run (caller should exit with exitCode),
// or (false, 0) to run the TUI.
func Run(db *sql.DB, args []string) (handled bool, exitCode int) {
	if len(args) < 2 {
		return false, 0
	}
	sub := strings.TrimSpace(strings.ToLower(args[1]))
	switch sub {
	case "add":
		return runAdd(db, args[2:])
	case "delete", "del":
		return runDelete(db, args[2:])
	case "list", "ls":
		return runList(db, args[2:])
	default:
		return false, 0
	}
}

func runAdd(db *sql.DB, args []string) (handled bool, exitCode int) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	amountF := fs.Float64("amount", 0, "Expense amount (required)")
	descF := fs.String("description", "", "Expense description (required)")
	typeF := fs.String("type", string(types.ExpenseTypeOther), "Category: food, transport, bills, shopping, health, personal_care, entertainment, education, other")
	dateF := fs.String("date", "", "Date as YYYY-MM-DD or 'today' (default: today)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sana add -amount <n> -description <text> [-type <cat>] [-date YYYY-MM-DD]\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return true, 1
	}
	desc := strings.TrimSpace(*descF)
	if desc == "" {
		fmt.Fprintln(os.Stderr, "Error: -description is required")
		fs.Usage()
		return true, 1
	}

	id, err := expense.AddExpense(db, fmt.Sprint(*amountF), desc, *typeF, *dateF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true, 1
	}
	expType, _ := types.ParseExpenseType(*typeF)
	fmt.Printf("Created expense id=%d (%.2f %s - %s)\n", id, *amountF, expType.String(), desc)
	return true, 0
}

func runDelete(db *sql.DB, args []string) (handled bool, exitCode int) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	idF := fs.Int64("id", 0, "Expense ID to delete (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sana delete -id <expense_id>\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return true, 1
	}
	if *idF <= 0 {
		fmt.Fprintln(os.Stderr, "Error: -id must be a positive integer")
		fs.Usage()
		return true, 1
	}
	err := database.DeleteExpense(db, *idF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting expense: %v\n", err)
		return true, 1
	}
	fmt.Printf("Deleted expense id=%d\n", *idF)
	return true, 0
}

func runList(db *sql.DB, args []string) (handled bool, exitCode int) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	monthF := fs.String("month", "", "Month as YYYY-MM (default: current month)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: sana list [-month YYYY-MM]\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return true, 1
	}
	month, err := expense.ParseMonth(*monthF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true, 1
	}

	expenses, err := database.ListExpenses(db, month)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing expenses: %v\n", err)
		return true, 1
	}
	total, err := database.GetTotalExpenses(db, month)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting total: %v\n", err)
		return true, 1
	}

	monthStr := month.Format("2006-01")
	fmt.Printf("Expenses for %s (total: %.2f)\n", monthStr, total)
	if len(expenses) == 0 {
		fmt.Println("(none)")
		return true, 0
	}
	// Align columns: id, date, amount, type, description
	fmt.Printf("%-6s %-19s %10s %-10s %s\n", "ID", "Date", "Amount", "Type", "Description")
	fmt.Println(strings.Repeat("-", 80))
	for _, e := range expenses {
		dateStr := e.Date.Format("2006-01-02 15:04:05")
		fmt.Printf("%-6d %-19s %10.2f %-10s %s\n", e.ID, dateStr, e.Amount, e.Type.String(), e.Description)
	}
	return true, 0
}

// PrintUsage prints a short usage line when the user passes an unknown subcommand or -h.
func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: sana [add|delete|list] [flags]\n")
	fmt.Fprintf(os.Stderr, "  add    -amount <n> -description <text> [-type <cat>] [-date YYYY-MM-DD]\n")
	fmt.Fprintf(os.Stderr, "  delete -id <expense_id>\n")
	fmt.Fprintf(os.Stderr, "  list   [-month YYYY-MM]\n")
	fmt.Fprintf(os.Stderr, "Run with no arguments to start the TUI.\n")
}
