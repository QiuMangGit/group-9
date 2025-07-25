package repository

import (
	"github.com/gin-gonic/gin"
	"encoding/json"
	"group-9/model"
	"log"
	"net/http"
	"time"
)

type VoiceConversation struct {
	RobotId   int       `json:"robotId" gorm:"column:Robot_Id"`
	MessageId uint      `gorm:"primaryKey;autoIncrement;column:Message_Id"`
	Context   string    `json:"context" gorm:"column:Context"`
	Time      time.Time `json:"time" gorm:"autoCreateTime;column:Time"`
	From      string    `json:"from" gorm:"column:From"`
}

type VoiceConversationHandler struct {
	repo *Repository
}

func NewConversationHandler(repo *Repository) *VoiceConversationHandler {
	return &VoiceConversationHandler{repo: repo}
}


func (h *VoiceConversationHandler) DeleteVoiceConversation(c *gin.Context) {
	var voiceConversation VoiceConversation
	if err := c.ShouldBindJSON(&voiceConversation); err != nil {
		log.Printf("Invalid parameters: %v", err)
		c.JSON(http.StatusBadRequest, model.Response{Code: 0, Msg: "Invalid parameters"})
		return
	}
	if err := h.repo.DeleteVoiceConversation(voiceConversation.RobotId); err != nil {
		log.Printf("Failed to delete conversation: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{Code: 0, Msg: "Delete failed"})
		return
	}
	// 添加成功响应
	c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Delete successful"})
}

var RobotId int
func (h *VoiceConversationHandler) GetVoiceConversation(c *gin.Context) {
	var voiceConversation VoiceConversation
	if err := c.ShouldBindJSON(&voiceConversation); err != nil {
		log.Printf("Invalid parameters: %v", err)
		c.JSON(http.StatusBadRequest, model.Response{Code: 0, Msg: "Invalid parameters"})
		return
	}
	RobotId = voiceConversation.RobotId
	voiceConversations, err := h.repo.GetAllVoiceConversations(&voiceConversation.RobotId)
	if err != nil {
		log.Printf("Failed to get conversations: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{Code: 0, Msg: "Get failed"})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Get successful", Data: voiceConversations})
}

func (r *Repository) AddVoiceConversation(conversation *VoiceConversation) error {
	return r.db.Create(conversation).Error
}

func (r *Repository) DeleteVoiceConversation(id int) error {
	return r.db.Delete(&VoiceConversation{}, "robot_id = ?", id).Error
}

func (r *Repository) GetAllVoiceConversations(id *int) ([]VoiceConversation, error) {
	var conversations []VoiceConversation

	query := r.db.Order("time ASC")

	if id != nil {
		query = query.Where("robot_id = ?", *id)
	}

	if err := query.Find(&conversations).Error; err != nil {
		return nil, err
	}

	return conversations, nil
}

// MarshalJSON 自定义 JSON 序列化格式
func (v VoiceConversation) MarshalJSON() ([]byte, error) {
	type Alias VoiceConversation // 使用别名避免递归调用
	return json.Marshal(&struct {
		Time string `json:"time"`
		*Alias
	}{
		Time:  v.Time.Format("2006/1/2 15:04:05"), // 格式化为 2025/7/22 14:57:04
		Alias: (*Alias)(&v),
	})
}