// Go package
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
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
		ID         int            `gorm:"primarykey" json:"id"`
		ActivityID int            `json:"activity_group_id" validate:"required"`
		Title      string         `json:"title" validate:"required"`
		IsActive   string         `json:"is_active"`
		Priority   string         `json:"priority"`
		CreatedAt  time.Time      `json:"created_at"`
		UpdatedAt  time.Time      `json:"updated_at"`
		DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	}

	TodoFilter struct {
		ActivityID uint `json:"activity_group_id"`
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
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	// Using gorm, connecting to mysql
	dsn := "root:123@tcp(127.0.0.1:3306)/skyshidb?charset=utf8mb4&parseTime=True&loc=Local"
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
		todos := new([]Todo)
		if activityID != "" {
			activityIDInt, _ = strconv.Atoi(activityID)
			db.Where("activity_id=?", activityIDInt).Find(&todos)
		} else {
			db.Find(&todos)
		}

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todos,
		})
	})

	// Get One :
	todoItems.GET("/:id", func(c echo.Context) error {
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

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todoDB,
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
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "title and activity_group_id cannot be null",
				Data:    map[string]any{},
			})
		}

		// Checking if activity id exists
		activityDB := new(Activity)
		db.First(&activityDB, todo.ActivityID)
		if int(activityDB.ID) != todo.ActivityID {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: fmt.Sprintf("Activity with ID %v Not Found", todo.ActivityID),
				Data:    map[string]any{},
			})
		}

		// Inserting Todo
		result := db.Create(&todo)
		if result.Error != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "activity_group_id not found",
				Data:    map[string]any{},
			})
		}
		return c.JSON(http.StatusCreated, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todo,
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
		db.First(&todoDB, id)
		idInt, _ := strconv.Atoi(id)
		if int(todoDB.ID) != idInt {
			return c.JSON(http.StatusNotFound, Response{
				Status:  "Not Found",
				Message: fmt.Sprintf("Todo with ID %v Not Found", id),
				Data:    map[string]any{},
			})
		}

		// Updating Todo with req
		todo := new(Todo)
		if err := c.Bind(todo); err != nil {
			return err
		}
		if todo.ActivityID != 0 {
			todoDB.ActivityID = todo.ActivityID
		}
		if todo.IsActive != "" {
			todoDB.IsActive = todo.IsActive
		}
		if todo.Priority != "" {
			todoDB.Priority = todo.Priority
		}

		// Validating Request
		if todo.ActivityID == 0 {
			todo.ActivityID = todoDB.ActivityID
		}
		if err := c.Validate(todo); err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				Status:  "Bad Request",
				Message: "title cannot be null",
				Data:    map[string]any{},
			})
		}

		todoDB.Title = todo.Title
		db.Save(todoDB)

		return c.JSON(http.StatusOK, Response{
			Status:  "Success",
			Message: "Success",
			Data:    todoDB,
		})
	})

	e.Logger.Fatal(e.Start(":3030"))
}
