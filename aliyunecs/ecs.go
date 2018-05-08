package aliyunecs

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	mrand "math/rand"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"

	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	//"os"
)

const (
	driverName               = "aliyunecs"
	defaultRegion            = "cn-hangzhou"
	defaultInstanceType      = "ecs.n4.small"
	defaultRootSize          = 20
	internetChargeType       = string(common.PayByTraffic)
	ipRange                  = "0.0.0.0/0"
	machineSecurityGroupName = "rancher-machine"
	vpcCidrBlock             = "10.0.0.0/8"
	vSwitchCidrBlock         = "10.1.0.0/24"
	timeout                  = 300
	defaultSSHUser           = "root"
	maxRetry                 = 20
)

var (
	dockerPort = 2376
	swarmPort  = 3376
)

type Driver struct {
	*drivers.BaseDriver
	Id                      string
	AccessKey               string
	SecretKey               string
	Region                  common.Region
	ImageID                 string
	SSHPassword             string
	SSHKeyPairName          string
	SSHPrivateKeyPath       string
	PublicKey               []byte
	InstanceId              string
	InstanceType            string
	PrivateIPAddress        string
	SecurityGroupId         string
	SecurityGroupName       string
	ReservationId           string
	VpcId                   string
	VSwitchId               string
	Zone                    string
	PrivateIPOnly           bool
	InternetMaxBandwidthOut int
	InternetChargeType      common.InternetChargeType
	RouteCIDR               string
	SLBID                   string
	SLBIPAddress            string
	Tags                    map[string]string
	DiskSize                int
	DiskFS                  string
	DiskCategory            ecs.DiskCategory
	Description             string
	IoOptimized             bool
	APIEndpoint             string
	SystemDiskCategory      ecs.DiskCategory
	SystemDiskSize          int

	client    *ecs.Client
	slbClient *slb.Client
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "aliyunecs-access-key-id",
			Usage:  "ECS Access Key ID",
			Value:  "",
			EnvVar: "ECS_ACCESS_KEY_ID",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-access-key-secret",
			Usage:  "ECS Access Key Secret",
			Value:  "",
			EnvVar: "ECS_ACCESS_KEY_SECRET",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-image-id",
			Usage:  "ECS machine image",
			EnvVar: "ECS_IMAGE_ID",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-region",
			Usage:  "ECS region, default cn-hangzhou",
			Value:  defaultRegion,
			EnvVar: "ECS_REGION",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-vpc-id",
			Usage:  "ECS VPC id",
			Value:  "",
			EnvVar: "ECS_VPC_ID",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-vswitch-id",
			Usage:  "ECS VSwitch id",
			Value:  "",
			EnvVar: "ECS_VSWITCH_ID",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-zone",
			Usage:  "ECS zone for instance",
			Value:  "",
			EnvVar: "ECS_ZONE",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-security-group",
			Usage:  "ECS VPC security group",
			Value:  "docker-machine",
			EnvVar: "ECS_SECURITY_GROUP",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-instance-type",
			Usage:  "ECS instance type",
			Value:  defaultInstanceType,
			EnvVar: "ECS_INSTANCE_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-private-ip",
			Usage:  "ECS VPC instance private IP",
			Value:  "",
			EnvVar: "ECS_VPC_PRIVATE_IP",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-description",
			Usage:  "Description for instance",
			Value:  "",
			EnvVar: "ECS_DESCRIPTION",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-ssh-password",
			Usage:  "Set the password of the ssh user",
			EnvVar: "ECS_SSH_PASSWORD",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-ssh-keypair",
			Usage:  "Set the SSH key pair name",
			EnvVar: "ECS_SSH_KEYPAIR",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-ssh-keypath",
			Usage:  "File path of SSH private key",
			EnvVar: "ECS_SSH_KEYPATH",
		},
		mcnflag.BoolFlag{
			Name:   "aliyunecs-private-address-only",
			EnvVar: "ECS_PRIVATE_ADDR_ONLY",
			Usage:  "Only use a private IP address",
		},
		mcnflag.IntFlag{
			Name:   "aliyunecs-internet-max-bandwidth",
			Usage:  "Maximum bandwidth for Internet access (in Mbps), default 1",
			Value:  1,
			EnvVar: "ECS_INTERNET_MAX_BANDWIDTH",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-internet-charge-type",
			Usage:  "Internet charge type, the valid values are PayByTraffic (default) or PayByBandwidth",
			Value:  internetChargeType,
			EnvVar: "ECS_INTERNET_CHARGE_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-route-cidr",
			Usage:  "Docker bridge CIDR for route entry in VPC",
			EnvVar: "ECS_ROUTE_CIDR",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-slb-id",
			Usage:  "SLB id for instance association",
			EnvVar: "ECS_SLB_ID",
		},
		mcnflag.StringSliceFlag{
			Name:   "aliyunecs-tag",
			Usage:  "Tags for instance",
			Value:  []string{},
			EnvVar: "ECS_TAGS",
		},
		mcnflag.IntFlag{
			Name:   "aliyunecs-disk-size",
			Usage:  "Data disk size for instance in GB",
			Value:  0,
			EnvVar: "ECS_DISK_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-disk-fs",
			Usage:  "File system for data disk (ext4 or xfs)",
			Value:  "ext4",
			EnvVar: "ECS_DISK_FS",
		},
		mcnflag.IntFlag{
			Name:   "aliyunecs-system-disk-size",
			Usage:  "System disk size for instance in GB",
			Value:  40,
			EnvVar: "ECS_SYSTEM_DISK_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-system-disk-category",
			Usage:  "System disk category for instance",
			EnvVar: "ECS_SYSTEM_DISK_CATEGORY",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-disk-category",
			Usage:  "Data disk category for instance",
			EnvVar: "ECS_DISK_CATEGORY",
		},
		mcnflag.BoolFlag{
			Name:   "aliyunecs-upgrade-kernel",
			Usage:  "Upgrade kernel for Ubuntu 14.04 instance (deprecated)",
			EnvVar: "ECS_UPGRADE_KERNEL",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-io-optimized",
			Usage:  "I/O optimized instance",
			Value:  "true",
			EnvVar: "ECS_IO_OPTIMIZED",
		},
		mcnflag.StringFlag{
			Name:   "aliyunecs-api-endpoint",
			Usage:  "Custom API endpoint",
			Value:  "",
			EnvVar: "ECS_API_ENDPOINT",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	id := generateId()
	return &Driver{
		Id: id,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		}}
}

func (d *Driver) GetImageID(image string) string {

	if len(image) != 0 {
		return image
	}
	args := ecs.DescribeImagesArgs{
		RegionId:        d.Region,
		ImageOwnerAlias: ecs.ImageOwnerSystem,
	}

	// Scan registed images with prefix of default Ubuntu image
	for {
		images, pagination, err := d.getClient().DescribeImages(&args)
		if err != nil {
			log.Errorf("%s | Failed to describe images: %v", d.MachineName, err)
			break
		} else {
			for _, image := range images {
				if strings.HasPrefix(image.ImageId, defaultUbuntuImagePrefix) {
					return image.ImageId
				}
			}
			nextPage := pagination.NextPage()
			if nextPage == nil {
				break
			}
			args.Pagination = *nextPage
		}
	}

	//Use default image

	image = defaultUbuntuImageID

	return image
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIEndpoint = flags.String("aliyunecs-api-endpoint")
	var region common.Region
	var err error
	regionId := flags.String("aliyunecs-region")
	if d.APIEndpoint != "" {
		// Ignore the Region validation
		region = common.Region(regionId)
	} else {
		region, err = validateECSRegion(flags.String("aliyunecs-region"))
		if err != nil {
			return err
		}
	}
	d.AccessKey = flags.String("aliyunecs-access-key-id")
	d.SecretKey = flags.String("aliyunecs-access-key-secret")
	d.Region = region
	d.ImageID = flags.String("aliyunecs-image-id")
	d.InstanceType = flags.String("aliyunecs-instance-type")
	d.VpcId = flags.String("aliyunecs-vpc-id")
	d.VSwitchId = flags.String("aliyunecs-vswitch-id")
	d.SecurityGroupName = flags.String("aliyunecs-security-group")
	d.Zone = flags.String("aliyunecs-zone")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = defaultSSHUser
	d.SSHPassword = flags.String("aliyunecs-ssh-password")
	d.SSHKeyPairName = flags.String("aliyunecs-ssh-keypair")
	d.SSHPrivateKeyPath = flags.String("aliyunecs-ssh-keypath")
	d.SSHPort = 22
	d.PrivateIPOnly = flags.Bool("aliyunecs-private-address-only")
	d.InternetMaxBandwidthOut = flags.Int("aliyunecs-internet-max-bandwidth")
	d.InternetChargeType = common.InternetChargeType(flags.String("aliyunecs-internet-charge-type"))
	d.RouteCIDR = flags.String("aliyunecs-route-cidr")
	d.SLBID = flags.String("aliyunecs-slb-id")
	d.DiskSize = flags.Int("aliyunecs-disk-size")
	d.DiskFS = flags.String("aliyunecs-disk-fs")
	d.DiskCategory = ecs.DiskCategory(flags.String("aliyunecs-disk-category"))
	tags := flags.StringSlice("aliyunecs-tag")
	//upgradeKernel := false //flags.Bool("aliyunecs-upgrade-kernel")
	//if upgradeKernel {
	//	log.Warnf("%s | The --aliyunecs-upgrade-kernel is deprecated. Please use the Ubuntu 16.04 as the OS image for ECS instance", d.MachineName)
	//}

	ioOptimized := strings.ToLower(flags.String("aliyunecs-io-optimized"))

	d.IoOptimized = (ioOptimized == "true" || ioOptimized == "optimized")
	d.Description = flags.String("aliyunecs-description")
	d.SystemDiskCategory = ecs.DiskCategory(flags.String("aliyunecs-system-disk-category"))
	d.SystemDiskSize = flags.Int("aliyunecs-system-disk-size")

	if d.SystemDiskCategory == "" && d.IoOptimized {
		d.SystemDiskCategory = ecs.DiskCategoryCloudSSD
	}

	if d.DiskCategory == "" && d.IoOptimized {
		d.DiskCategory = ecs.DiskCategoryCloudSSD
	}

	tagMap := make(map[string]string)
	if len(tags) > 0 {
		for _, tag := range tags {
			s := strings.Split(tag, "=")
			if len(s) != 2 {
				log.Infof("%s | Invalid tag for --aliyunecs-tag", tag)
				return fmt.Errorf("%s | Invalid tag for --aliyunecs-tag", tag)
			}
			k := strings.TrimSpace(s[0])
			v := strings.TrimSpace(s[1])
			tagMap[k] = v
		}
	}

	if len(tagMap) > 0 {
		d.Tags = tagMap
	}

	if d.RouteCIDR != "" {
		_, _, err := net.ParseCIDR(d.RouteCIDR)
		if err != nil {
			return fmt.Errorf("%s | Invalid CIDR value for --aliyunecs-route-cidr", d.MachineName)
		}
	}

	if d.InternetMaxBandwidthOut < 0 || d.InternetMaxBandwidthOut > 100 {
		return fmt.Errorf("%s | aliyunecs driver --aliyunecs-internet-max-bandwidth: The value should be in 1 ~ 100", d.MachineName)
	}

	if d.InternetMaxBandwidthOut == 0 {
		d.InternetMaxBandwidthOut = 1
	}

	if d.InternetChargeType != common.PayByTraffic && d.InternetChargeType != common.PayByBandwidth {
		return fmt.Errorf("Unsupported internet charge type: %s", d.InternetChargeType)
	}

	if d.AccessKey == "" {
		return fmt.Errorf("%s | aliyunecs driver requires the --aliyunecs-access-key-id option", d.MachineName)
	}

	if d.SecretKey == "" {
		return fmt.Errorf("%s | aliyunecs driver requires the --aliyunecs-access-key-secret option", d.MachineName)
	}

	//VpcId and VSwitchId are optional or required together
	if (d.VpcId == "" && d.VSwitchId != "") || (d.VpcId != "" && d.VSwitchId == "") {
		return fmt.Errorf("%s | aliyunecs driver requires both the --aliyunecs-vpc-id and --aliyunecs-vswitch-id for Virtual Private Cloud", d.MachineName)
	}

	if d.isSwarmMaster() {
		u, err := url.Parse(d.SwarmHost)
		if err != nil {
			return fmt.Errorf("error parsing swarm host: %s", err)
		}

		parts := strings.Split(u.Host, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		swarmPort = port
	}

	if d.APIEndpoint != "" {
		if d.SLBID != "" {
			return fmt.Errorf("Unsupport 'aliyunecs-slb-id' flag when the custom API endpoint is specified")
		}
	}

	if d.DiskFS != "xfs" && d.DiskFS != "ext4" {
		return fmt.Errorf("Unsupport file system for data disk: %s", d.DiskFS)
	}

	if d.SSHPrivateKeyPath == "" && d.SSHKeyPairName != "" {
		return fmt.Errorf("using --aliyunecs-keypair-name also requires --aliyunecs-ssh-keypath")
	}

	return nil
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) checkPrereqs() error {

	if d.SLBID != "" {
		loadBalancer, err := d.getSLBClient().DescribeLoadBalancerAttribute(d.SLBID)
		if err != nil {
			return fmt.Errorf("%s | Invalid --aliyunecs-slb-id: %v", d.MachineName, err)
		}
		d.SLBIPAddress = loadBalancer.Address
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
}

func (d *Driver) Create() error {

	var (
		err error
	)
	VpcId := d.VpcId
	VSwitchId := d.VSwitchId

	if err := d.checkPrereqs(); err != nil {
		return err
	}

	log.Infof("%s | Creating key pair for instance ...", d.MachineName)

	if err := d.createKeyPair(); err != nil {
		return fmt.Errorf("%s | Failed to create key pair: %v", d.MachineName, err)
	}

	log.Infof("%s | Configuring security groups instance ...", d.MachineName)
	if err := d.configureSecurityGroup(VpcId, d.SecurityGroupName); err != nil {
		return err
	}

	// Create random password if no input
	if d.SSHPassword == "" && d.SSHKeyPairName == "" {
		d.SSHPassword = randomPassword()
		log.Infof("%s | Launching instance with generated password, please update password in console or log in with ssh key.", d.MachineName)
	}

	imageID := d.GetImageID(d.ImageID)
	log.Infof("%s | Creating instance with image %s ...", d.MachineName, imageID)

	ioOptimized := ecs.IoOptimizedNone
	if d.IoOptimized {
		ioOptimized = ecs.IoOptimizedOptimized
	}

	args := ecs.CreateInstanceArgs{
		RegionId:                d.Region,
		InstanceName:            d.GetMachineName(),
		Description:             d.Description,
		ImageId:                 imageID,
		InstanceType:            d.InstanceType,
		SecurityGroupId:         d.SecurityGroupId,
		InternetChargeType:      d.InternetChargeType,
		Password:                d.SSHPassword,
		KeyPairName:             d.SSHKeyPairName,
		VSwitchId:               VSwitchId,
		ZoneId:                  d.Zone,
		IoOptimized:             ioOptimized,
		InternetMaxBandwidthOut: d.InternetMaxBandwidthOut,
		ClientToken:             d.getClient().GenerateClientToken(),
	}

	if d.SystemDiskCategory != "" {
		args.SystemDisk.Category = d.SystemDiskCategory
	}

	if d.SystemDiskSize > 0 {
		args.SystemDisk.Size = d.SystemDiskSize
	}

	if d.DiskSize > 0 { // Allocate Data Disk

		disk := ecs.DataDiskType{
			DiskName:           d.MachineName + "_data",
			Description:        "Data volume for Docker",
			Size:               d.DiskSize,
			Category:           d.DiskCategory,
			Device:             "/dev/xvdb",
			DeleteWithInstance: true,
		}

		args.DataDisk = []ecs.DataDiskType{disk}

	}

	// Create instance
	instanceId, err := d.getClient().CreateInstance(&args)

	if err != nil {
		err = fmt.Errorf("%s | Failed to create instance: %s", d.MachineName, err)
		log.Error(err)
		return err
	}
	log.Infof("%s | Create instance %s successfully", d.MachineName, instanceId)

	d.InstanceId = instanceId

	// Wait for creation successfully
	err = d.getClient().WaitForInstance(instanceId, ecs.Stopped, timeout)

	if err != nil {
		err = fmt.Errorf("%s | Failed to wait instance to 'stopped': %s", d.MachineName, err)
		log.Error(err)
	}

	if err == nil {
		err = d.configNetwork(VpcId, instanceId)
	}

	if err == nil {
		// Start instance
		log.Infof("%s | Starting instance %s ...", d.MachineName, instanceId)
		err = d.getClient().StartInstance(instanceId)
		if err == nil {
			// Wait for running
			err = d.getClient().WaitForInstance(instanceId, ecs.Running, timeout)

			if err == nil {
				log.Infof("%s | Start instance %s successfully", d.MachineName, instanceId)
				instance, err := d.getInstance()

				if err == nil {
					d.Zone = instance.ZoneId
					d.PrivateIPAddress = d.GetPrivateIP(instance)

					d.IPAddress = d.getIP(instance)

					ssh.SetDefaultClient(ssh.Native)

					d.configECSInstance(imageID)

					log.Infof("%s | Created instance %s successfully with public IP address %s and private IP address %s",
						d.MachineName,
						d.InstanceId,
						d.IPAddress,
						d.PrivateIPAddress,
					)
				}
			} else {
				err = fmt.Errorf("%s | Failed to wait instance to running state: %s", d.MachineName, err)
			}
		} else {
			err = fmt.Errorf("%s | Failed to start instance %s: %v", d.MachineName, instanceId, err)
		}
	}

	// Add instance tags
	if len(d.Tags) > 0 {
		log.Infof("%s | Adding tags %v to instance %s ...", d.MachineName, d.Tags, instanceId)
		args := ecs.AddTagsArgs{
			RegionId:     d.Region,
			ResourceId:   instanceId,
			ResourceType: ecs.TagResourceInstance,
			Tag:          d.Tags,
		}
		err2 := d.getClient().AddTags(&args)
		if err2 != nil {
			log.Warnf("%s | Failed to add tags %v to instance %s: %v", d.MachineName, d.Tags, instanceId, err)
		}
	}

	return err
}

func (d *Driver) configNetwork(vpcId string, instanceId string) error {
	var err error
	if vpcId == "" {
		// Assign public IP if not private IP only

		if !d.PrivateIPOnly {
			// Allocate public IP address for classic network
			var ipAddress string
			ipAddress, err = d.getClient().AllocatePublicIpAddress(instanceId)
			if err != nil {
				err = fmt.Errorf("%s | Error allocate public IP address for instance %s: %v", d.MachineName, instanceId, err)
			} else {
				log.Infof("%s | Allocate publice IP address %s for instance %s successfully", d.MachineName, ipAddress, instanceId)
			}
		}
	} else {
		err := d.addRouteEntry(vpcId)
		if err != nil {
			return err
		}
		if !d.PrivateIPOnly {
			// Create EIP for virtual private cloud
			eipArgs := ecs.AllocateEipAddressArgs{
				RegionId:           d.Region,
				Bandwidth:          d.InternetMaxBandwidthOut,
				InternetChargeType: d.InternetChargeType,
				ClientToken:        d.getClient().GenerateClientToken(),
			}
			log.Infof("%s | Allocating Eip address for instance %s ...", d.MachineName, instanceId)

			_, allocationId, err := d.getClient().AllocateEipAddress(&eipArgs)
			if err != nil {
				return fmt.Errorf("%s | Failed to allocate EIP address: %v", d.MachineName, err)
			}
			err = d.getClient().WaitForEip(d.Region, allocationId, ecs.EipStatusAvailable, 60)
			if err != nil {
				log.Infof("%s | Releasing Eip address %s for ...", d.MachineName, allocationId)
				err2 := d.getClient().ReleaseEipAddress(allocationId)
				if err2 != nil {
					log.Warnf("%s | Failed to release EIP address: %v", d.MachineName, err2)
				}
				return fmt.Errorf("%s | Failed to wait EIP %s: %v", d.MachineName, allocationId, err)
			}
			log.Infof("%s | Associating Eip address %s for instance %s ...", d.MachineName, allocationId, instanceId)
			err = d.getClient().AssociateEipAddress(allocationId, instanceId)
			if err != nil {
				return fmt.Errorf("%s | Failed to associate EIP address: %v", d.MachineName, err)
			}
			err = d.getClient().WaitForEip(d.Region, allocationId, ecs.EipStatusInUse, 60)
			if err != nil {
				return fmt.Errorf("%s | Failed to wait EIP %s: %v", d.MachineName, allocationId, err)
			}
		}
	}

	if d.SLBID != "" { // Add the instance to SLB
		log.Infof("%s | Adding instance %s to SLB %s ...", d.MachineName, instanceId, d.SLBID)
		count := 0
		for {
			backendServers := []slb.BackendServerType{
				slb.BackendServerType{
					ServerId: instanceId,
					Weight:   100,
				},
			}
			_, err = d.getSLBClient().AddBackendServers(d.SLBID, backendServers)
			if err != nil {
				log.Errorf("%s | Failed to add instance to SLB: %v", d.MachineName, err)
				count++
				if count <= maxRetry {
					time.Sleep(time.Duration(5000+mrand.Int63n(2000)) * time.Millisecond)
					continue
				} else {
					return fmt.Errorf("%s | Failed to delete route entry after %d times", d.MachineName, maxRetry)
				}
			}
			break
		}
	}

	return err
}

func (d *Driver) removeRouteEntry(vpcId string, regionId common.Region, instanceId string) error {

	client := d.getClient()

	describeArgs := ecs.DescribeVpcsArgs{
		VpcId:    vpcId,
		RegionId: regionId,
	}

	vpcs, _, err := client.DescribeVpcs(&describeArgs)
	if err != nil {
		return fmt.Errorf("%s | Failed to describe VPC %s in region %s: %v", d.MachineName, d.VpcId, d.Region, err)
	}
	vrouterId := vpcs[0].VRouterId

	describeRouteTablesArgs := ecs.DescribeRouteTablesArgs{
		VRouterId: vrouterId,
	}

	routeTables, _, err := client.DescribeRouteTables(&describeRouteTablesArgs)
	if err != nil {
		return fmt.Errorf("%s | Failed to describe route tables: %v", d.MachineName, err)
	}

	routeEntries := routeTables[0].RouteEntrys.RouteEntry

	// Find route entry associated with instance
	for _, routeEntry := range routeEntries {
		count := 0

		if routeEntry.InstanceId == instanceId {
			for {
				deleteArgs := ecs.DeleteRouteEntryArgs{
					RouteTableId:         routeEntry.RouteTableId,
					DestinationCidrBlock: routeEntry.DestinationCidrBlock,
					NextHopId:            routeEntry.InstanceId,
				}
				log.Infof("%s | Deleting route entry for instance %s ...", d.MachineName, d.InstanceId)

				err := client.DeleteRouteEntry(&deleteArgs)
				if err != nil {
					log.Errorf("%s | Failed to delete route entry: %v", d.MachineName, err)
					count++
					if count <= maxRetry {
						time.Sleep(time.Duration(5000+mrand.Int63n(2000)) * time.Millisecond)
						continue
					} else {
						return fmt.Errorf("%s | Failed to delete route entry after %d times", d.MachineName, maxRetry)
					}
				}
				return nil
			}
		}
	}
	return nil
}

func (d *Driver) addRouteEntry(vpcId string) error {

	if d.RouteCIDR != "" {
		client := d.getClient()

		describeArgs := ecs.DescribeVpcsArgs{
			VpcId:    vpcId,
			RegionId: d.Region,
		}
		vpcs, _, err := client.DescribeVpcs(&describeArgs)
		if err != nil {
			return fmt.Errorf("%s | Failed to describe VPC %s in region %s: %v", d.MachineName, d.VpcId, d.Region, err)
		}
		vrouterId := vpcs[0].VRouterId
		describeVRoutersArgs := ecs.DescribeVRoutersArgs{
			VRouterId: vrouterId,
			RegionId:  d.Region,
		}
		vrouters, _, err := client.DescribeVRouters(&describeVRoutersArgs)
		if err != nil {
			return fmt.Errorf("%s | Failed to describe VRouters: %v", d.MachineName, err)
		}
		routeTableId := vrouters[0].RouteTableIds.RouteTableId[0]
		count := 0

		for {
			createArgs := ecs.CreateRouteEntryArgs{
				RouteTableId:         routeTableId,
				DestinationCidrBlock: d.RouteCIDR,
				NextHopType:          ecs.NextHopInstance,
				NextHopId:            d.InstanceId,
				ClientToken:          client.GenerateClientToken(),
			}
			err = client.CreateRouteEntry(&createArgs)
			if err == nil {
				break
			}

			ecsErr, _ := err.(*common.Error)
			//Retry for IncorretRouteEntryStatus or Internal Error
			if ecsErr != nil && (ecsErr.StatusCode == 500 || (ecsErr.StatusCode == 400 && ecsErr.Code == "IncorrectRouteEntryStatus")) {
				count++
				if count <= maxRetry {
					time.Sleep(time.Duration(5000+mrand.Int63n(2000)) * time.Millisecond)
					continue
				}

			}
			return fmt.Errorf("%s | Failed to create route entry: %v", d.MachineName, err)
		}
	}
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

func (d *Driver) GetIP() (string, error) {
	inst, err := d.getInstance()
	if err != nil {
		return "", err
	}

	return d.getIP(inst), nil
}

func (d *Driver) GetPrivateIP(inst *ecs.InstanceAttributesType) string {
	if inst.InnerIpAddress.IpAddress != nil && len(inst.InnerIpAddress.IpAddress) > 0 {
		return inst.InnerIpAddress.IpAddress[0]
	}

	if inst.VpcAttributes.PrivateIpAddress.IpAddress != nil && len(inst.VpcAttributes.PrivateIpAddress.IpAddress) > 0 {
		return inst.VpcAttributes.PrivateIpAddress.IpAddress[0]
	}
	return ""
}

func (d *Driver) getIP(inst *ecs.InstanceAttributesType) string {
	if d.PrivateIPOnly {
		return d.GetPrivateIP(inst)
	}
	if inst.PublicIpAddress.IpAddress != nil && len(inst.PublicIpAddress.IpAddress) > 0 {
		return inst.PublicIpAddress.IpAddress[0]
	}
	if len(inst.EipAddress.IpAddress) > 0 {
		return inst.EipAddress.IpAddress
	}
	return ""
}

func (d *Driver) GetState() (state.State, error) {
	inst, err := d.getInstance()
	if err != nil {
		return state.Error, err
	}
	switch ecs.InstanceStatus(inst.Status) {
	case ecs.Starting:
		return state.Starting, nil
	case ecs.Running:
		return state.Running, nil
	case ecs.Stopping:
		return state.Stopping, nil
	case ecs.Stopped:
		return state.Stopped, nil
	default:
		return state.Error, nil
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) Start() error {
	if err := d.getClient().StartInstance(d.InstanceId); err != nil {
		log.Errorf("%s | Failed to start instance %s: %v", d.MachineName, d.InstanceId, err)
		return err
	}

	// Wait for running
	err := d.getClient().WaitForInstance(d.InstanceId, ecs.Running, timeout)

	if err != nil {
		log.Errorf("%s | Failed to wait instance %s running: %v", d.MachineName, d.InstanceId, err)
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	if err := d.getClient().StopInstance(d.InstanceId, false); err != nil {
		log.Errorf("%s | Failed to stop instance %s: %v", d.MachineName, d.InstanceId, err)
		return err
	}

	// Wait for stopped
	err := d.getClient().WaitForInstance(d.InstanceId, ecs.Stopped, timeout)

	if err != nil {
		log.Errorf("%s | Failed to wait instance %s stopped: %v", d.MachineName, d.InstanceId, err)
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	log.Infof("%s | Remove instance %s ...", d.MachineName, d.InstanceId)

	if d.InstanceId == "" {
		return fmt.Errorf("%s | Unknown instance id", d.MachineName)
	}

	s, err := d.GetState()
	if err == nil && s == state.Running {
		if err := d.Stop(); err != nil {
			log.Errorf("%s | Unable to removed the instance %s: %s", d.MachineName, d.InstanceId, err)
		}
	}

	instance, err := d.getInstance()
	if err != nil {
		log.Errorf("%s | Unable to describe the instance %s: %s", d.MachineName, d.InstanceId, err)
	} else {
		// Check and release EIP if exists
		if len(instance.EipAddress.AllocationId) != 0 {

			allocationId := instance.EipAddress.AllocationId

			err = d.getClient().UnassociateEipAddress(allocationId, instance.InstanceId)
			if err != nil {
				log.Errorf("%s | Failed to unassociate EIP address from instance %s: %v", d.MachineName, d.InstanceId, err)
			}
			err = d.getClient().WaitForEip(instance.RegionId, allocationId, ecs.EipStatusAvailable, 0)
			if err != nil {
				log.Errorf("%s | Failed to wait EIP %s available: %v", d.MachineName, allocationId, err)
			}
			err = d.getClient().ReleaseEipAddress(allocationId)
			if err != nil {
				log.Errorf("%s | Failed to release EIP address: %v", d.MachineName, err)
			}
		}
		log.Debugf("%s | instance.VpcAttributes: %++v\n", d.MachineName, instance.VpcAttributes)

		vpcId := instance.VpcAttributes.VpcId
		if vpcId != "" {
			// Remove route entry firstly
			d.removeRouteEntry(vpcId, instance.RegionId, instance.InstanceId)
		}
	}

	log.Infof("%s | Deleting instance: %s", d.MachineName, d.InstanceId)
	if err := d.getClient().DeleteInstance(d.InstanceId); err != nil {
		return fmt.Errorf("%s | Unable to delete instance %s: %s", d.MachineName, d.InstanceId, err)
	}
	d.InstanceId = ""
	d.IPAddress = ""
	d.PrivateIPAddress = ""
	d.Zone = ""
	return nil
}

func (d *Driver) Restart() error {
	if err := d.getClient().RebootInstance(d.InstanceId, false); err != nil {
		return fmt.Errorf("%s | Unable to restart instance %s: %s", d.MachineName, d.InstanceId, err)
	}
	return nil
}

func (d *Driver) Kill() error {
	log.Debugf("%s | Killing instance ...", d.MachineName)

	if err := d.getClient().StopInstance(d.InstanceId, true); err != nil {
		return fmt.Errorf("%s | Unable to kill instance %s: %s", d.MachineName, d.InstanceId, err)
	}
	return nil
}

func (d *Driver) getSLBClient() *slb.Client {
	if d.slbClient == nil {
		client := slb.NewClient(d.AccessKey, d.SecretKey)
		client.SetDebug(false)
		d.slbClient = client
	}
	return d.slbClient
}

func (d *Driver) getClient() *ecs.Client {
	if d.client == nil {
		client := ecs.NewClient(d.AccessKey, d.SecretKey)
		if d.APIEndpoint != "" {
			client.SetEndpoint(d.APIEndpoint)
		}
		client.SetUserAgent("AliyunContainerService/docker-machine")
		client.SetDebug(false)
		d.client = client
	}
	return d.client
}

func (d *Driver) getInstance() (*ecs.InstanceAttributesType, error) {
	return d.getClient().DescribeInstanceAttribute(d.InstanceId)
}

func (d *Driver) createKeyPair() error {

	if d.SSHPrivateKeyPath == "" {
		log.Debugf("%s | SSH key path: %s", d.MachineName, d.GetSSHKeyPath())

		if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			return err
		}

		publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
		if err != nil {
			return err
		}

		d.PublicKey = publicKey
	} else {
		log.Debugf("%s | Using SSHPrivateKeyPath: %s", d.MachineName, d.SSHPrivateKeyPath)
		if err := mcnutils.CopyFile(d.SSHPrivateKeyPath, d.GetSSHKeyPath()); err != nil {
			return err
		}
		if err := mcnutils.CopyFile(d.SSHPrivateKeyPath+".pub", d.GetSSHKeyPath()+".pub"); err != nil {
			return err
		}
		if d.SSHKeyPairName != "" {
			log.Debugf("%s | Using existing ECS key pair: %s", d.MachineName, d.SSHKeyPairName)
			return nil
		}
	}

	return nil
}

func (d *Driver) isSwarmMaster() bool {
	return d.SwarmMaster
}

func (d *Driver) getSecurityGroup(id string) (sg *ecs.DescribeSecurityGroupAttributeResponse, err error) {
	args := ecs.DescribeSecurityGroupAttributeArgs{
		SecurityGroupId: id,
		RegionId:        d.Region,
	}
	return d.getClient().DescribeSecurityGroupAttribute(&args)
}

func (d *Driver) securityGroupAvailableFunc(id string) func() bool {
	return func() bool {
		_, err := d.getSecurityGroup(id)
		if err == nil {
			return true
		}
		log.Debug(err)
		return false
	}
}

func (d *Driver) configureSecurityGroup(vpcId string, groupName string) error {
	log.Debugf("%s | Configuring security group in %s", d.MachineName, d.VpcId)

	var securityGroup *ecs.DescribeSecurityGroupAttributeResponse

	args := ecs.DescribeSecurityGroupsArgs{
		RegionId: d.Region,
		VpcId:    vpcId,
	}

	for {
		groups, pagination, err := d.getClient().DescribeSecurityGroups(&args)
		if err != nil {
			return err
		}
		//log.Debugf("DescribeSecurityGroups: %++v\n", groups)

		for _, grp := range groups {
			if grp.SecurityGroupName == groupName && grp.VpcId == d.VpcId {
				log.Debugf("%s | Found existing security group (%s) in %s", d.MachineName, groupName, d.VpcId)
				securityGroup, _ = d.getSecurityGroup(grp.SecurityGroupId)
				break
			}
		}

		if securityGroup != nil {
			break
		}

		nextPage := pagination.NextPage()
		if nextPage == nil {
			break
		}
		args.Pagination = *nextPage
	}

	// if not found, create
	if securityGroup == nil {
		log.Debugf("%s | Creating security group (%s) in %s", d.MachineName, groupName, d.VpcId)
		creationArgs := ecs.CreateSecurityGroupArgs{
			RegionId:          d.Region,
			SecurityGroupName: groupName,
			Description:       "Rancher Machine",
			VpcId:             vpcId,
			ClientToken:       d.getClient().GenerateClientToken(),
		}

		groupId, err := d.getClient().CreateSecurityGroup(&creationArgs)
		if err != nil {
			return err
		}

		// wait until created (dat eventual consistency)
		log.Debugf("%s | Waiting for group (%s) to become available", d.MachineName, groupId)
		if err := mcnutils.WaitFor(d.securityGroupAvailableFunc(groupId)); err != nil {
			return err
		}
		securityGroup, err = d.getSecurityGroup(groupId)
		if err != nil {
			return err
		}
	}

	d.SecurityGroupId = securityGroup.SecurityGroupId

	perms := d.configureSecurityGroupPermissions(securityGroup)

	for _, permission := range perms {
		log.Debugf("%s | Authorizing group %s with permission: %v", d.MachineName, securityGroup.SecurityGroupName, permission)
		args := permission.createAuthorizeSecurityGroupArgs(d.Region, d.SecurityGroupId)
		args.NicType = ecs.NicTypeInternet
		if err := d.getClient().AuthorizeSecurityGroup(args); err != nil {
			log.Warnf("%s | Failed to authorizing group %s with permission: %v", d.MachineName, securityGroup.SecurityGroupName, permission, err)
			return err
		}
		args.NicType = ecs.NicTypeIntranet
		//如果是经典网络，则需要去掉2376的内网入规则
		if (d.VpcId == "" && d.VSwitchId == "") && permission.FromPort != dockerPort && permission.ToPort != dockerPort {
			if err := d.getClient().AuthorizeSecurityGroup(args); err != nil {
				log.Warnf("%s | Failed to authorizing group %s with permission: %v", d.MachineName, securityGroup.SecurityGroupName, permission, err)
				return err
			}
		}
	}

	return nil
}

type IpPermission struct {
	IpProtocol ecs.IpProtocol
	FromPort   int
	ToPort     int
	IpRange    string
}

func (p *IpPermission) createAuthorizeSecurityGroupArgs(regionId common.Region, securityGroupId string) *ecs.AuthorizeSecurityGroupArgs {
	args := ecs.AuthorizeSecurityGroupArgs{
		RegionId:        regionId,
		SecurityGroupId: securityGroupId,
		IpProtocol:      p.IpProtocol,
		SourceCidrIp:    p.IpRange,
		PortRange:       fmt.Sprintf("%d/%d", p.FromPort, p.ToPort),
	}
	return &args
}

func (d *Driver) configureSecurityGroupPermissions(group *ecs.DescribeSecurityGroupAttributeResponse) []IpPermission {
	hasSSHPort := false
	hasDockerPort := false
	for _, p := range group.Permissions.Permission {
		portRange := strings.Split(p.PortRange, "/")

		log.Debugf("%s | portRange %v", d.MachineName, portRange)
		fromPort, _ := strconv.Atoi(portRange[0])
		switch fromPort {
		case 22:
			hasSSHPort = true
		case dockerPort:
			hasDockerPort = true
		}
	}

	perms := []IpPermission{}

	if !hasSSHPort {
		perms = append(perms, IpPermission{
			IpProtocol: ecs.IpProtocolTCP,
			FromPort:   22,
			ToPort:     22,
			IpRange:    ipRange,
		})
	}

	if !hasDockerPort {
		perms = append(perms, IpPermission{
			IpProtocol: ecs.IpProtocolTCP,
			FromPort:   dockerPort,
			ToPort:     dockerPort,
			IpRange:    ipRange,
		})
	}

	//80
	perms = append(perms, IpPermission{
		IpProtocol: ecs.IpProtocolTCP,
		FromPort:   80,
		ToPort:     80,
		IpRange:    ipRange,
	})

	//443
	perms = append(perms, IpPermission{
		IpProtocol: ecs.IpProtocolTCP,
		FromPort:   443,
		ToPort:     443,
		IpRange:    ipRange,
	})

	//ICMP
	perms = append(perms, IpPermission{
		IpProtocol: ecs.IpProtocolICMP,
		FromPort:   -1,
		ToPort:     -1,
		IpRange:    ipRange,
	})

	//rke begin
	//apiserver
	perms = append(perms, IpPermission{
		IpProtocol: ecs.IpProtocolTCP,
		FromPort:   6443,
		ToPort:     6443,
		IpRange:    ipRange,
	})

	//etcd
	perms = append(perms, IpPermission{
		IpProtocol: ecs.IpProtocolTCP,
		FromPort:   2379,
		ToPort:     2380,
		IpRange:    ipRange,
	})

	//rke end
	//如果是容器网段的话，需要设置容器网段开放安全组
	if d.VpcId != "" || d.VSwitchId != "" {
		containerIPRange, err := getContainerCIDR(d.RouteCIDR)
		if err == nil {
			perms = append(perms, IpPermission{
				IpProtocol: ecs.IpProtocolAll,
				FromPort:   -1,
				ToPort:     -1,
				IpRange:    containerIPRange,
			})
		} else {
			log.Errorf("%s failed to get container bip %++v", group.SecurityGroupId, err)
		}
	}

	log.Debugf("%s | Configuring new permissions: %v", d.MachineName, perms)

	return perms
}

func getContainerCIDR(cidrBlock string) (string, error) {
	ip, _, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return "", err
	}

	ip = ip.To4()
	ip[2] = 0
	ip[3] = 0

	return fmt.Sprintf("%s/16", ip.String()), nil
}

func (d *Driver) deleteSecurityGroup() error {
	log.Infof("%s | Deleting security group %s", d.MachineName, d.SecurityGroupId)
	if err := d.getClient().DeleteSecurityGroup(d.Region, d.SecurityGroupId); err != nil {
		return err
	}

	return nil
}

func generateId() string {
	rb := make([]byte, 10)
	_, err := rand.Read(rb)
	if err != nil {
		log.Errorf("Unable to generate id: %s", err)
	}

	h := md5.New()
	io.WriteString(h, string(rb))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (d *Driver) configECSInstance(imageId string) error {
	ipAddr := d.IPAddress
	port, _ := d.GetSSHPort()
	tcpAddr := fmt.Sprintf("%s:%d", ipAddr, port)

	log.Infof("%s | Waiting SSH service %s is ready to connect ...", d.MachineName, tcpAddr)

	log.Infof("%s | Uploading SSH keypair to %s ...", d.MachineName, tcpAddr)

	var auth *ssh.Auth

	if d.SSHPrivateKeyPath == "" {
		auth = &ssh.Auth{
			Passwords: []string{d.SSHPassword},
		}
	} else {
		auth = &ssh.Auth{
			Keys: []string{d.SSHPrivateKeyPath},
		}
	}

	sshClient, err := ssh.NewClient(d.GetSSHUsername(), ipAddr, port, auth)

	if err != nil {
		return err
	}

	if d.SSHKeyPairName == "" {
		err = d.uploadKeyPair(sshClient)
		if err != nil {
			return err
		}
	}

	if isUbuntuImage(imageId) {
		d.fixAptConf(sshClient)
	}

	d.fixRoutingRules(sshClient)

	if d.DiskSize > 0 {
		d.autoFdisk(sshClient)
	}

	return nil
}

func (d *Driver) uploadKeyPair(sshClient ssh.Client) error {
	command := fmt.Sprintf("mkdir -p ~/.ssh; echo '%s' > ~/.ssh/authorized_keys", string(d.PublicKey))

	log.Debugf("%s | Upload the public key with command: %s", d.MachineName, command)

	output, err := sshClient.Output(command)

	log.Debugf("%s | Upload command err, output: %v: %s", d.MachineName, err, output)

	return err
}

func (d *Driver) fixAptConf(sshClient ssh.Client) {
	output, err := sshClient.Output("sed -i 's/Acquire::http::Proxy/#Acquire::http::Proxy/' /etc/apt/apt.conf")
	log.Debugf("%s | Update the apt.conf command err, output: %v: %s", d.MachineName, err, output)
}

// Fix the routing rules
func (d *Driver) fixRoutingRules(sshClient ssh.Client) {
	output, err := sshClient.Output("route del -net 172.16.0.0/12")
	log.Debugf("%s | Delete route command err, output: %v: %s", d.MachineName, err, output)

	output, err = sshClient.Output("if [ -e /etc/network/interfaces ]; then sed -i '/^up route add -net 172.16.0.0 netmask 255.240.0.0 gw/d' /etc/network/interfaces; fi")
	log.Debugf("%s | Fix route in /etc/network/interfaces command err, output: %v: %s", d.MachineName, err, output)

	output, err = sshClient.Output("if [ -e /etc/sysconfig/network-scripts/route-eth0 ]; then sed -i '/^172.16.0.0\\/12 via /d' /etc/sysconfig/network-scripts/route-eth0; fi")
	log.Debugf("%s | Fix route in /etc/sysconfig/network-scripts/route-eth0 command err, output: %v: %s", d.MachineName, err, output)
}

// Mount the addtional disk
func (d *Driver) autoFdisk(sshClient ssh.Client) {

	s := autoFdiskScriptExt4

	if d.DiskFS == "xfs" {
		s = autoFdiskScriptXFS
	}

	script := fmt.Sprintf("cat > ~/machine_autofdisk.sh <<MACHINE_EOF\n%s\nMACHINE_EOF\n", s)
	output, err := sshClient.Output(script)
	output, err = sshClient.Output("bash ~/machine_autofdisk.sh")
	log.Debugf("%s | Auto Fdisk command err, output: %v: %s", d.MachineName, err, output)
}
