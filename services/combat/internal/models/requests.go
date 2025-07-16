package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateCombatRequest représente une demande de création de combat
type CreateCombatRequest struct {
	CombatType      CombatType           `json:"combat_type" binding:"required"`
	ZoneID          string               `json:"zone_id,omitempty"`
	MaxParticipants int                  `json:"max_participants,omitempty"`
	TurnTimeLimit   int                  `json:"turn_time_limit,omitempty"`
	MaxDuration     int                  `json:"max_duration,omitempty"`
	Settings        *CombatSettings      `json:"settings,omitempty"`
	Participants    []ParticipantRequest `json:"participants,omitempty"`
}

// ParticipantRequest représente une demande d'ajout de participant
type ParticipantRequest struct {
	CharacterID uuid.UUID `json:"character_id" binding:"required"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Team        int       `json:"team"`
	Position    int       `json:"position"`
}

// JoinCombatRequest représente une demande de rejoindre un combat
type JoinCombatRequest struct {
	CharacterID uuid.UUID `json:"character_id" binding:"required"`
	Team        int       `json:"team,omitempty"`
	Position    int       `json:"position,omitempty"`
}

// LeaveCombatRequest représente une demande de quitter un combat
type LeaveCombatRequest struct {
	Reason string `json:"reason,omitempty"`
}

// StartCombatRequest représente une demande de démarrage de combat
type StartCombatRequest struct {
	ForceStart bool `json:"force_start,omitempty"`
}

// EndCombatRequest représente une demande de fin de combat
type EndCombatRequest struct {
	Reason   string `json:"reason" binding:"required"`
	WinnerID *int   `json:"winner_id,omitempty"`
	ForceEnd bool   `json:"force_end,omitempty"`
}

// GetCombatStatusRequest représente une demande de statut de combat
type GetCombatStatusRequest struct {
	IncludeParticipants bool       `json:"include_participants,omitempty"`
	IncludeActions      bool       `json:"include_actions,omitempty"`
	IncludeEffects      bool       `json:"include_effects,omitempty"`
	IncludeLogs         bool       `json:"include_logs,omitempty"`
	LastActionID        *uuid.UUID `json:"last_action_id,omitempty"`
}

// UpdateParticipantRequest représente une demande de mise à jour de participant
type UpdateParticipantRequest struct {
	Health   *int  `json:"health,omitempty"`
	Mana     *int  `json:"mana,omitempty"`
	IsReady  *bool `json:"is_ready,omitempty"`
	Position *int  `json:"position,omitempty"`
	Team     *int  `json:"team,omitempty"`
}

// SearchCombatsRequest représente une demande de recherche de combats
type SearchCombatsRequest struct {
	CombatType      *CombatType   `json:"combat_type,omitempty"`
	Status          *CombatStatus `json:"status,omitempty"`
	ZoneID          *string       `json:"zone_id,omitempty"`
	ParticipantID   *uuid.UUID    `json:"participant_id,omitempty"`
	CreatedAfter    *time.Time    `json:"created_after,omitempty"`
	CreatedBefore   *time.Time    `json:"created_before,omitempty"`
	Limit           int           `json:"limit,omitempty"`
	Offset          int           `json:"offset,omitempty"`
	IncludeFinished bool          `json:"include_finished,omitempty"`
}

// GetCombatHistoryRequest représente une demande d'historique de combat
type GetCombatHistoryRequest struct {
	CharacterID *uuid.UUID  `json:"character_id,omitempty"`
	UserID      *uuid.UUID  `json:"user_id,omitempty"`
	CombatType  *CombatType `json:"combat_type,omitempty"`
	DateFrom    *time.Time  `json:"date_from,omitempty"`
	DateTo      *time.Time  `json:"date_to,omitempty"`
	WinsOnly    bool        `json:"wins_only,omitempty"`
	LossesOnly  bool        `json:"losses_only,omitempty"`
	Limit       int         `json:"limit,omitempty"`
	Offset      int         `json:"offset,omitempty"`
}

// GetStatisticsRequest représente une demande de statistiques
type GetStatisticsRequest struct {
	CharacterID *uuid.UUID  `json:"character_id,omitempty"`
	UserID      *uuid.UUID  `json:"user_id,omitempty"`
	CombatType  *CombatType `json:"combat_type,omitempty"`
	Period      string      `json:"period,omitempty"` // "day", "week", "month", "year", "all"
	Detailed    bool        `json:"detailed,omitempty"`
}

// ApplyEffectRequest représente une demande d'application d'effet
type ApplyEffectRequest struct {
	EffectID    string                 `json:"effect_id" binding:"required"`
	TargetID    uuid.UUID              `json:"target_id" binding:"required"`
	CasterID    *uuid.UUID             `json:"caster_id,omitempty"`
	Duration    *int                   `json:"duration,omitempty"`
	Stacks      *int                   `json:"stacks,omitempty"`
	CustomValue *int                   `json:"custom_value,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RemoveEffectRequest représente une demande de suppression d'effet
type RemoveEffectRequest struct {
	EffectID    *uuid.UUID  `json:"effect_id,omitempty"`
	EffectType  *EffectType `json:"effect_type,omitempty"`
	TargetID    uuid.UUID   `json:"target_id" binding:"required"`
	RemoveAll   bool        `json:"remove_all,omitempty"`
	OnlyDebuffs bool        `json:"only_debuffs,omitempty"`
	OnlyBuffs   bool        `json:"only_buffs,omitempty"`
}

// ProcessTurnRequest représente une demande de traitement de tour
type ProcessTurnRequest struct {
	AdvanceTurn        bool `json:"advance_turn,omitempty"`
	ProcessEffects     bool `json:"process_effects,omitempty"`
	CheckWinConditions bool `json:"check_win_conditions,omitempty"`
}

// ValidateActionRequest représente une demande de validation d'action
type ValidateActionRequest struct {
	Action         *ActionRequest `json:"action" binding:"required"`
	ActorID        uuid.UUID      `json:"actor_id" binding:"required"`
	Strict         bool           `json:"strict,omitempty"`
	CheckCooldowns bool           `json:"check_cooldowns,omitempty"`
	CheckResources bool           `json:"check_resources,omitempty"`
}

// BulkActionRequest représente une demande d'actions multiples
type BulkActionRequest struct {
	Actions     []ActionRequest `json:"actions" binding:"required"`
	Atomic      bool            `json:"atomic,omitempty"`     // Toutes ou aucune
	Sequential  bool            `json:"sequential,omitempty"` // Exécuter dans l'ordre
	StopOnError bool            `json:"stop_on_error,omitempty"`
}

// ReplayRequest représente une demande de rejeu de combat
type ReplayRequest struct {
	CombatID      uuid.UUID `json:"combat_id" binding:"required"`
	Speed         float64   `json:"speed,omitempty"` // Vitesse de rejeu (1.0 = normal)
	StartFromTurn int       `json:"start_from_turn,omitempty"`
	EndAtTurn     int       `json:"end_at_turn,omitempty"`
	IncludeUI     bool      `json:"include_ui,omitempty"`
}

// AdminActionRequest représente une demande d'action administrative
type AdminActionRequest struct {
	Action     string                 `json:"action" binding:"required"`
	TargetID   uuid.UUID              `json:"target_id" binding:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	Silent     bool                   `json:"silent,omitempty"`
}

// Validate valide une demande de création de combat
func (r *CreateCombatRequest) Validate() error {
	// Validation du type de combat
	validTypes := []CombatType{CombatTypePvE, CombatTypePvP, CombatTypeDungeon, CombatTypeRaid}
	isValid := false
	for _, t := range validTypes {
		if r.CombatType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("type de combat invalide")
	}

	// Validation des participants
	if r.MaxParticipants > 0 && len(r.Participants) > r.MaxParticipants {
		return fmt.Errorf("trop de participants: %d/%d", len(r.Participants), r.MaxParticipants)
	}

	// Validation des limites de temps
	if r.TurnTimeLimit > 0 && r.TurnTimeLimit < 5 {
		return fmt.Errorf("limite de temps de tour trop courte (minimum 5 secondes)")
	}
	if r.TurnTimeLimit > 300 {
		return fmt.Errorf("limite de temps de tour trop longue (maximum 300 secondes)")
	}

	if r.MaxDuration > 0 && r.MaxDuration < 60 {
		return fmt.Errorf("durée maximale trop courte (minimum 60 secondes)")
	}
	if r.MaxDuration > 3600 {
		return fmt.Errorf("durée maximale trop longue (maximum 3600 secondes)")
	}

	return nil
}

// Validate valide une demande de recherche de combats
func (r *SearchCombatsRequest) Validate() error {
	// Validation de la limite
	if r.Limit < 0 {
		return fmt.Errorf("la limite ne peut pas être négative")
	}
	if r.Limit > 100 {
		return fmt.Errorf("la limite ne peut pas dépasser 100")
	}
	if r.Limit == 0 {
		r.Limit = 20 // Valeur par défaut
	}

	// Validation de l'offset
	if r.Offset < 0 {
		return fmt.Errorf("l'offset ne peut pas être négatif")
	}

	// Validation des dates
	if r.CreatedAfter != nil && r.CreatedBefore != nil {
		if r.CreatedAfter.After(*r.CreatedBefore) {
			return fmt.Errorf("la date de début doit être antérieure à la date de fin")
		}
	}

	return nil
}

// Validate valide une demande d'historique de combat
func (r *GetCombatHistoryRequest) Validate() error {
	// Au moins un critère de filtre requis
	if r.CharacterID == nil && r.UserID == nil {
		return fmt.Errorf("character_id ou user_id requis")
	}

	// Validation de la limite
	if r.Limit < 0 {
		return fmt.Errorf("la limite ne peut pas être négative")
	}
	if r.Limit > 200 {
		return fmt.Errorf("la limite ne peut pas dépasser 200")
	}
	if r.Limit == 0 {
		r.Limit = 50 // Valeur par défaut
	}

	// Validation de l'offset
	if r.Offset < 0 {
		return fmt.Errorf("l'offset ne peut pas être négatif")
	}

	// Validation des dates
	if r.DateFrom != nil && r.DateTo != nil {
		if r.DateFrom.After(*r.DateTo) {
			return fmt.Errorf("la date de début doit être antérieure à la date de fin")
		}
	}

	// Validation des filtres exclusifs
	if r.WinsOnly && r.LossesOnly {
		return fmt.Errorf("wins_only et losses_only ne peuvent pas être vrais en même temps")
	}

	return nil
}

// Validate valide une demande de statistiques
func (r *GetStatisticsRequest) Validate() error {
	// Au moins un critère de filtre requis
	if r.CharacterID == nil && r.UserID == nil {
		return fmt.Errorf("character_id ou user_id requis")
	}

	// Validation de la période
	validPeriods := []string{"day", "week", "month", "year", "all"}
	if r.Period != "" {
		isValid := false
		for _, p := range validPeriods {
			if r.Period == p {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("période invalide, valeurs acceptées: %v", validPeriods)
		}
	} else {
		r.Period = "all" // Valeur par défaut
	}

	return nil
}

// Validate valide une demande d'application d'effet
func (r *ApplyEffectRequest) Validate() error {
	// Validation de l'ID d'effet
	if r.EffectID == "" {
		return fmt.Errorf("effect_id requis")
	}

	// Validation de la durée
	if r.Duration != nil {
		if *r.Duration < 0 {
			return fmt.Errorf("la durée ne peut pas être négative")
		}
		if *r.Duration > 100 {
			return fmt.Errorf("la durée ne peut pas dépasser 100 tours")
		}
	}

	// Validation des stacks
	if r.Stacks != nil {
		if *r.Stacks < 1 {
			return fmt.Errorf("le nombre de stacks doit être au moins 1")
		}
		if *r.Stacks > 10 {
			return fmt.Errorf("le nombre de stacks ne peut pas dépasser 10")
		}
	}

	return nil
}

// Validate valide une demande d'actions multiples
func (r *BulkActionRequest) Validate() error {
	// Validation du nombre d'actions
	if len(r.Actions) == 0 {
		return fmt.Errorf("au moins une action requise")
	}
	if len(r.Actions) > 10 {
		return fmt.Errorf("impossible d'exécuter plus de 10 actions à la fois")
	}

	// Validation de chaque action
	for i, action := range r.Actions {
		if validation := action.Validate(); !validation.IsValid {
			return fmt.Errorf("action %d invalide: %v", i, validation.Errors)
		}
	}

	return nil
}

// Validate valide une demande de rejeu
func (r *ReplayRequest) Validate() error {
	// Validation de la vitesse
	if r.Speed <= 0 {
		r.Speed = 1.0 // Valeur par défaut
	}
	if r.Speed > 10.0 {
		return fmt.Errorf("la vitesse ne peut pas dépasser 10x")
	}

	// Validation des tours
	if r.StartFromTurn < 0 {
		return fmt.Errorf("le tour de début ne peut pas être négatif")
	}
	if r.EndAtTurn > 0 && r.EndAtTurn <= r.StartFromTurn {
		return fmt.Errorf("le tour de fin doit être supérieur au tour de début")
	}

	return nil
}

// Validate valide une demande d'action administrative
func (r *AdminActionRequest) Validate() error {
	// Validation de l'action
	validActions := []string{
		"force_end_combat",
		"kick_participant",
		"reset_combat",
		"apply_effect",
		"remove_effect",
		"modify_stats",
		"grant_victory",
		"cancel_combat",
	}

	isValid := false
	for _, action := range validActions {
		if r.Action == action {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("action administrative invalide: %s", r.Action)
	}

	// Validation des paramètres selon l'action
	switch r.Action {
	case "kick_participant":
		if r.Parameters == nil || r.Parameters["reason"] == nil {
			return fmt.Errorf("paramètre 'reason' requis pour kick_participant")
		}
	case "apply_effect", "remove_effect":
		if r.Parameters == nil || r.Parameters["effect_id"] == nil {
			return fmt.Errorf("paramètre 'effect_id' requis pour %s", r.Action)
		}
	case "modify_stats":
		if r.Parameters == nil {
			return fmt.Errorf("paramètres requis pour modify_stats")
		}
	case "grant_victory":
		if r.Parameters == nil || r.Parameters["winner_id"] == nil {
			return fmt.Errorf("paramètre 'winner_id' requis pour grant_victory")
		}
	}

	return nil
}

// GetDefaultCreateCombatRequest retourne une demande par défaut
func GetDefaultCreateCombatRequest() *CreateCombatRequest {
	// Créer une variable temporaire pour pouvoir prendre son adresse
	defaultSettings := GetDefaultCombatSettings()

	return &CreateCombatRequest{
		CombatType:      CombatTypePvE,
		MaxParticipants: 4,
		TurnTimeLimit:   30,
		MaxDuration:     300,
		Settings:        &defaultSettings, // Maintenant on peut prendre l'adresse
		Participants:    []ParticipantRequest{},
	}
}

// GetDefaultSearchRequest retourne une demande de recherche par défaut
func GetDefaultSearchRequest() *SearchCombatsRequest {
	return &SearchCombatsRequest{
		Limit:           20,
		Offset:          0,
		IncludeFinished: false,
	}
}
