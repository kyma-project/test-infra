#!/usr/bin/env bash

VERSION=v16.20.0
DISTRO=linux-x64

curl -Lo node.tar.gz https://nodejs.org/download/release/$VERSION/node-$VERSION-$DISTRO.tar.xz
mkdir -p /usr/local/lib/nodejs
tar -xJvf node.tar.xz -C /usr/local/lib/nodejs 

export PATH=/usr/local/lib/nodejs/node-$VERSION-$DISTRO/bin:$PATH

echo "Node.js version: $(node -v)"
echo "NPM version: $(npm -v)"

echo "Copying Kubeconfig"
k3d kubeconfig get k3d > tests/integration/fixtures/kubeconfig.yaml

echo "Installing Busola"
npm install

echo "Starting Busola"
npm start 2>&1 &
sleep 80

echo "Run Cypress"
cd tests/integration
npm ci --no-optional
npm run test:cluster
