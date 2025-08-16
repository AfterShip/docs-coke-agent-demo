package main

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// 直接使用 anthropic SDK 的参数结构
type Request = anthropic.MessageNewParams

type Config struct {
	AIGCHubBaseURL string `json:"aigc_hub_base_url,omitempty"`
	LLMModel       string `json:"llm_model,omitempty"`
	APIKey         string `json:"api_key,omitempty"`
}
