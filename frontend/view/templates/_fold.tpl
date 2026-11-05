{{- define "fold" }}
{{- $id := index . 0 }}
<button type="button" class="fold-toggle" aria-expanded="true" aria-controls="{{ $id }}"><svg class="ic chev"><use href="#i-chev"/></svg>{{ index . 1 }}</button>
{{- end }}
