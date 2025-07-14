package service

import (
	"world/internal/config"
)

// Stubs pour les services en attendant leur implémentation complète

// NPCService gère la logique métier des NPCs
type NPCService struct {
	config *config.Config
}

func NewNPCService(config *config.Config) *NPCService {
	return &NPCService{config: config}
}

// WorldEventService gère la logique métier des événements du monde
type WorldEventService struct {
	config *config.Config
}

func NewWorldEventService(config *config.Config) *WorldEventService {
	return &WorldEventService{config: config}
}

// WeatherService gère la logique métier de la météo
type WeatherService struct {
	config *config.Config
}

func NewWeatherService(config *config.Config) *WeatherService {
	return &WeatherService{config: config}
}