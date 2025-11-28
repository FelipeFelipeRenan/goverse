package main

import (
	"net/http"
	"os"

	"github.com/FelipeFelipeRenan/goverse/common/pkg/logger"

	_ "github.com/FelipeFelipeRenan/goverse/api-gateway/docs"

	_ "github.com/FelipeFelipeRenan/goverse/api-gateway/doc_generators"
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

// --- DEFINIÇÕES DE SEGURANÇA CORRIGIDAS (Formato Multi-Linha) ---

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description Cookie de autenticação HttpOnly (obtido via /auth/login)

// @securityDefinitions.apikey CsrfAuth
// @in header
// @name X-CSRF-TOKEN
// @description Token CSRF (obtido via /auth/login, necessário para POST/PUT/PATCH/DELETE)
func main() {

	logger.Init("info", "documentation-service")

	mux := http.NewServeMux()

	// A única rota que este serviço terá: servir a UI do Swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// redirect amigável da raiz "/" para a documentação
	mux.Handle("/", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))

	port := os.Getenv("DOCS_PORT")
	if port == "" {
		port = "8090" // Uma porta interna padrão
	}

	logger.Info("Serviço de Documentação (Swagger) rodando", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("Erro ao iniciar o serviço de documentação", "err", err)
	}
}
