package main

// Stop is the function to stop the plugin logic.
func Stop() error {

	defer Killed.Store(true)

	// ...

	return nil
}
