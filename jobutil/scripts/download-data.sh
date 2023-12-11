# Copyright (C) Alt Research Ltd. All Rights Reserved.
#
# This source code is licensed under the limited license found in the LICENSE file
# in the root directory of this source tree.

set -ex
DATADIR=${DATADIR:-/data-dir}
BUCKET=${BUCKET:-operator-private}
OBJECT_KEY=${OBJECT_KEY:-$POD_NAME}
DATA_S3_URI=s3://$BUCKET/$OBJECT_KEY
AWS_ENDPOINT=${AWS_ENDPOINT:-}
NEW_DATA_ON_RETRY=${NEW_DATA_ON_RETRY:-false}

if [[ "$(which aws)" == "" ]]; then
    apt update && apt install -y unzip
    AWSCLI_DIR=/tmp/awscli
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    ./aws/install -u -i $AWSCLI_DIR -b $AWSCLI_DIR/bin
    export PATH=$AWSCLI_DIR/bin:$PATH
fi

AWS=aws
if [[ "$AWS_ENDPOINT" != "" ]]; then
    AWS="aws --endpoint-url=$AWS_ENDPOINT"
fi

# skip when data exists on s3 and NEW_DATA_ON_RETRY is true
if [[ "$NEW_DATA_ON_RETRY" == "true" ]]; then
    if [[ "$DATA_S3_URI" != "" ]]; then
        if $AWS s3 ls $DATA_S3_URI; then
            echo "[jobutil] skip downloading data, cause NEW_DATA_ON_RETRY is true"
            exit 0
        fi
    fi
fi

set -x
$AWS s3 ls $DATA_S3_URI && $AWS s3 cp $DATA_S3_URI - | tar -xvzf - -C $DATADIR/
chmod -vR 777 $DATADIR
chmod -vR 777 /tmp/marker
rm -rf /tmp/marker/*
