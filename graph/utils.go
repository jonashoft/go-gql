package graph

import (
	"errors"
	"graphql-go/persistence"
	"time"
	"gorm.io/gorm"
)

// FindFirst returns the first element in the slice that matches the predicate function.
// If no element matches, it returns the zero value for the type and false.
func FindFirst[T any](slice []T, match func(T) bool) (T, bool) {
	for _, item := range slice {
		if match(item) {
			return item, true
		}
	}
	var zero T // Default zero value of type T
	return zero, false
}

// LastElement returns the last element of a slice of any type.
// It returns the zero value and false if the slice is empty.
func LastElement[T any](s []T) (T, bool) {
	if len(s) == 0 {
		var zero T // Default zero value of type T
		return zero, false
	}
	return s[len(s)-1], true
}

func current_burger_day(r *gorm.DB) (*persistence.BurgerDay, error) {
	currentDateString := time.Now().Format("2006-01-02") // Format the date as a string in the format "YYYY-MM-DD"
	burgerDay := &persistence.BurgerDay{}
	res := r.Where("date = ?", currentDateString).First(burgerDay)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no record is found for today
		}
		return nil, res.Error
	}

	return burgerDay, nil
}
