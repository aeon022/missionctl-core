package palette

import "testing"

func TestMatchPrefixBeforeContains(t *testing.T) {
	cmds := []Command{
		{Name: "settings", Desc: "d1", Key: "S"},
		{Name: "stats", Desc: "d2", Key: "t"},
		{Name: "delete", Desc: "d3", Key: "d"},
	}
	got := Match(cmds, "s")
	if len(got) != 2 || got[0].Name != "settings" || got[1].Name != "stats" {
		t.Errorf("expected [settings, stats] (prefix matches, in order), got %v", got)
	}
}

func TestMatchContainsFallsBackAfterPrefix(t *testing.T) {
	cmds := []Command{
		{Name: "archive", Desc: "d1", Key: "a"},
		{Name: "search", Desc: "d2", Key: "/"},
	}
	got := Match(cmds, "hive") // suffix of "archive", not a prefix of anything
	if len(got) != 1 || got[0].Name != "archive" {
		t.Errorf("expected contains-match [archive], got %v", got)
	}

	got = Match(cmds, "ear") // substring of "search" only
	if len(got) != 1 || got[0].Name != "search" {
		t.Errorf("expected contains-match [search], got %v", got)
	}
}

func TestMatchEmptyQueryReturnsAll(t *testing.T) {
	cmds := []Command{{Name: "a"}, {Name: "b"}}
	got := Match(cmds, "")
	if len(got) != 2 {
		t.Errorf("expected all commands back for empty query, got %v", got)
	}
}

func TestMatchNoHits(t *testing.T) {
	cmds := []Command{{Name: "archive"}, {Name: "delete"}}
	if got := Match(cmds, "zzz"); len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}
