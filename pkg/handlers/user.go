package handlers

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/dal"
	"github.com/daheishandemao/Tiktok-E-commerce/pkg/middleware"
	"github.com/golang-jwt/jwt/v5" // 修正导入路径
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 请求体结构
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=4,max=20"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=6"`
}

// Register 用户注册
func Register(ctx context.Context, c *app.RequestContext) {
	var req RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, map[string]interface{}{
			"error":   "参数绑定失败",
			"details": err.Error(), // 显示具体错误信息
		})
		return
	}
	// 增加空值防御
	if len(req.Username) == 0 || len(req.Password) == 0 {
		c.JSON(400, map[string]string{"error": "用户名和密码不能为空",
			"req.Usernam": req.Username, "req.Passwor": req.Password})
		return
	}
	// 手动校验
	if len(req.Username) < 4 || len(req.Password) < 6 {
		c.JSON(400, map[string]string{"error": "用户名需4-20字符，密码需6-32字符"})
		return
	}

	// 检查用户名是否存在
	var existUser dal.User
	err := dal.DB.Where("username = ?", req.Username).First(&existUser).Error
	if err == nil {
		c.JSON(409, map[string]string{"error": "用户名已存在"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(500, map[string]string{"error": "数据库查询失败"})
		return
	}

	// 密码加密存储
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // 默认cost=10，约100ms计算时间
	if err != nil {
		c.JSON(500, map[string]string{"error": "密码加密失败"})
		return
	}

	// 创建用户
	now := time.Now() // 获取当前时间
	newUser := dal.User{
		Username:  req.Username,
		Password:  string(hashedPassword),
		LastLogin: &now, // 显式设置有效时间
	}
	if err := dal.DB.Create(&newUser).Error; err != nil {
		c.JSON(500, map[string]interface{}{"error": "用户创建失败", "details": err.Error()}) // 显示具体错误信息})
		return
	}

	// 生成JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": newUser.ID,
		"exp":    time.Now().Add(2 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		c.JSON(500, map[string]string{"error": "令牌生成失败"})
		return
	}

	c.JSON(200, map[string]interface{}{
		"user_id": newUser.ID,
		"token":   tokenString,
	})
}

// Login 用户登录
func Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(400, map[string]interface{}{
			"error":   "参数校验失败",
			"details": err.Error(),
		})
		return
	}
	// 增加空值防御
	if len(req.Username) == 0 || len(req.Password) == 0 {
		c.JSON(400, map[string]string{"error": "用户名和密码不能为空"})
		return
	}

	// 查询用户
	var user dal.User
	if err := dal.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(401, map[string]string{"error": "用户名或密码错误"})
		} else {
			c.JSON(500, map[string]string{"error": "数据库查询失败"})
		}
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(401, map[string]string{"error": "用户名或密码错误"})
		return
	}

	// 生成新令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(2 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		c.JSON(500, map[string]string{"error": "令牌生成失败"})
		return
	}

	c.JSON(200, map[string]interface{}{
		"user_id": user.ID,
		"token":   tokenString,
	})
}

func GetUserInfo(_ context.Context, c *app.RequestContext) {
	// 从中间件获取注入的userID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, map[string]string{"error": "用户未认证"})
		return
	}

	// 转换为uint类型
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(500, map[string]string{"error": "用户ID类型错误"})
		return
	}

	// 查询真实数据库
	var user dal.User
	if err := dal.DB.First(&user, uid).Error; err != nil {
		c.JSON(404, map[string]string{"error": "用户不存在"})
		return
	}

	// 示例数据返回（需替换为真实数据库查询）
	c.JSON(200, map[string]interface{}{
		"user_id":    user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}
