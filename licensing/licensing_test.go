package licensing

import "testing"

func TestGrants(t *testing.T) {
	cases := []struct {
		name      string
		result    Result
		benefitID string
		bundleID  string
		want      bool
	}{
		{
			name:      "inactive status never grants",
			result:    Result{Status: "invalid", BenefitID: "tool-x"},
			benefitID: "tool-x",
			bundleID:  "bundle-1",
			want:      false,
		},
		{
			name:      "unscoped (no IDs configured yet) grants on any active key",
			result:    Result{Status: "active", BenefitID: "whatever"},
			benefitID: "",
			bundleID:  "",
			want:      true,
		},
		{
			name:      "matches own tool benefit",
			result:    Result{Status: "active", BenefitID: "tool-x"},
			benefitID: "tool-x",
			bundleID:  "bundle-1",
			want:      true,
		},
		{
			name:      "matches bundle benefit",
			result:    Result{Status: "active", BenefitID: "bundle-1"},
			benefitID: "tool-x",
			bundleID:  "bundle-1",
			want:      true,
		},
		{
			name:      "active key for a different tool's product does not grant",
			result:    Result{Status: "active", BenefitID: "tool-y"},
			benefitID: "tool-x",
			bundleID:  "bundle-1",
			want:      false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.result.Grants(c.benefitID, c.bundleID); got != c.want {
				t.Errorf("Grants(%q, %q) = %v, want %v", c.benefitID, c.bundleID, got, c.want)
			}
		})
	}
}

func TestMaskKey(t *testing.T) {
	cases := map[string]string{
		"":         "",
		"ab":       "****",
		"abcd":     "****",
		"abcdefgh": "abcd****",
	}
	for in, want := range cases {
		if got := MaskKey(in); got != want {
			t.Errorf("MaskKey(%q) = %q, want %q", in, got, want)
		}
	}
}
