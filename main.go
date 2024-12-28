package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Post struct {
	ID          string `json:"ID"`
	Imagen      string `json:"Imagen"`
	Nombre      string `json:"Nombre"`
	Descripcion string `json:"Descripcion"`
}

// Estructura para la respuesta estándar
type Response struct {
	Estado    bool        `json:"Estado"`
	Respuesta interface{} `json:"Respuesta"` // Cambiar a `interface{}` para poder manejar cualquier tipo de dato
}

// Función para cargar las variables de entorno desde el archivo .env
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}
}

// Función para verificar si la base de datos existe, y crearla si no es así
func createDatabaseIfNotExists() error {
	// Cargar las variables de entorno
	loadEnv()

	// Cadena de conexión a PostgreSQL para conectarse a la base de datos por defecto (postgres)
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	// Verificar si la base de datos ya existe usando una consulta preparada con parámetros
	var exists bool
	query := `SELECT 1 FROM pg_database WHERE datname = $1`
	err = db.QueryRow(query, os.Getenv("DB_NAME")).Scan(&exists)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("Error al verificar si la base de datos existe: %v", err)
	}

	if exists {
		// Si la base de datos ya existe, no hacemos nada
		fmt.Printf("La base de datos '%s' ya existe.\n", os.Getenv("DB_NAME"))
		return nil
	}

	// Crear la base de datos si no existe usando parámetros para evitar inyección SQL
	createQuery := `CREATE DATABASE $1`
	_, err = db.Exec(createQuery, os.Getenv("DB_NAME"))
	if err != nil {
		return fmt.Errorf("Error al crear la base de datos: %v", err)
	}

	fmt.Printf("Base de datos '%s' creada exitosamente.\n", os.Getenv("DB_NAME"))
	return nil
}

// Función para conectar con la base de datos
func connectToDB() (*sql.DB, error) {
	// Cargar las variables de entorno
	loadEnv()

	// Cadena de conexión a PostgreSQL usando variables de entorno
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("Error al conectar a la base de datos: %v", err)
	}
	return db, nil
}

// Función para verificar si la tabla existe, y crearla si no es así
func createTableIfNotExists() error {
	// Cargar las variables de entorno
	loadEnv()

	// Cadena de conexión a PostgreSQL para conectarse a la base de datos
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido (puedes agregar más validaciones si es necesario)
	if tableName == "" {
		return fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Intentamos crear la tabla si no existe, con el nombre de tabla proporcionado
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			ID TEXT PRIMARY KEY,
			Imagen TEXT,
			Nombre TEXT,
			Descripcion TEXT
		)`, tableName)

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("Error al crear la tabla: %v", err)
	}

	return nil
}

// Función para guardar el post en la base de datos
func savePostToDB(post Post) error {
	// Cargar las variables de entorno
	loadEnv()
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido
	if tableName == "" {
		return fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Construir la consulta SQL dinámicamente con el nombre de la tabla
	insertQuery := fmt.Sprintf("INSERT INTO %s (ID, Imagen, Nombre, Descripcion) VALUES ($1, $2, $3, $4)", tableName)

	// Ejecutar la consulta SQL
	_, err = db.Exec(insertQuery, post.ID, post.Imagen, post.Nombre, post.Descripcion)
	if err != nil {
		return fmt.Errorf("Error al guardar el post: %v", err)
	}

	return nil
}

// Función para leer todos los posts desde la base de datos
func getAllPostsFromDB() ([]Post, error) {
	// Cargar las variables de entorno
	loadEnv()
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido
	if tableName == "" {
		return nil, fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Construir la consulta SQL dinámicamente con el nombre de la tabla
	query := fmt.Sprintf("SELECT ID, Imagen, Nombre, Descripcion FROM %s", tableName)

	// Ejecutar la consulta SQL
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener los posts: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Imagen, &post.Nombre, &post.Descripcion); err != nil {
			return nil, fmt.Errorf("Error al escanear el post: %v", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error al iterar sobre los posts: %v", err)
	}

	return posts, nil
}

// Función para leer un post desde la base de datos por ID
func getPostByIDFromDB(postID string) (Post, error) {
	// Cargar las variables de entorno
	loadEnv()
	db, err := connectToDB()
	if err != nil {
		return Post{}, err
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido
	if tableName == "" {
		return Post{}, fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Validar que el ID no esté vacío
	if postID == "" {
		return Post{}, fmt.Errorf("El ID de la información no puede estar vacío.")
	}

	// Construir la consulta SQL para obtener un post por ID
	query := fmt.Sprintf("SELECT ID, Imagen, Nombre, Descripcion FROM %s WHERE ID = $1", tableName)

	// Definir una variable para almacenar el post
	var post Post

	// Ejecutar la consulta SQL con el parámetro
	err = db.QueryRow(query, postID).Scan(&post.ID, &post.Imagen, &post.Nombre, &post.Descripcion)
	if err != nil {
		if err == sql.ErrNoRows {
			return Post{}, fmt.Errorf("No se encontró una información con el ID proporcionado")
		}
		return Post{}, fmt.Errorf("Error al obtener el post: %v", err)
	}

	// Retornar el post encontrado
	return post, nil
}

// Función para actualizar un post en la base de datos
func updatePostInDB(post Post) error {
	// Cargar las variables de entorno
	loadEnv()
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido
	if tableName == "" {
		return fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Construir la consulta SQL dinámicamente con el nombre de la tabla
	query := fmt.Sprintf("UPDATE %s SET Imagen = $1, Nombre = $2, Descripcion = $3 WHERE ID = $4", tableName)

	// Ejecutar la consulta SQL
	_, err = db.Exec(query, post.Imagen, post.Nombre, post.Descripcion, post.ID)
	if err != nil {
		return fmt.Errorf("Error al actualizar el post: %v", err)
	}

	return nil
}

// Función para eliminar un post de la base de datos
func deletePostFromDB(post Post) error {
	// Cargar las variables de entorno
	loadEnv()
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Obtener el nombre de la tabla desde las variables de entorno
	tableName := os.Getenv("DB_TABLE")

	// Validar que el nombre de la tabla sea válido
	if tableName == "" {
		return fmt.Errorf("El nombre de la tabla no está definido en las variables de entorno.")
	}

	// Construir la consulta SQL dinámicamente con el nombre de la tabla
	query := fmt.Sprintf("DELETE FROM %s WHERE ID = $1", tableName)

	// Ejecutar la consulta SQL
	_, err = db.Exec(query, post.ID)
	if err != nil {
		return fmt.Errorf("Error al eliminar el post: %v", err)
	}

	return nil
}

// Handler para recibir el webhook (crear un post)
func webhookCreateHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	// Decodificar el cuerpo de la solicitud JSON
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: "Solicitud no válida",
		})
		return
	}

	// Guardar el post en la base de datos
	err = savePostToDB(post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: fmt.Sprintf("Error al guardar la información: %v", err),
		})
		return
	}

	// Enviar una respuesta de éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Estado:    true,
		Respuesta: "Información guardado exitosamente",
	})
}

// Handler para obtener todos los posts
func webhookGetHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := getAllPostsFromDB()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: fmt.Sprintf("Error al obtener la información: %v", err),
		})
		return
	}

	// Codificar los posts en formato JSON y enviar la respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Estado:    true,
		Respuesta: posts,
	})
}

// Handler para obtener solo el ID de un post por ID
func webhookGetPostIDByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del post desde el cuerpo de la solicitud (supuesto que es un JSON)
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: "Solicitud no válida",
		})
		return
	}

	// Llamar a la función para obtener el ID por ID
	postID, err := getPostByIDFromDB(post.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: fmt.Sprintf("Error al obtener el ID: %v", err),
		})
		return
	}

	// Codificar el ID en formato JSON y enviar la respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Estado:    true,
		Respuesta: postID,
	})
}

// Handler para actualizar un post
func webhookUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	// Decodificar el cuerpo de la solicitud JSON
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: "Solicitud no válida",
		})
		return
	}

	// Actualizar el post en la base de datos
	err = updatePostInDB(post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: fmt.Sprintf("Error al actualizar la información: %v", err),
		})
		return
	}

	// Enviar una respuesta de éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Estado:    true,
		Respuesta: "Información actualizado exitosamente",
	})
}

// Handler para eliminar un post
func webhookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var post Post
	// Decodificar el cuerpo de la solicitud JSON
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: "Solicitud no válida",
		})
		return
	}

	// Eliminar el post de la base de datos
	err = deletePostFromDB(post)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Estado:    false,
			Respuesta: fmt.Sprintf("Error al eliminar la información: %v", err),
		})
		return
	}

	// Enviar una respuesta de éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Estado:    true,
		Respuesta: "Información eliminado exitosamente",
	})
}

func main() {
	// Verificar y crear la base de datos si no existe
	err := createDatabaseIfNotExists()
	if err != nil {
		log.Fatalf("Error al verificar o crear la base de datos: %v", err)
	}

	// Crear la tabla si no existe
	err = createTableIfNotExists()
	if err != nil {
		log.Fatalf("Error al crear la tabla: %v", err)
	}

	// Rutas para los webhooks
	http.HandleFunc("/notify", webhookCreateHandler)           // Crear post
	http.HandleFunc("/posts", webhookGetHandler)               // Obtener posts
	http.HandleFunc("/posts_uni", webhookGetPostIDByIDHandler) // Obtener posts
	http.HandleFunc("/update", webhookUpdateHandler)           // Actualizar post
	http.HandleFunc("/delete", webhookDeleteHandler)           // Eliminar post

	// Iniciar el servidor
	log.Println("Servidor receptor escuchando en http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
