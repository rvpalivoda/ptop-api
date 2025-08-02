package utils

import gonanoid "github.com/matoous/go-nanoid/v2"

func GenerateNanoID() (string, error) {
	return gonanoid.New()
}
