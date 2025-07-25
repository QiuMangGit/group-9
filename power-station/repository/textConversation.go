package repository

import (
	"time"
)

type TextConversation struct {
    RobotId      int      `json:"robot_id"`
    ContentId   int      `json:"content_id"`
    Content string   `json:"content"`
    From    string   `json:"from"`
    Time    time.Time `json:"time" gorm:"autoCreateTime"`
}

