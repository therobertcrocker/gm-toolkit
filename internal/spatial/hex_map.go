package spatial

import "errors"

type HexMap struct{}

func (hexMap *HexMap) Location(_ string) (Location, bool)       { return nil, false }
func (hexMap *HexMap) Distance(_, _ string, _ int) (int, error) { return 0, errors.New("not implemented") }
