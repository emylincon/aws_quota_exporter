#!/bin/bash

#   6e5204f2806598c7ffb77b39686add790951a8ee
run="no"
include="yml go"
for i in $(git --no-pager diff --name-only --diff-filter=ACMRT 4637363fce013781a11448c3b29c1cb18efbe493 825fafed155c05c0c3c04a79b0123493b9d3844d); do
    extension=${i##*.}
    echo "running $i - extension=${i##*.}"
    if [[ "$include" == *"$extension"* ]]; then
        run="yes"
    fi
done

echo "run=$run"
