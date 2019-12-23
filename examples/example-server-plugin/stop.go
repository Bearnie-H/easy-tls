
// Stop is the function to stop the plugin logic.
func Stop() error {

	defer Killed.Store(true)
	var err error

	// ...

	WriteStatus("Stopped module %s", err, false, PluginName)
	return err
}
