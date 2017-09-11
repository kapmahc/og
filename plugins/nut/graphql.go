package nut

import (
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

var _rootMutation = make(graphql.Fields)
var _rootQuery = make(graphql.Fields)

// AddQuery add graphql query
func AddQuery(n string, f *graphql.Field) {
	_rootQuery[n] = f
}

// AddMutation add graphql mutation
func AddMutation(n string, f *graphql.Field) {
	_rootQuery[n] = f
}

func handleGraphql(c *gin.Context) {

}

func init() {
	Router().GET("/graphql", handleGraphql)
}
