# Resume — ZFS PR #2209 integration

Created: 2026-08-25T1600Z
Updated: 2026-08-25T1606Z
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
- Committed the main feature port as 590cf5f3.
- Added ZFS tooling to the Alpine agent image in 2f44ad86.
- Applied the PR's TypeScript follow-up in ba011f56.
- Preserved the fork release workflow in 5172c7f6.
- Replayed reboot-alert behavior; its focused alert and hub tests pass.
- Confirmed the broader WebSocket status timing failures reproduce unchanged on current upstream main, so they are baseline failures rather than an integration regression.
- Integrated configured-filesystem disappearance tracking with ZFS dataset usage: availability refresh runs first, unavailable mounts are skipped, and available ZFS mounts use `zfs list` accounting.
- Verified focused agent, ZFS, disk, alert, hub, and migration tests after the semantic merge.

## Next action

Commit configured-filesystem disappearance alerts, then inspect the complete candidate and address integration defects.

## Open decisions or blockers

- None.

## Last focused verification

- `go test -tags='testing no_ui' ./agent/... ./internal/alerts/... ./internal/hub/systems/... ./internal/migrations/... ./internal/hub -run 'Reboot|Filesystem|Zfs|Disk' -count=1` — passed.

## Pointers

- Issue / PR: https://github.com/henrygd/beszel/pull/2209
- Transcript: none
- Other source of truth: live virusdesk Beszel hub and agent
