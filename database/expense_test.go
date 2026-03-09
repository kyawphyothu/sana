package database

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/kyawphyothu/sana/types"
)

// testDB returns an in-memory SQLite DB (":memory:") with migrations applied.
// Using a real DB catches SQL/schema bugs; no mocks or temp files needed.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := Migrate(db); err != nil {
		db.Close()
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCreateExpense(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	date := time.Date(2025, 3, 15, 10, 30, 0, 0, time.Local)
	id, err := CreateExpense(db, date, 99.50, "lunch", types.ExpenseTypeFood)
	if err != nil {
		t.Fatalf("CreateExpense: %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateExpense id = %d, want positive", id)
	}

	// Second insert gets next ID
	id2, err := CreateExpense(db, date, 1, "other", types.ExpenseTypeOther)
	if err != nil {
		t.Fatalf("CreateExpense second: %v", err)
	}
	if id2 != id+1 {
		t.Errorf("CreateExpense second id = %d, want %d", id2, id+1)
	}
}

func TestListExpenses(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	mar := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	feb := time.Date(2025, 2, 10, 0, 0, 0, 0, time.Local)

	// Empty month
	list, err := ListExpenses(db, mar)
	if err != nil {
		t.Fatalf("ListExpenses: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("ListExpenses empty month: got %d, want 0", len(list))
	}

	// Insert in March
	_, err = CreateExpense(db, mar, 10, "m1", types.ExpenseTypeFood)
	if err != nil {
		t.Fatalf("CreateExpense: %v", err)
	}
	_, err = CreateExpense(db, time.Date(2025, 3, 20, 0, 0, 0, 0, time.Local), 20, "m2", types.ExpenseTypeBills)
	if err != nil {
		t.Fatalf("CreateExpense: %v", err)
	}
	// Insert in February
	_, err = CreateExpense(db, feb, 5, "feb1", types.ExpenseTypeOther)
	if err != nil {
		t.Fatalf("CreateExpense: %v", err)
	}

	list, err = ListExpenses(db, mar)
	if err != nil {
		t.Fatalf("ListExpenses March: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ListExpenses March: got %d, want 2", len(list))
	}
	// Order: date DESC, id DESC — so m2 (Mar 20) then m1 (Mar 15)
	if list[0].Description != "m2" || list[0].Amount != 20 {
		t.Errorf("first row: got %q %.2f, want m2 20", list[0].Description, list[0].Amount)
	}
	if list[1].Description != "m1" || list[1].Amount != 10 {
		t.Errorf("second row: got %q %.2f, want m1 10", list[1].Description, list[1].Amount)
	}

	list, err = ListExpenses(db, feb)
	if err != nil {
		t.Fatalf("ListExpenses Feb: %v", err)
	}
	if len(list) != 1 || list[0].Description != "feb1" {
		t.Errorf("ListExpenses Feb: got %d rows, want 1 (feb1)", len(list))
	}
}

func TestGetTotalExpenses(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local)
	total, err := GetTotalExpenses(db, mar)
	if err != nil {
		t.Fatalf("GetTotalExpenses empty: %v", err)
	}
	if total != 0 {
		t.Errorf("GetTotalExpenses empty: got %.2f, want 0", total)
	}

	_, _ = CreateExpense(db, time.Date(2025, 3, 10, 0, 0, 0, 0, time.Local), 100, "a", types.ExpenseTypeFood)
	_, _ = CreateExpense(db, time.Date(2025, 3, 11, 0, 0, 0, 0, time.Local), 50, "b", types.ExpenseTypeFood)
	total, err = GetTotalExpenses(db, mar)
	if err != nil {
		t.Fatalf("GetTotalExpenses: %v", err)
	}
	if total != 150 {
		t.Errorf("GetTotalExpenses: got %.2f, want 150", total)
	}
}

func TestDeleteExpense(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	mar := time.Date(2025, 3, 15, 0, 0, 0, 0, time.Local)
	id, _ := CreateExpense(db, mar, 10, "to delete", types.ExpenseTypeFood)
	list, _ := ListExpenses(db, mar)
	if len(list) != 1 {
		t.Fatalf("before delete: want 1 row, got %d", len(list))
	}

	err := DeleteExpense(db, id)
	if err != nil {
		t.Fatalf("DeleteExpense: %v", err)
	}
	list, _ = ListExpenses(db, mar)
	if len(list) != 0 {
		t.Errorf("after delete: want 0 rows, got %d", len(list))
	}
}

func TestGetExpensesSummary(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	mar := time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local)
	_, _ = CreateExpense(db, time.Date(2025, 3, 10, 0, 0, 0, 0, time.Local), 30, "a", types.ExpenseTypeFood)
	_, _ = CreateExpense(db, time.Date(2025, 3, 11, 0, 0, 0, 0, time.Local), 20, "b", types.ExpenseTypeFood)
	_, _ = CreateExpense(db, time.Date(2025, 3, 12, 0, 0, 0, 0, time.Local), 15, "c", types.ExpenseTypeBills)

	summary, err := GetExpensesSummary(db, mar)
	if err != nil {
		t.Fatalf("GetExpensesSummary: %v", err)
	}
	if len(summary) != 2 {
		t.Fatalf("GetExpensesSummary: got %d groups, want 2", len(summary))
	}
	// Order: total DESC — Food 50, Bills 15
	if summary[0].Category != "Food" || summary[0].Total != 50 || summary[0].Count != 2 {
		t.Errorf("first summary: got %s %.2f count %d, want Food 50 2", summary[0].Category, summary[0].Total, summary[0].Count)
	}
	if summary[1].Category != "Bills" || summary[1].Total != 15 || summary[1].Count != 1 {
		t.Errorf("second summary: got %s %.2f count %d, want Bills 15 1", summary[1].Category, summary[1].Total, summary[1].Count)
	}
}

func TestGetMonthlyReport(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	_, _ = CreateExpense(db, time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local), 100, "feb", types.ExpenseTypeFood)
	_, _ = CreateExpense(db, time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local), 50, "mar1", types.ExpenseTypeFood)
	_, _ = CreateExpense(db, time.Date(2025, 3, 2, 0, 0, 0, 0, time.Local), 25, "mar2", types.ExpenseTypeFood)

	report, err := GetMonthlyReport(db)
	if err != nil {
		t.Fatalf("GetMonthlyReport: %v", err)
	}
	if len(report) != 2 {
		t.Fatalf("GetMonthlyReport: got %d months, want 2", len(report))
	}
	// Order: month DESC — 2025-03 then 2025-02
	if report[0].Month.Year() != 2025 || report[0].Month.Month() != 3 || report[0].Total != 75 {
		t.Errorf("first month: got %v %.2f, want 2025-03 75", report[0].Month.Format("2006-01"), report[0].Total)
	}
	if report[1].Month.Year() != 2025 || report[1].Month.Month() != 2 || report[1].Total != 100 {
		t.Errorf("second month: got %v %.2f, want 2025-02 100", report[1].Month.Format("2006-01"), report[1].Total)
	}
}
