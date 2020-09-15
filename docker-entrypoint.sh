#!/bin/bash
set -e

if [ "${1:0:1}" = '-' ] || [ "${1:0:1}" = "{" ]
then
  set -- /harpocrates "$@"
fi

exec "$@"