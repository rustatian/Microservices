# Microservices
Samples of different microservices architectures written in GO

# Authorization
Creates jwt token after login request, optioinally can save it in consul k/v storage, also has heath checks

# Vault
Vault microservice is used for generate hashes from passwords and then compare

# Registration
Registration microservice is used for register users, also connects to vault service to generate hashes (and after compare) from passwords to store in db


All services contains /health endpoint to connect to consul and /metrics endpoint to pull metrics to Prometheus
