<!--[metadata]>
+++
title = "Aliyun Elastic Compute Service"
description = "Aliyun driver for machine"
keywords = ["machine, aliyun, driver, ecs"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->
# Docker Machine Driver of Aliyun ECS
Create machines on [Aliyun Elastic Compute Service (ECS)](http://www.aliyun.com/).  You will need an Access Key ID, Secret Access Key and a Region ID.  If you want to setup instance on the VPC network, you will need the VPC ID and VSwitch ID; Please login to the Aliyun console -> Products and Services -> VPC and select the one where you would like to launch the instance.

creates docker instances on Aliyun ECS.

```bash
docker-machine create -d aliyunecs machine-aliyunecs
```

## Installation

The easiest way to install the clc docker-machine driver is to:

```
go install github.com/denverdino/docker-machine-driver-aliyunecs

```

binaries also available,you can download from [drone.io](https://drone.io/github.com/denverdino/machine/files)


## Example Usage 

## Options

```bash
docker-machine create -d aliyunecs --help
```
 Option Name                                          | Description                                           | required 
------------------------------------------------------|------------------------------------------------------|----
``--aliyunecs-access-key-id`` | Your access key ID for the Aliyun ECS API.  |**yes**|
``--aliyunecs-access-key-secret``|Your secret access key for the Aliyun ECS API.| **yes** |
``--aliyunecs-api-endpoint``|The custom API endpoint.|**no** |
``--aliyunecs-description`` | The description of instance.|**no** |
 ``--aliyunecs-disk-size``| The data disk size for /var/lib/docker (in GB),[more detail]()||
 ``--aliyunecs-disk-category``|The category of data disk, the valid values could be `cloud` (default), `cloud_efficiency` or `cloud_ssd`.|| 
``--aliyunecs-image-id``| The image ID of the instance to use Default is the latest Ubuntu 14.04 provided by system||
``--aliyunecs-io-optimized``| The I/O optimized instance type, the valid values could be `none` (default) or `optimized`||
``--aliyunecs-instance-type``| The instance type to run.  Default: `ecs.t1.small`||
``--aliyunecs-internet-max-bandwidth``| Maxium bandwidth for Internet access (in Mbps), default 1||
``--aliyunecs-private-address-only``| Use the private IP address only||
``--aliyunecs-region``| The region to use when launching the instance. Default: `cn-hangzhou`||
``--aliyunecs-route-cidr``| The CIDR to use configure the route entry for the instance in VPC. Sample: 192.168.200.0/24||
``--aliyunecs-security-group``| Aliyun security group name. Default: `docker-machine`||
``--aliyunecs-slb-id``|SLB id for instance association||
``--aliyunecs-ssh-password``| SSH password for created virtual machine. Default is random generated.||
``--aliyunecs-system-disk-category``|System disk category for instance||
``--aliyunecs-tag``| Tag for the instance.||
``--aliyunecs-vpc-id``| Your VPC ID to launch the instance in. (required for VPC network only)||
``--aliyunecs-vswitch-id``| Your VSwitch ID to launch the instance with. (required for VPC network only)||
``--aliyunecs-zone``| The availabilty zone to launch the instance||

## Environment variables and default values:

| CLI option                          | Environment variable        | Default          |
|-------------------------------------|-----------------------------|------------------|
| **`--aliyunecs-access-key-id`**     | `ECS_ACCESS_KEY_ID`         | -                |
| **`--aliyunecs-access-key-key`**    | `ECS_ACCESS_KEY_SECRET`     | -                |
| `--aliyunecs-api-endpoint`          | `ECS_API_ENDPOINT`          | -                |
| `--aliyunecs-description`           | `ECS_DESCRIPTION`           | -                |
| `--aliyunecs-disk-size`             | `ECS_DISK_SIZE`             | -                |
| `--aliyunecs-disk-category`         | `ECS_DISK_CATEGORY`         | -                |
| `--aliyunecs-image-id`              | `ECS_IMAGE_ID`              | -                |
| `--aliyunecs-aliyunecs-io-optimized`| `ECS_IO_OPTIMIZED`          | `none`           |
| `--aliyunecs-instance-type`         | `ECS_INSTANCE_TYPE`         | `ecs.t1.small`   |
| `--aliyunecs-internet-max-bandwidth`| `ECS_INTERNET_MAX_BANDWIDTH`| `1`              |
| `--aliyunecs-private-address-only`  | `ECS_PRIVATE_ADDR_ONLY`     | `false`          |
| `--aliyunecs-region`                | `ECS_REGION`                | `cn-hangzhou`    |
| `--aliyunecs-route-cidr`            | `ECS_ROUTE_CIDR`            | -                |
| `--aliyunecs-security-group`        | `ECS_SECURITY_GROUP`        | -                |
| `--aliyunecs-slb-id`                | `ECS_SLB_ID`                | -                |
| `--aliyunecs-ssh-password`          | `ECS_SSH_PASSWORD`          | Random generated |
| `--aliyunecs-tag`                   | `ECS_TAGS`                  | -                |
| `--aliyunecs-vpc-id`                | `ECS_VPC_ID`                | -                |
| `--aliyunecs-vswitch-id`            | `ECS_VSWITCH_ID`            | -                |
| `--aliyunecs-zone`                  | `ECS_ZONE`                  | -                |

Each environment variable may be overloaded by its option equivalent at runtime.

## Kernels

The default ubuntu image runs kernel 3.18.xx but advanced swarm/networking features require a newer kernel.

Optionally, `docker-machine ssh` in and `apt-get install -qy linux-image-generic-lts-wily && reboot`

## Hacking

### Get the sources

```bash
go get github.com/denverdino/docker-machine-driver-aliyunecs
cd $GOPATH/src/github.com/denverdino/docker-machine-driver-aliyunecs
```

### Test the driver

To test the driver make sure your current build directory has the highest
priority in your ``$PATH`` so that docker-machine can find it.

```
export PATH=$GOPATH/src/github.com/denverdino/docker-machine-driver-aliyunecs:$PATH
```

## Related links

- **Docker Machine**: https://docs.docker.com/machine/
- **Contribute**: https://github.com/denverdino/docker-machine-driver-aliyunecs
- **Report bugs**: https://github.com/denverdino/docker-machine-driver-aliyunecs/issues

## License

Apache 2.0