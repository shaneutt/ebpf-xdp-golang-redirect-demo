#!/bin/bash

set -euox pipefail

docker run --name udp-listener -d ubuntu /bin/bash -c 'apt-get update -q && apt-get install netcat -yq && nc -kul 172.17.0.2 9875'
