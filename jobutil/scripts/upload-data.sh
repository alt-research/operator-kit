set -ex
DONE_MARKER=/tmp/marker/done
UPLOADED_MARKER=/tmp/marker/uploaded

BUCKET=${BUCKET:-operator-private}
OBJECT_KEY=${OBJECT_KEY:-$POD_NAME}
DATA_S3_URI=s3://$BUCKET/$OBJECT_KEY
# https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl
OBJECT_ACL=${OBJECT_ACL:-private}
STORAGE_CLASS=${STORAGE_CLASS:-STANDARD}
DATADIR=${DATADIR:-/data-dir}

if [[ "$DATA_S3_URI" == "" ]]; then
    exit "[jobutil] DATA_S3_URI is not set"
fi

if [[ "$(which aws)" == "" ]]; then
    apt update && apt install -y unzip
    AWSCLI_DIR=/tmp/awscli
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    ./aws/install -u -i $AWSCLI_DIR -b $AWSCLI_DIR/bin
    export PATH=$AWSCLI_DIR/bin:$PATH
fi

AWS_ENDPOINT=${AWS_ENDPOINT:-}
AWS=aws
if [[ "$AWS_ENDPOINT" != "" ]]; then
    AWS="aws --endpoint-url=$AWS_ENDPOINT"
fi

echo "[jobutil] waiting workload to complete"
set +x
while [[ ! -f $DONE_MARKER ]]; do
    sleep 0.1
done
return_code=$(cat $DONE_MARKER)

echo "[jobutil] compressing and upload to s3"
set +e
set -x
tar -cz --directory=$DATADIR . | $AWS s3 cp - "$DATA_S3_URI" --storage-class $STORAGE_CLASS --acl $OBJECT_ACL
return_code=$?
echo $return_code >$UPLOADED_MARKER
echo $return_code >$DATADIR/workload-status
ls /tmp/marker
exit $return_code
