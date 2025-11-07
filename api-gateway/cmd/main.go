package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/logger"

	_ "github.com/FelipeFelipeRenan/goverse/api-gateway/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Goverse API (Unificada)
// @version 1.0
// @description Documentação unificada dos microsserviços do Goverse.
// @termsOfService http://swagger.io/terms/

// @contact.name Felipe Renan
// @contact.email feliperenanqwerty@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	logger.Init()

	mux := http.NewServeMux()

	// 1. A única rota que este serviço terá: servir a UI do Swagger
	//    Isso usa os arquivos em /api-gateway/docs/
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// 2. Um redirect amigável da raiz "/" para a documentação
	mux.Handle("/", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))

	port := os.Getenv("DOCS_PORT")
	if port == "" {
		port = "8090" // Uma porta interna padrão
	}

	logger.Info.Info("Serviço de Documentação (Swagger) rodando", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error.Error("Erro ao iniciar o serviço de documentação", "err", err)
	}
}
