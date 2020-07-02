package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	//cloudflare "github.com/tc80/cloudflare-go"

	"github.com/cdnjs/tools/util"
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
	// r, err := api.WriteWorkersKV(
	// 	context.Background(),
	// 	namespaceID,
	// 	"test-key",
	// 	[]byte("hello world"),
	// )
	// for i := 0; i < 100; i++ {
	// 	resp, err := api.WriteWorkersKV(context.Background(), namespaceID, fmt.Sprintf("hello-world %d", i), []byte(fmt.Sprintf("hello world %d", i)))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(resp)
	// }
	// os.Exit(1)

	payload := make([]byte, 300)
	_, err := rand.Read(payload)
	util.Check(err)

	// kvs := []*cloudflare.WorkersKVPair{
	// 	&cloudflare.WorkersKVPair{
	// 		Key:    "test1",
	// 		Value:  base64.StdEncoding.EncodeToString([]byte("hello world")),
	// 		Base64: true,
	// 		Metadata: struct {
	// 			Prefix []string `json:"prefix"`
	// 			SRI    string   `json:"sri,omitempty"` // only calculate for files
	// 		}{
	// 			Prefix: []string{
	// 				"longpckname",
	// 			},
	// 			SRI: "something",
	// 		},
	// 		Expiration: 1693662399,
	// 		// ExpirationTTL: 100,
	// 	},
	// }

	r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, []*cloudflare.WorkersKVPair{
		&cloudflare.WorkersKVPair{
			Key:      "testing kv",
			Value:    "testing kv value",
			Metadata: `{"lib/firebugx.js":"sha512-13eEcmSqkWS/f8/3km30e2kRsRhj5KeRUzeItzT84Q2MoMGYjbgM809gxpZjTz/8gq7U6pXBhToeQwixaIn4xg==","lib/firebugx.min.js":"sha512-FgL/Banpp0kMTq84SnLD2XS4i4LGXl1ivshmjcpHSPrgcqUTB/O+Y9gOBiC9/DZ4ZjAxhhkkceCG8T45y2oyKw==","lib/jquery.event.drag-2.0.min.js":"sha512-yoEpUmDArIgVDQsxxbnWRbq1167BuCyKEB77z9eIcXl7SCj31j42xUE/XxPptJ/SCm5O+8YfyaRAKkaSEIrkTA==","slick.columnpicker.css":"sha512-c1ncG04DuSJr+ZZjIMcIlY866jdwuB/hXNzgdaEJYMUo1D+gfsbzK0puz+ee6Wzn7ejDSC4IvGrwmLgaJqVaHg==","slick.columnpicker.js":"sha512-RQcsjQbLDHf0yRMIo8vm1kSbgfdw8ZWRXUFQ4Hz5sT2wg4wjENeEN5iW7pjUuizfQOHQKfcOAT3GG43TU9RTGQ==","slick.columnpicker.min.css":"sha512-4NjADEaBz1/WytTMfSxVzD9zAQJOHbqsxkznuuAxrBLUN+U03R4nIhmOfX+cVbCUaxD1NitC80XV5AKm1vQSuQ==","slick.columnpicker.min.js":"sha512-RozwTYDeGYuKZupxBNh/kQ4qz8taEin0TgHj+4P7GUGidHwIjWhgamZdr+uIrrcmztmmup/ovPEVvtLKmWytNA==","slick.editors.js":"sha512-0mMsTnGaueMF/kTAVrdi0PSnkhLkVcLiyMI+FxwUHjktx7AcwvlfQqMO3mDBNSMOsDagd/xJZdwy4ZuCZ5oYiQ==","slick.editors.min.js":"sha512-Nw1vSBkvwhhi2ykjEGBPvZtz8wscr/vwYavqX/kiOVmWTklhYfXDbqQvacfxQqMcmiOsHGb9fUU97m4cDQGbLQ==","slick.grid.css":"sha512-f7xL1tX1vwAJBKyQzUJ/th0X4XMKYfmbDtOFiOtl7oYIDpEJBL2zEecMGKHXWzfaXA6Y6StEX2ezO6nERY8LVg==","slick.grid.js":"sha512-DhzlPOf+8sr3W72/KUA6glW7N3FCiu93yTvBEA7Fzae6L+slm2qgvJ9cQa2DDLePniKYY3c+bO9qrCDXJ9SPxQ==","slick.grid.min.css":"sha512-j67NkJGpw77pjXm2rhKoavp7zwB01WNUms19hPK688BRORTiPhDaKGAXpD9agG8mYeLdJA90/ym8p+iYvRXQmA==","slick.grid.min.js":"sha512-q/+eMv3pp3LjJ2ACkLG6kb8Vzq0dYbzwhKh1B6lKpyklX44u9xxLna1QfzhfWmtvimot/FyViK7xytHBuB7APQ==","slick.model.js":"sha512-/kr+tbxtF5ni6R/KDxRrAwrLZ+bk7sRxhl/GNc+9sODOd7FJ6xRgFZ6y2Qgpe5ZreignRUO49yNCmBI7ez6Pdw==","slick.model.min.js":"sha512-qFBFq1dFDUP16nikRMJUYGBJQtw670KsbLuO7T68ISJdjegnh2Wu5a4XkFEAOijDZonHqjKn6/0v8ebIkK99kQ==","slick.pager.css":"sha512-AdCIcy7EiFWKqq8DZ1ttZKudkNPaEp7b+YLlv9iAvhPRUJuoyVMEL1XdMLFALahXqP7EMCJ+fEYE/OnC/EQdYQ==","slick.pager.js":"sha512-Om4bpXh9w1GJHXdmpn1Jvq/4B2H0eXDxlZK8ExtslNRHGTdbc1ZCrG7HOP7lUTLIvL+0Khqk3Jx1mox04eNX4Q==","slick.pager.min.css":"sha512-eu34HbO1MHZPCEM756tOXC86vgsbafLDlbuRAFe9Zm9BNnbz2L3xgscTVt7Rhu7z6fUJQaq4la8hqMRFaaRk+Q==","slick.pager.min.js":"sha512-GBpGNecCi625mYE5uedbMKYsEdNtlCrBhMy8rjYORAvWzdV3ko9cFESyoiQ2K5N22nt42ilswYHGbQfPvE31WQ==","slick.remotemodel.js":"sha512-S8mOmjGC7SdBssxnYZYR47pWMI6jxBxOICIKXJK2HBpN2mrMC8JIFJBxbsgjxSRNjGPrKFoA3EAj+x37qVsmvQ==","slick.remotemodel.min.js":"sha512-tp0n0SEMmzADqBhB4WeK8079Lgaqq+lgp0NnlSn5EZg7GHhUbJ5BqqUd8BwEeidHlNSZ+2oHntWUVh5mRLYxQQ=="}`,
		},
	})

	fmt.Println(r)
	util.Check(err)

	limit := 10
	pref := "testing kv"

	re, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, cloudflare.ListWorkersKVsOptions{
		Limit:  &limit,
		Prefix: &pref,
	})

	fmt.Println(re)
	util.Check(err)

	// type Test struct {
	// 	Hello string `json:"hello"`
	// }
	// r, err := api.WriteWorkersKVWithOptions(context.Background(), namespaceID, "write with options test2", cloudflare.WriteWorkersKVOptions{
	// 	Value:    "test options2",
	// 	Metadata: Test{"hey"},
	// })
	// // r, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)

	// //bytes, err := api.ReadWorkersKV(context.Background(), namespaceID, "test2")
	// //
	// // // re, err := api.ListWorkersKVs(context.Background(), namespaceID, cloudflare.Option{Key: "limit", Value: "10"})
	// // count := 11
	// // pref := "hello"
	// c := "AAAAANSOuYgr4HmfGH02-cfDN8Cr9ejOwkd_Ai5rsZ7SANEqVJBenU9-gYRlrsziywKLx48RNDLvtYzBAmRPsLGdye8ECr5PqFYWIO8UITdhdyTc1x6bV8pmyjz5DO-XaZH4kYY1KfqT8NRBIe5sic6yYt3FUDttGZafy0ivi-UpmTkVdRB0OxCf3O3OB-svG6DXheV5XTNDNrNx1o_CVqy2l2j0F4iKV1qFe_KhdkjC7Y6QohUZ1MOb3J_uznNYVCo7Z-bVAAsJmXA"
	// re, err := api.ListWorkersKVsWithOptions(context.Background(), namespaceID, cloudflare.ListWorkersKVsOptions{
	// 	// Limit:  &count,
	// 	// Prefix: &pref,
	// 	// Cursor: &c,
	// })

	// fmt.Printf("%v\n", re)

	util.Check(err)
	//fmt.Println(r)
	// api.WriteWorkersKVBulk()
	// cloudflare.WorkersKVBulkWriteRequest{}
	//api.ListWorkersKVs()
	//api.UpdateWorkersKVNamespace()
	// cloudflare.WorkersKVNamespaceRequest{

	// }

	// update matadata
	// add function to api?

	// api.WriteWorkersKV()

	//api.UpdateWorkersKVNamespace
	// for i := 0; i < 100; i++ {
	// 	resp, err := api.WriteWorkersKV(context.Background(), namespaceID, "hello-world", []byte(fmt.Sprintf("hello world %d", i)))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(resp)
	// }
	// deleteTestEntries("3D")

	os.Exit(1)

	// basePath := util.GetCDNJSPackages()
	// files, err := filepath.Glob(path.Join(basePath, "*", "package.json"))
	// util.Check(err)

	// // for i, f := range files {
	// // 	fmt.Printf("%d - %s\n", i, f)
	// // }
	// paths := make(chan string)
	// kvPairs := make(chan *cloudflare.WorkersKVPair)

	// for i := 0; i < runtime.NumCPU(); i++ {
	// 	go worker(basePath, paths, kvPairs)
	// }

	// p, err := packages.ReadPackageJSON(context.Background(), files[4])
	// util.Check(err)

	// var kvs []*cloudflare.WorkersKVPair

	// // make api call to get kv entry for this particular package

	// for _, v := range p.Versions() {
	// 	versionPath := path.Join(p.Name, v)
	// 	strs, err := util.ListFilesInVersion(context.Background(), path.Join(p.Path(), v))
	// 	util.Check(err)
	// 	for _, s := range strs {
	// 		paths <- path.Join(versionPath, s)
	// 	}
	// 	for i := 0; i < len(strs); i++ {
	// 		// <-kvPairs
	// 		k := <-kvPairs
	// 		fmt.Println("received! ", k.Key, len(k.Value))
	// 		kvs = append(kvs, k)
	// 	}
	// 	// break
	// }
	// fmt.Println("waiting ...")
	// resp, err := api.WriteWorkersKVBulk(context.Background(), namespaceID, kvs)
	// util.Check(err)
	// fmt.Println(resp)

	os.Exit(1)

}
