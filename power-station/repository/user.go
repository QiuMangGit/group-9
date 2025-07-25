package repository

import (
	"crypto/rand"
	"group-9/model"
	"log"
	"net/http"

	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mail.v2"
)

const (
	FromEmail  = "3186965058@qq.com" // 发件人邮箱（如 QQ 邮箱）
	EmailPass  = "wvcfqxenodtrddeh"  // 邮箱 SMTP 授权码
	SMTPServer = "smtp.qq.com"       // 邮箱服务器（QQ 邮箱用 smtp.qq.com）
	SMTPPort   = 587                 // 端口（QQ 邮箱 587 或 465，看配置）
)

type UserHandler struct {
	repo *Repository
}

func NewUserHandler(repo *Repository) *UserHandler {
	return &UserHandler{repo: repo}
}

var VerificationCode string

// 生成随机验证码（6位数字）
func generateVerificationCode() string {
	// 创建一个字节切片来存储随机数
	b := make([]byte, 3)

	// 使用crypto/rand生成强随机数
	if _, err := rand.Read(b); err != nil {
		panic("生成随机数失败: " + err.Error())
	}

	return fmt.Sprintf("%02d%02d%02d", int(b[0])%100, int(b[1])%100, int(b[2])%100)
}

func (h *UserHandler) SendEmailCode(c *gin.Context) {
	// 1. 绑定请求参数（只需要邮箱地址，不需要主题和正文）
	type EmailReq struct {
		EmailTo string `json:"emailTo" binding:"required,email"` // 收件人邮箱
	}
	var req EmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 0, Msg: "参数错误：" + err.Error()})
		return
	}

	// 2. 生成随机验证码
	code := generateVerificationCode()

	VerificationCode = code

	// 3. 构建邮件内容（主题和正文由后端固定，包含验证码）
	subject := "【系统】您的验证码"
	body := fmt.Sprintf(`您的验证码是：%s该验证码5分钟内有效，请不要泄露给他人。`, code)

	err := h.sendEmail(req.EmailTo, subject, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 0, Msg: "邮箱发送失败：" + err.Error()})
		return
	}

	// 5. 将验证码存入Redis（有效期5分钟）
	// 注意：实际项目中需要实现这部分，用于后续验证
	// redisKey := fmt.Sprintf("email_code:%s", req.EmailTo)
	// redisClient.Set(redisKey, code, 5*time.Minute)

	c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "验证码已发送，请注意查收"})
}

func (h *UserHandler) sendEmail(to, subject, body string) error {
	// 1. 构建邮件内容
	m := mail.NewMessage()
	m.SetHeader("From", FromEmail)  // 发件人
	m.SetHeader("To", to)           // 收件人
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/plain", body)   // 正文（纯文本，也可改成 text/html 发 HTML 邮件）

	// 2. 配置 SMTP 客户端
	d := mail.NewDialer(SMTPServer, SMTPPort, FromEmail, EmailPass)
	// 跳过 TLS 验证（如果是正式环境，建议关闭 InsecureSkipVerify，用正规证书）
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 3. 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username         string `json:"username" binding:"required"`
		Password         string `json:"password" binding:"required,min=6"`
		Email            string `json:"email" binding:"required,email"`
		VerificationCode string `json:"verificationCode" `
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "参数错误: " + err.Error()})
		return
	}

	

	// 检查用户是否已存在（同时检查用户名和邮箱）
	var count int64
	if err := h.repo.db.Model(&model.User{}).
		Where("email = ?",req.Email).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "数据库查询失败"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "邮箱已经存在,请直接登录"})
		return
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "密码加密失败"})
		return
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
	}

	if err := h.repo.db.Create(user).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "注册失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "注册成功"})
}

// 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "参数错误: " + err.Error()})
		return
	}

	var user model.User
	if err := h.repo.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "用户名或密码错误11"})
		return
	}
	log.Println(user.Email)
	c.JSON(http.StatusOK, model.Response{Code: 1, Msg: "登录成功", Data: gin.H{"email": user.Email}})
}

// 查看用户详细信息
func (h *UserHandler) GetDetail(c *gin.Context){
        var req struct{
                Username string `json:"username" binding:"required"`
                Password string `json:"password" binding:"required"`
                Email string `json:"email" binding:"required"   `
        }
        // 
        if err := c.ShouldBindJSON(&req); err != nil{
                c.JSON(http.StatusOK,model.Response{Code:0,Msg:"参数不匹配" + err.Error()})
                return
        }
        var user model.User
        if err := h.repo.db.Where("email = ?",req.Email).First(&user).Error; err != nil{
                c.JSON(http.StatusOK,model.Response{Code:0,Msg:"邮箱错误"})
                return
        }

        c.JSON(http.StatusOK,model.Response{Code:1,Msg: "查询成功",Data:gin.H{"data":req}})
}
func (h *UserHandler) SubmitDetail(c *gin.Context){
        var req struct{
                Username string `json:"username" binding:"required"`
                Password string `json:"password" binding:"required"`
                Email string `json:"email" binding:"required"   `
        }
        // 绑定请求
        if err :=c.ShouldBindJSON(&req); err != nil{
                c.JSON(http.StatusOK,model.Response{Code:0,Msg:"参数不匹配" + err.Error()})
        }
        var user model.User
        if err := h.repo.db.Model(&user).Where("email = ?", req.Email).Updates(req).Error; err != nil {
                c.JSON(http.StatusOK, model.Response{Code: 0, Msg: "数据更新失败: " + err.Error()})
                return
        }
        c.JSON(http.StatusOK,model.Response{Code:1,Msg:"更新成功",Data:gin.H{"data":req}})
}
