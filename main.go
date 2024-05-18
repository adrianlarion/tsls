package main

//Takes a dir as arg
//displays categorized information about size, type, etc

import (
	"cmp"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
	"os"
	"path"
	"path/filepath"
	"slices"
)

const NOEXT = "__NOEXT__"

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
	var args struct {
		Dir     string `arg:"positional" help:"target directory"`
		Bytes   bool   `arg:"-b, --bytes" help:"show bytes instead of human readable size"`
		Reverse bool   `arg:"-r, --reverse" help:"reverse sort"`
	}
	arg.MustParse(&args)

	//now := time.Now()

	//if no dir supplied, use current dir
	if len(args.Dir) == 0 {
		ex, err := os.Executable()
		if err != nil {
			fmt.Println(err)
			return
		}
		args.Dir = filepath.Dir(ex)
	} else {
		//check if dir exists
		if _, err := os.Stat(args.Dir); err != nil {
			fmt.Println(err)
			return
		}
	}

	ch := putInfo(args.Dir)
	rawMap := processInfoIntoRawMap(ch)
	resultSlice := rawMapToResultSlice(rawMap)
	sortResultSlic(resultSlice, args.Reverse)
	printResultSlice(resultSlice, args.Bytes)

	//fmt.Println("Elapsed ", time.Since(now))

}

func sortResultSlic(resultSlice []Result, reverse bool) {

	slices.SortFunc(resultSlice, func(a, b Result) int {
		if reverse {
			return cmp.Or(
				-cmp.Compare(a.TotalSize, b.TotalSize),
			)

		} else {
			return cmp.Or(
				cmp.Compare(a.TotalSize, b.TotalSize),
			)
		}
	})
}

func printResultSlice(resultSlice []Result, bytes bool) {
	//fmt.Println("EXT | SIZE | NUM")
	for _, v := range resultSlice {
		//show either human readable or bytes
		size := humanize.Bytes(uint64(v.TotalSize))
		if bytes {
			size = fmt.Sprintf("%v", v.TotalSize)
		}

		fmt.Printf("%s | ", v.Type)
		fmt.Printf("%v | ", v.Num)
		fmt.Printf("%s\n", size)
	}
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
