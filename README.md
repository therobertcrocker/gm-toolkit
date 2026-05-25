# GM Toolkit

Tools for running Stars Without Number Inspired factions and worlds.

## Bootstrap

A fresh checkout doesn't ship a campaign. Create one before running the toolkit:

    gm-toolkit campaign create my-game --path ~/games --rules ./rulebooks/swn/

This scaffolds the campaign directory, seeds it with the bundled SWN starter rulebook, and registers it as your active campaign. After this, `gm-toolkit faction` opens the TUI against your campaign.

If you receive another GM's campaign directory, register it without modifying its contents:

    gm-toolkit campaign register ~/incoming/their-campaign

See `gm-toolkit campaign --help` for the full command surface.
