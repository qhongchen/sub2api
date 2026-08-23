package service

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	// OpenAIFirstOutputTimeoutExtraKey 保存账号级首语义输出超时；缺失或 null 表示继承全局配置。
	OpenAIFirstOutputTimeoutExtraKey   = "openai_first_output_timeout_seconds"
	OpenAIFirstOutputTimeoutMinSeconds = 30
	OpenAIFirstOutputTimeoutMaxSeconds = 1800
)

// ValidateOpenAIFirstOutputTimeoutExtra 校验账号级超时；null 表示清除覆盖值。
func ValidateOpenAIFirstOutputTimeoutExtra(platform string, extra map[string]any) error {
	if extra == nil {
		return nil
	}
	raw, exists := extra[OpenAIFirstOutputTimeoutExtraKey]
	if !exists {
		return nil
	}
	if platform != PlatformOpenAI {
		return infraerrors.BadRequest(
			"OPENAI_FIRST_OUTPUT_TIMEOUT_PLATFORM_INVALID",
			"openai_first_output_timeout_seconds can only be set on OpenAI accounts",
		)
	}
	if raw == nil {
		return nil
	}
	seconds, ok := openAIFirstOutputTimeoutSecondsValue(raw)
	if !ok || seconds < 0 || (seconds != 0 && (seconds < OpenAIFirstOutputTimeoutMinSeconds || seconds > OpenAIFirstOutputTimeoutMaxSeconds)) {
		return infraerrors.BadRequest(
			"OPENAI_FIRST_OUTPUT_TIMEOUT_INVALID",
			fmt.Sprintf("%s must be 0 or between %d-%d seconds, or null", OpenAIFirstOutputTimeoutExtraKey, OpenAIFirstOutputTimeoutMinSeconds, OpenAIFirstOutputTimeoutMaxSeconds),
		)
	}
	return nil
}

// normalizeOpenAIFirstOutputTimeoutExtra 复制 Extra，并在完整账号写入时删除继承态的 null 键。
func normalizeOpenAIFirstOutputTimeoutExtra(platform string, extra map[string]any) (map[string]any, error) {
	if extra == nil {
		return extra, nil
	}
	if err := ValidateOpenAIFirstOutputTimeoutExtra(platform, extra); err != nil {
		return nil, err
	}
	if platform != PlatformOpenAI {
		return extra, nil
	}
	normalized := maps.Clone(extra)
	if normalized[OpenAIFirstOutputTimeoutExtraKey] == nil {
		delete(normalized, OpenAIFirstOutputTimeoutExtraKey)
	}
	return normalized, nil
}

func normalizeOpenAIFirstOutputTimeoutUpdateExtra(account *Account, input *UpdateAccountInput, normalized map[string]any) (map[string]any, error) {
	if account == nil || input == nil {
		return normalized, nil
	}
	raw, provided := input.Extra[OpenAIFirstOutputTimeoutExtraKey]
	if provided {
		if err := ValidateOpenAIFirstOutputTimeoutExtra(account.Platform, input.Extra); err != nil {
			return nil, err
		}
		if normalized == nil {
			normalized = make(map[string]any)
		}
		if raw == nil {
			delete(normalized, OpenAIFirstOutputTimeoutExtraKey)
		} else {
			normalized[OpenAIFirstOutputTimeoutExtraKey] = raw
		}
		return normalized, nil
	}
	if account.Platform == PlatformOpenAI {
		if current, exists := account.Extra[OpenAIFirstOutputTimeoutExtraKey]; exists && current != nil {
			if normalized == nil {
				normalized = make(map[string]any)
			}
			normalized[OpenAIFirstOutputTimeoutExtraKey] = current
		}
	}
	return normalized, nil
}

func openAIFirstOutputTimeoutSecondsValue(raw any) (int, bool) {
	var value float64
	switch v := raw.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		if int64(int(v)) != v {
			return 0, false
		}
		return int(v), true
	case uint:
		if uint(int(v)) != v {
			return 0, false
		}
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		if uint32(int(v)) != v {
			return 0, false
		}
		return int(v), true
	case uint64:
		if uint64(int(v)) != v {
			return 0, false
		}
		return int(v), true
	case float32:
		value = float64(v)
	case float64:
		value = v
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0, false
		}
		value = parsed
	default:
		return 0, false
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
		return 0, false
	}
	seconds := int(value)
	if float64(seconds) != value {
		return 0, false
	}
	return seconds, true
}

// openAIFirstOutputTimeoutOverride 返回秒数、是否显式配置和配置是否合法。
// 数字 0 属于显式配置；缺失和 null 都表示继承。
func openAIFirstOutputTimeoutOverride(extra map[string]any) (int, bool, bool) {
	if len(extra) == 0 {
		return 0, false, true
	}
	raw, exists := extra[OpenAIFirstOutputTimeoutExtraKey]
	if !exists || raw == nil {
		return 0, false, true
	}
	seconds, ok := openAIFirstOutputTimeoutSecondsValue(raw)
	if !ok || seconds < 0 || (seconds != 0 && (seconds < OpenAIFirstOutputTimeoutMinSeconds || seconds > OpenAIFirstOutputTimeoutMaxSeconds)) {
		return 0, true, false
	}
	return seconds, true, true
}
