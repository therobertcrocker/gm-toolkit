package narrative

import "github.com/therobertcrocker/gm-toolkit/internal/faction/narrative/digest"

type Renderer interface {
	Render(cycleDigest digest.CycleDigest, seed int64) (string, error)
}

func NewWireRenderer() Renderer {
	return &wireRenderer{}
}
