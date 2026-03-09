package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// 1. COLECCIÓN: USUARIOS
// ==========================================

type Usuario struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	NombreCompleto string             `bson:"nombre_completo" json:"nombre_completo"`
	Email          string             `bson:"email" json:"email"`
	Telefono       string             `bson:"telefono" json:"telefono"`
	Direcciones    []Direccion        `bson:"direcciones" json:"direcciones"`
	FechaRegistro  time.Time          `bson:"fecha_registro" json:"fecha_registro"`
}

type Direccion struct {
	Etiqueta        string `bson:"etiqueta" json:"etiqueta"`
	DireccionLinea1 string `bson:"direccion_linea1" json:"direccion_linea1"`
	Coordenadas     Point  `bson:"coordenadas" json:"coordenadas"`
}

// ==========================================
// 2. COLECCIÓN: RESTAURANTES (Incluye Menú)
// ==========================================

type Restaurante struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Nombre               string             `bson:"nombre" json:"nombre"`
	Categoria            string             `bson:"categoria" json:"categoria"`
	ImagenPerfilID       primitive.ObjectID `bson:"imagen_perfil_id,omitempty" json:"imagen_perfil_id,omitempty"` // Para GridFS
	Ubicacion            Point              `bson:"ubicacion" json:"ubicacion"`
	Menu                 []MenuItem         `bson:"menu" json:"menu"` // Documentos Embebidos
	CalificacionPromedio float64            `bson:"calificacion_promedio" json:"calificacion_promedio"`
}

type MenuItem struct {
	ItemID      primitive.ObjectID `bson:"item_id,omitempty" json:"item_id,omitempty"`
	Nombre      string             `bson:"nombre" json:"nombre"`
	Descripcion string             `bson:"descripcion" json:"descripcion"`
	Precio      float64            `bson:"precio" json:"precio"`
	Disponible  bool               `bson:"disponible" json:"disponible"`
}

// ==========================================
// 3. COLECCIÓN: ÓRDENES (Transaccional)
// ==========================================

type Orden struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UsuarioID     primitive.ObjectID `bson:"usuario_id" json:"usuario_id"`         // Referencia a Usuario
	RestauranteID primitive.ObjectID `bson:"restaurante_id" json:"restaurante_id"` // Referencia a Restaurante
	Estado        string             `bson:"estado" json:"estado"`                 // ej: creado, preparacion, entregado
	ItemsPedido   []ItemPedido       `bson:"items_pedido" json:"items_pedido"`     // Embebido Histórico
	TotalOrden    float64            `bson:"total_orden" json:"total_orden"`
	MetodoPago    string             `bson:"metodo_pago" json:"metodo_pago"`
	FechaPedido   time.Time          `bson:"fecha_pedido" json:"fecha_pedido"`
}

type ItemPedido struct {
	ItemID         primitive.ObjectID `bson:"item_id" json:"item_id"` // Ref histórica
	Nombre         string             `bson:"nombre" json:"nombre"`
	Cantidad       int                `bson:"cantidad" json:"cantidad"`
	PrecioUnitario float64            `bson:"precio_unitario" json:"precio_unitario"`
	Subtotal       float64            `bson:"subtotal" json:"subtotal"`
}

// ==========================================
// 4. COLECCIÓN: RESEÑAS
// ==========================================

type Resena struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	RestauranteID primitive.ObjectID  `bson:"restaurante_id" json:"restaurante_id"`
	UsuarioID     primitive.ObjectID  `bson:"usuario_id" json:"usuario_id"`
	OrdenID       *primitive.ObjectID `bson:"orden_id,omitempty" json:"orden_id,omitempty"` // Puntero porque es opcional
	Calificacion  int                 `bson:"calificacion" json:"calificacion"`             // 1 a 5
	Comentario    string              `bson:"comentario" json:"comentario"`
	FechaResena   time.Time           `bson:"fecha_resena" json:"fecha_resena"`
}

// ==========================================
// UTILIDADES: GEOJSON
// ==========================================

// Point define la estructura estándar de GeoJSON que MongoDB necesita para índices geoespaciales
type Point struct {
	Type        string    `bson:"type" json:"type"`               // Siempre debe ser "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [Longitud, Latitud]
}
