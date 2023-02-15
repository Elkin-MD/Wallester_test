package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"wallester_test/models"
	"wallester_test/storage"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	DB *gorm.DB
}

func Time(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func (r *Repository) CreateCustomer(context *fiber.Ctx) error {
	customer := models.Customer{}

	err1 := context.BodyParser(&customer)

	if err1 != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Request failed",
		})
		return err1
	}

	validate := validator.New()
	err2 := validate.Struct(&customer)
	if err2 != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Validation failed",
		})
		return err2
	}

	now := time.Now()
	from := now.AddDate(-60, 0, 0)
	to := now.AddDate(-18, 0, 0)
	date, err1 := time.Parse("02-01-2006", customer.DateOfBirth)
	if err1 != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Date of birth must be of format DD-MM-YYYY",
		})
		return err1
	}

	err3 := Time(from, to, date)
	if !err3 {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
			"message": "Customer must be older than 18 and younger than 60 years old",
		})
		return nil
	}

	err4 := r.DB.Create(&customer).Error
	if err4 != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not create customer",
		})
		return err4
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customer has been added",
	})
	return nil
}

func (r *Repository) UpdateCustomer(context *fiber.Ctx) error {

	customerModel := models.Customer{}

	id, err1 := context.ParamsInt("id")

	if err1 != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Invalid id format",
		})
		return err1
	}

	r.DB.First(&customerModel, id)

	type UpdateCustomer struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Gender      string `json:"gender"`
		Email       string `json:"e_mail"`
		Address     string `json:"address"`
	}

	updateData := UpdateCustomer{}

	err2 := context.BodyParser(&updateData)
	if err2 != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not parse update data",
		})
		return err2
	}

	validate := validator.New()

	if updateData.FirstName != "" {
		err := validate.Var(updateData.FirstName, "max=100")
		if err != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "First name validation failed",
			})
			return err
		}
		customerModel.FirstName = updateData.FirstName
	}

	if updateData.LastName != "" {
		err := validate.Var(updateData.LastName, "max=100")
		if err != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Last name validation failed",
			})
			return err
		}
		customerModel.LastName = updateData.LastName
	}

	if updateData.DateOfBirth != "" {
		now := time.Now()
		from := now.AddDate(-60, 0, 0)
		to := now.AddDate(-18, 0, 0)
		date, err1 := time.Parse("02-01-2006", updateData.DateOfBirth)
		if err1 != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Date of birth must be of format DD-MM-YYYY",
			})
			return err1
		}
		err2 := Time(from, to, date)
		if !err2 {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Customer must be older than 18 and younger than 60 years old",
			})
			return nil
		}
		customerModel.DateOfBirth = updateData.DateOfBirth
	}

	if updateData.Gender != "" {
		err := validate.Var(updateData.Gender, "oneof=Male Female")
		if err != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Gender validation failed",
			})
			return err
		}
		customerModel.Gender = updateData.Gender
	}

	if updateData.Email != "" {
		err := validate.Var(updateData.Email, "email")
		if err != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Email validation failed",
			})
			return err
		}
		customerModel.Email = updateData.Email
	}

	if updateData.Address != "" {
		err := validate.Var(updateData.Address, "max=200")
		if err != nil {
			context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{
				"message": "Date of birth validation failed",
			})
			return err
		}
		customerModel.Address = updateData.Address
	}

	r.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"first_name": customerModel.FirstName, "last_name": customerModel.LastName,
			"date_of_birth": customerModel.DateOfBirth, "gender": customerModel.Gender,
			"e_mail": customerModel.Email, "address": customerModel.Address,
		}),
	}).Save(&customerModel)

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customer updated successfully",
	})
	return nil
}

func (r *Repository) DeleteCustomer(context *fiber.Ctx) error {
	customerModel := models.Customer{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "ID cannot be empty",
		})
		return nil
	}

	err := r.DB.Delete(customerModel, id)

	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Could not delete customer",
		})
		return err.Error
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customer deleted successfully",
	})
	return nil
}

func (r *Repository) GetCustomers(context *fiber.Ctx) error {
	customerModels := &[]models.Customer{}

	err := r.DB.Find(customerModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not get customers"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customers fetched successfully",
		"data":    customerModels,
	})
	return nil
}

func (r *Repository) GetCustomerByID(context *fiber.Ctx) error {

	id := context.Params("id")
	customerModel := &models.Customer{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "ID cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(customerModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not get the customer"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customer id fetched successfully",
		"data":    customerModel,
	})
	return nil
}

func (r *Repository) GetCustomersByFirstName(context *fiber.Ctx) error {

	FirstName := context.Params("first_name")
	customerModels := &[]models.Customer{}

	err := r.DB.Where("first_name = ?", FirstName).Order("first_name").Find(customerModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not get customers"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customers fetched successfully",
		"data":    customerModels,
	})
	return nil
}

func (r *Repository) GetCustomersByLastName(context *fiber.Ctx) error {

	LastName := context.Params("last_name")
	customerModels := &[]models.Customer{}

	err := r.DB.Where("last_name = ?", LastName).Order("last_name").Find(customerModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Could not get customers"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Customers fetched successfully",
		"data":    customerModels,
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_customer", r.CreateCustomer)
	api.Put("/update_customer/:id", r.UpdateCustomer)
	api.Delete("/delete_customer/:id", r.DeleteCustomer)
	api.Get("/get_customer/:id", r.GetCustomerByID)
	api.Get("/customers", r.GetCustomers)
	api.Get("/get_customers_fn/:first_name", r.GetCustomersByFirstName)
	api.Get("/get_customers_ln/:last_name", r.GetCustomersByLastName)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Could not load the database")
	}
	err = models.MigrateCustomers(db)
	if err != nil {
		log.Fatal("Could not migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
