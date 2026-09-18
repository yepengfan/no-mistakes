package agent

import (
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestAcpxAgent_BuildArgs_SuppressesCursorProjectConfigsUnderOptOut(t *testing.T) {
	a := &acpxAgent{
		bin:                    "acpx",
		target:                 "cursor",
		rawCommand:             "cursor-agent acp",
		disableProjectSettings: true,
	}
	args := a.buildArgs(RunOpts{CWD: "/work"})
	if !argsContainPair(args, "--agent", "cursor-agent --disable-project-configs acp") {
		t.Fatalf("buildArgs = %v, want --agent \"cursor-agent --disable-project-configs acp\"", args)
	}
}

func TestAcpxAgent_BuildArgs_NoSuppressionWithoutOptOut(t *testing.T) {
	a := &acpxAgent{
		bin:        "acpx",
		target:     "cursor",
		rawCommand: "cursor-agent acp",
	}
	args := a.buildArgs(RunOpts{CWD: "/work"})
	if !argsContainPair(args, "--agent", "cursor-agent acp") {
		t.Fatalf("buildArgs = %v, want unchanged default raw command", args)
	}
}

func TestAcpxAgent_BuildArgs_NoSuppressionForNonCursorTargets(t *testing.T) {
	a := &acpxAgent{
		bin:                    "acpx",
		target:                 "gemini",
		rawCommand:             "gemini --acp",
		disableProjectSettings: true,
	}
	args := a.buildArgs(RunOpts{CWD: "/work"})
	if !argsContainPair(args, "--agent", "gemini --acp") {
		t.Fatalf("buildArgs = %v, must not inject cursor flag for other targets", args)
	}
}

func TestAcpxAgent_BuildArgs_PreservesExplicitDisableProjectConfigsPin(t *testing.T) {
	a := &acpxAgent{
		bin:                    "acpx",
		target:                 "cursor",
		rawCommand:             "cursor-agent --disable-project-configs acp",
		disableProjectSettings: true,
	}
	args := a.buildArgs(RunOpts{})
	if !argsContainPair(args, "--agent", "cursor-agent --disable-project-configs acp") {
		t.Fatalf("buildArgs = %v, must not duplicate the operator pin", args)
	}
}

func TestAcpxAgent_NeutralizesGateInstructions_CursorUnderOptOut(t *testing.T) {
	for _, name := range []types.AgentName{types.AgentCursor, "acp:cursor"} {
		a, err := NewWithOptions(name, "acpx", nil, Options{DisableProjectSettings: true})
		if err != nil {
			t.Fatalf("NewWithOptions(%s): %v", name, err)
		}
		if !NeutralizesGateInstructions(a) {
			t.Fatalf("%s must neutralize under opt-out", name)
		}
	}
}

func TestAcpxAgent_NeutralizesGateInstructions_FalseWithoutOptOut(t *testing.T) {
	a, err := NewWithOptions(types.AgentCursor, "acpx", nil, Options{})
	if err != nil {
		t.Fatalf("NewWithOptions(cursor): %v", err)
	}
	if NeutralizesGateInstructions(a) {
		t.Fatal("cursor must not report neutralized when the repo did not opt out")
	}
}

func TestAcpxAgent_NeutralizesGateInstructions_DefeatedOverrideFailsClosed(t *testing.T) {
	a, err := NewWithOptions(types.AgentCursor, "acpx", nil, Options{
		DisableProjectSettings: true,
		ACPRegistryOverrides: map[string]string{
			"cursor": "cursor-agent --disable-project-configs=false acp",
		},
	})
	if err != nil {
		t.Fatalf("NewWithOptions(cursor): %v", err)
	}
	if NeutralizesGateInstructions(a) {
		t.Fatal("cursor with --disable-project-configs=false must fail closed under opt-out")
	}
	if err := EnsureGateNeutralized(a); err == nil {
		t.Fatal("EnsureGateNeutralized must refuse a defeated cursor override")
	}
}

func TestCursorEffectiveRawCommand_QuotesPathsWithSpaces(t *testing.T) {
	got := cursorEffectiveRawCommand(`"/opt/cursor bins/cursor-agent" acp`)
	want := `"/opt/cursor bins/cursor-agent" --disable-project-configs acp`
	if got != want {
		t.Fatalf("cursorEffectiveRawCommand() = %q, want %q", got, want)
	}
	if !strings.Contains(got, "--disable-project-configs") {
		t.Fatalf("effective command missing suppression flag: %q", got)
	}
}
