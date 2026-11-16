package engine

type Degree int

const (
	CriticalFailure Degree = iota
	Failure
	Success
	CriticalSuccess
)

func (degree Degree) String() string {
	switch degree {
	case CriticalFailure:
		return "critical failure"
	case Failure:
		return "failure"
	case Success:
		return "success"
	default:
		return "critical success"
	}
}

func DegreeOf(die, total, dc int) Degree {
	var degree Degree
	switch {
	case total >= dc+10:
		degree = CriticalSuccess
	case total >= dc:
		degree = Success
	case total <= dc-10:
		degree = CriticalFailure
	default:
		degree = Failure
	}
	switch die {
	case 20:
		degree = min(degree+1, CriticalSuccess)
	case 1:
		degree = max(degree-1, CriticalFailure)
	}
	return degree
}
