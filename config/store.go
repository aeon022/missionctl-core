package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// ErrNotFound is returned by Store.Read when no config file exists on any
// search path — usually fine, since defaults then apply.
var ErrNotFound = errors.New("config file not found")

// Store is a tiny YAML-backed key/value config, replacing the global viper
// instance each tool used for the same handful of calls. Keys are
// case-insensitive and may be dotted ("profiles.work.data_dir"). A value is
// resolved from the environment first (<PREFIX>_<KEY>, upper-cased, only
// for top-level keys), then from the file/defaults/Set calls.
//
// Not safe for concurrent use — same as the viper globals it replaces.
type Store struct {
	name, prefix string
	paths        []string
	data         map[string]any
}

func NewStore(name, envPrefix string) *Store {
	return &Store{name: name, prefix: envPrefix, data: map[string]any{}}
}

// Reset clears everything but name and prefix (for tests).
func (s *Store) Reset() { s.paths, s.data = nil, map[string]any{} }

// AddPath adds a directory to search for <name>.yaml; $VARS are expanded.
func (s *Store) AddPath(dir string) { s.paths = append(s.paths, os.ExpandEnv(dir)) }

// SetDefault sets key only if nothing (file, earlier default, Set) has.
func (s *Store) SetDefault(key string, v any) {
	if _, ok := lookup(s.data, key); !ok {
		s.Set(key, v)
	}
}

// Read merges the first <name>.yaml (or .yml) found on the search paths over
// whatever is already set, so defaults survive. Returns ErrNotFound if none.
func (s *Store) Read() error {
	for _, dir := range s.paths {
		for _, ext := range []string{".yaml", ".yml"} {
			b, err := os.ReadFile(filepath.Join(dir, s.name+ext))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return err
			}
			var file map[string]any
			if err := yaml.Unmarshal(b, &file); err != nil {
				return fmt.Errorf("parse %s: %w", filepath.Join(dir, s.name+ext), err)
			}
			merge(s.data, lower(file))
			return nil
		}
	}
	return ErrNotFound
}

// Get returns the raw value for key (env override first), or nil.
func (s *Store) Get(key string) any {
	if !strings.Contains(key, ".") && s.prefix != "" {
		if e := os.Getenv(s.prefix + "_" + strings.ToUpper(key)); e != "" {
			return e
		}
	}
	v, _ := lookup(s.data, key)
	return v
}

// GetString returns key as a string ("" if unset or not a scalar).
func (s *Store) GetString(key string) string {
	switch v := s.Get(key).(type) {
	case nil:
		return ""
	case string:
		return v
	case map[string]any, []any:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

// GetMap returns a shallow copy of the map at key (nil if unset/not a map),
// so callers can delete entries and Set it back.
func (s *Store) GetMap(key string) map[string]any {
	m, _ := s.Get(key).(map[string]any)
	return maps.Clone(m)
}

// Set stores v at key, creating intermediate maps for dotted keys.
func (s *Store) Set(key string, v any) {
	parts := strings.Split(strings.ToLower(key), ".")
	m := s.data
	for _, p := range parts[:len(parts)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		m = next
	}
	m[parts[len(parts)-1]] = v
}

// Write persists the current state (without env overrides) to path.
func (s *Store) Write(path string) error {
	b, err := yaml.Marshal(s.data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Unmarshal decodes the current state, with env overrides applied to
// top-level keys that exist, into out via its `yaml` struct tags.
func (s *Store) Unmarshal(out any) error {
	m := maps.Clone(s.data)
	for k := range m {
		if e := os.Getenv(s.prefix + "_" + strings.ToUpper(k)); e != "" && s.prefix != "" {
			var parsed any
			if yaml.Unmarshal([]byte(e), &parsed) != nil || parsed == nil {
				parsed = e
			}
			m[k] = parsed
		}
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

func lookup(m map[string]any, key string) (any, bool) {
	var v any = m
	for _, p := range strings.Split(strings.ToLower(key), ".") {
		mm, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		if v, ok = mm[p]; !ok {
			return nil, false
		}
	}
	return v, true
}

func lower(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if sub, ok := v.(map[string]any); ok {
			v = lower(sub)
		}
		out[strings.ToLower(k)] = v
	}
	return out
}

func merge(dst, src map[string]any) {
	for k, v := range src {
		if sm, ok := v.(map[string]any); ok {
			if dm, ok := dst[k].(map[string]any); ok {
				merge(dm, sm)
				continue
			}
		}
		dst[k] = v
	}
}
