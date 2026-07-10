package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
)

// New API 的可借鉴点是按模型合并 header；这里用前缀列表表达 1m context 能力范围。
var claudeContext1MModelPrefixes = []string{
	"claude-sonnet-4",
	"claude-opus-4-8",
	"claude-fable-5",
	"fable-5",
}

func shouldForceClaudeContext1M(modelID string) bool {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	modelID = strings.TrimSuffix(modelID, "-thinking")
	for _, prefix := range claudeContext1MModelPrefixes {
		if strings.HasPrefix(modelID, prefix) {
			return true
		}
	}
	return false
}

func forcedClaudeContextBetasForModel(modelID string, enabled bool) []string {
	if !enabled || !shouldForceClaudeContext1M(modelID) {
		return nil
	}
	return []string{claude.BetaContext1M}
}

func appendForcedClaudeContextBetas(modelID string, betas []string, enabled bool) []string {
	forcedBetas := forcedClaudeContextBetasForModel(modelID, enabled)
	if len(forcedBetas) == 0 {
		return betas
	}
	out := make([]string, 0, len(betas)+len(forcedBetas))
	out = append(out, betas...)
	out = append(out, forcedBetas...)
	return out
}

func dropSetWithoutForcedClaudeContextBetas(modelID string, drop map[string]struct{}, enabled bool) map[string]struct{} {
	forcedBetas := forcedClaudeContextBetasForModel(modelID, enabled)
	if len(forcedBetas) == 0 || len(drop) == 0 {
		return drop
	}

	forcedSet := make(map[string]struct{}, len(forcedBetas))
	for _, beta := range forcedBetas {
		forcedSet[beta] = struct{}{}
	}

	out := make(map[string]struct{}, len(drop))
	for beta := range drop {
		if _, forced := forcedSet[beta]; forced {
			continue
		}
		out[beta] = struct{}{}
	}
	return out
}

func mergeAnthropicBetaDroppingForModel(modelID string, required []string, incoming string, drop map[string]struct{}, forceContext1M bool) string {
	effectiveDropSet := dropSetWithoutForcedClaudeContextBetas(modelID, drop, forceContext1M)
	return mergeAnthropicBetaDropping(appendForcedClaudeContextBetas(modelID, required, forceContext1M), incoming, effectiveDropSet)
}

func computeForcedClaudeContextBeta(modelID string, incoming string, drop map[string]struct{}, forceContext1M bool) (string, bool) {
	forcedBetas := forcedClaudeContextBetasForModel(modelID, forceContext1M)
	if len(forcedBetas) == 0 {
		if incoming == "" {
			return "", false
		}
		return stripBetaTokensWithSet(incoming, drop), true
	}
	finalBeta := mergeAnthropicBetaDropping(forcedBetas, incoming, dropSetWithoutForcedClaudeContextBetas(modelID, drop, forceContext1M))
	return finalBeta, finalBeta != "" || incoming != ""
}

func explicitClaudeContext1MBeta(incoming string) string {
	if anthropicBetaTokensContains(incoming, claude.BetaContext1M) {
		return claude.BetaContext1M
	}
	return ""
}

func requestUsesClaudeContext1M(req *http.Request) bool {
	return req != nil && anthropicBetaTokensContains(getHeaderRaw(req.Header, "anthropic-beta"), claude.BetaContext1M)
}

func isClaudeContext1MForceEnabled(settingService *SettingService, ctx context.Context) bool {
	if settingService == nil {
		return false
	}
	return settingService.IsClaudeContext1MForceEnabled(ctx)
}

func (s *GatewayService) checkForcedClaudeContextBetaPolicy(ctx context.Context, account *Account, modelID string, forceContext1M bool) error {
	forcedBetas := forcedClaudeContextBetasForModel(modelID, forceContext1M)
	if len(forcedBetas) == 0 {
		return nil
	}
	if blockErr := s.checkBetaPolicyBlockForTokens(ctx, forcedBetas, account, modelID); blockErr != nil {
		return blockErr
	}
	return nil
}
