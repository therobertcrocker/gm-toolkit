# Encounter Lab — Arc Concept

## In Plain Words

- **What** — a set of tools for a GM who is planning a PF2e encounter. The tools play the fight many times, show how balanced it is, and help build the flowchart the GM uses to run it at the table.
- **Why** — the first fight simulator was one script written for one fight. The second fight needed a second script built on top of the first; it took a few hours and most of a 5-hour usage block, and it still missed problems a GM would feel at the table. The tools have to be cheaper and more honest than that.
- **How it fits** — the arc lives in the `campaigns` product. It starts from `F.041` (the PF2e encounter simulator), which was planned as one feature and turned out to be several different kinds of work. `F.041` stays, narrowed, as the arc's first initiative.

## Vision

Robert plans a boss fight. He exports the creatures from Foundry, adds or changes a few abilities, and asks the tool to play the fight. The tool follows the PF2e rules in full; it is a simulation, not a projection. It shows what happened: who hit, who saved, where the damage came from, who went down, how long the fight ran.

On top of those raw facts sit measures that say something a GM can act on, such as "the party never makes this save" or "one ability carries the whole fight". Each measure is there because it was checked against evidence, not because it sounded right. The same tools hold the boss's tactics as data, so the flowchart for the table comes from the tactics the fight was tested with.

Future bosses are standard stat blocks with added or changed abilities. So a new fight costs an import, a few ability edits and a tactics file, not a new script.

The arc is worth its cost only if Robert uses the tools for fights after The Fall of Luna.

## Problem

- **One script per fight** — `cmd/scripts/hawk_finale_sim.py` simulates one fight. The second fight (`seeds/second_fight/ythrel_sim.py`) is a second script of about 750 lines. It imports the first and adds its own boss, six hard-coded strategies and its own report.
- **The scoring was invented, then retired** — the `F.041` discovery doc gives scoring about 15 lines that copy the script's threat and triumph scheme. Robert's later patches to the script (`seeds/second_fight/`) stop printing that verdict and add plain measures and a win-chance curve. Two scripts beside them compare the measures with win chance. They print rates and record no verdict, and both have fallen behind the patches, so the approach exists but no measure has passed a check yet.
- **The second fight shows what measures must catch** — Robert's critique of it: the sim likely ignores that the party never makes a save and that the party struggles to hit, and it leaned too long on one ability to carry the fight. That is not fun at the table, and the scoring did not say so.
- **One Discovery did not own the whole design** — rules coverage grew until the `F.041.E3` Plan session split it into three efforts and wrote a rules-scope doc. Description conversion still carries an open question. Scoring has no evidence behind it.
- **No simulated fight has been played yet** — the Ythrel table session on 2026-11-23 will be the first real test of what the tools say against what happens at the table.

## Why an Arc

- **The parts are different kinds of work** — rules and engine have a spec (PF2e). Description conversion is a tooling experiment. Tactics and flowcharts are design. Measurement has no spec and needs evidence. Each needs its own Discovery, and one Discovery could not hold all four.
- **Each part merges as a working tool** — sim core plays a fight and prints raw facts; measures, flowcharts and conversion each add to it and can ship on their own. The raw facts alone are not enough to prepare a fight; that takes the whole arc.
- **Some decisions bind every part** — the `F.041` discovery doc records decisions D-1 to D-26. Many of them cross the parts (data formats, the layer model, the event stream, where campaign data lives). Arc-Discovery ratifies the ones that bind the whole arc.
- **The order is a design question** — sim core comes first. Robert hopes to have the whole arc done before the Ythrel table session on 2026-11-23, so that session tests the tools; the date is a hope, not a deadline that changes the process. The session's logs are evidence for measurement, whenever they arrive. Whether tactics or conversion comes next is open.

## Candidate Initiatives

Rough and non-binding; Arc-Plan does the real slicing.

- **Sim core** (`F.041`, narrowed) — the engine, the Foundry import, full rules coverage (Efforts 1 to 5 of the current overview plan), and a basic `run` that plays a fight and prints raw facts: hit rates, save rates, damage share by source, knockdowns, rounds, a replay log
- **Measurement** — its own full Discovery; every measure is checked against evidence before it counts
- **Tactics and flowcharts** — boss and party tactics as data, and the flowchart for running the fight
- **Description conversion** — turn the prose of a custom ability into data the engine can run

Out of the arc: a boss builder. Robert wants one eventually, but it is far off.

## Trigger

`ready`
