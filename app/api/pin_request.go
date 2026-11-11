package api

// PinRequest is what pins a folder, or a job of it, to the sidebar.
type PinRequest struct {
	// Folder is a folder's ID, or a directory's inside one.
	Folder string `json:"folder"`
	// Job pins a job of the folder; empty pins the folder.
	Job string `json:"job"`
	// Dir is the directory inside the folder the job runs in, as a run's.
	Dir string `json:"dir"`
	// Args are what the job runs with after --, for a job that takes args.
	Args RunArgs `json:"args"`
}
