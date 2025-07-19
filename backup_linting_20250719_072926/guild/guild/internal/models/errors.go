package models

import "errors"

// Erreurs spécifiques aux guildes
var (
	ErrGuildNotFound           = errors.New("guilde non trouvée")
	ErrGuildAlreadyExists      = errors.New("une guilde avec ce nom ou tag existe déjà")
	ErrAlreadyInGuild          = errors.New("le joueur est déjà dans une guilde")
	ErrNotInGuild              = errors.New("le joueur n'est pas dans cette guilde")
	ErrGuildFull               = errors.New("la guilde est pleine")
	ErrInsufficientPermissions = errors.New("permissions insuffisantes")
	ErrInvalidRole             = errors.New("rôle invalide")
	ErrMemberNotFound          = errors.New("membre non trouvé")
	ErrInvitationNotFound      = errors.New("invitation non trouvée")
	ErrApplicationNotFound     = errors.New("candidature non trouvée")
	ErrWarNotFound             = errors.New("guerre non trouvée")
	ErrAllianceNotFound        = errors.New("alliance non trouvée")
	ErrBankTransactionFailed   = errors.New("transaction bancaire échouée")
	ErrInsufficientFunds       = errors.New("fonds insuffisants")
)
