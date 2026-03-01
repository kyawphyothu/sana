package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/kyawphyothu/sana/config"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/types"
)

type template struct {
	Description string
	Type        types.ExpenseType
	MinAmount   float64
	MaxAmount   float64
}

var templates = []template{
	// Food
	{"Groceries", types.ExpenseTypeFood, 5, 80},
	{"Lunch", types.ExpenseTypeFood, 5, 25},
	{"Dinner", types.ExpenseTypeFood, 10, 50},
	{"Coffee", types.ExpenseTypeFood, 3, 8},
	{"Snacks", types.ExpenseTypeFood, 2, 15},
	{"Breakfast", types.ExpenseTypeFood, 4, 18},
	{"Takeout", types.ExpenseTypeFood, 8, 35},
	{"Bubble tea", types.ExpenseTypeFood, 4, 8},

	// Transport
	{"Bus fare", types.ExpenseTypeTransport, 1, 5},
	{"Taxi", types.ExpenseTypeTransport, 5, 30},
	{"Grab ride", types.ExpenseTypeTransport, 3, 20},
	{"Fuel", types.ExpenseTypeTransport, 20, 60},
	{"Parking", types.ExpenseTypeTransport, 2, 10},

	// Bills
	{"Electricity", types.ExpenseTypeBills, 30, 120},
	{"Water bill", types.ExpenseTypeBills, 10, 40},
	{"Internet", types.ExpenseTypeBills, 20, 50},
	{"Phone bill", types.ExpenseTypeBills, 10, 40},
	{"Rent", types.ExpenseTypeBills, 300, 800},

	// Shopping
	{"Clothes", types.ExpenseTypeShopping, 15, 100},
	{"Shoes", types.ExpenseTypeShopping, 20, 120},
	{"Electronics", types.ExpenseTypeShopping, 10, 200},
	{"Home supplies", types.ExpenseTypeShopping, 5, 50},
	{"Books", types.ExpenseTypeShopping, 5, 30},

	// Health
	{"Medicine", types.ExpenseTypeHealth, 5, 50},
	{"Doctor visit", types.ExpenseTypeHealth, 20, 100},
	{"Gym membership", types.ExpenseTypeHealth, 20, 50},
	{"Vitamins", types.ExpenseTypeHealth, 10, 30},

	// Other
	{"Entertainment", types.ExpenseTypeOther, 5, 40},
	{"Subscription", types.ExpenseTypeOther, 5, 20},
	{"Gift", types.ExpenseTypeOther, 10, 80},
	{"Haircut", types.ExpenseTypeOther, 5, 25},
	{"Laundry", types.ExpenseTypeOther, 3, 15},
}

func randomDate(r *rand.Rand, monthsBack int) time.Time {
	now := time.Now()
	month := now.AddDate(0, -r.Intn(monthsBack+1), 0)
	day := r.Intn(28) + 1
	hour := r.Intn(14) + 7 // 7am - 9pm
	minute := r.Intn(60)
	second := r.Intn(60)
	return time.Date(month.Year(), month.Month(), day, hour, minute, second, 0, time.Local)
}

func randomAmount(r *rand.Rand, min, max float64) float64 {
	val := min + r.Float64()*(max-min)
	return float64(int(val*100)) / 100
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	const (
		monthsBack   = 6
		totalExpenses = 200
	)

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		fmt.Println("Error creating database:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Println("Error running migrations:", err)
		os.Exit(1)
	}

	for i := 0; i < totalExpenses; i++ {
		t := templates[r.Intn(len(templates))]
		date := randomDate(r, monthsBack)
		amount := randomAmount(r, t.MinAmount, t.MaxAmount)

		_, err := database.CreateExpense(db, date, amount, t.Description, t.Type)
		if err != nil {
			fmt.Println("Error inserting expense:", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Database seeded with %d expenses across %d months\n", totalExpenses, monthsBack+1)
}
