package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

	"github.com/joho/godotenv"
)

type Cell struct {
	ID          string   `json:"ID"`
	Admins      []string `json:"admins"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
}

type CellResponse struct {
	*Cell
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

var firebaseClient *firestore.Client = new(firestore.Client)
var apiKey string
var networkId string

var tokenAuth *jwtauth.JWTAuth

func helloworld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(30*time.Second))

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

func main() {
	err := godotenv.Load() // 👈 load .env file
	if err != nil {
		log.Fatal(err)
	}

	tgBotToken := os.Getenv("TG_BOT_TOKEN")
	apiKey = os.Getenv("MERAKI_API_KEY")
	networkId = os.Getenv("MERAKI_NETWORK_ID")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	sa := option.WithCredentialsFile("./merakle-dev-firebase-adminsdk-vlqb4-65dba08b77.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	firebaseClient, err = app.Firestore(ctx)
	if err != nil {
		fmt.Printf("Error connecting to firebase:\n")
		log.Fatalln(err)
	}
	defer firebaseClient.Close()
	fmt.Printf("Firebase Client initialised: %s\n", firebaseClient)

	go StartBot(ctx, tgBotToken, firebaseClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Protected area
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
		})
	})

	r.Route("/cells", func(r chi.Router) {
		r.Get("/my", FindCellsByAdminUsername) // GET /cells/my
		// r.With(paginate).Get("/", ListArticles)
		// r.Post("/", CreateArticle)       // POST /articles
		// r.Get("/search", SearchArticles) // GET /articles/search

		// r.Route("/{articleID}", func(r chi.Router) {
		// 	r.Use(ArticleCtx)            // Load the *Article on the request context
		// 	r.Get("/", GetArticle)       // GET /articles/123
		// 	r.Put("/", UpdateArticle)    // PUT /articles/123
		// 	r.Delete("/", DeleteArticle) // DELETE /articles/123
		// })

		// // GET /articles/whats-up
		// r.With(ArticleCtx).Get("/{articleSlug:[a-z-]+}", GetArticle)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	// http.HandleFunc("/", printMerakiNetworkName)
	// http.HandleFunc("/cells", cellsHandler)
	http.ListenAndServe(":8080", r)
}

func FindCellsByAdminUsername(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "article", "_article_1234")
	incomingUsername := "KonstantinIlinov"
	userCells, _ := fetchCellsByTelegramUsername(ctx, incomingUsername)
	if err := render.RenderList(w, r, CellListResponse(userCells)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func CellListResponse(cells []Cell) []render.Renderer {
	list := []render.Renderer{}
	for _, cell := range cells {
		list = append(list, &cell)
	}
	return list
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func (rd *Cell) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func initFirebase() {

}
