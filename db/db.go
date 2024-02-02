package db

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"graphql-go/graph/model"
	"os"
	"strings"
)

func EnsureMigrated(db *pg.DB) {
	// Create tables and apply foreign key constraints
	err := createSchema(db)
	if err != nil {
		panic(err)
	}
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*User)(nil),      // Ensure User is created first, assuming it doesn't depend on other tables
		(*BurgerDay)(nil), // BurgerDay likely depends on User
		(*Order)(nil),     // Order depends on both User and BurgerDay
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false, // Change to true if you want to create a temporary table
			IfNotExists:   true,  // Only create the table if it does not already exist
			FKConstraints: true,  // Ensure foreign key constraints are created
		})
		if err != nil {
			// Check if error is because the table already exists or other errors
			if !strings.Contains(err.Error(), "already exists") {
				return err // Return the error if it's not about the table already existing
			}
		}
	}

	return nil
}

func Connect() *pg.DB {
	// Check if the DB_URL environment variable is set.
	dbURL := os.Getenv("DB_URL")
	var db *pg.DB

	if dbURL != "" {
		// Parse the DATABASE_URL environment variable.
		opt, err := pg.ParseURL(dbURL)
		if err != nil {
			panic(err) // Handle error appropriately in real applications
		}
		db = pg.Connect(opt)
	} else {
		// Use hardcoded values for local development.
		db = pg.Connect(&pg.Options{
			Addr:     "localhost:5432",
			User:     "postgres",
			Password: "changeme",
			Database: "gogql",
		})
	}

	EnsureMigrated(db)

	return db
}

func UserToModel(user *User) *model.User {
	return &model.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func BurgerDayToModel(burgerDay *BurgerDay) *model.BurgerDay {
	return &model.BurgerDay{
		ID:       burgerDay.ID,
		Date:     burgerDay.Date,
		AuthorId: burgerDay.AuthorId,
	}
}
func StringsToSpecialOrders(strings []string) ([]model.SpecialOrders, error) {
	specialOrders := make([]model.SpecialOrders, len(strings))
	for i, s := range strings {
		found := false
		for _, validOrder := range model.AllSpecialOrders {
			if string(validOrder) == s {
				specialOrders[i] = validOrder
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid special order: %s", s)
		}
	}
	return specialOrders, nil
}

func SpecialOrdersToStrings(specialOrders []model.SpecialOrders) []string {
	stringsToFormat := make([]string, len(specialOrders))
	for i, order := range specialOrders {
		stringsToFormat[i] = string(order)
	}
	return stringsToFormat
}

func OrderToModel(order *Order) *model.Order {
	special, _ := StringsToSpecialOrders(order.SpecialRequest)
	return &model.Order{
		ID:             order.ID,
		BurgerDayId:    order.BurgerDayId,
		UserId:         order.UserId,
		SpecialRequest: special,
	}
}

func OrdersToModels(orders []*Order) []*model.Order {
	models := make([]*model.Order, len(orders))
	for i, order := range orders {
		models[i] = OrderToModel(order)
	}
	return models
}

type BurgerDay struct {
	ID       string   `pg:"id,pk" json:"id"`
	AuthorId string   `pg:"author_id" json:"authorId"`
	Author   *User    `pg:"rel:has-one" json:"author"`
	Date     string   `pg:"type:text,unique" json:"date"`
	Orders   []*Order `pg:"rel:has-many,join_fk:burger_day_id" json:"orders"`
}

type User struct {
	ID         string       `pg:"id,pk" json:"id"`
	Name       string       `json:"name"`
	Email      string       `json:"email"`
	BurgerDays []*BurgerDay `pg:"rel:has-many,join_fk:author_id" json:"burgerDays"`
	Orders     []*Order     `pg:"rel:has-many" json:"orders"`
}

type Order struct {
	ID             string     `pg:"id,pk" json:"id"`
	BurgerDayId    string     `pg:"burger_day_id" json:"burgerDayId"`
	BurgerDay      *BurgerDay `pg:"rel:has-one" json:"BurgerDay"`
	UserId         string     `pg:"user_id" json:"userId"`
	User           *User      `pg:"rel:has-one" json:"user"`
	SpecialRequest []string   `pg:"type:text[],array" json:"specialRequest"`
}
