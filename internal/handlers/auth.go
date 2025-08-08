package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/utils"
)

// Общие структуры запросов и ответов для Swagger и тестов

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type MnemonicWord struct {
	Position int    `json:"position"`
	Word     string `json:"word"`
}

type RegisterResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	Mnemonic     []MnemonicWord `json:"mnemonic"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RecoverChallengeResponse struct {
	Positions []int `json:"positions"`
}

type RecoverPhrase struct {
	Position int    `json:"position"`
	Word     string `json:"word"`
}

type RecoverRequest struct {
	Username        string          `json:"username"`
	Phrases         []RecoverPhrase `json:"phrases"`
	NewPassword     string          `json:"new_password"`
	PasswordConfirm string          `json:"password_confirm"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ChangeUsernameRequest struct {
	Password    string `json:"password"`
	NewUsername string `json:"new_username"`
}

type SetPinCodeRequest struct {
	Password string `json:"password"`
	PinCode  string `json:"pincode"`
}

type Enable2FARequest struct {
	Password string `json:"password"`
}

type Enable2FAResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

type VerifyPasswordRequest struct {
	Password string `json:"password"`
}

type VerifyPasswordResponse struct {
	Verified bool `json:"verified"`
}

type ProfileResponse struct {
	Username     string `json:"username"`
	TwoFAEnabled bool   `json:"twofa_enabled"`
	PinCodeSet   bool   `json:"pincode_set"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Register godoc
// @Summary Регистрация клиента
// @Description Создаёт нового клиента с уникальным именем, хешем пароля и мнемонической фразой
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "данные регистрации"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func Register(db *gorm.DB, ttl map[string]time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r RegisterRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		if r.Password != r.PasswordConfirm {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "passwords do not match"})
			return
		}
		var count int64
		db.Model(&models.Client{}).Where("username = ?", r.Username).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "username exists"})
			return
		}
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "hash error"})
			return
		}
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "entropy error"})
			return
		}
		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "mnemonic error"})
			return
		}
		words := strings.Split(mnemonic, " ")
		hashes := make([]string, len(words))
		respMn := make([]MnemonicWord, len(words))
		for i, w := range words {
			h := sha256.Sum256([]byte(w))
			hashes[i] = hex.EncodeToString(h[:])
			respMn[i] = MnemonicWord{Position: i + 1, Word: w}
		}
		hashesJSON, _ := json.Marshal(hashes)
		pwd := string(pwdHash)
		client := models.Client{
			Username: r.Username,
			Password: &pwd,
			Bip39:    datatypes.JSON(hashesJSON),
		}
		if err := db.Create(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		accessStr, _ := utils.GenerateNanoID()
		refreshStr, _ := utils.GenerateNanoID()
		access := models.Token{ClientID: client.ID, Token: accessStr, Type: "access", ExpiresAt: time.Now().Add(ttl["access"])}
		refresh := models.Token{ClientID: client.ID, Token: refreshStr, Type: "refresh", ExpiresAt: time.Now().Add(ttl["refresh"])}
		db.Create(&access)
		db.Create(&refresh)
		c.JSON(http.StatusOK, RegisterResponse{
			AccessToken:  accessStr,
			RefreshToken: refreshStr,
			Mnemonic:     respMn,
		})
	}
}

// Login godoc
// @Summary Вход клиента
// @Description Аутентифицирует клиента и выдаёт пару токенов. При включённой 2FA требуется код.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "учётные данные"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func Login(db *gorm.DB, ttl map[string]time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r LoginRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		var client models.Client
		if err := db.Where("username = ?", r.Username).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.Password)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
			return
		}
		if client.TwoFAEnabled {
			if r.Code == "" || client.TOTPSecret == nil || !totp.Validate(r.Code, *client.TOTPSecret) {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid code"})
				return
			}
		}
		accessStr, _ := utils.GenerateNanoID()
		refreshStr, _ := utils.GenerateNanoID()
		access := models.Token{ClientID: client.ID, Token: accessStr, Type: "access", ExpiresAt: time.Now().Add(ttl["access"])}
		refresh := models.Token{ClientID: client.ID, Token: refreshStr, Type: "refresh", ExpiresAt: time.Now().Add(ttl["refresh"])}
		db.Create(&access)
		db.Create(&refresh)
		c.JSON(http.StatusOK, TokenResponse{AccessToken: accessStr, RefreshToken: refreshStr})
	}
}

// Refresh godoc
// @Summary Обновление access токена
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshRequest true "refresh токен"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func Refresh(db *gorm.DB, ttl map[string]time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r RefreshRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		var token models.Token
		if err := db.Where("token = ? AND type = ?", r.RefreshToken, "refresh").First(&token).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			return
		}
		if token.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "token expired"})
			return
		}
		db.Delete(&token)
		accessStr, _ := utils.GenerateNanoID()
		refreshStr, _ := utils.GenerateNanoID()
		access := models.Token{ClientID: token.ClientID, Token: accessStr, Type: "access", ExpiresAt: time.Now().Add(ttl["access"])}
		refresh := models.Token{ClientID: token.ClientID, Token: refreshStr, Type: "refresh", ExpiresAt: time.Now().Add(ttl["refresh"])}
		db.Create(&access)
		db.Create(&refresh)
		c.JSON(http.StatusOK, TokenResponse{AccessToken: accessStr, RefreshToken: refreshStr})
	}
}

// RecoverChallenge godoc
// @Summary Запрос позиций для восстановления
// @Tags auth
// @Produce json
// @Param username path string true "имя пользователя"
// @Success 200 {object} RecoverChallengeResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/recover/{username} [get]
func RecoverChallenge(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		var client models.Client
		if err := db.Where("username = ?", username).First(&client).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		var hashes []string
		if err := json.Unmarshal(client.Bip39, &hashes); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "mnemonic error"})
			return
		}
		if len(hashes) < 3 {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "mnemonic too short"})
			return
		}
		rand.Seed(time.Now().UnixNano())
		perm := rand.Perm(len(hashes))
		idx := []int{perm[0] + 1, perm[1] + 1, perm[2] + 1}
		c.JSON(http.StatusOK, RecoverChallengeResponse{Positions: idx})
	}
}

// Recover godoc
// @Summary Восстановление доступа
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RecoverRequest true "фразы и новый пароль"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/recover [post]
func Recover(db *gorm.DB, ttl map[string]time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r RecoverRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		var client models.Client
		if err := db.Where("username = ?", r.Username).First(&client).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
			return
		}
		var hashes []string
		if err := json.Unmarshal(client.Bip39, &hashes); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "mnemonic error"})
			return
		}
		if len(r.Phrases) != 3 {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "need 3 phrases"})
			return
		}
		for _, p := range r.Phrases {
			if p.Position <= 0 || p.Position > len(hashes) {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid position"})
				return
			}
			h := sha256.Sum256([]byte(p.Word))
			if hex.EncodeToString(h[:]) != hashes[p.Position-1] {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid phrase"})
				return
			}
		}
		if r.NewPassword != r.PasswordConfirm {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "passwords do not match"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(r.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "hash error"})
			return
		}
		pwd := string(hash)
		client.Password = &pwd
		if err := db.Save(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		accessStr, _ := utils.GenerateNanoID()
		refreshStr, _ := utils.GenerateNanoID()
		access := models.Token{ClientID: client.ID, Token: accessStr, Type: "access", ExpiresAt: time.Now().Add(ttl["access"])}
		refresh := models.Token{ClientID: client.ID, Token: refreshStr, Type: "refresh", ExpiresAt: time.Now().Add(ttl["refresh"])}
		db.Create(&access)
		db.Create(&refresh)
		c.JSON(http.StatusOK, TokenResponse{AccessToken: accessStr, RefreshToken: refreshStr})
	}
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
			return
		}
		tokenStr := parts[1]
		var token models.Token
		if err := db.Where("token = ? AND type = ?", tokenStr, "access").First(&token).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if token.ExpiresAt.Before(time.Now()) {
			db.Delete(&token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}
		c.Set("client_id", token.ClientID)
		c.Next()
	}
}

// Logout godoc
// @Summary Выход клиента
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func Logout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		db.Where("client_id = ?", clientID).Delete(&models.Token{})
		c.JSON(http.StatusOK, StatusResponse{Status: "logged out"})
	}
}

// VerifyPassword godoc
// @Summary Проверка текущего пароля
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body VerifyPasswordRequest true "пароль"
// @Success 200 {object} VerifyPasswordResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/verify-password [post]
func VerifyPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r VerifyPasswordRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.Password)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
			return
		}
		c.JSON(http.StatusOK, VerifyPasswordResponse{Verified: true})
	}
}

// ChangePassword godoc
// @Summary Смена пароля
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ChangePasswordRequest true "пароли"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/password [post]
func ChangePassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r ChangePasswordRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		if r.NewPassword != r.ConfirmPassword {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "passwords do not match"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.OldPassword)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(r.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "hash error"})
			return
		}
		pwd := string(hash)
		client.Password = &pwd
		if err := db.Save(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, StatusResponse{Status: "password updated"})
	}
}

// Profile godoc
// @Summary Профиль клиента
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/profile [get]
func Profile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		c.JSON(http.StatusOK, ProfileResponse{
			Username:     client.Username,
			TwoFAEnabled: client.TwoFAEnabled,
			PinCodeSet:   client.PinCode != nil,
		})
	}
}

// ChangeUsername godoc
// @Summary Смена имени пользователя
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ChangeUsernameRequest true "новое имя"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/username [post]
func ChangeUsername(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r ChangeUsernameRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.Password)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
			return
		}
		var count int64
		db.Model(&models.Client{}).Where("username = ?", r.NewUsername).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "username exists"})
			return
		}
		client.Username = r.NewUsername
		if err := db.Save(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, StatusResponse{Status: "username updated"})
	}
}

// SetPinCode godoc
// @Summary Установка PIN-кода
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body SetPinCodeRequest true "пин-код"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/pincode [post]
func SetPinCode(db *gorm.DB) gin.HandlerFunc {
	re := regexp.MustCompile(`^[0-9]{4}$`)
	return func(c *gin.Context) {
		var r SetPinCodeRequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		if !re.MatchString(r.PinCode) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid pincode"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.Password)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(r.PinCode), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "hash error"})
			return
		}
		s := string(hash)
		client.PinCode = &s
		if err := db.Save(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, StatusResponse{Status: "pincode set"})
	}
}

// Enable2FA godoc
// @Summary Включение двухфакторной аутентификации
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body Enable2FARequest true "подтверждение пароля"
// @Success 200 {object} Enable2FAResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/2fa/enable [post]
func Enable2FA(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var r Enable2FARequest
		if err := c.BindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
			return
		}
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID, _ := clientIDVal.(string)
		var client models.Client
		if err := db.Where("id = ?", clientID).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid client"})
			return
		}
		if client.Password == nil || bcrypt.CompareHashAndPassword([]byte(*client.Password), []byte(r.Password)) != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid password"})
			return
		}
		if client.TwoFAEnabled {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "2fa already enabled"})
			return
		}
		key, err := totp.Generate(totp.GenerateOpts{Issuer: "ptop", AccountName: client.Username})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "totp error"})
			return
		}
		secret := key.Secret()
		client.TwoFAEnabled = true
		client.TOTPSecret = &secret
		if err := db.Save(&client).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
		c.JSON(http.StatusOK, Enable2FAResponse{Secret: secret, URL: key.URL()})
	}
}
