package foundry

import (
	"fmt"
	"strings"
)

func slugify(name string) string {
	lowered := strings.NewReplacer("'", "", "’", "").Replace(strings.ToLower(name))
	var builder strings.Builder
	pendingDash := false
	for _, character := range lowered {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			if pendingDash && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(character)
			pendingDash = false
			continue
		}
		pendingDash = true
	}
	return builder.String()
}

func uniqueID(name string, taken map[string]bool) string {
	base := slugify(name)
	if base == "" {
		base = "item"
	}
	id := base
	for suffix := 2; taken[id]; suffix++ {
		id = fmt.Sprintf("%s-%d", base, suffix)
	}
	taken[id] = true
	return id
}
