package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PUT /api/restaurantes/precio
// UpdateOne: Cambia el precio de UN platillo específico dentro del menú embebido
// Body: { "restaurante_id": "...", "item_id": "...", "nuevo_precio": 85.50 }
func PutPrecioPlatillo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido. Usa PUT.", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		RestauranteID string  `json:"restaurante_id"`
		ItemID        string  `json:"item_id"`
		NuevoPrecio   float64 `json:"nuevo_precio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if body.NuevoPrecio <= 0 {
		http.Error(w, "nuevo_precio debe ser mayor a 0", http.StatusBadRequest)
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

	// Usamos el operador posicional "$" para actualizar solo el elemento del array que coincide
	filter := bson.M{
		"_id":          restauranteObjID,
		"menu.item_id": itemObjID,
	}
	update := bson.M{
		"$set": bson.M{
			"menu.$.precio": body.NuevoPrecio,
		},
	}

	result, err := RestaurantesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error al actualizar precio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Restaurante o platillo no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":      "Precio actualizado correctamente",
		"nuevo_precio": body.NuevoPrecio,
	})
}

// PUT /api/restaurantes/descuento
// UpdateMany: Aplica un descuento del 10% a TODOS los platillos de restaurantes
// de una categoría específica (ej: todos los "Mariscos")
// Body: { "categoria": "Mariscos" }
func PutDescuentoCategoria(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido. Usa PUT.", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Categoria  string  `json:"categoria"`
		Descuento  float64 `json:"descuento"` // Ej: 0.10 para 10%
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	if body.Categoria == "" {
		http.Error(w, "Debes proporcionar una 'categoria'", http.StatusBadRequest)
		return
	}

	// Si no mandan descuento, aplicamos 10% por defecto
	descuento := body.Descuento
	if descuento <= 0 || descuento >= 1 {
		descuento = 0.10
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// $mul multiplica el campo por el factor dado: precio * (1 - descuento) = precio con descuento
	// Afecta a TODOS los restaurantes de esa categoría (UpdateMany)
	filter := bson.M{"categoria": body.Categoria}
	update := bson.M{
		"$mul": bson.M{
			"menu.$[].precio": 1 - descuento,
		},
	}

	result, err := RestaurantesCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error aplicando descuento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":               "Descuento aplicado",
		"categoria":             body.Categoria,
		"descuento_aplicado":    descuento * 100,
		"restaurantes_afectados": result.ModifiedCount,
	})
}
