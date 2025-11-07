package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	userPb "live-stream-platform/gen/proto/user"
	"live-stream-platform/pkg/config"
	"live-stream-platform/pkg/database"
	"live-stream-platform/pkg/jwt"
	pkgRedis "live-stream-platform/pkg/redis"
	"live-stream-platform/services/user-service/internal/handler"
	"live-stream-platform/services/user-service/internal/repository"
	"live-stream-platform/services/user-service/internal/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting User Service...")
	//1. 加载配置
	cfg := config.Load()

	// 2. 初始化数据库
	dbConfig := &config.DatabaseConfig{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Database:     getEnv("DB_NAME", cfg.Database.Database),
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxLifetime:  time.Hour,
	}

	if err := database.Init(dbConfig); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()
	log.Println("Database initialized")

	if err := pkgRedis.Init(&cfg.Redis); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer pkgRedis.Close()
	log.Println("Redis initialized")
	//4. 初始化 JWT
	jwt.Init(cfg.JWT.Secret)
	log.Println("JWT initialized")
	// 5. 创建依赖实例
	userRepo := repository.NewUserRepository(database.DB)
	//service 层
	userService := service.NewUserService(userRepo, pkgRedis.GetClient(), cfg.JWT.ExpireHours)
	//Handler 层
	userHandler := handler.NewUserHandler(userService)
	log.Println("User service initialized")
	// 6. 创建 gRPC 服务器
	list, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), //4MB
		grpc.MaxSendMsgSize(4*1024*1024), //4MB
	)
	// 7. 注册服务
	userPb.RegisterUserServiceServer(grpcServer, userHandler)
	//8. 启动 gRPC 反射 （用于调试）
	reflection.Register(grpcServer)
	// 9. 启动服务
	go func() {
		log.Printf("✓ User service listening on port %s", cfg.Server.Port)
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("🚀 User Service Started Successfully!")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	// 10.优雅关停
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down User Service...")
	grpcServer.GracefulStop()
	log.Println("User Service stopped")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
