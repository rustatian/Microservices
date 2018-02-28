# TaskManager
Backend for TaskManager (https://github.com/ArtsiomPiashchynski/TaskManager) written in GO, hosted in http://0xdev.me.

# Authorization

Authorization microservice with consul, go-kit. Create jwt token after login request, optioinally can save it in consul k/v storage, also have heath checks

# Vault

Vault microservice to generate hashes from passwords and then compare

# Registration

Registration microservice using to register users, also connects to vault service to generate hashes (and after compare) from passwords to store in db

# Task calendar

Microservice that would be used for adding, showing and deleting tasks from task (or events) calendar


All services contains /health endpoint to connect to consul and /metrics endpoint to pull metrics to Prometheus


Sample of counsul health checks you can see at:
http://207.154.217.167:8500/ui/#/dc1/services

Sample of grafana graphics (with prometheus) you can see at (login test, password test):
http://207.154.217.167:3000/dashboard/db/taskcalendar
