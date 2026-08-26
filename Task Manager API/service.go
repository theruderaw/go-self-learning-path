package main

import "errors"

func createTaskService(title string, createdBy string) (int , error) {
	if title == "" {
		return -1, errors.New("title is required")
	}

		return createTask(title, createdBy), nil
}

func getTasksService() []Task {
	return getTasks()
}

func getTaskService(id int) (Task, error) {
	return getTask(id)
}

func completeTaskService(id int) error {
	return completeTask(id)
}

func updateTaskService(id int, title string) error {
	if title == "" {
		return errors.New("title is required")
	}

	return updateTask(id, title)
}

func deleteTaskService(id int) error {
	return deleteTask(id)
}
