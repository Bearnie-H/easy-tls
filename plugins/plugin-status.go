package plugins

import "fmt"

// PluginStatus represents a single status message from a given EasyTLS-compliant plugin.
type PluginStatus struct {
	Message string
	Error   error
	IsFatal bool
}

func (S PluginStatus) String() string {
	switch {
	case S.IsFatal:
		return fmt.Sprintf("FATAL ERROR: %s - %s", S.Message, S.Error)
	case S.Error != nil:
		return fmt.Sprintf("Warning: %s - %s", S.Message, S.Error)
	default:
		return S.Message
	}
}
