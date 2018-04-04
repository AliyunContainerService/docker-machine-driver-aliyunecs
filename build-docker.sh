#!/bin/bash
docker build -t registry.cn-hangzhou.aliyuncs.com/denverdino/docker-machine-aliyunecs .
docker create --name builder registry.cn-hangzhou.aliyuncs.com/denverdino/docker-machine-aliyunecs
docker cp builder:/go/src/github.com/AliyunContainerService/docker-machine-driver-aliyunecs/bin .
docker rm builder

