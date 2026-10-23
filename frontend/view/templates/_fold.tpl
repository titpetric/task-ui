{{- /* The toggle in a folding block's heading: (list id label), where id
     names the body it shows and hides. The block carries data-fold
     with the same id, and the script folds it and remembers it. */}}
{{- define "fold" }}
{{- $id := index . 0 }}
<button type="button" class="fold-toggle" aria-expanded="true" aria-controls="{{ $id }}"><svg class="ic chev"><use href="#i-chev"/></svg>{{ index . 1 }}</button>
{{- end }}
