package models

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// Zone représente une zone/map du monde
type Zone struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	DisplayName string `json:"display_name" db:"display_name"`
	Description string `json:"description" db:"description"`
	Type        string `json:"type" db:"type"`   // city, dungeon, wilderness, pvp, safe
	Level       int    `json:"level" db:"level"` // niveau recommandé

	// Géométrie de la zone
	MinX float64 `json:"min_x" db:"min_x"`
	MinY float64 `json:"min_y" db:"min_y"`
	MinZ float64 `json:"min_z" db:"min_z"`
	MaxX float64 `json:"max_x" db:"max_x"`
	MaxY float64 `json:"max_y" db:"max_y"`
	MaxZ float64 `json:"max_z" db:"max_z"`

	// Points de spawn
	SpawnX float64 `json:"spawn_x" db:"spawn_x"`
	SpawnY float64 `json:"spawn_y" db:"spawn_y"`
	SpawnZ float64 `json:"spawn_z" db:"spawn_z"`

	// Configuration
	MaxPlayers int          `json:"max_players" db:"max_players"`
	IsPvP      bool         `json:"is_pvp" db:"is_pvp"`
	IsSafeZone bool         `json:"is_safe_zone" db:"is_safe_zone"`
	Settings   ZoneSettings `json:"settings" db:"settings"`

	// État
	Status      string `json:"status" db:"status"`  // active, maintenance, disabled
	PlayerCount int    `json:"player_count" db:"-"` // calculé dynamiquement

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ZoneSettings configuration spécifique de la zone
type ZoneSettings struct {
	Weather              string   `json:"weather"`
	TimeOfDay            string   `json:"time_of_day"`
	BackgroundMusic      string   `json:"background_music"`
	AllowedClasses       []string `json:"allowed_classes"`
	RestrictedItems      []string `json:"restricted_items"`
	ExperienceMultiplier float64  `json:"experience_multiplier"`
	LootMultiplier       float64  `json:"loot_multiplier"`
	DeathPenalty         string   `json:"death_penalty"` // none, durability, item_loss
}

// NPC représente un personnage non-joueur
type NPC struct {
	ID      uuid.UUID `json:"id" db:"id"`
	ZoneID  string    `json:"zone_id" db:"zone_id"`
	Name    string    `json:"name" db:"name"`
	Type    string    `json:"type" db:"type"`       // merchant, guard, quest_giver, monster
	Subtype string    `json:"subtype" db:"subtype"` // specific type like blacksmith, banker

	// Position
	PositionX float64 `json:"position_x" db:"position_x"`
	PositionY float64 `json:"position_y" db:"position_y"`
	PositionZ float64 `json:"position_z" db:"position_z"`
	Rotation  float64 `json:"rotation" db:"rotation"`

	// Apparence
	Model   string  `json:"model" db:"model"`
	Texture string  `json:"texture" db:"texture"`
	Scale   float64 `json:"scale" db:"scale"`

	// Comportement
	Behavior NPCBehavior `json:"behavior" db:"behavior"`

	// Combat (pour les monstres)
	Level     int `json:"level" db:"level"`
	Health    int `json:"health" db:"health"`
	MaxHealth int `json:"max_health" db:"max_health"`

	// État
	Status   string    `json:"status" db:"status"` // active, inactive, dead
	LastSeen time.Time `json:"last_seen" db:"last_seen"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NPCBehavior comportement d'un NPC
type NPCBehavior struct {
	IsStationary    bool      `json:"is_stationary"`
	PatrolRoute     []Point3D `json:"patrol_route"`
	InteractionText string    `json:"interaction_text"`
	AggroRange      float64   `json:"aggro_range"`
	ChaseRange      float64   `json:"chase_range"`
	RespawnTime     int       `json:"respawn_time"` // en secondes
	IsHostile       bool      `json:"is_hostile"`
	Faction         string    `json:"faction"`
}

// PlayerPosition position d'un joueur dans le monde
type PlayerPosition struct {
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	ZoneID      string    `json:"zone_id" db:"zone_id"`

	// Position actuelle
	X        float64 `json:"x" db:"x"`
	Y        float64 `json:"y" db:"y"`
	Z        float64 `json:"z" db:"z"`
	Rotation float64 `json:"rotation" db:"rotation"`

	// Mouvement
	VelocityX float64 `json:"velocity_x" db:"velocity_x"`
	VelocityY float64 `json:"velocity_y" db:"velocity_y"`
	VelocityZ float64 `json:"velocity_z" db:"velocity_z"`
	IsMoving  bool    `json:"is_moving" db:"is_moving"`

	// État du joueur
	IsOnline   bool      `json:"is_online" db:"is_online"`
	LastUpdate time.Time `json:"last_update" db:"last_update"`

	// Informations additionnelles
	CharacterName  string `json:"character_name,omitempty" db:"-"`
	CharacterLevel int    `json:"character_level,omitempty" db:"-"`
}

// WorldEvent événement du monde
type WorldEvent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ZoneID      string    `json:"zone_id" db:"zone_id"` // null si global
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // boss_spawn, treasure_hunt, pvp_tournament

	// Timing
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Duration  int       `json:"duration" db:"duration"` // en minutes

	// Configuration
	MaxParticipants int           `json:"max_participants" db:"max_participants"`
	MinLevel        int           `json:"min_level" db:"min_level"`
	MaxLevel        int           `json:"max_level" db:"max_level"`
	Rewards         []EventReward `json:"rewards" db:"rewards"`

	// État
	Status           string `json:"status" db:"status"` // scheduled, active, completed, canceled
	ParticipantCount int    `json:"participant_count" db:"-"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// EventReward récompense d'événement
type EventReward struct {
	Type   string `json:"type"` // experience, gold, item
	Amount int    `json:"amount"`
	ItemID string `json:"item_id,omitempty"`
	Rarity string `json:"rarity,omitempty"`
}

// Weather météo d'une zone
type Weather struct {
	ZoneID        string  `json:"zone_id" db:"zone_id"`
	Type          string  `json:"type" db:"type"`                     // clear, rain, storm, snow, fog
	Intensity     float64 `json:"intensity" db:"intensity"`           // 0.0 à 1.0
	Temperature   float64 `json:"temperature" db:"temperature"`       // en Celsius
	WindSpeed     float64 `json:"wind_speed" db:"wind_speed"`         // km/h
	WindDirection float64 `json:"wind_direction" db:"wind_direction"` // degrés
	Visibility    float64 `json:"visibility" db:"visibility"`         // distance en mètres

	// Timing
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`

	// État
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Point3D représente un point dans l'espace 3D
type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// ZoneTransition transition entre zones
type ZoneTransition struct {
	ID         uuid.UUID `json:"id" db:"id"`
	FromZoneID string    `json:"from_zone_id" db:"from_zone_id"`
	ToZoneID   string    `json:"to_zone_id" db:"to_zone_id"`

	// Point de départ dans la zone source
	TriggerX      float64 `json:"trigger_x" db:"trigger_x"`
	TriggerY      float64 `json:"trigger_y" db:"trigger_y"`
	TriggerZ      float64 `json:"trigger_z" db:"trigger_z"`
	TriggerRadius float64 `json:"trigger_radius" db:"trigger_radius"`

	// Point d'arrivée dans la zone cible
	DestinationX float64 `json:"destination_x" db:"destination_x"`
	DestinationY float64 `json:"destination_y" db:"destination_y"`
	DestinationZ float64 `json:"destination_z" db:"destination_z"`

	// Conditions
	RequiredLevel int    `json:"required_level" db:"required_level"`
	RequiredQuest string `json:"required_quest" db:"required_quest"`

	// État
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Spawn point de spawn pour NPCs ou ressources
type Spawn struct {
	ID       uuid.UUID `json:"id" db:"id"`
	ZoneID   string    `json:"zone_id" db:"zone_id"`
	Type     string    `json:"type" db:"type"`           // npc, resource, item
	EntityID string    `json:"entity_id" db:"entity_id"` // ID de l'entité à spawner

	// Position
	X      float64 `json:"x" db:"x"`
	Y      float64 `json:"y" db:"y"`
	Z      float64 `json:"z" db:"z"`
	Radius float64 `json:"radius" db:"radius"` // zone de spawn aléatoire

	// Timing
	RespawnTime int       `json:"respawn_time" db:"respawn_time"` // secondes
	LastSpawn   time.Time `json:"last_spawn" db:"last_spawn"`
	NextSpawn   time.Time `json:"next_spawn" db:"next_spawn"`

	// Configuration
	MaxSpawns   int     `json:"max_spawns" db:"max_spawns"`     // nombre max simultané
	SpawnChance float64 `json:"spawn_chance" db:"spawn_chance"` // probabilité 0-1

	// Conditions
	RequiredTime    string `json:"required_time" db:"required_time"` // day, night, any
	RequiredWeather string `json:"required_weather" db:"required_weather"`

	// État
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DTOs pour les requêtes

// CreateZoneRequest requête de création de zone
type CreateZoneRequest struct {
	ID          string       `json:"id" binding:"required,min=3,max=50"`
	Name        string       `json:"name" binding:"required,min=3,max=100"`
	DisplayName string       `json:"display_name" binding:"required,min=3,max=100"`
	Description string       `json:"description" binding:"max=500"`
	Type        string       `json:"type" binding:"required,oneof=city dungeon wilderness pvp safe"`
	Level       int          `json:"level" binding:"min=1,max=100"`
	MinX        float64      `json:"min_x"`
	MinY        float64      `json:"min_y"`
	MinZ        float64      `json:"min_z"`
	MaxX        float64      `json:"max_x"`
	MaxY        float64      `json:"max_y"`
	MaxZ        float64      `json:"max_z"`
	SpawnX      float64      `json:"spawn_x"`
	SpawnY      float64      `json:"spawn_y"`
	SpawnZ      float64      `json:"spawn_z"`
	MaxPlayers  int          `json:"max_players" binding:"min=1,max=1000"`
	IsPvP       bool         `json:"is_pvp"`
	IsSafeZone  bool         `json:"is_safe_zone"`
	Settings    ZoneSettings `json:"settings"`
}

// UpdateZoneRequest requête de mise à jour de zone
type UpdateZoneRequest struct {
	Name        *string       `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	DisplayName *string       `json:"display_name,omitempty" binding:"omitempty,min=3,max=100"`
	Description *string       `json:"description,omitempty" binding:"omitempty,max=500"`
	Level       *int          `json:"level,omitempty" binding:"omitempty,min=1,max=100"`
	MaxPlayers  *int          `json:"max_players,omitempty" binding:"omitempty,min=1,max=1000"`
	IsPvP       *bool         `json:"is_pvp,omitempty"`
	IsSafeZone  *bool         `json:"is_safe_zone,omitempty"`
	Settings    *ZoneSettings `json:"settings,omitempty"`
	Status      *string       `json:"status,omitempty" binding:"omitempty,oneof=active maintenance disabled"`
}

// UpdatePositionRequest requête de mise à jour de position
type UpdatePositionRequest struct {
	ZoneID    string  `json:"zone_id" binding:"required"`
	X         float64 `json:"x" binding:"required"`
	Y         float64 `json:"y" binding:"required"`
	Z         float64 `json:"z" binding:"required"`
	Rotation  float64 `json:"rotation"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	VelocityZ float64 `json:"velocity_z"`
	IsMoving  bool    `json:"is_moving"`
}

// CreateNPCRequest requête de création de NPC
type CreateNPCRequest struct {
	ZoneID    string      `json:"zone_id" binding:"required"`
	Name      string      `json:"name" binding:"required,min=2,max=50"`
	Type      string      `json:"type" binding:"required,oneof=merchant guard quest_giver monster"`
	Subtype   string      `json:"subtype" binding:"max=50"`
	PositionX float64     `json:"position_x" binding:"required"`
	PositionY float64     `json:"position_y" binding:"required"`
	PositionZ float64     `json:"position_z" binding:"required"`
	Model     string      `json:"model" binding:"required"`
	Texture   string      `json:"texture"`
	Scale     float64     `json:"scale" binding:"min=0.1,max=10"`
	Behavior  NPCBehavior `json:"behavior"`
	Level     int         `json:"level" binding:"min=1,max=100"`
	MaxHealth int         `json:"max_health" binding:"min=1"`
}

// CreateEventRequest requête de création d'événement
type CreateEventRequest struct {
	ZoneID          *string       `json:"zone_id,omitempty"`
	Name            string        `json:"name" binding:"required,min=3,max=100"`
	Description     string        `json:"description" binding:"max=500"`
	Type            string        `json:"type" binding:"required"`
	StartTime       time.Time     `json:"start_time" binding:"required"`
	Duration        int           `json:"duration" binding:"min=1,max=1440"` // max 24h
	MaxParticipants int           `json:"max_participants" binding:"min=1"`
	MinLevel        int           `json:"min_level" binding:"min=1,max=100"`
	MaxLevel        int           `json:"max_level" binding:"min=1,max=100"`
	Rewards         []EventReward `json:"rewards"`
}

// SetWeatherRequest requête de définition météo
type SetWeatherRequest struct {
	Type          string  `json:"type" binding:"required,oneof=clear rain storm snow fog"`
	Intensity     float64 `json:"intensity" binding:"min=0,max=1"`
	Temperature   float64 `json:"temperature" binding:"min=-50,max=50"`
	WindSpeed     float64 `json:"wind_speed" binding:"min=0,max=200"`
	WindDirection float64 `json:"wind_direction" binding:"min=0,max=360"`
	Duration      int     `json:"duration" binding:"min=1,max=1440"` // minutes
}

// Constantes pour les types

// Types de zones
const (
	ZoneTypeCity       = "city"
	ZoneTypeDungeon    = "dungeon"
	ZoneTypeWilderness = "wilderness"
	ZoneTypePvP        = "pvp"
	ZoneTypeSafe       = "safe"
)

// Types de NPCs
const (
	NPCTypeMerchant   = "merchant"
	NPCTypeGuard      = "guard"
	NPCTypeQuestGiver = "quest_giver"
	NPCTypeMonster    = "monster"
)

// Types d'événements
const (
	EventTypeBossSpawn     = "boss_spawn"
	EventTypeTreasureHunt  = "treasure_hunt"
	EventTypePvPTournament = "pvp_tournament"
	EventTypeRaid          = "raid"
	EventTypeSiege         = "siege"
)

// Types de météo
const (
	WeatherTypeClear = "clear"
	WeatherTypeRain  = "rain"
	WeatherTypeStorm = "storm"
	WeatherTypeSnow  = "snow"
	WeatherTypeFog   = "fog"
)

// status
const (
	StatusActive      = "active"
	StatusInactive    = "inactive"
	StatusMaintenance = "maintenance"
	StatusDisabled    = "disabled"
	StatusScheduled   = "scheduled"
	StatusCompleted   = "completed"
	StatusCancelled   = "canceled"
)

// Fonctions utilitaires

// GetDefaultZoneSettings retourne les paramètres par défaut d'une zone
func GetDefaultZoneSettings() ZoneSettings {
	return ZoneSettings{
		Weather:              WeatherTypeClear,
		TimeOfDay:            "day",
		BackgroundMusic:      "ambient_forest",
		AllowedClasses:       []string{"warrior", "mage", "archer", "rogue"},
		RestrictedItems:      []string{},
		ExperienceMultiplier: 1.0,
		LootMultiplier:       1.0,
		DeathPenalty:         "durability",
	}
}

// GetDefaultNPCBehavior retourne le comportement par défaut d'un NPC
func GetDefaultNPCBehavior() NPCBehavior {
	return NPCBehavior{
		IsStationary:    true,
		PatrolRoute:     []Point3D{},
		InteractionText: "Hello, traveler!",
		AggroRange:      5.0,
		ChaseRange:      10.0,
		RespawnTime:     300, // 5 minutes
		IsHostile:       false,
		Faction:         "neutral",
	}
}

// Distance calcule la distance 3D entre deux points
func (p1 Point3D) Distance(p2 Point3D) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	dz := p1.Z - p2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// IsInZone vérifie si un point est dans les limites d'une zone
func (zone *Zone) IsInZone(x, y, z float64) bool {
	return x >= zone.MinX && x <= zone.MaxX &&
		y >= zone.MinY && y <= zone.MaxY &&
		z >= zone.MinZ && z <= zone.MaxZ
}

// GetSpawnPoint retourne le point de spawn de la zone
func (zone *Zone) GetSpawnPoint() Point3D {
	return Point3D{
		X: zone.SpawnX,
		Y: zone.SpawnY,
		Z: zone.SpawnZ,
	}
}

