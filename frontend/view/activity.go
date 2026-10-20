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
	// Failed counts the failed and killed runs, Other those
	// still running or stopped.
	Failed int
	Other  int
	// The heights of the parts, in percent of the chart.
	OKStyle     template.CSS
	FailedStyle template.CSS
	OtherStyle  template.CSS
	// Average is the mean duration of the bucket's finished runs, Min
	// and Max the fastest and the slowest of them, and Finished their
	// count; AverageTitle is what the duration graph's hover says.
	Average      time.Duration
	Min          time.Duration
	Max          time.Duration
	Finished     int
	AverageTitle string

	elapsed time.Duration
}

// ActivityLines is the duration graph under a count chart: a line each
// for the average, the fastest and the slowest run of the buckets, as SVG
// paths in a 100 by 100 viewBox, a point at the middle of each bucket's
// column; a bucket without finished runs sits on the baseline. Slowest is
// the top of the scale.
type ActivityLines struct {
	Average string
	Min     string
	Max     string
	Slowest time.Duration
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

// ActivityTick is a time marker under a chart: its label, and At, where
// it sits along the chart in percent, the middle of its column.
type ActivityTick struct {
	Label string
	At    string
}

// Activity is when runs ran: a bar per day of the last ActivityDays,
// a bar per hour of the day over the same days, the counts, and the
// duration line.
type Activity struct {
	Days  []ActivityBar
	Hours []ActivityBar
	// DayLines and HourLines are the durations of the days and the hours
	// as lines, nil when none of their runs finished.
	DayLines  *ActivityLines
	HourLines *ActivityLines
	// DayTicks mark a day a week, today the last; HourTicks every third
	// hour, labelled at 03:00, 09:00, 15:00 and 21:00.
	DayTicks  []ActivityTick
	HourTicks []ActivityTick
	Total     int
	Failed    int
	Median    time.Duration
	// Spent is the time the finished runs took, together.
	Spent     time.Duration
	Durations *ActivitySpark
}

// NewActivity charts the runs that started in a window of days up to now,
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
	for i := days - 1; i >= 0; i -= 7 {
		day := first.AddDate(0, 0, i)
		a.DayTicks = append([]ActivityTick{{Label: day.Format("Jan 2"), At: svgNumber((float64(i) + 0.5) * 100 / float64(days))}}, a.DayTicks...)
	}
	for h := 0; h < 24; h += 3 {
		tick := ActivityTick{At: svgNumber((float64(h) + 0.5) * 100 / 24)}
		if h%6 == 3 {
			tick.Label = twoDigits(h) + ":00"
		}
		a.HourTicks = append(a.HourTicks, tick)
	}

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
			a.Spent += run.Elapsed
			day.add(run.Elapsed)
			hour.add(run.Elapsed)
		}
	}
	if a.Total == 0 {
		return nil
	}
	scale(a.Days)
	scale(a.Hours)
	a.DayLines = lines(a.Days)
	a.HourLines = lines(a.Hours)
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

// lines draws the buckets' average, fastest and slowest durations as
// lines against the slowest run, a point at the middle of each column.
func lines(bars []ActivityBar) *ActivityLines {
	l := &ActivityLines{}
	for i := range bars {
		b := &bars[i]
		if b.Finished > 0 {
			b.Average = b.elapsed / time.Duration(b.Finished)
			l.Slowest = max(l.Slowest, b.Max)
		}
	}
	if l.Slowest == 0 && !anyFinished(bars) {
		return nil
	}
	y := func(d time.Duration) float64 {
		if l.Slowest == 0 {
			return 100
		}
		return 100 - float64(d)/float64(l.Slowest)*(100-activityHeadroom)
	}
	l.Average = linePath(bars, func(b ActivityBar) float64 { return y(b.Average) })
	l.Min = linePath(bars, func(b ActivityBar) float64 { return y(b.Min) })
	l.Max = linePath(bars, func(b ActivityBar) float64 { return y(b.Max) })
	return l
}

func anyFinished(bars []ActivityBar) bool {
	for _, b := range bars {
		if b.Finished > 0 {
			return true
		}
	}
	return false
}

// linePath draws a line through every bucket, a point at the middle of
// each column: a bucket with no finished run sits on the baseline, so the
// line runs the width of the chart. The line is a monotone cubic curve
// (Fritsch and Carlson): smooth, and it never swings past its points,
// below the baseline or over the slowest run. A bucket alone is a dot.
func linePath(bars []ActivityBar, y func(ActivityBar) float64) string {
	n := len(bars)
	if n == 0 {
		return ""
	}
	step := 100 / float64(n)
	xs, ys := make([]float64, n), make([]float64, n)
	for i, b := range bars {
		xs[i] = (float64(i) + 0.5) * step
		ys[i] = 100
		if b.Finished > 0 {
			ys[i] = y(b)
		}
	}
	point := func(x, y float64) string { return svgNumber(x) + " " + svgNumber(y) }
	if n == 1 {
		return "M " + point(xs[0], ys[0]) + " L " + point(xs[0], ys[0])
	}
	tangents := monotoneTangents(ys, step)
	var out strings.Builder
	out.WriteString("M " + point(xs[0], ys[0]))
	for i := 0; i < n-1; i++ {
		third := step / 3
		out.WriteString(" C " + point(xs[i]+third, ys[i]+tangents[i]*third) +
			" " + point(xs[i+1]-third, ys[i+1]-tangents[i+1]*third) +
			" " + point(xs[i+1], ys[i+1]))
	}
	return out.String()
}

// monotoneTangents returns the slope of the curve at each point, spaced
// step apart: zero at a peak, a valley or a flat, the harmonic mean of the
// slopes on either side elsewhere, so the curve stays between its points.
func monotoneTangents(ys []float64, step float64) []float64 {
	n := len(ys)
	slopes := make([]float64, n-1)
	for i := range slopes {
		slopes[i] = (ys[i+1] - ys[i]) / step
	}
	tangents := make([]float64, n)
	tangents[0], tangents[n-1] = slopes[0], slopes[n-2]
	for i := 1; i < n-1; i++ {
		before, after := slopes[i-1], slopes[i]
		if before*after <= 0 {
			continue
		}
		tangents[i] = 2 * before * after / (before + after)
	}
	return tangents
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
// "5 runs" and the hour, then the ones that did not end ok.
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
