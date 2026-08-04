package ai

import "testing"

func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "GEMINI_MODEL",
		"OLLAMA_HOST", "OLLAMA_MODEL", "TESTTOOL_PROVIDER",
	} {
		t.Setenv(k, "")
	}
}

func TestDetect_PicksAnthropicFirst(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-xxx")
	t.Setenv("OPENAI_API_KEY", "sk-oai-xxx")

	info, err := Detect("TESTTOOL")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if info.Name != ProviderAnthropic {
		t.Errorf("Name = %q, want anthropic", info.Name)
	}
}

func TestDetect_FallsBackToOllamaWithNoKeys(t *testing.T) {
	clearProviderEnv(t)

	info, err := Detect("TESTTOOL")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if info.Name != ProviderOllama {
		t.Errorf("Name = %q, want ollama (should always be available as a no-key fallback)", info.Name)
	}
}

func TestDetect_GeminiFreeTeirDetected(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("GEMINI_API_KEY", "AIzaxxx")

	info, err := Detect("TESTTOOL")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if info.Name != ProviderGemini {
		t.Errorf("Name = %q, want gemini", info.Name)
	}
}

func TestDetect_OverrideRequiresMatchingCredential(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("TESTTOOL_PROVIDER", "anthropic")
	// no ANTHROPIC_API_KEY set

	_, err := Detect("TESTTOOL")
	if err == nil {
		t.Fatal("expected error when override provider has no credential")
	}
}

func TestDetect_OverrideSelectsAmongMultiple(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-xxx")
	t.Setenv("GEMINI_API_KEY", "AIzaxxx")
	t.Setenv("TESTTOOL_PROVIDER", "gemini")

	info, err := Detect("TESTTOOL")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if info.Name != ProviderGemini {
		t.Errorf("Name = %q, want gemini (explicit override)", info.Name)
	}
}
