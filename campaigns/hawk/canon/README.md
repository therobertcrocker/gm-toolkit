# H.A.W.K Canon

The written canon of the H.A.W.K campaign and its setting: the fiction behind its factions, places, tags and assets. The campaign's rulebook and map are built from it. It is plain Markdown, so the Codex can take it over later.

## Files

| File | Holds |
|------|-------|
| [campaign.md](campaign.md) | The campaign, the story so far, and what the faction game does in it |
| [setting.md](setting.md) | Pitch, conflicts, history, tone, magic as technology, travel and scale |
| [glossary.md](glossary.md) | H.A.W.K terms, one entry per term |
| [places.md](places.md) | The map, its regions and its places |
| [factions.md](factions.md) | One entry per faction |
| [people.md](people.md) | Named characters |
| [tags.md](tags.md) | Tag fiction, one entry per `HAWK-T-` ID |
| [assets.md](assets.md) | Asset fiction, one entry per `HAWK-` asset ID |
| [images/myr-world-map.webp](images/myr-world-map.webp) | The labelled world map of Myr, the Material Plane |
| [images/core-sector-grid.webp](images/core-sector-grid.webp) | The Core Sector of the Astral Sea on a hex grid |
| [images/magical-traditions.webp](images/magical-traditions.webp) | The Magical Traditions chart: how each kind of magic works, on Myr and in the Astral Sea |

The images are compressed copies at full size. The original files live outside the repo.

## Entries

`campaign.md` and `setting.md` are prose under topic headings. Every other file holds one `##` entry per entity. The heading is the name, an em dash, and the entity's ID in backticks:

    ## Iron Concord — `iron-concord`

- **Factions** — the faction ID in `state/state.seed.toml`
- **Places** — the world `id` in `spatial/source/data.toml`
- **Tags and assets** — the `HAWK-` ID in the rulebook (`HAWK-T-001`, `HAWK-F1-001`)
- **Everything else** (the map, regions, terms, people) — a kebab-case slug of the name, without a leading "The"

An ID never changes once committed, even if the name does. The ID of a cut entity is never reused.

A secret the players must not learn sits in its own callout, so it can be kept out of anything player-facing:

    > **GM-only:** <the secret, in a sentence or two>

## Canon and rulebook

The canon entry holds an element's full fiction. A rulebook `description` summarizes the entry and is never its only copy. When a rulebook entry gains new fiction, its canon entry gains it first.
