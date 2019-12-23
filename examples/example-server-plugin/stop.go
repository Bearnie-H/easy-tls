package main

// Stop is the function to stop the plugin logic.
func Stop() error {

	if Killed.Load().(bool) {
		return nil
	}

	Killed.Store(true)

	var err error

	// Put your plugin stop logic here!

	WriteStatus("Stopped module %s", err, false, PluginName)
	close(StatusChannel)
	return err
}
