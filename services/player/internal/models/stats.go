package models

import (
	"time"

	"github.com/google/uuid"
)

// CharacterStats représente les statistiques d'un personnage
type CharacterStats struct {
	CharacterID  uuid.UUID `json:"character_id" db:"character_id"`
	
	// Statistiques de base
	Health       int `json:"health" db:"health"`
	MaxHealth    int `json:"max_health" db:"max_health"`
	Mana         int `json:"mana" db:"mana"`
	MaxMana      int `json:"max_mana" db:"max_mana"`
	
	// Attributs principaux
	Strength     int `json:"strength" db:"strength"`
	Agility      int `json:"agility" db:"agility"`
	Intelligence int `json:"intelligence" db:"intelligence"`
	Vitality     int `json:"vitality" db:"vitality"`
	
	// Points disponibles
	StatPoints   int `json:"stat_points" db:"stat_points"`
	SkillPoints  int `json:"skill_points" db:"skill_points"`
	
	// Statistiques dérivées (calculées)
	PhysicalDamage  int `json:"physical_damage" db:"physical_damage"`
	MagicalDamage   int `json:"magical_damage" db:"magical_damage"`
	PhysicalDefense int `json:"physical_defense" db:"physical_defense"`
	MagicalDefense  int `json:"magical_defense" db:"magical_defense"`
	CriticalChance  int `json:"critical_chance" db:"critical_chance"`   // en pourcentage
	AttackSpeed     int `json:"attack_speed" db:"attack_speed"`        // en pourcentage
	MovementSpeed   int `json:"movement_speed" db:"movement_speed"`    // en pourcentage
	
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CombatStats représente les statistiques de combat
type CombatStats struct {
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	
	// Statistiques PvE
	MonstersKilled    int `json:"monsters_killed" db:"monsters_killed"`
	BossesKilled      int `json:"bosses_killed" db:"bosses_killed"`
	Deaths            int `json:"deaths" db:"deaths"`
	DamageDealt       int64 `json:"damage_dealt" db:"damage_dealt"`
	DamageTaken       int64 `json:"damage_taken" db:"damage_taken"`
	HealingDone       int64 `json:"healing_done" db:"healing_done"`
	
	// Statistiques PvP
	PvPKills         int `json:"pvp_kills" db:"pvp_kills"`
	PvPDeaths        int `json:"pvp_deaths" db:"pvp_deaths"`
	PvPDamageDealt   int64 `json:"pvp_damage_dealt" db:"pvp_damage_dealt"`
	PvPDamageTaken   int64 `json:"pvp_damage_taken" db:"pvp_damage_taken"`
	
	// Statistiques générales
	QuestsCompleted  int `json:"quests_completed" db:"quests_completed"`
	ItemsLooted      int `json:"items_looted" db:"items_looted"`
	GoldEarned       int64 `json:"gold_earned" db:"gold_earned"`
	DistanceTraveled float64 `json:"distance_traveled" db:"distance_traveled"` // en kilomètres
	
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateStatsRequest représente une demande de mise à jour des stats
type UpdateStatsRequest struct {
	Strength     *int `json:"strength,omitempty"`
	Agility      *int `json:"agility,omitempty"`
	Intelligence *int `json:"intelligence,omitempty"`
	Vitality     *int `json:"vitality,omitempty"`
}

// StatsResponse représente la réponse complète des statistiques
type StatsResponse struct {
	BaseStats   *CharacterStats   `json:"base_stats"`
	CombatStats *CombatStats      `json:"combat_stats"`
	Modifiers   []*StatModifier   `json:"modifiers,omitempty"` // Changé pour utiliser des pointeurs
}

// StatModifier représente un modificateur temporaire de statistiques
type StatModifier struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CharacterID uuid.UUID `json:"character_id" db:"character_id"`
	Type        string    `json:"type" db:"type"`         // buff, debuff, equipment
	Source      string    `json:"source" db:"source"`     // spell_name, item_id, etc.
	StatName    string    `json:"stat_name" db:"stat_name"`
	Value       int       `json:"value" db:"value"`       // peut être négatif
	Duration    int       `json:"duration" db:"duration"` // en secondes, 0 = permanent
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CalculateMaxHealth calcule la santé maximale basée sur la vitalité
func (s *CharacterStats) CalculateMaxHealth() int {
	// Formule: 100 HP de base + (vitality * 10)
	return 100 + (s.Vitality * 10)
}

// CalculateMaxMana calcule le mana maximal basé sur l'intelligence
func (s *CharacterStats) CalculateMaxMana() int {
	// Formule: 50 MP de base + (intelligence * 5)
	return 50 + (s.Intelligence * 5)
}

// CalculatePhysicalDamage calcule les dégâts physiques
func (s *CharacterStats) CalculatePhysicalDamage() int {
	// Formule: strength * 2 + (agility * 0.5)
	return (s.Strength * 2) + (s.Agility / 2)
}

// CalculateMagicalDamage calcule les dégâts magiques
func (s *CharacterStats) CalculateMagicalDamage() int {
	// Formule: intelligence * 2.5
	return int(float64(s.Intelligence) * 2.5)
}

// CalculatePhysicalDefense calcule la défense physique
func (s *CharacterStats) CalculatePhysicalDefense() int {
	// Formule: (strength + vitality) / 3
	return (s.Strength + s.Vitality) / 3
}

// CalculateMagicalDefense calcule la défense magique
func (s *CharacterStats) CalculateMagicalDefense() int {
	// Formule: (intelligence + vitality) / 3
	return (s.Intelligence + s.Vitality) / 3
}

// CalculateCriticalChance calcule le chance de critique
func (s *CharacterStats) CalculateCriticalChance() int {
	// Formule: (agility / 10) + 5% de base, max 50%
	critChance := (s.Agility / 10) + 5
	if critChance > 50 {
		critChance = 50
	}
	return critChance
}

// CalculateAttackSpeed calcule la vitesse d'attaque
func (s *CharacterStats) CalculateAttackSpeed() int {
	// Formule: 100% de base + (agility / 5)%, max 200%
	attackSpeed := 100 + (s.Agility / 5)
	if attackSpeed > 200 {
		attackSpeed = 200
	}
	return attackSpeed
}

// CalculateMovementSpeed calcule la vitesse de déplacement
func (s *CharacterStats) CalculateMovementSpeed() int {
	// Formule: 100% de base + (agility / 10)%, max 150%
	moveSpeed := 100 + (s.Agility / 10)
	if moveSpeed > 150 {
		moveSpeed = 150
	}
	return moveSpeed
}

// RecalculateAll recalcule toutes les statistiques dérivées
func (s *CharacterStats) RecalculateAll() {
	s.MaxHealth = s.CalculateMaxHealth()
	s.MaxMana = s.CalculateMaxMana()
	s.PhysicalDamage = s.CalculatePhysicalDamage()
	s.MagicalDamage = s.CalculateMagicalDamage()
	s.PhysicalDefense = s.CalculatePhysicalDefense()
	s.MagicalDefense = s.CalculateMagicalDefense()
	s.CriticalChance = s.CalculateCriticalChance()
	s.AttackSpeed = s.CalculateAttackSpeed()
	s.MovementSpeed = s.CalculateMovementSpeed()
	s.UpdatedAt = time.Now()
}

// AddStatPoints ajoute des points de statistiques
func (s *CharacterStats) AddStatPoints(points int) {
	s.StatPoints += points
	s.UpdatedAt = time.Now()
}

// SpendStatPoint dépense un point de statistique
func (s *CharacterStats) SpendStatPoint(stat string, amount int) bool {
	if s.StatPoints < amount {
		return false
	}
	
	switch stat {
	case "strength":
		s.Strength += amount
	case "agility":
		s.Agility += amount
	case "intelligence":
		s.Intelligence += amount
	case "vitality":
		s.Vitality += amount
	default:
		return false
	}
	
	s.StatPoints -= amount
	s.RecalculateAll()
	return true
}

// RestoreHealth restaure de la santé
func (s *CharacterStats) RestoreHealth(amount int) {
	s.Health += amount
	if s.Health > s.MaxHealth {
		s.Health = s.MaxHealth
	}
	s.UpdatedAt = time.Now()
}

// RestoreMana restaure du mana
func (s *CharacterStats) RestoreMana(amount int) {
	s.Mana += amount
	if s.Mana > s.MaxMana {
		s.Mana = s.MaxMana
	}
	s.UpdatedAt = time.Now()
}

// TakeDamage applique des dégâts
func (s *CharacterStats) TakeDamage(amount int) bool {
	s.Health -= amount
	if s.Health < 0 {
		s.Health = 0
	}
	s.UpdatedAt = time.Now()
	return s.Health == 0 // true si mort
}

// SpendMana dépense du mana
func (s *CharacterStats) SpendMana(amount int) bool {
	if s.Mana < amount {
		return false
	}
	s.Mana -= amount
	s.UpdatedAt = time.Now()
	return true
}

// GetTotalStatPoints calcule le total des points de statistiques dépensés
func (s *CharacterStats) GetTotalStatPoints() int {
	return s.Strength + s.Agility + s.Intelligence + s.Vitality
}

// GetStatValue récupère la valeur d'une statistique par nom
func (s *CharacterStats) GetStatValue(statName string) int {
	switch statName {
	case "health":
		return s.Health
	case "max_health":
		return s.MaxHealth
	case "mana":
		return s.Mana
	case "max_mana":
		return s.MaxMana
	case "strength":
		return s.Strength
	case "agility":
		return s.Agility
	case "intelligence":
		return s.Intelligence
	case "vitality":
		return s.Vitality
	case "physical_damage":
		return s.PhysicalDamage
	case "magical_damage":
		return s.MagicalDamage
	case "physical_defense":
		return s.PhysicalDefense
	case "magical_defense":
		return s.MagicalDefense
	case "critical_chance":
		return s.CriticalChance
	case "attack_speed":
		return s.AttackSpeed
	case "movement_speed":
		return s.MovementSpeed
	default:
		return 0
	}
}

// IsExpired vérifie si un modificateur est expiré
func (m *StatModifier) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false // permanent
	}
	return time.Now().After(*m.ExpiresAt)
}