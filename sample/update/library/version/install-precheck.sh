#!/bin/bash
# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

myprg=$0
myDir=$(dirname ${myprg})

. ${myDir}/version-lib.sh

if [ -e ${rollbackFile} ]; then
    echo "An earlier update is in-progress. Commit or roll back the update before installing a new one."
    exit 1
fi
