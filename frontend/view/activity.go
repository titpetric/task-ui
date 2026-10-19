package view

import (
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
)

// ActivityDays is how many days the charts cover, today the last.
const ActivityDays = 30

// ActivityDurations is how many of the latest finished runs the
// duration line draws.
const ActivityDurations = 30

// activityStub is the least height of a bar with runs, in percent, so a
// quiet day next to a busy one still shows.
const activityStub = 4.0

// activityHeadroom keeps the longest run off the top edge of the
// duration line, in percent.
const activityHeadroom = 10.0

// ActivityRun is a run as the charts read it: when it started, how long
// it took, how it ended.
type ActivityRun struct {
	StartedAt time.Time
	Elapsed   time.Duration
	State     string
}

// ActivityBar is a bucket of the charts, a day or an hour of the day:
// its runs by how they ended, the height of each part, and the text a
// hover shows.
type ActivityBar struct {
	Label string
	Title string
	OK    int
	// Failed counts the runs that failed or were killed, Other those
	// still running or stopped.
	Failed int
	Other  int
	// The heights of the parts, in percent of the chart.
	OKStyle     template.CSS
	FailedStyle template.CSS
	OtherStyle  template.CSS
	// Average is the mean duration of the bucket's finished runs, Min
	// and Max the fastest and the slowest of them, and Finished how many
	// there were. AverageStyle is the height of its bar and RangeStyle
	// the line from Min to Max, both against the slowest Max of the
	// chart; AverageTitle what its hover says.
	Average      time.Duration
	Min          time.Duration
	Max          time.Duration
	Finished     int
	AverageStyle template.CSS
	RangeStyle   template.CSS
	AverageTitle string

	elapsed time.Duration
}

// ActivityPoint is a run on the duration line: the stretch of the plot
// that hovers it, from X as wide as Width, both in percent; At is where
// the run is on the line; Title the text a hover shows.
type ActivityPoint struct {
	X      string
	Width  string
	At     string
	Title  string
	Failed bool
}

// ActivitySpark is the duration line: the latest finished runs, oldest
// first, as SVG paths in a 100 by 100 viewBox drawn from zero to the
// longest of them.
type ActivitySpark struct {
	Line    string
	Area    string
	Points  []ActivityPoint
	Longest time.Duration
}

// Activity is when runs ran: a bar per day of the last ActivityDays,
// a bar per hour of the day over the same days, the counts, and the
// duration line.
type Activity struct {
	Days  []ActivityBar
	Hours []ActivityBar
	// Axis labels the first, middle and last day.
	Axis      []string
	Total     int
	Failed    int
	Median    time.Duration
	Durations *ActivitySpark
}

// NewActivity charts the runs that started in the days days up to now,
// counted in now's time zone. Nil when none did.
func NewActivity(runs []ActivityRun, now time.Time, days int) *Activity {
	if days < 1 {
		days = ActivityDays
	}
	if now.IsZero() {
		now = time.Now()
	}
	loc := now.Location()
	first := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1-days)
	end := first.AddDate(0, 0, days)

	a := &Activity{Days: make([]ActivityBar, days), Hours: make([]ActivityBar, 24)}
	index := map[string]int{}
	for i := range a.Days {
		day := first.AddDate(0, 0, i)
		index[day.Format(time.DateOnly)] = i
		a.Days[i].Label = day.Format("Mon 2006-01-02")
	}
	for h := range a.Hours {
		a.Hours[h].Label = strconv.Itoa(h)
	}
	a.Axis = []string{a.Days[0].Label[4:], a.Days[days/2].Label[4:], a.Days[days-1].Label[4:]}

	var finished []ActivityRun
	for _, run := range runs {
		at := run.StartedAt.In(loc)
		if at.Before(first) || !at.Before(end) {
			continue
		}
		a.Total++
		day := &a.Days[index[at.Format(time.DateOnly)]]
		hour := &a.Hours[at.Hour()]
		switch run.State {
		case model.StateOK:
			day.OK++
			hour.OK++
		case model.StateFailed, model.StateKilled:
			day.Failed++
			hour.Failed++
			a.Failed++
		default:
			day.Other++
			hour.Other++
		}
		if run.State != model.StateRunning {
			finished = append(finished, run)
			day.add(run.Elapsed)
			hour.add(run.Elapsed)
		}
	}
	if a.Total == 0 {
		return nil
	}
	scale(a.Days)
	scale(a.Hours)
	average(a.Days)
	average(a.Hours)
	for i := range a.Days {
		a.Days[i].Title = counts(a.Days[i], a.Days[i].Label)
		a.Days[i].AverageTitle = averageText(a.Days[i], a.Days[i].Label)
	}
	for h := range a.Hours {
		when := twoDigits(h) + ":00–" + twoDigits(h+1) + ":00"
		a.Hours[h].Title = counts(a.Hours[h], when)
		a.Hours[h].AverageTitle = averageText(a.Hours[h], when)
	}
	sort.SliceStable(finished, func(i, j int) bool { return finished[i].StartedAt.Before(finished[j].StartedAt) })
	a.Median = median(finished)
	if len(finished) > ActivityDurations {
		finished = finished[len(finished)-ActivityDurations:]
	}
	a.Durations = spark(loc, finished)
	return a
}

// scale sets the heights of the bars against the busiest of them; a bar
// with runs is at least activityStub tall.
func scale(bars []ActivityBar) {
	busiest := 0
	for _, b := range bars {
		busiest = max(busiest, b.OK+b.Failed+b.Other)
	}
	if busiest == 0 {
		return
	}
	for i := range bars {
		b := &bars[i]
		total := b.OK + b.Failed + b.Other
		if total == 0 {
			continue
		}
		unit := 100 / float64(busiest)
		if float64(total)*unit < activityStub {
			unit = activityStub / float64(total)
		}
		b.OKStyle = height(float64(b.OK) * unit)
		b.FailedStyle = height(float64(b.Failed) * unit)
		b.OtherStyle = height(float64(b.Other) * unit)
	}
}

// add counts a finished run's duration into the bar.
func (b *ActivityBar) add(elapsed time.Duration) {
	if b.Finished == 0 || elapsed < b.Min {
		b.Min = elapsed
	}
	b.Max = max(b.Max, elapsed)
	b.Finished++
	b.elapsed += elapsed
}

// average sets each bar's mean duration, its bar and the line from its
// fastest to its slowest run, all against the slowest run of the chart;
// a bar with finished runs is at least activityStub tall.
func average(bars []ActivityBar) {
	var slowest time.Duration
	for i := range bars {
		b := &bars[i]
		if b.Finished > 0 {
			b.Average = b.elapsed / time.Duration(b.Finished)
			slowest = max(slowest, b.Max)
		}
	}
	for i := range bars {
		b := &bars[i]
		if b.Finished == 0 {
			continue
		}
		if slowest == 0 {
			b.AverageStyle = height(activityStub)
			continue
		}
		of := func(d time.Duration) float64 { return float64(d) / float64(slowest) * 100 }
		b.AverageStyle = height(max(of(b.Average), activityStub))
		low, high := of(b.Min), of(b.Max)
		b.RangeStyle = template.CSS("bottom:" + svgNumber(low) + "%;height:" + svgNumber(high-low) + "%")
	}
}

// averageText says a bar's durations and when, for its hover.
func averageText(b ActivityBar, when string) string {
	if b.Finished == 0 {
		return "no finished runs, " + when
	}
	if b.Finished == 1 {
		return Duration(b.Average) + ", 1 run, " + when
	}
	return "avg " + Duration(b.Average) + " · min " + Duration(b.Min) + " · max " + Duration(b.Max) + " over " + plural(b.Finished, "run") + ", " + when
}

// counts says what a bar holds and when, for its hover: the runs first,
// "5 runs, 09:00–10:00", then the ones that did not end ok.
func counts(b ActivityBar, when string) string {
	total := b.OK + b.Failed + b.Other
	if total == 0 {
		return "no runs, " + when
	}
	text := plural(total, "run") + ", " + when
	if b.Failed > 0 {
		text += " · " + strconv.Itoa(b.Failed) + " failed"
	}
	if b.Other > 0 {
		text += " · " + strconv.Itoa(b.Other) + " running or stopped"
	}
	return text
}

// median is the middle duration of the runs, the mean of the middle two
// for an even count.
func median(runs []ActivityRun) time.Duration {
	if len(runs) == 0 {
		return 0
	}
	d := make([]time.Duration, len(runs))
	for i, run := range runs {
		d[i] = run.Elapsed
	}
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	mid := len(d) / 2
	if len(d)%2 == 0 {
		return (d[mid-1] + d[mid]) / 2
	}
	return d[mid]
}

// spark draws the runs' durations as a line, a run a point spread evenly
// from left to right, and the plot under it as a wash. A single run is
// a flat line across.
func spark(loc *time.Location, runs []ActivityRun) *ActivitySpark {
	if len(runs) == 0 {
		return nil
	}
	s := &ActivitySpark{}
	for _, run := range runs {
		s.Longest = max(s.Longest, run.Elapsed)
	}
	xs := make([]float64, len(runs))
	ys := make([]float64, len(runs))
	for i, run := range runs {
		xs[i] = 50
		if len(runs) > 1 {
			xs[i] = float64(i) * 100 / float64(len(runs)-1)
		}
		ys[i] = 100
		if s.Longest > 0 {
			ys[i] = 100 - float64(run.Elapsed)/float64(s.Longest)*(100-activityHeadroom)
		}
	}
	var line strings.Builder
	if len(runs) == 1 {
		line.WriteString("M 0 " + svgNumber(ys[0]) + " L 100 " + svgNumber(ys[0]))
	} else {
		for i := range runs {
			cmd := " L "
			if i == 0 {
				cmd = "M "
			}
			line.WriteString(cmd + svgNumber(xs[i]) + " " + svgNumber(ys[i]))
		}
	}
	s.Line = strings.TrimSpace(line.String())
	if len(runs) == 1 {
		s.Area = "M 0 100 L 0 " + svgNumber(ys[0]) + " L 100 " + svgNumber(ys[0]) + " L 100 100 Z"
	} else {
		s.Area = "M 0 100 L" + strings.TrimPrefix(s.Line, "M") + " L 100 100 Z"
	}
	for i, run := range runs {
		from, to := 0.0, 100.0
		if i > 0 {
			from = (xs[i-1] + xs[i]) / 2
		}
		if i < len(runs)-1 {
			to = (xs[i] + xs[i+1]) / 2
		}
		failed := run.State == model.StateFailed || run.State == model.StateKilled
		s.Points = append(s.Points, ActivityPoint{
			X:      svgNumber(from),
			Width:  svgNumber(to - from),
			At:     svgNumber(xs[i]),
			Title:  Duration(run.Elapsed) + ", " + run.StartedAt.In(loc).Format("Mon 2006-01-02 15:04") + " · " + run.State,
			Failed: failed,
		})
	}
	return s
}

// height is a bar part's CSS height.
func height(percent float64) template.CSS {
	return template.CSS("height:" + svgNumber(percent) + "%")
}

// svgNumber writes a coordinate or a percentage of the 100 by 100 plot.
func svgNumber(v float64) string {
	v = min(max(v, 0), 100)
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
