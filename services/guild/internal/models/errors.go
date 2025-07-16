package models

// GuildError représente une erreur de guilde
type GuildError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implémente l'interface error
func (e *GuildError) Error() string {
	return e.Message
}

// Erreurs communes
var (
	ErrGuildNotFound           = &GuildError{Code: "GUILD_NOT_FOUND", Message: "Guilde non trouvée"}
	ErrGuildAlreadyExists      = &GuildError{Code: "GUILD_ALREADY_EXISTS", Message: "Une guilde avec ce nom existe déjà"}
	ErrMemberNotFound          = &GuildError{Code: "MEMBER_NOT_FOUND", Message: "Membre non trouvé"}
	ErrAlreadyInGuild          = &GuildError{Code: "ALREADY_IN_GUILD", Message: "Le joueur est déjà dans une guilde"}
	ErrNotInGuild              = &GuildError{Code: "NOT_IN_GUILD", Message: "Le joueur n'est pas dans cette guilde"}
	ErrInsufficientPermissions = &GuildError{Code: "INSUFFICIENT_PERMISSIONS", Message: "Permissions insuffisantes"}
	ErrGuildFull               = &GuildError{Code: "GUILD_FULL", Message: "La guilde est pleine"}
	ErrInvitationNotFound      = &GuildError{Code: "INVITATION_NOT_FOUND", Message: "Invitation non trouvée"}
	ErrInvitationExpired       = &GuildError{Code: "INVITATION_EXPIRED", Message: "L'invitation a expiré"}
	ErrApplicationNotFound     = &GuildError{Code: "APPLICATION_NOT_FOUND", Message: "Candidature non trouvée"}
	ErrBankTransactionFailed   = &GuildError{Code: "BANK_TRANSACTION_FAILED", Message: "Transaction bancaire échouée"}
	ErrInsufficientFunds       = &GuildError{Code: "INSUFFICIENT_FUNDS", Message: "Fonds insuffisants"}
	ErrWarAlreadyExists        = &GuildError{Code: "WAR_ALREADY_EXISTS", Message: "Une guerre existe déjà entre ces guildes"}
	ErrAllianceAlreadyExists   = &GuildError{Code: "ALLIANCE_ALREADY_EXISTS", Message: "Une alliance existe déjà entre ces guildes"}
)
