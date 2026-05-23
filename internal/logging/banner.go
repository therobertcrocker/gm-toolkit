package logging

import "log/slog"

func RunHeader(log *slog.Logger, tool, version, campaign string, factions int) {
	log.Info("", "_banner", "header", "tool", tool, "version", version, "campaign", campaign, "factions", factions)
}

func TurnStart(log *slog.Logger, turn int, faction string) {
	log.Info("", "_banner", "turn", "turn", turn, "faction", faction)
}

func PhaseStart(log *slog.Logger, phase string) {
	log.Info("", "_banner", "phase", "phase", phase)
}
