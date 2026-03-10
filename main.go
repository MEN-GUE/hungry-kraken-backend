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

func main() {
	// 1. Cargar .env
	_ = godotenv.Load()
	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		log.Fatal("ERROR: MONGO_URI no está definido")
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

	// 3. Inicializar GridFS
	db := client.Database("hungry_kraken_db")
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		log.Fatal("Error inicializando GridFS: ", err)
	}
	
	handlers.RestaurantesCollection = db.Collection("restaurantes")
	handlers.ResenasCollection = db.Collection("resenas")

	// 4. Configurar las rutas HTTP (Nuestra API)
	gridfsHandler := &handlers.GridFSHandler{Bucket: bucket}

	http.HandleFunc("/api/upload", gridfsHandler.UploadImage)
	http.HandleFunc("/api/image", gridfsHandler.GetImage)

	http.HandleFunc("/api/restaurantes", handlers.GetRestaurantes)

	http.HandleFunc("/api/resenas", handlers.GetResenas)

	// 5. Encender el servidor
	fmt.Println("🚀 Servidor API corriendo en http://localhost:8080")
	fmt.Println("   - POST /api/upload (Para subir imagen)")
	fmt.Println("   - GET  /api/image?id=<imagen_id> (Para ver imagen)")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("El servidor falló: ", err)
	}
}
