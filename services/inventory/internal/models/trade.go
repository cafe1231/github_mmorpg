package models

import (
	"time"

	"github.com/google/uuid"
)

type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"
	TradeStatusAccepted  TradeStatus = "accepted"
	TradeStatusCancelled TradeStatus = "canceled"
	TradeStatusCompleted TradeStatus = "completed"
	TradeStatusExpired   TradeStatus = "expired"
)

type Trade struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	InitiatorID    uuid.UUID   `json:"initiator_id" db:"initiator_id"`
	RecipientID    uuid.UUID   `json:"recipient_id" db:"recipient_id"`
	Status         TradeStatus `json:"status" db:"status"`
	InitiatorReady bool        `json:"initiator_ready" db:"initiator_ready"`
	RecipientReady bool        `json:"recipient_ready" db:"recipient_ready"`
	InitiatorGold  int64       `json:"initiator_gold" db:"initiator_gold"`
	RecipientGold  int64       `json:"recipient_gold" db:"recipient_gold"`
	ExpiresAt      time.Time   `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty" db:"completed_at"`

	// Relations
	InitiatorItems []TradeItem `json:"initiator_items,omitempty"`
	RecipientItems []TradeItem `json:"recipient_items,omitempty"`
}

type TradeItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TradeID   uuid.UUID `json:"trade_id" db:"trade_id"`
	ItemID    uuid.UUID `json:"item_id" db:"item_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	OwnerID   uuid.UUID `json:"owner_id" db:"owner_id"` // initiator or recipient
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	Item *Item `json:"item,omitempty"`
}

type TradeOffer struct {
	Items []TradeItemOffer `json:"items"`
	Gold  int64            `json:"gold"`
}

type TradeItemOffer struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity int       `json:"quantity"`
}

// IsExpired checks if the trade has expired
func (t *Trade) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CanModify checks if the trade can still be modified
func (t *Trade) CanModify() bool {
	return t.Status == TradeStatusPending && !t.IsExpired()
}

// IsBothPlayersReady checks if both players are ready to complete the trade
func (t *Trade) IsBothPlayersReady() bool {
	return t.InitiatorReady && t.RecipientReady
}

// CanComplete checks if the trade can be completed
func (t *Trade) CanComplete() bool {
	return t.Status == TradeStatusPending &&
		t.IsBothPlayersReady() &&
		!t.IsExpired()
}

// GetPlayerRole returns the role of a player in the trade
func (t *Trade) GetPlayerRole(playerID uuid.UUID) string {
	if t.InitiatorID == playerID {
		return "initiator"
	} else if t.RecipientID == playerID {
		return "recipient"
	}
	return ""
}

// IsPlayerInTrade checks if a player is part of the trade
func (t *Trade) IsPlayerInTrade(playerID uuid.UUID) bool {
	return t.InitiatorID == playerID || t.RecipientID == playerID
}

// GetPlayerItems returns items offered by a specific player
func (t *Trade) GetPlayerItems(playerID uuid.UUID) []TradeItem {
	if t.InitiatorID == playerID {
		return t.InitiatorItems
	} else if t.RecipientID == playerID {
		return t.RecipientItems
	}
	return []TradeItem{}
}

// GetPlayerGold returns gold offered by a specific player
func (t *Trade) GetPlayerGold(playerID uuid.UUID) int64 {
	if t.InitiatorID == playerID {
		return t.InitiatorGold
	} else if t.RecipientID == playerID {
		return t.RecipientGold
	}
	return 0
}

// IsPlayerReady checks if a specific player is ready
func (t *Trade) IsPlayerReady(playerID uuid.UUID) bool {
	if t.InitiatorID == playerID {
		return t.InitiatorReady
	} else if t.RecipientID == playerID {
		return t.RecipientReady
	}
	return false
}

// SetPlayerReady sets the ready status for a player
func (t *Trade) SetPlayerReady(playerID uuid.UUID, ready bool) {
	if t.InitiatorID == playerID {
		t.InitiatorReady = ready
	} else if t.RecipientID == playerID {
		t.RecipientReady = ready
	}
}

// ValidateTradeOffer validates a trade offer
func (t *Trade) ValidateTradeOffer(playerID uuid.UUID, offer TradeOffer) error {
	if !t.IsPlayerInTrade(playerID) {
		return NewValidationError("player is not part of this trade")
	}

	if !t.CanModify() {
		return NewValidationError("trade cannot be modified")
	}

	if offer.Gold < 0 {
		return NewValidationError("gold amount cannot be negative")
	}

	for _, item := range offer.Items {
		if item.Quantity <= 0 {
			return NewValidationError("item quantity must be positive")
		}
	}

	return nil
}

// GetOtherPlayerID returns the ID of the other player in the trade
func (t *Trade) GetOtherPlayerID(playerID uuid.UUID) uuid.UUID {
	if t.InitiatorID == playerID {
		return t.RecipientID
	}
	return t.InitiatorID
}
