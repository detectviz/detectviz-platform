// Package contracts provides contract version validation for go-platform
// Ensures SSOT compliance by validating proto generation consistency
package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// VersionMetadata represents the contract generation metadata
type VersionMetadata struct {
	BufVersion   string    `json:"buf_version"`
	GeneratedAt  time.Time `json:"generated_at"`
	ProtoHash    string    `json:"proto_hash"`
}

const (
	// Expected minimum versions (should match contracts/Makefile)
	MinBufVersion = "v1.47.0"
	
	// Maximum age for generated code (hours)
	MaxGeneratedAge = 72
)

// ValidateContractVersion validates proto generation version consistency
// Returns error if validation fails, causing startup to abort
func ValidateContractVersion() error {
	metadata, err := loadVersionMetadata()
	if err != nil {
		return fmt.Errorf("failed to load contract version metadata: %w", err)
	}

	// Check buf version compatibility
	if !isVersionCompatible(metadata.BufVersion, MinBufVersion) {
		return fmt.Errorf("contract version mismatch: generated with %s, minimum required %s. "+
			"Please regenerate contracts with 'cd contracts && make gen'", 
			metadata.BufVersion, MinBufVersion)
	}

	// Check generation age (prevent stale contracts)
	age := time.Since(metadata.GeneratedAt)
	if age > MaxGeneratedAge*time.Hour {
		zap.L().Warn("Contract generation is old, consider regenerating",
			zap.Duration("age", age),
			zap.String("generated_at", metadata.GeneratedAt.Format(time.RFC3339)),
		)
	}

	zap.L().Info("Contract version validation passed",
		zap.String("buf_version", metadata.BufVersion),
		zap.String("proto_hash", metadata.ProtoHash[:16]+"..."),
		zap.Time("generated_at", metadata.GeneratedAt),
	)

	return nil
}

// loadVersionMetadata loads version metadata from contracts generation
func loadVersionMetadata() (*VersionMetadata, error) {
	// Search for metadata file in common locations
	searchPaths := []string{
		"contracts/gen/metadata/version.json",
		"../contracts/gen/metadata/version.json",
		"../../contracts/gen/metadata/version.json",
	}

	var metadataPath string
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			metadataPath = path
			break
		}
	}

	if metadataPath == "" {
		return nil, fmt.Errorf("contract version metadata not found. "+
			"Please run 'cd contracts && make gen' to generate metadata. "+
			"Searched paths: %v", searchPaths)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	var metadata VersionMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &metadata, nil
}

// isVersionCompatible checks if generated version is compatible with minimum required
func isVersionCompatible(generated, minimum string) bool {
	// Simple version comparison (can be enhanced with semver library if needed)
	return generated >= minimum
}

// GetContractInfo returns contract information for health checks
func GetContractInfo() map[string]interface{} {
	metadata, err := loadVersionMetadata()
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"status":       "valid",
		"buf_version":  metadata.BufVersion,
		"proto_hash":   metadata.ProtoHash,
		"generated_at": metadata.GeneratedAt.Format(time.RFC3339),
		"age_hours":    int(time.Since(metadata.GeneratedAt).Hours()),
	}
}

// ValidateContractConsistency performs deeper consistency checks
func ValidateContractConsistency() error {
	// Check if generated Go files exist
	goGenPath := findContractGoPath()
	if goGenPath == "" {
		return fmt.Errorf("generated Go contract files not found. Run 'cd contracts && make gen'")
	}

	// Check if Python files exist (for full cross-language consistency)
	pythonGenPath := findContractPythonPath()
	if pythonGenPath == "" {
		zap.L().Warn("Generated Python contract files not found, Python runtime may fail")
	}

	zap.L().Info("Contract consistency validation passed",
		zap.String("go_path", goGenPath),
		zap.String("python_path", pythonGenPath),
	)

	return nil
}

// findContractGoPath locates generated Go contract files
func findContractGoPath() string {
	searchPaths := []string{
		"contracts/gen/go/detectviz/contracts/v1",
		"../contracts/gen/go/detectviz/contracts/v1", 
		"../../contracts/gen/go/detectviz/contracts/v1",
	}

	for _, path := range searchPaths {
		if files, err := filepath.Glob(filepath.Join(path, "*.pb.go")); err == nil && len(files) > 0 {
			return path
		}
	}
	return ""
}

// findContractPythonPath locates generated Python contract files
func findContractPythonPath() string {
	searchPaths := []string{
		"contracts/gen/python/detectviz/contracts/v1",
		"../contracts/gen/python/detectviz/contracts/v1",
		"../../contracts/gen/python/detectviz/contracts/v1",
	}

	for _, path := range searchPaths {
		if files, err := filepath.Glob(filepath.Join(path, "*_pb2.py")); err == nil && len(files) > 0 {
			return path
		}
	}
	return ""
}