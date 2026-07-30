package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/oss"
	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/oss/ossfake"
)

// recordingOSS wraps the in-memory fake and records Mirror calls so tests can
// assert which skill directories pushBuiltinSkills mirrored to the worker prefix.
type recordingOSS struct {
	*ossfake.Memory
	mirrors [][2]string // {src, dst}
}

func (r *recordingOSS) Mirror(ctx context.Context, src, dst string, opts oss.MirrorOptions) error {
	r.mirrors = append(r.mirrors, [2]string{src, dst})
	return nil
}

// repoAgentDir locates the repo's manager/agent directory relative to this test
// file (internal/service -> repo root is three levels up).
func repoAgentDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// internal/service -> <repo>/agentteams-controller/internal/service
	// repo root is three dirs up from agentteams-controller.
	controllerRoot := filepath.Dir(filepath.Dir(wd))
	repoRoot := filepath.Dir(controllerRoot)
	agentDir := filepath.Join(repoRoot, "manager", "agent")
	if _, err := os.Stat(agentDir); err != nil {
		t.Fatalf("manager/agent not found at %s: %v", agentDir, err)
	}
	return agentDir
}

// TestPushBuiltinSkillsIncludesBetterHarness proves the builtin better-harness
// skill shipped in each agent template is mirrored to the worker's OSS skills
// prefix for every runtime — the mechanism behind the spec's "skill sync" test.
func TestPushBuiltinSkillsIncludesBetterHarness(t *testing.T) {
	ctx := context.Background()
	agentDir := repoAgentDir(t)
	workerAgentDir := filepath.Join(agentDir, "worker-agent")

	cases := []struct {
		name    string
		role    string
		runtime string
	}{
		{"openclaw worker", "worker", "openclaw"},
		{"copaw worker", "worker", "copaw"},
		{"hermes worker", "worker", "hermes"},
		{"openhuman worker", "worker", "openhuman"},
		{"qwenpaw worker", "worker", "qwenpaw"},
		{"team leader", "team_leader", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingOSS{Memory: ossfake.NewMemory()}
			d := NewDeployer(DeployerConfig{
				OSS:            store,
				WorkerAgentDir: workerAgentDir,
			})
			if err := d.pushBuiltinSkills(ctx, "alice", "agents/alice", tc.role, tc.runtime); err != nil {
				t.Fatalf("pushBuiltinSkills: %v", err)
			}
			var found bool
			for _, m := range store.mirrors {
				dst := m[1]
				if strings.HasSuffix(strings.TrimSuffix(dst, "/"), "/skills/better-harness") {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("better-harness skill not mirrored for role=%s runtime=%s; mirrors=%v", tc.role, tc.runtime, store.mirrors)
			}
		})
	}
}
