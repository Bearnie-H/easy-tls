package main

// Stop is the function to stop the plugin logic.
func Stop() (err error) {

	defer Killed.Store(true)
	if Killed.Load().(bool) {
		return nil
	}

	// Put your plugin stop logic here!

	err = customStopLogic()

	// End your plugin stop logic here!

	WriteStatus("Stopped module %s", err, false, PluginName)
	close(StatusChannel)
	return err
}

// Place all of the customer Plugin-Stop logic here.
func customStopLogic() (err error) {

	// ...

	return
}
