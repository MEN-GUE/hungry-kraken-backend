package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ResenasCollection *mongo.Collection

// GET: Obtener reseñas con Lookup
func GetResenas(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	restauranteID := r.URL.Query().Get("restaurante_id")
	objID, err := primitive.ObjectIDFromHex(restauranteID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "restaurante_id", Value: objID}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "usuarios"},
			{Key: "localField", Value: "usuario_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "usuario"},
		}}},
		{{Key: "$unwind", Value: "$usuario"}},
		{{Key: "$project", Value: bson.D{
			{Key: "comentario", Value: 1},
			{Key: "calificacion", Value: 1},
			{Key: "fecha_resena", Value: 1},
			{Key: "usuario.nombre_completo", Value: 1},
		}}},
	}

	cursor, err := ResenasCollection.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// POST: Crear nueva reseña
func CreateResena(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var data struct {
		RestauranteID string `json:"restaurante_id"`
		UsuarioID     string `json:"usuario_id"`
		Comentario    string `json:"comentario"`
		Calificacion  int    `json:"calificacion"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	restID, _ := primitive.ObjectIDFromHex(data.RestauranteID)
	usrID, _ := primitive.ObjectIDFromHex(data.UsuarioID)

	nuevaResena := bson.M{
		"restaurante_id": restID,
		"usuario_id":     usrID,
		"comentario":     data.Comentario,
		"calificacion":   data.Calificacion,
		"fecha_resena":   time.Now(),
	}

	_, err := ResenasCollection.InsertOne(ctx, nuevaResena)
	if err != nil {
		http.Error(w, "Error al crear reseña", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"mensaje": "Reseña publicada"})
}
