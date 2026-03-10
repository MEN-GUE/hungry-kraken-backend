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

// Función para habilitar CORS y permitir que el frontend se conecte
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Permitir conexiones desde cualquier origen
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Permitir los métodos HTTP que usaremos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Manejar peticiones "pre-vuelo" (OPTIONS) que hacen los navegadores por seguridad
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continuar con la petición original hacia el endpoint
		next(w, r)
	}
}

func main() {
	// 1. Cargar .env
	_ = godotenv.Load()
	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		log.Fatal("ERROR: MONGO_URI no está definido en el .env")
	}

	// 2. Conectar a Atlas
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

	// 3. Inicializar GridFS y Bases de Datos
	db := client.Database("hungry_kraken_db")
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		log.Fatal("Error inicializando GridFS: ", err)
	}

	// Inicializar las colecciones para que los endpoints de José Pablo las encuentren
	handlers.RestaurantesCollection = db.Collection("restaurantes")
	handlers.ResenasCollection = db.Collection("resenas")
	handlers.OrdenesCollection = db.Collection("ordenes")
	handlers.UsuariosCollection = db.Collection("usuarios")

	// 4. Configurar las rutas HTTP (Nuestra API) ENVUELTAS EN CORS
	gridfsHandler := &handlers.GridFSHandler{Bucket: bucket}

	// Endpoints tuyos (GridFS)
	http.HandleFunc("/api/upload", enableCORS(gridfsHandler.UploadImage))
	http.HandleFunc("/api/image", enableCORS(gridfsHandler.GetImage))

	// Endpoints de José Pablo (y el tuyo nuevo de traer 1 solo)
	http.HandleFunc("/api/restaurantes", enableCORS(handlers.GetRestaurantes))
	http.HandleFunc("/api/restaurante", enableCORS(handlers.GetRestauranteByID)) // <-- LA NUEVA RUTA
	http.HandleFunc("/api/resenas", enableCORS(handlers.GetResenas))
	http.HandleFunc("/api/reportes/mejores-restaurantes", enableCORS(handlers.GetMejoresRestaurantes))
	http.HandleFunc("/api/reportes/restaurantes-mas-ventas", enableCORS(handlers.GetRestaurantesMasVentas))
	http.HandleFunc("/api/reportes/usuarios-mas-activos", enableCORS(handlers.GetUsuariosMasActivos))

	// 5. Encender el servidor
	fmt.Println("🚀 Servidor API corriendo en http://localhost:8080")
	fmt.Println("   - POST /api/upload (Para subir imagen)")
	fmt.Println("   - GET  /api/image?id=<imagen_id> (Para ver imagen)")
	fmt.Println("   - GET  /api/restaurantes (Catálogo con filtros y skip/limit)")
	fmt.Println("   - GET  /api/restaurante?id=<id> (Traer 1 restaurante con Menú)")
	fmt.Println("   - GET  /api/resenas?restaurante_id=<id> (Lookups)")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("El servidor falló: ", err)
	}
}
