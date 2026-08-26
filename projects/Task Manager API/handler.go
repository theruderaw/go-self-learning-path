package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type CreateTaskPayload struct{
	Title string `json:"title"`
	CreatedBy string `json:"createdBy"`
}

type UpdateTaskPayload struct{
	Title string `json:"title"`
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskPayload

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"Invalid JSON",http.StatusBadRequest)
		return
	}

	id, err := createTaskService(req.Title, req.CreatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]int{
		"id":id,
	})
}

func getTasksHandler(w http.ResponseWriter,r *http.Request) {
	tasks := getTasksService()

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(tasks)
}

func getTaskHandler(w http.ResponseWriter,r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w,"Invalid ID",http.StatusBadRequest)
		return
	}
	
	task,err := getTaskService(id)
	if err != nil {
		http.Error(w, err.Error(),http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(task)
}

func updateTaskHandler(w http.ResponseWriter,r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var req UpdateTaskPayload

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w,"Invalid JSON", http.StatusBadRequest)
		return 
	}

	err = updateTaskService(id, req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = completeTaskService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = deleteTaskService(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


func getID(r *http.Request) (int,error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return -1,err
	}
	return id, nil
}
