package aliyunecs

const autoFdiskScript = `#/bin/bash
#fdisk ,formating and create the file system on /dev/xvdb or /dev/vdb
DISK_ATTACH_POINT="/dev/xvdb"
fdisk_fun()
{
fdisk -S 56 \$DISK_ATTACH_POINT << EOF
n
p
1


wq
EOF

sleep 5
mkfs.ext4 -i 8192 \${DISK_ATTACH_POINT}1
}

#config /etc/fstab and mount device
main()
{
  if [ -b "/dev/vdb" ]; then
    DISK_ATTACH_POINT="/dev/vdb"
  fi

  fdisk_fun
  flag=0
  if [ -d "/var/lib/docker" ];then
    flag=1
    service docker stop
    rm -fr /var/lib/docker
  fi
  mkdir /var/lib/docker
  echo "\${DISK_ATTACH_POINT}1    /var/lib/docker     ext4    defaults        0 0" >>/etc/fstab
  mount -a

  if [ \$flag==1 ]; then
    service docker start
  fi
}

main
df -h

`
