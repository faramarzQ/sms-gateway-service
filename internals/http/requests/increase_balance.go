package requests

type IncreaseBalanceRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}
