package main

// Stop is the function to stop the plugin logic.
// In all cases, this function MUST leave the plugin in a state where it does not continue to function.
func Stop() (err error) {

	defer Killed.Store(true)
	if dead, ok := Killed.Load().(bool); ok {
		if dead {
			return nil
		}
	} else {
		Killed.Store(false)
	}

	// Put your plugin stop logic here!

	err = customStopLogic()

	// End your plugin stop logic here!

	if StatusChannel != nil {
		WriteStatus("Stopped module %s", err, false, PluginName)
		close(StatusChannel)
		StatusChannel = nil
	}

	return err
}

// Place all of the customer Plugin-Stop logic here.
func customStopLogic() (err error) {

	// ...

	return
}
