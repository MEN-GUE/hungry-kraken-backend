package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MEN-GUE/hungry-kraken-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var RestaurantesCollection *mongo.Collection

// GET: Todos los restaurantes (Catálogo)
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
		filter["nombre"] = bson.M{"$regex": nombre, "$options": "i"}
	}
	if categoria != "" {
		filter["categoria"] = categoria
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "calificacion_promedio", Value: -1}})
	opts.SetProjection(bson.M{"menu": 0})
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

// GET: Un solo restaurante con su menú (Perfil)
func GetRestauranteByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	idStr := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var restaurante models.Restaurante
	err = RestaurantesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&restaurante)
	if err != nil {
		http.Error(w, "Restaurante no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurante)
}

// POST: Crear nuevo restaurante (GridFS Integrado y Ubicación GeoJSON)
func CreateRestaurante(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var data struct {
		Nombre         string  `json:"nombre"`
		Categoria      string  `json:"categoria"`
		ImagenPerfilID string  `json:"imagen_perfil_id"`
		Latitud        float64 `json:"latitud"`
		Longitud       float64 `json:"longitud"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	// Construimos el objeto GeoJSON exacto como lo diseñaron en el PDF
	ubicacion := bson.M{
		"type": "Point",
		// OJO: MongoDB requiere el orden [longitud, latitud] para índices 2dsphere
		"coordinates": []float64{data.Longitud, data.Latitud},
	}

	nuevoRestaurante := bson.M{
		"nombre":                data.Nombre,
		"categoria":             data.Categoria,
		"calificacion_promedio": 0.0,
		"menu":                  []bson.M{},
		"ubicacion":             ubicacion,
	}

	if data.ImagenPerfilID != "" {
		objID, err := primitive.ObjectIDFromHex(data.ImagenPerfilID)
		if err == nil {
			nuevoRestaurante["imagen_perfil_id"] = objID
		}
	}

	result, err := RestaurantesCollection.InsertOne(ctx, nuevoRestaurante)
	if err != nil {
		http.Error(w, "Error al crear: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"mensaje": "Creado exitosamente", "id": result.InsertedID})
}

// POST: Agregar platillo al menú ($push) (Ahora recibe descripción)
func AddMenuItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var data struct {
		RestauranteID string  `json:"restaurante_id"`
		Nombre        string  `json:"nombre"`
		Descripcion   string  `json:"descripcion"`
		Precio        float64 `json:"precio"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	objID, _ := primitive.ObjectIDFromHex(data.RestauranteID)

	nuevoItem := bson.M{
		"item_id":     primitive.NewObjectID(),
		"nombre":      data.Nombre,
		"descripcion": data.Descripcion,
		"precio":      data.Precio,
		"disponible":  true,
	}

	update := bson.M{"$push": bson.M{"menu": nuevoItem}}
	_, err := RestaurantesCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)

	if err != nil {
		http.Error(w, "Error al agregar platillo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Platillo agregado al menú"})
}
