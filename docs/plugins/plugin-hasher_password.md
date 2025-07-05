# PasswordHasher Plugin

## 概述

PasswordHasher 插件為 Detectviz 平台提供安全的密碼散列和驗證功能。此插件封裝了密碼散列演算法的實現細節，提供統一的介面來處理用戶密碼的安全存儲和驗證。

## 功能特性

- **安全散列**: 使用業界標準的散列演算法（如 bcrypt）對密碼進行不可逆散列
- **密碼驗證**: 提供密碼驗證功能，確保用戶輸入的密碼與存儲的散列值匹配
- **可配置成本**: 支援調整散列演算法的成本參數，平衡安全性和性能
- **上下文感知**: 支援 Go context，可處理請求取消和超時
- **錯誤處理**: 提供詳細的錯誤資訊和適當的錯誤類型

## 支援的散列演算法

### bcrypt
- **推薦使用**: 是目前最廣泛使用和信任的密碼散列演算法
- **自適應成本**: 可調整成本參數以應對計算能力的提升
- **內建鹽值**: 每次散列都會生成唯一的鹽值
- **成本範圍**: 4-31（推薦 10-15）

## 配置說明

### 基本配置

```yaml
password_hasher:
  name: "default_bcrypt_hasher"
  type: "bcrypt"
  config:
    cost: 12
  enabled: true
```

### 高安全性配置

```yaml
password_hasher:
  name: "high_security_bcrypt_hasher"
  type: "bcrypt"
  config:
    cost: 15
    salt_length: 32
  enabled: true
```

### 配置參數

| 參數 | 類型 | 必需 | 默認值 | 說明 |
|------|------|------|--------|------|
| `name` | string | 是 | - | 插件的唯一標識符 |
| `type` | string | 是 | - | 散列演算法類型 (`bcrypt`, `argon2`, `scrypt`) |
| `config.cost` | integer | 否 | 10 | bcrypt 成本參數 (4-31) |
| `config.salt_length` | integer | 否 | 16 | 鹽值長度 (16-32) |
| `enabled` | boolean | 否 | true | 是否啟用此插件 |

## 使用範例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "detectviz-platform/internal/auth/hasher"
)

func main() {
    // 創建密碼散列器
    passwordHasher := hasher.NewDefaultBcryptPasswordHasher()
    ctx := context.Background()
    
    // 散列密碼
    plainPassword := "mySecurePassword123"
    hashedPassword, err := passwordHasher.HashPassword(ctx, plainPassword)
    if err != nil {
        log.Fatalf("Failed to hash password: %v", err)
    }
    
    fmt.Printf("Hashed password: %s\n", hashedPassword)
    
    // 驗證密碼
    isValid, err := passwordHasher.VerifyPassword(ctx, plainPassword, hashedPassword)
    if err != nil {
        log.Fatalf("Failed to verify password: %v", err)
    }
    
    if isValid {
        fmt.Println("Password is valid!")
    } else {
        fmt.Println("Password is invalid!")
    }
}
```

### 自定義成本參數

```go
package main

import (
    "context"
    "log"
    
    "detectviz-platform/internal/auth/hasher"
)

func main() {
    // 創建高安全性密碼散列器
    passwordHasher, err := hasher.NewBcryptPasswordHasher(15)
    if err != nil {
        log.Fatalf("Failed to create password hasher: %v", err)
    }
    
    ctx := context.Background()
    
    // 使用高成本參數散列密碼
    hashedPassword, err := passwordHasher.HashPassword(ctx, "verySecurePassword")
    if err != nil {
        log.Fatalf("Failed to hash password: %v", err)
    }
    
    // 驗證密碼
    isValid, err := passwordHasher.VerifyPassword(ctx, "verySecurePassword", hashedPassword)
    if err != nil {
        log.Fatalf("Failed to verify password: %v", err)
    }
    
    if isValid {
        log.Println("High-security password verified successfully!")
    }
}
```

### 在用戶服務中使用

```go
package userservice

import (
    "context"
    "fmt"
    
    "detectviz-platform/internal/auth/hasher"
    "detectviz-platform/pkg/domain/entities"
)

type UserService struct {
    passwordHasher hasher.PasswordHasher
    userRepo       UserRepository
}

func NewUserService(passwordHasher hasher.PasswordHasher, userRepo UserRepository) *UserService {
    return &UserService{
        passwordHasher: passwordHasher,
        userRepo:       userRepo,
    }
}

func (s *UserService) CreateUser(ctx context.Context, name, email, plainPassword string) (*entities.User, error) {
    // 散列密碼
    hashedPassword, err := s.passwordHasher.HashPassword(ctx, plainPassword)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    
    // 創建用戶實體
    user, err := entities.NewUser(generateID(), name, email, hashedPassword)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // 保存用戶
    if err := s.userRepo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user: %w", err)
    }
    
    return user, nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, email, plainPassword string) (*entities.User, error) {
    // 查找用戶
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    
    // 驗證密碼
    isValid, err := s.passwordHasher.VerifyPassword(ctx, plainPassword, user.PasswordHash)
    if err != nil {
        return nil, fmt.Errorf("failed to verify password: %w", err)
    }
    
    if !isValid {
        return nil, fmt.Errorf("invalid credentials")
    }
    
    return user, nil
}
```

## 安全考慮

### 成本參數選擇
- **開發環境**: 使用較低成本（4-8）以加快開發速度
- **測試環境**: 使用中等成本（8-10）平衡速度和安全性
- **生產環境**: 使用較高成本（12-15）確保最大安全性

### 最佳實踐
1. **定期更新成本**: 隨著硬體性能提升，定期增加成本參數
2. **監控性能**: 監控散列操作的性能，確保不影響用戶體驗
3. **錯誤處理**: 適當處理散列和驗證過程中的錯誤
4. **日誌記錄**: 記錄散列操作的關鍵事件，但不記錄敏感資訊

## 效能考慮

### 成本與時間關係
- **成本 4**: ~1ms
- **成本 10**: ~100ms
- **成本 12**: ~400ms
- **成本 15**: ~3s

### 最佳化建議
1. **異步處理**: 在可能的情況下使用異步處理密碼散列
2. **快取策略**: 對於頻繁驗證的場景，考慮實施適當的快取策略
3. **負載均衡**: 在高負載環境中，考慮將密碼處理分散到多個服務實例

## 故障排除

### 常見錯誤

#### 1. 成本參數無效
```
Error: invalid bcrypt cost: 32, must be between 4 and 31
```
**解決方案**: 確保成本參數在有效範圍內（4-31）

#### 2. 密碼為空
```
Error: password cannot be empty
```
**解決方案**: 確保傳入的密碼不為空字符串

#### 3. 散列值格式無效
```
Error: failed to verify password: crypto/bcrypt: hashedPassword is not the hash of the given password
```
**解決方案**: 確保散列值是有效的 bcrypt 散列格式

### 調試技巧
1. **啟用詳細日誌**: 在開發環境中啟用詳細的錯誤日誌
2. **測試不同成本**: 使用不同的成本參數測試性能和安全性
3. **驗證散列格式**: 確保散列值符合預期的格式

## 相關資源

- [bcrypt 官方文檔](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [OWASP 密碼存儲指南](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [Detectviz 平台安全最佳實踐](../security/best-practices.md)

## 版本歷史

- **v1.0.0**: 初始版本，支援 bcrypt 散列
- **v1.1.0**: 添加上下文支援和錯誤處理改進
- **v1.2.0**: 添加可配置成本參數和詳細測試 