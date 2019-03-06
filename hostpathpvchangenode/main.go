package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"sort"
	"strings"
	"time"

	"k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath"
	"k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath/xfsquotamanager/common"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	createSavePVOnly = flag.Bool("saveonly", true, "create save-mapping-info only")
	nodemappingfile  = flag.String("nodemappingfile", "map.txt", "node ip mapping file")
	kubeconfig       = flag.String("kubeconfig", getHomeDir()+"/.kube/config", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	namespace        = flag.String("namespace", "", "pvs bound by namespace pvc(namespace=all will for allnamespace)")
)

func getHomeDir() string {
	u, _ := user.Current()
	return u.HomeDir
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func parseMapFile(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return map[string]string{}, err
	}
	defer f.Close()
	read := bufio.NewReader(f)
	ret := make(map[string]string)
	for {
		line, e := read.ReadString('\n')
		if e != nil || io.EOF == e {
			break
		}
		strs := strings.Split(line, ":")
		if len(strs) == 2 {
			ret[strings.Trim(strs[0], " ")] = strings.Trim(strings.Trim(strs[1], "\n"), " ")
		}
	}
	return ret, nil
}

func mapToSliceAndSort(m map[string]string) []string {
	ret := make([]string, 0, len(m))
	for k, v := range m {
		ret = append(ret, fmt.Sprintf("%s -> %s", k, v))
	}
	sort.Strings(ret)
	return ret
}

func sureMapping(m map[string]string) bool {
	strs := mapToSliceAndSort(m)
	for _, s := range strs {
		fmt.Printf("    %s\n", s)
	}
	fmt.Printf("create save tmp pv: %t\n", *createSavePVOnly)
	fmt.Printf("Are you sure the node ip mapping(y/n): ")
	var ans byte
	fmt.Scanf("%c", &ans)
	if ans == 'y' || ans == 'Y' {
		return true
	}
	return false
}

func isPVHostpathPV(pv *v1.PersistentVolume) bool {
	if pv.Spec.HostPath != nil {
		return true
	}
	if pv.Spec.CSI != nil && strings.Contains(strings.ToLower(pv.Spec.CSI.Driver), "hostpath") == true {
		return true
	}
	return false
}

func getHostpathPVs(client kubernetes.Interface, ns string) ([]*v1.PersistentVolume, error) {
	pvs, err := client.Core().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		return []*v1.PersistentVolume{}, err
	}
	ret := make([]*v1.PersistentVolume, 0, len(pvs.Items))
	for i, _ := range pvs.Items {
		if isPVHostpathPV(&pvs.Items[i]) {
			if ns == "all" || (pvs.Items[i].Spec.ClaimRef != nil && pvs.Items[i].Spec.ClaimRef.Namespace == ns) {
				ret = append(ret, &pvs.Items[i])
			}
		}
	}
	return ret, nil
}

func getPVMountInfo(pv *v1.PersistentVolume) (hostpath.HostPathPVMountInfoList, error) {
	ret := make(hostpath.HostPathPVMountInfoList, 0, 10)
	if pv.Annotations == nil || pv.Annotations[common.PVVolumeHostPathMountNode] == "" {
		return ret, nil
	}
	err := json.Unmarshal([]byte(pv.Annotations[common.PVVolumeHostPathMountNode]), &ret)

	return ret, err
}

func getPVSMountInfo(pvs []*v1.PersistentVolume, filter func(nodename string) bool) (hostpath.HostPathPVMountInfoList, error) {
	ret := make(hostpath.HostPathPVMountInfoList, 0, 5*len(pvs))
	mapInfo := make(map[string]hostpath.HostPathPVMountInfo)
	for _, pv := range pvs {
		info, err := getPVMountInfo(pv)
		if err != nil {
			glog.Errorf("get pv %s mountinfo err:%v", pv.Name, err)
			continue
		}
		for _, i := range info {
			if filter(i.NodeName) {
				continue
			}
			if existNodeinfo, exist := mapInfo[i.NodeName]; exist {
				existNodeinfo.MountInfos = append(existNodeinfo.MountInfos, i.MountInfos...)
				mapInfo[i.NodeName] = existNodeinfo
			} else {
				mapInfo[i.NodeName] = i
			}
		}
	}
	for _, infos := range mapInfo {
		ret = append(ret, infos)
	}
	return ret, nil
}
func newTmpHostPathPV(name string) *v1.PersistentVolume {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"name": name, "app": "hostpathpvchangenode"},
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Capacity:    v1.ResourceList{v1.ResourceStorage: *resource.NewQuantity(1, resource.DecimalSI)},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "\\"},
			},
		},
	}
	return pv
}

func mountInfoMapToSlice(mapInfo map[string]hostpath.HostPathPVMountInfo) []hostpath.HostPathPVMountInfo {
	ret := make([]hostpath.HostPathPVMountInfo, 0, len(mapInfo))
	for _, info := range mapInfo {
		ret = append(ret, info)
	}
	return ret
}

func mergeMountInfo(existInfo, info hostpath.HostPathPVMountInfo) hostpath.HostPathPVMountInfo {
	if existInfo.NodeName != info.NodeName {
		return existInfo
	}
	pathSet := make(map[string]bool)
	for _, i := range existInfo.MountInfos {
		pathSet[i.HostPath] = true
	}
	for _, i := range info.MountInfos {
		if _, exist := pathSet[i.HostPath]; exist == false {
			existInfo.MountInfos = append(existInfo.MountInfos, i)
		}
	}
	return existInfo
}

func replacePVHostpath(client kubernetes.Interface, pvs []*v1.PersistentVolume, replaceMap map[string]string) error {
	for _, pv := range pvs {
		if isPVHostpathPV(pv) {
			curPV, err := client.Core().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get pv %s err:%v", pv.Name, err)
			}
			info, errGet := getPVMountInfo(curPV)
			if errGet != nil {
				return fmt.Errorf("get pv %s mount info err:%v", pv.Name, errGet)
			}
			changed := false
			infoMap := make(map[string]hostpath.HostPathPVMountInfo)
			for i, _ := range info {
				nodeName := info[i].NodeName

				if newName, exist := replaceMap[nodeName]; exist && nodeName != "" && newName != nodeName {
					info[i].NodeName = newName
					changed = true
				}

				existInfo, exist := infoMap[info[i].NodeName]
				if exist {
					infoMap[info[i].NodeName] = mergeMountInfo(existInfo, info[i])
				} else {
					infoMap[info[i].NodeName] = info[i]
				}
			}
			if changed {
				errUpdate := updateHostpathPVMountInfo(client, curPV.Name, mountInfoMapToSlice(infoMap), 5)
				if errUpdate != nil {
					return errUpdate
				}
			}
			fmt.Printf("PV %s change OK\n", curPV.Name)
		}
	}
	return nil
}

func updateHostpathPVMountInfo(client kubernetes.Interface, pvName string, mountinfos hostpath.HostPathPVMountInfoList, tryTimes int) error {
	buf, err := json.Marshal(mountinfos)
	if err != nil {
		return err
	}

	for i := 0; i < tryTimes; i++ {
		pv, err := client.Core().PersistentVolumes().Get(pvName, metav1.GetOptions{})
		if pv.Annotations == nil {
			pv.Annotations = make(map[string]string)
		}
		pv.Annotations[common.PVVolumeHostPathMountNode] = string(buf)

		_, err = client.Core().PersistentVolumes().Update(pv)
		if err == nil {
			return nil
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return fmt.Errorf("update pv %s err:%v, after try %d times", pvName, err, tryTimes)
}

func createTmpPV(client kubernetes.Interface, pvName string, mountinfos hostpath.HostPathPVMountInfoList) error {
	pv := newTmpHostPathPV(pvName)
	buf, err := json.Marshal(mountinfos)
	if err != nil {
		return err
	}
	pv.Annotations = map[string]string{
		common.PVVolumeHostPathMountNode: string(buf),
		common.PVHostPathMountPolicyAnn:  common.PVHostPathKeep,
		common.PVHostPathQuotaForOnePod:  "true",
	}

	client.Core().PersistentVolumes().Delete(pv.Name, nil)
	_, err = client.Core().PersistentVolumes().Create(pv)

	return err
}

func main() {
	flag.Parse()

	if *nodemappingfile == "" {
		glog.Error("nodemappingfile not define")
		os.Exit(1)
	}
	if *kubeconfig == "" {
		glog.Error("kubeconfig not define")
		os.Exit(1)
	}
	if *namespace == "" {
		glog.Error("namespace not define")
		os.Exit(1)
	}

	nodeMap, errMap := parseMapFile(*nodemappingfile)
	if errMap != nil {
		glog.Error("read mapping file err:%v", errMap)
		os.Exit(1)
	}

	if sureMapping(nodeMap) == false {
		os.Exit(0)
	}

	config, err := buildConfig(*kubeconfig)
	if err != nil {
		glog.Error(err.Error())
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Error(err.Error())
		os.Exit(2)
	}

	pvs, errGet := getHostpathPVs(clientset, *namespace)
	if errGet != nil {
		glog.Error(errGet.Error())
		os.Exit(3)
	}

	infos, err := getPVSMountInfo(pvs, func(nodeName string) bool {
		_, exist := nodeMap[nodeName]
		return !exist
	})
	tmppvname := "save-mapping-info"
	if *createSavePVOnly {
		tmppvname = "saveonly-mapping-info"
	}
	err = createTmpPV(clientset, tmppvname, infos)
	if err != nil {
		glog.Error("createTmpPV err:%v", err)
		os.Exit(4)
	}
	if *createSavePVOnly == false {
		if err = replacePVHostpath(clientset, pvs, nodeMap); err != nil {
			glog.Error("replacePVHostpath err:%v", err)
			os.Exit(5)
		}
	}
}
