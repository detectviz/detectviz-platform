package mysql

import (
	"context"
	"database/sql"

	"detectviz-platform/pkg/domain/entities"
	"detectviz-platform/pkg/domain/interfaces"
	"detectviz-platform/pkg/domain/valueobjects"
	"detectviz-platform/pkg/platform/contracts"
)

// UserRepository 實現了 interfaces.UserRepository 介面
// 職責: 提供用戶實體的 MySQL 數據庫操作
type UserRepository struct {
	db     *sql.DB
	logger contracts.Logger
}

// NewUserRepository 創建新的用戶倉儲實例
func NewUserRepository(db *sql.DB, logger contracts.Logger) interfaces.UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create 創建新用戶
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `INSERT INTO users (id, email, password_hash, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		r.logger.Error("創建用戶失敗", "user_id", user.ID, "error", err)
		return err
	}

	r.logger.Debug("用戶創建成功", "user_id", user.ID)
	return nil
}

// GetByID 根據 ID 查找用戶
func (r *UserRepository) GetByID(ctx context.Context, id valueobjects.IDVO) (*entities.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?`

	var user entities.User
	row := r.db.QueryRowContext(ctx, query, id.String())
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("用戶未找到", "user_id", id.String())
			return nil, nil
		}
		r.logger.Error("查找用戶失敗", "user_id", id.String(), "error", err)
		return nil, err
	}

	r.logger.Debug("用戶查找成功", "user_id", id.String())
	return &user, nil
}

// GetByEmail 根據郵箱查找用戶
func (r *UserRepository) GetByEmail(ctx context.Context, email valueobjects.EmailVO) (*entities.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?`

	var user entities.User
	row := r.db.QueryRowContext(ctx, query, email.String())
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("用戶未找到", "email", email.String())
			return nil, nil
		}
		r.logger.Error("根據郵箱查找用戶失敗", "email", email.String(), "error", err)
		return nil, err
	}

	r.logger.Debug("根據郵箱查找用戶成功", "email", email.String())
	return &user, nil
}

// Update 更新用戶信息
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `UPDATE users SET email = ?, password_hash = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, user.Email, user.PasswordHash, user.UpdatedAt, user.ID)
	if err != nil {
		r.logger.Error("更新用戶失敗", "user_id", user.ID, "error", err)
		return err
	}

	r.logger.Debug("用戶更新成功", "user_id", user.ID)
	return nil
}

// Delete 刪除用戶
func (r *UserRepository) Delete(ctx context.Context, id valueobjects.IDVO) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		r.logger.Error("刪除用戶失敗", "user_id", id.String(), "error", err)
		return err
	}

	r.logger.Debug("用戶刪除成功", "user_id", id.String())
	return nil
}

// List 列出所有用戶
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users`
	args := []interface{}{}

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}
	if offset > 0 {
		query += ` OFFSET ?`
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("列出用戶失敗", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
			r.logger.Error("掃描用戶記錄失敗", "error", err)
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("遍歷用戶記錄失敗", "error", err)
		return nil, err
	}

	r.logger.Debug("列出用戶成功", "count", len(users))
	return users, nil
}
