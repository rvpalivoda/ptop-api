package services

import (
	"fmt"
	"ptop/internal/utils"
)

func GetAddress(clientID, assetID string) (string, error) {
	id, err := utils.GenerateNanoID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("addr_%s_%s_%s", assetID, clientID, id), nil
}
