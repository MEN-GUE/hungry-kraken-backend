package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// 1. Cargar las variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: No se encontró archivo .env, usando variables del sistema")
	}

	// 2. Obtener la URI de conexión
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("ERROR: La variable MONGO_URI no está configurada. Revisa tu archivo .env")
	}

	// 3. Configurar el contexto con un timeout de 10 segundos para no quedarnos colgados
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Se asegura de cancelar el contexto al salir de la función

	// 4. Crear el cliente y conectar a MongoDB Atlas
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error creando el cliente de MongoDB: ", err)
	}

	// Buena práctica: desconectar el cliente cuando la aplicación se apague
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Error desconectando de MongoDB: ", err)
		}
	}()

	// 5. Hacer "Ping" para verificar que la conexión realmente funciona a través de la red
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("No se pudo conectar al clúster de Atlas. Verifica tu IP en Atlas, Usuario y Contraseña: ", err)
	}

	// Si llegamos aquí, ¡la conexión fue un éxito total!
	fmt.Println("=================================================")
	fmt.Println("🐙 ¡El Hungry Kraken se ha conectado a MongoDB Atlas!")
	fmt.Println("=================================================")

	// Aquí más adelante inicializaremos los endpoints de nuestra API,
	// la instancia de GridFS y el servidor web...
}
