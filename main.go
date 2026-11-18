package main

import (
	"financialExchange/enum"
	storage "financialExchange/store"
	"fmt"
	"log/slog"
	"os"
	"time"
	"uuid"
)

func main() {
	options := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, &options)
	logger := slog.New(handler)

	orderBook := storage.NewIndexedStorage[uuid.UUID, enum.Order](100)
	//orderBook := storage.NewStorage[uuid.UUID, enum.Order](10)

	uuidOrder1 := uuid.NewV7()
	uuidOrder2 := uuid.NewV7()
	uuidOrder3 := uuid.NewV7()

	orderBuy := enum.Order{
		ID:        uuidOrder1,
		TraderID:  uuid.NewV4(),
		Side:      enum.SideBuy,
		Type:      enum.OrderTypeLimit,
		Symbo:     "BTC",
		Price:     100.00,
		Quantity:  4.0,
		Filled:    0.00,
		Status:    enum.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	orderSell := enum.Order{
		ID:        uuidOrder2,
		TraderID:  uuid.NewV4(),
		Side:      enum.SideSell,
		Type:      enum.OrderTypeLimit,
		Symbo:     "BTC",
		Price:     98.00,
		Quantity:  2.0,
		Filled:    0.00,
		Status:    enum.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	orderSell2 := enum.Order{
		ID:        uuidOrder3,
		TraderID:  uuid.NewV4(),
		Side:      enum.SideSell,
		Type:      enum.OrderTypeLimit,
		Symbo:     "BTC",
		Price:     99.00,
		Quantity:  3.0,
		Filled:    0.00,
		Status:    enum.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	orderBook.Put(uuidOrder1, orderBuy)
	orderBook.Put(uuidOrder2, orderSell)
	orderBook.Put(uuidOrder3, orderSell2)

	// logger.Info("Зарегистрированы новая сделка",
	// 	slog.String("Type", orderBuy.Type.String()),
	// 	slog.Float64("Value", orderBuy.Price),
	// 	slog.Float64("Quantity", orderBuy.Quantity),
	// 	slog.String("Symbo", orderBuy.Symbo),
	// 	slog.String("TraderID", orderBuy.TraderID.String()),
	// )
	// logger.Info("Зарегистрированы новая сделка",
	// 	slog.String("Type", orderSell.Type.String()),
	// 	slog.Float64("Value", orderSell.Price),
	// 	slog.Float64("Quantity", orderSell.Quantity),
	// 	slog.String("Symbo", orderSell.Symbo),
	// 	slog.String("TraderID", orderSell.TraderID.String()),
	// )
	// logger.Info("Зарегистрированы новая сделка",
	// 	slog.String("Type", orderSell2.Type.String()),
	// 	slog.Float64("Value", orderSell2.Price),
	// 	slog.Float64("Quantity", orderSell2.Quantity),
	// 	slog.String("Symbo", orderSell2.Symbo),
	// 	slog.String("TraderID", orderSell2.TraderID.String()),
	// )

	// fmt.Println("\n----\nСделка 1: ", orderBook.GetOrder(uuidOrder1).Side.String(), orderBook.GetOrder(uuidOrder1).Type.String(), orderBook.GetOrder(uuidOrder1).Price, orderBook.GetOrder(uuidOrder1).Symbo, orderBook.GetOrder(uuidOrder1).Quantity)
	// fmt.Println("----\nСделка 2: ", orderBook.GetOrder(uuidOrder2).Side.String(), orderBook.GetOrder(uuidOrder2).Type.String(), orderBook.GetOrder(uuidOrder2).Price, orderBook.GetOrder(uuidOrder1).Symbo, orderBook.GetOrder(uuidOrder2).Quantity)
	// fmt.Println("----\nСделка 3: ", orderBook.GetOrder(uuidOrder3).Side.String(), orderBook.GetOrder(uuidOrder3).Type.String(), orderBook.GetOrder(uuidOrder3).Price, orderBook.GetOrder(uuidOrder3).Symbo, orderBook.GetOrder(uuidOrder3).Quantity)
	// fmt.Println("----\n")

	// orderBook.Delete(uuidOrder2)
	// logger.Info("Зарегистрировано удаление сделки",
	// 	slog.String("Type", orderBuy.Type.String()),
	// 	slog.Float64("Value", orderBuy.Price),
	// 	slog.Float64("Quantity", orderBuy.Quantity),
	// 	slog.String("Symbo", orderBuy.Symbo),
	// 	slog.String("TraderID", orderBuy.TraderID.String()),
	// )

	// success := storage.OrderBuy(orderBook, uuidOrder1, logger)

	// fmt.Println("Покупка полностью исполнена:", success)

	// updatedBuy := orderBook.GetOrder(uuidOrder1)
	// updatedSell := orderBook.GetOrder(uuidOrder2)
	// updatedSell2 := orderBook.GetOrder(uuidOrder3)

	// fmt.Println("Куплено BTC:", updatedBuy.Filled)
	// fmt.Println("Продано BTC:", updatedSell.Filled)
	// fmt.Println("Продано BTC:", updatedSell2.Filled)
	// fmt.Println("Статус покупателя:", updatedBuy.Status.String())
	// fmt.Println("Статус продавца:", updatedSell.Status.String())
	// fmt.Println("Статус продавца 2:", updatedSell2.Status.String())

	updatedOrder, ok := orderBook.Update(uuidOrder1, func(order enum.Order) enum.Order {
		order.Filled += 1.0

		if order.Filled >= order.Quantity {
			order.Status = enum.OrderStatusFilled
		} else {
			order.Status = enum.OrderStatusPartiallyFilled
		}
		return order
	})

	if ok {
		logger.Info("Статус сделаки на покупку изменен",
			slog.String("Type:", updatedOrder.Type.String()),
			slog.String("Status:", updatedOrder.Status.String()),
			slog.Float64("Value:", updatedOrder.Price),
			slog.Float64("Quantity:", updatedOrder.Quantity),
			slog.String("Symbo:", updatedOrder.Symbo),
			slog.String("TraderID:", updatedOrder.TraderID.String()),
		)
	}

	updatedOrder, ok = orderBook.Update(uuidOrder1, func(order enum.Order) enum.Order) {
		order.Filled += orderSell.Quantity

		if order.Filled >= order.Quantity {
			order.Status = enum.OrderStatusFilled
		} else {
			order.Status = enum.OrderStatusPartiallyFilled
		}
		return order
	}

	if ok {
		logger.Info("Статус сделаки на покупку изменен",
			slog.String("Type:", updatedOrder.Type.String()),
			slog.String("Status:", updatedOrder.Status.String()),
			slog.Float64("Value:", updatedOrder.Price),
			slog.Float64("Quantity:", updatedOrder.Quantity),
			slog.String("Symbo:", updatedOrder.Symbo),
			slog.String("TraderID:", updatedOrder.TraderID.String()),
			slog.String("Date:", updatedOrder.UpdatedAt.String()),
		)
	}

	fmt.Scanln()
}
