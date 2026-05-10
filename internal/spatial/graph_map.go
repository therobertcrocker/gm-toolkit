package spatial

type GraphMap struct{}

func (graphMap *GraphMap) Location(_ string) (Location, bool)       { return nil, false }
func (graphMap *GraphMap) Distance(_, _ string, _ int) (int, error) { return 0, ErrNotImplemented }
