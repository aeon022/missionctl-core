package syncdir

import (
	"os"
	"strings"
)

// Every tool's store imports this package, so a test binary always gets here: drop the developer's
// exported <TOOL>_DATA_DIR variables so tests can never read or write the real databases.
func init() {
	if !strings.HasSuffix(os.Args[0], ".test") {
		return
	}
	for _, kv := range os.Environ() {
		if k, _, _ := strings.Cut(kv, "="); strings.HasSuffix(k, "_DATA_DIR") {
			os.Unsetenv(k)
		}
	}
}
