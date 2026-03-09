package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Cargar el .env (como el script corre desde la raíz, buscará el .env ahí)
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: No se encontró archivo .env local, usando variables del sistema")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("Error: MONGO_URI no está definido en el .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Seleccionamos nuestra base de datos y la colección de usuarios
	db := client.Database("hungry_kraken_db")
	usuariosCollection := db.Collection("usuarios")

	fmt.Println("⏳ Generando 50,000 documentos en memoria...")

	// 2. Crear un slice para almacenar los 50,000 documentos
	var usuariosDocs []interface{}

	// Bucle para generar data falsa pero estructurada
	for i := 1; i <= 50000; i++ {
		nuevoUsuario := bson.M{
			"nombre_completo": fmt.Sprintf("Usuario Kraken %d", i),
			"email":           fmt.Sprintf("usuario%d@hungrykraken.com", i),
			"telefono":        fmt.Sprintf("+502 5555%04d", i%10000),
			"fecha_registro":  time.Now().Add(-time.Duration(i) * time.Hour),
			"direcciones": []bson.M{
				{
					"etiqueta":         "Casa",
					"direccion_linea1": fmt.Sprintf("Zona %d, Ciudad de Guatemala", (i%20)+1),
					"coordenadas": bson.M{
						"type":        "Point",
						"coordinates": []float64{-90.5 + (float64(i) * 0.00001), 14.6 + (float64(i) * 0.00001)},
					},
				},
			},
		}
		usuariosDocs = append(usuariosDocs, nuevoUsuario)
	}

	fmt.Println("🚀 Ejecutando operación BULK (InsertMany) hacia Atlas...")

	// 3. Inserción masiva (Bulkwrite)
	inicio := time.Now()
	_, err = usuariosCollection.InsertMany(ctx, usuariosDocs)
	if err != nil {
		log.Fatal("Error en la inserción masiva: ", err)
	}
	duracion := time.Since(inicio)

	fmt.Println("=====================================================")
	fmt.Printf("✅ ¡ÉXITO! 50,000 usuarios insertados en %v\n", duracion)
	fmt.Println("=====================================================")
}
