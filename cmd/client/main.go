package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NikolosHGW/goph-keeper/internal/client/command"
	"github.com/NikolosHGW/goph-keeper/internal/client/entity"
	"github.com/NikolosHGW/goph-keeper/internal/client/infrastructure/config"
	"github.com/NikolosHGW/goph-keeper/internal/client/service"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
)

func main() {
	config := config.NewConfig()

	myLogger, err := logger.NewLogger("info")
	if err != nil {
		log.Fatalf("ошибка инициализации логгер: %v", err)
	}

	grpcClient, err := service.NewGRPCClient(config.GetServerAddress(), myLogger, config.GetRootCertPath())
	if err != nil {
		myLogger.LogInfo("Ошибка инициализации gRPC клиента", err)
		os.Exit(1)
	}
	defer func() {
		err := grpcClient.Close()
		if err != nil {
			myLogger.LogInfo("не удалось закрыть соединение клиента gRPC", err)
		}
	}()

	tokenHolder := &entity.TokenHolder{}

	authService := service.NewAuthService(grpcClient, myLogger)
	dataService := service.NewDataService(grpcClient, myLogger)

	commands := []command.Command{
		command.NewRegisterCommand(authService, tokenHolder, os.Stdin, os.Stdout),
		command.NewLoginCommand(authService, tokenHolder, os.Stdin, os.Stdout),
		command.NewAddCommand(dataService, tokenHolder, os.Stdin, os.Stdout),
		command.NewGetCommand(dataService, tokenHolder, os.Stdin, os.Stdout),
		command.NewUpdateCommand(dataService, tokenHolder, os.Stdin, os.Stdout),
		command.NewDeleteCommand(dataService, tokenHolder, os.Stdin, os.Stdout),
	}

	commandNames := make([]string, len(commands))

	commandMap := make(map[string]command.Command)
	for i, cmd := range commands {
		commandMap[cmd.Name()] = cmd
		commandNames[i] = cmd.Name()
	}

	fmt.Println("Доступные команды: ", strings.Join(commandNames, ", "))
	for {
		fmt.Print("Введите команду: ")
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			myLogger.LogInfo("Ошибка ввода команды", err)
		}

		cmd, exists := commandMap[input]
		if !exists {
			fmt.Println("Неизвестная команда:", input)
			continue
		}

		err = cmd.Execute()
		if err != nil {
			myLogger.LogInfo("Ошибка вызова команды", err)
		}
	}
}
