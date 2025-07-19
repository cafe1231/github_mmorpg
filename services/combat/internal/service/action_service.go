package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"combat/internal/config"
	"combat/internal/models"
	"combat/internal/repository"
)

// Constantes pour les actions anti-cheat
const (
	AntiCheatActionWarn  = "warn"
	AntiCheatActionBlock = "block"
)

// ActionServiceInterface définit les méthodes du service d'actions
type ActionServiceInterface interface {
	// Exécution d'actions
	ExecuteAction(combat *models.CombatInstance, actor *models.CombatParticipant, req *models.ActionRequest) (*models.ActionResult, error)
	ValidateAction(combat *models.CombatInstance, actor *models.CombatParticipant,
		req *models.ValidateActionRequest) (*models.ValidationResponse, error)

	// Actions disponibles
	GetAvailableActions(combat *models.CombatInstance, actor *models.CombatParticipant) ([]*models.ActionTemplate, error)
	GetActionTemplates() []*models.ActionTemplate

	// Traitement des actions
	ProcessAction(action *models.CombatAction, combat *models.CombatInstance) (*models.ActionResult, error)
	CalculateActionResult(action *models.CombatAction, actor, target *models.CombatParticipant, skill *models.SkillInfo) error

	// Cooldowns et restrictions
	IsActionOnCooldown(actorID uuid.UUID, actionType models.ActionType, skillID string) (bool, time.Duration, error)
	SetActionCooldown(actorID uuid.UUID, actionType models.ActionType, skillID string, duration time.Duration) error

	// Statistiques
	GetActionStatistics(actorID uuid.UUID, timeWindow time.Duration) (*models.ActionStatistics, error)
}

// ActionService implémente l'interface ActionServiceInterface
type ActionService struct {
	actionRepo repository.ActionRepositoryInterface
	combatRepo repository.CombatRepositoryInterface
	effectRepo repository.EffectRepositoryInterface
	damageCalc DamageCalculatorInterface
	config     *config.Config
	cooldowns  map[string]*models.ActionCooldown // Cache des cooldowns
}

// NewActionService crée un nouveau service d'actions
func NewActionService(
	actionRepo repository.ActionRepositoryInterface,
	combatRepo repository.CombatRepositoryInterface,
	effectRepo repository.EffectRepositoryInterface,
	damageCalc DamageCalculatorInterface,
	config *config.Config,
) ActionServiceInterface {
	return &ActionService{
		actionRepo: actionRepo,
		combatRepo: combatRepo,
		effectRepo: effectRepo,
		damageCalc: damageCalc,
		config:     config,
		cooldowns:  make(map[string]*models.ActionCooldown),
	}
}

// ExecuteAction exécute une action de combat
func (s *ActionService) ExecuteAction(combat *models.CombatInstance, actor *models.CombatParticipant,
	req *models.ActionRequest,
) (*models.ActionResult, error) {
	startTime := time.Now()

	// Créer l'action
	action := models.CreateAction(combat.ID, actor.CharacterID, req)
	action.TurnNumber = combat.CurrentTurn
	action.ServerTimestamp = time.Now()

	// Déterminer l'ordre d'action (basé sur la vitesse d'attaque)
	action.ActionOrder = s.calculateActionOrder(actor)

	result := &models.ActionResult{
		Success: true,
		Action:  action,
		StateChanges: &models.StateChanges{
			ParticipantChanges: make(map[uuid.UUID]*models.ParticipantChange),
		},
		Logs: []*models.CombatLog{},
	}

	// Traiter l'action selon son type
	var err error
	switch req.ActionType {
	case models.ActionTypeAttack:
		err = s.executeAttack(action, combat, actor, result)
	case models.ActionTypeSkill:
		err = s.executeSkill(action, combat, actor, result)
	case models.ActionTypeItem:
		err = s.executeItem(action, combat, actor, result)
	case models.ActionTypeDefend:
		err = s.executeDefend(action, combat, actor, result)
	case models.ActionTypeFlee:
		err = s.executeFlee(action, combat, actor, result)
	case models.ActionTypeWait:
		err = s.executeWait(action, combat, actor, result)
	default:
		err = fmt.Errorf("unknown action type: %s", req.ActionType)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		action.IsValidated = false
		errMsg := err.Error()
		action.ValidationNotes = &errMsg
	}

	// Calculer le temps de traitement
	processingTime := int(time.Since(startTime).Milliseconds())
	action.ProcessingTimeMs = &processingTime

	// Sauvegarder l'action
	if saveErr := s.actionRepo.Create(action); saveErr != nil {
		logrus.WithError(saveErr).Error("Failed to save action")
	}

	// Mettre à jour les participants affectés
	if result.Success {
		for participantID, change := range result.StateChanges.ParticipantChanges {
			if err := s.applyParticipantChanges(participantID, change); err != nil {
				logrus.WithError(err).WithField("participant_id", participantID).Error("Failed to apply participant changes")
			}
		}
	}

	// Ajouter un log de l'action
	logEntry := &models.CombatLog{
		ID:         uuid.New(),
		CombatID:   combat.ID,
		LogType:    "action",
		ActorName:  actor.Character.Name, // TODO: Récupérer le nom
		Message:    action.GetDescription(),
		TurnNumber: &action.TurnNumber,
		Timestamp:  time.Now(),
	}
	result.Logs = append(result.Logs, logEntry)

	return result, nil
}

// executeAttack exécute une attaque de base
func (s *ActionService) executeAttack(action *models.CombatAction, combat *models.CombatInstance,
	actor *models.CombatParticipant, result *models.ActionResult,
) error {
	if action.TargetID == nil {
		return fmt.Errorf("target required for attack")
	}

	// Récupérer la cible
	target, err := s.combatRepo.GetParticipant(combat.ID, *action.TargetID)
	if err != nil {
		return fmt.Errorf("target not found: %w", err)
	}

	if !target.IsAlive {
		return fmt.Errorf("cannot attack dead target")
	}

	// Vérifier si c'est un allié (selon les règles du combat)
	if !combat.Settings.TeamDamage && actor.Team == target.Team {
		return fmt.Errorf("cannot attack ally")
	}

	// Calculer la chance de toucher
	hitChance := models.CalculateHitChance(actor, target, nil)
	hit := rand.Float64() < hitChance

	if !hit {
		action.IsMiss = true
		result.Logs = append(result.Logs, &models.CombatLog{
			Message: fmt.Sprintf("%s rate son attaque sur %s", actor.Character.Name, target.Character.Name),
		})
		return nil
	}

	// Calculer la chance de critique
	critChance := models.CalculateCriticalChance(actor, nil)
	action.IsCritical = rand.Float64() < critChance

	// Calculer les dégâts
	damage := action.CalculateDamage(actor, target, nil)
	action.DamageDealt = damage

	// Appliquer les dégâts
	s.applyDamage(target, damage, result)

	// Mettre à jour les statistiques de l'acteur
	change := result.StateChanges.ParticipantChanges[actor.CharacterID]
	if change == nil {
		change = &models.ParticipantChange{}
		result.StateChanges.ParticipantChanges[actor.CharacterID] = change
	}

	// Gestion de l'énergie/mana pour l'attaque
	if actor.Mana > 0 {
		change.ManaChange = -1 // Coût minimal pour une attaque
	}

	return nil
}

// executeSkill exécute une compétence
func (s *ActionService) executeSkill(action *models.CombatAction, combat *models.CombatInstance, actor *models.CombatParticipant, result *models.ActionResult) error {
	if action.SkillID == nil {
		return fmt.Errorf("skill ID required")
	}

	// Récupérer les informations de la compétence
	skillTemplates := models.GetSkillTemplates()
	skill, exists := skillTemplates[*action.SkillID]
	if !exists {
		return fmt.Errorf("unknown skill: %s", *action.SkillID)
	}

	// Vérifier les prérequis
	if err := s.validateSkillRequirements(actor, skill); err != nil {
		return err
	}

	// Vérifier le coût en mana
	if actor.Mana < skill.ManaCost {
		return fmt.Errorf("insufficient mana: %d/%d", actor.Mana, skill.ManaCost)
	}

	// Vérifier le cooldown
	if onCooldown, remaining, _ := s.IsActionOnCooldown(actor.CharacterID, models.ActionTypeSkill, *action.SkillID); onCooldown {
		return fmt.Errorf("skill on cooldown for %v", remaining)
	}

	// Déterminer la cible
	var target *models.CombatParticipant
	if skill.TargetType != "self" && action.TargetID != nil {
		var err error
		target, err = s.combatRepo.GetParticipant(combat.ID, *action.TargetID)
		if err != nil {
			return fmt.Errorf("target not found: %w", err)
		}

		if !target.IsAlive && skill.TargetType != "dead" {
			return fmt.Errorf("cannot target dead participant")
		}
	} else {
		target = actor // Auto-ciblage
	}

	// Calculer la chance de toucher
	hitChance := models.CalculateHitChance(actor, target, skill)
	hit := rand.Float64() < hitChance

	if !hit {
		action.IsMiss = true
		action.ManaUsed = skill.ManaCost / config.DefaultArmorDivisor // Coût réduit en cas d'échec
		return nil
	}

	// Calculer la chance de critique
	critChance := models.CalculateCriticalChance(actor, skill)
	action.IsCritical = rand.Float64() < critChance

	// Consommer la mana
	action.ManaUsed = skill.ManaCost

	// Calculer et appliquer les effets
	if skill.BaseDamage > 0 {
		damage := action.CalculateDamage(actor, target, skill)
		action.DamageDealt = damage
		s.applyDamage(target, damage, result)
	}

	if skill.BaseHealing > 0 {
		healing := action.CalculateHealing(actor, skill)
		action.HealingDone = healing
		s.applyHealing(target, healing, result)
	}

	// Appliquer les effets de la compétence
	for _, effect := range skill.Effects {
		if rand.Float64() < effect.Probability {
			s.applySkillEffect(actor, target, effect, result)
		}
	}

	// Définir le cooldown
	if skill.Cooldown > 0 {
		if err := s.SetActionCooldown(actor.CharacterID, models.ActionTypeSkill, *action.SkillID,
			time.Duration(skill.Cooldown)*time.Second); err != nil {
			logrus.WithError(err).Error("Failed to set action cooldown")
		}
	}

	// Mettre à jour la mana de l'acteur
	change := result.StateChanges.ParticipantChanges[actor.CharacterID]
	if change == nil {
		change = &models.ParticipantChange{}
		result.StateChanges.ParticipantChanges[actor.CharacterID] = change
	}
	change.ManaChange = -action.ManaUsed

	return nil
}

// executeItem utilize un objet
func (s *ActionService) executeItem(action *models.CombatAction, combat *models.CombatInstance,
	actor *models.CombatParticipant, result *models.ActionResult,
) error {
	if action.ItemID == nil {
		return fmt.Errorf("item ID required")
	}

	// TODO: Récupérer les informations de l'objet depuis le service inventory
	// Pour l'instant, simuler un objet de soin simple
	if *action.ItemID == "health_potion" {
		healing := 30
		action.HealingDone = healing
		s.applyHealing(actor, healing, result)

		// Log d'utilisation
		result.Logs = append(result.Logs, &models.CombatLog{
			Message: fmt.Sprintf("%s utilize une potion de soin (+%d PV)", actor.Character.Name, healing),
		})
	} else {
		return fmt.Errorf("unknown item: %s", *action.ItemID)
	}

	return nil
}

// executeDefend exécute une action de défense
func (s *ActionService) executeDefend(action *models.CombatAction, combat *models.CombatInstance,
	actor *models.CombatParticipant, result *models.ActionResult,
) error {
	// Appliquer un effet de défense temporaire
	defenseEffect := &models.EffectApplication{
		EffectTemplate: &models.EffectTemplate{
			ID:            "defense_bonus",
			Name:          "Défense renforcée",
			Description:   "Réduit les dégâts reçus de 50%",
			EffectType:    models.EffectTypeBuff,
			StatAffected:  "damage_reduction",
			ModifierValue: config.DefaultImprovementScore,
			ModifierType:  models.ModifierTypePercentage,
			BaseDuration:  1, // Un tour
			MaxStacks:     1,
			IsDispellable: false,
			IsBeneficial:  true,
		},
		TargetID: actor.CharacterID,
		Duration: 1,
	}

	// TODO: Appliquer l'effet via le service d'effets
	_ = defenseEffect

	result.Logs = append(result.Logs, &models.CombatLog{
		Message: fmt.Sprintf("%s se défend", actor.Character.Name),
	})

	return nil
}

// executeFlee tente de fuir le combat
func (s *ActionService) executeFlee(action *models.CombatAction, combat *models.CombatInstance,
	actor *models.CombatParticipant, result *models.ActionResult,
) error {
	if !combat.Settings.AllowFlee {
		return fmt.Errorf("fleeing not allowed in this combat")
	}

	// Chance de réussite de la fuite (basée sur l'agilité)
	fleeChance := config.DefaultFleeChanceBase + (float64(actor.PhysicalDamage) / config.DefaultFleeChanceDivisor) // Formule simple
	// Limiter la chance de fuite à 90% maximum
	if fleeChance > config.DefaultMaxFleeChance {
		fleeChance = config.DefaultMaxFleeChance
	}

	success := rand.Float64() < fleeChance

	if success {
		// Retirer le participant du combat
		result.StateChanges.ParticipantChanges[actor.CharacterID] = &models.ParticipantChange{
			StatusChange: "fled",
		}

		result.Logs = append(result.Logs, &models.CombatLog{
			Message: fmt.Sprintf("%s fuit le combat avec succès", actor.Character.Name),
		})

		// TODO: Marquer le participant comme ayant fui
	} else {
		result.Logs = append(result.Logs, &models.CombatLog{
			Message: fmt.Sprintf("%s échoue à fuir", actor.Character.Name),
		})
	}

	return nil
}

// executeWait attend et récupère de la mana
func (s *ActionService) executeWait(action *models.CombatAction, combat *models.CombatInstance,
	actor *models.CombatParticipant, result *models.ActionResult,
) error {
	// Récupérer un pourcentage de mana
	// Récupération de mana (10% de la mana max)
	manaRecovery := int(float64(actor.MaxMana) * config.DefaultManaRecoveryPercent)
	if manaRecovery < 1 {
		manaRecovery = 1
	}

	action.HealingDone = 0          // Pas de soins
	action.ManaUsed = -manaRecovery // Récupération de mana

	change := &models.ParticipantChange{
		ManaChange: manaRecovery,
	}
	result.StateChanges.ParticipantChanges[actor.CharacterID] = change

	result.Logs = append(result.Logs, &models.CombatLog{
		Message: fmt.Sprintf("%s attend et récupère %d points de mana", actor.Character.Name, manaRecovery),
	})

	return nil
}

// Helper methods

func (s *ActionService) calculateActionOrder(actor *models.CombatParticipant) int {
	// Ordre basé sur la vitesse d'attaque (plus élevé = agit en premier)
	baseOrder := 100
	// Bonus de vitesse d'attaque
	speedBonus := int(actor.AttackSpeed * config.DefaultSpeedBonusMultiplier)
	return baseOrder + speedBonus + rand.Intn(config.DefaultRandomFactor) // Ajout d'un facteur aléatoire
}

func (s *ActionService) applyDamage(target *models.CombatParticipant, damage int, result *models.ActionResult) {
	if damage <= 0 {
		return
	}

	// Récupérer ou créer le changement pour la cible
	change := result.StateChanges.ParticipantChanges[target.CharacterID]
	if change == nil {
		change = &models.ParticipantChange{}
		result.StateChanges.ParticipantChanges[target.CharacterID] = change
	}

	// Appliquer les dégâts
	change.HealthChange -= damage

	// Vérifier si la cible meurt
	newHealth := target.Health + change.HealthChange
	if newHealth <= 0 {
		change.HealthChange = -target.Health // Assurer que la vie ne descend pas en dessous de 0
		change.StatusChange = "dead"

		result.Logs = append(result.Logs, &models.CombatLog{
			LogType: "death",
			Message: fmt.Sprintf("%s est vaincu", target.Character.Name),
		})
	}
}

func (s *ActionService) applyHealing(target *models.CombatParticipant, healing int, result *models.ActionResult) {
	if healing <= 0 {
		return
	}

	// Récupérer ou créer le changement pour la cible
	change := result.StateChanges.ParticipantChanges[target.CharacterID]
	if change == nil {
		change = &models.ParticipantChange{}
		result.StateChanges.ParticipantChanges[target.CharacterID] = change
	}

	// Appliquer les soins (ne peut pas dépasser la vie maximale)
	maxHealing := target.MaxHealth - target.Health
	if healing > maxHealing {
		healing = maxHealing
	}

	change.HealthChange += healing

	if healing > 0 {
		result.Logs = append(result.Logs, &models.CombatLog{
			LogType: "healing",
			Message: fmt.Sprintf("%s récupère %d points de vie", target.Character.Name, healing),
		})
	}
}

func (s *ActionService) applySkillEffect(caster, target *models.CombatParticipant, effect models.SkillEffect, result *models.ActionResult) {
	// Créer un effet de combat basé sur l'effet de compétence
	effectApp := &models.EffectApplication{
		EffectTemplate: &models.EffectTemplate{
			ID:            fmt.Sprintf("skill_effect_%s", effect.Type),
			Name:          effect.Type,
			Description:   fmt.Sprintf("Effet de compétence: %s", effect.Type),
			EffectType:    s.mapSkillEffectType(effect.Type),
			StatAffected:  effect.StatAffected,
			ModifierValue: effect.Value,
			ModifierType:  models.ModifierTypeFlat,
			BaseDuration:  effect.Duration,
			MaxStacks:     1,
			IsDispellable: true,
			IsBeneficial:  s.isEffectBeneficial(effect.Type),
		},
		TargetID: target.CharacterID,
		CasterID: &caster.CharacterID,
		Duration: effect.Duration,
	}

	// Récupérer ou créer le changement pour la cible
	change := result.StateChanges.ParticipantChanges[target.CharacterID]
	if change == nil {
		change = &models.ParticipantChange{}
		result.StateChanges.ParticipantChanges[target.CharacterID] = change
	}

	// Créer l'effet de combat
	combatEffect := models.CreateEffectFromTemplate(effectApp.EffectTemplate, effectApp)
	combatEffect.CombatID = caster.CombatID

	// Ajouter l'effet aux changements
	change.EffectsAdded = append(change.EffectsAdded, combatEffect)

	result.Logs = append(result.Logs, &models.CombatLog{
		LogType: "effect",
		Message: fmt.Sprintf("%s applique %s sur %s", caster.Character.Name, combatEffect.EffectName, target.Character.Name),
	})
}

func (s *ActionService) mapSkillEffectType(skillEffectType string) models.EffectType {
	switch skillEffectType {
	case "damage_over_time":
		return models.EffectTypeDot
	case "heal_over_time":
		return models.EffectTypeHot
	case "stun":
		return models.EffectTypeStun
	case "silence":
		return models.EffectTypeSilence
	case "buff":
		return models.EffectTypeBuff
	case "debuff":
		return models.EffectTypeDebuff
	default:
		return models.EffectTypeBuff
	}
}

func (s *ActionService) isEffectBeneficial(effectType string) bool {
	beneficialEffects := []string{"heal_over_time", "buff", "shield"}
	for _, beneficial := range beneficialEffects {
		if effectType == beneficial {
			return true
		}
	}
	return false
}

func (s *ActionService) validateSkillRequirements(actor *models.CombatParticipant, skill *models.SkillInfo) error {
	for requirement, value := range skill.Requirements {
		switch requirement {
		case "shield_equipped":
			// TODO: Vérifier si le joueur a un bouclier équipé
			// Pour l'instant, on assume que c'est OK
			_ = value
		case "weapon_type":
			// TODO: Vérifier le type d'arme
		default:
			logrus.WithField("requirement", requirement).Warn("Unknown skill requirement")
		}
	}
	return nil
}

func (s *ActionService) applyParticipantChanges(participantID uuid.UUID, change *models.ParticipantChange) error {
	// Récupérer le participant actuel
	// Note: Nous aurions besoin du combat ID ici, mais pour simplifier on va faire une recherche
	// Dans une vraie implémentation, il faudrait passer le combat ID ou restructurer

	// TODO: Implémenter la mise à jour réelle du participant
	// participant, err := s.combatRepo.GetParticipant(combatID, participantID)
	// if err != nil {
	//     return err
	// }

	// Appliquer les changements
	// participant.Health += change.HealthChange
	// participant.Mana += change.ManaChange

	// Vérifier les limites
	// if participant.Health < 0 {
	//     participant.Health = 0
	//     participant.IsAlive = false
	// }
	// if participant.Health > participant.MaxHealth {
	//     participant.Health = participant.MaxHealth
	// }
	// if participant.Mana < 0 {
	//     participant.Mana = 0
	// }
	// if participant.Mana > participant.MaxMana {
	//     participant.Mana = participant.MaxMana
	// }

	// return s.combatRepo.UpdateParticipant(participant)

	logrus.WithFields(logrus.Fields{
		"participant_id": participantID,
		"health_change":  change.HealthChange,
		"mana_change":    change.ManaChange,
	}).Debug("Applied participant changes")

	return nil
}

// ValidateAction valide une action sans l'exécuter
func (s *ActionService) ValidateAction(combat *models.CombatInstance, actor *models.CombatParticipant,
	req *models.ValidateActionRequest,
) (*models.ValidationResponse, error) {
	validation := req.Action.Validate()

	response := &models.ValidationResponse{
		Valid:    validation.IsValid,
		Errors:   validation.Errors,
		Warnings: validation.Warnings,
	}

	// Validations supplémentaires selon le contexte
	if req.CheckCooldowns {
		if req.Action.SkillID != nil {
			if onCooldown, remaining, _ := s.IsActionOnCooldown(actor.CharacterID, models.ActionTypeSkill, *req.Action.SkillID); onCooldown {
				response.Valid = false
				response.Errors = append(response.Errors, fmt.Sprintf("Skill on cooldown for %v", remaining))
			}
		}
	}

	if req.CheckResources {
		if req.Action.SkillID != nil {
			skillTemplates := models.GetSkillTemplates()
			if skill, exists := skillTemplates[*req.Action.SkillID]; exists {
				if actor.Mana < skill.ManaCost {
					response.Valid = false
					response.Errors = append(response.Errors, fmt.Sprintf("Insufficient mana: %d/%d", actor.Mana, skill.ManaCost))
				}
			}
		}
	}

	// Validation anti-cheat
	if req.Strict {
		antiCheatResult := &models.AntiCheatResult{
			Suspicious: false,
			Score:      0,
			Action:     "allow",
		}

		// Vérifier la fréquence d'actions
		recentActions, err := s.actionRepo.GetRecentActionsByActor(actor.CharacterID, 1*time.Minute)
		if err == nil && len(recentActions) > s.config.AntiCheat.MaxActionsPerSecond*60 {
			antiCheatResult.Suspicious = true
			antiCheatResult.Score += 30
			antiCheatResult.Flags = append(antiCheatResult.Flags, "high_action_frequency")
		}

		// Vérifier le timestamp
		if req.Action.ClientTimestamp.IsZero() {
			antiCheatResult.Score += 10
			antiCheatResult.Flags = append(antiCheatResult.Flags, "missing_timestamp")
		} else {
			timeDiff := time.Since(req.Action.ClientTimestamp).Abs()
			if timeDiff > 5*time.Second {
				antiCheatResult.Score += 20
				antiCheatResult.Flags = append(antiCheatResult.Flags, "suspicious_timestamp")
			}
		}

		if antiCheatResult.Score > config.DefaultMinScore2 {
			antiCheatResult.Action = AntiCheatActionWarn
		} else if antiCheatResult.Score > config.DefaultMinScore3 {
			antiCheatResult.Action = AntiCheatActionBlock
			response.Valid = false
		}

		response.AntiCheat = antiCheatResult
	}

	return response, nil
}

// GetAvailableActions récupère les actions disponibles pour un participant
func (s *ActionService) GetAvailableActions(combat *models.CombatInstance,
	actor *models.CombatParticipant,
) ([]*models.ActionTemplate, error) {
	templates := models.GetActionTemplates()
	available := []*models.ActionTemplate{}

	for _, template := range templates {
		// Vérifier la disponibilité selon le contexte
		isAvailable := true

		switch template.Type {
		case models.ActionTypeFlee:
			if !combat.Settings.AllowFlee {
				isAvailable = false
			}
		case models.ActionTypeItem:
			if !combat.Settings.AllowItems {
				isAvailable = false
			}
		}

		// Vérifier les ressources
		if template.ManaCost > actor.Mana {
			isAvailable = false
		}

		// Vérifier les cooldowns
		if template.Type == models.ActionTypeSkill {
			// Pour les compétences, on vérifierait le cooldown ici
			// Pour l'instant, on assume qu'elles sont disponibles
			logrus.Debug("Skill cooldown check not fully implemented")
		}

		template.Available = isAvailable
		available = append(available, template)
	}

	// Ajouter les compétences disponibles
	skillTemplates := models.GetSkillTemplates()
	for skillID, skill := range skillTemplates {
		// Vérifier si le joueur connaît cette compétence
		// TODO: Intégrer avec le système de compétences du joueur

		template := &models.ActionTemplate{
			Type:            models.ActionTypeSkill,
			Name:            skill.Name,
			Description:     skill.Description,
			Icon:            skill.Icon,
			RequiredTargets: 1,
			TargetType:      skill.TargetType,
			Range:           skill.Range,
			Cooldown:        skill.Cooldown,
			ManaCost:        skill.ManaCost,
			Available:       actor.Mana >= skill.ManaCost,
		}

		// Vérifier le cooldown
		if onCooldown, _, _ := s.IsActionOnCooldown(actor.CharacterID, models.ActionTypeSkill, skillID); onCooldown {
			template.Available = false
		}

		available = append(available, template)
	}

	return available, nil
}

// GetActionTemplates retourne tous les modèles d'actions
func (s *ActionService) GetActionTemplates() []*models.ActionTemplate {
	return models.GetActionTemplates()
}

// ProcessAction traite une action déjà créée
func (s *ActionService) ProcessAction(action *models.CombatAction, combat *models.CombatInstance) (*models.ActionResult, error) {
	// Cette méthode pourrait être utilisée pour retraiter des actions
	// ou pour des actions différées
	return &models.ActionResult{
		Success: true,
		Action:  action,
	}, nil
}

// CalculateActionResult calcule le résultat d'une action
func (s *ActionService) CalculateActionResult(action *models.CombatAction, actor, target *models.CombatParticipant,
	skill *models.SkillInfo,
) error {
	if action.ActionType == models.ActionTypeAttack || (action.ActionType == models.ActionTypeSkill && skill != nil && skill.BaseDamage > 0) {
		// Calculer les dégâts
		damage := action.CalculateDamage(actor, target, skill)
		action.DamageDealt = damage
	}

	if action.ActionType == models.ActionTypeSkill && skill != nil && skill.BaseHealing > 0 {
		// Calculer les soins
		healing := action.CalculateHealing(actor, skill)
		action.HealingDone = healing
	}

	return nil
}

// IsActionOnCooldown vérifie si une action est en cooldown
func (s *ActionService) IsActionOnCooldown(actorID uuid.UUID, actionType models.ActionType, skillID string) (bool, time.Duration, error) {
	key := fmt.Sprintf("%s_%s_%s", actorID.String(), actionType, skillID)

	cooldown, exists := s.cooldowns[key]
	if !exists {
		return false, 0, nil
	}

	if time.Now().After(cooldown.ExpiresAt) {
		// Cooldown expiré, le supprimer
		delete(s.cooldowns, key)
		return false, 0, nil
	}

	remaining := time.Until(cooldown.ExpiresAt)
	return true, remaining, nil
}

// SetActionCooldown définit un cooldown pour une action
func (s *ActionService) SetActionCooldown(actorID uuid.UUID, actionType models.ActionType, skillID string, duration time.Duration) error {
	key := fmt.Sprintf("%s_%s_%s", actorID.String(), actionType, skillID)

	cooldownKey := fmt.Sprintf("%s_%s_%s", actorID.String(), actionType, skillID)
	cooldown := &models.ActionCooldown{
		ID:         cooldownKey,
		ActorID:    actorID,
		ActionType: actionType,
		SkillID:    skillID,
		ExpiresAt:  time.Now().Add(duration),
		Duration:   duration,
	}

	s.cooldowns[key] = cooldown
	return nil
}

// GetActionStatistics récupère les statistiques d'actions
func (s *ActionService) GetActionStatistics(actorID uuid.UUID, timeWindow time.Duration) (*models.ActionStatistics, error) {
	// Déléguer au repository qui a déjà cette méthode
	actionStats, err := s.actionRepo.GetActionStatistics(actorID, timeWindow)
	if err != nil {
		return nil, err
	}

	// Convertir le type interne vers le type de modèle
	stats := &models.ActionStatistics{
		TotalActions:      actionStats.TotalActions,
		CriticalHits:      actionStats.CriticalHits,
		Misses:            actionStats.Misses,
		Blocks:            actionStats.Blocks,
		AvgDamage:         actionStats.AvgDamage,
		MaxDamage:         actionStats.MaxDamage,
		AvgHealing:        actionStats.AvgHealing,
		AvgProcessingTime: actionStats.AvgProcessingTime,
		CriticalRate:      actionStats.CriticalRate,
		MissRate:          actionStats.MissRate,
		BlockRate:         actionStats.BlockRate,
		AccuracyRate:      actionStats.AccuracyRate,
	}

	return stats, nil
}

// CleanupExpiredCooldowns nettoie les cooldowns expirés
func (s *ActionService) CleanupExpiredCooldowns() {
	now := time.Now()
	for key, cooldown := range s.cooldowns {
		if now.After(cooldown.ExpiresAt) {
			delete(s.cooldowns, key)
		}
	}
}

// StartCooldownCleanupRoutine démarre une routine de nettoyage des cooldowns
func (s *ActionService) StartCooldownCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			s.CleanupExpiredCooldowns()
		}
	}()
}
