#!/bin/bash
# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

myprg=$0
myDir=$(dirname ${myprg})

. ${myDir}/version-lib.sh

if [ -e ${rollbackFile} ]; then
    echo "Version info is already backed up."
    exit 0
fi

cp ${releaseFile} ${rollbackFile}
echo "Successfully saved system version info for rollback purposes."
exit 0
