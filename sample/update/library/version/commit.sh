#!/bin/bash
# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

myprg=$0
myDir=$(dirname ${myprg})

. ${myDir}/version-lib.sh

rm -f ${rollbackFile}
echo "Successfully removed ${rollbackFile}."
exit 0
