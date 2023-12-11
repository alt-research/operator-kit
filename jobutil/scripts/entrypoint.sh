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
