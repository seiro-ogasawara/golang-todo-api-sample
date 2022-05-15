package model

import (
	"fmt"
	"strings"
	"time"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusNotReady
	StatusReady
	StatusDoing
	StatusDone
)

func (s Status) String() string {
	switch s {
	case StatusNotReady:
		return "Not Ready"
	case StatusReady:
		return "Ready"
	case StatusDoing:
		return "Doing"
	case StatusDone:
		return "Done"
	default:
		return "Unknown"
	}
}

func (s Status) GoString() string {
	return s.String()
}

func ToStatus(v int) Status {
	switch v {
	case 1:
		return StatusNotReady
	case 2:
		return StatusReady
	case 3:
		return StatusDoing
	case 4:
		return StatusDone
	default:
		return StatusUnknown
	}
}

type Priority int

const (
	PriorityUnknown Priority = iota
	PriorityHigh
	PriorityMiddle
	PriorityLow
)

func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "High"
	case PriorityMiddle:
		return "Middle"
	case PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}

func (p Priority) GoString() string {
	return p.String()
}

func ToPriority(v int) Priority {
	switch v {
	case 1:
		return PriorityHigh
	case 2:
		return PriorityMiddle
	case 3:
		return PriorityLow
	default:
		return PriorityUnknown
	}
}

type Todo struct {
	ID          int    `gorm:"primaryKey"`
	UserID      string `gorm:"not null"`
	Title       string `gorm:"not null"`
	Description string
	Status      Status    `gorm:"not null"`
	Priority    Priority  `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	User        *User
}

func (Todo) TableName() string {
	return "todos"
}

type Sorter string

const (
	SortByID       Sorter = "id"
	SortByPriority Sorter = "priority"
)

func ToSorter(v string) (Sorter, error) {
	lv := strings.ToLower(v)
	switch lv {
	case "id":
		return SortByID, nil
	case "priority":
		return SortByPriority, nil
	default:
		return "", fmt.Errorf("sorter must be id or priority, but %s", v)
	}
}

type Order string

const (
	OrderByASC  Order = "ASC"
	OrderByDESC Order = "DESC"
)

func ToOrder(v string) (Order, error) {
	lv := strings.ToLower(v)
	switch lv {
	case "asc":
		return OrderByASC, nil
	case "desc":
		return OrderByDESC, nil
	default:
		return "", fmt.Errorf("order must be asc or desc, but %s", v)
	}
}
