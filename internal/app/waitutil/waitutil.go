package waitutil

import (
	"fmt"
	"strings"
	"time"
)

type WaitSpec struct {
	ID       string
	Interval time.Duration
	Timeout  time.Duration
}

type WaitExecutionResult struct {
	Rows   []map[string]interface{}
	Failed bool
}

func ValidateSingleWaitSpec(spec WaitSpec) error {
	if strings.TrimSpace(spec.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if spec.Interval < 0 {
		return fmt.Errorf("interval-seconds must be >= 0")
	}
	if spec.Timeout < 0 {
		return fmt.Errorf("timeout-seconds must be >= 0")
	}
	return nil
}

func WaitResult(id, operation, result, status, displayStatus string, elapsedSeconds int, errText string) map[string]interface{} {
	return map[string]interface{}{
		"id":              id,
		"operation":       operation,
		"result":          result,
		"status":          status,
		"display_status":  displayStatus,
		"elapsed_seconds": elapsedSeconds,
		"error":           errText,
	}
}

func StringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func MapFromData(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func FirstNonEmptyString(values ...interface{}) string {
	for _, value := range values {
		if s := strings.TrimSpace(StringValue(value)); s != "" {
			return s
		}
	}
	return ""
}

func NormalizeStorageKind(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch {
	case lower == "":
		return ""
	case strings.Contains(lower, "objectstorage"), strings.Contains(lower, "oss"), strings.Contains(lower, "对象存储"):
		return "objectstorage"
	case strings.Contains(lower, "hypergate"), strings.Contains(lower, "blockstorage"), strings.Contains(lower, "block"), strings.Contains(lower, "cloud-sync-gateway"), strings.Contains(lower, "云同步网关"), strings.Contains(lower, "块存储"):
		return "hypergate"
	default:
		return lower
	}
}

func IsFailureStatus(values ...string) bool {
	return hasStatusKeyword(values,
		"fail", "failed", "error", "invalid", "denied", "expired", "exception",
		"失败", "错误", "异常", "无效", "拒绝", "过期",
	)
}

func IsSuccessStatus(values ...string) bool {
	return hasStatusKeyword(values,
		"available", "active", "valid", "success", "successful", "ready", "normal", "ok", "online", "enabled",
		"创建成功", "成功", "正常", "可用", "有效", "就绪",
	)
}

func IsRunningStatus(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return hasStatusKeyword(values,
		"creating", "checking", "verifying", "pending", "running", "doing", "processing", "initializing", "queued", "progress",
		"创建中", "校验中", "处理中", "执行中", "等待中", "进行中",
	)
}

func hasStatusKeyword(values []string, keywords ...string) bool {
	for _, value := range values {
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "" {
			continue
		}
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				return true
			}
		}
	}
	return false
}
