package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/NikolosHGW/goph-keeper/api/authpb"
	"github.com/NikolosHGW/goph-keeper/api/datapb"
	"github.com/NikolosHGW/goph-keeper/api/registerpb"
	"github.com/NikolosHGW/goph-keeper/internal/server/handler"
	"github.com/NikolosHGW/goph-keeper/internal/server/infrastructure/config"
	"github.com/NikolosHGW/goph-keeper/internal/server/infrastructure/db"
	"github.com/NikolosHGW/goph-keeper/internal/server/infrastructure/repository"
	"github.com/NikolosHGW/goph-keeper/internal/server/interceptor"
	"github.com/NikolosHGW/goph-keeper/internal/server/service"
	"github.com/NikolosHGW/goph-keeper/internal/server/usecase"
	"github.com/NikolosHGW/goph-keeper/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(fmt.Errorf("не удалось запустить сервер: %w", err))
	}
}

func run() error {
	config := config.NewConfig()

	myLogger, err := logger.NewLogger("info")
	if err != nil {
		return fmt.Errorf("не удалось инициализировать логгер: %w", err)
	}

	database, err := db.InitDB(config.GetDatabaseURI(), &db.DBConnector{}, &db.Migrator{})
	if err != nil {
		return fmt.Errorf("не удалось инициализировать базу данных: %w", err)
	}

	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			myLogger.LogInfo("ошибка при закрытии базы данных: ", closeErr)
		}
	}()

	userRepo := repository.NewUser(database, myLogger)
	dataRepo := repository.NewDataRepository(database, myLogger)

	registerService := service.NewRegister(myLogger)
	tokenService := service.NewToken(myLogger, config.GetSecretKey())
	encryptionService := service.NewEncryptionService([]byte(config.GetCryptoKeyPath()))
	dataService := service.NewDataService(dataRepo, encryptionService)

	registerUsecase := usecase.NewRegister(registerService, tokenService, userRepo)
	authUsecase := usecase.NewAuth(tokenService, userRepo)

	listen, err := net.Listen("tcp", config.GetRunAddress())
	if err != nil {
		return fmt.Errorf("не удалось прослушать TCP: %w", err)
	}

	noAuthMethods := []string{
		"/register.Register/RegisterUser",
		"/auth.Auth/LoginUser",
	}

	creds, err := credentials.NewServerTLSFromFile(config.GetServerCrtPath(), config.GetServerKeyPath())
	if err != nil {
		return fmt.Errorf("не удалось загрузить TLS сертификаты: %w", err)
	}

	srv := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			interceptor.NewAuthInterceptor(tokenService, noAuthMethods).Unary(),
		),
	)

	reflection.Register(srv)

	registerpb.RegisterRegisterServer(srv, handler.NewRegisterServer(registerUsecase))
	authpb.RegisterAuthServer(srv, handler.NewAuthServer(authUsecase))
	datapb.RegisterDataServiceServer(srv, handler.NewDataServer(dataService, myLogger))

	errChan := make(chan error, 1)

	go func() {
		myLogger.LogStringInfo("Запуск сервера", "address", config.GetRunAddress())
		if err := srv.Serve(listen); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errChan <- fmt.Errorf("ошибка при запуске сервера: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		myLogger.LogStringInfo("Получен сигнал завершения, отключаем сервер...", "address", config.GetRunAddress())
	case err := <-errChan:
		myLogger.LogInfo("Сервер завершился с ошибкой: ", err)

		return fmt.Errorf("горутина с запуском сервера вернула ошибку: %w", err)
	}

	srv.GracefulStop()

	myLogger.LogStringInfo("Сервер успешно остановлен", "address", config.GetRunAddress())

	return nil
}
