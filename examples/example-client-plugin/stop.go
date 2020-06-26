package main

// Stop is the function to stop the plugin logic.
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

	WriteStatus("Stopped module %s", err, false, PluginName)
	StatusLock.Lock()
	close(StatusChannel)

	return err
}

// Place all of the customer Plugin-Stop logic here.
func customStopLogic() (err error) {

	// ...

	return
}
