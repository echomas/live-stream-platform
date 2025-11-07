package handler

import (
	"context"
	commonPb "live-stream-platform/gen/proto/common"
	userPb "live-stream-platform/gen/proto/user"
	"live-stream-platform/services/user-service/internal/service"
)

type UserHandler struct {
	userPb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(ctx context.Context, req *userPb.RegisterRequest) (*userPb.RegisterResponse, error) {
	userID, err := h.userService.Register(ctx, req)
	if err != nil {
		return &userPb.RegisterResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &userPb.RegisterResponse{
		Code:    0,
		Message: "success",
		UserId:  userID,
	}, nil
}

// Login 用户登录
func (h *UserHandler) Login(ctx context.Context, req *userPb.LoginRequest) (*userPb.LoginResponse, error) {
	token, user, err := h.userService.Login(ctx, req)
	if err != nil {
		return &userPb.LoginResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &userPb.LoginResponse{
		Code:    0,
		Message: "success",
		Token:   token,
		User:    user,
	}, nil
}

// Logout 用户登出
func (h *UserHandler) Logout(ctx context.Context, req *userPb.LogoutRequest) (*commonPb.Response, error) {
	err := h.userService.Logout(ctx, req.UserId, req.Token)
	if err != nil {
		return &commonPb.Response{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &commonPb.Response{
		Code:    0,
		Message: "success",
	}, nil
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(ctx context.Context, req *userPb.GetUserInfoRequest) (*userPb.GetUserInfoResponse, error) {
	user, err := h.userService.GetUserInfo(ctx, req.UserId)
	if err != nil {
		return &userPb.GetUserInfoResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &userPb.GetUserInfoResponse{
		Code:    0,
		Message: "success",
		User:    user,
	}, nil
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(ctx context.Context, req *userPb.UpdateUserInfoRequest) (*commonPb.Response, error) {
	err := h.userService.UpdateUserInfo(ctx, req)
	if err != nil {
		return &commonPb.Response{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &commonPb.Response{
		Code:    0,
		Message: "success",
	}, nil
}

// VerifyToken 验证 Token
func (h *UserHandler) VerifyToken(ctx context.Context, req *userPb.VerifyTokenRequest) (*userPb.VerifyTokenResponse, error) {
	claims, err := h.userService.VerifyToken(ctx, req.Token)
	if err != nil {
		return &userPb.VerifyTokenResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &userPb.VerifyTokenResponse{
		Code:     0,
		Message:  "success",
		UserId:   claims.UserID,
		Username: claims.Username,
	}, nil
}

// GetUsersByIds 批量获取用户消息
func (h *UserHandler) GetUsersByIds(ctx context.Context, req *userPb.GetUsersByIdsRequest) (*userPb.GetUsersByIdsResponse, error) {
	users, err := h.userService.GetUsersByIds(ctx, req.UserIds)
	if err != nil {
		return &userPb.GetUsersByIdsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &userPb.GetUsersByIdsResponse{
		Code:    0,
		Message: "success",
		Users:   users,
	}, nil
}

// Health 健康检查
func (h *UserHandler) Health(ctx context.Context, req *userPb.HealthRequest) (*userPb.HealthResponse, error) {
	return &userPb.HealthResponse{
		Status: "ok",
	}, nil
}
