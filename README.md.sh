#!/bin/bash
cat README.md.tpl
echo
echo "# Development"
echo
echo "The packages, the lint and test bar and the screenshots are in [docs/development.md](docs/development.md). The tasks:"
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
	# vars: and env: are how a task is wired, not what it is for: the
	# summary prints both and a reader of the README wants neither.
	echo "## $(task ${NAME} --summary | awk '/^vars:/ || /^env:/ { skip = 1; next } skip && /^$/ { skip = 0; next } !skip' | perl -p -e 's/^ - (.*?)\s*$/- \`\1\`/')"
	echo
done
