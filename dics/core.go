package dics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	amqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	configs "github.com/crowdeco/skeleton/configs"
	events "github.com/crowdeco/skeleton/events"
	handlers "github.com/crowdeco/skeleton/handlers"
	interfaces "github.com/crowdeco/skeleton/interfaces"
	"github.com/crowdeco/skeleton/middlewares"
	"github.com/crowdeco/skeleton/paginations"
	"github.com/crowdeco/skeleton/routes"
	"github.com/crowdeco/skeleton/utils"
	"github.com/gadelkareem/cachita"
	"github.com/sarulabs/dingo/v4"
	"github.com/sirupsen/logrus"
	mongodb "github.com/weekface/mgorus"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var Core = []dingo.Def{
	{
		Name: "core:event:dispatcher",
		Build: func(
			todo events.Listener,
		) (*events.Dispatcher, error) {
			return events.NewDispatcher([]events.Listener{todo}), nil
		},
		Params: dingo.Params{
			"0": dingo.Service("module:todo:listener:search"),
		},
	},
	{
		Name: "core:interface:database",
		Build: func(
			todo configs.Server,
		) (*interfaces.Database, error) {
			database := interfaces.Database{
				Servers: []configs.Server{
					todo,
				},
			}

			return &database, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("module:todo:server"),
		},
	},
	{
		Name: "core:interface:grpc",
		Build: func(
			todo configs.Server,
			server *grpc.Server,
			dispatcher *events.Dispatcher,
		) (*interfaces.GRpc, error) {
			grpc := interfaces.GRpc{
				GRpc:       server,
				Dispatcher: dispatcher,
			}

			grpc.Register([]configs.Server{
				todo,
			})

			return &grpc, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("module:todo:server"),
		},
	},
	{
		Name: "core:interface:queue",
		Build: func(
			todo configs.Server,
		) (*interfaces.Queue, error) {
			queue := interfaces.Queue{
				Servers: []configs.Server{
					todo,
				},
			}

			return &queue, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("module:todo:server"),
		},
	},
	{
		Name:  "core:interface:rest",
		Build: (*interfaces.Rest)(nil),
		Params: dingo.Params{
			"Middleware": dingo.Service("core:handler:middleware"),
			"Router":     dingo.Service("core:handler:router"),
			"Server":     dingo.Service("core:http:mux"),
			"Context":    dingo.Service("core:context:background"),
		},
	},
	{
		Name: "core:handler:logger",
		Build: func() (*handlers.Logger, error) {
			logger := logrus.New()
			logger.SetFormatter(&logrus.JSONFormatter{})

			mongodb, err := mongodb.NewHooker(fmt.Sprintf("%s:%d", configs.Env.MongoDbHost, configs.Env.MongoDbPort), configs.Env.MongoDbName, "logs")
			if err == nil {
				logger.AddHook(mongodb)
			} else {
				return nil, err
			}

			return &handlers.Logger{
				Logger: logger,
			}, nil
		},
	},
	{
		Name:  "core:handler:messager",
		Build: (*handlers.Messenger)(nil),
		Params: dingo.Params{
			"Logger":    dingo.Service("core:handler:logger"),
			"Publisher": dingo.Service("core:message:publisher"),
			"Consumer":  dingo.Service("core:message:consumer"),
		},
	},
	{
		Name:  "core:handler:handler",
		Build: (*handlers.Handler)(nil),
		Params: dingo.Params{
			"Dispatcher": dingo.Service("core:event:dispatcher"),
			"Context":    dingo.Service("core:context:background"),
		},
	},
	{
		Name: "core:handler:middleware",
		Build: func(
			auth configs.Middleware,
		) (*handlers.Middleware, error) {
			return &handlers.Middleware{
				Middlewares: []configs.Middleware{
					auth,
				},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("core:middleware:auth"),
		},
	},
	{
		Name: "core:handler:router",
		Build: func(
			gateway configs.Router,
			mux configs.Router,
		) (*handlers.Router, error) {
			return &handlers.Router{
				Routes: []configs.Router{
					gateway,
					mux,
				},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("core:router:gateway"),
			"1": dingo.Service("core:router:mux"),
		},
	},
	{
		Name:  "core:middleware:auth",
		Build: (*middlewares.Auth)(nil),
	},
	{
		Name:  "core:router:gateway",
		Build: (*routes.GRpcGateway)(nil),
	},
	{
		Name:  "core:router:mux",
		Build: (*routes.MuxRouter)(nil),
	},
	{
		Name: "core:http:mux",
		Build: func() (*http.ServeMux, error) {
			return http.NewServeMux(), nil
		},
	},
	{
		Name: "core:grpc:server",
		Build: func() (*grpc.Server, error) {
			return grpc.NewServer(), nil
		},
	},
	{
		Name: "core:context:background",
		Build: func() (context.Context, error) {
			return context.Background(), nil
		},
	},
	{
		Name: "core:gorm:db",
		Build: func() (*gorm.DB, error) {
			return configs.Database, nil
		},
	},
	{
		Name: "core:message:config",
		Build: func() (amqp.Config, error) {
			address := fmt.Sprintf("amqp://%s:%s@%s:%d/", configs.Env.AmqpUser, configs.Env.AmqpPassword, configs.Env.AmqpHost, configs.Env.AmqpPort)

			return amqp.NewDurableQueueConfig(address), nil
		},
	},
	{
		Name: "core:message:publisher",
		Build: func(config amqp.Config) (*amqp.Publisher, error) {
			publisher, err := amqp.NewPublisher(config, watermill.NewStdLogger(configs.Env.Debug, configs.Env.Debug))
			if err != nil {
				return nil, err
			}

			return publisher, nil
		},
	},
	{
		Name: "core:message:consumer",
		Build: func(config amqp.Config) (*amqp.Subscriber, error) {
			consumer, err := amqp.NewSubscriber(config, watermill.NewStdLogger(false, false))
			if err != nil {
				return nil, err
			}

			return consumer, nil
		},
	},
	{
		Name:  "core:pagination:paginator",
		Build: (*paginations.Pagination)(nil),
	},
	{
		Name:  "core:cache:memory",
		Build: (*utils.Cache)(nil),
		Params: dingo.Params{
			"Pool": dingo.Service("core:cachita:cache"),
		},
	},
	{
		Name:  "core:number:formatter",
		Build: (*utils.Number)(nil),
	},
	{
		Name:  "core:string:formatter",
		Build: (*utils.Word)(nil),
	},
	{
		Name: "core:cachita:cache",
		Build: func() (cachita.Cache, error) {
			return cachita.Memory(), nil
		},
	},
}
