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

func main() {
	// CORRECCIÓN: Busca el .env en la carpeta actual
	_ = godotenv.Load(".env")
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("Error: MONGO_URI no definido. Asegúrate de tener el archivo .env en la raíz del proyecto.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("hungry_kraken_db")
	usuariosColl := db.Collection("usuarios")
	restaurantesColl := db.Collection("restaurantes")
	ordenesColl := db.Collection("ordenes")

	fmt.Println("Obteniendo usuarios y restaurantes existentes...")

	// Obtener IDs de usuarios (Límite de 50 para no saturar memoria)
	cursorUsr, _ := usuariosColl.Find(ctx, bson.M{}, options.Find().SetLimit(50).SetProjection(bson.M{"_id": 1}))
	var usuarios []bson.M
	cursorUsr.All(ctx, &usuarios)

	// Obtener restaurantes con su menú
	cursorRest, _ := restaurantesColl.Find(ctx, bson.M{})
	var restaurantes []bson.M
	cursorRest.All(ctx, &restaurantes)

	if len(usuarios) == 0 || len(restaurantes) == 0 {
		log.Fatal("Necesitas tener usuarios y restaurantes en la BD antes de correr este script.")
	}

	fmt.Println("Generando 500 órdenes aleatorias...")
	var ordenes []interface{}
	rand.Seed(time.Now().UnixNano())
	metodosPago := []string{"tarjeta_credito", "efectivo", "puntos"}

	for i := 0; i < 500; i++ {
		usr := usuarios[rand.Intn(len(usuarios))]
		rest := restaurantes[rand.Intn(len(restaurantes))]

		// Extraer menú de forma segura
		menuBruto, ok := rest["menu"].(primitive.A)
		if !ok || len(menuBruto) == 0 {
			continue // Saltar si no tiene menú
		}

		// Seleccionar 1 o 2 platillos aleatorios
		numItems := rand.Intn(2) + 1
		itemsPedido := []bson.M{}
		var totalOrden float64

		for j := 0; j < numItems; j++ {
			platillo, ok := menuBruto[rand.Intn(len(menuBruto))].(primitive.M)
			if !ok {
				continue
			}

			precio, _ := platillo["precio"].(float64)
			cantidad := rand.Intn(3) + 1
			subtotal := precio * float64(cantidad)

			itemsPedido = append(itemsPedido, bson.M{
				"item_id":         platillo["item_id"],
				"nombre":          platillo["nombre"],
				"cantidad":        cantidad,
				"precio_unitario": precio,
				"subtotal":        subtotal,
			})
			totalOrden += subtotal
		}

		if len(itemsPedido) == 0 {
			continue
		}

		// Generar fecha aleatoria en el último mes
		diasAtras := time.Duration(rand.Intn(30)) * 24 * time.Hour
		fechaPedido := time.Now().Add(-diasAtras)

		orden := bson.M{
			"usuario_id":     usr["_id"],
			"restaurante_id": rest["_id"],
			"estado":         "entregado", // Estado necesario para las agregaciones de ventas
			"items":          itemsPedido,
			"total_orden":    totalOrden,
			"metodo_pago":    metodosPago[rand.Intn(len(metodosPago))],
			"fecha_pedido":   fechaPedido,
		}
		ordenes = append(ordenes, orden)
	}

	result, err := ordenesColl.InsertMany(ctx, ordenes)
	if err != nil {
		log.Fatal("Error insertando órdenes:", err)
	}

	fmt.Printf("¡Se insertaron %d órdenes exitosamente!\n", len(result.InsertedIDs))
}
