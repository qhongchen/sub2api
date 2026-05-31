package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertTomlToLiteLLMJSONSelectsOfficialDeepSeekProvider(t *testing.T) {
	body := []byte(`
[models."deepseek-v4-pro"]
display_name = "DeepSeek V4 Pro"
model_family = "deepseek-thinking"
mode = "chat"
input_cost_per_token = 0.00000174
output_cost_per_token = 0.00000348
cache_read_input_token_cost = 0.000000145
litellm_provider = "auriko"
providers = ["auriko", "deepseek", "digitalocean"]

[models."deepseek-v4-pro".pricing."auriko"]
input_cost_per_token = 0.00000174
output_cost_per_token = 0.00000348
cache_read_input_token_cost = 0.000000145

[models."deepseek-v4-pro".pricing."deepseek"]
input_cost_per_token = 0.000000435
output_cost_per_token = 0.00000087
cache_read_input_token_cost = 0.000000003625
`)

	out, stats, err := convertTomlToLiteLLMJSON(body, convertOptions{
		PreferredProviders: defaultProviderPreference,
	})
	require.NoError(t, err)
	require.Equal(t, 1, stats.WrittenModels)
	require.Equal(t, 1, stats.SelectedFromProvider)

	got := out["deepseek-v4-pro"]
	require.NotNil(t, got)
	require.Equal(t, "deepseek", got["litellm_provider"])
	require.Equal(t, "deepseek", got["selected_pricing_provider"])
	require.Equal(t, 0.000000435, got["input_cost_per_token"])
	require.Equal(t, 0.00000087, got["output_cost_per_token"])
	require.Equal(t, 0.000000003625, got["cache_read_input_token_cost"])
}

func TestConvertTomlToLiteLLMJSONSkipsEntriesWithoutTokenPrice(t *testing.T) {
	body := []byte(`
[models."image-only-model"]
mode = "image_generation"
litellm_provider = "openai"
output_cost_per_image = 0.04

[models."seconds-only-model"]
mode = "chat"
litellm_provider = "bedrock"
input_cost_per_second = 0.01
output_cost_per_second = 0.01
`)

	out, stats, err := convertTomlToLiteLLMJSON(body, convertOptions{
		PreferredProviders: defaultProviderPreference,
	})
	require.Error(t, err)
	require.Nil(t, out)
	require.Equal(t, 2, stats.TotalModels)
	require.Equal(t, 0, stats.WrittenModels)
	require.Equal(t, 2, stats.SkippedNoPrice)
}
