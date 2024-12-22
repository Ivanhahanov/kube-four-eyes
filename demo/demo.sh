set -e

DEX_HOSTNAME=dex.local
TEMP_SSL_DIR=".ssl"
CLUSTER_NAME="auth"
# bcrypt hash of the string "password"
DEMO_PASSWORD='$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W'

echo "# do some clean up"
rm -rf ${TEMP_SSL_DIR}

echo "# create a folder to store certificates"
mkdir -p ${TEMP_SSL_DIR}

echo "# generate an rsa key"
openssl genrsa -out .ssl/root-ca-key.pem 2048

echo "# generate root certificate"
openssl req -x509 -new -nodes -key .ssl/root-ca-key.pem \
  -days 3650 -sha256 -out .ssl/root-ca.pem -subj "/CN=kube-ca"

if kind get clusters | grep -q "$CLUSTER_NAME"; then
  echo "Kind cluster $CLUSTER_NAME exists. Deleting..."

  # Delete the kind cluster
  kind delete cluster --name "$CLUSTER_NAME"

  echo "Kind cluster $CLUSTER_NAME deleted."
fi

kind create cluster --name $CLUSTER_NAME --image kindest/node:v1.32.0 --config  - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        oidc-client-id: kube
        oidc-issuer-url: https://${DEX_HOSTNAME}
        oidc-username-claim: email
        oidc-groups-claim: groups
        oidc-username-prefix: "oidc:"
        oidc-ca-file: /etc/ca-certificates/custom/root-ca.pem
        authorization-config: /etc/configs/authorization-config.yml
      extraVolumes:
        - name: configs
          hostPath: /etc/configs
          mountPath: /etc/configs
          readOnly: true
          pathType: "DirectoryOrCreate"
  extraMounts:
  - hostPath: $PWD/.ssl/root-ca.pem
    containerPath: /etc/ca-certificates/custom/root-ca.pem
    readOnly: true
  - hostPath: $PWD/configs
    containerPath: /etc/configs
    readOnly: true
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
    listenAddress: "0.0.0.0"
  - containerPort: 443
    hostPort: 443
    protocol: TCP
    listenAddress: "0.0.0.0"
EOF

echo "# Create a kubernetes secret containing the Root CA certificate and its key"
kubectl create ns cert-manager
kubectl create secret tls -n cert-manager ca-key-pair \
  --cert=.ssl/root-ca.pem \
  --key=.ssl/root-ca-key.pem

echo "# Install ingress controller"
kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml

echo "# Deploy the certificate manager"
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager --namespace cert-manager jetstack/cert-manager --set crds.enabled=true

echo "# Create an Issuer where you specify the secret: ca-key-pair"
cat <<EOF | kubectl apply -f -
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: ca-issuer
spec:
  ca:
    secretName: ca-key-pair
EOF

kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

helm upgrade --install dex dex --wait --timeout 15m \
  --namespace dex --create-namespace \
  --repo https://charts.dexidp.io \
  --values - <<EOF
config:
  issuer: https://${DEX_HOSTNAME}

  storage:
    type: memory

  staticPasswords:
  - email: "admin@example.com"
    hash: "$DEMO_PASSWORD"
    username: "admin"
    userID: "08a8684b-db88-4b73-90a9-3cd1661f5461"
  - email: "user1@example.com"
    hash: "$DEMO_PASSWORD"
    username: "user1"
    userID: "08a8684b-db88-4b73-90a9-3cd1661f5462"
  - email: "user2@example.com"
    hash: "$DEMO_PASSWORD"
    username: "user2"
    userID: "08a8684b-db88-4b73-90a9-3cd1661f5463"

  enablePasswordDB: true

  # oauth2:
  #   skipApprovalScreen: true
  #   passwordConnector: local

  expiry:
    signingKeys: "4h"
    idTokens: "1h"

  staticClients:
  - id: oauth2-proxy
    redirectURIs:
    - 'https://auth.dev.local/oauth2/callback'
    name: 'OAuth2 Proxy'
    secret: b2F1dGgyLXByb3h5LWNsaWVudC1zZWNyZXQK

  - id: kube
    redirectURIs:
    - http://localhost:8000
    name: 'Kubernetes'
    secret: ZXhhbXBsZS1hcHAtc2VjcmV0

ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: ca-issuer
  className: "nginx"
  hosts:
    - host: ${DEX_HOSTNAME}
      paths:
        - path: /
          pathType: ImplementationSpecific
  tls:
    - secretName: dex-tls
      hosts:
        - ${DEX_HOSTNAME}
EOF

echo "# Setup ClusterRoleBinding for admin and users"
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: admin@example.com
EOF

kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-users
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: user1@example.com
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: user2@example.com
EOF

echo "# Set OIDC credentials for kubectl"
kubectl config set-credentials oidc \
    --exec-api-version=client.authentication.k8s.io/v1beta1 \
    --exec-command=kubectl \
    --exec-arg=oidc-login \
    --exec-arg=get-token \
    --exec-arg=--oidc-issuer-url=https://${DEX_HOSTNAME} \
    --exec-arg=--oidc-client-id=kube \
    --exec-arg=--oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
    --exec-arg=--oidc-extra-scope=email \
    --exec-arg=--oidc-extra-scope=groups \
    --exec-arg=--certificate-authority-data=$(cat .ssl/root-ca.pem | base64 | tr -d '\n')

echo "# Install oauth2-proxy"
helm upgrade -i oauth2-proxy oauth2-proxy \
  --namespace oauth2-proxy --create-namespace \
  --repo https://oauth2-proxy.github.io/manifests/ \
  --values - <<EOF
proxyVarsAsSecrets: false
config:
  # Config file
  configFile: |-
    client_id="oauth2-proxy"
    client_secret="b2F1dGgyLXByb3h5LWNsaWVudC1zZWNyZXQK"
    cookie_secure="false"

    # Provider config
    provider="oidc"
    provider_display_name="Dex"
    redirect_url="https://auth.dev.local/oauth2/callback"
    oidc_issuer_url="https://dex.local"
    ssl_insecure_skip_verify=true

    # Upstream config
    http_address="0.0.0.0:4180"
    upstreams=["file:///dev/null"]
    email_domains=["example.com"]

    cookie_secret="OQINaROshtE9TcZkNAm-5Zs2Pv3xaWytBmc5W7sPX7w="
    cookie_domains=["dev.local"]
    
    whitelist_domains=[".dev.local"]
    
    set_xauthrequest= true
    set_authorization_header="true"

ingress:
  enabled: true
  className: "nginx"
  pathType: Prefix
  path: /oauth2
  annotations:
    cert-manager.io/cluster-issuer: ca-issuer
    nginx.ingress.kubernetes.io/proxy-buffer-size: "32k"
  hosts:
    - auth.dev.local
  tls:
    - hosts:
        - auth.dev.local
      secretName: oauth2-proxy-tls
EOF

echo "# Next steps

Add auth endpoint to /etc.hosts:

echo '<primary IP> dex.local' | sudo tee -a /etc/hosts

Connect to cluster with oidc:

kubectl --user oidc get pods
"