package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GridFSHandler contiene el "Bucket" de GridFS para interactuar con la BD
type GridFSHandler struct {
	Bucket *gridfs.Bucket
}

// 1. ENDPOINT PARA SUBIR IMÁGENES (POST /api/upload)
func (h *GridFSHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Restringimos para que solo acepte método POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido. Usa POST.", http.StatusMethodNotAllowed)
		return
	}

	// Parsear el archivo enviado en el formulario (limite 10MB en memoria)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error al procesar el archivo", http.StatusBadRequest)
		return
	}

	// Obtener el archivo del campo "imagen"
	file, header, err := r.FormFile("imagen")
	if err != nil {
		http.Error(w, "No se encontró el archivo 'imagen' en la petición", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Opciones: Guardar el nombre original del archivo
	opts := options.GridFSUpload().SetMetadata(map[string]string{"filename": header.Filename})

	// Subir el archivo al Bucket de GridFS
	uploadStream, err := h.Bucket.OpenUploadStream(header.Filename, opts)
	if err != nil {
		http.Error(w, "Error al abrir el stream de GridFS", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	// Copiar los datos del archivo al stream de Mongo
	_, err = io.Copy(uploadStream, file)
	if err != nil {
		http.Error(w, "Error guardando el archivo en la base de datos", http.StatusInternalServerError)
		return
	}

	// CORRECCIÓN AQUÍ: FileID es un atributo, no una función (sin paréntesis)
	fileID := uploadStream.FileID.(primitive.ObjectID)

	// Responder al frontend con el ID generado en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensaje":   "Imagen subida exitosamente con GridFS",
		"imagen_id": fileID.Hex(),
	})
}

// 2. ENDPOINT PARA DESCARGAR/VER IMÁGENES (GET /api/image?id=...)
func (h *GridFSHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido. Usa GET.", http.StatusMethodNotAllowed)
		return
	}

	// Obtener el ID de la imagen desde la URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Debes proporcionar el parámetro 'id'", http.StatusBadRequest)
		return
	}

	// Convertir el string a un ObjectID de MongoDB
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "ID de imagen inválido", http.StatusBadRequest)
		return
	}

	// Abrir el stream de descarga desde GridFS
	downloadStream, err := h.Bucket.OpenDownloadStream(objID)
	if err != nil {
		http.Error(w, "No se encontró la imagen en GridFS", http.StatusNotFound)
		return
	}
	defer downloadStream.Close()

	// Enviar los pedazos (chunks) directamente como respuesta HTTP al navegador
	w.Header().Set("Content-Type", "image/jpeg") // Asumimos JPEG por simplicidad
	_, err = io.Copy(w, downloadStream)
	if err != nil {
		http.Error(w, "Error al enviar la imagen", http.StatusInternalServerError)
		return
	}
}
