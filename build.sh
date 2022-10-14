#!/bin/bash


go build -o container-upgrader

docker image build -t oscarzhou/container-upgrader:0.0.1 .
