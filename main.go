package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"gin-examples/project/handlers"
	"gin-examples/project/middleware"
	"gin-examples/project/models"
	"gin-examples/project/services"
)

func main() {

	// 打开数据库
	db, err := gorm.Open(sqlite.Open("userPost.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化服务
	userService := services.NewUserService(db)
	userHandler := handlers.NewUserHandler(userService, "secret")

	// 创建 Gin 引擎
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 公开路由
	public := r.Group("/api/v1")
	{
		//创建用户
		public.POST("/users/register", userHandler.Register) //pass
		//登录
		public.POST("/users/login", userHandler.Login) //pass
		//支持获取所有文章列表和单个文章的详细信息
		public.GET("/users/findPost/:id", userHandler.FindPost) //pass
		public.GET("/users/findPost", userHandler.FindPostList) //pass
		//支持获取某篇文章的所有评论列表
		public.GET("/users/findCommontsByPostId/:id", userHandler.FindCommuntsByPostId)

	}

	// 需要认证的路由
	protected := r.Group("/api/private/v1")
	protected.Use(middleware.Auth("secret"))
	{
		//发布文章
		protected.POST("/users/create-post", userHandler.CreatePost) //pass
		//发布评论
		protected.POST("/users/create-comment", userHandler.CreateComment) //pass
		//只有文章的作者才能更新自己的文章
		protected.PUT("/users/updatePost", userHandler.UpdatePost) //pass
		//只有文章的作者才能删除自己的文章
		protected.DELETE("/users/delete/:id", userHandler.Delete) //pass

	}

	// 启动服务器
	r.Run(":8080")
}
