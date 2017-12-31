# TaskManager
Backend for TaskManager (https://github.com/ArtsiomPiashchynski/TaskManager) written in GO

# Auth

Auth microservice with consul, go-kit. Create jwt token after login request, optioinally can save it in consul k/v storage

# Vault

Vault microservice to generate hashes from passwords and then compare

# Registration

Registration microservice using to register users, also connects to vault service to generate hashes (and after compare) from passwords to store in db

# Authorization

Authorization microservice using to login (return jwt token auth, for now only 1 token)


All services contains /health endpoint to connect to consul and /metrics endpoint to pull metrics to Prometheus
