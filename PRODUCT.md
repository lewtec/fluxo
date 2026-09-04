# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Assumed from the repo and the redesign brief, not a live interview: one operator of a self-hosted BitTorrent box. They add a magnet or `.torrent`, watch progress, start/stop/remove. They sit at a desk with the machine nearby, often in a dark room, and glance at speeds more than they browse.

## Product Purpose

Fluxo is a local BitTorrent client. Success is: add something, see it move, act on it, without another app or a GraphQL client.

## Positioning

The UI is rendered by the same Go process that holds the Rain session. Live HTML arrives over SSE. A neighboring web client that talks to an API cannot truthfully claim that.

## Operating Context

Runs on localhost (default `127.0.0.1:8080`). Session state is in-memory plus Rain’s database. Watcher ticks publish torrent and speed changes. The operator already chose to run a binary; there is no signup, no multi-user, no cloud.

## Capabilities and Constraints

Confirmed in code:

- List torrents with status, progress, speeds, ETA, peer count
- Add via magnet URI or `.torrent` file
- Detail: files, trackers, start/stop/remove
- SSE events: `stats`, `list`, `detail`, `removed`
- Detail accordions must keep open state across live ticks
- daisyUI 5 + Tailwind 4, templ, embedded CSS
- `--dev-mode` / `--dev-proxy` are ignored leftovers

Undecided (not asked): phone-first vs desk-first. Assumed desk-first, usable down to a phone width.

## Brand Commitments

Name: Fluxo. No logo file, no brand guide. Voice is short and functional (“No torrents.”, “Add”).

Standing preference (2026-09-04): look like a classical torrent index (The Pirate Bay / Kickass / Torrentz table grammar). No ads. No porn. This is a local client, not a search site.

## Evidence on Hand

Real data is the operator’s own torrents. No marketing screenshots, testimonials, or stock photography. Do not invent customer claims or benchmark numbers.

## Product Principles

- The transfer list is the product, not a dashboard around it.
- Live numbers must stay readable at a glance.
- Add, start, stop, remove stay one tap from the current job.
- Do not pretend this is a multi-tenant cloud service.

## Accessibility & Inclusion

No product-specific standard was set. Keyboard-usable forms and live regions that do not reset operator state (open accordions) are required by the current implementation.
