#!/bin/bash
docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/docker-machine-driver-aliyunecs .
docker create --name builder registry.cn-hangzhou.aliyuncs.com/denverdino/docker-machine-driver-aliyunecs
docker cp builder:/go/src/github.com/denverdino/docker-machine-driver-aliyunecs/bin .
docker rm builder

