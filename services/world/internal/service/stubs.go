package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"world/internal/config"
	"world/internal/models"
	"world/internal/repository"

	"github.com/google/uuid"
)

// NPCService gère la logique métier des NPCs
type NPCService struct {
	config   *config.Config
	npcRepo  repository.NPCRepositoryInterface
	zoneRepo repository.ZoneRepositoryInterface
}

func NewNPCService(config *config.Config, npcRepo repository.NPCRepositoryInterface, zoneRepo repository.ZoneRepositoryInterface) *NPCService {
	return &NPCService{
		config:   config,
		npcRepo:  npcRepo,
		zoneRepo: zoneRepo,
	}
}

// CreateNPC crée un nouveau NPC avec validation
func (s *NPCService) CreateNPC(ctx context.Context, npc *models.NPC) error {
	// Validation des données
	if err := s.validateNPCData(ctx, npc); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Génération d'un ID unique
	npc.ID = uuid.New()
	npc.CreatedAt = time.Now()
	npc.UpdatedAt = time.Now()

	// Valeurs par défaut
	if npc.Status == "" {
		npc.Status = "active"
	}
	if npc.Scale == 0 {
		npc.Scale = 1.0
	}
	if npc.Health == 0 && npc.MaxHealth > 0 {
		npc.Health = npc.MaxHealth
	}

	return s.npcRepo.Create(npc)
}

// GetNPCsByZone récupère tous les NPCs d'une zone avec filtres
func (s *NPCService) GetNPCsByZone(ctx context.Context, zoneID string, npcType string) ([]*models.NPC, error) {
	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	if npcType != "" {
		// Si un type spécifique est demandé, filtrer par type
		allNPCs, err := s.npcRepo.GetByZone(zoneID)
		if err != nil {
			return nil, err
		}

		var filteredNPCs []*models.NPC
		for _, npc := range allNPCs {
			if npc.Type == npcType {
				filteredNPCs = append(filteredNPCs, npc)
			}
		}
		return filteredNPCs, nil
	}

	return s.npcRepo.GetByZone(zoneID)
}

// GetNearbyNPCs trouve les NPCs proches d'une position
func (s *NPCService) GetNearbyNPCs(ctx context.Context, zoneID string, x, y, z, radius float64) ([]*models.NPC, error) {
	return s.npcRepo.GetNearbyNPCs(zoneID, x, y, z, radius)
}

// UpdateNPCPosition met à jour la position d'un NPC
func (s *NPCService) UpdateNPCPosition(ctx context.Context, npcID uuid.UUID, x, y, z, rotation float64) error {
	npc, err := s.npcRepo.GetByID(npcID.String())
	if err != nil {
		return fmt.Errorf("NPC not found: %w", err)
	}

	// Vérifier que la position est valide dans la zone
	if err := s.validatePositionInZone(ctx, npc.ZoneID, x, y, z); err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	npc.PositionX = x
	npc.PositionY = y
	npc.PositionZ = z
	npc.Rotation = rotation
	npc.UpdatedAt = time.Now()
	npc.LastSeen = time.Now()

	return s.npcRepo.Update(npc)
}

// UpdateNPCHealth met à jour la santé d'un NPC
func (s *NPCService) UpdateNPCHealth(ctx context.Context, npcID uuid.UUID, health int) error {
	npc, err := s.npcRepo.GetByID(npcID.String())
	if err != nil {
		return fmt.Errorf("NPC not found: %w", err)
	}

	if health < 0 {
		health = 0
	}
	if health > npc.MaxHealth {
		health = npc.MaxHealth
	}

	npc.Health = health
	npc.UpdatedAt = time.Now()
	npc.LastSeen = time.Now()

	// Si le NPC meurt, changer son statut
	if health == 0 {
		npc.Status = "dead"
	} else if npc.Status == "dead" && health > 0 {
		npc.Status = "active"
	}

	return s.npcRepo.Update(npc)
}

// RespawnNPC fait réapparaître un NPC mort
func (s *NPCService) RespawnNPC(ctx context.Context, npcID uuid.UUID) error {
	npc, err := s.npcRepo.GetByID(npcID.String())
	if err != nil {
		return fmt.Errorf("NPC not found: %w", err)
	}

	if npc.Status != "dead" {
		return fmt.Errorf("NPC is not dead")
	}

	npc.Health = npc.MaxHealth
	npc.Status = "active"
	npc.UpdatedAt = time.Now()
	npc.LastSeen = time.Now()

	return s.npcRepo.Update(npc)
}

// validateNPCData valide les données d'un NPC
func (s *NPCService) validateNPCData(ctx context.Context, npc *models.NPC) error {
	if npc.Name == "" {
		return fmt.Errorf("name is required")
	}
	if npc.ZoneID == "" {
		return fmt.Errorf("zone_id is required")
	}
	if npc.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(npc.ZoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	// Valider la position dans la zone
	return s.validatePositionInZone(ctx, npc.ZoneID, npc.PositionX, npc.PositionY, npc.PositionZ)
}

// validatePositionInZone vérifie qu'une position est valide dans une zone
func (s *NPCService) validatePositionInZone(ctx context.Context, zoneID string, x, y, z float64) error {
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return err
	}

	if x < zone.MinX || x > zone.MaxX ||
		y < zone.MinY || y > zone.MaxY ||
		z < zone.MinZ || z > zone.MaxZ {
		return fmt.Errorf("position outside zone bounds")
	}

	return nil
}

// WorldEventService gère la logique métier des événements du monde
type WorldEventService struct {
	config    *config.Config
	eventRepo repository.WorldEventRepositoryInterface
	zoneRepo  repository.ZoneRepositoryInterface
}

func NewWorldEventService(config *config.Config, eventRepo repository.WorldEventRepositoryInterface, zoneRepo repository.ZoneRepositoryInterface) *WorldEventService {
	return &WorldEventService{
		config:    config,
		eventRepo: eventRepo,
		zoneRepo:  zoneRepo,
	}
}

// CreateEvent crée un nouvel événement
func (s *WorldEventService) CreateEvent(ctx context.Context, event *models.WorldEvent) error {
	// Validation
	if err := s.validateEventData(ctx, event); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Génération d'un ID unique
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	// Valeurs par défaut
	if event.Status == "" {
		event.Status = "scheduled"
	}
	if event.Duration > 0 && event.EndTime.IsZero() {
		event.EndTime = event.StartTime.Add(time.Duration(event.Duration) * time.Minute)
	}

	return s.eventRepo.Create(event)
}

// GetActiveEvents récupère les événements actifs
func (s *WorldEventService) GetActiveEvents(ctx context.Context, zoneID string) ([]*models.WorldEvent, error) {
	if zoneID == "" {
		// Tous les événements actifs
		return s.eventRepo.GetActive()
	}

	// Filtrer par zone
	allActive, err := s.eventRepo.GetActive()
	if err != nil {
		return nil, err
	}

	var filteredEvents []*models.WorldEvent
	for _, event := range allActive {
		if event.ZoneID == zoneID {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents, nil
}

// GetUpcomingEvents récupère les événements à venir
func (s *WorldEventService) GetUpcomingEvents(ctx context.Context, zoneID string, limit int) ([]*models.WorldEvent, error) {
	allUpcoming, err := s.eventRepo.GetUpcoming()
	if err != nil {
		return nil, err
	}

	var filteredEvents []*models.WorldEvent
	count := 0
	for _, event := range allUpcoming {
		if count >= limit && limit > 0 {
			break
		}
		if zoneID == "" || event.ZoneID == zoneID {
			filteredEvents = append(filteredEvents, event)
			count++
		}
	}
	return filteredEvents, nil
}

// StartEvent démarre un événement programmé
func (s *WorldEventService) StartEvent(ctx context.Context, eventID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(eventID.String())
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	if event.Status != "scheduled" {
		return fmt.Errorf("event is not scheduled")
	}

	now := time.Now()
	if event.StartTime.After(now) {
		return fmt.Errorf("event start time is in the future")
	}

	event.Status = "active"
	event.UpdatedAt = now

	// Si pas d'heure de fin définie, calculer à partir de la durée
	if event.EndTime.IsZero() && event.Duration > 0 {
		event.EndTime = now.Add(time.Duration(event.Duration) * time.Minute)
	}

	return s.eventRepo.Update(event)
}

// EndEvent termine un événement actif
func (s *WorldEventService) EndEvent(ctx context.Context, eventID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(eventID.String())
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	if event.Status != "active" {
		return fmt.Errorf("event is not active")
	}

	event.Status = "completed"
	event.UpdatedAt = time.Now()
	event.EndTime = time.Now()

	return s.eventRepo.Update(event)
}

// CancelEvent annule un événement
func (s *WorldEventService) CancelEvent(ctx context.Context, eventID uuid.UUID) error {
	event, err := s.eventRepo.GetByID(eventID.String())
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	if event.Status == "completed" {
		return fmt.Errorf("cannot cancel completed event")
	}

	event.Status = "cancelled"
	event.UpdatedAt = time.Now()

	return s.eventRepo.Update(event)
}

// ProcessScheduledEvents traite les événements programmés pour démarrage automatique
func (s *WorldEventService) ProcessScheduledEvents(ctx context.Context) error {
	now := time.Now()

	// Récupérer les événements programmés dont l'heure de début est passée
	events, err := s.eventRepo.GetByStatus("scheduled")
	if err != nil {
		return fmt.Errorf("failed to get scheduled events: %w", err)
	}

	for _, event := range events {
		if !event.StartTime.After(now) {
			if err := s.StartEvent(ctx, event.ID); err != nil {
				log.Printf("Failed to start event %s: %v", event.ID, err)
			}
		}
	}

	// Terminer les événements actifs dont l'heure de fin est passée
	activeEvents, err := s.eventRepo.GetActive()
	if err != nil {
		return fmt.Errorf("failed to get active events: %w", err)
	}

	for _, event := range activeEvents {
		if !event.EndTime.IsZero() && event.EndTime.Before(now) {
			if err := s.EndEvent(ctx, event.ID); err != nil {
				log.Printf("Failed to end event %s: %v", event.ID, err)
			}
		}
	}

	return nil
}

// validateEventData valide les données d'un événement
func (s *WorldEventService) validateEventData(ctx context.Context, event *models.WorldEvent) error {
	if event.Name == "" {
		return fmt.Errorf("name is required")
	}
	if event.Type == "" {
		return fmt.Errorf("type is required")
	}
	if event.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}

	// Vérifier que la zone existe si spécifiée
	if event.ZoneID != "" {
		_, err := s.zoneRepo.GetByID(event.ZoneID)
		if err != nil {
			return fmt.Errorf("zone not found: %w", err)
		}
	}

	// Validation des niveaux
	if event.MinLevel < 0 {
		return fmt.Errorf("min_level cannot be negative")
	}
	if event.MaxLevel > 0 && event.MaxLevel < event.MinLevel {
		return fmt.Errorf("max_level cannot be less than min_level")
	}

	return nil
}

// WeatherService gère la logique métier de la météo
type WeatherService struct {
	config      *config.Config
	weatherRepo repository.WeatherRepositoryInterface
	zoneRepo    repository.ZoneRepositoryInterface
}

func NewWeatherService(config *config.Config, weatherRepo repository.WeatherRepositoryInterface, zoneRepo repository.ZoneRepositoryInterface) *WeatherService {
	return &WeatherService{
		config:      config,
		weatherRepo: weatherRepo,
		zoneRepo:    zoneRepo,
	}
}

// SetWeather définit la météo pour une zone
func (s *WeatherService) SetWeather(ctx context.Context, weather *models.Weather) error {
	// Validation
	if err := s.validateWeatherData(ctx, weather); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Valeurs par défaut
	if weather.StartTime.IsZero() {
		weather.StartTime = time.Now()
	}
	if weather.EndTime.IsZero() {
		// Météo par défaut pendant 2 heures
		weather.EndTime = weather.StartTime.Add(2 * time.Hour)
	}
	weather.IsActive = true
	weather.CreatedAt = time.Now()
	weather.UpdatedAt = time.Now()

	return s.weatherRepo.Upsert(weather)
}

// GetCurrentWeather récupère la météo actuelle d'une zone
func (s *WeatherService) GetCurrentWeather(ctx context.Context, zoneID string) (*models.Weather, error) {
	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	return s.weatherRepo.GetByZone(zoneID)
}

// GenerateRandomWeather génère une météo aléatoire pour une zone
func (s *WeatherService) GenerateRandomWeather(ctx context.Context, zoneID string, duration time.Duration) error {
	zone, err := s.zoneRepo.GetByID(zoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	// Types de météo possibles selon le type de zone
	var weatherTypes []string
	switch zone.Type {
	case "city", "safe":
		weatherTypes = []string{"clear", "rain", "fog"}
	case "wilderness":
		weatherTypes = []string{"clear", "rain", "storm", "fog"}
	case "dungeon":
		weatherTypes = []string{"fog", "clear"}
	default:
		weatherTypes = []string{"clear", "rain", "storm", "snow", "fog"}
	}

	// Sélection aléatoire
	weatherType := weatherTypes[rand.Intn(len(weatherTypes))]

	// Paramètres aléatoires selon le type
	weather := &models.Weather{
		ZoneID:    zoneID,
		Type:      weatherType,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(duration),
	}

	switch weatherType {
	case "clear":
		weather.Intensity = 0.0
		weather.Temperature = 15.0 + rand.Float64()*20.0 // 15-35°C
		weather.WindSpeed = rand.Float64() * 10.0        // 0-10 km/h
		weather.Visibility = 10000.0                     // 10km
	case "rain":
		weather.Intensity = 0.3 + rand.Float64()*0.4        // 0.3-0.7
		weather.Temperature = 5.0 + rand.Float64()*15.0     // 5-20°C
		weather.WindSpeed = 5.0 + rand.Float64()*15.0       // 5-20 km/h
		weather.Visibility = 1000.0 + rand.Float64()*4000.0 // 1-5km
	case "storm":
		weather.Intensity = 0.7 + rand.Float64()*0.3      // 0.7-1.0
		weather.Temperature = 10.0 + rand.Float64()*10.0  // 10-20°C
		weather.WindSpeed = 20.0 + rand.Float64()*30.0    // 20-50 km/h
		weather.Visibility = 300.0 + rand.Float64()*700.0 // 300m-1km
	case "snow":
		weather.Intensity = 0.2 + rand.Float64()*0.6       // 0.2-0.8
		weather.Temperature = -10.0 + rand.Float64()*10.0  // -10-0°C
		weather.WindSpeed = rand.Float64() * 20.0          // 0-20 km/h
		weather.Visibility = 500.0 + rand.Float64()*2000.0 // 500m-2.5km
	case "fog":
		weather.Intensity = 0.4 + rand.Float64()*0.4     // 0.4-0.8
		weather.Temperature = 5.0 + rand.Float64()*20.0  // 5-25°C
		weather.WindSpeed = rand.Float64() * 5.0         // 0-5 km/h
		weather.Visibility = 50.0 + rand.Float64()*450.0 // 50-500m
	}

	weather.WindDirection = rand.Float64() * 360.0 // 0-360°

	return s.SetWeather(ctx, weather)
}

// UpdateWeatherSystem met à jour le système météo de toutes les zones
func (s *WeatherService) UpdateWeatherSystem(ctx context.Context) error {
	// Récupérer toutes les zones actives
	zones, err := s.zoneRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}

	now := time.Now()

	for _, zone := range zones {
		if zone.Status != "active" {
			continue
		}

		// Vérifier la météo actuelle
		currentWeather, err := s.weatherRepo.GetByZone(zone.ID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error getting weather for zone %s: %v", zone.ID, err)
			continue
		}

		// Si pas de météo active ou météo expirée, en générer une nouvelle
		if currentWeather == nil || now.After(currentWeather.EndTime) {
			// Durée aléatoire entre 1 et 6 heures
			duration := time.Duration(1+rand.Intn(5)) * time.Hour

			if err := s.GenerateRandomWeather(ctx, zone.ID, duration); err != nil {
				log.Printf("Failed to generate weather for zone %s: %v", zone.ID, err)
			}
		}
	}

	return nil
}

// validateWeatherData valide les données météo
func (s *WeatherService) validateWeatherData(ctx context.Context, weather *models.Weather) error {
	if weather.ZoneID == "" {
		return fmt.Errorf("zone_id is required")
	}
	if weather.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Vérifier que la zone existe
	_, err := s.zoneRepo.GetByID(weather.ZoneID)
	if err != nil {
		return fmt.Errorf("zone not found: %w", err)
	}

	// Validation des valeurs
	if weather.Intensity < 0.0 || weather.Intensity > 1.0 {
		return fmt.Errorf("intensity must be between 0.0 and 1.0")
	}
	if weather.WindSpeed < 0.0 {
		return fmt.Errorf("wind_speed cannot be negative")
	}
	if weather.WindDirection < 0.0 || weather.WindDirection >= 360.0 {
		return fmt.Errorf("wind_direction must be between 0.0 and 359.9")
	}
	if weather.Visibility < 0.0 {
		return fmt.Errorf("visibility cannot be negative")
	}

	// Validation des types de météo
	validTypes := []string{"clear", "rain", "storm", "snow", "fog"}
	isValidType := false
	for _, validType := range validTypes {
		if weather.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid weather type: %s", weather.Type)
	}

	return nil
}
