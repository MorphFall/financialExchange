package modifiers
import (
	//"fmt"
	"financialExchange/enum"
)


func ModifierForOrderBuy(traderQuantity float64) func(enum.Order) enum.Order {
	return func(order enum.Order) enum.Order {
		order.Filled += traderQuantity

		if order.Filled >= order.Quantity {
			order.Status = enum.OrderStatusFilled
		} else {
			order.Status = enum.OrderStatusPartiallyFilled
		}
		return order
	}
}