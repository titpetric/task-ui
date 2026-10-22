package view

// SetupTrail is the footer of a step of the settings: the step before
// it and the one after, among those the user may open, and every step
// between them as a trail with the page's step current.
type SetupTrail struct {
	Previous *SetupStep
	Next     *SetupStep
	Steps    []SetupStep
}

// SetupTrail returns the footer of the page's step, or nil: for a page
// of no step, a step the user may not open, or one without another
// step the user may open.
func (p *Page) SetupTrail() *SetupTrail {
	if p.Step == "" {
		return nil
	}
	steps := p.SetupSteps()
	for i, step := range steps {
		if !step.Current {
			continue
		}
		if len(steps) < 2 {
			return nil
		}
		trail := &SetupTrail{Steps: steps}
		if i > 0 {
			trail.Previous = &steps[i-1]
		}
		if i+1 < len(steps) {
			trail.Next = &steps[i+1]
		}
		return trail
	}
	return nil
}
