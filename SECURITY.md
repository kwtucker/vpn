# Security policy

## Supported versions

Security fixes are provided for the latest tagged release. Older releases are
not maintained; upgrade with `mise`, `go install`, or a current release
archive.

## Reporting a vulnerability

Please report vulnerabilities privately through GitHub's private vulnerability
reporting:

1. Open the repository's **Security** tab.
2. Choose **Report a vulnerability**.
3. Describe the issue, affected version, platform, and reproduction steps.

Do not open a public issue for an unpatched vulnerability. Do not include
private keys, passphrases, tokens, or other secrets in a report.

Reports are handled on a best-effort basis. There is no bug bounty program.
Please allow time for a fix and coordinated disclosure before publishing
details.

## Threat model

`vpn` coordinates two existing VPN clients so that only one is active at a
time. It does not manage credentials, keys, or daemon configuration, and it
never uses a shell or requests elevation.

In scope:

- Leaving both VPNs active when a switch fails, including incorrect parsing of
  `mullvad status --json` or `tailscale status --json`.
- Connecting the target before the other client has actually finished going
  down, or returning before the final state has been verified.
- Command-line arguments being misinterpreted or unexpectedly forwarded to
  `mullvad` or `tailscale`, including arguments that change VPN settings.
- Unsafe handling of captured child-process output, which can contain relay,
  location, or daemon diagnostic details.
- Cleanup or rollback affecting processes and state the tool does not own.

Out of scope:

- A compromised operating system, firmware, boot chain, or hypervisor.
- Malicious `mullvad` or `tailscale` binaries earlier on `PATH`, or a
  compromised VPN daemon.
- Vulnerabilities in Mullvad, Tailscale, their daemons, or their configuration.
- Losing network access or dropping a remote session as a consequence of
  disconnect-first switching. Use a local session for the first switch.
- A third party changing VPN state outside this tool. `vpn status` detects the
  conflict but there is no background watchdog.
- Partial switches if the process is killed or the machine loses power midway.
  Run `vpn status` afterwards and repair with `vpn off` if needed.

## Security model

The README documents the intended operating model and remaining risks. This
tool is a convenience wrapper, not a security boundary between the two VPN
clients.
