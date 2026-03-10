package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MEN-GUE/hungry-kraken-backend/handlers"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	_ = godotenv.Load()
	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		log.Fatal("ERROR: MONGO_URI no está definido en el .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("🐙 ¡Conectado a MongoDB Atlas!")

	// 🔑 CLAVE PARA EL CHECKOUT: Pasamos el cliente global a handlers para la Transacción ACID
	handlers.MongoClient = client

	db := client.Database("hungry_kraken_db")
	bucket, _ := gridfs.NewBucket(db)

	handlers.RestaurantesCollection = db.Collection("restaurantes")
	handlers.ResenasCollection = db.Collection("resenas")
	handlers.OrdenesCollection = db.Collection("ordenes")
	handlers.UsuariosCollection = db.Collection("usuarios")

	gridfsHandler := &handlers.GridFSHandler{Bucket: bucket}

	// --- RUTAS GET Y GRIDFS ---
	http.HandleFunc("/api/upload", enableCORS(gridfsHandler.UploadImage))
	http.HandleFunc("/api/image", enableCORS(gridfsHandler.GetImage))
	http.HandleFunc("/api/restaurantes", enableCORS(handlers.GetRestaurantes))
	http.HandleFunc("/api/restaurante", enableCORS(handlers.GetRestauranteByID))
	http.HandleFunc("/api/resenas", enableCORS(handlers.GetResenas))
	http.HandleFunc("/api/reportes/mejores-restaurantes", enableCORS(handlers.GetMejoresRestaurantes))
	http.HandleFunc("/api/reportes/restaurantes-mas-ventas", enableCORS(handlers.GetRestaurantesMasVentas))
	http.HandleFunc("/api/reportes/usuarios-mas-activos", enableCORS(handlers.GetUsuariosMasActivos))

	// --- RUTAS POST (Creaciones) ---
	http.HandleFunc("/api/restaurantes/crear", enableCORS(handlers.CreateRestaurante))
	http.HandleFunc("/api/menu/agregar", enableCORS(handlers.AddMenuItem))
	http.HandleFunc("/api/resenas/crear", enableCORS(handlers.CreateResena))
	http.HandleFunc("/api/checkout", enableCORS(handlers.PostCheckout)) // <-- TRANSACCIÓN ACID

	// --- RUTAS PUT (Actualizaciones) ---
	http.HandleFunc("/api/restaurantes/precio", enableCORS(handlers.PutPrecioPlatillo))
	http.HandleFunc("/api/restaurantes/descuento", enableCORS(handlers.PutDescuentoCategoria))

	// --- RUTAS DELETE (Eliminaciones) ---
	http.HandleFunc("/api/menu", enableCORS(handlers.DeleteMenuItem))
	http.HandleFunc("/api/resena", enableCORS(handlers.DeleteResena))
	http.HandleFunc("/api/resenas/masivo", enableCORS(handlers.DeleteResenasMasivo))

	fmt.Println("🚀 Servidor API corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
