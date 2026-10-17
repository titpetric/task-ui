#!/bin/bash
cat README.md.tpl
echo
echo "# Development"
echo
task -l | sed 's/:$/:\n/g'
echo

function taskList {
	yq -r '.tasks | to_entries | map(select(.value.internal != true)) | .[].key' Taskfile.yml
}

for NAME in $(taskList); do
	if [ -z "$NAME" ]; then
		continue
	fi
	echo "## $(task ${NAME} --summary | awk '/^vars:/ { skip = 1; next } skip && /^$/ { skip = 0; next } !skip' | perl -p -e 's/^ - (.*)/- \`\1\`/g')"
	echo
done
