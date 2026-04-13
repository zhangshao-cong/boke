package services

import (
	"errors"

	"gorm.io/gorm"

	"gin-examples/project/models"
	"gin-examples/project/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(req models.User) (*models.User, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, utils.NewAppError(409, "Username already exists")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, utils.NewAppError(409, "Email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	req.Password = string(hashedPassword)

	if err := s.db.Create(&req).Error; err != nil {
		return nil, err
	}

	return &req, nil
}

func (s *UserService) CreatePost(req models.Post) (*models.Post, error) {
	if err := s.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *UserService) FindPost(id int) (*models.Post, error) {
	var post models.Post
	if err := s.db.First(&post, id).Find(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *UserService) FindCommentsByPostId(id int) ([]*models.Comment, error) {
	var communts []*models.Comment
	if err := s.db.Model(&models.Comment{}).Where("post_id = ?", id).Find(&communts).Error; err != nil {
		return nil, err
	}
	return communts, nil
}

func (s *UserService) SearchPostList() ([]*models.Post, error) {
	var post []*models.Post
	if err := s.db.Model(&models.Post{}).Find(&post).Error; err != nil {
		return nil, err
	}

	return post, nil
}

func (s *UserService) UpdatePost(req models.UpdetePostRequest) (*models.Post, error) {
	var post models.Post
	if err := s.db.Model(&post).Where("id = ?", req.PostId).
		Updates(map[string]any{"title": req.Title, "content": req.Content}).
		Error; err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *UserService) Delete(id int) (*models.Post, error) {
	var post models.Post
	if err := s.db.Delete(&post, id).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *UserService) CreateComment(req models.Comment) (*models.Comment, error) {
	if err := s.db.First(&req.Post, req.PostID).Error; err != nil {
		return nil, err
	}
	if err := s.db.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(401, "Invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, utils.NewAppError(401, "Invalid credentials")
	}

	return &user, nil
}

//先不做分页
// func paginate(page, size int) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		// Validate and normalize page number
// 		if page <= 0 {
// 			page = 1
// 		}
// 		// Validate and normalize page size (max 100, min 10)
// 		switch {
// 		case size > 100:
// 			size = 100
// 		case size <= 0:
// 			size = 10
// 		}
// 		// Calculate offset: (page - 1) * size
// 		// Example: page 1, size 10 -> offset 0
// 		//          page 2, size 10 -> offset 10
// 		offset := (page - 1) * size
// 		// Offset: Skip N records
// 		// Limit: Return at most N records
// 		return db.Offset(offset).Limit(size)
// 	}
// }
