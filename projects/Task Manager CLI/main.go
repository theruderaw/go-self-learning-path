package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type TaskItem struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

var tasks []TaskItem

const STORE_FILE = "tasks.json"

func nextID() int {
	maxId := 0

	for _, task := range tasks {
		if task.ID > maxId {
			maxId = task.ID
		}
	}

	return maxId + 1
}

func GetTasks() []TaskItem {
	data, err := os.ReadFile(STORE_FILE)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = json.Unmarshal(data, &tasks)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return tasks
}

func addTask(
	text string,
	done bool,
) error {
	tasks = GetTasks()

	task := TaskItem{
		ID:   nextID(),
		Text: text,
		Done: done,
	}
	tasks = append(tasks, task)

	data, err := json.Marshal(tasks)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(STORE_FILE, data, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetTask(id int) (item TaskItem, err error) {
	tasks = GetTasks()
	for _, value := range tasks {
		if value.ID == id {
			return value, nil
		}
	}
	return TaskItem{}, errors.New("ID not found")
}

func MarkDone(id int) error {

	tasks = GetTasks()
	found := false

	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Done = true
			found = true
			break
		}
	}

	if !found {
		return errors.New("ID not found")
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		fmt.Println(err)
		return errors.New("JSON encoding error")
	}

	err = os.WriteFile(STORE_FILE, data, 0644)
	if err != nil {
		fmt.Println(err)
		return errors.New("File write error")
	}

	return nil

}

func DeleteTask(id int) error {
	tasks = GetTasks()
	found := false

	for index, value := range tasks {
		if value.ID == id {
			tasks = append(tasks[:index], tasks[index+1:]...)
			found = true
			break
		}
	}

	if !found {
		return errors.New("ID not found")
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = os.WriteFile(STORE_FILE, data, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage: godo <command>")
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("usage: godo add <task>")
			return
		}

		fmt.Println("Adding:", os.Args[2])
		err := addTask(
			os.Args[2],
			false,
		)
		if err != nil {
			return
		}

	case "list":
		fmt.Println("Listing tasks")
		for _, task := range GetTasks() {
			fmt.Printf("%d [%t] %s\n", task.ID, task.Done, task.Text)
		}

	case "done":
		if len(os.Args) < 3 {
			fmt.Println("usage: godo done <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		err = MarkDone(id)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Completing:", os.Args[2])

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("usage: godo delete <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		err = DeleteTask(id)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Deleting", os.Args[2])

	default:
		fmt.Println("unknown command:", os.Args[1])
	}
}
