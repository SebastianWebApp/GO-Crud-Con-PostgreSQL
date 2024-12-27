# Usar una imagen base de Go para compilar la aplicación
FROM golang:1.20-alpine AS builder

# Directorio de trabajo
WORKDIR / 

# Copiar el código fuente al contenedor
COPY . .

# Descargar las dependencias de Go
RUN go mod tidy

# Compilar la aplicación Go
RUN go build -o app .

# Usar una imagen más liviana para ejecutar la aplicación
FROM alpine:latest

# Instalar las dependencias necesarias para ejecutar la aplicación (como libpq para PostgreSQL)
RUN apk --no-cache add ca-certificates postgresql-client

# Directorio de trabajo
WORKDIR /

# Copiar el binario compilado desde la imagen del builder
COPY --from=builder /app .
COPY --from=builder /.env .

# Exponer el puerto en el que se ejecuta la aplicación
EXPOSE 3000

# Comando para ejecutar la aplicación
CMD ["./app"]
