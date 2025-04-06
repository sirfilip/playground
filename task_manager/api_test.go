package task_manager

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestApiScheduleTask(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Error(err)
	}
	db := connect(t)
	t.Cleanup(func() {
		db.Close()
		devNull.Close()
	})

	logger := slog.New(slog.NewJSONHandler(devNull, nil))
	api := NewApi(&SqliteRepo{db: db}, logger)

	srv := httptest.NewServer(api)
	defer srv.Close()

	client := srv.Client()
	res, err := client.Post(fmt.Sprintf("%s%s", srv.URL, "/schedule-task"), "Content-type: application/json", strings.NewReader(fmt.Sprintf(`
		{
			"author": "filip",
			"due_date": %d
		}
	`, time.Now().Add(1*time.Minute).Unix())))
	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Errorf("want status 201 got: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if string(body) != `{"status": "created"}` {
		t.Errorf(`want: {"status": "created"} got: %v`, string(body))
	}
}
