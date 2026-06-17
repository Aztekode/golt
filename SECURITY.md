# Security Policy

## Supported Versions

At the moment, only the latest tagged release is considered supported for
security fixes.

## Reporting a Vulnerability

Please do not open public issues for suspected vulnerabilities.

Instead:

1. Prepare a minimal report with impact, affected files, and reproduction steps
2. Contact the maintainers privately through the repository security reporting
   channel if available
3. If GitHub private vulnerability reporting is not enabled, contact the
   maintainers directly before public disclosure

## Response Goals

- Acknowledge receipt as soon as possible
- Confirm severity and impact
- Prepare a fix and release notes
- Credit the reporter if they want public attribution

## Scope

Security-sensitive areas currently include:

- Release integrity (`SHA256SUMS`, `SHA256SUMS.minisig`, release workflow)
- Installer behavior (`installer.iss`)
- Runtime filesystem, environment, and network access
- Future package resolution and module installation logic
