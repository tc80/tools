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

func deleteTestEntries() {
	kvs := getKVs()
	for _, res := range kvs.Result {
		if key := res.Name; strings.HasPrefix(key, "test_") {
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
		kvPairs <- &cloudflare.WorkersKVPair{
			Key:   p,
			Value: string(bytes),
		}
	}
}

func main() {
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
			k := <-kvPairs
			fmt.Println("received! ", k.Key, len(k.Value))
			kvs = append(kvs, k)
			// os.Exit(1)
		}
		break
	}

	resp, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)
	util.Check(err)
	fmt.Println(resp)

	//	fmt.Println(strings.Replace(files[2], libsPath, "", 1))
	os.Exit(1)

	// best way to get a file to a string ?

	// deleteTestEntries()
	// os.Exit(1)

	//

	// api, err := cloudflare.New(apiKey, user, cloudflare.UsingAccount(accountID))
	// if err != nil {
	// 	log.Fatal("fail1", err)
	// }

	// //rand.Read(payload)

	// for i := 0; i < 100; i++ {
	// 	//payload := make([]byte, 10485761)
	// 	//key := "small_file"
	// 	resp, err := api.WriteWorkersKV(context.Background(), namespace, fmt.Sprintf("test_%d", i), []byte(fmt.Sprintf("value_%d", i)))
	// 	if err != nil {
	// 		log.Fatal("fail2", err)
	// 	}
	// 	fmt.Println(resp.Success)
	// }

	// bulk request fast
	// use worker pool to generate bulk request
	// push bulk request
	//ioutil.Read

	// list files glob
	// get list of files -- push that number of jobs
	// receive path, return cloudflare.WorkersKVPair, or error ??? or nil ???

	// try normally then with bulk and compare time

	// var kvs []*cloudflare.WorkersKVPair
	// for i := 0; i < 100; i++ {
	// 	//payload := make([]byte, 10485761)
	// 	//key := "small_file"
	// 	kvs = append(kvs, &cloudflare.WorkersKVPair{
	// 		Key:   fmt.Sprintf("test_%d", i),
	// 		Value: fmt.Sprintf("value_%d", i),
	// 	})
	// 	// resp, err := api.WriteWorkersKV(context.Background(), namespace, fmt.Sprintf("test_%d", i), []byte(fmt.Sprintf("value_%d", i)))
	// 	// if err != nil {
	// 	// 	log.Fatal("fail2", err)
	// 	// }
	// 	// fmt.Println(resp.Success)
	// }

	// resp, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)
	// util.Check(err)
	// fmt.Println(resp)

	// 	fmt.Println(resp)
	// }

	// fmt.Printf("`%s`\n", resp)

	// fmt.Println(resp)
}
