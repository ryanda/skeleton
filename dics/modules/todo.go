package modules

import (
	configs "github.com/crowdeco/skeleton/configs"
	events "github.com/crowdeco/skeleton/events"
	"github.com/crowdeco/skeleton/handlers"
	"github.com/crowdeco/skeleton/paginations"
	todos "github.com/crowdeco/skeleton/todos"
	listeners "github.com/crowdeco/skeleton/todos/listeners"
	models "github.com/crowdeco/skeleton/todos/models"
	services "github.com/crowdeco/skeleton/todos/services"
	validations "github.com/crowdeco/skeleton/todos/validations"
	"github.com/crowdeco/skeleton/utils"
	"github.com/sarulabs/dingo/v4"
	"gorm.io/gorm"
)

var Todo = []dingo.Def{
	{
		Name: "module:todo",
		Build: func(
			dispatcher *events.Dispatcher,
			service configs.Service,
			logger *handlers.Logger,
			messenger *handlers.Messenger,
			handler *handlers.Handler,
			validator *validations.Todo,
			cache *utils.Cache,
			paginator *paginations.Pagination,
		) (*todos.TodoModule, error) {
			return todos.NewTodoModule(
				dispatcher,
				service,
				logger,
				messenger,
				handler,
				validator,
				cache,
				paginator,
			), nil
		},
	},
	{
		Name:  "module:todo:server",
		Build: (*todos.Server)(nil),
		Params: dingo.Params{
			"Module": dingo.Service("module:todo"),
		},
	},
	{
		Name:  "module:todo:model",
		Build: (*models.Todo)(nil),
	},
	{
		Name:  "module:todo:validator",
		Build: (*validations.Todo)(nil),
	},
	{
		Name: "module:todo:service",
		Build: func(db *gorm.DB, model *models.Todo) (configs.Service, error) {
			return &services.Service{
				Db:        db,
				TableName: model.TableName(),
			}, nil
		},
	},
	{
		Name:  "module:todo:listener:search",
		Build: (*listeners.TodoSearch)(nil),
	},
}
