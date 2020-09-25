#!/usr/bin/env bash

echo "--> Generate Certificate Signing Request"
cat << EOF | cfssl genkey - &> /dev/null | cfssljson -bare server
{
  "hosts": [
    "image-guard-svc.default.svc",
    "image-guard-svc.default.svc.cluster.local"
  ],
  "CN": "image-guard-svc.default.svc",
  "key": {
    "algo": "ecdsa",
    "size": 256
  }
}
EOF

echo "  Creating new Certificate Signing Request"
cat << EOF | kubectl apply -f - &> /dev/null
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: image-guard-svc.default
spec:
  request: $(base64 < server.csr | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

echo "  Approving Certificate Signing Request"
kubectl certificate approve image-guard-svc.default &> /dev/null
echo "  Gettinc cert and CA cert"
CERT=$(kubectl get csr image-guard-svc.default -ojsonpath='{.status.certificate}' | base64 --decode)
CA_CERT=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')
echo "--> Certificate created successfully"

echo "--> Creating secret with a certificate"
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: image-guard-certs
data:
  cert.crt: $(echo "$CERT" | base64)
  key.pem: $(base64 < server-key.pem)
type: Opaque
EOF
echo "--> Done"

echo "--> Deploy admission server to the cluster"
kubectl apply -f deployment.yaml
echo "--> Done"

echo "--> Configure server webhooks"
cat << EOF | kubectl apply -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: enforce-image-registry
webhooks:
  - name: enforce-image-registry.image-guard.admission
    admissionReviewVersions: ["v1", "v1beta1"]
    sideEffects: None
    rules:
      - apiVersions: ["*"]
        apiGroups: ["*"]
        operations: ["CREATE"]
        resources: ["pods"]
    clientConfig:
      service:
        name: image-guard-svc
        namespace: default
        path: /admission-control/enforce-image-registry
      caBundle: ${CA_CERT}
EOF
