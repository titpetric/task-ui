<?php $_v=&$this->vars;?>
	<div class="field">
		<label for="sch-every">Repeat</label>
		<select id="sch-every" data-sched-mode>
			<option value="hourly">Hourly</option>
			<option value="daily" selected>Daily</option>
			<option value="weekly">Weekly</option>
			<option value="monthly">Monthly</option>
			<option value="expression">Expression</option>
		</select>
		<small class="faint">Expression leaves the five fields to you.</small>
	</div>
	<div class="field" data-sched-when="hourly">
		<label for="sch-minute">At minute</label>
		<input type="number" id="sch-minute" min="0" max="59" step="1" value="0" data-sched-minute>
		<small class="faint">Past every hour, 0 to 59.</small>
	</div>
	<div class="field" data-sched-when="daily weekly monthly">
		<label for="sch-time">At</label>
		<input type="time" id="sch-time" value="03:00" data-sched-time>
		<small class="faint">The server's clock, not yours.</small>
	</div>
	<div class="field whole" data-sched-when="weekly">
		<span class="label">On these days</span>
		<div class="checks">
			<label class="check"><input type="checkbox" value="1" data-sched-day> Mon</label>
			<label class="check"><input type="checkbox" value="2" data-sched-day> Tue</label>
			<label class="check"><input type="checkbox" value="3" data-sched-day> Wed</label>
			<label class="check"><input type="checkbox" value="4" data-sched-day> Thu</label>
			<label class="check"><input type="checkbox" value="5" data-sched-day> Fri</label>
			<label class="check"><input type="checkbox" value="6" data-sched-day> Sat</label>
			<label class="check"><input type="checkbox" value="0" data-sched-day> Sun</label>
			<button class="btn btn-sm" type="button" data-sched-days="1,2,3,4,5">Weekdays</button>
			<button class="btn btn-sm" type="button" data-sched-days="0,6">Weekend</button>
			<button class="btn btn-sm" type="button" data-sched-days="">Every day</button>
		</div>
	</div>
	<div class="field whole" data-sched-when="monthly">
		<span class="label">On this day of the month</span>
		<div class="checks">
			<input type="number" id="sch-day" min="1" max="28" step="1" value="1" aria-label="Day of the month" data-sched-dom>
			<button class="btn btn-sm" type="button" data-sched-pick="1">First</button>
			<button class="btn btn-sm" type="button" data-sched-pick="15">Middle</button>
			<button class="btn btn-sm" type="button" data-sched-pick="28">Last</button>
			<span class="faint">1 to 28, the days every month has.</span>
		</div>
	</div>
	<div class="field whole">
		<label for="sch-expr">Expression</label>
		<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required data-sched-expr>
		<small class="faint" data-sched-says></small>
	</div>
	<div class="field whole">
		<div class="table-wrap fit">
		<table class="table fit-cards sched-legend">
			<thead><tr><th>Position</th><th>Field</th><th>Range</th><th>This one</th><th>What it says</th></tr></thead>
			<tbody>
				<tr><td class="key" data-label="Position">1</td><td data-label="Field">minute</td><td data-label="Range"><code>0-59</code></td><td data-label="This one"><code data-sched-pos="0"></code></td><td data-label="What it says">The minute of the hour it starts on.</td></tr>
				<tr><td class="key" data-label="Position">2</td><td data-label="Field">hour</td><td data-label="Range"><code>0-23</code></td><td data-label="This one"><code data-sched-pos="1"></code></td><td data-label="What it says">The hour of the day, on the server's clock.</td></tr>
				<tr><td class="key" data-label="Position">3</td><td data-label="Field">day of the month</td><td data-label="Range"><code>1-31</code></td><td data-label="This one"><code data-sched-pos="2"></code></td><td data-label="What it says">The date. A day past the end of a short month is skipped that month.</td></tr>
				<tr><td class="key" data-label="Position">4</td><td data-label="Field">month</td><td data-label="Range"><code>1-12</code></td><td data-label="This one"><code data-sched-pos="3"></code></td><td data-label="What it says">The month, by number or by name, <code>jan</code> to <code>dec</code>.</td></tr>
				<tr><td class="key" data-label="Position">5</td><td data-label="Field">day of the week</td><td data-label="Range"><code>0-6</code></td><td data-label="This one"><code data-sched-pos="4"></code></td><td data-label="What it says">Sunday is 0, and 7 is Sunday too, or by name, <code>sun</code> to <code>sat</code>.</td></tr>
			</tbody>
		</table>
		</div>
		<small class="faint"><code>*</code> is every, <code>*/5</code> every fifth, <code>1-5</code> a range and <code>1,15</code> a list. A day of the month and a day of the week both named must both hold, so <code>0 3 1 * mon</code> is the 1st when the 1st is a Monday. The whole-period names are <code>@hourly</code>, <code>@daily</code>, <code>@weekly</code>, <code>@monthly</code> and <code>@yearly</code>; <code>@midnight</code> is <code>@daily</code>.</small>
	</div>
	<div class="form-foot">
		<button class="btn" type="button" data-sched-cancel hidden>Cancel</button>
		<button class="btn btn-primary" type="submit"><svg class="ic" data-sched-icon><use href="#i-plus"/></svg><span data-sched-label>Add schedule</span></button>
	</div>
