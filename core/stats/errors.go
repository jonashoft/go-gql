package stats

type ErrBurgerDayClosed struct {
}

func (e ErrBurgerDayClosed) Error() string {
	return "burger day is closed"
}

func (e ErrBurgerDayClosed) ToString() string {
	return "burger day is closed"
}
