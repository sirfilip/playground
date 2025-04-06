package task_manager

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Api struct {
	mux    *http.ServeMux
	tm     *TaskManager
	logger *slog.Logger
}

func NewApi(repo Repository, logger *slog.Logger) *Api {
	api := &Api{
		mux:    http.NewServeMux(),
		tm:     NewTaskManager(repo),
		logger: logger,
	}
	api.registerRoutes()
	return api
}

func (api *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mux.ServeHTTP(w, r)
}

func (api *Api) registerRoutes() {
	api.mux.HandleFunc("POST /schedule-task", scheduleTask(api.tm, api.logger))
}

func scheduleTask(tm *TaskManager, logger *slog.Logger) http.HandlerFunc {
	type createTaskForm struct {
		Author      string `json:"author"`
		Title       string `json:"title"`
		Description string `json:"description"`
		DueDate     int64  `json:"due_date"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var form createTaskForm
		// limit body to max 1MB
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&form); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		owner, err := NewOwner(form.Author)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		taskTitle, err := NewTaskTitle(form.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		taskDesc, err := NewTaskDescription(form.Description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dueDate, err := NewDueDate(time.Unix(form.DueDate, 0))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		id, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			logger.Error(err.Error())
			return

		}
		taskId, err := NewTaskID(id.String())

		cmd := CreateTaskCmd{
			Owner:       owner,
			Title:       taskTitle,
			Description: taskDesc,
			DueDate:     dueDate,
			TaskId:      taskId,
		}

		if err := tm.Schedule(r.Context(), cmd); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status": "created"}`))
	}
}
