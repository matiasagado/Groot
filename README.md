# Groot

AI log analysis. Logs come in, an LLM reads them, anomalies and plain-English summaries come out.

Groot is the intelligence layer over my personal project portfolio — it ingests logs from four of my projects, weighs them by user impact, and surfaces what's worth my attention.

---

## Origins

The original proof of concept was a senior team project sponsored by [Erik (@StealthBadger747)](https://github.com/StealthBadger747). That codebase established the core architecture — a Go log ingest pipeline, a Redis queue feeding a Python AI worker, ClickHouse for long-term storage, and a Go web UI — and validated that a locally-hosted LLM could produce useful analysis on streaming log data.

This repository is a fork of that work, migrated onto my own infrastructure and evolved into a portfolio piece.

## What I Changed

| Original PoC | This rebuild |
|---|---|
| Ran on the senior team's headscale network | Runs on my own Tailscale mesh against my homelab |
| LLM endpoint pointed at a teammate's desktop | Points at the Ollama instance on my homelab Dell XPS |
| `deepseek-coder-v2` as the analysis model | `qwen2.5-coder:7b` — better JSON-mode compliance, lower memory footprint |
| Single-source ingest (one app's logs) | Multi-source ingest across four of my projects, with per-source severity weighting |
| Token-auth plumbing for the team's LLM gateway | Removed — Ollama on the Tailnet is unauthenticated |
| Vector clients ran from team-issued machines | Logs ship through the homelab's Promtail/Loki stack and Groot reads from there |

The migration was the easy part. The interesting work is the per-project weighting and the integration back into the homelab's observability surface.

## What Groot Reads

Four projects feed Groot, in two tiers:

**Real-user projects (production-class signal):**
- **Foothold** — Social connectivity app for college students, licensed to universities.
- **Totem** — Festival companion app. Friend groups coordinate who is seeing which artist across a festival's multiple stages.

**Showcase projects (low-signal traffic):**
- **Portfolio** — Public site showcasing my work. Visitors only, no accounts.
- **Atlas** — Inverted search index from a class project, being finished out as a portfolio piece.

Groot weights findings by user impact, not log volume. A 500 error in Foothold during finals week is treated differently than a 500 from a Portfolio crawler.

## Architecture

```
Project logs (Foothold, Totem, Portfolio, Atlas)
        │  Promtail / Vector
        ▼
Loki  (on the homelab)
        │  Loki query API over Tailscale
        ▼
log-preprocessor  (Go)
        ├──►  Redis queue       — short-term work queue for the AI worker
        └──►  ClickHouse        — long-term storage, queryable from the UI

ai-core  (Python)
        │  pops one log at a time, calls Ollama, writes the result back
        ▼
        ClickHouse + UI

frontend  (Go, Echo)
        │  reads ClickHouse
        ▼
        Web UI for browsing logs and AI annotations
```

Groot does not host the model. Inference runs on the homelab's Ollama instance — Groot calls it over the Tailnet.

## Stack

| Layer | Tech |
|---|---|
| Log ingest | Go (Echo-style HTTP sink) |
| Work queue | Redis |
| Long-term storage | ClickHouse |
| Schema migrations | goose |
| AI worker | Python — Ollama HTTP client |
| LLM backend | Ollama (`qwen2.5-coder:7b`) — runs on the homelab, not in this repo |
| Web UI | Go (Echo) |
| Log shipping | Vector.dev / Promtail (configured on the homelab side) |
| Access | Tailscale — no public exposure |

## Status

Migration in progress. The PoC architecture is intact and the pipeline runs end-to-end against synthetic logs; the active work is repointing ingest to real project sources, adding per-project labels and severity weights, and deciding how Groot's findings flow back to the homelab dashboard.

## Acknowledgments

Thanks to [Erik (@StealthBadger747)](https://github.com/StealthBadger747) for sponsoring the original senior project and to the senior team I worked with on the initial PoC. The foundation came from that work; the homelab integration, multi-project ingest, and continued development is mine.
