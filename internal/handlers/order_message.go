package handlers

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ptop/internal/models"
	"ptop/internal/orderchat"
	"ptop/internal/services"
	storage "ptop/internal/services/storage"
	"ptop/internal/utils"
)

// OrderMessageRequest используется для текстовых сообщений.
type OrderMessageRequest struct {
    Content string `json:"content"`
}

// ReadOrderMessageRequest используется для отметки прочтения
type ReadOrderMessageRequest struct {
    ReadAt *time.Time `json:"readAt"`
}

// ListOrderMessages godoc
// @Summary Список сообщений ордера
// @Description Возвращает историю переписки. В каждом сообщении присутствует поле `senderName` — имя отправителя (по `client_id`). Новые сообщения приходят в реальном времени через WebSocket `/ws/orders/{id}/chat`.
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID ордера"
// @Param cursor query string false "cursor"
// @Param after query string false "after"
// @Success 200 {array} models.OrderMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages [get]
func ListOrderMessages(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
		orderID := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var order models.Order
		if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}
		if clientID != order.BuyerID && clientID != order.SellerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var chat models.OrderChat
		if err := db.Where("order_id = ?", order.ID).First(&chat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, []models.OrderMessage{})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
			return
		}
        q := db.Where("chat_id = ?", chat.ID)
		if cursor := c.Query("cursor"); cursor != "" {
			q = q.Where("id > ?", cursor)
		}
		if after := c.Query("after"); after != "" {
			if t, err := time.Parse(time.RFC3339, after); err == nil {
				q = q.Where("created_at > ?", t)
			}
		}
        var msgs []models.OrderMessage
        if err := q.Preload("Client").Order("created_at asc").Limit(50).Find(&msgs).Error; err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
            return
        }
        // обогащаем ответ именем отправителя
        for i := range msgs {
            msgs[i].SenderName = msgs[i].Client.Username
        }
        c.JSON(http.StatusOK, msgs)
    }
}

// CreateOrderMessage godoc
// @Summary Отправить сообщение в ордер
// @Description Поддерживаются типы файлов: image/jpeg, image/png, application/pdf. В ответе возвращается `senderName` — имя отправителя. После сохранения сообщение мгновенно рассылается всем участникам через WebSocket `/ws/orders/{id}/chat`.
// @Tags orders
// @Security BearerAuth
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "ID ордера"
// @Param input body OrderMessageRequest false "данные"
// @Param file formData file false "файл (image/jpeg, image/png, application/pdf)"
// @Success 200 {object} models.OrderMessage
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages [post]
func CreateOrderMessage(db *gorm.DB, st storage.Storage, cache *services.ChatCache) gin.HandlerFunc {
    return func(c *gin.Context) {
		orderID := c.Param("id")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var order models.Order
		if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}
		if clientID != order.BuyerID && clientID != order.SellerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var chat models.OrderChat
		if err := db.Where("order_id = ?", order.ID).First(&chat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chat = models.OrderChat{OrderID: order.ID}
				if err := db.Create(&chat).Error; err != nil {
					c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
				return
			}
		}
		var msg models.OrderMessage
		if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid file"})
				return
			}
			f, err := file.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid file"})
				return
			}
			defer f.Close()
			buf := make([]byte, 512)
			n, _ := f.Read(buf)
			mimeType := http.DetectContentType(buf[:n])
			ext := strings.ToLower(filepath.Ext(file.Filename))
			allowed := map[string]string{
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".png":  "image/png",
				".pdf":  "application/pdf",
			}
			if allowed[ext] != mimeType {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unsupported file type"})
				return
			}
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "file error"})
				return
			}
			id, err := utils.GenerateNanoID()
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "id error"})
				return
			}
			objectName := id + ext
			if _, err := st.Upload(c.Request.Context(), objectName, f, file.Size, mimeType); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			url, err := st.GetURL(c.Request.Context(), objectName, time.Hour)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
			ftype := mimeType
			size := file.Size
			msg = models.OrderMessage{
				ChatID:   chat.ID,
				ClientID: clientID,
				Type:     models.MessageTypeFile,
				Content:  file.Filename,
				FileURL:  &url,
				FileType: &ftype,
				FileSize: &size,
			}
		} else {
			var r OrderMessageRequest
			if err := c.BindJSON(&r); err != nil || r.Content == "" {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json"})
				return
			}
			msg = models.OrderMessage{ChatID: chat.ID, ClientID: clientID, Type: models.MessageTypeText, Content: r.Content}
		}
        if err := db.Create(&msg).Error; err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
            return
        }
        // подставляем имя отправителя
        var snd models.Client
        if err := db.Select("username").Where("id = ?", clientID).First(&snd).Error; err == nil {
            msg.SenderName = snd.Username
        }
        if cache != nil {
            _ = cache.AddMessage(c.Request.Context(), chat.ID, msg)
        }
        orderchat.Broadcast(chat.ID, msg)
        c.JSON(http.StatusOK, msg)
    }
}

// ReadOrderMessage godoc
// @Summary Отметить сообщение прочитанным
// @Tags orders
// @Security BearerAuth
// @Produce json
// @Accept json
// @Param id path string true "ID ордера"
// @Param msgId path string true "ID сообщения"
// @Param input body handlers.ReadOrderMessageRequest false "опционально: время прочтения в RFC3339"
// @Success 200 {object} models.OrderMessage
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /orders/{id}/messages/{msgId}/read [patch]
func ReadOrderMessage(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
		orderID := c.Param("id")
		msgID := c.Param("msgId")
		clientIDVal, ok := c.Get("client_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "no client"})
			return
		}
		clientID := clientIDVal.(string)
		var order models.Order
		if err := db.Where("id = ?", orderID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid order"})
			return
		}
		if clientID != order.BuyerID && clientID != order.SellerID {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		var chat models.OrderChat
		if err := db.Where("order_id = ?", order.ID).First(&chat).Error; err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid message"})
			return
		}
        var msg models.OrderMessage
        if err := db.Where("id = ? AND chat_id = ?", msgID, chat.ID).First(&msg).Error; err != nil {
            c.JSON(http.StatusNotFound, ErrorResponse{Error: "invalid message"})
            return
        }
        // берём readAt из тела запроса, если передано, иначе используем текущее время
        var r ReadOrderMessageRequest
        _ = c.BindJSON(&r)
        when := time.Now()
        if r.ReadAt != nil {
            when = *r.ReadAt
        }
        if err := db.Model(&msg).Update("read_at", when).Error; err != nil {
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "db error"})
            return
        }
        msg.ReadAt = &when
        // подставляем имя отправителя в ответ
        var snd models.Client
        if err := db.Select("username").Where("id = ?", msg.ClientID).First(&snd).Error; err == nil {
            msg.SenderName = snd.Username
        }
        // уведомляем по WS о прочтении
        orderchat.BroadcastRead(chat.ID, msg)
        c.JSON(http.StatusOK, msg)
    }
}
