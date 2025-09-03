#!/bin/sh

echo "Starting app"
/bin/sh -c $@
exit_code=$?
echo "App exit code: ${exit_code}"
exit ${exit_code}