package config

// UI represents configs for the UI
type UI struct {
	Enable bool `validate:"required"`
}
