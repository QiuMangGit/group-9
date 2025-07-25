package repository

import (
	"group-9/model"
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
)



type RobotHandler struct {
    repo *Repository
}

func NewRobotHandler(repo *Repository) *RobotHandler {
    return &RobotHandler{repo: repo}
}

func (h *RobotHandler) AddAssistant(c *gin.Context) {
    var robot model.PostAddRobot
    if err := c.ShouldBindJSON(&robot); err != nil {
        log.Printf("Parameter error: %v", err)
        c.Status(http.StatusBadRequest)
        return
    }
    log.Printf("Request body: %+v", robot)
    var robotSql model.ChatRobotSql
    robotSql.Email = robot.Email
    robotSql.Name = robot.Name
    robotSql.Description = robot.Description
    if err := h.repo.CreateRobot(&robotSql); err != nil {
        c.Status(http.StatusInternalServerError)
        return
    }
    c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Created successfully"})
}



func (h *RobotHandler) DeleteAssistant(c *gin.Context) {
    var robotSql model.ChatRobotSql
    if err := c.ShouldBindJSON(&robotSql); err != nil {
        log.Printf("Parameter error: %v", err)
        c.Status(http.StatusBadRequest)
        return
    }
    log.Println("是这里",robotSql)
    if err := h.repo.DeleteRobot(int(robotSql.Id)); err != nil {
        log.Println("Delete failed: ", err)
        c.Status(http.StatusInternalServerError)
        return
    }
    c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Deleted successfully"})
}

func (h *RobotHandler) UpdateAssistant(c *gin.Context) {
    var robotSql model.ChatRobotSql
    if err := c.ShouldBindJSON(&robotSql); err != nil {
        c.Status(http.StatusBadRequest)
        return
    }
    if err := h.repo.UpdateRobot(&robotSql); err != nil {
        log.Println("Update failed: ", err)
        c.Status(http.StatusInternalServerError)
        return
    }
    c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Updated successfully"})
}

func (h *RobotHandler) GetAssistant(c *gin.Context) {
    var robot model.ChatRobot
    if err := c.ShouldBindJSON(&robot); err != nil {
        log.Printf("Parameter error: %v", err)
        c.Status(http.StatusBadRequest)
        return
    }
    log.Println(robot)
    var robotSql model.ChatRobotSql
    robotSql.Email = robot.Email
    log.Println("111111111111111111111",robotSql.Email)
    robots, err := h.repo.GetRobotsByEmail(robotSql.Email)
    if err != nil {
        log.Println("Get failed: ", err)
        c.Status(http.StatusInternalServerError)
        return
    }
    log.Println("机器人列表",robots)
    c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "Get successfully", Data: robots})
}

func (r *Repository) CreateRobot(robot *model.ChatRobotSql) error {
	return r.db.Create(robot).Error
}

func (r *Repository) DeleteRobot(id int) error {
	return r.db.Delete(&model.ChatRobotSql{}, id).Error
}

func (r *Repository) UpdateRobot(robot *model.ChatRobotSql) error {
	return r.db.Save(robot).Error
}

func (r *Repository) GetRobotsByEmail(email string) ([]model.ChatRobotSql, error) {
	var robots []model.ChatRobotSql
	// 查询邮箱为email的所有机器人
	if err := r.db.Where("Email = ?", email).Find(&robots).Error; err != nil {
		return nil, err
	}
    log.Println("机器人列表2",robots)
	return robots, nil
}