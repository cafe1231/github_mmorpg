package service

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/models"
)

type DamageCalculatorInterface interface {
	CalculatePhysicalDamage(attacker *models.CombatParticipant, defender *models.CombatParticipant, actionData map[string]interface{}) (*models.DamageResult, error)
	CalculateMagicalDamage(attacker *models.CombatParticipant, defender *models.CombatParticipant, spell *models.Spell) (*models.DamageResult, error)
	CalculateHealing(caster *models.CombatParticipant, target *models.CombatParticipant, spell *models.Spell) (*models.HealingResult, error)
	ApplyResistances(baseDamage int, damageType string, resistances map[string]float64) int
	CalculateCriticalHit(attacker *models.CombatParticipant, baseDamage int) (int, bool)
	CalculateArmorReduction(damage int, armor int) int
}

type DamageCalculator struct {
	rng *rand.Rand
}

func NewDamageCalculator() DamageCalculatorInterface {
	return &DamageCalculator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (dc *DamageCalculator) CalculatePhysicalDamage(attacker *models.CombatParticipant, defender *models.CombatParticipant, actionData map[string]interface{}) (*models.DamageResult, error) {
	// TODO: Récupérer les vraies stats depuis le service Player
	baseDamage := 20 // Valeur par défaut
	
	// Variance
	variance := dc.rng.Float64()*0.2 - 0.1
	finalDamage := int(float64(baseDamage) * (1.0 + variance))
	
	// Critique
	criticalDamage, isCritical := dc.CalculateCriticalHit(attacker, finalDamage)
	if isCritical {
		finalDamage = criticalDamage
	}
	
	if finalDamage < 1 {
		finalDamage = 1
	}

	return &models.DamageResult{
		Amount:      finalDamage,
		Type:        "physical",
		IsCritical:  isCritical,
		IsResisted:  false,
		Penetration: 0,
		Source:      "melee_attack",
		ElementType: "",
	}, nil
}

func (dc *DamageCalculator) CalculateMagicalDamage(attacker *models.CombatParticipant, defender *models.CombatParticipant, spell *models.Spell) (*models.DamageResult, error) {
	// Utiliser les effets du sort pour calculer les dégâts
	baseDamage := 0
	elementType := "arcane" // Par défaut
	
	// Parcourir les effets du sort pour trouver les dégâts
	for _, effect := range spell.Effects {
		if effect.Type == "damage" {
			baseDamage = effect.Value
			elementType = effect.Element
			break
		}
	}
	
	if baseDamage == 0 {
		baseDamage = 25 // Valeur par défaut
	}
	
	// Variance
	variance := dc.rng.Float64()*0.15 - 0.075
	finalDamage := int(float64(baseDamage) * (1.0 + variance))
	
	// Critique
	criticalDamage, isCritical := dc.CalculateCriticalHit(attacker, finalDamage)
	if isCritical {
		finalDamage = criticalDamage
	}
	
	if finalDamage < 1 {
		finalDamage = 1
	}

	return &models.DamageResult{
		Amount:      finalDamage,
		Type:        "magical",
		IsCritical:  isCritical,
		IsResisted:  false,
		Penetration: 0,
		Source:      spell.Name,
		ElementType: elementType,
	}, nil
}

func (dc *DamageCalculator) CalculateHealing(caster *models.CombatParticipant, target *models.CombatParticipant, spell *models.Spell) (*models.HealingResult, error) {
	// Utiliser les effets du sort pour calculer les soins
	baseHealing := 0
	
	for _, effect := range spell.Effects {
		if effect.Type == "heal" {
			baseHealing = effect.Value
			break
		}
	}
	
	if baseHealing == 0 {
		baseHealing = 30 // Valeur par défaut
	}
	
	// Variance
	variance := dc.rng.Float64()*0.1 - 0.05
	finalHealing := int(float64(baseHealing) * (1.0 + variance))
	
	// Critique
	criticalHealing, isCritical := dc.CalculateCriticalHit(caster, finalHealing)
	if isCritical {
		finalHealing = criticalHealing
	}
	
	if finalHealing < 0 {
		finalHealing = 0
	}

	return &models.HealingResult{
		Amount:       finalHealing,
		IsCritical:   isCritical,
		Source:       spell.Name,
		Overheal:     0,
		TargetHealth: 100, // TODO: Récupérer la vraie valeur
	}, nil
}

func (dc *DamageCalculator) ApplyResistances(baseDamage int, damageType string, resistances map[string]float64) int {
	resistance, exists := resistances[damageType]
	if !exists {
		resistance = 0.0
	}
	
	if resistance < 0 {
		resistance = 0
	}
	if resistance > 0.95 {
		resistance = 0.95
	}
	
	finalDamage := int(float64(baseDamage) * (1.0 - resistance))
	return finalDamage
}

func (dc *DamageCalculator) CalculateCriticalHit(attacker *models.CombatParticipant, baseDamage int) (int, bool) {
	// TODO: Récupérer les vraies stats
	criticalChance := 10 // 10%
	criticalMultiplier := 2.0
	
	roll := dc.rng.Intn(100) + 1
	isCritical := roll <= criticalChance
	
	if isCritical {
		criticalDamage := int(float64(baseDamage) * criticalMultiplier)
		return criticalDamage, true
	}
	
	return baseDamage, false
}

func (dc *DamageCalculator) CalculateArmorReduction(damage int, armor int) int {
	if armor <= 0 {
		return damage
	}
	
	reduction := float64(armor) / (float64(armor) + 100.0)
	if reduction > 0.9 {
		reduction = 0.9
	}
	
	finalDamage := int(float64(damage) * (1.0 - reduction))
	return finalDamage
}