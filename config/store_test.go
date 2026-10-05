package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testCfg struct {
	Name string `yaml:"name"`
	N    int    `yaml:"n"`
}

func TestStore(t *testing.T) {
	dir := t.TempDir()
	s := NewStore("app", "APPT")
	s.AddPath(dir)
	s.SetDefault("n", 30)
	s.SetDefault("name", "")
	if err := s.Read(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Read = %v, want ErrNotFound", err)
	}

	s.Set("profiles.Work.data_dir", "/x")
	s.Set("name", "a")
	if got := s.GetString("profiles.work.data_dir"); got != "/x" {
		t.Errorf("dotted get = %q", got)
	}
	if err := s.Write(filepath.Join(dir, "app.yaml")); err != nil {
		t.Fatal(err)
	}

	s2 := NewStore("app", "APPT")
	s2.AddPath(dir)
	if err := s2.Read(); err != nil {
		t.Fatal(err)
	}
	m := s2.GetMap("profiles")
	delete(m, "work")
	if _, ok := s2.GetMap("profiles")["work"]; !ok {
		t.Error("GetMap must return a copy")
	}

	t.Setenv("APPT_N", "7")
	var c testCfg
	if err := s2.Unmarshal(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "a" || c.N != 7 {
		t.Errorf("Unmarshal = %+v, want {a 7} (env override)", c)
	}
	if s2.GetString("n") != "7" {
		t.Errorf("env override via GetString = %q", s2.GetString("n"))
	}
	_ = os.Remove(filepath.Join(dir, "app.yaml"))
}
