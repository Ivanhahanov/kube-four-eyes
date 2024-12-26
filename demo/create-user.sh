set -e
USERNAME=${1:-}
USERS_CERTS_DIR=".users"
CLUSTER_NAME="kind-auth"
# Creating users
mkdir -p .users/
openssl genrsa -out $USERS_CERTS_DIR/$USERNAME.key 2048
openssl req -new -key $USERS_CERTS_DIR/$USERNAME.key -out $USERS_CERTS_DIR/$USERNAME.csr -subj "/CN=$USERNAME"

cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: $USERNAME-csr
spec:
  request: $(cat $USERS_CERTS_DIR/$USERNAME.csr | base64 | tr -d '\n')
  signerName: kubernetes.io/kube-apiserver-client
  usages:
  - client auth
EOF

kubectl certificate approve $USERNAME-csr
kubectl get csr $USERNAME-csr -o jsonpath='{.status.certificate}' | base64 -d > $USERS_CERTS_DIR/$USERNAME.crt

kubectl config set-credentials $USERNAME --client-certificate=$USERS_CERTS_DIR/$USERNAME.crt --client-key=$USERS_CERTS_DIR/$USERNAME.key
kubectl config set-context $USERNAME-context --cluster=$CLUSTER_NAME --user=$USERNAME
# kubectl config use-context $USERNAME-context


