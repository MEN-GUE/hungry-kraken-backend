package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MEN-GUE/hungry-kraken-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /api/menu
// Agrega un nuevo platillo al arreglo embebido "menu" de un restaurante usando $push
func PostMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido. Usa POST.", http.StatusMethodNotAllowed)
		return
	}

	// Esperamos: restaurante_id + los campos del nuevo item
	var body struct {
		RestauranteID string `json:"restaurante_id"`
		models.MenuItem
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.RestauranteID)
	if err != nil {
		http.Error(w, "restaurante_id inválido", http.StatusBadRequest)
		return
	}

	// Generamos un nuevo ObjectID para el item
	body.MenuItem.ItemID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// $push agrega el nuevo documento al final del array "menu"
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$push": bson.M{
			"menu": body.MenuItem,
		},
	}

	result, err := RestaurantesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error al agregar platillo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Restaurante no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":  "Platillo agregado al menú",
		"item_id":  body.MenuItem.ItemID.Hex(),
	})
}

// DELETE /api/menu
// Elimina un platillo del arreglo "menu" de un restaurante usando $pull
func DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido. Usa DELETE.", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		RestauranteID string `json:"restaurante_id"`
		ItemID        string `json:"item_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	restauranteObjID, err := primitive.ObjectIDFromHex(body.RestauranteID)
	if err != nil {
		http.Error(w, "restaurante_id inválido", http.StatusBadRequest)
		return
	}

	itemObjID, err := primitive.ObjectIDFromHex(body.ItemID)
	if err != nil {
		http.Error(w, "item_id inválido", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// $pull elimina del array todos los elementos que coincidan con el filtro
	filter := bson.M{"_id": restauranteObjID}
	update := bson.M{
		"$pull": bson.M{
			"menu": bson.M{"item_id": itemObjID},
		},
	}

	result, err := RestaurantesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error al eliminar platillo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Restaurante no encontrado", http.StatusNotFound)
		return
	}

	if result.ModifiedCount == 0 {
		http.Error(w, "Platillo no encontrado en el menú", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje": "Platillo eliminado del menú",
	})
}
