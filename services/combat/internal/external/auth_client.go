// internal/external/auth_client.go
package external

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"combat/internal/config"
)

// AuthClientInterface définit les méthodes pour communiquer avec le service Auth
type AuthClientInterface interface {
	ValidateToken(token string) (*TokenInfo, error)
	GetUserInfo(userID uuid.UUID) (*UserInfo, error)
}

// TokenInfo représente les informations d'un token validé
type TokenInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	Valid       bool      `json:"valid"`
}

// UserInfo représente les informations d'un utilisateur
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	IsActive bool      `json:"is_active"`
}

// AuthClient implémente l'interface AuthClientInterface
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
	config     *config.Config
}

// NewAuthClient crée une nouvelle instance du client Auth
func NewAuthClient(cfg *config.Config) AuthClientInterface {
	return &AuthClient{
		baseURL: cfg.Services.Auth.URL,
		httpClient: &http.Client{
			Timeout: cfg.Services.Auth.Timeout,
		},
		config: cfg,
	}
}

// ValidateToken valide un token JWT auprès du service Auth
func (c *AuthClient) ValidateToken(token string) (*TokenInfo, error) {
	// Dans notre architecture, la validation JWT se fait localement
	// Ce client pourrait être utilisé pour des validations spéciales
	// ou des vérifications de révocation de tokens
	
	// Pour l'instant, on retourne une implémentation simple
	return &TokenInfo{
		Valid: true,
	}, nil
}

// GetUserInfo récupère les informations d'un utilisateur
func (c *AuthClient) GetUserInfo(userID uuid.UUID) (*UserInfo, error) {
	url := fmt.Sprintf("%s/api/v1/services/user/%s", c.baseURL, userID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	// Pour l'instant, retourner une implémentation basique
	// Dans un vrai système, on décode la réponse JSON
	return &UserInfo{
		ID:       userID,
		IsActive: true,
	}, nil
}