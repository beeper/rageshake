#!/bin/bash
set -euo pipefail
set -x

BUCKET="rageshake-storage-20250625043023267600000001"
REGION="eu-central-1"

FILES_TO_DELETE=("ax-dumps.log.gz" "platform-imessage.log.gz")

PATHS=(
  # Put paths here on each line for rageshake reports to clean the above FILES_TO_DELETE from
  # ex: 2026-02-04/184627-QSN66FL2
)

export AWS_ACCESS_KEY_ID="$RAGESHAKE_STORAGE_IAM_ID"
export AWS_SECRET_ACCESS_KEY="$RAGESHAKE_STORAGE_IAM_SECRET"

for prefix in "${PATHS[@]}"; do
  for file in "${FILES_TO_DELETE[@]}"; do
    key="${prefix}/${file}"
    echo "Deleting: s3://$BUCKET/$key"
    aws s3 rm "s3://$BUCKET/$key" --region "$REGION"
  done
done
