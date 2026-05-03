package history

import (
	"fmt"
	"time"

	"github.com/therobertcrocker/gm-toolkit/internal/faction/domain"
	"github.com/therobertcrocker/gm-toolkit/internal/faction/state"
)

func buildEventRecord(factionState *state.FactionState, faction *domain.Faction, mutations []domain.Mutation) (domain.EventRecord, error) {
	records := make([]domain.MutationRecord, 0, len(mutations))
	for _, mutation := range mutations {
		record, err := domain.NewMutationRecord(mutation)
		if err != nil {
			return domain.EventRecord{}, fmt.Errorf("building mutation record: %w", err)
		}
		records = append(records, record)
	}
	return domain.EventRecord{
		Cycle:     factionState.CycleNumber,
		FactionID: faction.ID,
		Timestamp: time.Now().UTC(),
		Mutations: records,
	}, nil
}
