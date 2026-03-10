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

	db := client.Database("hungry_kraken_db")
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		log.Fatal("Error inicializando GridFS: ", err)
	}

	// Inicializar colecciones
	handlers.RestaurantesCollection = db.Collection("restaurantes")
	handlers.ResenasCollection = db.Collection("resenas")
	handlers.OrdenesCollection = db.Collection("ordenes")
	handlers.UsuariosCollection = db.Collection("usuarios")

	// *** NUEVO: Pasar el client para las transacciones ACID ***
	handlers.MongoClient = client

	// Rutas GridFS
	gridfsHandler := &handlers.GridFSHandler{Bucket: bucket}
	http.HandleFunc("/api/upload", enableCORS(gridfsHandler.UploadImage))
	http.HandleFunc("/api/image", enableCORS(gridfsHandler.GetImage))

	// Rutas Restaurantes (CRUD)
	http.HandleFunc("/api/restaurantes", enableCORS(handlers.GetRestaurantes))
	http.HandleFunc("/api/restaurante", enableCORS(handlers.GetRestauranteByID))
	http.HandleFunc("/api/restaurantes/precio", enableCORS(handlers.PutPrecioPlatillo))        // PUT - UpdateOne
	http.HandleFunc("/api/restaurantes/descuento", enableCORS(handlers.PutDescuentoCategoria)) // PUT - UpdateMany

	// Rutas Menú ($push / $pull)
	http.HandleFunc("/api/menu", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.PostMenuItem(w, r)
		case http.MethodDelete:
			handlers.DeleteMenuItem(w, r)
		default:
			http.Error(w, "Método no soportado", http.StatusMethodNotAllowed)
		}
	}))

	// Ruta Checkout (Transacción ACID)
	http.HandleFunc("/api/checkout", enableCORS(handlers.PostCheckout)) // POST

	// Rutas Reseñas
	http.HandleFunc("/api/resenas", enableCORS(handlers.GetResenas))
	http.HandleFunc("/api/resena", enableCORS(handlers.DeleteResena))                // DELETE - DeleteOne
	http.HandleFunc("/api/resenas/masivo", enableCORS(handlers.DeleteResenasMasivo)) // DELETE - DeleteMany

	// Reportes
	http.HandleFunc("/api/reportes/mejores-restaurantes", enableCORS(handlers.GetMejoresRestaurantes))
	http.HandleFunc("/api/reportes/restaurantes-mas-ventas", enableCORS(handlers.GetRestaurantesMasVentas))
	http.HandleFunc("/api/reportes/usuarios-mas-activos", enableCORS(handlers.GetUsuariosMasActivos))

	fmt.Println("🚀 Servidor API corriendo en http://localhost:8080")
	fmt.Println("   --- GridFS ---")
	fmt.Println("   POST   /api/upload")
	fmt.Println("   GET    /api/image?id=<id>")
	fmt.Println("   --- Restaurantes ---")
	fmt.Println("   GET    /api/restaurantes")
	fmt.Println("   GET    /api/restaurante?id=<id>")
	fmt.Println("   PUT    /api/restaurantes/precio      (UpdateOne: precio de platillo)")
	fmt.Println("   PUT    /api/restaurantes/descuento   (UpdateMany: descuento por categoría)")
	fmt.Println("   --- Menú ($push / $pull) ---")
	fmt.Println("   POST   /api/menu   (agregar platillo)")
	fmt.Println("   DELETE /api/menu   (quitar platillo)")
	fmt.Println("   --- Checkout (ACID) ---")
	fmt.Println("   POST   /api/checkout")
	fmt.Println("   --- Reseñas ---")
	fmt.Println("   GET    /api/resenas?restaurante_id=<id>")
	fmt.Println("   DELETE /api/resena          (DeleteOne)")
	fmt.Println("   DELETE /api/resenas/masivo  (DeleteMany: spam de 1 estrella)")

	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("El servidor falló: ", err)
	}
}
