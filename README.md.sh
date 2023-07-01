#!/bin/bash
cat README.md.tpl
echo
echo "# Development"
echo
task -l | sed 's/:$/:\n/g'
echo

for NAME in build test run install fix; do
	echo "## $(task ${NAME} --summary)"
	echo
done
