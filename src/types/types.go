package types

type AlertmanagerSilence struct {
	Duration int
	Labels   string
	Comment  string
}
type Job struct {
	Namespace           string
	Type                string
	Workload            string
	Schedule            string
	AlertmanagerSilence *AlertmanagerSilence
}
