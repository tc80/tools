package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	user = "tcaslin@cloudflare.com"
)

var (
	namespaceID = util.GetEnv("WORKERS_KV_NAMESPACE_ID")
	accountID   = util.GetEnv("WORKERS_KV_ACCOUNT_ID")
	apiKey      = util.GetEnv("WORKERS_KV_API_KEY")
	api         = getAPI()
)

func getAPI() *cloudflare.API {
	a, err := cloudflare.New(apiKey, user, cloudflare.UsingAccount(accountID))
	util.Check(err)
	return a
}

func getKVs() cloudflare.ListStorageKeysResponse {
	resp, err := api.ListWorkersKVs(context.Background(), namespaceID)
	util.Check(err)
	return resp
}

func delete(key string) {
	fmt.Printf("deleting %s\n", key)
	resp, err := api.DeleteWorkersKV(context.Background(), namespaceID, key)
	util.Check(err)
	if !resp.Success {
		log.Fatalf("delete failure %v\n", resp)
	}
}

func deleteTestEntries(startsWith string) {
	kvs := getKVs()
	for _, res := range kvs.Result {
		if key := res.Name; strings.HasPrefix(key, startsWith) {
			delete(key)
		}
	}
}

func worker(basePath string, paths <-chan string, kvPairs chan<- *cloudflare.WorkersKVPair) {
	fmt.Println("worker start!", basePath)
	for p := range paths {
		bytes, err := ioutil.ReadFile(path.Join(basePath, p))
		if err != nil {
			panic(err)
		}
		// resp, err := api.WriteWorkersKV(context.Background(), namespaceID, p, bytes)
		// util.Check(err)
		// fmt.Println(resp.Success, p)
		// kvPairs <- nil
		kvPairs <- &cloudflare.WorkersKVPair{
			Key:   p,
			Value: string(bytes),
		}
	}
}

func main() {
	// deleteTestEntries("3D")
	// os.Exit(1)

	basePath := util.GetCDNJSPackages()
	files, err := filepath.Glob(path.Join(basePath, "*", "package.json"))
	util.Check(err)

	// for i, f := range files {
	// 	fmt.Printf("%d - %s\n", i, f)
	// }
	paths := make(chan string)
	kvPairs := make(chan *cloudflare.WorkersKVPair)

	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(basePath, paths, kvPairs)
	}

	p, err := packages.ReadPackageJSON(context.Background(), files[4])
	util.Check(err)

	var kvs []*cloudflare.WorkersKVPair

	for _, v := range p.Versions() {
		versionPath := path.Join(p.Name, v)
		strs, err := util.ListFilesInVersion(context.Background(), path.Join(p.Path(), v))
		util.Check(err)
		for _, s := range strs {
			paths <- path.Join(versionPath, s)
		}
		for i := 0; i < len(strs); i++ {
			// <-kvPairs
			k := <-kvPairs
			fmt.Println("received! ", k.Key, len(k.Value))
			kvs = append(kvs, k)
		}
		// break
	}
	fmt.Println("waiting ...")
	resp, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)
	util.Check(err)
	fmt.Println(resp)

	os.Exit(1)

}
