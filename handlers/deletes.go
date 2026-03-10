package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DELETE /api/resena
// DeleteOne: Borra una reseña específica por su ID
// Body: { "resena_id": "..." }
func DeleteResena(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido. Usa DELETE.", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ResenaID string `json:"resena_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(body.ResenaID)
	if err != nil {
		http.Error(w, "resena_id inválido", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ResenasCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Error al eliminar reseña: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Reseña no encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje": "Reseña eliminada correctamente",
	})
}

// DELETE /api/resenas/masivo
// DeleteMany: Elimina TODAS las reseñas con calificación de 1 estrella (consideradas spam)
// Opcionalmente puede recibir un restaurante_id para limitar el borrado a ese restaurante
// Body (opcional): { "restaurante_id": "..." }
func DeleteResenasMasivo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido. Usa DELETE.", http.StatusMethodNotAllowed)
		return
	}

	// El body es opcional: si viene restaurante_id, borrar solo las de ese restaurante
	var body struct {
		RestauranteID string `json:"restaurante_id"`
	}
	// Ignoramos error de decode porque el body es opcional
	json.NewDecoder(r.Body).Decode(&body)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Filtro base: calificacion == 1 (spam)
	filter := bson.M{"calificacion": 1}

	// Si nos mandaron un restaurante_id válido, lo añadimos al filtro
	if body.RestauranteID != "" {
		objID, err := primitive.ObjectIDFromHex(body.RestauranteID)
		if err != nil {
			http.Error(w, "restaurante_id inválido", http.StatusBadRequest)
			return
		}
		filter["restaurante_id"] = objID
	}

	result, err := ResenasCollection.DeleteMany(ctx, filter)
	if err != nil {
		http.Error(w, "Error en borrado masivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":          "Reseñas spam eliminadas",
		"resenas_borradas": result.DeletedCount,
	})
}
