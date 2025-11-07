package repository

import (
	"context"
	"gorm.io/gorm"
	"live-stream-platform/services/user-service/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	GetByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (ur *userRepository) Create(ctx context.Context, user *model.User) error {
	return ur.db.WithContext(ctx).Create(user).Error
}

func (ur *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := ur.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := ur.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := ur.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) Update(ctx context.Context, user *model.User) error {
	return ur.db.WithContext(ctx).Save(user).Error
}

func (ur *userRepository) GetByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	var users []*model.User
	if err := ur.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ur *userRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	return ur.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}
