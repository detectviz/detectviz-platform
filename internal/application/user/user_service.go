package user

import (
	"context"
	"fmt"
	"time"

	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/interfaces"
	"detectviz-platform/pkg/domain/valueobjects"
	"detectviz-platform/pkg/platform/contracts"
)

// UserService 實現用戶相關的業務邏輯
// 職責: 協調用戶實體的創建、更新、查詢等業務操作
type UserService struct {
	userRepo interfaces.UserRepository
	logger   contracts.Logger
}

// NewUserService 創建新的用戶服務實例
func NewUserService(userRepo interfaces.UserRepository, logger contracts.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// CreateUser 創建新用戶
func (s *UserService) CreateUser(ctx context.Context, email, password string) (*entities.User, error) {
	s.logger.Info("開始創建用戶", "email", email)

	// 檢查用戶是否已存在
	emailVO, err := valueobjects.NewEmailVO(email)
	if err != nil {
		s.logger.Error("無效的郵箱格式", "email", email, "error", err)
		return nil, fmt.Errorf("無效的郵箱格式: %w", err)
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		s.logger.Error("檢查用戶存在性失敗", "email", email, "error", err)
		return nil, fmt.Errorf("檢查用戶存在性失敗: %w", err)
	}

	if existingUser != nil {
		s.logger.Warn("用戶已存在", "email", email)
		return nil, fmt.Errorf("用戶已存在: %s", email)
	}

	// 創建新用戶實體
	user := &entities.User{
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 設置密碼哈希
	if err := user.SetPassword(password); err != nil {
		s.logger.Error("設置用戶密碼失敗", "email", email, "error", err)
		return nil, fmt.Errorf("設置用戶密碼失敗: %w", err)
	}

	// 保存用戶
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("保存用戶失敗", "email", email, "error", err)
		return nil, fmt.Errorf("保存用戶失敗: %w", err)
	}

	s.logger.Info("用戶創建成功", "user_id", user.ID, "email", email)
	return user, nil
}

// GetUser 根據 ID 獲取用戶
func (s *UserService) GetUser(ctx context.Context, id string) (*entities.User, error) {
	s.logger.Debug("查找用戶", "user_id", id)

	idVO, err := valueobjects.NewIDVO(id)
	if err != nil {
		s.logger.Error("無效的用戶ID格式", "user_id", id, "error", err)
		return nil, fmt.Errorf("無效的用戶ID格式: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, idVO)
	if err != nil {
		s.logger.Error("查找用戶失敗", "user_id", id, "error", err)
		return nil, fmt.Errorf("查找用戶失敗: %w", err)
	}

	if user == nil {
		s.logger.Debug("用戶未找到", "user_id", id)
		return nil, fmt.Errorf("用戶未找到: %s", id)
	}

	s.logger.Debug("用戶查找成功", "user_id", id)
	return user, nil
}

// GetUserByEmail 根據郵箱獲取用戶
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	s.logger.Debug("根據郵箱查找用戶", "email", email)

	emailVO, err := valueobjects.NewEmailVO(email)
	if err != nil {
		s.logger.Error("無效的郵箱格式", "email", email, "error", err)
		return nil, fmt.Errorf("無效的郵箱格式: %w", err)
	}

	user, err := s.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		s.logger.Error("根據郵箱查找用戶失敗", "email", email, "error", err)
		return nil, fmt.Errorf("根據郵箱查找用戶失敗: %w", err)
	}

	if user == nil {
		s.logger.Debug("用戶未找到", "email", email)
		return nil, fmt.Errorf("用戶未找到: %s", email)
	}

	s.logger.Debug("根據郵箱查找用戶成功", "email", email)
	return user, nil
}

// UpdateUser 更新用戶信息
func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) error {
	s.logger.Info("更新用戶信息", "user_id", user.ID)

	// 檢查用戶是否存在
	idVO, err := valueobjects.NewIDVO(user.ID)
	if err != nil {
		s.logger.Error("無效的用戶ID格式", "user_id", user.ID, "error", err)
		return fmt.Errorf("無效的用戶ID格式: %w", err)
	}

	existingUser, err := s.userRepo.GetByID(ctx, idVO)
	if err != nil {
		s.logger.Error("檢查用戶存在性失敗", "user_id", user.ID, "error", err)
		return fmt.Errorf("檢查用戶存在性失敗: %w", err)
	}

	if existingUser == nil {
		s.logger.Warn("用戶不存在", "user_id", user.ID)
		return fmt.Errorf("用戶不存在: %s", user.ID)
	}

	// 更新時間戳
	user.UpdatedAt = time.Now()

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("更新用戶失敗", "user_id", user.ID, "error", err)
		return fmt.Errorf("更新用戶失敗: %w", err)
	}

	s.logger.Info("用戶更新成功", "user_id", user.ID)
	return nil
}

// DeleteUser 刪除用戶
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("刪除用戶", "user_id", id)

	idVO, err := valueobjects.NewIDVO(id)
	if err != nil {
		s.logger.Error("無效的用戶ID格式", "user_id", id, "error", err)
		return fmt.Errorf("無效的用戶ID格式: %w", err)
	}

	// 檢查用戶是否存在
	existingUser, err := s.userRepo.GetByID(ctx, idVO)
	if err != nil {
		s.logger.Error("檢查用戶存在性失敗", "user_id", id, "error", err)
		return fmt.Errorf("檢查用戶存在性失敗: %w", err)
	}

	if existingUser == nil {
		s.logger.Warn("用戶不存在", "user_id", id)
		return fmt.Errorf("用戶不存在: %s", id)
	}

	// 刪除用戶
	if err := s.userRepo.Delete(ctx, idVO); err != nil {
		s.logger.Error("刪除用戶失敗", "user_id", id, "error", err)
		return fmt.Errorf("刪除用戶失敗: %w", err)
	}

	s.logger.Info("用戶刪除成功", "user_id", id)
	return nil
}

// ListUsers 列出用戶
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	s.logger.Debug("列出用戶", "limit", limit, "offset", offset)

	users, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		s.logger.Error("列出用戶失敗", "error", err)
		return nil, fmt.Errorf("列出用戶失敗: %w", err)
	}

	s.logger.Debug("列出用戶成功", "count", len(users))
	return users, nil
}

// AuthenticateUser 驗證用戶身份
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*entities.User, error) {
	s.logger.Debug("驗證用戶身份", "email", email)

	emailVO, err := valueobjects.NewEmailVO(email)
	if err != nil {
		s.logger.Error("無效的郵箱格式", "email", email, "error", err)
		return nil, fmt.Errorf("無效的郵箱格式: %w", err)
	}

	// 根據郵箱查找用戶
	user, err := s.userRepo.GetByEmail(ctx, emailVO)
	if err != nil {
		s.logger.Error("查找用戶失敗", "email", email, "error", err)
		return nil, fmt.Errorf("查找用戶失敗: %w", err)
	}

	if user == nil {
		s.logger.Warn("用戶不存在", "email", email)
		return nil, fmt.Errorf("用戶不存在或密碼錯誤")
	}

	// 驗證密碼
	if !user.CheckPassword(password) {
		s.logger.Warn("密碼驗證失敗", "email", email)
		return nil, fmt.Errorf("用戶不存在或密碼錯誤")
	}

	s.logger.Info("用戶身份驗證成功", "user_id", user.ID, "email", email)
	return user, nil
}
