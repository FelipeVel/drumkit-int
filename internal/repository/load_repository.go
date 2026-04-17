package repository

import "github.com/FelipeVel/drumkit-int/internal/model"

// LoadRepository defines the data-access contract for loads.
// The service layer depends on this interface, never on a concrete type,
// which makes the external provider swappable without touching business logic.
type LoadRepository interface {
	GetAll() ([]model.Load, error)
	Create(load model.Load) (int, error)
}
