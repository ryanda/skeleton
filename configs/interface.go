package configs

import (
	"net/http"
	"time"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type (
	Base struct {
		Id        string         `gorm:"type:varchar(36);primary_key"`
		CreatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
		UpdatedAt time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
		CreatedBy string         `gorm:"type:varchar(36);default:null"`
		UpdatedBy string         `gorm:"type:varchar(36);default:null"`
		DeletedAt gorm.DeletedAt `gorm:"default:null;index"`
		DeletedBy string         `gorm:"type:varchar(36);default:null"`
	}

	Model interface {
		TableName() string
		SetCreatedBy(user *User)
		SetUpdatedBy(user *User)
		SetDeletedBy(user *User)
		IsSoftDelete() bool
	}

	Service interface {
		Name() string
		Create(value interface{}, id string) error
		Update(value interface{}, id string) error
		Bind(value interface{}, id string) error
		Delete(value interface{}, id string) error
	}

	Module interface {
		Consume()
	}

	Server interface {
		RegisterGRpc(server *grpc.Server)
		RegisterAutoMigrate()
		RegisterQueueConsumer()
	}

	Middleware interface {
		Attach(request *http.Request, response http.ResponseWriter) bool
	}

	Application interface {
		Run()
	}
)
