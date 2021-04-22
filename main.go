package main

import (
	"context"
	"log"
	"net/http"
	"opg-search-service/cache"
	"opg-search-service/middleware"
	"opg-search-service/person"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-go-healthcheck/healthcheck"
)

func main() {
	healthcheck.Register("http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check")

	// Create a Logger
	l := log.New(os.Stdout, "opg-search-service ", log.LstdFlags)

	// Create new serveMux
	sm := mux.NewRouter().PathPrefix(os.Getenv("PATH_PREFIX")).Subrouter()

	// swagger:operation GET /health-check health-check
	// Check if the service is up and running
	// ---
	// responses:
	//   '200':
	//     description: Search service is up and running
	//   '404':
	//     description: Not found
	sm.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	secretsCache := cache.New()

	// Create a sub-router for protected handlers
	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.Use(middleware.JwtVerify(secretsCache))

	// Register protected handlers
	iph, err := person.NewIndexHandler(l)
	if err != nil {
		l.Fatal(err)
	}
	// swagger:operation POST /persons post-persons
	// Index one or many Persons
	// ---
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - in: "body"
	//   name: "body"
	//   description: ""
	//   required: true
	//   schema:
	//     type: object
	//     properties:
	//       persons:
	//         type: array
	//         items:
	//           type: object
	//           properties:
	//             uId:
	//               type: string
	//             normalizedUid:
	//               type: integer
	//               format: int64
	//             workPhoneNumber:
	//               type: object
	//               properties:
	//                 id:
	//                   type: integer
	//                   format: int32
	//                 phoneNumber:
	//                   type: string
	//                 type:
	//                   type: string
	//                 default:
	//                   type: boolean
	//                 className:
	//                   type: string
	//             homePhoneNumber:
	//               type: object
	//               properties:
	//                 id:
	//                   type: integer
	//                   format: int32
	//                 phoneNumber:
	//                   type: string
	//                 type:
	//                   type: string
	//                 default:
	//                   type: boolean
	//                 className:
	//                   type: string
	//             mobilePhoneNumber:
	//               type: object
	//               properties:
	//                 id:
	//                   type: integer
	//                   format: int32
	//                 phoneNumber:
	//                   type: string
	//                 type:
	//                   type: string
	//                 default:
	//                   type: boolean
	//                 className:
	//                   type: string
	//             email:
	//               type: string
	//             dob:
	//               type: string
	//             firstname:
	//               type: string
	//             middlenames:
	//               type: string
	//             surname:
	//               type: string
	//             addressLine1:
	//               type: string
	//             addressLine2:
	//               type: string
	//             addressLine3:
	//               type: string
	//             town:
	//               type: string
	//             county:
	//               type: string
	//             postcode:
	//               type: string
	//             country:
	//               type: string
	//             isAirmailRequired:
	//               type: boolean
	//             addresses:
	//               type: array
	//               items:
	//                 type: object
	//                 properties:
	//                   addressLines:
	//                     type: array
	//                     items:
	//                       type: string
	//                   postcode:
	//                     type: string
	//                   className:
	//                     type: string
	//             phoneNumber:
	//               type: string
	//             phoneNumbers:
	//               type: array
	//               items:
	//                 type: object
	//                 properties:
	//                   id:
	//                     type: integer
	//                     format: int32
	//                   phoneNumber:
	//                     type: string
	//                   type:
	//                     type: string
	//                   default:
	//                     type: boolean
	//                   className:
	//                     type: string
	//             personType:
	//               type: string
	//             cases:
	//               type: array
	//               items:
	//                 type: object
	//                 properties:
	//                   uId:
	//                     type: string
	//                   normalizedUid:
	//                     type: integer
	//                     format: int64
	//                   caseRecNumber:
	//                     type: string
	//                   batchId:
	//                     type: string
	//                   className:
	//                     type: string
	//             orders:
	//               type: array
	//               items:
	//                 type: object
	//                 properties:
	//                   order:
	//                     type: object
	//                     properties:
	//                       uId:
	//                         type: string
	//                       normalizedUid:
	//                         type: integer
	//                         format: int64
	//                       caseRecNumber:
	//                         type: string
	//                       batchId:
	//                         type: string
	//                       className:
	//                         type: string
	//                   className:
	//                     type: string
	//             className:
	//               type: string
	// responses:
	//   '202':
	//     description: The request has been handled and individual index responses are included in the response body
	//     schema:
	//       type: object
	//       properties:
	//         results:
	//           type: array
	//           items:
	//             type: object
	//             properties:
	//               id:
	//                 type: integer
	//                 format: int64
	//               statusCode:
	//                 type: integer
	//               message:
	//                 type: string
	//   '400':
	//     description: Request failed validation
	//     schema:
	//       type: object
	//       properties:
	//         message:
	//           type: string
	//         errors:
	//           type: array
	//           items:
	//             type: object
	//             properties:
	//               name:
	//                 type: string
	//               description:
	//                 type: string
	//   '404':
	//     description: Not found
	//   '500':
	//     description: Unexpected error occurred
	postRouter.Handle("/persons", iph)

	sph, err := person.NewSearchHandler(l)
	if err != nil {
		l.Fatal(err)
	}
	postRouter.Handle("/persons/search", sph)

	s := &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     l,                 // Set the logger for the server
		IdleTimeout:  120 * time.Second, // max time fro connections using TCP Keep-Alive
		ReadTimeout:  1 * time.Second,   // max time to read request from the client
		WriteTimeout: 1 * time.Minute,   // max time to write response to the client
	}

	// start the server
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	// Gracefully shutdown when signal received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	sig := <-c
	l.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_ = s.Shutdown(tc)
}
