package main

import (
	"cmp"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"
)

//todo
// --raw bytes option
// -- reverse sort option
// -- top 10 option

import (
	"github.com/dustin/go-humanize"
)

const NOEXT = "__NOEXT__"

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
	dir := "/home/me/temp/learngo/testts2"
	//dir := "/home/me"
	//dir := "/home/me/temp/learngo/recapgoxxx"
	now := time.Now()

	ch := putInfo(dir)
	rawMap := processInfoIntoRawMap(ch)
	resultSlice := rawMapToResultSlice(rawMap)
	sortResultSlic(resultSlice)
	printResultSlice(resultSlice)

	fmt.Println("Elapsed ", time.Since(now))

}

func sortResultSlic(resultSlice []Result) {
	slices.SortFunc(resultSlice, func(a, b Result) int {
		return cmp.Or(
			cmp.Compare(a.TotalSize, b.TotalSize),
		)
	})
}

func printResultSlice(resultSlice []Result) {
	fmt.Println("EXT | SIZE | NUM")
	for _, v := range resultSlice {
		//fmt.Println(k, v)
		//fmt.Printf("%s val: %v\n", k, v)
		fmt.Printf("%s | ", v.Type)
		fmt.Printf("%v | ", v.Num)
		fmt.Printf("%s\n", humanize.Bytes(uint64(v.TotalSize)))
	}
	fmt.Println()
}

func rawMapToResultSlice(rawMap map[string][]FInfo) []Result {
	var resultSlice []Result

	//using worker pool pattern
	resultsCh := make(chan Result, len(rawMap))
	defer close(resultsCh)

	for k, v := range rawMap {
		go func(k string, v []FInfo) {
			resultsCh <- finfoSliceToResult(k, v)
		}(k, v)
	}
	//note how we don't use range but the len of teh jobs
	for a := 0; a < len(rawMap); a++ {
		v := <-resultsCh
		//resultMap[v.Type] = v
		resultSlice = append(resultSlice, v)

	}

	return resultSlice
}

func finfoSliceToResult(fType string, fSlice []FInfo) Result {
	res := Result{}
	if len(fSlice) > 0 {
		res.Type = fType
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
