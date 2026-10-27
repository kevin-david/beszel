# Beszel update

Branch: `update/v0.21.1-2026-10-27`

## Scope

Update the deployed `0.18.8` fork (`c4387a84`) to upstream `0.21.1`
(`07e42d63`). Keep reboot alerts, configured-filesystem alerts, failed ZFS
reading suppression, and fork image publication. Add monitoring configurations
after the update. Live services remain on the old version.

## Pending choice

Kevin has been asked whether to use upstream's pool charts and current dataset
tables, or preserve the fork's historical per-dataset charts. The current
candidate uses upstream's charts. Do not deploy before that choice is answered.

## Evidence

- Web build includes TypeScript checking and completed successfully.
- Focused tests cover both old fork CBOR collisions, current upstream counters,
  alert dispatch, uptime transitions, filesystem disappearance, and failed ZFS
  readings followed by recovery.
- The two numeric system-alert tests pass after limiting previous-state parsing
  to the uptime and filesystem fields it actually uses.
- A copied-data migration boot retained monitored systems, configured alerts,
  history rows, SMART devices, and ZFS records. SQLite integrity and both custom
  select options passed. The final candidate still needs this check repeated.
- The broad system-manager suite crashes in SSH authentication on both the
  candidate and an untouched `upstream/main` checkout. The broad agent suite
  also reported a lifecycle test port-binding failure. Neither is a full pass.

## Next steps

1. Apply the chart choice. If preserving dataset history, add its telemetry and
   averaging at fork-only CBOR field numbers and adapt the current pool charts.
2. Review the final diff. Build the final hub image and host agent. Repeat the
   copied-data check with the final image.
3. Before deployment, save the current image identity, agent binary, Compose
   file, and a consistent complete data backup. Keep credentials private.
4. Deploy the compatibility-capable hub first, then replace the host agent.
   Keep the Mac agent's existing configuration; its old messages remain
   decodable, although its version cannot refresh the new ZFS detail protocol.
5. Verify the serving revision, fresh host samples, both custom alert records,
   and all live ZFS pools with refreshed dataset details. Roll back the hub
   image, data, and agent binary together if verification fails.
6. Add the agreed memory warning, then settle and implement Frigate recovery.
   No alert configuration or automatic restart has been installed yet.
