package main

import "errors"

var tasks []Task
var counter int


func createTask(title string,createdBy string) int {
	counter++
	task := Task{
		ID: counter,
		Title: title,
		Completed: false,
		CreatedBy: createdBy,
	}
	tasks = append(tasks, task)
	return counter
}

func getTasks() []Task {
	return tasks
}

func getTask(id int) (Task, error) {
	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return Task{}, errors.New("Task not found")
}

func completeTask(id int) error {
	for i, task := range tasks {
		if task.ID == id {
			if !task.Completed {
				tasks[i].Completed = true
				return nil
			}

			return errors.New("Already completed")
		}
	}

	return errors.New("Task not found")
}

func updateTask(id int, title string) error {
	for i, task := range tasks {
		if task.ID == id {
			if task.Completed {
				return errors.New("Already completed")
			}

			tasks[i].Title = title
			return nil
		}
	}

	return errors.New("Task not found")
}

func deleteTask(id int) error {
	for index, task := range tasks {
		if task.ID == id {
			if task.Completed {
				tasks = append(tasks[:index], tasks[index+1:]...)
				return nil
			}

			return errors.New("Task not Completed")
		}
	}

	return errors.New("Task not found")
}
