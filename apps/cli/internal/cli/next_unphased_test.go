package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/next"
)

// unphasedIDs runs `next` over strictPhaseFiles with phases v0.2, v0.3 and
// next.unphased set to value (nil leaves the key absent), returning the ranked
// IDs and the command error.
func unphasedIDs(t *testing.T, value any, args ...string) ([]string, error) {
	t.Helper()
	repo := newTaskRepo(t, strictPhaseFiles())
	res := repo.RunWith(func() {
		setPhaseOrder([]string{"v0.2", "v0.3"})
		if value != nil {
			viper.Set(nextUnphasedConfigKey, value)
		}
	}, append([]string{"next", "--format", "json", "--limit", "10"}, args...)...)
	if res.Err != nil {
		return nil, res.Err
	}
	return recIDs(t, res.Stdout), nil
}

func TestNext_UnphasedSetting(t *testing.T) {
	// strictPhaseFiles: 001 low v0.2, 002 critical v0.3, 003 medium v0.2,
	// 004 high unphased.
	cases := []struct {
		name  string
		value any
		want  []string
	}{
		{"absent defaults to current", nil, []string{"003", "001", "004", "002"}},
		{"current joins the current tier", "current", []string{"003", "001", "004", "002"}},
		{"last ranks after every phase", "last", []string{"003", "001", "002", "004"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := unphasedIDs(t, tc.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIDOrder(t, ids, tc.want...)
		})
	}
}

func TestNext_UnphasedSetting_LastBeforeUnknownPhase(t *testing.T) {
	files := strictPhaseFiles()
	files["005.md"] = "---\nid: \"005\"\ntitle: \"Typo phase\"\nstatus: pending\npriority: critical\nphase: v9.9\n---"
	repo := newTaskRepo(t, files)

	res := repo.RunWith(func() {
		setPhaseOrder([]string{"v0.2", "v0.3"})
		viper.Set(nextUnphasedConfigKey, "last")
	}, "next", "--format", "json", "--limit", "10")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	assertIDOrder(t, recIDs(t, res.Stdout), "003", "001", "002", "004", "005")
}

func TestNext_UnphasedSetting_NoEffectWithStrictPhasesOff(t *testing.T) {
	last, err := unphasedIDs(t, "last", "--strict-phases=false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	current, err := unphasedIDs(t, nil, "--strict-phases=false")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertIDOrder(t, last, current...)
}

func TestNext_UnphasedSetting_NoEffectWithoutPhases(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	withLast := repo.RunWith(func() { viper.Set(nextUnphasedConfigKey, "last") },
		"next", "--format", "json", "--limit", "10")
	if withLast.Err != nil {
		t.Fatalf("unexpected error: %v", withLast.Err)
	}
	plain := nextStdout(t, repo, "--format", "json", "--limit", "10")

	if withLast.Stdout != plain {
		t.Errorf("without phases next.unphased must not change ranking\nlast: %s\ndefault: %s", withLast.Stdout, plain)
	}
}

func TestNext_UnphasedSetting_InvalidValue(t *testing.T) {
	cases := []struct {
		name    string
		value   any
		wantMsg []string
	}{
		{"unknown string", "lst", []string{".taskmd.yaml", "next.unphased", `"lst"`, "current, last", `did you mean "last"`}},
		{"non-string", []any{"last"}, []string{".taskmd.yaml", "next.unphased", "found a list"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := unphasedIDs(t, tc.value)
			if err == nil {
				t.Fatal("expected a config error")
			}
			for _, want := range tc.wantMsg {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q missing %q", err, want)
				}
			}
		})
	}
}

func TestParseUnphasedPlacement(t *testing.T) {
	cases := []struct {
		raw     any
		want    next.UnphasedPlacement
		wantErr bool
	}{
		{nil, next.UnphasedCurrent, false},
		{"current", next.UnphasedCurrent, false},
		{"last", next.UnphasedLast, false},
		{"", "", true},
		{"LAST", "", true},
		{true, "", true},
	}
	for _, tc := range cases {
		got, err := parseUnphasedPlacement(tc.raw)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseUnphasedPlacement(%v) error = %v, wantErr %v", tc.raw, err, tc.wantErr)
			continue
		}
		if got != tc.want {
			t.Errorf("parseUnphasedPlacement(%v) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// createUnphasedProject writes a registered project whose own .taskmd.yaml
// ranks unphased tasks last. Unphased 001 is critical, so only the setting
// keeps it behind the low-priority early-phase 002.
func createUnphasedProject(t *testing.T, unphased string) string {
	t.Helper()
	repo := newTaskRepo(t, nil)
	repo.Write(configFilename, "task-dir: tasks\nphases:\n  - id: early\nnext:\n  unphased: "+unphased+"\n")
	repo.Write(filepath.Join("tasks", "001-bug.md"),
		"---\nid: \"001\"\ntitle: \"Bug\"\nstatus: pending\npriority: critical\n---\n# Bug\n")
	repo.Write(filepath.Join("tasks", "002-early.md"), phasedTaskMD("002", "Early", "low", "early"))
	return repo.Dir
}

func TestNextAllProjects_UsesEachProjectsUnphasedSetting(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+createUnphasedProject(t, "last")+
		"\n  - id: beta\n    path: "+createUnphasedProject(t, "current")+"\n")

	recs := runNextAllProjectsJSON(t)

	if bug, early := indexOfRec(recs, "alpha", "001"), indexOfRec(recs, "alpha", "002"); bug < early {
		t.Errorf("alpha ranks unphased last: 002 must precede 001; got %+v", recs)
	}
	if bug, early := indexOfRec(recs, "beta", "001"), indexOfRec(recs, "beta", "002"); bug > early {
		t.Errorf("beta ranks unphased in the current tier: 001 must precede 002; got %+v", recs)
	}
}

func TestNextAllProjects_InvalidUnphasedSettingSkipsProject(t *testing.T) {
	resetCLIState()
	defer resetCLIState()
	setupProjectFlagRegistry(t, "  - id: alpha\n    path: "+createUnphasedProject(t, "sideways")+
		"\n  - id: beta\n    path: "+createUnphasedProject(t, "last")+"\n")

	res := newTaskRepo(t, nil).Run("next", "--all-projects", "--format", "json")
	if res.Err != nil {
		t.Fatalf("next --all-projects: %v", res.Err)
	}
	if !strings.Contains(res.Stderr, `skipping project "alpha"`) || !strings.Contains(res.Stderr, "next.unphased") {
		t.Errorf("expected a skip warning naming next.unphased, got stderr: %q", res.Stderr)
	}
	if strings.Contains(res.Stdout, `"alpha"`) || !strings.Contains(res.Stdout, `"beta"`) {
		t.Errorf("expected only beta recommendations, got: %s", res.Stdout)
	}
}
