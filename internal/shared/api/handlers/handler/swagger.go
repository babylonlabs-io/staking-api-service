package handler

import (
	"github.com/babylonlabs-io/staking-api-service/docs"
	"github.com/swaggo/swag"
	"net/http"
)

func SwaggerDoc(w http.ResponseWriter, req *http.Request) {
	doc, err := swag.ReadDoc(docs.SwaggerInfo.InstanceName())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(doc))
}
