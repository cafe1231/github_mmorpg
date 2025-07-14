// internal/service/damage.go - Service de calcul des dégâts
package service

import (
	"math"
	"math/rand"

	"combat/internal/config"
	"combat/internal/models"
)

// DamageCalculatorInterface définit les méthodes de calcul des dégâts
type DamageCalculatorInterface interface {
	CalculatePhysicalDamage(attacker, defender *models.CombatParticipant, powerLevel float64) int
	CalculateMagicalDamage(caster, target *models.CombatParticipant, baseDamage int, element string) int
	IsCriticalHit(critChance float64) bool
	CalculateResistance(defender *models.CombatParticipant, damageType string) float64
}

// DamageCalculator implémente l'interface DamageCalculatorInterface
type DamageCalculator struct {
	config *config.Config
}

// NewDamageCalculator crée une nouvelle instance du calculateur de dégâts
func NewDamageCalculator(cfg *config.Config) DamageCalculatorInterface {
	return &DamageCalculator{
		config: cfg,
	}
}

// CalculatePhysicalDamage calcule les dégâts physiques
func (dc *DamageCalculator) CalculatePhysicalDamage(attacker, defender *models.CombatParticipant, powerLevel float64) int {
	// Dégâts de base de l'attaquant
	baseDamage := float64(attacker.Damage) * powerLevel * dc.config.Combat.BaseDamageMultiplier

	// Réduction par la défense du défenseur
	defense := float64(defender.Defense)
	damageReduction := defense / (defense + 100) // Formule de réduction asymptotique

	// Dégâts finaux
	finalDamage := baseDamage * (1 - damageReduction)

	// Variance aléatoire (±10%)
	variance := 0.9 + rand.Float64()*0.2
	finalDamage *= variance

	// Minimum 1 dégât
	if finalDamage < 1 {
		finalDamage = 1
	}

	return int(math.Round(finalDamage))
}

// CalculateMagicalDamage calcule les dégâts magiques
func (dc *DamageCalculator) CalculateMagicalDamage(caster, target *models.CombatParticipant, baseDamage int, element string) int {
	// Dégâts de base
	damage := float64(baseDamage)

	// Bonus d'intelligence du lanceur (les sorts scalent avec l'intelligence)
	intBonus := float64(caster.Damage) * 0.3 // 30% de l'intelligence en bonus
	damage += intBonus

	// Résistance magique du défenseur
	resistance := dc.CalculateResistance(target, "magical")
	damage *= (1 - resistance)

	// Résistance élémentaire spécifique
	elementalResistance := dc.getElementalResistance(target, element)
	damage *= (1 - elementalResistance)

	// Variance aléatoire
	variance := 0.85 + rand.Float64()*0.3 // ±15% pour la magie
	damage *= variance

	// Minimum 1 dégât
	if damage < 1 {
		damage = 1
	}

	return int(math.Round(damage))
}

// IsCriticalHit détermine si une attaque est critique
func (dc *DamageCalculator) IsCriticalHit(critChance float64) bool {
	return rand.Float64() < critChance
}

// CalculateResistance calcule la résistance générale
func (dc *DamageCalculator) CalculateResistance(defender *models.CombatParticipant, damageType string) float64 {
	// Résistance de base basée sur la défense
	baseResistance := float64(defender.Defense) * 0.001 // 0.1% par point de défense

	// Cap à 75% de résistance
	if baseResistance > 0.75 {
		baseResistance = 0.75
	}

	return baseResistance
}

// getElementalResistance calcule la résistance élémentaire
func (dc *DamageCalculator) getElementalResistance(target *models.CombatParticipant, element string) float64 {
	// Pour l'instant, résistance basique
	// Dans un vrai système, on récupérerait les résistances depuis les effets actifs
	switch element {
	case "fire":
		return 0.0
	case "ice":
		return 0.0
	case "lightning":
		return 0.0
	case "poison":
		return 0.1 // 10% de résistance de base au poison
	default:
		return 0.0
	}
}