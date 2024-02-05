package stats

import (
	"graphql-go/graph/model"
	"graphql-go/persistence"

	"gorm.io/gorm"
)

type UserBurgerStats struct {
	TotalOrders     int         `json:"totalOrders"`
	TotalBurgerDays int         `json:"totalBurgerDays"`
	User            *model.User `json:"user"`
}

type BurgerStats struct {
	TotalOrders     int                `json:"totalOrders"`
	TotalBurgerDays int                `json:"totalBurgerDays"`
	UserStats       []*UserBurgerStats `json:"userStats"`
}

func BurgerStatsToModel(bStats *BurgerStats) *model.BurgerStats {
	return &model.BurgerStats{
		TotalOrders:     bStats.TotalOrders,
		TotalBurgerDays: bStats.TotalBurgerDays,
		TopConsumers:    UserStatsToConsumerModel(bStats.UserStats),
	}
}

func UserStatsToConsumerModel(userStats []*UserBurgerStats) []*model.Consumer {
	consumers := make([]*model.Consumer, len(userStats))
	for i, userStat := range userStats {
		consumers[i] = &model.Consumer{
			User:            userStat.User,
			TotalOrders:     userStat.TotalOrders,
			TotalBurgerDays: userStat.TotalBurgerDays,
		}
	}
	return consumers
}

func CalculateBurgerStats(db *gorm.DB) (*BurgerStats, error) {
	var totalOrders int64
	var totalBurgerDays int64
	var userStats []*UserBurgerStats

	// Calculate total number of orders
	if err := db.Model(&persistence.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, err
	}

	// Calculate total number of unique burger days
	if err := db.Model(&persistence.BurgerDay{}).Distinct("id").Count(&totalBurgerDays).Error; err != nil {
		return nil, err
	}

	// Calculate stats per user
	// This involves a bit more complex query to aggregate data per user
	type userStatsQueryResult struct {
		UserId          string
		Email           string
		Name            string
		TotalOrders     int
		TotalBurgerDays int
	}

	var queryResults []userStatsQueryResult
	if err := db.Table("orders").
		Select("users.id as user_id, users.name, users.email, count(distinct orders.id) as total_orders, count(distinct burger_day_id) as total_burger_days").
		Joins("join users on users.id = orders.user_id").
		Group("users.id, users.name, users.email").
		Find(&queryResults).Error; err != nil {
		return nil, err
	}

	// Convert query results to UserBurgerStats
	for _, result := range queryResults {
		user := &model.User{
			ID:    result.UserId,
			Name:  result.Name,
			Email: result.Email,
		}

		stat := &UserBurgerStats{
			User:            user,
			TotalOrders:     result.TotalOrders,
			TotalBurgerDays: result.TotalBurgerDays,
		}
		userStats = append(userStats, stat)
	}

	// You might need additional queries to fill in the Name and Email for each user in userStats
	// if not already done in the aggregation query above.

	stats := &BurgerStats{
		TotalOrders:     int(totalOrders),
		TotalBurgerDays: int(totalBurgerDays),
		UserStats:       userStats,
	}

	return stats, nil
}
