package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/capgainschristian/go_api_ds/database"
	"github.com/capgainschristian/go_api_ds/models"
	"gorm.io/gorm"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "API is up and running.")
}

func ListCustomers(w http.ResponseWriter, r *http.Request) {
	// Pagination: listcustomers?limit=10&offset=0
	customers := []models.Customer{}

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	// Provide defaults so no input required
	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	result := database.DB.Db.Limit(limit).Offset(offset).Find(&customers)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	jsonResponse, err := json.Marshal(customers)
	if err != nil {
		http.Error(w, "Failed to marshal customers", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = database.DB.Db.Create(&customer).Error
	if err != nil {
		http.Error(w, "Failed to add customer to the database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Customer added successfully."))
}

func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if customer.Email == "" {
		http.Error(w, "Missing customer email", http.StatusBadRequest)
		return
	} else {
		err = database.DB.Db.Where("email = ?", customer.Email).First(&customer).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Customer not found", http.StatusNotFound)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	err = database.DB.Db.Delete(&customer).Error
	if err != nil {
		http.Error(w, "Failed to delete customer from database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer deleted successfully."))

}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	// Representation of the updated info
	var updatedinfo models.Customer

	// Representation of what's currently in the DB
	customer := new(models.Customer)

	err := json.NewDecoder(r.Body).Decode(&updatedinfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Must provide valid email; cannot update email via curl
	if updatedinfo.Email == "" {
		http.Error(w, "Missing customer email", http.StatusBadRequest)
		return
	}

	err = database.DB.Db.Where("email = ?", updatedinfo.Email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	customer.Name = updatedinfo.Name
	customer.Email = updatedinfo.Email
	customer.Address = updatedinfo.Address
	customer.Number = updatedinfo.Number

	err = database.DB.Db.Save(&customer).Error
	if err != nil {
		http.Error(w, "Failed to update customer in database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Customer's information updated successfully."))
}
