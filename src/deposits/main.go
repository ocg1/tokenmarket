package main

import (
	"deposits/handlers"
	"deposits/services"
	"fmt"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-plugins/broker/kafka"
	"github.com/micro/go-plugins/registry/zookeeper"
	"github.com/micro/go-plugins/transport/tcp"
	"helpers/database"
	"log"
	protoSchema "protos"
)

// inits microservice
func initService() micro.Service {
	registryAddress := registry.Addrs("token-kafka")
	brokerAddress := broker.Addrs("token-kafka:9092")

	registry := zookeeper.NewRegistry(registryAddress)
	transport := tcp.NewTransport()
	broker := kafka.NewBroker(brokerAddress)

	// Create a new service. Optionally include some options here.
	service := micro.NewService(
		micro.Name("deposits"),
		micro.Version("latest"),
		micro.Registry(registry),
		micro.Transport(transport),
		micro.Broker(broker),
	)

	// Init will parse the command line flags. Any flags set will
	// override the above settings. Options defined here will
	// override anything set on the command line.
	service.Init()

	return service
}

// Inits the shared profile storage (connection to database)
func initDepositsStorage() services.DepositsStorageService {
	db := database.Open()
	logger := &log.Logger{}
	queries := services.GetQueries()
	storage := services.NewDepositsStorageService(db, logger, queries)
	return storage
}

func main() {
	// Storage service init
	storage := initDepositsStorage()
	defer storage.Close()

	service := initService()

	fmt.Println("Deposits microservice is up and running.")

	// Register handler
	protoSchema.RegisterDepositsHandler(service.Server(), handlers.NewDepositsHandler(storage))

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
