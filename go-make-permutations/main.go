package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/labstack/gommon/log"
)

const (
	DIRROOT        = "../images"
	CATEGORYPREFIX = "_"
)

type temp struct {
	ID     int
	Name   string
	Path   string
	Choose bool
}

type D struct {
	ID         int
	CategoryID int
	Category   string
	NameID     int
	Name       string
	Path       string
	PermHash   string
	CreatedAt  time.Time
}

type NFT struct {
	ID   int    `csv:"id"`
	Hash string `csv:"hash"`
	Name string `csv:"name"`
	// 構成部をkeyにし構成内容をmap
	InnerHash map[int]D `csv:"inner_hash"`
	CreatedAt time.Time `csv:"created_at"`
}

/*
カテゴリからユニークなファイルを選択し（a
さらに次カテゴリでファイルを選択し（aを頂点にしたツリーを作る
*/

func main() {
	// 画像数
	var filen int
	// images内フォルダ（各部カテゴリ）
	var index int
	var patterns []temp

	if err := filepath.WalkDir(DIRROOT, func(s string, info fs.DirEntry, _ error) error {
		if !info.IsDir() {
			// ファイル総数を数える
			filen++
			return nil
		}

		// RootDirを捨てる
		if strings.Contains(DIRROOT, info.Name()) {
			return nil
		}

		patterns = append(patterns, temp{
			ID:   index,
			Name: info.Name(),
			Path: s,
		})
		index++

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	log.Infof("%#v\n", patterns)

	// ファイルを内包しているもののみを使用する
	var categoriesN int
	for i, part := range patterns {
		if err := filepath.WalkDir(part.Path, func(_ string, info fs.DirEntry, _ error) error {
			if info.IsDir() {
				return nil
			}

			if !patterns[i].Choose {
				patterns[i].Choose = true
				categoriesN++
			}

			return nil
		}); err != nil {
			log.Fatal(err)
		}
	}

	// ファイル管理用データ
	var df = make([]D, filen)
	index = 0
	for categoryID, pattern := range patterns {
		if !pattern.Choose {
			continue
		}

		if err := filepath.WalkDir(pattern.Path, func(s string, info fs.DirEntry, _ error) error {
			if info.IsDir() {
				return nil
			}

			df[index] = D{
				ID:         index,
				CategoryID: categoryID,
				Category:   pattern.Name,
				NameID:     index,
				Name:       info.Name(),
				Path:       s,
				CreatedAt:  time.Now(),
			}
			index++

			return nil
		}); err != nil {
			log.Fatal(err)
		}
	}

	// ユニークなHashを作る
	str := make([]string, len(df))
	for i := range df {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%d-%s", df[i].CategoryID, df[i].NameID, df[i].Name)))
		df[i].PermHash = hex.EncodeToString([]byte(hash[:]))
	}

	fmt.Println("使用部品種: ", categoriesN)
	fmt.Printf("ファイル数: %d, hash数: %d\n", filen, len(str))

	index = 0

	var inputs = make([][]string, categoriesN)

	for i := range df {
		inputs[df[i].CategoryID] = append(inputs[df[i].CategoryID], df[i].PermHash)
	}

	results := pairs(inputs)

	nfts := make([]NFT, len(results))
	for i := range results {
		nfts[i] = NFT{
			ID:        i,
			Name:      fmt.Sprintf("hoge_%d", i),
			InnerHash: map[int]D{},
			CreatedAt: time.Now(),
		}

		nfts[i].Hash = fmt.Sprintf("hoge_%d_%s", i, nfts[i].CreatedAt.String())

		for j := range results[i] {
			for k := range df {
				if results[i][j] == df[k].PermHash {
					nfts[i].InnerHash[df[k].CategoryID] = df[k]
				}
			}
		}
	}

	for i := range nfts {
		fmt.Printf("%d - %v\n", i, nfts[i])
	}

	// Save the data
	f, err := os.Create("./data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := gocsv.MarshalFile(nfts, f); err != nil {
		log.Fatal()
	}

}

func pairs(words [][]string) [][]string {
	var pairs [][]string
	for i := range words {
		for _, word := range words[i] {
			if i+1 >= len(words) {
				break
			}
			for _, word1 := range words[i+1] {
				pairs = append(pairs, []string{word, word1})
			}
		}
	}
	return pairs
}
