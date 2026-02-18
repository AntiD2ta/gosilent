package testjson

// Action represents a go test -json event action.
type Action string

const (
	ActionPass        Action = "pass"
	ActionFail        Action = "fail"
	ActionSkip        Action = "skip"
	ActionRun         Action = "run"
	ActionOutput      Action = "output"
	ActionPause       Action = "pause"
	ActionCont        Action = "cont"
	ActionBuildOutput Action = "build-output"
	ActionBuildFail   Action = "build-fail"
	ActionArtifacts   Action = "artifacts"
)

// TestEvent represents a single event from go test -json output.
type TestEvent struct {
	Action     Action  `json:"Action"`
	Package    string  `json:"Package"`
	Test       string  `json:"Test"`
	Output     string  `json:"Output"`
	Elapsed    float64 `json:"Elapsed"`
	ImportPath string  `json:"ImportPath"`
}

// IsPackageLevel reports whether this event applies to the package as a whole
// (not a specific test).
func (e TestEvent) IsPackageLevel() bool {
	return e.Test == ""
}
