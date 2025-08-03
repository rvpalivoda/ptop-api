package services

import (
	"bytes"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	btcec "github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
	solana "github.com/gagliardetto/solana-go"
	bip39 "github.com/tyler-smith/go-bip39"
	ed25519hd "github.com/wealdtech/go-ed25519hd"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func nextIndex(db *gorm.DB, assetID string) (uint32, error) {
	var max sql.NullInt64
	if err := db.Model(&models.Wallet{}).Where("asset_id = ?", assetID).Select("MAX(derivation_index)").Scan(&max).Error; err != nil {
		return 0, err
	}
	if max.Valid {
		return uint32(max.Int64 + 1), nil
	}
	return 0, nil
}

func GetAddress(db *gorm.DB, clientID, assetID string) (string, uint32, error) {
	var asset models.Asset
	if err := db.Where("id = ?", assetID).First(&asset).Error; err != nil {
		return "", 0, err
	}

	if strings.EqualFold(os.Getenv("DEBUG_FAKE_NETWORK"), "true") {
		idx, err := nextIndex(db, assetID)
		if err != nil {
			return "", 0, err
		}
		return fmt.Sprintf("fake:%s:%s:%d", assetID, clientID, idx), idx, nil
	}

	switch {
	case strings.HasPrefix(asset.Name, "BTC"):
		idx, err := nextIndex(db, assetID)
		if err != nil {
			return "", 0, err
		}
		if asset.Xpub == "" {
			return "", 0, errors.New("no xpub")
		}
		key, err := hdkeychain.NewKeyFromString(asset.Xpub)
		if err != nil {
			return "", 0, err
		}
		child, err := key.Child(idx)
		if err != nil {
			return "", 0, err
		}
		addr, err := child.Address(&chaincfg.MainNetParams)
		if err != nil {
			return "", 0, err
		}
		return addr.EncodeAddress(), idx, nil

	case strings.HasPrefix(asset.Name, "ETH"):
		idx, err := nextIndex(db, assetID)
		if err != nil {
			return "", 0, err
		}
		if asset.Xpub == "" {
			return "", 0, errors.New("no xpub")
		}
		key, err := hdkeychain.NewKeyFromString(asset.Xpub)
		if err != nil {
			return "", 0, err
		}
		change, err := key.Child(0)
		if err != nil {
			return "", 0, err
		}
		child, err := change.Child(idx)
		if err != nil {
			return "", 0, err
		}
		pk, err := child.ECPubKey()
		if err != nil {
			return "", 0, err
		}
		ecdsaPK := ecdsa.PublicKey{Curve: btcec.S256(), X: pk.X, Y: pk.Y}
		addr := crypto.PubkeyToAddress(ecdsaPK)
		return addr.Hex(), idx, nil

	case strings.HasPrefix(asset.Name, "USDC"):
		idx, err := nextIndex(db, assetID)
		if err != nil {
			return "", 0, err
		}
		if asset.Xpub == "" {
			return "", 0, errors.New("no seed")
		}
		seed := bip39.NewSeed(asset.Xpub, "")
		path := fmt.Sprintf("m/44'/501'/0'/0'/%d", idx)
		pub, _, err := ed25519hd.Keys(seed, path)
		if err != nil {
			return "", 0, err
		}
		pubKey := solana.PublicKeyFromBytes(pub)
		mintAddr := os.Getenv("USDC_MINT_ADDRESS")
		if mintAddr == "" {
			return "", 0, errors.New("no usdc mint")
		}
		mint := solana.MustPublicKeyFromBase58(mintAddr)
		ata, _, err := solana.FindAssociatedTokenAddress(pubKey, mint)
		if err != nil {
			return "", 0, err
		}
		return ata.String(), idx, nil

	case strings.HasPrefix(asset.Name, "XMR"):
		rpcURL := os.Getenv("MONERO_RPC_URL")
		if rpcURL == "" {
			rpcURL = "http://localhost:18083/json_rpc"
		}
		payload := map[string]any{
			"jsonrpc": "2.0",
			"id":      "0",
			"method":  "create_address",
			"params": map[string]any{
				"account_index": 0,
				"label":         clientID,
			},
		}
		body, _ := json.Marshal(payload)
		resp, err := http.Post(rpcURL, "application/json", bytes.NewReader(body))
		if err != nil {
			return "", 0, err
		}
		defer resp.Body.Close()
		var res struct {
			Result struct {
				Address      string `json:"address"`
				AddressIndex uint32 `json:"address_index"`
			} `json:"result"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return "", 0, err
		}
		if res.Error != nil {
			return "", 0, errors.New(res.Error.Message)
		}
		return res.Result.Address, res.Result.AddressIndex, nil

	default:
		return "", 0, fmt.Errorf("unsupported asset")
	}
}
