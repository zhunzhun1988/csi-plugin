package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
	//	"fmt"
	"github.com/k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath/xfsquotamanager"
)

type k8s struct {
}

func (k8s *k8s) IsPodExist(podId string) bool {
	return false
}
func (k8s *k8s) GetParentId(podId string) string {
	return ""
}

func main() {
	manager := xfsquotamanager.NewXFSQuotaManager("/xfs",
		&k8s{})
	starttime := time.Now()
	defer func() {
		fmt.Printf("use time:%v\n", time.Since(starttime))
	}()
	if len(os.Args) > 1 {
		if os.Args[1] == "add" {
			if len(os.Args) < 6 {
				panic("add not enough args")
			} else {
				volumeId := os.Args[2]
				podId := os.Args[3]

				quota, err1 := strconv.Atoi(os.Args[4])
				if err1 != nil {
					panic(fmt.Sprintf("parse quota %s err:%v", os.Args[4], err1))
				}
				recycle := false
				if os.Args[5] == "true" {
					recycle = true
				}
				if podId == "empty" {
					podId = ""
				}
				if _, path, err := manager.AddQuotaPath(volumeId, podId, int64(quota), int64(quota), recycle); err != nil {
					fmt.Printf("add quota err:%v\n", err)
				} else {
					fmt.Printf("add quota success %s\n", path)
				}
			}
		} else if os.Args[1] == "delete" {
			if len(os.Args) < 3 {
				panic("delete not enough args")
			} else {
				if _, err := manager.DeleteQuotaByPath(os.Args[2]); err != nil {
					fmt.Printf("delete quota err:%v\n", err)
				} else {
					fmt.Printf("delete quota %s success\n", os.Args[2])
				}
			}
		} else if os.Args[1] == "watch" {
			if len(os.Args) < 3 {
				panic("watch not enough args")
			} else {
				watchpath := os.Args[2]
				fmt.Printf("start watch path:%s\n", watchpath)

				for {
					if exist, info := manager.GetQuotaInfoByPath(watchpath); exist == false {
						fmt.Printf("%s is not a valid quotapath\n", watchpath)
					} else {
						fmt.Printf("%s quotainfo %d/%d\n", watchpath, info.UsedSize, info.HardQuota)
					}
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			fmt.Printf("unknow command\n")
		}
	} else {

		diskinfos := manager.GetQuotaDiskInfos()
		buf, _ := json.MarshalIndent(diskinfos, " ", "  ")
		fmt.Printf("%s\n", string(buf))
	}

}
