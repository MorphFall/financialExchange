package storage

import (
	//"fmt"
	"financialExchange/enum"
	"log/slog"
	"sync"
	"time"
	"uuid"
)

// sync.Mutex полностью блокирует доступ и для чтения и для записи.
// RWMutex позволяет потокам читать данные одновременно, но запись осуществляется только одним потоком

// генерик и потокобезопасное хранилище.
type Store[K comparable, V any] struct {
	mutex sync.RWMutex // Защищает map от одновременного доступа
	items map[K]V      // Хранилище элементов
}

func NewStore[K comparable, V any](capacity int) *Store[K, V] {
	return &Store[K, V]{
		items: make(map[K]V, capacity),
	}
}

// добавляет элемент или перезаписывает существующий по ключу
func (store *Store[K, V]) Put(key K, value V) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.items[key] = value
}

// возвращает значение по ключу
func (store *Store[K, V]) Get(key K) (V, bool) {
	store.mutex.RLock() // RLock разрешает нескольким горутинам читать одновременно
	defer store.mutex.RUnlock()

	value, ok := store.items[key]
	return value, ok
}

// Удаление элемента по ключу
func (store *Store[K, V]) Delete(key K, value V) bool {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.items[key]; !ok {
		return false
	}

	delete(store.items, key)
	return true
}

// изменяет существующее значение с помощью переданной функции.
func (store *Store[K, V]) Update(key K, modif func(V) V) (V, bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	value, ok := store.items[key]
	if !ok {
		var zero V // Нулевое значение типа V, например 0 для int
		return zero, false
	}

	updatedValue := modif(value) // modif принимает старое значение и возвращает изменённое

	store.items[key] = updatedValue // Сохраняем изменение в map
	return updatedValue, true
}

// изменяет существующее значение с помощью переданной функции. ---Для IndexedStorage---
func (indexStore *IndexedStorage[K, V]) Update(key K, tradeQuantity float64, modif func(V) V) (V, bool) {
	indexStore.mutex.Lock()
	defer indexStore.mutex.Unlock()

	value, ok := indexStore.data[key]
	if !ok {
		var zero V // Нулевое значение типа V, например 0 для int
		return zero, false
	}

	updatedValue := modif(value) // modif принимает старое значение и возвращает изменённое

	indexStore.data[key] = updatedValue // Сохраняем изменение в map
	return updatedValue, true
}

// хранилище с сохранением порядка элементов.
type IndexedStorage[K comparable, V any] struct {
	mutex sync.RWMutex
	order []K
	byKey map[K]int // Позволяет за O(1) найти индекс ключа в order
	data  map[K]V   // Значения элементов по ключу
}

func (indexStore *IndexedStorage[K, V]) GetOrder(key K) V {
	indexStore.mutex.RLock()
	defer indexStore.mutex.RUnlock()

	return indexStore.data[key]
}

func (indexStore *IndexedStorage[K, V]) Ordered() []V { //Возвращает тип данных V - генерик
	indexStore.mutex.RLock() //Блокировка на чтение, чтобы другие потоки могли читать данные одновременно
	defer indexStore.mutex.RUnlock()

	out := make([]V, len(indexStore.order))    //Срез типа V с длиной indStor.order
	for index, key := range indexStore.order { //Заполняем out по индексу и ключу
		out[index] = indexStore.data[key] //Заполняем out в том же порядке, используя индекс элемента
	}
	return out //Возвращаем срез out
}

func NewIndexedStorage[K comparable, V any](capacity int) *IndexedStorage[K, V] {
	return &IndexedStorage[K, V]{
		order: make([]K, 0, capacity),
		byKey: make(map[K]int, capacity),
		data:  make(map[K]V, capacity),
	}
}

// Добавляет элемент в хранилище
func (storage *IndexedStorage[K, V]) Put(key K, newValue V) {
	storage.mutex.Lock() //Лок и анлок мьютекса
	defer storage.mutex.Unlock()

	if _, ok := storage.byKey[key]; ok { //Проверяет, есть ли ключ в хранилище. Если есть, то обновляет значение
		storage.data[key] = newValue
		return
	}

	storage.byKey[key] = len(storage.order)    //Если, то устанавливает индекс ключа
	storage.order = append(storage.order, key) //Добавляет ключ в срез order
	storage.data[key] = newValue               //Добавляет значение в data по ключу
}

func (storage *IndexedStorage[K, V]) Delete(key K) bool {
	storage.mutex.Lock() //Лок и анлок мьютекса
	defer storage.mutex.Unlock()
	indexInSlice, ok := storage.byKey[key]

	if !ok {
		return false
	}
	lastIndex := len(storage.order) - 1

	if indexInSlice != lastIndex {
		lastIndex := storage.order[lastIndex]   //Получает последний элемент среза order
		storage.order[indexInSlice] = lastIndex //Заменяет удаляемый элемент на последний элемент среза order
		storage.byKey[lastIndex] = indexInSlice //Обновляет индекс последнего элемента в мапе byKey
	}

	storage.order = storage.order[:lastIndex] //Удаляет последний элемент из среза order
	delete(storage.byKey, key)                //Удаляет ключ из мапы byKey
	delete(storage.data, key)                 //Удаляет ключ из мапы data

	return true
}

func OrderBuy(orderBook *IndexedStorage[uuid.UUID, enum.Order], key uuid.UUID, logger *slog.Logger) bool {
	orderBook.mutex.Lock()
	defer orderBook.mutex.Unlock()

	// Поиск заявки покупателя
	buy, ok := orderBook.data[key]
	if !ok || buy.Side != enum.SideBuy || !buy.IsActive() {
		logger.Warn("Заявка на покупку не найдена или неактивна")
		return false
	}

	for _, sellID := range orderBook.order {
		if buy.Remaining() <= 0 {
			break
		}

		sell := orderBook.data[sellID]

		// Проверяем, подходит ли заявка продавца
		if sell.Side != enum.SideSell || !sell.IsActive() || sell.Symbo != buy.Symbo || sell.Price > buy.Price || sell.TraderID == buy.TraderID {
			continue
		}

		// Сколько валюты осталось на покупку и продажу
		buyRemaining := buy.Remaining()
		sellRemaining := sell.Remaining()

		// Берем меньшее количество
		quantity := buyRemaining
		if sellRemaining < quantity {
			quantity = sellRemaining
		}
		if quantity <= 0 {
			continue
		}

		tradePrice := sell.Price

		// Обновляем исполненное количество
		buy.Filled += quantity
		sell.Filled += quantity

		now := time.Now().UTC()
		buy.UpdatedAt = now
		sell.UpdatedAt = now

		// Обновляем статус покупателя
		if buy.Remaining() <= 0 {
			buy.Status = enum.OrderStatusFilled
		} else {
			buy.Status = enum.OrderStatusPartiallyFilled
		}

		// Обновляем статус продавца.
		if sell.Remaining() <= 0 {
			sell.Status = enum.OrderStatusFilled
		} else {
			sell.Status = enum.OrderStatusPartiallyFilled
		}

		// Сохраняем изменённую заявку продавца.
		orderBook.data[sellID] = sell

		// Логируем совершённую сделку.
		logger.Info("Сделка исполнена",
			slog.String("BuyOrderID", buy.ID.String()),
			slog.String("SellOrderID", sell.ID.String()),
			slog.String("Symbol", buy.Symbo),
			slog.Float64("Price", tradePrice),
			slog.Float64("Quantity", quantity),
			slog.Float64("Total", tradePrice*quantity),
			slog.String("BuyStatus", buy.Status.String()),
			slog.String("SellStatus", sell.Status.String()),
		)
	}

	// Сохраняем итоговое состояние покупателя.
	orderBook.data[key] = buy

	logger.Info("Обработка покупки завершена",
		slog.String("OrderID", buy.ID.String()),
		slog.Float64("Filled", buy.Filled),
		slog.Float64("Remaining", buy.Remaining()),
		slog.String("Status", buy.Status.String()),
	)

	return buy.Status == enum.OrderStatusFilled
}

type Storage[K comparable] struct {
	storage map[K]struct{}
}

func NewStorage[T comparable](capacity int) *Storage[T] {
	return &Storage[T]{
		storage: make(map[T]struct{}, capacity),
	}
}

func (s *Storage[T]) Add(item T) {
	s.storage[item] = struct{}{}
}

func (s *Storage[T]) Set(value T) {
	s.storage[value] = struct{}{}
}

func (s *Storage[T]) Remove(value T) {
	delete(s.storage, value)
}

func (s *Storage[T]) Has(value T) bool {
	_, ok := s.storage[value]
	return ok
}

func (s *Storage[T]) Length() int {
	return len(s.storage)
}

func (s *Storage[T]) Items() []T {
	outSlice := make([]T, 0, len(s.storage))
	for value := range s.storage {
		outSlice = append(outSlice, value)
	}
	return outSlice
}
