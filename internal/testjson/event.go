package testjson

// Action represents a go test -json event action.
type Action string

const (
	ActionPass   Action = "pass"
	ActionFail   Action = "fail"
	ActionSkip   Action = "skip"
	ActionRun    Action = "run"
	ActionOutput Action = "output"
	ActionPause  Action = "pause"
	ActionCont   Action = "cont"
)

// TestEvent represents a single event from go test -json output.
type TestEvent struct {
	Action  Action  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

// IsPackageLevel reports whether this event applies to the package as a whole
// (not a specific test).
func (e TestEvent) IsPackageLevel() bool {
	return false
}
