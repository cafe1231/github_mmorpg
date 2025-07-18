package service

import (
	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/utils"
	"fmt"
)

// Constantes pour les types de dégâts
const (
	DamageTypePhysical = "physical"
	DamageTypeMagical  = "magical"

	// Constantes pour la variance des dégâts
	varianceMin = 0.9
	varianceMax = 0.2
)

// DamageCalculatorInterface définit les méthodes du calculateur de dégâts
type DamageCalculatorInterface interface {
	// calculations de base
	CalculateDamage(attacker, defender *models.CombatParticipant, skill *models.SkillInfo, modifiers map[string]float64) *DamageResult
	CalculateHealing(caster, target *models.CombatParticipant, skill *models.SkillInfo, modifiers map[string]float64) *HealingResult

	// calculations spécialisés
	CalculatePhysicalDamage(attacker, defender *models.CombatParticipant, baseDamage int, modifiers map[string]float64) int
	CalculateMagicalDamage(attacker, defender *models.CombatParticipant, baseDamage int, modifiers map[string]float64) int
	CalculateTrueDamage(baseDamage int, modifiers map[string]float64) int

	// Chances et critiques
	CalculateCriticalChance(attacker *models.CombatParticipant, skill *models.SkillInfo, modifiers map[string]float64) float64
	CalculateHitChance(attacker, defender *models.CombatParticipant, skill *models.SkillInfo, modifiers map[string]float64) float64
	CalculateBlockChance(defender *models.CombatParticipant, modifiers map[string]float64) float64

	// Modificateurs et résistances
	ApplyArmorReduction(damage, armor int, damageType string) int
	ApplyResistances(damage int, resistances map[string]float64, damageType string) int
	ApplyVulnerabilities(damage int, vulnerabilities map[string]float64, damageType string) int

	// calculations avancés
	CalculateElementalDamage(attacker *models.CombatParticipant, element string, baseDamage int) int
	CalculateDamageOverTime(effect *models.CombatEffect, target *models.CombatParticipant) int
	CalculateStatusEffectChance(caster, target *models.CombatParticipant, effect *models.SkillEffect) float64
}

// DamageCalculator implémente l'interface DamageCalculatorInterface
type DamageCalculator struct {
	config *config.Config
}

// DamageResult représente le résultat d'un calcul de dégâts
type DamageResult struct {
	FinalDamage    int                `json:"final_damage"`
	BaseDamage     int                `json:"base_damage"`
	DamageType     string             `json:"damage_type"`
	IsCritical     bool               `json:"is_critical"`
	IsBlocked      bool               `json:"is_blocked"`
	IsMiss         bool               `json:"is_miss"`
	ArmorReduction int                `json:"armor_reduction"`
	Modifiers      map[string]float64 `json:"modifiers"`
	Elements       map[string]int     `json:"elements,omitempty"`
	Breakdown      []DamageComponent  `json:"breakdown"`
}

// HealingResult représente le résultat d'un calcul de soins
type HealingResult struct {
	FinalHealing int                `json:"final_healing"`
	BaseHealing  int                `json:"base_healing"`
	IsCritical   bool               `json:"is_critical"`
	Modifiers    map[string]float64 `json:"modifiers"`
	Overheal     int                `json:"overheal"`
}

// DamageComponent représente un composant de dégâts
type DamageComponent struct {
	Type       string  `json:"type"`
	Value      int     `json:"value"`
	Source     string  `json:"source"`
	Multiplier float64 `json:"multiplier,omitempty"`
}

// NewDamageCalculator crée un nouveau calculateur de dégâts
func NewDamageCalculator(config *config.Config) DamageCalculatorInterface {
	return &DamageCalculator{
		config: config,
	}
}

// CalculateDamage calcule les dégâts totaux d'une attaque
func (dc *DamageCalculator) CalculateDamage(
	attacker, defender *models.CombatParticipant,
	skill *models.SkillInfo,
	modifiers map[string]float64,
) *DamageResult {
	result := &DamageResult{
		Modifiers: make(map[string]float64),
		Elements:  make(map[string]int),
		Breakdown: []DamageComponent{},
	}

	// Déterminer le type de dégâts et les dégâts de base
	var baseDamage int
	var damageType string

	if skill != nil {
		baseDamage = skill.BaseDamage
		damageType = skill.Type
		result.DamageType = damageType
	} else {
		// Attaque de base - dégâts physiques
		baseDamage = attacker.PhysicalDamage
		damageType = DamageTypePhysical
		result.DamageType = "physical"
	}

	result.BaseDamage = baseDamage

	// Vérifier si l'attaque touche
	hitChance := dc.CalculateHitChance(attacker, defender, skill, modifiers)
	if utils.SecureRandFloat64() > hitChance {
		result.IsMiss = true
		result.FinalDamage = 0
		return result
	}

	// Vérifier si l'attaque est bloquée
	blockChance := dc.CalculateBlockChance(defender, modifiers)
	if utils.SecureRandFloat64() < blockChance {
		result.IsBlocked = true
		result.FinalDamage = int(float64(baseDamage) * config.DefaultDamageReduction)
		return result
	}

	// Calculer les dégâts selon le type
	var rawDamage int
	switch damageType {
	case "physical":
		rawDamage = dc.CalculatePhysicalDamage(attacker, defender, baseDamage, modifiers)
	case DamageTypeMagical:
		rawDamage = dc.CalculateMagicalDamage(attacker, defender, baseDamage, modifiers)
	case "true":
		rawDamage = dc.CalculateTrueDamage(baseDamage, modifiers)
	default:
		rawDamage = dc.CalculatePhysicalDamage(attacker, defender, baseDamage, modifiers)
	}

	result.Breakdown = append(result.Breakdown, DamageComponent{
		Type:   damageType,
		Value:  rawDamage,
		Source: "base_calculation",
	})

	// Vérifier les critiques
	critChance := dc.CalculateCriticalChance(attacker, skill, modifiers)
	if utils.SecureRandFloat64() < critChance {
		result.IsCritical = true
		critMultiplier := 1.5 // Multiplicateur de base

		// Vérifier si la compétence a un multiplicateur spécial
		if skill != nil {
			if customCrit, exists := skill.Modifiers["critical_multiplier"]; exists {
				critMultiplier = customCrit
			}
		}

		rawDamage = int(float64(rawDamage) * critMultiplier)
		result.Breakdown = append(result.Breakdown, DamageComponent{
			Type:       "critical",
			Value:      rawDamage - result.Breakdown[len(result.Breakdown)-1].Value,
			Source:     "critical_hit",
			Multiplier: critMultiplier,
		})
	}

	// Appliquer les modificateurs d'effets
	if buff, exists := modifiers["damage_buff"]; exists {
		buffDamage := int(float64(rawDamage) * buff)
		rawDamage += buffDamage
		result.Breakdown = append(result.Breakdown, DamageComponent{
			Type:   "buff",
			Value:  buffDamage,
			Source: "damage_buff",
		})
	}

	// Appliquer la variance (±10%)
	variance := varianceMin + (utils.SecureRandFloat64() * varianceMax)
	rawDamage = int(float64(rawDamage) * variance)

	result.FinalDamage = rawDamage

	// S'assurer que les dégâts ne sont jamais négatifs
	if result.FinalDamage < 0 {
		result.FinalDamage = 0
	}

	return result
}

// CalculateHealing calcule les soins
func (dc *DamageCalculator) CalculateHealing(
	caster, target *models.CombatParticipant,
	skill *models.SkillInfo,
	modifiers map[string]float64,
) *HealingResult {
	result := &HealingResult{
		Modifiers: make(map[string]float64),
	}

	if skill == nil || skill.BaseHealing <= 0 {
		return result
	}

	baseHealing := skill.BaseHealing
	result.BaseHealing = baseHealing

	// Appliquer les modificateurs du lanceur
	healing := float64(baseHealing)

	if skill.Type == "magical" {
		// Les soins magiques sont améliorés par la puissance magique
		magicalBonus := float64(caster.MagicalDamage) * config.DefaultHealingVarianceRange
		healing += magicalBonus
	}

	// Vérifier les critiques pour les soins
	if skill.Type == "magical" {
		critChance := dc.CalculateCriticalChance(caster, skill, modifiers)
		if utils.SecureRandFloat64() < critChance {
			result.IsCritical = true
			healing *= 1.3 // Les soins critiques sont moins puissants que les dégâts critiques
		}
	}

	// Appliquer les modificateurs d'effets
	if healingBuff, exists := modifiers["healing_buff"]; exists {
		healing *= (1.0 + healingBuff)
	}

	// Appliquer la variance (±10%)
	variance := varianceMin + (utils.SecureRandFloat64() * varianceMax)
	healing *= variance

	result.FinalHealing = int(healing)

	// Calculer l'overheal
	if target.Health+result.FinalHealing > target.MaxHealth {
		result.Overheal = (target.Health + result.FinalHealing) - target.MaxHealth
		result.FinalHealing = target.MaxHealth - target.Health
	}

	return result
}

// CalculatePhysicalDamage calcule les dégâts physiques
func (dc *DamageCalculator) CalculatePhysicalDamage(
	attacker, defender *models.CombatParticipant,
	baseDamage int,
	modifiers map[string]float64,
) int {
	// Ajouter la puissance d'attaque physique
	damage := float64(baseDamage) + (float64(attacker.PhysicalDamage) * config.DefaultElementalPowerFire)

	// Appliquer la réduction d'armure
	armorReduction := dc.ApplyArmorReduction(int(damage), defender.PhysicalDefense, "physical")
	damage = float64(armorReduction)

	// Appliquer les modificateurs spécifiques
	if physicalMod, exists := modifiers["physical_damage_modifier"]; exists {
		damage *= (1.0 + physicalMod)
	}

	return int(damage)
}

// CalculateMagicalDamage calcule les dégâts magiques
func (dc *DamageCalculator) CalculateMagicalDamage(
	attacker, defender *models.CombatParticipant,
	baseDamage int,
	modifiers map[string]float64,
) int {
	// Ajouter la puissance magique
	damage := float64(baseDamage) + (float64(attacker.MagicalDamage) * config.DefaultElementalPowerFire)

	// Appliquer la résistance magique
	magicReduction := dc.ApplyArmorReduction(int(damage), defender.MagicalDefense, "magical")
	damage = float64(magicReduction)

	// Appliquer les modificateurs spécifiques
	if magicalMod, exists := modifiers["magical_damage_modifier"]; exists {
		damage *= (1.0 + magicalMod)
	}

	return int(damage)
}

// CalculateTrueDamage calcule les dégâts purs (ignorent les défenses)
func (dc *DamageCalculator) CalculateTrueDamage(baseDamage int, modifiers map[string]float64) int {
	damage := float64(baseDamage)

	// Les dégâts purs ne sont affectés que par des modificateurs spéciaux
	if trueMod, exists := modifiers["true_damage_modifier"]; exists {
		damage *= (1.0 + trueMod)
	}

	return int(damage)
}

// CalculateCriticalChance calcule la chance de critique
func (dc *DamageCalculator) CalculateCriticalChance(
	attacker *models.CombatParticipant,
	skill *models.SkillInfo,
	modifiers map[string]float64,
) float64 {
	baseCritChance := attacker.CriticalChance

	// Bonus de la compétence
	if skill != nil {
		if critBonus, exists := skill.Modifiers["critical_chance_bonus"]; exists {
			baseCritChance += critBonus
		}
	}

	// Modificateurs d'effets
	if critMod, exists := modifiers["critical_chance_modifier"]; exists {
		baseCritChance += critMod
	}

	// Limiter entre 0% et 95%
	if baseCritChance < 0 {
		baseCritChance = 0
	}
	if baseCritChance > config.DefaultMaxCriticalChance {
		baseCritChance = config.DefaultMaxCriticalChance
	}

	return baseCritChance
}

// CalculateHitChance calcule la chance de toucher
func (dc *DamageCalculator) CalculateHitChance(
	attacker, defender *models.CombatParticipant,
	skill *models.SkillInfo,
	modifiers map[string]float64,
) float64 {
	// Chance de base selon le type d'attaque
	baseHitChance := 0.85

	if skill != nil {
		// Certaines compétences ont des chances de toucher différentes
		if hitBonus, exists := skill.Modifiers["hit_chance_bonus"]; exists {
			baseHitChance += hitBonus
		}
	}

	// Facteur d'agilité/précision
	attackerDamage := float64(attacker.PhysicalDamage - defender.PhysicalDefense)
	ratingRange := float64(config.DefaultRatingRange * config.DefaultMaxParticipantsPvP)
	agilityDiff := attackerDamage / ratingRange
	hitChance := baseHitChance + agilityDiff

	// Modificateurs d'effets
	if hitMod, exists := modifiers["hit_chance_modifier"]; exists {
		hitChance += hitMod
	}

	// Limiter entre 5% et 95%
	if hitChance < config.DefaultMinHitChance {
		hitChance = config.DefaultMinHitChance
	}
	if hitChance > config.DefaultMaxHitChance {
		hitChance = config.DefaultMaxHitChance
	}

	return hitChance
}

// CalculateBlockChance calcule la chance de bloquer
func (dc *DamageCalculator) CalculateBlockChance(defender *models.CombatParticipant, modifiers map[string]float64) float64 {
	// Chance de base très faible
	baseBlockChance := 0.05

	// Bonus basé sur la défense physique
	defenseBonus := float64(defender.PhysicalDefense) / float64(config.DefaultVarianceDivisor*config.DefaultVarianceDivisor)
	blockChance := baseBlockChance + defenseBonus

	// Modificateurs d'effets
	if blockMod, exists := modifiers["block_chance_modifier"]; exists {
		blockChance += blockMod
	}

	// Limiter entre 0% et 75%
	if blockChance < 0 {
		blockChance = 0
	}
	if blockChance > config.DefaultMaxBlockChance {
		blockChance = config.DefaultMaxBlockChance
	}

	return blockChance
}

// ApplyArmorReduction applique la réduction d'armure
func (dc *DamageCalculator) ApplyArmorReduction(damage, armor int, damageType string) int {
	if damage <= 0 {
		return 0
	}

	// Formule de réduction: damage * (100 / (100 + armor))
	reduction := float64(armor) / (float64(armor) + float64(config.DefaultVarianceDivisor))
	finalDamage := float64(damage) * (1.0 - reduction)

	// Les dégâts purs ignorent l'armure
	if damageType == "true" {
		return damage
	}

	// S'assurer qu'il reste au moins 10% des dégâts originaux
	minDamage := float64(damage) * config.DefaultAverageDamageMultiplier2
	if finalDamage < minDamage {
		finalDamage = minDamage
	}

	return int(finalDamage)
}

// ApplyResistances applique les résistances élémentaires
func (dc *DamageCalculator) ApplyResistances(damage int, resistances map[string]float64, damageType string) int {
	if resistance, exists := resistances[damageType]; exists {
		// La résistance réduit les dégâts (0.0 = aucune résistance, 1.0 = immunité)
		finalDamage := float64(damage) * (1.0 - resistance)
		return int(finalDamage)
	}
	return damage
}

// ApplyVulnerabilities applique les vulnérabilités élémentaires
func (dc *DamageCalculator) ApplyVulnerabilities(damage int, vulnerabilities map[string]float64, damageType string) int {
	if vulnerability, exists := vulnerabilities[damageType]; exists {
		// La vulnérabilité augmente les dégâts
		finalDamage := float64(damage) * (1.0 + vulnerability)
		return int(finalDamage)
	}
	return damage
}

// CalculateElementalDamage calcule les dégâts élémentaires
func (dc *DamageCalculator) CalculateElementalDamage(attacker *models.CombatParticipant, element string, baseDamage int) int {
	elementalPower := dc.getElementalPower(attacker, element)

	// Les dégâts élémentaires sont basés sur la puissance magique et l'affinité élémentaire
	elementalDamage := float64(baseDamage) + (float64(attacker.MagicalDamage) * elementalPower)

	return int(elementalDamage)
}

// CalculateDamageOverTime calcule les dégâts sur la durée
func (dc *DamageCalculator) CalculateDamageOverTime(effect *models.CombatEffect, target *models.CombatParticipant) int {
	if effect.EffectType != models.EffectTypeDot {
		return 0
	}

	// Dégâts de base * nombre de stacks
	baseDamage := effect.ModifierValue * effect.CurrentStacks

	// Réduction basée sur la résistance magique (la plupart des DoTs sont magiques)
	finalDamage := dc.ApplyArmorReduction(baseDamage, target.MagicalDefense/config.DefaultArmorDivisor, "magical")

	return finalDamage
}

// CalculateStatusEffectChance calcule la chance d'appliquer un effet de statut
func (dc *DamageCalculator) CalculateStatusEffectChance(
	caster *models.CombatParticipant,
	target *models.CombatParticipant,
	effect *models.SkillEffect,
) float64 {
	baseChance := effect.Probability

	// Modifier selon la différence de puissance magique
	powerDiff := float64(caster.MagicalDamage-target.MagicalDefense) / float64(config.DefaultRatingRange*config.DefaultMaxParticipantsPvP)
	finalChance := baseChance + powerDiff

	// Limiter entre 5% et 95%
	if finalChance < config.DefaultMinFinalChance {
		finalChance = config.DefaultMinFinalChance
	}
	if finalChance > config.DefaultMaxFinalChance {
		finalChance = config.DefaultMaxFinalChance
	}

	return finalChance
}

// Helper methods

func (dc *DamageCalculator) getElementalPower(_ *models.CombatParticipant, element string) float64 {
	// Pour l'instant, retourner une valeur par défaut
	// Dans un vrai jeu, cela serait basé sur l'équipement et les stats du personnage
	elementalPowers := map[string]float64{
		"fire":      config.DefaultElementalPowerFire,
		"ice":       config.DefaultElementalPowerIce,
		"lightning": config.DefaultElementalPowerLightning,
		"earth":     config.DefaultElementalPowerEarth,
		"wind":      config.DefaultElementalPowerWind,
		"dark":      config.DefaultElementalPowerDark,
		"light":     config.DefaultElementalPowerLight,
	}

	if power, exists := elementalPowers[element]; exists {
		return power
	}

	return config.DefaultFleeChanceBase // Valeur par défaut
}

// CalculateLifesteal calcule le vol de vie
func (dc *DamageCalculator) CalculateLifesteal(attacker *models.CombatParticipant, damageDealt int, lifestealPercent float64) int {
	if lifestealPercent <= 0 || damageDealt <= 0 {
		return 0
	}

	lifestealAmount := float64(damageDealt) * lifestealPercent

	// Limiter le vol de vie (ne peut pas dépasser la vie manquante)
	missingHealth := attacker.MaxHealth - attacker.Health
	if int(lifestealAmount) > missingHealth {
		lifestealAmount = float64(missingHealth)
	}

	return int(lifestealAmount)
}

// CalculateManaBurn calcule la destruction de mana
func (dc *DamageCalculator) CalculateManaBurn(target *models.CombatParticipant, burnAmount int) int {
	if target.Mana <= 0 || burnAmount <= 0 {
		return 0
	}

	actualBurn := burnAmount
	if actualBurn > target.Mana {
		actualBurn = target.Mana
	}

	return actualBurn
}

// CalculateKnockback calcule la force de projection
func (dc *DamageCalculator) CalculateKnockback(attacker, target *models.CombatParticipant, baseKnockback float64) float64 {
	// La force de projection dépend de la différence de "poids" ou de résistance
	massRatio := float64(attacker.PhysicalDamage) / float64(target.PhysicalDefense+config.DefaultMassRatioDivisor)

	knockback := baseKnockback * massRatio

	// Limiter le knockback
	if knockback > config.DefaultMaxKnockback {
		knockback = config.DefaultMaxKnockback
	}
	if knockback < config.DefaultMinKnockback {
		knockback = config.DefaultMinKnockback
	}

	return knockback
}

// CalculateComboMultiplier calcule le multiplicateur de combo
func (dc *DamageCalculator) CalculateComboMultiplier(comboCount int) float64 {
	// Multiplicateur qui augmente avec le combo mais avec des rendements décroissants
	if comboCount <= 0 {
		return 1.0
	}

	// Formule: 1 + (combo * 0.1) / (1 + combo * 0.05)
	comboFloat := float64(comboCount)
	numerator := comboFloat * config.DefaultAverageDamageMultiplier2
	denominator := 1.0 + comboFloat*config.DefaultAverageDamageMultiplier2/config.DefaultMaxParticipantsPvP
	multiplier := 1.0 + numerator/denominator

	// Limiter le multiplicateur
	if multiplier > config.DefaultMaxMultiplier {
		multiplier = config.DefaultMaxMultiplier
	}

	return multiplier
}

// CalculateExperienceGain calcule l'expérience gagnée
func (dc *DamageCalculator) CalculateExperienceGain(victorLevel, defeatedLevel int, damageContribution float64) int {
	// Expérience de base selon le niveau de l'ennemi
	baseExp := defeatedLevel * config.DefaultBaseExperience

	// Modifier selon la différence de niveau
	levelDiff := defeatedLevel - victorLevel
	levelMultiplier := 1.0 + (float64(levelDiff) * config.DefaultAverageDamageMultiplier2)

	// Appliquer la contribution aux dégâts
	finalExp := float64(baseExp) * levelMultiplier * damageContribution

	// S'assurer d'un minimum d'expérience
	if finalExp < 1 {
		finalExp = 1
	}

	return int(finalExp)
}

// CalculateThreat calcule le niveau de menace généré
func (dc *DamageCalculator) CalculateThreat(action *models.CombatAction, participant *models.CombatParticipant) int {
	threat := 0

	switch action.ActionType {
	case models.ActionTypeAttack:
		// Les attaques génèrent de la menace basée sur les dégâts
		threat = action.DamageDealt

	case models.ActionTypeSkill:
		// Les compétences offensives génèrent plus de menace
		threat = action.DamageDealt
		if action.DamageDealt > 0 {
			threat = int(float64(action.DamageDealt) * config.DefaultThreatMultiplier)
		}

	case models.ActionTypeItem:
		// Les objets de soin génèrent de la menace
		if action.HealingDone > 0 {
			threat = action.HealingDone / config.DefaultMaxParticipantsPvP
		}

	case models.ActionTypeDefend:
		// La défense génère peu de menace
		threat = 1

	default:
		threat = 0
	}

	// Modifier selon le rôle/classe du personnage
	// TODO: Intégrer avec le système de classes

	return threat
}

// AdvancedDamageResult représente un résultat de calcul avancé
type AdvancedDamageResult struct {
	*DamageResult
	Lifesteal       int     `json:"lifesteal,omitempty"`
	ManaBurn        int     `json:"mana_burn,omitempty"`
	Knockback       float64 `json:"knockback,omitempty"`
	ComboMultiplier float64 `json:"combo_multiplier,omitempty"`
	ThreatGenerated int     `json:"threat_generated,omitempty"`
}

// CalculateAdvancedDamage effectue un calcul de dégâts avec effets avancés
func (dc *DamageCalculator) CalculateAdvancedDamage(
	attacker, defender *models.CombatParticipant,
	skill *models.SkillInfo,
	modifiers map[string]float64,
	comboCount int,
) *AdvancedDamageResult {
	// Calcul de base
	baseResult := dc.CalculateDamage(attacker, defender, skill, modifiers)

	result := &AdvancedDamageResult{
		DamageResult: baseResult,
	}

	// Effets avancés seulement si l'attaque touche
	if !result.IsMiss && !result.IsBlocked && result.FinalDamage > 0 {
		// Multiplicateur de combo
		if comboCount > 0 {
			result.ComboMultiplier = dc.CalculateComboMultiplier(comboCount)
			result.FinalDamage = int(float64(result.FinalDamage) * result.ComboMultiplier)
		}

		// Vol de vie
		if lifestealPercent, exists := modifiers["lifesteal"]; exists {
			result.Lifesteal = dc.CalculateLifesteal(attacker, result.FinalDamage, lifestealPercent)
		}

		// Destruction de mana
		if manaBurnAmount, exists := modifiers["mana_burn"]; exists {
			result.ManaBurn = dc.CalculateManaBurn(defender, int(manaBurnAmount))
		}

		// Projection
		if knockbackForce, exists := modifiers["knockback"]; exists {
			result.Knockback = dc.CalculateKnockback(attacker, defender, knockbackForce)
		}
	}

	return result
}

// GetDamageBreakdown retourne une analyze détaillée des dégâts
func (dc *DamageCalculator) GetDamageBreakdown(result *DamageResult) string {
	if len(result.Breakdown) == 0 {
		return "No damage breakdown available"
	}

	breakdown := fmt.Sprintf("Base damage: %d\n", result.BaseDamage)

	for _, component := range result.Breakdown {
		if component.Multiplier > 0 {
			breakdown += fmt.Sprintf("%s: %d (x%.2f)\n", component.Type, component.Value, component.Multiplier)
		} else {
			breakdown += fmt.Sprintf("%s: %d\n", component.Type, component.Value)
		}
	}

	breakdown += fmt.Sprintf("Final damage: %d", result.FinalDamage)

	return breakdown
}
