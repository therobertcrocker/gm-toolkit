package tags

import "github.com/therobertcrocker/gm-toolkit/internal/faction/engine"

func RegisterDefaultTags(eng *engine.Engine) {
	eng.Tag.Register(ScavengersHandler{})
	eng.Tag.Register(WarlikeHandler{})
	eng.Tag.Register(FanaticalHandler{})
	eng.Tag.Register(PreceptorArchiveHandler{})
}
