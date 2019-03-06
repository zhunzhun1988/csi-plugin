package main

import (
	"strconv"
	//"encoding/json"
	"fmt"
	"k8s-plugins/csi-plugin/hostpathpv/pkg/hostpath/xfsquotamanager/prjquota"
	"os"
	"strings"
)

var pq prjquota.Interface

func addQuota(projectid int64, path string, quota int64) error {
	prjName := fmt.Sprintf("k8spro%d", projectid)
	if err := pq.QuotaAddProjectInfo(projectid, prjName, path); err != nil {
		return err
	}
	if ret, err := pq.QuotaSetupProject(prjName); err != nil {
		return fmt.Errorf("QuotaSetupProject err:%v, ret:%s", err, ret)
	}
	if ret, _, _, err := pq.LimitQuotaProject(prjName, quota, quota); err != nil {
		return fmt.Errorf("LimitQuotaProject err:%v, ret:%s", err, ret)
	}
	return nil
}

func deleteQuota(projectid int64) error {
	projects := pq.QuotaProjects()
	var path string
	for _, pro := range projects {
		if int64(pro.ProjectId) == projectid {
			path = string(pro.Path)
		}
	}
	if path == "" {
		return fmt.Errorf("project %d is not exist", projectid)
	}
	if err := pq.QuotaDeleteProjectInfo(projectid); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func main() {
	host := prjquota.NewHostRunner()
	pq = prjquota.New(host)
	if len(os.Args) > 1 {
		if os.Args[1] == "add" {
			if len(os.Args) < 5 {
				panic("add not enough args")
			} else {
				fmt.Printf("run add %s\n", strings.Join(os.Args[1:], " "))
				projectid, err1 := strconv.Atoi(os.Args[2])
				if err1 != nil {
					panic(fmt.Sprintf("project id %s err %v", os.Args[2], err1))
				}
				path := os.Args[3]
				quota, err2 := strconv.Atoi(os.Args[4])
				if err2 != nil {
					panic(fmt.Sprintf("quota %s err %v", os.Args[4], err2))
				}
				if host.IsPathExist(path) == false {
					if err := os.MkdirAll(path, 0755); err != nil {
						panic(fmt.Errorf("mkdir %s err :%v", path, err))
					}
				}
				if err3 := addQuota(int64(projectid), path, int64(quota)); err3 != nil {
					fmt.Printf("addQuota fail %v\n", err3)
				} else {
					fmt.Printf("addQuota success\n")
				}
			}
		} else if os.Args[1] == "delete" {
			if len(os.Args) < 3 {
				panic("delete not enough args")
			} else {
				fmt.Printf("run delete %s\n", strings.Join(os.Args[1:], " "))
				projectid, err1 := strconv.Atoi(os.Args[2])
				if err1 != nil {
					panic(fmt.Sprintf("project id %s err %v", os.Args[2], err1))
				}
				if err2 := deleteQuota(int64(projectid)); err2 != nil {
					fmt.Printf("deleteQuota fail %v\n", err2)
				} else {
					fmt.Printf("deleteQuota success\n")
				}
			}

		} else {
			fmt.Printf("unknow command\n")
		}
	}
}
