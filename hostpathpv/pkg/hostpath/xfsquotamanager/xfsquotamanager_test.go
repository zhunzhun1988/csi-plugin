package xfsquotamanager

import (
	"path"
	"testing"

	"github.com/k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath/xfsquotamanager/common"
	"github.com/k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath/xfsquotamanager/prjquota"
)

type fakeK8sInterface struct {
	podidToOwnerid map[string]string
	activePodids   []string
}

func (fk *fakeK8sInterface) IsPodExist(podId string) bool {
	for _, pod := range fk.activePodids {
		if pod == podId {
			return true
		}
	}
	return false
}

func (fk *fakeK8sInterface) GetParentId(podId string) string {
	for pid, oid := range fk.podidToOwnerid {
		if pid == podId {
			return oid
		}
	}
	return ""
}

func TestXFSQuotaManagerAddTest(t *testing.T) {
	device1 := prjquota.XFSStateInfo{
		MountPath:                 prjquota.XFSMountPath("/xfs/disk1"),
		Device:                    prjquota.XFSDevice("/dev/sda1"),
		UserQuotaAccountingOn:     false,
		UserQuotaEnforcementOn:    false,
		GroupQuotaAccountingOn:    false,
		GroupQuotaEnforcementOn:   false,
		ProjectQuotaAccountingOn:  true,
		ProjectQuotaEnforcementOn: true,
	}
	device2 := prjquota.XFSStateInfo{
		MountPath:                 prjquota.XFSMountPath("/xfs/disk2"),
		Device:                    prjquota.XFSDevice("/dev/sdb1"),
		UserQuotaAccountingOn:     false,
		UserQuotaEnforcementOn:    false,
		GroupQuotaAccountingOn:    false,
		GroupQuotaEnforcementOn:   false,
		ProjectQuotaAccountingOn:  true,
		ProjectQuotaEnforcementOn: true,
	}
	device3 := prjquota.XFSStateInfo{
		MountPath:                 prjquota.XFSMountPath("/xfs/disk3"),
		Device:                    prjquota.XFSDevice("/dev/sdc1"),
		UserQuotaAccountingOn:     false,
		UserQuotaEnforcementOn:    false,
		GroupQuotaAccountingOn:    false,
		GroupQuotaEnforcementOn:   false,
		ProjectQuotaAccountingOn:  true,
		ProjectQuotaEnforcementOn: true,
	}
	diskSizeMap := map[string]int64{string(device1.Device): 1000, string(device2.Device): 1000, string(device3.Device): 1000}
	quotaProject1 := prjquota.FakePathQuotaInfo{
		Path:           "/xfs/disk1/k8squota_volumeid1_kfp-podid1",
		VolumeId:       "volumeid1",
		PodId:          "podid1",
		OwnerId:        "podid1",
		ProjectId:      1,
		ProjectName:    "k8spro1",
		UsedSize:       10,
		SoftQuota:      200,
		HardQuota:      200,
		IsKeep:         true,
		IsShare:        false,
		IsCSIQuotaPath: true,
	}
	quotaProject2 := prjquota.FakePathQuotaInfo{
		Path:           "/xfs/disk2/k8squota_volumeid2_kfp-podid1",
		VolumeId:       "volumeid2",
		PodId:          "podid1",
		OwnerId:        "podid1",
		ProjectId:      2,
		ProjectName:    "k8spro2",
		UsedSize:       10,
		SoftQuota:      100,
		HardQuota:      100,
		IsKeep:         true,
		IsShare:        false,
		IsCSIQuotaPath: true,
	}
	quotaProject3 := prjquota.FakePathQuotaInfo{
		Path:           "/xfs/disk3/k8squota_volumeid3_kfp-podid1",
		VolumeId:       "volumeid3",
		PodId:          "podid1",
		OwnerId:        "podid1",
		ProjectId:      3,
		ProjectName:    "k8spro3",
		UsedSize:       10,
		SoftQuota:      200,
		HardQuota:      200,
		IsKeep:         true,
		IsShare:        false,
		IsCSIQuotaPath: true,
	}

	addProjectsTest := []struct {
		Quota           int64
		CanAdd          bool
		TestName        string
		VolumeId        string
		PodId           string
		PodIdToWonerId  map[string]string
		ActivePods      []string
		ShouldMountPath string
		Devices         []prjquota.XFSStateInfo
		DisableDevices  []prjquota.XFSStateInfo
		ExistsProject   []prjquota.FakePathQuotaInfo
		Recycle         bool
	}{
		{
			Quota:           100 * 1024,
			CanAdd:          true,
			TestName:        "Normal add at empty disk3",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk3/k8squota_volumeid11_kfp-podid11/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject2},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           100 * 1024,
			CanAdd:          true,
			TestName:        "Normal add at empty disk2",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk2/k8squota_volumeid11_kfp-podid11/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           100 * 1024,
			CanAdd:          true,
			TestName:        "Normal add at empty disk1",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk1/k8squota_volumeid11_kfp-podid11/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject2, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           100 * 1024,
			CanAdd:          true,
			TestName:        "Normal add at quota less disk2",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk2/k8squota_volumeid11_kfp-podid11/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject2, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           900 * 1024,
			CanAdd:          true,
			TestName:        "Normal add only disk2 match",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk2/k8squota_volumeid11_kfp-podid11/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject2, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           900 * 1024,
			CanAdd:          false,
			TestName:        "Normal add no disk match",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "",
			Devices:         []prjquota.XFSStateInfo{device1, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           900 * 1024,
			CanAdd:          false,
			TestName:        "Normal add use exist quota path but owner pod is active",
			VolumeId:        "volumeid1",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "",
			Devices:         []prjquota.XFSStateInfo{device1, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{"podid1"},
			Recycle:         true,
		},
		{
			Quota:           900 * 1024,
			CanAdd:          true,
			TestName:        "Normal add use exist quota path and owner pod is not active",
			VolumeId:        "volumeid1",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk1/k8squota_volumeid1_kfp-podid1/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device3},
			DisableDevices:  []prjquota.XFSStateInfo{},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{""},
			Recycle:         true,
		},
		{
			Quota:           100 * 1024,
			CanAdd:          false,
			TestName:        "Normal add no enable disk can be used",
			VolumeId:        "volumeid11",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "",
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{device1, device2, device3},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{""},
			Recycle:         true,
		},
		{
			Quota:           100 * 1024,
			CanAdd:          true,
			TestName:        "Normal add no enable disk can be used but can recycle the old quota path",
			VolumeId:        "volumeid1",
			PodId:           "podid11",
			PodIdToWonerId:  map[string]string{"podid11": "ownerid11"},
			ShouldMountPath: "/xfs/disk1/k8squota_volumeid1_kfp-podid1/" + common.XfsKeepForOnePodInnerDir,
			Devices:         []prjquota.XFSStateInfo{device1, device2, device3},
			DisableDevices:  []prjquota.XFSStateInfo{device1, device2, device3},
			ExistsProject:   []prjquota.FakePathQuotaInfo{quotaProject1, quotaProject3},
			ActivePods:      []string{""},
			Recycle:         true,
		},
	}

	for _, test := range addProjectsTest {
		host := prjquota.NewFakeHostRunner([]string{"/xfs/"})
		fakeK8sInterface := &fakeK8sInterface{podidToOwnerid: test.PodIdToWonerId, activePodids: test.ActivePods}
		qm := NewFakeXFSQuotaManager("/xfs/", host, test.Devices, diskSizeMap, test.ExistsProject, fakeK8sInterface)

		for _, disableDisk := range test.DisableDevices {
			if err := qm.SetQuotaDiskDisabled(string(disableDisk.MountPath), true); err != nil {
				t.Errorf("[%s] set disk %s disable err:%v", test.TestName, string(disableDisk.MountPath), err)
			}
		}
		ok, volumePath, err := qm.AddQuotaPath(test.VolumeId, test.PodId, test.Quota, test.Quota, test.Recycle)
		if ok != test.CanAdd {
			t.Errorf("[%s] expect can added %t, but add %t, volumePath:%s , err:%v", test.TestName, test.CanAdd, ok, volumePath, err)
		} else if ok == true && path.Clean(volumePath) != path.Clean(test.ShouldMountPath) {
			t.Errorf("[%s] expect add quota path at %s but added at %s", test.TestName, test.ShouldMountPath, volumePath)
		}
		fakeK8sInterface.activePodids = []string{}
		if ok == true {
			if exist, info := qm.GetQuotaInfoByVolumePodId(test.VolumeId, test.PodId); exist == false {
				t.Errorf("[%s] add quotapath success but not find quota info", test.TestName)
			} else {
				qm.DeleteQuotaByPath(info.Path)
			}
		}

		// test file clean
		for _, existProject := range test.ExistsProject {
			qm.DeleteQuotaByPath(existProject.Path)
		}
		if dirs := host.GetDirs("/xfs/disk"); len(dirs) > 0 {
			t.Errorf("[%s] dir %v is not be cleaned", test.TestName, dirs)
		}
		if files := host.GetFiles("/xfs", common.XfsDiskDisablefilename); len(files) > 0 {
			t.Errorf("[%s] files %v is not be cleaned", test.TestName, files)
		}
	}
}
