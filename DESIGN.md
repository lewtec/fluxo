# Design

<!-- impeccable:design-schema 1 -->

## World

Classical torrent-index table (The Pirate Bay / Kickass / Torrentz grammar) applied to a local client. No ads. Black `bay` theme by default; pale `listing` as the day toggle.

## Type

Verdana, Geneva, Tahoma, sans-serif. One size scale. Tabular numbers in tables.

## Color

- Ground: black (`bay`) or #f6f6f6 (`listing`)
- Wordmark: `#ee0000` (primary, once)
- Seeds / download: `#00cc00`
- Links: default daisyUI `link`
- Radius: 0

## Components

daisyUI 5: `navbar`, `table`, `link`, `progress`, `fieldset`, `textarea`, `file-input`, `collapse`, `alert`, `btn`, `swap`, `theme-controller`.

## Surfaces

- `/` — transfer table
- `/add` — magnet or file form
- `/torrents/{id}` — definition table plus file/tracker collapses

## Motion

None except swap/theme. Live updates patch table rows; collapse checkboxes stay mounted.
