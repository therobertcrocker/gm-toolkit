# [Subsystem Name]

<!--
COPY THIS FILE TO: docs/architecture/<side>/<subsystem>.md
  where <side> is `engine` or `interface`.

An architecture page is the durable home for one subsystem's "what it is" and
"why it's shaped this way." It is the primary read surface of the push-not-pull
decision system (session-modes.md section 10): Discovery and Plan sessions open
the relevant page during re-grounding, and pre-merge promotion appends durable
decisions to Key Decisions here.

Routing rule: anything about a single subsystem lives on that subsystem's page;
only narrative that spans subsystems may live in the side's overview.md spine.

Prose-first. Reach for a small code signature only where it carries more than
the sentence would. No status line, no last-verified line — the page is current
because we keep it current.

Delete these comments before writing.
-->

> **Code:** `internal/<path>` (`, internal/<path>`…)

<!--
REQUIRED. The code-map header: the package(s) this page describes, as inline
backticked paths. A reader following the page to source starts here.
-->

## Purpose

<!--
REQUIRED. 2–5 sentences: what this subsystem is for and the one-line version of
how it does it. A Discovery reader should know after this section whether this
is the page they need.
-->

## Shape

<!--
REQUIRED. The structural story: the package's key types/interfaces, the data
flow through them, and the boundaries with neighbours. Prose-first; small code
signatures only where they carry more than prose. This is the longest section.

Leave a `<!-- diagram: ... -->` breadcrumb where a diagram is wanted (Effort 5
fills these); do not draw it here.
-->

## [Subsystem] Catalogue

<!--
OPTIONAL — data-backed pages only. Pages whose subsystem is driven by rulebook
data (actions → emitted mutations, goals, the mutation vocabulary) carry a
catalogue: a table enumerating the data instances, with a pointer to
docs/rules/swn-faction-mechanics.md for the rules behind the numbers. Keeps the prose
at structural altitude. Omit where there's no rulebook data behind the page
(interface pages, orchestrator, persistence).
-->

## Key Decisions

<!--
REQUIRED. The durable "why this subsystem is shaped this way," topic-organized.
One bold lead-in per decision + rationale. This is the surface pre-merge
promotion appends to. Provenance links (plan docs, frozen-log decision numbers)
welcome but optional.
-->

- **<Decision>** — <rationale>.

## Dependencies

<!--
REQUIRED. Both axes, briefly: what this subsystem depends on and what depends
on it (code axis), each named subsystem linked to its page where one exists
(doc axis). A short prose paragraph or two split lists ("Depends on" / "Depended
on by") — whichever reads cleaner for the subsystem.
-->
