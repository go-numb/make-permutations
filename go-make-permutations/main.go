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

	json "github.com/dustin/gojson"
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
	InnerHash []D       `csv:"inner_hash"`
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

	log.Infof("%v", patterns[len(patterns)-1])

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

	// カテゴリ数を数える
	numbers := make([]int, categoriesN)

	for i := range df {
		numbers[df[i].CategoryID]++
	}

	fmt.Printf("カテゴリ別ファイル数: %v", numbers)

	// map[pattern][4]int
	var pattern int = numbers[0]
	for i := range numbers {
		if len(numbers) == i+1 {
			break
		}
		pattern *= numbers[i+1]
	}

	fmt.Printf("パターン総数: %d\n", pattern)

	// カテゴリごとの配列を作る
	arrays := make(map[int][]int, categoriesN)
	for i := range df {
		arrays[df[i].CategoryID] = append(arrays[df[i].CategoryID], df[i].ID)
	}

	fmt.Printf("カテゴリ別ファイルID配列%v\n", arrays)

	// var currentCategories int
	results := make([][4]int, pattern)
	fmt.Printf("%d -- %#v", len(results), results)
	var (
		// i = パターン総数
		// j = カテゴリID
		// k = カテゴリ内ファイル配列
		k int = 0
	)

	fmt.Printf("%d\n", len(arrays))

	for j := 0; j < len(arrays); j++ {
		for i := range results {
			if i != 0 && i%len(arrays[j]) == 0 {
				k = 0
			}

			results[i][j] = arrays[j][k]
			// カテゴリ内のファイルを一通り代入したら繰り上げる
			k++
		}
		k = 0
	}

	if pattern != len(results) {
		log.Fatal("事前計算したパターン数と実際組み合わせたパターン数が違う")
	}
	fmt.Printf("事前計算したパターン数 == 実際組み合わせたパターン数\n%d == %d\n", pattern, len(results))

	NFTs := make([]NFT, len(results))
	for i := range results {
		inner := make([]D, categoriesN)
		name := "" // The 適当
		for j := range results[i] {
			for k := range df {
				if results[i][j] == df[k].ID {
					inner[df[k].CategoryID] = df[k]
					name += df[k].Name
				}
			}
		}

		// 組み合わせた名前から適当に
		hash := sha256.Sum256([]byte(name))
		NFTs[i] = NFT{
			ID:        i,
			Hash:      hex.EncodeToString([]byte(hash[:])),
			Name:      name,
			InnerHash: inner,
			CreatedAt: time.Now(),
		}
	}

	f, _ := os.Create("data.csv")
	defer f.Close()

	b, err := json.Marshal(NFTs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", string(b))

	f.Write(b)
}
