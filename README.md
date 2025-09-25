# Database Classification API

REST API written in Go that discovers sensitive information in MySQL schemas using configurable regular expressions. The solution stores every configuration, scan, and pattern in a dedicated MySQL metadata database—no MongoDB dependency.

---

## 1. Características Clave
- **Gestión de conexiones**: guarda credenciales cifradas (AES-256-GCM) y valida la conectividad antes de persistir.
- **Escaneo asincrónico**: explora INFORMATION_SCHEMA, clasifica columnas por nombre y calcula riesgo agregado.
- **Patrones configurables**: CRUD en tiempo real sobre regex mediante /api/v1/patterns; las expresiones viven en MySQL y se inicializan desde configs/patterns.json si la tabla está vacía.
- **Persistencia SQL**: tablas database_connections, scan_results, classification_patterns en el esquema classifier_meta (docker/mysql-init.sql).
- **Documentación y pruebas**: colección Postman (postman_collection.json) y guía paso a paso incluida.

---

## 2. Arquitectura
    cmd/api/main.go        // punto de entrada
    internal/
      config               // carga de variables de entorno
      domain               // entidades + interfaces
      infrastructure
        database           // conexiones MySQL (target y metadata)
        http               // enrutador Gin + middlewares
      repository           // persistencia SQL
      service              // casos de uso (database, scan, patterns)
    pkg/
      classifier           // motor de regex y scoring
      security             // cifrado AES-256-GCM
    configs/patterns.json  // semillas de patrones
    docker/mysql-init.sql  // datos de prueba + esquema metadata

---

## 3. Requisitos Previos
- Go 1.21+ (para ejecución local).
- Docker Compose (opción recomendada) o una instancia MySQL 8 con acceso para crear el esquema classifier_meta.
- Variables de entorno listadas en env.example.

---

## 4. Puesta en Marcha Rápida (Docker)
    cp env.example .env                 # opcional, personaliza credenciales
    docker compose up -d                # levanta mysql_test y la API
    # espera ~30 s hasta que MySQL importe seed data
    curl http://localhost:8080/health   # verifica estado

Servicios levantados:
- mysql_test: MySQL con datos ficticios y esquema classifier_meta.
- app: API en http://localhost:8080.

Para apagar:
    docker compose down -v

---

## 5. Ejecución Local (sin Docker)
1. Provisiona MySQL 8 y ejecuta docker/mysql-init.sql manualmente (crea datos de demostración + esquema metadata + usuario metauser/metapass).
2. Configura variables en .env (ver sección 6).
3. Instala dependencias y ejecuta la API:
        go mod tidy
        go run ./cmd/api

---

## 6. Variables de Entorno
Archivo de referencia: env.example.

| Variable | Descripción |
|----------|-------------|
| PORT | Puerto HTTP de la API (default 8080). |
| GIN_MODE | release recomendado para producción. |
| METADATA_DB_HOST | Host del MySQL metadata (por ejemplo mysql_test). |
| METADATA_DB_PORT | Puerto del metadata DB (3306). |
| METADATA_DB_USER / METADATA_DB_PASSWORD | Usuario/clave con permisos sobre classifier_meta. |
| METADATA_DB_NAME | Nombre del esquema metadata (default classifier_meta). |
| METADATA_DB_PARAMS | Parámetros extra (parseTime=true&charset=utf8mb4&loc=UTC). |
| ENCRYPTION_KEY | Cadena exacta de 32 caracteres para AES-256-GCM. |
| JWT_SECRET | Reservado para auth futura (obligatorio por validación). |
| API_VERSION | Prefijo de versión (v1). |
| API_TIMEOUT | Timeout por request (ej. 30s). |

---

## 7. Esquema Metadata (MySQL)
- database_connections: almacena conexiones target (UUID, host, puerto, usuario, password cifrada, timestamps, last_scanned_at).
- scan_results: resultados completos del último escaneo (schemas y summary en columnas JSON, estado, errores, timestamps).
- classification_patterns: regex activos con prioridad, descripción y estado.

Las tablas se crean automáticamente al ejecutar docker/mysql-init.sql (Docker Compose ya lo hace).

---

## 8. Flujo de Trabajo Recomendado
1. Crear conexión: POST /api/v1/database con host/credenciales del target MySQL.
2. Lanzar escaneo: POST /api/v1/database/{databaseId}/scan.
3. Monitorizar: GET /api/v1/scan/{scanId}.
4. Consultar resultados: GET /api/v1/database/{databaseId}/classification y, si se requiere, GET /api/v1/database/{databaseId}/scan/history.
5. Ajustar patrones: CRUD sobre /api/v1/patterns para incorporar nuevos tipos de datos sensibles.

---

## 9. Cobertura de Endpoints
- Health: GET /health.
- Database connections: alta, consulta, listado, actualización, eliminación y prueba (/api/v1/database).
- Scans: iniciar, ver historial, obtener último resultado, obtener detalle por scan, cancelar.
- Patterns: crear, listar, obtener, actualizar y eliminar expresiones regulares activas.

Detalles de payload y respuestas en API_DOCUMENTATION.md.

---

## 10. Pruebas y Verificación
1. Go tooling (ejecutar localmente):
        gofmt ./...
        go test ./...
        go build ./...
2. Colección Postman: importa postman_collection.json, setea baseUrl (default http://localhost:8080) y usa las variables databaseId, scanId, patternId según las respuestas que obtengas.
3. Comprobación manual: revisa las tablas classifier_meta para validar la persistencia (SELECT * FROM database_connections;, etc.).

---