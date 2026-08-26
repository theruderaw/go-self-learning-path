package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileItem struct {
	Path string
	Filename  string
	Extension string
}

type OrganiseOutput struct {
	Files   int
	Folders int
}

var files = make(map[string][]FileItem)

func MakeFolders(output string) error {
	for key,value := range files {

		folder := filepath.Join(output,key[1:])

		err := os.MkdirAll(folder,0755)
		if err != nil {
			return err
		}
		err = WriteFiles(folder,value)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteFiles(loc string,files []FileItem) error {
	for _,file := range files {
		err := os.Rename(file.Path, filepath.Join(loc,file.Filename))
		if err != nil {
			return err
		}
	}
	return nil
}

func StoreFiles(dirPath string, filename string) (FileItem) {
	file := FileItem{
		Path: filepath.Join(dirPath,filename),
		Filename: filename,
		Extension: filepath.Ext(filename),
	}
	files[file.Extension] = append(files[file.Extension],file)
	return file
}

func GetFiles(path string) (OrganiseOutput, error) {
	data, err := os.ReadDir(path)

	var output OrganiseOutput

	if err != nil {
		return output, err
	}

	for _, value := range data {
		if value.IsDir() {
			data,err := GetFiles(
				filepath.Join(path, value.Name()),
			)
			if err != nil {
				return OrganiseOutput{},err
			}
			output.Files += data.Files
			output.Folders += data.Folders
			output.Folders++
		} else {
			output.Files++
			
			data := StoreFiles(
				path, value.Name(),
			)
			fmt.Printf("File %s\n", value.Name())
			fmt.Println(data)
		}
	}

	return output, nil
}



func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: gofile path/to/folder/")
		return
	}

	data, err := GetFiles(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	err = MakeFolders(os.Args[2]) 
	if err != nil {
		fmt.Println(err)
		return
	}	
	fmt.Println("Memory has",files,"files")
	fmt.Printf("Organised %d files into %d folders", data.Files, data.Folders)
}
