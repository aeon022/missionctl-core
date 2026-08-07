// Package licensing wraps the missionctl suite's Polar.sh license-key
// integration — the `<tool> license activate/status` commands every tool
// duplicated independently as an identical copy-pasted cmd/license.go.
//
// Beyond deduplication, this package makes license checks product-aware.
// The previous per-tool code only checked license-key status ("active" or
// not) against a shared Polar organization ID — any valid key under that
// org unlocked every tool identically, because the validate/activate
// response's benefit_id (Polar's own field for telling license-key types
// apart — see https://polar.sh/docs, "be sure to validate their unique
// benefit_id") was read from the API but never inspected. That was fine
// while the missionctl Bundle was the only product, but breaks the moment
// per-tool products exist: a $9 calctl-only key would otherwise also
// unlock budgetctl/notectl/mailctl/habctl's gated AI features for free.
package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultOrgID is aeon022's Polar.sh organization — shared across the
// missionctl suite.
const DefaultOrgID = "aa792ea4-650e-492e-a955-9b3d564e943e"

// Result is what a Polar activate/validate call resolves to, and what
// gets cached to disk between runs.
type Result struct {
	Status    string `mapstructure:"license_status" json:"status"`
	BenefitID string `mapstructure:"license_benefit_id" json:"benefit_id"`
}

// Grants reports whether this Result unlocks Pro features for a tool,
// given that tool's own individual-product benefit ID and the shared
// Bundle's benefit ID (either may legitimately grant access).
//
// If BOTH benefitID and bundleID are empty, per-product scoping hasn't
// been wired up yet for this tool (the individual product doesn't exist
// in Polar, or its benefit ID hasn't been recorded here) — in that case
// any active key under our org grants access, matching the suite's
// original behavior so existing Bundle customers are never locked out
// mid-rollout. Once both IDs are set, only a key whose benefit_id
// actually matches one of them grants access.
func (r Result) Grants(benefitID, bundleID string) bool {
	if !r.isValid() {
		return false
	}
	if benefitID == "" && bundleID == "" {
		return true
	}
	return r.BenefitID == benefitID || r.BenefitID == bundleID
}

// isValid reports whether Status represents a usable key. Polar's real
// license-key resource uses "granted" (confirmed against the live API —
// this package originally assumed "active" without ever having checked a
// real response, which silently rejected every genuinely valid key).
// "active" is still accepted too, in case a different Polar response
// shape or a manually-set local override uses it.
func (r Result) isValid() bool {
	return r.Status == "granted" || r.Status == "active"
}

type activateRequest struct {
	Key            string `json:"key"`
	OrganizationID string `json:"organization_id"`
	Label          string `json:"label"`
}

type validateRequest struct {
	Key            string `json:"key"`
	OrganizationID string `json:"organization_id"`
}

type polarError struct {
	Error  string      `json:"error"`
	Detail interface{} `json:"detail"`
}

// polarResponse covers both the activate and validate endpoints' response
// shapes — both return the license key resource, which includes benefit_id.
type polarResponse struct {
	Status    string `json:"status"`
	BenefitID string `json:"benefit_id"`
}

// MaskKey shows only the first 4 characters of a secret, for safe display.
func MaskKey(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}

// Label builds the per-machine activation label Polar shows in the
// customer's dashboard (which devices have activated a key).
func Label(toolName string) string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = toolName + "-terminal"
	}
	return fmt.Sprintf("%s (%s)", hostname, time.Now().Format("02.01.2006"))
}

// Activate registers a license key for this machine with Polar. On a
// network error the key is still worth recording locally (status
// "offline_pending") so a later `license status` can verify it once
// online — this mirrors what every tool's original license.go already did.
//
// Not every License Key benefit has per-device activation tracking turned
// on in Polar's dashboard (confirmed live: a real Bundle purchase key came
// back 403 "NotPermitted" — "This license key does not support
// activations. Use the /validate endpoint instead."). When that's the
// case, this transparently falls back to Validate: a successful validate
// is just as good a signal that the key is real and paid-for as an
// activation record would be, and the customer shouldn't be blocked just
// because the product wasn't configured to track individual devices.
func Activate(orgID, key, label string) (Result, error) {
	body, err := json.Marshal(activateRequest{Key: key, OrganizationID: orgID, Label: label})
	if err != nil {
		return Result{}, err
	}

	resp, err := http.Post("https://api.polar.sh/v1/customer-portal/license-keys/activate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return Result{Status: "offline_pending"}, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var pr polarResponse
		_ = json.Unmarshal(respBytes, &pr)
		status := pr.Status
		if status == "" {
			status = "granted"
		}
		return Result{Status: status, BenefitID: pr.BenefitID}, nil
	}

	var polarErr polarError
	_ = json.Unmarshal(respBytes, &polarErr)
	if polarErr.Error == "NotPermitted" {
		return Validate(orgID, key)
	}

	errMsg := "Invalid or inactive key"
	if polarErr.Error != "" {
		errMsg = polarErr.Error
	}
	err = fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	if isTransientStatus(resp.StatusCode) {
		// Rate-limited or Polar's own server hiccupping — not proof the
		// key is bad. Empty Status (like the network-error path above)
		// tells the caller to leave whatever's already cached alone
		// instead of overwriting a known-good status with "invalid".
		return Result{}, err
	}
	return Result{Status: "invalid"}, err
}

// isTransientStatus reports whether an HTTP status from Polar indicates a
// temporary problem (rate limiting, their server erroring) rather than a
// definitive answer about the key itself. Learned the hard way: a 429
// during `license status` was overwriting a perfectly valid cached
// "granted" status with "invalid" simply because Polar was asked too many
// times in quick succession.
func isTransientStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

// Validate re-checks a previously activated key's current status with
// Polar. A network error is reported so the caller can fall back to a
// cached Result rather than treating it as invalid.
func Validate(orgID, key string) (Result, error) {
	body, err := json.Marshal(validateRequest{Key: key, OrganizationID: orgID})
	if err != nil {
		return Result{}, err
	}

	resp, err := http.Post("https://api.polar.sh/v1/customer-portal/license-keys/validate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("server returned status %d", resp.StatusCode)
		if isTransientStatus(resp.StatusCode) {
			return Result{}, err
		}
		return Result{Status: "invalid"}, err
	}

	var pr polarResponse
	_ = json.Unmarshal(respBytes, &pr)
	status := pr.Status
	if status == "" {
		status = "granted"
	}
	return Result{Status: status, BenefitID: pr.BenefitID}, nil
}
