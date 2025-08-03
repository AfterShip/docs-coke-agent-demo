package tools

import (
	"testing"
)

func TestWeatherTool(t *testing.T) {
	tool := NewWeatherTool()

	// Test tool metadata
	if tool.Name() != "get_weather" {
		t.Errorf("Expected name 'get_weather', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestWeatherRequestResponse(t *testing.T) {
	tests := []struct {
		name     string
		request  WeatherRequest
		response WeatherResponse
	}{
		{
			name: "weather data structures",
			request: WeatherRequest{
				Location: "北京",
			},
			response: WeatherResponse{
				Weather: "天气晴朗",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test request structure
			if tt.request.Location != "北京" {
				t.Errorf("Expected location '北京', got '%s'", tt.request.Location)
			}

			// Test response structure
			if tt.response.Weather != "天气晴朗" {
				t.Errorf("Expected weather '天气晴朗', got '%s'", tt.response.Weather)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if len(registry.GetRegisteredTools()) != 0 {
		t.Error("New registry should be empty")
	}

	// Test tool registration
	weatherTool := NewWeatherTool()
	registry.Register(weatherTool)

	tools := registry.GetRegisteredTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 registered tool, got %d", len(tools))
	}

	if tools[0].Name() != "get_weather" {
		t.Errorf("Expected registered tool name 'get_weather', got '%s'", tools[0].Name())
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	registry := GetDefaultRegistry()

	tools := registry.GetRegisteredTools()
	if len(tools) == 0 {
		t.Error("Default registry should have at least one tool")
	}

	// Check that weather tool is registered
	foundWeatherTool := false
	for _, tool := range tools {
		if tool.Name() == "get_weather" {
			foundWeatherTool = true
			break
		}
	}

	if !foundWeatherTool {
		t.Error("Default registry should include weather tool")
	}
}
