package prompt

import "testing"

func TestBuildAgentCommand(t *testing.T) {
	t.Setenv("COPILOT_MODEL", "")

	cases := []struct {
		name    string
		command string
		want    string
	}{
		{"claude passthrough", "claude", "claude"},
		{"codex passthrough", "codex", "codex"},
		{"aider passthrough", "aider", "aider"},
		{"copilot non-interactive", "copilot", "copilot -s --allow-all-tools --no-ask-user -p"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildAgentCommand(tc.command); got != tc.want {
				t.Errorf("buildAgentCommand(%q) = %q, want %q", tc.command, got, tc.want)
			}
		})
	}
}

func TestBuildAgentCommandWithModel(t *testing.T) {
	t.Setenv("COPILOT_MODEL", "claude-sonnet-4.5")

	want := "copilot -s --allow-all-tools --no-ask-user --model claude-sonnet-4.5 -p"
	if got := buildAgentCommand("copilot"); got != want {
		t.Errorf("buildAgentCommand(\"copilot\") = %q, want %q", got, want)
	}

	if got := buildAgentCommand("claude"); got != "claude" {
		t.Errorf("model should not affect non-copilot agents, got %q", got)
	}
}

func TestParseAgents(t *testing.T) {
	configs, err := parseAgents("copilot:2,claude:1")
	if err != nil {
		t.Fatalf("parseAgents returned error: %v", err)
	}

	if got := configs["copilot"]; got.Command != "copilot" || got.Count != 2 {
		t.Errorf("copilot config = %+v, want {copilot 2}", got)
	}
	if got := configs["claude"]; got.Command != "claude" || got.Count != 1 {
		t.Errorf("claude config = %+v, want {claude 1}", got)
	}
}
