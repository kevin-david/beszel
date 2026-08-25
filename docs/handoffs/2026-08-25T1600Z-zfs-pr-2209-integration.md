# Resume — ZFS PR #2209 integration

Created: 2026-08-25T1600Z
Updated: 2026-08-25T1602Z
Branch: kd/zfs-pr-2209-test

## Goal

Port upstream Beszel PR #2209 onto current upstream, preserve the fork-specific behavior, verify it on virusdesk, then update the fork's main branch safely.

## Accepted decisions

- Bring in all of upstream PR #2209 so virusdesk can test native ZFS monitoring for its author.
- Keep the fork's release workflow, reboot alerts, and generic configured-filesystem disappearance alerts unless integrated behavior proves one redundant.
- Preserve the current fork tip under a backup ref before updating main.
- After verification, force-push origin/main with an exact force-with-lease.
- Do not post feedback upstream without separate approval.

## Completed

- Verified PR #2209 is open at e69f9660, 83 commits behind current upstream main, conflicting, unreviewed, and has a failed VulnCheck.
- Created local backup ref backup/main-pre-zfs-pr2209-2026-08-25 at f4d55212.
- Created kd/zfs-pr-2209-test from upstream main 4c48fe0c.
- Ported the main ZFS feature commit ba12d8d5 onto current upstream, preserving both the newer manager shutdown context and the ZFS fetch cache.
- Verified the tagged agent, alert, system-manager, and migration package tests pass.

## Next action

Commit the main ZFS feature port, then apply the Alpine image and TypeScript follow-up commits.

## Open decisions or blockers

- None.

## Last focused verification

- `go test -tags='testing no_ui' ./agent/... ./internal/alerts/... ./internal/hub/systems/... ./internal/migrations/...` — passed.

## Pointers

- Issue / PR: https://github.com/henrygd/beszel/pull/2209
- Transcript: none
- Other source of truth: live virusdesk Beszel hub and agent
