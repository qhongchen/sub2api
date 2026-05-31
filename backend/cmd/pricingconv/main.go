package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

type cloudPriceTable struct {
	Models map[string]map[string]any `toml:"models"`
}

type convertOptions struct {
	PreferredProviders []string
}

type convertStats struct {
	TotalModels          int            `json:"total_models"`
	WrittenModels        int            `json:"written_models"`
	SkippedNoPrice       int            `json:"skipped_no_price"`
	SelectedFromProvider int            `json:"selected_from_provider"`
	ProviderCounts       map[string]int `json:"provider_counts"`
}

var priceFields = map[string]struct{}{
	"input_cost_per_token":                              {},
	"input_cost_per_token_priority":                     {},
	"output_cost_per_token":                             {},
	"output_cost_per_token_priority":                    {},
	"cache_creation_input_token_cost":                   {},
	"cache_creation_input_token_cost_above_1hr":         {},
	"cache_read_input_token_cost":                       {},
	"cache_read_input_token_cost_priority":              {},
	"input_cost_per_token_above_200k_tokens":            {},
	"output_cost_per_token_above_200k_tokens":           {},
	"cache_creation_input_token_cost_above_200k_tokens": {},
	"cache_read_input_token_cost_above_200k_tokens":     {},
	"input_cost_per_token_above_272k_tokens":            {},
	"output_cost_per_token_above_272k_tokens":           {},
	"cache_creation_input_token_cost_above_272k_tokens": {},
	"cache_read_input_token_cost_above_272k_tokens":     {},
	"input_cost_per_second":                             {},
	"output_cost_per_second":                            {},
	"input_cost_per_image":                              {},
	"output_cost_per_image":                             {},
	"input_cost_per_image_token":                        {},
	"output_cost_per_image_token":                       {},
	"input_cost_per_pixel":                              {},
	"output_cost_per_pixel":                             {},
	"input_cost_per_query":                              {},
	"selected_pricing_provider":                         {},
}

var defaultProviderPreference = []string{
	"openai",
	"anthropic",
	"deepseek",
	"google",
	"vertex_ai",
	"vertex",
	"cohere",
	"mistral",
	"xai",
	"bedrock",
	"azure",
	"azure_ai",
	"openrouter",
	"vercel",
	"kilo",
	"novita-ai",
	"siliconflow",
	"fireworks-ai",
	"fireworks_ai",
	"deepinfra",
}

func main() {
	input := flag.String("input", "", "Input prices-base.toml path, URL, or '-' for stdin")
	output := flag.String("output", "-", "Output JSON path, or '-' for stdout")
	prefer := flag.String("prefer", "", "Comma-separated provider preference override")
	pretty := flag.Bool("pretty", true, "Pretty-print JSON")
	statsPath := flag.String("stats", "", "Optional path for conversion stats JSON")
	flag.Parse()

	if strings.TrimSpace(*input) == "" {
		exitf("missing -input")
	}

	body, err := readInput(*input)
	if err != nil {
		exitf("read input: %v", err)
	}

	options := convertOptions{PreferredProviders: defaultProviderPreference}
	if strings.TrimSpace(*prefer) != "" {
		options.PreferredProviders = splitCSV(*prefer)
	}

	result, stats, err := convertTomlToLiteLLMJSON(body, options)
	if err != nil {
		exitf("convert: %v", err)
	}

	var encoded []byte
	if *pretty {
		encoded, err = json.MarshalIndent(result, "", "  ")
	} else {
		encoded, err = json.Marshal(result)
	}
	if err != nil {
		exitf("marshal output: %v", err)
	}
	encoded = append(encoded, '\n')

	if err := writeOutput(*output, encoded); err != nil {
		exitf("write output: %v", err)
	}

	if strings.TrimSpace(*statsPath) != "" {
		statsBytes, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			exitf("marshal stats: %v", err)
		}
		statsBytes = append(statsBytes, '\n')
		if err := os.WriteFile(*statsPath, statsBytes, 0644); err != nil {
			exitf("write stats: %v", err)
		}
	}
}

func convertTomlToLiteLLMJSON(body []byte, options convertOptions) (map[string]map[string]any, convertStats, error) {
	var table cloudPriceTable
	if err := toml.Unmarshal(body, &table); err != nil {
		return nil, convertStats{}, err
	}
	if len(table.Models) == 0 {
		return nil, convertStats{}, fmt.Errorf("missing or empty models table")
	}

	out := make(map[string]map[string]any, len(table.Models))
	stats := convertStats{
		TotalModels:    len(table.Models),
		ProviderCounts: map[string]int{},
	}

	modelNames := make([]string, 0, len(table.Models))
	for modelName := range table.Models {
		modelNames = append(modelNames, modelName)
	}
	sort.Strings(modelNames)

	for _, modelName := range modelNames {
		entry := table.Models[modelName]
		if len(entry) == 0 {
			stats.SkippedNoPrice++
			continue
		}

		converted, selectedProvider := flattenModel(modelName, entry, options)
		if !hasUsablePrice(converted) {
			stats.SkippedNoPrice++
			continue
		}

		if selectedProvider != "" {
			stats.SelectedFromProvider++
			stats.ProviderCounts[selectedProvider]++
		} else if provider := stringValue(converted["litellm_provider"]); provider != "" {
			stats.ProviderCounts[provider]++
		}

		out[modelName] = converted
		stats.WrittenModels++
	}

	if len(out) == 0 {
		return nil, stats, fmt.Errorf("no models with usable prices found")
	}

	return out, stats, nil
}

func flattenModel(modelName string, entry map[string]any, options convertOptions) (map[string]any, string) {
	base := copyWithoutPricing(entry)
	pricing := pricingMap(entry["pricing"])
	if len(pricing) == 0 {
		return base, ""
	}

	selectedProvider, node := selectProvider(modelName, entry, pricing, options)
	if selectedProvider == "" || node == nil {
		return base, ""
	}

	for field := range priceFields {
		delete(base, field)
	}
	for key, value := range node {
		base[key] = value
	}
	base["litellm_provider"] = selectedProvider
	base["selected_pricing_provider"] = selectedProvider
	return base, selectedProvider
}

func copyWithoutPricing(entry map[string]any) map[string]any {
	out := make(map[string]any, len(entry))
	for key, value := range entry {
		if key == "pricing" || key == "metadata" {
			continue
		}
		out[key] = value
	}
	return out
}

func pricingMap(value any) map[string]map[string]any {
	raw, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]map[string]any, len(raw))
	for provider, node := range raw {
		if nodeMap, ok := node.(map[string]any); ok {
			out[provider] = nodeMap
		}
	}
	return out
}

func selectProvider(modelName string, entry map[string]any, pricing map[string]map[string]any, options convertOptions) (string, map[string]any) {
	candidates := providerCandidates(modelName, entry, options)
	for _, candidate := range candidates {
		if provider, node := lookupProvider(pricing, candidate); provider != "" {
			return provider, node
		}
	}

	providers := make([]string, 0, len(pricing))
	for provider := range pricing {
		providers = append(providers, provider)
	}
	sort.Slice(providers, func(i, j int) bool {
		leftScore := providerScore(providers[i], pricing[providers[i]], options.PreferredProviders)
		rightScore := providerScore(providers[j], pricing[providers[j]], options.PreferredProviders)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return providers[i] < providers[j]
	})
	if len(providers) == 0 {
		return "", nil
	}
	return providers[0], pricing[providers[0]]
}

func providerCandidates(modelName string, entry map[string]any, options convertOptions) []string {
	var candidates []string
	add := func(value string) {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			return
		}
		for _, existing := range candidates {
			if existing == value {
				return
			}
		}
		candidates = append(candidates, value)
	}

	text := strings.ToLower(modelName + " " + stringValue(entry["model_family"]) + " " + stringValue(entry["display_name"]))
	switch {
	case strings.Contains(text, "deepseek"):
		add("deepseek")
	case strings.Contains(text, "claude") || strings.Contains(text, "anthropic"):
		add("anthropic")
	case strings.Contains(text, "gemini"):
		add("google")
		add("vertex_ai")
		add("vertex")
	case strings.Contains(text, "gpt") || strings.Contains(text, "openai") || strings.HasPrefix(strings.ToLower(modelName), "o"):
		add("openai")
	case strings.Contains(text, "cohere"):
		add("cohere")
	case strings.Contains(text, "mistral"):
		add("mistral")
	case strings.Contains(text, "grok") || strings.Contains(text, "xai"):
		add("xai")
	}

	add(stringValue(entry["litellm_provider"]))
	for _, provider := range options.PreferredProviders {
		add(provider)
	}
	return candidates
}

func lookupProvider(pricing map[string]map[string]any, candidate string) (string, map[string]any) {
	if node, ok := pricing[candidate]; ok {
		return candidate, node
	}
	normalizedCandidate := normalizeProviderKey(candidate)
	for provider, node := range pricing {
		if normalizeProviderKey(provider) == normalizedCandidate {
			return provider, node
		}
	}
	return "", nil
}

func providerScore(provider string, node map[string]any, preferred []string) int {
	score := 0
	for field := range priceFields {
		if isFiniteNumber(node[field]) {
			score += 10
		}
	}
	for i, preferredProvider := range preferred {
		if normalizeProviderKey(provider) == normalizeProviderKey(preferredProvider) {
			score += 1000 - i
			break
		}
	}
	return score
}

func hasUsablePrice(entry map[string]any) bool {
	for _, field := range []string{
		"input_cost_per_token",
		"output_cost_per_token",
	} {
		if _, ok := entry[field]; ok {
			return true
		}
	}
	return false
}

func isFiniteNumber(value any) bool {
	switch v := value.(type) {
	case int:
		return true
	case int64:
		return true
	case float64:
		return !isNaNOrInf(v)
	default:
		return false
	}
}

func isNaNOrInf(value float64) bool {
	return math.IsNaN(value) || math.IsInf(value, 0)
}

func stringValue(value any) string {
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func normalizeProviderKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func readInput(input string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "-" {
		return io.ReadAll(os.Stdin)
	}
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Get(input)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(input)
}

func writeOutput(output string, body []byte) error {
	output = strings.TrimSpace(output)
	if output == "" || output == "-" {
		_, err := os.Stdout.Write(body)
		return err
	}
	return os.WriteFile(output, body, 0644)
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
