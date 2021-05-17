package db

import (
	apiModel "github.com/anchamber/genetics-api/model"
	"github.com/anchamber/genetics-tank/db/model"
)

type Options struct {
	Pageination *apiModel.Pageination
	Filters     []*apiModel.Filter
}

type SystemDB interface {
	Select(Options) ([]*model.System, error)
	SelectByName(name string) (*model.System, error)
	Insert(system *model.System) error
	Update(system *model.System) error
	Delete(name string) error
}

type ErrorCode string

const (
	SystemAlreadyExists ErrorCode = "system already exists"
	Unknown             ErrorCode = "unknown error with db occured"
)
