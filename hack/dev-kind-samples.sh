#!/usr/bin/env bash

# scripts must be run from project root
. hack/1-bin.sh || exit 1

# main

"$KUBECTL" delete -f config/samples || true
"$KUBECTL" apply -f config/samples
