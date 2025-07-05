package http_handlers

import (
	"net/http"
	"strconv"

	"detectviz-platform/internal/application/user"
	"detectviz-platform/pkg/application/shared"
	"detectviz-platform/pkg/platform/contracts"

	"github.com/labstack/echo/v4"
)

// UserHandler 處理用戶相關的 HTTP 請求
// 職責: 將 HTTP 請求轉換為服務層調用，並將結果返回給客戶端
type UserHandler struct {
	userService *user.UserService
	userMapper  *shared.UserMapper
	logger      contracts.Logger
}

// NewUserHandler 創建新的用戶處理器
func NewUserHandler(userService *user.UserService, logger contracts.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		userMapper:  shared.NewUserMapper(),
		logger:      logger,
	}
}

// CreateUser 創建新用戶
func (h *UserHandler) CreateUser(c echo.Context) error {
	var req shared.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// 驗證請求數據
	if err := h.userMapper.ValidateCreateRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 調用應用服務創建用戶（使用原始參數）
	user, err := h.userService.CreateUser(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 使用 Mapper 轉換為響應 DTO
	response := h.userMapper.ToResponse(user)
	return c.JSON(http.StatusCreated, response)
}

// GetUser 獲取用戶信息
func (h *UserHandler) GetUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "User ID is required",
		})
	}

	user, err := h.userService.GetUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	// 使用 Mapper 轉換為響應 DTO
	response := h.userMapper.ToResponse(user)
	return c.JSON(http.StatusOK, response)
}

// GetUsers 獲取用戶列表
func (h *UserHandler) GetUsers(c echo.Context) error {
	// 解析查詢參數
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	// 計算 offset
	offset := (page - 1) * limit

	users, err := h.userService.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 使用 Mapper 轉換為響應 DTO 列表
	responses := h.userMapper.ToResponseList(users)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": responses,
		"page":  page,
		"limit": limit,
		"total": len(responses),
	})
}

// UpdateUser 更新用戶信息
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "User ID is required",
		})
	}

	var req shared.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// 驗證請求數據
	if err := h.userMapper.ValidateUpdateRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 先獲取現有用戶
	existingUser, err := h.userService.GetUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	// 使用 Mapper 更新實體
	if err := h.userMapper.UpdateEntityFromDTO(existingUser, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// 調用服務更新
	if err := h.userService.UpdateUser(c.Request().Context(), existingUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 使用 Mapper 轉換為響應 DTO
	response := h.userMapper.ToResponse(existingUser)
	return c.JSON(http.StatusOK, response)
}

// DeleteUser 刪除用戶
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "User ID is required",
		})
	}

	err := h.userService.DeleteUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, nil)
}

// AuthenticateUser 用戶認證
func (h *UserHandler) AuthenticateUser(c echo.Context) error {
	var req shared.AuthenticateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("綁定認證請求失敗", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "無效的請求格式",
		})
	}

	user, err := h.userService.AuthenticateUser(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("用戶認證失敗", "email", req.Email, "error", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "用戶名或密碼錯誤",
		})
	}

	// 使用 Mapper 轉換為響應 DTO
	response := h.userMapper.ToResponse(user)

	h.logger.Info("用戶認證成功", "user_id", user.ID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "認證成功",
		"user":    response,
	})
}
