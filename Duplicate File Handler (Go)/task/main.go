package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	. "fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		Println("Directory is not specified")
		return
	}
	var ext, option string

	pathMap := make(map[int][]string)

	Println("Enter file format:")
	Scanf("%s\n", &ext)

	Println("Size sorting options:")
	Println("1. Descending")
	Println("2. Ascending")

	Println()

	for {
		Println("Enter a sorting option:")
		Scan(&option)
		Println()

		if option != "1" && option != "2" {
			Println("Wrong option")
			Println()
		} else {
			runFileTree(ext, option, pathMap)
			break
		}
	}
}

func runFileTree(ext, option string, m map[int][]string) {
	var sortedKeys []int
	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if !info.IsDir() && (filepath.Ext(path) == "."+ext || ext == "") {
			sortedKeys = make([]int, len(m))

			if _, ok := m[int(info.Size())]; ok {
				m[int(info.Size())] = append(m[int(info.Size())], path)
			} else {
				m[int(info.Size())] = []string{path}
			}

			for keys := range m {
				sortedKeys = append(sortedKeys, keys)
			}

			switch option {
			case "1":
				sort.Slice(m[int(info.Size())], func(i, j int) bool {
					return m[int(info.Size())][i] > m[int(info.Size())][j]
				})

				sort.Slice(sortedKeys, func(i, j int) bool {
					return sortedKeys[i] > sortedKeys[j]
				})

			case "2":
				sort.Strings(m[int(info.Size())])
				sort.Slice(sortedKeys, func(i, j int) bool {
					return sortedKeys[i] < sortedKeys[j]
				})
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, sizes := range sortedKeys {
		if sizes == 0 {
			continue
		}
		Println(sizes, "bytes")
		for _, el := range m[sizes] {
			Println(el)
		}
		Println()
	}

	checkDuplicates(sortedKeys, m)
}

func checkDuplicates(keys []int, pathMap map[int][]string) {
	Println("Check for duplicates?")
	var yesOrNo string
	Scan(&yesOrNo)
	Println()

	switch strings.ToLower(yesOrNo) {
	case "yes":
		// we will use this 2D map to group the paths by hash values
		hashedFileMap := make(map[int]map[string][]string)

		for _, sizes := range keys {
			if sizes == 0 || len(pathMap[sizes]) < 2 {
				continue
			}
			hashedFileMap[sizes] = make(map[string][]string)
			for _, el := range pathMap[sizes] {
				file, err := os.Open(el)
				if err != nil {
					log.Fatal(err)
				}

				md5Hash := md5.New()
				if _, errorz := io.Copy(md5Hash, file); errorz != nil {
					log.Fatal(errorz)
				}

				hashString := hex.EncodeToString(md5Hash.Sum(nil))
				if _, ok := hashedFileMap[sizes][hashString]; ok {
					hashedFileMap[sizes][hashString] = append(hashedFileMap[sizes][hashString], el)
				} else {
					hashedFileMap[sizes][hashString] = []string{el}
				}
			}
		}

		var count int
		var storedDupes []string

		for _, sizes := range keys {
			if sizes == 0 || len(pathMap[sizes]) < 1 {
				continue
			}
			Println(sizes, "bytes")
			for hash, el := range hashedFileMap[sizes] {
				if len(el) <= 1 {
					continue
				}
				Printf("Hash: %s\n", hash)
				for i, pathEl := range el {
					count++
					pathEl = strings.ReplaceAll(hashedFileMap[sizes][hash][i], pathEl, Sprintf("%d. %s", count, pathEl))
					hashedFileMap[sizes][hash][i] = pathEl
					storedDupes = append(storedDupes, pathEl)
					Println(pathEl)
				}
			}
			Println()
		}
		deleteDuplicates(storedDupes)

	case "no":
		return
	default:
		Println("Wrong option")
		checkDuplicates(keys, pathMap)
	}
}

func deleteDuplicates(dupes []string) {
	var yesOrNo string
	Println("Delete files?")
	Scan(&yesOrNo)

	switch yesOrNo {
	case "yes":
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		if input == "" {
			Println("Wrong format")
			deleteDuplicates(dupes)
		}

		inputSlice := strings.Fields(input)

		var memFreed int
		for inpEl := range inputSlice {
			_, err := strconv.Atoi(inputSlice[inpEl])
			if err != nil {
				Println("Wrong format")
				deleteDuplicates(dupes)
				return
			}
			for _, el := range dupes {
				elNoSuffix := strings.TrimSuffix(el, filepath.Ext(el))
				elSlice := strings.Split(elNoSuffix, ". ")
				if elSlice[0] == inputSlice[inpEl] {
					rawPath := strings.Join(elSlice[1:], "")
					rawPath = Sprintf("%s%s", rawPath, filepath.Ext(el))

					fileStat, _ := os.Stat(rawPath)
					size := int(fileStat.Size())
					memFreed += size
					os.Remove(rawPath)
				}
			}
		}

		Printf("Total freed up space: %d bytes\n", memFreed)

	case "no":
		return
	default:
		Println("Wrong option")
		deleteDuplicates(dupes)
	}
}
