package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"path"
	"path/filepath"
	"time"
)

const NOEXT = "NO EXTENSION"

//Takes a dir as arg
//displays categorized information about size, type, etc

type FInfo struct {
	Bytes int64
	Name  string
	Err   error
}

type Result struct {
	Type            string
	TotalSize       int64
	Num             uint64
	Top5FilesBySize []FInfo
}

func main() {
	//var args struct {
	//	Dir string `arg:"positional, required"`
	//}
	//arg.MustParse(&args)
	//fmt.Println(args.Dir)
	//dir := "/home/me/temp/learngo/recapgo"
	dir := "/home/me"
	//dir := "/home/me/temp/learngo/testts"
	//dir := "/home/me/temp/learngo/recapgoxxx"
	now := time.Now()

	ch := putInfo(dir)
	rawMap := processInfoIntoRawMap(ch)
	resultMap := rawMapToResultMap(rawMap)
	printResultMap(resultMap)

	fmt.Println("Elapsed ", time.Since(now))

}

func printResultMap(resultMap map[string]Result) {
	fmt.Println("EXT | SIZE | NUM")
	for k, v := range resultMap {
		//fmt.Println(k, v)
		//fmt.Printf("%s val: %v\n", k, v)
		fmt.Printf("%s | ", k)
		fmt.Printf("%v | ", v.Num)
		fmt.Printf("%s\n", humanize.Bytes(uint64(v.TotalSize)))
	}
	fmt.Println()
}

func rawMapToResultMap(rawMap map[string][]FInfo) map[string]Result {
	resultMap := make(map[string]Result)

	resultsCh := make(chan Result, len(rawMap))
	defer close(resultsCh)

	for k, v := range rawMap {
		if _, ok := resultMap[k]; !ok {
			resultMap[k] = Result{}
		}
		go func(k string, v []FInfo) {
			resultsCh <- finfoSliceToResult(k, v)
		}(k, v)
	}
	for a := 0; a < len(rawMap); a++ {
		v := <-resultsCh
		resultMap[v.Type] = v
	}

	return resultMap
}

func finfoSliceToResult(fType string, fSlice []FInfo) Result {
	res := Result{}
	if len(fSlice) > 0 {
		res.Type = path.Ext(fSlice[0].Name)
	}

	for _, v := range fSlice {
		if v.Err != nil {
			continue
		}
		res.TotalSize += v.Bytes
		res.Num++
	}
	return res
}

func processInfoIntoRawMap(in <-chan FInfo) map[string][]FInfo {
	rawMap := make(map[string][]FInfo)
	for finfo := range in {
		ext := path.Ext(finfo.Name)
		if len(ext) == 0 {
			ext = NOEXT
		}

		if _, ok := rawMap[ext]; !ok {
			rawMap[ext] = make([]FInfo, 1)
			rawMap[ext][0] = finfo
		} else {
			rawMap[ext] = append(rawMap[ext], finfo)
		}
	}
	return rawMap
}

func putInfo(dir string) <-chan FInfo {
	ch := make(chan FInfo)
	go func() {
		defer close(ch)
		walk := func(s string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			finfo := FInfo{}
			if err != nil {
				finfo.Err = err
			} else {
				finfo.Name = info.Name()
				finfo.Bytes = info.Size()
				finfo.Err = nil
			}
			ch <- finfo
			return nil
		}

		err := filepath.Walk(dir, walk)
		if err != nil {
			fmt.Println(err)
			return
		}

	}()
	return ch
}
