package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MEN-GUE/hungry-kraken-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MongoClient se debe inicializar desde main.go (igual que las colecciones)
var MongoClient *mongo.Client

// CheckoutRequest define el body que esperamos del frontend
type CheckoutRequest struct {
	UsuarioID     string               `json:"usuario_id"`
	RestauranteID string               `json:"restaurante_id"`
	Items         []models.ItemPedido  `json:"items"`
	MetodoPago    string               `json:"metodo_pago"`
}

// POST /api/checkout
// Transacción ACID: inserta una orden Y actualiza puntos de lealtad del usuario.
// Si cualquier operación falla, se hace rollback automático.
func PostCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido. Usa POST.", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Body inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	usuarioObjID, err := primitive.ObjectIDFromHex(req.UsuarioID)
	if err != nil {
		http.Error(w, "usuario_id inválido", http.StatusBadRequest)
		return
	}

	restauranteObjID, err := primitive.ObjectIDFromHex(req.RestauranteID)
	if err != nil {
		http.Error(w, "restaurante_id inválido", http.StatusBadRequest)
		return
	}

	// Calcular el total sumando los subtotales de cada item
	var totalOrden float64
	for _, item := range req.Items {
		totalOrden += item.Subtotal
	}

	// Los puntos de lealtad son 1 punto por cada Q1 gastado (redondeado)
	puntosGanados := int(totalOrden)

	// --- INICIO DE TRANSACCIÓN ACID ---
	// Usamos WriteConcern "majority" para garantizar durabilidad en el cluster
	wc := writeconcern.Majority()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadPreference(readpref.Primary())

	session, err := MongoClient.StartSession()
	if err != nil {
		http.Error(w, "Error al iniciar sesión de MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.EndSession(context.Background())

	var ordenInsertada *mongo.InsertOneResult

	// WithTransaction ejecuta la función y hace rollback automático si retorna error
	_, txnErr := session.WithTransaction(context.Background(), func(ctx mongo.SessionContext) (interface{}, error) {

		// OPERACIÓN 1: Insertar la nueva orden
		nuevaOrden := models.Orden{
			UsuarioID:     usuarioObjID,
			RestauranteID: restauranteObjID,
			Estado:        "creado",
			ItemsPedido:   req.Items,
			TotalOrden:    totalOrden,
			MetodoPago:    req.MetodoPago,
			FechaPedido:   time.Now(),
		}

		result, err := OrdenesCollection.InsertOne(ctx, nuevaOrden)
		if err != nil {
			return nil, err // <-- ROLLBACK
		}
		ordenInsertada = result

		// OPERACIÓN 2: Sumarle puntos de lealtad al usuario
		// Usamos $inc para incrementar el campo "puntos_lealtad" de forma atómica
		filter := bson.M{"_id": usuarioObjID}
		update := bson.M{
			"$inc": bson.M{"puntos_lealtad": puntosGanados},
		}

		_, err = UsuariosCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, err // <-- ROLLBACK (la orden insertada arriba también se deshace)
		}

		return nil, nil // Todo OK -> COMMIT
	}, txnOpts)

	if txnErr != nil {
		http.Error(w, "Transacción fallida (rollback ejecutado): "+txnErr.Error(), http.StatusInternalServerError)
		return
	}
	// --- FIN DE TRANSACCIÓN ACID ---

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":        "Orden creada exitosamente",
		"orden_id":       ordenInsertada.InsertedID.(primitive.ObjectID).Hex(),
		"total":          totalOrden,
		"puntos_ganados": puntosGanados,
	})
}
