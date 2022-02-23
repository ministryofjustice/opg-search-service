package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-search-service/internal/cache"
	"github.com/ministryofjustice/opg-search-service/internal/cmd"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/middleware"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	personIndex, personConfig, err := person.IndexConfig()
	if err != nil {
		l.Fatal(err)
	}

	firmIndex, firmConfig, err := firm.IndexConfigFirm()
	if err != nil {
		l.Fatal(err)
	}

	secretsCache := cache.New()

	esClient, err := elasticsearch.NewClient(&http.Client{}, l)
	if err != nil {
		l.Fatal(err)
	}
	l.Println("kate just in front of cmd run")
	l.Println(firmIndex)
	cmd.Run(l,
		cmd.NewHealthCheck(l),
		cmd.NewCreateIndices(esClient, firmIndex, firmConfig),
		cmd.NewIndex(l, esClient, secretsCache, firmIndex),
		cmd.NewUpdateAlias(l, esClient, firmIndex),
		cmd.NewCleanupIndices(l, esClient, firmIndex),
		cmd.NewCreateIndices(esClient, personIndex, personConfig),
		cmd.NewIndex(l, esClient, secretsCache, personIndex),
		cmd.NewUpdateAlias(l, esClient, personIndex),
		cmd.NewCleanupIndices(l, esClient, personIndex),
	)

	if err := esClient.CreateIndex(personIndex, personConfig, false); err != nil {
		l.Fatal(err)
	}

	aliasedIndex, err := esClient.ResolveAlias(person.AliasName)
	if err == elasticsearch.ErrAliasMissing {
		if err := esClient.CreateAlias(person.AliasName, personIndex); err != nil {
			l.Fatal(err)
		}
		aliasedIndex = personIndex
	} else if err != nil {
		l.Fatal(err)
	}

	indices := []string{person.AliasName}
	if aliasedIndex != personIndex {
		indices = append(indices, personIndex)
	}

	l.Println("indexing to", indices)

	//firm specific
	if err := esClient.CreateIndex(firmIndex, firmConfig, false); err != nil {
		l.Fatal(err)
	}

	aliasedIndexFirm, err := esClient.ResolveAlias(firm.AliasName)
	if err == elasticsearch.ErrAliasMissing {
		if err := esClient.CreateAlias(firm.AliasName, firmIndex); err != nil {
			l.Fatal(err)
		}
		aliasedIndexFirm = firmIndex
	} else if err != nil {
		l.Fatal(err)
	}

	indicesFirm := []string{firm.AliasName}
	if aliasedIndexFirm != firmIndex {
		indicesFirm = append(indicesFirm, firmIndex)
	}

	l.Println("indexing to", indicesFirm)

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

	// Create a sub-router for protected handlers
	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.Use(middleware.JwtVerify(secretsCache, l))

	// Register protected handlers

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
	//             id:
	//               type: integer
	//               format: int64
	//             uId:
	//               type: string
	//             normalizedUid:
	//               type: integer
	//               format: int64
	//             sageId:
	//               type: string
	//             caseRecNumber:
	//               type: string
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
	//             companyName:
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
	//                   onlineLpaId:
	//                     type: string
	//                   batchId:
	//                     type: string
	//                   className:
	//                     type: string
	//                   caseType:
	//                     type: string
	//                   caseSubtype:
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
	//         successful:
	//           type: integer
	//           format: int64
	//         failed:
	//           type: integer
	//           format: int64
	//         errors:
	//           type: array
	//           items:
	//             type: string
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
	postRouter.Handle("/persons", person.NewIndexHandler(l, esClient, indices))

	postRouter.Handle("/persons/search", person.NewSearchHandler(l, esClient))

	postRouter.Handle("/firms", firm.NewIndexHandler(l, esClient, indicesFirm))

	postRouter.Handle("/firms/search", firm.NewSearchHandler(l, esClient))

	w := l.Writer()
	defer w.Close()

	s := &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     log.New(w, "", 0), // Set the logger for the server
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
