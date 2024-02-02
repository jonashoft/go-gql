package model

type Order struct {
	ID             string          `json:"id"`
	BurgerDay      *BurgerDay      `json:"BurgerDay"`
	BurgerDayId    string          `json:"burgerDayId"`
	User           *User           `json:"user"`
	UserId         string          `json:"userId"`
	SpecialRequest []SpecialOrders `json:"specialRequest"`
}
type BurgerDay struct {
	ID       string   `json:"id"`
	Author   *User    `json:"author"`
	AuthorId string   `json:"authorId"`
	Date     string   `json:"date"`
	Orders   []*Order `json:"orders"`
}
