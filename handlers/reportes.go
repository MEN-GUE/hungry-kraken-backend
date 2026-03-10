package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var OrdenesCollection *mongo.Collection
var UsuariosCollection *mongo.Collection

// restaurantes mejor calificados
func GetMejoresRestaurantes(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{

		{{Key: "$sort", Value: bson.D{
			{Key: "calificacion_promedio", Value: -1},
		}}},

		{{Key: "$limit", Value: 5}},

		{{Key: "$project", Value: bson.D{
			{Key: "nombre", Value: 1},
			{Key: "categoria", Value: 1},
			{Key: "calificacion_promedio", Value: 1},
		}}},
	}

	cursor, err := RestaurantesCollection.Aggregate(ctx, pipeline)
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

// restaurantes con más ventas
func GetRestaurantesMasVentas(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{

		{{Key: "$match", Value: bson.D{
			{Key: "estado", Value: "entregado"},
		}}},

		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$restaurante_id"},
			{Key: "total_ventas", Value: bson.D{
				{Key: "$sum", Value: "$total_orden"},
			}},
			{Key: "total_ordenes", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "restaurantes"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "restaurante"},
		}}},

		{{Key: "$unwind", Value: "$restaurante"}},

		{{Key: "$project", Value: bson.D{
			{Key: "restaurante", Value: "$restaurante.nombre"},
			{Key: "total_ventas", Value: 1},
			{Key: "total_ordenes", Value: 1},
		}}},

		{{Key: "$sort", Value: bson.D{
			{Key: "total_ventas", Value: -1},
		}}},
	}

	cursor, err := OrdenesCollection.Aggregate(ctx, pipeline)
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

// usuarios más activos
func GetUsuariosMasActivos(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{

		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$usuario_id"},
			{Key: "total_gastado", Value: bson.D{
				{Key: "$sum", Value: "$total_orden"},
			}},
			{Key: "total_pedidos", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "usuarios"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "usuario"},
		}}},

		{{Key: "$unwind", Value: "$usuario"}},

		{{Key: "$project", Value: bson.D{
			{Key: "usuario", Value: "$usuario.nombre_completo"},
			{Key: "total_gastado", Value: 1},
			{Key: "total_pedidos", Value: 1},
		}}},

		{{Key: "$sort", Value: bson.D{
			{Key: "total_pedidos", Value: -1},
		}}},
	}

	cursor, err := OrdenesCollection.Aggregate(ctx, pipeline)
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