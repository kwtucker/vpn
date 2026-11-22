package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

type call struct {
	name string
	args []string
}

type fakeRunner struct {
	calls          []call
	mullvadState   string
	tailscaleState string
	fail           map[string]error
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		mullvadState:   "disconnected",
		tailscaleState: "Down",
		fail:           make(map[string]error),
	}
}

func (f *fakeRunner) Run(_ context.Context, _ io.Reader, name string, args []string, stdout, _ io.Writer) error {
	f.calls = append(f.calls, call{name: name, args: append([]string(nil), args...)})
	if len(args) > 0 {
		if err, ok := f.fail[name+" "+args[0]]; ok {
			return err
		}
	}
	switch {
	case name == "mullvad" && len(args) > 0 && args[0] == "status":
		fmt.Fprintf(stdout, "{\"state\":%q}\n", f.mullvadState)
	case name == "tailscale" && len(args) > 0 && args[0] == "status":
		fmt.Fprintf(stdout, "{\"BackendState\":%q}\n", f.tailscaleState)
	case name == "mullvad" && len(args) > 0 && args[0] == "connect":
		f.mullvadState = "connected"
	case name == "mullvad" && len(args) > 0 && args[0] == "disconnect":
		f.mullvadState = "disconnected"
	case name == "tailscale" && len(args) > 0 && args[0] == "up":
		f.tailscaleState = "Running"
	case name == "tailscale" && len(args) > 0 && args[0] == "down":
		f.tailscaleState = "Down"
	}
	return nil
}

func (f *fakeRunner) index(name, firstArg string) int {
	for i, c := range f.calls {
		if c.name == name && len(c.args) > 0 && c.args[0] == firstArg {
			return i
		}
	}
	return -1
}

func (f *fakeRunner) has(name, firstArg string) bool {
	return f.index(name, firstArg) >= 0
}

func (f *fakeRunner) argsFor(name, firstArg string) []string {
	if i := f.index(name, firstArg); i >= 0 {
		return f.calls[i].args
	}
	return nil
}

func newTestApp(f *fakeRunner) (*App, *bytes.Buffer, *bytes.Buffer) {
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	app := NewApp(f, strings.NewReader(""), out, errOut)
	app.WaitInterval = time.Millisecond
	app.CommandTimeout = time.Second
	app.ConnectTimeout = time.Second
	app.DownTimeout = time.Second
	return app, out, errOut
}

func contains(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func TestConnectMullvadBringsTailscaleDownFirst(t *testing.T) {
	f := newFakeRunner()
	f.tailscaleState = "Running"
	app, out, _ := newTestApp(f)

	if err := app.ConnectMullvad(context.Background(), nil); err != nil {
		t.Fatalf("ConnectMullvad: %v", err)
	}
	down, connect := f.index("tailscale", "down"), f.index("mullvad", "connect")
	if down < 0 || connect < 0 || down > connect {
		t.Fatalf("expected tailscale down before mullvad connect, calls: %v", f.calls)
	}
	if f.mullvadState != "connected" || f.tailscaleState != "Down" {
		t.Fatalf("final states: mullvad=%q tailscale=%q", f.mullvadState, f.tailscaleState)
	}
	if !strings.Contains(out.String(), "Mullvad connected") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestConnectTailscaleDisconnectsMullvadFirst(t *testing.T) {
	f := newFakeRunner()
	f.mullvadState = "connected"
	app, out, _ := newTestApp(f)

	if err := app.ConnectTailscale(context.Background(), nil); err != nil {
		t.Fatalf("ConnectTailscale: %v", err)
	}
	disconnect, up := f.index("mullvad", "disconnect"), f.index("tailscale", "up")
	if disconnect < 0 || up < 0 || disconnect > up {
		t.Fatalf("expected mullvad disconnect before tailscale up, calls: %v", f.calls)
	}
	if f.mullvadState != "disconnected" || f.tailscaleState != "Running" {
		t.Fatalf("final states: mullvad=%q tailscale=%q", f.mullvadState, f.tailscaleState)
	}
	if !strings.Contains(out.String(), "Tailscale up") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestConnectMullvadAbortsWhenTailscaleDownFails(t *testing.T) {
	f := newFakeRunner()
	f.tailscaleState = "Running"
	f.fail["tailscale down"] = errors.New("permission denied")
	app, _, _ := newTestApp(f)

	if err := app.ConnectMullvad(context.Background(), nil); err == nil {
		t.Fatal("expected an error when tailscale down fails")
	}
	if f.has("mullvad", "connect") {
		t.Fatalf("must not connect Mullvad if Tailscale is still up, calls: %v", f.calls)
	}
}

func TestConnectTailscaleAbortsWhenMullvadDisconnectFails(t *testing.T) {
	f := newFakeRunner()
	f.mullvadState = "connected"
	f.fail["mullvad disconnect"] = errors.New("daemon unavailable")
	app, _, _ := newTestApp(f)

	if err := app.ConnectTailscale(context.Background(), nil); err == nil {
		t.Fatal("expected an error when mullvad disconnect fails")
	}
	if f.has("tailscale", "up") {
		t.Fatalf("must not bring Tailscale up if Mullvad is still connected, calls: %v", f.calls)
	}
}

func TestConnectTailscaleUsesNoFlagsByDefault(t *testing.T) {
	f := newFakeRunner()
	app, _, _ := newTestApp(f)

	if err := app.ConnectTailscale(context.Background(), nil); err != nil {
		t.Fatalf("ConnectTailscale: %v", err)
	}
	args := f.argsFor("tailscale", "up")
	if len(args) != 1 || args[0] != "up" {
		t.Fatalf("expected plain 'tailscale up', got %v", args)
	}
}

func TestConnectTailscaleForwardsExtraArgs(t *testing.T) {
	f := newFakeRunner()
	app, _, _ := newTestApp(f)

	if err := app.ConnectTailscale(context.Background(), []string{"--exit-node=auto:any", "--timeout=5s"}); err != nil {
		t.Fatalf("ConnectTailscale: %v", err)
	}
	args := f.argsFor("tailscale", "up")
	if !contains(args, "--exit-node=auto:any") {
		t.Fatalf("extra flag not forwarded: %v", args)
	}
	if !contains(args, "--timeout=5s") {
		t.Fatalf("user timeout not forwarded: %v", args)
	}
}

func TestConnectMullvadDoesNotDuplicateWait(t *testing.T) {
	f := newFakeRunner()
	app, _, _ := newTestApp(f)

	if err := app.ConnectMullvad(context.Background(), []string{"--wait"}); err != nil {
		t.Fatalf("ConnectMullvad: %v", err)
	}
	args := f.argsFor("mullvad", "connect")
	count := 0
	for _, arg := range args {
		if arg == "--wait" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one --wait, got %v", args)
	}
}

func TestStatusDetectsConflict(t *testing.T) {
	f := newFakeRunner()
	f.mullvadState = "connected"
	f.tailscaleState = "Running"
	app, out, errOut := newTestApp(f)

	if err := app.Status(context.Background()); err == nil {
		t.Fatal("expected a conflict error when both VPNs are active")
	}
	if !strings.Contains(errOut.String(), "CONFLICT") {
		t.Fatalf("missing conflict warning: %q", errOut.String())
	}
	if !strings.Contains(out.String(), "Mullvad:   connected") {
		t.Fatalf("unexpected status output: %q", out.String())
	}
}

func TestStatusClean(t *testing.T) {
	f := newFakeRunner()
	f.tailscaleState = "Running"
	app, out, _ := newTestApp(f)

	if err := app.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !strings.Contains(out.String(), "Tailscale: up") {
		t.Fatalf("unexpected status output: %q", out.String())
	}
}

func TestRootWithNoArgsShowsStatus(t *testing.T) {
	f := newFakeRunner()
	f.tailscaleState = "Running"
	app, out, _ := newTestApp(f)
	root := NewRootCmd(app)
	root.SetArgs([]string{})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out.String(), "Tailscale: up") {
		t.Fatalf("unexpected status output: %q", out.String())
	}
}

func TestPassthroughHelpDoesNotRunCommands(t *testing.T) {
	f := newFakeRunner()
	app, out, _ := newTestApp(f)
	root := NewRootCmd(app)
	root.SetArgs([]string{"mull", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(f.calls) != 0 {
		t.Fatalf("help must not run commands, calls: %v", f.calls)
	}
	if !strings.Contains(out.String(), "Connect Mullvad") {
		t.Fatalf("unexpected help output: %q", out.String())
	}
}
