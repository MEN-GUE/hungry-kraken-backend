package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MEN-GUE/hungry-kraken-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var RestaurantesCollection *mongo.Collection

func GetRestaurantes(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nombre := r.URL.Query().Get("nombre")
	categoria := r.URL.Query().Get("categoria")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	skip := (page - 1) * limit

	filter := bson.M{}

	if nombre != "" {
		filter["nombre"] = bson.M{
			"$regex":   nombre,
			"$options": "i",
		}
	}

	if categoria != "" {
		filter["categoria"] = categoria
	}

	opts := options.Find()

	opts.SetSort(bson.D{{"calificacion_promedio", -1}})

	opts.SetProjection(bson.M{
		"menu": 0,
	})

	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))

	cursor, err := RestaurantesCollection.Find(ctx, filter, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var restaurantes []models.Restaurante

	if err := cursor.All(ctx, &restaurantes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurantes)
}