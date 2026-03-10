package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var categorias = []string{"Pizza", "Hamburguesas", "Mariscos", "Sushi", "Tacos", "Postres", "Comida China", "Comida India", "Parrillada", "Vegetariano"}
var adjetivos = []string{"Kraken", "Supremo", "Dorado", "Mágico", "Gourmet", "Express", "Fuego", "Oceánico", "Urbano", "Real"}
var nombresBase = []string{"El rincón", "La taberna", "Bistro", "Cocina", "Sabores", "Palacio", "Estación", "Paraíso", "Taller", "Cueva"}

func generarMenu(categoria string) []bson.M {
	menu := []bson.M{}
	for i := 1; i <= 5; i++ {
		precio := float64(rand.Intn(150) + 30) // Precio entre 30 y 180
		item := bson.M{
			"item_id":     primitive.NewObjectID(),
			"nombre":      fmt.Sprintf("%s Especial %d", categoria, i),
			"descripcion": "Delicioso platillo preparado con los mejores ingredientes.",
			"precio":      precio,
			"disponible":  true,
		}
		menu = append(menu, item)
	}
	return menu
}

func main() {
	// CORRECCIÓN: Busca el .env en la carpeta actual desde donde se ejecuta el comando
	_ = godotenv.Load(".env")
	mongoURI := os.Getenv("MONGO_URI")

	if mongoURI == "" {
		log.Fatal("Error: MONGO_URI no definido. Asegúrate de tener el archivo .env en la raíz del proyecto.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("hungry_kraken_db")
	restaurantesColl := db.Collection("restaurantes")

	fmt.Println("Generando 50 restaurantes...")

	var restaurantes []interface{}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 50; i++ {
		cat := categorias[rand.Intn(len(categorias))]
		nombre := fmt.Sprintf("%s del %s %s", nombresBase[rand.Intn(len(nombresBase))], adjetivos[rand.Intn(len(adjetivos))], cat)

		// Coordenadas aleatorias cerca de Ciudad de Guatemala
		lat := 14.6 + (rand.Float64() * 0.1)
		lon := -90.6 + (rand.Float64() * 0.2)

		rest := bson.M{
			"nombre":                nombre,
			"categoria":             cat,
			"calificacion_promedio": float64(rand.Intn(20)+30) / 10.0, // Calificación 3.0 - 5.0
			"menu":                  generarMenu(cat),
			"ubicacion": bson.M{
				"type":        "Point",
				"coordinates": []float64{lon, lat},
			},
		}
		restaurantes = append(restaurantes, rest)
	}

	result, err := restaurantesColl.InsertMany(ctx, restaurantes)
	if err != nil {
		log.Fatal("Error insertando restaurantes:", err)
	}

	fmt.Printf("¡Se insertaron %d restaurantes exitosamente!\n", len(result.InsertedIDs))
}
