AUTH_SERVER_GRPC_PORT=:8080                 #For TLS
TRAIT_SERVICE_GRPC_PORT=:8084
TRAIT_SERVICE_REST_PORT=:8085
DATABASE_URL=postgresql://<NAME>:<PASSWORD>@<ADDRESS:<PORT>/<DB_NAME>?sslmode=disable
TRAIT_SERVICE_CERT_FILE=certs/server.crt    #Path to server.crt
TRAIT_SERVICE_KEY_FILE=certs/server.key     #Path to server.ket
TRAIT_SERVICE_CA_FILE=certs/ca.crt          #Path to ca.crt