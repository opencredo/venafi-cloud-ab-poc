#!/bin/bash

if [[ -z "$BUCKET_NAME" ]]; then
  BUCKET_NAME="vault_script_state"
fi
echo "Checking state in $BUCKET_NAME"

gsutil ls -b gs://$BUCKET_NAME || gsutil mb gs://$BUCKET_NAME

sum_name="$(basename $0).sum"

gsutil cp gs://$BUCKET_NAME/$sum_name /tmp/$sum_name >& /dev/null
if [[ $? > 0 ]]; then
  echo "No existing hash"
  touch /tmp/$sum_name
fi

old_sum_value=$(cat /tmp/$sum_name)
new_sum_value=$(sha256sum $0 | cut -b -64)

if [[ "$old_sum_value" == "$new_sum_value" ]]; then
  echo "This version of the script has already been run successfully"
  exit
fi

##
# Main
##

echo "Doing nothing"

##
# End Main
##
echo "$new_sum_value" > /tmp/$sum_name
gsutil cp /tmp/$sum_name gs://$BUCKET_NAME/$sum_name
rm /tmp/$sum_name
