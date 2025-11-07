package service

import (
	"context"
	"errors"
	"fmt"
	"live-stream-platform/services/user-service/internal/model"
	"live-stream-platform/services/user-service/internal/repository"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	commonPb "live-stream-platform/gen/proto/common"
	userPb "live-stream-platform/gen/proto/user"
	"live-stream-platform/pkg/jwt"
	"live-stream-platform/pkg/utils"
)

// UserService 用户服务接口
type UserService interface {
	//Register 用户注册
	Register(ctx context.Context, req *userPb.RegisterRequest) (int64, error)
	// Login 用户登录
	Login(ctx context.Context, req *userPb.LoginRequest) (string, *commonPb.UserInfo, error)
	// Logout 用户登出
	Logout(ctx context.Context, userID int64, token string) error
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID int64) (*commonPb.UserInfo, error)
	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(ctx context.Context, req *userPb.UpdateUserInfoRequest) error
	// VerifyToken 验证 Token
	VerifyToken(ctx context.Context, token string) (*jwt.Claims, error)
	// GetUsersByIds 批量获取用户信息
	GetUsersByIds(ctx context.Context, userIDs []int64) ([]*commonPb.UserInfo, error)
}

// userService 用户服务实现
type userService struct {
	userRepo    repository.UserRepository
	redisClient *redis.Client
	jwtExpire   int
}

func NewUserService(userRepo repository.UserRepository, redisClient *redis.Client, jwtExpire int) UserService {
	return &userService{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtExpire:   jwtExpire,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, req *userPb.RegisterRequest) (int64, error) {
	if !utils.ValidateEmail(req.Email) {
		return 0, errors.New("invalid email")
	}
	if !utils.ValidateUsername(req.Username) {
		return 0, errors.New("invalid username:3-20 characters, alphanumeric and underscore only")
	}
	if !utils.ValidatePassword(req.Password) {
		return 0, errors.New("invalid password: at least 8 characters with uppercase, lowercase and number")
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return 0, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check email: %w", err)
	}

	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return 0, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check username: %w", err)
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}
	user := &model.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Nickname:     req.Nickname,
		Gender:       int(req.Gender),
		Status:       1,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return user.ID, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, req *userPb.LoginRequest) (string, *commonPb.UserInfo, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("username or password incorrect")
		}
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.Status != 1 {
		return "", nil, errors.New("user account is disabled")
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return "", nil, errors.New("username or password incorrect")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtExpire)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}
	tokenKey := fmt.Sprintf("token:%s", token)
	if err := s.redisClient.Set(ctx, tokenKey, user.ID, time.Duration(s.jwtExpire)*time.Hour).Err(); err != nil {
		fmt.Printf("Warning: Failed to cache token: %v\n", err)
	}

	userInfo := &commonPb.UserInfo{
		Id:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Gender:    int32(user.Gender),
		Avatar:    user.Avatar,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt.Unix(),
	}
	return token, userInfo, nil
}

// Logout 用户登出
func (s *userService) Logout(ctx context.Context, userID int64, token string) error {
	// 删除 Redis 中的 Token
	tokenKey := fmt.Sprintf("token:%s", token)
	if err := s.redisClient.Del(ctx, tokenKey).Err(); err != nil {
		fmt.Printf("Warning: Failed to cache token: %v\n", err)
	}
	// 将 token 加入黑名单
	blacklistKey := fmt.Sprintf("token:blacklist:%s", token)
	if err := s.redisClient.Set(ctx, blacklistKey, userID, time.Duration(s.jwtExpire)*time.Hour).Err(); err != nil {
		fmt.Printf("Warning: Failed to cache blacklist: %v\n", err)
	}
	return nil
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, userID int64) (*commonPb.UserInfo, error) {
	// 1. 尝试从 Redis 获取缓存
	// TODO 实现缓存
	//cacheKey := fmt.Sprintf("user:info:%d", userID)
	//res, err := s.redisClient.Get(ctx, cacheKey).Result()
	//if err == nil {
	//
	//}
	// TODO 这里简化了缓存实现，实际应该序列化后存储
	// 暂时直接从数据库中获取
	//2.从数据库中获取
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	// 3. 转换成 protobuf 消息
	userInfo := &commonPb.UserInfo{
		Id:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Gender:    int32(user.Gender),
		Avatar:    user.Avatar,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt.Unix(),
	}
	// 4. 写入缓存 (简化版)
	//TODO 实现完整的缓存逻辑
	return userInfo, nil
}

// UpdateUserInfo 更新用户信息
func (s *userService) UpdateUserInfo(ctx context.Context, req *userPb.UpdateUserInfoRequest) error {
	// 1. 获取用户
	user, err := s.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}
	// 2. 更新字段
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Gender > 0 {
		user.Gender = int(req.Gender)
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	// 3. 保存更新
	if err = s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	// 4. 删除缓存
	cacheKey := fmt.Sprintf("user:info:%d", user.ID)
	s.redisClient.Del(ctx, cacheKey)
	return nil
}

// VerifyToken 验证 Token
func (s *userService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	// 1. 检查 token 是否在黑名单（已登出）
	blacklistKey := fmt.Sprintf("token:blacklist:%s", token)
	exists, err := s.redisClient.Exists(ctx, blacklistKey).Result()
	if err == nil && exists > 0 {
		return nil, errors.New("token has been revoked")
	}
	// 2. 解析和验证 token
	claims, err := jwt.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return claims, nil
}

func (s *userService) GetUsersByIds(ctx context.Context, userIDs []int64) ([]*commonPb.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*commonPb.UserInfo{}, nil
	}
	// 1. 从数据库批量查询
	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	//2. 转换为 protobuf 消息
	userInfos := make([]*commonPb.UserInfo, 0, len(users))
	for _, user := range users {
		userInfo := &commonPb.UserInfo{
			Id:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Gender:    int32(user.Gender),
			Avatar:    user.Avatar,
			Status:    int32(user.Status),
			CreatedAt: user.CreatedAt.Unix(),
		}
		userInfos = append(userInfos, userInfo)
	}
	return userInfos, nil
}
