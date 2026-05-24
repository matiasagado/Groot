# Client-side log shipping

Vector configs deployed on each project host. Each project tails its own
logs and ships them to the homelab Loki, stamped with a `project` label so
Groot (and Grafana) can route by source.

## Layout

| Path | Purpose |
|---|---|
| `foothold/vector.yaml`  | Foothold — real users, production-class signal |
| `totem/vector.yaml`     | Totem — real users, festival-spiky traffic |
| `portfolio/vector.yaml` | Portfolio site — visitors only, low-signal |
| `atlas/vector.yaml`     | Atlas search index — demo, low-signal |
| `_shared/nginx-parser.yaml`        | Reusable VRL transform for projects behind nginx |
| `_shared/nginx-enhanced-json.conf` | Optional nginx `log_format` block for structured access logs |

Each per-project `vector.yaml` is a skeleton. Source `include` paths are
placeholders — fill them in when the project lands on the homelab.

## Deploy on a project host

1. Install Vector (https://vector.dev/docs/setup/installation/).
2. Drop the project's `vector.yaml` at `/etc/vector/vector.yaml`.
3. Set `LOKI_PUSH_URL` in Vector's environment to the homelab Loki, reached
   over Tailscale:
   ```
   LOKI_PUSH_URL=http://<xps-tailscale-ip>:3100
   ```
4. Adjust `sources.app_logs.include` to point at the project's real log
   paths.
5. Start Vector (`systemctl start vector`, or however Vector is run on the
   host).

If the project is behind nginx and nginx logs should also be ingested, copy
the `nginx_parser` transform from `_shared/nginx-parser.yaml` into the
project's `vector.yaml`, add an `nginx_logs` file source for
`/var/log/nginx/*.log`, and wire the parser into the existing `enrich`
transform.

## Severity weighting

Foothold and Totem are production-class sources (real users). Portfolio and
Atlas are showcase traffic. Groot weighs alerts accordingly — the `project`
label is what the weighting logic reads.
