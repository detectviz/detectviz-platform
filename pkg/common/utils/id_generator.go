package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IDGenerator 提供各種 ID 生成功能
// 職責: 統一管理平台內所有 ID 生成邏輯
// AI_SCAFFOLD_HINT: ID 生成工具，支援多種 ID 格式，AI 可自動選擇合適的生成策略
type IDGenerator struct{}

// NewIDGenerator 創建新的 ID 生成器
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// GenerateUUID 生成標準 UUID
// AI_SCAFFOLD_HINT: 用於實體主鍵生成
func (g *IDGenerator) GenerateUUID() string {
	return uuid.New().String()
}

// GenerateShortID 生成短 ID（8 字符）
// AI_SCAFFOLD_HINT: 用於用戶友好的短 ID
func (g *IDGenerator) GenerateShortID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateTimestampID 生成基於時間戳的 ID
// AI_SCAFFOLD_HINT: 用於需要時間排序的場景
func (g *IDGenerator) GenerateTimestampID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}

// GeneratePluginID 生成插件專用 ID
// AI_SCAFFOLD_HINT: 插件實例 ID 生成，格式：{type}-{timestamp}-{random}
func (g *IDGenerator) GeneratePluginID(pluginType string) string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 2)
	rand.Read(randomBytes)
	return fmt.Sprintf("%s-%d-%s", pluginType, timestamp, hex.EncodeToString(randomBytes))
}

// GenerateSessionID 生成會話 ID
// AI_SCAFFOLD_HINT: 用於用戶會話管理
func (g *IDGenerator) GenerateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateUUID 驗證 UUID 格式
// AI_SCAFFOLD_HINT: UUID 格式驗證工具
func (g *IDGenerator) ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
