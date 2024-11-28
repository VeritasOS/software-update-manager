#!/bin/sh
# Copyright (c) 2020 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

scriptStatus=0

echo "Pausing 10 sec for updating task messages in WebUI ..."
sleep 10

echo "Stopping VERITAS Cluster Server (VCS) service...";
/bin/systemctl stop vcs;
status=$?;
if [ ${status} -ne 0 ]; then
  echo "Failed to stop VERITAS Cluster Server (VCS) service.";
  scriptStatus=1
fi
echo "Successfully stopped VERITAS Cluster Server (VCS) service.";

# Run VERITAS Cluster Server (VCS) status to log output for debugging purposes.
echo "Display VERITAS Cluster Server (VCS) status...";
/bin/systemctl status vcs;

exit ${scriptStatus};
