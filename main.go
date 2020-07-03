package main

import (
	"fmt"
	"io/ioutil"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/util"
)

func main() {
	files, err := ioutil.ReadDir(util.GetCDNJSPackages())
	util.Check(err)

	fail := []string{}

	for i := 2118; i < len(files); i++ {
		f := files[i]
		vs, latest := npm.GetVersions(f.Name())
		// fmt.Println(i, f.Name())
		//fmt.Printf("%d - %s\n", i, f.Name())
		if latest == "failure" || latest == "time_failure" {
			fmt.Printf("FAIL - %d - %s (%d - %s)\n", i, f.Name(), len(vs), latest)
			fail = append(fail, f.Name())
		}
	}

	fmt.Println(fail)
}
