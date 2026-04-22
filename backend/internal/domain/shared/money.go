package shared

// Money 描述金额。MVP 阶段使用 float64；未来若接入真实支付，需要改为 decimal 或整数最小单位。
type Money struct {
	Amount   float64 `bson:"amount"   json:"amount"`
	Currency string  `bson:"currency" json:"currency"`
}
