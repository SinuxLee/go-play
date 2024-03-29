#!/usr/bin/env bash
set -ue
cd "$(dirname "$0")" || exit 1

rm -rf ./*.log
daemonize -c $PWD -a -e error.log -o output.log -p daeminze.pid -l lockfile -E GOTRACEBACK=crash $PWD/server
