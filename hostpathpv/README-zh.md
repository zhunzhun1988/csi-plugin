# CSI HostpathPV
　
## 说明

[English](README.md) | [中文](README-zh.md)

**该模块主要对有hostpath的Node上的Hostpath quota进行分配/回收/删除等，它也会将hostpath的使用情况更新到对应PV的annotation的io.enndata.kubelet/alpha-pvchostpathnode字段，如果quota目录没有PV对其引用则会将其quota目录删除．该模块要想正常工作需要[调度模块](https://gitlab.cloud.enndata.cn/kubernetes/k8s-plugins/blob/master/extender-scheduler/README-zh.md)的配合．**

## 部署
+ **1) 下载代码及编译：**

		$ git clone ssh://git@gitlab.cloud.enndata.cn:10885/kubernetes/k8s-plugins.git
		$ cd k8s-plugins/csi-plugin/driver-registrar
		$ make release REGISTRY=10.19.140.200:29006  # 编译CSI插件所需要的registrar
		$ cd ../external-attacher/
		$ make release REGISTRY=10.19.140.200:29006  # 编译CSI插件所需要的attacher模块
		$ cd ../hostpathpv
		$ make release REGISTRY=10.19.140.200:29006 # 编译CSI hostpathpv模块
    
    （make release 将编译代码且制作相应docker image ihub.helium.io:29006/library/hostpathcsi:$TAG , 并将其push到registry．也可以只执行make build 生成hostpathcsi可执行文件，详情可以查看该目录下的Makefile）
	
+ **2) Hostpath Node部署：**
因为hostpath quota管理是通过xfs文件系统的prjquota来实现的，所以需要一个独立的分区．如：node有如下分区/dev/vdb1, /dev/vdc1, /dev/vdd1 是给hostpath pv使用的, 准备工作如下：

		# mkfs.xfs /dev/vdb1                                                                              # 格式化vdb1为xfs文件系统
		# mkfs.xfs /dev/vdc1                                                                              # 格式化vdc1为xfs文件系统
		# mkfs.xfs /dev/vdd1                                                                              # 格式化vdd1为xfs文件系统
		# mkdir -p /xfs/disk1                                                                              # 创建/dev/vdb1 mount目录
		# mount -o prjquota /dev/vdb1 /xfs/disk1                                              # mount /dev/vdb1
		# echo "/dev/vdb1 /xfs/disk1    xfs prjquota 0 0" >> /etc/fstab              # 确保重启之后/dev/vdb1 mount
		# mkdir -p /xfs/disk2                                                                             # 创建/dev/vdc1 mount目录
		# mount -o prjquota /dev/vdc1 /xfs/disk2                                             # mount /dev/vdc1
		# echo "/dev/vdc1 /xfs/disk2    xfs prjquota 0 0" >> /etc/fstab             # 确保重启之后/dev/vdc1 mount
		# mkdir -p /xfs/disk3                                                                            # 创建/dev/vdd1 mount目录
		# mount -o prjquota /dev/vdd1 /xfs/disk3                                            # mount /dev/vdd1
		# echo "/dev/vdd1 /xfs/disk3    xfs prjquota 0 0" >> /etc/fstab            # 确保重启之后/dev/vdd1 mount
		# xfs_quota -x -c "state"                                                                     # 确认quota盘正确的mount
		User quota state on /xfs/disk2 (/dev/vdc1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #0 (0 blocks, 0 extents)
		Group quota state on /xfs/disk2 (/dev/vdc1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #67 (1315 blocks, 11 extents)
		Project quota state on /xfs/disk2 (/dev/vdc1)
		  Accounting: ON
		  Enforcement: ON
		  Inode: #67 (1315 blocks, 11 extents)
		Blocks grace time: [7 days]
		Inodes grace time: [7 days]
		Realtime Blocks grace time: [7 days]
		User quota state on /xfs/disk3 (/dev/vdd1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #0 (0 blocks, 0 extents)
		Group quota state on /xfs/disk3 (/dev/vdd1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #67 (1315 blocks, 15 extents)
		Project quota state on /xfs/disk3 (/dev/vdd1)
		  Accounting: ON
		  Enforcement: ON
		  Inode: #67 (1315 blocks, 15 extents)
		Blocks grace time: [7 days]
		Inodes grace time: [7 days]
		Realtime Blocks grace time: [7 days]
		User quota state on /xfs/disk1 (/dev/vdb1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #0 (0 blocks, 0 extents)
		Group quota state on /xfs/disk1 (/dev/vdb1)
		  Accounting: OFF
		  Enforcement: OFF
		  Inode: #67 (1315 blocks, 15 extents)
		Project quota state on /xfs/disk1 (/dev/vdb1)
		  Accounting: ON
		  Enforcement: ON
		  Inode: #67 (1315 blocks, 15 extents)
		Blocks grace time: [7 days]
		Inodes grace time: [7 days]
		Realtime Blocks grace time: [7 days]
		# kubectl label node 192.168.122.9 io.enndata/hasquotapath=true --overwrite   # 给node 192.168.122.9 打标签，标记该Node支持hostpath quota，下面部署CSI plugin时用的DaemonSet只会运行在有该标签的Node上．
(其他支持hostpath quota 的Node也类似上面的步骤)

+ **3) 安装：**

		$ make install REGISTRY=10.19.140.200:29006
		$ kubectl -n k8splugin get pod -o wide
		NAME                                               READY     STATUS    RESTARTS   AGE       IP                NODE                 NOMINATED NODE
        csi-xfshostpath-attacher-0                1/1       Running                    0          43s       10.9.29.5    192.168.122.196   <none>
        csi-xfshostpath-attacher-1                1/1       Running                    0          43s       10.9.6.4      192.168.122.9       <none>
        csi-xfshostpath-attacher-2                1/1       Running                    0          9s         10.9.29.6    192.168.122.196   <none>
        csi-xfshostpath-plugin-5vn2d            2/2       Running                    0          44s       10.9.29.3    192.168.122.196   <none>
        csi-xfshostpath-plugin-hqsxj             2/2       Running                    0          44s       10.9.6.2      192.168.122.9       <none>


+ **4) 卸载：**

		$ make uninstall

## 测试
	为了测试模块的稳定性，专门写了一个测试脚本csi-plugin/hostpathpv/test/csikeeptest.sh，根据集群环境修改csikeeptest.sh里的kubeconfig．
    
		$ cd csi-plugin/hostpathpv/test/
		$ ./csikeeptest.sh