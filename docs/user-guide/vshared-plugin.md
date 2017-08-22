---
title: vShared volume plugin for Docker

---
## Overview
Depending on the underlying block storage device system, it might not be possible to access the same 
persistent volume across different hosts/nodes simultanously.
For example, currently users cannot mount the same persistent volume which is created through
vSphere Docker Volume Service (vDVS) on containers running on two different hosts at the same time.

This can be solved through distributed file systems, such as NFS, Ceph, Gluster, etc.
However, setting up and maintaining those distributed file systems for docker persistent data usage is not a trivial work. 
Furthermore, users can face more challenges in order to achieve high availability, scalability, and load balancing. 

__vShared volume plugin for Docker__ provides simultanous persistent volume access between hosts in the
same Docker Swarm cluster for the base volume plugin service such as vDVS, with zero configuration effort,
along with high availability, scalability, and load balancing support.

## Detailed documentation
Detailed documentation can be found on our [GitHub Documentation Page](http://vmware.github.io/docker-volume-vsphere/documentation/).

## Prerequisites
* Docker version 1.30/17.06.0 is required.
* To use vShared plugin, hosts must be Docker Swarm mode.
[How to create a swarm](https://docs.docker.com/engine/swarm/swarm-tutorial/create-swarm/)
[How to add nodes to the swarm](https://docs.docker.com/engine/swarm/swarm-tutorial/add-nodes/)
* Base docker volume plugin (e.g. [vSphere Docker Volume Service](https://github.com/vmware/docker-volume-vsphere))

## Usage examples

#### Creating a persistent volume from vShared plugin
```
$ docker volume create --driver=vshared --name=SharedVol -o size=10gb
$ docker volume ls
$ docker volume inspect SharedVol
```
Options for creation will be the same for the base volume plugin.
Please refer to the base volume plugin for proper options.

#### Mounting this volume to a container running on the first host
```
# ssh to node1
$ docker run --rm -it -v SharedVol:/mnt/myvol --name busybox-on-node1 busybox
/ # cd /mnt/myvol
# write data into mounted shared volume
```

#### Mounting this volume to a container running on the second host
```
# ssh to node2
$ docker run --rm -it -v SharedVol:/mnt/myvol --name busybox-on-node2 busybox
/ # cd /mnt/myvol
# read data from mounted shared volume
```

#### Stopping the two containers on each host
```
# docker stop busybox-on-node1
# docker stop busybox-on-node2
```

#### Removing the vShared volume
```
$ docker volume rm SharedVol
```

## Installing
The recommended way to install vShared plugin is from docker cli:
```
docker plugin install --grant-all-permissions --alias vshared vmware/vsphere-shared:latest
```

## Configuration
### Options for vShared plugin
Users can choose the base volume plugin for vShared plugin, by setting configuration during install process.
<!---
* Through CLI flag can only be done through non-managed plugin.
--->

* Default config file location: `/etc/vsphere-shared.conf`.
* Default base volume plugin: vSphere Docker Volume Service
* Sample config file:
```
{
        "InternalDriver": "vsphere"
}
```

The user can override the default configuration by providing a different configuration file, 
via the `--config` option, specifying the full path of the file.

### Options for logging
* Default log location: `/var/log/vsphere-shared.log`.
* Logs retention, size for rotation and log location can be set in the config file too:
```
 {
	"MaxLogAgeDays": 28,
	"MaxLogSizeMb": 100,
	"LogPath": "/var/log/vsphere-shared.log"
}
```

## Q&A

### How to install and use the driver?
Please see README.md in the for the release by clicking on the tag for the release. Example: TBD.

### Can I use another base volume plugin other than vDVS?
Currently vShared volume service is only developed and tested with vDVS as the base volume plugin.
