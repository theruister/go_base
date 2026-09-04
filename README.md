To start this is the base of an app. This should be able to easily be forked and used as the basis for future services
where infrastructure is ready to update/go. 
 
It launches a gRPC server with a gRPC-Gateway for REST APIs. 
Handles setting up config and logging
Connects to a postgreSQL database.
Connects to a RabbitMQ messagequeue. 

To start it has a simple 'User' definition to demonstrate basic CRUD.

This service can be run in a number of ways

1) directly as a standalone CLI applicatoin
cd cmd/base
go build base.go 
go run base

2) docker app using the provided Dockerfile.
docker build -t base-app .
docker run base-app

can also expose ports with -p or detach to run in the background.
i.e. docker run -d -p 11000:11000 base-app 

3) docker compose - currently can be run at the base of the repo for just this service. Move the docker-compose.yml
to parent directory and add other services there to package multiple services into one deployable package.
--I usually have a separate tmux pane (I keep it small at the bottom of the window) running 'watch docker ps' to show my running containers when developing.
docker-compose -f docker-compose.yml build
docker-compose -f docker-compose.yml up -d 
docker-compose -f docker-compose.yml down
