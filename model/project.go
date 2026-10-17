package model

// AllowList returns the users and roles that may run the project's jobs.
// Empty allows every runner.
func (p *Project) AllowList() []string {
	if p == nil {
		return nil
	}
	return DecodeList(p.Allow)
}

// SourceList returns the conventions the project is limited to, by kind.
// Empty reads them all.
func (p *Project) SourceList() []string {
	if p == nil {
		return nil
	}
	return DecodeList(p.Sources)
}

// CommandMap returns the runner overrides per convention kind.
func (p *Project) CommandMap() map[string]string {
	if p == nil {
		return map[string]string{}
	}
	return DecodeStrings(p.Commands)
}

// JobAllowMap returns the allow lists of single jobs, by job ID.
func (p *Project) JobAllowMap() map[string][]string {
	if p == nil {
		return map[string][]string{}
	}
	return DecodeMap(p.JobAllow)
}
