package domain

import "time"

type OrderCursor struct {
	CreatedAt time.Time
	OrderID   string
}

type ListOrdersFilter struct {
	UserID       string
	PageSize     int
	Cursor       *OrderCursor
	StatusFilter *OrderStatus
}

type OrderPage struct {
	Orders        []Order
	NextCursor    *OrderCursor
	NextPageToken string
}
