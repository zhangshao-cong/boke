package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gin-examples/project/models"
	"gin-examples/project/services"
	"gin-examples/project/utils"
)

type UserHandler struct {
	userService *services.UserService
	jwtSecret   string
}

func NewUserHandler(userService *services.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

// 创建用户
func (h *UserHandler) Register(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, parseValidationErrors(err))
		return
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.Success(c, "用户创建成功", models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// 查询文章
func (h *UserHandler) FindPost(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Error(c, 401, "请输入文章id")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := h.userService.FindPost(idInt)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.Success(c, "查询成功", post)
}

// 查询文章列表
func (h *UserHandler) FindPostList(c *gin.Context) {
	dataList, err := h.userService.SearchPostList()
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Success(c, "查询成功", dataList)
}

func (h *UserHandler) FindCommuntsByPostId(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Error(c, 401, "请输入文章id")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	communts, err := h.userService.FindCommentsByPostId(idInt)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.Success(c, "查询成功", communts)
}

// 删除文章
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Error(c, 401, "请输入文章id")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := h.userService.FindPost(idInt)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	userId, flag := c.Get("userID")
	if !flag {
		utils.Error(c, 401, "未获取到用户Id")
		return
	}

	if post.UserID != userId {
		utils.Error(c, 401, "只能删自己发布的文章")
		return
	}
	_, errDelete := h.userService.Delete(idInt)
	if errDelete != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Success(c, "删除成功", nil)
}

// 发布文章
func (h *UserHandler) CreatePost(c *gin.Context) {
	var req models.Post
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, parseValidationErrors(err))
		return
	}
	userId, flag := c.Get("userID")
	if !flag {
		utils.Error(c, 401, "未获取到用户Id")
		return
	}
	id, ok := userId.(uint)
	if !ok {
		utils.Error(c, 500, "用户ID类型错误")
		return
	}
	req.UserID = id

	_, err := h.userService.CreatePost(req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Success(c, "文章发布成功", nil)
}

// 发布评论
func (h *UserHandler) CreateComment(c *gin.Context) {
	var req models.Comment
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, parseValidationErrors(err))
		return
	}
	if req.PostID <= 0 {
		utils.Error(c, 401, "请输入文章id")
		return
	}
	userId, flag := c.Get("userID")
	if !flag {
		utils.Error(c, 401, "未获取到用户Id")
		return
	}
	id, ok := userId.(uint)
	if !ok {
		utils.Error(c, 500, "用户ID类型错误")
		return
	}
	req.UserID = id

	_, err := h.userService.CreateComment(req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	utils.Success(c, "评论发布成功", nil)
}

// 更新文章
func (h *UserHandler) UpdatePost(c *gin.Context) {
	var req models.UpdetePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, parseValidationErrors(err))
		return
	}
	if req.PostId == 0 {
		utils.Error(c, 401, "请输入文章id")
		return
	}
	post, err := h.userService.FindPost(int(req.PostId))
	if err != nil {
		utils.HandleError(c, err)
		return
	}
	userId, flag := c.Get("userID")
	if !flag {
		utils.Error(c, 401, "未获取到用户Id")
		return
	}
	if post.UserID != userId {
		utils.Error(c, 401, "只能更新自己发布的文章")
		return
	}
	_, errUp := h.userService.UpdatePost(req)
	if errUp != nil {
		utils.HandleError(c, err)
		return
	}

	utils.Success(c, "文章更新成功", nil)
}

// 登录
func (h *UserHandler) Login(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, parseValidationErrors(err))
		return
	}

	user, err := h.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	token, err := utils.GenerateToken(h.jwtSecret, user.ID, user.Username)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.Success(c, "登录成功", gin.H{
		"token": token,
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

func parseValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	// 简化处理，实际应该解析 binding 错误
	errors["general"] = err.Error()
	return errors
}
