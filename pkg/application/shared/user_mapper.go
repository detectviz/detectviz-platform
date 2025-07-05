package shared

import (
	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/valueobjects"
)

// UserMapper 提供用戶相關的 DTO 與實體之間的轉換功能
// 職責: 封裝 DTO 與領域實體之間的映射邏輯，確保數據格式的正確轉換
// AI_SCAFFOLD_HINT: 此 Mapper 負責 DTO 與 Entity 的雙向轉換，AI 可自動生成映射邏輯
type UserMapper struct{}

// NewUserMapper 創建新的用戶映射器
func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

// ToEntity 將 CreateUserRequest DTO 轉換為 User 實體
// AI_SCAFFOLD_HINT: 自動處理 DTO 驗證和 ValueObject 創建
func (m *UserMapper) ToEntity(req *CreateUserRequest) (*entities.User, error) {
	// 創建 Email 值對象
	emailVO, err := valueobjects.NewEmailVO(req.Email)
	if err != nil {
		return nil, err
	}

	// 創建用戶實體
	user := &entities.User{
		Email: emailVO.String(),
	}

	// 設置密碼
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	return user, nil
}

// ToResponse 將 User 實體轉換為 UserResponse DTO
// AI_SCAFFOLD_HINT: 自動處理實體字段到 DTO 的映射
func (m *UserMapper) ToResponse(user *entities.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToResponseList 將 User 實體列表轉換為 UserResponse DTO 列表
// AI_SCAFFOLD_HINT: 批量轉換邏輯，可自動生成
func (m *UserMapper) ToResponseList(users []*entities.User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = m.ToResponse(user)
	}
	return responses
}

// UpdateEntityFromDTO 使用 UpdateUserRequest DTO 更新 User 實體
// AI_SCAFFOLD_HINT: 部分更新邏輯，只更新非零值字段
func (m *UserMapper) UpdateEntityFromDTO(user *entities.User, req *UpdateUserRequest) error {
	// 更新 Email（如果提供）
	if req.Email != "" {
		emailVO, err := valueobjects.NewEmailVO(req.Email)
		if err != nil {
			return err
		}
		user.Email = emailVO.String()
	}

	return nil
}

// ValidateCreateRequest 驗證創建用戶請求的業務邏輯
// AI_SCAFFOLD_HINT: 業務規則驗證，可根據領域規則自動生成
func (m *UserMapper) ValidateCreateRequest(req *CreateUserRequest) error {
	// 這裡可以添加額外的業務邏輯驗證
	// 例如：密碼強度檢查、Email 格式特殊要求等

	// 基本驗證已在 DTO 的 validate 標籤中處理
	// 這裡主要處理複雜的業務邏輯驗證

	return nil
}

// ValidateUpdateRequest 驗證更新用戶請求的業務邏輯
// AI_SCAFFOLD_HINT: 更新驗證邏輯，可自動生成
func (m *UserMapper) ValidateUpdateRequest(req *UpdateUserRequest) error {
	// 更新請求的業務邏輯驗證
	return nil
}
