package keymap

import "testing"

func TestKeysAreSingleTokens(t *testing.T) {
	for name, k := range map[string]string{
		"SearchKey": SearchKey, "HelpKey": HelpKey, "DeleteKey": DeleteKey,
		"ConfirmKey": ConfirmKey, "QuitKey": QuitKey, "BackKey": BackKey, "SyncKey": SyncKey,
	} {
		if k == "" {
			t.Errorf("keymap.%s is empty", name)
		}
	}
}
