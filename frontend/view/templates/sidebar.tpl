{{ define "content" }}
{{- $admin := and .User .User.CanAdmin }}
<div class="page-head">
	<div>
		<h2>Sidebar</h2>
		<p>The sidebar's top block, in the order it lists them. Drag a row by its grip, or focus the grip and press the up and down arrows. <strong>Shown</strong> leaves an item out of the sidebar without deleting it: the page behind it still answers, and its permission still decides who may open it.{{ if not $admin }} An admin saves the order.{{ end }}</p>
	</div>
</div>

<form class="form sidebar-form" {{ if $admin }}data-api="/api/sidebar" data-method="PUT" data-next="reload" data-notice="Sidebar saved" data-sidebar{{ end }} novalidate>
<section class="block">
	<h3>The items <small>a row per link, the order the sidebar reads</small></h3>
	<div class="table-wrap">
	<table class="table sidebar-rows"{{ if $admin }} data-sort-rows{{ end }}>
		<thead><tr><th class="num">Order</th><th>Item</th><th>Link</th><th class="num">Shown</th><th></th></tr></thead>
		<tbody>
		{{- range $i, $row := .SidebarRows }}
			<tr data-row data-href="{{ $row.Item.Href }}">
				<td class="num">
					{{- if $admin }}
					<button type="button" class="grip" data-row-grip title="Drag to move the item, or focus it and press the up and down arrows" aria-label="Move {{ $row.Title }}"><svg class="ic"><use href="#i-grip"/></svg></button>
					{{- end }}
					<span class="mono" data-row-num>{{ add $i 1 }}</span>
				</td>
				<td class="key">
					<svg class="ic"><use href="#i-{{ $row.Item.Icon }}"/></svg>
					{{ $row.Title }}
					{{- with $row.Note }}<small class="faint">{{ . }}</small>{{ end }}
				</td>
				<td class="mono"><a href="{{ $row.Item.Href }}">{{ $row.Item.Href }}</a></td>
				{{/* The switch carries no name: serialize splits a field name
				     on a dot into a map, and a module's link can hold one.
				     sidebarBody reads the switches out of the DOM instead,
				     keyed by the link, and data-row-shown remembers the state
				     the page loaded with so the foot can count what changed. */}}
				<td class="num">
					<label class="switch" title="{{ if $row.Item.Hidden }}Left out of the sidebar{{ else }}Listed in the sidebar{{ end }}">
						<input type="checkbox" data-row-shown="{{ if $row.Item.Hidden }}false{{ else }}true{{ end }}"{{ if not $row.Item.Hidden }} checked{{ end }}{{ if not $admin }} disabled{{ end }} aria-label="Show {{ $row.Title }} in the sidebar">
						<i aria-hidden="true"></i>
					</label>
				</td>
				<td class="row-actions">
					{{- if $admin }}
					{{ template "menu" (menu (print "Actions on " $row.Title)
						(detail "Link" $row.Item.Href "mono" "")
						(detail "Kind" $row.Item.Kind)
						(item "up" "Move up" "data-row-move" "up")
						(item "down" "Move down" "data-row-move" "down")) }}
					{{- end }}
				</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- if $admin }}
	<div class="form-foot">
		<span class="faint" data-sidebar-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
	<p class="wf-live" aria-live="polite" data-rows-live></p>
	{{- end }}
</section>
</form>
{{ end }}
