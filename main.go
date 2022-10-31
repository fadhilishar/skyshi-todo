// Go package
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type (
	Activity struct {
		ID        int            `gorm:"primarykey" json:"id"`
		Email     string         `json:"email"`
		Title     string         `json:"title" validate:"required"`
		CreatedAt time.Time      `json:"created_at"`
		UpdatedAt time.Time      `json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	}

	Todo struct {
		ID              int            `gorm:"primarykey" json:"id"`
		ActivityGroupID int            `json:"activity_group_id" validate:"required"`
		Title           string         `json:"title" validate:"required"`
		IsActive        string         `json:"is_active" default:"1"`
		Priority        string         `json:"priority" default:"very-high"`
		CreatedAt       time.Time      `json:"created_at"`
		UpdatedAt       time.Time      `json:"updated_at"`
		DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	}

	GetTodoResponse struct {
		ID              int            `gorm:"primarykey" json:"id"`
		ActivityGroupID string         `json:"activity_group_id" validate:"required"`
		Title           string         `json:"title" validate:"required"`
		IsActive        bool           `json:"is_active" default:"1"`
		Priority        string         `json:"priority" default:"very-high"`
		CreatedAt       time.Time      `json:"created_at"`
		UpdatedAt       time.Time      `json:"updated_at"`
		DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	}

	CreateTodoReq struct {
		ActivityGroupID string         `json:"activity_group_id" validate:"required"`
		Title           string         `json:"title" validate:"required"`
		IsActive        string         `json:"is_active" default:"1"`
		Priority        string         `json:"priority" default:"very-high"`
		CreatedAt       time.Time      `json:"created_at"`
		UpdatedAt       time.Time      `json:"updated_at"`
		DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	}

	TodoFilter struct {
		ActivityGroupID uint `json:"activity_group_id"`
	}

	Response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}

	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Using gorm, connecting to mysql
	dsn := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`, os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), "3306", os.Getenv("MYSQL_DBNAME"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to mysql database")
	}

	// Migrate the schema
	db.AutoMigrate(
		&Activity{},
		&Todo{},
	)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to API Todo",
		})
	})

	////////////////////////
	// 1. ACTIVITY ROUTES //
	////////////////////////
	activityGroups := e.Group("/activity-groups")
	// Get All :
	activityGroups.GET("", func(c echo.Context) error {
		activities := new([]Activity)
		db.Find(&activities)
		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    activities,
		})
	})

	// Get One :
	activityGroups.GET("/:id", func(c echo.Context) error {
		id := c.Param("id")
		activityDB := new(Activity)
		db.First(&activityDB, id)
		idInt, _ := strconv.Atoi(id)
		if int(activityDB.ID) != idInt {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Activity with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    activityDB,
		})
	})

	// Create :
	activityGroups.POST("", func(c echo.Context) error {
		// Validate Request
		activity := new(Activity)
		if err := c.Bind(activity); err != nil {
			return err
		}
		if err := c.Validate(activity); err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "title cannot be null",
				Data:    map[string]any{},
			})
		}

		// Inserting Activity
		db.Create(&activity)

		return c.JSON(http.StatusCreated, Response{
			Status:  "Success",
			Message: "Success",
			Data:    activity,
		})
	})

	// Delete :
	activityGroups.DELETE("/:id", func(c echo.Context) error {
		// Finding Activity
		id := c.Param("id")
		activityDB := new(Activity)
		db.First(&activityDB, id)
		idInt, _ := strconv.Atoi(id)
		if int(activityDB.ID) != idInt {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Activity with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		// Deleting Activity
		db.Delete(&activityDB)

		db.Where("activity_id = ?", activityDB.ID).Delete(&Todo{})

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    map[string]any{},
		})
	})

	// Update :
	activityGroups.PATCH("/:id", func(c echo.Context) error {
		// Validate Request
		activity := new(Activity)
		if err := c.Bind(activity); err != nil {
			return err
		}
		if err := c.Validate(activity); err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "title cannot be null",
				Data:    map[string]any{},
			})
		}

		// Finding Activity
		id := c.Param("id")
		activityDB := new(Activity)
		db.First(&activityDB, id)
		idInt, _ := strconv.Atoi(id)
		if int(activityDB.ID) != idInt {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Activity with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		// Updating Activity
		activityDB.Title = activity.Title
		if activity.Email != "" {
			activityDB.Email = activity.Email
		}
		db.Save(&activityDB)

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    activityDB,
		})
	})

	////////////////////
	// 2. TODO ROUTES //
	////////////////////
	todoItems := e.Group("/todo-items")

	// Get All :
	todoItems.GET("", func(c echo.Context) error {
		// Validate Request
		activityID := c.QueryParam("activity_group_id")
		activityIDInt := 0
		todos := []Todo{}
		if activityID != "" {
			activityIDInt, _ = strconv.Atoi(activityID)
			db.Where("activity_id=?", activityIDInt).Find(&todos)
		} else {
			db.Find(&todos)
		}
		todosResp := []GetTodoResponse{}
		for _, todo := range todos {
			isActive := true
			if todo.IsActive != "1" {
				isActive = false
			}
			todosResp = append(todosResp, GetTodoResponse{
				ID:              todo.ID,
				ActivityGroupID: strconv.Itoa(todo.ActivityGroupID),
				Title:           todo.Title,
				IsActive:        isActive,
				Priority:        todo.Priority,
				CreatedAt:       todo.CreatedAt,
				UpdatedAt:       todo.UpdatedAt,
				DeletedAt:       todo.DeletedAt,
			})
		}

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todosResp,
		})
	})

	// Get One :
	todoItems.GET("/:id", func(c echo.Context) error {
		id := c.Param("id")
		todoDB := new(Todo)
		res := db.First(&todoDB, id)

		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, Response{
					Status:  "Not Found",
					Message: fmt.Sprintf("Todo with ID %v Not Found", id),
					Data:    map[string]any{},
				})
			}
		}
		todoResp := GetTodoResponse{}
		isActive := true
		if todoDB.IsActive != "1" {
			isActive = false
		}
		todoResp = GetTodoResponse{
			ID:              todoDB.ID,
			ActivityGroupID: strconv.Itoa(todoDB.ActivityGroupID),
			Title:           todoDB.Title,
			IsActive:        isActive,
			Priority:        todoDB.Priority,
			CreatedAt:       todoDB.CreatedAt,
			UpdatedAt:       todoDB.UpdatedAt,
			DeletedAt:       todoDB.DeletedAt,
		}
		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todoResp,
		})
	})

	// Create :
	todoItems.POST("", func(c echo.Context) error {
		// Validating Request
		todo := new(Todo)
		if err := c.Bind(todo); err != nil {
			return err
		}
		if err := c.Validate(todo); err != nil {
			for _, err2 := range err.(validator.ValidationErrors) {
				errorField := ""
				if err2.Field() == "Title" {
					errorField = "title"
				} else if err2.Field() == "ActivityGroupID" {
					errorField = "activity_group_id"
				}
				return c.JSON(http.StatusBadRequest, Response{
					Status:  "Bad Request",
					Message: fmt.Sprintf("%v cannot be null", errorField),
					Data:    map[string]any{},
				})
			}
		}

		// Inserting Todo
		todo.IsActive = "1"
		todo.Priority = "very-high"

		result := db.Create(&todo)
		if result.Error != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "activity_group_id not found",
				Data:    map[string]any{},
			})
		}

		type todoResponse struct {
			ID              int            `gorm:"primarykey" json:"id"`
			ActivityGroupID int            `json:"activity_group_id" validate:"required"`
			Title           string         `json:"title" validate:"required"`
			IsActive        bool           `json:"is_active" default:"1"`
			Priority        string         `json:"priority" default:"very-high"`
			CreatedAt       time.Time      `json:"created_at"`
			UpdatedAt       time.Time      `json:"updated_at"`
			DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"`
		}
		todoResp := todoResponse{}
		isActive := false
		if todo.IsActive == "1" {
			isActive = true
		}
		todoResp = todoResponse{
			ID:              todo.ID,
			ActivityGroupID: todo.ActivityGroupID,
			Title:           todo.Title,
			IsActive:        isActive,
			Priority:        todo.Priority,
			CreatedAt:       todo.CreatedAt,
			UpdatedAt:       todo.UpdatedAt,
			DeletedAt:       todo.DeletedAt,
		}
		return c.JSON(http.StatusCreated, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todoResp,
		})
	})

	// Delete :
	todoItems.DELETE("/:id", func(c echo.Context) error {
		// Finding Todo
		id := c.Param("id")
		todoDB := new(Todo)
		db.First(&todoDB, id)
		idInt, _ := strconv.Atoi(id)
		if int(todoDB.ID) != idInt {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Todo with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		// Deleting Todo
		db.Delete(&todoDB)
		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    map[string]any{},
		})
	})

	// Update :
	todoItems.PATCH("/:id", func(c echo.Context) error {
		// Finding Todo
		id := c.Param("id")
		todoDB := new(Todo)
		res := db.First(&todoDB, id)
		// idInt, _ := strconv.Atoi(id)
		if res.Error != nil {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Todo with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}
		if strconv.Itoa(todoDB.ID) != id {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Todo with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		// Updating Todo with req
		todo := new(Todo)
		if err := c.Bind(todo); err != nil {

		}
		if todo.ActivityGroupID != 0 {
			todoDB.ActivityGroupID = todo.ActivityGroupID
		}
		if todo.IsActive != "" {
			todoDB.IsActive = todo.IsActive
		}
		if todo.Priority != "" {
			todoDB.Priority = todo.Priority
		}

		// Validating Request
		if todo.ActivityGroupID == 0 {
			todo.ActivityGroupID = todoDB.ActivityGroupID
		}

		if todo.Title != "" {
			todoDB.Title = todo.Title
		}
		db.Save(todoDB)

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todoDB,
		})
	})

	e.Logger.Fatal(e.Start(":3030"))
}
