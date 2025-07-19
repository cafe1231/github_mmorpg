package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	// Constantes pour la génération aléatoire sécurisée
	randomPrecision     = 1000000
	randomPrecisionF64  = 1000000.0
	fallbackRandomValue = 0.5
)

// SecureRandFloat64 génère un nombre aléatoire sécurisé entre 0.0 et 1.0
func SecureRandFloat64() float64 {
	maxVal := big.NewInt(randomPrecision)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		// Fallback en cas d'erreur (ne devrait pas arriver)
		return fallbackRandomValue
	}
	return float64(n.Int64()) / randomPrecisionF64
}

// SecureRandIntn génère un entier aléatoire sécurisé entre 0 et n-1
func SecureRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	maxVal := big.NewInt(int64(n))
	result, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		// Fallback en cas d'erreur
		return 0
	}
	return int(result.Int64())
}
