# vpn

A tiny CLI that connects to **exactly one** of Mullvad or Tailscale at a time.
`vpn mull` brings Tailscale down before connecting Mullvad; `vpn tail`
disconnects Mullvad before bringing Tailscale up. If the other VPN cannot be
disconnected, the switch is aborted rather than leaving both active.

The CLI shells out to the existing `mullvad` and `tailscale` binaries. It does
not manage accounts, keys, or VPN settings.

## Install

With mise:

```sh
mise use -g github:kwtucker/vpn@latest
```

With Go:

```sh
go install github.com/kwtucker/vpn/cmd/vpn@latest
```

Both `mullvad` and `tailscale` must be installed and usable without `sudo`.

## Commands

| Command | Purpose |
| --- | --- |
| `vpn` / `vpn status` | Print both states; exits non-zero if both are active |
| `vpn mull` | Disconnect Tailscale, connect Mullvad |
| `vpn tail` | Disconnect Mullvad, bring Tailscale up |
| `vpn off` / `vpn down` | Disconnect both |
| `vpn --version` | Print the version |

`m`, `mullvad`, `t`, `tailscale`, `st`, and `disconnect` are accepted aliases.

Extra flags after `mull` or `tail` are passed straight to the underlying CLI:

```sh
vpn tail --exit-node=auto:any
vpn mull --wait
```

Mullvad's progress chatter is captured so `vpn mull` prints one summary line;
if the command fails, the captured output is included in the error.

The wrapper adds `--wait` to `mullvad connect`. It calls plain `tailscale up`
so your existing Tailscale preferences are preserved; flags you supply are
forwarded verbatim. Command execution and the state checks are bounded by the
wrapper's own timeouts. It then verifies both the target state and that the
other VPN is still down; a failed check rolls the target back and returns an
error.

Note: Tailscale treats any flag passed to `up` as a settings change and then
requires every non-default preference to be repeated. If you pass flags, check
`tailscale get` first or add `--reset` deliberately; plain `vpn tail` avoids
that entirely.

## Why mutual exclusion?

Tailscale and Mullvad are separate WireGuard clients that both manage routes,
firewall rules, and DNS. Running them as independent full clients is not
reliable: they can fight over the default route, and Tailscale's DNS management
(`accept-dns`) conflicts with Mullvad's DNS enforcement. The supported way to
combine them is Tailscale's Mullvad exit-node add-on, which is a different
architecture. This tool deliberately enforces one-at-a-time instead.

Two caveats:

- Plain `tailscale up` gives you tailnet access, **not** an internet VPN. Use
  `vpn tail --exit-node=...` if you want traffic routed through Tailscale.
- Disconnect-first can drop your session if you are connected remotely through
  the VPN being disconnected. First tests should be run locally.

## Development

```sh
make build
make check   # fmt-check, lint, tests, govulncheck
make install
```

Unit tests use a fake command runner and assert call order, argument
forwarding, conflict detection, and abort-on-failure.
