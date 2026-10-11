package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	newIssue      = "https://github.com/example/todo/issues/12"
	oldIssue      = "https://github.com/example/todo/issues/3"
	defaultConfig = "repo: example/todo\nlabel: adw\ninterval_seconds: 20\nprompt: /adw {{issue_url}}\n"
)

func TestTriggerHelperProcess(t *testing.T) {
	if os.Getenv("TEST_TRIGGER_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			if err := run(os.Args[i+1:]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exitCode(err))
			}
			os.Exit(0)
		}
	}
	os.Exit(90)
}

func runTrigger(t *testing.T, issues, configBody string, env []string, args ...string) (string, string, string, error) {
	t.Helper()
	project, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if configBody == "" {
		configBody = defaultConfig
	}
	write("adw.yml", configBody)
	write("issues", issues)
	write("gh", `#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" > "$TEST_ROOT/gh-args"
[[ "${TEST_GH_EXIT:-0}" == 0 ]] || exit "$TEST_GH_EXIT"
cat "$TEST_ROOT/issues"
`)
	write("claude", `#!/usr/bin/env bash
set -euo pipefail
[[ $# == 2 && "$1" == -p ]] || exit 90
if IFS= read -r unexpected; then exit 91; fi
printf 'START:%s\n' "$2" >> "$TEST_ROOT/events"
printf '%s' "$PWD" > "$TEST_ROOT/cwd"
[[ "${TEST_CLAUDE_EXIT:-0}" == 0 ]] || exit "$TEST_CLAUDE_EXIT"
/bin/sleep 0.02
printf 'END:%s\n' "$2" >> "$TEST_ROOT/events"
`)
	write("sleep", `#!/usr/bin/env bash
set -euo pipefail
printf 'SLEEP:%s\n' "$1" >> "$TEST_ROOT/events"
if [[ -f "$TEST_ROOT/slept" ]]; then exit 73; fi
touch "$TEST_ROOT/slept"
`)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	childArgs := []string{"-test.run=^TestTriggerHelperProcess$", "--", "--config", filepath.Join(root, "adw.yml")}
	cmd := exec.CommandContext(ctx, os.Args[0], append(childArgs, args...)...)
	cmd.Dir = project
	for _, value := range env {
		if value == "TEST_USE_WRAPPER=1" {
			wrapperArgs := []string{filepath.Join(project, "adws/trigger-github.sh"), "--config", filepath.Join(root, "adw.yml")}
			cmd = exec.CommandContext(ctx, "bash", append(wrapperArgs, args...)...)
			cmd.Dir = root
		}
	}
	cmd.Env = append(os.Environ(),
		"PATH="+root+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TEST_ROOT="+root, "TEST_TRIGGER_HELPER=1", "TEST_GH_EXIT=0", "TEST_CLAUDE_EXIT=0",
	)
	cmd.Env = append(cmd.Env, env...)
	output, runErr := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("trigger exceeded test timeout: %s", output)
	}
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, name))
		if os.IsNotExist(err) {
			return ""
		}
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	if cwd := read("cwd"); cwd != "" && cwd != project {
		t.Fatalf("unexpected working directory: %s", cwd)
	}
	return string(output), read("events"), read("gh-args"), runErr
}

func TestTriggerOnceRunsIssuesInSequence(t *testing.T) {
	output, events, args, err := runTrigger(t, newIssue+"\n"+oldIssue+"\n", "", nil, "--once")
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	wantArgs := []string{"issue", "list", "--repo", "example/todo", "--state", "open", "--label", "adw", "--search", "sort:created-desc", "--limit", "100", "--json", "url", "--jq", ".[].url"}
	if got := strings.Split(strings.TrimSpace(args), "\n"); !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("query must request newest open adw issues: %q", got)
	}
	want := "START:/adw " + newIssue + "\nEND:/adw " + newIssue + "\nSTART:/adw " + oldIssue + "\nEND:/adw " + oldIssue + "\n"
	if events != want {
		t.Fatalf("want sequential invocations %q, got %q", want, events)
	}
}

func TestTriggerShellWrapper(t *testing.T) {
	output, events, _, err := runTrigger(t, newIssue+"\n", "", []string{"TEST_USE_WRAPPER=1", "GOFLAGS=-mod=readonly"}, "--once")
	want := "START:/adw " + newIssue + "\nEND:/adw " + newIssue + "\n"
	if err != nil || events != want {
		t.Fatalf("wrapper must load YAML and launch the configured prompt: err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerConfigurablePromptIsLiteral(t *testing.T) {
	body := `repo: example/todo
label: mi-etiqueta
interval_seconds: 7
prompt: >-
  /mi-flujo {{issue_url}} --remote origin --base main;
  $(exit 99) "{{issue_url}}"
`
	_, events, args, err := runTrigger(t, newIssue+"\n", body, []string{"ADW_PROMPT=este valor antiguo no se usa"})
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v", err)
	}
	template := `/mi-flujo {{issue_url}} --remote origin --base main; $(exit 99) "{{issue_url}}"`
	prompt := strings.ReplaceAll(template, "{{issue_url}}", newIssue)
	pass := "START:" + prompt + "\nEND:" + prompt + "\nSLEEP:7\n"
	if events != pass+pass || !strings.Contains(args, "--label\nmi-etiqueta\n") {
		t.Fatalf("YAML config must control the literal prompt, label and interval: %q, %q", events, args)
	}
}

func TestTriggerEmptyListDoesNotInvokeClaude(t *testing.T) {
	output, events, _, err := runTrigger(t, "", "", nil, "--once")
	if err != nil || events != "" {
		t.Fatalf("err=%v, events=%q, output=%s", err, events, output)
	}
}

func TestTriggerStopsOnGitHubError(t *testing.T) {
	_, events, _, err := runTrigger(t, newIssue+"\n", "", []string{"TEST_GH_EXIT=7"}, "--once")
	if err == nil || events != "" {
		t.Fatalf("err=%v, events=%q", err, events)
	}
}

func TestTriggerStopsOnClaudeError(t *testing.T) {
	_, events, _, err := runTrigger(t, newIssue+"\n"+oldIssue+"\n", "", []string{"TEST_CLAUDE_EXIT=9"}, "--once")
	if err == nil || events != "START:/adw "+newIssue+"\n" {
		t.Fatalf("must stop before launching second issue: err=%v, events=%q", err, events)
	}
}

func TestTriggerPollsWithoutRememberingIssues(t *testing.T) {
	_, events, _, err := runTrigger(t, newIssue+"\n", "", nil)
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 73 {
		t.Fatalf("unexpected exit: %v", err)
	}
	pass := "START:/adw " + newIssue + "\nEND:/adw " + newIssue + "\nSLEEP:20\n"
	if events != pass+pass {
		t.Fatalf("this minimal trigger deliberately has no memory: %q", events)
	}
}

func TestTriggerInvalidConfigurationDoesNotLaunch(t *testing.T) {
	cases := []struct {
		name string
		body string
		args []string
	}{
		{name: "missing-repo", body: strings.Replace(defaultConfig, "example/todo", "\"\"", 1)},
		{name: "missing-label", body: strings.Replace(defaultConfig, "label: adw", "label: ''", 1)},
		{name: "zero-interval", body: strings.Replace(defaultConfig, "interval_seconds: 20", "interval_seconds: 0", 1)},
		{name: "negative-interval", body: strings.Replace(defaultConfig, "interval_seconds: 20", "interval_seconds: -1", 1)},
		{name: "missing-issue-marker", body: strings.Replace(defaultConfig, "{{issue_url}}", "", 1)},
		{name: "malformed-yaml", body: "repo: [\n"},
		{name: "unknown-field", body: defaultConfig + "unexpected: true\n"},
		{name: "duplicate-field", body: defaultConfig + "repo: otro/repo\n"},
		{name: "multiple-documents", body: defaultConfig + "---\n" + defaultConfig},
		{name: "unknown-option", args: []string{"--unknown"}},
		{name: "old-repo-argument", args: []string{"example/todo"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, events, ghArgs, err := runTrigger(t, newIssue+"\n", tc.body, nil, tc.args...)
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 || events != "" || ghArgs != "" {
				t.Fatalf("invalid configuration must stop before querying: err=%v, events=%q, args=%q, output=%s", err, events, ghArgs, output)
			}
		})
	}
}

func TestTriggerMissingConfig(t *testing.T) {
	if _, err := readConfig(filepath.Join(t.TempDir(), "missing.yml")); err == nil {
		t.Fatal("missing config must report an error")
	}
}

func TestTriggerProjectConfig(t *testing.T) {
	_, err := readConfig("../config/adw.yml")
	// El fichero puede estar todavía sin configurar o contener el repo del usuario.
	if err != nil && !strings.Contains(err.Error(), "configura repo") {
		t.Fatalf("invalid project config: %v", err)
	}
}
