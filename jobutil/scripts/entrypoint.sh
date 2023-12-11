# Copyright (C) Alt Research Ltd. All Rights Reserved.
#
# This source code is licensed under the limited license found in the LICENSE file
# in the root directory of this source tree.

DATADIR=${DATADIR:-/data-dir}
DONE_MARKER=/tmp/marker/done
UPLOADED_MARKER=/tmp/marker/uploaded
STARTER_SCRIPT=${STARTER_SCRIPT:-/scripts/starter.sh}
set +e
# set -x
cp $STARTER_SCRIPT ${DATADIR}/starter.sh
{ bash ${DATADIR}/starter.sh ; echo $? > $DONE_MARKER ; } | tee ${DATADIR}/workload.log
return_code=$(cat $DONE_MARKER)
echo $return_code > ${DATADIR}/workload-status
echo $return_code > ${DATADIR}/workload-status-$(date +%s)

# wait uploader to complete
while [[ ! -f $UPLOADED_MARKER ]]; do
    sleep 0.1
done
exit $return_code
