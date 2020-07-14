#!/bin/bash
set -e

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

echo "Running script..."

service=$(kubectl get pod -A -o custom-columns=name:metadata.name,namespace:metadata.namespace | grep vault-0 | awk '{print $1}')
namespace=$(kubectl get pod -A -o custom-columns=name:metadata.name,namespace:metadata.namespace | grep vault-0 | awk '{print $2}')

#kubectl exec $service -n $namespace -- sh -c 'vault secrets enable -path=internal kv-v2'
#kubectl exec $service -n $namespace -- sh -c 'vault kv put internal/database/config username="db-readonly-username" password="db-secret-password"'
#kubectl exec $service -n $namespace -- sh -c 'vault auth enable kubernetes'
#kubectl exec $service -n $namespace -- sh -c 'vault write auth/kubernetes/config \
#                                         token_reviewer_jwt="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
#                                         kubernetes_host="https://$KUBERNETES_PORT_443_TCP_ADDR:443" \
#                                         kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'

##
# End Main
##
echo "$new_sum_value" > /tmp/$sum_name
gsutil cp /tmp/$sum_name gs://$BUCKET_NAME/$sum_name
rm /tmp/$sum_name
