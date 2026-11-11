// Package detector reads the conventions a folder follows: a pipeline file,
// a Makefile, package scripts, a scripts directory. Each convention is a
// Detector that returns a Source, the targets it declares and the command
// that runs them. Adding a convention is adding a Detector to Detectors;
// nothing else in task-ui asks which kind a target is.
//
// Discovery asks HasPipeline and Any of every folder under the root, so
// both answer the atkins pipeline by its file name: reading each pipeline
// runs atkins once a folder.
package detector
