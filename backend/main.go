package main;

import (
  //"fmt"
	"net/http"
	"log"
	"github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/go-chi/render"
	"github.com/go-chi/cors"
	"database/sql"
	"./services"
	"./storage"
)

const ServerAddr = "localhost:3000";
//const connString = "postgresql://julian@localhost:26257/dbprueba?ssl=true"
const connString = "postgresql://root@localhost:26257?sslcert=certs%5Cclient.root.crt&sslkey=certs%5Cclient.root.key&sslmode=verify-full&sslrootcert=certs%5Cca.crt"
//const connString = "postgresql://julian@localhost:26257/dbprueba?ssl=true&sslmode=require&sslrootcert=certs/ca.crt&sslkey=certs/client.julian.key&sslcert=certs/client.julian.crt";

func Routes(db *sql.DB) *chi.Mux {
	router := chi.NewRouter();
	cors := cors.New(cors.Options{
    // AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
    AllowedOrigins:   []string{"*"},
    // AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           300, // Maximum value not ignored by any of major browsers
  })
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
		cors.Handler,
	);

  router.Route("/v1", func(r chi.Router) {
    r.Mount("/test", services.Routes(db));
  });
  return router;
};

func main() {
	db := storage.ConnectionBD(connString);
  router := Routes(db);

  walkFunc := func(method string, route string, handler http.Handler, middleware ...func(http.Handler) http.Handler) error {
    log.Printf("%s %s\n", method, route);
    return nil;
  };
  if err := chi.Walk(router, walkFunc); err != nil {
    log.Panicf("Logging err: %s\n", err.Error());
  };

  log.Fatal(http.ListenAndServe(ServerAddr, router));
}
