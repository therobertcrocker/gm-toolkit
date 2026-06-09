package chrome

// Severity ranks a status-line message for the root bottom bar. The root maps
// it to a style (Recoverable -> Warning, Fatal -> Danger).
type Severity int

const (
	Info Severity = iota
	Recoverable
	Fatal
)

// StatusLiner is implemented by sub-models that surface a status line in the
// root's bottom bar. Empty text means "no status — fall through to help".
type StatusLiner interface {
	StatusLine() (text string, severity Severity)
}
