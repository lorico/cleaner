package main

import (
	"COMMON/db"
	//"github.com/djherbis/times"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var dbFile = "right.db"

func main() {
	//rightPath := "./public" // the one we KEEP
	//leftPath := "./left"    // the one we can REMOVE
	// === PARAMETERS
	usage := `USAGE=
		cleaner -left=[path to folder where duplicates will be REMOVED !] -right=[path to master folder (NOTHING gets removed in that one...)]
		`
	paramsLeftPtr := flag.String("left", "", "folder to CLEAN (ie: duplicate files here will be removed!")
	paramsRightPtr := flag.String("right", "", "folder to ANALYSE (ie: that's your master folder, nothing will be removed from here...")
	flag.Parse()
	tail := flag.Args()
	if *paramsLeftPtr == "" {
		log.Println("Parameters incomplete (missing -left=...).")
		log.Println(usage)
		return
	}
	if *paramsRightPtr == "" {
		log.Println("Parameters incomplete (missing -right=...).")
		log.Println(usage)
		return
	}
	if len(tail) != 0 {
		log.Println("Wrong Parameters (too many).")
		log.Println(usage)
		return
	}
	leftPath := *paramsLeftPtr
	rightPath := *paramsRightPtr

	// === GO !
	err := os.Remove(dbFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	var createStmts []string
	// TABLE: master
	createStmts = append(createStmts, `
	CREATE TABLE IF NOT EXISTS right (
		Id INTEGER PRIMARY KEY, 
		Name TEXT NOT NULL, 
		Size INT,
		ModifDate TEXT,
		Path TEXT
		)
	;
	`)
	if err := db.InitDB(dbFile, createStmts); err != nil {
		log.Println("ERROR: ", err.Error())
		return
	}

	// ANALYSE right (generate DB)
	target(rightPath)
	log.Println("============================================")
	log.Println("============================================")
	log.Println("============================================")
	// REMOVE from left
	source(leftPath)

}

func target(searchDir string) {
	cnt := 0
	log.Println("will check folder (right): ", searchDir)
	//fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		log.Println("***")
		/*
			log.Println("R-Path: ", path)
			log.Println("R-IsDir: ", f.IsDir())
			log.Println("R-Name: ", f.Name())
			log.Println("R-Size: ", f.Size())
			log.Println("R-ModTime: ", f.ModTime())
		*/
		if !f.IsDir() {
			add := `INSERT INTO right (Name, Size, ModifDate, Path)
					VALUES('` + f.Name() + `', '` + strconv.FormatInt(f.Size(), 10) + `', '` + f.ModTime().Format(time.RFC3339) + `', '` + path + `')
					`
			db.Write(add)
			log.Println("=> Written! ", add)
			cnt++
		}
		/*
			t, err := times.Stat(path)
			if err != nil {
				log.Fatal(err.Error())
			}
			log.Println("Birth time: ", t.BirthTime())
			log.Println("Modification Time: ", t.ModTime())
			log.Println("Change Time: ", t.ChangeTime())
			log.Println("Access Time: ", t.AccessTime())
		*/
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("+++ +++ +++")
	log.Println("SUMMARY RIGHT: # of FILES= ", cnt)
	log.Println("+++ +++ +++")
	return
}

func source(searchDir string) {
	cnt := 0
	log.Println("will analyse folder (left): ", searchDir)
	//fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		log.Println("***")
		log.Println("L-Path: ", path)
		log.Println("L-IsDir: ", f.IsDir())
		log.Println("L-Name: ", f.Name())
		log.Println("L-Size: ", f.Size())
		log.Println("L-ModTime: ", f.ModTime())
		if !f.IsDir() {
			sel := `SELECT Name, Size, ModifDate FROM right
					WHERE
						Name = '` + f.Name() + `'
						AND Size = '` + strconv.FormatInt(f.Size(), 10) + `'
						AND ModifDate = '` + f.ModTime().Format(time.RFC3339) + `'
					`
			res, err := db.Select(sel)
			if err != nil {
				log.Fatal(err)
				return err
			}

			// Print results (not mandatory here, as we only check if exists YES or NO)
			cntRows := len(res)
			log.Println("LEN= ", cntRows)
			if cntRows != 0 {
				log.Println("===> REMOVE file: ", path)
				if err := os.Remove(path); err != nil {
					log.Fatal("ERROR trying to remove: ", path, err)
				}
				cnt++
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("+++ +++ +++")
	log.Println("SUMMARY LEFT: # of files REMOVED= ", cnt)
	log.Println("+++ +++ +++")
	return
}
