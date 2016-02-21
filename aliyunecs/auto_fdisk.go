package aliyunecs

const autoFdiskScript = `#/bin/bash
#fdisk ,formating and create the file system on /dev/xvdb
fdisk_fun()
{
fdisk -S 56 /dev/xvdb << EOF
n
p
1


wq
EOF

sleep 5
mkfs.ext4 /dev/xvdb1
}

#config /etc/fstab and mount device
main()
{
  fdisk_fun
  mkdir /var/lib/docker
  echo "/dev/xvdb1    /var/lib/docker     ext4    defaults        0 0" >>/etc/fstab
  mount -a
}

main
df -h


`
