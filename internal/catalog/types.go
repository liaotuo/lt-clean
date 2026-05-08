package catalog

// SafetyLevel classifies the risk of cleaning an item.
type SafetyLevel int

const (
	Safe SafetyLevel = iota
	Costly
	Destructive
)

func (s SafetyLevel) String() string {
	switch s {
	case Safe:
		return "Safe"
	case Costly:
		return "Costly"
	case Destructive:
		return "Destructive"
	}
	return "Unknown"
}

// ActionKind identifies the action variant.
type ActionKind int

const (
	ActRmDir ActionKind = iota
	ActRmGlobInDir
	ActCmd
	ActMultiPath
	ActDsStoreSweep
)

// Action describes how to clean a specific Item.
//
// Field usage by Kind:
//
//	ActRmDir         Paths[0] is the directory to remove.
//	ActRmGlobInDir   Dir is the directory; Exts are extensions to delete.
//	ActCmd           Program + Args run as a subprocess.
//	ActMultiPath     Paths get RmDir; if Program != "", it also runs.
//	ActDsStoreSweep  Paths[0] is the base for "find ... -name .DS_Store -delete".
type Action struct {
	Kind    ActionKind
	Paths   []string
	Dir     string
	Exts    []string
	Program string
	Args    []string
}

// Item is one cleanable entry in the catalog.
type Item struct {
	ID        string
	Group     string
	Title     string
	Hint      string // one-line plain-language description shown in TUI footer
	Level     SafetyLevel
	SizePaths []string
	Action    Action
	Probe     func(*Item) bool
}

// Available reports whether this item is applicable on the current machine.
func (i *Item) Available() bool {
	if i.Probe == nil {
		return true
	}
	return i.Probe(i)
}
